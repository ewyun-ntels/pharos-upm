/**
 * Dashboard URL Manager
 *
 * URL ↔ Store 동기화의 단일 진입점(Single Point of Control).
 * 이전에 setFilter, setDatetime, setStep, StepFilter, RefreshFilter 등
 * 4곳 이상에서 분산된 URL 쓰기를 이 모듈 하나로 통합한다.
 *
 * 설계 원칙:
 * - URL 읽기/쓰기는 반드시 이 모듈을 통해서만 수행
 * - requestAnimationFrame으로 같은 프레임 내 여러 변경을 batch 처리
 * - window.history.replaceState만 사용 (react-router navigate 사용 금지)
 */
import {
  dateTimeValueToUrlString,
} from '@pharos/shared/components/ui-extension';
import type { FilterMeta } from '../use-dashboard-store';
import type { DateTime, Step } from '../../types/filter.types';
import {
  URL_PARAM_FROM,
  URL_PARAM_TO,
  URL_PARAM_STEP,
  URL_PARAM_REFRESH,
  getFilterParamName,
} from './url-param-constants';

// ─── Batch 처리를 위한 내부 상태 ─────────────────────────────────────────────

let pendingUpdate: (() => void) | null = null;
let rafId: number | null = null;

/**
 * 같은 프레임 내 여러 URL 변경을 하나의 replaceState로 합친다.
 * queueMicrotask 대신 rAF를 사용하여 렌더 사이클과 맞춘다.
 */
function scheduleUrlUpdate(updateFn: () => void) {
  pendingUpdate = updateFn;

  if (rafId === null) {
    rafId = requestAnimationFrame(() => {
      rafId = null;
      if (pendingUpdate) {
        pendingUpdate();
        pendingUpdate = null;
      }
    });
  }
}

// ─── URL 쓰기 (Store → URL) ─────────────────────────────────────────────────

function replaceUrl(params: URLSearchParams) {
  const newUrl = params.toString()
    ? `${window.location.pathname}?${params.toString()}`
    : window.location.pathname;
  const currentUrl = `${window.location.pathname}${window.location.search}`;
  if (newUrl !== currentUrl) {
    window.history.replaceState({}, '', newUrl);
  }
}

/** 현재 URL params의 snapshot을 반환 */
function currentParams(): URLSearchParams {
  return new URLSearchParams(window.location.search);
}

// ─── Public API ──────────────────────────────────────────────────────────────

export const DashboardUrlManager = {
  /**
   * filterMetas 전체를 URL에 동기화 (batch)
   * setFilter() 후 subscribe에서 호출
   */
  syncFilters(filterMetas: Map<string, FilterMeta>) {
    if (typeof window === 'undefined') return;

    scheduleUrlUpdate(() => {
      const params = currentParams();

      filterMetas.forEach((meta, key) => {
        const paramName = getFilterParamName(key);
        if (meta.value !== undefined) {
          const paramValue = Array.isArray(meta.value)
            ? meta.value.join(',')
            : String(meta.value);
          params.set(paramName, paramValue);
        } else {
          params.delete(paramName);
        }
      });

      replaceUrl(params);
    });
  },

  /**
   * datetime을 URL에 동기화 (즉시)
   * setDatetime() 후 subscribe에서 호출
   */
  syncDatetime(datetime: DateTime) {
    if (typeof window === 'undefined') return;

    const params = currentParams();
    params.set(URL_PARAM_FROM, dateTimeValueToUrlString(datetime.startTime));
    params.set(URL_PARAM_TO, dateTimeValueToUrlString(datetime.endTime));
    replaceUrl(params);
  },

  /**
   * step을 URL에 동기화 (즉시)
   * setStep() 후 subscribe에서 호출
   */
  syncStep(step: Step) {
    if (typeof window === 'undefined') return;

    const params = currentParams();
    params.set(URL_PARAM_STEP, step.step);
    replaceUrl(params);
  },

  /**
   * refreshInterval을 URL에 동기화 (즉시)
   */
  syncRefreshInterval(value: string) {
    if (typeof window === 'undefined') return;

    const params = currentParams();
    params.set(URL_PARAM_REFRESH, value);
    replaceUrl(params);
  },

  /**
   * 대시보드 언로드 시 대시보드 관련 파라미터 모두 제거
   */
  clearDashboardParams() {
    if (typeof window === 'undefined') return;

    // pending update 취소
    if (rafId !== null) {
      cancelAnimationFrame(rafId);
      rafId = null;
      pendingUpdate = null;
    }
  },
} as const;
