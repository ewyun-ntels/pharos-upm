import {useMemo} from 'react';
import {useDashboardStore} from './use-dashboard-store';

/**
 * Dashboard Variable State Custom Hook
 * 
 * 중복되는 variable state 구독 로직을 한 곳에 모음
 */
export function useVariableState() {
  const filterMetas = useDashboardStore((state) => state.filterState?.filterMetas ?? new Map());
  const datetime = useDashboardStore((state) => state.filterState?.datetime);
  const step = useDashboardStore((state) => state.filterState?.step);
  const refreshCount = useDashboardStore((state) => state.filterState?.refreshCount ?? 0);

  const filters = useMemo(() => {
    const map = new Map();
    filterMetas.forEach((meta, key) => {
      if (meta.value !== undefined) map.set(key, meta.value);
    });
    return map;
  }, [filterMetas]);

  const filterIds = useMemo(() => Array.from(filterMetas.keys()), [filterMetas]);

  return {
    filterMetas,
    filterValues: filters,  // ✅ filters → filterValues로 변경 (FilterConfig[]와 충돌 방지)
    filterIds,
    datetime,
    step,
    refreshCount,
  };
}
