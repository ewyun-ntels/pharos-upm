/**
 * useDashboardQuery - Simple wrapper for QueryPipeline
 */

import {useEffect, useState, useRef, useMemo} from 'react';
import {QueryPipeline} from './QueryPipeline';
import {useDashboardStore} from './use-dashboard-store';
import {getDependencyGraph} from './slices/types';
import {registerOnLogout} from '@pharos/shared/features/auth';
import type {ChartQuery} from '@pharos/shared/types/dashboard';
import type {ChartQueryRequest} from '@types';

/**
 * SWR (Stale-While-Revalidate) 캐시 — 패널 쿼리 전용
 *
 * - 대상: useFilterMeta=false 경로 (패널 쿼리)
 * - 목적: 에디터 ↔ 대시보드 이동 시 스피너 없이 이전 데이터 즉시 표시
 * - 동작: mount 시 캐시 데이터로 isLoading=false 시작 → 백그라운드 재조회 → 완료 시 교체
 * - 캐시 키: panel id (UUID, per-panel 고유)
 * - 수명: 로그아웃 시 전체 clear
 *
 * 메모리: 쿼리 결과(수백KB~수MB)를 저장하므로 MAX_PANEL_SWR_CACHE로 항목 수 제한 (LRU)
 */
const MAX_PANEL_SWR_CACHE = 30;
const panelSWRCache = new Map<string, unknown>();

function setPanelSWRCache(key: string, value: unknown) {
  // LRU: 이미 있으면 삭제 후 들락말에 다시 삽입 (가장 최근 사용로 갱신)
  if (panelSWRCache.has(key)) panelSWRCache.delete(key);
  panelSWRCache.set(key, value);
  // 크기 초과 시 가장 오래된 항목 제거
  if (panelSWRCache.size > MAX_PANEL_SWR_CACHE) {
    panelSWRCache.delete(panelSWRCache.keys().next().value!);
  }
}

registerOnLogout(() => {
  panelSWRCache.clear();
});

// ─── 대시보드 레벨 쿼리 취소 ────────────────────────────────────────────────
// 대시보드 페이지 unmount 시 모든 pending 쿼리를 즉시 abort하여
// 브라우저 HTTP 연결 풀을 해제한다. (네비게이션 지연 방지)
//
// Set 기반: 각 useDashboardQuery가 자신의 AbortController를 등록하고,
// 개별 abort 시 자동 제거된다. useEffect 실행 순서(자식→부모)에 무관하게 동작.

const activeQueryControllers = new Set<AbortController>();

/**
 * 쿼리의 AbortController를 등록한다.
 * useDashboardQuery의 useEffect 내부에서 호출.
 * abort 시 자동으로 Set에서 제거된다.
 */
function registerQueryController(ctrl: AbortController): void {
  activeQueryControllers.add(ctrl);
  ctrl.signal.addEventListener('abort', () => activeQueryControllers.delete(ctrl), { once: true });
}

/**
 * 대시보드 페이지 unmount 시 호출.
 * 모든 pending 쿼리를 즉시 취소하여 HTTP 연결 해제.
 */
export function unmountDashboardQueries(): void {
  activeQueryControllers.forEach((ctrl) => ctrl.abort());
  activeQueryControllers.clear();
}

export function toChartQueryRequests(
  queries: ChartQuery[],
  dashboardId: string,
): ChartQueryRequest[] {
  return queries.map((q, idx) => ({
    dashboardId,
    queryName: `query-${idx}`,
    label: q.label || '',
    query: q.query || '',
    datasourceName: q.datasourceName || '',
    forceRefresh: q.forceRefresh,
  }));
}

// Helper: Extract value from item using valueKey
function extractValue(item: unknown, valueKey: string): unknown {
  if (typeof item === 'object' && item !== null && valueKey in item) {
    return (item as Record<string, unknown>)[valueKey];
  }
  return item;
}

// Auto-select first option from results
function autoSelectFirst(
  id: string,
  results: unknown[],
  currentValue: unknown,
  valueKey: string,
  setValue: (v: string | string[]) => void,
) {
  if (!Array.isArray(results) || results.length === 0) return;

  // Skip if current value is valid
  if (currentValue != null) {
    const valid = results.some((item) => {
      const v = extractValue(item, valueKey);
      return Array.isArray(currentValue) ? currentValue.includes(v) : currentValue === v;
    });
    if (valid) {
      return;
    }
  }

  // Select first
  const val = extractValue(results[0], valueKey);
  const strVal = String(val);
  setValue(strVal);
}

export interface UseDashboardQueryOptions {
  id: string;
  dashboardId: string;
  queries: ChartQueryRequest[];
  resource: 'timeseries' | 'raw';
  enabled?: boolean;
  useFilterMeta?: boolean;
  useDependencyGraph?: boolean;
  autoSelectFirst?: boolean;
  valueKey?: string;
  labelKey?: string; // For filter options
  dataProviderName?: string; // For panel queries
  refetchInterval?: number | false; // For auto-refresh (ignored for now)
  args?: Map<string, string | number | (() => number)>; // Panel에서 전달하는 추가 args
}

export type QueryError = Error & { statusCode?: number };

export interface UseDashboardQueryResult<TData = unknown> {
  data: TData | undefined;
  isLoading: boolean;
  isFetching: boolean;
  isSuccess: boolean;
  isError: boolean;
  error: QueryError | undefined;
  value?: string | string[];
  setValue?: (value: string | string[]) => void;
}

export function useDashboardQuery<TData = unknown>(
  opts: UseDashboardQueryOptions,
): UseDashboardQueryResult<TData> {
  const {
    id,
    dashboardId,
    queries,
    resource,
    enabled = true,
    useFilterMeta = false,
    useDependencyGraph = false,
    autoSelectFirst: shouldAutoSelect = false,
    valueKey = 'value',
    args = new Map(),
  } = opts;

  // mount 시 캐시된 결과가 있으면 초기값으로 사용 (remount re-fetch 방지)
  const cachedMeta = useFilterMeta
    ? useDashboardStore.getState().filterState?.filterMetas?.get(id)
    : undefined;
  const hasCachedResult = cachedMeta?.isQuerySuccess && cachedMeta?.data !== undefined;

  // ✅ SWR: 패널 쿼리 캐시 (useFilterMeta=false 경로)
  // hasCachedResult가 false일 때(=패널 쿼리)만 SWR 캐시 확인
  const swrCachedData = !hasCachedResult && !useFilterMeta ? panelSWRCache.get(id) : undefined;
  const hasSWRCache = swrCachedData !== undefined;

  const [data, setData] = useState<TData>(
    hasCachedResult
      ? (cachedMeta!.data as TData)
      : hasSWRCache
        ? (swrCachedData as TData)
        : (undefined as TData),
  );
  const [isLoading, setIsLoading] = useState(
    // SWR 캐시가 있으면 isLoading=false로 시작 → 스피너 없이 이전 데이터 즉시 표시
    hasCachedResult || hasSWRCache ? false : enabled && queries.length > 0,
  );
  const [isFetching, setIsFetching] = useState(false);
  const [isSuccess, setIsSuccess] = useState(
    hasCachedResult ? true : (cachedMeta?.isQuerySuccess ?? false),
  );
  const [isError, setIsError] = useState(false);
  const [error, setError] = useState<Error>();
  const abortRef = useRef<AbortController | null>(null);

  const graph = getDependencyGraph();
  const startTime = useDashboardStore((s) => s.filterState?.datetime?.startTime);
  const endTime = useDashboardStore((s) => s.filterState?.datetime?.endTime);
  const refreshCount = useDashboardStore((s) => s.filterState?.refreshCount ?? 0);
  const meta = useDashboardStore((s) =>
    useFilterMeta ? s.filterState?.filterMetas?.get(id) : undefined,
  );

  // ✅ queries를 stringfy해서 변경 감지 (forceRefresh 포함)
  const queriesKey = useMemo(() => JSON.stringify(queries), [queries]);

  // 의존하는 변수들의 상태 추적 (ready 상태 변경 감지용)
  const depsReadyKey = useDashboardStore((s) => {
    if (!s.filterState?.filterMetas) return '';
    const metas = s.filterState.filterMetas;

    // 쿼리에서 변수 추출
    try {
      const p = QueryPipeline.for(id, queries).with({
        graph: getDependencyGraph(),
        filterMetas: metas,
        startTime: s.filterState?.datetime?.startTime,
        endTime: s.filterState?.datetime?.endTime,
        enabled,
        dashboardId,
        resource,
      });
      return p.deps.requiredVars
        .map((v) => {
          const m = metas.get(v);
          return `${v}:${m?.value}:${m?.isFetching}:${m?.isQuerySuccess}`;
        })
        .join('|');
    } catch {
      return '';
    }
  });

  const setValue = (v: string | string[]) => {
    if (!useFilterMeta) return;
    useDashboardStore.getState().setFilter(id, v);
  };

  // Filter 초기 등록
  useEffect(() => {
    if (!useFilterMeta || !queries[0]) return;
    const store = useDashboardStore.getState();
    const exists = store.filterState?.filterMetas?.get(id);
    if (!exists) {
      store.setFilterMeta(id, {
        id,
        query: queries[0].query || '',
        datasourceName: queries[0].datasourceName || '',
      });
    }
  }, [id, queries, useFilterMeta]);

  // Pipeline 실행
  useEffect(() => {
    abortRef.current?.abort();
    const ctrl = new AbortController();
    abortRef.current = ctrl;

    // Set에 등록: 대시보드 unmount 시 일괄 abort 대상
    registerQueryController(ctrl);

    const metas = useDashboardStore.getState().filterState?.filterMetas ?? new Map();

    // ✅ 캐시 히트: 같은 refreshCount로 이미 성공한 쿼리면 re-fetch 스킵 (Editor remount 등)
    if (useFilterMeta) {
      const currentMeta = metas.get(id);
      if (
        currentMeta?.isQuerySuccess &&
        currentMeta?.data !== undefined &&
        currentMeta?.lastRefreshCount === refreshCount
      ) {
        return;
      }
    }

    const p = QueryPipeline.for(id, queries).with({
      graph,
      filterMetas: metas,
      startTime,
      endTime,
      enabled,
      dashboardId,
      resource,
      panelArgs: args,
    });

    if (!p.ready) {
      // ✅ Waiting 상태: 로딩 표시 유지 (dependencies 기다리는 중)
      setIsFetching(false);
      if (useFilterMeta) useDashboardStore.getState().setFilterMeta(id, {isFetching: false});
      return;
    }

    // ✅ SWR: 캐시된 데이터가 있으면 isLoading은 올리지 않음 (스피너 없이 백그라운드 재조회)
    //    캐시 없는 최초 조회만 isLoading=true (스피너 표시)
    const hasPanelSWRCache = !useFilterMeta && panelSWRCache.has(id);
    if (!hasPanelSWRCache) {
      setIsLoading(true);
    }
    setIsFetching(true);
    setIsError(false);
    if (useFilterMeta) useDashboardStore.getState().setFilterMeta(id, {isFetching: true});

    p.execute<TData>(ctrl.signal)
      .then((result) => {
        if (ctrl.signal.aborted) return;

        setData(result);
        setIsSuccess(true);
        setIsError(false);
        setError(undefined);
        setIsLoading(false);
        setIsFetching(false);

        // ✅ SWR 캐시 갱신: 패널 쿼리 성공 시 캐시에 저장
        if (!useFilterMeta) {
          setPanelSWRCache(id, result);
        }

        if (useFilterMeta) {
          useDashboardStore.getState().setFilterMeta(id, {
            isFetching: false,
            isQuerySuccess: true,
            data: result,
            lastRefreshCount: refreshCount,
          });
          if (shouldAutoSelect && Array.isArray(result)) {
            autoSelectFirst(id, result, meta?.value, valueKey, setValue);
          }
        }
      })
      .catch((err) => {
        if (ctrl.signal.aborted) return;

        setError(Object.assign(new Error(err?.message || 'Error'), { statusCode: err?.statusCode }));
        setIsError(true);
        setIsSuccess(false);
        setIsLoading(false);
        setIsFetching(false);

        if (useFilterMeta)
          useDashboardStore
            .getState()
            .setFilterMeta(id, {isFetching: false, isQuerySuccess: false});
      });

    return () => ctrl.abort();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    id,
    graph,
    startTime,
    endTime,
    enabled,
    dashboardId,
    resource,
    useDependencyGraph,
    useFilterMeta,
    shouldAutoSelect,
    valueKey,
    refreshCount,
    depsReadyKey,
    queriesKey,
  ]);

  return {data, isLoading, isFetching, isSuccess, isError, error, value: meta?.value, setValue};
}
