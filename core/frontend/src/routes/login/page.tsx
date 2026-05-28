import { authActionPluginRegistry } from '@/features/login';
import type { AuthActionContext, LoginResult, ChangePasswordResult } from '@/features/login';
import { useUser, useAuthStore } from '@pharos/shared/features/auth';
import { axiosInstance } from '@lib/axios';
import { useNavigate } from 'react-router-dom';
import { t } from 'i18next';
import { userProvider, USER_RESOURCES } from '@providers/user-provider';
import { UserSchema } from '@pharos/shared/types/user';
import { useDashboardStore } from '@features/dashboard/hooks/use-dashboard-store';

// OAuth2 client_id
const CLIENT_ID = '7a1e1782d4162e17dfc6f2bedd2c74ca';

export default function AuthPage() {
  const user = useUser();
  const navigate = useNavigate();
  const setUser = useAuthStore((state) => state.setUser);

  // 현재 액션 결정: prepare가 있으면 해당 액션, 없으면 login
  const currentAction = user?.prepare?.[0] || 'login';

  // 로그인 처리 (Core에서 구현)
  const handleLogin = async (username: string, password: string): Promise<LoginResult> => {
    try {
      const params = new URLSearchParams();
      params.append('grant_type', 'password');
      params.append('username', username);
      params.append('password', password);
      params.append('client_id', CLIENT_ID);

      await axiosInstance.post('/auth/token', params, {
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      });

      return { success: true };
    } catch (error: unknown) {
      let errorKey = 'userNotFound';

      if (typeof error === 'object' && error !== null && 'response' in error) {
        const axiosError = error as { response?: { data?: { error_description?: string } } };
        const msg = (axiosError.response?.data?.error_description || '').toLowerCase();

        if (msg.includes('user not found') || msg.includes('password mismatch')) {
          errorKey = 'userNotFound';
        } else if (msg.includes('user is blocked')) {
          errorKey = 'userBlocked';
        } else if (msg.includes('user is temporarily blocked')) {
          errorKey = 'userTemporarilyBlocked';
        }
      }

      return { success: false, error: errorKey };
    }
  };

  // 비밀번호 변경 처리 (Core에서 구현)
  const handleChangePassword = async (
    newPassword: string,
    currentPassword?: string,
  ): Promise<ChangePasswordResult> => {
    try {
      await axiosInstance.put('/me/password', {
        new_password: newPassword,
        current_password: currentPassword || '',
      });
      return { success: true };
    } catch (error: unknown) {
      let errorMessage = 'An unexpected error occurred.';

      if (typeof error === 'object' && error !== null && 'response' in error) {
        const axiosError = error as { response?: { data?: { error?: string } } };
        errorMessage = axiosError.response?.data?.error || errorMessage;
      }

      return { success: false, error: errorMessage };
    }
  };

  // Context 생성
  const context: AuthActionContext = {
    login: handleLogin,
    changePassword: handleChangePassword,
    dataProvider: {
      getOne: async ({ resource, id }) => {
        const url = id ? `/${resource}/${id}` : `/${resource}`;
        const res = await axiosInstance.get(url);
        return res.data;
      },
      create: async ({ resource, values }) => {
        const res = await axiosInstance.post(`/${resource}`, values);
        return res.data;
      },
      update: async ({ resource, id, values }) => {
        const url = id ? `/${resource}/${id}` : `/${resource}`;
        const res = await axiosInstance.put(url, values);
        return res.data;
      },
      custom: async ({ url, method, payload }) => {
        const res = await axiosInstance.request({
          url,
          method,
          data: payload,
          headers:
            payload instanceof URLSearchParams
              ? { 'Content-Type': 'application/x-www-form-urlencoded' }
              : undefined,
        });
        return res.data;
      },
    },
    utils: {
      t,
      toast: (message, type = 'info') => {
        console.log(`[Toast ${type}]: ${message}`);
      },
      navigate,
    },
  };

  const handleComplete = async () => {
    if (currentAction === 'login') {
      // 로그인 완료 후 /me API 호출하여 사용자 정보 가져오기
      try {
        localStorage.removeItem('pwExp-dialog');

        const response = await userProvider.getOne({
          resource: USER_RESOURCES.ME,
          id: '',
        });

        // Zod로 파싱하여 타입 안전성 확보
        const parseResult = UserSchema.safeParse(response.data);
        if (!parseResult.success) {
          console.error('Failed to parse user data:', parseResult.error);
          return;
        }
        const fetchedUser = parseResult.data;

        // 이전 사용자의 대시보드 상태 초기화
        useDashboardStore.getState().reset();

        // Auth store에 사용자 정보 저장
        setUser(fetchedUser, true);

        // prepare 필드 체크하여 필요한 액션으로 전환
        // setUser로 이미 스토어 업데이트됨 → currentAction이 prepare[0]으로 재렌더링
        if (fetchedUser.prepare && fetchedUser.prepare.length > 0) {
          return;
        }

        navigate('/home-upm');
      } catch (error) {
        console.error('Failed to fetch user info:', error);
      }
    } else {
      // required action 완료 → 홈으로
      navigate('/home-upm');
    }
  };

  const handleCancel = () => {
    // 로그아웃 처리
    window.location.href = '/ui/login';
  };

  // Extension이 전체 페이지를 렌더링
  return authActionPluginRegistry.render(currentAction, {
    user,
    context,
    onComplete: handleComplete,
    onCancel: currentAction !== 'login' ? handleCancel : undefined,
  });
}
