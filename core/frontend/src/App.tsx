// CRITICAL: Import extension loader FIRST
import '@pharos/meta/extension-loader';

import React from 'react';
import {BrowserRouter, Routes, Route, Navigate} from 'react-router-dom';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import {RefineContext} from '@/routes/_refine_context';
import {getExtensionPages} from '@features/extension/registry';
import {Toaster} from '@pharos/shared/components/ui/sonner';
import {LoadingIndicator} from '@pharos/shared/components/ui-extension';

import './locales/i18next';

// Create a QueryClient instance
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

// Import actual page components
const HomePage = React.lazy(() => import('@/routes/home/page'));
const HomeV2Page = React.lazy(() => import('@/routes/home-v2/page'));
const LoginPage = React.lazy(() => import('@/routes/login/page'));
const DashboardsPage = React.lazy(() => import('@/routes/dashboards/page'));
const AlertPage = React.lazy(() => import('@/routes/alert/page'));
const AlertRuleEditPage = React.lazy(() => import('@/routes/alert/edit/page'));
const NotificationPage = React.lazy(() => import('@/routes/notification/page'));
const UsersPage = React.lazy(() => import('@/routes/users/page'));
const UserEditPage = React.lazy(() => import('@/routes/users/edit/page'));
const RolesSettingsPage = React.lazy(() => import('@/routes/settings/roles/page'));
const UserFieldsSettingsPage = React.lazy(() => import('@/routes/settings/user-fields/page'));
// const SetThemePage = React.lazy(() => import('@/routes/setTheme/page'));
const NotFoundPage = () => <div>404 - Not Found</div>;

function App() {
  // Get extension pages dynamically
  const extensionPages = getExtensionPages();

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter basename="/ui">
        <RefineContext>
          <React.Suspense fallback={<LoadingIndicator className="h-screen" />}>
            <Routes>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/" element={<Navigate to="/home" replace />} />
              <Route path="/home" element={<HomePage />} />
              <Route path="/home-v2" element={<HomeV2Page />} />
              <Route path="/dashboards/*" element={<DashboardsPage />} />
              <Route path="/alert/edit" element={<AlertRuleEditPage />} />
              <Route path="/alert/*" element={<AlertPage />} />
              <Route path="/notification/*" element={<NotificationPage />} />
              <Route path="/users/edit" element={<UserEditPage />} />
              <Route path="/users/*" element={<UsersPage />} />
              <Route path="/settings/roles/*" element={<RolesSettingsPage />} />
              <Route path="/settings/user-fields" element={<UserFieldsSettingsPage />} />
              {/*<Route path="/setTheme" element={<SetThemePage />} />*/}

              {/* Extension Routes - dynamically added */}
              {extensionPages.map((page) => (
                <Route key={page.path} path={page.path} element={<page.component />} />
              ))}

              <Route path="*" element={<NotFoundPage />} />
            </Routes>
          </React.Suspense>
        </RefineContext>
        <Toaster />
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;
