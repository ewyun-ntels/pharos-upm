import React from 'react';
import { PanelPlugin, panelPluginRegistry } from '@pharos/core/panel-registry';
import iconUrl from './icon.svg';
import { Options } from './options';
import { setParamToChart, setDefaultParam as toFormData } from './setParam';
import { PieCardChart } from './pieCard';
import type { PiePanelOptions } from './types';

export const piePlugin: PanelPlugin<PiePanelOptions> = {
  info: {
    id: 'pie',
    label: 'Pie chart',
    icon: React.createElement('img', { src: iconUrl, alt: 'Pie Chart' }),
    description: 'Pie chart for displaying proportions',
    category: 'distribution',
  },
  component: PieCardChart, // 직접 전달 - 타입 힌트 가능!
  editor: async () => ({
    OptionsComponent: Options,
    toPanelData: setParamToChart,
    toFormData,
  }),
};

// 🔥 직접 등록
panelPluginRegistry.register(piePlugin);
