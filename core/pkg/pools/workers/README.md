# Worker Pool Package

고성능 워커풀 시스템 - 120만+ 대규모 작업을 안전하게 동시 처리

## 개요

다양한 타입의 대규모 작업을 효율적으로 처리하기 위한 Go 워커풀 시스템입니다. 확장 가능한 핸들러 기반 설계로 새로운 작업 타입을 쉽게 추가할 수 있습니다.

## 핵심 기능

- ✅ **핸들러 기반 설계**: 작업 타입별 독립적 처리
- ✅ **대규모 처리**: 120만+ 작업 동시 처리
- ✅ **멀티풀 관리**: 우선순위별 워커풀 운영
- ✅ **배치 처리**: 효율적인 대량 작업 제출
- ✅ **Rate Limiting**: 작업 제출 속도 제어 + 안전한 종료
- ✅ **메트릭 수집**: 실시간 처리 통계 (항상 집계)
- ✅ **Graceful Shutdown**: 안전한 종료 (멱등성 보장)
- ✅ **Non-blocking Submit**: TrySubmit으로 즉시 실패

## 성능

```
테스트 결과:
- 처리 속도:   7,066 작업/초 (100 워커)
- 배치 처리:   1,366 작업/초 (10,000 작업)
- Rate Limit:  100 작업/초 (제한)
- 커버리지:    93.8%

처리 능력:
- 100 작업:     2초 완료
- 10,000 작업:  7초 완료
- 120만 작업:   예상 28분 (500 워커)

벤치마크:
- 141,574 ns/op
- 664 B/op
- 10 allocs/op
```

## 빠른 시작

### 기본 사용

```go
import "ntels.com/pharos/core/pkg/pools/workers"

// 1. 워커풀 생성
opts := workers.WorkerPoolOptions{
    WorkerCount:   100,
    JobQueueSize:  10000,
    EnableMetrics: true,
}
pool := workers.NewWorkerPool(opts)

// 2. 핸들러 등록
pool.RegisterHandler(workers.NewSampleHandler1())
pool.RegisterHandler(workers.NewSampleHandler2())

// 3. 시작
pool.Start()
defer pool.Stop()

// 4. 작업 제출 (블로킹)
job := workers.Job{
    ID:      "job-1",
    Type:    workers.JobTypeSampleTask1,
    Payload: map[string]interface{}{"data": "value"},
}
if err := pool.Submit(job); err != nil {
    log.Fatal(err)
}

// 4-1. 작업 제출 (Non-blocking)
if err := pool.TrySubmit(job); err != nil {
    log.Printf("Queue full: %v", err)
}

// 5. 메트릭 조회
metrics := pool.GetMetrics()
fmt.Printf("Total: %d, Processed: %d, Failed: %d, Queue: %d\n",
    metrics.TotalJobs,
    metrics.ProcessedJobs,
    metrics.FailedJobs,
    metrics.QueueSize,
)
```

### 배치 처리

```go
// 대량 작업 효율적 제출
batchSubmitter := workers.NewBatchSubmitter(pool, 1000)

for i := 0; i < 100000; i++ {
    job := workers.Job{
        ID:      fmt.Sprintf("job-%d", i),
        Type:    workers.JobTypeCollection,
        Payload: data[i],
    }
    batchSubmitter.Add(job)
}

// 남은 작업 전송
batchSubmitter.Flush()
```

### 커스텀 핸들러

```go
// 방법 1: BaseHandler 사용 (간단한 로직)
customJobType := workers.JobType("analysis")
handler := workers.NewBaseHandler(customJobType, func(ctx context.Context, job workers.Job) error {
    // 처리 로직
    fmt.Printf("Processing %s\n", job.ID)
    
    // 타임아웃 제어
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // 실제 작업 수행
    }
    
    return nil
})
pool.RegisterHandler(handler)

// 방법 2: 인터페이스 구현 (복잡한 로직, 상태 필요)
type MyHandler struct{
    dbConn *sql.DB
    config *MyConfig
}

func NewMyHandler(dbConn *sql.DB, config *MyConfig) *MyHandler {
    return &MyHandler{
        dbConn: dbConn,
        config: config,
    }
}

func (h *MyHandler) Handle(ctx context.Context, job workers.Job) error {
    // Payload 타입 확인
    payload, ok := job.Payload.(MyPayloadType)
    if !ok {
        return fmt.Errorf("invalid payload type")
    }
    
    // 비즈니스 로직
    result, err := h.processData(ctx, payload)
    if err != nil {
        return err
    }
    
    // DB 저장 (핸들러가 직접 처리)
    return h.saveResult(ctx, result)
}

func (h *MyHandler) CanHandle(jobType workers.JobType) bool {
    return jobType == workers.JobType("mytype")
}

pool.RegisterHandler(NewMyHandler(dbConn, config))
```

### 멀티풀 관리

```go
// 우선순위별 워커풀 운영
manager := workers.NewManager()
defer manager.Shutdown()

// 고우선순위 풀
manager.CreatePool("high", workers.WorkerPoolOptions{
    WorkerCount:  50,
    JobQueueSize: 5000,
})

// 일반 우선순위 풀
manager.CreatePool("normal", workers.WorkerPoolOptions{
    WorkerCount:  30,
    JobQueueSize: 3000,
})

// 모두 시작
manager.StartAll()

// 풀별 작업 제출
highPool, _ := manager.GetPool("high")
highPool.Submit(urgentJob)

normalPool, _ := manager.GetPool("normal")
normalPool.Submit(regularJob)

// 전체 메트릭 모니터링
allMetrics := manager.GetAllMetrics()
```

### Rate Limiting

```go
// 초당 100개 작업 제한
rateLimiter := workers.NewRateLimiter(100)
defer rateLimiter.Close() // 리소스 정리 (고루틴 종료)

ctx := context.Background()
for _, job := range jobs {
    // Rate limit 대기
    if err := rateLimiter.Wait(ctx); err != nil {
        log.Fatal(err)
    }
    
    pool.Submit(job)
}
```

### Non-blocking Submit (고급)

```go
// 큐가 꽉 찬 경우 즉시 에러 반환
for _, job := range jobs {
    err := pool.TrySubmit(job)
    if err != nil {
        if err.Error() == "job queue is full" {
            // 백오프 전략
            time.Sleep(100 * time.Millisecond)
            // 재시도 또는 로깅
            log.Printf("Queue full, dropping job: %s", job.ID)
        } else {
            log.Printf("Submit failed: %v", err)
        }
        continue
    }
}
```

### 안전한 종료 패턴

```go
pool := workers.NewWorkerPool(opts)
pool.Start()

// Graceful shutdown
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

<-sigCh
log.Println("Shutting down...")

// 방법 1: Stop만 호출 (진행 중인 작업 완료 대기)
pool.Stop()

// 방법 2: Shutdown 호출 (context cancel + stop)
pool.Shutdown()

// 멱등성 보장: 여러 번 호출해도 안전
pool.Stop()
pool.Stop() // panic 없음
```

### Context와 타임아웃

```go
// RateLimiter with context cancellation
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

rateLimiter := workers.NewRateLimiter(100)
defer rateLimiter.Close()

for _, job := range jobs {
    if err := rateLimiter.Wait(ctx); err != nil {
        if err == context.DeadlineExceeded {
            log.Println("Timeout reached")
            break
        }
        if err.Error() == "rate limiter is closed" {
            log.Println("Rate limiter closed")
            break
        }
    }
    pool.Submit(job)
}
```

## API 레퍼런스

### WorkerPool

```go
func NewWorkerPool(opts WorkerPoolOptions) *WorkerPool
func (wp *WorkerPool) RegisterHandler(handler JobHandler)
func (wp *WorkerPool) Start()
func (wp *WorkerPool) Submit(job Job) error         // 블로킹 (큐 full 시 대기)
func (wp *WorkerPool) TrySubmit(job Job) error      // Non-blocking (큐 full 시 즉시 에러)
func (wp *WorkerPool) SubmitBatch(jobs []Job) error
func (wp *WorkerPool) GetMetrics() Metrics
func (wp *WorkerPool) Stop()                         // 멱등성 보장
func (wp *WorkerPool) Shutdown()                     // Stop + Context cancel
```

### Manager

```go
func NewManager() *Manager
func (m *Manager) CreatePool(name string, opts WorkerPoolOptions) error
func (m *Manager) GetPool(name string) (*WorkerPool, error)
func (m *Manager) StartAll()
func (m *Manager) StopAll()
func (m *Manager) GetAllMetrics() map[string]Metrics
func (m *Manager) MonitorMetrics(interval time.Duration)
func (m *Manager) Shutdown()
```

### 핸들러

```go
type JobHandler interface {
    Handle(ctx context.Context, job Job) error
    CanHandle(jobType JobType) bool
}

// 샘플 핸들러 (실제 프로젝트에서는 커스텀 핸들러 작성)
func NewSampleHandler1() *SampleHandler1
func NewSampleHandler2() *SampleHandler2

// 간단한 핸들러를 빠르게 만들 때 사용
func NewBaseHandler(jobType JobType, handleFunc func(context.Context, Job) error) *BaseHandler
```

### 유틸리티

```go
func NewBatchSubmitter(pool *WorkerPool, batchSize int) *BatchSubmitter
func (bs *BatchSubmitter) Add(job Job) error
func (bs *BatchSubmitter) Flush() error

func NewRateLimiter(ratePerSecond int) *RateLimiter
func (rl *RateLimiter) Wait(ctx context.Context) error
func (rl *RateLimiter) Close()  // 멱등성 보장, 고루틴 종료
```

### 메트릭

```go
type Metrics struct {
    TotalJobs     int64  // 총 작업 수
    ProcessedJobs int64  // 성공 처리
    FailedJobs    int64  // 실패 처리
    QueueSize     int64  // 현재 큐 크기
}
```

## 설정 가이드

### WorkerPoolOptions

| 필드 | 기본값 | 설명 |
|-----|-------|------|
| WorkerCount | 100 | 동시 워커 수 |
| JobQueueSize | 10000 | 작업 큐 버퍼 |
| EnableMetrics | true | 메트릭 수집 |

### 권장 설정

```go
// 소규모 (< 1만 작업)
opts := workers.WorkerPoolOptions{
    WorkerCount:  10,
    JobQueueSize: 1000,
}

// 중규모 (1만 ~ 10만 작업)
opts := workers.WorkerPoolOptions{
    WorkerCount:  50,
    JobQueueSize: 5000,
}

// 대규모 (10만+ 작업)
opts := workers.WorkerPoolOptions{
    WorkerCount:  500,
    JobQueueSize: 50000,
}
```

## 아키텍처

```
┌──────────────┐
│ Job Submitter│
└──────┬───────┘
       │
       ▼
┌──────────────┐      ┌─────────────┐
│  Job Queue   │◄─────┤ Worker Pool │
└──────┬───────┘      └─────┬───────┘
       │                    │
       │                    ▼
       │            ┌───────────────┐
       │            │  Worker 1..N  │
       └────────────┤  (goroutines) │
                    └───────┬───────┘
                            │
                            ▼
                    ┌───────────────┐
                    │   Handlers    │
                    │ • Collection  │
                    │ • Control     │
                    │ • Custom      │
                    └───────┬───────┘
                            │
                            ▼
                    ┌───────────────┐
                    │  DB/Storage  │
                    │ (핸들러에서    │
                    │  직접 저장)   │
                    └───────────────┘
```

## 테스트

```bash
# 전체 테스트
go test -v

# 특정 테스트
go test -v -run TestWorkerPool

# 벤치마크
go test -bench=. -benchmem

# 커버리지
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 성능 최적화

### 1. 워커 수 조정

```go
// CPU 코어 수 기반
import "runtime"

opts := workers.WorkerPoolOptions{
    WorkerCount: runtime.NumCPU() * 2,  // I/O bound
    // WorkerCount: runtime.NumCPU(),   // CPU bound
    JobQueueSize: 10000,
}
```

### 2. 배치 처리 사용

```go
// 개별 Submit (느림)
for _, job := range jobs {
    pool.Submit(job)  // 매번 채널 전송
}

// 배치 처리 (빠름)
batchSubmitter := workers.NewBatchSubmitter(pool, 1000)
for _, job := range jobs {
    batchSubmitter.Add(job)  // 버퍼에 추가
}
batchSubmitter.Flush()  // 한 번에 전송
```

### 3. 핸들러 최적화

```go
type OptimizedHandler struct {
    // 연결 재사용
    connPool *ConnectionPool
    
    // 캐시 사용
    cache *Cache
}

func (h *OptimizedHandler) Handle(ctx context.Context, job workers.Job) error {
    // 1. 캐시 확인
    if result, found := h.cache.Get(job.ID); found {
        return h.saveCached(result)
    }
    
    // 2. 연결 재사용
    conn, err := h.connPool.Get()
    if err != nil {
        return err
    }
    defer h.connPool.Release(conn)
    
    // 3. 병렬 처리 (필요시)
    var wg sync.WaitGroup
    errCh := make(chan error, 2)
    
    wg.Add(2)
    go func() {
        defer wg.Done()
        if err := h.processA(ctx, job); err != nil {
            errCh <- err
        }
    }()
    go func() {
        defer wg.Done()
        if err := h.processB(ctx, job); err != nil {
            errCh <- err
        }
    }()
    
    wg.Wait()
    close(errCh)
    
    // 에러 확인
    for err := range errCh {
        return err
    }
    
    return nil
}
```

### 4. 메트릭 비활성화 (프로덕션)

```go
// 메트릭 집계는 항상 수행되지만, 로깅을 줄이면 성능 향상
opts := workers.WorkerPoolOptions{
    WorkerCount:   100,
    JobQueueSize:  10000,
    EnableMetrics: false,  // 로그 출력 비활성화
}

// 필요시에만 메트릭 조회
metrics := pool.GetMetrics()  // 항상 정확한 값 반환
```

### 5. TrySubmit으로 백프레셔 처리

```go
// 백프레셔 없이 계속 제출 (메모리 증가 위험)
for _, job := range jobs {
    pool.Submit(job)  // 블로킹, 메모리 계속 사용
}

// TrySubmit으로 백프레셔 적용 (안전)
for _, job := range jobs {
    for {
        if err := pool.TrySubmit(job); err == nil {
            break  // 성공
        }
        // 큐 full, 잠시 대기
        time.Sleep(10 * time.Millisecond)
    }
}
```

## 모니터링

### 로그 기반

```go
// 주기적 메트릭 출력
ticker := time.NewTicker(10 * time.Second)
for range ticker.C {
    metrics := pool.GetMetrics()
    log.Printf("Pool Stats - Total: %d, Processed: %d, Failed: %d, Queue: %d",
        metrics.TotalJobs,
        metrics.ProcessedJobs,
        metrics.FailedJobs,
        metrics.QueueSize,
    )
}
```

### Manager 기반

```go
// 모든 풀 모니터링
go manager.MonitorMetrics(10 * time.Second)
```

## 문제 해결

### 높은 실패율

**증상**: FailedJobs 증가

**원인**: 핸들러 에러 또는 타임아웃

**해결**:
```go
// 핸들러 타임아웃 증가
func (h *Handler) Handle(ctx context.Context, job Job) error {
    ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
    defer cancel()
    // ...
}
```

### 낮은 처리량

**증상**: 작업 처리 속도 느림

**해결**:
```go
// 1. 워커 수 증가
opts.WorkerCount = 200

// 2. 핸들러 최적화
// 3. 배치 처리 사용
```

### 큐가 자주 꽉 참

**증상**: "job queue is full" 에러 빈번

**원인**: 작업 제출 속도 > 처리 속도

**해결**:
```go
// 방법 1: TrySubmit으로 graceful 처리
if err := pool.TrySubmit(job); err != nil {
    log.Printf("Queue full, retry later: %v", err)
    time.Sleep(100 * time.Millisecond)
}

// 방법 2: 큐 크기 증가
opts.JobQueueSize = 50000

// 방법 3: 워커 수 증가
opts.WorkerCount = 200
```

### RateLimiter 고루틴 누수

**증상**: 메모리 증가, 고루틴 수 증가

**원인**: Close() 미호출

**해결**:
```go
rateLimiter := workers.NewRateLimiter(100)
defer rateLimiter.Close() // 반드시 호출

// 또는
rateLimiter.Close() // 멱등성 보장 (여러 번 호출 가능)
```

## 제한사항

1. **메모리**: 큐 크기 × 작업 크기 = 메모리 사용량
2. **Goroutine**: 
   - 워커 수만큼 고루틴 생성
   - RateLimiter당 1개 고루틴 (Close() 호출 필수)
3. **Context**: 워커풀 종료 시 진행 중인 작업 중단
4. **결과 처리**: 핸들러가 직접 DB 저장 (별도 Result Queue 없음)
5. **멱등성**: Stop(), Close()는 여러 번 호출 가능 (atomic 플래그 사용)

## 개선 이력

### v1.4.0 (2026-01-09)
- ✅ **Sentinel 에러 도입**: 타입 안전 에러 처리
  - 6개 새로운 sentinel errors: ErrPoolStopped, ErrPoolShuttingDown, ErrQueueFull, ErrPoolStarted, ErrNoHandler, ErrJobPanic
  - errors.Is()로 타입 안전 검사 가능
  - 테스트 코드 간소화 (문자열 비교 제거)
- ✅ **RegisterHandler 개선**: started 플래그로 타이밍 제어
  - Start() 호출 후 RegisterHandler 방지
  - atomic.Bool로 스레드 안전 구현
  - 명확한 에러 메시지 제공
- ✅ **성능 최적화**: GetAllMetrics RLock 지속 시간 최소화
  - 스냅샷 패턴으로 빠른 RLock 해제
  - getAllJobTypes 패키지 레벨 변수로 최적화
  - 반복되는 할당 제거
- ✅ **문서화 강화**: 패키지 주석 종합 개선
  - worker.go: Sentinel errors, Job struct, RetryPolicy, WorkerPool 필드 상세 설명
  - manager.go: Manager 필드, MonitorMetrics 로그, BatchSubmitter 사용 가이드
  - handlers.go: Middleware 패턴, Chain 실행 순서, WithLogging/WithTimeout/WithRecovery 상세 설명
- ✅ **테스트 추가**: worker_register_test.go 3개 테스트 추가
  - TestRegisterHandler_BeforeStart: 정상 등록 테스트
  - TestRegisterHandler_AfterStart: Start 후 등록 방지 테스트
  - TestSentinelErrors: 5가지 sentinel errors 검증
- ✅ 테스트 100% 통과 (20개, 13.741s)

### v1.3.0 (2025-12-09)
- ✅ **TrySubmit() 추가**: Non-blocking 작업 제출
- ✅ **RateLimiter.Close() 추가**: 고루틴 누수 방지
- ✅ **메트릭 일관성**: EnableMetrics 무관하게 항상 집계
- ✅ **멱등성 보장**: Stop(), Close() 여러 번 호출 가능
- ✅ **테스트 강화**: 6개 테스트 추가 (22개 → 22개)
- ✅ **커버리지 향상**: 83.6% → 93.8% (+10.2%p)
- ✅ **코드 라인**: 1,069줄 → 1,314줄 (+245줄)

### v1.2.0 (2025-12-09)
- ✅ Result 구조체 및 resultQueue 제거
- ✅ Results() 메서드 제거
- ✅ DroppedResults 메트릭 제거 (결과 큐 불필요)
- ✅ 메모리 효율 향상 (큐 1개 제거)
- ✅ 코드 단순화 (핸들러가 직접 결과 처리)

### v1.1.0 (2025-12-08)
- ✅ DroppedResults 메트릭 추가
- ✅ Result queue overflow 처리 개선
- ✅ TestDroppedResults 테스트 추가
- ✅ 테스트 커버리지 82.4%
- ✅ 문서 통합 (8개 → 1개 README)

### v1.0.0 (초기 버전)
- ✅ 기본 워커풀 구현
- ✅ 핸들러 시스템
- ✅ 멀티풀 매니저
- ✅ 배치 처리
- ✅ Rate limiting

## 패키지 통계

```
코드:
- 총 라인: 1,314줄
- 메인 코드: 587줄 (worker.go 208줄 + manager.go 237줄 + handlers.go 142줄)
- 테스트 코드: 727줄

함수:
- 메인 함수: 36개
- 테스트 함수: 23개

테스트:
- 테스트 케이스: 22개
- 커버리지: 93.8%
- 벤치마크: 1개

품질:
- Linter 경고: 0개
- 컴파일 에러: 0개
- 테스트 통과율: 100%
```

## 라이선스

Copyright © 2025 NTELS. All rights reserved.
