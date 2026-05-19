# Metrics Package

Prometheus 메트릭 수집 및 제공을 위한 코어 패키지입니다.

## 개요

이 패키지는 Prometheus 메트릭 수집기(collector)를 등록하고 HTTP 엔드포인트를 통해 메트릭을 제공하는 기능을 제공합니다.

### 주요 기능

- **Collector 등록 관리**: Prometheus collector를 동적으로 등록 및 관리
- **HTTP API**: Gin 프레임워크를 통한 메트릭 엔드포인트 제공
- **멀티 레지스트리**: 독립적인 Prometheus registry 관리
- **표준 메트릭**: Go runtime 및 프로세스 메트릭 자동 수집

## 아키텍처

```
pkg/metrics/
├── metrics.go      # Collector 등록 관리
├── api.go          # HTTP API 및 엔드포인트
├── metrics_test.go # Collector 관리 테스트
└── api_test.go     # API 통합 테스트
```

## 사용 방법

### 1. Collector 등록

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "ntels.com/phylax/core/pkg/metrics"
)

// Custom collector 생성
myCollector := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "my_custom_counter",
        Help: "A custom counter metric",
    },
    []string{"label1", "label2"},
)

// Collector 등록
metrics.AddCollector("my_collector", myCollector)
```

### 2. API 초기화 및 사용

```go
import (
    "github.com/gin-gonic/gin"
    "ntels.com/phylax/core/pkg/common"
    "ntels.com/phylax/core/pkg/metrics"
)

func main() {
    // API 인스턴스 생성
    api := &metrics.Api{}
    
    // 설정
    config := common.Config{
        Serve: common.ServeConfig{
            Type: common.ServeTypeMaster,
        },
        Metrics: common.MetricsConfig{
            Use: true,
        },
    }
    
    // 초기화
    api.Init("/path/to/config.yaml", config)
    
    // Collector 로드
    if err := api.Load(); err != nil {
        log.Fatal(err)
    }
    
    // Gin 라우터에 등록
    router := gin.Default()
    metricsGroup := router.Group("/metrics")
    api.RegisterRoutes(metricsGroup)
    
    // 서버 시작
    router.Run(":8080")
}
```

### 3. 메트릭 확인

```bash
# 메트릭 엔드포인트 접근
curl http://localhost:8080/metrics/

# 응답 예시
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 10
# HELP my_custom_counter A custom counter metric
# TYPE my_custom_counter counter
my_custom_counter{label1="value1",label2="value2"} 42
```

## 테스트

### 테스트 실행

```bash
# 전체 테스트 실행
go test ./pkg/metrics/...

# 상세 로그와 함께 실행
go test -v ./pkg/metrics/...

# 커버리지 측정
go test -cover ./pkg/metrics/...
go test -coverprofile=coverage.out ./pkg/metrics/...
go tool cover -html=coverage.out
```

### 벤치마크 실행

```bash
# 모든 벤치마크 실행
go test -bench=. -benchmem ./pkg/metrics/...

# 특정 벤치마크만 실행
go test -bench=BenchmarkAddCollector -benchmem ./pkg/metrics/...
```

### 테스트 커버리지

현재 테스트 커버리지: **100%**

| 파일 | 커버리지 | 테스트 수 |
|------|----------|-----------|
| metrics.go | 100% | 6개 |
| api.go | 100% | 12개 |

### 벤치마크 결과

```
BenchmarkApiLoad-6               10000    162939 ns/op    29281 B/op    268 allocs/op
BenchmarkApiRegisterRoutes-6    670414      1732 ns/op     1690 B/op     27 allocs/op
BenchmarkAddCollector-6       52652715      25.32 ns/op       0 B/op      0 allocs/op
BenchmarkGetCollector-6       56861209      19.09 ns/op       0 B/op      0 allocs/op
BenchmarkGetAllCollectors-6   87090147      19.92 ns/op       0 B/op      0 allocs/op
```

## 테스트 구조

### metrics_test.go

Collector 등록 및 관리 기능을 테스트합니다.

- `TestAddCollector`: 단일 collector 등록
- `TestAddCollector_Multiple`: 여러 collector 등록
- `TestAddCollector_Overwrite`: Collector 덮어쓰기
- `TestCollectors_GetAll`: 전체 collector 조회
- `TestCollectors_ThreadSafety`: 동시성 안전성 검증
- `BenchmarkAddCollector`: Collector 등록 성능 측정
- `BenchmarkGetCollector`: Collector 조회 성능 측정
- `BenchmarkGetAllCollectors`: 전체 조회 성능 측정

### api_test.go

HTTP API 및 Gin 통합 기능을 테스트합니다.

- `TestApiInit`: API 초기화
- `TestApiUse_Master`: Master 모드에서 메트릭 활성화
- `TestApiUse_NotMaster`: Worker 모드에서 메트릭 비활성화
- `TestApiUse_MetricsDisabled`: 메트릭 설정 비활성화
- `TestApiLoad`: Collector 로드 및 레지스트리 등록
- `TestApiUnload`: Collector 언로드
- `TestApiGetRelativePath`: 엔드포인트 경로 확인
- `TestApiRegisterRoutes`: Gin 라우터 등록
- `TestApiRegisterRoutes_FullPath`: 그룹 경로 포함 등록
- `TestApiIntegration`: 전체 통합 테스트
- `TestApiLoad_CreatesNewRegistry`: 레지스트리 재생성 확인
- `BenchmarkApiLoad`: API 로드 성능 측정
- `BenchmarkApiRegisterRoutes`: 라우터 등록 성능 측정

## API Reference

### metrics.go

#### AddCollector

```go
func AddCollector(name string, collector prometheus.Collector)
```

Prometheus collector를 등록합니다.

**Parameters:**
- `name`: Collector의 고유 이름
- `collector`: Prometheus Collector 인터페이스 구현체

**Example:**
```go
counter := prometheus.NewCounter(prometheus.CounterOpts{
    Name: "my_counter",
    Help: "My counter metric",
})
metrics.AddCollector("my_counter", counter)
```

### api.go

#### Api.Init

```go
func (a *Api) Init(configPath string, config common.Config)
```

API를 초기화하고 설정을 저장합니다.

#### Api.Use

```go
func (a *Api) Use() bool
```

메트릭 기능 활성화 여부를 반환합니다.

**Returns:**
- `true`: Master 모드이고 메트릭이 활성화된 경우
- `false`: 그 외의 경우

#### Api.Load

```go
func (a *Api) Load() error
```

등록된 모든 collector를 Prometheus registry에 로드합니다.

#### Api.Unload

```go
func (a *Api) Unload()
```

등록된 모든 collector를 레지스트리에서 제거합니다.

#### Api.RegisterRoutes

```go
func (a *Api) RegisterRoutes(routes gin.IRoutes)
```

Gin 라우터에 메트릭 엔드포인트를 등록합니다.

#### Api.GetRelativePath

```go
func (a *Api) GetRelativePath() string
```

메트릭 엔드포인트의 상대 경로를 반환합니다.

**Returns:** `"/metrics"`

## 설정

### Config 구조

```go
type Config struct {
    Serve common.ServeConfig
    Metrics common.MetricsConfig
}

type ServeConfig struct {
    Type string  // "master" 또는 "worker"
}

type MetricsConfig struct {
    Use bool  // 메트릭 활성화 여부
}
```

### 설정 예시

```yaml
serve:
  type: master

metrics:
  use: true
```

## 동시성 안전성

- `AddCollector`와 `Get` 작업은 내부적으로 mutex를 사용하여 스레드 안전성을 보장합니다.
- 여러 goroutine에서 동시에 collector를 등록하거나 조회할 수 있습니다.
- Prometheus registry 자체도 동시성을 지원합니다.

## 성능 최적화

### Collector 등록

- Collector 등록은 매우 빠릅니다 (~25ns/op, 0 allocs)
- 애플리케이션 시작 시 한 번만 등록하는 것을 권장합니다

### 레지스트리 관리

- `Load()` 호출 시마다 새로운 registry가 생성됩니다
- 불필요한 `Load()` 호출을 피하세요

### HTTP 엔드포인트

- `/metrics/` 엔드포인트는 매우 가볍습니다 (~1.7μs/req)
- 높은 요청 빈도에도 안정적으로 동작합니다

## 확장 패키지

이 코어 패키지는 다음과 같은 확장 패키지와 함께 사용됩니다:

- **CATV Extension** (`ntels.com/phylax/extensions/catv/pkg/metrics`)
  - CATVCollector: ClickHouse 데이터베이스 메트릭
  - ControlCollector: TCP 제어 작업 메트릭
  - UDPCollector: UDP 메시지 수신 메트릭

확장 패키지 사용 예시:

```go
import (
    coremetrics "ntels.com/phylax/core/pkg/metrics"
    catvmetrics "ntels.com/phylax/extensions/catv/pkg/metrics"
)

// CATV 메트릭 로드
if err := catvmetrics.Load(config); err != nil {
    log.Fatal(err)
}

// Core API 설정
api := &coremetrics.Api{}
api.Init(configPath, config)
api.Load()
```

## 문제 해결

### Collector 중복 등록 오류

**문제:** `duplicate metrics collector registration attempted`

**원인:** 같은 이름의 metric을 여러 번 등록

**해결:**
```go
// 싱글톤 패턴 사용
var (
    myMetrics *Metrics
    once      sync.Once
)

func GetMetrics() *Metrics {
    once.Do(func() {
        myMetrics = createMetrics()
    })
    return myMetrics
}
```

### 메트릭이 표시되지 않음

**문제:** `/metrics/` 엔드포인트에 메트릭이 없음

**원인:**
1. `api.Load()`가 호출되지 않음
2. Collector가 등록되지 않음
3. `Metrics.Use`가 `false`

**해결:**
```go
// 1. Collector 등록 확인
metrics.AddCollector("my_collector", myCollector)

// 2. Load 호출 확인
if err := api.Load(); err != nil {
    log.Fatal(err)
}

// 3. 설정 확인
config.Metrics.Use = true
config.Serve.Type = common.ServeTypeMaster
```

## 참고 자료

- [Prometheus Client Go Documentation](https://github.com/prometheus/client_golang)
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)

## 라이선스

이 패키지는 프로젝트의 라이선스를 따릅니다.

## 변경 이력

### 2025-12-15
- 초기 버전 작성
- 테스트 커버리지 100% 달성
- 벤치마크 테스트 추가
- README.md 문서화 완료
