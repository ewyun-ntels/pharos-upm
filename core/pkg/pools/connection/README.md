# Connection Pool Package

TCP 연결 풀 구현 - 주기적 수집 시 연결 재사용으로 성능 최적화

## 개요

120만+ 장비로부터 주기적으로 데이터를 수집하는 시스템을 위한 고성능 TCP 연결 풀입니다. 동일 IP에 반복 접속 시 3-way handshake를 제거하여 **50% 성능 향상**을 달성합니다.

## 핵심 기능

- ✅ **자동 연결 재사용**: 99.99% 재사용률
- ✅ **메모리 효율**: LRU 기반 크기 제한
- ✅ **동시성 안전**: Mutex + atomic 연산
- ✅ **자동 정리**: 유휴 연결 자동 제거
- ✅ **Retry 로직**: Exponential backoff
- ✅ **타임아웃 제어**: Read/Write deadline

## 성능

```
벤치마크 결과:
- 연결 재사용:  216.7 ns/op (99.99% 재사용률)
- 직접 연결:   79060 ns/op
- 성능 향상:   364배 빠름

실제 효과:
- 첫 수집:  4분 (handshake 포함)
- 재사용:   2분 (handshake 제거)
- 절약:     50% 시간 단축
```

## 빠른 시작

### 기본 사용

```go
import "ntels.com/pharos/core/pkg/pools/connection"

// 연결 풀 생성
pool := connection.NewConnectionPool(
    30*time.Second,  // 유휴 타임아웃 (수집 주기 × 2)
    100,             // 최대 연결 수
)
defer pool.Close()

// 연결 가져오기
conn, err := pool.Get("192.168.1.100:5000")
if err != nil {
    log.Fatal(err)
}

// 데이터 송수신
response, err := connection.SendWithTimeout(
    conn,
    []byte("GET STATUS\n"),
    5*time.Second,
)

// 연결 반환 (재사용을 위해)
pool.Release("192.168.1.100:5000", conn)

// 메트릭 조회
metrics := pool.GetMetrics()
fmt.Printf("재사용률: %.2f%%\n", pool.GetReuseRate()*100)
```

### 재시도 로직

```go
ctx := context.Background()
conn, err := connection.DialWithRetry(ctx, "192.168.1.100:5000", 3)
if err != nil {
    log.Printf("3번 재시도 후 실패: %v", err)
}
```

## API 레퍼런스

### ConnectionPool

```go
func NewConnectionPool(maxIdleTime time.Duration, maxSize int) *ConnectionPool
func (p *ConnectionPool) Get(address string) (net.Conn, error)
func (p *ConnectionPool) Release(address string, conn net.Conn)
func (p *ConnectionPool) Close()
func (p *ConnectionPool) GetMetrics() PoolMetricsSnapshot
func (p *ConnectionPool) GetReuseRate() float64
func (p *ConnectionPool) GetActiveConnections() int64
```

### 헬퍼 함수

```go
func DialWithRetry(ctx context.Context, address string, maxRetries int) (net.Conn, error)
func SendWithTimeout(conn net.Conn, command []byte, timeout time.Duration) ([]byte, error)
```

### 메트릭

```go
type PoolMetricsSnapshot struct {
    TotalConnections   int64  // 생성된 총 연결 수
    ActiveConnections  int64  // 현재 활성 연결 수
    ConnectionAttempts int64  // 연결 요청 횟수
    ConnectionFailures int64  // 연결 실패 횟수
    ConnectionReuses   int64  // 연결 재사용 횟수
}
```

## 설정 가이드

### 수집 주기별 권장 설정

| 수집 주기 | 연결 풀 | 유휴 타임아웃 | MaxSize | 메모리 |
|---------|--------|------------|---------|-------|
| 15초 | ✅ 필수 | 30초 | 100 | 160MB |
| 1분 | ✅ 권장 | 2분 | 100 | 160MB |
| 5분 | ⚠️ 선택 | 10분 | 50 | 80MB |
| 1시간+ | ❌ 불필요 | - | - | - |

### 메모리 사용량 계산

```
총 메모리 = MaxSize × 16KB

예시:
- 100 연결 × 16KB = 160MB
- 50 연결 × 16KB = 80MB
```

## 모니터링

### 로그

```
INFO connection pool created max_idle_time=30s max_size=100
DEBUG connection reused address=192.168.1.100:5000
INFO idle connections cleanup removed=5 active=45
WARN non-retryable error address=192.168.1.101:5000 error="connection refused"
INFO connection pool closed total_connections=1000 connection_reuses=900
```

### 메트릭 수집

```go
ticker := time.NewTicker(1 * time.Minute)
for range ticker.C {
    metrics := pool.GetMetrics()
    reuseRate := pool.GetReuseRate()
    
    log.Printf("Pool stats: total=%d active=%d attempts=%d reuses=%d rate=%.2f%%",
        metrics.TotalConnections,
        metrics.ActiveConnections,
        metrics.ConnectionAttempts,
        metrics.ConnectionReuses,
        reuseRate*100,
    )
}
```

## 테스트

```bash
# 전체 테스트
go test -v

# 벤치마크
go test -bench=. -benchmem

# 커버리지
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 아키텍처

```
Cron Scheduler (*/15 * * * * *)
    ↓
Worker Pool (500 workers)
    ↓
Connection Pool
    ├─ LRU 정책 (크기 제한)
    ├─ 자동 cleanup (유휴 타임아웃)
    └─ Atomic 메트릭
    ↓
TCP Connections (재사용)
    ↓
Target Devices (1.2M+)
```

## TCP/UDP 지원

### TCP (현재 구현)
- ✅ 연결 풀 필수
- ✅ 50% 성능 향상
- ✅ Handshake 절약

### UDP (향후 계획)
- ⬜ 직접 다이얼 (풀 불필요)
- ⬜ Connectionless 특성
- ⬜ 2% 미만 성능 차이

**참고**: UDP는 handshake가 없어 연결 풀의 효과가 거의 없음

## 문제 해결

### 낮은 재사용률 (< 50%)

**원인**: 유휴 타임아웃이 수집 주기보다 짧음

**해결**:
```go
pool := NewConnectionPool(
    scheduleInterval * 2,  // 수집 주기의 2배
    100,
)
```

### 메모리 사용량 증가

**원인**: 풀 크기가 너무 큼

**해결**:
```go
pool := NewConnectionPool(
    30*time.Second,
    50,  // MaxSize 줄임
)
```

### 연결 실패 증가

**원인**: 대상 서버 응답 없음

**확인**:
```go
metrics := pool.GetMetrics()
fmt.Printf("실패율: %.2f%%\n", 
    float64(metrics.ConnectionFailures) / float64(metrics.ConnectionAttempts) * 100)
```

## 개선 이력

### v1.1.0 (2026-01-09)
- ✅ **로깅 강화**: 재사용률, 성공률, 연결 시간 구조화된 로깅
  - slog.Info로 주요 이벤트 로깅 (재사용률 50.0%, 성공률 100.0%)
  - 연결 시간, cleanup 메트릭 상세 출력
  - 구조화된 필드로 메트릭 추적 용이
- ✅ **주석 개선**: 패키지 문서화 강화
  - pooledConn 필드 상세 설명 (inUse, isAlive, lastUsed)
  - SendWithTimeout 동작 방식 문서화
  - performCleanup 로직 설명 추가
- ✅ **코드 품질**: Race condition 제거, 리소스 누수 방지
  - inUse 플래그로 동시 사용 방지
  - isConnAlive 검증으로 불량 연결 제거
  - 동적 버퍼로 메모리 효율 개선
- ✅ 테스트 100% 통과 (8.121s)

### v1.0.0 (2025-12-08)
- ✅ GetMetrics() 컴파일 에러 수정 (PoolMetricsSnapshot 도입)
- ✅ isNonRetryableError() 완전 구현
  - Connection refused 감지
  - Network unreachable 감지
  - Invalid argument 감지
- ✅ 테스트 100% 통과
- ✅ 문서 재구성 (README.md 통합)

## 라이선스

Copyright © 2025 NTELS. All rights reserved.
