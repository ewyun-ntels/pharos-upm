'use client';
import React, {useCallback, useMemo} from 'react';
// Import variables to trigger plugin registration
import '@features/dashboard/variables';
import {variablePluginRegistry} from '@features/dashboard/variables';
import type {FilterConfig} from './types';
import {useDashboardStore} from '../../hooks/use-dashboard-store';
import type {FilterValue} from '../../types/filter.types';
import {useDashboardData} from '../../hooks/use-dashboard-store';

/**
 * HeaderItem - Optimized Registry 기반
 * 
 * PanelRenderer 패턴 적용:
 * 1. 컴포넌트 직접 렌더링 (registry.render 제거)
 * 2. useMemo로 컴포넌트 캐싱 (type 변경 시에만 재생성)
 * 3. props memoization (불필요한 리렌더 방지)
 */
const HeaderItemComponent = ({
  item,
  dashboardId,
  index,
}: {
  item: FilterConfig;
  index: number;
  dashboardId: string;
}) => {
  const setFilter = useDashboardStore((state) => state.setFilter);
  const { permission } = useDashboardData();

  // ✅ OPTIMIZATION 1: Stable callback reference
  const handleVariableChange = useCallback(
    (value: FilterValue) => {
      setFilter(item.id, value);
    },
    [item.id, setFilter],
  );

  // ✅ OPTIMIZATION 2: Cached Variable Component (type 변경 시에만 재생성)
  const VariableComponent = useMemo(() => {
    const plugin = variablePluginRegistry.get(item.type);
    return plugin?.component || null;
  }, [item.type]);

  // ✅ OPTIMIZATION 3: Memoized variable props (안정적인 참조로 리렌더 방지)
  const variableProps = useMemo(() => ({
    ...item,
    dashboardId,
    index,
    onChange: handleVariableChange,
    permission,
  }), [
    item,
    dashboardId,
    index,
    handleVariableChange,
    permission,
  ]);

  // Early return if component not found
  if (!VariableComponent) {
    return null;
  }

  // ✅ OPTIMIZATION 4: Direct rendering (no registry.render overhead)
  return <VariableComponent {...variableProps} />;
};

/**
 * Memoized wrapper to prevent re-renders when parent updates
 * but props haven't changed
 */
export const HeaderItem = React.memo(HeaderItemComponent);

HeaderItem.displayName = 'HeaderItem';

/**
 * VariableItems - Registry 기반
 * 
 * ✅ 관심사 분리 (Separation of Concerns):
 * - VariableItems: 변수 렌더링만 담당
 * - DateTimeVariable: datetime 관련 로직 처리
 * - StepVariable: step 관련 로직 처리
 * - useDateTimeStepSync: datetime-step 연동 로직
 */
export const VariableItems = ({
  dashboardId,
  items,
  wrap = true,
}: {
  dashboardId: string;
  items?: FilterConfig[];
  wrap?: boolean;
}) => {
  return (
    <div className={`flex ${wrap ? 'flex-wrap' : 'flex-nowrap'} gap-2`}>
      {items?.map((item, index) => (
        <HeaderItem
          key={item.id}
          item={item}
          index={index}
          dashboardId={dashboardId}
        />
      ))}
    </div>
  );
};
