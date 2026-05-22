import {ChartQueryArgs} from '@types';
import {
  QUERY_PARAM_START_TIME,
  QUERY_PARAM_END_TIME,
  QUERY_PARAM_START_TIME_MS,
  QUERY_PARAM_END_TIME_MS,
  QUERY_PARAM_STEP,
  QUERY_PARAM_INTERVAL,
  QUERY_PARAM_INTERVAL_MS,
} from '@lib/query-params';
import {getAbsoluteValueTimestamp, getAbsoluteValueTimestampMs} from '@pharos/shared/components/ui-extension';

/**
 * 차트 쿼리 args 생성 유틸리티
 * 대시보드와 패널 에디터에서 공통으로 사용
 *
 * @description
 * - __start_time, __end_time을 실제 timestamp 값으로 즉시 변환
 * - 모든 값이 이미 resolved되어 있어 useDashboardQuery에서 추가 처리 불필요
 * - 단일 책임: args 생성의 모든 로직을 한 곳에서 관리
 */
export interface ChartArgsBuilderOptions {
  startTime?: any; // DateTime object
  endTime?: any; // DateTime object
  step?: number;
  refreshCount?: number;
  dynamicArgs?: Record<string, string>;
}

/**
 * 차트 쿼리에 사용되는 args 객체를 생성합니다.
 *
 * @example
 * ```typescript
 * const args = buildChartQueryArgs({
 *   startTime: filterState?.datetime?.startTime,
 *   endTime: filterState?.datetime?.endTime,
 *   step: 300,
 *   dynamicArgs: { Host: "'host1','host2'" }
 * });
 * // Result: {
 * //   __start_time: 1234567890,  // 실제 값 바로 반환
 * //   __end_time: 1234567999,
 * //   __step: 300,
 * //   Host: "'host1','host2'"
 * // }
 * ```
 */
// Grafana auto step과 동일한 방식: 시간범위 / max data points, 최소 15초
function calcAutoStep(startTime: any, endTime: any): number {
  const MAX_DATA_POINTS = 1000;
  const MIN_STEP = 30; // Prometheus 스크레이프 간격(15~30s) 고려, 30s 미만이면 빈 구간 발생
  try {
    const start = getAbsoluteValueTimestamp(startTime);
    const end = getAbsoluteValueTimestamp(endTime);
    if (!start || !end || end <= start) return 300;
    return Math.max(MIN_STEP, Math.ceil((end - start) / MAX_DATA_POINTS));
  } catch {
    return 300;
  }
}

export function buildChartQueryArgs(options: ChartArgsBuilderOptions): ChartQueryArgs {
  const {startTime, endTime, step, refreshCount, dynamicArgs = {}} = options;

  // step 명시 없으면 시간범위 기반 auto 계산 (Grafana auto step 방식)
  const resolvedStep = step ?? calcAutoStep(startTime, endTime);

  const result = new Map<string, string | number | (() => number)>();

  result.set(QUERY_PARAM_START_TIME, startTime ? getAbsoluteValueTimestamp(startTime) : 0);
  result.set(QUERY_PARAM_END_TIME, endTime ? getAbsoluteValueTimestamp(endTime) : 0);
  result.set(QUERY_PARAM_START_TIME_MS, startTime ? getAbsoluteValueTimestampMs(startTime) : 0);
  result.set(QUERY_PARAM_END_TIME_MS, endTime ? getAbsoluteValueTimestampMs(endTime) : 0);
  result.set(QUERY_PARAM_STEP, resolvedStep);
  result.set(QUERY_PARAM_INTERVAL, `${resolvedStep}s`);
  result.set(QUERY_PARAM_INTERVAL_MS, resolvedStep * 1000);

  // refreshCount가 있을 때만 추가 (패널 에디터용)
  if (refreshCount !== undefined) {
    result.set('refreshCount', refreshCount);
  }

  // 동적 args 추가
  Object.entries(dynamicArgs).forEach(([key, value]) => {
    result.set(key, value);
  });

  return result;
}

/**
 * Map 기반 Filter로부터 동적 args를 생성합니다.
 *
 * 단일 문자열 값은 raw 그대로 전달합니다.
 * 쿼리 템플릿에서 직접 따옴표를 처리해야 합니다: '{{SO}}'
 *
 * 배열(다중 선택)은 IN 절에 사용할 수 있도록 따옴표를 포함한 CSV 형태로 변환합니다.
 * 쿼리 템플릿: column IN ({{tags}})
 *
 * @example
 * ```typescript
 * const filters = new Map([['Host', 'server1'], ['Tags', ['tag1', 'tag2']]]);
 * const dynamicArgs = buildDynamicArgsFromMap(['Host', 'Tags'], filters);
 * // Result: { Host: "server1", Tags: "'tag1','tag2'" }
 * ```
 */
export function buildDynamicArgsFromMap(
  filterIds: string[],
  filters: Map<string, any> | undefined,
  format: 'sql' | 'regex' = 'sql',
): Record<string, string> {
  if (!filters) return {};

  return filterIds.reduce<Record<string, string>>((acc, id) => {
    const filterValue = filters.get(id);

    if (filterValue !== undefined) {
      if (Array.isArray(filterValue)) {
        if (format === 'regex') {
          // Prometheus regex OR 형식: val1|val2
          acc[id] = filterValue.length === 0 ? '' : filterValue.join('|');
        } else if (filterValue.length === 0) {
          acc[id] = "''";
        } else {
          // SQL IN 절 형식: 'val1','val2'
          acc[id] = filterValue.map((value) => `'${value}'`).join(',');
        }
      } else {
        acc[id] = String(filterValue);
      }
    }

    return acc;
  }, {});
}
