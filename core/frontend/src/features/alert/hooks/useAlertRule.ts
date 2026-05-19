import {useOne} from '@/lib/data-provider';
import {ALERT_RESOURCES, ALERT_PROVIDER_NAME} from '@providers/alert-provider/types';
import {AlertRuleSchema, type AlertRule} from '@pharos/shared/types/alert';
import {useMemo} from 'react';

interface UseAlertRuleOptions {
  id: string | null;
  enabled?: boolean;
}

interface UseAlertRuleResult {
  data: AlertRule | null;
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
  validationError: Error | null;
}

/**
 * Fetch and validate alert rule data
 * 
 * @example
 * ```tsx
 * const {data, isLoading, isError, validationError} = useAlertRule({
 *   id: ruleId,
 *   enabled: !!ruleId
 * });
 * 
 * if (validationError) {
 *   console.error('Invalid data format:', validationError);
 * }
 * ```
 */
export function useAlertRule({id, enabled = true}: UseAlertRuleOptions): UseAlertRuleResult {
  const {query: {data: rawData, isLoading, isError, error}} = useOne({
    resource: ALERT_RESOURCES.RULE,
    id: id || '',
    queryOptions: {
      enabled: enabled && !!id,
    },
    meta: {
      dataProviderName: ALERT_PROVIDER_NAME,
    },
  });

  const {data, validationError} = useMemo(() => {
    if (!rawData) {
      return {data: null, validationError: null};
    }

    const responseData = rawData.data;
    
    if (!responseData) {
      return {data: null, validationError: null};
    }

    try {
      // Validate with Zod schema
      const validatedData = AlertRuleSchema.parse(responseData);
      return {data: validatedData, validationError: null};
    } catch (err) {
      return {
        data: null,
        validationError: err instanceof Error ? err : new Error('Validation failed'),
      };
    }
  }, [rawData]);

  return {
    data,
    isLoading,
    isError,
    error: error as Error | null,
    validationError,
  };
}
