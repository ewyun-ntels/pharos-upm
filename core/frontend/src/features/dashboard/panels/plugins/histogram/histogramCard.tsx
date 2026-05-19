'use client';

import React from 'react';
import { withCardChartPanel } from '../common/withCardChartPanel';
import { PanelContentProps } from '../common/withChartPanel';
import { HistogramChart } from './histogram';
import type { HistogramPanelOptions } from './types';

function HistogramContent({ chartData, options, chartWidth, chartHeight }: PanelContentProps<HistogramPanelOptions>) {
  return <HistogramChart chartData={chartData} options={options} chartWidth={chartWidth} chartHeight={chartHeight} />;
}

export const HistogramCard = withCardChartPanel(HistogramContent);
