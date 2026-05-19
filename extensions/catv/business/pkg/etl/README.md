# ETL Package

## 개요

ETL(Extract, Transform, Load) 패키지는 Pharos 시스템에서 데이터 처리 및 전송을 담당하는 핵심 모듈입니다. NATS 메시지 브로커를 통해 수신한 데이터를 Elasticsearch로 전송하여 인덱싱하는 기능을 제공합니다.

### 주요 기능

- **NATS 기반 메시지 구독**: NATS 서버로부터 실시간 데이터 스트림 수신
- **Elasticsearch 인덱싱**: 수신한 데이터를 Elasticsearch에 자동 인덱싱
- **확장 가능한 아키텍처**: Extension 시스템을 통한 모듈 확장
- **데몬 모드 지원**: 백그라운드 프로세스로 실행 가능
- **설정 기반 동작**: TOML 설정 파일을 통한 유연한 구성

## 아키텍처

### 주요 컴포넌트

```
etl/
├── command.go          # CLI 명령어 정의 및 실행
├── service.go          # ETL 서비스 라이프사이클 관리
├── load.go            # 설정 및 초기화
├── extension.go       # 확장 모듈 시스템
└── subjects/          # NATS Subject 핸들러
    └── elasticsearch_index.go  # Elasticsearch 인덱싱 핸들러
```

### 데이터 흐름

```
┌─────────────┐      NATS      ┌─────────────┐    Elasticsearch    ┌──────────────┐
│   Source    │ ─────────────> │ ETL Service │ ─────────────────> │ Elasticsearch│
│  (ClickHouse│   TabSeparated │             │    Bulk API        │              │
│   NATS MV)  │                └─────────────┘                    └──────────────┘
└─────────────┘
                     Subject: elasticsearch.index
                     Format: index\tdocument_id\tdocument
```

1. **데이터 수신**: NATS Subject `elasticsearch.index`에서 TabSeparated 형식의 데이터 수신
2. **데이터 파싱**: `index`, `document_id`, `document` 필드 추출
3. **데이터 검증**: Elasticsearch 인덱스 이름 규칙 검증
4. **데이터 인덱싱**: Elasticsearch Bulk API를 사용한 효율적인 인덱싱

## 시작하기

### 설정 파일

ETL 서비스는 TOML 형식의 설정 파일을 사용합니다:

```toml
[etl.elasticsearch]
version = 9
addresses = ["http://192.168.15.105:9200"]
[etl.elasticsearch.bulk]
flush_bytes = 5000000
flush_interval = "30s"
```

### CLI 사용법

#### 기본 실행
```bash
# 기본 설정 파일 사용
./pharos etl

# 커스텀 설정 파일 지정
./pharos etl -c /path/to/config.toml

# 여러 설정 파일 병합
./pharos etl -c config1.toml -c config2.toml
```

#### 데몬 모드 실행
```bash
# 데몬으로 실행
./pharos etl -d

# PID 파일 및 로그 파일 지정
./pharos etl -d --pid-file /var/run/etl.pid --log-file /var/log/etl.log
```

#### Systemd 서비스 등록
```ini
[Unit]
Description=Pharos ETL Service
After=network.target nats.service elasticsearch.service

[Service]
Type=simple
ExecStart=/usr/local/bin/pharos etl -c /etc/pharos/config.toml
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

## ClickHouse NATS 엔진 연동

### NATS 엔진 테이블 개요

ClickHouse의 NATS 엔진은 NATS 메시지 브로커로 데이터를 발행(publish)하는 특수 테이블 엔진입니다. 이 테이블에 INSERT된 데이터는 자동으로 지정된 NATS Subject로 전송됩니다.

#### 중요 사항

⚠️ **마이그레이션 제약사항**
- NATS 엔진 테이블 생성 시점에 NATS 서버 연결이 필수
- 연결 실패 시 테이블 생성 자체가 실패
- 현재 마이그레이션은 서버 실행 전에 수행되므로 마이그레이션으로 생성 불가
- **수동으로 별도 생성 필요**

### NATS 엔진 테이블 생성

```sql
CREATE TABLE nats
(
    `index` String,
    `document_id` String,
    `document` String
)
ENGINE = NATS
SETTINGS 
    nats_url = 'nats://192.168.15.103:4222',
    nats_subjects = 'elasticsearch.index',
    nats_format = 'TabSeparated'
```

#### 파라미터 설명

| 파라미터 | 설명 | 예시 |
|---------|------|------|
| `nats_url` | NATS 서버 주소 | `nats://192.168.15.103:4222` |
| `nats_subjects` | 발행할 Subject 이름 | `elasticsearch.index` |
| `nats_format` | 데이터 포맷 | `TabSeparated`, `JSONEachRow` 등 |

#### 컬럼 구조

- `index`: Elasticsearch 인덱스 이름
- `document_id`: 문서 고유 ID (UUID)
- `document`: JSON 형식의 문서 본문

### MATERIALIZED VIEW를 통한 자동 전송

MATERIALIZED VIEW를 사용하면 원본 테이블의 데이터를 NATS 테이블로 자동 전송할 수 있습니다.

#### 기본 예제

```sql
CREATE MATERIALIZED VIEW qmvoc_nats_mv TO nats
(
    `index` String,
    `document_id` String,
    `document` String
)
AS SELECT 
    'qmvoc-from-nats' AS index,
    generateUUIDv4() AS document_id,
    toJSONString(map(
        '@timestamp', toString(`@timestamp`), 
        'CELL_NO', toString(CELL_NO), 
        'CMPL_CNS_NM', toString(CMPL_CNS_NM)
        -- ... 추가 필드
    )) AS document 
FROM qmvoc_queue
```

#### 고급 패턴 - 동적 인덱스 생성

날짜별 또는 조건별로 다른 인덱스에 데이터를 분산할 수 있습니다:

```sql
CREATE MATERIALIZED VIEW metrics_nats_mv TO nats
AS SELECT 
    -- 날짜별 인덱스 생성
    concat('metrics-', formatDateTime(`@timestamp`, '%Y.%m.%d')) AS index,
    generateUUIDv4() AS document_id,
    toJSONString(map(
        '@timestamp', toString(`@timestamp`),
        'metric_name', toString(metric_name),
        'value', toString(value),
        'host', toString(host)
    )) AS document
FROM metrics_queue
```

#### 실전 예제 - CATV 품질 데이터

```sql
-- 1. NATS 엔진 테이블 생성
CREATE TABLE catv_nats
(
    `index` String,
    `document_id` String,
    `document` String
)
ENGINE = NATS
SETTINGS 
    nats_url = 'nats://nats.example.com:4222',
    nats_subjects = 'elasticsearch.index',
    nats_format = 'TabSeparated';

-- 2. MATERIALIZED VIEW 생성
CREATE MATERIALIZED VIEW qmvoc_nats_mv TO catv_nats
AS SELECT 
    'qmvoc-' || formatDateTime(now(), '%Y.%m') AS index,  -- 월별 인덱스
    generateUUIDv4() AS document_id,
    toJSONString(map(
        '@timestamp', toString(`@timestamp`), 
        'CELL_NO', toString(CELL_NO), 
        'CMPL_CNS_NM', toString(CMPL_CNS_NM), 
        'CNS_ACPN_DT', toString(CNS_ACPN_DT), 
        'CNS_CONT', toString(CNS_CONT), 
        'CNS_FINI_NM', toString(CNS_FINI_NM), 
        'ICENTER', toString(ICENTER), 
        'IP_CA', toString(IP_CA), 
        'ISU_ORG_NM', toString(ISU_ORG_NM), 
        'JURS_TPO_NM', toString(JURS_TPO_NM), 
        'KPI', toString(KPI), 
        'KUC_SA_NM', toString(KUC_SA_NM), 
        'LCL_NM', toString(LCL_NM), 
        'LOCALCENTER', toString(LOCALCENTER), 
        'MCL_NM', toString(MCL_NM), 
        'SCL_NM', toString(SCL_NM), 
        'STB_MAC_NO', toString(STB_MAC_NO), 
        'STB_MDL_NM_KOR', toString(STB_MDL_NM_KOR), 
        'STB_MDL_NM_KOR2', toString(STB_MDL_NM_KOR2), 
        'STRT_DT', toString(STRT_DT), 
        'SUP_ADDR', toString(SUP_ADDR), 
        'SVC_CD', toString(SVC_CD), 
        'SVC_DTL_CL_CD', toString(SVC_DTL_CL_CD), 
        'SVC_DTL_CL_NM', toString(SVC_DTL_CL_NM), 
        'SVC_NM', toString(SVC_NM), 
        'TEAM', toString(TEAM), 
        'so_code', toString(so_code), 
        'so_name', toString(so_name), 
        'stb_mac_no2', toString(stb_mac_no2)
    )) AS document 
FROM qmvoc_queue;
```

### 데이터 흐름 전체 구조

```
┌──────────────┐    INSERT     ┌──────────────┐    Auto-Insert    ┌────────────┐
│   Source     │ ──────────> │ qmvoc_queue  │ ─────────────────> │ NATS Table │
│   Table      │               │  (MergeTree) │  (Materialized   │  (Engine)  │
└──────────────┘               └──────────────┘      View)         └────────────┘
                                                                         │
                                                                         │ Publish
                                                                         v
                                                                   ┌────────────┐
                                                                   │    NATS    │
                                                                   │   Server   │
                                                                   └────────────┘
                                                                         │
                                                                         │ Subscribe
                                                                         v
                                                                   ┌────────────┐
                                                                   │ ETL Service│
                                                                   └────────────┘
                                                                         │
                                                                         │ Index
                                                                         v
                                                                   ┌────────────┐
                                                                   │Elasticsearch│
                                                                   └────────────┘
```

### 모니터링 및 문제 해결

#### NATS 테이블 상태 확인

```sql
-- 테이블 존재 여부 확인
SHOW TABLES LIKE 'nats%';

-- 테이블 정의 확인
SHOW CREATE TABLE nats;
```

#### 일반적인 문제 해결

1. **테이블 생성 실패**
   ```
   Error: Cannot connect to NATS server
   ```
   - NATS 서버가 실행 중인지 확인
   - 네트워크 연결 및 방화벽 설정 확인
   - `nats_url` 파라미터 정확성 검증

2. **데이터가 전송되지 않음**
   ```sql
   -- MATERIALIZED VIEW 확인
   SELECT * FROM system.tables 
   WHERE name LIKE '%_nats_mv';
   
   -- 소스 테이블에 데이터가 있는지 확인
   SELECT count() FROM qmvoc_queue;
   ```

3. **성능 최적화**
   - Batch insert 사용
   - MATERIALIZED VIEW에 WHERE 절로 필터링
   - TTL 설정으로 오래된 데이터 자동 삭제

## Extension 시스템

ETL 패키지는 확장 가능한 모듈 시스템을 제공합니다.

### Extension 인터페이스

```go
type Extension interface {
    Name() string
    Load(config common.Config) error
    Unload()
}
```

### Extension 등록 예제

```go
package myextension

import "ntels.com/pharos/core/pkg/etl"

type MyExtension struct{}

func (e *MyExtension) Name() string {
    return "my-extension"
}

func (e *MyExtension) Load(config common.Config) error {
    // 초기화 로직
    return nil
}

func (e *MyExtension) Unload() {
    // 정리 로직
}

func init() {
    etl.Register(&MyExtension{})
}
```

## API 참조

### Service

```go
// NewService creates a new ETL service instance
func NewService(config common.Config) (*Service, error)

// Start starts the ETL service
func (s *Service) Start() error

// Stop stops the ETL service
func (s *Service) Stop() error
```

### Elasticsearch Index Subject

Subject Key: `elasticsearch.index`

#### 메시지 포맷
```
TabSeparated 형식:
index\tdocument_id\tdocument

예시:
logs-2024.12.12\t550e8400-e29b-41d4-a716-446655440000\t{"@timestamp":"2024-12-12T10:00:00Z","message":"test"}
```

#### 인덱스 이름 규칙
- 길이: 1-255자
- 허용 문자: 소문자(a-z), 숫자(0-9), 하이픈(-), 언더스코어(_), 플러스(+)
- 시작 문자: 소문자 또는 숫자 (특수문자로 시작 불가)
- 예시:
  - ✅ `logs-2024.12.12`
  - ✅ `metrics_prod`
  - ✅ `app01+data`
  - ❌ `-invalid` (하이픈으로 시작)
  - ❌ `Invalid` (대문자 포함)
  - ❌ `index@name` (허용되지 않는 문자)

## 테스트

### 테스트 실행

```bash
# 전체 테스트
go test -v ./...

# 커버리지 리포트
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Race condition 검사
go test -race ./...
```

### 테스트 구조

- `command_test.go`: CLI 명령어 테스트
- `service_test.go`: 서비스 라이프사이클 테스트
- `extension_test.go`: Extension 시스템 테스트
- `load_test.go`: 설정 로드 테스트
- `subjects/elasticsearch_index_test.go`: Elasticsearch 인덱싱 테스트

자세한 내용은 [TESTING.md](TESTING.md)를 참조하세요.

## 모범 사례

### 1. 설정 관리
- 환경별로 별도 설정 파일 유지
- 민감한 정보는 환경 변수 사용 고려
- 설정 파일은 버전 관리에서 제외 (.gitignore)

### 2. 에러 핸들링
- 재시도 로직 구현 (Elasticsearch 연결 실패 등)
- 적절한 로깅으로 문제 추적
- Circuit breaker 패턴 고려

### 3. 성능 최적화
- Elasticsearch Bulk API 활용
- 적절한 버퍼 크기 설정
- NATS 연결 풀 관리

### 4. 모니터링
- 처리량 메트릭 수집
- 에러율 모니터링
- NATS 및 Elasticsearch 연결 상태 확인

## 참고 자료

### 관련 문서
- [NATS 가이드](../../../../docs/NATS_가이드.md)
- [마이그레이션 가이드](../migration/README.md)
- [플러그인 개발 가이드](../plugins/README.md)

### 외부 링크
- [NATS Documentation](https://docs.nats.io/)
- [Elasticsearch Documentation](https://www.elastic.co/guide/index.html)
- [ClickHouse NATS Engine](https://clickhouse.com/docs/en/engines/table-engines/integrations/nats)
- [ClickHouse Materialized Views](https://clickhouse.com/docs/en/guides/developer/cascading-materialized-views)

## 라이선스

이 프로젝트는 회사 내부 사용을 위한 것입니다.

## 기여

개선 사항이나 버그 발견 시 이슈를 등록하거나 Pull Request를 제출해주세요.
