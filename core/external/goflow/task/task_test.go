package task

import (
	"context"
	"testing"

	"ntels.com/pharos/core/pkg/common"
)

// MockTask is a mock implementation of Task interface for testing
type MockTask struct {
	name       string
	configPath string
	config     common.Config
	data       any
	dataFormat any
	runFunc    func(context.Context) (any, error)
}

func (m *MockTask) GetName() string {
	return m.name
}

func (m *MockTask) SetConfig(configPath string, config common.Config) {
	m.configPath = configPath
	m.config = config
}

func (m *MockTask) SetData(data any) {
	m.data = data
}

func (m *MockTask) GetDataFormat() any {
	return m.dataFormat
}

func (m *MockTask) Run(ctx context.Context) (any, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx)
	}
	return nil, nil
}

func TestAdd(t *testing.T) {
	// Create a new mock task
	mockTask := &MockTask{
		name:       "test-task",
		dataFormat: "test-format",
	}

	// Add the task
	Add(mockTask)

	// Verify the task was added
	if !Exist("test-task") {
		t.Error("Expected task to exist after adding")
	}

	// Verify we can get the task
	retrievedTask := Get("test-task")
	if retrievedTask == nil {
		t.Fatal("Expected to retrieve task, got nil")
	}

	if retrievedTask.GetName() != "test-task" {
		t.Errorf("Expected task name 'test-task', got '%s'", retrievedTask.GetName())
	}
}

func TestExist(t *testing.T) {
	tests := []struct {
		name     string
		taskName string
		setup    func()
		want     bool
	}{
		{
			name:     "existing task",
			taskName: "existing-task",
			setup: func() {
				Add(&MockTask{name: "existing-task"})
			},
			want: true,
		},
		{
			name:     "non-existing task",
			taskName: "non-existing-task",
			setup:    func() {},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got := Exist(tt.taskName)
			if got != tt.want {
				t.Errorf("Exist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGet(t *testing.T) {
	// Setup: Add a task
	mockTask := &MockTask{
		name:       "get-test-task",
		dataFormat: "test-format",
	}
	Add(mockTask)

	// Test: Get the task
	retrievedTask := Get("get-test-task")
	if retrievedTask == nil {
		t.Fatal("Expected to retrieve task, got nil")
	}

	if retrievedTask.GetName() != "get-test-task" {
		t.Errorf("Expected task name 'get-test-task', got '%s'", retrievedTask.GetName())
	}

	// Test: Get non-existing task
	nonExistingTask := Get("non-existing")
	if nonExistingTask != nil {
		t.Error("Expected nil for non-existing task")
	}
}

func TestGets(t *testing.T) {
	// Setup: Add multiple tasks
	task1 := &MockTask{name: "task-1"}
	task2 := &MockTask{name: "task-2"}

	Add(task1)
	Add(task2)

	// Test: Get all tasks
	allTasks := Gets()
	if allTasks == nil {
		t.Fatal("Expected map of tasks, got nil")
	}

	// Verify at least our test tasks are present
	if _, ok := allTasks["task-1"]; !ok {
		t.Error("Expected 'task-1' to be in the map")
	}

	if _, ok := allTasks["task-2"]; !ok {
		t.Error("Expected 'task-2' to be in the map")
	}
}

func TestMockTask_SetConfig(t *testing.T) {
	mockTask := &MockTask{name: "config-test"}
	config := common.Config{}
	configPath := "/test/config/path"

	mockTask.SetConfig(configPath, config)

	if mockTask.configPath != configPath {
		t.Errorf("Expected configPath '%s', got '%s'", configPath, mockTask.configPath)
	}
}

func TestMockTask_SetData(t *testing.T) {
	mockTask := &MockTask{name: "data-test"}
	testData := map[string]string{"key": "value"}

	mockTask.SetData(testData)

	if mockTask.data == nil {
		t.Error("Expected data to be set, got nil")
	}
}

func TestMockTask_GetDataFormat(t *testing.T) {
	expectedFormat := "test-format"
	mockTask := &MockTask{
		name:       "format-test",
		dataFormat: expectedFormat,
	}

	format := mockTask.GetDataFormat()
	if format != expectedFormat {
		t.Errorf("Expected format '%v', got '%v'", expectedFormat, format)
	}
}

func TestMockTask_Run(t *testing.T) {
	tests := []struct {
		name    string
		runFunc func(context.Context) (any, error)
		wantErr bool
	}{
		{
			name: "successful run",
			runFunc: func(ctx context.Context) (any, error) {
				return "result", nil
			},
			wantErr: false,
		},
		{
			name:    "default run (nil runFunc)",
			runFunc: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTask := &MockTask{
				name:    "run-test",
				runFunc: tt.runFunc,
			}

			ctx := context.Background()
			result, err := mockTask.Run(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.runFunc != nil && result == nil {
				t.Error("Expected result, got nil")
			}
		})
	}
}
