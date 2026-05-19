
import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { PageBreadcrumb } from '@components/breadcrumb';
import { VariableBarTemplate } from '@features/dashboard/components/VariableBar/VariableBarTemplate';
import { IconButton, LoadingIndicator } from '@pharos/shared/components/ui-extension';
import { IconFullscreen } from '@components/icons';
import { useFullscreenStore } from '@features/dashboard/hooks/fullscreenStore';
import GridLayout from '@features/dashboard/components/DashboardGrid/gridLayout';
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogCancel,
  AlertDialogAction,
  Checkbox,
  Label,
  Button,
} from '@pharos/shared/components/ui';
import { Save, Settings, Plus } from '@pharos/shared/components';
import { useToast } from '@hooks/use-toast';
import '@pharos/core/panel-registry/corePlugins'; // Register all core panels
import {
  useDashboardData,
  useDashboardActions,
  useDashboardAutoLoad,
  useDashboardStore,
} from '@features/dashboard/hooks/use-dashboard-store';
import { useDashboardNavigationGuard } from '@features/dashboard/hooks/use-dashboard-navigation-guard';
import { DashboardUrlManager } from '@features/dashboard/hooks/url';
import { unmountDashboardQueries } from '@features/dashboard/hooks/useDashboardQuery';
import { ActionSchema } from '@pharos/shared/types/dashboard';

/**
 * Dashboard 보기 페이지
 * URL: /dashboards/:dashboardId
 * VariableBarTemplate을 사용하여 대시보드를 표시
 */
interface DashboardPageHeaderProps {
  permission: string | undefined;
  setShowAlert: (show: boolean) => void;
  router: { push: (href: string) => void };
  dashboardId?: string;
}

function DashboardPageHeader({
  permission,
  setShowAlert,
  router,
  dashboardId,
}: DashboardPageHeaderProps) {
  const { toggleFullscreen } = useFullscreenStore();
  const isFullscreen = useFullscreenStore((s) => s.isFullscreen);

  if (isFullscreen) return null;

  return (
    <div className="flex items-center justify-between border-b px-3 h-11 shrink-0 bg-sidebar text-sidebar-foreground">
      <div className="flex items-center ml-2">
        <PageBreadcrumb />
      </div>
      <div className="flex items-center gap-1">
        {permission !== ActionSchema.enum.viewer && (
          <>
            <Button
              variant="default"
            className="h-8 shadow-none px-3 py-1"
            onClick={() => {
              const store = useDashboardStore.getState();
              const newPanelId = store.createPanel();
              if (!dashboardId) return;
              router.push(`/dashboards/${dashboardId}/edit/${newPanelId}${window.location.search}`);
            }}
          >
            <span className="[&>*]:w-4 [&>*]:h-4">
              <Plus className="stroke-[2]" />
            </span>
            Add panel
          </Button>
          <div className="h-8 w-8 flex items-center">
            <IconButton variant="ghost" icon={<Save />} onClick={() => setShowAlert(true)}>
              Save dashboard
            </IconButton>
          </div>
          <div className="h-8 w-8 flex items-center">
            <IconButton 
              variant="ghost" 
              icon={<Settings />}
              onClick={() => {
                if (!dashboardId) return;
                router.push(`/dashboards/${dashboardId}/settings${window.location.search}`);
              }}
            >
              Settings
            </IconButton>
          </div>
        </>
      )}
      <IconButton
        icon={<IconFullscreen />}
        variant="ghost"
        size="icon"
        className="cursor-pointer **:dark:stroke-white"
        onClick={toggleFullscreen}
      >
        Full screen
      </IconButton>
      </div>
    </div>
  );
}

export default function DashboardShowPage() {
  const { toast } = useToast();
  const router = useNavigate();
  const { dashboardId } = useParams<{ dashboardId: string }>();

  // ✅ 자동 대시보드 로딩 (URL 파라미터 포함, setFilter에서 URL 동기화)
  useDashboardAutoLoad(dashboardId || '');

  // ✅ 다른 페이지로 이동 시 저장 확인 (로그아웃은 header에서 처리)
  useDashboardNavigationGuard(dashboardId);

  // ✅ 대시보드 페이지 떠날 때 pending URL 업데이트 취소 + 모든 쿼리 abort
  React.useEffect(() => {
    return () => {
      unmountDashboardQueries();
      DashboardUrlManager.clearDashboardParams();
    };
  }, []);

  // ✅ 관련 페이지 청크를 미리 다운로드 (패널 쿼리로 네트워크가 포화될 때 클릭 → URL만 바뀌고
  //    lazy 청크 다운로드가 지연되는 문제 방지)
  React.useEffect(() => {
    void Promise.all([
      import('@/routes/dashboards/[dashboardId]/settings/page'),
      import('@/routes/dashboards/[dashboardId]/edit/[panelId]/page'),
      import('@/routes/dashboards/[dashboardId]/variables/new/page'),
    ]);
  }, []);

  const { loading, dashboard: dashboardData, permission } = useDashboardData();
  const { saveDashboard } = useDashboardActions();

  const [showAlert, setShowAlert] = useState(false);
  const [saveTimeRange, setSaveTimeRange] = useState(false);
  const [saveStep, setSaveStep] = useState(false);
  const [saveRefreshInterval, setSaveRefreshInterval] = useState(false);

  const filters = React.useMemo(() => {
    return dashboardData?.filters || [];
  }, [dashboardData?.filters]);

  // Early return after all hooks
  if (!dashboardId) {
    return (
      <div className="p-6">
        <h1 className="text-2xl font-bold mb-6 text-red-600">오류</h1>
        <div className="bg-red-50 p-4 rounded-lg border border-red-200">
          <p>Dashboard ID가 필요합니다.</p>
          <p className="text-sm text-gray-600 mt-2">URL 예시: /dashboards/123</p>
        </div>
      </div>
    );
  }

  const handleSave = async () => {
    try {
      await saveDashboard({ saveTimeRange, saveStep, saveRefreshInterval });
      toast({ description: 'Successfully saved.' });
      setShowAlert(false);
      // 저장 후 체크박스 초기화
      setSaveTimeRange(false);
      setSaveStep(false);
      setSaveRefreshInterval(false);
    } catch (error) {
      console.error('Error:', error);
      toast({ description: 'Failed to save dashboard', variant: 'destructive' });
    }
  };

  return (
    <div key={dashboardId} className="flex flex-col h-full w-full">
      <DashboardPageHeader
        permission={permission}
        setShowAlert={setShowAlert}
        router={{ push: (href: string) => router(href) }}
        dashboardId={dashboardId}
      />
      <div className="flex-1 min-h-0">
        {loading && <LoadingIndicator className="h-full" />}
        {!loading && dashboardData && (
          <VariableBarTemplate
            id={dashboardId}
            filters={filters}
            body={
              !dashboardData || dashboardData?.panels?.length === 0 ? (
                <div className="w-full h-full flex items-center justify-center flex-col">
                  <p className="font-bold text-secondary-foreground text-lg leading-16">
                    Add new panel.
                  </p>
                  <p>Start visualizing your data by adding a panel.</p>
                </div>
              ) : (
                <GridLayout />
              )
            }
          />
        )}
      </div>
      <AlertDialog open={showAlert} onOpenChange={setShowAlert}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Do you want to save changes to the dashboard?</AlertDialogTitle>
            <AlertDialogDescription>Unsaved changes will be lost if you leave without saving.</AlertDialogDescription>
          </AlertDialogHeader>
          <div className="py-4 space-y-3">
            <div className="flex items-center space-x-2">
              <Checkbox
                id="save-time-range"
                checked={saveTimeRange}
                onCheckedChange={(checked) => setSaveTimeRange(checked === true)}
              />
              <Label
                htmlFor="save-time-range"
                className="text-sm font-normal cursor-pointer"
              >
                Save current time range as dashboard default
              </Label>
            </div>
            <div className="flex items-center space-x-2">
              <Checkbox
                id="save-step"
                checked={saveStep}
                onCheckedChange={(checked) => setSaveStep(checked === true)}
              />
              <Label
                htmlFor="save-step"
                className="text-sm font-normal cursor-pointer"
              >
                Save current step as dashboard default
              </Label>
            </div>
            <div className="flex items-center space-x-2">
              <Checkbox
                id="save-refresh-interval"
                checked={saveRefreshInterval}
                onCheckedChange={(checked) => setSaveRefreshInterval(checked === true)}
              />
              <Label
                htmlFor="save-refresh-interval"
                className="text-sm font-normal cursor-pointer"
              >
                Save current refresh interval as dashboard default
              </Label>
            </div>
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleSave}>Save</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
