# Go 백엔드 상세

## 진입점

```
core/cmd/pharos/main.go
  cobra CLI 루트 커맨드
  ├── server     → HTTP/NATS/UDP 서버 실행 (주 커맨드)
  ├── daemon     → 서버를 백그라운드(데몬) 으로 실행
  ├── util       → 유틸리티 서브커맨드
  ├── version    → 버전 정보 출력
  └── (Extension 커맨드)
      예) catv-control-settopbox
          → K8s Job Pod에서 실행되는 STB 제어 워커
```

`import_extensions.go`는 `go generate` 시 `meta/scripts/generate-metadata.go`가 자동 생성하며, `SITE_MODE`에 해당하는 Extension만 import합니다.

## 패키지 구조

### `core/pkg/` — 공개 패키지

#### `server/` — 서버 시작 조율

- `main.go` — cobra `serve` 커맨드. 설정 로드 → NATS 서버 시작 → Gin HTTP 서버 시작 → UDP 서버 시작 → 시그널 대기
- `extension.go` — `Register()` / `GetExtensions()` — init() 기반 Extension 등록 레지스트리
- `load.go` — 설정 파일 파싱 (Viper)
- `http/gin/` — Gin 서버, 미들웨어(인증, CORS, 요청 로그), 정적 파일 서빙
- `tcp/`, `udp/` — TCP/UDP 서버 인터페이스 (gnet 래핑)

#### `common/` — 공통 설정 시스템

전체 애플리케이션 설정은 `Config` 구조체 하나로 관리합니다.

```go
type Config struct {
    Serve         ServeConfig              // HTTP/NATS 스키마
    Servers       map[string]ServerConfig  // http, https, nats 포트·TLS
    Clients       map[string]ClientConfig  // NATS 클라이언트 접속 정보
    Database      orm.DatabaseConfig       // 메타 DB (SQLite/PostgreSQL)
    Auth          AuthConfig               // OAuth2, JWT, Cookie, CSRF
    User          UsersConfig              // 비밀번호 규칙, 로그인 잠금
    Role          map[string]CompositeRole // RBAC Role 정의
    Master        MasterConfig             // 통계 DB 설정
    Logger        RotationConfig           // 로그 로테이션
    Migration     MigrationConfig          // DB 마이그레이션 설정
    Nats          NatsConfig               // NATS JetStream 클러스터 설정
    Pool          PoolConfig               // 워커풀 / 커넥션풀 설정
    Catv          CatvConfig               // CATV Extension 전용 설정
    Metrics       MetricsConfig            // Prometheus 노출 여부
    ETL           ETLConfig                // Elasticsearch 접속 정보
    ...
}
```

`CatvConfig`는 다음을 포함합니다:

- `Database` — ClickHouse 접속 (통계 DB)
- `Collect.Transmission.*` — UDP 포트 (periodic/daily/diagnostic/quality/nqt)
- `Collect.Weather` — FTP 접속, 크론 스펙, 시뮬레이션 모드
- `Control` — STB 제어 설정 (K8s Job, 배치 크기, 재시도, Rate Limit)

#### `auth/` — 인증·인가 핸들러 (HTTP 레이어)

- OAuth2 토큰 발급(`/oauth/token`), 갱신, 폐기 엔드포인트 등록

#### `authhandler/` — Gin 미들웨어

- `GetAuthenticationHandler(required bool, roles ...string)` — JWT 검증 + Casbin 권한 확인

#### `message_router/` — NATS 메시징 추상화

```
NatsServer  ← NATS JetStream 내장 서버 (clustering 지원)
NatsClient  ← 자동 재연결, Request(5초 타임아웃), Publish, Subscribe
Subject     ← {Key, RequestSubject, RequestHandler, ReplyHandler} 구조체
Subjects    ← sync.Map 기반 전역 레지스트리
```

주요 Subject Key:

- `AgentPing` — Agent 연결 확인 (Heartbeat)
- `ElasticsearchIndex` — NATS → Elasticsearch ETL

#### `dashboard/` — 대시보드 CRUD API

- RESTful API (목록, 생성, 수정, 삭제, 이력, 즐겨찾기)
- Repository 패턴 — SQLite/PostgreSQL 이중 지원 (`factory/factory.go`로 런타임 선택)

#### `plugins/` — 외부 데이터소스 플러그인 관리

- HashiCorp go-plugin 기반, 별도 프로세스로 실행
- 지원 드라이버: `altibase`(CGO), `clickhouse`, `postgresql`, `http-receiver`, `elasticsearch`, `prometheus`, `vertica`

#### `pools/workers/` — 워커풀 시스템

- `WorkerPool` — 고정 크기 고루틴 풀
- `BatchSubmitter` — 여러 Job을 배치로 묶어 제출
- `RateLimiter` — token bucket 방식 초당 연결 제한
- STB 제어 시 `MaxConnectionsPerSecond` 설정으로 포트 고갈 방지

#### `metrics/` — Prometheus 메트릭 서비스

- Extension이 `RegisterCollector()`로 커스텀 Collector 등록
- `/metrics` 엔드포인트 노출

#### `websocket/` — Centrifuge WebSocket

- 실시간 알림, 로그 스트리밍에 사용

#### `workflow/` — DAG 워크플로우

- GoVisual(third_party) 기반 작업 그래프 실행

### `core/internal/` — 내부 전용 패키지

#### `auth/`

- `oauth2.go` — fosite Compose: ROPC + Refresh Token + JWT Introspection + Revocation + Client Credentials
- `token.go` — JWT 생성·검증 헬퍼
- `storage.go` — fosite Storage 구현 (SQLite/PostgreSQL 기반)

#### `casbin/`

- `enforcer.go` — sqlx 어댑터(`Blank-Xu/sqlx-adapter`) 기반 Casbin Enforcer 생성
- 정책 저장소: 메타 DB (`casbin_rule` 테이블)

#### `cert/`

- TLS 인증서 자동 생성 (`generate.go`) — RSA 2048 또는 직접 파일 지정
- JWT 서명 키가 없으면 자동 생성 (`cert/jwt/` 디렉토리에 저장)

#### `repositories/`

- 인터페이스: `DashboardRepository`, `UserRepository`, `DashboardFolderRepository`, ...
- 구현체: `sqlite/`, `postgres/`
- `factory/factory.go` — `orm.DatabaseConfig.Driver`에 따라 구현체 선택

#### `user/`

- `password_validator.go` — `PasswordRule` 검증 (대/소문자, 숫자, 특수문자 최소 개수)
- `username_validator.go` — 이메일 허용 여부 설정 포함

### `core/external/` — 외부 시스템 어댑터

#### `orm/`

- `orm.go` — `Handler()` / `StatisticsHandler()` — 풀에서 DB 연결 획득 후 핸들러 실행
- `database_pool.go` — 드라이버별 `*sqlx.DB` 풀 (singleton 3개: Default / Statistics / Service)
- `config.go` — `DatabaseConfig` 구조체 (ClickHouse, PostgreSQL, SQLite, Altibase 각각 세부 설정)
- `elasticsearch_client.go` — `go-elasticsearch` 클라이언트 래핑 (IndexBulk)
- `datetime.go` — ClickHouse DateTime 파싱 유틸

```
orm.DatabasePool          // 메타 DB (사용자·대시보드)
orm.StatisticsDatabasePool // 통계 DB (ClickHouse → STB 데이터)
orm.ServiceDatabasePool    // 서비스 DB (Agent 전용)
```

#### `goflow/`

- `goflow.go` — goflow 워크플로우 엔진 래핑
- `task/sqlite/` — SQLite Backup / Vacuum 내장 태스크

#### `config_loader/`

- `loader.go` — Viper 기반 TOML 파일 로더. 환경 변수 오버라이드 지원

## 설정 부팅 순서

```
1. cobra PreRunE
   → config 파일 로드 (Viper: TOML → 환경변수 오버라이드)
   → (daemon 모드라면) DaemonizeWithConfig()

2. RunE → runServerWithRetry()
   → Extension Load() 호출 (core_server.GetExtensions())
   → NATS 서버 시작
   → 마이그레이션 실행 (goose)
   → OAuth2 Provider 초기화 (fosite)
   → Casbin Enforcer 초기화
   → Gin 라우트 등록
     ├── /api/catv/* (CATV Extension)
     ├── /api/dashboard/*
     ├── /api/users/*
     ├── /api/auth/*
     └── /ui/* (embedded SPA)
   → UDP 서버 시작 (각 gnet 고루틴)
   → Weather 크론 시작
   → os.SIGTERM/SIGINT 대기
   → Graceful Shutdown (Extension Unload → UDP 서버 종료 → NATS 종료)
```

## 버전 관리

`core/pkg/version/` 패키지. `make` 빌드 시 ldflags로 주입:

```
-X 'ntels.com/pharos/core/pkg/version.Version=$(GIT_TAG)'
-X 'ntels.com/pharos/core/pkg/version.Commit=$(GIT_COMMIT)'
-X 'ntels.com/pharos/core/pkg/version.BuildTime=$(BUILD_TIME)'
```

태그 규칙: `vX.Y.Z-catv` 또는 `vX.Y.Z-catv-beta.N`
