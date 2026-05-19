import * as z from "zod";

// Type of alert rule

export const AlertTypeSchema = z.enum([
    "event-history",
    "event-status",
    "query",
]);
export type AlertType = z.infer<typeof AlertTypeSchema>;

// Alert severity level
//
// Severity level when threshold is met

export const SeveritySchema = z.enum([
    "Critical",
    "Major",
    "Minor",
    "Normal",
]);
export type Severity = z.infer<typeof SeveritySchema>;

// Current alert status
//
// Filter by alert status

export const StatusSchema = z.enum([
    "alerting",
    "event",
    "normal",
]);
export type Status = z.infer<typeof StatusSchema>;


export const EventHistoryAlertRuleAlertTypeSchema = z.enum([
    "event-history",
]);
export type EventHistoryAlertRuleAlertType = z.infer<typeof EventHistoryAlertRuleAlertTypeSchema>;


export const EventStatusAlertRuleAlertTypeSchema = z.enum([
    "event-status",
]);
export type EventStatusAlertRuleAlertType = z.infer<typeof EventStatusAlertRuleAlertTypeSchema>;


export const QueryAlertRuleAlertTypeSchema = z.enum([
    "query",
]);
export type QueryAlertRuleAlertType = z.infer<typeof QueryAlertRuleAlertTypeSchema>;

// Type of check to perform on query results

export const CheckTypeSchema = z.enum([
    "last",
]);
export type CheckType = z.infer<typeof CheckTypeSchema>;

// Comparison operation type

export const ConditionSchema = z.enum([
    "is_above",
    "is_below",
    "outside_range",
    "within_range",
]);
export type Condition = z.infer<typeof ConditionSchema>;

export const AlertEventSchema = z.object({
    "alert_id": z.string(),
    "description": z.string().optional(),
    "labels": z.record(z.string(), z.string()).optional(),
    "severity": z.string(),
    "value": z.number().optional(),
});
export type AlertEvent = z.infer<typeof AlertEventSchema>;

export const LabelsSchema = z.object({
});
export type Labels = z.infer<typeof LabelsSchema>;

export const AlertHistoryRequestSchema = z.object({
    "count": z.number().optional(),
    "endTime": z.coerce.date().optional(),
    "name": z.string().optional(),
    "startTime": z.coerce.date().optional(),
});
export type AlertHistoryRequest = z.infer<typeof AlertHistoryRequestSchema>;

export const QuerySchema = z.object({
    "query": z.string(),
    "variables": z.record(z.string(), z.any()).optional(),
});
export type Query = z.infer<typeof QuerySchema>;

export const StatisticsSchema = z.object({
    "elapsed": z.number().optional(),
});
export type Statistics = z.infer<typeof StatisticsSchema>;

export const AlertValueSchema = z.object({
    "alert_id": z.string(),
    "alert_type": AlertTypeSchema,
    "check_time": z.union([z.coerce.date(), z.null()]).optional(),
    "description": z.union([z.null(), z.string()]).optional(),
    "id": z.string(),
    "labels": z.record(z.string(), z.string()).optional(),
    "mask": z.boolean(),
    "name": z.string(),
    "previous_severity": z.union([SeveritySchema, z.null()]).optional(),
    "previous_timestamp": z.union([z.coerce.date(), z.null()]).optional(),
    "previous_value": z.union([z.number(), z.null()]).optional(),
    "severity": SeveritySchema,
    "status": StatusSchema,
    "timestamp": z.coerce.date(),
    "updated_at": z.coerce.date(),
    "value": z.number(),
});
export type AlertValue = z.infer<typeof AlertValueSchema>;

export const AlertStatusMaskRequestSchema = z.object({
    "mask": z.boolean(),
});
export type AlertStatusMaskRequest = z.infer<typeof AlertStatusMaskRequestSchema>;

export const AlertStatusRequestSchema = z.object({
    "alertType": AlertTypeSchema.optional(),
    "limit": z.number().optional(),
    "name": z.string().optional(),
    "offset": z.number().optional(),
    "severity": z.string().optional(),
    "status": StatusSchema.optional(),
});
export type AlertStatusRequest = z.infer<typeof AlertStatusRequestSchema>;

export const EventHistoryAlertRuleSchema = z.object({
    "alert_type": EventHistoryAlertRuleAlertTypeSchema.optional(),
    "description": z.string().optional(),
    "id": z.string().optional(),
    "name": z.string().optional(),
    "notifications": z.array(z.string()).optional(),
    "retention_period": z.number(),
});
export type EventHistoryAlertRule = z.infer<typeof EventHistoryAlertRuleSchema>;

export const EventStatusAlertRuleSchema = z.object({
    "alert_type": EventStatusAlertRuleAlertTypeSchema.optional(),
    "description": z.string().optional(),
    "id": z.string().optional(),
    "name": z.string().optional(),
    "notifications": z.array(z.string()).optional(),
});
export type EventStatusAlertRule = z.infer<typeof EventStatusAlertRuleSchema>;

export const DatasourceQuerySchema = z.object({
    "query": z.string(),
    "time_label": z.string(),
    "variable_label": z.string(),
    "variables": z.record(z.string(), z.any()).optional(),
});
export type DatasourceQuery = z.infer<typeof DatasourceQuerySchema>;

export const ThresholdSchema = z.object({
    "condition": ConditionSchema,
    "end_value": z.number().optional(),
    "id": z.string(),
    "labels": z.record(z.string(), z.string()).optional(),
    "severity": SeveritySchema,
    "start_value": z.number(),
});
export type Threshold = z.infer<typeof ThresholdSchema>;

export const AlertEventRequestSchema = z.object({
    "alertId": z.string(),
    "description": z.string().optional(),
    "labels": LabelsSchema.optional(),
    "severity": z.string(),
    "value": z.number().optional(),
});
export type AlertEventRequest = z.infer<typeof AlertEventRequestSchema>;

export const AlertQueryRequestSchema = z.object({
    "datasource": z.string(),
    "query": QuerySchema,
});
export type AlertQueryRequest = z.infer<typeof AlertQueryRequestSchema>;

export const AlertQueryResponseSchema = z.object({
    "data": z.array(z.record(z.string(), z.any())),
    "meta": z.array(z.record(z.string(), z.string())),
    "rows": z.number(),
    "sql": z.string().optional(),
    "statistics": StatisticsSchema.optional(),
});
export type AlertQueryResponse = z.infer<typeof AlertQueryResponseSchema>;

export const AlertRuleSchema = z.object({
    "alert_type": AlertTypeSchema,
    "rule": z.record(z.string(), z.any()),
    "status": z.array(AlertValueSchema).optional(),
    "timestamp": z.coerce.date().optional(),
    "updated_at": z.coerce.date().optional(),
});
export type AlertRule = z.infer<typeof AlertRuleSchema>;

export const QueryAlertRuleSchema = z.object({
    "alert_type": QueryAlertRuleAlertTypeSchema.optional(),
    "check_type": CheckTypeSchema.optional(),
    "datasource": z.string(),
    "datasource_query": DatasourceQuerySchema,
    "description": z.string().optional(),
    "evaluation_interval": z.string(),
    "id": z.string().optional(),
    "name": z.string().optional(),
    "notifications": z.array(z.string()).optional(),
    "query_delay_offset": z.number().optional(),
    "threshold": z.array(ThresholdSchema),
});
export type QueryAlertRule = z.infer<typeof QueryAlertRuleSchema>;
