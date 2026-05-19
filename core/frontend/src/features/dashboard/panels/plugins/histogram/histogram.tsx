'use client';

import {
  XAxis,
  YAxis,
  CartesianGrid,
  ReferenceArea,
  BarChart as ReBarChart,
  Bar,
} from 'recharts';
import {
  ChartConfig,
  ChartContext,
  ChartTooltip,
  ChartTooltipContent,
} from '@pharos/shared/components/ui';
import { formatLegendLabel, PromLegend } from '@pharos/shared/components/charts/utils/legendFormatter';
import { PromFormatter } from '@pharos/shared/components/charts/utils/tootlipFormatter';
import React, { useCallback, useMemo, useRef, useState } from 'react';
import { convertUnit, UnitType } from '@pharos/shared/lib/unitUtils';
import useScrollStore, { PromLegendTable } from '@pharos/shared/components/charts/utils/tableLegendFormatter';
import { cn } from '@lib/utils';
import { ChartTooltipContentOnlyLine, DataNotExistMessage } from '@pharos/shared/components/charts/utils/common';
import { UniqueMetric } from '@pharos/shared/components/charts/utils/chartType';
import { v4 as uuidv4 } from 'uuid';
import { ChartMetricData } from '@types';
import { useTableCheckStore } from '../common/tableCheckStore';
import { useTranslation } from '@/lib/data-provider';
import { HistogramPanelOptions } from './types';
import { GetColor } from '@pharos/shared/components/charts/utils/color';

const chartConfig = {} satisfies ChartConfig;

type HistogramChartProps = {
  chartData: ChartMetricData | undefined;
  options?: HistogramPanelOptions;
  chartWidth?: number;
  chartHeight?: number;
};

function HistogramChart({ chartData, options, chartWidth = 0, chartHeight = 0 }: HistogramChartProps) {
  const unit = options?.unit || 'short';
  const unitFn = (value: number | undefined) => convertUnit(value ?? 0, unit as UnitType).toString();
  const stackId = options?.stackId ?? false;
  const yaxis = options?.yaxis ?? false;
  const tickFont = options?.tickFont ?? 12;
  const fillOpacity = options?.fillOpacity ?? 0.8;
  const bucketCount = options?.bucketCount ?? 10;
  const showGrid = options?.showGrid ?? true;
  const legendRule = options?.legendRule;
  const showAverage = options?.showAverage ?? false;
  const showLast = options?.showLast ?? false;
  const showMax = options?.showMax ?? false;
  const showMin = options?.showMin ?? false;
  const showTotal = options?.showTotal ?? false;
  const legendEnabled = options?.legendEnabled ?? true;
  const legendAsTable = options?.legendAsTable ?? false;
  const legendOption = options?.legendOption ?? 'bottom';

  const { regId, set } = useTableCheckStore();
  const [selectedLegend, setSelectedLegend] = useState<string[] | undefined>(undefined);
  const { translate: t } = useTranslation();
  const [id] = useState(uuidv4());
  const [refAreaLeft, setRefAreaLeft] = useState<string | undefined>();
  const [refAreaRight, setRefAreaRight] = useState<string | undefined>();
  const removeScrollPosition = useScrollStore((state) => state.removeScrollPosition);

  const chartAreaObserverRef = useRef<ResizeObserver | null>(null);
  const [chartAreaSize, setChartAreaSize] = useState({ width: 0, height: 0 });
  const chartAreaRef = useCallback((node: HTMLDivElement | null) => {
    if (chartAreaObserverRef.current) {
      chartAreaObserverRef.current.disconnect();
      chartAreaObserverRef.current = null;
    }
    if (!node) {
      setChartAreaSize({ width: 0, height: 0 });
      return;
    }
    setChartAreaSize({ width: node.offsetWidth, height: node.offsetHeight });
    const ro = new ResizeObserver(() => {
      setChartAreaSize({ width: node.offsetWidth, height: node.offsetHeight });
    });
    ro.observe(node);
    chartAreaObserverRef.current = ro;
  }, []);

  React.useEffect(() => {
    return () => {
      removeScrollPosition(id);
    };
  }, [id, removeScrollPosition]);

  const { histogramData, uniqueKeys } = useMemo(() => {
    if (!chartData?.chartMetric || chartData.chartType !== 'timeseries') {
      return { histogramData: [], uniqueKeys: [] };
    }

    let chartRowData: Array<{ range: string; [key: string]: string | number }> = [];
    let keyArray: string[] = [];
    let valArray: number[] = [];

    for (let i = 0; i < chartData.chartMetric.length; i++) {
      const chartVal = Object.values(chartData.chartMetric[i]);
      const chartKey = Object.keys(chartData.chartMetric[i]);
      for (let s = 0; s < chartVal.length; s++) {
        if (s > 0) valArray.push(Number(chartVal[s]));
      }
      for (let s = 0; s < chartKey.length; s++) {
        if (s > 0) keyArray.push(chartKey[s]);
      }
    }

    if (valArray.length === 0) return { histogramData: [], uniqueKeys: [] };

    const minNum = Math.floor(Math.min(...valArray));
    const maxNum = Math.ceil(Math.max(...valArray));
    const intervalSize = Math.ceil((maxNum - minNum + 1) / bucketCount);
    const uniqueKeySet = [...new Set(keyArray)];

    for (let i = 0; i < bucketCount; i++) {
      const start = minNum + i * intervalSize;
      const end = Math.min(start + intervalSize - 1, maxNum);
      const chartObj: { range: string; [key: string]: string | number } = { range: Math.floor(start) + ' ~ ' + Math.floor(end) };
      for (let s = 0; s < uniqueKeySet.length; s++) {
        chartObj[uniqueKeySet[s]] = 0;
      }
      chartRowData.push(chartObj);
      if (end === maxNum) break;
    }

    chartRowData = chartRowData.filter(
      (obj, index, self) =>
        index === self.findIndex((o) => JSON.stringify(o) === JSON.stringify(obj)),
    );

    for (let i = 0; i < chartData.chartMetric.length; i++) {
      const entryKeyObj = Object.keys(chartData.chartMetric[i]);
      entryKeyObj.forEach((dataKey, index) => {
        if (index > 0) {
          for (let s = 0; s < chartRowData.length; s++) {
            const [rangeStart, rangeEnd] = chartRowData[s].range.split(' ~ ').map(Number);
            if (
              Math.ceil(Number(chartData.chartMetric?.[i][dataKey])) >= rangeStart &&
              Math.ceil(Number(chartData.chartMetric?.[i][dataKey])) <= rangeEnd
            ) {
              chartRowData[s][dataKey] = (chartRowData[s][dataKey] as number) + 1;
            }
          }
        }
      });
    }

    return {
      histogramData: chartRowData,
      uniqueKeys: chartRowData.length > 0 ? Object.keys(chartRowData[0]).filter((k) => k !== 'range') : [],
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [chartData?.chartMetric, bucketCount]);

  if (!chartData?.chartMetric || chartData.chartMetric.length === 0) {
    return DataNotExistMessage(t('label.chart.data_not_exists'));
  }

  if (chartData?.chartType !== 'timeseries') {
    return <div>Unsupported chart data(${chartData?.chartType})</div>;
  }

  if (chartWidth <= 0 || chartHeight <= 0) return null;

  // 색상 및 formatted label 계산
  const keyMeta: Record<string, { color: string; label: string }> = {};
  uniqueKeys.forEach((key, i) => {
    const label = formatLegendLabel(key, legendRule);
    keyMeta[key] = { color: GetColor(i, label), label };
  });

  const metrics = uniqueKeys.reduce((acc: UniqueMetric, key: string) => {
    const values = histogramData.map((d) => d[key]).filter((v): v is number => typeof v === 'number');
    acc[key] = {
      key: keyMeta[key].label,
      fill: keyMeta[key].color,
      min: values.length ? Math.min(...values) : undefined,
      max: values.length ? Math.max(...values) : undefined,
      average: values.length ? values.reduce((s, v) => s + v, 0) / values.length : undefined,
      last: values.length ? values[values.length - 1] : undefined,
      total: values.reduce((s, v) => s + v, 0),
    };
    return acc;
  }, {} as UniqueMetric);

  const legendPayload = uniqueKeys.map((key) => ({
    value: keyMeta[key].label,
    color: keyMeta[key].color,
    type: 'square' as const,
  }));

  const legendMaxWidth = legendOption === 'right'
    ? 150 + [showAverage, showMax, showMin, showLast, showTotal].filter(Boolean).length * 50
    : undefined;

  const selectedLegendHandler = (key: string | undefined) => {
    if (key === undefined) {
      setSelectedLegend(undefined);
    } else {
      setSelectedLegend((prev) => {
        if (!prev) return [key];
        if (prev.includes(key)) {
          const filtered = prev.filter((k) => k !== key);
          return filtered.length === 0 ? undefined : filtered;
        }
        return [...prev, key];
      });
    }
  };

  return (
    <ChartContext.Provider value={{ config: chartConfig }}>
      <div
        style={{
          display: 'flex',
          flexDirection: legendOption === 'right' ? 'row' : 'column',
          width: chartWidth,
          height: chartHeight,
          overflow: 'hidden',
        }}
      >
        <div
          ref={chartAreaRef}
          className="[&_.recharts-cartesian-axis-tick_text]:fill-muted-foreground [&_.recharts-layer]:outline-hidden [&_.recharts-surface]:outline-hidden"
          style={{ flex: '1 1 0%', minHeight: 0, minWidth: 0, overflow: 'hidden' }}
        >
        {chartAreaSize.width > 0 && chartAreaSize.height > 0 && (
        <ReBarChart
          width={chartAreaSize.width}
          height={chartAreaSize.height}
          onMouseDown={(e) => {
            setRefAreaLeft(e.activeLabel != null ? String(e.activeLabel) : undefined);
          }}
          onMouseMove={(e) => {
            id !== regId && set(id);
            refAreaLeft && setRefAreaRight(e.activeLabel != null ? String(e.activeLabel) : undefined);
          }}
          onMouseUp={() => {
            setRefAreaRight(undefined);
            setRefAreaLeft(undefined);
          }}
          data={histogramData}
          barCategoryGap={1}
          margin={{ top: 4, left: 0, right: 4, bottom: 0 }}
        >
          {showGrid && (
            <CartesianGrid
              yAxisId="1"
              vertical={false}
              strokeDasharray="3 3"
              stroke="var(--muted-foreground)"
              strokeOpacity={0.2}
              strokeWidth={0.5}
            />
          )}
          <XAxis
            tick={{ fontSize: tickFont }}
            dataKey="range"
            tickLine={false}
            axisLine={{ stroke: 'var(--border)' }}
            tickMargin={8}
            tickFormatter={(value: string) => value.split(' ~ ')[0]}
          />
          {yaxis ? (
            <YAxis
              tick={{ fontSize: tickFont }}
              yAxisId="1"
              axisLine={false}
              tickLine={false}
              width={35}
              domain={[0, (dataMax: number) => Math.ceil(dataMax) + 1]}
              allowDecimals={false}
            />
          ) : (
            <YAxis
              yAxisId={'1'}
              tickLine={false}
              axisLine={false}
              tick={false}
              width={0}
              domain={[0, (dataMax: number) => Math.ceil(dataMax) + 1]}
              allowDecimals={false}
            />
          )}
          <ChartTooltip
            cursor={true}
            wrapperClassName={'border'}
            content={
              id === regId ? (
                <ChartTooltipContent
                  formatter={(value, name, payload) => (
                    <PromFormatter
                      value={value ?? 0}
                      color={keyMeta[name?.toString() ?? '']?.color}
                      name={formatLegendLabel((name ?? '').toLocaleString(), legendRule)}
                      payload={payload}
                      unitFn={unitFn}
                    />
                  )}
                />
              ) : (
                <ChartTooltipContentOnlyLine />
              )
            }
            wrapperStyle={{ zIndex: 1000 }}
          />
          {uniqueKeys.map(
            (key) =>
              (selectedLegend === undefined || selectedLegend.includes(key)) && (
                <Bar
                  yAxisId={'1'}
                  key={key}
                  dataKey={key}
                  fill={keyMeta[key].color}
                  fillOpacity={fillOpacity}
                  stroke={keyMeta[key].color}
                  strokeWidth={1}
                  stackId={stackId ? 'stackedBarChart' : undefined}
                  isAnimationActive={false}
                  activeBar={{ fillOpacity: fillOpacity, strokeWidth: 1 }}
                />
              ),
          )}
          {refAreaLeft && refAreaRight ? (
            <ReferenceArea
              yAxisId="1"
              x1={refAreaLeft}
              x2={refAreaRight}
              strokeOpacity={0.3}
            />
          ) : null}
        </ReBarChart>
        )}
        </div>

        {legendEnabled && metrics && (
          <div
            className={cn(
              'overflow-auto text-xs flex-none',
              legendOption === 'right' ? '' : 'w-full',
            )}
            style={{
              ...(legendMaxWidth ? { maxWidth: legendMaxWidth } : {}),
              ...(legendOption === 'bottom' ? { maxHeight: Math.floor(chartHeight * 0.35) } : {}),
            }}
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
                unitFn={unitFn}
                selectedLegend={selectedLegend}
                selectedLegendFn={selectedLegendHandler}
              />
            ) : (
              <PromLegend
                className={cn(
                  legendOption === 'right' && '[&>div]:flex-col [&>div]:items-start',
                  legendOption === 'bottom' && 'justify-center flex flex-wrap',
                )}
                payload={legendPayload}
                metrics={metrics}
                showAverage={showAverage}
                showLast={showLast}
                showMax={showMax}
                showMin={showMin}
                showTotal={showTotal}
                unitFn={unitFn}
                selectedLegend={selectedLegend}
                selectedLegendFn={selectedLegendHandler}
              />
            )}
          </div>
        )}
      </div>
    </ChartContext.Provider>
  );
}

export { HistogramChart };
export type { HistogramChartProps };
