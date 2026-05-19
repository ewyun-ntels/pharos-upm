import { useOne } from '@/lib/data-provider';
import { DASHBOARD_PROVIDER_NAME, DASHBOARD_RESOURCES } from '@providers/dashboard-provider';
import type { Kind, QueryInspectResponse } from '@pharos/shared/types/dashboard';

interface UseQueryInspectorParams {
  dashboardId: string;
  // Panel mode (for saved panels in dashboard view)
  panelId?: string;
  kind?: Kind;
  // Run mode (for unsaved panels in panel editor)
  datasourceName?: string;
  query?: string;
  variables?: Record<string, any>;
  enabled?: boolean;
}

/**
 * Query Inspector hook - fetches query execution details with inspect info
 * 
 * Uses panel mode to fetch query execution details from the backend.
 * The backend will resolve panel data (query, datasourceName) from the panel ID.
 */
export const useQueryInspector = ({
  dashboardId,
  panelId,
  kind,
  datasourceName,
  query,
  variables,
  enabled = true,
}: UseQueryInspectorParams) => {
  // ✅ Convert variables to Map with proper number formatting
  // Prevent scientific notation for large numbers (timestamps)
  const argsMap = new Map<string, string | number>();
  if (variables) {
    Object.entries(variables).forEach(([key, value]) => {
      if (typeof value === 'number') {
        // Keep as number (no scientific notation issues with Map)
        argsMap.set(key, value);
      } else if (value != null) {
        argsMap.set(key, String(value));
      }
    });
  }

  // ✅ Create unique query key that includes query content
  // This ensures cache invalidation when query changes
  const queryKey = [
    DASHBOARD_RESOURCES.RAW,
    dashboardId,
    panelId,
    datasourceName,
    query, // Include query content in cache key
    variables,
  ].filter(Boolean);

  return useOne<QueryInspectResponse>({
    resource: DASHBOARD_RESOURCES.RAW,
    id: dashboardId,
    queryOptions: {
      enabled,
      // 🔄 Enable refetch on window focus and always get fresh data
      refetchOnWindowFocus: true,
      staleTime: 0, // Always consider data stale - refetch when enabled changes
      queryKey, // 🔑 Custom query key including query content
      // 🚫 Disable retry for inspector - we want to see errors immediately
      retry: false,
      retryDelay: 0,
      // ⚡ Get error feedback immediately without delay
      refetchOnMount: 'always',
    },
    meta: {
      variables: {
        value: [
          {
            dashboardId,
            // Panel mode: use panelId + kind
            ...(panelId && kind ? { id: panelId, kind } : {}),
            // Run mode: use datasourceName + query
            ...(datasourceName && query ? { datasourceName, query } : {}),
            args: argsMap,
          }
        ]
      },
      inspect: true, // 🔍 Enable inspect mode
    },
    dataProviderName: DASHBOARD_PROVIDER_NAME,
  });
};
