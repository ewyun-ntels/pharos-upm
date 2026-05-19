import { useList, useOne, type HttpError } from '@/lib/data-provider';
import { NOTIFICATION_RESOURCES, NOTIFICATION_PROVIDER_NAME } from '@providers/notification-provider';
import type { NotificationRule } from '@pharos/shared/types/notification';

/**
 * Notification Rule 목록 조회 Hook
 *
 * @example
 * const { data, isLoading, refetch } = useNotificationRuleList({
 *   pagination: { currentPage: 1, pageSize: 20 }
 * });
 */
export function useNotificationRuleList(config?: {
  filters?: any[];
  pagination?: { currentPage: number; pageSize: number };
  sorters?: any[];
  meta?: Record<string, any>;
}) {
  return useList<NotificationRule, HttpError>({
    resource: NOTIFICATION_RESOURCES.RULE,
    dataProviderName: NOTIFICATION_PROVIDER_NAME,
    filters: config?.filters,
    pagination: config?.pagination || { currentPage: 1, pageSize: 20, mode: 'server' },
    sorters: config?.sorters,
    meta: config?.meta,
  });
}

/**
 * Notification Rule 단일 조회 Hook
 *
 * @example
 * const { data, isLoading } = useNotificationRule('rule-123');
 */
export function useNotificationRule(id?: string) {
  return useOne<NotificationRule, HttpError>({
    resource: NOTIFICATION_RESOURCES.RULE,
    dataProviderName: NOTIFICATION_PROVIDER_NAME,
    id: id || '',
    queryOptions: {
      enabled: !!id,
    },
  });
}
