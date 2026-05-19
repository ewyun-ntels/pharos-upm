import React from 'react';
import {VariablePlugin, variablePluginRegistry, VariableProps} from '@features/dashboard/variables';
import {
  DatetimeRange,
  DateTimeRangeValue,
  getRelativeRangeValue,
} from '@pharos/shared/components/ui-extension';
import {useDashboardStore} from '@features/dashboard/hooks/use-dashboard-store';
import {useDateTimeStepSync} from '@features/dashboard/variables/hooks/useDateTimeStepSync';
import type {DateTimeFilterOptions} from './types';
import {DatetimeVariableEditor} from './DatetimeVariableEditor';
import {DatetimeVariablePreview} from './DatetimeVariablePreview';

/**
 * DateTime Variable Component
 *
 * DateTime과 Step 변수의 연동을 자동으로 처리합니다.
 * - rangeStepOptions가 있으면 datetime 변경 시 자동으로 stepOptions 업데이트
 * - URL 파라미터: from, to (Grafana 호환, 자동 동기화)
 */
export const DateTimeVariable: React.FC<VariableProps<DateTimeFilterOptions>> = () => {
  const datetime = useDashboardStore((state) => state.filterState?.datetime);
  const setDatetime = useDashboardStore((state) => state.setDatetime);

  useDateTimeStepSync();

  const handleChange = (startTime: DateTimeRangeValue, endTime: DateTimeRangeValue) => {
    setDatetime(startTime, endTime);
  };

  return (
    <DatetimeRange
      startTime={datetime?.startTime ?? getRelativeRangeValue(1, 'Hours ago')}
      endTime={datetime?.endTime ?? getRelativeRangeValue(true)}
      onChange={handleChange}
      size="small"
      className="w-auto min-w-[280px]"
    />
  );
};

/**
 * DateTime Variable Plugin
 */
export const dateTimeVariablePlugin: VariablePlugin<DateTimeFilterOptions> = {
  info: {
    id: 'datetime',
    label: 'Date Time Range',
    description: 'Date and time range selector',
    category: 'time',
    requiresData: false,
  },
  component: DateTimeVariable,
  editor: DatetimeVariableEditor,
  preview: DatetimeVariablePreview,
};

// Register
variablePluginRegistry.register(dateTimeVariablePlugin);
