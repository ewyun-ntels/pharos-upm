package goflow

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/third_party/goflow"
)

var GlobalGoflow Goflow

var UpdateJobInfos = internal.NewMap[UpdateJobInfo]()

type UpdateJobInfo struct {
	Use      bool
	Schedule string
}

type Goflow struct {
	gf         *goflow.GoflowWithMutex
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func (g *Goflow) Start(configPath string, config common.Config) error {
	if !config.Workflow.Use || (g.gf != nil && g.gf.IsRunning()) {
		return nil
	}

	g.gf = goflow.NewWithMutex(goflow.Options{
		Store:        &DataStore{config: config},
		UIPath:       "ui/",
		Streaming:    true,
		ShowExamples: false,
		WithSeconds:  true,
	})
	g.ctx, g.cancelFunc = context.WithCancel(context.Background())

	if err := g.jobLoad(configPath, config); err != nil {
		return err
	}

	go func() {
		if err := g.gf.Run(g.ctx, config); err != nil && err.Error() != "context cancelled" {
			slog.Error("goflow run error", "error", err.Error())
		}
	}()

	return nil
}

func (g *Goflow) Stop() {
	if g.gf == nil || !g.gf.IsRunning() {
		return
	}

	g.cancelFunc()
}

func (g *Goflow) AddJob(jobFunc ...func() *goflow.Job) error {
	if g.gf == nil || !g.gf.IsRunning() {
		return nil
	}

	return g.gf.AddJob(jobFunc...)
}

func (g *Goflow) UpdateJob(jobFunc ...func() *goflow.Job) error {
	if g.gf == nil || !g.gf.IsRunning() {
		return nil
	}

	return g.gf.UpdateJob(jobFunc...)
}

func (g *Goflow) RemoveJob(jobName string) error {
	if g.gf == nil || !g.gf.IsRunning() {
		return nil
	}

	err := g.gf.RemoveJob(jobName)
	if err != nil {
		slog.Error("failed to remove job", "name", jobName, "error", err)
		return err
	}

	return nil
}

func (g *Goflow) jobLoad(configPath string, config common.Config) error {
	if err := g.updateJob(configPath, config); err != nil {
		return err
	}

	if jobs, err := GetJobs(configPath, config); err != nil {
		return err
	} else if err := AddFlows(jobs...); err != nil {
		return err
	}

	return nil
}

func (g *Goflow) updateJob(configPath string, config common.Config) error {
	AddUpdateJobInfo("sqlite-vacuum", UpdateJobInfo{
		Use:      config.Database.SQLite.Backup.Use,
		Schedule: config.Database.SQLite.Backup.Schedule,
	})

	for name, info := range UpdateJobInfos.GetAll() {
		job := Job{ConfigPath: configPath, Config: config}
		if err := job.SetFromName(name); errors.Is(err, sql.ErrNoRows) {
			slog.Warn("not found job", "name", name)
			continue
		} else if err != nil {
			slog.Error("failed to set job from name", "name", name, "error", err)
			return err
		}

		if job.Active == info.Use && job.Schedule == info.Schedule {
			continue
		}

		job.Active = info.Use
		job.Schedule = info.Schedule

		if err := job.Update(false); err != nil {
			return err
		}
	}

	return nil
}

func AddUpdateJobInfo(name string, info UpdateJobInfo) {
	UpdateJobInfos.Set(name, info)
}
