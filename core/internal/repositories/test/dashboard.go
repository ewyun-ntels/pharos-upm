package test

import (
	"context"
	"encoding/json"
	"testing"

	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/shared/types/dashboard"
)

func configEqual(a, b dashboard.DashboardConfig) bool {
	// Convert both configs to JSON and compare
	jsonA, errA := json.Marshal(a)
	jsonB, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return false
	}

	var objA, objB map[string]any
	if err := json.Unmarshal(jsonA, &objA); err != nil {
		return false
	}
	if err := json.Unmarshal(jsonB, &objB); err != nil {
		return false
	}

	jsonStrA, _ := json.Marshal(objA)
	jsonStrB, _ := json.Marshal(objB)
	return string(jsonStrA) == string(jsonStrB)
}

type DashboardTestable interface {
	repositories.DashboardRepository
	SetUpDb() error
	TearDownDb() error
	ClearTable() error
	GetDashboard(id string) (*repositories.DashboardEntity, error)
	GetDashboardCount() (int, error)
}

func DashboardSuite(t *testing.T, testable DashboardTestable) {
	t.Run("Dashboard Repository Suite", func(t *testing.T) {
		t.Run("Create", func(t *testing.T) { DashboardCreate(t, testable) })
		t.Run("GetByID", func(t *testing.T) { DashboardGetByID(t, testable) })
		t.Run("ListAll", func(t *testing.T) { DashboardListAll(t, testable) })
		t.Run("Update", func(t *testing.T) { DashboardUpdate(t, testable) })
		t.Run("DeleteByID", func(t *testing.T) { DashboardDeleteByID(t, testable) })
		t.Run("UpdateAnnotations", func(t *testing.T) { DashboardUpdateAnnotations(t, testable) })
	})
}

func DashboardCreate(t *testing.T, testable DashboardTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})

	var countBefore int
	countBefore, err = testable.GetDashboardCount()
	if err != nil {
		t.Fatalf("GetDashboardCount failed: %v", err)
	}

	id := "test-dashboard"
	config := dashboard.DashboardConfig{
		Title:       "Test Dashboard",
		Description: "A test dashboard",
		Type:        "default",
	}
	dashboardEntity := &repositories.DashboardEntity{ID: id, Config: config}
	err = testable.Create(context.Background(), dashboardEntity)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var countAfter int
	countAfter, err = testable.GetDashboardCount()
	if err != nil {
		t.Fatalf("GetDashboardCount after create failed: %v", err)
	}
	if countAfter != countBefore+1 {
		t.Errorf("Expected dashboard count to increase by 1, but got %d -> %d", countBefore, countAfter)
	}

	d, err := testable.GetDashboard(id)
	if err != nil {
		t.Fatalf("GetDashboard failed: %v", err)
	}
	if d == nil || d.ID != id || !configEqual(d.Config, config) {
		t.Errorf("Created dashboard not found or data mismatch\ngot: %+v\nexpected: %+v", d, dashboardEntity)
	}
}

func DashboardGetByID(t *testing.T, testable DashboardTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})

	id := "test-dashboard"
	config := dashboard.DashboardConfig{
		Title:       "Test Dashboard",
		Description: "A test dashboard",
		Type:        "default",
	}
	dashboardEntity := &repositories.DashboardEntity{ID: id, Config: config}
	err = testable.Create(context.Background(), dashboardEntity)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	d, err := testable.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if d == nil || d.ID != id || !configEqual(d.Config, config) {
		t.Errorf("GetByID returned incorrect data: %+v", d)
	}
}

func DashboardListAll(t *testing.T, testable DashboardTestable) {
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

	ids := []string{"dash1", "dash2", "dash3"}
	for i, id := range ids {
		config := dashboard.DashboardConfig{
			Title: id,
			Type:  "default",
		}
		dashboardEntity := &repositories.DashboardEntity{ID: id, Config: config}
		err = testable.Create(context.Background(), dashboardEntity)
		if err != nil {
			t.Fatalf("Create failed for %s (%d): %v", id, i, err)
		}
	}

	list, err := testable.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(list) != len(ids) {
		t.Errorf("Expected %d dashboards, got %d", len(ids), len(list))
	}

	idSet := make(map[string]struct{})
	for _, d := range list {
		idSet[d.ID] = struct{}{}
	}
	for _, id := range ids {
		if _, exists := idSet[id]; !exists {
			t.Errorf("Dashboard %s not found in ListAll result", id)
		}
	}
}

func DashboardUpdate(t *testing.T, testable DashboardTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})

	id := "test-dashboard"
	config := dashboard.DashboardConfig{
		Title:       "Test Dashboard",
		Description: "Dark theme",
		Type:        "default",
	}
	dashboardEntity := &repositories.DashboardEntity{ID: id, Config: config}
	err = testable.Create(context.Background(), dashboardEntity)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	newConfig := dashboard.DashboardConfig{
		Title:       "Test Dashboard",
		Description: "Light theme",
		Type:        "default",
	}
	dashboardEntity.Config = newConfig
	err = testable.Update(context.Background(), dashboardEntity)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	result, err := testable.GetDashboard(id)
	if err != nil {
		t.Fatalf("GetDashboard after update failed: %v", err)
	}
	if !configEqual(result.Config, newConfig) {
		t.Errorf("Update did not change config: got %+v, want %+v", result.Config, newConfig)
	}
}

func DashboardDeleteByID(t *testing.T, testable DashboardTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})

	id := "test-dashboard"
	config := dashboard.DashboardConfig{
		Title:       "Test Dashboard",
		Description: "To be deleted",
		Type:        "default",
	}
	dashboardEntity := &repositories.DashboardEntity{ID: id, Config: config}
	err = testable.Create(context.Background(), dashboardEntity)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var countBefore int
	countBefore, err = testable.GetDashboardCount()
	if err != nil {
		t.Fatalf("GetDashboardCount failed: %v", err)
	}

	err = testable.DeleteByID(context.Background(), id)
	if err != nil {
		t.Fatalf("DeleteByID failed: %v", err)
	}

	result, err := testable.GetDashboard(id)
	if err == nil && result != nil {
		t.Errorf("Dashboard should be deleted, but found: %+v", result)
	}

	var countAfter int
	countAfter, err = testable.GetDashboardCount()
	if err != nil {
		t.Fatalf("GetDashboardCount after delete failed: %v", err)
	}
	if countAfter != countBefore-1 {
		t.Errorf("Expected dashboard count to decrease by 1, but got %d -> %d", countBefore, countAfter)
	}
}

func DashboardUpdateAnnotations(t *testing.T, testable DashboardTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})

	id := "test-dashboard-annotations"
	config := dashboard.DashboardConfig{
		Title: "Annotation Test Dashboard",
		Type:  "default",
	}
	err = testable.Create(context.Background(), &repositories.DashboardEntity{ID: id, Config: config})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	panelID := "panel-1"
	annotations := []dashboard.Annotation{
		{ID: "ann-1", Time: 1700000000000, Text: "global event", Scope: dashboard.Global},
		{ID: "ann-2", Time: 1700001000000, Text: "panel note", Scope: dashboard.ScopePanel, PanelID: &panelID},
	}

	// 저장
	err = testable.UpdateAnnotations(context.Background(), id, annotations)
	if err != nil {
		t.Fatalf("UpdateAnnotations failed: %v", err)
	}

	// 조회해서 검증
	d, err := testable.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID after UpdateAnnotations failed: %v", err)
	}
	if len(d.Annotations) != 2 {
		t.Fatalf("Expected 2 annotations, got %d", len(d.Annotations))
	}
	if d.Annotations[0].ID != "ann-1" || d.Annotations[0].Text != "global event" {
		t.Errorf("Annotation[0] mismatch: %+v", d.Annotations[0])
	}
	if d.Annotations[1].ID != "ann-2" || d.Annotations[1].PanelID == nil || *d.Annotations[1].PanelID != panelID {
		t.Errorf("Annotation[1] mismatch: %+v", d.Annotations[1])
	}

	// 빈 리스트로 덮어쓰기
	err = testable.UpdateAnnotations(context.Background(), id, []dashboard.Annotation{})
	if err != nil {
		t.Fatalf("UpdateAnnotations (clear) failed: %v", err)
	}

	d, err = testable.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID after clear failed: %v", err)
	}
	if len(d.Annotations) != 0 {
		t.Errorf("Expected 0 annotations after clear, got %d", len(d.Annotations))
	}

	// Update는 annotations에 영향 없음 검증
	annotations2 := []dashboard.Annotation{
		{ID: "ann-3", Time: 1700002000000, Text: "preserved", Scope: dashboard.Global},
	}
	err = testable.UpdateAnnotations(context.Background(), id, annotations2)
	if err != nil {
		t.Fatalf("UpdateAnnotations (set before update) failed: %v", err)
	}
	err = testable.Update(context.Background(), &repositories.DashboardEntity{ID: id, Config: dashboard.DashboardConfig{Title: "Updated", Type: "default"}})
	if err != nil {
		t.Fatalf("Update (config) failed: %v", err)
	}
	d, err = testable.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID after config update failed: %v", err)
	}
	if len(d.Annotations) != 1 || d.Annotations[0].ID != "ann-3" {
		t.Errorf("Annotations should be preserved after config update, got: %+v", d.Annotations)
	}
}
