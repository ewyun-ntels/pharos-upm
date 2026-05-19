# Dashboard Repository 마이그레이션 가이드

## 기존 코드 vs 새로운 Repository 패턴

### 기존 방식 (dashboard.go)

```go
// 데이터베이스 직접 접근
func (dashboard *Dashboard) getConfig(check func(string) bool, hide bool) (map[string]any, error) {
    // ...
    handler := func(db *sqlx.DB) error {
        return db.Get(dashboard, `SELECT * FROM dashboard WHERE id = $1;`, dashboard.ID)
    }
    if err := orm.Handler(orm.DriverDefault, &config.Database, handler); errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    } else if err != nil {
        return nil, err
    }
    // ...
}
```

**문제점:**
- 비즈니스 로직과 데이터 접근 로직이 혼재
- 테스트 시 실제 데이터베이스 필요
- SQL 쿼리가 여러 곳에 분산
- 데이터베이스 변경 시 여러 파일 수정 필요

### 새로운 방식 (Repository 패턴)

```go
// Repository를 통한 데이터 접근
func (s *DashboardService) getConfig(id string, check func(string) bool, hide bool) (map[string]any, error) {
    // ...
    
    // Repository를 통해 대시보드 조회
    dashboard, err := s.repo.GetByID(context.Background(), id)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    } else if err != nil {
        return nil, err
    }
    // ...
}
```

**장점:**
- 비즈니스 로직과 데이터 접근 로직 분리
- Repository를 Mock으로 대체하여 테스트 가능
- SQL 쿼리가 Repository에 집중
- 데이터베이스 변경 시 Repository 구현만 수정

## 단계별 마이그레이션

### 1단계: Repository 생성

```go
import (
    "ntels.com/pharos/core/internal/repositories/factory"
)

repo, err := factory.NewDashboardRepository(factory.RepositoryOptions{
    DatabaseConfig: config.Database,
})
if err != nil {
    return err
}
```

### 2단계: 기존 SQL 쿼리를 Repository 메서드로 변환

#### Before (기존)
```go
handler := func(db *sqlx.DB) error {
    return db.Get(dashboard, `SELECT * FROM dashboard WHERE id = $1;`, dashboard.ID)
}
if err := orm.Handler(orm.DriverDefault, &config.Database, handler); err != nil {
    return err
}
```

#### After (Repository)
```go
dashboard, err := repo.GetByID(context.Background(), id)
if err != nil {
    return err
}
```

### 3단계: CRUD 작업 변환 예제

#### Create (생성)

**Before:**
```go
dashboard.ID = uuid.New().String()

handler := func(db *sqlx.DB) error {
    _, err := db.NamedExec(`INSERT INTO dashboard(id, config) VALUES (:id, :config);`, dashboard)
    return err
}
if err := orm.Handler(orm.DriverDefault, &config.Database, handler); err != nil {
    return err
}
```

**After:**
```go
dashboard := &repositories.DashboardEntity{
    ID:     uuid.New().String(),
    Config: configJSON,
}

if err := repo.Create(context.Background(), dashboard); err != nil {
    return err
}
```

#### Read (조회)

**Before:**
```go
handler := func(db *sqlx.DB) error {
    return db.Get(dashboard, `SELECT * FROM dashboard WHERE id = $1;`, dashboard.ID)
}
if err := orm.Handler(orm.DriverDefault, &config.Database, handler); err != nil {
    return err
}
```

**After:**
```go
dashboard, err := repo.GetByID(context.Background(), id)
if err != nil {
    return err
}
```

#### Update (수정)

**Before:**
```go
handler := func(db *sqlx.DB) error {
    _, err := db.NamedExec(`UPDATE dashboard SET config = :config WHERE id = :id;`, dashboard)
    return err
}
if err := orm.Handler(orm.DriverDefault, &config.Database, handler); err != nil {
    return err
}
```

**After:**
```go
dashboard := &repositories.DashboardEntity{
    ID:     id,
    Config: newConfigJSON,
}

if err := repo.Update(context.Background(), dashboard); err != nil {
    return err
}
```

#### Delete (삭제)

**Before:**
```go
handler := func(db *sqlx.DB) error {
    _, err := db.Exec(`DELETE FROM dashboard WHERE id = $1;`, dashboard.ID)
    return err
}
if err := orm.Handler(orm.DriverDefault, &config.Database, handler); err != nil {
    return err
}
```

**After:**
```go
if err := repo.DeleteByID(context.Background(), id); err != nil {
    return err
}
```

#### List All (전체 조회)

**Before:**
```go
dashboards := []Dashboard{}

handler := func(db *sqlx.DB) error {
    return db.Select(&dashboards, `SELECT * FROM dashboard;`)
}
if err := orm.Handler(orm.DriverDefault, &config.Database, handler); err != nil {
    return err
}
```

**After:**
```go
dashboards, err := repo.ListAll(context.Background())
if err != nil {
    return err
}
```

## 점진적 마이그레이션 전략

### 방법 1: 병렬 실행 (추천)

기존 코드를 유지하면서 새 코드를 추가합니다.

```go
// handler.go - 기존 라우트 (그대로 유지)
routes.GET("/:id", authhandler.GetAuthenticationHandler(false), dashboardHandler)

// 새로운 라우트 추가 (선택적)
routes.GET("/v2/:id", authhandler.GetAuthenticationHandler(false), dashboardServiceHandler)
```

**장점:**
- 기존 기능에 영향 없음
- 점진적으로 전환 가능
- 롤백 용이

### 방법 2: 단계적 교체

1. **준비 단계**: DashboardService 구현
2. **테스트 단계**: 새 서비스 검증
3. **전환 단계**: 기존 핸들러에서 새 서비스 호출
4. **정리 단계**: 중복 코드 제거

```go
// 전환 단계 예시
func dashboardHandler(c *gin.Context) {
    // 새 서비스 사용
    service, err := handler.NewDashboardService(config, enforcer, getJWTClaims(c))
    if err != nil {
        c.JSON(http.StatusInternalServerError, resources.ErrorResponse{Message: err.Error()})
        return
    }
    
    switch c.Request.Method {
    case http.MethodGet:
        c.JSON(service.GetHandler(c))
    // ...
    }
}
```

## 테스트 전략

### Repository 단위 테스트

```go
func TestDashboardRepository_Create(t *testing.T) {
    // 테스트 데이터베이스 설정
    repo := setupTestRepository(t)
    
    dashboard := &repositories.DashboardEntity{
        ID:     uuid.New().String(),
        Config: `{"title": "Test"}`,
    }
    
    err := repo.Create(context.Background(), dashboard)
    assert.NoError(t, err)
    
    // 검증
    retrieved, err := repo.GetByID(context.Background(), dashboard.ID)
    assert.NoError(t, err)
    assert.Equal(t, dashboard.Config, retrieved.Config)
}
```

### Service 단위 테스트 (Mock Repository)

```go
type MockDashboardRepository struct {
    mock.Mock
}

func (m *MockDashboardRepository) GetByID(ctx context.Context, id string) (*repositories.DashboardEntity, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*repositories.DashboardEntity), args.Error(1)
}

func TestDashboardService_GetHandler(t *testing.T) {
    mockRepo := new(MockDashboardRepository)
    service := &DashboardService{
        repo: mockRepo,
        // ...
    }
    
    mockRepo.On("GetByID", mock.Anything, "test-id").Return(&repositories.DashboardEntity{
        ID:     "test-id",
        Config: `{"title": "Test"}`,
    }, nil)
    
    // 테스트 실행
    // ...
}
```

## 주의사항

### 1. Context 사용
모든 Repository 메서드는 `context.Context`를 첫 번째 파라미터로 받습니다.

```go
// Good
dashboard, err := repo.GetByID(context.Background(), id)

// 요청 컨텍스트 전달 (타임아웃, 취소 가능)
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
dashboard, err := repo.GetByID(ctx, id)
```

### 2. 에러 처리
Repository는 데이터베이스 에러를 그대로 반환합니다.

```go
dashboard, err := repo.GetByID(ctx, id)
if errors.Is(err, sql.ErrNoRows) {
    // 레코드가 없는 경우
    return http.StatusNotFound, nil
} else if err != nil {
    // 기타 데이터베이스 에러
    return http.StatusInternalServerError, err
}
```

### 3. 트랜잭션
현재 구현은 단일 작업만 지원합니다. 트랜잭션이 필요한 경우 추가 구현이 필요합니다.

```go
// 향후 트랜잭션 지원 예시
type DashboardRepository interface {
    // ...
    WithTransaction(ctx context.Context, fn func(repo DashboardRepository) error) error
}
```

## 다음 단계

1. **테스트 작성**: Repository와 Service의 단위 테스트
2. **통합 테스트**: 전체 플로우 검증
3. **성능 테스트**: Repository 패턴의 오버헤드 측정
4. **문서화**: API 문서 업데이트
5. **점진적 마이그레이션**: 기존 코드를 새 패턴으로 전환

## 참고 자료

- [Repository Pattern](https://martinfowler.com/eaaCatalog/repository.html)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Dependency Injection in Go](https://blog.drewolson.org/dependency-injection-in-go)
