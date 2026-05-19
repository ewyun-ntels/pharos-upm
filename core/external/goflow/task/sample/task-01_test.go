package sample

import (
	"context"
	"testing"

	"ntels.com/pharos/core/pkg/common"
)

func TestTask01Task_GetName(t *testing.T) {
	task := &Task01Task{}
	expected := "sample-task-01"

	got := task.GetName()
	if got != expected {
		t.Errorf("GetName() = %v, want %v", got, expected)
	}
}

func TestTask01Task_SetConfig(t *testing.T) {
	task := &Task01Task{}
	configPath := "/test/config/path"
	config := common.Config{}

	task.SetConfig(configPath, config)

	if task.configPath != configPath {
		t.Errorf("Expected configPath '%s', got '%s'", configPath, task.configPath)
	}
}

func TestTask01Task_SetData(t *testing.T) {
	task := &Task01Task{}
	testData := map[string]any{"key": "value"}

	task.SetData(testData)

	if task.data == nil {
		t.Error("Expected data to be set, got nil")
	}
}

func TestTask01Task_GetDataFormat(t *testing.T) {
	task := &Task01Task{}

	format := task.GetDataFormat()

	// Verify the format is a map with expected structure
	formatMap, ok := format.(map[string]any)
	if !ok {
		t.Fatal("Expected GetDataFormat() to return map[string]any")
	}

	if val, exists := formatMap["field-01"]; !exists || val != "value-01" {
		t.Errorf("Expected field-01 to be 'value-01', got %v", val)
	}
}

func TestTask01Task_Run(t *testing.T) {
	task := &Task01Task{}
	ctx := context.Background()

	result, err := task.Run(ctx)

	if err != nil {
		t.Errorf("Run() returned unexpected error: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
}

func TestTask01Task_Interface(t *testing.T) {
	// Verify that Task01Task implements all required methods
	task := &Task01Task{}

	// Test all interface methods exist
	_ = task.GetName()
	task.SetConfig("", common.Config{})
	task.SetData(nil)
	_ = task.GetDataFormat()
	_, _ = task.Run(context.Background())
}

func TestTask01Task_ConfigPersistence(t *testing.T) {
	task := &Task01Task{}
	configPath := "/test/path/config.yaml"
	config := common.Config{}

	task.SetConfig(configPath, config)

	if task.configPath != configPath {
		t.Errorf("configPath not persisted correctly, got %s", task.configPath)
	}
}

func TestTask01Task_DataPersistence(t *testing.T) {
	task := &Task01Task{}
	testData := map[string]string{"test": "data"}

	task.SetData(testData)

	if task.data == nil {
		t.Error("data not persisted correctly, got nil")
	}

	dataMap, ok := task.data.(map[string]string)
	if !ok {
		t.Fatal("data type not preserved")
	}

	if dataMap["test"] != "data" {
		t.Error("data content not preserved")
	}
}

func TestTask01Task_MultipleRuns(t *testing.T) {
	task := &Task01Task{}
	ctx := context.Background()

	// Run multiple times to ensure idempotency
	for i := range 3 {
		result, err := task.Run(ctx)
		if err != nil {
			t.Errorf("Run #%d returned error: %v", i+1, err)
		}
		if result != nil {
			t.Errorf("Run #%d returned non-nil result: %v", i+1, result)
		}
	}
}
