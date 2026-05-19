import {create} from 'zustand';
import {devtools} from 'zustand/middleware';
import type {AuthCallbacks, AuthState, FetchUserFn} from './types';
import type {User} from '../../types/user';
import {PermissionKeys} from '../../types/role';

/**
 * Permission type - all available permission keys (roles and attributes)
 */
export type Permission = PermissionKeys;

/**
 * Auth Store
 * 
 * 사용자 인증 및 권한 정보를 관리하는 전역 Store
 * 
 * - Core, Extension 모두에서 사용 가능
 * - HttpOnly Cookie 대응 (API 호출 기반)
 * - Lifecycle callbacks 지원 (로그인/로그아웃 시 자동 실행)
 * - Schema 기반 타입 (shared/schema/user/User.schema)
 * 
 * @example
 * // 초기화 + 콜백 등록
 * const initialize = useAuthStore(state => state.initialize);
 * await initialize(
 *   async () => {
 *     const res = await fetch('/me', { credentials: 'include' });
 *     return await res.json();
 *   },
 *   {
 *     onLogin: async (user) => {
 *       await dashboardStore.loadUserDashboards(user.name);
 *     },
 *     onLogout: async () => {
 *       dashboardStore.reset();
 *     }
 *   }
 * );
 * 
 * @example
 * // 권한 체크
 * const hasPermission = useAuthStore(state => state.hasPermission);
 * const canAdd = hasPermission('dashboard_add');
 * 
 * @example
 * // 여러 권한 체크
 * const hasAny = useAuthStore(state => state.hasAnyPermission);
 * const canManage = hasAny(['manage_user', 'super_admin']);
 */
/**
 * super_admin 권한 체크 헬퍼 함수
 */
const isSuperAdmin = (user: User | null): boolean => {
  if (!user?.attributes?.roles) return false;
  const super_admin: Permission = 'role:super_admin';
  return user.attributes.roles[super_admin];
};

export const useAuthStore = create<AuthState>()(
  devtools(
    (set, get) => ({
      // ========================================
      // State
      // ========================================
      user: null,
      isAuthenticated: false,
      isLoading: false,
      isInitialized: false,
      error: null,
      callbacks: {},
      onLogoutListeners: [],

      // ========================================
      // Actions
      // ========================================
      
      /**
       * Store 초기화
       */
      initialize: async (fetchUser: FetchUserFn, callbacks?: AuthCallbacks) => {
        // 이미 초기화 중이면 중복 호출 방지
        if (get().isLoading) {
          console.warn('Auth store is already initializing');
          return;
        }

        // 콜백 등록
        if (callbacks) {
          set({ callbacks });
        }

        set({ isLoading: true, error: null });

        try {
          const user = await fetchUser();
          
          if (user) {
            set({ 
              user,
              isAuthenticated: true,
              isInitialized: true,
              error: null,
            });
            
            // onLogin 콜백 실행
            const { callbacks: registeredCallbacks } = get();
            if (registeredCallbacks.onLogin) {
              try {
                await registeredCallbacks.onLogin(user);
              } catch (err) {
                console.error('onLogin callback error:', err);
              }
            }
          } else {
            set({ 
              user: null,
              isAuthenticated: false,
              isInitialized: true,
              error: null,
            });
          }
        } catch (error) {
          console.error('Failed to initialize auth store:', error);
          set({ 
            user: null,
            isInitialized: true,
            error: error instanceof Error ? error : new Error('Unknown error'),
          });
        } finally {
          set({ isLoading: false });
        }
      },

      /**
       * 사용자 정보 직접 설정
       */
      setUser: (user: User | null, triggerCallback = false) => {
        const prevUser = get().user;
        set({ 
          user,
          isAuthenticated: !!user,
        });
        
        // onUserUpdate 콜백 실행
        if (triggerCallback && user && user !== prevUser) {
          const { callbacks } = get();
          if (callbacks.onUserUpdate) {
            Promise.resolve(callbacks.onUserUpdate(user)).catch(err => {
              console.error('onUserUpdate callback error:', err);
            });
          }
        }
      },

      /**
       * 로그아웃 (콜백 실행 포함)
       */
      logout: async () => {
        const { callbacks, onLogoutListeners } = get();

        // 기존 onLogout 콜백 실행
        if (callbacks.onLogout) {
          try {
            await callbacks.onLogout();
          } catch (err) {
            console.error('onLogout callback error:', err);
          }
        }

        // 동적으로 등록된 모든 logout 리스너 실행
        for (const listener of onLogoutListeners) {
          try {
            await listener();
          } catch (err) {
            console.error('onLogout listener error:', err);
          }
        }

        // Store 초기화
        set({
          user: null,
          isAuthenticated: false,
          isInitialized: false,
          error: null,
        });
      },

      /**
       * Store 초기화 (콜백 없이)
       */
      reset: () => {
        set({
          user: null,
          isAuthenticated: false,
          isInitialized: false,
          error: null,
        });
      },

      /**
       * 콜백 등록/업데이트
       */
      setCallbacks: (callbacks: AuthCallbacks) => {
        set({ callbacks });
      },

      /**
       * logout 콜백 등록
       */
      registerOnLogout: (callback: () => void | Promise<void>) => {
        set((state) => ({
          onLogoutListeners: [...state.onLogoutListeners, callback],
        }));

        // unsubscribe 함수 반환
        return () => {
          set((state) => ({
            onLogoutListeners: state.onLogoutListeners.filter((cb) => cb !== callback),
          }));
        };
      },

      // ========================================
      // Permission Checkers
      // ========================================
      
      /**
       * 특정 권한 보유 여부 확인
       *
       * super_admin인 경우 모든 권한을 가진 것으로 간주합니다.
       * user.attributes.roles에서 해당 키의 값이 true인지 확인합니다.
       *
       * Backend attributes 구조:
       * {
       *   "roles": {
       *     "role:super_admin": true,
       *     "role:manage_user": false
       *   },
       *   "info": {
       *     "department": "engineering"
       *   }
       * }
       *
       * roles 객체에 선언된 키는 권한이 있는 것으로 취급합니다.
       */
      hasPermission: (permission: string) => {
        const user = get().user;
        if (!user?.attributes?.roles) return false;

        // super_admin은 모든 권한을 가짐
        if (isSuperAdmin(user)) return true;

        // 권한이 roles 객체에 선언되어 있으면 권한이 있는 것으로 판단
        return permission in user.attributes.roles;
      },

      /**
       * 여러 권한 중 하나라도 보유 (OR)
       * super_admin인 경우 항상 true를 반환합니다.
       */
      hasAnyPermission: (permissions: Permission[]) => {
        if (permissions.length === 0) return false;
        
        const user = get().user;
        // super_admin은 모든 권한을 가짐
        if (isSuperAdmin(user)) return true;

        const { hasPermission } = get();
        return permissions.some(permission => hasPermission(permission));
      },

      /**
       * 모든 권한 보유 (AND)
       * super_admin인 경우 항상 true를 반환합니다.
       */
      hasAllPermissions: (permissions: Permission[]) => {
        if (permissions.length === 0) return false;
        
        const user = get().user;
        // super_admin은 모든 권한을 가짐
        if (isSuperAdmin(user)) return true;

        const { hasPermission } = get();
        return permissions.every(permission => hasPermission(permission));
      },
    }),
    { 
      name: 'auth-store',
      enabled: process.env.NODE_ENV === 'development',
    }
  )
);

/**
 * Auth Store Hooks (React 환경에서 사용)
 */

/**
 * 현재 사용자 정보 가져오기
 */
export const useUser = () => useAuthStore(state => state.user);

/**
 * 특정 권한 보유 여부 확인 Hook
 *
 * @example
 * const canAdd = useHasPermission('dashboard_add');
 */
export const useHasPermission = (permission: string) =>
  useAuthStore(state => state.hasPermission(permission));

/**
 * 로그아웃 시 실행될 콜백 등록
 *
 * @example
 * // store 파일에서 등록
 * registerOnLogout(() => {
 *   useMyStore.getState().reset();
 * });
 */
export const registerOnLogout = (callback: () => void | Promise<void>) =>
  useAuthStore.getState().registerOnLogout(callback);
