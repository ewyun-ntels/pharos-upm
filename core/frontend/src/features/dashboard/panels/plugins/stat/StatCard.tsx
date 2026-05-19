'use client';

import uPlot from 'uplot';
import UplotReact from 'uplot-react';
import 'uplot/dist/uPlot.min.css';
import React, { JSX } from 'react';
import { calculateMetrics, DataNotExistMessage } from '@pharos/shared/components/charts/utils/common';
import { UniqueMetric } from '@pharos/shared/components/charts/utils/chartType';
import { useTranslation } from '@/lib/data-provider';
import type { StatPanelOptions } from './types';
import { withCardChartPanel } from '../common/withCardChartPanel';
import { PanelContentProps } from '../common/withChartPanel';
import type { ChartMetricData } from '@types';

function extractMetricName(key: string): string {
  try {
    const parsed = JSON.parse(key);
    return parsed?.metric?.columnName || key;
  } catch {
    return key;
  }
}

const chartColorPalette: { [key: string]: string } = {
  chart_1: 'hsl(258, 90%, 66%, fillOpacity)',
  chart_2: 'hsl(221, 83%, 53%, fillOpacity)',
  chart_3: 'hsl(160, 84%, 39%, fillOpacity)',
  chart_4: 'hsl(12, 76%, 61%, fillOpacity)',
  chart_5: 'hsl(27, 96%, 61%, fillOpacity)',
  chart_6: 'hsl(45, 93%, 47%, fillOpacity)',
  chart_7: 'hsl(347, 77%, 50%, fillOpacity)',
  chart_8: 'hsl(350, 80%, 72%, fillOpacity)',
  chart_9: 'hsl(351, 83%, 82%, fillOpacity)',
  chart_10: 'hsl(142, 87%, 28%, fillOpacity)',
  chart_11: 'hsl(142, 71%, 45%, fillOpacity)',
  chart_12: 'hsl(212, 95%, 68%, fillOpacity)',
  chart_13: 'hsl(210, 98%, 78%, fillOpacity)',
  chart_14: 'hsl(24, 34%, 28%, fillOpacity)',
  chart_15: 'hsl(24, 34%, 28%, fillOpacity)',
  chart_16: 'hsl(31, 41%, 48%, fillOpacity)',
  chart_17: 'hsl(25, 5%, 45%, fillOpacity)',
};

type LegendProps = {
  showType: string;
  textMode?: string;
  metricName?: string;
  metrics: UniqueMetric;
  chartData?: ChartMetricData;  // first 계산을 위해 필요
  unitFn?: (value: number) => string;
  width?: number;
  height?: number;
  textAlignment?: string;
  thresholds?: {value: number; color: string}[];
  wideLayout?: boolean;
};

const Legend = ({
  showType,
  textMode = 'auto',
  metricName = '',
  metrics,
  chartData,
  unitFn = (value: number) =>
    value.toLocaleString(undefined, {
      minimumFractionDigits: 1,
      maximumFractionDigits: 1,
    }),
  width = 240,
  height = 40,
  textAlignment = 'center',
  thresholds = [],
  wideLayout = false,
}: LegendProps): JSX.Element | null => {
  // ===== 메트릭 데이터 추출 =====
  if (!metrics) {
    return (
      <div className="w-full p-4 mb-0 font-semibold" style={{ fontSize: '20px' }}>
        N/A
      </div>
    );
  }

  const key = Object.keys(metrics)[0];
  const metric = metrics[key];

  let value;
  switch (showType) {
    case 'min': value = metric.min; break;
    case 'max': value = metric.max; break;
    case 'last': value = metric.last; break;
    case 'first':
      // calculateMetrics에 first가 없으므로 직접 계산
      const firstValue = chartData?.chartMetric?.[0]?.[key];
      value = typeof firstValue === 'number' ? firstValue : undefined;
      break;
    case 'average': value = metric.average; break;
    case 'total': value = metric.total; break;
    default: value = undefined; break;
  }

  const output = value !== undefined ? unitFn(value) : 'N/A';

  // ===== Threshold 색상 계산 =====
  let currentColor = '';
  if (value !== undefined && thresholds.length > 0) {
    const sortedThresholds = [...thresholds].sort((a, b) => b.value - a.value);
    for (const threshold of sortedThresholds) {
      if (value >= threshold.value) {
        currentColor = threshold.color;
        break;
      }
    }
  }

  // ===== 폰트 사이즈 계산 =====
  const isSingleLine = textMode === 'value' || textMode === 'name';
  const fontSizeFactor = wideLayout ? (isSingleLine ? 0.6 : 0.4) : (isSingleLine ? 0.85 : 0.60);

  // 높이 기반 폰트 크기
  const heightBasedFontSize = Math.max(14, Math.min(200, height * fontSizeFactor));

  // 너비 기반 폰트 크기 (텍스트가 패널을 벗어나지 않도록)
  let widthBasedFontSize = 200;
  if (width) {
    let text = '';
    let widthFactor = 1.0; // 너비 여유 계수

    if (textMode === 'value') {
      text = output;
      widthFactor = 0.95; // 단일 텍스트는 95% 사용
    } else if (textMode === 'name') {
      text = metricName;
      widthFactor = 0.95;
    } else if (textMode === 'auto' || textMode === 'value_and_name') {
      // auto / value_and_name 모드: 값의 폰트(fontSize)가 이름(nameFontSize)보다 크므로
      // 항상 값을 기준으로 계산해야 함
      text = output;
      // 두 줄이므로 더 작은 폰트 필요
      widthFactor = 0.50; // value only의 0.95보다 훨씬 작게
    }

    if (text) {
      const textLength = String(text).length;
      // 대략적인 문자 너비: 영문/숫자 0.6em, 한글/기타 1em
      // padding 고려 (p-4 = 16px * 2 = 32px)
      // wide layout일 경우 px-6 = 24px * 2 = 48px
      const avgCharWidth = 0.65; // 평균 문자 너비 비율
      const paddingX = wideLayout ? 48 : 32;
      const availableWidth = (width - paddingX) * widthFactor;
      widthBasedFontSize = Math.floor(availableWidth / (textLength * avgCharWidth));
    }
  }

  // 높이와 너비 기반 중 작은 값 선택
  const fontSize = Math.max(14, Math.min(heightBasedFontSize, widthBasedFontSize));
  const nameFontSize = Math.max(11, Math.min(36, fontSize * 0.35));

  // ===== 렌더링 헬퍼 함수들 =====
  //
  // 전체 케이스 맵:
  // ┌─ textMode: 'none' → 렌더링 없음
  // ├─ textMode: 'name' → renderSingleText (이름만)
  // ├─ textMode: 'value' → renderSingleText (값만, threshold 색상)
  // └─ textMode: 'auto' | 'value_and_name' (이름 + 값)
  //    ├─ wideLayout: false (세로)
  //    │  ├─ textAlignment: 'left' | 'justify' → 왼쪽 정렬
  //    │  └─ textAlignment: 'auto' | 'center' → 중앙 정렬
  //    └─ wideLayout: true (가로)
  //       ├─ textAlignment: 'auto' | 'justify' → justify-between
  //       └─ textAlignment: 'center' → justify-center

  // 1. 단일 텍스트 렌더링 (name only 또는 value only)
  const renderSingleText = (text: string, applyThresholdColor: boolean = false) => {
    const style: React.CSSProperties = { fontSize: `${fontSize}px` };
    if (applyThresholdColor && currentColor) {
      style.color = currentColor;
    }

    return (
      <div className="w-full h-full p-4 mb-0 font-semibold flex items-center justify-center" style={style}>
        {text}
      </div>
    );
  };

  // 2. 세로 레이아웃 렌더링 (이름 위, 값 아래)
  const renderVerticalLayout = () => {
    // Vertical Layout의 textAlignment:
    // - 'left' 또는 'justify' → 왼쪽 정렬
    // - 'auto' 또는 'center' → 중앙 정렬
    const isLeftAligned = textAlignment === 'left' || textAlignment === 'justify';

    // 컨테이너: flex-col로 세로 배치, justify-center로 세로 중앙, items로 가로 정렬
    const containerClasses = `w-full h-full p-4 mb-0 flex flex-col justify-center ${
      isLeftAligned ? 'items-start pl-6' : 'items-center'
    }`;

    // 텍스트 정렬
    const textClasses = `text-muted-foreground ${isLeftAligned ? 'text-left' : 'text-center'}`;
    const valueClasses = `font-semibold ${isLeftAligned ? 'text-left' : 'text-center'}`;

    const valueStyle: React.CSSProperties = { fontSize: `${fontSize}px` };
    if (currentColor) {
      valueStyle.color = currentColor;
    }

    return (
      <div className={containerClasses}>
        {metricName && (
          <div className={textClasses} style={{fontSize: `${nameFontSize}px`}}>
            {metricName}
          </div>
        )}
        <div className={valueClasses} style={valueStyle}>
          {output}
        </div>
      </div>
    );
  };

  // 3. 가로 레이아웃 렌더링 (이름 왼쪽, 값 오른쪽)
  const renderHorizontalLayout = () => {
    // textAlignment: 'auto' → justify-between (text 왼쪽, value 오른쪽)
    // textAlignment: 'center' → justify-center (가운데)
    const useSpaceBetween = textAlignment === 'auto' || textAlignment === 'justify';

    const containerClasses = `w-full h-full p-4 mb-0 flex flex-row items-center px-6 ${
      useSpaceBetween ? 'justify-between' : 'justify-center gap-4'
    }`;

    const valueStyle: React.CSSProperties = { fontSize: `${fontSize}px` };
    if (currentColor) {
      valueStyle.color = currentColor;
    }

    return (
      <div className={containerClasses}>
        {metricName && (
          <div className="text-muted-foreground" style={{fontSize: `${nameFontSize}px`}}>
            {metricName}
          </div>
        )}
        <div className="font-semibold" style={valueStyle}>
          {output}
        </div>
      </div>
    );
  };

  // ===== 옵션 트리 구조에 따른 렌더링 분기 =====

  // textMode: 'none' → 렌더링 없음
  if (textMode === 'none') {
    return null;
  }

  // textMode: 'name' → 이름만 표시
  if (textMode === 'name') {
    return renderSingleText(metricName, false);
  }

  // textMode: 'value' → 값만 표시 (threshold 색상 적용)
  if (textMode === 'value') {
    return renderSingleText(output, true);
  }

  // textMode: 'auto' | 'value_and_name' → 이름 + 값 표시
  // wideLayout에 따라 가로/세로 레이아웃 선택
  if (wideLayout) {
    return renderHorizontalLayout();
  } else {
    return renderVerticalLayout();
  }
};

function StatContent({
  chartData,
  chartWidth,
  chartHeight,
  options: userOptions,
}: PanelContentProps<StatPanelOptions>) {
  const { translate: t } = useTranslation();

  const defaultChartOptions = {
    chartColor: 'chart_1',
    chartType: 'area',
    fillOpacity: 0.3,
    showType: 'average',
    unitFn: (value: number) =>
      value.toLocaleString(undefined, {
        minimumFractionDigits: 1,
        maximumFractionDigits: 1,
      }),
  };
  const chartOptions = { ...defaultChartOptions, ...userOptions };

  if (userOptions?.decimals !== undefined) {
    const d = userOptions.decimals;
    chartOptions.unitFn = (value: number) =>
      Number(value).toLocaleString(undefined, {
        minimumFractionDigits: d,
        maximumFractionDigits: d,
      });
  }

  if (chartData === undefined || chartData.chartMetric?.length == 0) {
    return chartOptions.chartDataNotExistMessage
      ? chartOptions.chartDataNotExistMessage
      : DataNotExistMessage(t ? t('label.chart.data_not_exists') : 'No data');
  }

  if (chartData?.chartType !== 'timeseries') {
    return <div>Unsupported chart data(${chartData?.chartType})</div>;
  }

  const metrics: UniqueMetric =
    chartData.uniqueKeys?.reduce((acc: UniqueMetric, key: string) => {
      if (chartData.chartMetric) acc[key] = calculateMetrics(chartData.chartMetric, key);
      return acc;
    }, {} as UniqueMetric) || {};

  const metricsCount = Object.keys(metrics).length;
  if (metricsCount !== 1) {
    console.warn(`There should be exactly one metric, but found ${metricsCount}.`);
    return <div>There should be exactly one metric, but found ${metricsCount}.</div>;
  }

  let statChartData: uPlot.AlignedData = [];
  const uniqArray = ['timestamp'].concat(chartData.uniqueKeys || []);
  uniqArray.map((key) => {
    const arrData = chartData.chartMetric?.map((cData) => {
      return key === 'timestamp' ? cData[key] * 1000 : cData[key];
    });
    arrData && statChartData.push(Float64Array.from(arrData));
  });

  const bars = () => {
    if (!uPlot.paths?.bars) {
      console.error('uPlot.paths.bars is not available');
      return () => null;
    }
    return uPlot.paths.bars();
  };

  const cursorOpts = {
    x: false,
    y: false,
  };

  const legendHeight = Math.max(40, Math.floor((chartHeight || 90) * 0.55));

  // chartColor 처리: palette에 있으면 사용, 없으면 직접 color 값으로 간주
  const getChartColor = (colorKey: string | undefined, opacity: string | number) => {
    const key = colorKey ?? 'chart_1';
    const paletteColor = chartColorPalette[key];

    if (paletteColor) {
      // palette에 있으면 fillOpacity를 opacity 값으로 치환
      return paletteColor.replace('fillOpacity', String(opacity));
    } else {
      // palette에 없으면 직접 color 값으로 간주 (hex, hsl, rgb 등)
      // opacity 적용하기 위해 rgba/hsla로 변환
      const opacityNum = typeof opacity === 'string' ? parseFloat(opacity) : opacity;

      // hex 색상 처리 (#rgb, #rrggbb)
      if (key.startsWith('#')) {
        const hex = key.replace('#', '');
        let r: number, g: number, b: number;

        if (hex.length === 3) {
          r = parseInt(hex[0] + hex[0], 16);
          g = parseInt(hex[1] + hex[1], 16);
          b = parseInt(hex[2] + hex[2], 16);
        } else {
          r = parseInt(hex.substring(0, 2), 16);
          g = parseInt(hex.substring(2, 4), 16);
          b = parseInt(hex.substring(4, 6), 16);
        }

        return `rgba(${r}, ${g}, ${b}, ${opacityNum})`;
      }

      // hsl 색상 처리 (hsl(...))
      if (key.startsWith('hsl(')) {
        return key.replace('hsl(', 'hsla(').replace(')', `, ${opacityNum})`);
      }

      // rgb 색상 처리 (rgb(...))
      if (key.startsWith('rgb(')) {
        return key.replace('rgb(', 'rgba(').replace(')', `, ${opacityNum})`);
      }

      // rgba 색상 처리 (rgba(..., 0.x)) - 기존 alpha를 새 opacity로 교체
      if (key.startsWith('rgba(')) {
        return key.replace(/,\s*[\d.]+\s*\)$/, `, ${opacityNum})`);
      }

      // hsla 색상 처리 (hsla(..., 0.x)) - 기존 alpha를 새 opacity로 교체
      if (key.startsWith('hsla(')) {
        return key.replace(/,\s*[\d.]+\s*\)$/, `, ${opacityNum})`);
      }

      // 기타 경우 (named color 등) 그대로 반환
      return key;
    }
  };

  const uPlotOptions: uPlot.Options = {
    width: chartWidth ? chartWidth : 240,
    height: (chartHeight ? chartHeight : 90) - legendHeight,
    pxAlign: false,
    legend: { show: false },
    scales: { x: { time: false }, y: {} },
    axes: [{ show: false }, { show: false }],
    series: [
      {},
      {
        paths: chartOptions.chartType === 'bar' ? bars() : undefined,
        points: { show: chartData.chartMetric?.length === 1 },
        stroke: getChartColor(chartOptions.chartColor, 1),
        fill: getChartColor(chartOptions.chartColor, chartOptions.fillOpacity ?? 0.3),
      },
    ],
    cursor: cursorOpts,
  };

  return (
    <div className="flex flex-col h-full w-full">
      <div style={{height: chartOptions.chartType === 'none' ? '100%' : `${legendHeight}px`}}>
        <Legend
          showType={chartOptions.showType || 'last'}
          textMode={chartOptions.textMode || 'auto'}
          metricName={chartOptions.displayName || extractMetricName(chartData.uniqueKeys?.[0] || '')}
          metrics={metrics}
          chartData={chartData}
          unitFn={chartOptions.unitFn}
          height={chartOptions.chartType === 'none' ? chartHeight : legendHeight}
          width={chartWidth}
          textAlignment={chartOptions.textAlignment || 'center'}
          thresholds={chartOptions.thresholds}
          wideLayout={chartOptions.wideLayout}
        />
      </div>
      {chartOptions.chartType !== 'none' && (
        <div className="shrink-0">
          <UplotReact
            options={uPlotOptions}
            data={statChartData}
            onCreate={() => {}}
            onDelete={() => {}}
          />
        </div>
      )}
    </div>
  );
}

export const StatuPlotCardChart = withCardChartPanel(StatContent);
