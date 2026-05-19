import React from 'react';
import {HeaderItem} from '@features/dashboard/components/VariableBar/VariableItems';
import type {VariablePreviewProps} from '../../registry/VariablePluginRegistry';
import type {RefreshFilterOptions} from './types';
import type {FilterConfig} from '@pharos/shared/types/dashboard';

export function RefreshVariablePreview({data, dashboardId}: VariablePreviewProps<RefreshFilterOptions>) {
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
