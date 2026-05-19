# Dashboard Repository 구현

## 개요

Dashboard 패키지에 Clean Architecture 원칙을 적용하여 데이터 레이어와 비즈니스 로직을 분리했습니다.

## 아키텍처

```
core/pkg/dashboard/handler/
├── dashboard.go              # 기존 핸들러 (변경 없음)
├── dashboard_service.go      # 새로운 서비스 구조체 (Repository 패턴 사용)
└── ...

core/internal/repositories/
├── dashboard.go              # Dashboard 엔티티 및 Repository 인터페이스
├── postgres/
│   └── dashboard.go         # PostgreSQL 구현
├── sqlite/
│   └── dashboard.go         # SQLite 구현
└── factory/
    └── factory.go           # Repository 생성 팩토리
```

## 주요 컴포넌트

### 1. Repository 인터페이스 (`core/internal/repositories/dashboard.go`)

```go
type DashboardRepository interface {
    Create(ctx context.Context, dashboard *DashboardEntity) error
    GetByID(ctx context.Context, id string) (*DashboardEntity, error)
    ListAll(ctx context.Context) ([]*DashboardEntity, error)
    Update(ctx context.Context, dashboard *DashboardEntity) error
    DeleteByID(ctx context.Context, id string) error
}
```

Dashboard 데이터 접근을 위한 표준 인터페이스를 정의합니다.

### 2. DashboardEntity

```go
type DashboardEntity struct {
    ID     string
    Config string
}
```

데이터베이스의 dashboard 테이블과 매핑되는 엔티티입니다.

### 3. 데이터베이스별 구현

- **PostgreSQL 구현** (`core/internal/repositories/postgres/dashboard.go`)
  - PostgreSQL 특화 쿼리 사용 (`$1`, `$2` 플레이스홀더)
  
- **SQLite 구현** (`core/internal/repositories/sqlite/dashboard.go`)
  - SQLite 특화 쿼리 사용 (`?` 플레이스홀더)

### 4. Factory 패턴

```go
repo, err := factory.NewDashboardRepository(factory.RepositoryOptions{
    DatabaseConfig: config.Database,
})
```

데이터베이스 드라이버에 따라 적절한 구현체를 자동으로 생성합니다.

### 5. DashboardService

```go
service, err := NewDashboardService(config, enforcer, jwtClaims)
```

Repository를 사용하는 새로운 서비스 레이어입니다. 기존 `Dashboard` 구조체와 별도로 동작하며:

- 데이터 레이어(Repository)와 비즈니스 로직(Service) 분리
- 권한 검증 로직 포함
- 기존 코드에 영향 없이 독립적으로 동작

## 사용 예제

### Repository 직접 사용

```go
import (
    "context"
    "ntels.com/pharos/core/internal/repositories/factory"
)

// Repository 생성
repo, err := factory.NewDashboardRepository(factory.RepositoryOptions{
    DatabaseConfig: config.Database,
})

// Dashboard 생성
dashboard := &repositories.DashboardEntity{
    ID:     uuid.New().String(),
    Config: `{"title": "My Dashboard"}`,
}
err = repo.Create(context.Background(), dashboard)

// Dashboard 조회
dashboard, err = repo.GetByID(context.Background(), id)

// Dashboard 업데이트
dashboard.Config = `{"title": "Updated Dashboard"}`
err = repo.Update(context.Background(), dashboard)

// Dashboard 삭제
err = repo.DeleteByID(context.Background(), id)
```

### DashboardService 사용

```go
import (
    "ntels.com/pharos/core/pkg/dashboard/handler"
)

// Service 생성
service, err := handler.NewDashboardService(config, enforcer, jwtClaims)

// Gin 핸들러에서 사용
func dashboardHandler(c *gin.Context) {
    service, _ := handler.NewDashboardService(config, enforcer, getJWTClaims(c))
    
    switch c.Request.Method {
    case http.MethodGet:
        c.JSON(service.GetHandler(c))
    case http.MethodPost:
        c.JSON(service.PostHandler(c))
    // ...
    }
}
```

## 기존 코드와의 호환성

- **기존 코드는 전혀 수정되지 않았습니다**
- `dashboard.go`의 기존 `Dashboard` 구조체는 그대로 유지
- 새로운 `DashboardService`는 별도의 파일(`dashboard_service.go`)에 구현
- 기존 핸들러와 라우트는 계속 정상 동작

## 장점

### 1. 관심사의 분리 (Separation of Concerns)
- 데이터 접근 로직 → Repository
- 비즈니스 로직 → Service
- API 레이어 → Handler

### 2. 테스트 용이성
- Repository를 Mock으로 대체하여 Service 레이어 단위 테스트 가능
- 각 레이어를 독립적으로 테스트 가능

### 3. 유지보수성
- 데이터베이스 변경 시 Repository 구현만 수정
- 비즈니스 로직 변경 시 Service만 수정
- 명확한 책임 분리로 코드 이해도 향상

### 4. 확장성
- 새로운 데이터베이스 지원: 새 Repository 구현 추가
- 새로운 비즈니스 로직: Service에 메서드 추가
- Factory 패턴으로 쉬운 구현체 교체

## 데이터베이스 스키마

```sql
CREATE TABLE IF NOT EXISTS dashboard (
    id     TEXT PRIMARY KEY NOT NULL,
    config JSONB NOT NULL  -- PostgreSQL
    -- config TEXT NOT NULL  -- SQLite
);
```

## 설계 원칙

이 구현은 `core/internal/repositories/README.md`에 정의된 다음 원칙을 따릅니다:

1. **Interface-First Design**: 인터페이스 우선 설계
2. **Separation of Concerns**: 관심사의 분리
3. **Testability**: 테스트 가능성
4. **Extensibility**: 확장성

## 참고 자료

- [Repository Pattern 가이드](../../../internal/repositories/README.md)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [User Repository 구현 예제](../../../internal/repositories/user.go)
