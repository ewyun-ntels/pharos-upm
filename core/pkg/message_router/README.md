# Message Router

메시지 라우터는 NATS 기반의 분산 메시징 시스템을 위한 추상화 레이어를 제공하는 패키지입니다. 클라이언트와 서버 간의 메시지 통신을 단순화하고, Request-Reply 패턴 및 Publish-Subscribe 패턴을 지원합니다.

## 목차

- [개요](#개요)
- [아키텍처](#아키텍처)
- [주요 컴포넌트](#주요-컴포넌트)
  - [Client](#client)
  - [Server](#server)
  - [Message & MessageHandler](#message--messagehandler)
  - [Subject](#subject)
- [설치 및 설정](#설치-및-설정)
- [사용 예제](#사용-예제)
  - [클라이언트 사용](#클라이언트-사용)
  - [서버 시작](#서버-시작)
  - [Subject 등록](#subject-등록)
- [고급 기능](#고급-기능)
  - [Leader Election](#leader-election)
  - [Agent 중복 확인](#agent-중복-확인)
  - [TLS 지원](#tls-지원)
- [테스트](#테스트)

## 개요

Message Router는 다음과 같은 기능을 제공합니다:

- **클라이언트-서버 추상화**: NATS 특정 구현을 숨기고 범용 인터페이스 제공
- **Request-Reply 패턴**: 동기식 요청-응답 통신
- **Publish-Subscribe 패턴**: 비동기식 메시지 발행-구독
- **Subject 관리**: 메시지 주제와 핸들러의 중앙 집중식 관리
- **Leader Election**: JetStream 클러스터에서 리더 자동 선출
- **Agent 중복 방지**: 동일한 이름의 에이전트 실행 방지
- **TLS/보안 연결**: 암호화된 통신 지원

## 아키텍처

```
┌─────────────────────────────────────────────┐
│           Message Router Package            │
├─────────────────────────────────────────────┤
│                                             │
│  ┌──────────┐          ┌──────────┐        │
│  │  Client  │          │  Server  │        │
│  │Interface │          │Interface │        │
│  └────┬─────┘          └────┬─────┘        │
│       │                     │              │
│  ┌────▼────────┐      ┌────▼────────────┐  │
│  │ NatsClient  │      │  NatsServer     │  │
│  │             │      │                 │  │
│  │ - Request   │      │ - Start/Stop    │  │
│  │ - Publish   │      │ - Clustering    │  │
│  │ - Subscribe │      │ - JetStream     │  │
│  └─────────────┘      │ - LeaderChecker │  │
│                       └─────────────────┘  │
│                                             │
│  ┌──────────────────────────────────────┐  │
│  │       Subject Management             │  │
│  │  - Request/Reply Handlers            │  │
│  │  - Subject Registration              │  │
│  └──────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
            │
            ▼
    ┌───────────────┐
    │  NATS Server  │
    │  (JetStream)  │
    └───────────────┘
```

## 주요 컴포넌트

### Client

클라이언트는 NATS 서버에 연결하여 메시지를 송수신하는 인터페이스입니다.

#### Interface

```go
type Client interface {
    Connect(config common.Config) error
    Close() error
    Request(subjectKey, target string, data []byte) (message_router.Message, error)
    Publish(subjectKey string, data []byte) error
    Subscribe(subjectKey string, handler message_router.MessageHandler) error
}
```

#### 주요 메서드

- **Connect**: NATS 서버에 연결
  - TLS 지원
  - 자동 재연결 설정
  - Agent 중복 확인
  
- **Request**: 동기식 요청-응답 (타임아웃: 5초)
  ```go
  msg, err := client.GlobalClient.Request("subjectKey", "targetAgent", data)
  ```

- **Publish**: 비동기식 메시지 발행
  ```go
  err := client.GlobalClient.Publish("subjectKey", data)
  ```

- **Subscribe**: 메시지 구독
  ```go
  err := client.GlobalClient.Subscribe("subject", handler)
  ```

#### Global Client

```go
var GlobalClient Client = &NatsClient{}
```

전역 클라이언트 인스턴스를 사용하여 애플리케이션 전체에서 단일 연결을 공유합니다.

### Server

NATS 서버를 임베디드 모드로 실행하고 관리합니다.

#### Interface

```go
type Server interface {
    Start(configPath string, config common.Config) error
    Stop() error
}
```

#### 주요 기능

- **임베디드 NATS 서버**: 별도의 NATS 서버 없이 애플리케이션 내에서 실행
- **클러스터링**: 여러 서버 간 클러스터 구성
- **JetStream**: 영속성 메시징 및 스트리밍 지원
- **Leader Election**: 클러스터 내에서 리더 자동 선출

#### 설정 예시

```go
config := common.Config{
    Servers: map[string]common.ServerConfig{
        common.ServerTypeNats: {
            Port: 4222,
        },
    },
    Nats: common.NatsConfig{
        JetStream: true,
        JetStreamMaxMemory: 1024 * 1024 * 100,
        JetStreamMaxStore: 1024 * 1024 * 100,
        StoreDir: "/data/nats",
        Routes: []string{"nats://server1:6222", "nats://server2:6222"},
        Cluster: struct {
            Name   string
            Listen string
        }{
            Name:   "my-cluster",
            Listen: "nats://0.0.0.0:6222",
        },
    },
}

server := NewServer()
err := server.Start("/config/path", config)
```

### Message & MessageHandler

메시징 시스템의 핵심 데이터 구조입니다.

#### Message 구조체

```go
type Message struct {
    Data    []byte              // 메시지 데이터
    Subject string              // 메시지 주제
    Reply   string              // 응답 주제 (Request-Reply용)
    Headers map[string][]string // 메시지 헤더
    Raw     any                 // 원본 메시지 (*nats.Msg)
}
```

#### MessageHandler

```go
type MessageHandler func(msg Message)
```

메시지를 수신할 때 호출되는 콜백 함수입니다.

### Subject

Subject는 메시지 주제와 관련 핸들러를 정의합니다.

#### 구조체

```go
type Subject struct {
    Key            string         // 고유 식별자
    ServeType      string         // "agent" 또는 "server"
    RequestSubject string         // 요청 주제 패턴
    RequestHandler MessageHandler // 요청 핸들러
    ReplySubject   string         // 응답 주제 패턴
    ReplyHandler   MessageHandler // 응답 핸들러
}
```

#### 패턴 변수

- `{{.name}}`: 에이전트/서버 이름으로 치환됨
- 예: `agent.ping.{{.name}}` → `agent.ping.my-agent`

#### Subject 관리

```go
// Subject 등록
message_router.AddSubject(subject)

// Subject 조회
subject := message_router.Subjects.Get(key)

// 모든 Subject 조회
subjects := message_router.Subjects.GetAll()
```

## 설치 및 설정

### 의존성

```go
import (
    "ntels.com/pharos/core/pkg/message_router"
    "ntels.com/pharos/core/pkg/message_router/client"
    "ntels.com/pharos/core/pkg/message_router/server"
    "ntels.com/pharos/core/pkg/common"
)
```

### 기본 설정

1. **클라이언트 설정**

```go
config := common.Config{
    Clients: map[string]common.ClientConfig{
        common.ClientTypeNats: {
            URL: "nats://localhost:4222",
        },
    },
}
```

2. **서버 설정**

```go
config := common.Config{
    Serve: common.ServeConfig{
        Type:       common.ServeTypeAgent,
        NatsSchema: common.SchemaNats,
    },
    Servers: map[string]common.ServerConfig{
        common.ServerTypeNats: {
            Port: 4222,
        },
    },
}
```

## 사용 예제

### 클라이언트 사용

#### 1. 연결

```go
import "ntels.com/pharos/core/pkg/message_router/client"

// 클라이언트 연결
err := client.GlobalClient.Connect(config)
if err != nil {
    log.Fatal(err)
}
defer client.GlobalClient.Close()
```

#### 2. Request-Reply 패턴

```go
// 요청 보내기
data := []byte("request data")
response, err := client.GlobalClient.Request("mySubject", "targetAgent", data)
if err != nil {
    log.Printf("Request failed: %v", err)
    return
}

// 응답 처리
log.Printf("Response: %s", string(response.Data))
```

#### 3. Publish-Subscribe 패턴

```go
// 메시지 발행
data := []byte("publish data")
err := client.GlobalClient.Publish("mySubject", data)
if err != nil {
    log.Printf("Publish failed: %v", err)
}

// 메시지 구독
handler := func(msg message_router.Message) {
    log.Printf("Received: %s", string(msg.Data))
}

err := client.GlobalClient.Subscribe("mySubject", handler)
if err != nil {
    log.Printf("Subscribe failed: %v", err)
}
```

### 서버 시작

```go
import "ntels.com/pharos/core/pkg/message_router/server"

// 서버 생성
srv := server.NewServer()

// 서버 시작
err := srv.Start("/config/path", config)
if err != nil {
    log.Fatal(err)
}

// 종료 시 서버 중지
defer srv.Stop()
```

### Subject 등록

#### 기본 Subject 정의

```go
import "ntels.com/pharos/core/pkg/message_router"

subject := message_router.Subject{
    Key:       "myService",
    ServeType: common.ServeTypeAgent,
    
    // 요청 주제와 핸들러
    RequestSubject: "service.request.{{.name}}",
    RequestHandler: func(msg message_router.Message) {
        log.Printf("Request received: %s", string(msg.Data))
        
        // NATS 원본 메시지로 응답
        if natsMsg, ok := msg.Raw.(*nats.Msg); ok {
            natsMsg.Respond([]byte("response"))
        }
    },
    
    // 응답 주제와 핸들러
    ReplySubject: "service.reply.{{.name}}",
    ReplyHandler: func(msg message_router.Message) {
        log.Printf("Reply received: %s", string(msg.Data))
    },
}

// Subject 등록
message_router.AddSubject(subject)
```

#### Agent Ping Subject 예제

```go
import "ntels.com/pharos/core/pkg/message_router/subjects"

// Agent Ping Subject 가져오기
pingSubject := subjects.GetAgentPingSubject(config)

// 등록
message_router.AddSubject(pingSubject)
```

## 고급 기능

### Leader Election

JetStream 클러스터 환경에서 자동으로 리더를 선출합니다.

#### 동작 원리

1. **10초마다 리더 확인**
   - JetStream 활성화 여부
   - 클러스터링 여부
   - 현재 노드의 리더 상태

2. **리더인 경우**
   - 클라이언트 시작
   - Goflow(작업 흐름) 시작

3. **리더가 아닌 경우**
   - 클라이언트 중지
   - Goflow 중지

#### 리더 판단 로직

```go
func (leaderChecker *LeaderChecker) isLeader(server *NatsServer) bool {
    if !server.natsServer.JetStreamEnabled() || 
       !server.natsServer.JetStreamIsClustered() {
        return true  // 단일 노드는 항상 리더
    }
    return server.natsServer.JetStreamIsLeader()
}
```

### Agent 중복 확인

동일한 이름의 Agent가 이미 실행 중인지 확인합니다.

#### 동작 방식

```go
func (c *NatsClient) agentDuplicateCheck(config common.Config) error {
    if config.Serve.Type != common.ServeTypeAgent {
        return nil
    }
    
    // AgentPing 요청 시도
    if _, err := c.Request(message_router.AgentPing, "", nil); err == nil {
        return errors.New("there are agents with the same name")
    } else if !errors.Is(err, nats.ErrNoResponders) {
        return err
    }
    
    return nil
}
```

- **성공 응답**: 동일 이름의 Agent 존재 → 에러 반환
- **ErrNoResponders**: 중복 없음 → 정상 진행
- **기타 에러**: 연결 문제 등 → 에러 전파

### TLS 지원

암호화된 연결을 위한 TLS 설정을 지원합니다.

#### 클라이언트 TLS 설정

```go
config := common.Config{
    Clients: map[string]common.ClientConfig{
        common.ClientTypeNats: {
            URL:      "tls://localhost:4222",
            CaFile:   "/path/to/ca.pem",
            CertFile: "/path/to/cert.pem",
            KeyFile:  "/path/to/key.pem",
        },
    },
}
```

#### 서버 TLS 설정

```go
config := common.Config{
    Serve: common.ServeConfig{
        NatsSchema: common.SchemaTls,
    },
    Servers: map[string]common.ServerConfig{
        common.ServerTypeNats: {
            Port: 4222,
            Cert: common.CertConfig{
                CaFile:             "/path/to/ca.pem",
                CertFile:           "/path/to/cert.pem",
                KeyFile:            "/path/to/key.pem",
                InsecureSkipVerify: false,
            },
        },
    },
}
```

## 테스트

### 테스트 실행

```bash
# 모든 테스트 실행
go test ./...

# 특정 패키지 테스트
go test ./client
go test ./server
go test ./subjects

# 커버리지와 함께 실행
go test -cover ./...

# 상세 출력
go test -v ./...
```

### 주요 테스트 케이스

#### Client 테스트

- 연결 및 종료
- Request-Reply 패턴
- Publish-Subscribe 패턴
- 에러 처리
- Agent 중복 확인

#### Server 테스트

- 서버 시작 및 종료
- 기본 포트 설정
- 클러스터 설정
- TLS 설정

#### Leader Election 테스트

- 시작 및 정지
- 다중 시작/정지 사이클
- 동시성 테스트
- 리더 판단 로직

#### Subject 테스트

- Agent Ping 핸들러
- 빈 데이터 처리
- 대용량 데이터 처리
- 에러 시나리오

### 통합 테스트 예제

```go
func TestIntegration_RequestReply(t *testing.T) {
    // 서버 시작
    srv := server.NewServer()
    config := getTestConfig()
    err := srv.Start("/test/config", config)
    require.NoError(t, err)
    defer srv.Stop()
    
    // Subject 등록
    message_router.AddSubject(testSubject)
    
    // 클라이언트 연결
    err = client.GlobalClient.Connect(config)
    require.NoError(t, err)
    defer client.GlobalClient.Close()
    
    // Request-Reply 테스트
    response, err := client.GlobalClient.Request("testKey", "target", []byte("test"))
    require.NoError(t, err)
    assert.Equal(t, "expected response", string(response.Data))
}
```

## 환경 변수

- `HOSTNAME`: 에이전트/서버 이름 (기본값: 시스템 호스트명)

## 로깅

`log/slog`를 사용하여 구조화된 로그를 출력합니다.

```go
slog.Info("message", "key", value)
slog.Error("error message", "error", err.Error())
```

## 성능 고려사항

- **MaxPayload**: 메시지 최대 크기 설정 가능
- **Timeout**: Request 타임아웃 기본 5초
- **Reconnect**: 자동 재연결 활성화 (무제한 재시도)
- **Connection Pool**: 전역 클라이언트로 연결 재사용

## 문제 해결

### 연결 실패

```go
// 연결 상태 확인
if !conn.IsConnected() {
    log.Println("Not connected to NATS")
}
```

### Agent 중복 에러

```
Error: there are agents with the same name
```

→ 동일한 이름의 Agent가 이미 실행 중입니다. Agent 이름을 변경하거나 기존 Agent를 종료하세요.

### Request 타임아웃

```
Error: nats: timeout
```

→ 대상 서버/Agent가 응답하지 않습니다. 대상이 실행 중인지 확인하세요.

### ErrNoResponders

```
Error: nats: no responders available for request
```

→ 해당 Subject를 구독하는 서비스가 없습니다. Subject 등록 및 구독을 확인하세요.

## 라이선스

이 프로젝트는 NTELS의 내부 프로젝트입니다.

## 기여

버그 리포트 및 기능 제안은 이슈 트래커를 사용해주세요.
