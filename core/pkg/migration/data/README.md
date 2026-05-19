# Data Migrations

## 개요

**데이터 마이그레이션**은 애플리케이션이 실행된 후 기존 데이터의 구조를 변환하는 마이그레이션입니다.

### 스키마 vs 데이터 마이그레이션

| 구분 | 스키마 마이그레이션 | 데이터 마이그레이션 |
|------|-------------------|-------------------|
| 위치 | `pkg/migration/schema/` | `pkg/migration/data/` |
| 실행 시점 | **앱 시작 전** (테이블 생성) | **앱 시작 후** (데이터 변환) |
| 테이블명 | `core_db_version` | `core_data_version` |
| 도구 | Goose (embed.FS) | Custom (plugin-style) |
| 목적 | DDL (CREATE TABLE, ALTER TABLE) | DML (UPDATE, data transformation) |
| 예시 | dashboard 테이블 생성 | dashboard.config JSON 구조 변경 |

## 특징

### ✅ 커스텀 플러그인 시스템
- **파일명 기반 버전 자동 추출**: `YYYYMMDDHHMMSS_description.go` → Version 자동 감지
- **자동 실행**: 실행되지 않은 마이그레이션만 자동 실행
- **트랜잭션 안전**: 각 마이그레이션은 트랜잭션 내에서 실행
- **멱등성**: 여러 번 실행해도 안전 (이미 실행된 것은 스킵)
- **Zero Dependency**: stdlib만 사용 (goose 불필요)

### ✅ 개발 편의성
- 개발 중에 새 마이그레이션 추가 → **자동으로 반영됨**
- 버전 중복 입력 불필요 (파일명에서 자동 추출)
- 간단한 `Register()` 함수로 등록

## 구조

```
pkg/migration/data/
├── migrations.go                                  # 패키지 문서
├── runner.go                                      # 실행 엔진 (버전 자동 추출 포함)
├── 20251017000000_filter_and_panel_migration.go  # Filter & Panel 마이그레이션
├── 20251017000000_filter_and_panel_migration_test.go
└── README.md                                      # 이 문서
```

## 사용 방법

### 1. 자동 실행 (권장)

애플리케이션 시작 시 자동으로 실행됩니다:

```go
// daemon/start.go (이미 적용됨)
if err := migration.Load(config); err != nil {
    log.Fatal(err)
}

// 내부적으로:
// 1. schema.Up(config)    - 스키마 마이그레이션 (Goose)
// 2. data.Run(config)     - 데이터 마이그레이션 (Custom)
```

### 2. 상태 확인

```go
import "ntels.com/pharos/core/pkg/migration/data"

if err := data.Status(config); err != nil {
    log.Fatal(err)
}
```

출력 예시:
```
Data Migration Status:
    =======================================
    [Applied] 20251017000000 - filter_and_panel_migration
    [Pending] 20251018120000 - next_migration
    =======================================
```

### 3. 데이터베이스 확인

```sql
-- 실행된 마이그레이션 확인
SELECT * FROM core_data_version ORDER BY id DESC;
```

| id | version_id | is_applied | tstamp |
|----|------------|------------|--------|
| 1 | 20251017000000 | 1 | 2025-10-17 23:30:00 |

## 새 마이그레이션 추가

### 1. 파일 생성

```bash
cd core/pkg/migration/data
touch 20251018120000_your_migration_name.go
```

**파일명 규칙**:
- `YYYYMMDDHHMMSS`: 타임스탬프 (유니크해야 함)
- `_description`: 설명 (snake_case)
- `.go`: Go 파일

### 2. 마이그레이션 작성

```go
package data

import (
	"context"
	"database/sql"
	"fmt"
)

func init() {
	Register(&Migration{
		Name:    "your_migration_name",  // 버전은 파일명에서 자동 추출!
		Up:      upYourMigration,
		Down:    downYourMigration,
	})
}

func upYourMigration(ctx context.Context, tx *sql.Tx) error {
	// 마이그레이션 로직 작성
	query := `UPDATE dashboard SET config = json_set(config, '$.newField', 'value') WHERE ...`
	
	result, err := tx.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("✅ Your migration completed: %d rows updated\n", rowsAffected)
	return nil
}

func downYourMigration(ctx context.Context, tx *sql.Tx) error {
	// 롤백 로직 (선택사항)
	// 데이터 마이그레이션은 일반적으로 롤백 복잡
	return fmt.Errorf("rollback not supported - use backup table")
}
```

**핵심 포인트**:
- ✅ `Version` 필드 **생략 가능** - 파일명 `20251018120000_*.go`에서 자동 추출
- ✅ `init()` 함수에서 `Register()` 호출
- ✅ 새 파일 추가 시 **다른 파일 수정 불필요**

### 3. 실행

앱을 재시작하면 자동으로 실행됩니다:

```bash
cd core
./pharos serve --config ./config/master_config.toml

# 출력:
# INFO Applying migration version=20251018120000 name=your_migration_name
# ✅ Your migration completed: 15 rows updated
# INFO Migration applied successfully version=20251018120000 name=your_migration_name
```

## 기존 마이그레이션

### 20251017000000_filter_and_panel_migration

**목적**: Dashboard config 구조 변경

**변환 내용**:
1. Filter 통합:
   - `headerL` → `filters[].position = "headerL"`
   - `headerR` → `filters[].position = "headerR"`
   - `left` → `filters[].position = "headerL"`
   - `rangeStepOptions` → `filters[type=step].options.rangeStepOptions`

2. Panel type 변경:
   - `panels[].layout.type: "card"` → `"panel"`

**통계 출력 예시**:
```
20251017 Filter & Panel Migration Statistics:
  Total dashboards migrated: 12
  === Filter Migration ===
  headerL filters moved: 10
  headerR filters moved: 12
  left filters moved: 6
  Empty left removed: 6
  rangeStepOptions moved: 12
  Step filters created: 3
  Total filters added: 43
  === Panel Migration ===
  Panel card→panel migrated: 98
```

## 주의사항

### ✅ DO

- 트랜잭션 사용 (`tx *sql.Tx` 파라미터 활용)
- 에러 핸들링 철저히
- 통계 출력으로 변경 내역 추적
- 파일명에 유니크한 타임스탬프 사용
- 실제 DB 테스트 후 커밋

### ❌ DON'T

- 스키마 변경 (DDL) - `pkg/migration/schema/`에서 처리
- 트랜잭션 외부에서 DB 작업
- 같은 타임스탬프로 여러 파일 생성
- 이미 실행된 마이그레이션 파일 수정
- `Version` 필드 수동 입력 (파일명과 불일치 위험)

## 내부 동작 원리

### 버전 자동 추출

`runner.go`의 `Register()` 함수:

```go
var versionPattern = regexp.MustCompile(`^(\d{14})_`)

func Register(m *Migration) {
	if m.Version == 0 {
		_, file, _, _ := runtime.Caller(1)  // 호출자 파일명 가져오기
		filename := filepath.Base(file)     // 20251017000000_filter_and_panel_migration.go
		matches := versionPattern.FindStringSubmatch(filename)
		version, _ := strconv.ParseInt(matches[1], 10, 64)  // 20251017000000
		m.Version = version
	}
	migrations = append(migrations, m)
}
```

### 실행 흐름

1. **등록**: 각 마이그레이션 파일의 `init()` → `Register()` 호출
2. **정렬**: 버전순으로 정렬 (`sort.Slice(migrations, ...)`)
3. **적용 여부 확인**: `core_data_version` 테이블 조회
4. **실행**: 미적용 마이그레이션만 트랜잭션 내에서 실행
5. **기록**: 성공 시 `core_data_version`에 기록

## 트러블슈팅

### "migration already applied" 로그

이미 실행된 마이그레이션입니다. 정상입니다.

### 마이그레이션이 실행되지 않음

1. 파일명이 올바른지 확인 (`YYYYMMDDHHMMSS_*.go`)
2. `init()` 함수에서 `Register()` 호출하는지 확인
3. 빌드를 다시 했는지 확인 (파일 추가 후 재빌드 필요)
4. 버전이 중복되지 않는지 확인

### 파일명 형식 오류

```
panic: migration filename your_migration.go does not match pattern YYYYMMDDHHMMSS_name.go
```

→ 파일명을 `20251018120000_your_migration.go` 형식으로 수정

### 롤백이 필요한 경우

```sql
-- 1. 백업 테이블에서 복구 (백업이 있는 경우)
UPDATE dashboard 
SET config = (SELECT config FROM dashboard_backup WHERE dashboard.id = dashboard_backup.id);

-- 2. 마이그레이션 기록 삭제
DELETE FROM core_data_version WHERE version_id = 20251017000000;

-- 3. 앱 재시작 시 다시 실행됨
```

## 참고

- **패키지 문서**: `migrations.go` 파일 참조
- **테스트**: `*_test.go` 파일로 단위 테스트 작성 권장
- **실제 DB 테스트**: 프로덕션 데이터 복사본으로 테스트 필수

---

**Last Updated**: 2025-10-18  
**Status**: ✅ Production Ready  
**Migration System**: Custom (Goose-free)

## 사용 방법

### 1. 자동 실행 (권장)

애플리케이션 시작 시 자동으로 실행됩니다:

```go
// main.go (이미 적용됨)
if err := migration.Load(config); err != nil {
    log.Fatal(err)
}

// 내부적으로:
// 1. schema.Up(config)    - 스키마 마이그레이션
// 2. data.Run(config)     - 데이터 마이그레이션
```

### 2. 상태 확인

```go
import "ntels.com/pharos/core/pkg/migration/data"

if err := data.Status(config); err != nil {
    log.Fatal(err)
}
```

출력 예시:
```
Applied At                  Migration
=======================================
2025-01-17 10:30:00 UTC  -- 20250117000000_filter_and_panel_migration.go
Pending                  -- 20250118000000_next_migration.go
```

### 3. 데이터베이스 확인

```sql
-- 실행된 마이그레이션 확인
SELECT * FROM pharos_data ORDER BY id DESC;
```

| id | version_id | is_applied | tstamp |
|----|------------|------------|--------|
| 1 | 20250117000000 | 1 | 2025-01-17 10:30:00 |

## 새 마이그레이션 추가

### 1. 파일 생성

```bash
cd core/pkg/migration/data
touch 20250118120000_your_migration_name.go
```

**파일명 규칙**:
- `YYYYMMDDHHMMSS`: 타임스탬프 (유니크해야 함)
- `_description`: 설명 (snake_case)
- `.go`: Go 파일

### 2. 마이그레이션 작성

```go
package data

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"

	"github.com/pressly/goose/v3"
)

func init() {
	// 파일명에서 버전 자동 추출
	_, filename, _, _ := runtime.Caller(0)
	version, err := goose.NumericComponent(filename)
	if err != nil {
		panic(err)
	}

	// 마이그레이션 함수 등록
	up := &goose.GoFunc{RunTx: upYourMigration}
	down := &goose.GoFunc{RunTx: downYourMigration}

	// 마이그레이션 생성 및 등록
	migration := goose.NewGoMigration(version, up, down)
	migration.Source = filename
	Migrations = append(Migrations, migration)
}

func upYourMigration(ctx context.Context, tx *sql.Tx) error {
	// 마이그레이션 로직 작성
	query := `UPDATE dashboard SET config = ...`
	
	if _, err := tx.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	
	fmt.Println("✅ Your migration completed")
	return nil
}

func downYourMigration(ctx context.Context, tx *sql.Tx) error {
	// 롤백 로직 (선택사항)
	// 데이터 마이그레이션은 일반적으로 롤백 복잡
	return fmt.Errorf("rollback not supported")
}
```

**중요**: 각 마이그레이션 파일은 `init()` 함수에서 자동으로 `Migrations` 슬라이스에 추가됩니다.
새 파일을 추가하면 **다른 파일 수정 없이** 자동으로 인식됩니다!

### 3. 실행

앱을 재시작하면 자동으로 실행됩니다:

```bash
cd core
./pharos

# 출력:
# 2025/01/18 12:00:00 goose: applying 20250118120000_your_migration_name.go
# ✅ Your migration completed
# 2025/01/18 12:00:00 goose: successfully applied 20250118120000_your_migration_name.go
```

## 기존 마이그레이션

### 20250117000000_filter_and_panel_migration.go

**목적**: Dashboard config 구조 변경

**변환 내용**:
1. Filter 통합:
   - `headerL` → `filters[].position = "headerL"`
   - `headerR` → `filters[].position = "headerR"`
   - `left` → `filters[].position = "headerL"`
   - `rangeStepOptions` → `filters[type=step].options.rangeStepOptions`

2. Panel type 변경:
   - `panels[].layout.type: "card"` → `"panel"`

**통계 출력**:
```
20250117 Filter & Panel Migration Statistics:
  Total dashboards migrated: 15
  === Filter Migration ===
  headerL filters moved: 12
  headerR filters moved: 15
  left filters moved: 8
  Empty left removed: 7
  rangeStepOptions moved: 15
  Step filters created: 5
  Total filters added: 55
  === Panel Migration ===
  Panel card→panel migrated: 142
```

## 주의사항

### ✅ DO

- 트랜잭션 사용 (`tx *sql.Tx` 파라미터)
- 에러 핸들링
- 통계 출력으로 변경 내역 추적
- 파일명에 유니크한 타임스탬프 사용

### ❌ DON'T

- 스키마 변경 (DDL) - `databases/`에서 처리
- 트랜잭션 외부에서 DB 작업
- 같은 타임스탬프로 여러 파일 생성
- 이미 실행된 마이그레이션 파일 수정

## Goose 명령어 (선택사항)

직접 goose CLI를 사용할 수도 있습니다:

```bash
cd core/pkg/migration/data

# 상태 확인
goose -dir . status

# 특정 마이그레이션까지 실행
goose -dir . up-to 20250117000000

# 롤백
goose -dir . down

# 강제 버전 설정
goose -dir . set-version 20250117000000
```

## 트러블슈팅

### "migration already applied" 에러

이미 실행된 마이그레이션입니다. 정상입니다.

### 마이그레이션이 실행되지 않음

1. 파일명이 올바른지 확인 (`YYYYMMDDHHMMSS_*.go`)
2. `init()` 함수에서 `Migrations`에 추가하는지 확인:
   ```go
   func init() {
       _, filename, _, _ := runtime.Caller(0)
       version, _ := goose.NumericComponent(filename)
       up := &goose.GoFunc{RunTx: upYourMigration}
       down := &goose.GoFunc{RunTx: downYourMigration}
       migration := goose.NewGoMigration(version, up, down)
       migration.Source = filename
       Migrations = append(Migrations, migration)  // ✅ 이 줄이 있어야 함
   }
   ```
3. 빌드를 다시 했는지 확인 (파일 추가 후 재빌드 필요)

### 롤백이 필요한 경우

```sql
-- 백업 테이블에서 복구
UPDATE dashboard 
SET config = (
  SELECT config 
  FROM dashboards_backup_filter_migration 
  WHERE dashboard.id = dashboards_backup_filter_migration.id
);

-- 마이그레이션 기록 삭제
DELETE FROM pharos_data WHERE version_id = '20250117000000';
```

## 참고 문서

- [Goose Documentation](https://github.com/pressly/goose)
- [Filter Migration Summary](../../../migrations/v2.0.2_MIGRATION_SUMMARY.md)
- [Test Results](../../../migrations/v2.0.2_TEST_RESULTS.md)

---

**Last Updated**: 2025-01-17
**Status**: ✅ Production Ready
