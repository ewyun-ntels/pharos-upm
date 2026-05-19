import React from 'react';
import type { PanelProps } from '@pharos/core/panel-registry';
import { MarkdownViewerPanelOptions } from './types';
import { MarkdownViewer } from './markdownViewer';

/**
 * MarkdownViewer Card Wrapper
 * Wraps MarkdownViewer component with PanelProps interface
 */
export function MarkdownViewerCard(props: PanelProps<MarkdownViewerPanelOptions>) {
  return <MarkdownViewer {...props} />;
}
