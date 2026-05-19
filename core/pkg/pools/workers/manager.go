package workers

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"sync"
	"sync/atomic"
	"time"
)

// Manager manages multiple worker pools for different priorities or types
// Provides centralized control, monitoring, and metrics collection for multiple pools
//
// Thread-safety: All methods are safe for concurrent use
//
// Typical usage:
//
//	manager := workers.NewManager()
//	manager.CreatePool("high-priority", highPriorityOpts)
//	manager.CreatePool("normal", normalOpts)
//	manager.StartAll()
//	defer manager.Shutdown()
type Manager struct {
	pools  map[string]*WorkerPool // Named worker pools
	mu     sync.RWMutex           // Protects pools map
	ctx    context.Context        // Cancellation context for monitoring
	cancel context.CancelFunc     // Cancellation function
}

// NewManager creates a new worker pool manager
func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		pools:  make(map[string]*WorkerPool),
		ctx:    ctx,
		cancel: cancel,
	}
}

// CreatePool creates a new worker pool with the given name and options
func (m *Manager) CreatePool(name string, opts WorkerPoolOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.pools[name]; exists {
		return fmt.Errorf("pool %s already exists", name)
	}

	pool := NewWorkerPool(opts)
	m.pools[name] = pool

	return nil
}

// GetPool returns a worker pool by name
func (m *Manager) GetPool(name string) (*WorkerPool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pool, exists := m.pools[name]
	if !exists {
		return nil, fmt.Errorf("pool %s not found", name)
	}

	return pool, nil
}

// StartAll starts all worker pools
func (m *Manager) StartAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, pool := range m.pools {
		slog.Info("Starting worker pool", "pool_name", name)
		pool.Start()
	}
}

// StopAll stops all worker pools gracefully
func (m *Manager) StopAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, pool := range m.pools {
		slog.Info("Stopping worker pool", "pool_name", name)
		pool.Stop()
	}
}

// Shutdown shuts down the manager and all worker pools
func (m *Manager) Shutdown() {
	m.cancel()
	m.StopAll()
}

// GetAllMetrics returns metrics for all worker pools
// Optimized to hold RLock for minimal duration
func (m *Manager) GetAllMetrics() map[string]Metrics {
	m.mu.RLock()
	// Quick snapshot of pools under lock
	pools := make(map[string]*WorkerPool, len(m.pools))
	maps.Copy(pools, m.pools)
	m.mu.RUnlock() // Release lock quickly

	// Collect metrics without holding lock
	metrics := make(map[string]Metrics)
	for name, pool := range pools {
		metrics[name] = pool.GetMetrics()
	}

	return metrics
}

// MonitorMetrics starts a goroutine that periodically logs metrics
// The goroutine runs until the manager is shut down
//
// Metrics are logged using slog.Info with the following fields:
//   - pool_name: Name of the worker pool
//   - total_jobs: Total jobs submitted
//   - processed_jobs: Successfully completed jobs
//   - success_rate: Percentage of successful jobs
//   - failed_jobs: Number of failed jobs
//   - retried_jobs: Number of retry attempts
//   - queue_size: Current queue depth
//   - avg_time_ms: Average processing time in milliseconds
//
// Usage:
//
//	manager.MonitorMetrics(10 * time.Second) // Log every 10 seconds
func (m *Manager) MonitorMetrics(interval time.Duration) {
	go m.monitorMetrics(interval)
}

func (m *Manager) monitorMetrics(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			metrics := m.GetAllMetrics()
			for name, metric := range metrics {
				avgTimeMs := float64(metric.AvgProcessingTime.Nanoseconds()) / 1e6
				successRate := float64(0)
				if metric.TotalJobs > 0 {
					successRate = float64(metric.ProcessedJobs) / float64(metric.TotalJobs) * 100
				}

				slog.Info("Worker pool metrics",
					"pool_name", name,
					"total_jobs", metric.TotalJobs,
					"processed_jobs", metric.ProcessedJobs,
					"success_rate", fmt.Sprintf("%.1f%%", successRate),
					"failed_jobs", metric.FailedJobs,
					"retried_jobs", metric.RetriedJobs,
					"queue_size", metric.QueueSize,
					"avg_time_ms", fmt.Sprintf("%.2f", avgTimeMs),
				)
			}
		}
	}
}

// GetAllHealthStatus returns health status for all worker pools
// Optimized to hold RLock for minimal duration
func (m *Manager) GetAllHealthStatus() map[string]HealthStatus {
	m.mu.RLock()
	// Quick snapshot of pools under lock
	pools := make(map[string]*WorkerPool, len(m.pools))
	maps.Copy(pools, m.pools)
	m.mu.RUnlock() // Release lock quickly

	// Collect health status without holding lock
	status := make(map[string]HealthStatus)
	for name, pool := range pools {
		status[name] = pool.HealthCheck()
	}

	return status
}

// BatchSubmitter helps submit large batches of jobs efficiently
// Automatically flushes jobs when batch size is reached
//
// Benefits:
//   - Reduces lock contention by batching submissions
//   - Amortizes channel send overhead
//   - Provides automatic flushing
//
// Usage:
//
//	submitter := workers.NewBatchSubmitter(pool, 100)
//	defer submitter.Flush() // Flush remaining jobs
//
//	for _, data := range largeDataset {
//	    job := workers.Job{...}
//	    if err := submitter.Add(job); err != nil {
//	        return err
//	    }
//	}
type BatchSubmitter struct {
	pool      *WorkerPool // Target worker pool
	batchSize int         // Number of jobs before auto-flush
	buffer    []Job       // Buffered jobs
	mu        sync.Mutex  // Protects buffer
}

// NewBatchSubmitter creates a new batch submitter
func NewBatchSubmitter(pool *WorkerPool, batchSize int) *BatchSubmitter {
	return &BatchSubmitter{
		pool:      pool,
		batchSize: batchSize,
		buffer:    make([]Job, 0, batchSize),
	}
}

// Add adds a job to the batch buffer
func (bs *BatchSubmitter) Add(job Job) error {
	return bs.AddWithContext(context.Background(), job)
}

// AddWithContext adds a job to the batch buffer with context support
func (bs *BatchSubmitter) AddWithContext(ctx context.Context, job Job) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.buffer = append(bs.buffer, job)

	if len(bs.buffer) >= bs.batchSize {
		return bs.flushLocked()
	}

	return nil
}

// Flush submits all buffered jobs to the worker pool
func (bs *BatchSubmitter) Flush() error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	return bs.flushLocked()
}

func (bs *BatchSubmitter) flushLocked() error {
	if len(bs.buffer) == 0 {
		return nil
	}

	err := bs.pool.SubmitBatch(bs.buffer)
	bs.buffer = bs.buffer[:0] // Clear buffer

	return err
}

// RateLimiter provides rate limiting for job submission
type RateLimiter struct {
	limiter chan struct{}
	stopCh  chan struct{}
	closed  atomic.Bool
}

// NewRateLimiter creates a new rate limiter
//
// CRITICAL: Must call Close() when done to prevent goroutine leak
// The rate limiter starts a background goroutine that must be cleaned up.
//
// Usage pattern:
//
//	limiter := NewRateLimiter(100)  // 100 requests per second
//	defer limiter.Close()           // Always close to prevent leak
//
//	for _, job := range jobs {
//	    if err := limiter.Wait(ctx); err != nil {
//	        return err
//	    }
//	    pool.Submit(job)
//	}
//
// Parameters:
//   - ratePerSecond: Maximum number of operations allowed per second
//
// Returns:
//   - *RateLimiter: Must be closed with Close() when done
func NewRateLimiter(ratePerSecond int) *RateLimiter {
	limiter := make(chan struct{}, ratePerSecond)
	stopCh := make(chan struct{})

	// Fill the limiter
	for range ratePerSecond {
		limiter <- struct{}{}
	}

	rl := &RateLimiter{
		limiter: limiter,
		stopCh:  stopCh,
	}

	// Refill the limiter every second
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-rl.stopCh:
				return
			case <-ticker.C:
				for range ratePerSecond {
					select {
					case limiter <- struct{}{}:
					default:
						// Limiter is full
					}
				}
			}
		}
	}()

	return rl
}

// Wait waits for rate limit permission
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-rl.stopCh:
		return fmt.Errorf("rate limiter is closed")
	case <-rl.limiter:
		return nil
	}
}

// Close stops the rate limiter and releases resources (idempotent)
func (rl *RateLimiter) Close() {
	if !rl.closed.CompareAndSwap(false, true) {
		return // Already closed
	}
	close(rl.stopCh)
}
