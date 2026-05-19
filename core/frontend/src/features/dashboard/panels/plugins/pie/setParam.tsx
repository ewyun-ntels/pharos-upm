
import type { Panel } from '@pharos/shared/types/dashboard';
import type { PanelProps } from '@pharos/core/panel-registry';
import type { PiePanelOptions } from './types';
import { createDashboardChartProvider } from '../common/setParam';

export const setParamToChart = (panel: Panel): Partial<PanelProps<PiePanelOptions>> => {
  return {
    ...createDashboardChartProvider(panel.dataProvider?.chartQuery),
    options: panel.options,
  };
};

export const setDefaultParam = (panelData: PanelProps<PiePanelOptions>) => {
  return {
    ...createDashboardChartProvider(panelData.dataProvider?.chartQuery),
    options: {
      ...panelData.options,
      innerRadius: panelData.options?.innerRadius || 40,
      outerRadius: panelData.options?.outerRadius || 55,
      strokeWidth: panelData.options?.strokeWidth ?? 2,
    },
  };
};
