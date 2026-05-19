import type { Panel } from '@pharos/shared/types/dashboard';
import type { PanelProps } from '@pharos/core/panel-registry';
import type { HistogramPanelOptions } from './types';
import { createDashboardChartProvider } from '../common/setParam';

export const setParamToChart = (panel: Panel): Partial<PanelProps<HistogramPanelOptions>> => {
  return {
    ...createDashboardChartProvider(panel.dataProvider?.chartQuery),
    options: {
      bucketCount: panel.options?.bucketCount,
      legendRule: panel.options?.legendRule || '{col} {col_value}',
      yaxis: panel.options?.yaxis,
      tickFont: panel.options?.tickFont,
      stackId: panel.options?.stackId,
      fillOpacity: panel.options?.fillOpacity,
      showAverage: panel.options?.showAverage,
      showLast: panel.options?.showLast,
      showMax: panel.options?.showMax,
      showMin: panel.options?.showMin,
      showTotal: panel.options?.showTotal,
      legendAsTable: panel.options?.legendAsTable,
      legendEnabled: panel.options?.legendEnabled,
      legendOption: (panel.options?.legendOption || 'bottom') as 'bottom' | 'right',
      unit: panel.options?.unit || 'short',
      selectedUnit: panel.options?.selectedUnit,
    },
  };
};

export const setDefaultParam = (panelData: PanelProps<HistogramPanelOptions>) => {
  return {
    ...createDashboardChartProvider(panelData.dataProvider?.chartQuery),
    options: panelData.options,
  };
};
