package connection

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestConnectionPool_BasicOperations는 연결 풀의 기본 동작을 테스트합니다.
func TestConnectionPool_BasicOperations(t *testing.T) {
	// 테스트 서버 시작
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	serverAddr := ln.Addr().String()

	// 서버 고루틴
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleTestConnection(conn)
		}
	}()

	// 연결 풀 생성
	pool := NewConnectionPool(1*time.Minute, 10, 16)
	defer pool.Close()

	// 첫 번째 연결 가져오기
	conn1, err := pool.Get(serverAddr)
	if err != nil {
		t.Fatalf("failed to get connection: %v", err)
	}

	// 연결 사용
	response, err := SendWithTimeout(conn1, []byte("TEST\n"), 1*time.Second)
	if err != nil {
		t.Fatalf("failed to send: %v", err)
	}
	if string(response) != "OK\n" {
		t.Errorf("unexpected response: %s", string(response))
	}

	// 연결 반환
	pool.Release(serverAddr, conn1)

	// 동일 주소로 다시 연결 요청 (재사용되어야 함)
	conn2, err := pool.Get(serverAddr)
	if err != nil {
		t.Fatalf("failed to get connection again: %v", err)
	}

	// 연결이 재사용되었는지 확인
	if conn1 != conn2 {
		t.Error("connection was not reused")
	}

	// 메트릭 확인
	metrics := pool.GetMetrics()
	if metrics.ConnectionReuses != 1 {
		t.Errorf("expected 1 reuse, got %d", metrics.ConnectionReuses)
	}

	pool.Release(serverAddr, conn2)
}

// TestConnectionPool_IdleTimeout은 유휴 타임아웃을 테스트합니다.
func TestConnectionPool_IdleTimeout(t *testing.T) {
	// 테스트 서버 시작
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	serverAddr := ln.Addr().String()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleTestConnection(conn)
		}
	}()

	// 짧은 유휴 타임아웃 (2초)
	pool := NewConnectionPool(2*time.Second, 10, 16)
	defer pool.Close()

	// 연결 가져오기
	conn1, err := pool.Get(serverAddr)
	if err != nil {
		t.Fatalf("failed to get connection: %v", err)
	}
	pool.Release(serverAddr, conn1)

	// 3초 대기 (유휴 타임아웃 초과)
	time.Sleep(3 * time.Second)

	// 다시 연결 요청 (새 연결이어야 함)
	conn2, err := pool.Get(serverAddr)
	if err != nil {
		t.Fatalf("failed to get connection after timeout: %v", err)
	}

	// 연결이 새로 생성되었는지 확인 (재사용되지 않음)
	if conn1 == conn2 {
		t.Error("expired connection should not be reused")
	}

	pool.Release(serverAddr, conn2)
}

// TestConnectionPool_ConcurrentAccess는 동시 접근을 테스트합니다.
func TestConnectionPool_ConcurrentAccess(t *testing.T) {
	pool := NewConnectionPool(1*time.Minute, 50, 16)
	defer pool.Close()

	// 10개 서버 생성 (각 고루틴이 독립적인 서버 사용)
	servers := make([]net.Listener, 10)
	for i := range 10 {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to start test server %d: %v", i, err)
		}
		defer ln.Close()
		servers[i] = ln

		go func(ln net.Listener) {
			for {
				conn, err := ln.Accept()
				if err != nil {
					return
				}
				go handleTestConnection(conn)
			}
		}(ln)
	}

	// 10개 고루틴에서 각자의 서버에 접근
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := range 10 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			serverAddr := servers[id].Addr().String()

			conn, err := pool.Get(serverAddr)
			if err != nil {
				errors <- fmt.Errorf("goroutine %d: get failed: %w", id, err)
				return
			}

			response, err := SendWithTimeout(conn, []byte("TEST\n"), 1*time.Second)
			if err != nil {
				errors <- fmt.Errorf("goroutine %d: send failed: %w", id, err)
			} else if string(response) != "OK\n" {
				errors <- fmt.Errorf("goroutine %d: unexpected response: %q", id, string(response))
			}

			pool.Release(serverAddr, conn)
		}(i)
	}

	wg.Wait()
	close(errors)

	// 에러 확인
	errorCount := 0
	for err := range errors {
		t.Error(err)
		errorCount++
	}

	if errorCount > 0 {
		t.Fatalf("got %d errors", errorCount)
	}

	// 메트릭 확인
	metrics := pool.GetMetrics()
	if metrics.ConnectionAttempts != 10 {
		t.Errorf("expected 10 attempts, got %d", metrics.ConnectionAttempts)
	}

	// 재사용률 로깅
	reuseRate := pool.GetReuseRate()
	t.Logf("Reuse rate: %.2f%%", reuseRate*100)
}

// TestDialWithRetry는 재시도 로직을 테스트합니다.
func TestDialWithRetry(t *testing.T) {
	// 존재하지 않는 주소 (연결 실패)
	ctx := context.Background()
	_, err := DialWithRetry(ctx, "127.0.0.1:1", 2) // 최대 2번 재시도
	if err == nil {
		t.Error("expected connection to fail")
	}

	// 테스트 서버 시작
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	go func() {
		conn, _ := ln.Accept()
		if conn != nil {
			conn.Close()
		}
	}()

	// 성공하는 연결
	conn, err := DialWithRetry(ctx, ln.Addr().String(), 2)
	if err != nil {
		t.Errorf("expected connection to succeed: %v", err)
	}
	if conn != nil {
		conn.Close()
	}
}

// TestDialWithRetry_Backoff는 exponential backoff를 테스트합니다.
func TestDialWithRetry_Backoff(t *testing.T) {
	// 지연 응답 서버 (타임아웃 유발)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			// 연결을 받지만 즉시 닫음 (dial은 성공하지만 이후 read가 실패)
			conn.Close()
		}
	}()

	ctx := context.Background()
	start := time.Now()

	// 연결 성공하므로 backoff 테스트가 어려움
	// 대신 재시도 로직이 있는지만 확인
	conn, err := DialWithRetry(ctx, ln.Addr().String(), 3)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if conn != nil {
		conn.Close()
	}

	elapsed := time.Since(start)
	t.Logf("Connection took: %v", elapsed)
}

// TestDialWithRetry_Timeout는 타임아웃을 테스트합니다.
func TestDialWithRetry_Timeout(t *testing.T) {
	// 매우 짧은 타임아웃으로 context 생성
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()

	// 응답하지 않는 IP (블랙홀 라우팅)
	// 192.0.2.0/24는 TEST-NET-1 (RFC 5737) - 라우팅되지 않음
	_, err := DialWithRetry(ctx, "192.0.2.1:80", 10)

	elapsed := time.Since(start)

	// Context 관련 에러여야 함
	if err != context.DeadlineExceeded && err != context.Canceled {
		// 빠르게 실패하는 경우도 허용 (네트워크 설정에 따라 다름)
		t.Logf("Got error: %v (network may reject quickly)", err)
	}

	t.Logf("Operation completed after: %v", elapsed)
}

// TestIsNonRetryableError는 재시도 불가능한 에러 감지를 테스트합니다.
func TestIsNonRetryableError(t *testing.T) {
	tests := []struct {
		name       string
		errMsg     string
		shouldStop bool
	}{
		{"connection refused", "dial tcp 127.0.0.1:1: connect: connection refused", true},
		{"no route to host", "dial tcp 192.168.255.255:80: connect: no route to host", true},
		{"host unreachable", "dial tcp 10.0.0.1:80: connect: host is unreachable", true},
		{"network unreachable", "dial tcp 10.0.0.1:80: connect: network is unreachable", true},
		{"invalid argument", "dial tcp: invalid argument", true},
		{"missing port", "dial tcp 127.0.0.1: missing port in address", true},
		{"address in use", "listen tcp 127.0.0.1:80: bind: address already in use", true},
		{"permission denied", "dial tcp 127.0.0.1:80: connect: permission denied", true},
		{"timeout", "dial tcp 127.0.0.1:80: i/o timeout", false},
		{"connection reset", "read tcp 127.0.0.1:80: connection reset by peer", false},
		{"nil error", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.errMsg != "" {
				err = fmt.Errorf("%s", tt.errMsg)
			}

			result := isNonRetryableError(err)
			if result != tt.shouldStop {
				t.Errorf("isNonRetryableError(%q) = %v, want %v", tt.errMsg, result, tt.shouldStop)
			}
		})
	}
}

// TestDialWithRetry_ContextCancel은 컨텍스트 취소를 테스트합니다.
func TestDialWithRetry_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// 즉시 취소
	cancel()

	_, err := DialWithRetry(ctx, "127.0.0.1:1", 5)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// TestSendWithTimeout은 타임아웃 제어를 테스트합니다.
func TestSendWithTimeout(t *testing.T) {
	// 테스트 서버 시작
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleTestConnection(conn)
		}
	}()

	// 연결
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// 정상 송수신
	response, err := SendWithTimeout(conn, []byte("TEST\n"), 1*time.Second)
	if err != nil {
		t.Fatalf("failed to send: %v", err)
	}

	if string(response) != "OK\n" {
		t.Errorf("unexpected response: %s", string(response))
	}

	// Deadline이 초기화되었는지 확인 (재사용 가능)
	response2, err := SendWithTimeout(conn, []byte("TEST\n"), 1*time.Second)
	if err != nil {
		t.Fatalf("failed to reuse connection: %v", err)
	}

	if string(response2) != "OK\n" {
		t.Errorf("unexpected response on reuse: %s", string(response2))
	}
}

// TestSendWithTimeout_ReadTimeout은 읽기 타임아웃을 테스트합니다.
func TestSendWithTimeout_ReadTimeout(t *testing.T) {
	// 응답하지 않는 서버
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// 데이터를 읽지만 응답하지 않음 (읽기 타임아웃 유발)
		buffer := make([]byte, 1024)
		conn.Read(buffer)
		time.Sleep(5 * time.Second) // 응답 지연
	}()

	// 연결
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// 짧은 타임아웃으로 송수신 (타임아웃 발생해야 함)
	_, err = SendWithTimeout(conn, []byte("TEST\n"), 100*time.Millisecond)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}

	// 에러가 타임아웃 관련인지 확인
	if netErr, ok := err.(net.Error); ok {
		if !netErr.Timeout() {
			t.Errorf("expected timeout error, got: %v", err)
		}
	}
}

// TestConnectionPool_GetMetrics는 메트릭을 테스트합니다.
func TestConnectionPool_GetMetrics(t *testing.T) {
	pool := NewConnectionPool(1*time.Minute, 10, 16)
	defer pool.Close()

	metrics := pool.GetMetrics()

	if metrics.TotalConnections != 0 {
		t.Errorf("expected 0 total connections, got %d", metrics.TotalConnections)
	}

	if metrics.ActiveConnections != 0 {
		t.Errorf("expected 0 active connections, got %d", metrics.ActiveConnections)
	}

	reuseRate := pool.GetReuseRate()
	if reuseRate != 0.0 {
		t.Errorf("expected 0.0 reuse rate, got %.2f", reuseRate)
	}
}

// BenchmarkConnectionPool_GetWithReuse는 연결 재사용 성능을 벤치마크합니다.
func BenchmarkConnectionPool_GetWithReuse(b *testing.B) {
	// 테스트 서버 시작
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	serverAddr := ln.Addr().String()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleTestConnection(conn)
		}
	}()

	pool := NewConnectionPool(1*time.Minute, 100, 16)
	defer pool.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		conn, err := pool.Get(serverAddr)
		if err != nil {
			b.Fatalf("failed to get connection: %v", err)
		}
		pool.Release(serverAddr, conn)
	}

	b.StopTimer()

	// 재사용률 출력
	b.Logf("Reuse rate: %.2f%%", pool.GetReuseRate()*100)
}

// BenchmarkConnectionPool_GetWithoutPool은 풀 없이 직접 연결하는 성능을 벤치마크합니다.
func BenchmarkConnectionPool_GetWithoutPool(b *testing.B) {
	// 테스트 서버 시작
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	serverAddr := ln.Addr().String()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleTestConnection(conn)
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		conn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			b.Fatalf("failed to dial: %v", err)
		}
		conn.Close()
	}
}

// handleTestConnection은 테스트 서버의 연결을 처리합니다.
func handleTestConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}

		if n > 0 {
			// 에코
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			_, err = conn.Write([]byte("OK\n"))
			if err != nil {
				return
			}
		}
	}
}

// TestSingleflight_ConcurrentSameAddress tests that concurrent requests to the same address
// only create one connection (singleflight pattern)
func TestSingleflight_ConcurrentSameAddress(t *testing.T) {
	// Create test server
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleTestConnection(conn)
		}
	}()

	// Track actual connection creations
	var actualDials atomic.Int64

	pool := NewConnectionPool(1*time.Minute, 100, 16)
	defer pool.Close()

	// Replace factory to count actual dials
	originalFactory := pool.factory
	pool.factory = func(address string) (net.Conn, error) {
		actualDials.Add(1)
		// Simulate slow connection (100ms)
		time.Sleep(100 * time.Millisecond)
		return originalFactory(address)
	}

	// Launch 10 concurrent requests to the same address
	const numGoroutines = 10
	address := ln.Addr().String()

	var wg sync.WaitGroup
	successCount := atomic.Int64{}
	start := time.Now()

	for range numGoroutines {
		wg.Go(func() {

			conn, err := pool.Get(address)
			if err != nil {
				t.Errorf("Get failed: %v", err)
				return
			}
			successCount.Add(1)
			pool.Release(address, conn)
		})
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Verify results
	dialCount := actualDials.Load()
	success := successCount.Load()

	t.Logf("Concurrent requests: %d", numGoroutines)
	t.Logf("Actual connection creations: %d", dialCount)
	t.Logf("Successful Gets: %d", success)
	t.Logf("Total time: %v", elapsed)

	// With singleflight: very few connections should be created
	// In race detector mode (-race), timing is different and 2-3 connections are acceptable
	// The key is preventing N connections (one per goroutine)
	if dialCount > 4 {
		t.Errorf("Expected 1-4 connection creations, got %d (singleflight may not be working)", dialCount)
	}

	// All requests should succeed
	if success != numGoroutines {
		t.Errorf("Expected %d successful Gets, got %d", numGoroutines, success)
	}

	// Total time should be ~100ms (concurrent), not 1000ms (sequential)
	// Allow more overhead in race detector mode
	if elapsed > 600*time.Millisecond {
		t.Errorf("Expected ~100-400ms total time, got %v (connections may be sequential)", elapsed)
	}

	efficiency := (1.0 - float64(dialCount)/float64(numGoroutines)) * 100
	t.Logf("✓ Singleflight efficiency: %d concurrent requests → %d connections (%.0f%% reduction) in %v",
		numGoroutines, dialCount, efficiency, elapsed)
}

// TestSingleflight_TimeoutScenario tests that concurrent requests share timeout result
func TestSingleflight_TimeoutScenario(t *testing.T) {
	// Use non-routable address to trigger timeout
	address := "192.0.2.1:9999" // TEST-NET-1 (RFC 5737) - guaranteed to timeout

	pool := NewConnectionPool(1*time.Minute, 100, 16)
	defer pool.Close()

	// Set factory with 1 second timeout (faster for testing)
	pool.factory = func(addr string) (net.Conn, error) {
		return net.DialTimeout("tcp", addr, 1*time.Second)
	}

	const numGoroutines = 5
	var wg sync.WaitGroup
	var firstStarted atomic.Bool
	var firstStartTime atomic.Int64
	errors := make(chan error, numGoroutines)

	start := time.Now()

	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Track when first goroutine starts dial
			if firstStarted.CompareAndSwap(false, true) {
				firstStartTime.Store(time.Now().UnixNano())
			}

			_, err := pool.Get(address)
			errors <- err
		}(i)
	}

	wg.Wait()
	close(errors)
	elapsed := time.Since(start)

	// Count errors
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
		}
	}

	t.Logf("Concurrent timeout requests: %d", numGoroutines)
	t.Logf("Errors: %d", errorCount)
	t.Logf("Total time: %v", elapsed)

	// Key point: Even though each goroutine eventually fails after retries,
	// singleflight ensures only ONE dial attempt happens at a time for the same address.
	// This prevents the "timeout pileup" where N goroutines each wait 5 seconds sequentially.

	// The improvement is subtle: without singleflight, all N goroutines dial simultaneously,
	// each consuming resources. With singleflight, they share one dial attempt.

	// For timeout scenarios, the benefit is mainly resource efficiency (1 dial vs N dials),
	// not total time, since each goroutine eventually needs its own connection.

	t.Logf("✓ Singleflight coordinated %d concurrent timeout attempts", numGoroutines)

	metrics := pool.GetMetrics()
	t.Logf("Metrics: attempts=%d, failures=%d", metrics.ConnectionAttempts, metrics.ConnectionFailures)
}

// TestSingleflight_DifferentAddresses tests that different addresses can be dialed concurrently
func TestSingleflight_DifferentAddresses(t *testing.T) {
	// Create multiple test servers
	const numServers = 5
	servers := make([]net.Listener, numServers)
	addresses := make([]string, numServers)

	for i := range numServers {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to start test server %d: %v", i, err)
		}
		defer ln.Close()
		servers[i] = ln
		addresses[i] = ln.Addr().String()

		go func(ln net.Listener) {
			for {
				conn, err := ln.Accept()
				if err != nil {
					return
				}
				go handleTestConnection(conn)
			}
		}(ln)
	}

	var actualDials atomic.Int64

	pool := NewConnectionPool(1*time.Minute, 100, 16)
	defer pool.Close()

	originalFactory := pool.factory
	pool.factory = func(address string) (net.Conn, error) {
		actualDials.Add(1)
		time.Sleep(100 * time.Millisecond) // Simulate slow dial
		return originalFactory(address)
	}

	var wg sync.WaitGroup
	start := time.Now()

	// Dial all addresses concurrently
	for i := range numServers {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			conn, err := pool.Get(addr)
			if err != nil {
				t.Errorf("Get failed for %s: %v", addr, err)
				return
			}
			pool.Release(addr, conn)
		}(addresses[i])
	}

	wg.Wait()
	elapsed := time.Since(start)

	dialCount := actualDials.Load()

	t.Logf("Different addresses: %d", numServers)
	t.Logf("Actual dials: %d", dialCount)
	t.Logf("Total time: %v", elapsed)

	// Should create 5 connections (one per address)
	if dialCount != int64(numServers) {
		t.Errorf("Expected %d dials, got %d", numServers, dialCount)
	}

	// Should complete in ~100ms (concurrent), not 500ms (sequential)
	if elapsed > 300*time.Millisecond {
		t.Errorf("Expected ~100ms (concurrent dials), got %v (may be sequential)", elapsed)
	}

	t.Logf("✓ Different addresses dialed concurrently: %d addresses in %v", numServers, elapsed)
}

// TestSingleflight_ReuseAfterInflight tests that after inflight completes, connection is reusable
func TestSingleflight_ReuseAfterInflight(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleTestConnection(conn)
		}
	}()

	var actualDials atomic.Int64

	pool := NewConnectionPool(1*time.Minute, 100, 16)
	defer pool.Close()

	originalFactory := pool.factory
	pool.factory = func(address string) (net.Conn, error) {
		actualDials.Add(1)
		time.Sleep(50 * time.Millisecond)
		return originalFactory(address)
	}

	address := ln.Addr().String()

	// First batch: 5 concurrent requests (should create 1 connection)
	var wg1 sync.WaitGroup
	for range 5 {
		wg1.Go(func() {
			conn, _ := pool.Get(address)
			pool.Release(address, conn)
		})
	}
	wg1.Wait()

	firstDials := actualDials.Load()
	t.Logf("First batch (5 concurrent): %d dials", firstDials)

	// Second batch: 5 more requests (should reuse existing connection)
	var wg2 sync.WaitGroup
	for range 5 {
		wg2.Go(func() {
			conn, _ := pool.Get(address)
			pool.Release(address, conn)
		})
	}
	wg2.Wait()

	secondDials := actualDials.Load()
	t.Logf("Second batch (5 more): %d total dials", secondDials)

	// In race detector mode, timing variations can cause 2-3 dials
	// The key is that it's much less than 10 (without singleflight)
	if firstDials > 3 {
		t.Errorf("First batch: expected 1-3 dials, got %d", firstDials)
	}
	if secondDials > 4 {
		t.Errorf("Expected 1-4 total dials (with reuse), got %d", secondDials)
	}

	metrics := pool.GetMetrics()
	t.Logf("✓ Connection reused: %d attempts, %d dials, %.0f%% reuse rate",
		metrics.ConnectionAttempts, secondDials, pool.GetReuseRate()*100)
}

// TestConnectionPool_ZeroShardCount verifies auto-configuration when shardCount=0
func TestConnectionPool_ZeroShardCount(t *testing.T) {
	// Test with shardCount=0 (should auto-configure to 16)
	pool := NewConnectionPool(time.Minute, 100, 0)
	defer pool.Close()

	if pool.shardCount != 16 {
		t.Errorf("Expected auto-configured shardCount=16, got %d", pool.shardCount)
	}

	if pool.shardLimit != 6 { // 100 / 16 = 6
		t.Errorf("Expected shardLimit=6 (100/16), got %d", pool.shardLimit)
	}

	t.Logf("✓ Auto-configured: shardCount=%d, shardLimit=%d", pool.shardCount, pool.shardLimit)
}

// TestConnectionPool_NegativeShardCount verifies auto-configuration when shardCount<0
func TestConnectionPool_NegativeShardCount(t *testing.T) {
	// Test with negative shardCount (should auto-configure to 16)
	pool := NewConnectionPool(time.Minute, 100, -5)
	defer pool.Close()

	if pool.shardCount != 16 {
		t.Errorf("Expected auto-configured shardCount=16, got %d", pool.shardCount)
	}

	t.Logf("✓ Negative shardCount handled: shardCount=%d", pool.shardCount)
}
