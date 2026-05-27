import {Index} from '@pharos/shared/types/datasource';
import {ChartMetric, ChartMetricData, ChartMetricLabel} from '../types';
import {parseTimestampToMs, toUnixMilliseconds} from '@pharos/shared/lib';

// 파이프라인 단계별 데이터 타입
interface ProcessedRow {
  timestamp: number;  // 밀리초
  numericFields: Record<string, number>;
  stringFields: Record<string, string>;
}

interface ChartPoint {
  timestamp: number;  // 밀리초
  label: string;
  value: number;
}

const numericStringPattern = /^[+-]?(?:(?:\d+\.?\d*)|(?:\.\d+))(?:e[+-]?\d+)?$/i;

const isNumericString = (value: string): boolean => numericStringPattern.test(value.trim());

// 1단계: 시간 파싱 (항상 밀리초 반환)
/**
 * 타임스탬프를 밀리초로 파싱
 * 
 * @deprecated 새 코드에서는 parseTimestampToMs를 직접 import하여 사용하세요
 * @see {@link parseTimestampToMs}
 */
const parseTimestamp = parseTimestampToMs;

// 2단계: 필드 분류
const classifyFields = (row: Record<string, unknown>): ProcessedRow => {
  const timestamp = parseTimestamp(row.time ?? row.TIME ?? row.timestamp);
  const numericFields: Record<string, number> = {};
  const stringFields: Record<string, string> = {};

  const isTimeField = (key: string) => ['time', 'timestamp'].includes(key.toLowerCase());

  for (const key in row) {
    const value = row[key];
    if (isTimeField(key)) continue;

    if (typeof value === 'number') {
      numericFields[key] = value;
    } else if (typeof value === 'string' && isNumericString(value)) {
      numericFields[key] = Number(value);
    } else if (typeof value === 'string') {
      stringFields[key] = value;
    }
  }

  return { timestamp, numericFields, stringFields };
};

// 3단계: 차트 포인트 생성
const createChartPoints = (processed: ProcessedRow[], query: string, name: string, label: string): ChartPoint[] => {
  const points: ChartPoint[] = [];

  processed.forEach(({ timestamp, numericFields, stringFields }) => {
    Object.entries(numericFields).forEach(([numKey, numValue]) => {
      // 라벨 데이터 구성
      const metricData = Object.keys(stringFields).length > 0
        ? stringFields
        : { columnName: numKey };

      const chartLabel = JSON.stringify({
        metric: metricData,
        _query: query,
        _name: name,
        _label: label,
      } as ChartMetricLabel);

      points.push({
        timestamp,
        label: chartLabel,
        value: numValue
      });
    });
  });

  return points;
};

// 4단계: 피봇 포인트 생성
const createPivotChartPoints = (
  processed: ProcessedRow[],
  pivotKey: string,
  query: string,
  name: string,
  label: string
): ChartPoint[] => {
  const points: ChartPoint[] = [];

  for (const row of processed) {
    const { timestamp, numericFields, stringFields } = row;
    const pivotValue = stringFields[pivotKey];

    if (!pivotValue) continue;

    for (const numKey in numericFields) {
      const metricData = { [pivotKey]: pivotValue };
      const chartLabel = JSON.stringify({
        metric: metricData,
        _query: query,
        _name: name,
        _label: label,
      });

      points.push({ timestamp, label: chartLabel, value: numericFields[numKey] });
    }
  }

  return points;
};

// 5단계: 최종 결과 구성
const buildChartData = (points: ChartPoint[]): ChartMetricData => {
  const output = new Map<number, ChartMetric>();
  const uniqueKeys: string[] = [];

  points.forEach(({ timestamp, label, value }) => {
    if (!uniqueKeys.includes(label)) uniqueKeys.push(label);

    const existing = output.get(timestamp);
    if (existing) {
      existing[label] = value;
    } else {
      const metric: ChartMetric = { timestamp };
      metric[label] = value;
      output.set(timestamp, metric);
    }
  });

  return {
    chartMetric: Array.from(output.values()),
    chartType: 'timeseries',
    endTime: 0,
    startTime: 0,
    step: 0,
    chartMetricMap: output,
    uniqueKeys,
  };
};

// 피봇 설정 추출
const extractPivotConfig = (label: string): { usePivot: boolean; pivotKey?: string } => {
  try {
    const parsed = JSON.parse(label);
    if (parsed.pivot_key && typeof parsed.pivot_key === 'string') {
      return { usePivot: true, pivotKey: parsed.pivot_key };
    }
  } catch {
    // JSON 파싱 실패시 피봇 없음
  }
  return { usePivot: false };
};

// 메인 변환 파이프라인
function transformDataToMap(
  inputData: Index,
  query: string,
  name: string,
  label: string,
): ChartMetricData {
  if (!inputData.data) return buildChartData([]);

  const processed = inputData.data.map(classifyFields);
  const pivotConfig = extractPivotConfig(label);

  const points = pivotConfig.usePivot
    ? createPivotChartPoints(processed, pivotConfig.pivotKey!, query, name, label)
    : createChartPoints(processed, query, name, label);

  return buildChartData(points);
}

export function getNumber(time: number | string | (() => number) | undefined) {
  switch (typeof time) {
    case 'number':
      return time;
    case 'string':
      return parseInt(time, 10);
    case 'function':
      return time();
    default:
      return 0;
  }
}

export function responseConvert(
  src: ChartMetricData,
  data: Index,
  query: string,
  name: string,
  label: string,
  startTime: number,
  endTime: number,
  step: number,
): ChartMetricData {
  // chart data 형태로 data를 변환
  const result = transformDataToMap(data, query, name, label);

  // ✅ 다중 쿼리 결과 병합 (덮어쓰기 → 병합)
  // chartMetricMap 초기화 (없으면 생성)
  if (!src.chartMetricMap) {
    src.chartMetricMap = new Map();
  }

  // chartMetricMap 병합: 같은 timestamp면 필드 병합, 없으면 추가
  if (result.chartMetricMap) {
    result.chartMetricMap.forEach((metric, timestamp) => {
      const existing = src.chartMetricMap!.get(timestamp);
      if (existing) {
        // 같은 timestamp: 기존 객체에 새 필드 병합
        // ⚠️ 필드 이름 충돌 감지 (개발 모드에서만)
        if (process.env.NODE_ENV !== 'production') {
          const newKeys = Object.keys(metric as Record<string, unknown>).filter(
            key => key !== 'timestamp' // timestamp는 당연히 같으므로 제외
          );
          const overlappingKeys = newKeys.filter(
            (key) =>
              key in existing &&
              (existing as Record<string, unknown>)[key] !==
                (metric as Record<string, unknown>)[key],
          );
          if (overlappingKeys.length > 0) {
            console.warn(
              `[Metric Merge] Field collision at timestamp ${timestamp}:\n` +
              `  Conflicting fields: ${overlappingKeys.join(', ')}\n` +
              `  Values will be overwritten by the latest query result.`
            );
          }
        }
        Object.assign(existing, metric);
      } else {
        // 새 timestamp: 복사하여 추가
        src.chartMetricMap!.set(timestamp, { ...metric });
      }
    });
  }

  // chartMetricMap에서 chartMetric 배열 재생성 (timestamp 순 정렬)
  src.chartMetric = Array.from(src.chartMetricMap.values())
    .sort((a, b) => a.timestamp - b.timestamp);

  // uniqueKeys 중복 제거하여 병합
  const mergedKeys = new Set([...(src.uniqueKeys || []), ...(result.uniqueKeys || [])]);
  src.uniqueKeys = Array.from(mergedKeys);

  // ✅ 시간 범위 업데이트
  // - 입력: startTime/endTime/step은 초 단위 (백엔드 API 규격)
  // - 내부: 밀리초로 통일 (parseTimestampToMs가 밀리초를 반환하므로)
  const startTimeMs = toUnixMilliseconds(startTime);
  const endTimeMs = toUnixMilliseconds(endTime);
  const stepMs = toUnixMilliseconds(step);
  
  if (src.startTime === 0) {
    src.startTime = startTimeMs;
  } else if (src.startTime <= startTimeMs && startTimeMs != 0) {
    src.startTime = startTimeMs;
  }

  if (src.endTime === 0) {
    src.endTime = endTimeMs;
  } else if (src.endTime >= endTimeMs && endTimeMs != 0) {
    src.endTime = endTimeMs;
  }
  src.step = stepMs;

  return src;
}
