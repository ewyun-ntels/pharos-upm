import type { Panel } from '@pharos/shared/types/dashboard';
import type { PanelProps } from '@pharos/core/panel-registry';
import type { CustomAlertPanelOptions } from './types';

export const setParamToChart = (_panel: Panel): Partial<PanelProps<CustomAlertPanelOptions>> => {
  return {};
};

export const setDefaultParam = (_panelData: PanelProps<CustomAlertPanelOptions>) => {
  return {};
};
