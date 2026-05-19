import React, { useState, useEffect } from 'react';
import { useParams, useLocation, useNavigate } from 'react-router-dom';
import { LoadingIndicator } from '@pharos/shared/components/ui-extension';
import { PageBreadcrumb } from '@components/breadcrumb';
import { TableTabsTemplate, Tabs, Tab } from '@pharos/shared/components/template/table-tabs';
import {
  useDashboardData,
  useDashboardAutoLoad,
} from '@features/dashboard/hooks/use-dashboard-store';
import { useDashboardNavigationGuard } from '@features/dashboard/hooks/use-dashboard-navigation-guard';
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from '@pharos/shared/components/ui';
import { AlertTriangle } from 'lucide-react';
import { ActionSchema } from '@pharos/shared/types/dashboard';
import { GeneralSettingTab, PermissionTab, JsonEditorTab, VariablesTab } from '@features/dashboard/settings';

export default function DashboardSettingsPage() {
  const { dashboardId } = useParams<{ dashboardId: string }>();
  const location = useLocation();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<string>('General');

  useDashboardAutoLoad(dashboardId || '');
  useDashboardNavigationGuard(dashboardId);

  const { loading, dashboard: dashboardData, permission } = useDashboardData();

  useEffect(() => {
    const rawHash = location.hash.replace('#', '').toLowerCase();
    if (!rawHash) {
      return;
    }
    
    const hashToTab: Record<string, string> = {
      general: 'General',
      permission: 'Permission',
      'json-editor': 'JSON Editor',
      variables: 'Variables',
    };
    
    const requestedTab = hashToTab[rawHash];
    
    const availableTabs: string[] = [
      'General',
      ...(permission === ActionSchema.enum.owner ? ['Permission'] : []),
      'JSON Editor',
      'Variables',
    ];
    
    if (requestedTab && availableTabs.includes(requestedTab)) {
      setActiveTab(requestedTab);
    } else {
      setActiveTab(availableTabs[0]);
    }
  }, [location.hash, permission]);

  const handleTabChange = (tabName: string) => {
    setActiveTab(tabName);
    
    const tabToHash: Record<string, string> = {
      General: 'general',
      Permission: 'permission',
      'JSON Editor': 'json-editor',
      Variables: 'variables',
    };
    
    const hash = tabToHash[tabName] || tabName.toLowerCase();
    navigate(`#${hash}`, { replace: true });
  };

  if (loading) return <main className="w-full h-full"><LoadingIndicator /></main>;
  
  if (!dashboardId) {
    return (
      <main className="w-full h-full p-4">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>Dashboard ID is required.</AlertDescription>
        </Alert>
      </main>
    );
  }

  if (!dashboardData) {
    return (
      <main className="w-full h-full p-4">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>Dashboard not found.</AlertDescription>
        </Alert>
      </main>
    );
  }

  return (
    <main className="flex flex-col w-full h-full">
      <TableTabsTemplate name="Dashboard Settings" breadcrumb={<PageBreadcrumb />} activeTab={activeTab} onTabChange={handleTabChange}>
        <Tabs>
          <Tab name="General">
            <GeneralSettingTab />
          </Tab>
          {permission === ActionSchema.enum.owner && (
            <Tab name="Permission">
              <PermissionTab dashboardId={dashboardId} />
            </Tab>
          )}
          <Tab name="JSON Editor">
            <JsonEditorTab />
          </Tab>
          <Tab name="Variables">
            <VariablesTab dashboardId={dashboardId} />
          </Tab>
        </Tabs>
      </TableTabsTemplate>
    </main>
  );
}
