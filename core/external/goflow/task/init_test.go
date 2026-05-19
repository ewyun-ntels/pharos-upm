package task

import (
	"testing"

	"ntels.com/pharos/core/pkg/common"
)

func TestInit_TasksRegistered(t *testing.T) {
	// Test that init() registered all expected tasks
	expectedTasks := []string{
		"sample-task-01",
		"sample-task-02",
		"sqlite-backup",
		"sqlite-vacuum",
	}

	for _, taskName := range expectedTasks {
		t.Run(taskName, func(t *testing.T) {
			if !Exist(taskName) {
				t.Errorf("Expected task '%s' to be registered in init()", taskName)
			}

			task := Get(taskName)
			if task == nil {
				t.Errorf("Expected to retrieve task '%s', got nil", taskName)
			}

			if task.GetName() != taskName {
				t.Errorf("Task name mismatch: expected '%s', got '%s'", taskName, task.GetName())
			}
		})
	}
}

func TestInit_AllTasksRetrievable(t *testing.T) {
	allTasks := Gets()

	if allTasks == nil {
		t.Fatal("Expected tasks map, got nil")
	}

	// Verify that we have at least the expected number of tasks
	expectedMinCount := 4
	if len(allTasks) < expectedMinCount {
		t.Errorf("Expected at least %d tasks, got %d", expectedMinCount, len(allTasks))
	}
}

func TestInit_SampleTask01(t *testing.T) {
	task := Get("sample-task-01")
	if task == nil {
		t.Fatal("Expected sample-task-01 to be registered")
	}

	// Test GetDataFormat returns expected structure
	dataFormat := task.GetDataFormat()
	if dataFormat == nil {
		t.Error("Expected non-nil data format")
	}

	formatMap, ok := dataFormat.(map[string]any)
	if !ok {
		t.Error("Expected data format to be map[string]any")
	} else if val, exists := formatMap["field-01"]; !exists || val != "value-01" {
		t.Error("Expected field-01 in data format")
	}
}

func TestInit_SampleTask02(t *testing.T) {
	task := Get("sample-task-02")
	if task == nil {
		t.Fatal("Expected sample-task-02 to be registered")
	}

	// Test GetDataFormat returns expected type
	dataFormat := task.GetDataFormat()
	if dataFormat == nil {
		t.Error("Expected non-nil data format")
	}

	if _, ok := dataFormat.(string); !ok {
		t.Error("Expected data format to be string")
	}
}

func TestInit_BackupTask(t *testing.T) {
	task := Get("sqlite-backup")
	if task == nil {
		t.Fatal("Expected sqlite-backup to be registered")
	}

	// Test GetDataFormat returns empty string
	dataFormat := task.GetDataFormat()
	if dataFormat != "" {
		t.Errorf("Expected empty data format, got %v", dataFormat)
	}
}

func TestInit_VacuumTask(t *testing.T) {
	task := Get("sqlite-vacuum")
	if task == nil {
		t.Fatal("Expected sqlite-vacuum to be registered")
	}

	// Test GetDataFormat returns empty string
	dataFormat := task.GetDataFormat()
	if dataFormat != "" {
		t.Errorf("Expected empty data format, got %v", dataFormat)
	}
}

func TestInit_TasksAreUnique(t *testing.T) {
	allTasks := Gets()

	taskNames := make(map[string]bool)
	for name := range allTasks {
		if taskNames[name] {
			t.Errorf("Duplicate task name found: %s", name)
		}
		taskNames[name] = true
	}
}

func TestInit_TaskInterfaceCompliance(t *testing.T) {
	allTasks := Gets()

	for name, task := range allTasks {
		t.Run(name, func(t *testing.T) {
			// Verify all interface methods work
			if task.GetName() == "" {
				t.Error("GetName() returned empty string")
			}

			// These should not panic
			task.SetConfig("", common.Config{})
			task.SetData(nil)
			_ = task.GetDataFormat()
		})
	}
}
