package common

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	alertresources "ntels.com/pharos/core/pkg/alert/resources"
	"ntels.com/pharos/core/pkg/common"
)

type RuleShim struct {
	id       string
	name     string
	execHits int32
	runHits  int32
}

func (d *RuleShim) Load(_ common.Config, _ map[string]any) error { return nil }
func (d *RuleShim) Normalize() error                             { return nil }
func (d *RuleShim) GetId() string                                { return d.id }
func (d *RuleShim) SetId(id string)                              { d.id = id }
func (d *RuleShim) GetName() string                              { return d.name }
func (d *RuleShim) Validate() error                              { return nil }
func (d *RuleShim) Run(_ context.Context) error {
	atomic.AddInt32(&d.runHits, 1)
	return nil
}
func (d *RuleShim) Destroy() error { return nil }
func (d *RuleShim) Execute(_ context.Context) error {
	atomic.AddInt32(&d.execHits, 1)
	return nil
}
func (d *RuleShim) EventHandler(_ *alertresources.Event) error { return nil }

func TestCron_AddRemoveJob_And_Run(t *testing.T) {
	// Ensure controller is started
	if err := CronRun(); err != nil {
		t.Fatalf("CronRun error: %v", err)
	}

	// 1) AddJob should fail if ID empty
	rEmpty := &RuleShim{id: "", name: "empty"}
	if err := CronInstance.AddJob("@every 1s", rEmpty); err == nil {
		t.Fatalf("expected error for empty job id")
	}

	// 2) Add a job with seconds spec; Execute should be called
	r := &RuleShim{id: "job1", name: "job1"}
	if err := CronInstance.AddJob("@every 1s", r); err != nil {
		t.Fatalf("AddJob error: %v", err)
	}

	// Wait up to 2 seconds for at least one Execute
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&r.execHits) > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if atomic.LoadInt32(&r.execHits) == 0 {
		t.Fatalf("expected Execute to be called at least once")
	}

	// 3) Reschedule should switch to using Run()
	if err := CronInstance.RescheduleJob("job1", "@every 1s"); err != nil {
		t.Fatalf("RescheduleJob error: %v", err)
	}

	// Wait up to 2 seconds for Run to be called
	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&r.runHits) > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if atomic.LoadInt32(&r.runHits) == 0 {
		t.Fatalf("expected Run to be called after reschedule")
	}

	// 4) RemoveJob should succeed, and second removal should fail
	if err := CronInstance.RemoveJob("job1"); err != nil {
		t.Fatalf("RemoveJob error: %v", err)
	}
	if err := CronInstance.RemoveJob("job1"); err == nil {
		t.Fatalf("expected error on removing non-existent job")
	}
}
