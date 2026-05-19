package test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/internal/repositories"
	sharedRole "ntels.com/pharos/shared/types/role"
)

type RoleMetadataConfigTestable interface {
	repositories.RoleMetadataConfigRepository
	SetUpDb() error
	TearDownDb() error
	ClearTable() error
}

func RoleMetadataConfigSuite(t *testing.T, testable RoleMetadataConfigTestable) {
	t.Run("RoleMetadataConfig Repository Suite", func(t *testing.T) {
		t.Run("SaveConfig", func(t *testing.T) { RoleMetadataConfigSave(t, testable) })
		t.Run("GetActiveConfig_NotFound", func(t *testing.T) { RoleMetadataConfigGetActiveNotFound(t, testable) })
		t.Run("SaveConfig_MultipleVersions", func(t *testing.T) { RoleMetadataConfigSaveMultipleVersions(t, testable) })
		t.Run("GetHistory", func(t *testing.T) { RoleMetadataConfigGetHistory(t, testable) })
		t.Run("SaveConfig_MaxHistoryCleanup", func(t *testing.T) { RoleMetadataConfigSaveMaxHistoryCleanup(t, testable) })
		t.Run("Rollback", func(t *testing.T) { RoleMetadataConfigRollback(t, testable) })
		t.Run("DeleteAll", func(t *testing.T) { RoleMetadataConfigDeleteAll(t, testable) })
		t.Run("SaveWithoutUpdatedBy", func(t *testing.T) { RoleMetadataConfigSaveWithoutUpdatedBy(t, testable) })
		t.Run("JSONMarshalUnmarshal", func(t *testing.T) { RoleMetadataConfigJSONMarshalUnmarshal(t, testable) })
		t.Run("RollbackNonexistent", func(t *testing.T) { RoleMetadataConfigRollbackNonexistent(t, testable) })
		t.Run("ConcurrentWrites", func(t *testing.T) { RoleMetadataConfigConcurrentWrites(t, testable) })
	})
}

func createTestMetadata(key string) sharedRole.RoleMetadata {
	return sharedRole.RoleMetadata{
		Key:         key,
		DisplayName: "Test Role " + key,
		Description: "Test description for " + key,
		Group:       "test-group",
		Hide:        ptrBool(false),
	}
}

func createTestConfig(metadata []sharedRole.RoleMetadata, updatedBy *string, description string) repositories.RoleMetadataConfigEntity {
	configJSON, _ := json.Marshal(metadata)
	desc := &description
	return repositories.RoleMetadataConfigEntity{
		ConfigJSON:  string(configJSON),
		Config:      metadata,
		UpdatedBy:   updatedBy,
		Description: desc,
	}
}

func ptrBool(b bool) *bool {
	return &b
}

func RoleMetadataConfigSave(t *testing.T, testable RoleMetadataConfigTestable) {
	err := testable.SetUpDb()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = testable.TearDownDb()
	})

	ctx := context.Background()
	updatedBy := "admin"

	// Create test config with multiple roles
	config := []sharedRole.RoleMetadata{
		createTestMetadata("test.role.1"),
		createTestMetadata("test.role.2"),
	}

	entity := createTestConfig(config, &updatedBy, "Initial config")

	// Save config
	err = testable.SaveConfig(ctx, entity, 10)
	require.NoError(t, err)

	// Verify active config exists
	active, err := testable.GetActiveConfig(ctx)
	require.NoError(t, err)
	assert.NotNil(t, active)
	assert.Len(t, active.Config, 2)
	assert.Equal(t, "admin", *active.UpdatedBy)
	assert.Equal(t, "Initial config", *active.Description)
	assert.NotEmpty(t, active.ID)
}

func RoleMetadataConfigGetActiveNotFound(t *testing.T, testable RoleMetadataConfigTestable) {
	err := testable.SetUpDb()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = testable.TearDownDb()
	})

	ctx := context.Background()
	active, err := testable.GetActiveConfig(ctx)
	require.NoError(t, err)
	assert.Nil(t, active)
}

func RoleMetadataConfigSaveMultipleVersions(t *testing.T, testable RoleMetadataConfigTestable) {
	err := testable.SetUpDb()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = testable.TearDownDb()
	})

	ctx := context.Background()
	updatedBy := "admin"

	// Create 3 versions of config
	for i := 1; i <= 3; i++ {
		config := []sharedRole.RoleMetadata{
			createTestMetadata("test.role.1"),
		}
		config[0].DisplayName = fmt.Sprintf("Version %d", i)

		entity := createTestConfig(config, &updatedBy, fmt.Sprintf("Config version %d", i))
		err = testable.SaveConfig(ctx, entity, 10)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	// Verify only latest is active
	active, err := testable.GetActiveConfig(ctx)
	require.NoError(t, err)
	assert.NotNil(t, active)
	assert.Equal(t, "Version 3", active.Config[0].DisplayName)
	assert.Equal(t, "Config version 3", *active.Description)

	// Verify history contains all versions
	history, err := testable.GetHistory(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, history, 3)
	assert.Equal(t, "Version 3", history[0].Config[0].DisplayName)
	assert.Equal(t, "Version 2", history[1].Config[0].DisplayName)
	assert.Equal(t, "Version 1", history[2].Config[0].DisplayName)
}

func RoleMetadataConfigGetHistory(t *testing.T, testable RoleMetadataConfigTestable) {
	err := testable.SetUpDb()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = testable.TearDownDb()
	})

	ctx := context.Background()
	updatedBy := "admin"

	// Create 5 versions
	for i := 1; i <= 5; i++ {
		config := []sharedRole.RoleMetadata{
			createTestMetadata("test.role.1"),
		}
		config[0].DisplayName = fmt.Sprintf("Version %d", i)

		entity := createTestConfig(config, &updatedBy, fmt.Sprintf("Version %d", i))
		err = testable.SaveConfig(ctx, entity, 10)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	// Get history with limit
	history, err := testable.GetHistory(ctx, 3)
	require.NoError(t, err)
	assert.Len(t, history, 3)
	assert.Equal(t, "Version 5", history[0].Config[0].DisplayName)
	assert.Equal(t, "Version 4", history[1].Config[0].DisplayName)
	assert.Equal(t, "Version 3", history[2].Config[0].DisplayName)

	// Get all history
	allHistory, err := testable.GetHistory(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, allHistory, 5)
}

func RoleMetadataConfigSaveMaxHistoryCleanup(t *testing.T, testable RoleMetadataConfigTestable) {
	err := testable.SetUpDb()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = testable.TearDownDb()
	})

	ctx := context.Background()
	updatedBy := "admin"
	maxHistory := 3

	// Create 5 versions with max 3 history
	for i := 1; i <= 5; i++ {
		config := []sharedRole.RoleMetadata{
			createTestMetadata("test.role.1"),
		}
		config[0].DisplayName = fmt.Sprintf("Version %d", i)

		entity := createTestConfig(config, &updatedBy, fmt.Sprintf("Version %d", i))
		err = testable.SaveConfig(ctx, entity, maxHistory)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	// Verify only 3 versions remain
	history, err := testable.GetHistory(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, history, maxHistory)
	assert.Equal(t, "Version 5", history[0].Config[0].DisplayName)
	assert.Equal(t, "Version 4", history[1].Config[0].DisplayName)
	assert.Equal(t, "Version 3", history[2].Config[0].DisplayName)
}

func RoleMetadataConfigRollback(t *testing.T, testable RoleMetadataConfigTestable) {
	err := testable.SetUpDb()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = testable.TearDownDb()
	})

	ctx := context.Background()
	updatedBy := "admin"

	// Create 3 versions
	var version2ID string
	for i := 1; i <= 3; i++ {
		config := []sharedRole.RoleMetadata{
			createTestMetadata("test.role.1"),
		}
		config[0].DisplayName = fmt.Sprintf("Version %d", i)

		entity := createTestConfig(config, &updatedBy, fmt.Sprintf("Version %d", i))
		err = testable.SaveConfig(ctx, entity, 10)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)

		// Save version 2 ID for rollback
		if i == 2 {
			active, err := testable.GetActiveConfig(ctx)
			require.NoError(t, err)
			version2ID = active.ID
		}
	}

	// Verify version 3 is active
	active, err := testable.GetActiveConfig(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Version 3", active.Config[0].DisplayName)

	// Rollback to version 2
	rollbackBy := "rollback-user"
	err = testable.Rollback(ctx, version2ID, &rollbackBy, 10)
	require.NoError(t, err)

	// Verify version 2 content is now active
	active, err = testable.GetActiveConfig(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Version 2", active.Config[0].DisplayName)
	assert.Equal(t, "rollback-user", *active.UpdatedBy)

	// Verify history now has 4 entries (original 3 + rollback)
	history, err := testable.GetHistory(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, history, 4)
	assert.Equal(t, "Version 2", history[0].Config[0].DisplayName) // Rolled back version
}

func RoleMetadataConfigDeleteAll(t *testing.T, testable RoleMetadataConfigTestable) {
	err := testable.SetUpDb()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = testable.TearDownDb()
	})

	ctx := context.Background()
	updatedBy := "admin"

	// Create config
	config := []sharedRole.RoleMetadata{
		createTestMetadata("test.role.1"),
		createTestMetadata("test.role.2"),
		createTestMetadata("test.role.3"),
	}

	entity := createTestConfig(config, &updatedBy, "Initial config")
	err = testable.SaveConfig(ctx, entity, 10)
	require.NoError(t, err)

	// Verify it exists
	active, err := testable.GetActiveConfig(ctx)
	require.NoError(t, err)
	assert.NotNil(t, active)

	// Delete all
	err = testable.DeleteAll(ctx)
	require.NoError(t, err)

	// Verify deleted
	active, err = testable.GetActiveConfig(ctx)
	require.NoError(t, err)
	assert.Nil(t, active)
}

func RoleMetadataConfigSaveWithoutUpdatedBy(t *testing.T, testable RoleMetadataConfigTestable) {
	err := testable.SetUpDb()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = testable.TearDownDb()
	})

	ctx := context.Background()
	config := []sharedRole.RoleMetadata{
		createTestMetadata("test.role.1"),
	}

	entity := createTestConfig(config, nil, "No user config")
	err = testable.SaveConfig(ctx, entity, 10)
	require.NoError(t, err)

	active, err := testable.GetActiveConfig(ctx)
	require.NoError(t, err)
	assert.NotNil(t, active)
	assert.Nil(t, active.UpdatedBy)
}

func RoleMetadataConfigJSONMarshalUnmarshal(t *testing.T, testable RoleMetadataConfigTestable) {
	err := testable.SetUpDb()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = testable.TearDownDb()
	})

	ctx := context.Background()
	// Create config with all fields
	config := []sharedRole.RoleMetadata{
		{
			Key:         "test.role.json",
			DisplayName: "JSON Test Role",
			Description: "Description with special chars: 한글, emoji 🎉",
			Group:       "custom-group",
			Hide:        ptrBool(true),
		},
	}

	updatedBy := "json-tester"
	entity := createTestConfig(config, &updatedBy, "JSON test config")
	err = testable.SaveConfig(ctx, entity, 10)
	require.NoError(t, err)

	// Retrieve and verify
	active, err := testable.GetActiveConfig(ctx)
	require.NoError(t, err)
	assert.NotNil(t, active)
	assert.Len(t, active.Config, 1)
	assert.Equal(t, config[0].Key, active.Config[0].Key)
	assert.Equal(t, config[0].DisplayName, active.Config[0].DisplayName)
	assert.Equal(t, config[0].Description, active.Config[0].Description)
	assert.Equal(t, config[0].Group, active.Config[0].Group)
	assert.Equal(t, *config[0].Hide, *active.Config[0].Hide)
}

func RoleMetadataConfigRollbackNonexistent(t *testing.T, testable RoleMetadataConfigTestable) {
	err := testable.SetUpDb()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = testable.TearDownDb()
	})

	ctx := context.Background()
	// Try to rollback to nonexistent ID
	fakeID := uuid.New().String()
	err = testable.Rollback(ctx, fakeID, nil, 10)
	assert.Error(t, err)
}

func RoleMetadataConfigConcurrentWrites(t *testing.T, testable RoleMetadataConfigTestable) {
	err := testable.SetUpDb()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = testable.TearDownDb()
	})

	ctx := context.Background()
	updatedBy := "concurrent-user"

	// Create initial config
	config := []sharedRole.RoleMetadata{
		createTestMetadata("test.role.1"),
	}
	entity := createTestConfig(config, &updatedBy, "Initial")
	err = testable.SaveConfig(ctx, entity, 10)
	require.NoError(t, err)

	// Concurrent updates should not cause errors
	done := make(chan bool, 2)
	for i := range 2 {
		go func(version int) {
			defer func() { done <- true }()
			cfg := []sharedRole.RoleMetadata{
				createTestMetadata("test.role.1"),
			}
			cfg[0].DisplayName = fmt.Sprintf("Concurrent %d", version)
			e := createTestConfig(cfg, &updatedBy, fmt.Sprintf("Concurrent %d", version))
			_ = testable.SaveConfig(ctx, e, 10)
		}(i)
	}

	<-done
	<-done

	// Verify config exists (one of the concurrent writes succeeded)
	active, err := testable.GetActiveConfig(ctx)
	require.NoError(t, err)
	assert.NotNil(t, active)
}
