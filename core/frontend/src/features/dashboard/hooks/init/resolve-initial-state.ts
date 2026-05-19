/**
 * Dashboard Initial State Resolver
 *
 * 대시보드 로드 시 초기 상태를 결정하는 단일 함수.
 * _loadDashboard와 useDashboardAutoLoad 양쪽에서 공유하여
 * 중복된 3단계 우선순위 로직(URL > saved > default)을 통합한다.
 *
 * 이전에 _loadDashboard(170줄)과 useDashboardAutoLoad(80줄)에
 * 거의 동일한 코드가 복붙되어 있었고, step 파라미터명 불일치
 * (step vs var_step) 같은 미묘한 차이가 버그 원인이었다.
 */
import type { FilterConfig } from '@pharos/shared/types/dashboard';
import type { FilterValue, DateTime, Step } from '../../types/filter.types';
import { getStepValue, parseRefreshInterval } from '../../utils/datetime-step-utils';
import {
  URL_PARAM_FROM,
  URL_PARAM_TO,
  URL_PARAM_STEP,
  URL_PARAM_REFRESH,
  getFilterParamName,
} from '../url/url-param-constants';

export interface DashboardInitSources {
  urlParams: URLSearchParams | undefined;
  filters: FilterConfig[];
}

export interface ResolvedInitialState {
  filterValues: Map<string, FilterValue>;
  datetime: DateTime | undefined;
  step: Step | undefined;
  refreshInterval: number | false | undefined;
}

/**
 * 3단계 우선순위로 초기 상태를 결정한다.
 * 우선순위: URL > 저장된 값(savedValue) > 기본값
 */
export async function resolveInitialState(
  sources: DashboardInitSources,
): Promise<ResolvedInitialState> {
  const { urlParams, filters } = sources;

  return {
    filterValues: resolveFilterValues(urlParams, filters),
    datetime: await resolveDatetime(urlParams, filters),
    step: resolveStep(urlParams, filters),
    refreshInterval: resolveRefreshInterval(urlParams, filters),
  };
}

// ─── Select Filter Values ────────────────────────────────────────────────────

function resolveFilterValues(
  urlParams: URLSearchParams | undefined,
  filters: FilterConfig[],
): Map<string, FilterValue> {
  const values = new Map<string, FilterValue>();
  if (!urlParams) return values;

  filters.forEach((f) => {
    const paramName = getFilterParamName(f.id);
    const urlValue = urlParams.get(paramName);
    if (urlValue) {
      if (f.type === 'select' && f.options?.isMulti) {
        values.set(f.id, urlValue.split(',').filter((v) => v));
      } else {
        values.set(f.id, urlValue);
      }
    }
  });

  return values;
}

// ─── DateTime ────────────────────────────────────────────────────────────────

async function resolveDatetime(
  urlParams: URLSearchParams | undefined,
  filters: FilterConfig[],
): Promise<DateTime | undefined> {
  // 1순위: URL 파라미터
  const fromParam = urlParams?.get(URL_PARAM_FROM);
  const toParam = urlParams?.get(URL_PARAM_TO);

  if (fromParam && toParam) {
    try {
      const { urlStringToDateTimeValue } = await import(
        '@pharos/shared/components/ui-extension'
      );
      const startTime = urlStringToDateTimeValue(fromParam);
      const endTime = urlStringToDateTimeValue(toParam);
      if (startTime && endTime) {
        return { startTime, endTime };
      }
    } catch (err) {
      console.warn('[resolveInitialState] Invalid from/to URL params:', err);
    }
  }

  // 2순위: 저장된 값 (filter.options.savedStartTime/savedEndTime)
  const datetimeFilter = filters.find((f) => f.type === 'datetime');

  // datetime 필터가 아예 없는 대시보드는 default도 적용하지 않음.
  // (불필요한 async import 방지 + 패널 쿼리에 datetime이 없는 설계 존중)
  if (!datetimeFilter) return undefined;

  const savedStartTime = datetimeFilter?.options?.savedStartTime as string | undefined;
  const savedEndTime = datetimeFilter?.options?.savedEndTime as string | undefined;

  if (savedStartTime && savedEndTime) {
    try {
      const { urlStringToDateTimeValue } = await import(
        '@pharos/shared/components/ui-extension'
      );
      const startTime = urlStringToDateTimeValue(savedStartTime);
      const endTime = urlStringToDateTimeValue(savedEndTime);
      if (startTime && endTime) {
        return { startTime, endTime };
      }
    } catch (err) {
      console.warn('[resolveInitialState] Invalid saved datetime:', err);
    }
  }

  // 3순위: 기본값 (1시간 전 ~ 지금)
  try {
    const { getRelativeRangeValue } = await import(
      '@pharos/shared/components/ui-extension'
    );
    const startTime = getRelativeRangeValue(1, 'Hours ago');
    const endTime = getRelativeRangeValue(true);
    return { startTime, endTime };
  } catch {
    return undefined;
  }
}

// ─── Step ────────────────────────────────────────────────────────────────────

function resolveStep(
  urlParams: URLSearchParams | undefined,
  filters: FilterConfig[],
): Step | undefined {
  // 1순위: URL 파라미터 (var_step으로 통일)
  const stepParam = urlParams?.get(URL_PARAM_STEP);
  if (stepParam) {
    const stepValue = getStepValue(stepParam);
    if (stepValue !== null) {
      return { step: stepParam , stepValue };
    }
  }

  // 2순위: 저장된 값
  const stepFilter = filters.find((f) => f.type === 'step');
  const savedValue = stepFilter?.options?.savedValue as string | undefined;
  if (savedValue) {
    const stepValue = getStepValue(savedValue);
    if (stepValue !== null) {
      return { step: savedValue , stepValue };
    }
  }

  // 3순위: 기본값은 StepFilter 컴포넌트에서 처리 (initValue 또는 '5m')
  return undefined;
}

// ─── RefreshInterval ─────────────────────────────────────────────────────────

function resolveRefreshInterval(
  urlParams: URLSearchParams | undefined,
  filters: FilterConfig[],
): number | false | undefined {
  // 1순위: URL 파라미터
  const refreshParam = urlParams?.get(URL_PARAM_REFRESH);
  if (refreshParam) {
    const interval = parseRefreshInterval(refreshParam);
    if (interval !== null) {
      return interval;
    }
  }

  // 2순위: 저장된 값
  const refreshFilter = filters.find((f) => f.type === 'refreshInterval');
  const savedValue = refreshFilter?.options?.savedValue as string | undefined;
  if (savedValue) {
    const interval = parseRefreshInterval(savedValue);
    if (interval !== null) {
      return interval;
    }
  }

  // 3순위: 기본값은 RefreshVariable 컴포넌트에서 처리 (initValue 또는 'Off')
  return undefined;
}
