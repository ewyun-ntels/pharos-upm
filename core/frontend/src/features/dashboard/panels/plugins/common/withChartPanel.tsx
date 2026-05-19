import React, { useCallback, useEffect, useRef, useState } from 'react';
import type { ChartMetricData } from '@types';
import type { PanelProps } from '@pharos/core/panel-registry';
import type { DataProvider, ChartQuery } from '@pharos/shared/types/dashboard';

import { Alert, AlertDescription, AlertTitle } from '@pharos/shared/components/ui';
import { AlertTriangle, Info } from 'lucide-react';
import { LoadingIndicator, PanelLoadingBar } from '@pharos/shared/components/ui-extension';
import { useDashboardStore } from '@features/dashboard/hooks/use-dashboard-store';
import type { FilterValue } from '@features/dashboard/types/filter.types';
import { DASHBOARD_PROVIDER_NAME, DASHBOARD_RESOURCES } from '@providers/dashboard-provider';
import { ActionSchema } from '@pharos/shared/types/dashboard';

// ─── usePanelSize ─────────────────────────────────────────────────────────────

/**
 * 패널의 실제 렌더링 크기를 측정하는 커스텀 훅
 *
 * ⚠️ IMPORTANT: Callback ref 패턴을 사용해야 함 (useLayoutEffect + useRef 금지)
 *
 * 문제 시나리오 (useLayoutEffect + dependency [] 사용 시):
 * 1. 초기 마운트: hasValidQuery=false → Alert 렌더링 → divRef.current=null
 * 2. useLayoutEffect 한 번 실행 후 종료 (dependency [])
 * 3. "Run All Queries" 클릭 → hasValidQuery=true → 실제 div 렌더링
 * 4. 하지만 useLayoutEffect는 재실행 안 됨 (dependency []이므로)
 * 5. chartWidth, chartHeight가 영구히 0으로 유지 → timeseries 차트 안 그려짐
 *
 * 해결 방법 (Callback ref):
 * - useCallback으로 ref 함수를 만들면, DOM 노드가 실제로 마운트될 때마다 호출됨
 * - Alert → 실제 차트 전환 시에도 자동으로 치수 측정
 * - 컴포넌트가 조건부 렌더링되는 경우에도 안전하게 동작
 *
 * @returns {Object} divRef - DOM 노드에 attach할 callback ref
 * @returns {number} chartWidth - 측정된 너비
 * @returns {number} chartHeight - 측정된 높이
 */
export function usePanelSize() {
  const [chartWidth, setChartWidth] = useState(0);
  const [chartHeight, setChartHeight] = useState(0);
  const resizeObserverRef = useRef<ResizeObserver | null>(null);

  // ✅ Callback ref 패턴: DOM이 실제로 마운트될 때마다 호출됨
  // ⚠️ 절대 useLayoutEffect + dependency []로 변경하지 말것!
  const divRef = useCallback((node: HTMLDivElement | null) => {
    // 이전 observer 정리
    if (resizeObserverRef.current) {
      resizeObserverRef.current.disconnect();
      resizeObserverRef.current = null;
    }

    if (!node) {
      // node가 null이면 unmount된 것 → 치수를 0으로 리셋
      setChartWidth(0);
      setChartHeight(0);
      return;
    }

    // 초기 치수 측정 (paint 이전에 실행하려면 동기적으로)
    const initialWidth = node.offsetWidth;
    const initialHeight = node.offsetHeight;

    setChartWidth(initialWidth);
    setChartHeight(initialHeight);

    // ResizeObserver 설정
    const resizeObserver = new ResizeObserver(() => {
      const newWidth = node.offsetWidth;
      const newHeight = node.offsetHeight;

      setChartWidth(newWidth);
      setChartHeight(newHeight);
    });

    resizeObserver.observe(node);
    resizeObserverRef.current = resizeObserver;
  }, []);

  return { divRef, chartWidth, chartHeight };
}

// ─── 내부 타입 ────────────────────────────────────────────────────────────────

export type GuaranteedDataProvider = Omit<DataProvider, 'chartQuery'> & { chartQuery: ChartQuery[] };

type ChartDataWrapperProps<TData> = Pick<
  PanelProps<unknown>,
  'pluginContext' | 'dataProvider' | 'args' | 'refetchInterval' | 'dashboardId' | 'id' | 'kind' | 'permission'
> & {
  onLoadingChange?: (isLoading: boolean) => void;
  children: (chartData: TData | undefined, setFilter: (id: string, value: FilterValue) => void, getFilter: (id: string) => FilterValue | undefined) => React.ReactNode;
};

// ─── ChartDataWrapper (내부 구현) ────────────────────────────────────────────

function parseDbError(message: string): { code: string | null; body: string } {
  // Pattern: "code: 62, message: Syntax error..."
  const match = message.match(/^code:\s*(\d+),\s*message:\s*([\s\S]*)/);
  if (match) {
    return { 
      code: match[1], 
      body: match[2].trim() 
    };
  }
  
  // No code found, return full message
  return { 
    code: null, 
    body: message 
  };
}

export function ChartDataWrapper<TData = ChartMetricData>({
  pluginContext,
  dataProvider,
  args,
  refetchInterval,
  dashboardId,
  id,
  kind,
  permission,
  onLoadingChange,
  children,
}: ChartDataWrapperProps<TData>) {
  const setFilter = useDashboardStore((state) => state.setFilter);
  const getFilter = useDashboardStore((state) => state.getFilter);
  const usePanelMode = permission === ActionSchema.enum.viewer && Boolean(id && kind);
  const queryRequests = usePanelMode
    ? [{
      dashboardId: dashboardId || 'preview',
      queryName: 'query-0',
      id,
      kind,
      query: '',
      datasourceName: '',
    }]
    : pluginContext?.hooks?.toChartQueryRequests?.(dataProvider?.chartQuery || [], dashboardId || 'preview') || [];

  // ✅ pluginContext를 통해 useDashboardQuery 사용 (플러그인 확장성)
  const { data: chartData, isLoading, isFetching, isError, error } = pluginContext?.hooks?.useDashboardQuery<TData>?.({
    id: id || `panel-preview-${dashboardId || 'temp'}`,
    dashboardId: dashboardId || 'preview',
    queries: queryRequests,
    resource: (dataProvider?.resource as 'timeseries' | 'raw') || DASHBOARD_RESOURCES.TIMESERIES,
    dataProviderName: dataProvider?.dataProviderName,
    args,
    refetchInterval,
  }) || { data: undefined, isLoading: false, isFetching: false, isError: false, error: undefined };

  // ✅ isFetching: 캐시 있어도 백그라운드 재조회 시 true (isLoading은 캐시 없을 때만 true)
  useEffect(() => {
    onLoadingChange?.(isFetching);
  }, [isFetching, onLoadingChange]);

  if (!pluginContext?.hooks?.useDashboardQuery) {
    return (
      <div className="flex justify-center items-center w-full h-full p-3">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Plugin Error</AlertTitle>
          <AlertDescription>useDashboardQuery not available in pluginContext</AlertDescription>
        </Alert>
      </div>
    );
  }

  if (isError && error) {
    if (error.statusCode === 408) {
      return (
        <div className="flex flex-col h-full w-full overflow-hidden p-3">
          <Alert variant="destructive">
            <AlertTriangle className="h-4 w-4" />
            <AlertTitle>Query Timeout</AlertTitle>
            <AlertDescription>쿼리 실행 시간이 초과되었습니다. 잠시 후 다시 시도해주세요.</AlertDescription>
          </Alert>
        </div>
      );
    }

    const rawMessage = error?.message || 'Something went wrong!';
    const { code, body } = parseDbError(rawMessage);
    return (
      <div className="flex flex-col h-full w-full overflow-hidden p-3">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Query Error{code ? ` (Code ${code})` : ''}</AlertTitle>
          <AlertDescription>
            <pre className="text-xs whitespace-pre-wrap font-mono leading-relaxed wrap-break-word mt-1">
              {body}
            </pre>
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  if (isLoading) {
    return <LoadingIndicator />;
  }

  return <>{children(chartData, setFilter, getFilter)}</>;
}

// ─── PanelContentProps (공개 타입) ────────────────────────────────────────────

/**
 * withChartPanel / withCardChartPanel로 감싼 컴포넌트가 받는 props.
 * - pluginContext: HOC가 내부적으로 처리하므로 제외
 * - dataProvider: chartQuery 존재 보장
 * - chartData: HOC가 데이터 fetching 후 주입
 * - chartWidth / chartHeight: HOC의 ResizeObserver가 측정하여 주입
 * - setFilter / getFilter: Optional - 인터랙티브 차트(클릭 등)에만 사용
 */
export type PanelContentProps<T, TData = ChartMetricData> = Omit<PanelProps<T>, 'pluginContext' | 'dataProvider'> & {
  dataProvider: GuaranteedDataProvider;
  chartData: TData | undefined;
  chartWidth: number;
  chartHeight: number;
  // ✅ Optional: 인터랙티브 차트만 사용 (BarChart onClickVariable 등)
  setFilter?: (id: string, value: FilterValue) => void;
  getFilter?: (id: string) => FilterValue | undefined;
};

// ─── withChartPanel HOC (공개) ────────────────────────────────────────────────

/**
 * withChartPanel HOC
 *
 * Card 없이 패널 전체 영역을 사용하는 패널용 HOC.
 * 마크다운처럼 Card chrome 없이 전체 공간이 필요한 경우에 사용합니다.
 *
 * 다음을 자동 처리합니다:
 * - dataProvider 체크 및 정규화
 * - 데이터 fetching + 로딩/에러 UI (ChartDataWrapper)
 * - ResizeObserver로 chartWidth / chartHeight 측정 및 주입
 *
 * 사용 예:
 *   component: withChartPanel(MyContent)          // Card 없이 전체 영역
 *   component: withCardChartPanel(MyContent)      // Card 포함, 차트 데이터
 */
export function withChartPanel<T, TData = ChartMetricData>(
  Component: React.ComponentType<PanelContentProps<T, TData>>,
): React.ComponentType<PanelProps<T>> {
  function WithChartPanel(props: PanelProps<T>) {
    const { divRef, chartWidth, chartHeight } = usePanelSize();
    // hooks는 early return 이전에 선언해야 함
    const [isPanelLoading, setIsPanelLoading] = useState(false);
    const dataProvider = props.dataProvider;
    const usePanelMode = props.permission === ActionSchema.enum.viewer && Boolean(props.id && props.kind);

    // ✅ Check if chartQuery exists and is not empty
    if (!usePanelMode && (!dataProvider?.chartQuery || dataProvider.chartQuery.length === 0)) {
      return (
        <div className="flex justify-center items-center w-full h-full p-3">
          <Alert>
            <Info className="h-4 w-4" />
            <AlertTitle>No Data Source</AlertTitle>
            <AlertDescription>No data source configured for this panel</AlertDescription>
          </Alert>
        </div>
      );
    }

    // ✅ Check if all queries are empty (newly created panels)
    const hasValidQuery = dataProvider?.chartQuery?.some(q => q.query && q.query.trim() !== '') ?? false;
    if (!usePanelMode && !hasValidQuery) {
      return (
        <div className="flex justify-center items-center w-full h-full p-3">
          <Alert>
            <Info className="h-4 w-4" />
            <AlertTitle>No Query Configured</AlertTitle>
            <AlertDescription>Click to edit this panel and add a query</AlertDescription>
          </Alert>
        </div>
      );
    }

    const guaranteedProvider = {
      chartQuery: dataProvider?.chartQuery || [],
      dataProviderName: dataProvider?.dataProviderName || DASHBOARD_PROVIDER_NAME,
      resource: dataProvider?.resource || DASHBOARD_RESOURCES.TIMESERIES,
    } as GuaranteedDataProvider;

    return (
      <div ref={divRef} className="w-full h-full relative">
        {isPanelLoading && <PanelLoadingBar />}
        <ChartDataWrapper<TData>
          pluginContext={props.pluginContext}
          dataProvider={guaranteedProvider}
          args={props.args || new Map()}
          refetchInterval={props.refetchInterval}
          dashboardId={props.dashboardId}
          id={props.id}
          kind={props.kind}
          permission={props.permission}
          onLoadingChange={setIsPanelLoading}
        >
          {(chartData, setFilter, getFilter) => (
            <Component
              {...props}
              dataProvider={guaranteedProvider}
              chartData={chartData}
              chartWidth={chartWidth}
              chartHeight={chartHeight}
              setFilter={setFilter}
              getFilter={getFilter}
            />
          )}
        </ChartDataWrapper>
      </div>
    );
  }

  WithChartPanel.displayName = `WithChartPanel(${Component.displayName || Component.name || 'Component'})`;

  return WithChartPanel;
}
