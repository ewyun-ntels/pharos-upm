import type { Panel } from '@pharos/shared/types/dashboard';
import type { PanelProps } from '@pharos/core/panel-registry';
import type { MarkdownViewerPanelOptions } from './types';

export const setParamToChart = (panel: Panel): Partial<PanelProps<MarkdownViewerPanelOptions>> => {
  const content = panel.dataProvider?.chartQuery?.[0]?.query || '';

  return {
    options: {
      content,
      chartDataNotExistMessage: panel.options?.chartDataNotExistMessage,
    },
  };
};

export const setDefaultParam = (panelData: PanelProps<MarkdownViewerPanelOptions>) => {
  return {
    dataProvider: {
      chartQuery: [{
        query: panelData.options?.content || '',
      }],
    },
    chartDataNotExistMessage: panelData.options?.chartDataNotExistMessage,
  };
};
