package common

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/robfig/cron/v3"
)

type CronJob struct {
	entryId cron.EntryID
	rule    Rule
}

type Cron struct {
	mutex      sync.Mutex
	Jobs       map[string]CronJob
	controller *cron.Cron
}

var CronInstance Cron

func init() {
	CronInstance = Cron{
		Jobs: make(map[string]CronJob),
	}
}

func CronRun() error {
	return CronInstance.Run()
}

func (c *Cron) AddJob(spec string, rule Rule) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if rule.GetId() == "" {
		return errors.New("job ID is empty")
	}

	entryId, err := c.controller.AddJob(spec, cron.FuncJob(func() {
		err := rule.Execute(context.Background())
		if err != nil {
			slog.Error("cron job failed", "error", err)
			return
		}
	}))
	if err != nil {
		return err
	}

	c.Jobs[rule.GetId()] = CronJob{
		entryId: entryId,
		rule:    rule,
	}

	return nil
}

func (c *Cron) RemoveJob(id string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.Jobs[id]; !ok {
		return errors.New("job not found")
	}

	c.controller.Remove(c.Jobs[id].entryId)
	delete(c.Jobs, id)

	return nil
}

func (c *Cron) HasJob(id string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, ok := c.Jobs[id]
	return ok
}

func (c *Cron) JobIDs() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ids := make([]string, 0, len(c.Jobs))
	for id := range c.Jobs {
		ids = append(ids, id)
	}
	return ids
}

func (c *Cron) RescheduleJob(id string, spec string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	job, ok := c.Jobs[id]
	if !ok {
		return errors.New("job not found")
	}

	c.controller.Remove(job.entryId)

	entryId, err := c.controller.AddJob(spec, cron.FuncJob(func() {
		err := job.rule.Run(context.Background())
		if err != nil {
			slog.Error("cron job failed", "error", err)
			return
		}
	}))
	if err != nil {
		slog.Error("failed to reschedule job", "error", err)
		return err
	}

	job.entryId = entryId

	return nil
}

func (c *Cron) Run() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.controller == nil {
		c.controller = cron.New(cron.WithSeconds())
	}

	go c.controller.Start()

	return nil
}
