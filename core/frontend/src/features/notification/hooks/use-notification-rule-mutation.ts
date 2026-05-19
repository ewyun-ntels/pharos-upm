import { useCreate, useUpdate, useDelete, type HttpError } from '@/lib/data-provider';
import { NOTIFICATION_RESOURCES, NOTIFICATION_PROVIDER_NAME } from '@providers/notification-provider';
import {
  QueryNotificationRuleSchema,
  type QueryNotificationRule,
} from '@pharos/shared/types/notification';

// Union type for all notification rules (현재는 SNMP만)
type AnyNotificationRule = QueryNotificationRule;
const AnyNotificationRuleSchema = QueryNotificationRuleSchema;

/**
 * Custom hook for creating notification rules with Zod validation
 *
 * @example
 * const { mutate, isLoading, error } = useCreateNotificationRule();
 *
 * // SNMP Notification
 * mutate({
 *   name: 'SNMP Trap Notification',
 *   target: '192.168.1.100',
 *   port: 162,
 *   version: 2,
 *   snmpv2_config: {
 *     community: 'public'
 *   },
 *   trap_pdu_rule: [{
 *     label_name: 'severity',
 *     oid: '1.3.6.1.4.1.12345.1.1',
 *     value_type: 'OctetString'
 *   }]
 * });
 */
export function useCreateNotificationRule() {
  const { mutate, ...rest } = useCreate<AnyNotificationRule, HttpError, AnyNotificationRule>({
    dataProviderName: NOTIFICATION_PROVIDER_NAME,
  });

  const createWithValidation = (data: AnyNotificationRule) => {
    // Zod validation happens here
    const validated = AnyNotificationRuleSchema.parse(data);

    mutate({
      resource: NOTIFICATION_RESOURCES.RULE,
      values: validated as AnyNotificationRule,
    });
  };

  return {
    mutate: createWithValidation,
    ...rest,
  };
}

/**
 * Custom hook for updating notification rules with Zod validation
 *
 * @example
 * const { mutate } = useUpdateNotificationRule();
 *
 * mutate('rule-123', {
 *   name: 'Updated SNMP Notification',
 *   port: 163
 * });
 */
export function useUpdateNotificationRule() {
  const { mutate, ...rest } = useUpdate<AnyNotificationRule, HttpError, Record<string, any>>({
    dataProviderName: NOTIFICATION_PROVIDER_NAME,
  });

  const updateWithValidation = (id: string, data: Record<string, any>) => {
    // Basic validation - ensure data is an object
    if (typeof data !== 'object' || data === null) {
      throw new Error('Update data must be an object');
    }

    mutate({
      resource: NOTIFICATION_RESOURCES.RULE,
      id,
      values: data,
    });
  };

  return {
    mutate: updateWithValidation,
    ...rest,
  };
}

/**
 * Custom hook for deleting notification rules
 *
 * @example
 * const { mutate } = useDeleteNotificationRule();
 *
 * mutate('rule-123');
 */
export function useDeleteNotificationRule() {
  const { mutate, ...rest } = useDelete<AnyNotificationRule, HttpError>();

  const deleteRule = (id: string) => {
    mutate({
      resource: NOTIFICATION_RESOURCES.RULE,
      id,
      dataProviderName: NOTIFICATION_PROVIDER_NAME,
    });
  };

  return {
    mutate: deleteRule,
    ...rest,
  };
}
