# 날씨 데이터 수집 시스템

> CATV 품질 모니터링에 날씨 데이터를 통합

**최종 업데이트**: 2026-03-19  
**상태**: 🟡 배포 준비 완료 (SKB FTP 주소 확인됨 · 인증정보/파일 경로 확인 필요)

---

## 목차

- [🎯 개요](#-개요)
- [📊 최종 아키텍처](#-최종-아키텍처)
- [🧪 테스트 결과 (2026-03-06)](#-테스트-결과-2026-03-06)
- [📁 문서 구조](#-문서-구조)
- [🚀 사용법](#-사용법)
- [🔥 긴급 TODO](#-긴급-todo)
- [📊 통계](#-통계)
- [📞 문의](#-문의)

---

## 🎯 개요

날씨 데이터를 주기적으로 수집하여 ClickHouse에 저장하고, STB 품질과의 상관관계를 분석합니다.

### 빠른 정보

| 항목            | 내용                                             |
| --------------- | ------------------------------------------------ |
| **구현 방식**   | Go Command (Cobra)                               |
| **테이블 구조** | 개별 3개 + 통합 VIEW                             |
| **데이터 소스** | SKB FTP (주소 확인됨, 인증정보/파일 경로 미확정) |
| **매핑 단위**   | SO 단위 (23개)                                   |
| **쿼리 성능**   | 16ms (검증 완료)                                 |

---

## 📊 최종 아키텍처

### 테이블 구조

```
물리적 테이블 (3개):
├─ catv.dist_weather_obs (실시간 관측)
├─ catv.dist_weather_forecast_3h (3시간 예보)
└─ catv.dist_weather_forecast_daily (일별 예보)

논리적 통합:
└─ catv.v_weather_unified (Grafana용 VIEW)

매핑 테이블:
└─ catv.dist_so_weather_mapping (23 SO → 8 지역)
```

### 핵심 결정

**개별 테이블을 선택한 이유:**

- ✅ SKB FTP 방식 확정, 세부 스펙(인증정보/파일 경로) 확인 중 → 유연성 유지
- ✅ 데이터 중복 없음 (디스크 효율)
- ✅ NULL 없는 깨끗한 구조
- ✅ 향후 Materialized View 전환 가능

**Regular VIEW를 선택한 이유:**

- ✅ Real-time aggregation (UNION ALL + GROUP BY)
- ✅ 16ms 성능 (프로덕션 준비 완료)
- ✅ Grafana 대시보드 최적화
- ✅ 3개 테이블 독립적 업데이트 가능

---

## 🧪 테스트 결과 (2026-03-06)

### 성능 검증

```bash
# Cluster: 192.168.15.102:30123
# Query: SO-Weather JOIN with 23 mappings

{
  "rows": 14,
  "statistics": {
    "elapsed": 0.016468904,  ← 16ms!
    "rows_read": 25,
    "bytes_read": 949
  }
}
```

### 검증 항목

- ✅ 쿼리 속도: 16ms (매우 빠름)
- ✅ SO 매핑: 23개 정상 작동
- ✅ JOIN 성능: 검증 완료
- ✅ Grafana JSON: JSONCompact 지원
- ✅ NULL 처리: max() 함수로 자동 처리

---

## 📁 문서 구조

### 핵심 문서 (3개)

| 문서                                     | 내용                   | 대상                 |
| ---------------------------------------- | ---------------------- | -------------------- |
| **[ARCHITECTURE.md](./ARCHITECTURE.md)** | 아키텍처 결정 및 구현  | 개발자, DBA          |
| **[SCHEMA.md](./SCHEMA.md)**             | DB 스키마 및 매핑 전략 | DBA, 데이터 엔지니어 |
| [DATA_ANALYSIS.md](./DATA_ANALYSIS.md)   | 원본 데이터 분석       | 파서 개발자          |

### 실제 코드 문서

- [../../../pkg/weather/command/README.md](../../../pkg/weather/command/README.md) - 구현 코드 문서
- [../../../pkg/weather/command/](../../../pkg/weather/command/) - 실제 Go 코드

---

## 🚀 사용법

### 명령어 실행

```bash
# 단일 실행
./catv-controller weather-sync \
  --config config.yaml

# Daemon 모드
./catv-controller weather-sync \
  --config config.yaml \
  --daemon \
  --pid-file /var/run/weather-sync.pid
```

### Cron 설정

```bash
# 매시간 실행
0 * * * * /opt/pharos/bin/catv-controller weather-sync --config /etc/pharos/weather.yaml

# 3시간마다 실행
0 */3 * * * /opt/pharos/bin/catv-controller weather-sync --config /etc/pharos/weather.yaml
```

### Grafana 쿼리 예시

```sql
-- SO별 날씨 현황
SELECT
    toDateTime(w.datetime) AS time,
    m.so_nm,
    w.obs_temperature AS 기온,
    w.obs_humidity AS 습도,
    w.obs_weather_text AS 날씨
FROM catv.dist_so_weather_mapping m
LEFT JOIN catv.v_weather_unified w ON
    w.location = m.weather_location
WHERE $__timeFilter(w.datetime)
  AND m.so_nm = '$SO'
ORDER BY time
FORMAT JSONCompact;
```

---

## 🔥 긴급 TODO

### SKB 세부 스펙 확인 필요

```
[x] SKB 연동 방식: FTP 확정
[x] FTP 서버 주소: 확인 완료

[ ] 추가 확인 필요
    [ ] FTP 유저명
    [ ] FTP 패스워드
    [ ] 파일 디렉토리 경로
    [ ] 파일 포맷 및 네이밍 규칙
    [ ] 데이터 업데이트 주기
    [ ] 샘플 파일 요청

[ ] 스펙 확정 후
    [ ] Materialized View 전환 검토
    [ ] 또는 현재 구조 유지
```

**시나리오별 대응 (FTP 방식 확정 기준):**

| 시나리오     | 설명                     | 대응                          |
| ------------ | ------------------------ | ----------------------------- |
| A: 통합 파일 | 한 파일에 전체 데이터    | Materialized View 전환 가능   |
| B: 파일별    | 관측/예보 파일 분리 제공 | 현재 구조 유지 (NULL 57%)     |
| C: 부분 제공 | 일부 데이터만 포함       | VIEW 제거, 개별 테이블만 사용 |

---

## 📊 통계

- **SO 개수**: 23개 (서울 7, 부산 3, 대구 3, 인천 3, 수원 3, 기타 4)
- **날씨 지역**: 8개 (서울, 부산, 대구, 인천, 수원, 광주, 전주, 세종)
- **테이블**: 4개 (3개 날씨 + 1개 매핑)
- **쿼리 성능**: 16ms

---

## 📞 문의

- **Backend Team**: Go 구현
- **DBA Team**: 스키마 관리
- **SKB 연동 담당**: FTP 인증정보 및 파일 경로 확인 필요 🔥
