'use client';

import React from 'react';
import type { PanelProps } from '@pharos/core/panel-registry';
import { DataOverviewPanelOptions } from './types';
import { DataOverview } from './dataOverview';

/**
 * DataOverview Card Wrapper
 * Wraps DataOverview component with PanelProps interface
 */
export function DataOverviewCard(props: PanelProps<DataOverviewPanelOptions>) {
  return <DataOverview {...props} />;
}
