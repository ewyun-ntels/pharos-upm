import type { Panel } from '@pharos/shared/types/dashboard';
import type { PanelProps } from '@pharos/core/panel-registry';
import type { BarPercentGaugePanelOptions } from './types';
import { createDashboardChartProvider } from '../common/setParam';

export const setParamToChart = (panel: Panel): Partial<PanelProps<BarPercentGaugePanelOptions>> => {
  return {
    ...createDashboardChartProvider(panel.dataProvider?.chartQuery),
    options: {
      chartColor: panel.options?.chartColor,
      showType: panel.options?.showType || 'last',
      legendRule: panel.options?.legendRule,
      subName: panel.options?.subName,
      gaugeMaxValue: Number(panel.options?.gaugeMaxValue) || 100,
      unitType: panel.options?.unitType,
      chartDataNotExistMessage: panel.options?.chartDataNotExistMessage,
    },
  };
};

export const setDefaultParam = (panelData: PanelProps<BarPercentGaugePanelOptions>) => {
  return {
    ...createDashboardChartProvider(panelData.dataProvider?.chartQuery),
    options: panelData.options,
  };
};
