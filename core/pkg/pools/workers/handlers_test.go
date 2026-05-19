package workers_test

import (
	"context"
	"testing"

	"ntels.com/pharos/core/pkg/pools/workers"
)

func TestSampleHandler1_CanHandle(t *testing.T) {
	handler := workers.NewSampleHandler1()

	if !handler.CanHandle(workers.JobTypeSampleTask1) {
		t.Error("SampleHandler1 should handle JobTypeSampleTask1")
	}

	if handler.CanHandle(workers.JobTypeSampleTask2) {
		t.Error("SampleHandler1 should not handle JobTypeSampleTask2")
	}
}

func TestSampleHandler2_CanHandle(t *testing.T) {
	handler := workers.NewSampleHandler2()

	if !handler.CanHandle(workers.JobTypeSampleTask2) {
		t.Error("SampleHandler2 should handle JobTypeSampleTask2")
	}

	if handler.CanHandle(workers.JobTypeSampleTask1) {
		t.Error("SampleHandler2 should not handle JobTypeSampleTask1")
	}
}

func TestSampleHandler1_Handle(t *testing.T) {
	handler := workers.NewSampleHandler1()

	job := workers.Job{
		ID:      "test-1",
		Type:    workers.JobTypeSampleTask1,
		Payload: map[string]any{"data": "test"},
	}

	err := handler.Handle(context.Background(), job)
	if err != nil {
		t.Errorf("SampleHandler1 should not return error: %v", err)
	}
}

func TestSampleHandler2_Handle(t *testing.T) {
	handler := workers.NewSampleHandler2()

	job := workers.Job{
		ID:      "test-2",
		Type:    workers.JobTypeSampleTask2,
		Payload: map[string]any{"data": "test"},
	}

	err := handler.Handle(context.Background(), job)
	if err != nil {
		t.Errorf("SampleHandler2 should not return error: %v", err)
	}
}

func TestBaseHandler(t *testing.T) {
	customJobType := workers.JobType("custom-test")
	called := false

	handler := workers.NewBaseHandler(customJobType, func(ctx context.Context, job workers.Job) error {
		called = true
		return nil
	})

	// Test CanHandle
	if !handler.CanHandle(customJobType) {
		t.Error("BaseHandler should handle its custom job type")
	}

	if handler.CanHandle(workers.JobTypeSampleTask1) {
		t.Error("BaseHandler should not handle other job types")
	}

	// Test Handle
	job := workers.Job{
		ID:      "custom-1",
		Type:    customJobType,
		Payload: nil,
	}

	err := handler.Handle(context.Background(), job)
	if err != nil {
		t.Errorf("BaseHandler should not return error: %v", err)
	}

	if !called {
		t.Error("BaseHandler did not call the custom function")
	}
}

func TestBaseHandler_ContextCancellation(t *testing.T) {
	customJobType := workers.JobType("cancel-test")

	handler := workers.NewBaseHandler(customJobType, func(ctx context.Context, job workers.Job) error {
		<-ctx.Done()
		return ctx.Err()
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	job := workers.Job{
		ID:      "cancel-1",
		Type:    customJobType,
		Payload: nil,
	}

	err := handler.Handle(ctx, job)
	if err == nil {
		t.Error("Expected context cancellation error")
	}
}
