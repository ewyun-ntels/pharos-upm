'use client';

import React from 'react';
import { cn } from '@lib/utils';
import { Separator } from '@pharos/shared/components/ui';
import { UniqueMetric } from '@pharos/shared/components/charts/utils/chartType';
import { useTranslation } from '@/lib/data-provider';
import { formatLegendLabel } from '@pharos/shared/components/charts/utils/legendFormatter';
import { UnitType } from '@pharos/shared/lib/unitUtils';
import { BarPercentGaugePanelOptions } from './types';
import { calculateMetrics, CalculateUnit, DataNotExistMessage } from '@pharos/shared/components/charts/utils/common';
import { withCardChartPanel } from '../common/withCardChartPanel';
import { PanelContentProps } from '../common/withChartPanel';

type BarPercentGauge = {
  barName?: string;
  barsubName?: string;
  barValue: number;
  barMaxValue?: number;
  barKey: string;
};

function BarPercentGaugeContent({
  chartData,
  options,
}: PanelContentProps<BarPercentGaugePanelOptions>) {
  const { translate: t } = useTranslation();
  const chartColor = options?.chartColor ?? 'var(--chart-1)';
  const showType = options?.showType;
  const legendRule = options?.legendRule;
  const subName = options?.subName;
  const gaugeMaxValue = options?.gaugeMaxValue;
  const unitType = options?.unitType;
  const chartDataNotExistMessage = options?.chartDataNotExistMessage;

  if (chartData?.chartMetric === undefined || chartData.chartMetric?.length == 0) {
    return chartDataNotExistMessage
      ? chartDataNotExistMessage
      : DataNotExistMessage(t('label.chart.data_not_exists'));
  }

  if (chartData?.chartType !== 'timeseries') {
    return <div>Unsupported chart data(${chartData?.chartType})</div>;
  }

  const metrics: UniqueMetric =
    chartData.uniqueKeys?.reduce((acc: UniqueMetric, key: string) => {
      if (chartData.chartMetric) acc[key] = calculateMetrics(chartData.chartMetric, key);
      return acc;
    }, {} as UniqueMetric) || {};

  const barPercentGaugeList: BarPercentGauge[] = Object.entries(metrics).map(([key, value]) => {
    const barValue = (() => {
      switch (showType) {
        case 'min':
          return value.min ?? 0;
        case 'max':
          return value.max ?? 0;
        case 'last':
          return value.last ?? 0;
        case 'total':
          return value.total ?? 0;
        case 'average':
        default:
          return value.average ?? 0;
      }
    })();

    return {
      barName: formatLegendLabel(key, legendRule),
      barValue,
      barsubName: subName,
      barMaxValue: gaugeMaxValue,
      barKey: `bar-${key}`,
    };
  });

  return (
    <div className="flex flex-col items-center w-full h-full gap-5">
      {barPercentGaugeList.map((item, index) => {
        const varPercentValue = item.barMaxValue
          ? parseFloat(((item.barValue / item.barMaxValue) * 100).toFixed(1))
          : 0;

        const levelColor =
          varPercentValue >= 80
            ? 'var(--destructive)'
            : varPercentValue >= 50 && varPercentValue <= 79
              ? 'var(--chart-5)'
              : chartColor;

        return (
          <React.Fragment key={index}>
            <div className="flex flex-col w-full">
              <div className="flex flex-row justify-between">
                <div className="flex flex-col gap-1">
                  <div className="flex items-end flex-row h-5 space-x-2 text-xl font-bold leading-5">
                    <CalculateUnit value={item.barValue} unitKey={unitType as UnitType} />
                    {item.barMaxValue !== undefined && item.barMaxValue !== null && (
                      <>
                        <Separator orientation="vertical" />
                        <span className="text-xs font-medium">{varPercentValue.toFixed(1)}%</span>
                      </>
                    )}
                  </div>
                  <div className={cn('text-xs text-gray-700 dark:text-gray-200 truncate min-w-10')}>
                    {item.barName}
                  </div>
                </div>
                <div className="flex flex-col gap-1 items-end justify-end">
                  <div className="flex items-center text-md font-medium">
                    {item.barMaxValue !== undefined && item.barMaxValue !== null && (
                      <CalculateUnit value={item.barMaxValue} unitKey={unitType as UnitType} />
                    )}
                  </div>
                  <div className={cn('text-xs text-right text-gray-700 dark:text-gray-200 truncate min-w-10')}>
                    {item.barsubName}
                  </div>
                </div>
              </div>
              {item.barMaxValue !== undefined && item.barMaxValue !== null && (
                <div className="relative w-full h-full max-h-12 min-h-2 rounded-[2px] bg-muted my-3">
                  <div
                    className="absolute top-0 left-0 h-full min-h-2"
                    style={{
                      width: `${varPercentValue}%`,
                      background: levelColor,
                    }}
                  />
                </div>
              )}
            </div>
          </React.Fragment>
        );
      })}
    </div>
  );
}

export const BarPercentGaugeCard = withCardChartPanel(BarPercentGaugeContent);
