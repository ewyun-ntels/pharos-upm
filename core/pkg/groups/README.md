# Groups Package

사용자 그룹 관리를 위한 HTTP API와 Casbin 기반 권한 관리 시스템입니다.

## Features

- **권한 관리**: Casbin 기반 RBAC (Role-Based Access Control) 시스템
- **데이터베이스**: SQLite 기반 영구 저장소 지원
- **RESTful API**: 표준화된 HTTP API 엔드포인트
- **인증**: JWT 토큰 기반 인증 미들웨어
- **그룹 관리**: 그룹 생성, 조회, 삭제 기능
- **멤버 관리**: 그룹 내 사용자 멤버십 관리 (admin, member 역할)

## Architecture

```
┌─────────────┐    ┌──────────────┐    ┌────────────┐
│   Client    │───▶│   HTTP API   │───▶│  Handler   │
└─────────────┘    └──────────────┘    └────────────┘
                           │                   │
                    ┌──────▼──────┐    ┌──────▼──────┐
                    │ JWT Auth    │    │   Casbin    │
                    │ Middleware  │    │  Enforcer   │
                    └─────────────┘    └──────┬──────┘
                                              │
                                    ┌─────────▼─────────┐
                                    │ SQLite Database   │
                                    └───────────────────┘
```

## API Reference

### Groups API

#### 1. 모든 그룹 조회
- **Endpoint**: `GET /groups`
- **인증**: JWT Bearer Token 필요
- **응답**:
  ```json
  {
    "group-01": {
      "members": 2,
      "permission": "admin"
    },
    "group-02": {
      "members": 1,
      "permission": "admin"
    }
  }
  ```

#### 2. 특정 그룹 조회
- **Endpoint**: `GET /groups/:group-name`
- **인증**: JWT Bearer Token 필요
- **응답**:
  ```json
  {
    "members": 2,
    "permission": "admin"
  }
  ```

#### 3. 그룹 생성
- **Endpoint**: `POST /groups`
- **인증**: JWT Bearer Token 필요
- **요청 Body**:
  ```json
  {
    "name": "group-01"
  }
  ```
- **cURL 예제**:
  ```bash
  curl -X POST ${API_BASE_URL}/groups \
       -H "Authorization: Bearer ${JWT_TOKEN}" \
       -H "Content-Type: application/json" \
       -d '{"name":"group-01"}'
  ```

#### 4. 그룹 삭제
- **Endpoint**: `DELETE /groups/:group-name`
- **인증**: JWT Bearer Token 필요
- **권한**: 그룹 관리자 또는 Super Admin

### Members API

#### 멤버 역할 (Actions)
- `admin`: 그룹 관리자 권한
- `member`: 일반 멤버 권한

#### 1. 그룹 멤버 조회
- **Endpoint**: `GET /groups/:group-name/members`
- **인증**: JWT Bearer Token 필요
- **응답**:
  ```json
  [
    {
      "subject": "user-01",
      "object": "group-01", 
      "action": "admin"
    },
    {
      "subject": "user-02",
      "object": "group-01",
      "action": "member"
    }
  ]
  ```

#### 2. 멤버 추가/수정
- **Endpoint**: `PUT /groups/:group-name/members`
- **인증**: JWT Bearer Token 필요
- **요청 Body**:
  ```json
  {
    "subject": "user-02",
    "action": "member"
  }
  ```
- **cURL 예제**:
  ```bash
  curl -X PUT ${API_BASE_URL}/groups/group-01/members \
       -H "Authorization: Bearer ${JWT_TOKEN}" \
       -H "Content-Type: application/json" \
       -d '{"subject":"user-02","action":"member"}'
  ```

#### 3. 멤버 삭제
- **Endpoint**: `DELETE /groups/:group-name/members/:user-name`
- **인증**: JWT Bearer Token 필요
- **권한**: 그룹 관리자 또는 Super Admin

## Configuration

### Database Setup

#### Production Configuration
```go
import (
    "ntels.com/pharos/internal/casbin"
    "ntels.com/pharos/internal/orm"
)

// SQLite 파일 기반 설정
dbConfig := orm.DatabaseConfig{
    Driver: "sqlite",
    SQLite: orm.SQLiteConfig{
        Path: "/path/to/groups.db",
    },
}

enforcer, err := casbin.NewEnforcer(casbin.EnforcerTypeGroup, dbConfig)
```

#### Development/Test Configuration
```go
// 메모리 내 데이터베이스 (테스트용)
dbConfig := orm.DatabaseConfig{
    Driver: "sqlite",
    SQLite: orm.SQLiteConfig{
        Path: ":memory:",
    },
}

enforcer, err := casbin.NewEnforcer(casbin.EnforcerTypeGroup, dbConfig)
```

### Environment Variables
```bash
# API 서버 설정
SERVER_TYPE=master         # master, slave
SERVER_PORT=8080
SERVER_HOST=localhost

# 데이터베이스 설정
DB_DRIVER=sqlite
DB_SQLITE_PATH=/data/groups.db

# JWT 설정
JWT_SECRET=your-secret-key
JWT_ISSUER=pharos
JWT_EXPIRY=24h
```

### Module Loading
```go
import "ntels.com/pharos/pkg/groups"

// Gin 라우터에 그룹 API 등록
func main() {
    router := gin.Default()
    
    // 그룹 모듈 로드
    err := groups.Load("groups", config, router.Group("/api/v1"))
    if err != nil {
        log.Fatal(err)
    }
    
    router.Run(":8080")
}
```

## Security & Permissions

### JWT Authentication
- 모든 API 엔드포인트는 유효한 JWT 토큰이 필요합니다
- 토큰은 `Authorization: Bearer <token>` 헤더로 전달됩니다
- 토큰에는 사용자 정보와 권한이 포함되어 있습니다

### Role-Based Access Control (RBAC)
- **Super Admin**: 모든 그룹에 대한 전체 권한
- **Group Admin**: 특정 그룹에 대한 관리 권한
- **Group Member**: 특정 그룹의 읽기 권한

### Permission Matrix
| Operation | Super Admin | Group Admin | Group Member | Non-Member |
|-----------|-------------|-------------|--------------|------------|
| 그룹 조회 | ✅ | ✅ | ✅ | ❌ |
| 그룹 생성 | ✅ | ✅ | ❌ | ❌ |
| 그룹 삭제 | ✅ | ✅ | ❌ | ❌ |
| 멤버 조회 | ✅ | ✅ | ✅ | ❌ |
| 멤버 추가/수정 | ✅ | ✅ | ❌ | ❌ |
| 멤버 삭제 | ✅ | ✅ | ❌ | ❌ |

## Error Handling

### HTTP Status Codes
- `200 OK`: 성공적인 요청
- `400 Bad Request`: 잘못된 요청 데이터
- `401 Unauthorized`: 인증 실패
- `403 Forbidden`: 권한 없음
- `404 Not Found`: 리소스를 찾을 수 없음
- `500 Internal Server Error`: 서버 내부 오류

### Error Response Format
```json
{
  "message": "error description"
}
```

## Development

### Project Structure
```
pkg/groups/
├── handler.go          # HTTP 핸들러 구현
├── load.go             # 모듈 로딩 및 초기화
├── handler_test.go     # 핸들러 통합 테스트
├── load_test.go        # 로딩 테스트
└── README.md           # 이 문서
```

### Dependencies
```go
// 주요 의존성
github.com/gin-gonic/gin                    // HTTP 웹 프레임워크
github.com/ory/fosite/token/jwt            // JWT 토큰 처리
ntels.com/pharos/internal/casbin           // Casbin RBAC 구현
ntels.com/pharos/internal/orm              // ORM 및 데이터베이스 연결
ntels.com/pharos/pkg/auth                  // 인증 관련 유틸리티
```

### Contributing
1. 새로운 기능 추가 시 코드 리뷰 필수
2. 코드 변경 후 빌드 검증 확인
3. API 변경 시 README.md 업데이트
4. 성능 최적화 고려

## Troubleshooting

### 일반적인 문제

#### 1. SQLite 데이터베이스 연결 실패
```bash
Error: unable to open database file
```
**해결책**: 데이터베이스 파일 경로와 권한 확인

#### 2. JWT 토큰 검증 실패
```bash
Error: invalid or expired token
```
**해결책**: 토큰 유효성 및 만료 시간 확인

#### 3. 권한 부족 오류
```bash
Error: 403 Forbidden
```
**해결책**: 사용자의 그룹 멤버십 및 역할 확인

### 디버깅 팁
- 로그 레벨을 DEBUG로 설정하여 상세 정보 확인
- Casbin enforcer 정책 상태 확인: `enforcer.GetPolicy()`
- 데이터베이스 상태 직접 확인: SQLite CLI 도구 사용
