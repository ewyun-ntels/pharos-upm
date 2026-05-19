# UDP Server Package

Pharos Core의 UDP 서버 패키지로, gnet 프레임워크를 기반으로 한 고성능 UDP 서버 관리 시스템을 제공합니다.

## 목차

- [개요](#개요)
- [주요 기능](#주요-기능)
- [아키텍처](#아키텍처)
- [사용법](#사용법)

## 개요

이 패키지는 gnet 기반의 UDP 서버를 등록하고 관리하는 핵심 컴포넌트입니다. 여러 UDP 서버를 동시에 운영할 수 있으며, 각 서버는 독립적인 포트와 이벤트 핸들러를 가집니다.

## 주요 기능

### 1. 멀티 서버 관리
- 이름 기반의 서버 등록 및 조회
- Thread-safe 서버 맵 관리
- 독립적인 포트 할당

### 2. Gnet 통합
- 고성능 이벤트 기반 네트워킹
- 커스텀 이벤트 핸들러 지원
- Non-blocking I/O

### 3. UDP 특화 기능
- Connectionless 통신
- 데이터그램 기반 처리
- 브로드캐스트/멀티캐스트 지원

## 아키텍처

### 디렉토리 구조

```
udp/
├── udp.go         # UDP 서버 등록 및 관리
└── udp_test.go    # 단위 테스트
```

### 핵심 컴포넌트

#### 1. 서버 레지스트리

```go
var Servers = core_internal.NewMap[internal.Server]()
```

전역 서버 맵으로 등록된 모든 UDP 서버를 관리합니다.

#### 2. 서버 등록 함수

```go
func AddGnetServer(name string, port int, eventHandler gnet.EventHandler)
```

**매개변수**:
- `name`: 서버 식별자 (고유해야 함)
- `port`: 리스닝 포트 번호
- `eventHandler`: gnet 이벤트 핸들러 구현체

**특징**:
- 자동으로 프로토콜을 "udp"로 설정
- gnet 서버 인스턴스 생성 및 등록
- Thread-safe 동작

## 사용법

### UDP 서버 등록

```go
import (
    "github.com/panjf2000/gnet/v2"
    "ntels.com/pharos/core/pkg/server/udp"
)

// 커스텀 이벤트 핸들러 구현
type MyUDPHandler struct {
    gnet.BuiltinEventEngine
}

func (h *MyUDPHandler) OnTraffic(c gnet.Conn) gnet.Action {
    // UDP 데이터그램 처리 로직
    buf, _ := c.Next(-1)
    // 데이터 처리
    c.Write(buf) // 응답 전송 (선택적)
    return gnet.None
}

// 서버 등록
func init() {
    udp.AddGnetServer("my-udp-server", 9090, &MyUDPHandler{})
}
```

### 등록된 서버 조회

```go
import "ntels.com/pharos/core/pkg/server/udp"

func getServer() {
    server, exists := udp.Servers.Get("my-udp-server")
    if exists {
        // 서버 사용
    }
}
```

### 전체 서버 순회

```go
udp.Servers.Range(func(name string, server internal.Server) bool {
    // 각 서버에 대한 작업 수행
    return true // false 반환 시 순회 중단
})
```

## Gnet 이벤트 핸들러

### UDP 전용 이벤트

UDP 서버는 다음 이벤트를 처리할 수 있습니다:

#### OnBoot
```go
func (h *MyHandler) OnBoot(eng gnet.Engine) gnet.Action {
    // 서버 시작 시 초기화 로직
    return gnet.None
}
```

#### OnTraffic
```go
func (h *MyHandler) OnTraffic(c gnet.Conn) gnet.Action {
    // 데이터그램 수신 시 처리
    buf, _ := c.Next(-1)
    
    // 발신자 정보 확인
    remoteAddr := c.RemoteAddr()
    
    // 비즈니스 로직
    processData(buf, remoteAddr)
    
    // 응답 전송 (필요시)
    c.Write(response)
    
    return gnet.None
}
```

#### OnShutdown
```go
func (h *MyHandler) OnShutdown(eng gnet.Engine) {
    // 서버 종료 시 정리 작업
}
```

### UDP vs TCP 차이점

UDP 서버에서는 다음 이벤트를 사용하지 않습니다:
- `OnOpen`: UDP는 connectionless 프로토콜
- `OnClose`: 연결 종료 개념 없음

## 통합 예제

### 데이터 수집 서버

```go
package collector

import (
    "encoding/json"
    "log/slog"
    "github.com/panjf2000/gnet/v2"
    "ntels.com/pharos/core/pkg/common"
    "ntels.com/pharos/core/pkg/server/udp"
)

type MetricsCollectorHandler struct {
    gnet.BuiltinEventEngine
    Config common.Config
}

func (h *MetricsCollectorHandler) OnBoot(eng gnet.Engine) gnet.Action {
    slog.Info("UDP metrics collector started",
        slog.Int("port", h.Config.Collector.Port))
    return gnet.None
}

func (h *MetricsCollectorHandler) OnTraffic(c gnet.Conn) gnet.Action {
    buf, err := c.Next(-1)
    if err != nil {
        slog.Error("failed to read datagram", "error", err)
        return gnet.None
    }
    
    // 비동기 처리 (UDP는 빠른 응답이 중요)
    go h.processMetrics(buf, c.RemoteAddr().String())
    
    return gnet.None
}

func (h *MetricsCollectorHandler) processMetrics(data []byte, source string) {
    var metrics map[string]interface{}
    if err := json.Unmarshal(data, &metrics); err != nil {
        slog.Error("failed to parse metrics",
            "source", source,
            "error", err)
        return
    }
    
    // 메트릭 처리 로직
    slog.Debug("received metrics",
        slog.String("source", source),
        slog.Int("count", len(metrics)))
}

func Load(config common.Config) error {
    udp.AddGnetServer(
        "metrics-collector",
        config.Collector.Port,
        &MetricsCollectorHandler{Config: config},
    )
    return nil
}
```

### Request-Response 패턴

```go
type RequestResponseHandler struct {
    gnet.BuiltinEventEngine
}

func (h *RequestResponseHandler) OnTraffic(c gnet.Conn) gnet.Action {
    request, _ := c.Next(-1)
    
    // 요청 처리
    response := h.handleRequest(request)
    
    // 응답 전송
    c.Write(response)
    
    return gnet.None
}

func (h *RequestResponseHandler) handleRequest(data []byte) []byte {
    // 요청 처리 로직
    return []byte("ACK")
}
```

## 성능 최적화

### UDP 특성 활용

1. **Stateless 처리**: 연결 상태 관리 오버헤드 없음
2. **빠른 처리**: 핸드셰이크 없이 즉시 데이터 전송
3. **비동기 처리**: 고루틴을 활용한 병렬 처리

### 권장 사항

- **빠른 응답**: `OnTraffic`에서 최소한의 처리만 수행
- **비동기 처리**: 무거운 작업은 고루틴으로 분리
- **버퍼 크기**: 데이터그램 크기에 맞는 버퍼 설정
- **에러 핸들링**: 패킷 손실에 대비한 재전송 로직

### 주의사항

- UDP는 신뢰성을 보장하지 않음 (패킷 손실 가능)
- 순서가 보장되지 않음
- 대용량 데이터는 분할 전송 필요
- MTU 크기 고려 (일반적으로 1500 바이트)

## 모니터링

### 서버 상태 확인

```go
serverCount := 0
udp.Servers.Range(func(name string, server internal.Server) bool {
    serverCount++
    slog.Info("registered udp server", 
        slog.String("name", name),
        slog.Int("port", server.GetPort()),
    )
    return true
})
```

### 메트릭 수집

```go
type StatsHandler struct {
    gnet.BuiltinEventEngine
    receivedPackets int64
    receivedBytes   int64
}

func (h *StatsHandler) OnTraffic(c gnet.Conn) gnet.Action {
    buf, _ := c.Next(-1)
    
    atomic.AddInt64(&h.receivedPackets, 1)
    atomic.AddInt64(&h.receivedBytes, int64(len(buf)))
    
    return gnet.None
}
```

## 에러 처리

### 데이터 파싱 에러

```go
func (h *MyHandler) OnTraffic(c gnet.Conn) gnet.Action {
    buf, err := c.Next(-1)
    if err != nil {
        slog.Error("read error", "error", err)
        return gnet.None // UDP는 연결을 닫을 필요 없음
    }
    
    if err := validateData(buf); err != nil {
        slog.Warn("invalid data",
            "source", c.RemoteAddr().String(),
            "error", err)
        return gnet.None
    }
    
    processData(buf)
    return gnet.None
}
```

## 실제 사용 사례

### 1. 로그 수집
- 대량의 로그 데이터 수집
- 낮은 레이턴시 요구사항

### 2. 메트릭 모니터링
- 실시간 메트릭 전송
- 높은 처리량 필요

### 3. 알림 시스템
- 이벤트 통지
- Fire-and-forget 방식

### 4. IoT 데이터 수집
- 센서 데이터 수집
- 경량 프로토콜 필요

## 의존성

주요 외부 패키지:
- `github.com/panjf2000/gnet/v2`: 고성능 네트워킹 프레임워크
- 내부 패키지:
  - `ntels.com/pharos/core/internal`: 공통 유틸리티
  - `ntels.com/pharos/core/pkg/server/internal`: 서버 인터페이스
  - `ntels.com/pharos/core/pkg/server/internal/gnet`: Gnet 구현체

## 관련 문서

- [Server Package](../README.md) - 상위 서버 패키지 문서
- [TCP Server Package](../tcp/README.md) - TCP 서버 패키지
- [CATV UDP Extension](../../../extensions/catv/pkg/server/udp/README.md) - CATV UDP 확장 예제
- [Gnet Documentation](https://github.com/panjf2000/gnet) - Gnet 공식 문서

## 라이선스

Pharos Core 프로젝트의 라이선스를 따릅니다.
