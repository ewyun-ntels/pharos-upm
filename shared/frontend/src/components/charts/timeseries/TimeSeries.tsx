import uPlot, { Options as UPlotOptions } from 'uplot';
import UplotReact from 'uplot-react';
import 'uplot/dist/uPlot.min.css';
import { formatLegendLabel, PromLegend } from '../utils/legendFormatter';
import React, { useEffect, useLayoutEffect, useRef, useState, useMemo } from 'react';
import useScrollStore, {
  PromLegendTable,
  ScrollStore,
} from '../utils/tableLegendFormatter';
import { calculateMetrics, DataNotExistMessage } from '../utils/common';
import {
  LegendType,
  legendTypeBottom,
  legendTypeRight,
  UniqueMetric,
} from '../utils/chartType';
import { v4 as uuidv4 } from 'uuid';
import { getSortValue, createUniqueColorManager } from './utils';
import type { TimeSeriesChartProps, TimeSeriesPanelOptions, Threshold, ChartMetricData, ChartMetric, Annotation } from './types';
import { cn } from '../../../lib';
import { LoadingIndicator } from '../../ui-extension';
import type { LegendPayload as Payload } from 'recharts';

/**
 * Default cursor sync group key shared by all timeSeries panels on the same dashboard.
 * Set syncId to '' to disable sync, or any custom string to create a named group.
 */
export const TIMESERIES_DEFAULT_SYNC_KEY = 'pharos-timeseries-default';

// ============================================================================
// Default utility functions
// ============================================================================

const defaultFormatLocalTime = (date: Date) => {
  return date.toLocaleString('ko-KR', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  }).replace(/\. /g, '-').replace('.', '');
};

const defaultConvertUnit = (value: number, unit?: string, decimals: number = 2) => {
  if (!unit) return value.toFixed(decimals);
  return `${value.toFixed(decimals)} ${unit}`;
};

const defaultTranslate = (key: string) => key;

/**
 * x축 시간 범위(seconds)를 결정한다.
 * - chartData에 유효한 startTime/endTime이 있으면 그것을 우선 사용
 * - 없으면 chartMetric 전체 row에서 min/max timestamp를 구해 fallback
 *   (여러 쿼리 응답이 합쳐진 경우에도 정확히 커버)
 */
function resolveXRange(
  chartData: { startTime: number; endTime: number; step: number; chartMetric: { timestamp: number }[] }
): { sTime: number; eTime: number } {
  if (chartData.startTime && chartData.endTime && chartData.startTime !== chartData.endTime) {
    return {
      sTime: Math.floor(chartData.startTime / 1000),
      eTime: Math.floor(chartData.endTime / 1000),
    };
  }

  if (chartData.startTime && chartData.startTime === chartData.endTime) {
    const stepSec = Math.floor(chartData.step / 1000);
    return {
      sTime: Math.floor(chartData.startTime / 1000) - stepSec,
      eTime: Math.floor(chartData.endTime / 1000) + stepSec,
    };
  }

  // fallback: 응답된 모든 row의 timestamp(ms)에서 min/max 계산
  let minMs = Infinity;
  let maxMs = -Infinity;
  for (const row of chartData.chartMetric) {
    if (row.timestamp < minMs) minMs = row.timestamp;
    if (row.timestamp > maxMs) maxMs = row.timestamp;
  }
  let sTime = Math.floor(minMs / 1000);
  let eTime = Math.floor(maxMs / 1000);
  if (sTime === eTime) {
    const stepSec = Math.floor(chartData.step / 1000) || 60;
    sTime -= stepSec;
    eTime += stepSec;
  }
  return { sTime, eTime };
}

interface PluginOptions {
  className?: string;
  style?: React.CSSProperties;
  unitFn?: (value: number) => string;
  unit?: string;
  tooltipLimit?: number;
  formatLocalTime?: (date: Date) => string;
  onAddAnnotation?: (timeMs: number) => void;
  onDeleteAnnotation?: (id: string) => void;
  /** Stable ref object from the parent React component — survives chart recreation */
  lockedRef?: { current: boolean };
  annotations?: Annotation[];
  globalAnnotations?: Annotation[];
}

// converts the legend into a simple tooltip
export function legendAsTooltipPlugin({
  className,
  unitFn = (value: number) => value.toString(),
  tooltipLimit = 1,
  formatLocalTime = defaultFormatLocalTime,
  onAddAnnotation,
  onDeleteAnnotation,
  lockedRef,
  annotations = [],
  globalAnnotations = [],
}: PluginOptions = {}) {
  let legendEl: HTMLDivElement | null = null;
  let currentTimeMs: number | null = null;
  let annotBtn: HTMLButtonElement | null = null;
  let lockedCursorX: number | null = null; // x position (px) at the moment of locking

  function init(u: uPlot) {
    // Reset lock on every chart (re)creation to prevent DOM/ref desync
    if (lockedRef) lockedRef.current = false;

    legendEl = u.root.querySelector('.u-legend');
    if (!legendEl) return;

    legendEl.classList.remove('u-inline');
    className && legendEl.classList.add(className);

    Object.assign(legendEl.style, {
      textAlign: 'left',
      fontSize: '11px',
      pointerEvents: 'none',
      display: 'none',
      padding: '0px 15px 5px 15px',
      position: 'absolute',
      borderRadius: '5px',
      left: '0px',
      top: '0px',
      zIndex: 1000,
    });

    legendEl.classList.add(
      'shadow-lg',
      'backdrop-blur',
      'border',
      'border-gray-300',
      'dark:border-stone-700',
      'bg-white/90',
      'text-black',
      'dark:bg-neutral-900/90',
      'dark:text-white',
    );
    if (className) legendEl.classList.add(className);

    ['.u-label', '.u-marker', '.u-value'].forEach((cls) => legendEl?.querySelector(cls)?.remove());

    legendEl.querySelectorAll('.u-marker').forEach((marker, i) => {
      const series = u.series[i + 1];
      const seriesColor = series
        ? typeof series.stroke === 'function'
          ? series.stroke(u, i + 1)
          : series.stroke || series.fill
        : 'black';

      (marker as HTMLElement).style.cssText =
        `border: none; width: 6px; height: 6px; background-color: ${seriesColor}`;

      const label = legendEl?.querySelectorAll('.u-label')[i] as HTMLElement;
      if (label) {
        label.style.cssText = `color: ${seriesColor}; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 150px; display: inline-block;`;
        label.title = label.textContent || '';
      }
    });

    legendEl.querySelectorAll('.u-value').forEach((el) => {
      Object.assign((el as HTMLElement).style, {
        textAlign: 'right',
        display: 'inline-block',
        width: '100%',
        fontWeight: 'bold',
      });
    });

    u.over.appendChild(legendEl);

    u.over.addEventListener('mouseenter', () => {
      if (!lockedRef?.current) legendEl!.style.display = 'block';
    });
    u.over.addEventListener('mouseleave', () => {
      if (!lockedRef?.current) legendEl!.style.display = 'none';
    });

    if (lockedRef) {
      // Create add-annotation button only when the feature is enabled
      if (onAddAnnotation) {
        annotBtn = document.createElement('button');
        annotBtn.textContent = '+ Add Annotation';
        annotBtn.style.cssText = [
          'display:none',
          'width:calc(100% - 8px)',
          'margin:6px 4px 4px 4px',
          'padding:3px 0',
          'font-size:11px',
          'border:1px solid rgba(128,128,128,0.4)',
          'border-radius:4px',
          'cursor:pointer',
          'background:transparent',
          'color:inherit',
          'pointer-events:auto',
        ].join(';');
        annotBtn.addEventListener('click', (e) => {
          e.stopPropagation();
          if (currentTimeMs != null) onAddAnnotation(currentTimeMs);
          lockedRef.current = false;
          annotBtn!.style.display = 'none';
        });
        legendEl.appendChild(annotBtn);
      }

      // Detect real click vs drag: only toggle if mouse didn't move > 5px
      let mdX = 0, mdY = 0;
      let mdOnButton = false;
      u.over.addEventListener('mousedown', (e) => {
        mdOnButton = !!(e.target as HTMLElement).closest('button');
        mdX = e.clientX;
        mdY = e.clientY;
      });
      u.over.addEventListener('mouseup', (e) => {
        // Skip toggle when the interaction started or ended on a button
        if (mdOnButton || (e.target as HTMLElement).closest('button')) return;
        if (Math.abs(e.clientX - mdX) > 5 || Math.abs(e.clientY - mdY) > 5) return;
        lockedRef.current = !lockedRef.current;
        if (lockedRef.current) {
          // Capture cursor position at the exact moment of locking
          lockedCursorX = u.cursor.left ?? 0;
          legendEl!.style.display = 'block';
          // Show add button only if onAddAnnotation is wired up
          if (annotBtn) annotBtn.style.display = 'block';
          // Reveal delete button if an annotation is currently shown in tooltip
          const delBtn = legendEl!.querySelector('.tooltip-annot-del') as HTMLElement | null;
          if (delBtn) delBtn.style.display = 'inline-block';
        } else {
          if (annotBtn) annotBtn.style.display = 'none';
          // Hide delete button on unlock
          const delBtn = legendEl!.querySelector('.tooltip-annot-del') as HTMLElement | null;
          if (delBtn) delBtn.style.display = 'none';
          // Immediately reposition tooltip to current cursor
          update(u);
        }
      });
    }
  }

  //   uplotchart 툴팁 date 값 출력
  function update(u: uPlot) {
    if (!legendEl) return;

    if (lockedRef?.current) {
      // uPlot uses transform:translate(X,Y) via elTrans() — not style.left.
      // Our hook runs AFTER uPlot moves the cursor, so we can override it here.
      if (lockedCursorX !== null) {
        const vCursor = u.root.querySelector('.u-cursor-x') as HTMLElement | null;
        if (vCursor) {
          vCursor.style.left = ''; // clear any stale left set by old code
          vCursor.style.transform = `translate(${lockedCursorX}px, 0px)`;
        }
      }
      return;
    }

    let { left = 0, top = 0 } = u.cursor;
    if (u.width - (legendEl.offsetWidth + left + 50) < 0) left -= legendEl.offsetWidth;
    legendEl.style.transform = `translate(${left}px, ${top}px)`;

    const idx = u.cursor.idx;
    if (idx == null) return;

    const timestamp = u.data[0][idx];
    // ✅ timestamp가 밀리초면 그대로, 초면 1000 곱하기
    const timestampMs = timestamp > 10000000000 ? timestamp : timestamp * 1000;
    currentTimeMs = timestampMs;
    const date = new Date(timestampMs);
    const formatted = formatLocalTime(date);

    let timeEl = legendEl.querySelector('.tooltip-time') as HTMLDivElement;
    if (!timeEl) {
      timeEl = document.createElement('div');
      timeEl.className = 'tooltip-time';
      timeEl.style.cssText = 'font-weight:bold;padding-top:8px;margin-bottom:-4px;font-size:12px;';
      legendEl.prepend(timeEl);
    }
    timeEl.textContent = formatted;

    const mouseY = u.cursor.top ?? 0;
    const visibleSeries: Array<{ index: number; distance: number; rawValue: number }> = [];

    // 모든 시리즈의 거리 계산
    legendEl.querySelectorAll('.u-value').forEach((el, i) => {
      const currentIndex = i + 1;
      const series = u.series[currentIndex];
      if (!series?.show) return;

      const rawValue = u.data[currentIndex]?.[idx];
      if (rawValue == null || isNaN(rawValue)) return;

      const yPos = u.valToPos(rawValue, 'y');
      const distance = Math.abs(mouseY - yPos);

      visibleSeries.push({ index: currentIndex, distance, rawValue });
    });

    const shouldShowOnlyClosest = visibleSeries.length > tooltipLimit;

    let indicesToShow: Set<number>;
    if (shouldShowOnlyClosest) {
      visibleSeries.sort((a, b) => a.distance - b.distance);
      indicesToShow = new Set([visibleSeries[0]?.index].filter(Boolean));
    } else {
      indicesToShow = new Set(visibleSeries.map((s) => s.index));
    }

    // 표시/숨김 처리
    legendEl.querySelectorAll('.u-value').forEach((el, i) => {
      const currentIndex = i + 1;
      const shouldShow = indicesToShow.has(currentIndex);
      const seriesRow = el.closest('tr, .u-series') as HTMLElement;

      if (shouldShow) {
        const rawValue = u.data[currentIndex]?.[idx];
        el.textContent = typeof rawValue === 'number' ? unitFn(rawValue) : '';
        seriesRow.style.display = '';
      } else {
        seriesRow.style.display = 'none';
      }
    });

    // Annotation proximity: show annotation text in tooltip when cursor is near a marker
    const allAnnots = [
      ...globalAnnotations.map(a => ({ a, isGlobal: true })),
      ...annotations.map(a => ({ a, isGlobal: false })),
    ];
    const cursorX = u.cursor.left ?? 0;
    const nearAnnot = allAnnots.find(({ a }) =>
      Math.abs(u.valToPos(a.time / 1000, 'x') - cursorX) < 10
    );
    let annotEl = legendEl.querySelector('.tooltip-annot') as HTMLDivElement | null;
    if (nearAnnot) {
      if (!annotEl) {
        annotEl = document.createElement('div');
        annotEl.className = 'tooltip-annot';
        annotEl.style.cssText = [
          'padding:4px 4px 2px 4px',
          'font-size:13px',
          'font-weight:bold',
          'border-top:1px solid rgba(128,128,128,0.3)',
          'margin-top:4px',
          'white-space:nowrap',
          'display:flex',
          'align-items:center',
          'justify-content:space-between',
          'gap:6px',
        ].join(';');
        const insertTarget = annotBtn;
        if (insertTarget && legendEl.contains(insertTarget)) {
          legendEl.insertBefore(annotEl, insertTarget);
        } else {
          legendEl.appendChild(annotEl);
        }
      }
      annotEl.style.color = nearAnnot.a.color ?? '#ef4444';

      // Update or create the text span
      let textSpan = annotEl.querySelector('.tooltip-annot-text') as HTMLSpanElement | null;
      if (!textSpan) {
        textSpan = document.createElement('span');
        textSpan.className = 'tooltip-annot-text';
        annotEl.appendChild(textSpan);
      }
      textSpan.textContent = nearAnnot.a.text;

      // Update or create the delete button
      if (onDeleteAnnotation) {
        let delBtn = annotEl.querySelector('.tooltip-annot-del') as HTMLButtonElement | null;
        if (!delBtn) {
          delBtn = document.createElement('button');
          delBtn.className = 'tooltip-annot-del';
          delBtn.textContent = '✕';
          delBtn.style.cssText = [
            'display:none',
            'pointer-events:auto',
            'cursor:pointer',
            'background:transparent',
            'border:1px solid rgba(128,128,128,0.4)',
            'border-radius:3px',
            'font-size:10px',
            'line-height:1',
            'padding:1px 4px',
            'color:inherit',
            'flex-shrink:0',
          ].join(';');
          annotEl.appendChild(delBtn);
        }
        // Re-bind click so it always captures the current nearAnnot id
        delBtn.onclick = (e) => {
          e.stopPropagation();
          onDeleteAnnotation(nearAnnot.a.id);
        };
      }
    } else {
      annotEl?.remove();
    }
  }

  return { hooks: { init, setCursor: update } };
}

// ============================================================================
// Chart Data Builder
// ============================================================================

interface BuildChartDataResult {
  chartSeries: uPlot.Series[];
  payload: Payload[];
  metrics: UniqueMetric;
  statChartData: uPlot.AlignedData;
}

/**
 * 정렬된 키 목록을 기반으로 uPlot series 정의, legend payload, metric 통계,
 * 그리고 uPlot 형식 데이터(AlignedData)를 한 번에 생성한다.
 */
function buildChartData(
  sortedUniqueKeys: string[] | undefined,
  chartOptions: TimeSeriesPanelOptions,
  chartData: ChartMetricData,
  selectedLegend: string[] | undefined,
): BuildChartDataResult {
  const getUniqueColor = createUniqueColorManager();
  let chartSeries: uPlot.Series[] = [{}];
  const payload: Payload[] = [];

  const metrics = sortedUniqueKeys?.reduce((acc: UniqueMetric, key: string) => {
    if (chartData.chartMetric) acc[key] = calculateMetrics(chartData.chartMetric, key);

    const legendLabel = formatLegendLabel(key, chartOptions.legendRule);
    const color = getUniqueColor(legendLabel);

    if (selectedLegend === undefined || selectedLegend.includes(key)) {
      const isBars = chartOptions.chartType === 'bars';
      const isPoints = chartOptions.chartType === 'points';
      const isArea = chartOptions.chartType === 'area';

      chartSeries.push({
        label: legendLabel,
        stroke: isPoints ? undefined : color,
        width: (!isPoints && !isBars) ? (chartOptions.lineWidth ?? 1) : undefined,
        points: {
          show: isPoints || (chartOptions.chartType === 'line' && chartOptions.showPoints !== false),
          fill: color,
          size: isPoints ? (chartOptions.pointSize ?? 8) : 4,
        },
        paths: isBars
          ? (uPlot.paths.bars ? uPlot.paths.bars({ size: [chartOptions.barWidth ?? 0.6, 100] }) : undefined)
          : undefined,
        fill: (isArea || isBars) ? color.replace(/,\s*[\d.]+\)$/, `, ${chartOptions.fillOpacity ?? (isArea ? 0.1 : 1)})`) : undefined,
      });
    }

    payload.push({ value: legendLabel, type: 'square' as const, color });
    return acc;
  }, {} as UniqueMetric) ?? {};

  // chartMetric → uPlot AlignedData (ms → seconds 변환)
  // timestamp 열은 number, 데이터 열은 number|null (string 값은 uPlot에서 사용되지 않음)
  const cData: (number | null)[][] = [];
  ['timestamp'].concat(sortedUniqueKeys || []).forEach((key) => {
    if (key === 'timestamp' || selectedLegend === undefined || selectedLegend.includes(key)) {
      const arrData = chartData.chartMetric?.map((row: ChartMetric) =>
        key === 'timestamp'
          ? Math.floor((row[key] as number) / 1000)
          : typeof row[key] === 'number' ? (row[key] as number) : null,
      );
      if (arrData) cData.push(arrData);
    }
  });

  let statChartData: uPlot.AlignedData = cData as uPlot.AlignedData;

  // Bar stacking: 누적 합산 후 역순으로 그려 배경부터 전경 순으로 표시
  if (chartOptions.chartType === 'bars' && chartOptions.barStack && cData.length > 2) {
    const timestamps = cData[0] as number[];
    const seriesArrays = cData.slice(1) as (number | null)[][];
    const cumSeriesArrays = seriesArrays.map((_, s) =>
      timestamps.map((__, i) => {
        let sum = 0;
        for (let k = 0; k <= s; k++) {
          const v = seriesArrays[k][i];
          if (v != null) sum += v;
        }
        return sum;
      }),
    );
    statChartData = [cData[0], ...cumSeriesArrays.reverse()] as uPlot.AlignedData;
    chartSeries = [chartSeries[0], ...chartSeries.slice(1).reverse()];
  }

  return { chartSeries, payload, metrics, statChartData };
}

// ============================================================================
// Chart Dimensions Resolver
// ============================================================================

interface ChartDimensions {
  currentChartWidth: number;
  currentChartHeight: number;
  legendHeight: number;
}

/** legend 위치/크기 설정을 기반으로 실제 차트 및 legend 영역 크기를 계산한다. */
function resolveChartDimensions(
  chartOptions: TimeSeriesPanelOptions,
  chartWidth: number | undefined,
  chartHeight: number | undefined,
  payloadLength: number,
  legendWidthState: number,
): ChartDimensions {
  if (chartOptions.legendEnabled && (chartHeight ?? 0) > 159) {
    if (chartOptions.legendAlign === 'bottom') {
      if (payloadLength === 0) {
        return { currentChartWidth: chartWidth ?? 150, currentChartHeight: chartHeight ?? 98, legendHeight: 0 };
      }
      // legendAsTable일 때 payloadLength에 따른 높이 (높이가 많을수록 legend 영역 필요)
      const tableLegendHeights: Record<number, number> = { 1: 42, 2: 71 };
      const defaultTableHeight = 98;
      
      const legendH = chartOptions.legendAsTable 
        ? (tableLegendHeights[payloadLength] ?? defaultTableHeight)
        : 46;
      
      return { 
        currentChartWidth: chartWidth ?? 50, 
        currentChartHeight: (chartHeight ?? 50) - legendH, 
        legendHeight: legendH 
      };
    }
    if (chartOptions.legendAlign === 'right') {
      const h = chartHeight ?? 90;
      return {
        currentChartWidth: (chartWidth ?? 150) - legendWidthState,
        currentChartHeight: h,
        legendHeight: h,
      };
    }
  }
  return { currentChartWidth: chartWidth ?? 150, currentChartHeight: chartHeight ?? 98, legendHeight: 0 };
}

// ============================================================================
// uPlot Options Builder
// ============================================================================

interface BuildUPlotOptionsParams {
  width: number;
  height: number;
  sTime: number;
  eTime: number;
  yLabelWidth: number;
  setYLabelWidth: React.Dispatch<React.SetStateAction<number>>;
  chartSeries: uPlot.Series[];
  chartOptions: TimeSeriesPanelOptions;
  statChartData: uPlot.AlignedData;
  onTimeRangeChange?: (start: number, end: number) => void;
  unitFn: (value: number) => string;
  unit: string;
  formatLocalTime: (date: Date) => string;
  explicitYMin: number | undefined;
  explicitYMax: number | undefined;
  annotations: Annotation[];        // panel-scoped
  globalAnnotations: Annotation[];  // dashboard-scoped
  onAddAnnotation?: (timeMs: number) => void;
  onDeleteAnnotation?: (id: string) => void;
  lockedRef?: { current: boolean };
}

/** uPlot에 전달할 Options 객체를 생성한다. */
function buildUPlotOptions({
  width, height, sTime, eTime, yLabelWidth, setYLabelWidth,
  chartSeries, chartOptions, statChartData, onTimeRangeChange,
  unitFn, unit, formatLocalTime, explicitYMin, explicitYMax,
  annotations, globalAnnotations,
  onAddAnnotation, onDeleteAnnotation, lockedRef,
}: BuildUPlotOptionsParams): UPlotOptions {
  const yLabelDefault = 2;
  return {
    padding: [10, 24, -10, yLabelWidth],
    width,
    height,
    pxAlign: true,
    plugins: [legendAsTooltipPlugin({ unitFn, unit, tooltipLimit: chartOptions.tooltipLimit, formatLocalTime, onAddAnnotation, onDeleteAnnotation, lockedRef, annotations, globalAnnotations })],

    legend: { show: true },
    scales: {
      x: { time: true, min: sTime, max: eTime },
      y: {
        ...(explicitYMin !== undefined && { min: explicitYMin }),
        ...(explicitYMax !== undefined && { max: explicitYMax }),
      },
    },
    axes: [
      {
        show: true,
        grid: { show: false },
        font: '11px inter',
        stroke: 'gray',
        ticks: { size: 1 },
      },
      {
        show: true,
        grid: { show: true, stroke: 'border', width: 1 },
        values: (_u, vals) => {
          let strUnitLength = 1;
          return vals.map((value) => {
            const strUnit = `${unitFn ? unitFn(value) : value}`;
            if (strUnit.length > strUnitLength) {
              strUnitLength = strUnit.length;
              setYLabelWidth(strUnitLength * yLabelDefault);
            }
            return strUnit;
          });
        },
        font: '11px Inter',
        stroke: 'gray',
        ticks: { size: 1 },
      },
    ],
    series: chartSeries,
    cursor: {
      x: true,
      y: false,
      lock: false,
      sync: {
        key: chartOptions.syncId === '' ? uuidv4() : (chartOptions.syncId ?? TIMESERIES_DEFAULT_SYNC_KEY),
        setSeries: true,
      },
      points: { size: 9 },
    },
    hooks: {
      setSelect: [
        (u) => {
          if (u.select.width > 0) {
            const startIdx = u.posToIdx(u.select.left);
            const endIdx = u.posToIdx(u.select.left + u.select.width);
            const startTime = statChartData[0][startIdx];
            const endTime = statChartData[0][endIdx];
            onTimeRangeChange?.(startTime, endTime);
          }
        },
      ],
      draw: [
        (u) => {
          const ctx = u.ctx;
          const { left, top, width: bboxW, height: bboxH } = u.bbox;

          // ── Threshold lines ──────────────────────────────────────────────
          (chartOptions.thresholds || []).forEach(({ value, color, type, lineWidth }: Threshold) => {
            if (!value || !color) return;
            const yPos = u.valToPos(parseFloat(value), 'y', true);
            if (yPos < top || yPos > top + bboxH) return;
            ctx.save();
            ctx.strokeStyle = color;
            ctx.lineWidth = lineWidth ?? 2;
            ctx.setLineDash(type === 'dash' ? [5, 5] : type === 'dot' ? [2, 6] : []);
            ctx.beginPath();
            ctx.moveTo(left, yPos);
            ctx.lineTo(left + bboxW, yPos);
            ctx.stroke();
            ctx.restore();
          });

          // ── Annotation lines ─────────────────────────────────────────────
          // global: solid thick line; panel: dashed thin line
          const allAnnotations: Array<{ ann: Annotation; isGlobal: boolean }> = [
            ...globalAnnotations.map(ann => ({ ann, isGlobal: true })),
            ...annotations.map(ann => ({ ann, isGlobal: false })),
          ];
          allAnnotations.forEach(({ ann, isGlobal }) => {
            // annotation.time is ms; uPlot x-axis uses seconds
            const xPos = u.valToPos(ann.time / 1000, 'x', true);
            if (xPos < left || xPos > left + bboxW) return;
            const lineColor = ann.color ?? '#ef4444';
            ctx.save();
            ctx.beginPath();
            ctx.strokeStyle = lineColor;
            ctx.lineWidth = isGlobal ? 2 : 1;
            ctx.setLineDash(isGlobal ? [] : [4, 4]);
            ctx.moveTo(xPos, top);
            ctx.lineTo(xPos, top + bboxH);
            ctx.stroke();

            // Small downward-pointing triangle at the top — hover to see label in tooltip
            const mH = 9, mW = 6;
            ctx.fillStyle = lineColor;
            ctx.beginPath();
            ctx.moveTo(xPos - mW, top + 1);
            ctx.lineTo(xPos + mW, top + 1);
            ctx.lineTo(xPos, top + mH + 1);
            ctx.closePath();
            ctx.fill();
            ctx.restore();
          });
        },
      ],
    },
  };
}

/**
 * TimeSeriesChart - Pure Chart Component
 *
 * 순수 차트 렌더링 컴포넌트. 데이터를 외부에서 주입받아 시각화만 수행합니다.
 *
 * @param data - 차트에 표시할 시계열 데이터
 * @param loading - 로딩 상태
 * @param chartWidth - 차트 너비
 * @param chartHeight - 차트 높이
 * @param onTimeRangeChange - 시간 범위 변경 콜백
 * @param formatLocalTime - 시간 포맷팅 함수 (선택)
 * @param convertUnit - 단위 변환 함수 (선택)
 * @param translate - 번역 함수 (선택)
 * 
 * ⚡ React.memo로 최적화:
 * - data, loading, options 등이 실제로 변경될 때만 리랜더
 * - 부모의 불필요한 리랜더가 전파되지 않음
 */
export const TimeSeriesChart = React.memo(function TimeSeriesChart({
  data: chartData,
  loading = false,
  options: userOptions,
  chartWidth,
  chartHeight,
  onTimeRangeChange,
  formatLocalTime = defaultFormatLocalTime,
  convertUnit = defaultConvertUnit,
  translate = defaultTranslate,
  id: cardId,
  annotations: panelAnnotations = [],
  globalAnnotations = [],
  onChartClick,
  onDeleteAnnotation,
}: TimeSeriesChartProps) {
  const defaultChartOptions: Required<Omit<TimeSeriesPanelOptions, 'legendRule' | 'scales' | 'syncId' | 'thresholds' | 'annotations' | 'chartDataNotExistMessage' | 'showPoints' | 'lineWidth' | 'pointSize' | 'barWidth' | 'barStack' | 'decimals' | 'annotationLabelPosition'>> = {
    legendEnabled: false,
    legendAlign: 'right',
    legendAsTable: false,
    showAverage: false,
    showLast: false,
    showMax: false,
    showMin: false,
    showTotal: false,
    chartColor: 'chart_1',
    chartType: 'line',
    fillOpacity: 0.1,
    showLink: false,
    tooltipLimit: 1,
    showAnnotationButton: true,
    unit: 'short',
  };
  const chartOptions = { ...defaultChartOptions, ...userOptions }; // 기본 옵션과 사용자 옵션 병합

  const [selectedLegend, setSelectedLegend] = useState<string[] | undefined>(undefined);

  // uniqueKeys 내용이 바뀌면 selectedLegend 리셋
  // - pod=single 선택 후 pod=All로 바꾸면 새 키가 추가됨 → 이전 선택 유지 시 나머지 시리즈가 숨겨짐
  // - 같은 키 집합으로 auto-refresh된 경우(내용 변경 없음)는 리셋하지 않음
  const currentKeysStr = JSON.stringify(chartData?.uniqueKeys?.slice().sort());
  const prevKeysStrRef = React.useRef<string | undefined>(undefined);

  useEffect(() => {
    const prev = prevKeysStrRef.current;
    prevKeysStrRef.current = currentKeysStr;

    if (prev === undefined || prev === currentKeysStr) return;
    setSelectedLegend((prev) => (prev === undefined ? prev : undefined));
  }, [currentKeysStr]);

  // ✅ PanelRenderer에서 전달된 고정 id 사용 (새 UUID 생성하지 않음)
  const componentId = useMemo(() => cardId || uuidv4(), [cardId]);

  const removeScrollPosition = useScrollStore(
    (state: ScrollStore) => state.removeScrollPosition
  );

  const uPlotRef = useRef<uPlot>(null); // uPlot 인스턴스를 저장할 ref
  const uPlotRefDiv = useRef<HTMLDivElement>(null); // Chart div 참조
  const uPlotLegendRef = useRef<HTMLDivElement>(null); // legend div 참조
  // Stable ref for cursor lock state — survives chart recreation by uplot-react
  const lockedRef = useRef(false);

  const [yLabelWidth, setYLabelWidth] = useState(10);
  // ✅ legend 너비를 ResizeObserver로 추적 (render 중 ref.offsetWidth 직접 읽기의 타이밍 버그 방지)
  const [legendWidthState, setLegendWidthState] = useState(0);

  useEffect(() => {
    return () => {
      removeScrollPosition(componentId);
    };
  }, [componentId, removeScrollPosition]);

  // ✅ 데이터 유무 추적: legend div는 데이터가 있을 때만 DOM에 렌더됨
  const hasData = !!(chartData && chartData.chartMetric && chartData.chartMetric.length > 0);

  // ✅ legend 오른쪽 정렬 시 너비를 동기적으로 측정
  // useLayoutEffect: paint 이전에 실행 → 첫 렌더부터 올바른 chartWidth 계산
  // hasData: 데이터 없음→있음 전환 시 legend div가 최초 마운트되므로 재실행 필요
  useLayoutEffect(() => {
    const legendEl = uPlotLegendRef.current;
    if (!legendEl || chartOptions.legendAlign !== 'right' || !chartOptions.legendEnabled) {
      setLegendWidthState(0);
      return;
    }
    // 즉시 동기 측정 (페인트 전, legend div 마운트 직후)
    setLegendWidthState(legendEl.offsetWidth);
    const observer = new ResizeObserver(() => {
      if (uPlotLegendRef.current) {
        setLegendWidthState(uPlotLegendRef.current.offsetWidth);
      }
    });
    observer.observe(legendEl);
    return () => observer.disconnect();
    // chartOptions는 매 렌더마다 새 객체이므로 실제 값으로 의존
  }, [chartOptions.legendAlign, chartOptions.legendEnabled, hasData]);

  // ✅ 데이터는 외부에서 주입받음 (useChartData 제거)
  const unit = chartOptions.unit || 'short';
  const unitFn = (value: number) => convertUnit(value, unit, chartOptions.decimals ?? 2).toString();
  const unitFnForLegend: (value: number | undefined) => string = (value) => unitFn(value ?? 0);

  useEffect(() => {
    if (
      loading ||
      chartData?.chartMetric === undefined ||
      chartData?.chartMetric.length == 0 ||
      !uPlotRefDiv.current
    )
      return;

    // ✅ uPlot은 seconds를 기대하므로 ms → seconds 변환
    if (chartData.startTime === chartData.endTime) {
      const stepSec = Math.floor(chartData.step / 1000);
      uPlotRef.current?.setScale('x', {
        min: Math.floor(chartData.startTime / 1000) - stepSec,
        max: Math.floor(chartData.endTime / 1000) + stepSec,
      });
    } else {
      uPlotRef.current?.setScale('x', {
        min: Math.floor(chartData.startTime / 1000),
        max: Math.floor(chartData.endTime / 1000),
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loading, chartData?.chartMetric, uPlotRefDiv.current]);

  // ============================================================================
  // Early returns - 조건부 렌더링
  // ============================================================================

  if (loading) {
    return <LoadingIndicator />;
  }

  if (!chartData || chartData?.chartMetric === undefined || chartData.chartMetric?.length === 0) {
    return chartOptions.chartDataNotExistMessage
      ? chartOptions.chartDataNotExistMessage
      : DataNotExistMessage(translate('label.chart.data_not_exists'));
  }

  if (chartData?.chartType !== 'timeseries') {
    return <div>Unsupported chart data(${chartData?.chartType})</div>;
  }

  const sortedUniqueKeys = chartData.uniqueKeys?.slice().sort((a: string, b: string) => {
    if (!chartData.chartMetric?.length) return 0;
    return (
      getSortValue(b, chartData.chartMetric, chartOptions) -
      getSortValue(a, chartData.chartMetric, chartOptions)
    );
  });

  const effectiveSelectedLegend = selectedLegend && sortedUniqueKeys?.some((key) => selectedLegend.includes(key))
    ? selectedLegend.filter((key) => sortedUniqueKeys.includes(key))
    : undefined;

  const { chartSeries, payload, metrics, statChartData } = buildChartData(
    sortedUniqueKeys, chartOptions, chartData, effectiveSelectedLegend,
  );

  const selectedLegendHandler = (key: string | undefined) => {
    if (key === undefined) {
      setSelectedLegend(undefined);
    } else {
      setSelectedLegend((prevSelectedLegend) => {
        if (!prevSelectedLegend) return [key];
        if (prevSelectedLegend.includes(key)) return prevSelectedLegend;
        return [...prevSelectedLegend, key];
      });
    }
  };

  const legendOptionAttribute: LegendType = chartOptions.legendAlign === 'right' ? legendTypeRight : legendTypeBottom;
  const legendMaxWidth = chartOptions.legendAlign === 'right'
    ? 150 + [chartOptions.showAverage, chartOptions.showMax, chartOptions.showMin, chartOptions.showLast, chartOptions.showTotal].filter(Boolean).length * 50
    : undefined;

  const { sTime, eTime } = resolveXRange(chartData);

  const { currentChartWidth, currentChartHeight, legendHeight } = resolveChartDimensions(
    chartOptions, chartWidth, chartHeight, payload.length, legendWidthState,
  );

  const explicitYMin = chartOptions.scales?.yMin ? Number(chartOptions.scales.yMin) : undefined;
  const explicitYMax = chartOptions.scales?.yMax ? Number(chartOptions.scales.yMax) : undefined;

  const options = buildUPlotOptions({
    width: currentChartWidth,
    height: currentChartHeight,
    sTime,
    eTime,
    yLabelWidth,
    setYLabelWidth,
    chartSeries,
    chartOptions,
    statChartData,
    onTimeRangeChange,
    unitFn,
    unit,
    formatLocalTime,
    explicitYMin,
    explicitYMax,
    annotations: panelAnnotations,
    globalAnnotations,
    onAddAnnotation: chartOptions.showAnnotationButton ? onChartClick : undefined,
    onDeleteAnnotation,
    lockedRef,
  });

  return (
    <div className="flex flex-col w-full h-full">
    <div
      ref={uPlotRefDiv}
      className={cn('flex flex-1 min-h-0', chartOptions.legendAlign === 'bottom' ? 'flex-col' : 'flex-row')}
    >
      <div className={cn(
        chartOptions.legendAlign === 'right' && 'flex-1',
        chartOptions.legendAlign !== 'right' && 'w-full'
      )}>
        <UplotReact
          options={options}
          data={statChartData}
          onCreate={(chart) => (uPlotRef.current = chart)}
          onDelete={() => { }}
        />
      </div>
      <div
        ref={uPlotLegendRef}
        className={cn(
          chartOptions.legendAlign === 'right' && 'flex-none text-xs',
          chartOptions.legendAlign !== 'right' && 'flex-none w-full text-xs',
          chartOptions.legendAlign === 'bottom' && 'overflow-y-auto',
          !(chartOptions.legendEnabled && chartHeight && chartHeight > 160) && 'hidden',
        )}
        style={{
          ...(chartOptions.legendAlign === 'bottom' && { height: `${legendHeight}px` }),
          ...(chartOptions.legendAlign === 'right' && { maxHeight: `${legendHeight}px` }),
          ...(legendMaxWidth !== undefined && { maxWidth: `${legendMaxWidth}px` }),
        }}
      >
        {chartOptions.legendEnabled &&
          (chartOptions.legendAsTable ? (
            <PromLegendTable
              id={componentId}
              className={cn(
                legendOptionAttribute.className,
                chartOptions.legendAlign === 'right' && 'w-full justify-start',
                chartOptions.legendAlign === 'bottom' && 'w-full h-full max-h-24.75',
              )}
              payload={payload}
              metrics={metrics}
              showAverage={chartOptions.showAverage || false}
              showLast={chartOptions.showLast || false}
              showMax={chartOptions.showMax || false}
              showMin={chartOptions.showMin || false}
              showTotal={chartOptions.showTotal || false}
              unitFn={unitFnForLegend}
              selectedLegend={effectiveSelectedLegend}
              selectedLegendFn={selectedLegendHandler}
            />
          ) : (
            <PromLegend
              className={cn(
                legendOptionAttribute.className,
                chartOptions.legendAlign === 'right' && 'w-full pr-2',
                chartOptions.legendAlign === 'bottom' &&
                'justify-center w-full max-w-full leading-none flex flex-wrap max-h-10.5',
              )}
              payload={payload}
              metrics={metrics}
              showAverage={chartOptions.showAverage || false}
              showLast={chartOptions.showLast || false}
              showMax={chartOptions.showMax || false}
              showMin={chartOptions.showMin || false}
              showTotal={chartOptions.showTotal || false}
              unitFn={unitFnForLegend}
              selectedLegend={effectiveSelectedLegend}
              selectedLegendFn={selectedLegendHandler}
            />
          ))}
      </div>
    </div>
    </div>
  );
});
