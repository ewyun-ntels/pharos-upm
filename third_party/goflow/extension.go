package goflow

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jmoiron/sqlx"
	"maze.io/x/duration"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

type Execution struct {
	ID         string       `json:"id"`
	Job        string       `json:"job"`
	StartTs    orm.Datetime `json:"startTs" db:"start_ts"`
	ModifiedTs orm.Datetime `json:"modifiedTs" db:"modified_ts"`
	State      state        `json:"state" db:"state"`

	Tasks      []taskExecution `json:"tasks" db:"-"`
	TasksForDB string          `json:"-" db:"tasks"`
}

func (e *Execution) Set(value any) (bool, error) {
	switch value.(type) {
	case *execution:
	case executionIndex:
		return false, nil
	default:
		return false, fmt.Errorf("invalid value type : %T", value)
	}

	ex := value.(*execution)

	e.ID = ex.ID.String()
	e.Job = ex.Job
	e.StartTs = orm.Datetime{Time: ex.StartTs}
	e.ModifiedTs = orm.Datetime{Time: ex.ModifiedTs}
	e.State = ex.State
	e.Tasks = ex.Tasks

	if err := e.jsonToDB(); err != nil {
		return false, err
	}

	return true, nil
}

func (e *Execution) GetsFromJob(config common.Config, job string) ([]Execution, error) {
	return e.gets(config, `SELECT * FROM workflow_execution WHERE job = $1;`, job)
}

func (e *Execution) Gets(config common.Config) ([]Execution, error) {
	return e.gets(config, `SELECT * FROM workflow_execution;`)
}

func (e *Execution) Exist(config common.Config) (bool, error) {
	count := 0

	handler := func(db *sqlx.DB) error {
		return db.Get(&count, `SELECT COUNT(*) FROM workflow_execution WHERE id = $1;`, e.ID)
	}
	if err := orm.Handler(orm.DriverDefault, &config.Database, handler); !errors.Is(err, nil) {
		return false, err
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}

func (e *Execution) Insert(config common.Config) error {
	handler := func(db *sqlx.DB) error {
		_, err := db.NamedExec(`
INSERT INTO workflow_execution(id, job, start_ts, modified_Ts, state, tasks) VALUES
(:id, :job, :start_ts, :modified_ts, :state, :tasks);
`, e)
		return err
	}
	if err := orm.Handler(orm.DriverDefault, &config.Database, handler); err != nil {
		return err
	}

	return nil
}

func (e *Execution) Update(config common.Config) error {
	handler := func(db *sqlx.DB) error {
		_, err := db.NamedExec(`
UPDATE workflow_execution SET
	job = :job,
	start_ts = :start_ts,
	modified_ts = :modified_ts,
	state = :state,
	tasks = :tasks
WHERE id = :id;`, e)

		return err
	}
	if err := orm.Handler(orm.DriverDefault, &config.Database, handler); err != nil {
		return err
	}

	return nil
}

func (e *Execution) Upsert(config common.Config) error {
	if exist, err := e.Exist(config); err != nil {
		return err
	} else if exist {
		return e.Update(config)
	}

	return e.Insert(config)
}

func (e *Execution) gets(config common.Config, query string, args ...any) ([]Execution, error) {
	executions := []Execution{}

	handler := func(db *sqlx.DB) error {
		return db.Select(&executions, query, args...)
	}
	if err := orm.Handler(orm.DriverDefault, &config.Database, handler); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if !errors.Is(err, nil) {
		return nil, err
	}

	for index := range executions {
		if err := executions[index].dbToJson(); err != nil {
			return nil, err
		}
	}

	return executions, nil
}

func (e *Execution) dbToJson() error {
	if err := json.Unmarshal([]byte(e.TasksForDB), &e.Tasks); err != nil {
		return err
	}

	return nil
}

func (e *Execution) jsonToDB() error {
	if bytes, err := json.Marshal(e.Tasks); err != nil {
		return err
	} else {
		e.TasksForDB = string(bytes)
	}

	return nil
}

func (g *Goflow) ExistJob(jobName string) bool {
	_, exist := g.Jobs[jobName]
	return exist
}

func (g *Goflow) RemoveJob(jobName string) {
	for job, entryID := range g.cronEntries {
		if job == jobName {
			g.cron.Remove(entryID)
			delete(g.cronEntries, job)
		}
	}

	for index, value := range g.jobs {
		if value == jobName {
			g.jobs[index] = g.jobs[len(g.jobs)-1]
			g.jobs = g.jobs[:len(g.jobs)-1]
			break
		}
	}

	delete(g.Jobs, jobName)
}

func NewWithMutex(opts Options) *GoflowWithMutex {
	goflowMutex := GoflowWithMutex{}

	goflowMutex.gf = New(opts)

	return &goflowMutex
}

type GoflowWithMutex struct {
	mutex sync.Mutex
	gf    *Goflow
	stop  atomic.Bool
}

func (g *GoflowWithMutex) ExistJob(jobName string) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	return g.gf.ExistJob(jobName)
}

func (g *GoflowWithMutex) AddJob(jobFunc ...func() *Job) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	for _, f := range jobFunc {
		job := f()
		if g.gf.ExistJob(job.Name) {
			return errors.New("exist job")
		}
	}

	return g.gf.AddJob(jobFunc...)
}

func (g *GoflowWithMutex) UpdateJob(jobFunc ...func() *Job) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	for _, f := range jobFunc {
		g.gf.RemoveJob(f().Name)
	}

	return g.gf.AddJob(jobFunc...)
}

func (g *GoflowWithMutex) RemoveJob(jobName string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.gf.RemoveJob(jobName)

	return nil
}

func (g *GoflowWithMutex) Run(ctx context.Context, config common.Config) error {
	wg := new(sync.WaitGroup)

	g.stop.Store(false)

	if d, err := duration.ParseDuration(config.Workflow.Execution.TTL); !errors.Is(err, nil) {
		return err
	} else {
		wg.Add(1)
		go ttl(config, &g.stop, wg, int64(d.Seconds()))
	}

	if err := g.gf.Run(ctx); err != nil {
		return err
	}

	g.stop.Store(true)
	wg.Wait()

	return nil
}

func (g *GoflowWithMutex) IsRunning() bool {
	return !g.stop.Load()
}

func ttl(config common.Config, stop *atomic.Bool, wg *sync.WaitGroup, seconds int64) {
	defer wg.Done()

	for {
		if stop.Load() {
			break
		}

		time.Sleep(1 * time.Second)

		if time.Now().UTC().Unix()%seconds != 0 {
			continue
		}

		if err := ttlExecution(config.Database, seconds); err != nil {
			slog.Error("ttlExecution error", "error", err.Error())
		} else {
			slog.Info("ttlExecution run")
		}
	}
}

func ttlExecution(databaseConfig orm.DatabaseConfig, seconds int64) error {
	ttl := time.Now().UTC().Add(-time.Second * time.Duration(seconds))

	handler := func(db *sqlx.DB) error {
		_, err := db.Exec(`DELETE FROM workflow_execution WHERE start_ts < $1`, ttl)
		return err
	}

	if err := orm.Handler(orm.DriverDefault, &databaseConfig, handler); !errors.Is(err, nil) {
		return err
	}

	return nil
}
