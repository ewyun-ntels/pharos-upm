/**
 * Dashboard CRUD Slice
 *
 * 대시보드 로드/생성/수정/저장/삭제/복제/즐겨찾기 + 기본 상태.
 */
import { dashboardProvider, DASHBOARD_RESOURCES } from '@providers/dashboard-provider';
import { ActionSchema } from '@pharos/shared/types/dashboard';
import type { DashboardCreateResponse, DashboardData } from '@pharos/shared/types/dashboard';
import { dateTimeValueToUrlString } from '@pharos/shared/components/ui-extension';
import { withSuppressedUrlSyncAsync, DashboardUrlManager } from '../url';
import { resolveInitialState } from '../init/resolve-initial-state';
import { formatRefreshInterval } from '../../utils/datetime-step-utils';
import { createFilterState, resetDependencyGraph } from './types';
import type { StoreSet, StoreGet, DashboardStore } from './types';
import { fromArray, toList } from './panelCollection';

// 동시 로드 요청 중 최신 요청만 반영 (race condition 방지)
let _loadRequestId = 0;

export type DashboardCrudSlice = Pick<
  DashboardStore,
  | 'id' | 'permission' | 'dashboard' | 'hasUnsavedChanges' | 'loading' | 'error'
  | 'filters'
  | 'storeAnnotations'
  | '_loadDashboard' | 'createDashboard' | 'updateDashboard' | 'saveDashboard'
  | 'deleteDashboard' | 'duplicateDashboard' | 'updateFavorite' | 'setFilters'
  | 'setAnnotations' | 'addAnnotation' | 'removeAnnotation'
  | 'reset'
>;

export const createDashboardCrudSlice = (set: StoreSet, get: StoreGet): DashboardCrudSlice => ({
  // === 초기 상태 ===
  id: undefined,
  dashboard: undefined,
  permission: ActionSchema.enum.viewer,
  hasUnsavedChanges: false,
  loading: false,
  error: null,
  filters: [],
  storeAnnotations: [],

  // === 대시보드 로드 ===
  _loadDashboard: async (dashboardId, options) => {
    const forceReload = options?.forceReload ?? false;
    const urlParams = options?.urlParams;
    const datasourceTypeMap = options?.datasourceTypeMap ?? {};
    const { id: loadedId } = get();

    // 같은 대시보드 이미 로드됨: skip (강제 reload는 forceReload=true)
    if (dashboardId === loadedId && !forceReload) return;

    // 요청 ID로 race condition 방지: 더 최신 요청이 오면 이전 fetch 결과 무시
    _loadRequestId++;
    const myRequestId = _loadRequestId;

    set({
      loading: true,
      error: null,
      filterState: createFilterState(),
    });

    try {
      const result = await dashboardProvider.getOne({
        resource: DASHBOARD_RESOURCES.DASHBOARD,
        id: dashboardId,
      });

      // 이미 더 새로운 로드 요청이 시작된 경우 무시
      if (_loadRequestId !== myRequestId) return;

      const dashboardData = result.data as DashboardData;
      if (!dashboardData?.config) {
        throw new Error('No dashboard config returned');
      }

      const panels = dashboardData.config?.panels || [];
      const panelCollection = fromArray(panels);

      const filters = dashboardData.config?.filters || [];

      set({
        id: dashboardId,
        dashboard: dashboardData.config,
        permission: dashboardData.permission,
        panelOrder: panelCollection.ids,
        panelMap: panelCollection.entities,
        filters,
        storeAnnotations: dashboardData.annotations ?? [],
        loading: false,
        error: null,
      });

      const { setFilterMetas, setFilter } = get();

      const filterMetasData = filters
        .filter(f => f.query && f.datasourceName)
        .map(f => ({
          id: f.id,
          query: f.query || '',
          datasourceName: f.datasourceName || '',
          datasourceType: datasourceTypeMap[f.datasourceName || ''],
          options: f.options,
        }));

      // datasource filter 유무와 무관하게 항상 datetime/step 초기 상태를 적용한다.
      // 기존: filterMetasData.length > 0 일 때만 resolveInitialState를 호출하여
      //       datasource filter가 없는 대시보드는 savedValue가 완전히 무시되었다.
      await withSuppressedUrlSyncAsync(async () => {
        if (filterMetasData.length > 0) {
          const metaResult = setFilterMetas(filterMetasData);
          if (!metaResult.success) return;
        }

        const initialState = await resolveInitialState({
          urlParams: urlParams ?? new URLSearchParams(),
          filters,
        });

        if (filterMetasData.length > 0) {
          initialState.filterValues.forEach((value, filterId) => {
            setFilter(filterId, value, { skipCascade: true });
          });
        }

        if (initialState.step) {
          get().setStep(initialState.step);
        }

        if (initialState.datetime) {
          get().setDatetime(
            initialState.datetime.startTime,
            initialState.datetime.endTime,
            { skipRefreshCount: true },
          );
        }

        if (initialState.refreshInterval !== undefined) {
          get().setRefreshInterval(initialState.refreshInterval);
        }
      });

      // suppressSync 해제 전에 subscribe가 이미 억제되었으므로
      // 초기 상태(savedValue 포함)를 URL에 명시적으로 반영한다.
      const finalState = get().filterState;
      if (finalState.datetime) {
        DashboardUrlManager.syncDatetime(finalState.datetime);
      }
      if (finalState.step) {
        DashboardUrlManager.syncStep(finalState.step);
      }
      if (finalState.refreshInterval !== undefined) {
        DashboardUrlManager.syncRefreshInterval(formatRefreshInterval(finalState.refreshInterval));
      }
      if (finalState.filterMetas.size > 0) {
        DashboardUrlManager.syncFilters(finalState.filterMetas);
      }
    } catch (error) {
      if (_loadRequestId !== myRequestId) return;
      console.error('Error fetching dashboard:', error);
      set({
        error,
        loading: false,
        id: undefined,
        dashboard: undefined,
        permission: ActionSchema.enum.viewer,
        panelOrder: [],
        panelMap: {},
        filters: [],
      });
    }
  },

  createDashboard: async (data) => {
    set({ loading: true, error: null });

    try {
      const result = await dashboardProvider.create({
        resource: DASHBOARD_RESOURCES.DASHBOARD,
        variables: data,
      });

      set({ loading: false });
      return result.data as DashboardCreateResponse;
    } catch (error) {
      set({ error, loading: false });
      throw error;
    }
  },

  updateDashboard: (updates) => {
    const { dashboard, panelOrder: currentPanelOrder, panelMap: currentPanelMap } = get();
    if (!dashboard) return;

    const newDashboard = { ...dashboard, ...updates };
    let panelOrder = currentPanelOrder;
    let panelMap = currentPanelMap;

    if (Object.prototype.hasOwnProperty.call(updates, 'panels')) {
      const panelCollection = fromArray(newDashboard.panels || []);
      panelOrder = panelCollection.ids;
      panelMap = panelCollection.entities;
    }

    const filters = newDashboard.filters || [];

    set({
      dashboard: newDashboard,
      filters,
      panelOrder,
      panelMap,
      hasUnsavedChanges: true,
    });
  },

  saveDashboard: async (options) => {
    const { dashboard, panelOrder, panelMap, filters, filterState, id, loading } = get();
    if (!dashboard || !id || loading) return;

    try {
      const updatedFilters = filters.map(f => {
        if (f.type === 'datetime' && filterState.datetime && options?.saveTimeRange) {
          const { startTime, endTime } = filterState.datetime;

          const isStartNow = startTime.type === 'relative' && startTime.relativeNow === true;
          const isEndNow = endTime.type === 'relative' && endTime.relativeNow === true;

          if (isStartNow && isEndNow) {
            console.warn('[saveDashboard] Invalid datetime range: both start and end are "now". Skipping save.');
            return f;
          }

          const startTimeStr = dateTimeValueToUrlString(startTime);
          const endTimeStr = dateTimeValueToUrlString(endTime);

          return {
            ...f,
            options: {
              ...f.options,
              savedStartTime: startTimeStr,
              savedEndTime: endTimeStr,
            }
          };
        }
        if (f.type === 'step' && filterState.step && options?.saveStep) {
          return {
            ...f,
            options: {
              ...f.options,
              savedValue: filterState.step.step,
            }
          };
        }
        if (f.type === 'refreshInterval' && filterState.refreshInterval !== undefined && options?.saveRefreshInterval) {
          // refreshInterval을 문자열로 변환 (예: 30000 → '30s', false → 'Off')
          const savedValue = formatRefreshInterval(filterState.refreshInterval);
          return {
            ...f,
            options: {
              ...f.options,
              savedValue,
            }
          };
        }
        return f;
      });

      const panels = toList({ ids: panelOrder, entities: panelMap });

      const dashboardToSave = {
        ...dashboard,
        panels,
        filters: updatedFilters,
      };

      await dashboardProvider.update({
        resource: DASHBOARD_RESOURCES.DASHBOARD,
        id,
        variables: dashboardToSave,
      });

      set({
        filters: updatedFilters,
        hasUnsavedChanges: false
      });
    } catch (error) {
      set({ error });
      throw error;
    }
  },

  deleteDashboard: async (dashboardId) => {
    set({ loading: true, error: null });

    try {
      await dashboardProvider.deleteOne({
        resource: DASHBOARD_RESOURCES.DASHBOARD,
        id: dashboardId,
      });

      const { id } = get();
      if (id === dashboardId) {
        get().reset();
      }

      set({ loading: false });
    } catch (error) {
      set({ error, loading: false });
      throw error;
    }
  },

  duplicateDashboard: async (dashboardId) => {
    set({ loading: true, error: null });

    try {
      const original = await dashboardProvider.getOne({
        resource: DASHBOARD_RESOURCES.DASHBOARD,
        id: dashboardId,
      });

      const duplicated = {
        ...original.data.config,
        title: `${original.data.config.title} (Copy)`,
        displayName: `${original.data.config.displayName} (Copy)`,
      };

      const result = await dashboardProvider.create({
        resource: DASHBOARD_RESOURCES.DASHBOARD,
        variables: duplicated,
      });

      set({ loading: false });
      return result.data as DashboardCreateResponse;
    } catch (error) {
      set({ error, loading: false });
      throw error;
    }
  },

  updateFavorite: async (dashboardId, favorite, _data) => {
    set({ loading: true, error: null });

    try {
      await dashboardProvider.update({
        resource: DASHBOARD_RESOURCES.FAVORITES,
        id: dashboardId,
        variables: {},
      });

      const { id, dashboard } = get();
      if (id === dashboardId && dashboard) {
        set({
          dashboard: { ...dashboard, favorite },
        });
      }

      set({ loading: false });
    } catch (error) {
      set({ error, loading: false });
      throw error;
    }
  },

  setFilters: (newFilters) => {
    const { dashboard } = get();
    if (!dashboard) return;

    const filters = typeof newFilters === 'function' ? newFilters(get().filters) : newFilters;

    set({
      filters,
      dashboard: {
        ...dashboard,
        filters,
      },
      hasUnsavedChanges: true,
    });
  },

  reset: () => {
    resetDependencyGraph();
    set({
      id: undefined,
      dashboard: undefined,
      permission: ActionSchema.enum.viewer,
      panelOrder: [],
      panelMap: {},
      filters: [],
      storeAnnotations: [],
      filterState: createFilterState(),
      loading: false,
      error: null,
      hasUnsavedChanges: false,
    });
  },

  setAnnotations: (annotations) => set({ storeAnnotations: annotations }),

  addAnnotation: async (annotation) => {
    const { id, storeAnnotations } = get();
    if (!id) return;
    const next = [...storeAnnotations, annotation];
    set({ storeAnnotations: next });
    await dashboardProvider.update({
      resource: DASHBOARD_RESOURCES.ANNOTATIONS,
      id,
      variables: next,
    });
  },

  removeAnnotation: async (annotationId) => {
    const { id, storeAnnotations } = get();
    if (!id) return;
    const next = storeAnnotations.filter((a) => a.id !== annotationId);
    set({ storeAnnotations: next });
    await dashboardProvider.update({
      resource: DASHBOARD_RESOURCES.ANNOTATIONS,
      id,
      variables: next,
    });
  },
});
