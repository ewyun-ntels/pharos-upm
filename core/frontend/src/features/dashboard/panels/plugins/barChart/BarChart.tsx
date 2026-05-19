"use client";

import uPlot, { Options as UPlotOptions } from 'uplot';
import UplotReact from 'uplot-react';
import 'uplot/dist/uPlot.min.css';
import { LegendPayload as Payload } from 'recharts';
import {
    calculateMetrics,
    DataNotExistMessage,
    formatLegendLabel,
    LegendType,
    legendTypeBottom,
    legendTypeRight,
    PromLegend,
    PromLegendTable,
    UniqueMetric,
    useScrollStore,
} from '@pharos/shared/components/charts';
import React, { useEffect, useRef, useState } from 'react';
import { cn } from '@lib/utils';
import { v4 as uuidv4 } from 'uuid';
import { useTranslation } from '@/lib/data-provider';
import { convertUnit, UnitType } from '@pharos/shared/lib/unitUtils';
import { createUniqueColorManager, getSortValue } from './utils';
import type { BarChartPanelOptions } from './types';
import { ChartMetricData } from '@types';
import type { FilterValue } from '@features/dashboard/types/filter.types';

// 데이터 포인트 타입 정의
interface DataPoint {
    [key: string]: number | string | null | undefined;
}

// 상수 정의
const BAR_GROUP_WIDTH = 0.8; // 전체 그룹의 너비 비율
const BAR_GAP = 0.1; // 바 사이 간격 비율
const Y_AXIS_PADDING_RATIO = 1.1; // Y축 최대값 여유 비율 (10%)
const Y_LABEL_MULTIPLIER = 2; // Y축 라벨 너비 계산 승수

// 순수 렌더링 컴포넌트 props (data fetching 없음)
type BarChartProps = {
    chartData: ChartMetricData | undefined;
    chartWidth?: number;
    chartHeight?: number;
    options?: BarChartPanelOptions;
    // ✅ Optional: onClickVariable 사용 시에만 필요
    setFilter?: (id: string, value: FilterValue) => void;
    getFilter?: (id: string) => FilterValue | undefined;
};

interface PluginOptions {
    className?: string;
    style?: React.CSSProperties;
    unitFn?: (value: number | undefined) => string;
    unit?: string;
    tooltipLimit?: number;
    isHorizontal?: boolean;
    categoryLabels?: string[];
    barColorMode?: 'single' | 'category' | 'series';
    categoryColors?: string[];
}

// 스택된 데이터 계산 함수
function calculateStackedData(originalData: uPlot.AlignedData, barMode: string): uPlot.AlignedData {
    if (barMode !== 'stacked' && barMode !== 'percent') {
        return originalData;
    }

    const [xData, ...seriesData] = originalData;
    const stackedData: uPlot.AlignedData = [xData];

    for (let seriesIdx = 0; seriesIdx < seriesData.length; seriesIdx++) {
        const newSeriesData: (number | null)[] = [];

        for (let dataIdx = 0; dataIdx < (xData?.length || 0); dataIdx++) {
            let stackedValue = 0;

            // 현재 시리즈까지의 누적값 계산
            for (let stackIdx = 0; stackIdx <= seriesIdx; stackIdx++) {
                const value = seriesData[stackIdx]?.[dataIdx];
                if (typeof value === 'number' && !isNaN(value)) {
                    stackedValue += value;
                }
            }

            if (barMode === 'percent') {
                // 100% 스택: 해당 인덱스의 전체 합계 대비 비율로 변환
                let totalValue = 0;
                for (let stackIdx = 0; stackIdx < seriesData.length; stackIdx++) {
                    const value = seriesData[stackIdx]?.[dataIdx];
                    if (typeof value === 'number' && !isNaN(value)) {
                        totalValue += value;
                    }
                }
                newSeriesData.push(totalValue > 0 ? (stackedValue / totalValue) * 100 : 0);
            } else {
                newSeriesData.push(stackedValue);
            }
        }

        stackedData.push(newSeriesData);
    }

    return stackedData;
}

// uPlot 바 차트를 위한 경로 생성 함수
function buildBarPaths(u: uPlot, sidx: number, idx0: number, idx1: number, isHorizontal: boolean = false, barMode: string = 'single', seriesCount: number = 1) {
    const data = u.data[sidx];

    if (!data || data.length === 0) {
        return null;
    }

    let stroke = new Path2D();
    let fill = new Path2D();

    const groupWidth = BAR_GROUP_WIDTH;
    const barGap = BAR_GAP;

    // 모든 데이터를 그리도록 수정
    const startIdx = 0;
    const endIdx = data.length;

    for (let i = startIdx; i < endIdx; i++) {
        const value = data[i];
        if (value == null || typeof value !== 'number' || isNaN(value)) {
            continue; // null, undefined, NaN 값 건너뛰기
        }

        let x = i;
        let currentValue = value;

        // 바 너비 계산
        let barWidth: number;

        if (barMode === 'single' || barMode === 'stacked' || barMode === 'percent') {
            barWidth = groupWidth;
        } else { // grouped
            barWidth = (groupWidth - (seriesCount - 1) * barGap) / seriesCount;
        }

        // 바 위치 계산
        let barOffset = 0;
        if (barMode === 'grouped') {
            barOffset = (sidx - 1) * (barWidth + barGap) - (groupWidth - barWidth) / 2;
        }

        // 스택된 바의 경우 시작점 계산
        let baseValue = 0;
        if (barMode === 'stacked' || barMode === 'percent') {
            // 이전 시리즈들의 값을 기반으로 시작점 계산
            if (sidx > 1) {
                const prevSeriesValue = u.data[sidx - 1]?.[i];
                baseValue = typeof prevSeriesValue === 'number' && !isNaN(prevSeriesValue) ? prevSeriesValue : 0;
            }
        }

        if (isHorizontal) {
            // 가로 바 차트
            let x0 = u.valToPos(baseValue, "x", true);
            let x1 = u.valToPos(currentValue, "x", true);
            let y0 = u.valToPos(x - barWidth / 2 + barOffset, "y", true);
            let y1 = u.valToPos(x + barWidth / 2 + barOffset, "y", true);

            fill.rect(Math.min(x0, x1), Math.min(y0, y1), Math.abs(x1 - x0), Math.abs(y1 - y0));
        } else {
            // 세로 바 차트
            let x0 = u.valToPos(x - barWidth / 2 + barOffset, "x", true);
            let x1 = u.valToPos(x + barWidth / 2 + barOffset, "x", true);
            let y0 = u.valToPos(baseValue, "y", true);
            let y1 = u.valToPos(currentValue, "y", true);

            fill.rect(Math.min(x0, x1), Math.min(y0, y1), Math.abs(x1 - x0), Math.abs(y1 - y0));
        }
    }

    return {
        stroke,
        fill
    };
}

// 바 차트용 툴팁 플러그인
export function barTooltipPlugin({
    className,
    unitFn = (value: number | undefined) => (value ?? '').toString(),
    tooltipLimit = 1,
    categoryLabels = [],
    isHorizontal = false,
    barColorMode = 'series',
    categoryColors = [],
    barMode = 'single',
}: PluginOptions & { barMode?: string } = {}) {
    let legendEl: HTMLDivElement | null = null;

    function init(u: uPlot) {
        legendEl = u.root.querySelector(".u-legend");
        if (!legendEl) return;

        legendEl.classList.remove("u-inline");
        className && legendEl.classList.add(className);

        Object.assign(legendEl.style, {
            textAlign: "left",
            fontSize: "11px",
            pointerEvents: "none",
            display: "none",
            padding: "0px 15px 5px 15px",
            position: "absolute",
            borderRadius: "5px",
            left: "0px",
            top: "0px",
            zIndex: 1000,
        });

        legendEl.classList.add(
            "shadow-lg",
            "backdrop-blur",
            "border",
            "border-gray-300",
            "dark:border-stone-700",
            "bg-white/90",
            "text-black",
            "dark:bg-neutral-900/90",
            "dark:text-white"
        );
        if (className) legendEl.classList.add(className);

        [".u-label", ".u-marker", ".u-value"].forEach((cls) => legendEl?.querySelector(cls)?.remove());

        legendEl.querySelectorAll(".u-marker").forEach((marker, i) => {
            const series = u.series[i + 1];
            const seriesColor = series
                ? typeof series.stroke === "function"
                    ? series.stroke(u, i + 1)
                    : series.stroke || series.fill
                : "black";

            (marker as HTMLElement).style.cssText = `border: none; width: 6px; height: 6px; background-color: ${seriesColor}`;

            const label = legendEl?.querySelectorAll(".u-label")[i] as HTMLElement;
            if (label) {
                label.style.cssText = `color: ${seriesColor}; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 150px; display: inline-block;`;
                label.title = label.textContent || "";
            }
        });

        legendEl.querySelectorAll(".u-value").forEach((el) => {
            Object.assign((el as HTMLElement).style, {
                textAlign: "right",
                display: "inline-block",
                width: "100%",
                fontWeight: "bold",
            });
        });

        u.over.appendChild(legendEl);

        ["mouseenter", "mouseleave"].forEach((evt) =>
            u.over.addEventListener(
                evt,
                () => (legendEl!.style.display = evt === "mouseenter" ? "block" : "none"),
            ),
        );
    }

    //   바차트 툴팁 업데이트
    function update(u: uPlot) {
        if (!legendEl) return;

        let { left = 0, top = 0 } = u.cursor;

        let idx = u.cursor.idx;

        if (isHorizontal) {
            // Horizontal 바 차트의 경우, Y축 좌표로부터 인덱스 계산
            if (top === undefined || top < 0) return;

            // Y축 픽셀 좌표를 데이터 인덱스로 변환
            const yValue = u.posToVal(top, "y");
            idx = Math.round(yValue);

            // 유효한 범위 체크
            if (idx < 0 || idx >= u.data[0].length) return;
        } else {
            // Vertical 차트의 경우, cursor.idx 사용
            if (idx == null) {
                return;
            }
        }

        const visibleSeries: Array<{ index: number, distance: number, rawValue: number }> = [];

        // 모든 시리즈의 거리 계산 (바 차트용)
        legendEl.querySelectorAll(".u-value").forEach((el, i) => {
            const currentIndex = i + 1;
            const series = u.series[currentIndex];
            if (!series?.show) return;

            const rawValue = u.data[currentIndex]?.[idx];
            if (rawValue == null || isNaN(rawValue)) return;

            // 바 차트에서는 해당 인덱스의 모든 값을 표시
            visibleSeries.push({ index: currentIndex, distance: 0, rawValue });
        });

        // 표시할 값이 없으면 툴팁 숨기기
        if (visibleSeries.length === 0) {
            legendEl.style.display = 'none';
            return;
        }

        // 값이 있으면 툴팁 표시
        legendEl.style.display = 'block';

        if (u.width - (legendEl.offsetWidth + left + 50) < 0)
            left -= legendEl.offsetWidth;
        legendEl.style.transform = `translate(${left}px, ${top}px)`;

        // 카테고리 라벨 표시 (X축 값)
        const categoryLabel = categoryLabels[idx] || `Category ${idx + 1}`;

        let timeEl = legendEl.querySelector(".tooltip-time") as HTMLDivElement;
        if (!timeEl) {
            timeEl = document.createElement("div");
            timeEl.className = "tooltip-time";
            timeEl.style.cssText = "font-weight:bold;padding-top:8px;margin-bottom:-4px;font-size:12px;";
            legendEl.prepend(timeEl);
        }
        timeEl.textContent = categoryLabel;

        // Single/Grouped 모드에서는 tooltipLimit 무시하고 모든 시리즈 표시
        // Stacked/Percent 모드에서만 tooltipLimit 적용
        const shouldShowOnlyClosest = (barMode === 'stacked' || barMode === 'percent') && visibleSeries.length > tooltipLimit;

        let indicesToShow: Set<number>;
        if (shouldShowOnlyClosest) {
            // 값이 큰 순서로 정렬해서 상위 몇 개만 표시
            visibleSeries.sort((a, b) => Math.abs(b.rawValue) - Math.abs(a.rawValue));
            indicesToShow = new Set(visibleSeries.slice(0, tooltipLimit).map(s => s.index));
        } else {
            indicesToShow = new Set(visibleSeries.map(s => s.index));
        }

        // 표시/숨김 처리 및 카테고리 모드일 때 색상 업데이트
        legendEl.querySelectorAll(".u-value").forEach((el, i) => {
            const currentIndex = i + 1;
            const shouldShow = indicesToShow.has(currentIndex);
            const seriesRow = el.closest('tr, .u-series') as HTMLElement;

            if (shouldShow) {
                const rawValue = u.data[currentIndex][idx];
                el.textContent = unitFn(rawValue ?? undefined);
                seriesRow.style.display = '';

                // 카테고리 모드일 때 마커와 라벨 색상을 카테고리 색상으로 변경
                if (barColorMode === 'category' && categoryColors.length > idx) {
                    const categoryColor = categoryColors[idx];
                    const marker = seriesRow.querySelector('.u-marker') as HTMLElement;
                    const label = seriesRow.querySelector('.u-label') as HTMLElement;

                    if (marker) {
                        marker.style.backgroundColor = categoryColor;
                    }
                    if (label) {
                        label.style.color = categoryColor;
                    }
                }
            } else {
                seriesRow.style.display = 'none';
            }
        });
    }

    return { hooks: { init, setCursor: update } };
}

function BarChart({
    chartData,
    chartWidth,
    chartHeight,
    options: userOptions,
    setFilter,
    getFilter,
}: BarChartProps) {
    const defaultChartOptions = {
        legendEnabled: false,
        legendAlign: "right" as const,
        legendAsTable: false,
        barMode: "single" as const,
        orientation: "vertical" as const,
        showValueLabels: false,
        unit: "short",
        tooltipLimit: 1,
    };
    const chartOptions = { ...defaultChartOptions, ...userOptions }; // 기본 옵션과 사용자 옵션 병합

    const [selectedLegend, setSelectedLegend] = useState<string[] | undefined>(undefined);

    const [id] = useState(uuidv4());
    const { translate: t } = useTranslation();

    const removeScrollPosition = useScrollStore((state) => state.removeScrollPosition);

    const uPlotRef = useRef<uPlot>(null); // uPlot 인스턴스를 저장할 ref
    const uPlotRefDiv = useRef<HTMLDivElement>(null); // Chart div 참조
    const uPlotLegendRef = useRef<HTMLDivElement>(null); // legend div 참조
    const [yLabelWidth, setYLabelWidth] = useState(10);

    useEffect(() => {
        return () => {
            removeScrollPosition(id);
        };
    }, [id, removeScrollPosition]);

    const unit = chartOptions.unit || '';
    const unitFn = (value: number | undefined) => convertUnit(value ?? 0, unit as UnitType).toString();

    // Determine orientation: auto defaults to vertical
    const orientation = chartOptions.orientation || 'auto';
    const isHorizontal = orientation === 'horizontal';

    // Determine X axis column - 빈 문자열이면 자동으로 첫 번째 컬럼 사용
    let xAxisColumn = chartOptions.xAxis || '';

    // chartData 구조 파싱
    const chartMetric = chartData?.chartMetric;

    // uniqueKeys 추출
    let uniqueKeys = chartData?.uniqueKeys;
    if (!uniqueKeys && chartMetric && chartMetric.length > 0) {
        uniqueKeys = Object.keys(chartMetric[0]).filter(key => key !== 'timestamp');
    }

    // X Axis 컬럼 자동 결정
    if (!xAxisColumn && uniqueKeys && uniqueKeys.length > 0) {
        // X Axis가 지정되지 않은 경우, 첫 번째 컬럼을 X Axis로 사용
        xAxisColumn = uniqueKeys[0];
    }

    useEffect(() => {
        if (!chartMetric || chartMetric.length === 0 || !uPlotRefDiv.current)
            return;

        // 바 차트에서는 카테고리 인덱스 기반 스케일 설정
        const dataLength = chartMetric.length;
        if (isHorizontal) {
            // 가로 바: Y축이 카테고리 (데이터 개수만큼)
            uPlotRef.current?.setScale("y", { min: -0.5, max: dataLength - 0.5 });
        } else {
            // 세로 바: X축이 카테고리 (데이터 개수만큼)
            uPlotRef.current?.setScale("x", { min: -0.5, max: dataLength - 0.5 });
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [chartMetric, uPlotRefDiv.current, isHorizontal]);

    // Variable 값 변경 시 차트 다시 그리기
    useEffect(() => {
        if (!chartOptions.onClickVariable || !uPlotRef.current) return;

        // Variable 값이 변경되면 차트 다시 그리기
        uPlotRef.current.redraw();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [chartOptions.onClickVariable && getFilter ? getFilter(chartOptions.onClickVariable) : null]);

    if (!chartMetric || chartMetric.length === 0) {
        return chartOptions.chartDataNotExistMessage
            ? chartOptions.chartDataNotExistMessage
            : DataNotExistMessage(t("label.chart.data_not_exists"));
    }

    // 바 차트는 timeseries와 categorical 데이터 모두 지원

    // sortedUniqueKeys가 없거나 비어있는 경우 처리
    let sortedUniqueKeys = uniqueKeys?.slice() || [];

    // X Axis로 지정된 필드는 제외 (카테고리이므로 Y 값으로 사용하지 않음)
    sortedUniqueKeys = sortedUniqueKeys.filter((key: string) => {
        return !(xAxisColumn && key === xAxisColumn);

    });

    // 가로 바 차트이고 valueSortOrder 옵션이 설정된 경우 chartMetric 데이터 정렬
    let sortedChartMetric = chartMetric ? [...chartMetric] : [];
    if (isHorizontal && chartOptions.valueSortOrder && chartOptions.valueSortOrder !== 'none' && sortedUniqueKeys.length > 0) {
        // 첫 번째 시리즈 키를 기준으로 정렬
        const sortKey = sortedUniqueKeys[0];
        sortedChartMetric.sort((a: DataPoint, b: DataPoint) => {
            const aVal = typeof a[sortKey] === 'number' ? a[sortKey] as number : 0;
            const bVal = typeof b[sortKey] === 'number' ? b[sortKey] as number : 0;

            if (chartOptions.valueSortOrder === 'asc') {
                return aVal - bVal; // 오름차순
            } else {
                return bVal - aVal; // 내림차순
            }
        });
    }

    // 정렬 (옵션에 따라)
    if (sortedUniqueKeys.length > 0 && sortedChartMetric?.length > 0) {
        sortedUniqueKeys.sort((a: string, b: string) => {
            return getSortValue(b, sortedChartMetric, chartOptions) - getSortValue(a, sortedChartMetric, chartOptions);
        });
    }

    // 색상 모드 결정 (기본값: series)
    const barColorMode = chartOptions.barColorMode || 'series';
    const singleBarColor = chartOptions.barColor || 'hsl(221, 83%, 53%)'; // 기본 파란색

    const getUniqueColor = createUniqueColorManager();

    let chartSeries: uPlot.Series[] = [{ label: chartOptions.barMode }]; // barMode를 첫 번째 시리즈에 저장
    let payload: Payload[] = [];

    const metrics = sortedUniqueKeys?.reduce((acc: UniqueMetric, key: string) => {
        if (sortedChartMetric) acc[key] = calculateMetrics(sortedChartMetric, key);

        const legendLabel = formatLegendLabel(key, chartOptions.legendRule);
        const color = barColorMode === 'single' ? singleBarColor : getUniqueColor(legendLabel);

        if (selectedLegend === undefined || selectedLegend.includes(key)) {
            chartSeries.push({
                label: legendLabel,
                stroke: barColorMode === 'category' ? 'transparent' : color,
                fill: barColorMode === 'category' ? 'transparent' : color,
                width: 1,
                paths: (u: uPlot, sidx: number, idx0: number, idx1: number) => {
                    if (barColorMode === 'category') {
                        // 카테고리 모드: draw hook에서 그림
                        return {
                            stroke: new Path2D(),
                            fill: new Path2D()
                        };
                    }
                    // single 또는 series 모드: 기본 buildBarPaths 사용
                    return buildBarPaths(u, sidx, idx0, idx1,
                        chartOptions.orientation === 'horizontal',
                        chartOptions.barMode,
                        sortedUniqueKeys?.length || 1
                    );
                },
                points: { show: false },
            });
        }

        // payload에도 동일한 색상 추가 (범례에 표시)
        payload.push({
            value: legendLabel,
            type: "square" as const,
            color: color,
        });

        return acc;
    }, {} as UniqueMetric);

    const selectedLegendHandler = (key: string | undefined) => {
        if (key === undefined) {
            setSelectedLegend(undefined);
        } else {
            setSelectedLegend((prevSelectedLegend) => {
                if (!prevSelectedLegend) {
                    return [key];
                }
                if (prevSelectedLegend.includes(key)) {
                    return prevSelectedLegend;
                }
                return [...prevSelectedLegend, key];
            });
        }
    };

    let legendOptionAttribute: LegendType;
    if (chartOptions.legendAlign === "right") {
        legendOptionAttribute = legendTypeRight;
    } else {
        legendOptionAttribute = legendTypeBottom;
    }

    // 바 차트를 위한 데이터 변환 
    // 카테고리 데이터를 인덱스 기반으로 변환
    let statChartData: uPlot.AlignedData;
    let cData: (number | null)[][] = [];
    let categoryLabels: string[] = [];

    // 첫 번째 배열은 인덱스 (0, 1, 2, ...)
    const indexData = sortedChartMetric?.map((_: DataPoint, index: number) => index) || [];
    cData.push(indexData);

    // 카테고리 라벨 생성 (선택된 X axis 컬럼 사용)
    categoryLabels = sortedChartMetric?.map((dataPoint: DataPoint, index: number) => {
        // 1. 사용자가 지정한 X Axis 컬럼이 있고, 해당 값이 존재하는 경우
        if (xAxisColumn && dataPoint[xAxisColumn] !== undefined && dataPoint[xAxisColumn] !== null) {
            return String(dataPoint[xAxisColumn]);
        }

        // 2. Fallback: 인덱스 사용
        return `Category ${index + 1}`;
    }) || [];

    // 카테고리별 색상 및 메트릭 (barColorMode가 'category'일 때만 생성)
    let categoryColors: string[] = [];
    let categoryMetrics: UniqueMetric = {};

    if (barColorMode === 'category') {
        // 카테고리별 색상 생성
        const categoryColorManager = createUniqueColorManager();
        categoryColors = categoryLabels.map(label => categoryColorManager(label));

        // 범례를 카테고리별로 생성
        payload = categoryLabels.map((label, idx) => ({
            value: label,
            type: "square" as const,
            color: categoryColors[idx],
        }));

        // 카테고리별 메트릭 계산 (각 카테고리의 모든 시리즈 값 합산)
        categoryMetrics = categoryLabels.reduce((acc: UniqueMetric, label: string, idx: number) => {
            const values: number[] = [];

            sortedUniqueKeys?.forEach((key: string) => {
                const dataPoint = sortedChartMetric?.[idx];
                const value = dataPoint?.[key];
                if (typeof value === 'number' && !isNaN(value)) {
                    values.push(value);
                }
            });

            acc[label] = {
                key: label,
                min: values.length > 0 ? Math.min(...values) : undefined,
                max: values.length > 0 ? Math.max(...values) : undefined,
                average: values.length > 0 ? values.reduce((a, b) => a + b, 0) / values.length : undefined,
                last: values.length > 0 ? values[values.length - 1] : undefined,
                total: values.length > 0 ? values.reduce((a, b) => a + b, 0) : undefined,
            };

            return acc;
        }, {} as UniqueMetric);
    }

    // 각 시리즈 데이터 추가
    (sortedUniqueKeys || []).forEach((key: string) => {
        if (selectedLegend === undefined || selectedLegend.includes(key)) {
            const arrData = sortedChartMetric?.map((dataPoint: DataPoint) => {
                const value = dataPoint[key];
                return typeof value === 'number' ? value : null;
            });
            arrData && cData.push(arrData);
        }
    });

    // 바 모드에 따라 데이터 변환
    statChartData = calculateStackedData(cData as uPlot.AlignedData, chartOptions.barMode);

    // 바 차트에서는 인덱스 기반 스케일 사용
    const dataLength = sortedChartMetric?.length || 0;

    // Y축 값의 최대값 계산 (auto scale을 위해)
    let maxYValue = 0;

    if (chartOptions.barMode === 'percent') {
        // Percent 모드: 최대값은 100
        maxYValue = 100;
    } else if (chartOptions.barMode === 'stacked') {
        // Stacked 모드: 각 X축 포인트에서 모든 시리즈의 합계 중 최대값
        if (cData && cData.length > 1) {
            const seriesData = cData.slice(1); // x축 데이터 제외
            const dataPointCount = cData[0]?.length || 0;

            for (let i = 0; i < dataPointCount; i++) {
                let stackedSum = 0;
                for (const series of seriesData) {
                    const value = series[i];
                    if (typeof value === 'number' && !isNaN(value)) {
                        stackedSum += value;
                    }
                }
                if (stackedSum > maxYValue) {
                    maxYValue = stackedSum;
                }
            }
        }
    } else {
        // Single/Grouped 모드: 모든 시리즈에서 개별 최대값
        if (cData && cData.length > 1) {
            for (let i = 1; i < cData.length; i++) {
                const seriesData = cData[i];
                if (Array.isArray(seriesData)) {
                    for (const value of seriesData) {
                        // value는 number | null 타입
                        if (value != null && !isNaN(value) && value > maxYValue) {
                            maxYValue = value;
                        }
                    }
                }
            }
        }
    }

    // Y축 최대값에 10% 여유 추가
    const yMax = chartOptions.scales?.yMax ? Number(chartOptions.scales.yMax) : (maxYValue > 0 ? maxYValue * Y_AXIS_PADDING_RATIO : 100);
    const yMin = chartOptions.scales?.yMin ? Number(chartOptions.scales.yMin) : 0;

    // 가로 바 차트일 때 데이터 개수에 따라 최소 높이 계산

    let currentChartWidth: number;
    let currentChartHeight: number;
    let legendHeight = 0;
    if (chartOptions.legendEnabled && (chartHeight ?? 0) > 159) {
        if (chartOptions.legendAlign === "bottom") {
            // TODO 중복 코드 분리 필요
            currentChartWidth = chartWidth ? chartWidth : 50;
            if (chartOptions.legendAsTable) {
                if (payload.length === 2) {
                    currentChartHeight = (chartHeight ? chartHeight : 50) - 71;
                } else if (payload.length === 1) {
                    currentChartHeight = (chartHeight ? chartHeight : 50) - 42;
                } else {
                    currentChartHeight = (chartHeight ? chartHeight : 50) - 98;
                }
            } else {
                currentChartHeight = (chartHeight ? chartHeight : 50) - 46; //일반 legend bottom height
            }
        } else if (chartOptions.legendAlign === "right") {
            currentChartHeight = chartHeight ? chartHeight : 90;
            let legendWidth = 0;
            if (uPlotLegendRef.current) {
                const uPlotLegendDiv = uPlotLegendRef.current as HTMLDivElement;
                if (uPlotLegendDiv && chartOptions.legendAlign === "right") {
                    legendWidth = uPlotLegendRef.current.offsetWidth;
                }
            }
            currentChartWidth = (chartWidth ? chartWidth : 150) - legendWidth;

            legendHeight = currentChartHeight;
        } else {
            currentChartWidth = chartWidth ? chartWidth : 150;
            currentChartHeight = chartHeight ? chartHeight : 98;

        }
    } else {
        currentChartWidth = chartWidth ? chartWidth : 150;
        currentChartHeight = chartHeight ? chartHeight : 98;
        legendHeight = 0;
    }

    // cursor 옵션 - 바 차트용으로 조정
    const cursorOpts = {
        x: true,
        y: true,
        lock: false, // 바 차트에서는 lock하지 않음
        points: { show: false }, // 바 차트에서는 points 숨김
        drag: {
            setScale: false, // 드래그 선택으로 스케일 변경 비활성화
            x: false, // x축 드래그 비활성화
            y: false, // y축 드래그 비활성화
        },
    };

    // 옵션
    const options: UPlotOptions = {
        padding: isHorizontal
            ? [25, 15, 35, Math.max(yLabelWidth, 50)] // horizontal: 상단 여유 증가
            : [15, 25, 35, Math.max(yLabelWidth, 50)], // vertical: 오른쪽 여유 증가
        width: currentChartWidth,
        height: currentChartHeight,
        pxAlign: 1,
        plugins: [barTooltipPlugin({
            unitFn,
            tooltipLimit: chartOptions.tooltipLimit,
            isHorizontal,
            categoryLabels,
            barColorMode,
            categoryColors,
            barMode: chartOptions.barMode || 'single'
        })],
        legend: {
            show: true,
        },
        scales: isHorizontal ? {
            // 가로 바 차트: x는 값, y는 카테고리
            x: {
                min: yMin,
                max: yMax,
                auto: false, // 자동 스케일 조정 비활성화
            },
            y: {
                min: -0.5,
                max: dataLength - 0.5,
                auto: false, // 자동 스케일 조정 비활성화
            },
        } : {
            // 세로 바 차트: x는 카테고리, y는 값
            x: {
                min: -0.5,
                max: dataLength - 0.5,
                auto: false, // 자동 스케일 조정 비활성화
            },
            y: {
                min: yMin,
                max: yMax,
                auto: false, // 자동 스케일 조정 비활성화
            },
        },
        axes: isHorizontal ? [
            {   // x-axis (값)
                show: true,
                grid: { show: true, stroke: "border", width: 1 },
                font: "11px inter",
                stroke: "gray",
                ticks: { size: 1 },
                values: (_u, vals, _space) => {
                    let strUnitLength = 1;
                    return vals.map((value) => {
                        let strUnit = `${unitFn ? unitFn(value) : value}`;
                        if (strUnit.length > strUnitLength) {
                            strUnitLength = strUnit.length;
                            setYLabelWidth(strUnitLength * Y_LABEL_MULTIPLIER);
                        }
                        return strUnit;
                    });
                },
            },
            {   // y-axis (카테고리)
                show: true,
                grid: { show: false },
                font: "11px Inter",
                stroke: "gray",
                ticks: { size: 1 },
                values: (_u, vals, _space) => {
                    // 정수 인덱스만 필터링하여 카테고리 라벨 표시 (중복 방지)
                    return vals.map((value) => {
                        const index = Math.round(value);
                        // 정수가 아니거나 범위를 벗어나면 빈 문자열 반환
                        if (Math.abs(value - index) > 0.01 || index < 0 || index >= dataLength) {
                            return '';
                        }
                        return categoryLabels[index] || '';
                    });
                },
            },
        ] : [
            {   // x-axis (카테고리)
                show: true,
                grid: { show: false },
                font: "11px inter",
                stroke: "gray",
                ticks: { size: 1 },
                values: (_u, vals, _space) => {
                    // 정수 인덱스만 필터링하여 카테고리 라벨 표시
                    return vals.map((value) => {
                        const index = Math.round(value);
                        // 정수가 아니거나 범위를 벗어나면 빈 문자열 반환
                        if (Math.abs(value - index) > 0.01 || index < 0 || index >= dataLength) {
                            return '';
                        }
                        return categoryLabels[index] || '';
                    });
                },
            },
            {   // y-axis (값)
                show: true,
                grid: { show: true, stroke: "border", width: 1 },
                values: (_u, vals, _space) => {
                    let strUnitLength = 1;
                    return vals.map((value) => {
                        let strUnit = `${unitFn ? unitFn(value) : value}`;
                        if (strUnit.length > strUnitLength) {
                            strUnitLength = strUnit.length;
                            setYLabelWidth(strUnitLength * Y_LABEL_MULTIPLIER);
                        }
                        return strUnit;
                    });
                },
                font: "11px Inter",
                stroke: "gray",
                ticks: { size: 1 }
            },
        ],
        series: chartSeries,
        cursor: cursorOpts,
        hooks: {
            init: [
                (u) => {
                    // 더블 클릭 이벤트 비활성화 (캡처 단계에서 차단)
                    u.over.addEventListener('dblclick', (e) => {
                        e.preventDefault();
                        e.stopPropagation();
                        e.stopImmediatePropagation();
                    }, true);

                    // 마우스 휠 줌 비활성화
                    u.over.addEventListener('wheel', (e) => {
                        e.preventDefault();
                        e.stopPropagation();
                    }, { passive: false });

                    // bar 클릭 이벤트 핸들러 추가 (바 영역만 감지하는 토글 방식)
                    if (!chartOptions.onClickVariable) return;

                    const variableId = chartOptions.onClickVariable;

                    u.over.addEventListener('click', (e) => {
                        const rect = u.over.getBoundingClientRect();
                        const mouseX = e.clientX - rect.left;
                        const mouseY = e.clientY - rect.top;


                        // 클릭된 위치를 데이터 인덱스로 변환
                        const idx = Math.round(u.posToVal(isHorizontal ? mouseY : mouseX, isHorizontal ? "y" : "x"));

                        if (idx < 0 || idx >= categoryLabels.length) return;

                        // 바 영역 체크
                        const groupWidth = BAR_GROUP_WIDTH;
                        const barGap = BAR_GAP;
                        const barMode = chartOptions.barMode || 'single';
                        const seriesCount = u.series.length - 1;

                        let isInBarArea = false;

                        // 각 시리즈를 확인하여 클릭이 바 영역 내에 있는지 체크
                        for (let sidx = 1; sidx < u.series.length; sidx++) {
                            const series = u.series[sidx];
                            if (!series.show) continue;

                            const data = u.data[sidx];
                            const value = data[idx];

                            if (value == null || typeof value !== 'number' || isNaN(value)) continue;

                            // 바 너비 계산
                            let barWidth: number;
                            if (barMode === 'single' || barMode === 'stacked' || barMode === 'percent') {
                                barWidth = groupWidth;
                            } else {
                                barWidth = (groupWidth - (seriesCount - 1) * barGap) / seriesCount;
                            }

                            let barOffset = 0;
                            if (barMode === 'grouped') {
                                barOffset = (sidx - 1) * (barWidth + barGap) - (groupWidth - barWidth) / 2;
                            }

                            // 스택된 바의 경우 시작점 계산
                            let baseValue = 0;
                            if (barMode === 'stacked' || barMode === 'percent') {
                                if (sidx > 1) {
                                    const prevSeriesValue = u.data[sidx - 1]?.[idx];
                                    baseValue = typeof prevSeriesValue === 'number' && !isNaN(prevSeriesValue) ? prevSeriesValue : 0;
                                }
                            }

                            if (isHorizontal) {
                                // 가로 바: X축은 값, Y축은 카테고리
                                const xMin = u.valToPos(baseValue, "x");
                                const xMax = u.valToPos(value, "x");
                                const yMin = u.valToPos(idx - barWidth / 2 + barOffset, "y");
                                const yMax = u.valToPos(idx + barWidth / 2 + barOffset, "y");

                                if (mouseX >= Math.min(xMin, xMax) && mouseX <= Math.max(xMin, xMax) &&
                                    mouseY >= Math.min(yMin, yMax) && mouseY <= Math.max(yMin, yMax)) {
                                    isInBarArea = true;
                                    //console.log(`✓ Bar ${sidx} clicked!`);
                                    break;
                                }
                            } else {
                                // 세로 바: X축은 카테고리, Y축은 값
                                const xMin = u.valToPos(idx - barWidth / 2 + barOffset, "x");
                                const xMax = u.valToPos(idx + barWidth / 2 + barOffset, "x");
                                const yMin = u.valToPos(baseValue, "y");
                                const yMax = u.valToPos(value, "y");

                                if (mouseX >= Math.min(xMin, xMax) && mouseX <= Math.max(xMin, xMax) &&
                                    mouseY >= Math.min(yMin, yMax) && mouseY <= Math.max(yMin, yMax)) {
                                    isInBarArea = true;
                                    //console.log(`✓ Bar ${sidx} clicked!`);
                                    break;
                                }
                            }
                        }

                        // 바 영역 내에서만 variable 토글 (setFilter 있을 때만)
                        if (isInBarArea && setFilter && getFilter) {
                            const legendValue = categoryLabels[idx];

                            // 현재 설정된 값 가져오기
                            const currentValue = getFilter(variableId);

                            // 토글 방식: 같은 값이면 해제(빈 문자열), 다른 값이면 설정
                            if (currentValue === legendValue) {
                                setFilter(variableId, '');
                            } else {
                                setFilter(variableId, legendValue);
                            }
                        }
                    });
                }
            ],
            draw: [
                (u) => {
                    const ctx = u.ctx;
                    const { left, top, width, height } = u.bbox;

                    // Variable에서 선택된 카테고리 값 가져오기 (getFilter 있을 때만)
                    const variableId = chartOptions.onClickVariable;
                    const selectedCategory = variableId && getFilter ? getFilter(variableId) : null;
                    // 빈 문자열, null, undefined, 빈 배열은 선택되지 않은 것으로 처리
                    const hasSelection = !!selectedCategory &&
                        selectedCategory !== '' &&
                        (!Array.isArray(selectedCategory) || selectedCategory.length > 0);

                    // 카테고리 모드일 때만 bar를 다시 그림
                    if (barColorMode === 'category') {
                        const groupWidth = BAR_GROUP_WIDTH;
                        const barGap = BAR_GAP;
                        const barMode = chartOptions.barMode || 'single';
                        const seriesCount = u.series.length - 1; // 첫 번째는 x축

                        for (let sidx = 1; sidx < u.series.length; sidx++) {
                            const series = u.series[sidx];
                            if (!series.show) continue;

                            const data = u.data[sidx];

                            for (let i = 0; i < data.length; i++) {
                                const value = data[i];
                                if (value == null || typeof value !== 'number' || isNaN(value)) continue;

                                // 카테고리별 색상 적용
                                const categoryColor = categoryColors[i] || 'hsl(221, 83%, 53%)';
                                const categoryLabel = categoryLabels[i];
                                const isSelected = selectedCategory === categoryLabel;

                                let barWidth: number;
                                if (barMode === 'single' || barMode === 'stacked' || barMode === 'percent') {
                                    barWidth = groupWidth;
                                } else {
                                    barWidth = (groupWidth - (seriesCount - 1) * barGap) / seriesCount;
                                }

                                let barOffset = 0;
                                if (barMode === 'grouped') {
                                    barOffset = (sidx - 1) * (barWidth + barGap) - (groupWidth - barWidth) / 2;
                                }

                                let baseValue = 0;
                                if (barMode === 'stacked' || barMode === 'percent') {
                                    if (sidx > 1) {
                                        const prevSeriesValue = u.data[sidx - 1]?.[i];
                                        baseValue = typeof prevSeriesValue === 'number' && !isNaN(prevSeriesValue) ? prevSeriesValue : 0;
                                    }
                                }

                                ctx.save();

                                // 선택된 카테고리가 있을 때, 선택되지 않은 바는 opacity 0.3 적용
                                if (hasSelection && !isSelected) {
                                    ctx.globalAlpha = 0.3;
                                }

                                ctx.fillStyle = categoryColor;

                                let x0, x1, y0, y1;

                                if (isHorizontal) {
                                    x0 = u.valToPos(baseValue, "x", true);
                                    x1 = u.valToPos(value, "x", true);
                                    y0 = u.valToPos(i - barWidth / 2 + barOffset, "y", true);
                                    y1 = u.valToPos(i + barWidth / 2 + barOffset, "y", true);

                                    ctx.fillRect(Math.min(x0, x1), Math.min(y0, y1), Math.abs(x1 - x0), Math.abs(y1 - y0));
                                } else {
                                    x0 = u.valToPos(i - barWidth / 2 + barOffset, "x", true);
                                    x1 = u.valToPos(i + barWidth / 2 + barOffset, "x", true);
                                    y0 = u.valToPos(baseValue, "y", true);
                                    y1 = u.valToPos(value, "y", true);

                                    ctx.fillRect(Math.min(x0, x1), Math.min(y0, y1), Math.abs(x1 - x0), Math.abs(y1 - y0));
                                }

                                ctx.restore();
                            }
                        }
                    }

                    // 임계선 그리기
                    const thresholds = chartOptions.thresholds || [];
                    thresholds.forEach(({ value, color, type }) => {
                        if (!value || !color) return;

                        ctx.save();
                        ctx.strokeStyle = color;
                        ctx.lineWidth = 2;

                        let dash: number[];
                        switch (type) {
                            case "dash": dash = [5, 5]; break;
                            case "dot": dash = [2, 6]; break;
                            case "solid": dash = []; break;
                            default: dash = []; break;
                        }
                        ctx.setLineDash(dash);
                        ctx.beginPath();

                        if (isHorizontal) {
                            // 가로 바 차트에서는 세로 선
                            const xPos = u.valToPos(parseFloat(value), "x", true);
                            ctx.moveTo(xPos, top);
                            ctx.lineTo(xPos, top + height);
                        } else {
                            // 세로 바 차트에서는 가로 선
                            const yPos = u.valToPos(parseFloat(value), "y", true);
                            ctx.moveTo(left, yPos);
                            ctx.lineTo(left + width, yPos);
                        }

                        ctx.stroke();
                        ctx.restore();
                    });

                    // 값 라벨 그리기 (옵션)
                    if (chartOptions.showValueLabels) {
                        ctx.save();
                        ctx.fillStyle = "rgba(0, 0, 0, 0.8)";
                        ctx.font = "10px Inter";
                        ctx.textAlign = "center";

                        const barMode = chartOptions.barMode || 'single';
                        const groupWidth = BAR_GROUP_WIDTH;
                        const barGap = BAR_GAP;
                        const seriesCount = u.series.length - 1;

                        // 각 시리즈의 각 바에 값 표시
                        for (let sidx = 1; sidx < u.series.length; sidx++) {
                            const series = u.series[sidx];
                            if (!series.show) continue;

                            const data = u.data[sidx];

                            // Grouped 모드에서 바 위치 계산
                            let barWidth: number;
                            let barOffset = 0;

                            if (barMode === 'single' || barMode === 'stacked' || barMode === 'percent') {
                                barWidth = groupWidth;
                            } else { // grouped
                                barWidth = (groupWidth - (seriesCount - 1) * barGap) / seriesCount;
                                barOffset = (sidx - 1) * (barWidth + barGap) - (groupWidth - barWidth) / 2;
                            }

                            for (let i = 0; i < data.length; i++) {
                                const value = data[i];
                                if (value == null || typeof value !== 'number' || isNaN(value)) continue;

                                // 스택 모드일 때 기준값 계산
                                let baseValue = 0;
                                if (barMode === 'stacked' || barMode === 'percent') {
                                    if (sidx > 1) {
                                        const prevSeriesValue = u.data[sidx - 1]?.[i];
                                        baseValue = typeof prevSeriesValue === 'number' && !isNaN(prevSeriesValue) ? prevSeriesValue : 0;
                                    }
                                }

                                if (isHorizontal) {
                                    const xPos = u.valToPos(value, "x", true);
                                    const yPos = u.valToPos(i + barOffset, "y", true);
                                    ctx.fillText(unitFn(value), xPos + 5, yPos + 3);
                                } else {
                                    const xPos = u.valToPos(i + barOffset, "x", true);
                                    const yPos = barMode === 'stacked' || barMode === 'percent'
                                        ? u.valToPos((baseValue + value) / 2, "y", true) // 스택 모드: 바의 중앙
                                        : u.valToPos(value, "y", true); // 일반 모드: 바의 상단
                                    ctx.fillText(unitFn(value), xPos, yPos - 5);
                                }
                            }
                        }

                        ctx.restore();
                    }
                }
            ]
        },
    };

    return (
        <div className={cn("flex w-full h-full", chartOptions.legendAlign === "bottom" ? "flex-col" : "flex-row")} >
            <div ref={uPlotRefDiv} className={cn(
                "relative",
                chartOptions.legendAlign === "right" ? "flex-1" : "w-full"
            )} style={{ position: 'relative' }}>
                <UplotReact
                    key={`${categoryLabels.join('-')}-${chartMetric?.length || 0}`}
                    options={options}
                    data={statChartData}
                    onCreate={(chart) => (uPlotRef.current = chart)}
                    onDelete={() => { }}
                />
            </div>
            <div ref={uPlotLegendRef} className={cn(
                "flex text-xs",
                chartOptions.legendAlign === "bottom" ? "w-full" : "shrink-0",
                !(chartOptions.legendEnabled && chartHeight && chartHeight > 160) && "hidden"
            )}
                style={{ maxHeight: `${legendHeight}px` }}>
                {chartOptions.legendEnabled &&
                    (chartOptions.legendAsTable ? (
                        <PromLegendTable
                            id={id}
                            className={cn(
                                legendOptionAttribute.className,
                                chartOptions.legendAlign === "right" && "w-full justify-start",
                                chartOptions.legendAlign === "bottom" && "w-full h-full max-h-24.75"
                                //TODO legend bottom 정렬의 경우 각 프로젝트 사이즈에 따라 max-h-[$]수치 변경 필요 
                            )}
                            payload={payload}
                            metrics={barColorMode === 'category' ? categoryMetrics : metrics}
                            showAverage={false}
                            showLast={false}
                            showMax={false}
                            showMin={false}
                            showTotal={false}
                            unitFn={unitFn}
                            selectedLegend={selectedLegend}
                            selectedLegendFn={selectedLegendHandler}
                            isBarChart={true}
                        //showLink={showLink}
                        />
                    ) : (
                        <PromLegend
                            className={cn(
                                legendOptionAttribute.className,
                                chartOptions.legendAlign === "right" && "w-full pr-2",
                                chartOptions.legendAlign === "bottom" && "justify-center w-full max-w-full leading-none flex flex-wrap max-h-10.5" //TODO legend bottom 정렬의 경우 각 프로젝트 사이즈에 따라 max-h-[$]수치 변경 필요
                            )}
                            payload={payload}
                            metrics={barColorMode === 'category' ? categoryMetrics : metrics}
                            showAverage={false}
                            showLast={false}
                            showMax={false}
                            showMin={false}
                            showTotal={false}
                            unitFn={unitFn}
                            selectedLegend={selectedLegend}
                            selectedLegendFn={selectedLegendHandler}
                        />
                    ))}
            </div>
        </div>
    );
}

export { BarChart };
