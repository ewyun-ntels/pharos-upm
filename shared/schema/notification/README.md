# Notification Schema Documentation

Notification 시스템의 JSON Schema 정의와 사용법을 설명합니다.

## 📋 Schema Files

### 1. NotificationRule.schema
Notification 규칙 생성을 위한 요청 스키마입니다.

**위치**: `shared/schema/notification/NotificationRule.schema`  
**용도**: POST 요청에서 사용  
**생성 타입**: `shared/types/notificationrule.Types`

### 2. NotificationRuleUpdate.schema
Notification 규칙 수정을 위한 요청 스키마입니다.

**위치**: `shared/schema/notification/NotificationRuleUpdate.schema`  
**용도**: PUT 요청에서 사용  
**생성 타입**: `shared/types/notificationruleupdate.Types`

### 3. NotificationRuleResponse.schema  
Notification 규칙 조회 응답을 위한 스키마입니다.

**위치**: `shared/schema/notification/NotificationRuleResponse.schema`  
**용도**: GET 응답에서 사용  
**생성 타입**: `shared/types/notificationruleresponse.Types`

## 🏗️ 아키텍처 패턴

```
Frontend ←→ shared/types ←→ Adapter ←→ Internal Types ←→ Backend
            ↓ 자동 생성     ↓ 변환      ↓ 비즈니스 로직
        JSON Schema    FromAPI/ToAPI   resources.Rule
```

## 📖 사용 예제

### POST Notification Rule (생성)
```bash
curl -X POST /api/notification/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Critical Alert Email",
    "notification_type": "email",
    "priority": "high",
    "configuration": {
      "to": ["admin@company.com"],
      "subject": "Critical Alert Detected"
    },
    "enabled": true
  }'
```

### PUT Notification Rule (수정)
```bash
curl -X PUT /api/notification/rules/456 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Critical Alert Email",
    "notification_type": "email", 
    "priority": "medium",
    "configuration": {
      "to": ["admin@company.com", "ops@company.com"],
      "subject": "Updated: Critical Alert Detected"
    },
    "enabled": true
  }'
```

### GET Notification Rule (조회)
```bash
curl /api/notification/rules/456
```

**응답 예시**:
```json
{
  "id": "456",
  "name": "Critical Alert Email",
  "notification_type": "email",
  "priority": "high", 
  "configuration": {
    "to": ["admin@company.com"],
    "subject": "Critical Alert Detected"
  },
  "enabled": true,
  "created_at": "2025-09-27T00:00:00Z",
  "updated_at": "2025-09-27T00:00:00Z"
}
```

### GET Notification Rules List (목록 조회)
```bash
curl /api/notification/rules
```

## 🔄 Adapter 패턴

### Handler에서의 사용
```go
// POST Handler
var apiRule notificationrule.Types
json.Unmarshal(body, &apiRule)

adapter := adapters.NewApiAdapter()
internalRule, err := adapter.FromAPINotificationRule(apiRule)

// PUT Handler
var apiUpdateRule notificationruleupdate.Types
json.Unmarshal(body, &apiUpdateRule)

adapter := adapters.NewApiAdapter()
internalRule, err := adapter.FromAPINotificationRuleUpdate(apiUpdateRule)

// GET Handler
internalRule := getInternalRule(id)
adapter := adapters.NewApiAdapter()
apiResponse := adapter.ToAPINotificationRuleResponse(internalRule)
c.JSON(http.StatusOK, apiResponse)
```

### 타입 변환 흐름
1. **POST**: `JSON → notificationrule.Types → FromAPINotificationRule() → resources.Rule`
2. **PUT**: `JSON → notificationruleupdate.Types → FromAPINotificationRuleUpdate() → resources.Rule`
3. **GET**: `resources.Rule → ToAPINotificationRuleResponse() → notificationruleresponse.Types → JSON`

## 🎯 지원하는 Notification 타입

### Email
```json
{
  "notification_type": "email",
  "configuration": {
    "to": ["user@example.com"],
    "cc": ["manager@example.com"],
    "subject": "Alert Notification",
    "body": "Alert details..."
  }
}
```

### Slack  
```json
{
  "notification_type": "slack", 
  "configuration": {
    "channel": "#alerts",
    "webhook_url": "https://hooks.slack.com/...",
    "message": "Alert: {{alert.message}}"
  }
}
```

### Teams
```json
{
  "notification_type": "teams",
  "configuration": {
    "webhook_url": "https://outlook.office.com/...",
    "title": "Critical Alert",
    "text": "Alert details..."
  }
}
```

### Webhook
```json
{
  "notification_type": "webhook",
  "configuration": {
    "url": "https://api.example.com/webhook",
    "method": "POST", 
    "headers": {
      "Authorization": "Bearer token"
    }
  }
}
```

## ✅ 검증된 기능

- ✅ 완전한 CRUD 작업에 shared/types 사용
- ✅ POST/PUT 전용 스키마 분리로 정확한 타입 검증
- ✅ 4가지 notification 타입 지원 (email, slack, teams, webhook)
- ✅ Frontend와 Backend 타입 동기화
- ✅ 22개 테스트 케이스 통과

## 🔗 관련 파일

- **Handler**: `core/pkg/notification/handler.go`
- **Adapter**: `core/pkg/notification/adapters/api_adapter.go`
- **Tests**: `core/pkg/notification/adapters/api_types_test.go`
- **Types**: `shared/types/notificationrule/`, `shared/types/notificationruleupdate/`, `shared/types/notificationruleresponse/`

## 🎯 장점

1. **정확한 타입 분리**: POST/PUT 요청별 전용 스키마로 정확한 검증
2. **다양한 채널 지원**: Email, Slack, Teams, Webhook 통합 지원  
3. **우선순위 관리**: Priority 필드로 알림 중요도 관리
4. **설정 유연성**: 각 notification 타입별 맞춤 설정 지원
5. **자동 동기화**: Schema 변경 시 Frontend/Backend 자동 반영