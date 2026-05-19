# CATV 데이터 마이그레이션 도구

## 목차

- [개요](#개요)
- [배경 및 목적](#배경-및-목적)
- [주요 기능](#주요-기능)
- [사용 사례](#사용-사례)
- [설정 파일 구조](#설정-파일-구조)
- [사용 방법](#사용-방법)
- [동작 방식](#동작-방식)
- [실행 쿼리](#실행-쿼리)
- [주의사항](#주의사항)
- [성능 튜닝](#성능-튜닝)
- [문제 해결](#문제-해결)

## 개요

ClickHouse 테이블 간 데이터를 시간 범위 기반으로 마이그레이션하는 도구입니다. 설정 파일에서 원본/대상 테이블과 시간 조건을 지정하여 대량의 데이터를 안전하고 효율적으로 이동할 수 있습니다.

## 배경 및 목적

ClickHouse 환경에서 다음과 같은 상황에 필요한 마이그레이션 도구입니다:

- **Materialized View 재구성**: MV 스키마 변경 시 기존 데이터의 재처리 필요
- **테이블 스키마 변경**: 새로운 테이블 구조로 기존 데이터 이동
- **데이터 재처리**: ETL 로직 변경으로 인한 과거 데이터 재가공
- **테이블 분리/병합**: 데이터 재구성을 위한 테이블 간 이동

**해결하는 문제:**
- 메모리 제약으로 대량 데이터를 한 번에 처리 불가
- 수작업으로 시간 범위를 나눠 쿼리하는 것은 비효율적이고 오류 발생 가능
- 원본 테이블에 중복 삽입하는 방식은 부적절
- 병렬 처리로 성능을 최적화하면서도 리소스 관리 필요

## 주요 기능

- **시간 범위 기반 데이터 마이그레이션**: 설정한 시작/종료 시간 범위 내의 데이터를 자동으로 분할하여 마이그레이션
- **Worker Pool 패턴**: 동시 실행 고루틴 수를 제한하여 리소스 관리
- **병렬 처리**: 여러 시간 구간을 동시에 처리하여 성능 최적화
- **실시간 진행률 표시**: 작업 진행 상황을 주기적으로 출력하여 모니터링 용이
- **실패율 임계값 체크**: 설정한 임계값 초과 시 경고하여 조기 문제 감지
- **실패 구간 자동 저장**: 실패한 시간 범위를 파일로 저장하여 재실행 용이
- **설정 검증**: 프로그램 실행 전 설정값 유효성 검증으로 오류 사전 방지
- **SQL Injection 방지**: 테이블명/컬럼명 검증으로 보안 강화
- **에러 추적 및 리포팅**: 각 작업의 성공/실패 여부를 추적하고 상세한 로그 제공
- **유연한 설정**: YAML 기반 설정 파일로 손쉬운 커스터마이징

## 사용 사례

### 1. Materialized View 재구성

**시나리오**: MV 스키마 변경으로 기존 데이터를 새 테이블로 이동

```bash
# 1단계: 동일 스키마로 destination 테이블 생성
CREATE TABLE stb_monitoring_raw_new AS stb_monitoring_raw;
CREATE TABLE dist_stb_monitoring_raw_new AS dist_stb_monitoring_raw ENGINE = Distributed(...);

# 2단계: 새 테이블에 대한 Materialized View 생성
CREATE MATERIALIZED VIEW mv_new TO target_table AS
SELECT ... FROM stb_monitoring_raw_new;

# 3단계: 마이그레이션 도구로 데이터 이동
# config.yml 설정:
#   source: dist_stb_monitoring_raw
#   destination: dist_stb_monitoring_raw_new
./migrator -config config.yml
```

### 2. 시간 범위별 데이터 추출

**시나리오**: 특정 기간 데이터를 별도 테이블로 아카이빙

```yaml
migration:
  table:
    source: "events_table"
    destination: "events_archive_2026_q1"
  where:
    time:
      field: "event_time"
      start: "2026-01-01 00:00:00"
      end: "2026-03-31 23:59:59"
      interval: "1h"
```

### 3. 테이블 스키마 변경 시 데이터 이관

**시나리오**: 컬럼이 추가/변경된 새 테이블로 기존 데이터 복사

```sql
-- 새 스키마로 테이블 생성 (기존 컬럼 + 신규 컬럼)
CREATE TABLE metrics_v2 (
  insert_time DateTime,
  metric_value Float64,
  new_column String DEFAULT ''  -- 새로 추가된 컬럼
) ENGINE = MergeTree() ORDER BY insert_time;

-- 마이그레이션 (기존 컬럼만 복사, 새 컬럼은 기본값)
-- config.yml: source: metrics_v1, destination: metrics_v2
```

## 설정 파일 구조

```yaml
clickhouse:
  debug: false
  timezone: "UTC"  # ClickHouse 서버의 타임존 (UTC, Asia/Seoul 등)
  endpoints: ["192.168.15.102:30900"]
  username: "default" 
  password: "default"
  database: "catv"
  dial_timeout: "30s"
  max_open_connections: 10
  max_idle_max_open_connections: 5
  connection_max_lifetime: "1h"
  read_timeout: "30s"

migration:
  max_workers: 10  # 동시 실행 워커 수
  failure_threshold: 0.1  # 실패율 임계값 (10%)
  progress_interval: 100  # 진행률 출력 간격 (작업 개수)
  save_failed_ranges: true  # 실패 구간 파일 저장 여부
  table:
    source: "dist_stb_monitoring_raw"
    destination: "dist_stb_monitoring_raw_xxx"
  where:
    time:
      field: "insert_time"  # WHERE 절에 사용할 시간 컬럼명
      start: "2026-02-06 00:00:00"  # KST 기준, >=
      end: "2026-02-06 23:59:59"    # KST 기준, <=
      interval: "1m"  # 시간 분할 간격
```

### 설정 항목 설명

#### ClickHouse 설정
- `debug`: 디버그 모드 활성화 여부
- `timezone`: ClickHouse 서버의 타임존 (예: "UTC", "Asia/Seoul", "America/New_York")
  - ClickHouse에 저장된 시간 데이터의 타임존과 일치시켜야 함
  - 일반적으로 "UTC" 사용
- `endpoints`: ClickHouse 서버 주소 목록
- `username`, `password`: 인증 정보
- `database`: 대상 데이터베이스
- `dial_timeout`: 연결 타임아웃
- `max_open_connections`: 최대 오픈 커넥션 수
- `max_idle_max_open_connections`: 최대 유휴 커넥션 수
- `connection_max_lifetime`: 커넥션 최대 수명
- `read_timeout`: 읽기 타임아웃

#### 마이그레이션 설정
- `max_workers`: 동시에 실행할 최대 워커 수 (0 이하일 경우 기본값 5 사용)
- `failure_threshold`: 실패율 임계값 (0.0 ~ 1.0, 예: 0.1 = 10%)
  - 설정한 비율 초과 시 경고 출력 (작업은 계속 진행)
  - 0으로 설정 시 체크 비활성화
- `progress_interval`: 진행률 로그 출력 간격 (작업 개수 단위)
  - 예: 100으로 설정 시 100개 작업마다 진행률 출력
  - 0 이하일 경우 기본값 100 사용
- `save_failed_ranges`: 실패한 시간 범위를 파일로 저장할지 여부 (true/false)
  - true 시 `failed_ranges.txt` 파일에 저장
- `table.source`: 원본 테이블명
- `table.destination`: 대상 테이블명
- `where.time.field`: WHERE 절에 사용할 시간 컬럼명 (예: "insert_time", "created_at", "event_time")
- `where.time.start`: 마이그레이션 시작 시간 (KST 기준, 포함 >=)
- `where.time.end`: 마이그레이션 종료 시간 (KST 기준, 포함 <=)
- `where.time.interval`: 시간 분할 간격 (Go duration 형식: 1m, 5m, 1h 등)

## 사용 방법

### 1. 설정 파일 준비

`config.yml` 파일을 생성하고 필요한 설정을 입력합니다.

### 2. 실행

```bash
go run main.go -config config.yml
```

또는 빌드 후 실행:

```bash
go build -o migrator main.go
./migrator -config config.yml
```

### 3. 로그 확인

실행 중 다음과 같은 로그가 출력됩니다:

```
time=... level=INFO msg="Configuration loaded successfully" config_file=config.yml
time=... level=INFO msg="migration started" total_jobs=1440 max_workers=10
time=... level=INFO msg="migration progress" completed=100 total=1440 progress_percent=6.9% success=100 failed=0
time=... level=INFO msg="migration progress" completed=200 total=1440 progress_percent=13.9% success=199 failed=1
time=... level=INFO msg="job completed" where_start_time="2026-02-05 15:30:00" where_end_time="2026-02-05 15:30:59" elapsed_time_seconds=12.5
time=... level=ERROR msg="job failed" where_start_time="2026-02-05 16:00:00" where_end_time="2026-02-05 16:00:59" elapsed_time_seconds=5.2 error="connection timeout"
time=... level=WARN msg="failure threshold exceeded" failure_rate="12.00%" threshold="10.00%" failed_count=173 total_processed=1440
time=... level=INFO msg="migration summary" total_jobs=1440 success=1438 failed=2
time=... level=INFO msg="failed ranges saved to file" file=failed_ranges.txt
time=... level=INFO msg="process time info" elapsed_time_seconds=287.5
```

### 4. 실패 구간 확인

실패한 시간 범위는 `failed_ranges.txt`에 저장됩니다 (`save_failed_ranges: true`인 경우):

```
# Failed time ranges - 2026-02-09T15:30:00+09:00
# Format: START_TIME | END_TIME | ERROR

2026-02-05 15:30:00 | 2026-02-05 15:30:59 | connection timeout
2026-02-05 16:45:00 | 2026-02-05 16:45:59 | query execution error
```

이 파일을 참고하여 실패한 구간만 재실행할 수 있습니다.

## 동작 방식

### 시간 범위 분할

1. 설정된 `start`와 `end` 시간을 `interval` 단위로 분할
2. 각 구간은 중복 없이 1초 단위로 정확히 구분됨
3. 마지막 구간은 정확히 `end` 시간까지 포함

**예시** (interval: 1m, start: 12:01:00, end: 12:03:00):
- 구간 1: `12:01:00 ~ 12:01:59`
- 구간 2: `12:02:00 ~ 12:02:59`
- 구간 3: `12:03:00 ~ 12:03:00`

**예시** (interval: 1m, start: 12:01:00, end: 12:02:57):
- 구간 1: `12:01:00 ~ 12:01:59`
- 구간 2: `12:02:00 ~ 12:02:57`

### Worker Pool

1. 설정된 `max_workers` 수만큼 워커 고루틴 생성
2. 채널을 통해 작업 큐 관리
3. 각 워커는 작업을 순차적으로 처리
4. 모든 작업 완료 시까지 대기

### 타임존 처리

프로그램은 이중 타임존 변환을 수행합니다:

1. **입력 타임존 (Source)**: 항상 Asia/Seoul (KST)
   - 설정 파일의 `start`, `end` 시간은 한국 시간 기준으로 작성
   
2. **저장 타임존 (Destination)**: `clickhouse.timezone` 설정값
   - ClickHouse에 저장된 실제 시간의 타임존
   - 일반적으로 "UTC" 사용

3. **변환 과정**:
   ```
   설정: "2026-02-06 12:00:00" (KST)
   ↓ 파싱
   2026-02-06 12:00:00 +0900 KST
   ↓ timezone: "UTC"로 변환
   2026-02-06 03:00:00 +0000 UTC  (9시간 차이)
   ↓ 쿼리 실행
   WHERE insert_time >= toDateTime('2026-02-06 03:00:00')
   ```

**중요**: `clickhouse.timezone`은 실제 ClickHouse 서버의 시간 저장 방식과 일치해야 합니다.

### 로그 최적화

- **성공 작업**: 10초 이상 걸린 작업만 로깅 (과도한 로그 방지)
- **실패 작업**: 모든 실패는 즉시 로깅
- **진행률**: `progress_interval` 간격으로 진행 상황 출력
- **요약 정보**: 전체 작업 완료 후 성공/실패 통계 출력

### 실패율 모니터링

- 10개 이상 작업 처리 후부터 실패율 체크 시작
- `failure_threshold` 초과 시 경고 로그 출력
- 작업은 중단 없이 계속 진행 (모니터링 목적)

### 에러 처리

- 개별 작업 실패 시에도 다른 작업은 계속 진행
- 모든 작업 완료 후 성공/실패 집계
- 실패한 작업의 상세 정보 로깅
- 하나라도 실패 시 프로세스는 에러 반환

## 실행 쿼리

각 시간 구간에 대해 다음 쿼리가 실행됩니다:

```sql
INSERT INTO {destination_table}
SELECT * FROM {source_table}
WHERE {time_field} >= toDateTime('{start_time}') 
  AND {time_field} <= toDateTime('{end_time}');
```

**예시** (KST 12:00:00 ~ 12:00:59를 마이그레이션, timezone: "UTC"):
```sql
INSERT INTO dist_stb_monitoring_raw_xxx
SELECT * FROM dist_stb_monitoring_raw
WHERE insert_time >= toDateTime('2026-02-06 03:00:00') 
  AND insert_time <= toDateTime('2026-02-06 03:00:59');
```

> 주의: 쿼리의 시간은 `clickhouse.timezone`에 따라 자동 변환됩니다.

## 주의사항

1. **설정 검증**:
   - 프로그램 시작 시 모든 설정값 자동 검증
   - 잘못된 설정 시 실행 전에 오류 메시지 출력
   - 테이블명/컬럼명은 영숫자, 언더스코어, 점만 허용 (SQL Injection 방지)

2. **타임존 설정**: 
   - `clickhouse.timezone`은 실제 ClickHouse 서버의 시간 저장 타임존과 일치해야 함
   - 불일치 시 9시간(또는 다른 시간대) 차이 발생 가능
   - 설정 파일의 `start`, `end`는 항상 KST 기준으로 작성

3. **리소스 관리**: 
   - `max_workers` 값을 적절히 설정하여 시스템 리소스 고려
   - 워커 수가 많을수록 ClickHouse 서버 부하 증가

4. **시간 설정**: 
   - `interval`이 너무 크면 메모리 문제 발생 가능, 너무 작으면 오버헤드 증가
   - 권장: 1m ~ 5m (데이터 양에 따라 조정)

5. **중복 방지**: 
   - 같은 시간 범위로 여러 번 실행하면 데이터 중복 발생
   - destination 테이블 확인 후 실행

6. **컬럼 일치**: 
   - source와 destination 테이블의 스키마가 호환되어야 함
   - `where.time.field`에 지정한 컬럼이 두 테이블 모두에 존재해야 함
   - 시간 컬럼은 DateTime 또는 DateTime64 타입이어야 함

## 성능 튜닝

- **max_workers**: CPU 코어 수와 ClickHouse 서버 부하 고려 (권장: 5~20)
- **interval**: 데이터 양과 메모리 상황에 따라 조정 (권장: 1m ~ 1h)
- **max_open_connections**: 동시 실행 쿼리 수 고려 (max_workers와 균형)
- **read_timeout**: 대량 데이터 처리 시 충분한 시간 설정
- **progress_interval**: 로그 출력 빈도 조정 (너무 작으면 로그 과다)
- **failure_threshold**: 적절한 임계값 설정으로 문제 조기 감지

## 문제 해결

### 설정 검증 에러
```bash
level=ERROR msg="Configuration validation failed" error="invalid timezone \"WRONG\": ..."
```
- 설정 파일의 해당 항목을 수정
- 타임존, duration, 테이블명 등 형식 확인

### 시간이 9시간 차이 나는 경우
```bash
# ClickHouse에서 타임존 확인
SELECT timezone();

# 실제 데이터 확인
SELECT insert_time FROM dist_stb_monitoring_raw LIMIT 1;
```
- `clickhouse.timezone` 설정을 ClickHouse 서버 설정과 일치시키기
- 일반적으로 ClickHouse는 UTC로 저장하므로 `timezone: "UTC"` 사용

### 타임아웃 발생
- `read_timeout`, `dial_timeout` 값을 증가 (예: 60s, 120s)
- `interval`을 더 작은 단위로 조정
- `max_workers`를 줄여 동시 쿼리 수 감소

### 메모리 부족
- `max_workers` 값을 감소 (예: 10 → 5)
- `interval`을 더 작은 단위로 조정 (예: 5m → 1m)
- ClickHouse 서버 메모리 상태 확인

### 연결 실패
- ClickHouse 서버 주소 및 포트 확인
- 인증 정보 확인 (username, password)
- 네트워크 연결 상태 확인
- 방화벽 설정 확인

### 시간 컬럼을 찾을 수 없는 경우
- `where.time.field` 설정이 올바른지 확인
- 테이블에 해당 컬럼이 존재하는지 확인
```sql
DESC dist_stb_monitoring_raw;
```

### 로그가 너무 적게 나오는 경우
- 성공한 작업은 10초 이상 걸린 경우만 로그 출력됨
- 진행률은 `progress_interval` 설정에 따라 주기적으로 출력
- 모든 성공 로그를 보려면 코드의 `if result.Duration.Seconds() > 10` 조건 제거

### 실패율이 높은 경우
- `failure_threshold` 초과 시 경고 메시지 확인
- `failed_ranges.txt` 파일에서 실패 패턴 분석
- ClickHouse 서버 로그 확인
- `max_workers` 감소 시도
- `interval` 조정 시도