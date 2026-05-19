package adapters

import (
	"encoding/json"
	"time"

	"ntels.com/pharos/core/pkg/alert/resources"
	"ntels.com/pharos/shared/types/alert"
)

// RuleAdapter handles conversion between internal rule types and API response types
type RuleAdapter struct{}

// FromAPIRuleCreate converts API AlertRuleCreate to internal rule configuration
// This is used when creating rules via API
func (a *RuleAdapter) FromAPIRuleCreate(apiRule alert.AlertRule) (map[string]any, error) {
	ruleData := apiRule.Rule
	return ruleData, nil
}

// FromAPIRuleUpdate converts API AlertRuleUpdate to internal rule configuration
// This is used when updating rules via API (PUT requests)
func (a *RuleAdapter) FromAPIRuleUpdate(apiRule alert.AlertRule) (map[string]any, error) {
	// Convert AlertRuleUpdate to map for internal rule loading
	ruleData := apiRule.Rule
	return ruleData, nil
}

// MarshalRuleConfig marshals rule configuration for storage
func (a *RuleAdapter) MarshalRuleConfig(ruleConfig map[string]any) (string, error) {
	configBytes, err := json.Marshal(ruleConfig)
	if err != nil {
		return "", err
	}
	return string(configBytes), nil
}

// ToAPIRuleResponseFromInternal converts internal resources.Rule to API AlertRuleResponse
func ToAPIRuleResponseFromInternal(internalRule *resources.Rule) (*alert.AlertRule, error) {
	// Convert internal alert type to API alert type
	var alertType alert.AlertType
	switch internalRule.AlertType {
	case "query":
		alertType = alert.AlertQuery
	case "event-status":
		alertType = alert.AlertEventStatus
	case "event-history":
		alertType = alert.AlertEventHistory
	default:
		alertType = alert.AlertQuery // default fallback
	}

	// Convert status if present
	var apiStatus []alert.AlertValue
	if internalRule.Status != nil {
		apiAdapter := &APIAdapter{}
		for _, status := range internalRule.Status {
			alertValueAPI := apiAdapter.ToAPIAlertValue(&status)
			apiStatus = append(apiStatus, alertValueAPI)
		}
	}

	// Convert timestamps
	var timestamp *time.Time
	var updateAt *time.Time

	if !internalRule.Timestamp.IsZero() {
		timestamp = &internalRule.Timestamp
	}
	if !internalRule.UpdateAt.IsZero() {
		updateAt = &internalRule.UpdateAt
	}

	return &alert.AlertRule{
		AlertType: alertType,
		Rule:      internalRule.Rule,
		Status:    apiStatus,
		Timestamp: timestamp,
		UpdatedAt: updateAt,
	}, nil
}
