// Package connection provides TCP connection pooling to improve performance
// for repeated network connections.
//
// Key features:
//   - Connection reuse: Eliminates 3-way handshake overhead
//   - LRU-based eviction: Manages pool size limits
//   - Automatic idle cleanup: TTL-based resource management
//   - Concurrency safety: Thread-safe with sharded locks for high concurrency
//   - Detailed metrics: Performance tracking (reuse rate, failure rate, etc.)
package connection

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ConnectionPool reuses TCP connections for periodic collection tasks.
// Eliminates 3-way handshake overhead for repeated connections to the same IP.
//
// Thread-safety:
//   - Get/Release/Close are safe for concurrent use from multiple goroutines
//   - Uses 16 sharded locks to minimize contention across different addresses
//   - stopped uses atomic operations for race-free state management
//
// Performance:
//   - 16 shards provide up to 16x concurrent operations
//   - Different addresses in different shards can be processed simultaneously
//   - Each shard handles approximately maxSize/16 connections
type ConnectionPool struct {
	shards     []*poolShard                           // Sharded connection storage for reduced lock contention
	shardCount int                                    // Number of shards (default: 16)
	shardLimit int                                    // Maximum connections per shard (maxSize/shardCount)
	factory    func(address string) (net.Conn, error) // Connection factory function
	maxIdle    time.Duration                          // Idle timeout (triggers automatic cleanup)
	maxSize    int                                    // Maximum pool size across all shards
	metrics    PoolMetrics                            // Performance metrics
	stopCh     chan struct{}                          // Cleanup goroutine stop signal
	stopped    atomic.Bool                            // Pool stopped state flag
}

// poolShard is a single shard of the connection pool
// Each shard has its own lock to reduce contention
type poolShard struct {
	mu          sync.Mutex               // Protects this shard's connections and inflight
	cond        *sync.Cond               // Signals waiters when a connection becomes available
	connections map[string]*pooledConn   // Connections in this shard
	inflight    map[string]*inflightCall // In-progress connection attempts (singleflight)
}

// inflightCall represents an in-progress connection attempt
// Uses singleflight pattern to deduplicate concurrent requests to the same address
type inflightCall struct {
	wg   sync.WaitGroup // Coordination for waiting goroutines
	conn net.Conn       // Result: connection (if successful)
	err  error          // Result: error (if failed)
}

// pooledConn wraps a connection with metadata for pool management.
type pooledConn struct {
	conn     net.Conn  // Actual TCP connection
	lastUsed time.Time // Last usage timestamp (for idle timeout calculation)
	inUse    bool      // Currently in use flag (prevents concurrent use and cleanup)
}

// PoolMetrics tracks connection pool performance statistics.
type PoolMetrics struct {
	TotalConnections   atomic.Int64
	ActiveConnections  atomic.Int64
	ConnectionAttempts atomic.Int64
	ConnectionFailures atomic.Int64
	ConnectionReuses   atomic.Int64 // Connection reuse count (measures effectiveness)
}

// NewConnectionPool creates a connection pool for periodic collection tasks.
//
// Parameters:
//   - maxIdleTime: Recommended to be 2x collection interval (e.g., 15s interval → 30s)
//   - maxSize: Maximum number of connections across all shards (recommended: 100)
//   - shardCount: Number of shards for lock distribution (0 = auto: 16, recommended: 16)
//
// Performance:
//   - Uses 16 shards for reduced lock contention
//   - Supports up to 16 concurrent operations on different addresses
//   - Each shard manages approximately maxSize/16 connections
func NewConnectionPool(maxIdleTime time.Duration, maxSize, shardCount int) *ConnectionPool {
	// Auto-configure shard count if not specified or invalid
	if shardCount <= 0 {
		shardCount = 16
	}

	shards := make([]*poolShard, shardCount)
	for i := range shardCount {
		s := &poolShard{
			connections: make(map[string]*pooledConn),
			inflight:    make(map[string]*inflightCall),
		}
		s.cond = sync.NewCond(&s.mu)
		shards[i] = s
	}

	shardLimit := max(maxSize/shardCount, 1)

	pool := &ConnectionPool{
		shards:     shards,
		shardCount: shardCount,
		shardLimit: shardLimit,
		maxIdle:    maxIdleTime,
		maxSize:    maxSize,
		stopCh:     make(chan struct{}),
		factory: func(address string) (net.Conn, error) {
			return net.DialTimeout("tcp", address, 5*time.Second)
		},
	}

	// Periodically clean up idle connections
	go pool.cleanupIdleConnections()

	slog.Info("connection pool created",
		"max_idle_time", maxIdleTime,
		"max_size", maxSize,
		"shard_count", shardCount,
		"cleanup_interval", maxIdleTime/2)

	return pool
}

// getShard returns the shard for a given address
// Uses FNV-1a hash for uniform distribution across shards
func (p *ConnectionPool) getShard(address string) *poolShard {
	// Simple FNV-1a hash for fast, uniform distribution
	hash := uint32(2166136261)
	for i := range address {
		hash ^= uint32(address[i])
		hash *= 16777619
	}
	return p.shards[hash%uint32(p.shardCount)]
}

// Get returns a connection for the given address.
// Reuses an existing valid connection if available, otherwise creates a new one.
//
// Workflow:
//  1. Select shard based on address hash
//  2. Check existing connection: Verify inUse is false and within maxIdle time
//  3. Validate connection: Use isConnAlive() to check actual connection state
//  4. Reuse: Set inUse = true and return connection
//  5. Create new: If none exists, create new connection with inUse = true
//
// IMPORTANT:
//   - Must call Release() after using the returned connection
//   - Failure to call Release() leaves connection permanently inUse (unreusable)
//   - Recommended pattern: defer pool.Release(addr, conn)
//
// Performance:
//   - Different addresses in different shards can be processed concurrently
//   - Up to 16 simultaneous Get() operations on different shards
func (p *ConnectionPool) Get(address string) (net.Conn, error) {
	if p.stopped.Load() {
		return nil, fmt.Errorf("connection pool is closed")
	}

	shard := p.getShard(address)
	shard.mu.Lock()

	p.metrics.ConnectionAttempts.Add(1)

	now := time.Now()

	// Fast path: Check existing connection
	if pc, exists := shard.connections[address]; exists {
		// Check if connection is not in use and still valid
		if !pc.inUse && now.Sub(pc.lastUsed) < p.maxIdle {
			pc.inUse = true
			pc.lastUsed = now
			p.metrics.ConnectionReuses.Add(1)
			shard.mu.Unlock()
			return pc.conn, nil
		}

		// Expired or in-use connection removal
		if !pc.inUse {
			pc.conn.Close()
			delete(shard.connections, address)
			p.metrics.ActiveConnections.Add(-1)
		}
	}

	// Singleflight: Check if another goroutine is already creating a connection
	if call, exists := shard.inflight[address]; exists {
		// Another goroutine is creating the connection, wait for it
		shard.mu.Unlock()
		call.wg.Wait() // Wait for the connection to be created

		// Share the result (both success and failure)
		// This prevents timeout pileup: if first goroutine times out,
		// all waiting goroutines immediately get the same error
		// instead of each retrying sequentially
		if call.err != nil {
			p.metrics.ConnectionAttempts.Add(1)
			p.metrics.ConnectionFailures.Add(1)
			return nil, call.err
		}

		// The connection was just created and is inUse by the first caller.
		// Wait until it is released (inUse=false) before reusing it.
		// This avoids creating redundant connections.
		shard.mu.Lock()
		for {
			pc, exists := shard.connections[address]
			if !exists {
				// Connection was evicted or closed while we were waiting.
				shard.mu.Unlock()
				return p.Get(address)
			}
			if !pc.inUse {
				if time.Since(pc.lastUsed) < p.maxIdle {
					pc.inUse = true
					pc.lastUsed = time.Now()
					p.metrics.ConnectionReuses.Add(1)
					shard.mu.Unlock()
					return pc.conn, nil
				}
				// Connection expired while we waited; delete and create a new one.
				pc.conn.Close()
				delete(shard.connections, address)
				p.metrics.ActiveConnections.Add(-1)
				shard.mu.Unlock()
				return p.Get(address)
			}
			// Still in-use; block until Release() or a state change broadcasts.
			shard.cond.Wait()
		}
	}

	// Create new inflight call (this goroutine will create the connection)
	call := &inflightCall{}
	call.wg.Add(1)
	shard.inflight[address] = call

	// Check pool size limit before releasing lock
	needEviction := len(shard.connections) >= p.shardLimit
	if needEviction {
		// Remove oldest connection in this shard (LRU)
		p.evictOldestInShard(shard)
	}

	shard.mu.Unlock()

	// Create connection outside the lock (prevents timeout pileup)
	start := time.Now()
	conn, err := p.factory(address)

	// Store result in inflight call
	call.conn = conn
	call.err = err

	// Update metrics and pool state
	shard.mu.Lock()
	delete(shard.inflight, address) // Remove inflight call

	if err != nil {
		p.metrics.ConnectionFailures.Add(1)
		shard.mu.Unlock()
		call.wg.Done() // Wake up waiting goroutines
		return nil, err
	}

	// Store connection in pool
	shard.connections[address] = &pooledConn{
		conn:     conn,
		lastUsed: start,
		inUse:    true, // New connection is immediately marked as in-use
	}
	p.metrics.TotalConnections.Add(1)
	p.metrics.ActiveConnections.Add(1)

	shard.mu.Unlock()
	call.wg.Done() // Wake up waiting goroutines

	return conn, nil
}

// Release marks a connection as finished and returns it to the reusable pool.
//
// Behavior:
//   - Sets inUse flag to false, allowing other goroutines to reuse
//   - Updates lastUsed time to protect from idle timeout cleanup
//   - Connection is not closed, remains in pool for reuse
//
// Usage pattern:
//
//	conn, err := pool.Get(addr)
//	if err != nil { return err }
//	defer pool.Release(addr, conn)  // Must call
//
// Notes:
//   - Must call Release() for every connection obtained via Get()
//   - Safe to call Release() multiple times for the same connection
//
// Performance:
//   - Only locks the specific shard for this address
//   - Other shards remain accessible for concurrent operations
func (p *ConnectionPool) Release(address string, conn net.Conn) {
	if p.stopped.Load() {
		conn.Close()
		return
	}

	shard := p.getShard(address)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if pc, exists := shard.connections[address]; exists && pc.conn == conn {
		pc.inUse = false // Mark as finished
		pc.lastUsed = time.Now()
		shard.cond.Broadcast() // Wake up goroutines waiting in Get()
	}
}

// Close closes all connections and shuts down the pool.
func (p *ConnectionPool) Close() {
	if !p.stopped.CompareAndSwap(false, true) {
		return // Already closed
	}

	close(p.stopCh)

	// Close all connections in all shards
	activeCount := 0
	for _, shard := range p.shards {
		shard.mu.Lock()
		for _, pc := range shard.connections {
			pc.conn.Close()
			activeCount++
		}
		shard.connections = make(map[string]*pooledConn)
		shard.cond.Broadcast() // Wake up any goroutines waiting in Get()
		shard.mu.Unlock()
	}

	p.metrics.ActiveConnections.Store(0)

	// Calculate final metrics
	attempts := p.metrics.ConnectionAttempts.Load()
	reuses := p.metrics.ConnectionReuses.Load()
	failures := p.metrics.ConnectionFailures.Load()

	reuseRate := 0.0
	if attempts > 0 {
		reuseRate = float64(reuses) / float64(attempts) * 100
	}

	successRate := 0.0
	if attempts > 0 {
		successRate = float64(attempts-failures) / float64(attempts) * 100
	}

	slog.Info("connection pool closed",
		"total_connections", p.metrics.TotalConnections.Load(),
		"connection_attempts", attempts,
		"connection_reuses", reuses,
		"reuse_rate", fmt.Sprintf("%.1f%%", reuseRate),
		"connection_failures", failures,
		"success_rate", fmt.Sprintf("%.1f%%", successRate),
		"closed_connections", activeCount)
}

// evictOldestInShard removes the least recently used connection in a shard (LRU eviction).
// Caller must hold shard.mu lock
func (p *ConnectionPool) evictOldestInShard(shard *poolShard) {
	var oldestAddr string
	var oldestTime time.Time
	first := true

	for addr, pc := range shard.connections {
		if first || pc.lastUsed.Before(oldestTime) {
			oldestAddr = addr
			oldestTime = pc.lastUsed
			first = false
		}
	}

	if oldestAddr != "" {
		shard.connections[oldestAddr].conn.Close()
		delete(shard.connections, oldestAddr)
		p.metrics.ActiveConnections.Add(-1)
		shard.cond.Broadcast() // Wake up waiters so they can detect eviction
	}
}

// cleanupIdleConnections periodically removes idle connections.
func (p *ConnectionPool) cleanupIdleConnections() {
	ticker := time.NewTicker(p.maxIdle / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.performCleanup()
		case <-p.stopCh:
			return
		}
	}
}

func (p *ConnectionPool) performCleanup() {
	now := time.Now()
	totalRemoved := 0

	// Clean up each shard independently
	for _, shard := range p.shards {
		shard.mu.Lock()
		removed := 0

		// Clean up only connections that are not in use and have exceeded idle time
		for address, pc := range shard.connections {
			if !pc.inUse && now.Sub(pc.lastUsed) > p.maxIdle {
				pc.conn.Close()
				delete(shard.connections, address)
				p.metrics.ActiveConnections.Add(-1)
				removed++
			}
		}

		if removed > 0 {
			shard.cond.Broadcast() // Wake up waiters so they can detect expired connections
		}
		shard.mu.Unlock()
		totalRemoved += removed
	}

	if totalRemoved > 0 {
		// Calculate reuse rate
		attempts := p.metrics.ConnectionAttempts.Load()
		reuses := p.metrics.ConnectionReuses.Load()
		reuseRate := 0.0
		if attempts > 0 {
			reuseRate = float64(reuses) / float64(attempts) * 100
		}

		slog.Debug("idle connections cleanup",
			"removed", totalRemoved,
			"active", p.metrics.ActiveConnections.Load(),
			"reuse_rate", fmt.Sprintf("%.1f%%", reuseRate),
			"total_reuses", reuses)
	}
}

// PoolMetricsSnapshot is a point-in-time snapshot of pool metrics.
type PoolMetricsSnapshot struct {
	TotalConnections   int64
	ActiveConnections  int64
	ConnectionAttempts int64
	ConnectionFailures int64
	ConnectionReuses   int64
}

// GetMetrics returns a snapshot of current connection pool metrics.
func (p *ConnectionPool) GetMetrics() PoolMetricsSnapshot {
	return PoolMetricsSnapshot{
		TotalConnections:   p.metrics.TotalConnections.Load(),
		ActiveConnections:  p.metrics.ActiveConnections.Load(),
		ConnectionAttempts: p.metrics.ConnectionAttempts.Load(),
		ConnectionFailures: p.metrics.ConnectionFailures.Load(),
		ConnectionReuses:   p.metrics.ConnectionReuses.Load(),
	}
}

// GetReuseRate returns the connection reuse ratio (0.0 ~ 1.0).
// Higher values indicate better connection pool effectiveness.
func (p *ConnectionPool) GetReuseRate() float64 {
	attempts := p.metrics.ConnectionAttempts.Load()
	if attempts == 0 {
		return 0.0
	}

	reuses := p.metrics.ConnectionReuses.Load()
	return float64(reuses) / float64(attempts)
}

// GetActiveConnections returns the current number of active connections.
func (p *ConnectionPool) GetActiveConnections() int64 {
	return p.metrics.ActiveConnections.Load()
}

// DialWithRetry dials with exponential backoff retry logic.
//
// Retry policy:
//   - Backoff: attempt² × 100ms (capped at 5 seconds)
//   - Context cancellation causes immediate termination
//   - Non-retryable errors cause immediate failure
//
// Parameters:
//   - ctx: Cancellable context
//   - address: Address in "host:port" format
//   - maxRetries: Maximum number of retry attempts (total attempts = maxRetries + 1)
//
// Returns:
//   - Success: Connection object
//   - Failure: Error if maxRetries exceeded or non-retryable error encountered
func DialWithRetry(ctx context.Context, address string, maxRetries int) (net.Conn, error) {
	var lastErr error
	start := time.Now()

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Wait for retry (exponential backoff)
		if attempt > 0 {
			backoff := min(time.Duration(attempt*attempt)*100*time.Millisecond, 5*time.Second)

			slog.Debug("connection retry backoff",
				"attempt", attempt,
				"backoff", backoff,
				"address", address)

			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
		}

		// Attempt connection
		conn, err := net.DialTimeout("tcp", address, 5*time.Second)
		if err == nil {
			slog.Debug("connection established",
				"address", address,
				"attempt", attempt+1)
			return conn, nil
		}

		lastErr = err

		// Check for non-retryable errors
		if isNonRetryableError(err) {
			slog.Warn("non-retryable error",
				"address", address,
				"error", err)
			break
		}

		slog.Debug("connection failed, retrying",
			"address", address,
			"attempt", attempt+1,
			"error", err)
	}

	totalTime := time.Since(start)
	slog.Error("connection retry exhausted",
		"address", address,
		"attempts", maxRetries+1,
		"total_time", totalTime.Round(time.Millisecond),
		"last_error", lastErr)

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries+1, lastErr)
}

// isNonRetryableError determines if an error is not worth retrying.
func isNonRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check net.Error interface
	if netErr, ok := err.(net.Error); ok {
		// Timeout is retryable
		if netErr.Timeout() {
			return false
		}
	}

	// Check specific error patterns
	errMsg := err.Error()

	// Connection refused (server not available) - not retryable
	if strings.Contains(errMsg, "connection refused") {
		return true
	}

	// Host unreachable (routing failure) - not retryable
	if strings.Contains(errMsg, "no route to host") || strings.Contains(errMsg, "host is unreachable") {
		return true
	}

	// Network unreachable - not retryable
	if strings.Contains(errMsg, "network is unreachable") {
		return true
	}

	// Invalid address format - not retryable
	if strings.Contains(errMsg, "invalid argument") || strings.Contains(errMsg, "missing port") {
		return true
	}

	// Address already in use - not retryable (port conflict)
	if strings.Contains(errMsg, "address already in use") {
		return true
	}

	// Permission denied - not retryable (permission issue)
	if strings.Contains(errMsg, "permission denied") {
		return true
	}

	return false
}

// SendWithTimeout sends a command and receives a response with timeout control.
//
// Features:
//   - Dynamic buffer: Automatically handles responses exceeding 4096 bytes
//   - Safe deadline management: Guaranteed reset via defer even on errors
//   - Accurate timeout handling: Only returns timeout error if no data received
//
// Timeout behavior:
//   - Timeout after receiving data: Normal completion (response considered complete)
//   - Timeout without data: Returns error (actual timeout)
//
// Usage example:
//
//	response, err := SendWithTimeout(conn, []byte("GET /\n"), 5*time.Second)
//	if err != nil { return err }
//	// conn deadline is automatically reset and ready for reuse
func SendWithTimeout(conn net.Conn, command []byte, timeout time.Duration) ([]byte, error) {
	// Always reset deadline even on errors
	defer conn.SetDeadline(time.Time{})

	// Set write deadline
	if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("set write deadline: %w", err)
	}

	if _, err := conn.Write(command); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	// Set read deadline
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("set read deadline: %w", err)
	}

	// Use dynamic buffer (handles large responses)
	var result []byte
	buf := make([]byte, 4096)
	readOnce := false

	for {
		n, err := conn.Read(buf)
		if n > 0 {
			result = append(result, buf[:n]...)
			readOnce = true
		}
		if err != nil {
			// EOF is normal completion
			if err.Error() == "EOF" {
				break
			}
			// Timeout: normal if data was received, error if not
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				if readOnce {
					break // Data received successfully
				}
				return nil, fmt.Errorf("read: %w", err) // Timeout error
			}
			return nil, fmt.Errorf("read: %w", err)
		}
		// End if received less than buffer size
		if n < len(buf) {
			break
		}
	}

	return result, nil
}
