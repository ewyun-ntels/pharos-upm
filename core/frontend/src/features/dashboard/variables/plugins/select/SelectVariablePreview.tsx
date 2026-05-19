import React, {useEffect} from 'react';
import {HeaderItem} from '@features/dashboard/components/VariableBar/VariableItems';
import {useDashboardStore} from '@features/dashboard/hooks/use-dashboard-store';
import type {VariablePreviewProps} from '../../registry/VariablePluginRegistry';
import type {SelectFilterOptions} from './types';
import type {FilterConfig} from '@pharos/shared/types/dashboard';
import {getRelativeRangeValue} from '@pharos/shared/components/ui-extension';

export function SelectVariablePreview({data, dashboardId}: VariablePreviewProps<SelectFilterOptions>) {
  const setFilter = useDashboardStore((state) => state.setFilter);
  const setDatetime = useDashboardStore((state) => state.setDatetime);

  // Initialize variable store with default datetime (useEffect로 감싸야 함)
  useEffect(() => {
    if (data.id) {
      setFilter(data.id, '');
    }
    
    // ✅ Preview 환경: 기본 시간 범위 설정 (최근 1시간)
    const startTime = getRelativeRangeValue(1, 'Hours ago');
    const endTime = getRelativeRangeValue(true); // Now
    setDatetime(startTime, endTime);
  }, [data.id, setFilter, setDatetime]);

  // Validation
  if (!data.id?.trim()) {
    return (
      <div className="p-4 text-center text-muted-foreground">
        Please enter an ID to preview.
      </div>
    );
  }

  if (!data.query?.trim()) {
    return (
      <div className="p-4 text-center text-muted-foreground">
        Please enter a query to preview.
      </div>
    );
  }

  const filterItem: FilterConfig = {
    ...data,
    kind: data.kind as 'headerL' | 'headerR' | 'left',
  };

  return (
    <div className="p-4">
      <HeaderItem
        item={filterItem}
        index={0}
        dashboardId={dashboardId || 'preview'}
      />
    </div>
  );
}
