package adapters

import (
	"encoding/json"
	"log/slog"
	"time"

	"ntels.com/pharos/core/pkg/notification/common"
	"ntels.com/pharos/core/pkg/notification/model"
	"ntels.com/pharos/core/pkg/notification/resources"
	"ntels.com/pharos/shared/types/notification"
) // RuleAdapter handles conversion between internal rule types and API response types

type RuleAdapter struct{}

// NewRuleAdapter creates a new instance of RuleAdapter
func NewRuleAdapter() *RuleAdapter {
	return &RuleAdapter{}
}

// FromAPIRuleCreate converts API NotificationRuleCreate to internal rule configuration
func (a *RuleAdapter) FromAPIRuleCreate(apiRule notification.NotificationRule) (map[string]any, error) {
	ruleData := apiRule.Rule

	return ruleData, nil
}

// FromAPIRuleUpdate converts API NotificationRuleUpdate to internal rule configuration
func (a *RuleAdapter) FromAPIRuleUpdate(apiRule notification.NotificationRule) (map[string]any, error) {
	ruleData := apiRule.Rule

	return ruleData, nil
}

// ModelToResource converts model.Rule to resources.Rule with JSON parsing
func (a *RuleAdapter) ModelToResource(rule *model.Rule) (*resources.Rule, error) {
	if rule == nil {
		return nil, nil
	}

	var ruleMap map[string]any
	if rule.Rule != "" {
		if err := json.Unmarshal([]byte(rule.Rule), &ruleMap); err != nil {
			slog.Warn("Failed to parse rule JSON, using raw string", "rule", rule.Rule, "error", err)
			ruleMap = map[string]any{
				"raw_rule": rule.Rule,
			}
		}
	} else {
		ruleMap = make(map[string]any)
	}

	return &resources.Rule{
		NotificationType: common.Type(rule.NotificationType),
		Rule:             ruleMap,
		Timestamp:        rule.Timestamp.Time,
		UpdatedAt:        rule.UpdatedAt.Time,
	}, nil
}

// ToAPIRuleResponseFromInternal converts internal resources.Rule to API NotificationRuleResponse
func ToAPIRuleResponseFromInternal(internalRule *resources.Rule) (*notification.NotificationRule, error) {
	if internalRule == nil {
		return nil, nil
	}

	notificationType := notification.NotificationType(internalRule.NotificationType)

	// Convert timestamps
	var timestamp *time.Time
	var updatedAt *time.Time

	if !internalRule.Timestamp.IsZero() {
		timestamp = &internalRule.Timestamp
	}
	if !internalRule.UpdatedAt.IsZero() {
		updatedAt = &internalRule.UpdatedAt
	}

	return &notification.NotificationRule{
		NotificationType: notificationType,
		Rule:             internalRule.Rule,
		Timestamp:        timestamp,
		UpdatedAt:        updatedAt,
	}, nil
}
