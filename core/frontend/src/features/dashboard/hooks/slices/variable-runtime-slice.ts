/**
 * Variable Runtime Slice
 *
 * 변수 런타임 상태 관리 (filterMetas, datetime, step, refreshCount 등).
 * DependencyGraph는 store 외부 ref로 관리 (Immer freeze 호환을 위해).
 */
import { produce } from 'immer';
import { DependencyGraph } from '../../utils/dependency-graph';
import {
  createFilterState,
  getDependencyGraph,
  setDependencyGraph,
} from './types';
import type { FilterMeta, StoreSet, StoreGet, DashboardStore } from './types';

export type VariableRuntimeSlice = Pick<
  DashboardStore,
  | 'filterState'
  | 'setFilter' | 'getFilter' | 'setFilterMeta'
  | 'setDatetime' | 'setStep' | 'setRefreshInterval' | 'incrementRefreshCount'
  | 'setFilterMetas' | 'setStepOptions' | 'setRangeStepOptions'
>;

export const createVariableRuntimeSlice = (set: StoreSet, get: StoreGet): VariableRuntimeSlice => ({
  filterState: createFilterState(),

  // FilterMeta 기반 값 설정 (URL 동기화는 subscribe에서 자동 처리)
  setFilter: (id, value, options) => {
    const skipCascade = options?.skipCascade ?? false;

    set(
      produce((draft: DashboardStore) => {
        let meta = draft.filterState.filterMetas.get(id);

        if (!meta) {
          meta = {
            id,
            query: '',
            datasourceName: '',
            value: undefined,
            isFetching: false,
            isQuerySuccess: false,
          };
          draft.filterState.filterMetas.set(id, meta);
        }

        const currentValue = meta.value;
        meta.value = value;

        // Cascade: 외부 ref의 graph를 읽어서 dependents 초기화
        if (!skipCascade && currentValue !== value) {
          const graph = getDependencyGraph();
          const allDependents = graph.getAllDependents(id);
          if (allDependents.length > 0) {
            allDependents.forEach((depId: string) => {
              const depMeta = draft.filterState.filterMetas.get(depId);
              if (depMeta) {
                depMeta.value = undefined;
                depMeta.isQuerySuccess = false;
              }
            });
          }
        }
      }),
    );
  },

  getFilter: (id) => {
    const meta = get().filterState.filterMetas.get(id);
    return meta?.value;
  },

  setFilterMeta: (id, updates) => {
    set(
      produce((draft: DashboardStore) => {
        let meta = draft.filterState.filterMetas.get(id);

        if (!meta) {
          meta = {
            id,
            query: '',
            datasourceName: '',
            value: undefined,
            isFetching: false,
            isQuerySuccess: false,
          };
          draft.filterState.filterMetas.set(id, meta);
        }

        Object.assign(meta, updates);
      }),
    );

    // Self-registration 시 query가 추가되면 graph rebuild
    if (updates.query || updates.datasourceName) {
      queueMicrotask(() => {
        const state = get();
        const filterMetas = state.filterState.filterMetas;

        const validMetas = Array.from(filterMetas.values())
          .filter(m => m.query && m.datasourceName)
          .map(m => ({
            id: m.id,
            query: m.query,
            datasourceName: m.datasourceName,
          }));

        if (validMetas.length > 0) {
          const graph = new DependencyGraph();
          graph.buildFromVariables(validMetas);
          setDependencyGraph(graph);
        }
      });
    }
  },

  setDatetime: (startTime, endTime, options) => {
    const skipRefreshCount = options?.skipRefreshCount ?? false;

    set(
      produce((draft: DashboardStore) => {
        draft.filterState.datetime = { startTime, endTime };
        if (!skipRefreshCount) {
          draft.filterState.refreshCount += 1;
        }
      }),
    );
  },

  setStep: (step) => {
    set(
      produce((draft: DashboardStore) => {
        draft.filterState.step = step;
      }),
    );
  },

  setRefreshInterval: (interval) => {
    set(
      produce((draft: DashboardStore) => {
        draft.filterState.refreshInterval = interval;
      }),
    );
  },

  incrementRefreshCount: () => {
    set(
      produce((draft: DashboardStore) => {
        draft.filterState.refreshCount += 1;
      }),
    );
  },

  setFilterMetas: (filterMetas) => {
    const errors: string[] = [];

    try {
      const graph = new DependencyGraph();
      graph.buildFromVariables(filterMetas);

      const cycles = graph.detectCycles();
      if (cycles.length > 0) {
        cycles.forEach((cycle) => errors.push(`Circular dependency: ${cycle.join(' → ')}`));
        return { success: false, errors };
      }

      // graph는 store 외부 ref에 저장 (Immer freeze 우회)
      setDependencyGraph(graph);

      set(
        produce((draft: DashboardStore) => {
          const existingMetas = new Map<string, FilterMeta>(draft.filterState.filterMetas);
          draft.filterState.filterMetas.clear();

          filterMetas.forEach((meta) => {
            const existing = existingMetas.get(meta.id);
            draft.filterState.filterMetas.set(meta.id, {
              ...meta,
              value: existing?.value,
              isFetching: existing?.isFetching ?? false,
              isQuerySuccess: existing?.isQuerySuccess ?? false,
              error: existing?.error,
              data: existing?.data,
            });
          });
        }),
      );

      return { success: true, errors: [] };
    } catch (e) {
      errors.push(e instanceof Error ? e.message : String(e));
      return { success: false, errors };
    }
  },

  setStepOptions: (options) => {
    set(
      produce((draft: DashboardStore) => {
        draft.filterState.stepOptions = options;
      }),
    );
  },

  setRangeStepOptions: (options) => {
    set(
      produce((draft: DashboardStore) => {
        draft.filterState.rangeStepOptions = options;
      }),
    );
  },
});
