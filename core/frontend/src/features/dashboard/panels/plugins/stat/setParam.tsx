
import type { Panel } from '@pharos/shared/types/dashboard';
import type { PanelProps } from '@pharos/core/panel-registry';
import type { StatPanelOptions } from './types';
import { createDashboardChartProvider } from '../common/setParam';

export const setParamToChart = (panel: Panel): Partial<PanelProps<StatPanelOptions>> => {
  return {
    ...createDashboardChartProvider(panel.dataProvider?.chartQuery),
    options: panel.options,
  };
};

export const setDefaultParam = (panelData: PanelProps<StatPanelOptions>) => {
  return {
    ...createDashboardChartProvider(panelData.dataProvider?.chartQuery),
  };
};
