/**
 * Dashboard Store (Zustand)
 *
 * Slice 패턴으로 구성된 대시보드 상태 관리 스토어.
 *
 * Slice 구조:
 *   - dashboard-crud-slice: 대시보드 CRUD + 기본 상태
 *   - panel-slice: 패널 관리 (panelOrder, panelMap)
 *   - variable-runtime-slice: 변수 런타임 (filterMetas, datetime, step, refreshCount)
 *
 * URL 동기화:
 *   - url/dashboard-url-manager: URL 읽기/쓰기 단일 진입점
 *   - url/use-url-sync: Zustand subscribe 기반 자동 동기화
 *
 * 초기화:
 *   - init/resolve-initial-state: 3단계 우선순위 (URL > saved > default) 통합
 */
import { create } from 'zustand';
import { subscribeWithSelector } from 'zustand/middleware';
import { enableMapSet } from 'immer';
import React from 'react';
import { useLocation } from 'react-router-dom';
import { registerOnLogout } from '@pharos/shared/features/auth';
import type { Panel } from '@pharos/shared/types/dashboard';
import { useDatasourceList } from '@hooks/use-datasource-list';
import { setupUrlSync } from './url';
import { createDashboardCrudSlice } from './slices/dashboard-crud-slice';
import { createPanelSlice } from './slices/panel-slice';
import { createVariableRuntimeSlice } from './slices/variable-runtime-slice';

// Immer Map/Set 지원 활성화
enableMapSet();
// setAutoFreeze 제거: DependencyGraph를 store 외부 ref로 분리하여 더 이상 필요 없음

// ─── 타입 Re-export (기존 import 경로 유지) ─────────────────────────────────

export type { FilterMeta, FilterState, DashboardStore } from './slices/types';
export { createFilterState, getDependencyGraph } from './slices/types';

// ─── Store 생성 (Slice 조합) ─────────────────────────────────────────────────

import type { DashboardStore } from './slices/types';

const useDashboardStore = create<DashboardStore>()(
  subscribeWithSelector(
    (set, get) => ({
      ...createDashboardCrudSlice(set, get),
      ...createPanelSlice(set, get),
      ...createVariableRuntimeSlice(set, get),
    }),
  ),
);

export { useDashboardStore };

// ─── 선택적 구독을 위한 유틸리티 훅들 ───────────────────────────────────────

export const useDashboardData = () => {
  const dashboard = useDashboardStore((state) => state.dashboard);
  const permission = useDashboardStore((state) => state.permission);
  const loading = useDashboardStore((state) => state.loading);
  const error = useDashboardStore((state) => state.error);
  const id = useDashboardStore((state) => state.id);
  const panelOrder = useDashboardStore((state) => state.panelOrder);
  const panelMap = useDashboardStore((state) => state.panelMap);
  const filters = useDashboardStore((state) => state.filters);

  const panels = React.useMemo(
    () =>
      panelOrder
        .map((panelId) => panelMap[panelId])
        .filter((panel): panel is Panel => panel !== undefined),
    [panelOrder, panelMap],
  );

  return { dashboard, permission, loading, error, id, panels, filters };
};

export const useDashboardActions = () => {
  const createDashboard = useDashboardStore((state) => state.createDashboard);
  const updateDashboard = useDashboardStore((state) => state.updateDashboard);
  const saveDashboard = useDashboardStore((state) => state.saveDashboard);
  const deleteDashboard = useDashboardStore((state) => state.deleteDashboard);
  const duplicateDashboard = useDashboardStore((state) => state.duplicateDashboard);
  const addPanel = useDashboardStore((state) => state.addPanel);
  const updatePanel = useDashboardStore((state) => state.updatePanel);
  const updatePanelOptions = useDashboardStore((state) => state.updatePanelOptions);
  const batchUpdatePanels = useDashboardStore((state) => state.batchUpdatePanels);
  const deletePanel = useDashboardStore((state) => state.deletePanel);
  const duplicatePanel = useDashboardStore((state) => state.duplicatePanel);
  const getPanel = useDashboardStore((state) => state.getPanel);
  const setFilters = useDashboardStore((state) => state.setFilters);
  const updateFavorite = useDashboardStore((state) => state.updateFavorite);
  const reset = useDashboardStore((state) => state.reset);

  return {
    createDashboard,
    updateDashboard,
    saveDashboard,
    deleteDashboard,
    duplicateDashboard,
    addPanel,
    updatePanel,
    updatePanelOptions,
    batchUpdatePanels,
    deletePanel,
    duplicatePanel,
    getPanel,
    setFilters,
    updateFavorite,
    reset,
  };
};

// ─── 자동 대시보드 로딩 훅 ───────────────────────────────────────────────────

export const useDashboardAutoLoad = (dashboardId: string) => {
  const _loadDashboard = useDashboardStore((state) => state._loadDashboard);
  const currentId = useDashboardStore((state) => state.id);
  const { search } = useLocation();
  const { dataSourceList, isLoading: isDsLoading } = useDatasourceList();

  // search는 deps에 포함하지 않음:
  // Store→URL sync는 use-url-sync.ts subscribe(window.history.replaceState)가 처리하며,
  // React Router의 useLocation().search는 replaceState를 감지하지 못해 stale하다.
  // search가 deps에 있으면 sub-page 왕복 시 useEffect가 재실행되어
  // 이미 store에 올바른 상태가 있는데도 _loadDashboard를 다시 호출하거나
  // 불필요한 초기화가 발생한다.
  const searchRef = React.useRef(search);
  searchRef.current = search;

  const datasourceTypeMap = React.useMemo(
    () => Object.fromEntries(dataSourceList.map(ds => [ds.value, ds.type])),
    [dataSourceList],
  );

  React.useEffect(() => {
    if (!dashboardId || dashboardId === currentId) return;
    if (isDsLoading) return;
    void _loadDashboard(dashboardId, {
      urlParams: new URLSearchParams(searchRef.current),
      datasourceTypeMap,
    });
  }, [dashboardId, currentId, _loadDashboard, isDsLoading, datasourceTypeMap]);
};

// ─── 전역 설정 ───────────────────────────────────────────────────────────────

// 로그아웃 시 dashboard store 초기화
registerOnLogout(() => {
  useDashboardStore.getState().reset();
});

// URL 자동 동기화 설정 (Store → URL)
setupUrlSync(useDashboardStore);
