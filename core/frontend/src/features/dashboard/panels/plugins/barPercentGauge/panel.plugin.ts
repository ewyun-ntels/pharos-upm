import React from 'react';
import { PanelPlugin, panelPluginRegistry } from '@pharos/core/panel-registry';
import iconUrl from './icon.svg';
import { Options } from './options';
import { setParamToChart, setDefaultParam as toFormData } from './setParam';
import { BarPercentGaugeCard } from './barPercentGaugeCard';
import type { BarPercentGaugePanelOptions } from './types';

export const barPercentGaugePlugin: PanelPlugin<BarPercentGaugePanelOptions> = {
  info: {
    id: 'barPercentGauge',
    label: 'Bar Percent Gauge',
    icon: React.createElement('img', { src: iconUrl, alt: 'Bar Percent Gauge' }),
    description: 'Bar percentage gauge chart',
    category: 'gauge',
  },
  component: BarPercentGaugeCard,
  editor: async () => ({
    OptionsComponent: Options,
    toPanelData: setParamToChart,
    toFormData,
  }),
};

// 🔥 직접 등록
panelPluginRegistry.register(barPercentGaugePlugin);
