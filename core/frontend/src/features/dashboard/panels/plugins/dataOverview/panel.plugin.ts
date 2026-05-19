import { PanelPlugin, panelPluginRegistry } from '@pharos/core/panel-registry';
import { Options } from './options';
import { setParamToChart, setDefaultParam as toFormData } from './setParam';
import { DataOverviewCard } from './dataOverviewCard';
import type { DataOverviewPanelOptions } from './types';

export const dataOverviewPlugin: PanelPlugin<DataOverviewPanelOptions> = {
  info: {
    id: 'dataOverview',
    label: 'Data Overview',
    description: 'Data overview display',
    category: 'overview',
  },
  component: DataOverviewCard,
  editor: async () => ({
    OptionsComponent: Options,
    toPanelData: setParamToChart,
    toFormData,
  }),
};

// 🔥 직접 등록
panelPluginRegistry.register(dataOverviewPlugin);
