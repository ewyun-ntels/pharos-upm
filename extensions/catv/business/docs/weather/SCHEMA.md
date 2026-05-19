# 날씨 데이터 스키마 및 매핑 전략

**최종 업데이트**: 2026-03-06

---

## 목차

- [📋 테이블 구조](#-테이블-구조)
  - [1. weather_obs / dist_weather_obs (실시간 관측)](#1-weather_obs--dist_weather_obs-실시간-관측)
  - [2. weather_forecast_3h / dist_weather_forecast_3h (3시간 예보)](#2-weather_forecast_3h--dist_weather_forecast_3h-3시간-예보)
  - [3. weather_forecast_daily / dist_weather_forecast_daily (일별 예보)](#3-weather_forecast_daily--dist_weather_forecast_daily-일별-예보)
  - [4. v_weather_unified (통합 VIEW)](#4-v_weather_unified-통합-view)
- [🗺️ SO-날씨 매핑](#️-so-날씨-매핑)
- [🔍 쿼리 예시](#-쿼리-예시)
  - [Grafana 시계열 쿼리](#grafana-시계열-쿼리)
  - [STB-날씨 JOIN](#stb-날씨-join)
  - [날씨별 STB 장애율](#날씨별-stb-장애율)
- [📊 데이터 특성](#-데이터-특성)
- [🛠️ 유지보수](#️-유지보수)
  - [파티셔닝 전략](#파티셔닝-전략)
  - [TTL 설정](#ttl-설정)
  - [인덱스 최적화](#인덱스-최적화)
  - [ReplacingMergeTree 특징](#replacingmergetree-특징)
- [📝 마이그레이션](#-마이그레이션)
  - [초기 데이터 INSERT](#초기-데이터-insert)
  - [테이블 생성 순서](#테이블-생성-순서)
  - [Distributed vs Local 테이블](#distributed-vs-local-테이블)

---

## 📋 테이블 구조

### 1. weather_obs / dist_weather_obs (실시간 관측)

```sql
CREATE TABLE catv.weather_obs ON CLUSTER clickhouse_cluster_replicated
(
    `obs_date` DateTime COMMENT '관측날짜 (날짜 + 발표시각)',
    `location` String COMMENT '지역',
    `temperature` Float32 COMMENT '기온',
    `weather_text` String COMMENT '날씨텍스트',
    `weather_icon` Int32 COMMENT '날씨아이콘',
    `wind_speed` Float32 COMMENT '풍속',
    `wind_direction` String COMMENT '풍향',
    `rainfall` Nullable(Float32) COMMENT '강수량',
    `humidity` Int32 COMMENT '습도',

    `insert_time` DateTime DEFAULT now()
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/catv/weather_obs', '{replica}')
PARTITION BY toYYYYMM(obs_date)
ORDER BY (obs_date, location)
TTL obs_date + toIntervalYear(3);
```

```sql
CREATE TABLE catv.dist_weather_obs ON CLUSTER clickhouse_cluster_replicated
AS catv.weather_obs
ENGINE = Distributed(clickhouse_cluster_replicated, catv, weather_obs, rand());
```

**데이터 예시:**

```
2026/01/14#10:00#속초#-0.8#맑음#1#3.1#서#-#18#
→ obs_date='2026-01-14 10:00:00', location='속초', temperature=-0.8, rainfall=NULL
```

**⚠️ 강수량(rainfall) NULL 처리 이유:**

기상청 데이터에서 강수량이 '-' (하이픈)로 표시되는 경우는 **'데이터 없음'** 또는 **'관측되지 않음'**을 의미합니다.

- **0.0mm vs '-' 차이**:
  - `0.0`: 비가 내리지 않았거나 0.1mm 미만의 미량
  - `'-'`: 관측 불가/기기 장애/결측 (데이터 수집 실패)

- **Nullable로 처리하는 이유**:
  1. **데이터 정확성**: '-'를 0으로 치환하면 "비 안 옴"과 "관측 실패"를 구분할 수 없음
  2. **Grafana 집계**: NULL은 자동으로 제외되므로 avg(), sum() 계산 시 정확
  3. **문자열 방지**: 문자 '-'로 저장하면 Grafana에서 숫자 연산 불가

- **쿼리 예시**:

  ```sql
  -- NULL 자동 제외 (올바른 평균)
  SELECT avg(rainfall) FROM catv.dist_weather_obs;  -- 실제 강수가 있던 날만 계산

  -- NULL을 0으로 치환 (필요시)
  SELECT COALESCE(rainfall, 0) FROM catv.dist_weather_obs;
  ```

---

### 2. weather_forecast_3h / dist_weather_forecast_3h (3시간 예보)

```sql
CREATE TABLE catv.weather_forecast_3h ON CLUSTER clickhouse_cluster_replicated
(
    `forecast_hour` DateTime COMMENT '예보시간',
    `location` String COMMENT '지역명',
    `weather_icon` Int32 COMMENT '날씨아이콘',
    `weather_text` String COMMENT '날씨텍스트',
    `temperature` Int32 COMMENT '기온',
    `rain_prob` Int32 COMMENT '강수확률',
    `sunrise` DateTime COMMENT '일출시각',
    `sunset` DateTime COMMENT '일몰시각',

    `insert_time` DateTime DEFAULT now()
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/catv/weather_forecast_3h', '{replica}')
PARTITION BY toYYYYMM(forecast_hour)
ORDER BY (forecast_hour, location)
TTL forecast_hour + toIntervalYear(3);
```

```sql
CREATE TABLE catv.dist_weather_forecast_3h ON CLUSTER clickhouse_cluster_replicated
AS catv.weather_forecast_3h
ENGINE = Distributed(clickhouse_cluster_replicated, catv, weather_forecast_3h, rand());
```

**데이터 예시:**

```
백령도#09#04#흐림#-2#30#07:47#17:37#
→ forecast_hour='2026-01-14 09:00:00', location='백령도', temperature=-2, rain_prob=30
```

---

### 3. weather_forecast_daily / dist_weather_forecast_daily (일별 예보)

```sql
CREATE TABLE catv.weather_forecast_daily ON CLUSTER clickhouse_cluster_replicated
(
    `forecast_date` DateTime COMMENT '예보날짜',
    `location` String COMMENT '지역명',
    `temp_min` Int32 COMMENT '최저기온',
    `temp_max` Int32 COMMENT '최고기온',
    `rain_prob_am` Int32 COMMENT '오전강수확률',
    `rain_prob_pm` Int32 COMMENT '오후강수확률',
    `weather_text` String COMMENT '날씨텍스트',
    `weather_icon` Int32 COMMENT '날씨아이콘',
    `wind_direction` String COMMENT '풍향',

    `insert_time` DateTime DEFAULT now()
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/catv/weather_forecast_daily', '{replica}')
PARTITION BY toYYYYMM(forecast_date)
ORDER BY (forecast_date, location)
TTL forecast_date + toIntervalYear(3);
```

```sql
CREATE TABLE catv.dist_weather_forecast_daily ON CLUSTER clickhouse_cluster_replicated
AS catv.weather_forecast_daily
ENGINE = Distributed(clickhouse_cluster_replicated, catv, weather_forecast_daily, rand());
```

**데이터 예시:**

```
2026/01/14#백령도#-5#5#20#30#흐림#4#남동-남#
→ forecast_date='2026-01-14 00:00:00', location='백령도', temp_min=-5, temp_max=5
```

---

### 4. v_weather_unified (통합 VIEW)

```sql
CREATE VIEW catv.v_weather_unified ON CLUSTER clickhouse_cluster_replicated
AS
SELECT
    datetime,
    location,
    -- obs_* (weather_obs - 실시간 관측)
    max(obs_temperature) AS obs_temperature,
    max(obs_weather_text) AS obs_weather_text,
    max(obs_weather_icon) AS obs_weather_icon,
    max(obs_wind_speed) AS obs_wind_speed,
    max(obs_wind_direction) AS obs_wind_direction,
    max(obs_rainfall) AS obs_rainfall,
    max(obs_humidity) AS obs_humidity,
    -- forecast_3h_* (weather_forecast_3h - 3시간 예보)
    max(forecast_3h_weather_icon) AS forecast_3h_weather_icon,
    max(forecast_3h_weather_text) AS forecast_3h_weather_text,
    max(forecast_3h_temperature) AS forecast_3h_temperature,
    max(forecast_3h_rain_prob) AS forecast_3h_rain_prob,
    max(forecast_3h_sunrise) AS forecast_3h_sunrise,
    max(forecast_3h_sunset) AS forecast_3h_sunset,
    -- forecast_daily_* (weather_forecast_daily - 일별 예보)
    max(forecast_daily_temp_min) AS forecast_daily_temp_min,
    max(forecast_daily_temp_max) AS forecast_daily_temp_max,
    max(forecast_daily_rain_prob_am) AS forecast_daily_rain_prob_am,
    max(forecast_daily_rain_prob_pm) AS forecast_daily_rain_prob_pm,
    max(forecast_daily_weather_text) AS forecast_daily_weather_text,
    max(forecast_daily_weather_icon) AS forecast_daily_weather_icon,
    max(forecast_daily_wind_direction) AS forecast_daily_wind_direction,
    max(insert_time) AS insert_time
FROM (
    SELECT
        obs_date AS datetime,
        location,
        temperature AS obs_temperature,
        weather_text AS obs_weather_text,
        weather_icon AS obs_weather_icon,
        wind_speed AS obs_wind_speed,
        wind_direction AS obs_wind_direction,
        rainfall AS obs_rainfall,
        humidity AS obs_humidity,
        NULL AS forecast_3h_weather_icon,
        NULL AS forecast_3h_weather_text,
        NULL AS forecast_3h_temperature,
        NULL AS forecast_3h_rain_prob,
        NULL AS forecast_3h_sunrise,
        NULL AS forecast_3h_sunset,
        NULL AS forecast_daily_temp_min,
        NULL AS forecast_daily_temp_max,
        NULL AS forecast_daily_rain_prob_am,
        NULL AS forecast_daily_rain_prob_pm,
        NULL AS forecast_daily_weather_text,
        NULL AS forecast_daily_weather_icon,
        NULL AS forecast_daily_wind_direction,
        insert_time
    FROM catv.dist_weather_obs FINAL

    UNION ALL

    SELECT
        forecast_hour AS datetime,
        location,
        NULL AS obs_temperature,
        NULL AS obs_weather_text,
        NULL AS obs_weather_icon,
        NULL AS obs_wind_speed,
        NULL AS obs_wind_direction,
        NULL AS obs_rainfall,
        NULL AS obs_humidity,
        weather_icon AS forecast_3h_weather_icon,
        weather_text AS forecast_3h_weather_text,
        temperature AS forecast_3h_temperature,
        rain_prob AS forecast_3h_rain_prob,
        sunrise AS forecast_3h_sunrise,
        sunset AS forecast_3h_sunset,
        NULL AS forecast_daily_temp_min,
        NULL AS forecast_daily_temp_max,
        NULL AS forecast_daily_rain_prob_am,
        NULL AS forecast_daily_rain_prob_pm,
        NULL AS forecast_daily_weather_text,
        NULL AS forecast_daily_weather_icon,
        NULL AS forecast_daily_wind_direction,
        insert_time
    FROM catv.dist_weather_forecast_3h FINAL

    UNION ALL

    SELECT
        forecast_date AS datetime,
        location,
        NULL AS obs_temperature,
        NULL AS obs_weather_text,
        NULL AS obs_weather_icon,
        NULL AS obs_wind_speed,
        NULL AS obs_wind_direction,
        NULL AS obs_rainfall,
        NULL AS obs_humidity,
        NULL AS forecast_3h_weather_icon,
        NULL AS forecast_3h_weather_text,
        NULL AS forecast_3h_temperature,
        NULL AS forecast_3h_rain_prob,
        NULL AS forecast_3h_sunrise,
        NULL AS forecast_3h_sunset,
        temp_min AS forecast_daily_temp_min,
        temp_max AS forecast_daily_temp_max,
        rain_prob_am AS forecast_daily_rain_prob_am,
        rain_prob_pm AS forecast_daily_rain_prob_pm,
        weather_text AS forecast_daily_weather_text,
        weather_icon AS forecast_daily_weather_icon,
        wind_direction AS forecast_daily_wind_direction,
        insert_time
    FROM catv.dist_weather_forecast_daily FINAL
)
GROUP BY datetime, location;
```

**핵심 기능:**

- UNION ALL로 3개 테이블 통합
- max() 함수로 NULL 자동 제외
- datetime 기준 GROUP BY (시간축 통일)
- FINAL 키워드로 ReplacingMergeTree 최종 버전만 조회

---

## 🗺️ SO-날씨 매핑

### 5. so_weather_mapping / dist_so_weather_mapping

```sql
CREATE TABLE catv.so_weather_mapping ON CLUSTER clickhouse_cluster_replicated
(
    `so_nm` String COMMENT '가입자SO명 (SCRBR_SO_NM)',
    `weather_location` String COMMENT '날씨 지역명',
    `province` String COMMENT '광역시/도명 (선택적 분석용)',

    `insert_time` DateTime DEFAULT now()
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/catv/so_weather_mapping', '{replica}')
ORDER BY (so_nm);
```

```sql
CREATE TABLE catv.dist_so_weather_mapping ON CLUSTER clickhouse_cluster_replicated
AS catv.so_weather_mapping
ENGINE = Distributed(clickhouse_cluster_replicated, catv, so_weather_mapping, rand());
```

### 매핑 테이블 (23 SO → 8 지역)

| SO명       | weather_location | province       |
| ---------- | ---------------- | -------------- |
| 강서방송   | 서울             | 서울특별시     |
| 광진성동   | 서울             | 서울특별시     |
| 도봉강북   | 서울             | 서울특별시     |
| 동대문방송 | 서울             | 서울특별시     |
| 동작관악   | 서울             | 서울특별시     |
| 마포서대문 | 서울             | 서울특별시     |
| 서초강남   | 서울             | 서울특별시     |
| 노원방송   | 서울             | 서울특별시     |
| 송파       | 서울             | 서울특별시     |
| 종로중구   | 서울             | 서울특별시     |
| 서부산방송 | 부산             | 부산광역시     |
| 낙동방송   | 부산             | 부산광역시     |
| 동남방송   | 부산             | 부산광역시     |
| 대경방송   | 대구             | 대구광역시     |
| 대구방송   | 대구             | 대구광역시     |
| 티씨엔방송 | 대구             | 대구광역시     |
| 남동방송   | 인천             | 인천광역시     |
| 새롬방송   | 인천             | 인천광역시     |
| 서해방송   | 인천             | 인천광역시     |
| 수원방송   | 수원             | 경기도         |
| ABC방송    | 수원             | 경기도         |
| 기남방송   | 수원             | 경기도         |
| 한빛방송   | 광주             | 광주광역시     |
| 전주방송   | 전주             | 전라북도       |
| 중부방송   | 세종             | 세종특별자치시 |
| 세종방송   | 세종             | 세종특별자치시 |

**매핑 근거:**

- STB 테이블의 `SCRBR_SO_NM` (가입자SO명)이 유일한 지역 식별자
- 57개 필드 전수 조사 결과 다른 필드는 지역 정보 없음
- INFO_CNTR_CD는 부적합 (동일 SO에 여러 코드 존재)

---

## 🔍 쿼리 예시

### Grafana 시계열 쿼리

```sql
SELECT
    $__timeGroup(w.datetime, $__interval) AS time,
    m.so_nm,
    avg(w.obs_temperature) AS 기온,
    avg(w.obs_humidity) AS 습도
FROM catv.dist_so_weather_mapping m
LEFT JOIN catv.v_weather_unified w ON
    w.location = m.weather_location
WHERE $__timeFilter(w.datetime)
  AND m.so_nm IN ($SO_LIST)
GROUP BY time, m.so_nm
ORDER BY time
FORMAT JSONCompact;
```

### STB-날씨 JOIN

```sql
SELECT
    s.CM_MAC_ADDR,
    s.SCRBR_SO_NM,
    s.PWR_LVL,
    s.SNR,
    w.obs_temperature AS 기온,
    w.obs_weather_text AS 날씨
FROM catv.stb_monitoring s
LEFT JOIN catv.dist_so_weather_mapping m ON
    s.SCRBR_SO_NM = m.so_nm
LEFT JOIN catv.v_weather_unified w ON
    w.location = m.weather_location
    AND toStartOfHour(s.COLLECT_DATE) = w.datetime
WHERE s.COLLECT_DATE >= now() - INTERVAL 1 HOUR
LIMIT 1000;
```

### 날씨별 STB 장애율

```sql
SELECT
    COALESCE(w.obs_weather_text, w.forecast_3h_weather_text, '알수없음') AS 날씨,
    count(*) AS 측정수,
    countIf(s.PWR_LVL < -10) AS 장애수,
    round(countIf(s.PWR_LVL < -10) / count(*) * 100, 2) AS 장애율
FROM catv.stb_monitoring s
LEFT JOIN catv.dist_so_weather_mapping m ON s.SCRBR_SO_NM = m.so_nm
LEFT JOIN catv.v_weather_unified w ON
    w.location = m.weather_location
    AND toStartOfHour(s.COLLECT_DATE) = w.datetime
WHERE s.COLLECT_DATE >= now() - INTERVAL 24 HOUR
GROUP BY 날씨
ORDER BY 장애율 DESC;
```

---

## 📊 데이터 특성

### 수집 주기 (추정)

| 테이블                 | 주기      | 데이터량/일/지역 |
| ---------------------- | --------- | ---------------- |
| weather_obs            | 매시간    | 24개             |
| weather_forecast_3h    | 3시간마다 | 8개              |
| weather_forecast_daily | 하루 1번  | 1개              |

**총 데이터량 (251 지역 기준):**

- 관측: 24 × 251 = 6,024 rows/일
- 3시간: 8 × 251 = 2,008 rows/일
- 일별: 1 × 251 = 251 rows/일
- **합계**: ~8,300 rows/일

### 디스크 사용량

- **압축 전**: ~50KB/일
- **압축 후 (ZSTD)**: ~15KB/일
- **3년 보관**: ~16MB

---

## 🛠️ 유지보수

### 파티셔닝 전략

```sql
-- 월별 파티션
PARTITION BY toYYYYMM(obs_date)

-- 장점:
-- 1. 월별 DROP 가능
-- 2. 쿼리 최적화 (최근 월만 스캔)
-- 3. 백업/복원 단위

-- 파티션 확인
SELECT
    partition,
    rows,
    bytes_on_disk
FROM system.parts
WHERE table = 'weather_obs'
  AND active
ORDER BY partition DESC;
```

### TTL 설정

```sql
-- 3년 자동 삭제
TTL obs_date + toIntervalYear(3)

-- TTL 상태 확인
SELECT
    table,
    ttl_expr,
    ttl_type
FROM system.tables
WHERE database = 'catv'
  AND table LIKE 'weather_%';
```

### 인덱스 최적화

```sql
-- 현재 ORDER BY: (obs_date, location)
-- 효과:
-- 1. 시간 범위 스캔 최적화
-- 2. location 필터 보조
-- 3. ReplacingMergeTree 중복 제거

-- 쿼리 성능 분석
EXPLAIN indexes = 1
SELECT * FROM catv.weather_obs
WHERE location = '서울'
  AND obs_date >= today() - 7;
```

### ReplacingMergeTree 특징

```sql
-- ReplacingMergeTree는 동일 ORDER BY 키의 중복 데이터 자동 병합
-- ORDER BY (obs_date, location) → 같은 시간/지역의 중복 INSERT 방지

-- FINAL 키워드로 최종 버전만 조회 (VIEW에서 사용)
SELECT * FROM catv.weather_obs FINAL
WHERE obs_date >= today() - 7;
```

---

## 📝 마이그레이션

### 초기 데이터 INSERT

```sql
-- 1. SO 매핑 데이터
INSERT INTO catv.dist_so_weather_mapping
VALUES
    ('강서방송', '서울', '서울특별시'),
    ('광진성동', '서울', '서울특별시'),
    -- ... (26개)
;

-- 2. 날씨 데이터 (파서에서 처리)
-- Go Command에서 자동으로 INSERT
```

### 테이블 생성 순서

```bash
# 1. 날씨 테이블 및 매핑 테이블 생성
clickhouse-client -h 192.168.15.102 --port 30900 \
  < 20260305000000_create_weather.sql

# 생성되는 테이블:
# - catv.so_weather_mapping (local)
# - catv.dist_so_weather_mapping (distributed)
# - catv.weather_obs (local)
# - catv.dist_weather_obs (distributed)
# - catv.weather_forecast_3h (local)
# - catv.dist_weather_forecast_3h (distributed)
# - catv.weather_forecast_daily (local)
# - catv.dist_weather_forecast_daily (distributed)
# - catv.v_weather_unified (view)

# 2. SO 매핑 데이터 INSERT
clickhouse-client -h 192.168.15.102 --port 30900 \
  < 20260305000001_insert_dist_so_weather_mapping.sql
```

### Distributed vs Local 테이블

**Local 테이블** (weather_obs, weather_forecast_3h, weather_forecast_daily):

- 각 샤드에 실제 데이터 저장
- ReplicatedReplacingMergeTree 엔진
- 복제 및 중복 제거 처리

**Distributed 테이블** (dist_weather_obs, dist_weather_forecast_3h, dist_weather_forecast_daily):

- 모든 샤드의 데이터를 통합 조회
- 데이터는 저장하지 않음 (프록시 역할)
- 애플리케이션에서는 dist\_\* 테이블 사용
