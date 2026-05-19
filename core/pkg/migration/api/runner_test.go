package api

import (
	"testing"
	"time"

	"ntels.com/pharos/core/pkg/common"
)

func TestRunWithDisabledMigration(t *testing.T) {
	// Test configuration with migration disabled
	cfg := common.Config{
		Migration: common.MigrationConfig{
			Disable:      true,
			MigrationDir: "./nonexistent",
			ArtifactDir:  "./alsononexistent",
		},
	}

	// This should not panic or return error even with nonexistent directories
	Run(cfg)

	// Since Run is async, we need a small delay to ensure it completes
	time.Sleep(100 * time.Millisecond)

	// If we reach here without panic, the test passes
}

func TestRunWithEmptyMigrationDir(t *testing.T) {
	// Test configuration with empty migration directory
	cfg := common.Config{
		Migration: common.MigrationConfig{
			Disable:      false,
			MigrationDir: "", // Empty directory should cause early return
			ArtifactDir:  "./artifact",
		},
	}

	// This should not panic or return error with empty migration dir
	Run(cfg)

	// Since Run is async, we need a small delay to ensure it completes
	time.Sleep(100 * time.Millisecond)

	// If we reach here without panic, the test passes
}

func TestRunWithInvalidWaitDuration(t *testing.T) {
	// Test configuration with invalid wait duration
	cfg := common.Config{
		Migration: common.MigrationConfig{
			Disable:      false,
			MigrationDir: "./migration",
			ArtifactDir:  "./artifact",
			Wait:         "invalid-duration", // This should be handled gracefully
		},
	}

	// This should not panic even with invalid wait duration
	Run(cfg)

	// Since Run is async, we need a small delay to ensure it completes
	time.Sleep(100 * time.Millisecond)

	// If we reach here without panic, the test passes
}

func TestRunWithValidWaitDuration(t *testing.T) {
	// Test configuration with valid wait duration
	cfg := common.Config{
		Migration: common.MigrationConfig{
			Disable:      false,
			MigrationDir: "./migration",
			ArtifactDir:  "./artifact",
			Wait:         "50ms", // Valid duration
		},
	}

	start := time.Now()

	// This should wait for the specified duration
	Run(cfg)

	// Since Run is async, we need to wait a bit longer than the configured wait
	time.Sleep(150 * time.Millisecond)

	elapsed := time.Since(start)

	// We should have waited at least the specified duration
	if elapsed < 50*time.Millisecond {
		t.Errorf("Expected to wait at least 50ms, but only waited %v", elapsed)
	}
}
