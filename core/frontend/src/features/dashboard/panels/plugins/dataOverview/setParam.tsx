import type { Panel } from '@pharos/shared/types/dashboard';
import type { PanelProps } from '@pharos/core/panel-registry';
import type { DataOverviewPanelOptions } from './types';

export const setParamToChart = (panel: Panel): Partial<PanelProps<DataOverviewPanelOptions>> => {
  return {
    options: {
      datas: panel.options?.datas || [],
      layoutCol: panel.options?.layoutCol || false,
      titleTop: panel.options?.titleTop || false,
      largeText: panel.options?.largeText || false,
      right: panel.options?.right || false,
      chartDataNotExistMessage: panel.options?.chartDataNotExistMessage,
    },
  };
};

export const setDefaultParam = (panelData: PanelProps<DataOverviewPanelOptions>) => {
  return {
    options: panelData.options,
    datas: panelData.options?.datas || [],
  };
};
