# Authorization Header Processing

이 패키지는 OAuth 2.0 환경에서 사용되는 다양한 Authorization 헤더 형식을 처리합니다.

## 지원되는 인증 방식

### 1. Bearer Token (JWT)

일반 사용자 인증용 JWT 토큰:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

- JWT의 `sub` 클레임에서 사용자 ID 추출
- 반환값: JWT subject (예: "user123")

### 2. Basic Auth (OAuth Client Authentication)

OAuth 2.0 refresh token grant에서 클라이언트 인증용:

```
Authorization: Basic Y2xpZW50X2lkOmNsaWVudF9zZWNyZXQ=
```

- Base64 디코딩 후 `client_id:client_secret` 형식 파싱
- client_id만 추출하여 반환
- 반환값: "client:" 접두사가 붙은 클라이언트 ID (예: "client:my_oauth_app")

## 사용 예제

```go
import (
"net/http"
"ntels.com/pharos/third_party/history/internal/store"
)

// HTTP 헤더에서 사용자/클라이언트 식별자 추출
func processRequest(header http.Header) {
subject, err := store.getJwtSubject(header)
if err != nil {
log.Printf("Authorization processing failed: %v", err)
return
}

if subject == "" {
log.Println("No authorization header found")
return
}

if strings.HasPrefix(subject, "client:") {
// OAuth 클라이언트 인증
clientId := strings.TrimPrefix(subject, "client:")
log.Printf("OAuth client request from: %s", clientId)
} else {
// 일반 사용자 인증
log.Printf("User request from: %s", subject)
}
}
```

## 에러 처리

함수는 다음과 같은 경우 에러를 반환합니다:

- Authorization 헤더 형식이 잘못된 경우
- Basic Auth의 Base64 디코딩 실패
- Basic Auth에서 client_id가 비어있는 경우
- Bearer JWT 토큰 파싱 실패
- 지원하지 않는 인증 타입 (Digest 등)

## RFC 준수

이 구현은 다음 RFC 표준을 준수합니다:

- RFC 6749: OAuth 2.0 Authorization Framework
- RFC 7617: HTTP Basic Authentication
- RFC 7519: JSON Web Token (JWT)
