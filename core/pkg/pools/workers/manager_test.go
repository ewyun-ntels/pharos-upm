package workers_test

import (
	"testing"
	"time"

	"ntels.com/pharos/core/pkg/pools/workers"
)

func TestManager_CreatePool(t *testing.T) {
	manager := workers.NewManager()
	defer manager.Shutdown()

	opts := workers.WorkerPoolOptions{
		WorkerCount:  10,
		JobQueueSize: 100,
	}

	err := manager.CreatePool("test-pool", opts)
	if err != nil {
		t.Errorf("Failed to create pool: %v", err)
	}

	// 같은 이름으로 다시 생성 시도
	err = manager.CreatePool("test-pool", opts)
	if err == nil {
		t.Error("Expected error when creating pool with duplicate name")
	}
}

func TestManager_GetPool(t *testing.T) {
	manager := workers.NewManager()
	defer manager.Shutdown()

	opts := workers.DefaultOptions()
	manager.CreatePool("existing-pool", opts)

	// 존재하는 풀 가져오기
	pool, err := manager.GetPool("existing-pool")
	if err != nil {
		t.Errorf("Failed to get existing pool: %v", err)
	}
	if pool == nil {
		t.Error("Got nil pool")
	}

	// 존재하지 않는 풀 가져오기
	_, err = manager.GetPool("non-existent-pool")
	if err == nil {
		t.Error("Expected error when getting non-existent pool")
	}
}

func TestManager_StartAll_StopAll(t *testing.T) {
	manager := workers.NewManager()

	// 여러 풀 생성
	for i := range 3 {
		manager.CreatePool(
			string(rune('a'+i)),
			workers.DefaultOptions(),
		)
	}

	// 모두 시작
	manager.StartAll()

	// 작업 제출 테스트
	pool, _ := manager.GetPool("a")
	pool.RegisterHandler(workers.NewSampleHandler1())

	err := pool.Submit(workers.Job{
		ID:      "test-1",
		Type:    workers.JobTypeSampleTask1,
		Payload: nil,
	})
	if err != nil {
		t.Errorf("Failed to submit job: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// 모두 정지 (Shutdown 대신 StopAll 사용)
	manager.StopAll()
}

func TestManager_GetAllMetrics(t *testing.T) {
	manager := workers.NewManager()
	defer manager.Shutdown()

	// 풀 생성
	manager.CreatePool("metrics-test", workers.DefaultOptions())
	manager.StartAll()

	pool, _ := manager.GetPool("metrics-test")
	pool.RegisterHandler(workers.NewSampleHandler1())

	// 작업 제출
	for i := range 10 {
		pool.Submit(workers.Job{
			ID:      string(rune('0' + i)),
			Type:    workers.JobTypeSampleTask1,
			Payload: nil,
		})
	}

	time.Sleep(200 * time.Millisecond)

	// 메트릭 확인
	allMetrics := manager.GetAllMetrics()
	if len(allMetrics) != 1 {
		t.Errorf("Expected 1 pool metrics, got %d", len(allMetrics))
	}

	metrics, exists := allMetrics["metrics-test"]
	if !exists {
		t.Error("Expected metrics for 'metrics-test' pool")
	}

	if metrics.TotalJobs != 10 {
		t.Errorf("Expected 10 total jobs, got %d", metrics.TotalJobs)
	}
}

func TestManager_Shutdown(t *testing.T) {
	manager := workers.NewManager()

	// 풀 생성 및 시작
	manager.CreatePool("shutdown-test", workers.DefaultOptions())
	pool, _ := manager.GetPool("shutdown-test")
	pool.RegisterHandler(workers.NewSampleHandler1())
	manager.StartAll()

	// 작업 제출 (Shutdown 전)
	err := pool.Submit(workers.Job{
		ID:      "before-shutdown",
		Type:    workers.JobTypeSampleTask1,
		Payload: nil,
	})
	if err != nil {
		t.Errorf("Should be able to submit before shutdown: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Shutdown 호출
	manager.Shutdown()

	// Shutdown 완료 확인 (메트릭 검증)
	metrics := pool.GetMetrics()
	if metrics.ProcessedJobs == 0 {
		t.Error("Expected at least one processed job before shutdown")
	}
}
