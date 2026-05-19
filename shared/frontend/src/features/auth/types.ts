import { User } from '../../types/user';
import {Permission} from './auth-store';

/**
 * Auth Lifecycle Callbacks
 * 
 * 로그인/로그아웃 시 자동 실행할 콜백 함수들
 */
export interface AuthCallbacks {
  /**
   * 로그인 성공 후 자동 실행
   * 
   * @example
   * onLogin: async (user) => {
   *   // 대시보드 초기화
   *   await dashboardStore.loadUserDashboards(user.name);
   *   // 사용자 설정 로드
   *   await loadUserPreferences(user.name);
   * }
   */
  onLogin?: (user: User) => void | Promise<void>;
  
  /**
   * 로그아웃 전 자동 실행
   * 
   * @example
   * onLogout: async () => {
   *   // 대시보드 정리
   *   dashboardStore.reset();
   *   // 캐시 정리
   *   clearLocalCache();
   * }
   */
  onLogout?: () => void | Promise<void>;
  
  /**
   * 사용자 정보 업데이트 시 자동 실행
   * 
   * @example
   * onUserUpdate: async (user) => {
   *   // 권한 변경 시 대시보드 다시 로드
   *   if (user.attributes?.super_admin) {
   *     await loadAdminPanels();
   *   }
   * }
   */
  onUserUpdate?: (user: User) => void | Promise<void>;
}

/**
 * Token Response (OAuth2 Standard)
 */
export interface TokenResponse {
  access_token?: string;
  refresh_token?: string;
  token_type?: string;
  expires_in?: number;
}

/**
 * User Profile (JWT Standard Claims)
 */
export interface UserProfile {
  sub: string;
  name: string;
  email?: string;
  iss: string;
  aud: string[];
  iat: number;
  exp: number;
  [key: string]: unknown; // JWT의 다른 임의의 필드 허용
}

/**
 * User Identity (Standardized UI representation)
 */
export interface UserIdentity {
  id: string;
  name: string;
  email: string;
  accessToken?: string;
  expiresAt?: number;
  refreshToken?: string;
}

/**
 * Auth Store State
 * 
 * 사용자 인증 및 권한 정보 관리
 */
export interface AuthState {
  // ========================================
  // State
  // ========================================
  
  /**
   * 현재 로그인한 사용자 정보 (Backend GET /me Response)
   */
  user: User | null;
  
  /**
   * 인증 여부
   */
  isAuthenticated: boolean;

  /**
   * 로딩 상태
   */
  isLoading: boolean;
  
  /**
   * 초기화 완료 여부
   */
  isInitialized: boolean;
  
  /**
   * 에러 정보
   */
  error: Error | null;
  
  /**
   * 등록된 콜백
   */
  callbacks: AuthCallbacks;

  /**
   * 동적으로 등록된 logout 콜백 리스트
   */
  onLogoutListeners: Array<() => void | Promise<void>>;

  // ========================================
  // Actions
  // ========================================
  
  /**
   * Store 초기화
   * 
   * @param fetchUser - 사용자 정보를 가져오는 함수 (GET /me 호출)
   * @param callbacks - 로그인/로그아웃 시 실행할 콜백 (선택)
   * 
   * @example
   * // Refine authProvider 사용
   * initialize(async () => {
   *   const res = await fetch('/me', { credentials: 'include' });
   *   return await res.json();
   * });
   * 
   * @example
   * // 콜백과 함께 사용
   * initialize(
   *   async () => {
   *     const res = await fetch('/me', { credentials: 'include' });
   *     return await res.json();
   *   },
   *   {
   *     onLogin: async (user) => {
   *       console.log('Logged in:', user.name);
   *       await dashboardStore.loadUserDashboards(user.name);
   *     },
   *     onLogout: async () => {
   *       console.log('Cleaning up...');
   *       dashboardStore.reset();
   *     }
   *   }
   * );
   */
  initialize: (
    fetchUser: () => Promise<User | null>,
    callbacks?: AuthCallbacks
  ) => Promise<void>;
  
  /**
   * 사용자 정보 직접 설정
   * 
   * @param user - 설정할 사용자 정보
   * @param triggerCallback - onUserUpdate 콜백 실행 여부 (기본: false)
   */
  setUser: (user: User | null, triggerCallback?: boolean) => void;
  
  /**
   * 로그아웃 (콜백 실행 포함)
   * 
   * onLogout 콜백을 실행한 후 Store를 초기화합니다.
   */
  logout: () => Promise<void>;
  
  /**
   * Store 초기화 (콜백 없이)
   */
  reset: () => void;
  
  /**
   * 콜백 등록/업데이트
   */
  setCallbacks: (callbacks: AuthCallbacks) => void;

  /**
   * logout 콜백 등록
   *
   * 각 feature에서 logout 시 실행될 cleanup 로직을 등록합니다.
   * 반환된 함수를 호출하면 등록 해제됩니다.
   *
   * @param callback - logout 시 실행될 콜백
   * @returns unsubscribe 함수
   *
   * @example
   * // feature에서 등록
   * useEffect(() => {
   *   const unsubscribe = useAuthStore.getState().registerOnLogout(() => {
   *     myStore.reset();
   *   });
   *   return unsubscribe;
   * }, []);
   */
  registerOnLogout: (callback: () => void | Promise<void>) => () => void;

  // ========================================
  // Permission Checkers
  // ========================================
  
  /**
   * 특정 권한 보유 여부 확인
   *
   * super_admin인 경우 모든 권한을 가진 것으로 간주합니다.
   * user.attributes.roles에서 해당 키의 값이 true인지 확인
   *
   * @param permission - 확인할 권한 이름 (string으로 동적 권한 지원)
   * @returns boolean 값만 권한으로 취급
   *
   * @example
   * const canAdd = hasPermission('dashboard_add');
   */
  hasPermission: (permission: string) => boolean;
  
  /**
   * 여러 권한 중 하나라도 보유 (OR)
   * super_admin인 경우 항상 true를 반환합니다.
   * 
   * @example
   * const canManage = hasAnyPermission(['manage_user', 'super_admin']);
   */
  hasAnyPermission: (permissions: Permission[]) => boolean;
  
  /**
   * 모든 권한 보유 (AND)
   * super_admin인 경우 항상 true를 반환합니다.
   * 
   * @example
   * const canFullAccess = hasAllPermissions(['manage_user', 'manage_alert']);
   */
  hasAllPermissions: (permissions: Permission[]) => boolean;
}

/**
 * FetchUser 함수 타입
 */
export type FetchUserFn = () => Promise<User | null>;
