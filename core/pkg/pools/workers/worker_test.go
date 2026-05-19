package workers_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"ntels.com/pharos/core/pkg/pools/workers"
)

func TestWorkerPool(t *testing.T) {
	// Create worker pool with custom options
	opts := workers.WorkerPoolOptions{
		WorkerCount:   10,
		JobQueueSize:  100,
		EnableMetrics: true,
	}

	pool := workers.NewWorkerPool(opts)

	// Register sample handlers
	pool.RegisterHandler(workers.NewSampleHandler1())
	pool.RegisterHandler(workers.NewSampleHandler2())

	// Start the pool
	pool.Start()
	defer pool.Stop()

	// Submit test jobs
	for i := range 50 {
		job := workers.Job{
			ID:      fmt.Sprintf("sample1-%d", i),
			Type:    workers.JobTypeSampleTask1,
			Payload: map[string]any{"index": i},
		}

		if err := pool.Submit(job); err != nil {
			t.Errorf("Failed to submit job: %v", err)
		}
	}

	for i := range 50 {
		job := workers.Job{
			ID:      fmt.Sprintf("sample2-%d", i),
			Type:    workers.JobTypeSampleTask2,
			Payload: map[string]any{"index": i},
		}

		if err := pool.Submit(job); err != nil {
			t.Errorf("Failed to submit job: %v", err)
		}
	}

	// Collect results
	time.Sleep(2 * time.Second)

	metrics := pool.GetMetrics()
	t.Logf("Metrics - Total: %d, Processed: %d, Failed: %d",
		metrics.TotalJobs,
		metrics.ProcessedJobs,
		metrics.FailedJobs,
	)

	if metrics.TotalJobs != 100 {
		t.Errorf("Expected 100 total jobs, got %d", metrics.TotalJobs)
	}
}

func TestBatchSubmission(t *testing.T) {
	opts := workers.DefaultOptions()
	opts.WorkerCount = 20

	pool := workers.NewWorkerPool(opts)
	pool.RegisterHandler(workers.NewSampleHandler1())
	pool.Start()
	defer pool.Stop()

	// Create batch submitter
	batchSubmitter := workers.NewBatchSubmitter(pool, 1000)

	// Submit large batch
	jobCount := 10000
	for i := range jobCount {
		job := workers.Job{
			ID:      fmt.Sprintf("batch-job-%d", i),
			Type:    workers.JobTypeSampleTask1,
			Payload: map[string]any{"data": i},
		}

		if err := batchSubmitter.Add(job); err != nil {
			t.Errorf("Failed to add job to batch: %v", err)
		}
	}

	// Flush remaining jobs
	if err := batchSubmitter.Flush(); err != nil {
		t.Errorf("Failed to flush batch: %v", err)
	}

	// Wait for processing
	time.Sleep(3 * time.Second)

	metrics := pool.GetMetrics()
	t.Logf("Batch Metrics - Total: %d, Processed: %d, Failed: %d",
		metrics.TotalJobs,
		metrics.ProcessedJobs,
		metrics.FailedJobs,
	)
}

func TestManager(t *testing.T) {
	manager := workers.NewManager()
	defer manager.Shutdown()

	// Create multiple pools for different priorities
	err := manager.CreatePool("high-priority", workers.WorkerPoolOptions{
		WorkerCount:   50,
		JobQueueSize:  5000,
		EnableMetrics: true,
	})
	if err != nil {
		t.Fatalf("Failed to create high-priority pool: %v", err)
	}

	err = manager.CreatePool("normal-priority", workers.WorkerPoolOptions{
		WorkerCount:   30,
		JobQueueSize:  3000,
		EnableMetrics: true,
	})
	if err != nil {
		t.Fatalf("Failed to create normal-priority pool: %v", err)
	}

	// Get pools and register handlers
	highPool, _ := manager.GetPool("high-priority")
	highPool.RegisterHandler(workers.NewSampleHandler1())

	normalPool, _ := manager.GetPool("normal-priority")
	normalPool.RegisterHandler(workers.NewSampleHandler2())

	// Start all pools
	manager.StartAll()

	// Submit jobs to different pools
	for i := range 100 {
		job := workers.Job{
			ID:      fmt.Sprintf("high-priority-job-%d", i),
			Type:    workers.JobTypeSampleTask1,
			Payload: nil,
		}
		highPool.Submit(job)
	}

	for i := range 100 {
		job := workers.Job{
			ID:      fmt.Sprintf("normal-priority-job-%d", i),
			Type:    workers.JobTypeSampleTask2,
			Payload: nil,
		}
		normalPool.Submit(job)
	}

	// Wait and check metrics
	time.Sleep(2 * time.Second)

	allMetrics := manager.GetAllMetrics()
	for poolName, metrics := range allMetrics {
		t.Logf("Pool[%s] - Total: %d, Processed: %d, Failed: %d",
			poolName,
			metrics.TotalJobs,
			metrics.ProcessedJobs,
			metrics.FailedJobs,
		)
	}
}

func TestCustomHandler(t *testing.T) {
	opts := workers.DefaultOptions()
	pool := workers.NewWorkerPool(opts)

	// Create custom handler using BaseHandler
	customJobType := workers.JobType("custom")
	customHandler := workers.NewBaseHandler(customJobType, func(ctx context.Context, job workers.Job) error {
		// Custom processing logic
		t.Logf("Processing custom job: %s", job.ID)
		return nil
	})

	pool.RegisterHandler(customHandler)
	pool.Start()
	defer pool.Stop()

	// Submit custom job
	job := workers.Job{
		ID:      "custom-1",
		Type:    customJobType,
		Payload: "custom data",
	}

	if err := pool.Submit(job); err != nil {
		t.Errorf("Failed to submit custom job: %v", err)
	}

	time.Sleep(1 * time.Second)
}

func TestRateLimiter(t *testing.T) {
	rateLimiter := workers.NewRateLimiter(100) // 100 jobs per second

	ctx := context.Background()
	start := time.Now()

	for range 250 {
		if err := rateLimiter.Wait(ctx); err != nil {
			t.Errorf("Rate limiter wait failed: %v", err)
		}
	}

	elapsed := time.Since(start)
	t.Logf("Processed 250 jobs with rate limit in %v", elapsed)

	// Should take at least 2 seconds (250 jobs at 100/sec = 2.5s)
	if elapsed < 2*time.Second {
		t.Errorf("Rate limiter not working correctly, elapsed time: %v", elapsed)
	}
}

// Benchmark tests
func TestTrySubmit(t *testing.T) {
	// Synchronization: ensure worker is blocked before filling queue
	workerBlocked := make(chan struct{})
	blockChan := make(chan struct{})

	blockingHandler := workers.NewBaseHandler(workers.JobTypeSampleTask1, func(ctx context.Context, job workers.Job) error {
		// Signal that worker has started processing
		select {
		case workerBlocked <- struct{}{}:
		default:
		}
		// Block until test is done
		<-blockChan
		return nil
	})

	opts := workers.WorkerPoolOptions{
		WorkerCount:  1,
		JobQueueSize: 1, // Queue size 1: worker takes 1st job, 2nd job fills queue
	}

	pool := workers.NewWorkerPool(opts)
	pool.RegisterHandler(blockingHandler)
	pool.Start()
	defer func() {
		close(blockChan) // Unblock workers before stopping
		pool.Stop()
	}()

	// Submit first job and wait for worker to start processing it
	job1 := workers.Job{ID: "1", Type: workers.JobTypeSampleTask1, Payload: nil}
	if err := pool.TrySubmit(job1); err != nil {
		t.Fatalf("First TrySubmit should succeed: %v", err)
	}

	// Wait for worker to pick up job1 and enter blocking state
	select {
	case <-workerBlocked:
		// Worker is now blocked on job1
	case <-time.After(5 * time.Second):
		t.Fatal("Worker did not start processing job1")
	}

	// Now fill the queue - worker is busy with job1, job2 should fill the queue
	job2 := workers.Job{ID: "2", Type: workers.JobTypeSampleTask1, Payload: nil}
	if err := pool.TrySubmit(job2); err != nil {
		t.Errorf("Second TrySubmit should succeed: %v", err)
	}

	// Queue is now full (worker processing job1, job2 in queue)
	// job3 should fail
	job3 := workers.Job{ID: "3", Type: workers.JobTypeSampleTask1, Payload: nil}
	err := pool.TrySubmit(job3)
	if err == nil {
		t.Error("TrySubmit should fail when queue is full")
	} else if err.Error() != "job queue is full" {
		t.Errorf("Expected 'job queue is full', got: %v", err)
	}
}

func TestProcessJob_NoHandler(t *testing.T) {
	opts := workers.DefaultOptions()
	pool := workers.NewWorkerPool(opts)
	// Don't register any handlers
	pool.Start()
	defer pool.Stop()

	// Submit job with unknown type
	unknownJobType := workers.JobType("unknown-type")
	job := workers.Job{
		ID:      "no-handler",
		Type:    unknownJobType,
		Payload: nil,
	}

	if err := pool.Submit(job); err != nil {
		t.Errorf("Submit should not fail: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Check that job was marked as failed
	metrics := pool.GetMetrics()
	if metrics.FailedJobs != 1 {
		t.Errorf("Expected 1 failed job, got %d", metrics.FailedJobs)
	}
}

func TestShutdown(t *testing.T) {
	opts := workers.DefaultOptions()
	pool := workers.NewWorkerPool(opts)
	pool.RegisterHandler(workers.NewSampleHandler1())
	pool.Start()

	// Submit some jobs
	for i := range 10 {
		job := workers.Job{
			ID:      fmt.Sprintf("shutdown-job-%d", i),
			Type:    workers.JobTypeSampleTask1,
			Payload: nil,
		}
		pool.Submit(job)
	}

	time.Sleep(100 * time.Millisecond)

	// Call Shutdown (should cancel context and stop)
	pool.Shutdown()

	// Try to submit after shutdown
	job := workers.Job{
		ID:      "after-shutdown",
		Type:    workers.JobTypeSampleTask1,
		Payload: nil,
	}
	err := pool.Submit(job)
	if err == nil {
		t.Error("Submit should fail after shutdown")
	}
	// 에러 메시지는 "worker pool is stopped" 또는 "worker pool is shutting down" 모두 가능
	if err.Error() != "worker pool is stopped" && err.Error() != "worker pool is shutting down" {
		t.Errorf("Expected shutdown error, got: %v", err)
	}
}

func TestMonitorMetrics(t *testing.T) {
	manager := workers.NewManager()
	defer manager.Shutdown()

	manager.CreatePool("monitor-test", workers.DefaultOptions())
	pool, _ := manager.GetPool("monitor-test")
	pool.RegisterHandler(workers.NewSampleHandler1())
	manager.StartAll()

	// Start monitoring in background
	done := make(chan struct{})
	go func() {
		manager.MonitorMetrics(100 * time.Millisecond)
		close(done)
	}()

	// Submit some jobs
	for i := range 20 {
		pool.Submit(workers.Job{
			ID:      fmt.Sprintf("monitor-job-%d", i),
			Type:    workers.JobTypeSampleTask1,
			Payload: nil,
		})
	}

	// Wait for at least one monitoring cycle
	time.Sleep(250 * time.Millisecond)

	// Shutdown manager (should stop MonitorMetrics)
	manager.Shutdown()

	// Wait for MonitorMetrics to exit
	select {
	case <-done:
		// Success - MonitorMetrics exited
	case <-time.After(1 * time.Second):
		t.Error("MonitorMetrics did not exit after shutdown")
	}
}

func TestRateLimiter_Close(t *testing.T) {
	rateLimiter := workers.NewRateLimiter(10) // Smaller rate

	ctx := context.Background()

	// Drain all tokens first
	for range 10 {
		if err := rateLimiter.Wait(ctx); err != nil {
			t.Errorf("Wait should succeed before close: %v", err)
		}
	}

	// Close the rate limiter
	rateLimiter.Close()

	// Wait should fail after close (no more tokens, and closed)
	err := rateLimiter.Wait(ctx)
	if err == nil {
		t.Fatal("Wait should fail after close")
	}
	if err.Error() != "rate limiter is closed" {
		t.Errorf("Expected 'rate limiter is closed', got: %v", err)
	}

	// 멀등성 테스트: 두 번 Close 호출 가능
	rateLimiter.Close() // 에러 없이 종료되어야 함
}

func TestRateLimiter_ContextCancel(t *testing.T) {
	rateLimiter := workers.NewRateLimiter(10) // Low rate
	defer rateLimiter.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// Start waiting in background
	errCh := make(chan error, 1)
	go func() {
		// Drain all tokens first
		for range 10 {
			rateLimiter.Wait(context.Background())
		}
		// Now wait should block
		errCh <- rateLimiter.Wait(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Should get context error
	select {
	case err := <-errCh:
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Wait did not respond to context cancellation")
	}
}

func TestRegisterHandler_BeforeStart(t *testing.T) {
	pool := workers.NewWorkerPool(workers.DefaultOptions())
	handler := workers.NewSampleHandler1()
	err := pool.RegisterHandler(handler)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	pool.Start()
	defer pool.Stop()
}

func TestRegisterHandler_AfterStart(t *testing.T) {
	pool := workers.NewWorkerPool(workers.DefaultOptions())
	pool.Start()
	defer pool.Stop()
	handler := workers.NewSampleHandler1()
	err := pool.RegisterHandler(handler)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, workers.ErrPoolStarted) {
		t.Errorf("Expected ErrPoolStarted, got: %v", err)
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrPoolStopped", workers.ErrPoolStopped},
		{"ErrPoolShuttingDown", workers.ErrPoolShuttingDown},
		{"ErrQueueFull", workers.ErrQueueFull},
		{"ErrPoolStarted", workers.ErrPoolStarted},
		{"ErrNoHandler", workers.ErrNoHandler},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s is nil", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("%s has empty error message", tt.name)
			}
		})
	}
}

func BenchmarkWorkerPool(b *testing.B) {
	opts := workers.WorkerPoolOptions{
		WorkerCount:   100,
		JobQueueSize:  10000,
		EnableMetrics: false, // Disable for benchmark
	}

	pool := workers.NewWorkerPool(opts)
	pool.RegisterHandler(workers.NewSampleHandler1())
	pool.Start()
	defer pool.Stop()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		job := workers.Job{
			ID:      fmt.Sprintf("bench-%d", i),
			Type:    workers.JobTypeSampleTask1,
			Payload: nil,
		}
		pool.Submit(job)
	}
}
