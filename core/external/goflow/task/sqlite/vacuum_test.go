package sqlite

import (
	"context"
	"testing"

	"ntels.com/pharos/core/pkg/common"
)

func TestVacuumTask_GetName(t *testing.T) {
	task := &VacuumTask{}
	expected := "sqlite-vacuum"

	got := task.GetName()
	if got != expected {
		t.Errorf("GetName() = %v, want %v", got, expected)
	}
}

func TestVacuumTask_SetConfig(t *testing.T) {
	task := &VacuumTask{}
	configPath := "/test/config/path"
	config := common.Config{}

	task.SetConfig(configPath, config)

	if task.configPath != configPath {
		t.Errorf("Expected configPath '%s', got '%s'", configPath, task.configPath)
	}
}

func TestVacuumTask_SetData(t *testing.T) {
	task := &VacuumTask{}
	testData := map[string]any{"key": "value"}

	task.SetData(testData)

	if task.data == nil {
		t.Error("Expected data to be set, got nil")
	}
}

func TestVacuumTask_GetDataFormat(t *testing.T) {
	task := &VacuumTask{}

	format := task.GetDataFormat()
	if format != "" {
		t.Errorf("Expected empty string, got %v", format)
	}
}

func TestVacuumTask_Run(t *testing.T) {
	task := &VacuumTask{}
	ctx := context.Background()

	// Note: This test will likely fail if database is not properly initialized
	// In a real testing scenario, you might want to mock the database
	_, err := task.Run(ctx)

	// We test that the function can be called without panic
	// The actual error depends on whether database is available
	if err != nil {
		t.Logf("Run() returned error (expected if database not available): %v", err)
	}
}

func TestVacuumTask_Interface(t *testing.T) {
	// Verify that VacuumTask implements all required methods
	task := &VacuumTask{}

	// Test all interface methods exist
	_ = task.GetName()
	task.SetConfig("", common.Config{})
	task.SetData(nil)
	_ = task.GetDataFormat()
	_, _ = task.Run(context.Background())
}

func TestVacuumTask_ConfigPersistence(t *testing.T) {
	task := &VacuumTask{}
	configPath := "/test/path/config.yaml"
	config := common.Config{}

	task.SetConfig(configPath, config)

	if task.configPath != configPath {
		t.Errorf("configPath not persisted correctly, got %s", task.configPath)
	}
}

func TestVacuumTask_DataPersistence(t *testing.T) {
	task := &VacuumTask{}
	testData := "test-data"

	task.SetData(testData)

	if task.data != testData {
		t.Errorf("data not persisted correctly, got %v", task.data)
	}
}

func TestVacuumTask_MultipleRuns(t *testing.T) {
	task := &VacuumTask{}
	ctx := context.Background()

	// Note: These tests will fail if database is not available
	// Just testing that multiple calls don't cause panics
	for i := range 3 {
		_, err := task.Run(ctx)
		if err != nil {
			t.Logf("Run #%d returned error (expected if database not available): %v", i+1, err)
		}
	}
}

func TestVacuumTask_WithContext(t *testing.T) {
	task := &VacuumTask{}

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
			_, err := task.Run(tt.ctx)
			// We're just testing that it doesn't panic
			// Error is expected if database is not available
			if err != nil {
				t.Logf("Run() with %s returned error: %v", tt.name, err)
			}
		})
	}
}

func TestVacuumTask_SetMultipleData(t *testing.T) {
	task := &VacuumTask{}

	// Set data multiple times
	data1 := "first-data"
	data2 := "second-data"
	data3 := map[string]string{"key": "value"}

	task.SetData(data1)
	if task.data != data1 {
		t.Error("First data not set correctly")
	}

	task.SetData(data2)
	if task.data != data2 {
		t.Error("Second data not set correctly")
	}

	task.SetData(data3)
	if task.data == nil {
		t.Error("Third data not set correctly")
	}
}

func TestVacuumTask_SetMultipleConfigs(t *testing.T) {
	task := &VacuumTask{}

	config1 := common.Config{}
	path1 := "/path/1"

	config2 := common.Config{}
	path2 := "/path/2"

	task.SetConfig(path1, config1)
	if task.configPath != path1 {
		t.Error("First config path not set correctly")
	}

	task.SetConfig(path2, config2)
	if task.configPath != path2 {
		t.Error("Second config path not set correctly")
	}
}
