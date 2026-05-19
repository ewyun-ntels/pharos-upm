package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/pkg/common"
)

func TestLeaderChecker_StartStop(t *testing.T) {
	t.Parallel()

	// LeaderChecker의 시작과 정지 테스트
	checker := &LeaderChecker{}
	server := &NatsServer{
		config: common.Config{},
	}

	// 시작
	checker.Start(server)

	// 실행 중인지 확인
	isRunning := !checker.stop.Load()
	assert.True(t, isRunning, "LeaderChecker should be running after Start")

	// 짧은 대기
	time.Sleep(10 * time.Millisecond)

	// 정지
	checker.Stop()

	// 정지되었는지 확인
	isStopped := checker.stop.Load()
	assert.True(t, isStopped, "LeaderChecker should be stopped after Stop")
}

func TestLeaderChecker_MultipleStartStop(t *testing.T) {
	t.Parallel()

	// 여러 번 시작/정지 테스트
	checker := &LeaderChecker{}
	server := &NatsServer{
		config: common.Config{},
	}

	for range 3 {
		checker.Start(server)
		time.Sleep(5 * time.Millisecond)
		checker.Stop()
	}

	// 마지막 정지 후 상태 확인
	isStopped := checker.stop.Load()
	assert.True(t, isStopped, "LeaderChecker should be stopped after multiple cycles")
}

func TestLeaderChecker_StopWithoutStart(t *testing.T) {
	t.Parallel()

	// 시작하지 않고 정지하는 경우
	checker := &LeaderChecker{}

	// panic이 발생하지 않아야 함
	assert.NotPanics(t, func() {
		checker.Stop()
	}, "Stop without Start should not panic")
}

func TestLeaderChecker_IsLeader_NilServer(t *testing.T) {
	t.Parallel()

	// natsServer가 nil인 경우
	checker := &LeaderChecker{}
	server := &NatsServer{
		natsServer: nil,
	}

	isLeader := checker.isLeader(server)
	assert.False(t, isLeader, "Should return false when natsServer is nil")
}

func TestLeaderChecker_IsLeader_JetStreamDisabled(t *testing.T) {
	t.Parallel()

	// Skip this test as it requires a real NATS server
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// JetStream이 비활성화된 경우는 실제 서버가 필요하므로 스킵
	// 실제 구현에서는 JetStream이 비활성화되어 있으면 true를 반환
	t.Skip("Requires real NATS server with JetStream disabled")
}

func TestLeaderChecker_ConcurrentStartStop(t *testing.T) {
	t.Parallel()

	// 동시 시작/정지 테스트
	checker := &LeaderChecker{}
	server := &NatsServer{
		config: common.Config{},
	}

	// 시작
	checker.Start(server)
	time.Sleep(5 * time.Millisecond)

	// 동시에 여러 번 Stop 호출해도 안전해야 함
	done := make(chan bool, 3)
	for range 3 {
		go func() {
			checker.Stop()
			done <- true
		}()
	}

	// 모든 Stop이 완료될 때까지 대기
	for range 3 {
		<-done
	}

	isStopped := checker.stop.Load()
	assert.True(t, isStopped, "LeaderChecker should be stopped after concurrent stops")
}

func TestLeaderChecker_QuickStartStop(t *testing.T) {
	// 빠르게 시작하고 바로 정지하는 테스트
	checker := &LeaderChecker{}
	server := &NatsServer{
		config: common.Config{},
	}

	// 빠르게 시작/정지 반복
	for range 10 {
		checker.Start(server)
		checker.Stop()
	}

	// 최종적으로 정지 상태여야 함
	isStopped := checker.stop.Load()
	assert.True(t, isStopped, "LeaderChecker should be stopped after quick cycles")
}

func TestLeaderChecker_StateTransitions(t *testing.T) {
	// 상태 전환 테스트
	checker := &LeaderChecker{}
	server := &NatsServer{
		config: common.Config{},
	}

	// 초기 상태 (정지)
	initialState := checker.stop.Load()
	assert.False(t, initialState, "Initial state should be stopped (false)")

	// 시작
	checker.Start(server)
	runningState := checker.stop.Load()
	assert.False(t, runningState, "After Start, stop flag should be false")

	// 정지
	checker.Stop()
	stoppedState := checker.stop.Load()
	assert.True(t, stoppedState, "After Stop, stop flag should be true")
}

func TestLeaderChecker_RestartAfterStop(t *testing.T) {
	// 정지 후 재시작 테스트
	checker := &LeaderChecker{}
	server := &NatsServer{
		config: common.Config{},
	}

	// 첫 번째 사이클
	checker.Start(server)
	time.Sleep(20 * time.Millisecond)
	checker.Stop()

	// 완전히 정지될 때까지 대기
	time.Sleep(20 * time.Millisecond)

	// 두 번째 사이클
	checker.Start(server)
	isRunning := !checker.stop.Load()
	assert.True(t, isRunning, "LeaderChecker should restart successfully")

	checker.Stop()
	isStopped := checker.stop.Load()
	assert.True(t, isStopped, "LeaderChecker should stop successfully after restart")
}

func TestLeaderChecker_LongRunning(t *testing.T) {
	// 장시간 실행 테스트
	if testing.Short() {
		t.Skip("Skipping long-running test in short mode")
	}

	checker := &LeaderChecker{}
	server := &NatsServer{
		config: common.Config{},
	}

	checker.Start(server)

	// 1초 동안 실행
	time.Sleep(1 * time.Second)

	checker.Stop()

	isStopped := checker.stop.Load()
	assert.True(t, isStopped, "LeaderChecker should stop after long run")
}
