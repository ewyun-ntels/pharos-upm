package adapters

import (
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/alert/common"
	"ntels.com/pharos/core/pkg/alert/resources"
	"ntels.com/pharos/shared/types/alert"
)

// APIAdapter handles conversion between internal alert types and API response types
type APIAdapter struct{}

// ToAPIAlertValue converts internal common.Value to API alert.AlertValue
func (a *APIAdapter) ToAPIAlertValue(internal *common.Value) alert.AlertValue {
	apiAlert := alert.AlertValue{
		ID:        internal.Id,
		AlertID:   internal.AlertId,
		Name:      internal.Name,
		AlertType: alert.AlertType(internal.AlertType),
		Status:    alert.Status(internal.Status),
		Severity:  alert.Severity(internal.Severity),
		Value:     internal.Value,
		CheckTime: &internal.Timestamp,
		Mask:      internal.Mask,
		Labels:    internal.Labels,
		Timestamp: internal.Timestamp,
		UpdatedAt: internal.UpdatedAt,
	}

	// Handle optional fields
	if internal.Description != "" {
		apiAlert.Description = &internal.Description
	}

	if internal.CheckTime != nil {
		apiAlert.CheckTime = internal.CheckTime
	}

	if internal.PreviousSeverity != "" {
		severity := alert.Severity(internal.PreviousSeverity)
		apiAlert.PreviousSeverity = &severity
	}

	if internal.PreviousValue != nil {
		apiAlert.PreviousValue = internal.PreviousValue
	}

	if !internal.PreviousTimestamp.IsZero() {
		apiAlert.PreviousTimestamp = &internal.PreviousTimestamp
	}

	return apiAlert
}

// ToAPIAlertValues converts slice of internal common.Value to slice of API alert.AlertValue
func (a *APIAdapter) ToAPIAlertValues(internals []common.Value) []alert.AlertValue {
	apiAlerts := make([]alert.AlertValue, len(internals))
	for i, internal := range internals {
		apiAlerts[i] = a.ToAPIAlertValue(&internal)
	}
	return apiAlerts
}

// FromAPIAlertValue converts API alert.AlertValue to internal common.Value
// This is primarily used for API requests (POST/PUT)
func (a *APIAdapter) FromAPIAlertValue(api alert.AlertValue) *common.Value {
	internal := &common.Value{
		Id:        api.ID,
		AlertId:   api.AlertID,
		Name:      api.Name,
		AlertType: string(api.AlertType),
		Status:    string(api.Status),
		Severity:  string(api.Severity),
		Value:     api.Value,
		Mask:      api.Mask,
		Labels:    api.Labels,
		Timestamp: api.Timestamp,
		UpdatedAt: api.UpdatedAt,
	}

	// Handle optional fields
	if api.Description != nil {
		internal.Description = *api.Description
	}

	if api.CheckTime != nil {
		internal.CheckTime = api.CheckTime
	}

	if api.PreviousSeverity != nil {
		internal.PreviousSeverity = string(*api.PreviousSeverity)
	}

	if api.PreviousValue != nil {
		internal.PreviousValue = api.PreviousValue
	}

	if api.PreviousTimestamp != nil {
		internal.PreviousTimestamp = *api.PreviousTimestamp
	}

	// Convert labels to string format for DB storage
	_ = internal.ConvertLabelsToStringLabels()

	return internal
}

// FromAPIEventRequest converts API alert.AlertEventRequest to internal resources.Event
func (a *APIAdapter) FromAPIEventRequest(api alert.AlertEventRequest) *resources.Event {
	event := &resources.Event{
		AlertId:  api.AlertID,
		Severity: api.Severity,
	}

	// Handle optional/pointer fields
	if api.Value != nil {
		event.Value = *api.Value
	}

	if api.Description != nil {
		event.Description = *api.Description
	}

	// Labels type is generated as empty struct, need different handling
	// For now, create empty map - this would need proper schema definition
	if api.Labels != nil {
		event.Labels = make(map[string]string)
	}

	return event
}

// ToAPIHistoryResponse converts slice of alert.AlertValue to alert.AlertHistoryResponse
func (a *APIAdapter) ToAPIHistoryResponse(alerts []alert.AlertValue) alert.AlertHistoryResponse {
	return alert.AlertHistoryResponse(alerts)
}

// ToAPIQueryResponse converts orm.DatabaseResponse to API AlertQueryResponse format
// DatabaseResponse와 AlertQueryResponse는 동일한 구조
func (a *APIAdapter) ToAPIQueryResponse(frame orm.DatabaseResponse) alert.AlertQueryResponse {
	elapsed := frame.Statistics.Elapsed
	return alert.AlertQueryResponse{
		Meta: frame.Meta,
		Data: frame.Data,
		Rows: frame.Rows,
		SQL:  &frame.SQL,
		Statistics: &alert.Statistics{
			Elapsed: &elapsed,
		},
	}
}

// Request adapters - Convert API requests to internal types

// FromAPIStatusRequest converts API alert.AlertStatusRequest to internal filtering parameters
func (a *APIAdapter) FromAPIStatusRequest(api alert.AlertStatusRequest) map[string]any {
	filters := make(map[string]any)

	if api.Name != nil {
		filters["name"] = *api.Name
	}
	if api.Status != nil {
		filters["status"] = string(*api.Status)
	}
	if api.Severity != nil {
		filters["severity"] = *api.Severity
	}
	if api.AlertType != nil {
		filters["alert_type"] = string(*api.AlertType)
	}
	if api.Limit != nil {
		filters["limit"] = *api.Limit
	}
	if api.Offset != nil {
		filters["offset"] = *api.Offset
	}

	return filters
}

// FromAPIQueryRequest converts API alert.AlertQueryRequest to internal resources.Query
func (a *APIAdapter) FromAPIQueryRequest(api alert.AlertQueryRequest) *resources.Query {
	query := &resources.Query{
		Datasource: api.Datasource,
		Query: resources.QuerySpec{
			Query:     api.Query.Query,
			Variables: api.Query.Variables, // 직접 전달 (이미 map[string]interface{})
		},
	}

	return query
}

// FromAPIHistoryRequest converts API alert.AlertHistoryRequest to internal parameters
func (a *APIAdapter) FromAPIHistoryRequest(api alert.AlertHistoryRequest) map[string]any {
	params := make(map[string]any)

	if api.Name != nil {
		params["name"] = *api.Name
	}

	if api.StartTime != nil {
		params["start_time"] = *api.StartTime
	}

	if api.EndTime != nil {
		params["end_time"] = *api.EndTime
	}

	if api.Count != nil {
		params["count"] = int(*api.Count)
	}

	return params
}

// FromAPIStatusMaskRequest converts API alert.AlertStatusMaskRequest to boolean
func (a *APIAdapter) FromAPIStatusMaskRequest(api alert.AlertStatusMaskRequest) bool {
	return api.Mask
}
