/**
 * Query Preview Component
 * 
 * Alert query를 실행하고 결과를 차트로 표시합니다.
 * - Run 버튼만 노출
 * - Run 클릭 시: API 호출 후 바로 아래에 차트 표시
 */

'use client';

import React, {useState, useRef, useEffect} from 'react';
import {Button} from '@pharos/shared/components/ui';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import {TimeSeriesChart} from '@pharos/shared/components/charts';
import {
  QUERY_PARAM_START_TIME,
  QUERY_PARAM_END_TIME,
  QUERY_PARAM_STEP,
} from '@lib/query-params';
import type {ChartMetricData, ChartMetric} from '@pharos/shared/components/charts';
import type {AlertQueryRequest, AlertQueryResponse} from '@pharos/shared/types/alert';
import {Play, Loader2} from 'lucide-react';
import {useCreate} from '@/lib/data-provider';
import {ALERT_PROVIDER_NAME, ALERT_RESOURCES} from '@providers/alert-provider/types';

// Time range options (in seconds)
const TIME_RANGES = [
  {label: '5 minutes', value: 5 * 60},
  {label: '10 minutes', value: 10 * 60},
  {label: '30 minutes', value: 30 * 60},
  {label: '1 hour', value: 60 * 60},
  {label: '24 hours', value: 24 * 60 * 60},
] as const;

// Step options (in seconds)
const STEP_OPTIONS = [
  {label: '10 seconds', value: 10},
  {label: '30 seconds', value: 30},
  {label: '1 minute', value: 60},
  {label: '10 minutes', value: 10 * 60},
  {label: '30 minutes', value: 30 * 60},
  {label: '60 minutes', value: 60 * 60},
] as const;

interface QueryPreviewProps {
  datasource: string;
  query: string;
  variables?: Record<string, string | number | boolean>;
  disabled?: boolean;
}

/**
 * AlertQueryResponse (DatabaseResponse)를 ChartMetricData로 변환
 */
function convertToChartData(response: AlertQueryResponse): ChartMetricData {
  const {meta, data} = response;

  // timestamp 컬럼 찾기 (time, timestamp, datetime 등)
  const timeColumnIndex = meta.findIndex((col) =>
    ['time', 'timestamp', 'datetime', 't'].includes(col.name?.toLowerCase() || ''),
  );

  if (timeColumnIndex === -1) {
    throw new Error('No time column found in query result');
  }

  const timeColumnName = meta[timeColumnIndex].name;

  // 시계열 데이터 생성
  const chartMetric: ChartMetric[] = data.map((row) => {
    const metric: ChartMetric = {
      timestamp: new Date(row[timeColumnName] as string | number).getTime(),
    };

    // 다른 컬럼들을 메트릭으로 추가
    meta.forEach((col) => {
      if (col.name !== timeColumnName) {
        const value = row[col.name];
        metric[col.name] =
          typeof value === 'number' ? value : value === null ? null : String(value);
      }
    });

    return metric;
  });

  // 시간 범위 계산
  const timestamps = chartMetric.map((m) => m.timestamp);
  const startTime = Math.min(...timestamps);
  const endTime = Math.max(...timestamps);

  // step 계산 (평균 간격)
  let step = 60; // 기본값: 1분
  if (chartMetric.length > 1) {
    const intervals = [];
    for (let i = 1; i < chartMetric.length; i++) {
      intervals.push((chartMetric[i].timestamp - chartMetric[i - 1].timestamp) / 1000);
    }
    step = Math.round(intervals.reduce((a, b) => a + b, 0) / intervals.length);
  }

  // 메트릭 키 목록 (timestamp 제외)
  const uniqueKeys = meta.filter((col) => col.name !== timeColumnName).map((col) => col.name);

  return {
    chartMetric,
    startTime,
    endTime,
    step,
    chartType: 'timeseries', // TimeSeriesChart는 'timeseries' 타입을 기대
    uniqueKeys,
  };
}

export function QueryPreview({datasource, query, variables, disabled}: QueryPreviewProps) {
  const [showPreview, setShowPreview] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [chartData, setChartData] = useState<ChartMetricData | undefined>(undefined);
  const [timeRange, setTimeRange] = useState<number>(5 * 60); // Default: 5 minutes
  const [step, setStep] = useState<number>(30); // Default: 30 seconds
  const [chartWidth, setChartWidth] = useState<number>(800); // Default width
  const chartContainerRef = useRef<HTMLDivElement>(null);

  const { mutate: executeQuery, mutation: queryMutation } = useCreate<AlertQueryResponse>({
    resource: ALERT_RESOURCES.QUERY,
    dataProviderName: ALERT_PROVIDER_NAME,
  });
  const loading = queryMutation.isPending;

  // 차트 컨테이너 크기 감지
  useEffect(() => {
    if (!chartContainerRef.current) return;

    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        setChartWidth(entry.contentRect.width);
      }
    });

    resizeObserver.observe(chartContainerRef.current);

    return () => {
      resizeObserver.disconnect();
    };
  }, []);

  const handleRun = async () => {
    if (!datasource || !query) {
      setError('Datasource and query are required');
      return;
    }

    setError(null);

    // Calculate time range
    const endTime = Math.floor(Date.now() / 1000); // Current time in seconds
    const startTime = endTime - timeRange;

    const request: AlertQueryRequest = {
      datasource,
      query: {
        query,
        variables: {
          [QUERY_PARAM_START_TIME]: startTime,
          [QUERY_PARAM_END_TIME]: endTime,
          [QUERY_PARAM_STEP]: step,
          ...(variables && Object.keys(variables).length > 0 ? variables : {}),
        } as any,
      },
    };

    // Debug: Log the request payload
    console.log('[QueryPreview] Request payload:', JSON.stringify(request, null, 2));
    console.log('[QueryPreview] Variables types:', {
      [QUERY_PARAM_START_TIME]: typeof startTime,
      [QUERY_PARAM_END_TIME]: typeof endTime,
      [QUERY_PARAM_STEP]: typeof step,
    });

    executeQuery(
      {values: request as any},
      {
        onSuccess: (response) => {
          const data = response.data;

          if (!data || !data.data || data.data.length === 0) {
            setError('No data returned from query');
            setChartData(undefined);
            setShowPreview(false);
            return;
          }

          try {
            const chartData = convertToChartData(data);
            setChartData(chartData);
            setShowPreview(true);
          } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to convert data to chart format');
            setChartData(undefined);
            setShowPreview(false);
          }
        },
        onError: (err: any) => {
          setError(err?.message || 'Failed to execute query');
          setChartData(undefined);
          setShowPreview(false);
        },
      },
    );
  };

  return (
    <div className="space-y-4">
      {/* Time Range, Step Selection & Run Button */}
      <div className="flex items-center gap-3">
        <Select value={String(timeRange)} onValueChange={(val) => setTimeRange(Number(val))}>
          <SelectTrigger className="w-[180px] h-8">
            <SelectValue placeholder="Select time range" />
          </SelectTrigger>
          <SelectContent>
            {TIME_RANGES.map((range) => (
              <SelectItem key={range.value} value={String(range.value)}>
                {range.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={String(step)} onValueChange={(val) => setStep(Number(val))}>
          <SelectTrigger className="w-[180px] h-8">
            <SelectValue placeholder="Select step" />
          </SelectTrigger>
          <SelectContent>
            {STEP_OPTIONS.map((option) => (
              <SelectItem key={option.value} value={String(option.value)}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Button
          type="button"
          variant="secondary"
          className="h-8 shadow-none px-3 py-1"
          onClick={handleRun}
          disabled={disabled || loading || !datasource || !query}
        >
          {loading ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              Running...
            </>
          ) : (
            <>
              <Play className="h-4 w-4 mr-2" />
              Run Query
            </>
          )}
        </Button>
      </div>

      {/* Error Message */}
      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-md text-red-700 text-sm">
          {error}
        </div>
      )}

      {/* Chart Preview - Run 버튼 클릭 후에만 표시 */}
      {showPreview && chartData && (
        <div ref={chartContainerRef} className="border border-border rounded-md bg-muted/10 p-4">
          <TimeSeriesChart
            data={chartData}
            loading={loading}
            chartWidth={chartWidth}
            chartHeight={300}
            options={{
              legendEnabled: true,
              legendAlign: 'bottom',
              chartType: 'line',
              unit: 'short',
            }}
          />
        </div>
      )}
    </div>
  );
}
