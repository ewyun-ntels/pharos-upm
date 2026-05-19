
import React from 'react';
import { LoadingIndicator } from '@pharos/shared/components/ui-extension';
import { useParams } from 'react-router-dom';
import ChartEditorLayout from '@features/dashboard/panels/components/PanelEditor/chart-editor-layout';
import {
  useDashboardData,
  useDashboardAutoLoad
} from '@features/dashboard/hooks/use-dashboard-store';
import { useDashboardNavigationGuard } from '@features/dashboard/hooks/use-dashboard-navigation-guard';

/**
 * ⚠️ IMPORTANT: Core Plugins Registration
 * 
 * 모든 패널 플러그인(timeSeries, table, pie 등)을 등록합니다.
 * 이 import가 없으면 panelPluginRegistry.loadEditorConfig()가 실패하여
 * OptionsComponent를 찾을 수 없게 됩니다.
 * 
 * 📌 다른 페이지에서도 패널 플러그인을 사용한다면 반드시 추가해야 합니다:
 * - Panel 렌더링이 필요한 페이지
 * - Panel Editor/Options를 사용하는 페이지
 * - Panel 관련 기능을 사용하는 모든 페이지
 */
import '@pharos/core/panel-registry/corePlugins';

/**
 * Dashboard Panel 에디터 페이지
 * URL: /dashboards/edit/:dashboardId/:panelId
 */
export default function DashboardEditPage() {
  const { dashboardId, panelId } = useParams<{ dashboardId: string; panelId: string }>();

  // 자동 대시보드 로딩
  useDashboardAutoLoad(dashboardId || '');

  const { loading, dashboard: dashboardData } = useDashboardData();

  // ✅ 다른 페이지로 이동 시 저장 확인 (브라우저 새로고침/닫기 포함)
  useDashboardNavigationGuard(dashboardId);

  if (!dashboardId || !panelId) {
    return (
      <div className="p-6">
        <h1 className="text-2xl font-bold mb-6 text-red-600">오류</h1>
        <div className="bg-red-50 p-4 rounded-lg border border-red-200">
          <p className="mb-3">Dashboard ID와 Panel ID가 모두 필요합니다.</p>
          <div className="text-sm text-gray-600">
            <p className="mb-2">URL 예시: /dashboards/edit/123/456</p>
            <div className="border-t pt-2">
              <p><strong>현재 파라미터:</strong></p>
              <p>• dashboardId: {dashboardId || '없음'}</p>
              <p>• panelId: {panelId || '없음'}</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <React.Fragment key={dashboardId}>
      {loading && <LoadingIndicator className="h-full" />}
      {!loading && dashboardData && (
        <ChartEditorLayout
          key={dashboardId}
          id={dashboardId}
          panelId={panelId}
          filters={dashboardData?.filters ?? []}
        />
      )}
    </React.Fragment>
  );
}
