# TCP Server Package

Pharos Core의 TCP 서버 패키지로, gnet 프레임워크를 기반으로 한 고성능 TCP 서버 관리 시스템을 제공합니다.

## 목차

- [개요](#개요)
- [주요 기능](#주요-기능)
- [아키텍처](#아키텍처)
- [사용법](#사용법)

## 개요

이 패키지는 gnet 기반의 TCP 서버를 등록하고 관리하는 핵심 컴포넌트입니다. 여러 TCP 서버를 동시에 운영할 수 있으며, 각 서버는 독립적인 포트와 이벤트 핸들러를 가집니다.

## 주요 기능

### 1. 멀티 서버 관리
- 이름 기반의 서버 등록 및 조회
- Thread-safe 서버 맵 관리
- 독립적인 포트 할당

### 2. Gnet 통합
- 고성능 이벤트 기반 네트워킹
- 커스텀 이벤트 핸들러 지원
- Non-blocking I/O

## 아키텍처

### 디렉토리 구조

```
tcp/
├── tcp.go         # TCP 서버 등록 및 관리
└── tcp_test.go    # 단위 테스트
```

### 핵심 컴포넌트

#### 1. 서버 레지스트리

```go
var Servers = core_internal.NewMap[internal.Server]()
```

전역 서버 맵으로 등록된 모든 TCP 서버를 관리합니다.

#### 2. 서버 등록 함수

```go
func AddGnetServer(name string, port int, eventHandler gnet.EventHandler)
```

**매개변수**:
- `name`: 서버 식별자 (고유해야 함)
- `port`: 리스닝 포트 번호
- `eventHandler`: gnet 이벤트 핸들러 구현체

**특징**:
- 자동으로 프로토콜을 "tcp"로 설정
- gnet 서버 인스턴스 생성 및 등록
- Thread-safe 동작

## 사용법

### TCP 서버 등록

```go
import (
    "github.com/panjf2000/gnet/v2"
    "ntels.com/pharos/core/pkg/server/tcp"
)

// 커스텀 이벤트 핸들러 구현
type MyTCPHandler struct {
    gnet.BuiltinEventEngine
}

func (h *MyTCPHandler) OnTraffic(c gnet.Conn) gnet.Action {
    // TCP 트래픽 처리 로직
    buf, _ := c.Next(-1)
    // 데이터 처리
    c.Write(buf)
    return gnet.None
}

// 서버 등록
func init() {
    tcp.AddGnetServer("my-tcp-server", 8080, &MyTCPHandler{})
}
```

### 등록된 서버 조회

```go
import "ntels.com/pharos/core/pkg/server/tcp"

func getServer() {
    server, exists := tcp.Servers.Get("my-tcp-server")
    if exists {
        // 서버 사용
    }
}
```

### 전체 서버 순회

```go
tcp.Servers.Range(func(name string, server internal.Server) bool {
    // 각 서버에 대한 작업 수행
    return true // false 반환 시 순회 중단
})
```

## Gnet 이벤트 핸들러

### 주요 이벤트

TCP 서버는 다음 이벤트를 처리할 수 있습니다:

#### OnBoot
```go
func (h *MyHandler) OnBoot(eng gnet.Engine) gnet.Action {
    // 서버 시작 시 초기화 로직
    return gnet.None
}
```

#### OnOpen
```go
func (h *MyHandler) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
    // 새 연결 수립 시
    return nil, gnet.None
}
```

#### OnTraffic
```go
func (h *MyHandler) OnTraffic(c gnet.Conn) gnet.Action {
    // 데이터 수신 시 처리
    buf, _ := c.Next(-1)
    // 비즈니스 로직
    return gnet.None
}
```

#### OnClose
```go
func (h *MyHandler) OnClose(c gnet.Conn, err error) gnet.Action {
    // 연결 종료 시 정리 작업
    return gnet.None
}
```

#### OnShutdown
```go
func (h *MyHandler) OnShutdown(eng gnet.Engine) {
    // 서버 종료 시 정리 작업
}
```

## 통합 예제

```go
package myextension

import (
    "log/slog"
    "github.com/panjf2000/gnet/v2"
    "ntels.com/pharos/core/pkg/common"
    "ntels.com/pharos/core/pkg/server/tcp"
)

type DataCollectorHandler struct {
    gnet.BuiltinEventEngine
    Config common.Config
}

func (h *DataCollectorHandler) OnBoot(eng gnet.Engine) gnet.Action {
    slog.Info("TCP data collector started")
    return gnet.None
}

func (h *DataCollectorHandler) OnTraffic(c gnet.Conn) gnet.Action {
    buf, err := c.Next(-1)
    if err != nil {
        slog.Error("failed to read data", "error", err)
        return gnet.Close
    }
    
    // 데이터 처리
    go h.processData(buf)
    
    return gnet.None
}

func (h *DataCollectorHandler) processData(data []byte) {
    // 비즈니스 로직 구현
}

func Load(config common.Config) error {
    tcp.AddGnetServer(
        "data-collector",
        config.DataCollector.Port,
        &DataCollectorHandler{Config: config},
    )
    return nil
}
```

## 성능 최적화

### Gnet 특성 활용

1. **이벤트 기반 아키텍처**: Goroutine 오버헤드 최소화
2. **제로 카피**: 불필요한 메모리 복사 방지
3. **커넥션 풀링**: 효율적인 연결 관리

### 권장 사항

- CPU 바운드 작업은 별도 고루틴에서 처리
- I/O 바운드 작업은 이벤트 핸들러에서 직접 처리
- 큰 데이터는 스트리밍 방식으로 처리

## 모니터링

서버 상태 확인:
```go
serverCount := 0
tcp.Servers.Range(func(name string, server internal.Server) bool {
    serverCount++
    slog.Info("registered tcp server", 
        slog.String("name", name),
        slog.Int("port", server.GetPort()),
    )
    return true
})
```

## 에러 처리

### 연결 에러
- `OnClose` 이벤트에서 에러 파라미터 확인
- 적절한 로깅 및 정리 작업 수행

### 데이터 처리 에러
- 에러 발생 시 로깅
- 필요시 연결 종료 (`gnet.Close` 반환)

## 의존성

주요 외부 패키지:
- `github.com/panjf2000/gnet/v2`: 고성능 네트워킹 프레임워크
- 내부 패키지:
  - `ntels.com/pharos/core/internal`: 공통 유틸리티
  - `ntels.com/pharos/core/pkg/server/internal`: 서버 인터페이스
  - `ntels.com/pharos/core/pkg/server/internal/gnet`: Gnet 구현체

## 관련 문서

- [Server Package](../README.md) - 상위 서버 패키지 문서
- [UDP Server Package](../udp/README.md) - UDP 서버 패키지
- [Gnet Documentation](https://github.com/panjf2000/gnet) - Gnet 공식 문서

## 라이선스

Pharos Core 프로젝트의 라이선스를 따릅니다.
