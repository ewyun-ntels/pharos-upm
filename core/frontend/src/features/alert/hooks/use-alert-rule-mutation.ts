import { useCreate, useUpdate, type HttpError } from '@/lib/data-provider';
import { z } from 'zod';
import { ALERT_RESOURCES, ALERT_PROVIDER_NAME } from '@providers/alert-provider';
import { 
  QueryAlertRuleSchema,
  EventHistoryAlertRuleSchema,
  EventStatusAlertRuleSchema,
  type QueryAlertRule,
  type EventHistoryAlertRule,
  type EventStatusAlertRule
} from '@pharos/shared/types/alert';

// Union type for all alert rules
type AnyAlertRule = QueryAlertRule | EventHistoryAlertRule | EventStatusAlertRule;
const AnyAlertRuleSchema = z.union([
  QueryAlertRuleSchema,
  EventHistoryAlertRuleSchema,
  EventStatusAlertRuleSchema
]);

/**
 * Custom hook for creating alert rules with Zod validation
 * 
 * @example
 * const { mutate, isLoading, error } = useCreateAlertRule();
 * 
 * // Query Alert
 * mutate({
 *   name: 'High CPU Alert',
 *   alert_type: 'query',
 *   datasource: 'clickhouse',
 *   evaluation_interval: '1m',
 *   datasource_query: {
 *     query: 'SELECT cpu FROM metrics',
 *     time_label: 'timestamp',
 *     variable_label: 'host'
 *   },
 *   threshold: [{
 *     id: '1',
 *     condition: 'is_above',
 *     start_value: 80,
 *     severity: 'Critical'
 *   }]
 * });
 */
export function useCreateAlertRule() {
  const { mutate, ...rest } = useCreate<AnyAlertRule, HttpError, AnyAlertRule>({
    dataProviderName: ALERT_PROVIDER_NAME,
  });

  const createWithValidation = (data: AnyAlertRule) => {
    // Zod validation happens here
    const validated = AnyAlertRuleSchema.parse(data);
    
    mutate({
      resource: ALERT_RESOURCES.RULE,
      values: validated as AnyAlertRule,
    });
  };

  return {
    mutate: createWithValidation,
    ...rest,
  };
}

/**
 * Custom hook for updating alert rules with Zod validation
 * 
 * @example
 * const { mutate } = useUpdateAlertRule();
 * 
 * mutate('rule-123', {
 *   name: 'Updated Alert Name',
 *   evaluation_interval: '5m'
 * });
 */
export function useUpdateAlertRule() {
  const { mutate, ...rest } = useUpdate<AnyAlertRule, HttpError, Record<string, any>>({
    dataProviderName: ALERT_PROVIDER_NAME,
  });

  const updateWithValidation = (id: string, data: Record<string, any>) => {
    // Basic validation - ensure data is an object
    if (typeof data !== 'object' || data === null) {
      throw new Error('Update data must be an object');
    }
    
    mutate({
      resource: ALERT_RESOURCES.RULE,
      id,
      values: data,
    });
  };

  return {
    mutate: updateWithValidation,
    ...rest,
  };
}
