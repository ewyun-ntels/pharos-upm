
// CRITICAL: Import extension loader FIRST to ensure extensions are registered
// before any other code tries to access them
import '@pharos/meta/extension-loader';

import { DevtoolsProvider } from '@providers/devtools';
import React, { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { Layout } from '@components/layout/layout';

import { dashboardProvider, DASHBOARD_PROVIDER_NAME } from '@providers/dashboard-provider';
import { uiConfigProvider, UI_CONFIG_PROVIDER_NAME } from '@providers/ui-config-provider';
import { defaultProvider } from '@providers/default-provider';
import { ThemeProvider } from '@providers/theme-provider';
import { alertProvider, ALERT_PROVIDER_NAME } from '@providers/alert-provider';
import { notificationProvider, NOTIFICATION_PROVIDER_NAME } from '@providers/notification-provider';
import '@pharos/shared/styles/globals.css';

import { Toaster } from '@pharos/shared/components/ui';
import '../locales/i18next';
import { authProvider } from '@providers/auth-provider';
import { pluginProvider, PLUGIN_PROVIDER_NAME } from '@providers/plugin-provider';
import { userProvider, USER_PROVIDER_NAME } from '@providers/user-provider';
import { roleProvider, ROLE_PROVIDER_NAME } from '@providers/role-provider';
import { getExtensionDataProviders } from '@features/extension/registry';
import { AuthInitializer } from '@providers/auth-provider/initializer';
import { menuRegistry } from '@features/menu/registry';
import { LoadingIndicator } from '@pharos/shared/components/ui-extension';
import { MENU_ORDER } from '@pharos/meta/menu-order';
import { dataProviderRegistry, setAuthProvider } from '@/lib/data-provider';

// Import all feature menus (registers items automatically)
import '@/features/dashboard/menu';
import '@/features/alert/menu';
import '@/features/notification/menu';
import '@/features/user/menu';
import '@/features/settings/menu';

// 모든 feature/extension 등록 완료 후 순서 적용 (한 번만)
menuRegistry.applyOrder(MENU_ORDER);

// DataProviderRegistry에 모든 provider 등록 (한 번만)
setAuthProvider(authProvider);
dataProviderRegistry.register('default', defaultProvider);
dataProviderRegistry.register(DASHBOARD_PROVIDER_NAME, dashboardProvider);
dataProviderRegistry.register(UI_CONFIG_PROVIDER_NAME, uiConfigProvider);
dataProviderRegistry.register(ALERT_PROVIDER_NAME, alertProvider);
dataProviderRegistry.register(NOTIFICATION_PROVIDER_NAME, notificationProvider);
dataProviderRegistry.register(USER_PROVIDER_NAME, userProvider);
dataProviderRegistry.register(ROLE_PROVIDER_NAME, roleProvider);
dataProviderRegistry.register(PLUGIN_PROVIDER_NAME, pluginProvider);
// Extension data providers 자동 등록
Object.entries(getExtensionDataProviders()).forEach(([name, provider]) => {
  if (!dataProviderRegistry.has(name)) {
    dataProviderRegistry.register(name, provider as any);
  }
});

type RefineContextProps = {};

export const RefineContext = (props: React.PropsWithChildren<RefineContextProps>) => {
  return <App {...props} />;
};

type AppProps = {};

const App = (props: React.PropsWithChildren<AppProps>) => {
  const location = useLocation();

  const [isAuthChecking, setIsAuthChecking] = React.useState(true);

  // useEffect(() => {
  //   if (data.config && data.config['style-primary']) {
  //     document.documentElement.style.setProperty('--primary', data.config['style-primary']);
  //   } else {
  //     document.documentElement.style.setProperty('--primary', 'oklch(0.66 0.1197 156.75)');
  //   }
  // }, [data.config]);

  useEffect(() => {
    if (location.pathname === '/login') {
      setIsAuthChecking(false);
      return;
    }

    const checkAuth = async () => {
      try {
        const result = await authProvider.check();
        if (!result.authenticated && result.redirectTo) {
          // 인증되지 않은 경우 리다이렉트 (로딩 화면 유지)
          window.location.href = '/ui' + result.redirectTo;
          return;
        }
      } catch (error) {
        console.error('Authentication check failed:', error);
        // 인증 확인 중 오류 발생 시 로그인 페이지로 리다이렉트 (로딩 화면 유지)
        window.location.href = '/ui/login';
        return;
      }
      setIsAuthChecking(false);
    };

    checkAuth();
  }, []);

  if (isAuthChecking) {
    return <LoadingIndicator className="h-screen" />;
  }

  return (
    <DevtoolsProvider>
      <AuthInitializer />
      <Toaster />
      <ThemeProvider defaultTheme="system" storageKey="ui-theme">
        {location.pathname === '/login' ? (
          props.children
        ) : (
          <Layout>
            {props.children}
          </Layout>
        )}
      </ThemeProvider>
    </DevtoolsProvider>
  );
};
