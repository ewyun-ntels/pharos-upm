import React, {useMemo} from 'react';
import {FilterConfig} from '@features/dashboard/components/VariableBar/types';
import {VariableBarTemplate} from '@features/dashboard/components/VariableBar/VariableBarTemplate';
import {useDashboardStore, useDashboardData} from '@features/dashboard/hooks/use-dashboard-store';
import {panelPluginRegistry} from '@pharos/core/panel-registry';
import '@pharos/core/panel-registry/corePlugins';
import {usePanelArgs} from '@features/dashboard/hooks/use-panel-args';
import {useVariableState} from '@features/dashboard/hooks/use-variable-state';
import {useTransformedPanelData} from '@features/dashboard/hooks/use-transformed-panel-data';
import {DASHBOARD_PROVIDER_NAME, DASHBOARD_RESOURCES} from '@providers/dashboard-provider';
import {getPluginContext} from '@utils/pluginHooks';
import type {Panel} from '@pharos/shared/types/dashboard';

import '@features/dashboard/variables';

const LeftTopPanel = ({
  filters,
  dashboardId,
  panelId,
  overridePanel,
}: {
  filters: FilterConfig[];
  dashboardId: string;
  panelId: string;
  overridePanel?: Panel;
}) => {
  const storePanel = useDashboardStore((state) => state.panelMap[panelId]);
  const panel = overridePanel ?? storePanel;
  const { permission } = useDashboardData();

  const { filterMetas, filterValues, filterIds, datetime, step, refreshCount } = useVariableState();

  // args 생성 - PanelRenderer와 동일한 usePanelArgs 사용 (refreshCount 포함)
  const args = usePanelArgs(filterValues, filterIds, datetime, step, refreshCount);

  // Plugin Context 가져오기 - useMemo로 안정화 (매 렌더마다 새 참조 방지)
  const pluginContext = React.useMemo(() => getPluginContext(), []);

  // ✅ options 변경 감지를 위한 deep comparison
  const optionsStr = useMemo(() => JSON.stringify(panel?.options || {}), [panel?.options]);

  // ✅ 공통 훅: toPanelData() 변환
  // - watchOptions: true — 에디터에서 Fields 속성 변경 시 즉시 미리보기 반영
  // - enableCache: false — 에디터에서는 SWR 캐시 불필요
  const { transformedData: resolvedData, isResolved } = useTransformedPanelData({
    panel,
    permission,
    watchOptions: true,
    enableCache: false,
  });

  // 차트 props 생성 (타입 안전하게)
  const chartProps = useMemo(() => {
    if (!panel) return null;

    return {
      id: panel.id,
      title: panel.title || '',
      description: panel.description || '',
      bgTransparent: panel.bgTransparent || false,
      args,
      options: {
        ...(panel.options || {}),         // ① 모든 사용자 설정 보존 (leftItems, rightItems 등)
        ...(resolvedData?.options || {}), // ② toPanelData 결과로 오버라이드 (columnConfigs 등)
      },
      dataProvider: {
        chartQuery: panel.dataProvider?.chartQuery,
        dataProviderName: DASHBOARD_PROVIDER_NAME,
        resource: resolvedData?.dataProvider?.resource || DASHBOARD_RESOURCES.TIMESERIES,
      },
      dashboardId,
      permission,
      filterMetas,
      pluginContext,
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    panel?.id,
    panel?.title,
    panel?.description,
    panel?.bgTransparent,
    optionsStr,
    panel?.dataProvider,
    args,
    dashboardId,
    permission,
    filterMetas,
    pluginContext,
    panelId,
    resolvedData,
  ]);


  // ✅ PanelRenderer와 동일하게 registry 캐시 사용 (useMemo 내 React.lazy anti-pattern 제거)
  //
  // ⚠️ 절대 useMemo(() => React.lazy(...), [panelType])로 변경하지 말것!
  //
  // 문제 (useMemo + React.lazy 조합):
  // - 매번 새로운 React.lazy 인스턴스 생성 → 항상 Suspense 사이클 반복
  // - Suspense fallback → resolve → render 과정에서 useLayoutEffect 타이밍 불안정
  // - 특히 조건부 렌더링(Alert → 차트)과 결합 시 chartHeight 측정 실패 가능
  //
  // 해결 (registry 캐시):
  // - panelPluginRegistry.getLazyComponent()는 renderType별로 캐시된 lazy 인스턴스 반환
  // - 같은 renderType이면 항상 동일한 인스턴스 → 한 번 resolve된 후 동기 렌더링
  // - useLayoutEffect 타이밍 안정적
  const panelType = panel?.renderType;
  const PanelComponent = panelType ? panelPluginRegistry.getLazyComponent(panelType) : null;

  // 렌더링 조건 체크 (resolvedData 확정 전 렌더링 보류 — PanelRenderer와 동일 패턴)
  if (!panel || !chartProps || !PanelComponent || !isResolved) {
    return null;
  }

  const chartElement = (
    <React.Suspense fallback={<div className="flex items-center justify-center w-full h-full text-sm text-muted-foreground">Loading chart...</div>}>
      <PanelComponent {...chartProps} />
    </React.Suspense>
  );

  const panelBody = panelType === 'table' ? (
    <div className="w-full h-full relative p-4">
      <div className="w-full h-full absolute" style={{overflow: 'hidden'}}>
        {chartElement}
      </div>
    </div>
  ) : (
    // ✅ flex-1 대신 w-full h-full 사용: CSS height 체인을 명시적으로 유지
    // flex-1만으로는 자식의 h-full이 react-resizable-panels 초기화 전에 0으로 측정되어
    // chartHeight 데드락 발생 (TimeSeriesCard의 chartHeight <= 40 가드가 영구 null 반환)
    <div className="w-full h-full p-4 overflow-hidden">
      {chartElement}
    </div>
  );

  // ✅ 대시보드와 동일하게 VariableBarTemplate 사용
  return (
    <VariableBarTemplate
      id={dashboardId}
      filters={filters}
      body={panelBody}
    />
  );
};

export default LeftTopPanel;
