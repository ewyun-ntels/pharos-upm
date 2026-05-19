/**
 * Dashboard URL Sync (Zustand subscribe 기반)
 *
 * Store 상태 변경을 감지하여 URL을 자동 동기화한다.
 * 이전에 각 액션(setFilter, setDatetime, setStep) 내부에 인라인으로
 * 존재하던 URL 쓰기 코드를 subscribe 한 곳으로 통합.
 *
 * skipUrlSync 플래그가 제거되고, 대신 내부 suppressSync 플래그로
 * 초기 로드 시에만 동기화를 억제한다.
 */
import { DashboardUrlManager } from './dashboard-url-manager';
import { formatRefreshInterval } from '../../utils/datetime-step-utils';
import type { useDashboardStore as UseDashboardStoreType } from '../use-dashboard-store';

type DashboardStoreInstance = typeof UseDashboardStoreType;

/** URL 동기화 억제 플래그 (초기 로드 시 URL → Store 방향일 때 true) */
let suppressSync = false;

/**
 * URL 동기화를 일시적으로 억제한다. (_loadDashboard의 초기 로드 시 사용)
 * URL → Store 방향으로 값을 적용하는 동안 Store → URL subscribe를 억제하여
 * 무한 루프를 방지한다.
 */
export async function withSuppressedUrlSyncAsync<T>(fn: () => Promise<T>): Promise<T> {
  suppressSync = true;
  try {
    return await fn();
  } finally {
    queueMicrotask(() => {
      suppressSync = false;
    });
  }
}

/**
 * 현재 URL이 대시보드 페이지인지 확인.
 * 대시보드가 아닌 페이지에서 store 변경이 발생해도 URL을 오염시키지 않도록 한다.
 */
function isDashboardPath(): boolean {
  return window.location.pathname.includes('/dashboards/');
}

/**
 * Zustand store에 URL 동기화 subscription을 등록한다.
 * 앱 초기화 시 1회만 호출.
 */
export function setupUrlSync(store: DashboardStoreInstance) {
  // 1) filterMetas 변경 → URL 동기화
  store.subscribe(
    (state) => state.filterState.filterMetas,
    (filterMetas) => {
      if (suppressSync || !isDashboardPath()) return;
      DashboardUrlManager.syncFilters(filterMetas);
    },
  );

  // 2) datetime 변경 → URL 동기화
  store.subscribe(
    (state) => state.filterState.datetime,
    (datetime) => {
      if (suppressSync || !datetime || !isDashboardPath()) return;
      DashboardUrlManager.syncDatetime(datetime);
    },
  );

  // 3) step 변경 → URL 동기화
  store.subscribe(
    (state) => state.filterState.step,
    (step) => {
      if (suppressSync || !step || !isDashboardPath()) return;
      DashboardUrlManager.syncStep(step);
    },
  );

  // 4) refreshInterval 변경 → URL 동기화
  store.subscribe(
    (state) => state.filterState.refreshInterval,
    (refreshInterval) => {
      if (suppressSync || !isDashboardPath()) return;
      if (refreshInterval === undefined) return;

      // refreshInterval을 URL 형식으로 변환
      DashboardUrlManager.syncRefreshInterval(formatRefreshInterval(refreshInterval));
    },
  );
}
