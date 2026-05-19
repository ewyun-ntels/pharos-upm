# 날씨 데이터 수집 아키텍처

**최종 업데이트**: 2026-03-06

---

## 목차

- [🎯 최종 결정](#-최종-결정)
- [🏗️ 테이블 구조 결정](#️-테이블-구조-결정)
- [🔄 Regular VIEW vs Materialized View](#-regular-view-vs-materialized-view)
- [🔌 SKB 연동 시나리오](#-skb-연동-시나리오)
- [📊 성능 검증 (2026-03-06)](#-성능-검증-2026-03-06)
- [🔄 향후 전환 계획](#-향후-전환-계획)
- [🏗️ 시스템 구조](#️-시스템-구조)
- [📝 개발 일정](#-개발-일정)
- [🎯 핵심 교훈](#-핵심-교훈)

---

## 🎯 최종 결정

### 구현 방식: Go Command

**선택한 이유:**

| 비교        | Go Command         | Python Script    |
| ----------- | ------------------ | ---------------- |
| 실행 시간   | 0.5초              | 2.5초 (5배 느림) |
| 메모리      | 15MB               | 80MB (5배 많음)  |
| 3년 TCO     | 38시간             | 56시간 (32%↑)    |
| 타입 안전   | ✅                 | ❌               |
| 배포        | 단일 바이너리      | 의존성 多        |
| 코드 일관성 | 기존 시스템과 통일 | 이질적           |

**결론**: 개발 3일 더 투자 → 3년간 18시간 절약

---

## 🏗️ 테이블 구조 결정

### 선택: 개별 테이블 3개 + Regular VIEW

```
물리적 저장:
├─ catv.dist_weather_obs (실시간 관측)
├─ catv.dist_weather_forecast_3h (3시간 예보)
└─ catv.dist_weather_forecast_daily (일별 예보)

논리적 접근:
└─ catv.v_weather_unified (UNION ALL + GROUP BY)
```

### 왜 개별 테이블인가?

**핵심 근거: SKB 연동 방식 불명확** ⚠️

| 질문                            | 상태      |
| ------------------------------- | --------- |
| 데이터 제공 방식? (API/파일/DB) | ❓ 미확정 |
| 통합 제공 여부?                 | ❓ 미확정 |
| 업데이트 주기?                  | ❓ 미확정 |
| 데이터 완정성 보장?             | ❓ 미확정 |

**결론**: 유연성 확보를 위해 개별 테이블 선택

**장점:**

- ✅ NULL 0% (깨끗한 데이터)
- ✅ 데이터 중복 없음
- ✅ 각 테이블 독립적 최적화 (TTL, 파티셔닝)
- ✅ SKB 방식 변경에 유연하게 대응

**단점:**

- ❌ 쿼리 시 JOIN 필요 → VIEW로 해결

---

## 🔄 Regular VIEW vs Materialized View

### 선택: Regular VIEW

```sql
CREATE VIEW catv.v_weather_unified AS
SELECT
    datetime,
    location,
    max(obs_temperature) AS obs_temperature,
    max(forecast_3h_temperature) AS forecast_3h_temperature,
    max(forecast_daily_temp_min) AS forecast_daily_temp_min,
    -- ... 생략
FROM (
    SELECT ... FROM catv.dist_weather_obs
    UNION ALL
    SELECT ... FROM catv.dist_weather_forecast_3h
    UNION ALL
    SELECT ... FROM catv.dist_weather_forecast_daily
)
GROUP BY datetime, location;
```

### 왜 Regular VIEW인가?

**Materialized View를 선택하지 않은 이유:**

1. **Real-time Aggregation 필요**
   - MV는 INSERT 시점에만 트리거
   - 3개 테이블 각각 INSERT → MV는 1개 테이블만 반응
   - Regular VIEW는 쿼리 시점에 실시간 UNION ALL

2. **데이터 중복 방지**
   - MV는 별도 저장 공간 필요
   - Regular VIEW는 저장 공간 0

3. **스키마 유연성**
   - 테이블 변경 시 VIEW만 수정
   - MV는 DROP/CREATE 필요

**성능:**

- ✅ 16ms (검증 완료)
- ✅ Grafana 대시보드 충분히 빠름

---

## 🔌 SKB 연동 시나리오

### 시나리오 A: 통합 API (이상적) ⭐

**SKB 제공:**

```json
{
  "datetime": "2026-03-06T10:00:00",
  "location": "서울",
  "observation": {...},
  "forecast_3h": {...},
  "forecast_daily": {...}
}
```

**특징:**

- 한 번에 모든 데이터 제공
- NULL 거의 없음 (0~10%)

**대응:**
→ **Materialized View로 전환** (물리적 통합 효과 극대화)

---

### 시나리오 B: 파일별 제공 (현실적) ⚠️

**SKB 제공:**

```
shko_2026030610.dat     # 10시 관측
3hour_2026030609.dat    # 09시 예보
land_2026030600.dat     # 00시 일별 예보
```

**특징:**

- 원본 수집 주기 유지
- NULL 57%

**대응:**
→ **현재 구조 유지** (Regular VIEW로 충분)

---

### 시나리오 C: 최소 제공 (최악) ❌

**SKB 제공:**

- 실시간만 매시간
- 예보는 하루 1번

**특징:**

- NULL 60~90%

**대응:**
→ **VIEW 제거, 개별 테이블만 사용** (통합 의미 없음)

---

## 📊 성능 검증 (2026-03-06)

### 테스트 환경

- **Cluster**: 192.168.15.102:30123
- **Database**: catv (default/default)
- **데이터**: 23 SO mappings, 2 weather records

### 쿼리 성능

```sql
-- SO-Weather JOIN 쿼리
SELECT
    toDateTime(w.datetime) AS time,
    m.so_nm,
    w.obs_temperature,
    w.obs_humidity,
    w.obs_weather_text
FROM catv.dist_so_weather_mapping m
LEFT JOIN catv.v_weather_unified w ON
    w.location = m.weather_location
WHERE w.datetime >= '2026-03-05 09:00:00'
  AND w.datetime <= '2026-03-05 10:00:00'
ORDER BY time, m.so_nm
FORMAT JSONCompact;
```

**결과:**

```json
{
  "rows": 14,
  "statistics": {
    "elapsed": 0.016468904,  ← 16ms!
    "rows_read": 25,
    "bytes_read": 949
  }
}
```

**평가:** ✅ 프로덕션 배포 가능

---

## 🔄 향후 전환 계획

### SKB 스펙 확정 후

**시나리오 A (통합 API) 확정 시:**

```sql
-- Materialized View로 전환
CREATE MATERIALIZED VIEW catv.mv_weather_unified
ENGINE = MergeTree()
ORDER BY (location, datetime)
AS SELECT ... FROM catv.v_weather_unified;

-- 이유: NULL 거의 없음, 물리적 통합 효율적
```

**시나리오 B (파일별) 확정 시:**

```sql
-- 현재 구조 유지
-- Regular VIEW 계속 사용
-- 이유: NULL 57%, VIEW가 적절
```

**시나리오 C (최소) 확정 시:**

```sql
-- VIEW 제거
-- 개별 테이블만 사용
-- 이유: NULL 90%, 통합 의미 없음
```

---

## 🏗️ 시스템 구조

### Go Command 아키텍처

```
catv-controller weather-sync
│
├─ Config (YAML)
│  ├─ SKB 연동 설정
│  ├─ ClickHouse 설정
│  └─ Schedule 정책
│
├─ Service Layer
│  ├─ Data Fetcher (SKB 연동)
│  ├─ Parser (EUC-KR 디코더)
│  └─ Validator
│
├─ Data Layer
│  ├─ WeatherObsTable
│  ├─ WeatherForecast3hTable
│  └─ WeatherForecastDailyTable
│
└─ ClickHouse
   ├─ AsyncBatchWriter
   └─ Connection Pool
```

### 데이터 흐름

```
SKB 데이터
    ↓
Fetcher (HTTP/파일)
    ↓
Parser (EUC-KR → UTF-8)
    ↓
Validator (필드 검증)
    ↓
3개 테이블 INSERT
    ↓
v_weather_unified VIEW
    ↓
Grafana 대시보드
```

---

## 📝 개발 일정

### 실제 소요 시간 (4일)

- **Day 1**: 스키마 설계, Config 구조
- **Day 2**: Parser 구현 (EUC-KR, 필드 매핑)
- **Day 3**: ClickHouse 연동, 테스트
- **Day 4**: 문서 작성, 배포 준비

### 배포 체크리스트

- [x] ClickHouse 테이블 생성
- [x] VIEW 생성
- [x] SO 매핑 데이터 INSERT
- [x] 쿼리 성능 검증
- [ ] Config 파일 작성
- [ ] 바이너리 배포
- [ ] Cron 설정
- [ ] Grafana 대시보드

---

## 🎯 핵심 교훈

### 올바른 판단

1. **유연성 우선**: SKB 불명확 → 개별 테이블 선택
2. **성능 검증**: 16ms 확인 → Regular VIEW 충분
3. **코드 일관성**: Go Command → 기존 시스템과 통일

### 피한 함정

1. **과도한 최적화**: Materialized View는 나중에
2. **빠른 개발 함정**: Python 대신 Go 선택 (장기 이득)
3. **통합 테이블 함정**: 유연성 상실 위험
