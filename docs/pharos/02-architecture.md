# 시스템 아키텍처

## 전체 구조

```
┌─────────────────────────────────────────────────────────────────┐
│                          사용자 브라우저                          │
│                    React SPA (/ui/... 경로)                      │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTP/WebSocket
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Pharos Master Server                          │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Gin HTTP Server (:8080 / :8443)                         │   │
│  │  ├── /ui/*          → embedded React SPA (go:embed)      │   │
│  │  ├── /api/*         → REST API 핸들러                    │   │
│  │  ├── /api/ws        → Centrifuge WebSocket               │   │
│  │  └── /metrics       → Prometheus /metrics                 │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  NATS JetStream Server (:4222)                           │   │
│  │  ├── Request-Reply (동기 명령)                           │   │
│  │  └── Pub-Sub (비동기 이벤트, Leader Election)            │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  gnet UDP Server (다중 포트)                             │   │
│  │  ├── :periodic_port   → STB 주기 전송(10분)             │   │
│  │  ├── :daily_port      → STB 일일 전송                   │   │
│  │  ├── :diagnostic_port → STB 진단 전송                   │   │
│  │  ├── :quality_port    → STB 품질계측                    │   │
│  │  └── :nqt_port        → 네트워크 품질 전환              │   │
│  └──────────────────────────────────────────────────────────┘   │
└────────────────────┬──────────────────────┬─────────────────────┘
                     │                      │
          ┌──────────▼──────────┐  ┌────────▼──────────┐
          │   ClickHouse        │  │  SQLite / PostgreSQL│
          │  (통계 DB 클러스터)  │  │  (메타 DB)         │
          │  - STB 전송 데이터  │  │  - 사용자          │
          │  - 제어 작업 이력   │  │  - 대시보드        │
          │  - 기상 데이터      │  │  - Role/권한       │
          │  - Materialized View│  │  - 마이그레이션 이력│
          └─────────────────────┘  └───────────────────┘

              ┌────────────────────────────────────┐
              │     Kubernetes Cluster             │
              │  ┌──────────────────────────────┐  │
              │  │  catv-control-settopbox Job   │  │
              │  │  (N개 병렬 파드)              │  │
              │  │  STB IP → TCP 직접 연결       │  │
              │  │  제어 명령 전송               │  │
              │  │  결과 → ClickHouse 저장       │  │
              │  └──────────────────────────────┘  │
              └────────────────────────────────────┘

              ┌────────────────────────────────────┐
              │     기상청 FTP 서버                │
              │  (크론 스케줄 주기 수집)            │
              │  EUC-KR 이진 → UTF-8 → ClickHouse │
              └────────────────────────────────────┘
```

## Extension 시스템 아키텍처

```
┌──────────────────────────────────────────────────────────────┐
│                   SITE_MODE = catv                           │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  meta/site-config.json                                       │
│     └─ catv:                                                 │
│          extensions: [catv/business, clickhouse, demo]       │
│                                                              │
│              ↓ go generate (build time)                      │
│                                                              │
│  core/cmd/pharos/import_extensions.go (자동 생성)            │
│     import _ "ntels.com/pharos/extensions/catv/business"     │
│     import _ "ntels.com/pharos/extensions/dashboard/..."     │
│                                                              │
│  meta/generated/frontend/extension-loader.ts (자동 생성)    │
│     import '@pharos/catv-business'                           │
│     import '@pharos/clickhouse-datasource'                   │
│                                                              │
└──────────────────────────────────────────────────────────────┘

Go Extension 등록 흐름:
  extensions/catv/business/extension.go
    init() → core_server.Register(&Extension{})
           → core_plugins.RegisterExtension(...)

  Extension.Load(config)
    → control.Load(config)    # HTTP API 라우트 등록
    → metrics.Load(config)    # Prometheus Collector 등록
    → migrations.Load(config) # DB 마이그레이션 실행
    → collect.Load(config)    # UDP 서버 시작, Weather 크론 시작

Frontend Extension 등록 흐름:
  extension-loader.ts
    → extensionRegistry.register(...)  # 메뉴, 라우트, 패널, 로그인 등록

  App.tsx
    → getExtensionPages()   # 동적 라우트 주입
```

## 플러그인 (외부 데이터소스) 아키텍처

```
┌─────────────────────────────────────────────────────┐
│              Pharos Master Server                    │
│  core/pkg/plugins/                                   │
│    PluginManager                                     │
│      ├── 실행: ./bin/plugins/clickhouse (subprocess) │
│      ├── 실행: ./bin/plugins/postgresql             │
│      ├── 실행: ./bin/plugins/altibase  (CGO)        │
│      └── 실행: ./bin/plugins/http-receiver          │
│                                                      │
│    HashiCorp go-plugin (gRPC/net/rpc over stdio)     │
└─────────────────────────────────────────────────────┘
```

## 데이터 흐름 요약

### STB 데이터 수집

```
STB 기기 (수백만 대)
   │ UDP 전송 (10분 / 일별 / 이벤트 시)
   ▼
gnet UDP Handler
   │ JSON Unmarshal → sqlx NamedExec
   ▼
ClickHouse ReplicatedMergeTree
(stb_*_transmission 테이블)
   │ Materialized View
   ▼
stb_monitoring_hourly / stb_monitoring_day
(집계 테이블 — 시간별/일별 시청률, 불량률 등)
```

### STB 원격 제어

```
Web UI (POST /api/catv/control)
   │
   ▼
API Handler
  1. stb_information에서 대상 STB MAC/IP 조회
  2. stb_control_job에 작업 레코드 생성 (pending)
  3. Kubernetes Job 생성 (N개 파드, taskID 분배)
   │
   ▼
K8s Job Pod (catv-control-settopbox --task-id=xxx)
  1. DB에서 미처리 STB 목록 조회
  2. WorkerPool + RateLimiter (MaxConnPerSec)
  3. 각 STB IP → net.DialTimeout → 명령 전송
  4. AsyncBatchWriter → stb_control_job 결과 upsert
   │
   ▼
Web UI (GET /api/catv/control/results/:job_id)
  → 실시간 폴링으로 진행률/결과 확인
```

### ETL (NATS → Elasticsearch)

```
이벤트 발행자 (어떤 서비스든)
   │ NATS Publish "elasticsearch.index"
   │ Data: "{index}\t{docId}\t{json}"
   ▼
ETL Subject Handler (subjects/elasticsearch_index.go)
   │ 인덱스 이름 검증 (소문자, ^[a-z0-9][a-z0-9_\-+]*$)
   ▼
Elasticsearch IndexBulk API
```
