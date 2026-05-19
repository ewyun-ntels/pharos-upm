# User Role Schema Documentation

사용자 역할(Role)과 속성(Attribute) 시스템의 JSON Schema 정의와 사용법을 설명합니다.

## 📋 Schema Files

### 1. UserRole.schema
사용자 역할과 속성 정의를 위한 스키마입니다.

**위치**: `shared/schema/role/UserRole.schema`  
**용도**: 시스템 전반의 역할 및 속성 관리  
**생성 타입**: `shared/types/userrole.Types`

## 🏗️ 아키텍처 패턴

```
Frontend ←→ shared/types ←→ Backend Services
            ↓ 자동 생성     ↓ 역할 기반 접근 제어
        JSON Schema      RBAC System
```

## 📖 역할(Roles) 정의

### 관리자 역할
- **`role:super_admin`**: 시스템 전체 관리 권한
- **`role:manage_user`**: 사용자 관리 권한
- **`role:manage_notification`**: 알림 관리 권한  
- **`role:manage_alert`**: 경고 관리 권한

### 기능별 역할
- **`role:dashboard_add`**: 대시보드 추가 권한
- **`role:group_add`**: 그룹 추가 권한

## 📖 속성(Attributes) 정의

### 사용자 속성
- **`attr:temporary_user`**: 임시 사용자 표시
- **`attr:skip_temporarily_block`**: 임시 차단 건너뛰기
- **`attr:password_retry_unlimited`**: 무제한 비밀번호 재시도

## 📖 사용 예제

### TypeScript/JavaScript
```typescript
import { UserRole } from '@pharos/shared';

// 역할 체크
const isSuperAdmin = userRole === UserRole.roles.super_admin;
const canManageUsers = userRole === UserRole.roles.manage_user;

// 속성 체크  
const isTemporary = userAttributes.includes(UserRole.attributes.temporary_user);
```

### Go
```go
import "ntels.com/pharos/shared/types/userrole"

// 역할 체크
if user.Role == userrole.Types.Roles.SuperAdmin {
    // super admin logic
}

// 속성 체크
if slices.Contains(user.Attributes, userrole.Types.Attributes.TemporaryUser) {
    // temporary user logic  
}
```

## 🔒 권한 관리 예제

### 미들웨어에서 역할 체크
```go
func RequireRole(requiredRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := getUserRole(c)
        
        for _, role := range requiredRoles {
            if userRole == role {
                c.Next()
                return
            }
        }
        
        c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
        c.Abort()
    }
}

// 사용 예시
router.DELETE("/user/:id", 
    RequireRole(userrole.Types.Roles.SuperAdmin, userrole.Types.Roles.ManageUser),
    deleteUserHandler)
```

### 조건부 UI 렌더링
```tsx
import { UserRole } from '@pharos/shared';

function AdminPanel({ userRole }: { userRole: string }) {
    if (userRole === UserRole.roles.super_admin) {
        return <SuperAdminPanel />;
    }
    
    if (userRole === UserRole.roles.manage_user) {
        return <UserManagementPanel />;
    }
    
    return <AccessDenied />;
}
```

## ✅ 검증된 기능

- ✅ 일관된 역할 명명 규칙 (`role:` 접두사)
- ✅ 일관된 속성 명명 규칙 (`attr:` 접두사)  
- ✅ Frontend와 Backend 타입 동기화
- ✅ 중앙화된 권한 관리
- ✅ 타입 안전성 보장

## 🔗 관련 파일

- **Schema**: `shared/schema/role/UserRole.schema`
- **Types**: `shared/types/userrole/`
- **Auth Handler**: `core/pkg/authhandler/`
- **User Handler**: `core/pkg/userhandler/`

## 🎯 장점

1. **중앙 집중식 관리**: 모든 역할과 속성을 한 곳에서 관리
2. **일관된 명명**: `role:`, `attr:` 접두사로 명확한 구분
3. **타입 안전성**: 컴파일 타임에 역할/속성 검증
4. **자동 동기화**: Schema 변경 시 Frontend/Backend 자동 반영
5. **확장성**: 새로운 역할/속성 추가 용이
6. **문서화**: 각 역할과 속성의 목적이 명확히 기술됨