import {useMemo} from 'react';
import {buildChartQueryArgs, buildDynamicArgsFromMap} from '@utils/chartArgsBuilder';
import {ChartQueryArgs} from '@types';
import type {DateTime, FilterValue, Step} from '@features/dashboard/types/filter.types';
import {QUERY_PARAM_REFRESH_COUNT} from '@lib/query-params';

/**
 * Panel Args 생성 Custom Hook (Map 기반)
 * 
 * PanelRenderer의 복잡도를 줄이기 위해 args 생성 로직을 분리
 * Map 기반으로 stale closure 문제 완전 해결
 * 
 * @param filters - Map<string, FilterValue>
 * @param filterIds - string[]
 * @param datetime - DateTime | undefined
 * @param step - Step | undefined
 * @param refreshCount - number (refresh 버튼 클릭 감지, useFilterItemStore에서 항상 제공)
 * @returns 차트 쿼리에 사용할 args 객체
 */
export function usePanelArgs(
  filters: Map<string, FilterValue> | undefined,
  filterIds: string[],
  datetime: DateTime | undefined,
  step: Step | undefined,
  refreshCount: number,
  format: 'sql' | 'regex' = 'sql',
): ChartQueryArgs {
  const startTime = datetime?.startTime;
  const endTime = datetime?.endTime;
  const stepValue = step?.stepValue;

  return useMemo(() => {
    const dynamicArgs = buildDynamicArgsFromMap(filterIds, filters, format);

    const args = buildChartQueryArgs({
      startTime,
      endTime,
      step: stepValue,
      dynamicArgs,
    });

    args.set(QUERY_PARAM_REFRESH_COUNT, refreshCount.toString());

    return args;
  }, [startTime, endTime, stepValue, filterIds, filters, refreshCount, format]);
}
