/**
 * Alert Provider Types
 * shared/frontend/types/alert의 Zod 스키마 기반 타입을 re-export
 */

// Alert Provider 상수
export const ALERT_PROVIDER_NAME = 'alert' as const;

// Alert Provider Resource 상수
export const ALERT_RESOURCES = {
  RULE: 'alert/rule',
  STATUS: 'alert/status',
  HIST: 'alert/hist',
  QUERY: 'alert/query',
  STATUS_MASK: 'alert/status/mask',
} as const;

export type AlertResource = (typeof ALERT_RESOURCES)[keyof typeof ALERT_RESOURCES];

// shared/frontend/types/alert에서 모든 타입과 스키마를 가져옴
export {
  // Type exports (타입만)
  type AlertType,
  type Severity as AlertSeverity,
  type Status as AlertStatus,
  type CheckType,
  type Condition,
  type AlertValue,
  type AlertRule,
  type DatasourceQuery,
  type Threshold,
  type QueryAlertRule,
  type EventStatusAlertRule,
  type EventHistoryAlertRule,
  type AlertStatusRequest,
  type AlertStatusMaskRequest,
  type AlertQueryRequest,
  type AlertQueryResponse,
  type AlertHistoryRequest,
  type AlertEvent,
  type AlertEventRequest,
  type Query,
  
  // Zod Schemas (validation에 사용 가능, 런타임 값)
  AlertTypeSchema,
  SeveritySchema as AlertSeveritySchema,
  StatusSchema as AlertStatusSchema,
  CheckTypeSchema,
  ConditionSchema,
  AlertValueSchema,
  AlertRuleSchema,
  DatasourceQuerySchema,
  ThresholdSchema,
  QueryAlertRuleSchema,
  EventStatusAlertRuleSchema,
  EventHistoryAlertRuleSchema,
  AlertStatusRequestSchema,
  AlertStatusMaskRequestSchema,
  AlertQueryRequestSchema,
  AlertQueryResponseSchema,
  AlertHistoryRequestSchema,
  AlertEventSchema,
  AlertEventRequestSchema,
} from '@pharos/shared/types/alert';

// Alert History Response는 AlertValue 배열
import type {AlertValue} from '@pharos/shared/types/alert';
export type AlertHistoryResponse = AlertValue[];

// AlertQueryResponse의 meta 요소 타입
export type AlertQueryColumn = Record<string, string>;

// Create/Update Rule Request 타입 추가 (API 편의를 위해)
import type {AlertType, QueryAlertRule, EventStatusAlertRule, EventHistoryAlertRule} from '@pharos/shared/types/alert';

export interface CreateAlertRuleRequest {
  alert_type: AlertType;
  rule: Partial<QueryAlertRule | EventStatusAlertRule | EventHistoryAlertRule>;
}

export interface UpdateAlertRuleRequest {
  alert_type: AlertType;
  rule: Partial<QueryAlertRule | EventStatusAlertRule | EventHistoryAlertRule>;
}
