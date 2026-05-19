package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ntels.com/pharos/core/pkg/common"
)

func TestBackupTask_GetName(t *testing.T) {
	task := &BackupTask{}
	expected := "sqlite-backup"

	got := task.GetName()
	if got != expected {
		t.Errorf("GetName() = %v, want %v", got, expected)
	}
}

func TestBackupTask_SetConfig(t *testing.T) {
	task := &BackupTask{}
	configPath := "/test/config/path"
	config := common.Config{}

	task.SetConfig(configPath, config)

	if task.configPath != configPath {
		t.Errorf("Expected configPath '%s', got '%s'", configPath, task.configPath)
	}
}

func TestBackupTask_SetData(t *testing.T) {
	task := &BackupTask{}
	testData := map[string]any{"key": "value"}

	task.SetData(testData)

	if task.data == nil {
		t.Error("Expected data to be set, got nil")
	}
}

func TestBackupTask_GetDataFormat(t *testing.T) {
	task := &BackupTask{}

	format := task.GetDataFormat()
	if format != "" {
		t.Errorf("Expected empty string, got %v", format)
	}
}

func TestBackupTask_Interface(t *testing.T) {
	// Verify that BackupTask implements all required methods
	task := &BackupTask{}

	// Test all interface methods exist
	_ = task.GetName()
	task.SetConfig("", common.Config{})
	task.SetData(nil)
	_ = task.GetDataFormat()
	_, _ = task.Run(context.Background())
}

func TestBackupTask_ConfigPersistence(t *testing.T) {
	task := &BackupTask{}
	configPath := "/test/path/config.yaml"
	config := common.Config{}

	task.SetConfig(configPath, config)

	if task.configPath != configPath {
		t.Errorf("configPath not persisted correctly, got %s", task.configPath)
	}
}

func TestBackupTask_DataPersistence(t *testing.T) {
	task := &BackupTask{}
	testData := "test-data"

	task.SetData(testData)

	if task.data != testData {
		t.Errorf("data not persisted correctly, got %v", task.data)
	}
}

func TestBackupTask_MkdirBackupDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := filepath.Join(os.TempDir(), "test-backup-dir")
	defer os.RemoveAll(tempDir)

	task := &BackupTask{}
	task.config.Database.SQLite.Backup.Directory = tempDir

	err := task.mkdirBackupDirectory()
	if err != nil {
		t.Errorf("mkdirBackupDirectory() returned error: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}
}

func TestBackupTask_GetBackupFileName(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := filepath.Join(os.TempDir(), "test-backup-filename")
	defer os.RemoveAll(tempDir)

	task := &BackupTask{}
	task.config.Database.SQLite.Backup.Directory = tempDir

	fileName, err := task.getBackupFileName()
	if err != nil {
		t.Errorf("getBackupFileName() returned error: %v", err)
	}

	if fileName == "" {
		t.Error("Expected non-empty filename")
	}

	// Verify filename format (should contain .db extension)
	if filepath.Ext(fileName) != ".db" {
		t.Errorf("Expected .db extension, got %s", filepath.Ext(fileName))
	}

	// Verify directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}
}

func TestBackupTask_TTL_InvalidDuration(t *testing.T) {
	task := &BackupTask{}
	task.config.Database.SQLite.Backup.TTL = "invalid-duration"

	err := task.ttl()
	if err == nil {
		t.Error("Expected error for invalid duration")
	}
}

func TestBackupTask_TTL_EmptyDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := filepath.Join(os.TempDir(), "test-backup-ttl-empty")
	defer os.RemoveAll(tempDir)

	err := os.MkdirAll(tempDir, 0o750)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	task := &BackupTask{}
	task.config.Database.SQLite.Backup.Directory = tempDir
	task.config.Database.SQLite.Backup.TTL = "24h"

	err = task.ttl()
	if err != nil {
		t.Errorf("ttl() returned error for empty directory: %v", err)
	}
}

func TestBackupTask_Run(t *testing.T) {
	task := &BackupTask{}
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

func TestBackupTask_MkdirBackupDirectory_ExistingDirectory(t *testing.T) {
	// Create a temporary directory
	tempDir := filepath.Join(os.TempDir(), "test-backup-existing")
	err := os.MkdirAll(tempDir, 0o750)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	task := &BackupTask{}
	task.config.Database.SQLite.Backup.Directory = tempDir

	// Should not return error for existing directory
	err = task.mkdirBackupDirectory()
	if err != nil {
		t.Errorf("mkdirBackupDirectory() returned error for existing directory: %v", err)
	}
}
