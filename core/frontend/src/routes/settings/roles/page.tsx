import React from 'react';
import {Routes, Route, useLocation} from 'react-router-dom';
import {RoleConfigTab, CompositeRolesTab, RoleHistoryTab} from '@features/settings';
import {TableTabsTemplate, Tabs, Tab} from '@pharos/shared/components/template/table-tabs';
import {PageBreadcrumb} from '@components/breadcrumb';
import {LoadingIndicator} from '@pharos/shared/components/ui-extension';

const CompositeRoleFormPage = React.lazy(() => import('./composite/page'));

function RolesSettingsIndex() {
  const location = useLocation();
  const defaultTab =
    (location.state as {activeTab?: string} | null)?.activeTab ?? 'Base Roles';

  return (
    <main className="flex flex-col w-full h-full">
      <TableTabsTemplate
        name="Role Configuration"
        breadcrumb={<PageBreadcrumb />}
        defaultTab={defaultTab}
      >
        <Tabs>
          <Tab name="Base Roles">
            <RoleConfigTab />
          </Tab>
          <Tab name="Composite Roles">
            <CompositeRolesTab />
          </Tab>
          <Tab name="History">
            <RoleHistoryTab />
          </Tab>
        </Tabs>
      </TableTabsTemplate>
    </main>
  );
}

export default function RolesSettingsRouter() {
  return (
    <React.Suspense fallback={<LoadingIndicator className="h-screen" />}>
      <Routes>
        <Route index element={<RolesSettingsIndex />} />
        <Route path="composite" element={<CompositeRoleFormPage />} />
      </Routes>
    </React.Suspense>
  );
}
