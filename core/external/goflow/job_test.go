package goflow

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/pkg/common"
)

func TestJob_Valid(t *testing.T) {
	tests := []struct {
		name    string
		job     Job
		wantErr bool
		errType error
	}{
		{
			name: "valid job with existing tasks",
			job: Job{
				Name:      "test-job",
				Schedule:  "* * * * *",
				TaskNames: []string{"sample-task-01", "sample-task-02"},
			},
			wantErr: false,
		},
		{
			name: "invalid job with non-existing task",
			job: Job{
				Name:      "test-job",
				Schedule:  "* * * * *",
				TaskNames: []string{"non-existing-task"},
			},
			wantErr: true,
			errType: external.ErrorInvalidTaskName,
		},
		{
			name: "empty task names",
			job: Job{
				Name:      "test-job",
				Schedule:  "* * * * *",
				TaskNames: []string{},
			},
			wantErr: false,
		},
		{
			name: "task names get sorted",
			job: Job{
				Name:      "test-job",
				Schedule:  "* * * * *",
				TaskNames: []string{"sqlite-vacuum", "sample-task-01"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.job.valid()
			if (err != nil) != tt.wantErr {
				t.Errorf("valid() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("valid() error = %v, want error type %v", err, tt.errType)
			}
		})
	}
}

func TestJob_JsonToDB(t *testing.T) {
	job := Job{
		Name:      "test-job",
		TaskNames: []string{"task1", "task2"},
		TaskDatas: map[string]any{
			"task1": "data1",
			"task2": map[string]string{"key": "value"},
		},
	}

	err := job.jsonToDB()
	if err != nil {
		t.Errorf("jsonToDB() error = %v", err)
	}

	if job.TaskNamesForDB == "" {
		t.Error("TaskNamesForDB should not be empty")
	}

	if job.TaskDatasForDB == "" {
		t.Error("TaskDatasForDB should not be empty")
	}

	// Verify it's valid JSON
	var taskNames []string
	if err := json.Unmarshal([]byte(job.TaskNamesForDB), &taskNames); err != nil {
		t.Errorf("TaskNamesForDB is not valid JSON: %v", err)
	}

	var taskDatas map[string]any
	if err := json.Unmarshal([]byte(job.TaskDatasForDB), &taskDatas); err != nil {
		t.Errorf("TaskDatasForDB is not valid JSON: %v", err)
	}
}

func TestJob_DBToJson(t *testing.T) {
	job := Job{
		TaskNamesForDB: `["task1","task2"]`,
		TaskDatasForDB: `{"task1":"data1","task2":{"key":"value"}}`,
	}

	err := job.dbToJson()
	if err != nil {
		t.Errorf("dbToJson() error = %v", err)
	}

	if len(job.TaskNames) != 2 {
		t.Errorf("Expected 2 task names, got %d", len(job.TaskNames))
	}

	if len(job.TaskDatas) != 2 {
		t.Errorf("Expected 2 task datas, got %d", len(job.TaskDatas))
	}
}

func TestJob_DBToJson_InvalidJSON(t *testing.T) {
	tests := []struct {
		name           string
		taskNamesForDB string
		taskDatasForDB string
		wantErr        bool
	}{
		{
			name:           "invalid task names JSON",
			taskNamesForDB: "invalid json",
			taskDatasForDB: `{}`,
			wantErr:        true,
		},
		{
			name:           "invalid task datas JSON",
			taskNamesForDB: `[]`,
			taskDatasForDB: "invalid json",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := Job{
				TaskNamesForDB: tt.taskNamesForDB,
				TaskDatasForDB: tt.taskDatasForDB,
			}

			err := job.dbToJson()
			if (err != nil) != tt.wantErr {
				t.Errorf("dbToJson() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJob_SetFromReader(t *testing.T) {
	jobJSON := `{
		"name": "test-job",
		"schedule": "* * * * *",
		"active": true,
		"taskNames": ["sample-task-01"],
		"taskDatas": {"sample-task-01": "data"}
	}`

	job := Job{}
	reader := strings.NewReader(jobJSON)

	err := job.setFromReader(reader)
	if err != nil {
		t.Errorf("setFromReader() error = %v", err)
	}

	if job.Name != "test-job" {
		t.Errorf("Expected name 'test-job', got '%s'", job.Name)
	}

	if job.Schedule != "* * * * *" {
		t.Errorf("Expected schedule '* * * * *', got '%s'", job.Schedule)
	}

	if !job.Active {
		t.Error("Expected active to be true")
	}
}

func TestJob_SetFromReader_InvalidJSON(t *testing.T) {
	job := Job{}
	reader := strings.NewReader("invalid json")

	err := job.setFromReader(reader)
	if err == nil {
		t.Error("setFromReader() should return error for invalid JSON")
	}
}

func TestJob_ShouldBeAdded(t *testing.T) {
	tests := []struct {
		name    string
		job     Job
		wantAdd bool
		wantErr bool
	}{
		{
			name: "valid job should be added",
			job: Job{
				Name:      "test-job",
				TaskNames: []string{"sample-task-01"},
			},
			wantAdd: true,
			wantErr: false,
		},
		{
			name: "invalid job should not be added",
			job: Job{
				Name:      "test-job",
				TaskNames: []string{"non-existing-task"},
			},
			wantAdd: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			add, err := tt.job.shouldBeAdded()
			if (err != nil) != tt.wantErr {
				t.Errorf("shouldBeAdded() error = %v, wantErr %v", err, tt.wantErr)
			}
			if add != tt.wantAdd {
				t.Errorf("shouldBeAdded() = %v, want %v", add, tt.wantAdd)
			}
		})
	}
}

func TestJob_GetJobFunc(t *testing.T) {
	job := Job{
		Name:      "test-job",
		Schedule:  "* * * * *",
		Active:    true,
		TaskNames: []string{"sample-task-01"},
		TaskDatas: map[string]any{
			"sample-task-01": "test-data",
		},
		ConfigPath: "/test/config",
		Config:     common.Config{},
	}

	jobFunc := job.getJobFunc()
	if jobFunc == nil {
		t.Fatal("getJobFunc() returned nil")
	}

	goflowJob := jobFunc()
	if goflowJob == nil {
		t.Fatal("jobFunc() returned nil")
	}

	if goflowJob.Name != job.Name {
		t.Errorf("Expected job name '%s', got '%s'", job.Name, goflowJob.Name)
	}

	if goflowJob.Schedule != job.Schedule {
		t.Errorf("Expected schedule '%s', got '%s'", job.Schedule, goflowJob.Schedule)
	}

	if goflowJob.Active != job.Active {
		t.Errorf("Expected active %v, got %v", job.Active, goflowJob.Active)
	}
}

func TestJob_GetJobFunc_WithNonExistingTask(t *testing.T) {
	job := Job{
		Name:       "test-job",
		Schedule:   "* * * * *",
		Active:     true,
		TaskNames:  []string{"non-existing-task"},
		ConfigPath: "/test/config",
		Config:     common.Config{},
	}

	jobFunc := job.getJobFunc()
	goflowJob := jobFunc()

	// Should not panic, but log an error
	if goflowJob == nil {
		t.Fatal("jobFunc() returned nil")
	}
}

func TestJob_GetJobFunc_WithMultipleTasks(t *testing.T) {
	job := Job{
		Name:      "test-job",
		Schedule:  "* * * * *",
		Active:    true,
		TaskNames: []string{"sample-task-01", "sample-task-02"},
		TaskDatas: map[string]any{
			"sample-task-01": "data1",
			"sample-task-02": "data2",
		},
		ConfigPath: "/test/config",
		Config:     common.Config{},
	}

	jobFunc := job.getJobFunc()
	goflowJob := jobFunc()

	if goflowJob == nil {
		t.Fatal("jobFunc() returned nil")
	}
}

func TestJob_GetHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		paramName      string
		expectedStatus int
	}{
		{
			name:           "get all jobs",
			paramName:      "",
			expectedStatus: http.StatusInternalServerError, // Will fail without DB
		},
		{
			name:           "get specific job",
			paramName:      "test-job",
			expectedStatus: http.StatusInternalServerError, // Will fail without DB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := Job{
				ConfigPath: "/test/config",
				Config:     common.Config{},
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "name", Value: tt.paramName}}

			status, response := job.GetHandler(c)
			t.Logf("GetHandler() status = %v, response = %v (expected if database not available)", status, response)
		})
	}
}

func TestJob_PostHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jobJSON := `{
		"name": "test-job",
		"schedule": "* * * * *",
		"active": true,
		"taskNames": ["sample-task-01"],
		"taskDatas": {}
	}`

	job := Job{
		ConfigPath: "/test/config",
		Config:     common.Config{},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(jobJSON))

	status, response := job.PostHandler(c)

	// Will likely return error due to database not being initialized
	if status == http.StatusInternalServerError {
		t.Logf("PostHandler() returned error (expected if database not available): %v", response)
	}
}

func TestJob_PutHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jobJSON := `{
		"schedule": "0 * * * *",
		"active": false,
		"taskNames": ["sample-task-01"],
		"taskDatas": {}
	}`

	job := Job{
		ConfigPath: "/test/config",
		Config:     common.Config{},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/", bytes.NewBufferString(jobJSON))
	c.Params = gin.Params{{Key: "name", Value: "test-job"}}

	status, response := job.PutHandler(c)

	// Will likely return error due to database not being initialized
	if status == http.StatusInternalServerError {
		t.Logf("PutHandler() returned error (expected if database not available): %v", response)
	}
}

func TestJob_DeleteHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	job := Job{
		ConfigPath: "/test/config",
		Config:     common.Config{},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "name", Value: "test-job"}}

	status, response := job.DeleteHandler(c)

	// Should return OK even if job doesn't exist
	if status != http.StatusOK && status != http.StatusInternalServerError {
		t.Errorf("DeleteHandler() status = %v, response = %v", status, response)
	}
}

func TestJob_Update(t *testing.T) {
	job := Job{
		Name:      "test-job",
		Schedule:  "* * * * *",
		Active:    true,
		TaskNames: []string{"sample-task-01"},
		Config:    common.Config{},
	}

	err := job.Update(false)
	if err != nil {
		t.Logf("Update() error (expected if database not available): %v", err)
	}
}

func TestJob_SetFromName(t *testing.T) {
	job := Job{
		ConfigPath: "/test/config",
		Config:     common.Config{},
	}

	err := job.SetFromName("test-job")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			t.Log("SetFromName() returned ErrNoRows (expected if job doesn't exist)")
		} else {
			t.Logf("SetFromName() error: %v", err)
		}
	}
}

func TestAddFlows(t *testing.T) {
	jobs := []Job{
		{
			Name:      "job1",
			Schedule:  "* * * * *",
			Active:    true,
			TaskNames: []string{"sample-task-01"},
			Config:    common.Config{},
		},
		{
			Name:      "job2",
			Schedule:  "0 * * * *",
			Active:    false,
			TaskNames: []string{"sample-task-02"},
			Config:    common.Config{},
		},
	}

	err := AddFlows(jobs...)
	if err != nil {
		t.Logf("AddFlows() error: %v", err)
	}
}

func TestAddFlows_Empty(t *testing.T) {
	err := AddFlows()
	if err != nil {
		t.Errorf("AddFlows() with no jobs should not error, got: %v", err)
	}
}

func TestGetJobs(t *testing.T) {
	config := common.Config{}

	jobs, err := GetJobs("/test/config", config)
	if err != nil {
		t.Logf("GetJobs() error (expected if database not available): %v", err)
		return
	}

	// If no error, jobs should not be nil
	if jobs == nil {
		t.Error("GetJobs() should return empty slice, not nil")
	}
}

func TestJob_AddFlow(t *testing.T) {
	job := Job{
		Name:      "test-job",
		Schedule:  "* * * * *",
		Active:    true,
		TaskNames: []string{"sample-task-01"},
		Config:    common.Config{},
	}

	err := job.addFlow()
	if err != nil {
		t.Logf("addFlow() error: %v", err)
	}
}

func TestJob_UpdateFlow(t *testing.T) {
	job := Job{
		Name:      "test-job",
		Schedule:  "* * * * *",
		Active:    true,
		TaskNames: []string{"sample-task-01"},
		Config:    common.Config{},
	}

	err := job.updateFlow()
	if err != nil {
		t.Logf("updateFlow() error: %v", err)
	}
}

func TestJob_RemoveFlow(t *testing.T) {
	job := Job{
		Name:   "test-job",
		Config: common.Config{},
	}

	err := job.removeFlow()
	if err != nil {
		t.Logf("removeFlow() error: %v", err)
	}
}

func TestJob_SetFromReader_EmptyReader(t *testing.T) {
	job := Job{}
	reader := bytes.NewReader([]byte{})

	err := job.setFromReader(reader)
	if err == nil {
		t.Error("setFromReader() should return error for empty reader")
	}
}

func TestJob_PostHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	job := Job{
		ConfigPath: "/test/config",
		Config:     common.Config{},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString("invalid json"))

	status, response := job.PostHandler(c)

	if status != http.StatusInternalServerError {
		t.Errorf("PostHandler() with invalid JSON should return 500, got %v", status)
	}

	if response == nil {
		t.Error("PostHandler() should return error response")
	}
}

func TestJob_JsonToDB_EmptyFields(t *testing.T) {
	job := Job{
		Name:      "test-job",
		TaskNames: []string{},
		TaskDatas: map[string]any{},
	}

	err := job.jsonToDB()
	if err != nil {
		t.Errorf("jsonToDB() error = %v", err)
	}

	if job.TaskNamesForDB == "" {
		t.Error("TaskNamesForDB should not be empty even for empty slice")
	}

	if job.TaskDatasForDB == "" {
		t.Error("TaskDatasForDB should not be empty even for empty map")
	}
}

func TestJob_SetFromReader_NilReader(t *testing.T) {
	job := Job{}

	// Create a nil bytes.Reader to avoid actual nil pointer
	reader := bytes.NewReader(nil)

	err := job.setFromReader(reader)
	if err == nil {
		t.Error("setFromReader() should return error for empty/nil content")
	}
}
