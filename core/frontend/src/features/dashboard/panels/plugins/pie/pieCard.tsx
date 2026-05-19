'use client';

import * as React from 'react';
import { useState, useRef, useCallback } from 'react';
import { Pie, PieChart } from 'recharts';
import {
  ChartConfig,
  ChartContext,
  ChartTooltip,
  ChartTooltipContent,
} from '@pharos/shared/components/ui';
import { PromLegendTable } from '@pharos/shared/components/charts/utils/tableLegendFormatter';
import { cn } from '@lib/utils';
import { formatLegendLabel, PromLegend } from '@pharos/shared/components/charts/utils/legendFormatter';
import { calculateMetrics, DataNotExistMessage } from '@pharos/shared/components/charts/utils/common';
import { UniqueMetric } from '@pharos/shared/components/charts/utils/chartType';
import { PromFormatter } from '@pharos/shared/components/charts/utils/tootlipFormatter';
import { v4 as uuidv4 } from 'uuid';
import { useTranslation } from '@/lib/data-provider';
import { GetColor } from '@pharos/shared/components/charts/utils/color';
import { convertUnit, UnitType } from '@pharos/shared/lib/unitUtils';
import { PiePanelOptions } from './types';
import { withCardChartPanel } from '../common/withCardChartPanel';
import { PanelContentProps } from '../common/withChartPanel';

const chartConfig = {} satisfies ChartConfig;

function renderLabelLine(props: { points?: { x: number; y: number }[]; stroke?: string }) {
  return (
    <path
      d={props.points && props.points.length === 2 ? `M${props.points[0].x},${props.points[0].y}L${props.points[1].x},${props.points[1].y}` : ''}
      stroke={props.stroke}
      strokeWidth={2}
      fill="none"
    />
  );
}

function PieContent({ chartData, options, chartWidth, chartHeight }: PanelContentProps<PiePanelOptions>) {
  const showType = options?.showType ?? 'last';
  const legendEnabled = options?.legendEnabled ?? false;
  const legendAlign = options?.legendAlign ?? 'right';
  const legendAsTable = options?.legendAsTable ?? false;
  const legendRule = options?.legendRule;
  const showAverage = options?.showAverage ?? false;
  const showLast = options?.showLast ?? false;
  const showMax = options?.showMax ?? false;
  const showMin = options?.showMin ?? false;
  const showTotal = options?.showTotal ?? false;
  const chartType = options?.chartType ?? 'donut';
  const showLabel = options?.showLabel ?? 'none';
  const effectiveInnerRadius = chartType === 'pie' ? '0%' : `${options?.innerRadius ?? 40}%`;
  const outerRadius = `${options?.outerRadius ?? 80}%`;
  const useStatusPalette = options?.useStatusPalette;
  const aliasColors = options?.aliasColors;
  const unit = options?.unit;
  const decimals = options?.decimals ?? 2;
  const unitFn = (value: number | undefined) => convertUnit(value ?? 0, (unit && unit !== 'none' ? unit : 'short') as UnitType, decimals).toString();

  const [id] = useState(uuidv4());
  const { translate: t } = useTranslation();

  // 파이 영역 크기 측정 (flex-1 wrapper) - right 정렬 시 legend가 늘어나면 파이가 줄어듦
  const pieAreaObserverRef = useRef<ResizeObserver | null>(null);
  const [pieAreaWidth, setPieAreaWidth] = useState(0);
  const pieAreaRef = useCallback((node: HTMLDivElement | null) => {
    if (pieAreaObserverRef.current) {
      pieAreaObserverRef.current.disconnect();
      pieAreaObserverRef.current = null;
    }
    if (!node) { setPieAreaWidth(0); return; }
    setPieAreaWidth(node.offsetWidth);
    const ro = new ResizeObserver(() => setPieAreaWidth(node.offsetWidth));
    ro.observe(node);
    pieAreaObserverRef.current = ro;
  }, []);

  // bottom 정렬 시 legend 높이 측정
  const legendObserverRef = useRef<ResizeObserver | null>(null);
  const [legendHeight, setLegendHeight] = useState(0);
  const legendRef = useCallback((node: HTMLDivElement | null) => {
    if (legendObserverRef.current) {
      legendObserverRef.current.disconnect();
      legendObserverRef.current = null;
    }
    if (!node) { setLegendHeight(0); return; }
    setLegendHeight(node.offsetHeight);
    const ro = new ResizeObserver(() => setLegendHeight(node.offsetHeight));
    ro.observe(node);
    legendObserverRef.current = ro;
  }, []);

  if (chartData?.chartMetric === undefined || chartData.chartMetric?.length == 0) {
    return options?.chartDataNotExistMessage
      ? options.chartDataNotExistMessage
      : DataNotExistMessage(t('label.chart.data_not_exists'));
  }

  if (chartData?.chartType !== 'timeseries') {
    return <div>Unsupported chart data(${chartData?.chartType})</div>;
  }

  if (chartWidth <= 0 || chartHeight <= 0) return null;

  let index = 0;
  const metrics =
    chartData.uniqueKeys?.reduce((acc: UniqueMetric, key: string) => {
      if (chartData.chartMetric) {
        const formattedLabel = formatLegendLabel(key, legendRule);
        let metricValue = calculateMetrics(chartData.chartMetric, key);
        metricValue.key = formattedLabel;
        metricValue.fill = GetColor(index++, formattedLabel, aliasColors, useStatusPalette);
        acc[key] = metricValue;
      }
      return acc;
    }, {} as UniqueMetric) || {};

  const legendPayload = Object.values(metrics).map(m => ({
    value: m.key,
    color: m.fill,
    type: 'square' as const,
  }));

  const legendMaxWidth = legendAlign === 'right'
    ? 150 + [showAverage, showMax, showMin, showLast, showTotal].filter(Boolean).length * 50
    : undefined;

  // pieAreaWidth가 아직 0이면 (첫 렌더) fallback 계산값 사용
  const pieWidth = pieAreaWidth > 0
    ? pieAreaWidth
    : Math.max(0, chartWidth - (legendEnabled && legendAlign === 'right' ? (legendMaxWidth ?? 150) : 0));
  const pieHeight = legendEnabled && legendAlign === 'bottom' ? chartHeight - legendHeight : chartHeight;

  return (
    <ChartContext.Provider value={{ config: chartConfig }}>
      <div
        style={{
          display: 'flex',
          flexDirection: legendAlign === 'right' ? 'row' : 'column',
          width: chartWidth,
          height: chartHeight,
          overflow: 'hidden',
        }}
      >
        {/* 파이 차트 영역: flex-1로 남은 공간 차지, 실제 너비를 ResizeObserver로 측정 */}
        <div
          ref={pieAreaRef}
          style={{ flex: 1, minWidth: 0, height: '100%' }}
        >
          <PieChart width={Math.max(pieWidth, 0)} height={Math.max(pieHeight, 0)}>
          <ChartTooltip
            cursor={true}
            wrapperClassName={'border'}
            isAnimationActive={false}
            content={
              <ChartTooltipContent
                formatter={(value, name, payload) => (
                  <PromFormatter
                    value={value ?? 0}
                    color={payload.payload.fill}
                    name={name?.toLocaleString() ?? ''}
                    payload={payload}
                    unitFn={unitFn}
                  />
                )}
              />
            }
            wrapperStyle={{ zIndex: 1000 }}
          />
          <Pie
            data={Object.entries(metrics).map(([_key, value]) => value)}
            dataKey={showType}
            nameKey="key"
            strokeWidth={options?.strokeWidth ?? 2}
            stroke="var(--card)"
            innerRadius={effectiveInnerRadius}
            outerRadius={outerRadius}
            isAnimationActive={false}
            label={showLabel === 'none' ? false : showLabel === 'percent'
              ? ({ percent = 0 }: { percent?: number }) => `${((percent) * 100).toFixed(0)}%`
              : ({ value }: { value?: number }) => unitFn(value)
            }
            labelLine={showLabel === 'none' ? false : renderLabelLine}
          />
        </PieChart>
        </div>

        {legendEnabled && (
          <div
            ref={legendRef}
            className={cn(
              'text-xs',
              legendAlign === 'right' ? 'flex-none h-full' : 'w-full',
            )}
            style={legendAlign === 'bottom' && legendMaxWidth ? { maxWidth: `${legendMaxWidth}px` } : undefined}
          >
            {legendAsTable ? (
              <PromLegendTable
                id={id}
                payload={legendPayload}
                metrics={metrics}
                showAverage={showAverage}
                showLast={showLast}
                showMax={showMax}
                showMin={showMin}
                showTotal={showTotal}
                isPieChart={true}
                unitFn={unitFn}
              />
            ) : (
              <PromLegend
                className={cn(legendAlign === 'right' && 'h-full [&>div]:flex-col')}
                payload={legendPayload}
                metrics={metrics}
                showAverage={showAverage}
                showLast={showLast}
                showMax={showMax}
                showMin={showMin}
                showTotal={showTotal}
                isPieChart={true}
                unitFn={unitFn}
              />
            )}
          </div>
        )}
      </div>
    </ChartContext.Provider>
  );
}

export const PieCardChart = withCardChartPanel(PieContent);
