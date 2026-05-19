/**
 * Auth Store
 * 
 * 사용자 인증 및 권한 관리를 위한 Zustand Store
 * 
 * ## 특징
 * - Core와 Extension에서 공통 사용
 * - HttpOnly Cookie 대응 (API 호출 기반)
 * - Lifecycle callbacks 지원
 * - Schema 기반 타입 (shared/schema/user/User.schema)
 * 
 * ## 사용 방법
 * 
 * ### 1. Store 초기화 + 콜백 등록
 * ```typescript
 * import { useAuthStore } from '@pharos/shared/features/auth';
 * 
 * // App 시작 시
 * const initialize = useAuthStore(state => state.initialize);
 * await initialize(
 *   // GET /me API 호출
 *   async () => {
 *     const res = await fetch('/me', { credentials: 'include' });
 *     return await res.json();
 *   },
 *   // Lifecycle callbacks
 *   {
 *     onLogin: async (user) => {
 *       console.log('Logged in:', user.name);
 *       // 대시보드 초기화
 *       await dashboardStore.loadUserDashboards(user.name);
 *     },
 *     onLogout: async () => {
 *       console.log('Logging out...');
 *       // 대시보드 정리
 *       dashboardStore.reset();
 *     }
 *   }
 * );
 * ```
 * 
 * ### 2. 권한 체크
 * ```typescript
 * import { useHasPermission } from '@pharos/shared/features/auth';
 * 
 * function Dashboard() {
 *   const canAdd = useHasPermission('dashboard_add');
 *   return canAdd ? <CreateButton /> : null;
 * }
 * ```
 * 
 * ### 3. 여러 권한 체크
 * ```typescript
 * import { useHasAnyPermission } from '@pharos/shared/features/auth';
 * 
 * const canManage = useHasAnyPermission(['manage_user', 'super_admin']);
 * ```
 * 
 * ### 4. 로그아웃
 * ```typescript
 * const logout = useAuthStore(state => state.logout);
 * await logout(); // onLogout 콜백 자동 실행
 * ```
 */

export * from './auth-store';
export * from './types';
export * from './auth-action-types';

// Re-export User and Auth types for convenience
export type { User } from '../../types/user';
export type { PasswordPolicy, AuthConfig } from '../../types/auth';
