import React, {useEffect, useState} from 'react';
import parse from 'parse-duration';
import {VariablePlugin, variablePluginRegistry, VariableProps} from '@features/dashboard/variables';
import {RefreshCcw} from '@pharos/shared/components';
import {IconButton, VariableSelect} from '@pharos/shared/components/ui-extension';
import {useDashboardStore} from '@features/dashboard/hooks/use-dashboard-store';
import {formatRefreshInterval} from '@features/dashboard/utils/datetime-step-utils';
import type {RefreshFilterOptions, RefreshIntervalType} from './types';
import {REFRESH_INTERVAL, REFRESH_INTERVAL_DEFAULT} from './types';
import {RefreshVariableEditor} from './RefreshVariableEditor';
import {RefreshVariablePreview} from './RefreshVariablePreview';

/**
 * Refresh Variable Component
 * URL 동기화는 subscribe에서 자동 처리
 */
export const RefreshVariable: React.FC<VariableProps<RefreshFilterOptions>> = (props) => {
  const {options} = props;

  const refreshInterval = useDashboardStore((state) => state.filterState?.refreshInterval);
  const setRefreshInterval = useDashboardStore((state) => state.setRefreshInterval);
  const incrementRefreshCount = useDashboardStore((state) => state.incrementRefreshCount);

  const currentRefreshValue = formatRefreshInterval(refreshInterval) as RefreshIntervalType;
  const [refetchIntervalValue, setRefetchIntervalValue] = useState<RefreshIntervalType>(
    currentRefreshValue || options?.initValue || REFRESH_INTERVAL_DEFAULT
  );

  // refreshInterval이 store에서 변경되면 UI 업데이트
  useEffect(() => {
    if (currentRefreshValue) {
      setRefetchIntervalValue(currentRefreshValue);
    }
  }, [currentRefreshValue]);

  // Auto-refresh timer
  useEffect(() => {
    if (!refreshInterval) return;
    const timer = setInterval(() => {
      incrementRefreshCount();
    }, refreshInterval);
    return () => clearInterval(timer);
  }, [refreshInterval, incrementRefreshCount]);

  const handleBtnClick = () => {
    incrementRefreshCount();
  };

  // 사용 가능한 refresh interval 옵션 (설정된 refreshOptions 또는 기본값)
  const availableOptions = options?.refreshOptions && options.refreshOptions.length > 0
    ? options.refreshOptions
    : REFRESH_INTERVAL;

  return (
    <div className="flex flex-row h-8 items-center rounded-md border border-input bg-background overflow-hidden focus-within:border-ring focus-within:ring-ring/50 focus-within:ring-[3px] transition-[color,box-shadow]">
      <IconButton 
        onClick={handleBtnClick}
        variant="ghost"
        icon={<RefreshCcw className="!w-[15px] !h-[15px]"/>}
        className="border-0 border-r !border-input rounded-none w-8 h-8"
      >
        Refresh
      </IconButton>
      <VariableSelect
        options={availableOptions.map(v => ({ label: v, value: v }))}
        value={refetchIntervalValue}
        onChange={(v) => {
          const value = v as RefreshIntervalType;
          const duration = parse(value);
          if (duration) setRefreshInterval(duration);
          else setRefreshInterval(false);

          setRefetchIntervalValue(value);

          // URL 동기화는 subscribe에서 자동 처리 (use-url-sync.ts)
        }}
        searchable={false}
        showLabel={false}
      />
    </div>
  );
};

/**
 * Refresh Variable Plugin
 */
export const refreshVariablePlugin: VariablePlugin<RefreshFilterOptions> = {
  info: {
    id: 'refreshInterval',
    label: 'Refresh',
    description: 'Refresh interval selector with manual refresh button',
    category: 'control',
    requiresData: false,
  },
  component: RefreshVariable,
  editor: RefreshVariableEditor,
  preview: RefreshVariablePreview,
};

// Register
variablePluginRegistry.register(refreshVariablePlugin);
