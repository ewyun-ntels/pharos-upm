import React from 'react';
import { PanelPlugin, panelPluginRegistry } from '@pharos/core/panel-registry';
import iconUrl from './icon.svg';
import { Options } from './options';
import { setParamToChart, setDefaultParam as toFormData } from './setParam';
import { MarkdownViewerCard } from './markdownViewerCard';
import type { MarkdownViewerPanelOptions } from './types';

export const markdownViewerPlugin: PanelPlugin<MarkdownViewerPanelOptions> = {
  info: {
    id: 'markdownViewer',
    label: 'Markdown',
    icon: React.createElement('img', { src: iconUrl, alt: 'Markdown' }),
    description: 'Markdown viewer and editor',
    category: 'text',
  },
  component: MarkdownViewerCard,
  editor: async () => ({
    OptionsComponent: Options,
    toPanelData: setParamToChart,
    toFormData,
  }),
};

// 🔥 직접 등록
panelPluginRegistry.register(markdownViewerPlugin);
