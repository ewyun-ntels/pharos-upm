# Alert Schema Documentation

Alert 시스템의 JSON Schema 정의와 사용법을 설명합니다.

## 📋 Schema Files

### 1. AlertRule.schema
Alert 규칙 생성 및 수정을 위한 요청 스키마입니다.

**위치**: `shared/schema/alert/AlertRule.schema`  
**용도**: POST, PUT 요청에서 사용  
**생성 타입**: `shared/types/alertrule.Types`

### 2. AlertRuleResponse.schema  
Alert 규칙 조회 응답을 위한 스키마입니다.

**위치**: `shared/schema/alert/AlertRuleResponse.schema`  
**용도**: GET 응답에서 사용  
**생성 타입**: `shared/types/alertruleresponse.Types`

## 🏗️ 아키텍처 패턴

```
Frontend ←→ shared/types ←→ Adapter ←→ Internal Types ←→ Backend
            ↓ 자동 생성     ↓ 변환      ↓ 비즈니스 로직
        JSON Schema    FromAPI/ToAPI   resources.Rule
```

## 📖 사용 예제

### POST Alert Rule (생성)
```bash
curl -X POST /api/alert/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "High CPU Alert",
    "alert_type": "query",
    "rule": {
      "query": "cpu_usage > 90",
      "severity": "Critical"
    }
  }'
```

### GET Alert Rule (조회)
```bash
curl /api/alert/rules/123
```

**응답 예시**:
```json
{
  "id": "123",
  "name": "High CPU Alert", 
  "alert_type": "query",
  "rule": {
    "query": "cpu_usage > 90",
    "severity": "Critical"
  },
  "created_at": "2025-09-27T00:00:00Z",
  "updated_at": "2025-09-27T00:00:00Z"
}
```

### PUT Alert Rule (수정)
```bash
curl -X PUT /api/alert/rules/123 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated High CPU Alert",
    "alert_type": "query", 
    "rule": {
      "query": "cpu_usage > 85",
      "severity": "Major"
    }
  }'
```

## 🔄 Adapter 패턴

### Handler에서의 사용
```go
// POST Handler
var apiRule alertrule.Types
json.Unmarshal(body, &apiRule)

adapter := adapters.NewApiAdapter()
internalRule, err := adapter.FromAPIAlertRule(apiRule)

// GET Handler  
internalRule := getInternalRule(id)
adapter := adapters.NewApiAdapter()
apiResponse := adapter.ToAPIAlertRuleResponse(internalRule)
c.JSON(http.StatusOK, apiResponse)
```

### 타입 변환 흐름
1. **POST/PUT**: `JSON → alertrule.Types → FromAPIAlertRule() → resources.Rule`
2. **GET**: `resources.Rule → ToAPIAlertRuleResponse() → alertruleresponse.Types → JSON`

## ✅ 검증된 기능

- ✅ CRUD 전체 작업에 shared/types 사용
- ✅ 자동 타입 검증 및 마샬링/언마샬링  
- ✅ Frontend와 Backend 타입 동기화
- ✅ 25개 테스트 케이스 통과

## 🔗 관련 파일

- **Handler**: `core/pkg/alert/handler.go`
- **Adapter**: `core/pkg/alert/adapters/api_adapter.go`
- **Tests**: `core/pkg/alert/adapters/api_types_test.go`
- **Types**: `shared/types/alertrule/`, `shared/types/alertruleresponse/`

## 🎯 장점

1. **타입 안전성**: JSON Schema로 런타임 검증
2. **자동 동기화**: Schema 변경 시 Frontend/Backend 자동 반영  
3. **일관된 API**: 모든 엔드포인트가 동일한 패턴 사용
4. **유지보수성**: 중앙화된 타입 관리로 변경 추적 용이