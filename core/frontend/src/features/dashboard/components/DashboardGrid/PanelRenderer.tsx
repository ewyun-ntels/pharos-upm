'use client';

import React, { useMemo, useCallback } from 'react';
import { LayoutItem } from 'react-grid-layout';
import {useDashboardStore} from '@features/dashboard/hooks/use-dashboard-store';
import { Action, Panel } from '@pharos/shared/types/dashboard';

import {panelPluginRegistry} from '@pharos/core/panel-registry';
import type {PanelAction} from '@pharos/core/panel-registry';
import '@pharos/core/panel-registry/corePlugins'; // Register all core panels
import type { Annotation } from '@pharos/shared/types/dashboard';

const EMPTY_ANNOTATIONS: Annotation[] = [];
import { getPluginContext } from '@utils/pluginHooks';
import { useVariableState } from '@features/dashboard/hooks/use-variable-state';
import { usePanelArgs } from '@features/dashboard/hooks/use-panel-args';
import { useTransformedPanelData } from '@features/dashboard/hooks/use-transformed-panel-data';
import { LoadingIndicator } from '@pharos/shared/components/ui-extension';

const ChartLoadingSpinner = () => <LoadingIndicator />;

interface PanelRendererProps {
  item: LayoutItem;
  panel?: Panel;
  dashboardId?: string;
  displayName?: string;
  permission: Action;  // 스토어에서 항상 제공되므로 required
}

/**
 * Optimized PanelRenderer Component
 *
 * Key optimizations:
 * 1. Component caching with useMemo (prevents unnecessary re-creation)
 * 2. Direct rendering with React.lazy (eliminates registry.render overhead)
 * 3. Props memoization (stable references prevent child re-renders)
 * 4. Efficient dependency tracking (only re-compute when necessary)
 */

const PanelRendererComponent: React.FC<PanelRendererProps> = ({
  item,
  panel: panelProp,
  dashboardId,
  displayName,
  permission,
}) => {
  // ✅ 공통 훅: Variable State 구독
  const { filterMetas, filterValues, filterIds, datetime, step, refreshCount } = useVariableState();
  const refetchInterval = useDashboardStore((state) => state.filterState?.refreshInterval);
  const setDatetime = useDashboardStore((state) => state.setDatetime);
  const storeAnnotations = useDashboardStore((state) => state.storeAnnotations ?? EMPTY_ANNOTATIONS);
  const addAnnotation = useDashboardStore((state) => state.addAnnotation);
  const removeAnnotation = useDashboardStore((state) => state.removeAnnotation);
  const globalAnnotations = React.useMemo(
    () => storeAnnotations.filter((a) => a.scope === 'global'),
    [storeAnnotations],
  );

  // GridItem의 custom memo(레이아웃만 비교)로 인해 panel prop이 stale해지는 문제를 해결하기 위해
  // store에서 직접 최신 panel을 구독. updatePanel이 새 객체를 만들어도 항상 최신 데이터를 받음.
  const panelFromStore = useDashboardStore((state) => state.panelMap[item.i]);
  const panel = panelFromStore ?? panelProp;

  const panelAnnotations = React.useMemo(
    () => storeAnnotations.filter((a) => a.scope === 'panel' && a.panel_id === (panel?.id || item.i)),
    [storeAnnotations, panel?.id, item.i],
  );

  const updatePanel = useDashboardStore((state) => state.updatePanel);
  const updatePanelOptions = useDashboardStore((state) => state.updatePanelOptions);

  // ✅ 공통 훅: toPanelData() 변환 (SWR 캐시 포함)
  // - watchOptions: false — 대시보드에서 options는 toPanelData 재실행 불필요 (리렌더로 반영)
  // - enableCache: true — remount 시 스피너 제거
  const { transformedData } = useTransformedPanelData({
    panel,
    permission,
    watchOptions: false,
    enableCache: true,
  });

  // Plugin Context (hooks 포함) - useMemo로 안정화: 매 렌더마다 새 참조 생성 방지
  const pluginContext = React.useMemo(() => getPluginContext(), []);

  // ✅ 공통 훅: Map 기반 args 생성
  const args = usePanelArgs(filterValues, filterIds, datetime, step, refreshCount);

  const handleAnnotationCreate = useCallback(async (annotation: Annotation) => {
    await addAnnotation(annotation);
  }, [addAnnotation]);

  const handleAnnotationDelete = useCallback(async (annotationId: string) => {
    await removeAnnotation(annotationId);
  }, [removeAnnotation]);

  // ✅ OPTIMIZATION 1: Stable Panel Component from registry cache
  // Registry에서 renderType별 캐시된 React.lazy() 인스턴스를 반환하므로
  // useMemo 내부에서 React.lazy() 호출하는 anti-pattern 제거
  // - 같은 renderType → 항상 동일한 lazy 인스턴스 (panel remount 시 재로딩 없음)
  // - 다른 패널 컴포넌트들(e.g. 10개 timeSeries)이 모두 같은 lazy 인스턴스를 공유
  const renderType = panel?.renderType;
  const PanelComponent = renderType ? panelPluginRegistry.getLazyComponent(renderType) : null;

  // ✅ OPTIMIZATION 2: Stable callback reference
  const handlePanelAction = useCallback((action: PanelAction) => {
    const plugin = renderType ? panelPluginRegistry.get(renderType) : undefined;
    if (!plugin?.actionHandlers) return;
    const handler = plugin.actionHandlers[action.type];
    if (!handler) return;
    const panelId = panel?.id || '';
    handler(action.payload, {
      getOptions: () => panel?.options as unknown,
      updateOptions: (updates) => updatePanelOptions(panelId, updates),
    });
  }, [renderType, panel?.options, panel?.id, updatePanelOptions]);

  const handlePanelUpdate = useCallback((updates: Partial<Panel>) => {
    const panelId = panel?.id || '';

    const panelFields = ['title', 'description', 'bgTransparent', 'renderType'];
    const hasPanelFields = panelFields.some(field => updates.hasOwnProperty(field));

    if (hasPanelFields) {
      updatePanel(panelId, updates);
    } else {
      updatePanelOptions(panelId, updates);
    }
  }, [panel?.id, updatePanel, updatePanelOptions]);

  // ✅ OPTIMIZATION 3: Memoized panel props (stable reference prevents unnecessary re-renders)
  // ⚠️ IMPORTANT: Must be called before any early returns to follow Rules of Hooks
  const panelProps = useMemo(() => {
    if (!transformedData) return null;

    return {
      ...transformedData,
      // Dashboard context
      dashboardId,
      kind: 'panels' as const,
      id: panel?.id || item.i,
      title: panel?.title === 'Untitled' ? '' : (panel?.title || ''),
      description: panel?.description || '',
      displayName: displayName || '',
      permission,
      // Variable 치환용 (title/description template, 드릴다운 링크 등)
      filterMetas,
      // Runtime arguments (query execution)
      args,
      refetchInterval,
      // Plugin context (hooks)
      pluginContext,
      // Callbacks
      timeRangeCallback: setDatetime,
      onPanelUpdate: handlePanelUpdate,
      onPanelAction: handlePanelAction,
      // Annotation callbacks
      onAnnotationCreate: handleAnnotationCreate,
      onAnnotationDelete: handleAnnotationDelete,
      // Panel-scoped annotations (filtered by panel.id)
      annotations: panelAnnotations,
      // Global annotations (scope === 'global')
      globalAnnotations,
    };
  }, [
    transformedData,
    dashboardId,
    panel?.id,
    panel?.title,
    panel?.description,
    item.i,
    displayName,
    permission,
    filterMetas,
    args,
    refetchInterval,
    pluginContext,
    setDatetime,
    handlePanelUpdate,
    handlePanelAction,
    handleAnnotationCreate,
    handleAnnotationDelete,
    panelAnnotations,
    globalAnnotations,
  ]);

  // Early returns for loading/error states (after all hooks)
  if (!panel) {
    return (
      <div className="flex items-center justify-center w-full h-full text-muted-foreground text-sm">
        No data available
      </div>
    );
  }

  // renderType 누락 시 명시적 에러 표시
  if (!panel.renderType) {
    console.error('[PanelRenderer] Missing renderType for panel:', {
      panelId: panel.id || item.i,
      panel,
    });
    return (
      <div className="flex items-center justify-center w-full h-full text-red-500 text-sm border border-red-300 rounded bg-red-50 p-4">
        <div className="text-center">
          <p className="font-semibold">Panel Render Error</p>
          <p className="mt-1">Missing renderType for panel ID: {panel.id || item.i}</p>
          <p className="mt-2 text-xs text-red-600">
            Please check the dashboard configuration or contact an administrator.
          </p>
        </div>
      </div>
    );
  }

  // Wait for transformation to complete
  if (!transformedData || !PanelComponent || !panelProps) {
    return <ChartLoadingSpinner />;
  }

  // ✅ OPTIMIZATION 4: Direct rendering with React.lazy + Suspense
  // Benefits:
  // - Component reference stays stable (useMemo cache)
  // - queryKey in useDashboardQuery detects args changes
  // - No unnecessary remount (better UX)
  // - Code splitting with lazy loading
  return (
    <React.Suspense fallback={<ChartLoadingSpinner />}>
      <PanelComponent key={`panel-${panel.id}`} {...panelProps} />
    </React.Suspense>
  );
};

/**
 * Memoized wrapper to prevent re-renders when parent updates
 * but props haven't changed
 */
export const PanelRenderer = React.memo(PanelRendererComponent);
