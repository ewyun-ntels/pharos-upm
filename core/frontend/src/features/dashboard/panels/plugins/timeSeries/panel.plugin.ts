import React from 'react';
import { PanelPlugin, panelPluginRegistry } from '@pharos/core/panel-registry';
import iconUrl from './icon.svg';
import { Options } from './options';
import { setParamToChart, setDefaultParam as toFormData } from './setParam';
import type { TimeSeriesPanelOptions } from '@pharos/shared/components/charts';

export const timeSeriesPlugin: PanelPlugin<TimeSeriesPanelOptions> = {
  info: {
    id: 'timeSeries',
    label: 'Time series',
    icon: React.createElement('img', { src: iconUrl, alt: 'Time Series' }),
    description: 'Time series line/area/bar chart',
    category: 'timeseries',
  },
  component: () =>
    import('./TimeSeriesCard').then((m) => m.TimeSeriesCardChart),
  editor: async () => ({
    OptionsComponent: Options,
    toPanelData: setParamToChart,
    toFormData,
  }),
};

// 🔥 직접 등록
panelPluginRegistry.register(timeSeriesPlugin);
