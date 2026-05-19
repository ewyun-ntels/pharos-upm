import { useMemo } from 'react';
import type { FilterConfig } from '@pharos/shared/types/dashboard';

interface VariableLayoutGroups {
  headerL: FilterConfig[];
  headerR: FilterConfig[];
}

/**
 * 변수 목록을 위치(position)별로 분류하는 훅
 */
export function useVariableLayout(filters: FilterConfig[]): VariableLayoutGroups {
  return useMemo(() => {
    const headerL = filters.filter(f => f.kind === 'headerL');
    const headerR = filters.filter(f => f.kind === 'headerR');
    return {
      headerL,
      headerR,
    };
  }, [filters]);
}
