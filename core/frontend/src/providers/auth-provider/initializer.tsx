'use client';

import { useEffect } from 'react';
import { useAuthStore } from '@pharos/shared/features/auth';
import { userProvider, USER_RESOURCES } from '@/providers/user-provider';
import type { User } from '@pharos/shared/features/auth';
import { useFavoritesStore } from '@features/dashboard/hooks/use-favorites-store';
import { useDashboardNavStore } from '@features/dashboard/hooks/use-dashboard-nav-store';

/**
 * Auth Store 초기화 컴포넌트
 *
 * 앱 로딩 시 /me API를 호출하여 사용자 정보 복구 (새로고침 대응)
 *
 * 참고:
 * - 로그인 시 reset은 login/page.tsx에서 직접 처리
 * - 각 feature의 로그아웃 cleanup은 registerOnLogout으로 등록
 */
export const AuthInitializer = () => {
  const initialize = useAuthStore(state => state.initialize);
  const isInitialized = useAuthStore(state => state.isInitialized);
  const isAuthenticated = useAuthStore(state => state.isAuthenticated);

  useEffect(() => {
    // 로그인 페이지에서는 실행하지 않음
    if (window.location.pathname === '/ui/login') {
      console.log('[AuthInitializer] Skipping on login page');
      return;
    }

    // 이미 초기화되었으면 스킵
    if (isInitialized) {
      console.log('[AuthInitializer] Already initialized');
      return;
    }

    console.log('[AuthInitializer] Initializing auth store...');

    // Auth Store 초기화 (fetchUser 함수만 등록, 콜백은 각 feature에서 registerOnLogout으로 등록)
    initialize(
      async () => {
        try {
          console.log('[AuthInitializer] Fetching user info from /me...');
          const {data: user} = await userProvider.getOne({
            resource: USER_RESOURCES.ME,
            id: '',
          }) as {data: User};
          console.log('[AuthInitializer] User info loaded:', user?.name || 'unknown');
          return user;
        } catch (error) {
          console.error('[AuthInitializer] Failed to fetch user info:', error);
          return null;
        }
      }
    ).then(_ => {});
   }, [initialize, isInitialized]);

  // isAuthenticated가 true가 되는 시점에 favorites 로드
  // - 새로고침: initialize() 완료 후 isAuthenticated: true
  // - 로그인: login/page.tsx의 setUser() 호출 후 isAuthenticated: true
  useEffect(() => {
    if (isAuthenticated) {
      useFavoritesStore.getState().loadFavorites();
      useDashboardNavStore.getState().loadNav();
    }
  }, [isAuthenticated]);

  return null;
};
