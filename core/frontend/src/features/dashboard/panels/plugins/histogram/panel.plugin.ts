import React from 'react';
import { PanelPlugin, panelPluginRegistry } from '@pharos/core/panel-registry';
import iconUrl from './icon.svg';
import { Options } from './options';
import { setParamToChart, setDefaultParam as toFormData } from './setParam';
import { HistogramCard } from './histogramCard';
import type { HistogramPanelOptions } from './types';

export const histogramPlugin: PanelPlugin<HistogramPanelOptions> = {
  info: {
    id: 'histogram',
    label: 'Histogram',
    icon: React.createElement('img', { src: iconUrl, alt: 'Histogram' }),
    description: 'Histogram chart for distribution analysis',
    category: 'distribution',
  },
  component: HistogramCard,
  editor: async () => ({
    OptionsComponent: Options,
    toPanelData: setParamToChart,
    toFormData,
  }),
};

// 🔥 직접 등록
panelPluginRegistry.register(histogramPlugin);
