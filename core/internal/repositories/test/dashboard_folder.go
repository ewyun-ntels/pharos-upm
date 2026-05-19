package test

import (
	"context"
	"testing"
	"time"

	"ntels.com/pharos/core/internal/repositories"
)

type DashboardFolderTestable interface {
	repositories.DashboardFolderRepository
	SetUpDb() error
	TearDownDb() error
	ClearTable() error
	GetFolder(id string) (*repositories.DashboardFolderEntity, error)
	GetFolderCount() (int, error)
	CreateDashboard(id string) error
	GetDashboardFolderID(dashboardID string) (*string, error)
}

func DashboardFolderSuite(t *testing.T, testable DashboardFolderTestable) {
	t.Run("DashboardFolder Repository Suite", func(t *testing.T) {
		t.Run("Create", func(t *testing.T) { DashboardFolderCreate(t, testable) })
		t.Run("GetByID", func(t *testing.T) { DashboardFolderGetByID(t, testable) })
		t.Run("ListAll", func(t *testing.T) { DashboardFolderListAll(t, testable) })
		t.Run("Update", func(t *testing.T) { DashboardFolderUpdate(t, testable) })
		t.Run("DeleteByID", func(t *testing.T) { DashboardFolderDeleteByID(t, testable) })
		t.Run("MoveDashboard", func(t *testing.T) { DashboardFolderMoveDashboard(t, testable) })
	})
}

func DashboardFolderCreate(t *testing.T, testable DashboardFolderTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})

	countBefore, err := testable.GetFolderCount()
	if err != nil {
		t.Fatalf("GetFolderCount failed: %v", err)
	}

	id := "test-folder"
	folder := &repositories.DashboardFolderEntity{
		ID:        id,
		Name:      "Test Folder",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = testable.Create(context.Background(), folder)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	countAfter, err := testable.GetFolderCount()
	if err != nil {
		t.Fatalf("GetFolderCount after create failed: %v", err)
	}
	if countAfter != countBefore+1 {
		t.Errorf("Expected folder count to increase by 1, but got %d -> %d", countBefore, countAfter)
	}

	f, err := testable.GetFolder(id)
	if err != nil {
		t.Fatalf("GetFolder failed: %v", err)
	}
	if f == nil || f.ID != id || f.Name != folder.Name {
		t.Errorf("Created folder not found or data mismatch\ngot: %+v\nexpected: %+v", f, folder)
	}
}

func DashboardFolderGetByID(t *testing.T, testable DashboardFolderTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})

	id := "test-folder"
	folder := &repositories.DashboardFolderEntity{
		ID:        id,
		Name:      "Test Folder",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = testable.Create(context.Background(), folder)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	f, err := testable.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if f == nil || f.ID != id || f.Name != folder.Name {
		t.Errorf("GetByID returned incorrect data: %+v", f)
	}
}

func DashboardFolderListAll(t *testing.T, testable DashboardFolderTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})

	err = testable.ClearTable()
	if err != nil {
		t.Fatalf("ClearTable failed: %v", err)
	}

	ids := []string{"folder1", "folder2", "folder3"}
	for _, id := range ids {
		f := &repositories.DashboardFolderEntity{
			ID:        id,
			Name:      id,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = testable.Create(context.Background(), f)
		if err != nil {
			t.Fatalf("Create failed for %s: %v", id, err)
		}
	}

	list, err := testable.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(list) != len(ids) {
		t.Errorf("Expected %d folders, got %d", len(ids), len(list))
	}

	idSet := make(map[string]struct{})
	for _, f := range list {
		idSet[f.ID] = struct{}{}
	}
	for _, id := range ids {
		if _, exists := idSet[id]; !exists {
			t.Errorf("Folder %s not found in ListAll result", id)
		}
	}
}

func DashboardFolderUpdate(t *testing.T, testable DashboardFolderTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})

	id := "test-folder"
	folder := &repositories.DashboardFolderEntity{
		ID:        id,
		Name:      "Original Name",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = testable.Create(context.Background(), folder)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	folder.Name = "Updated Name"
	folder.UpdatedAt = time.Now()
	err = testable.Update(context.Background(), folder)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	result, err := testable.GetFolder(id)
	if err != nil {
		t.Fatalf("GetFolder after update failed: %v", err)
	}
	if result.Name != "Updated Name" {
		t.Errorf("Update did not change name: got %q, want %q", result.Name, "Updated Name")
	}
}

func DashboardFolderDeleteByID(t *testing.T, testable DashboardFolderTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})

	id := "test-folder"
	folder := &repositories.DashboardFolderEntity{
		ID:        id,
		Name:      "To Be Deleted",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = testable.Create(context.Background(), folder)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	countBefore, err := testable.GetFolderCount()
	if err != nil {
		t.Fatalf("GetFolderCount failed: %v", err)
	}

	err = testable.DeleteByID(context.Background(), id)
	if err != nil {
		t.Fatalf("DeleteByID failed: %v", err)
	}

	result, err := testable.GetFolder(id)
	if err == nil && result != nil {
		t.Errorf("Folder should be deleted, but found: %+v", result)
	}

	countAfter, err := testable.GetFolderCount()
	if err != nil {
		t.Fatalf("GetFolderCount after delete failed: %v", err)
	}
	if countAfter != countBefore-1 {
		t.Errorf("Expected folder count to decrease by 1, but got %d -> %d", countBefore, countAfter)
	}
}

func DashboardFolderMoveDashboard(t *testing.T, testable DashboardFolderTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})

	folderID := "test-folder"
	folder := &repositories.DashboardFolderEntity{
		ID:        folderID,
		Name:      "Test Folder",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = testable.Create(context.Background(), folder)
	if err != nil {
		t.Fatalf("Create folder failed: %v", err)
	}

	dashboardID := "test-dashboard"
	err = testable.CreateDashboard(dashboardID)
	if err != nil {
		t.Fatalf("CreateDashboard failed: %v", err)
	}

	// Move dashboard into folder
	err = testable.MoveDashboard(context.Background(), dashboardID, &folderID)
	if err != nil {
		t.Fatalf("MoveDashboard (into folder) failed: %v", err)
	}

	gotFolderID, err := testable.GetDashboardFolderID(dashboardID)
	if err != nil {
		t.Fatalf("GetDashboardFolderID failed: %v", err)
	}
	if gotFolderID == nil || *gotFolderID != folderID {
		t.Errorf("Expected folder_id %q, got %v", folderID, gotFolderID)
	}

	// Remove dashboard from folder
	err = testable.MoveDashboard(context.Background(), dashboardID, nil)
	if err != nil {
		t.Fatalf("MoveDashboard (remove from folder) failed: %v", err)
	}

	gotFolderID, err = testable.GetDashboardFolderID(dashboardID)
	if err != nil {
		t.Fatalf("GetDashboardFolderID after removal failed: %v", err)
	}
	if gotFolderID != nil {
		t.Errorf("Expected folder_id to be nil after removal, got %q", *gotFolderID)
	}
}
