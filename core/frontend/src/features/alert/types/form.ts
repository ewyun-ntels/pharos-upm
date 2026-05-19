/**
 * Alert Rule Form Types
 * 
 * React Hook Form용 타입 정의
 * - Zod 스키마 기반
 * - 각 Alert Type별 Form 타입
 */

import {z} from 'zod';
import {
  QueryAlertRuleSchema,
  EventStatusAlertRuleSchema,
  EventHistoryAlertRuleSchema,
  DatasourceQuerySchema,
  ThresholdSchema,
} from '@pharos/shared/types/alert';

/**
 * Query Alert Rule Form Schema
 * React Hook Form + zodResolver 사용
 */
export const QueryAlertRuleFormSchema = QueryAlertRuleSchema.omit({
  id: true,
});

export type QueryAlertRuleForm = z.infer<typeof QueryAlertRuleFormSchema>;

/**
 * Event Status Alert Rule Form Schema
 */
export const EventStatusAlertRuleFormSchema = EventStatusAlertRuleSchema.omit({
  id: true,
});

export type EventStatusAlertRuleForm = z.infer<typeof EventStatusAlertRuleFormSchema>;

/**
 * Event History Alert Rule Form Schema
 */
export const EventHistoryAlertRuleFormSchema = EventHistoryAlertRuleSchema.omit({
  id: true,
});

export type EventHistoryAlertRuleForm = z.infer<typeof EventHistoryAlertRuleFormSchema>;

/**
 * Datasource Query Form (Query Alert의 일부)
 */
export type DatasourceQueryForm = z.infer<typeof DatasourceQuerySchema>;

/**
 * Threshold Form (Query Alert의 일부)
 */
export type ThresholdForm = z.infer<typeof ThresholdSchema>;

/**
 * 전체 Alert Rule Form (Union)
 */
export type AlertRuleForm =
  | QueryAlertRuleForm
  | EventStatusAlertRuleForm
  | EventHistoryAlertRuleForm;
