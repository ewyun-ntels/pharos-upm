// Code generated from JSON Schema using quicktype. DO NOT EDIT.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    alertEvent, err := UnmarshalAlertEvent(bytes)
//    bytes, err = alertEvent.Marshal()
//
//    alertEventRequest, err := UnmarshalAlertEventRequest(bytes)
//    bytes, err = alertEventRequest.Marshal()
//
//    alertHistoryRequest, err := UnmarshalAlertHistoryRequest(bytes)
//    bytes, err = alertHistoryRequest.Marshal()
//
//    alertHistoryResponse, err := UnmarshalAlertHistoryResponse(bytes)
//    bytes, err = alertHistoryResponse.Marshal()
//
//    alertQueryRequest, err := UnmarshalAlertQueryRequest(bytes)
//    bytes, err = alertQueryRequest.Marshal()
//
//    alertQueryResponse, err := UnmarshalAlertQueryResponse(bytes)
//    bytes, err = alertQueryResponse.Marshal()
//
//    alertRule, err := UnmarshalAlertRule(bytes)
//    bytes, err = alertRule.Marshal()
//
//    alertSeverity, err := UnmarshalAlertSeverity(bytes)
//    bytes, err = alertSeverity.Marshal()
//
//    alertStatusMaskRequest, err := UnmarshalAlertStatusMaskRequest(bytes)
//    bytes, err = alertStatusMaskRequest.Marshal()
//
//    alertStatusRequest, err := UnmarshalAlertStatusRequest(bytes)
//    bytes, err = alertStatusRequest.Marshal()
//
//    alertType, err := UnmarshalAlertType(bytes)
//    bytes, err = alertType.Marshal()
//
//    alertValue, err := UnmarshalAlertValue(bytes)
//    bytes, err = alertValue.Marshal()
//
//    eventHistoryAlertRule, err := UnmarshalEventHistoryAlertRule(bytes)
//    bytes, err = eventHistoryAlertRule.Marshal()
//
//    eventStatusAlertRule, err := UnmarshalEventStatusAlertRule(bytes)
//    bytes, err = eventStatusAlertRule.Marshal()
//
//    queryAlertRule, err := UnmarshalQueryAlertRule(bytes)
//    bytes, err = queryAlertRule.Marshal()

package alert

import "time"

import "encoding/json"

func UnmarshalAlertEvent(data []byte) (AlertEvent, error) {
	var r AlertEvent
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AlertEvent) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAlertEventRequest(data []byte) (AlertEventRequest, error) {
	var r AlertEventRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AlertEventRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAlertHistoryRequest(data []byte) (AlertHistoryRequest, error) {
	var r AlertHistoryRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AlertHistoryRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type AlertHistoryResponse []AlertValue

func UnmarshalAlertHistoryResponse(data []byte) (AlertHistoryResponse, error) {
	var r AlertHistoryResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AlertHistoryResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAlertQueryRequest(data []byte) (AlertQueryRequest, error) {
	var r AlertQueryRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AlertQueryRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAlertQueryResponse(data []byte) (AlertQueryResponse, error) {
	var r AlertQueryResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AlertQueryResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAlertRule(data []byte) (AlertRule, error) {
	var r AlertRule
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AlertRule) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type AlertSeverity string

func UnmarshalAlertSeverity(data []byte) (AlertSeverity, error) {
	var r AlertSeverity
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AlertSeverity) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAlertStatusMaskRequest(data []byte) (AlertStatusMaskRequest, error) {
	var r AlertStatusMaskRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AlertStatusMaskRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAlertStatusRequest(data []byte) (AlertStatusRequest, error) {
	var r AlertStatusRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AlertStatusRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAlertType(data []byte) (AlertType, error) {
	var r AlertType
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AlertType) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAlertValue(data []byte) (AlertValue, error) {
	var r AlertValue
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AlertValue) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalEventHistoryAlertRule(data []byte) (EventHistoryAlertRule, error) {
	var r EventHistoryAlertRule
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *EventHistoryAlertRule) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalEventStatusAlertRule(data []byte) (EventStatusAlertRule, error) {
	var r EventStatusAlertRule
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *EventStatusAlertRule) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalQueryAlertRule(data []byte) (QueryAlertRule, error) {
	var r QueryAlertRule
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *QueryAlertRule) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Schema for external alert events sent to event-based alert rules
type AlertEvent struct {
	// Unique identifier for the alert event
	AlertID string `json:"alert_id"`
	// Human-readable description of the event
	Description *string `json:"description,omitempty"`
	// Key-value labels for categorizing and identifying the event
	Labels   map[string]string `json:"labels,omitempty"`
	Severity string            `json:"severity"`
	// Numeric value associated with the event
	Value *float64 `json:"value,omitempty"`
}

// Request schema for sending events to alert rules via external integration
type AlertEventRequest struct {
	// Unique identifier for the alert instance
	AlertID string `json:"alertId"`
	// Human-readable description of the event
	Description *string `json:"description,omitempty"`
	// Key-value pairs for additional event metadata
	Labels   *Labels `json:"labels,omitempty"`
	Severity string  `json:"severity"`
	// Numeric value associated with the event
	Value *float64 `json:"value,omitempty"`
}

// Key-value pairs for additional event metadata
type Labels struct {
}

// Request schema for querying alert history with filtering parameters
type AlertHistoryRequest struct {
	// Maximum number of history records to return
	Count *int64 `json:"count,omitempty"`
	// End time for history query (ISO 8601 format)
	EndTime *time.Time `json:"endTime,omitempty"`
	// Filter by alert rule name
	Name *string `json:"name,omitempty"`
	// Start time for history query (ISO 8601 format)
	StartTime *time.Time `json:"startTime,omitempty"`
}

// Request schema for testing alert queries against data sources
type AlertQueryRequest struct {
	// Name of the datasource to query against
	Datasource string `json:"datasource"`
	Query      Query  `json:"query"`
}

// Query specification with SQL and variables
type Query struct {
	// SQL query string with variable placeholders
	Query string `json:"query"`
	// Variables to substitute in the query
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// Response schema containing query results from datasource (same as DatabaseResponse)
type AlertQueryResponse struct {
	// Result rows
	Data []map[string]interface{} `json:"data"`
	// Column metadata
	Meta []map[string]string `json:"meta"`
	// Number of rows returned
	Rows int64 `json:"rows"`
	// SQL query executed
	SQL        *string     `json:"sql,omitempty"`
	Statistics *Statistics `json:"statistics,omitempty"`
}

type Statistics struct {
	// Query execution time in seconds
	Elapsed *float64 `json:"elapsed,omitempty"`
}

// Schema for creating new alert rules (POST requests)
type AlertRule struct {
	// Type of alert rule
	AlertType AlertType `json:"alert_type"`
	// Rule configuration object (varies by alert_type)
	Rule map[string]interface{} `json:"rule"`
	// Current alert status instances for this rule
	Status []AlertValue `json:"status,omitempty"`
	// Creation timestamp
	Timestamp *time.Time `json:"timestamp,omitempty"`
	// Last update timestamp
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// Array of alert history records ordered by timestamp descending
//
// API response schema for alert instances and their status information
type AlertValue struct {
	// Alert identifier based on rule name and labels
	AlertID string `json:"alert_id"`
	// Type of alert rule
	AlertType AlertType `json:"alert_type"`
	// Rule execution timestamp
	CheckTime *time.Time `json:"check_time,omitempty"`
	// Alert description
	Description *string `json:"description,omitempty"`
	// System-internal unique identifier for the alert instance
	ID string `json:"id"`
	// Key-value labels associated with the alert
	Labels map[string]string `json:"labels,omitempty"`
	// Whether the alert is masked (suppressed)
	Mask bool `json:"mask"`
	// Name of the alert rule that generated this alert
	Name string `json:"name"`
	// Previous severity level
	PreviousSeverity *Severity `json:"previous_severity,omitempty"`
	// Previous alert timestamp
	PreviousTimestamp *time.Time `json:"previous_timestamp,omitempty"`
	// Previous numeric value
	PreviousValue *float64 `json:"previous_value,omitempty"`
	// Alert severity level
	Severity Severity `json:"severity"`
	// Current alert status
	Status Status `json:"status"`
	// Alert occurrence timestamp
	Timestamp time.Time `json:"timestamp"`
	// Last update timestamp
	UpdatedAt time.Time `json:"updated_at"`
	// Current numeric value that triggered the alert
	Value float64 `json:"value"`
}

// Request schema for masking/unmasking alert status notifications
type AlertStatusMaskRequest struct {
	// Whether to mask (true) or unmask (false) the alert status
	Mask bool `json:"mask"`
}

// Request schema for retrieving alert status information with optional filtering and
// pagination
type AlertStatusRequest struct {
	AlertType *AlertType `json:"alertType,omitempty"`
	// Maximum number of results to return
	Limit *int64 `json:"limit,omitempty"`
	// Filter by alert rule name
	Name *string `json:"name,omitempty"`
	// Number of results to skip for pagination
	Offset *int64 `json:"offset,omitempty"`
	// Filter by alert severity level
	Severity *string `json:"severity,omitempty"`
	// Filter by alert status
	Status *Status `json:"status,omitempty"`
}

// Schema for event-history alert rules that store events and manage retention periods
type EventHistoryAlertRule struct {
	AlertType *EventHistoryAlertRuleAlertType `json:"alert_type,omitempty"`
	// Description shown when alert is triggered
	Description *string `json:"description,omitempty"`
	// Unique identifier for the alert rule
	ID *string `json:"id,omitempty"`
	// Name of the alert rule
	Name *string `json:"name,omitempty"`
	// List of notification destinations
	Notifications []string `json:"notifications,omitempty"`
	// Retention period in seconds for historical events
	RetentionPeriod int64 `json:"retention_period"`
}

// Schema for event-status alert rules that handle external events and maintain alert status
type EventStatusAlertRule struct {
	AlertType *EventStatusAlertRuleAlertType `json:"alert_type,omitempty"`
	// Description shown when alert is triggered
	Description *string `json:"description,omitempty"`
	// Unique identifier for the alert rule
	ID *string `json:"id,omitempty"`
	// Name of the alert rule
	Name *string `json:"name,omitempty"`
	// List of notification destinations
	Notifications []string `json:"notifications,omitempty"`
}

// Schema for query-based alert rules that evaluate database queries against thresholds
type QueryAlertRule struct {
	AlertType *QueryAlertRuleAlertType `json:"alert_type,omitempty"`
	// Type of check to perform on query results
	CheckType *CheckType `json:"check_type,omitempty"`
	// Name of the datasource to query
	Datasource string `json:"datasource"`
	// Query configuration for datasource
	DatasourceQuery DatasourceQuery `json:"datasource_query"`
	// Description shown when alert is triggered
	Description *string `json:"description,omitempty"`
	// Cron expression for evaluation schedule
	EvaluationInterval string `json:"evaluation_interval"`
	// Unique identifier for the alert rule
	ID *string `json:"id,omitempty"`
	// Name of the alert rule
	Name *string `json:"name,omitempty"`
	// List of notification destinations
	Notifications []string `json:"notifications,omitempty"`
	// Delay in seconds before executing query
	QueryDelayOffset *int64 `json:"query_delay_offset,omitempty"`
	// List of threshold conditions
	Threshold []Threshold `json:"threshold"`
}

// Query configuration for datasource
type DatasourceQuery struct {
	// SQL query or query template
	Query string `json:"query"`
	// Column name for timestamp values
	TimeLabel string `json:"time_label"`
	// Column name for the value to evaluate
	VariableLabel string `json:"variable_label"`
	// Variables for query templating
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type Threshold struct {
	// Comparison operation type
	Condition Condition `json:"condition"`
	// Upper bound for range operations (unused for single value operations)
	EndValue *float64 `json:"end_value,omitempty"`
	// Unique identifier for the threshold
	ID string `json:"id"`
	// Additional labels to attach to alerts from this threshold
	Labels map[string]string `json:"labels,omitempty"`
	// Severity level when threshold is met
	Severity Severity `json:"severity"`
	// Primary threshold value (single value for is_above/is_below, lower bound for range
	// operations)
	StartValue float64 `json:"start_value"`
}

// Type of alert rule
type AlertType string

const (
	AlertEventHistory AlertType = "event-history"
	AlertEventStatus  AlertType = "event-status"
	AlertQuery        AlertType = "query"
)

// Alert severity level
//
// Severity level when threshold is met
type Severity string

const (
	Critical Severity = "Critical"
	Major    Severity = "Major"
	Minor    Severity = "Minor"
	Normal   Severity = "Normal"
)

// Current alert status
//
// Filter by alert status
type Status string

const (
	Alerting     Status = "alerting"
	Event        Status = "event"
	StatusNormal Status = "normal"
)

type EventHistoryAlertRuleAlertType string

const (
	AlertTypeEventHistory EventHistoryAlertRuleAlertType = "event-history"
)

type EventStatusAlertRuleAlertType string

const (
	AlertTypeEventStatus EventStatusAlertRuleAlertType = "event-status"
)

type QueryAlertRuleAlertType string

const (
	AlertTypeQuery QueryAlertRuleAlertType = "query"
)

// Type of check to perform on query results
type CheckType string

const (
	Last CheckType = "last"
)

// Comparison operation type
type Condition string

const (
	IsAbove      Condition = "is_above"
	IsBelow      Condition = "is_below"
	OutsideRange Condition = "outside_range"
	WithinRange  Condition = "within_range"
)
