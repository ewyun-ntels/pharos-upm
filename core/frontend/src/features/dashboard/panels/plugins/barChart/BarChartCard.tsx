'use client';

import React from 'react';
import { withCardChartPanel } from '../common/withCardChartPanel';
import { PanelContentProps } from '../common/withChartPanel';
import { BarChart } from './BarChart';
import type { BarChartPanelOptions } from './types';

function BarChartContent({ chartData, options, chartWidth, chartHeight, setFilter, getFilter }: PanelContentProps<BarChartPanelOptions>) {
  if (chartHeight <= 40) return null;

  return (
    <BarChart
      chartData={chartData}
      chartWidth={chartWidth}
      chartHeight={chartHeight}
      options={options}
      setFilter={setFilter}
      getFilter={getFilter}
    />
  );
}

export const BarChartCard = withCardChartPanel(BarChartContent);
