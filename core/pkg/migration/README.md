# Pharos Migration System

Pharos의 마이그레이션 시스템은 **3단계 구조**로 되어 있습니다.

## 📋 목차

- [개요](#개요)
- [디렉토리 구조](#디렉토리-구조)
- [실행 흐름](#실행-흐름)
- [각 단계 설명](#각-단계-설명)

## 개요

```
애플리케이션 시작
    ↓
┌─────────────────────────────────────────┐
│ 1. Schema Migrations (schema/)         │  ← DDL: 테이블 생성
│    실행 시점: 앱 시작 전                 │
│    테이블: goose_db_version              │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ 2. Data Migrations (data/)             │  ← DML: 데이터 변환
│    실행 시점: 앱 시작 후                 │
│    테이블: pharos_data_migrations        │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ 3. API Migrations (api/)               │  ← REST: 초기 설정
│    실행 시점: 서버 시작 후 (백그라운드)  │
│    테이블: core_app_data_*               │
└─────────────────────────────────────────┘
    ↓
애플리케이션 정상 실행
```

## 디렉토리 구조

```
pkg/migration/
├── load.go                 # 진입점 (schema + data 실행)
├── schema/                 # 1️⃣ 스키마 마이그레이션 (DDL)
│   ├── README.md          # 상세 가이드
│   ├── databases.go       # 실행 로직
│   └── agent/master/      # 제품별/드라이버별 SQL 파일
├── data/                   # 2️⃣ 데이터 마이그레이션 (DML)
│   ├── README.md          # 상세 가이드
│   ├── runner.go          # 실행 로직
│   └── *.go               # Goose 마이그레이션 파일
└── api/                    # 3️⃣ API 마이그레이션 (REST)
    ├── README.md          # 상세 가이드
    └── runner.go          # apimigrate 실행기
```

## 실행 흐름

### 1. 자동 실행 (main.go)

```go
import "ntels.com/pharos/core/pkg/migration"

func main() {
    // 1️⃣ Schema + 2️⃣ Data 마이그레이션 (동기)
    if err := migration.Load(config); err != nil {
        log.Fatal(err)  // 실패 시 앱 종료
    }
    
    // 서버 시작
    server.Run()
    
    // 3️⃣ API 마이그레이션 (비동기, 별도 처리)
}
```

### 2. 내부 실행 순서

```go
// load.go
func Load(cfg common.Config) error {
    // 1️⃣ Schema: 테이블 생성 (필수)
    if err := schema.Up(config); err != nil {
        return err  // 실패 시 중단
    }

    // 2️⃣ Data: 데이터 변환
    if err := data.Run(config); err != nil {
        return err  // 실패 시 중단
    }

    return nil
}
```

## 각 단계 설명

### 1️⃣ Schema Migrations (schema/)

**목적**: 데이터베이스 테이블 구조 정의

- **실행 시점**: 앱 시작 **전** (필수)
- **작업**: CREATE TABLE, ALTER TABLE, DROP TABLE
- **파일 형식**: SQL 또는 Go
- **테이블**: `goose_db_version`

👉 [자세히 보기](schema/README.md)

**예시**:
```sql
-- 20250117000000_create_users.sql
CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);
```

### 2️⃣ Data Migrations (data/)

**목적**: 데이터 구조 변환 및 업데이트

- **실행 시점**: 앱 시작 **후**
- **작업**: UPDATE, JSON 변환, 데이터 정규화
- **파일 형식**: Go only
- **테이블**: `pharos_data_migrations`

👉 [자세히 보기](data/README.md)

**예시**:
```go
// 20250117000000_filter_migration.go
func up(ctx context.Context, tx *sql.Tx) error {
    // dashboard.config JSON 구조 변경
    // filters.step 필드 추가
    return nil
}
```

### 3️⃣ API Migrations (api/)

**목적**: REST API 호출로 초기 설정

- **실행 시점**: 서버 시작 **후** (백그라운드)
- **작업**: 사이트별 설정, 기본 데이터 투입
- **파일 형식**: YAML (apimigrate)
- **테이블**: `core_app_data_schema_migrations`, `core_app_data_migration_runs`

👉 [자세히 보기](api/README.md)

**예시**:
```yaml
# 0001_initial_setup.yaml
steps:
  - name: "Create default organization"
    method: POST
    url: "{{api_base}}/api/v1/organizations"
    body: '{"name": "Default"}'
```

## 빠른 시작

### 새 Schema 마이그레이션 추가

```bash
# 파일 생성
cd core/pkg/migration/schema/master/base/postgresql
touch 20250117120000_add_column.sql

# 내용 작성
cat > 20250117120000_add_column.sql << 'EOF'
-- +goose Up
ALTER TABLE users ADD COLUMN email TEXT;

-- +goose Down
ALTER TABLE users DROP COLUMN email;
EOF

# 실행 (앱 재시작 시 자동)
./pharos
```

### 새 Data 마이그레이션 추가

```bash
# 파일 생성
cd core/pkg/migration/data
touch 20250117120000_update_config.go

# 템플릿 복사
cp 20250117000000_filter_and_panel_migration.go 20250117120000_update_config.go

# up/down 함수 작성 후 실행 (앱 재시작 시 자동)
./pharos
```

### 새 API 마이그레이션 추가

```bash
# YAML 파일 생성
mkdir -p /opt/pharos/migrations
cat > /opt/pharos/migrations/0001_setup.yaml << 'EOF'
version: 1
steps:
  - name: "Setup"
    method: POST
    url: "{{api_base}}/api/v1/setup"
EOF

# config 설정
vim config/master_config.toml
# [migration]
# migration_dir = "/opt/pharos/migrations"

# 실행 (서버 시작 후 자동)
./pharos
```

## 주의사항

⚠️ **Schema는 필수**
- Schema 마이그레이션이 실패하면 앱이 시작되지 않음
- 테이블이 없으면 Data 마이그레이션도 실패

⚠️ **실행 순서 중요**
- Schema → Data → API 순서로 실행
- Data는 Schema가 완료된 후에만 실행됨
- API는 서버가 준비된 후 비동기 실행

⚠️ **버전 관리**
- Schema/Data는 Goose 타임스탬프 기반 (YYYYMMDDHHMMSS)
- 한 번 실행된 마이그레이션은 다시 실행되지 않음
- 롤백은 신중하게 (데이터 손실 가능)

## 참고 문서

- [Migration Architecture 전체 설명](../../../../docs/MIGRATION_ARCHITECTURE.md)
- [Schema Migrations 가이드](schema/README.md)
- [Data Migrations 가이드](data/README.md)
- [API Migrations 가이드](api/README.md)

## 문제 해결

### 마이그레이션 상태 확인

```sql
-- Schema 마이그레이션 이력
SELECT * FROM goose_db_version ORDER BY id DESC;

-- Data 마이그레이션 이력
SELECT * FROM pharos_data_migrations ORDER BY id DESC;

-- API 마이그레이션 이력
SELECT * FROM core_app_data_schema_migrations ORDER BY version DESC;
```

### 특정 버전으로 롤백 (주의!)

```bash
# Schema 롤백 (goose CLI 사용)
goose -dir pkg/migration/schema/master/base/postgresql \
      postgres "..." down

# Data 롤백 (수동)
# down 함수를 직접 실행하거나 백업에서 복구
```

### 마이그레이션 비활성화

```toml
# config/master_config.toml

# API 마이그레이션만 비활성화
[migration]
disable = true

# Schema/Data는 코드에서 migration.Load() 호출하지 않으면 됨
```
