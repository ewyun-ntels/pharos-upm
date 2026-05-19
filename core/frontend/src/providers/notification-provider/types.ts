/**
 * Notification Provider Types
 * shared/frontend/types/notification의 Zod 스키마 기반 타입을 re-export
 */

// Notification Provider 상수
export const NOTIFICATION_PROVIDER_NAME = 'notification' as const;

// Notification Provider Resource 상수
export const NOTIFICATION_RESOURCES = {
  RULE: 'notification/rule',
} as const;

export type NotificationResource = (typeof NOTIFICATION_RESOURCES)[keyof typeof NOTIFICATION_RESOURCES];

// shared/frontend/types/notification에서 모든 타입과 스키마를 가져옴
export {
  // Type exports (타입만)
  type NotificationType,
  type ValueType,
  type AuthProtocol,
  type PrivProtocol,
  type Transport,
  type NotificationRule,
  type AppendPdu,
  type Snmpv2Config,
  type Snmpv3Config,
  type TrapPduRule,
  type QueryNotificationRule,

  // Zod Schemas (validation에 사용 가능, 런타임 값)
  NotificationTypeSchema,
  ValueTypeSchema,
  AuthProtocolSchema,
  PrivProtocolSchema,
  TransportSchema,
  NotificationRuleSchema,
  AppendPduSchema,
  Snmpv2ConfigSchema,
  Snmpv3ConfigSchema,
  TrapPduRuleSchema,
  QueryNotificationRuleSchema,
} from '@pharos/shared/types/notification';

// Create/Update Rule Request 타입 추가 (API 편의를 위해)
import type {NotificationType, QueryNotificationRule} from '@pharos/shared/types/notification';

export interface CreateNotificationRuleRequest {
  notification_type: NotificationType;
  rule: Partial<QueryNotificationRule>;
}

export interface UpdateNotificationRuleRequest {
  notification_type: NotificationType;
  rule: Partial<QueryNotificationRule>;
}
