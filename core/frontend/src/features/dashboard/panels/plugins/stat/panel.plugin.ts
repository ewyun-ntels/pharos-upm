import React from 'react';
import { PanelPlugin, panelPluginRegistry } from '@pharos/core/panel-registry';
import iconUrl from './icon.svg';
import { Options } from './options';
import { setParamToChart, setDefaultParam as toFormData } from './setParam';
import type { StatPanelOptions } from './types';

export const statPlugin: PanelPlugin<StatPanelOptions> = {
  info: {
    id: 'stat',
    label: 'Stat',
    icon: React.createElement('img', { src: iconUrl, alt: 'Stat' }),
    description: 'Single stat value display',
    category: 'single-stat',
  },
  component: () => import('./StatCard').then((m) => m.StatuPlotCardChart),
  editor: async () => ({
    OptionsComponent: Options,
    toPanelData: setParamToChart,
    toFormData,
  }),
};

// 🔥 직접 등록
panelPluginRegistry.register(statPlugin);
