# User Schema & Auth Store

사용자 인증 및 권한 관리를 위한 Schema 기반 타입 시스템과 Auth Store입니다.

## 📋 목차

- [Schema 구조](#schema-구조)
- [타입 생성](#타입-생성)
- [Auth Store 사용법](#auth-store-사용법)
- [Lifecycle Callbacks](#lifecycle-callbacks)
- [권한 체크](#권한-체크)
- [HttpOnly Cookie 대응](#httponly-cookie-대응)
- [테스트](#테스트)

---

## Schema 구조

### User.schema (Backend GET /me Response)

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "User",
  "description": "User details from Backend GET /me endpoint",
  "properties": {
    "name": {
      "type": "string",
      "description": "Username"
    },
    "created_at": {
      "type": ["string", "null"],
      "format": "date-time"
    },
    "attributes": {
      "type": "object",
      "additionalProperties": true,
      "description": "Flexible user attributes including permissions"
    },
    "blocked": {
      "type": ["boolean", "null"]
    },
    "prepare": {
      "type": "array",
      "items": { "type": "string" }
    },
    "password_expired_at": {
      "type": ["string", "null"],
      "format": "date-time"
    },
    "temporary_blocked_expires_at": {
      "type": ["string", "null"],
      "format": "date-time"
    }
  },
  "required": ["name"]
}
```

### 핵심 특징

1. **additionalProperties: true** - Panel, Alert처럼 유연한 구조
2. **attributes 필드** - 권한 및 추가 속성 저장
3. **Backend와 동일** - `core/pkg/userhandler/resources.go` 구조

### Example Response

```json
{
  "name": "admin",
  "created_at": "2025-01-01T00:00:00Z",
  "attributes": {
    "super_admin": true,
    "manage_user": true,
    "dashboard_add": false,
    "group_name": "admins",
    "custom_field": "any_value"
  },
  "blocked": false,
  "prepare": [],
  "password_expired_at": "2025-12-31T23:59:59Z",
  "temporary_blocked_expires_at": null
}
```

**권한**: `attributes`의 boolean 값만 권한으로 취급  
**추가속성**: 다른 타입은 추가 속성

---

## 타입 생성

### 자동 생성 프로세스 (Quicktype)

```bash
# 1. Schema 수정
vim shared/schema/user/User.schema

# 2. 타입 자동 생성
cd shared/frontend
npm run generate:types

# 3. 생성된 타입 확인
cat types/user/index.ts
```

### 생성된 타입

```typescript
// shared/frontend/types/user/index.ts
export const IndexSchema = z.object({
  "attributes": z.record(z.string(), z.any()).optional(),
  "blocked": z.union([z.boolean(), z.null()]).optional(),
  "created_at": z.union([z.coerce.date(), z.null()]).optional(),
  "name": z.string(),
  // ... 기타 필드
});

export type Index = z.infer<typeof IndexSchema>;
```

**Note**: Quicktype이 `Index` 타입명을 생성하므로 `User`로 alias하여 사용

---

## Auth Store 사용법

### 위치

```
shared/frontend/hooks/use-auth/
  ├── auth-store.ts         # Zustand Store
  ├── types.ts              # Store 타입
  ├── index.ts              # Exports
  ├── README.md             # 상세 가이드
  └── __tests__/
      └── use-auth-store.test.ts
```

### Import

```typescript
// Core에서
import { useAuthStore } from '@pharos/shared/features/auth';

// Extension에서도 동일
import { useAuthStore } from '@pharos/shared/features/auth';
```

### 기본 사용

```typescript
import { useAuthStore } from '@pharos/shared/features/auth';
import { useEffect } from 'react';

function App() {
  const initialize = useAuthStore(state => state.initialize);

  useEffect(() => {
    // GET /me API 호출
    initialize(async () => {
      const res = await fetch('/me', { credentials: 'include' });
      return await res.json();
    });
  }, []);

  return <>{children}</>;
}
```

---

## Lifecycle Callbacks

### onLogin - 로그인 시 자동 실행

```typescript
import { useAuthStore } from '@pharos/shared/features/auth';
import { useDashboardStore } from '@stores/dashboard-store';

function AuthInitializer() {
  const initialize = useAuthStore(state => state.initialize);
  const dashboardStore = useDashboardStore();

  useEffect(() => {
    initialize(
      async () => {
        const res = await fetch('/me', { credentials: 'include' });
        return await res.json();
      },
      {
        // 🎯 로그인 시 대시보드 초기화
        onLogin: async (user) => {
          console.log('User logged in:', user.name);
          
          // 대시보드 로드
          await dashboardStore.loadUserDashboards(user.name);
          
          // 사용자 설정 로드
          await loadUserPreferences(user.name);
          
          // 알림 시스템 초기화
          await notificationStore.initialize(user.name);
        }
      }
    );
  }, []);

  return null;
}
```

### onLogout - 로그아웃 시 자동 실행

```typescript
const logout = useAuthStore(state => state.logout);

useEffect(() => {
  initialize(
    fetchUser,
    {
      // 🎯 로그아웃 시 정리 작업
      onLogout: async () => {
        console.log('User logging out...');
        
        // 대시보드 정리
        dashboardStore.reset();
        
        // 알림 정리
        notificationStore.reset();
        
        // 캐시 정리
        localStorage.clear();
        sessionStorage.clear();
      }
    }
  );
}, []);

// 호출 시 onLogout 자동 실행
await logout();
```

### onUserUpdate - 사용자 정보 업데이트 시

```typescript
const setUser = useAuthStore(state => state.setUser);

useEffect(() => {
  initialize(
    fetchUser,
    {
      // 🎯 권한 변경 시 대응
      onUserUpdate: async (user) => {
        console.log('User updated:', user);
        
        // 권한 변경 시 대시보드 다시 로드
        if (user.attributes?.super_admin) {
          await loadAdminPanels();
        }
      }
    }
  );
}, []);

// 호출 시 onUserUpdate 자동 실행
setUser(updatedUser);
```

---

## 권한 체크

### hasPermission - 단일 권한

```typescript
const hasPermission = useAuthStore(state => state.hasPermission);

if (hasPermission('super_admin')) {
  // Admin 전용 기능
}
```

### hasAnyPermission - OR 조건

```typescript
const hasAnyPermission = useAuthStore(state => state.hasAnyPermission);

if (hasAnyPermission(['manage_user', 'super_admin'])) {
  // 둘 중 하나라도 있으면
}
```

### hasAllPermissions - AND 조건

```typescript
const hasAllPermissions = useAuthStore(state => state.hasAllPermissions);

if (hasAllPermissions(['manage_user', 'dashboard_add'])) {
  // 모두 있어야만
}
```

### RoleGuard 컴포넌트

```typescript
// components/user/RoleGuard.tsx
import { useAuthStore } from '@pharos/shared/features/auth';

export function RoleGuard({ 
  permission, 
  children 
}: { 
  permission: string; 
  children: React.ReactNode;
}) {
  const hasPermission = useAuthStore(state => state.hasPermission);
  return hasPermission(permission) ? <>{children}</> : null;
}

// 사용
<RoleGuard permission="super_admin">
  <AdminPanel />
</RoleGuard>
```

---

## HttpOnly Cookie 대응

### ⚠️ 중요: JWT Token 사용 불가

HttpOnly Cookie를 사용하면 JavaScript에서 JWT Token에 접근할 수 없습니다.

```typescript
// ❌ 불가능
const token = user.access_token; // undefined
const decoded = decodeToken(token); // 에러

// ✅ 유일한 방법: API 호출
const response = await fetch('/me', { credentials: 'include' });
const userData = await response.json();
const permissions = userData.attributes;
```

### 올바른 구현

```typescript
// providers/auth-provider.tsx
export const authProvider: AuthBindings = {
  getIdentity: async () => {
    // ✅ GET /me API로 사용자 정보 조회
    const response = await axiosInstance.get('/me');
    return {
      id: response.data.name,
      name: response.data.name,
      permissions: response.data.attributes || {},
    };
  },
  
  logout: async () => {
    // ✅ Auth Store의 logout 호출 (onLogout callback 자동 실행)
    await useAuthStore.getState().logout();
    
    // Backend 로그아웃 API
    await axiosInstance.post('/logout');
  }
};
```

---

## Store API Reference

### State

```typescript
interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  callbacks: AuthCallbacks | null;
}
```

### Actions

```typescript
interface AuthActions {
  // 초기화 (callbacks 등록)
  initialize(
    fetchUser: () => Promise<User>, 
    callbacks?: AuthCallbacks
  ): Promise<void>;
  
  // 로그아웃 (onLogout 자동 실행)
  logout(): Promise<void>;
  
  // 사용자 정보 업데이트 (onUserUpdate 자동 실행)
  setUser(user: User | null): void;
  
  // Store 초기화
  reset(): void;
  
  // 권한 체크
  hasPermission(permission: string): boolean;
  hasAnyPermission(permissions: string[]): boolean;
  hasAllPermissions(permissions: string[]): boolean;
}
```

### Callbacks

```typescript
interface AuthCallbacks {
  onLogin?: (user: User) => void | Promise<void>;
  onLogout?: () => void | Promise<void>;
  onUserUpdate?: (user: User) => void | Promise<void>;
}
```

---

## 테스트

```bash
cd shared/frontend
npm test hooks/use-auth
```

### 테스트 커버리지

- ✅ initialize with callbacks
- ✅ logout with callback
- ✅ setUser with callback
- ✅ hasPermission
- ✅ hasAnyPermission
- ✅ hasAllPermissions
- ✅ Error handling

---

## 다음 단계 (Core 구현)

### 1. authProvider 개선

```typescript
// providers/auth-provider.tsx
getIdentity: async () => {
  const response = await axiosInstance.get('/me');
  return response.data;
}
```

### 2. RoleGuard 구현

```typescript
// components/user/RoleGuard.tsx
import { useAuthStore } from '@pharos/shared/features/auth';

export function RoleGuard({ permission, children }) {
  const hasPermission = useAuthStore(state => state.hasPermission);
  return hasPermission(permission) ? <>{children}</> : null;
}
```

### 3. AuthInitializer 통합

```typescript
// app/_refine_context.tsx
import { useAuthStore } from '@pharos/shared/features/auth';

function AuthInitializer({ children }) {
  const initialize = useAuthStore(state => state.initialize);
  
  useEffect(() => {
    initialize(
      () => authProvider.getIdentity?.(),
      {
        onLogin: async (user) => {
          // 대시보드 초기화
        },
        onLogout: async () => {
          // 대시보드 정리
        }
      }
    );
  }, []);
  
  return <>{children}</>;
}
```

---

## 참고 문서

- **[core/frontend/src/components/user/SUMMARY.md](../../../core/frontend/src/components/user/SUMMARY.md)** - 전체 요약
- **[shared/frontend/hooks/use-auth/README.md](../../frontend/hooks/use-auth/README.md)** - Auth Store 상세 가이드
- **[Backend: core/pkg/userhandler/resources.go](../../../core/pkg/userhandler/resources.go)** - GET /me Response 구조

---

**작성일**: 2025-01-18  
**상태**: ✅ Schema + Auth Store + Lifecycle Callbacks + Tests 완성  
**Import Path**: `@pharos/shared/features/auth`
