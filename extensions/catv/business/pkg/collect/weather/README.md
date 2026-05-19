# Weather Data Collection Package

SKB FTP 서버에서 날씨 데이터를 수집하여 ClickHouse에 저장하는 백그라운드 수집 패키지입니다.

---

## 📁 패키지 구조

```
weather/
├── load.go                           # 서비스 로드/언로드 진입점
├── command/
│   └── command.go                    # Cobra CLI 진입점
├── service/
│   ├── service.go                    # ETL 서비스 (FTP → 파싱 → ClickHouse)
│   └── type.go                       # 파일명 상수 및 관측원 상수
└── tables/
    ├── type.go                       # WeatherTable 인터페이스
    ├── weather_raw_table.go          # 원본 파일 데이터 보존 테이블
    ├── weather_obs_table.go          # 관측 데이터 테이블
    ├── weather_forecast_3h_table.go  # 3시간 예보 테이블
    ├── weather_forecast_daily_table.go # 일별 예보 테이블
    └── mock/                         # 테스트용 Mock 데이터
```

---

## 📋 개요

### ETL 흐름

```
cron 트리거 (CronSpec 설정값)
    ↓
service.etl()
    ├── extract()  – FTP 연결 → 4개 파일 다운로드 → EUC-KR → UTF-8 변환
    ├── transform() – 파일명으로 테이블 선택 (파일명 상수 기반)
    └── load()     – ClickHouse 배치 삽입
```

### 수집 파일 → 저장 테이블 매핑

| FTP 파일 | 저장 테이블 |
|----------|------------|
| `3hour.dat` | `dist_weather_forecast_3h` (3시간 예보) |
| `land.dat` | `dist_weather_forecast_daily` (일별 예보) |
| `shko.dat` | `dist_weather_obs` (관측 SHKO) |
| `shko2.dat` | `dist_weather_obs` (관측 AWS) |

원본 파일 데이터는 `dist_weather_raw` 테이블(파일명 + 원본 텍스트)에 별도 보존됩니다.

### WeatherTable 인터페이스

모든 테이블은 `tables.WeatherTable` 인터페이스를 구현합니다:

- `TableName()` – ClickHouse 테이블명 반환
- `Transform(data string)` – 원본 텍스트 파싱
- `Insert(rows)` – ClickHouse 배치 삽입
- `ConvertBatchFormat(rows)` – 배치 형식 변환
- `GetMockData()` – 시뮬레이션용 데이터 생성

---

## 🚀 사용 방법

```bash
./pharos weather-job --config /etc/pharos/config.toml
```

---

## 🧪 시뮬레이션 모드

FTP 서버 없이 Mock 데이터로 ClickHouse 삽입 동작을 확인할 수 있습니다.

```toml
[catv.collect.weather]
simulation = true
```

각 테이블의 `GetMockData()`로 더미 행을 생성하여 실제 삽입과 동일한 경로로 저장합니다.

---

## ⚙️ 주요 설정

```toml
[catv.collect.weather]
address    = "ftp.example.com"  # FTP 서버 주소
user       = "..."
password   = "..."
cron_spec  = "0 * * * * *"     # 6-필드 cron (초 단위)
simulation = false

[catv.collect.weather.timeout]
connection = "10s"  # FTP 제어 연결 수립 타임아웃
shut       = "30s"  # 파일 전송 완료 후 226 응답 대기 타임아웃
```

### `shut` 타임아웃이 필요한 이유

FTP는 제어 연결(control connection)과 데이터 연결(data connection)을 분리합니다.
대용량 파일 전송 중 제어 연결이 idle 상태로 유지되다가 서버 측 idle timeout에 걸릴 수 있습니다.
`shut` 값은 파일 수신 완료 직후 제어 연결의 deadline을 연장하여 `226 Closing data connection` 응답을 안전하게 수신하게 합니다.

### FTP 동시성 주의사항

`ftp.ServerConn`은 동시 호출이 불가능합니다. 현재 구현은 파일을 순차적으로 다운로드하므로 문제없습니다.
향후 병렬 다운로드가 필요하면 고루틴별 별도 연결을 생성해야 합니다.
