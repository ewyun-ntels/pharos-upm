import { AuthProvider } from '@/lib/data-provider';
import { useAuthStore } from '@pharos/shared/features/auth';
import { t } from 'i18next';
import { axiosInstance } from '@lib/axios';
import { UserSchema } from '@pharos/shared/types/user';
import { HOME_PATH } from '@pharos/meta/site-config';

// OAuth2 client_id (백엔드에서 필요)
const CLIENT_ID = '7a1e1782d4162e17dfc6f2bedd2c74ca';

/**
 * Auth Provider (HttpOnly 쿠키 기반)
 *
 * - 로그인: /auth/token 호출 → 백엔드가 Set-Cookie로 토큰 설정
 * - 인증 체크: /me 호출 → 쿠키 자동 전송
 * - 로그아웃: /auth/logout 호출 → 쿠키 삭제
 *
 * 프론트엔드는 토큰을 저장하지 않음 (HttpOnly 쿠키 사용)
 */
export const authProvider: AuthProvider = {
  login: async ({ email, password }) => {
    try {
      const params = new URLSearchParams();
      params.append('grant_type', 'password');
      params.append('username', email ?? '');
      params.append('password', password ?? '');
      params.append('client_id', CLIENT_ID);

      await axiosInstance.post('/auth/token', params, {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
      });

      // 쿠키가 자동 설정됨, 저장할 것 없음
      return { success: true, redirectTo: HOME_PATH };
    } catch (error: unknown) {
      console.error('Login error:', error);

      let errorMessage = t('msg.login.user_not_found');
      let returnedErrorMessage = '';

      if (typeof error === 'object' && error !== null && 'response' in error) {
        const axiosError = error as { response?: { data?: { error_description?: string } } };
        returnedErrorMessage = axiosError.response?.data?.error_description || '';
      }

      const msg = returnedErrorMessage.toLowerCase();
      if (msg.includes('user not found') || msg.includes('password mismatch')) {
        errorMessage = t('msg.login.user_not_found');
      } else if (msg.includes('user is blocked')) {
        errorMessage = t('msg.login.user_is_blocked');
      } else if (msg.includes('user is temporarily blocked')) {
        errorMessage = t('msg.login.user_is_temporarily_blocked');
      }

      return { success: false, error: { message: errorMessage, name: 'Invalid login' } };
    }
  },

  logout: async () => {
    try {
      // 토큰 무효화 + 쿠키 삭제
      await axiosInstance.post('/auth/revoke');
    } catch (error) {
      console.error('Logout API error:', error);
    }

    // logged_in 쿠키 명시적 삭제 (backend에서도 삭제하지만 즉시 반영)
    document.cookie = 'logged_in=; path=/; max-age=0';

    // auth store 정리 (등록된 콜백 실행)
    await useAuthStore.getState().logout();

    window.location.href = '/ui/login';
    return { success: true };
  },

  check: async () => {
    if (typeof window === 'undefined') return { authenticated: false };
    if (window.location.pathname === '/ui/login') return { authenticated: false };

    try {
      // /me API가 성공하면 인증됨 (쿠키 자동 전송)
      const res = await axiosInstance.get('/me', { withCredentials: true });
      const parsed = UserSchema.safeParse(res.data);
      // prepare 항목이 있으면 사전 처리 액션이 필요한 상태 → 로그인 페이지로
      if (parsed.success && (parsed.data.prepare?.length ?? 0) > 0) {
        return { authenticated: false, redirectTo: '/login' };
      }
      return { authenticated: true };
    } catch {
      return { authenticated: false, redirectTo: '/login' };
    }
  },

  getIdentity: async () => {
    try {
      const response = await axiosInstance.get('/me', { withCredentials: true });
      const parsed = UserSchema.safeParse(response.data);
      return parsed.success ? parsed.data : null;
    } catch {
      return null;
    }
  },

  getPermissions: async () => {
    console.warn('getPermissions is deprecated. Use useAuthStore hasPermission instead.');
    return null;
  },

  onError: async (error) => {
    console.error('Auth error:', error);
    return { error };
  },
};
