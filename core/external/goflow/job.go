package goflow

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/external/goflow/task"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	third_party_goflow "ntels.com/pharos/third_party/goflow"
)

type Job struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Active   bool   `json:"active"`

	TaskNames      []string `json:"taskNames" db:"-"`
	TaskNamesForDB string   `json:"-" db:"task_names"`

	TaskDatas      map[string]any `json:"taskDatas" db:"-"`
	TaskDatasForDB string         `json:"-" db:"task_datas"`

	UpdatedAt orm.Datetime `json:"updatedAt" db:"updated_at"`

	ConfigPath string        `json:"-" db:"-"`
	Config     common.Config `json:"-" db:"-"`
}

func (job *Job) GetHandler(c *gin.Context) (int, any) {
	var response any
	var err error

	name := c.Param("name")

	if len(name) != 0 {
		err = job.SetFromName(name)
		response = job
	} else {
		response, err = GetJobs(job.ConfigPath, job.Config)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return http.StatusOK, nil
	} else if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, response
}

func (job *Job) PostHandler(c *gin.Context) (int, any) {
	if err := job.setFromReader(c.Request.Body); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	if err := orm.Handler(orm.DriverDefault, &job.Config.Database, func(db *sqlx.DB) error {
		job.UpdatedAt = orm.Datetime{Time: time.Now().UTC()}

		_, err := db.NamedExec(`
INSERT INTO workflow_job(name, schedule, active, task_names, task_datas, updated_at) VALUES
(:name, :schedule, :active, :task_names, :task_datas, :updated_at);
`, job)
		return err
	}); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	if err := job.addFlow(); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

func (job *Job) PutHandler(c *gin.Context) (int, any) {
	if err := job.setFromReader(c.Request.Body); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}
	job.Name = c.Param("name")

	if err := job.Update(true); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

func (job *Job) DeleteHandler(c *gin.Context) (int, any) {
	name := c.Param("name")

	if err := job.SetFromName(name); errors.Is(err, sql.ErrNoRows) {
		return http.StatusOK, nil
	} else if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	if err := orm.Handler(orm.DriverDefault, &job.Config.Database, func(db *sqlx.DB) error {
		_, err := db.Exec(`DELETE FROM workflow_job WHERE name = $1`, name)
		return err
	}); errors.Is(err, sql.ErrNoRows) {
		return http.StatusOK, nil
	} else if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	if err := job.removeFlow(); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

func (job *Job) Update(updateFlow bool) error {
	handler := func(db *sqlx.DB) error {
		job.UpdatedAt = orm.Datetime{Time: time.Now().UTC()}

		_, err := db.NamedExec(`
UPDATE workflow_job SET
    schedule = :schedule,
    active = :active,
    task_names = :task_names,
    task_datas = :task_datas,
    updated_at = :updated_at
WHERE name = :name`, job)

		return err
	}
	if err := orm.Handler(orm.DriverDefault, &job.Config.Database, handler); err != nil {
		return err
	}

	if updateFlow {
		if err := job.updateFlow(); err != nil {
			return err
		}
	}

	return nil
}

func (job *Job) addFlow() error {
	if add, err := job.shouldBeAdded(); err != nil {
		return err
	} else if !add {
		slog.Info("job that should not be added", "job", job)
		return nil
	}

	if err := GlobalGoflow.AddJob(job.getJobFunc()); err != nil {
		return err
	}
	slog.Info("add job", "job", job)

	return nil
}

func (job *Job) updateFlow() error {
	if add, err := job.shouldBeAdded(); err != nil {
		return err
	} else if !add {
		slog.Info("job that should not be updated", "job", job)
		return nil
	}

	if err := GlobalGoflow.UpdateJob(job.getJobFunc()); err != nil {
		return err
	}
	slog.Info("update job", "job", job)

	return nil
}

func (job *Job) removeFlow() error {
	if err := GlobalGoflow.RemoveJob(job.Name); err != nil {
		return err
	}
	slog.Info("remove job", "name", job.Name)

	return nil
}

func (job *Job) valid() error {
	sort.Strings(job.TaskNames)

	for _, taskName := range job.TaskNames {
		if !task.Exist(taskName) {
			return external.ErrorInvalidTaskName
		}
	}

	return nil
}

func (job *Job) SetFromName(name string) error {
	if err := orm.Handler(orm.DriverDefault, &job.Config.Database, func(db *sqlx.DB) error {
		return db.Get(job, `SELECT * FROM workflow_job WHERE name = $1`, name)
	}); err != nil {
		return err
	}

	if err := job.dbToJson(); err != nil {
		return err
	}

	return job.valid()
}

func (job *Job) setFromReader(reader io.Reader) error {
	if body, err := io.ReadAll(reader); err != nil {
		return err
	} else if err := json.Unmarshal(body, job); err != nil {
		return err
	}

	if err := job.jsonToDB(); err != nil {
		return err
	}

	return job.valid()
}

func (job *Job) dbToJson() error {
	if err := json.Unmarshal([]byte(job.TaskNamesForDB), &job.TaskNames); err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(job.TaskDatasForDB), &job.TaskDatas); err != nil {
		return err
	}

	return nil
}

func (job *Job) jsonToDB() error {
	if bytes, err := json.Marshal(job.TaskNames); err != nil {
		return err
	} else {
		job.TaskNamesForDB = string(bytes)
	}

	if bytes, err := json.Marshal(job.TaskDatas); err != nil {
		return err
	} else {
		job.TaskDatasForDB = string(bytes)
	}

	return nil
}

func (job *Job) shouldBeAdded() (bool, error) {
	if err := job.valid(); err != nil {
		return false, err
	}

	return true, nil
}

func (job *Job) getJobFunc() func() *third_party_goflow.Job {
	jobFunc := func() *third_party_goflow.Job {
		goflowJob := &third_party_goflow.Job{
			Name:     job.Name,
			Schedule: job.Schedule,
			Active:   job.Active,
		}

		for _, taskName := range job.TaskNames {
			if !task.Exist(taskName) {
				slog.Error("not exist task", "job", job)
				continue
			}

			taskImpl := reflect.New(reflect.ValueOf(task.Get(taskName)).Elem().Type()).Interface().(task.Task)

			taskImpl.SetConfig(job.ConfigPath, job.Config)
			taskImpl.SetData(job.TaskDatas[taskName])
			if err := goflowJob.AddTask(&third_party_goflow.Task{Name: taskImpl.GetName(), Operator: taskImpl}); err != nil {
				slog.Error("add task error", "error", err, "job", job)

			}
		}

		return goflowJob
	}

	return jobFunc
}

func AddFlows(jobs ...Job) error {
	for _, job := range jobs {
		if err := job.addFlow(); err != nil {
			return err
		}
	}

	return nil
}

func GetJobs(configPath string, config common.Config) ([]Job, error) {
	var jobs []Job

	handler := func(db *sqlx.DB) error {
		return db.Select(&jobs, `SELECT * FROM workflow_job`)
	}

	if err := orm.Handler(orm.DriverDefault, &config.Database, handler); errors.Is(err, sql.ErrNoRows) {
		return jobs, nil
	} else if err != nil {
		return nil, err
	}

	for index, job := range jobs {
		jobs[index].ConfigPath = configPath
		jobs[index].Config = config

		if err := jobs[index].dbToJson(); err != nil {
			slog.Error("invalid job", "error", err, "name", job.Name)

			return nil, err
		}

		if err := jobs[index].valid(); err != nil {
			slog.Error("invalid job", "error", err, "name", job.Name)

			return nil, err
		}
	}

	return jobs, nil
}
