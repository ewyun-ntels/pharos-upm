import React from 'react';
import { PanelPlugin, panelPluginRegistry } from '@pharos/core/panel-registry';
import iconUrl from './icon.svg';
import { Options } from './options';
import { setParamToChart, setDefaultParam as toFormData } from './setParam';
import type { BarChartPanelOptions } from './types';

export const barChartPlugin: PanelPlugin<BarChartPanelOptions> = {
  info: {
    id: 'barChart',
    label: 'Bar Chart',
    icon: React.createElement('img', { src: iconUrl, alt: 'Bar Chart' }),
    description: 'Bar chart visualization',
    category: 'chart',
  },
  component: () =>
    import('./BarChartCard').then((m) => m.BarChartCard),
  editor: async () => ({
    OptionsComponent: Options,
    toPanelData: setParamToChart,
    toFormData,
  }),
};

// 🔥 직접 등록
panelPluginRegistry.register(barChartPlugin);
