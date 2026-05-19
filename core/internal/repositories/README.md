# Repositories Package

이 문서는 `internal/repositories` 패키지의 구조, 설계 패턴 및 새로운 repository를 추가하는 방법에 대한 가이드라인을 제공합니다.

## 📋 목차

- [패키지 구조](#패키지-구조)
- [설계 원칙](#설계-원칙)
- [디렉토리 구조](#디렉토리-구조)
- [새로운 Repository 추가 가이드](#새로운-repository-추가-가이드)
- [테스트 작성 가이드](#테스트-작성-가이드)
- [예제](#예제)

## 📁 패키지 구조

repositories 패키지는 **Clean Architecture**를 따르며, 인터페이스 레이어, 구현 레이어, 테스트 레이어로 명확하게 분리되어 있습니다.

### 핵심 개념

1. **Interface Layer**: 도메인 엔티티와 repository 인터페이스 정의
2. **Implementation Layer**: 데이터베이스별 구체적인 구현
3. **Test Layer**: 인터페이스 기반의 재사용 가능한 테스트 스위트
4. **Factory Pattern**: 데이터베이스 드라이버에 따라 적절한 구현체를 생성

## 🎯 설계 원칙

### 1. 인터페이스 우선 설계 (Interface-First Design)
- 모든 repository는 먼저 인터페이스로 정의됩니다
- 구현체는 인터페이스에 의존하며, 구체적인 데이터베이스 기술에 독립적입니다

### 2. 관심사의 분리 (Separation of Concerns)
- 도메인 로직과 데이터 접근 로직을 명확히 분리
- 각 데이터베이스 구현은 독립적인 패키지로 관리

### 3. 테스트 가능성 (Testability)
- 인터페이스 기반 테스트로 모든 구현체를 동일하게 검증
- 테스트 스위트 재사용으로 일관된 품질 보장

### 4. 확장성 (Extensibility)
- 새로운 데이터베이스 추가 시 최소한의 코드 변경
- Factory 패턴으로 새로운 구현체 통합이 간단함

## 🗂️ 디렉토리 구조

```
internal/repositories/
├── {entity_name}.go              # 인터페이스 정의
│   ├── {Entity}                  # 도메인 엔티티 구조체
│   └── {Entity}Repository        # Repository 인터페이스
│
├── base.go                       # 공통 기능을 제공하는 Base Repository
│
├── factory/
│   ├── factory.go                # Repository 생성을 위한 Factory
│   └── factory_test.go           # Factory 테스트
│
├── {db_type}/                    # 데이터베이스별 구현
│   ├── {db_type}.go              # (선택) DB 연결 관리
│   ├── {entity_name}.go          # Repository 구현
│   └── {entity_name}_test.go    # 구현별 테스트
│
└── test/
    └── {entity_name}.go          # 인터페이스 테스트 스위트
```

### 현재 구조 예시

```
internal/repositories/
├── user.go                       # UserEntity 및 UserRepository 인터페이스
├── base.go                       # BaseRepository (공통 기능)
├── factory/
│   ├── factory.go                # NewUserRepository() 팩토리 함수
│   └── factory_test.go
├── postgres/
│   ├── postgres.go               # PostgreSQL 연결 관리
│   ├── user.go                   # PostgreSQL UserRepository 구현
│   └── user_test.go
├── sqlite/
│   ├── sqlite.go                 # SQLite 연결 관리
│   ├── user.go                   # SQLite UserRepository 구현
│   └── user_test.go
└── test/
    └── user.go                   # UserRepository 테스트 스위트
```

## 🚀 새로운 Repository 추가 가이드

새로운 repository를 추가할 때는 다음 단계를 따르세요:

### Step 1: 인터페이스 정의

`internal/repositories/{entity_name}.go` 파일을 생성하고 엔티티와 인터페이스를 정의합니다.

```go
package repositories

import (
    "context"
    "time"
)

// {Entity}Entity represents a {entity} entity
type {Entity}Entity struct {
    ID        string
    Name      string
    CreatedAt time.Time
    // ... 기타 필드
}

// {Entity}Repository defines the interface for {entity} data access
type {Entity}Repository interface {
    Create(ctx context.Context, entity *{Entity}Entity) error
    GetByID(ctx context.Context, id string) (*{Entity}Entity, error)
    Update(ctx context.Context, entity *{Entity}Entity) error
    Delete(ctx context.Context, id string) error
    // ... 기타 메서드
}
```

### Step 2: 구현 레이어 작성

각 데이터베이스에 대한 구현을 작성합니다.

#### PostgreSQL 구현 예시

`internal/repositories/postgres/{entity_name}.go`:

```go
package postgres

import (
    "context"
    "github.com/jmoiron/sqlx"
    "ntels.com/pharos/core/external/orm"
    "ntels.com/pharos/core/internal/repositories"
)

var _ repositories.{Entity}Repository = (*{entity}Repository)(nil)

type {entity}Repository struct {
    repositories.BaseRepository
}

// New{Entity}Repository creates a new {entity} repository for PostgreSQL
func New{Entity}Repository(config orm.DatabaseConfig) repositories.{Entity}Repository {
    return &{entity}Repository{
        BaseRepository: repositories.NewBaseRepository(config),
    }
}

func (repo *{entity}Repository) Create(ctx context.Context, entity *repositories.{Entity}Entity) error {
    query := `INSERT INTO {table_name} (id, name, created_at) VALUES ($1, $2, $3)`
    
    return repo.Handler(func(db *sqlx.DB) error {
        _, err := db.ExecContext(ctx, query, entity.ID, entity.Name, entity.CreatedAt)
        return err
    })
}

// ... 기타 메서드 구현
```

#### SQLite 구현 예시

`internal/repositories/sqlite/{entity_name}.go`:

```go
package sqlite

import (
    "context"
    "github.com/jmoiron/sqlx"
    "ntels.com/pharos/core/external/orm"
    "ntels.com/pharos/core/internal/repositories"
)

var _ repositories.{Entity}Repository = (*{entity}Repository)(nil)

type {entity}Repository struct {
    repositories.BaseRepository
}

// New{Entity}Repository creates a new {entity} repository for SQLite
func New{Entity}Repository(config orm.DatabaseConfig) repositories.{Entity}Repository {
    return &{entity}Repository{
        BaseRepository: repositories.NewBaseRepository(config),
    }
}

func (repo *{entity}Repository) Create(ctx context.Context, entity *repositories.{Entity}Entity) error {
    query := `INSERT INTO {table_name} (id, name, created_at) VALUES (?, ?, ?)`
    
    return repo.Handler(func(db *sqlx.DB) error {
        _, err := db.ExecContext(ctx, query, entity.ID, entity.Name, entity.CreatedAt)
        return err
    })
}

// ... 기타 메서드 구현
```

### Step 3: Factory 함수 추가

`internal/repositories/factory/factory.go`에 새로운 factory 함수를 추가합니다:

```go
// New{Entity}Repository creates a new {entity} repository based on the database driver
func New{Entity}Repository(opts RepositoryOptions) (repositories.{Entity}Repository, error) {
    if opts.DatabaseConfig.GetDriverName() == "" || opts.DatabaseConfig.GetDataSourceName() == "" {
        return nil, errors.New("database configuration is incomplete")
    }

    switch opts.DatabaseConfig.Driver {
    case orm.DriverSqlite:
        return sqlite.New{Entity}Repository(opts.DatabaseConfig), nil
    case orm.DriverPostgreSQL:
        return postgres.New{Entity}Repository(opts.DatabaseConfig), nil
    default:
        return nil, errors.New("unsupported database driver: " + opts.DatabaseConfig.Driver)
    }
}
```

### Step 4: 테스트 스위트 작성

`internal/repositories/test/{entity_name}.go`에 인터페이스 테스트 스위트를 작성합니다:

```go
package test

import (
    "context"
    "testing"
    "ntels.com/pharos/core/internal/repositories"
)

type {Entity}Testable interface {
    repositories.{Entity}Repository
    SetUpDb() error
    TearDownDb() error
    // ... 테스트 헬퍼 메서드
}

func {Entity}Suite(t *testing.T, testable {Entity}Testable) {
    t.Run("Create", func(t *testing.T) { Create{Entity}(t, testable) })
    t.Run("GetByID", func(t *testing.T) { Get{Entity}ByID(t, testable) })
    // ... 기타 테스트
}

func Create{Entity}(t *testing.T, testable {Entity}Testable) {
    err := testable.SetUpDb()
    if err != nil {
        t.Fatalf("failed to set up test db: %v", err)
    }
    t.Cleanup(func() {
        if err := testable.TearDownDb(); err != nil {
            t.Errorf("failed to tear down test db: %v", err)
        }
    })

    entity := &repositories.{Entity}Entity{
        ID:   "test-id",
        Name: "test-name",
    }

    err = testable.Create(context.Background(), entity)
    if err != nil {
        t.Fatalf("failed to create entity: %v", err)
    }

    // 검증 로직...
}

// ... 기타 테스트 함수
```

### Step 5: 구현별 테스트 작성

각 데이터베이스 구현에 대한 테스트를 작성합니다.

`internal/repositories/postgres/{entity_name}_test.go`:

```go
package postgres

import (
    "testing"
    "ntels.com/pharos/core/internal/repositories/test"
)

type Test{Entity} struct {
    {entity}Repository
    // 테스트 헬퍼 필드...
}

func (t *Test{Entity}) SetUpDb() error {
    // DB 설정...
    return nil
}

func (t *Test{Entity}) TearDownDb() error {
    // DB 정리...
    return nil
}

func Test_{entity}(t *testing.T) {
    testable := &Test{Entity}{
        // 초기화...
    }
    test.{Entity}Suite(t, testable)
}
```

## 🧪 테스트 작성 가이드

### 테스트 구조

1. **Interface Tests**: `test/` 패키지에서 인터페이스를 검증하는 재사용 가능한 테스트 작성
2. **Implementation Tests**: 각 데이터베이스별 패키지에서 테스트 스위트 실행

### 테스트 원칙

- **DRY (Don't Repeat Yourself)**: 공통 테스트 로직은 `test/` 패키지의 함수로 추출
- **독립성**: 각 테스트는 독립적으로 실행 가능해야 함
- **정리**: `t.Cleanup()`을 사용하여 테스트 후 리소스 정리
- **명확성**: 테스트 실패 시 명확한 에러 메시지 제공

### 테스트 헬퍼 인터페이스

테스트를 위해 추가 메서드가 필요한 경우 Testable 인터페이스를 확장합니다:

```go
type {Entity}Testable interface {
    repositories.{Entity}Repository
    
    // 테스트 설정
    SetUpDb() error
    TearDownDb() error
    SetUp{Entity}(id string) error
    
    // 테스트 헬퍼
    ClearTable() error
    Get{Entity}Count() (int, error)
    // ... 기타 헬퍼 메서드
}
```

## 📚 예제

### 사용 예시

```go
package main

import (
    "context"
    "ntels.com/pharos/core/external/orm"
    "ntels.com/pharos/core/internal/repositories/factory"
)

func main() {
    // 1. 데이터베이스 설정
    dbConfig := orm.DatabaseConfig{
        Driver: orm.DriverPostgreSQL,
        PostgreSQL: orm.PostgreSQLConfig{
            Host:     "localhost",
            Port:     5432,
            Database: "mydb",
            Username: "user",
            Password: "pass",
        },
    }

    // 2. Factory를 통한 Repository 생성
    userRepo, err := factory.NewUserRepository(factory.RepositoryOptions{
        DatabaseConfig: dbConfig,
    })
    if err != nil {
        panic(err)
    }

    // 3. Repository 사용
    ctx := context.Background()
    users, err := userRepo.ListAll(ctx)
    if err != nil {
        panic(err)
    }

    for _, user := range users {
        println(user.Username)
    }
}
```

### BaseRepository 활용

공통 기능을 사용하려면 `BaseRepository`를 임베드합니다:

```go
type myRepository struct {
    repositories.BaseRepository
}

func (repo *myRepository) SomeMethod() error {
    // BaseRepository의 메서드 사용
    data, err := repo.MarshalAttributes(map[string]interface{}{
        "key": "value",
    })
    if err != nil {
        return err
    }

    // Handler를 통한 DB 작업
    return repo.Handler(func(db *sqlx.DB) error {
        _, err := db.Exec("INSERT INTO table (data) VALUES (?)", data)
        return err
    })
}
```

## 🔍 Best Practices

### 1. 인터페이스 설계
- **작고 집중적으로**: 각 인터페이스는 하나의 책임만 가져야 합니다
- **컨텍스트 사용**: 모든 메서드에 `context.Context`를 첫 번째 파라미터로 전달
- **에러 처리**: 적절한 에러 타입과 메시지를 반환

### 2. 구현 패턴
- **BaseRepository 활용**: 공통 기능은 BaseRepository에 구현
- **DB별 최적화**: 각 데이터베이스의 특성을 살린 쿼리 작성
- **트랜잭션 관리**: Handler를 통해 일관된 트랜잭션 처리

### 3. 명명 규칙
- **엔티티**: `{Entity}Entity` (예: `UserEntity`, `RoleEntity`)
- **인터페이스**: `{Entity}Repository` (예: `UserRepository`, `RoleRepository`)
- **구현체**: `{entity}Repository` (소문자, private) (예: `userRepository`, `roleRepository`)
- **Factory 함수**: `New{Entity}Repository` (예: `NewUserRepository`)
- **Public 생성자**: 각 DB 패키지의 `New{Entity}Repository` (예: `postgres.NewUserRepository`)

### 4. 에러 처리
```go
// 에러 정의
var (
    ErrNotFound = errors.New("entity not found")
    ErrDuplicate = errors.New("entity already exists")
)

// 사용
func (repo *myRepository) GetByID(ctx context.Context, id string) (*Entity, error) {
    // ...
    if errors.Is(err, sql.ErrNoRows) {
        return nil, ErrNotFound
    }
    return nil, err
}
```

### 5. 테스트 커버리지
- 모든 인터페이스 메서드에 대한 테스트 작성
- 성공 케이스와 실패 케이스 모두 검증
- 경계 조건 테스트 포함

## 🤝 기여 가이드

새로운 repository를 추가하거나 기존 코드를 수정할 때:

1. **이 가이드를 따라 구조를 일관되게 유지**
2. **모든 public 인터페이스에 대한 문서화 주석 작성**
3. **테스트 커버리지 유지 (최소 80% 이상)**
4. **코드 리뷰 전 로컬에서 모든 테스트 통과 확인**

## 📝 참고 자료

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Repository Pattern](https://martinfowler.com/eaaCatalog/repository.html)
- [Go Interface Best Practices](https://golang.org/doc/effective_go#interfaces)

---

**질문이나 제안사항이 있으시면 이슈를 생성해 주세요!**
