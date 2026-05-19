package sample

import (
	"context"
	"testing"

	"ntels.com/pharos/core/pkg/common"
)

func TestTask02Task_GetName(t *testing.T) {
	task := &Task02Task{}
	expected := "sample-task-02"

	got := task.GetName()
	if got != expected {
		t.Errorf("GetName() = %v, want %v", got, expected)
	}
}

func TestTask02Task_SetConfig(t *testing.T) {
	task := &Task02Task{}
	configPath := "/test/config/path"
	config := common.Config{}

	task.SetConfig(configPath, config)

	if task.configPath != configPath {
		t.Errorf("Expected configPath '%s', got '%s'", configPath, task.configPath)
	}
}

func TestTask02Task_SetData(t *testing.T) {
	task := &Task02Task{}
	testData := "test-string-data"

	task.SetData(testData)

	if task.data == nil {
		t.Error("Expected data to be set, got nil")
	}

	if task.data != testData {
		t.Errorf("Expected data '%v', got '%v'", testData, task.data)
	}
}

func TestTask02Task_GetDataFormat(t *testing.T) {
	task := &Task02Task{}

	format := task.GetDataFormat()

	// Verify the format is a string with expected value
	formatStr, ok := format.(string)
	if !ok {
		t.Fatal("Expected GetDataFormat() to return string")
	}

	if formatStr != "value" {
		t.Errorf("Expected 'value', got '%s'", formatStr)
	}
}

func TestTask02Task_Run(t *testing.T) {
	task := &Task02Task{}
	ctx := context.Background()

	result, err := task.Run(ctx)

	if err != nil {
		t.Errorf("Run() returned unexpected error: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
}

func TestTask02Task_Interface(t *testing.T) {
	// Verify that Task02Task implements all required methods
	task := &Task02Task{}

	// Test all interface methods exist
	_ = task.GetName()
	task.SetConfig("", common.Config{})
	task.SetData(nil)
	_ = task.GetDataFormat()
	_, _ = task.Run(context.Background())
}

func TestTask02Task_ConfigPersistence(t *testing.T) {
	task := &Task02Task{}
	configPath := "/test/path/config.yaml"
	config := common.Config{}

	task.SetConfig(configPath, config)

	if task.configPath != configPath {
		t.Errorf("configPath not persisted correctly, got %s", task.configPath)
	}
}

func TestTask02Task_DataPersistence(t *testing.T) {
	task := &Task02Task{}
	testData := "persistent-data"

	task.SetData(testData)

	if task.data != testData {
		t.Errorf("data not persisted correctly, got %v", task.data)
	}
}

func TestTask02Task_MultipleRuns(t *testing.T) {
	task := &Task02Task{}
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

func TestTask02Task_WithContext(t *testing.T) {
	task := &Task02Task{}

	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "with background context",
			ctx:  context.Background(),
		},
		{
			name: "with TODO context",
			ctx:  context.TODO(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := task.Run(tt.ctx)
			if err != nil {
				t.Errorf("Run() error = %v", err)
			}
			if result != nil {
				t.Errorf("Run() result = %v, want nil", result)
			}
		})
	}
}
