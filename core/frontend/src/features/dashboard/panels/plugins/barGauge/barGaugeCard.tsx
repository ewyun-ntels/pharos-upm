'use client';

import React from 'react';
import { withCardChartPanel } from '../common/withCardChartPanel';
import { PanelContentProps } from '../common/withChartPanel';
import { BarGaugePanelOptions } from './types';
import { UniqueMetric } from '@pharos/shared/components/charts/utils/chartType';
import { DataNotExistMessage, CalculateUnit, calculateMetrics } from '@pharos/shared/components/charts/utils/common';
import { TextWithTooltip } from '@pharos/shared/components/ui-extension';
import { formatLegendLabel } from '@pharos/shared/components/charts/utils/legendFormatter';
import { cn } from '@lib/utils';
import { UnitType } from '@pharos/shared/lib/unitUtils';
import { useTranslation } from '@/lib/data-provider';

type BarGauge = {
  barName: string;
  barValue: number | null;
  barKey: string;
};

function BarGaugeContent({ chartData, options }: PanelContentProps<BarGaugePanelOptions>) {
  const { translate: t } = useTranslation();
  const legendRule = options?.legendRule ?? 'default';
  const unit = options?.unit ?? 'default';
  const chartDataNotExistMessage = options?.chartDataNotExistMessage;

  if (chartData?.chartMetric === undefined || chartData.chartMetric?.length == 0) {
    return chartDataNotExistMessage
      ? chartDataNotExistMessage
      : DataNotExistMessage(t && t('label.chart.data_not_exists'));
  }

  if (chartData.chartType !== 'timeseries') {
    return <div>Unsupported chart data(${chartData.chartType})</div>;
  }

  let barGaugeList: BarGauge[] = [];
  let barGaugeMaxValue = 0;

  chartData.uniqueKeys?.reduce((acc: UniqueMetric, key: string) => {
    if (chartData.chartMetric) {
      acc[key] = calculateMetrics(chartData.chartMetric, key);
      const barObj = {
        barName: formatLegendLabel(key, legendRule),
        barValue: chartData.chartMetric[chartData.chartMetric?.length - 1][key]
          ? Number(chartData.chartMetric[chartData.chartMetric?.length - 1][key])
          : null,
        barKey: key,
      };
      if (Number(chartData.chartMetric[chartData.chartMetric.length - 1][key]) > barGaugeMaxValue)
        barGaugeMaxValue = Number(chartData.chartMetric[chartData.chartMetric.length - 1][key]);
      barGaugeList.push(barObj);
    }
    return acc;
  }, {} as UniqueMetric);

  return (
    <div className="grid grid-cols-[0.8fr_2fr_auto] gap-x-3 gap-y-1 items-center h-full p-4">
      {barGaugeList.map((item, index) => (
        <React.Fragment key={index}>
          <div className={cn('text-xs text-gray-700 dark:text-gray-200 truncate min-w-10')}>
            <TextWithTooltip text={item.barName} />
          </div>
          <div className="relative w-full h-full max-h-10 bg-muted rounded-[2px]">
            <div
              className="absolute left-0 top-0 h-full bg-primary rounded-[2px]"
              style={{
                width: `${(item.barValue === null ? 0 : item.barValue / barGaugeMaxValue) * 100}%`,
              }}
            />
          </div>
          <div className="text-right text-green-600 font-medium text-sm">
            <CalculateUnit value={item.barValue} unitKey={unit as UnitType} />
          </div>
        </React.Fragment>
      ))}
    </div>
  );
}

export const BarGaugeCardChart = withCardChartPanel(BarGaugeContent);
