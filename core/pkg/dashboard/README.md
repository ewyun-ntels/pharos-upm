# Dashboard API

이 문서는 pkg/dashboard 하위 API의 사용 방법을 설명합니다.

## 목차
- [기본 정보](#기본-정보)
- [엔드포인트 요약](#엔드포인트-요약)
- [오류 응답](#오류-응답)
- [대시보드 리소스](#대시보드-리소스)
- [IDs API](#ids-api)
- [Permission API](#permission-api)
- [Query API](#query-api)

## 기본 정보

**베이스 경로**: `/dashboard`  
**인증**: JWT 토큰 기반 인증 필요  
**권한**: 일부 작업은 특정 역할이 필요

### 공통 헤더
```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

### 권한 체계
- **SuperAdmin**: 모든 대시보드 관리 권한
- **DashboardAdd**: 새 대시보드 생성 권한
- **owner**: 대시보드 소유자 (모든 작업 가능)
- **editor**: 대시보드 편집 권한 (조회, 수정, 쿼리 실행)
- **viewer**: 대시보드 조회 권한 (조회, 미리 정의된 쿼리 실행만 가능)

## 엔드포인트 요약
| HTTP 메서드 | 엔드포인트 | 권한 요구사항 | 설명 |
|------------|-----------|--------------|------|
| `GET` | `/dashboard` | 인증된 사용자 | 접근 가능한 모든 대시보드 조회 |
| `GET` | `/dashboard/:id` | 대시보드 접근 권한 | 특정 대시보드 조회 |
| `POST` | `/dashboard` | SuperAdmin 또는 DashboardAdd | 새 대시보드 생성 |
| `PUT` | `/dashboard/:id` | owner 또는 editor | 대시보드 수정 |
| `DELETE` | `/dashboard/:id` | owner | 대시보드 삭제 |
| `GET` | `/dashboard/:id/permission` | owner | 대시보드 권한 조회 |
| `PUT` | `/dashboard/:id/permission` | owner | 대시보드 권한 설정 |
| `DELETE` | `/dashboard/:id/permission/:subject` | owner | 대시보드 권한 제거 |
| `POST` | `/dashboard/:id/query` | viewer 이상 | 대시보드 쿼리 실행 |
| `GET` | `/dashboard/ids` | 인증된 사용자 | 접근 가능한 대시보드 ID 목록 |

## 오류 응답

| 상태 코드 | 설명 | 응답 형식 |
|----------|------|-----------|
| `200` | 성공 | 요청된 데이터 |
| `201` | 생성 성공 | `{"id": "uuid"}` |
| `400` | 잘못된 요청 | `{"message": "오류 설명"}` |
| `401` | 인증 실패 | `{"message": "Unauthorized"}` |
| `403` | 권한 없음 | `{"message": "Forbidden"}` 또는 응답 바디 없음 |
| `404` | 리소스 없음 | `{"message": "Not Found"}` |
| `500` | 서버 오류 | `{"message": "Internal Server Error"}` |

---

## 대시보드 리소스

### `GET /dashboard`
현재 사용자가 접근 가능한 모든 대시보드를 반환합니다.

**권한**: 인증된 사용자  
**응답 코드**: `200 OK`

**응답 스키마**:
```json
{
  "dashboard_id": {
    "config": {}, // 대시보드 설정 객체
    "permission": "owner|editor|viewer"
  }
}
```

**응답 예시**:
```json
{
  "2efbd083-0379-4c36-88fe-80cec003ed0c": {
    "config": {
      "title": "시스템 모니터링",
      "panels": [...]
    },
    "permission": "owner"
  },
  "56175150-dcf0-41b8-b3f3-8791a3edefb0": {
    "config": {
      "title": "네트워크 대시보드",
      "panels": [...]
    },
    "permission": "editor"
  }
}
```

### `GET /dashboard/:id`
특정 대시보드를 반환합니다.

**권한**: 대시보드 접근 권한 (viewer 이상)  
**응답 코드**: `200 OK`

**경로 매개변수**:
- `id` (string): 대시보드 UUID

**응답 예시**:
```json
{
  "config": {
    "title": "시스템 모니터링",
    "panels": [...]
  },
  "permission": "owner"
}
```

### `POST /dashboard`
새 대시보드를 생성합니다.

**권한**: SuperAdmin 또는 DashboardAdd  
**응답 코드**: `201 Created`

**요청 바디**:
```json
{
  "title": "새 대시보드",
  "description": "대시보드 설명",
  "panels": [...]
}
```

**cURL 예시**:
```bash
curl -X POST ${DASHBOARD_API_URL}/dashboard \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "시스템 모니터링",
    "description": "서버 상태 모니터링 대시보드",
    "panels": []
  }'
```

**응답 예시**:
```json
{"id": "d11608c3-1e3f-4ce2-b447-bc7f359f1d89"}
```

### `PUT /dashboard/:id`
기존 대시보드를 수정합니다.

**권한**: owner 또는 editor  
**응답 코드**: `200 OK`

**경로 매개변수**:
- `id` (string): 대시보드 UUID

**요청 바디**: 대시보드 설정 객체 (JSON)

### `DELETE /dashboard/:id`
대시보드를 삭제합니다.

**권한**: owner  
**응답 코드**: `200 OK`

**경로 매개변수**:
- `id` (string): 대시보드 UUID

---

## IDs API

### `GET /dashboard/ids`
현재 사용자가 접근 가능한 대시보드 ID 목록을 반환합니다.

**권한**: 인증된 사용자  
**응답 코드**: `200 OK`

**응답 스키마**:
```json
[
  {"id": "dashboard_uuid"}
]
```

**응답 예시**:
```json
[
  {"id": "f2c4ec69-8465-45a4-ac17-0e7e581e1388"},
  {"id": "a10b3fc8-dcb1-4457-af62-df791c177d1a"}
]
```

---

## Permission API

권한 관리를 위한 API입니다.

### 권한 개념
- **action**: `owner`, `editor`, `viewer`
- **kind**: `user`, `group`
- **subject**: 사용자 ID, 그룹명, 또는 `__public` (공개 접근)

### `GET /dashboard/:id/permission`
대시보드의 모든 권한 정책을 조회합니다.

**권한**: owner  
**응답 코드**: `200 OK`

**경로 매개변수**:
- `id` (string): 대시보드 UUID

**응답 스키마**:
```json
[
  {
    "subject": "string",
    "object": "dashboard_uuid",
    "action": "owner|editor|viewer",
    "kind": "user|group"
  }
]
```

**응답 예시**:
```json
[
  {
    "subject": "__public",
    "object": "e09ee9fc-d52f-447a-9111-ae1fec3ca9be",
    "action": "viewer",
    "kind": "user"
  },
  {
    "subject": "user-01",
    "object": "e09ee9fc-d52f-447a-9111-ae1fec3ca9be",
    "action": "editor",
    "kind": "user"
  },
  {
    "subject": "user-02",
    "object": "e09ee9fc-d52f-447a-9111-ae1fec3ca9be",
    "action": "owner",
    "kind": "user"
  },
  {
    "subject": "admin-group",
    "object": "e09ee9fc-d52f-447a-9111-ae1fec3ca9be",
    "action": "owner",
    "kind": "group"
  }
]
```

### `PUT /dashboard/:id/permission`
특정 권한 정책을 설정합니다. 동일한 subject/object/kind 조합의 기존 정책은 제거 후 새로 추가됩니다.

**권한**: owner  
**응답 코드**: `200 OK`

**경로 매개변수**:
- `id` (string): 대시보드 UUID

**요청 바디**:
```json
{
  "subject": "user_id|group_name|__public",
  "action": "owner|editor|viewer",
  "kind": "user|group"
}
```

**cURL 예시**:
```bash
curl -X PUT ${DASHBOARD_API_URL}/dashboard/e09ee9fc-d52f-447a-9111-ae1fec3ca9be/permission \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "__public",
    "action": "viewer",
    "kind": "user"
  }'
```

### `DELETE /dashboard/:id/permission/:subject`
특정 subject의 대시보드 권한을 제거합니다.

**권한**: owner  
**응답 코드**: `200 OK`

**경로 매개변수**:
- `id` (string): 대시보드 UUID
- `subject` (string): 사용자 ID, 그룹명, 또는 `__public`

**쿼리 매개변수**:
- `kind` (string, 선택사항): `user` 또는 `group` (기본값: `user`)

**cURL 예시**:
```bash
# 사용자 권한 제거
curl -X DELETE "${DASHBOARD_API_URL}/dashboard/e09ee9fc-d52f-447a-9111-ae1fec3ca9be/permission/user-02?kind=user" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"

# 그룹 권한 제거
curl -X DELETE "${DASHBOARD_API_URL}/dashboard/e09ee9fc-d52f-447a-9111-ae1fec3ca9be/permission/admin-group?kind=group" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

---

## Query API

대시보드에서 데이터 쿼리를 실행하기 위한 API입니다.

### `POST /dashboard/:id/query`
대시보드 쿼리를 실행합니다.

**권한**: viewer 이상  
**응답 코드**: `200 OK`

**경로 매개변수**:
- `id` (string): 대시보드 UUID

**동작 방식**:
1. `run.datasourceName`과 `run.query`가 모두 제공된 경우: 직접 쿼리 실행
2. 둘 중 하나가 누락된 경우: `panel` 정보를 통해 미리 정의된 쿼리 실행
3. `viewer` 권한으로 `run.query`를 직접 명시하면 `403 Forbidden` 오류 발생

**요청 바디 스키마**:
```json
{
  "panel": {
    "kind": "string",
    "id": "string"
  },
  "run": {
    "datasourceName": "string",
    "query": "string"
  },
  "variables": {
    "variable_name": "value"
  }
}
```

**요청 예시**:
```json
{
  "panel": {
    "kind": "cards",
    "id": "6"
  },
  "run": {
    "datasourceName": "clickhouse-esn",
    "query": "SELECT Timestamp as Time, simpleJSONExtractString(Labels, 'name') as Name, Description, Severity, Labels as Information, if(Status = 'normal', 'cleared', Status) as Status FROM default.history_alert_row WHERE StartTimestamp BETWEEN toDateTime({{__start_time}}) AND toDateTime({{__end_time}}) ORDER BY StartTimestamp DESC LIMIT 80000"
  },
  "variables": {
    "__start_time": "5 Minutes ago",
    "__end_time": "Now"
  }
}
```

**응답 스키마**:
```json
{
  "meta": [
    {
      "name": "column_name",
      "type": "data_type"
    }
  ],
  "data": "array|null",
  "rows": "number",
  "statistics": {
    "elapsed": "number"
  }
}
```

**응답 예시**:
```json
{
  "meta": [
    {"name": "Time", "type": "DateTime"},
    {"name": "Name", "type": "Nullable(String)"},
    {"name": "Description", "type": "Nullable(String)"},
    {"name": "Severity", "type": "Nullable(String)"},
    {"name": "Information", "type": "Nullable(String)"},
    {"name": "Status", "type": "Nullable(String)"}
  ],
  "data": [
    ["2024-01-15 10:30:00", "서비스 A", "CPU 사용률 높음", "warning", "{\"instance\":\"server-01\"}", "active"]
  ],
  "rows": 1,
  "statistics": {
    "elapsed": 0.004324399
  }
}
```

**변수 사용법**:
- 쿼리에서 `{{variable_name}}` 형태로 변수 참조
- 시간 관련 변수 예시:
  - `{{__start_time}}`: 시작 시간
  - `{{__end_time}}`: 종료 시간
  - `{{__interval}}`: 시간 간격
