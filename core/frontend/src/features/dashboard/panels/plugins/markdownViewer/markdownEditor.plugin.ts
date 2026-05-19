import React from 'react';
import { PanelPlugin, panelPluginRegistry } from '@pharos/core/panel-registry';
import iconUrl from './icon.svg';
import { Options } from './options';
import { setParamToChart, setDefaultParam as toFormData } from './setParam';
import { MarkdownEditorCard } from './markdownEditor';
import type { MarkdownViewerPanelOptions } from './types';

export const markdownEditorPlugin: PanelPlugin<MarkdownViewerPanelOptions> = {
  info: {
    id: 'markdownEditor',
    label: 'Markdown Editor',
    icon: React.createElement('img', { src: iconUrl, alt: 'Markdown Editor' }),
    description: 'Markdown editor for rich text content',
    category: 'content',
  },
  component: MarkdownEditorCard,
  editor: async () => ({
    OptionsComponent: Options,
    toPanelData: setParamToChart,
    toFormData,
  }),
};

// 🔥 직접 등록
panelPluginRegistry.register(markdownEditorPlugin);
