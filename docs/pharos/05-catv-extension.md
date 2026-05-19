# CATV Extension 상세

## 구조 개요

```
extensions/catv/business/
├── extension.go            # Extension 진입점 (init() 자동 등록)
├── go.mod
├── frontend/               # React CATV UI
│   └── src/
│       ├── features/control/   # STB 제어 페이지
│       ├── providers/          # CATV 전용 데이터 Provider
│       ├── permissions.ts      # 권한 상수
│       └── index.ts            # Extension 등록
└── pkg/
    ├── collect/            # 데이터 수집
    ├── control/            # STB 원격 제어
    ├── etl/                # NATS → Elasticsearch
    ├── metrics/            # Prometheus Custom Collector
    ├── migrations/         # ClickHouse 스키마 마이그레이션
    ├── permissions/        # RBAC 권한 상수 (Go)
    └── plugins/            # Plugin Extension 구현체
```

## Extension 등록 (extension.go)

```go
func init() {
    core_server.Register(&Extension{}) // HTTP 라우트 등록 등 서버 Extension
    core_plugins.RegisterExtension(plugins.NewCatvPluginExtension()) // 플러그인 Extension
}

func (e *Extension) Load(config common.Config) error {
    control.Load(config)    // STB 제어 API 라우트 등록
    metrics.Load(config)    // Prometheus Collector 등록
    migrations.Load(config) // ClickHouse 스키마 마이그레이션 실행
    collect.Load(config)    // UDP 수집 서버 + Weather 크론 시작
    return nil
}
```

---

## 1. 데이터 수집 (`pkg/collect/`)

### UDP 전송 수집

STB가 UDP로 실시간 상태를 전송합니다. **gnet v2** 이벤트 기반 UDP 서버로 수신합니다.

| 메시지 타입                | 전송 주기/조건        | 저장 테이블                                        |
| -------------------------- | --------------------- | -------------------------------------------------- |
| `Periodic`                 | 10분마다              | `catv.stb_periodic_transmission`                   |
| `Daily`                    | 하루 1회              | `catv.stb_daily_transmission`                      |
| `Diagnostic`               | 진단 이벤트 시        | `catv.stb_diagnostic_transmission`                 |
| `QualityMeasurement`       | Sleep 모드 진입 시    | `catv.stb_quality_measurement_transmission`        |
| `NetworkQualityTransition` | 네트워크 품질 변화 시 | `catv.stb_network_quality_transition_transmission` |

**처리 흐름**:

```
UDP 패킷 수신 (gnet OnTraffic)
  → JSON Unmarshal
  → InstrumentUDPHandler (Prometheus 계측)
  → sqlx NamedExec → ClickHouse INSERT
```

**Periodic 주요 필드** (인터페이스 정의서 v2.8 기반):

```
mac_addr, cm_mac, stb_ip, cm_ip, stb_model, mw_ver,
ch_freq, ch_mode, pwr_lvl (Power Level), snr (신호 품질),
sig_weak (신호미약 팝업), sig_weak_cnt, stb_state, running_time
```

### 기상 데이터 수집 (`pkg/collect/weather/`)

기상청 FTP 서버에서 수치 예보 데이터를 주기적으로 수집합니다.

```
cron.New(cron.WithSeconds())
  → CronSpec (예: "0 30 6 * * *") 마다 cronHandler 실행
  → FTP 접속 (ftpConnectionTimeout)
  → 디렉토리: BaseDirectory + "YYYYMMDD/"
  → 파일: 3hour.dat (3시간 예보), land.dat (육상 예보), shko.dat (낙뢰)
  → EUC-KR → UTF-8 변환 (golang.org/x/text/encoding/korean)
  → 파서 (tables.WeatherTable 인터페이스) → ClickHouse INSERT
```

**시뮬레이션 모드** (`config.Catv.Collect.Weather.Simulation = true`):

- FTP 연결 없이 `sample/20260114/` 디렉토리의 내장 샘플 데이터 사용

**수집 테이블**:

- `catv.weather_forecast_3h` — 3시간 단위 예보 (기온, 강수, 풍속, 습도 등)
- `catv.weather_forecast_daily` — 일별 예보
- `catv.weather_obs` — 실시간 관측 데이터 (낙뢰 포함)

---

## 2. STB 원격 제어 (`pkg/control/`)

### HTTP API (`pkg/control/http/`)

**STB 계층 구조 조회**:

```
GET /api/catv/sos                              → SO 목록
GET /api/catv/sos/:so_id/l3s                   → L3 장비 목록
GET /api/catv/sos/:so_id/l3s/:l3_id/cells      → Cell 목록
GET /api/catv/sos/:so_id/.../cells/:cell_id/settopboxes → STB MAC 목록
```

**제어 요청**:

```
POST /api/catv/control
  Body: { work_type, targets: [{so_id, l3_id, cell_id, stb_mac},...] }
  → stb_information에서 IP 조회
  → stb_control_job 레코드 생성 (pending)
  → K8s Job N개 생성 (job_id 분배)
  Response: { job_id, task_ids, fail: {mac: "error",...} }

GET /api/catv/control/results
  Query: limit, offset, job_id, work_type, task_id, schedule_name
  → 전체 작업 목록 (상태 집계 포함)

GET /api/catv/control/results/:job_id
  Query: limit, offset, task_id, cm_mac, stb_mac, status, result, search
  → 작업별 상세 결과 (STB 단위)
```

**스케줄 제어 API** (`ControlScheduleApi`):

```
GET    /api/catv/control/schedule   → 스케줄 목록
POST   /api/catv/control/schedule   → 스케줄 생성 (cron 등록)
PUT    /api/catv/control/schedule   → 스케줄 수정
DELETE /api/catv/control/schedule/:name → 스케줄 삭제
```

`robfig/cron`으로 스케줄 등록. 해당 시각에 `POST /api/catv/control`을 내부적으로 실행합니다.

### Kubernetes Job 연동 (`pkg/control/http/kubernetes/`)

Master 서버가 클러스터 내에서 실행되므로 `rest.InClusterConfig()`로 자동 인증합니다.

```go
// K8s Job 생성 시 spec 예시
spec.JobSpec.Parallelism = &config.Catv.Control.K8sJob.Count
spec.JobSpec.ActiveDeadlineSeconds = &config.Catv.Control.K8sJob.ActiveDeadlineSeconds
spec.JobSpec.BackoffLimit = &config.Catv.Control.K8sJob.BackoffLimit
spec.JobSpec.TTLSecondsAfterFinished = &config.Catv.Control.K8sJob.TTLSecondsAfterFinished
// 컨테이너: 동일 pharos 이미지, 커맨드 catv-control-settopbox --task-id={uuid}
```

### 제어 워커 (`pkg/control/command/`)

K8s Job으로 실행되는 STB 제어 프로세서입니다.

```
NewService(config, taskID)
  → DB에서 미처리 STB 목록 조회 (stb_control_job WHERE task_id=? AND status='pending')
  → WorkerPool (config.Pool.Workers.WorkerCount 고루틴)
  → RateLimiter (config.Catv.Control.MaxConnectionsPerSecond)
  → AsyncBatchWriter (배치 결과 writer, ClickHouse upsert)

service.Start()
  → 각 STB에 대해 WorkerPool에 Job 제출
  → Job: net.DialTimeout(stb_ip, timeout) → 명령 전송 → 응답 수신
  → 결과를 resultChannel로 전송 → AsyncBatchWriter가 ClickHouse에 저장

service.Stop()
  → context 취소 → 워커 정리 → AsyncBatchWriter flush
```

**재시도 정책**:

- `config.Catv.Control.Retry.Total` — 전체 재시도 횟수·지연
- `config.Catv.Control.Retry.PortExhaustion` — 포트 고갈 시 별도 재시도 횟수·지연

**주의**: 각 STB는 고유 IP이므로 커넥션 풀 재사용률 0%. `net.DialTimeout` 직접 사용.

---

## 3. ETL (`pkg/etl/`)

NATS JetStream 메시지를 받아 Elasticsearch에 색인합니다.

```
NATS Subject: "elasticsearch.index"
메시지 포맷: "{index}\t{documentID}\t{json_document}"

검증:
  - 인덱스 이름: ^[a-z0-9][a-z0-9_\-+]*$ (최대 255자)
  - Tab 구분자 3개 이상

처리:
  - orm.DatabasePool.GetElasticsearchClient() → IndexBulk()
```

---

## 4. Prometheus 메트릭 (`pkg/metrics/`)

### `CATVCollector` (catv_collector.go)

ClickHouse를 주기적으로 쿼리하여 커스텀 메트릭을 노출합니다.

| 메트릭                               | 설명                                              |
| ------------------------------------ | ------------------------------------------------- |
| `catv_udp_periodic_received_total`   | stb_periodic_transmission 최근 24시간 레코드 수   |
| `catv_udp_daily_received_total`      | stb_daily_transmission 최근 24시간 레코드 수      |
| `catv_udp_diagnostic_received_total` | stb_diagnostic_transmission 최근 24시간 레코드 수 |
| `catv_stb_total`                     | 전체 STB 수 (stb_information)                     |
| `catv_stb_watching`                  | 시청 중인 STB 수 (stb_state='watching')           |
| `catv_stb_standby`                   | 대기 중인 STB 수 (stb_state='standby')            |
| `catv_stb_signal_weak`               | 신호미약(sig_weak='Y') STB 수                     |
| `catv_stb_power_level`               | 평균 Power Level (pwr_lvl 집계)                   |
| `catv_stb_snr`                       | 평균 SNR 값                                       |
| `catv_db_query_duration_seconds`     | ClickHouse 쿼리 소요 시간                         |
| `catv_scrape_duration_seconds`       | Collector 전체 스크래핑 시간                      |
| `catv_scrape_success`                | 스크래핑 성공 여부 (1/0)                          |

### `UDPCollector` (udp_collector.go)

UDP 수신 처리를 계측하는 카운터입니다. `InstrumentUDPHandler()` 함수로 UDP 핸들러마다 호출합니다.

### `ControlCollector` (control_collector.go)

제어 작업 성공/실패/처리 중 개수를 추적합니다.

---

## 5. 마이그레이션 (`pkg/migrations/`)

goose v3 기반 ClickHouse 스키마 마이그레이션입니다.

```
migrations/schema/master/statistics/clickhouse/
├── 00000000000001_initialize.go          # 마이그레이션 초기화 (goose embed)
├── 20251211000000_create_stb_information.sql
├── 20251211000001_create_stb_control_job.sql
├── 20251216000000_create_stb_transmission.sql   # 5종 전송 테이블
├── 20260305000000_create_weather.sql             # 날씨 테이블 3종
├── 20260305000001_insert_dist_so_weather_mapping.sql  # SO-날씨 매핑 데이터
├── 20260311000000_create_mv_stb_transmission.sql # Materialized View
└── 20260313000000_create_table_stb_control_schedule.sql
```

모든 테이블은 `ON CLUSTER clickhouse_cluster_replicated`와 분산 테이블(`dist_*`)을 함께 생성합니다.

---

## 6. 프론트엔드 — STB 제어 UI (`frontend/src/features/control/`)

```
control/
├── page.tsx                    # 최상위 STB 제어 페이지 (Tab 레이아웃)
├── components/
│   ├── control-list/
│   │   ├── ControlList.tsx     # 스케줄 목록 / 작업 목록 테이블
│   │   ├── ScheduleDetailSheet.tsx  # 스케줄 상세 슬라이드 패널
│   │   └── columns.tsx         # TanStack Table 컬럼 정의
│   ├── control-history/
│   │   ├── ControlHistory.tsx  # 작업 실행 결과 테이블 (무한 스크롤)
│   │   ├── ControlHistoryDetailSheet.tsx # 단일 STB 실행 결과 상세
│   │   ├── columns.tsx         # 결과 컬럼 (상태 뱃지 포함)
│   │   ├── columns-stb-info.tsx # STB 정보 컬럼
│   │   └── ColumnVisibilityPanel.tsx # 컬럼 표시/숨김 패널
│   └── control-dialog/
│       └── ControlDialog.tsx   # 제어 요청 입력 다이얼로그
└── index.ts                    # 라우트/메뉴 Export
```

**페이지 구조**:

- 탭 1 `스케줄 목록` — 등록된 스케줄 목록. 행 클릭 시 `상세` 탭 동적 추가
- 탭 2 `작업 목록` (선택 시 추가) — 특정 job_id의 STB별 결과 테이블
- `ControlDialog` — SO/L3/Cell/STB 계층 선택 → 제어 명령 제출

**권한 제어**:

```typescript
const canRead = useHasPermission(CATV_PERMISSIONS.Read); // extension:catv:read
const canCreate = useHasPermission(CATV_PERMISSIONS.Create); // extension:catv:create
```

---

## 7. 권한 (RBAC)

```go
// permissions/permissions.go
const (
    Read   = "extension:catv:read"
    Create = "extension:catv:create"
    Update = "extension:catv:update"
    Delete = "extension:catv:delete"
)
```

Casbin 정책에 따라 Super Admin 또는 해당 권한을 부여받은 Role이 각 API에 접근할 수 있습니다.

```go
// 예시: 스케줄 생성 API
routes.POST("/control/schedule",
    authhandler.GetAuthenticationHandler(true,
        string(role.RoleSuperAdmin),
        permissions.Create),
    a.post)
```
