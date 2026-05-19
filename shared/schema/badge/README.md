# Badge Schema Documentation

Badge 시스템의 JSON Schema 정의와 사용법을 설명합니다.

## 📋 Schema Files

### 1. BadgeSpec.schema
Badge 정의 생성 및 수정을 위한 요청 스키마입니다.

**위치**: `shared/schema/badge/BadgeSpec.schema`  
**용도**: POST, PUT 요청에서 사용  
**생성 타입**: `shared/types/badgespec.Types`

### 2. BadgeValueResponse.schema  
Badge 값 조회 응답을 위한 스키마입니다.

**위치**: `shared/schema/badge/BadgeValueResponse.schema`  
**용도**: GET 응답에서 사용  
**생성 타입**: `shared/types/badgevalueresponse.Types`

## 🏗️ 아키텍처 패턴

```
Frontend ←→ shared/types ←→ Adapter ←→ Internal Types ←→ Backend
            ↓ 자동 생성     ↓ 변환      ↓ 쿼리 실행
        JSON Schema    FromAPI/ToAPI   BadgeSpec
                                      query_builder.QuerySpec
```

## 📖 사용 예제

### POST Badge Spec (생성)
```bash
curl -X POST /api/badge \
  -H "Content-Type: application/json" \
  -d '{
    "name": "active_users",
    "datasource": "postgresql_main",
    "query": {
      "query": "SELECT COUNT(*) as count FROM users WHERE status = {{status}}",
      "variables": {
        "status": "active"
      },
      "time_label": "created_at"
    },
    "column": "count",
    "cache_ttl": 300
  }'
```

### PUT Badge Spec (수정)
```bash
curl -X PUT /api/badge/active_users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "active_users",
    "datasource": "postgresql_main",
    "query": {
      "query": "SELECT COUNT(*) as total FROM users WHERE status = {{status}} AND last_login > {{since}}",
      "variables": {
        "status": "active",
        "since": "2025-01-01"
      }
    },
    "column": "total",
    "cache_ttl": 180
  }'
```

### GET Badge Value (조회)
```bash
curl /api/badge/active_users
```

**응답 예시 (Cache Miss)**:
```json
{
  "name": "active_users",
  "value": 1543,
  "last_updated": "2025-09-27T10:30:00Z",
  "cache_hit": false
}
```

**응답 예시 (Cache Hit)**:
```json
{
  "name": "active_users", 
  "value": 1543,
  "last_updated": "2025-09-27T10:30:00Z",
  "cache_hit": true
}
```

### GET Badge List (목록 조회)
```bash
curl /api/badge
```

**응답 예시**:
```json
[
  "active_users",
  "system_health",
  "daily_orders"
]
```

### DELETE Badge (삭제)
```bash
curl -X DELETE /api/badge/active_users
```

## 🔄 Adapter 패턴

### Handler에서의 사용
```go
// POST/PUT Handler
var apiSpec badgespec.Types
json.Unmarshal(body, &apiSpec)

adapter := adapters.NewBadgeAdapter()
internalSpec := adapter.FromAPIBadgeSpec(apiSpec)

// GET Handler - Badge Value
internalResponse := adapters.BadgeValueResponse{
    Name: name,
    Value: extractedValue,
}
adapter := adapters.NewBadgeAdapter()
apiResponse := adapter.ToAPIBadgeValueResponse(internalResponse, cacheHit)
c.JSON(http.StatusOK, apiResponse)
```

### 타입 변환 흐름
1. **POST/PUT**: `JSON → badgespec.Types → FromAPIBadgeSpec() → internal BadgeSpec`
2. **GET**: `internal BadgeValueResponse → ToAPIBadgeValueResponse() → badgevalueresponse.Types → JSON`

## 🎯 QuerySpec 통합

Badge는 Alert와 공통된 `QuerySpec` 구조를 사용합니다:

### QuerySpec 구조
```json
{
  "query": "SELECT metric FROM table WHERE condition = {{variable}}",
  "variables": {
    "variable": "value"
  },
  "time_label": "timestamp_column",
  "variable_label": "variable_column"
}
```

### 공통 인터페이스 장점
- Alert와 Badge가 동일한 쿼리 빌더 사용
- 템플릿 변수 처리 통합
- 시간 기반 쿼리 지원

## 💾 Cache 시스템

### Cache 설정
- **cache_ttl**: 캐시 유지 시간 (초 단위)
- **0**: 캐시 비활성화  
- **기본값**: 60초

### Cache 동작
```go
// Cache Hit 시
if v, ok := badgeCache.Get(name); ok {
    // Cache에서 직접 반환, cache_hit: true
}

// Cache Miss 시  
// 1. 쿼리 실행
// 2. 결과를 캐시에 저장  
// 3. 응답 반환, cache_hit: false
```

## 📊 지원하는 Datasource

- **PostgreSQL**: `postgresql_main`
- **ClickHouse**: `clickhouse_analytics`  
- **Prometheus**: `prometheus_metrics`
- **Other**: 플러그인으로 확장 가능

## ✅ 검증된 기능

- ✅ POST/PUT 요청에 동일 스키마 사용으로 일관성 확보
- ✅ QuerySpec을 통한 Alert와의 통합 인터페이스
- ✅ Cache hit/miss 정보까지 스키마로 처리
- ✅ 템플릿 변수 처리 및 동적 쿼리 지원
- ✅ Frontend와 Backend 타입 동기화
- ✅ 26개 테스트 케이스 통과

## 🔗 관련 파일

- **Handler**: `core/pkg/badges/handler.go`
- **Adapter**: `core/pkg/badges/adapters/badge_adapter.go`
- **Tests**: `core/pkg/badges/adapters/badge_adapter_test.go`
- **Types**: `shared/types/badgespec/`, `shared/types/badgevalueresponse/`
- **Query Builder**: `core/internal/query_builder/`

## 🎯 장점

1. **Query 통합**: Alert와 공통 QuerySpec으로 통합된 쿼리 인터페이스
2. **Cache 최적화**: TTL 기반 캐시로 성능 최적화 및 상태 추적
3. **다양한 Datasource**: 플러그인 아키텍처로 확장 가능
4. **템플릿 변수**: 동적 쿼리 생성으로 유연한 Badge 정의  
5. **순환 Import 방지**: 독립적인 어댑터 타입으로 깔끔한 의존성