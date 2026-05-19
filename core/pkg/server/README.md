# Server Package

Pharos Core의 통합 서버 패키지로, HTTP/HTTPS, TCP, UDP 서버를 모두 지원하는 확장 가능한 멀티 프로토콜 서버 아키텍처를 제공합니다.

## 목차

- [개요](#개요)
- [주요 기능](#주요-기능)
- [아키텍처](#아키텍처)
- [서버 타입](#서버-타입)
- [사용법](#사용법)
- [확장 시스템](#확장-시스템)
- [보안](#보안)

## 개요

이 패키지는 Pharos 시스템의 Master 및 Agent 서버를 구동하는 핵심 컴포넌트입니다. HTTP/HTTPS, TCP, UDP 등 다양한 프로토콜을 지원하며, 설정 파일 기반의 유연한 구성, 데몬 모드 지원, 그리고 플러그인 방식의 확장 시스템을 제공합니다.

## 주요 기능

### 1. 멀티 프로토콜 지원
- **HTTP/HTTPS 서버**: Gin 기반 웹 서버 및 REST API
- **TCP 서버**: Gnet 기반 고성능 TCP 서버
- **UDP 서버**: Gnet 기반 데이터그램 서버
- 각 프로토콜별 독립적인 서버 인스턴스 관리

### 2. 서버 모드
- **Master 모드**: HTTP/HTTPS + TCP + UDP + Message Router 서버 운영
- **Agent 모드**: NATS 클라이언트로 동작하여 Master와 통신

### 3. 데몬 지원
- 백그라운드 프로세스로 실행 가능
- PID 파일 관리
- 로그 파일 리다이렉션
- Graceful shutdown 지원

### 4. 확장 시스템
- 플러그인 방식의 Extension 등록
- 초기화/정리 라이프사이클 관리
- 전역 레지스트리를 통한 중앙 관리

### 5. 보안 (HTTP/HTTPS)
- TLS/HTTPS 지원
- Content Security Policy (CSP)
- HSTS (HTTP Strict Transport Security)
- X-Content-Type-Options
- Request/Response 로깅

## 아키텍처

### 디렉토리 구조

```
server/
├── main.go           # Cobra 명령어 및 서버 실행 로직
├── extension.go      # 확장 모듈 시스템
├── load.go          # 설정 및 리소스 로딩
├── http/            # HTTP 서버 관리
│   ├── http.go      # HTTP 서버 레지스트리
│   ├── api/         # HTTP API 모듈 관리
│   │   └── api.go   # API 인터페이스 및 등록
│   └── gin/         # Gin 웹 프레임워크 구현
│       ├── server.go    # Gin 서버 구현
│       ├── middleware.go # 미들웨어 (로깅, 복구, 보안)
│       ├── slog.go      # 구조화된 로깅
│       └── static.go    # 정적 파일 서빙 (UI)
├── tcp/             # TCP 서버 관리
│   └── tcp.go       # TCP 서버 등록 및 관리
├── udp/             # UDP 서버 관리
│   └── udp.go       # UDP 서버 등록 및 관리
└── internal/        # 내부 공통 타입
    ├── types.go     # 서버 인터페이스 정의
    └── gnet/        # Gnet 기반 서버 구현
        └── server.go # TCP/UDP 서버 구현체
```

### 핵심 컴포넌트

#### 1. Server Interface (`internal/types.go`)

```go
type Server interface {
    Start(configPath string, config common.Config) error
    Stop() error
}
```

모든 서버가 구현해야 하는 기본 인터페이스입니다. HTTP, TCP, UDP 서버 모두 이 인터페이스를 구현합니다.

#### 2. Extension System (`extension.go`)

확장 모듈이 구현해야 하는 인터페이스:

```go
type Extension interface {
    Name() string
    Load(config common.Config) error
    Unload()
}
```

**특징**:
- 전역 레지스트리를 통한 중앙 관리
- Thread-safe 동작 (sync.RWMutex)
- 초기화 순서 보장
- 자동 정리 (defer 패턴)

**등록 방법**:
```go
func init() {
    server.Register(&MyExtension{})
}
```

#### 3. 서버 레지스트리

각 프로토콜별로 독립적인 서버 레지스트리를 관리합니다:

```go
// HTTP 서버 레지스트리
var http.Servers = core_internal.NewMap[internal.Server]()

// TCP 서버 레지스트리
var tcp.Servers = core_internal.NewMap[internal.Server]()

// UDP 서버 레지스트리
var udp.Servers = core_internal.NewMap[internal.Server]()
```

#### 4. Resource Loading (`load.go`)

시스템 리소스 초기화 순서:
1. 설정 파일 로딩 (`config_loader`)
2. 확장 모듈 로드 (`loadExtensions`)
3. 데이터베이스 마이그레이션 (`migration`)
4. ORM 초기화 (역할별 데이터베이스 풀)
   - **Default Pool**: 기본 데이터베이스
   - **Statistics Pool**: 통계 데이터베이스 (선택적)
   - **Service Pool**: Agent 서비스 데이터베이스 (Agent 모드)

## 사용법

### 기본 실행

```bash
# 기본 설정으로 실행
./pharos serve

# 커스텀 설정 파일로 실행
./pharos serve -c /path/to/config.toml

# 여러 설정 파일 사용
./pharos serve -c config1.toml -c config2.toml
```

### 데몬 모드

```bash
# 데몬으로 실행
./pharos serve -d

# PID 파일 및 로그 파일 지정
./pharos serve -d \
  --pid-file /var/run/pharos.pid \
  --log-file /var/log/pharos.log
```

### 명령줄 옵션

| 플래그 | 단축 | 기본값 | 설명 |
|--------|------|--------|------|
| `--config` | `-c` | `./config/config.toml` | 설정 파일 경로 (다중 지정 가능) |
| `--daemon` | `-d` | `false` | 데몬 모드로 실행 |
| `--pid-file` | | `/var/run/pharos.pid` | PID 파일 경로 |
| `--log-file` | | `/var/log/pharos.log` | 로그 파일 경로 |

## 확장 시스템

### Extension 구현 예제

```go
package myextension

import (
    "log/slog"
    "ntels.com/pharos/core/pkg/common"
    "ntels.com/pharos/core/pkg/server"
)

type MyExtension struct {
    // 확장 모듈의 상태
}

func init() {
    // 서버 시작 시 자동으로 등록됨
    server.Register(&MyExtension{})
}

func (e *MyExtension) Name() string {
    return "my-extension"
}

func (e *MyExtension) Load(config common.Config) error {
    slog.Info("Loading my extension")
    // 초기화 로직
    return nil
}

func (e *MyExtension) Unload() {
    slog.Info("Unloading my extension")
    // 정리 로직
}
```

### Extension 라이프사이클

1. **등록 단계**: `init()` 함수에서 `server.Register()` 호출
2. **로드 단계**: 서버 시작 시 `Load()` 호출
3. **실행 단계**: 서버 동작 중
4. **언로드 단계**: 서버 종료 시 `Unload()` 호출

## 서버 타입

### HTTP/HTTPS 서버

Gin 웹 프레임워크 기반의 HTTP/HTTPS 서버입니다.

#### 주요 기능
- REST API 엔드포인트
- 정적 파일 서빙 (UI)
- WebSocket 지원
- 미들웨어 체인 (로깅, 복구, 보안)
- TLS/HTTPS 지원

#### 등록 방법

```go
import (
    "ntels.com/pharos/core/pkg/server/http"
    "ntels.com/pharos/core/pkg/server/http/gin"
)

// 기본 Gin 서버 등록
http.AddServer("my-http-server", &gin.Server{})

// 커스텀 HTTP 서버 등록
http.AddServer("custom-server", &MyCustomServer{})
```

자세한 내용은 [HTTP Server 문서](http/gin/README.md)를 참조하세요.

### TCP 서버

Gnet 프레임워크 기반의 고성능 TCP 서버입니다.

#### 주요 기능
- 이벤트 기반 아키텍처
- Non-blocking I/O
- 커넥션 관리
- 커스텀 프로토콜 구현

#### 등록 방법

```go
import (
    "github.com/panjf2000/gnet/v2"
    "ntels.com/pharos/core/pkg/server/tcp"
)

type MyTCPHandler struct {
    gnet.BuiltinEventEngine
}

func (h *MyTCPHandler) OnTraffic(c gnet.Conn) gnet.Action {
    buf, _ := c.Next(-1)
    // 데이터 처리
    c.Write(response)
    return gnet.None
}

// TCP 서버 등록
tcp.AddGnetServer("my-tcp-server", 8080, &MyTCPHandler{})
```

자세한 내용은 [TCP Server 문서](tcp/README.md)를 참조하세요.

### UDP 서버

Gnet 프레임워크 기반의 UDP 데이터그램 서버입니다.

#### 주요 기능
- Connectionless 통신
- 데이터그램 기반 처리
- 브로드캐스트/멀티캐스트 지원
- 빠른 응답 시간

#### 등록 방법

```go
import (
    "github.com/panjf2000/gnet/v2"
    "ntels.com/pharos/core/pkg/server/udp"
)

type MyUDPHandler struct {
    gnet.BuiltinEventEngine
}

func (h *MyUDPHandler) OnTraffic(c gnet.Conn) gnet.Action {
    buf, _ := c.Next(-1)
    // 데이터그램 처리
    return gnet.None
}

// UDP 서버 등록
udp.AddGnetServer("my-udp-server", 9090, &MyUDPHandler{})
```

자세한 내용은 [UDP Server 문서](udp/README.md)를 참조하세요.

## HTTP API 구조

### API 인터페이스

```go
type Api interface {
    Init(string, common.Config)
    Use() bool
    Load() error
    Unload()
    RegisterRoutes(gin.IRoutes)
    GetRelativePath() string
}
```

### 기본 API 모듈

서버는 다음 API 모듈을 기본 제공합니다:

- `version`: 버전 정보
- `badges`: 뱃지 관리
- `authhandler`: 인증 처리
- `alert`: 알림 관리
- `notification`: 노티피케이션
- `workflow`: 워크플로우
- `websocket`: WebSocket 통신
- `dashboard`: 대시보드
- `ui_config`: UI 설정
- `plugins`: 플러그인 관리
- `groups`: 그룹 관리
- `userhandler`: 사용자 처리

### API 추가 방법

```go
import "ntels.com/pharos/core/pkg/server/http/api"

func init() {
    api.AddApi(&MyApi{})
}
```

## 서버 시작 흐름

### Master 모드

```
runServerWithRetry()
  ├─ Message Router 서버 시작
  │   └─ NATS 서버 초기화 및 Subject 등록
  │
  ├─ HTTP 서버 시작
  │   ├─ Gin 엔진 생성 및 미들웨어 등록
  │   ├─ 정적 파일 라우팅 (UI)
  │   ├─ API 로딩 및 라우팅 등록
  │   └─ HTTP/HTTPS 리스너 시작
  │
  ├─ TCP 서버 시작 (등록된 경우)
  │   └─ 각 TCP 서버 인스턴스 시작
  │
  ├─ UDP 서버 시작 (등록된 경우)
  │   └─ 각 UDP 서버 인스턴스 시작
  │
  └─ Graceful Shutdown 대기
```

### Agent 모드

```
runServerWithRetry()
  ├─ NATS 클라이언트 연결
  ├─ Master와 통신 채널 설정
  └─ Graceful Shutdown 대기
```

### Graceful Shutdown

서버는 다음 신호를 받으면 우아하게 종료됩니다:
- `SIGINT` (Ctrl+C)
- `SIGTERM`

종료 과정:
1. 신호 수신 감지
2. UDP 서버 종료 (5초 타임아웃)
3. TCP 서버 종료 (5초 타임아웃)
4. HTTP 서버 종료 (3초 타임아웃)
5. Message Router 종료
6. 확장 모듈 언로드
7. ORM 정리

## 멀티 서버 관리

### 서버 등록 및 시작

각 프로토콜 타입의 서버는 독립적으로 등록하고 시작할 수 있습니다:

```go
// Extension의 Load 메서드에서 등록
func (e *MyExtension) Load(config common.Config) error {
    // HTTP API 등록
    http_api.AddApi(&MyController{})
    
    // HTTP 서버 등록 (선택적, 기본 서버가 자동 등록됨)
    http.AddServer("custom-http", &gin.Server{})
    
    // TCP 서버 등록
    tcp.AddGnetServer("data-collector", 8080, &DataCollectorHandler{})
    
    // UDP 서버 등록
    udp.AddGnetServer("metrics-collector", 9090, &MetricsCollectorHandler{})
    
    return nil
}
```

### 서버 조회

등록된 서버는 레지스트리를 통해 조회할 수 있습니다:

```go
// HTTP 서버 조회
if httpServer, exists := http.Servers.Get("my-http-server"); exists {
    // 서버 사용
}

// TCP 서버 조회
if tcpServer, exists := tcp.Servers.Get("data-collector"); exists {
    // 서버 사용
}

// UDP 서버 조회
if udpServer, exists := udp.Servers.Get("metrics-collector"); exists {
    // 서버 사용
}

// 전체 서버 순회
http.Servers.Range(func(name string, server internal.Server) bool {
    slog.Info("registered server", "name", name)
    return true
})
```

## 보안 (HTTP/HTTPS)

### Content Security Policy

```
default-src 'self';
script-src 'self' 'unsafe-inline' 'unsafe-eval';
style-src 'self' 'unsafe-inline';
img-src 'self' data: blob:;
font-src 'self' data:;
connect-src 'self' ws: wss: http: https:;
frame-ancestors 'self';
object-src 'none';
base-uri 'self'
```

### HSTS (HTTP Strict Transport Security)

HTTPS 요청 시 자동으로 활성화:
```
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

### 추가 보안 헤더

- `X-Content-Type-Options: nosniff` - MIME 스니핑 방지

### TLS 설정

HTTPS 모드 시 인증서 자동 로딩:
- 설정 파일에서 인증서 경로 지정
- `cert.SetCertification()`을 통한 TLS 설정

## Message Router 통합

### Master 모드
- Message Router 서버 시작
- Subject 등록 및 관리
- Agent Ping Subject 자동 등록

### Agent 모드
- NATS 클라이언트 연결
- Master와 양방향 통신

## 설정 예시

### 설정 파일 (config.toml)

```toml
[serve]
type = "master"  # "master" or "agent"

[serve.http]
port = 8080
use_https = false
# cert_file = "/path/to/cert.pem"
# key_file = "/path/to/key.pem"

[daemon]
pid_file = "/var/run/pharos.pid"
log_file = "/var/log/pharos.log"

# TCP 서버는 Extension에서 동적으로 등록
# UDP 서버는 Extension에서 동적으로 등록
```

### Extension에서 서버 등록

```go
package myextension

import (
    "ntels.com/pharos/core/pkg/common"
    "ntels.com/pharos/core/pkg/server"
    "ntels.com/pharos/core/pkg/server/tcp"
    "ntels.com/pharos/core/pkg/server/udp"
)

type MyExtension struct{}

func init() {
    server.Register(&MyExtension{})
}

func (e *MyExtension) Name() string {
    return "my-extension"
}

func (e *MyExtension) Load(config common.Config) error {
    // TCP 서버 등록
    tcp.AddGnetServer("data-collector", 
        config.MyExtension.TcpPort, 
        &DataCollectorHandler{Config: config})
    
    // UDP 서버 등록
    udp.AddGnetServer("metrics-collector",
        config.MyExtension.UdpPort,
        &MetricsCollectorHandler{Config: config})
    
    return nil
}

func (e *MyExtension) Unload() {
    // 정리 로직
}
```

## 로깅

### 구조화된 로깅 (slog)

```go
// HTTP 요청 로깅
slog.Info("http request",
    slog.String("method", c.Request.Method),
    slog.String("path", path),
    slog.Int("status", status),
    slog.Duration("latency", latency),
    slog.String("client_ip", c.ClientIP()),
)

// 서버 시작 로깅
slog.Info("starting tcp server",
    slog.String("name", "data-collector"),
    slog.Int("port", 8080))

// 서버 종료 로깅
slog.Info("server stopped gracefully",
    slog.String("type", "udp"),
    slog.String("name", "metrics-collector"))
```

### 로그 회전

`slogrotate` 패키지를 통한 자동 로그 파일 회전 지원.

## History 기능 (HTTP)

요청/응답 히스토리를 ClickHouse에 저장:
- Request/Response Body 로깅
- 특정 경로 무시 가능 (`IgnorePaths`)
- Master 모드에서 선택적으로 활성화

## 마이그레이션 (HTTP)

서버 시작 후 백그라운드에서 API 마이그레이션 실행:
- HTTP 엔드포인트 준비 완료 후 실행 (200ms 지연)
- `github.com/loykin/apimigrate` 사용
- 설정 파일의 `[migration]` 섹션에서 제어

## 에러 처리

### Panic Recovery
- 모든 Panic은 자동으로 복구됨
- 에러 로그 기록
- HTTP 500 응답 반환

### 데몬 모드 Panic
- Panic 복구 및 로그 기록
- PID 파일 자동 제거
- systemd 재시작 정책에 따라 동작

## 성능 고려사항

### HTTP/HTTPS 서버
- Gin의 경량 라우팅
- 미들웨어 체인 최적화
- Keep-Alive 연결 지원

### TCP/UDP 서버 (Gnet)
- 이벤트 기반 아키텍처로 고루틴 오버헤드 최소화
- 멀티코어 지원 (`WithMulticore(true)`)
- 제로 카피 버퍼 관리
- CPU 바운드 작업은 별도 고루틴으로 분리

## 모니터링

### 서버 상태 확인

```go
// 등록된 모든 서버 확인
func checkAllServers() {
    slog.Info("=== HTTP Servers ===")
    http.Servers.Range(func(name string, _ internal.Server) bool {
        slog.Info("http server", "name", name)
        return true
    })
    
    slog.Info("=== TCP Servers ===")
    tcp.Servers.Range(func(name string, _ internal.Server) bool {
        slog.Info("tcp server", "name", name)
        return true
    })
    
    slog.Info("=== UDP Servers ===")
    udp.Servers.Range(func(name string, _ internal.Server) bool {
        slog.Info("udp server", "name", name)
        return true
    })
}
```

## 의존성

주요 외부 패키지:
- `github.com/gin-gonic/gin`: HTTP 웹 프레임워크
- `github.com/panjf2000/gnet/v2`: TCP/UDP 네트워킹 프레임워크
- `github.com/spf13/cobra`: CLI 프레임워크
- 내부 패키지:
  - `ntels.com/pharos/core/pkg/common`: 공통 타입 및 상수
  - `ntels.com/pharos/core/pkg/daemon`: 데몬 관리
  - `ntels.com/pharos/core/pkg/message_router`: 메시지 라우팅
  - `ntels.com/pharos/core/external/orm`: 데이터베이스 ORM

## 테스트

각 서브패키지는 단위 테스트를 포함합니다:

### HTTP 관련
- `http/api/api_test.go`
- `http/gin/middleware_test.go`
- `http/gin/server_test.go`
- `http/gin/slog_test.go`
- `http/gin/static_test.go`

### TCP/UDP 관련
- `tcp/tcp_test.go`
- `udp/udp_test.go`
- `internal/gnet/server_test.go`

## 관련 문서

- [HTTP Server](http/gin/README.md) - Gin 기반 HTTP/HTTPS 서버 상세
- [TCP Server](tcp/README.md) - TCP 서버 구현 및 사용법
- [UDP Server](udp/README.md) - UDP 서버 구현 및 사용법
- [CATV Extension](../../extensions/catv/pkg/server/README.md) - UDP 서버 활용 예제

## 라이선스

Pharos Core 프로젝트의 라이선스를 따릅니다.
