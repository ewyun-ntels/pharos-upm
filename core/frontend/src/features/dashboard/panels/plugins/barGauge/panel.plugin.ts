import React from 'react';
import { PanelPlugin, panelPluginRegistry } from '@pharos/core/panel-registry';
import iconUrl from './icon.svg';
import { Options } from './options';
import { setParamToChart, setDefaultParam as toFormData } from './setParam';
import type { BarGaugePanelOptions } from './types';

export const barGaugePlugin: PanelPlugin<BarGaugePanelOptions> = {
  info: {
    id: 'barGauge',
    label: 'Bar gauge',
    icon: React.createElement('img', { src: iconUrl, alt: 'Bar Gauge' }),
    description: 'Bar gauge chart',
    category: 'gauge',
  },
  component: () => import('./barGaugeCard').then((m) => m.BarGaugeCardChart),
  editor: async () => ({
    OptionsComponent: Options,
    toPanelData: setParamToChart,
    toFormData,
    info: barGaugePlugin.info,
  }),
};

// 🔥 직접 등록
panelPluginRegistry.register(barGaugePlugin);
