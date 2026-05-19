# 데이터베이스 스키마

## 메타 DB (SQLite / PostgreSQL)

Pharos 플랫폼 자체 운영 데이터를 저장합니다. `core/internal/repositories/`에서 관리합니다.

| 테이블                   | 설명                                        |
| ------------------------ | ------------------------------------------- |
| `users`                  | 사용자 계정 (ID, 비밀번호 해시, 메타데이터) |
| `user_password_history`  | 비밀번호 이력 (재사용 방지)                 |
| `casbin_rule`            | Casbin RBAC 정책 (`p`, `g` 타입)            |
| `role_metadata_override` | Role 메타데이터 (표시명, 설명, 그룹)        |
| `user_metadata_override` | 사용자 메타데이터 (확장 필드)               |
| `dashboards`             | 대시보드 JSON 정의                          |
| `dashboard_folders`      | 대시보드 폴더 트리                          |
| `dashboard_history`      | 대시보드 변경 이력                          |
| `data_migration`         | 마이그레이션 실행 이력                      |
| `login_history`          | 로그인 이력 (접속 IP, 시각, 성공 여부)      |
| `fosite_*`               | OAuth2 토큰·세션 저장소 (fosite)            |

---

## 통계 DB (ClickHouse 클러스터)

CATV STB 대용량 시계열 데이터를 저장합니다.  
모든 테이블은 `ON CLUSTER clickhouse_cluster_replicated`로 생성되며, 로컬 테이블(`ReplicatedMergeTree`)과 분산 테이블(`Distributed`)이 쌍으로 존재합니다.

### STB 기기 정보

#### `catv.stb_information` / `catv.dist_stb_information`

```sql
stb_mac_addr    String          -- STB MAC 주소 (Primary Key 포함)
cm_mac_addr     String          -- CM(케이블 모뎀) MAC
stb_ip          String          -- STB IP 주소
cm_ip           String          -- CM IP 주소
so_name         String          -- SO(가입자 사업소) 명
l3_id           String          -- L3 장비 ID
cell_id         String          -- Cell ID
stb_model       String          -- STB 모델명
updated_at      DateTime        -- 최종 업데이트 시각

ENGINE = ReplicatedReplacingMergeTree(..., updated_at)
ORDER BY (so_name, l3_id, cell_id, stb_mac_addr)
```

### STB 제어

#### `catv.stb_control_job` / `catv.dist_stb_control_job`

```sql
request_received_at   DateTime
schedule_id           UUID
schedule_name         String
schedule_type         String
job_id                String
task_id               String
task_data             String          -- 제어 요청 데이터 (JSON)
task_status           String          -- pending / running / completed / failed
task_error_message    String
task_created_at       DateTime
task_updated_at       DateTime        -- ReplacingMergeTree 버전 키
control_executed_at   Nullable(DateTime)
cm_mac_addr           String
stb_mac_addr          String
work_type             String          -- 제어 타입 (예: stb_request_info)
result_code           String          -- 1(성공) / 0(기기 오류) / -1(시스템 오류)
result_message        String

ENGINE = ReplicatedReplacingMergeTree(..., task_updated_at)
PARTITION BY toYYYYMM(request_received_at)
ORDER BY (request_received_at, schedule_id, job_id, task_id, task_data)
TTL request_received_at + toIntervalYear(3)
```

#### `catv.stb_control_schedule` / `catv.dist_stb_control_schedule`

스케줄 기반 제어 설정 저장 (cron 표현식, 대상 STB 그룹, 제어 명령 등).

### STB 전송 데이터

#### `catv.stb_periodic_transmission` / `catv.dist_stb_periodic_transmission`

10분 주기 STB 상태 전송.

```sql
timestamp       DateTime            -- 수신 시각 (서버)
host_id         String              -- HOST ID (10자리)
mac_addr        String              -- STB MAC
cm_mac          String              -- CM MAC
stb_ip          String
cm_ip           String
stb_model       String
mw_ver          String              -- 미들웨어 버전
local_ver       String
cloud_ver       String
logging_time    String              -- STB 측 수집 시간
sending_time    String              -- STB 측 전송 시간
ch_sid          Nullable(String)    -- 채널 Source ID
ch_num          Nullable(String)    -- 채널 번호
ch_name         Nullable(String)    -- 채널명
ch_prg          Nullable(String)    -- 프로그램명
ch_freq         Nullable(String)    -- 주파수 (MHz)
ch_mode         Nullable(String)    -- 변조방식 (8VSB/256QAM)
pwr_lvl         Nullable(String)    -- Power Level (신호 세기)
snr             Nullable(String)    -- SNR (신호 품질)
sig_weak        Nullable(String)    -- 신호미약 팝업 (Y/N)
sig_weak_cnt    Nullable(String)    -- 팝업 횟수
stb_state       Nullable(String)    -- STB 상태 (watching/standby)
running_time    Nullable(String)    -- 구동 시간

ENGINE = ReplicatedMergeTree(...)
ORDER BY (timestamp, host_id)
TTL timestamp + toIntervalYear(3)
```

#### `catv.stb_daily_transmission` / `catv.dist_stb_daily_transmission`

일별 STB 설정 전송. 주요 필드: `limit_age`, `tv_lock`, `fav_ch`, `resolution`, `audio_mode`, `hdmi_cec`, `hdcp`, `hdr` 등 40여 개 STB 설정 항목.

#### `catv.stb_diagnostic_transmission` / `catv.dist_stb_diagnostic_transmission`

진단 이벤트 전송. 오류 코드, 진단 항목별 측정값 포함.

#### `catv.stb_quality_measurement_transmission` / 분산 테이블

Sleep 모드 진입 시 수집되는 채널별 품질 계측 데이터.

#### `catv.stb_network_quality_transition_transmission` / 분산 테이블

네트워크 품질 급변 이벤트.

### 집계 테이블 (Materialized View 기반)

#### `catv.stb_monitoring_hourly` / `catv.dist_stb_monitoring_hourly`

```sql
-- stb_periodic_transmission에서 시간별 집계 Materialized View
-- 시간대별 STB 수, 시청률, 신호미약 발생률, 평균 SNR/Power Level 등
ENGINE = ReplicatedAggregatingMergeTree(...)
ORDER BY (hour_timestamp, so_name, l3_id, cell_id)
```

#### `catv.stb_monitoring_hourly_stb` — STB 단위 시간별 집계

#### `catv.stb_monitoring_hourly_freq` — 주파수 단위 시간별 집계

#### `catv.stb_monitoring_hourly_stb_freq` — STB×주파수 단위 시간별 집계

#### `catv.stb_monitoring_hourly_cell` — Cell 단위 시간별 집계

#### `catv.stb_monitoring_hourly_stb_perf` — STB 성능 지표 시간별 집계

#### `catv.stb_monitoring_day` 계열

위 `_hourly` 테이블들과 동일한 구조의 일별 집계 테이블.

### 기상 데이터

#### `catv.weather_forecast_3h` / 분산 테이블

```sql
-- 기상청 3시간 예보
base_time    DateTime        -- 기준 시각
fcst_time    DateTime        -- 예보 시각
nx, ny       Int32           -- 격자 좌표
tmp          Float32         -- 기온 (°C)
pop          Float32         -- 강수확률 (%)
reh          Float32         -- 습도 (%)
wsd          Float32         -- 풍속 (m/s)
sky          String          -- 하늘상태 (맑음/구름많음/흐림)
pty          String          -- 강수형태 (없음/비/비설/눈/소나기)
```

#### `catv.weather_forecast_daily` / 분산 테이블

일별 최고/최저 기온, 강수량, 일사량 등.

#### `catv.weather_obs` / 분산 테이블

기상청 실시간 관측 (기온, 습도, 풍속, 낙뢰 발생 여부, 관측 지점 코드).

#### `catv.dist_so_weather_mapping`

SO 사업소 코드와 기상청 관측 격자/지점 코드 매핑 테이블.

---

## 스키마 네이밍 규칙

| 접두사          | 설명                          |
| --------------- | ----------------------------- |
| `catv.stb_`     | STB 관련 테이블               |
| `catv.dist_`    | Distributed 엔진 분산 테이블  |
| `catv.weather_` | 기상 데이터 테이블            |
| `catv.mv_`      | Materialized View (집계 전용) |

## 파티셔닝 전략

- 전송 데이터: `PARTITION BY toYYYYMM(timestamp)` — 월 단위 파티션
- 제어 작업: `PARTITION BY toYYYYMM(request_received_at)` — 월 단위 파티션
- TTL: 모든 라우 데이터 **3년** 자동 삭제

## 분산 클러스터 구성

```
clickhouse_cluster_replicated (클러스터 이름)
  Shard 1:
    Replica 1: /clickhouse/tables/{shard}/{table}, replica={replica}
    Replica 2: ...
  Shard 2:
    ...

dist_* 테이블:
  ENGINE = Distributed(clickhouse_cluster_replicated, catv, {table}, rand())
  -- rand(): 데이터를 무작위로 샤드에 분산
```
