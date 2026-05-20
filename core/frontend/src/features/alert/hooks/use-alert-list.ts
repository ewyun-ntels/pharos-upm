import { useList, useInfiniteList, type HttpError } from '@/lib/data-provider';
import { ALERT_RESOURCES, ALERT_PROVIDER_NAME } from '@providers/alert-provider';
import type { AlertRule, AlertValue } from '@pharos/shared/types/alert';
import type { UseQueryOptions } from '@tanstack/react-query';

/**
 * Alert Rule 목록 조회 Hook
 * 
 * @example
 * const { data, isLoading, refetch } = useAlertRuleList({
 *   filters: [{ field: 'alert_type', operator: 'eq', value: 'query' }],
 *   pagination: { currentPage: 1, pageSize: 20 }
 * });
 */
export function useAlertRuleList(config?: {
  filters?: any[];
  pagination?: { currentPage: number; pageSize: number };
  sorters?: any[];
  meta?: Record<string, any>;
  queryOptions?: Omit<UseQueryOptions<any, HttpError>, 'queryKey' | 'queryFn'>;
}) {
  return useList<AlertRule, HttpError>({
    resource: ALERT_RESOURCES.RULE,
    dataProviderName: ALERT_PROVIDER_NAME,
    filters: config?.filters,
    pagination: config?.pagination || { currentPage: 1, pageSize: 20, mode: 'server' },
    sorters: config?.sorters,
    meta: {
      ...config?.meta,
      detail: true, // Alert rules에는 detail이 있음
    },
    queryOptions: config?.queryOptions,
  });
}

/**
 * Alert Status 목록 조회 Hook
 * 
 * @example
 * const { data, isLoading } = useAlertStatusList({
 *   filters: [
 *     { field: 'status', operator: 'eq', value: 'alerting' },
 *     { field: 'severity', operator: 'eq', value: 'Critical' }
 *   ]
 * });
 */
export function useAlertStatusList(config?: {
  filters?: any[];
  pagination?: { currentPage: number; pageSize: number };
  sorters?: any[];
  queryOptions?: Omit<UseQueryOptions<any, HttpError>, 'queryKey' | 'queryFn'>;
}) {
  return useList<AlertValue, HttpError>({
    resource: ALERT_RESOURCES.STATUS,
    dataProviderName: ALERT_PROVIDER_NAME,
    filters: config?.filters,
    pagination: config?.pagination || { currentPage: 1, pageSize: 20, mode: 'server' },
    sorters: config?.sorters,
    queryOptions: config?.queryOptions,
  });
}

/**
 * Alert History 목록 조회 Hook (Infinite Scroll)
 * 
 * ⚠️ 주의: 백엔드에서 페이징이 아직 지원되지 않습니다.
 * 추후 백엔드 API가 페이징을 지원하면 useInfiniteList를 활성화할 예정입니다.
 * 
 * @todo 백엔드 페이징 지원 후 useInfiniteList로 전환
 * 
 * @example
 * const { data, isLoading } = useAlertHistoryList({
 *   name: 'CPU Alert',
 *   startTime: new Date('2025-10-13'),
 *   endTime: new Date('2025-10-14'),
 *   count: 100
 * });
 */
export function useAlertHistoryList(params?: {
  name?: string;
  startTime?: Date;
  endTime?: Date;
  count?: number;
  queryOptions?: Omit<UseQueryOptions<any, HttpError>, 'queryKey' | 'queryFn'>;
}) {
  // 현재는 일반 useList 사용 (페이징 미지원)
  return useList<AlertValue, HttpError>({
    resource: ALERT_RESOURCES.HIST,
    dataProviderName: ALERT_PROVIDER_NAME,
    filters: [
      params?.name && { field: 'name', operator: 'eq', value: params.name },
      params?.startTime && { field: 'startTime', operator: 'eq', value: params.startTime.toISOString() },
      params?.endTime && { field: 'endTime', operator: 'eq', value: params.endTime.toISOString() },
      params?.count && { field: 'count', operator: 'eq', value: params.count },
    ].filter(Boolean) as any[],
    pagination: {
      currentPage: 1,
      pageSize: params?.count || 100,
      mode: 'server',
    },
    queryOptions: params?.queryOptions,
  });
}

/**
 * Alert History Infinite Scroll Hook (미래 사용 예정)
 * 
 * ⚠️ 현재 사용 불가: 백엔드 페이징 지원 필요
 * 
 * @todo 백엔드 페이징 API 준비되면 이 Hook 사용
 * 
 * @example
 * const { data, isLoading, fetchNextPage, hasNextPage } = useAlertHistoryInfinite({
 *   name: 'CPU Alert'
 * });
 * 
 * // Scroll event
 * if (hasNextPage && !isLoading) {
 *   fetchNextPage();
 * }
 */
export function useAlertHistoryInfinite(params?: {
  name?: string;
  startTime?: Date;
  endTime?: Date;
}) {
  // TODO: 백엔드 페이징 지원 후 활성화
  return useInfiniteList<AlertValue, HttpError>({
    resource: ALERT_RESOURCES.HIST,
    dataProviderName: ALERT_PROVIDER_NAME,
    filters: [
      params?.name && { field: 'name', operator: 'eq', value: params.name },
      params?.startTime && { field: 'startTime', operator: 'eq', value: params.startTime.toISOString() },
      params?.endTime && { field: 'endTime', operator: 'eq', value: params.endTime.toISOString() },
    ].filter(Boolean) as any[],
    pagination: {
      currentPage: 1,
      pageSize: 50,
      mode: 'server',
    },
  });
}
