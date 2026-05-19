package rule

import (
	"encoding/json"
	"errors"

	alert_common "ntels.com/pharos/core/pkg/alert/common"
	rule_common "ntels.com/pharos/core/pkg/alert/rule/common"
	"ntels.com/pharos/core/pkg/alert/rule/event_history"
	"ntels.com/pharos/core/pkg/alert/rule/event_status"
	"ntels.com/pharos/core/pkg/alert/rule/query"
	"ntels.com/pharos/core/pkg/common"
)

func GetRuleTemplate(alertType alert_common.Type) (rule_common.Rule, error) {
	var rule rule_common.Rule
	switch alertType {
	case alert_common.TypeQuery:
		rule = &query.Rule{}
	case alert_common.TypeEventStatus:
		rule = &event_status.Rule{}
	case alert_common.TypeEventHistory:
		rule = &event_history.Rule{}
	default:
		return nil, errors.New("unknown alert type")
	}

	return rule, nil
}

func GetRule(alertType alert_common.Type, config common.Config, data map[string]any) (rule_common.Rule, error) {
	rule, err := GetRuleTemplate(alertType)
	if err != nil {
		return nil, err
	}

	err = rule.Load(config, data)
	if err != nil {
		return nil, err
	}

	err = rule.Validate()
	if err != nil {
		return nil, err
	}

	return rule, nil
}

func GetRuleFromStr(alertType alert_common.Type, config common.Config, strRule []byte) (rule_common.Rule, error) {
	var decRule map[string]any
	err := json.Unmarshal(strRule, &decRule)
	if err != nil {
		return nil, err
	}

	return GetRule(alertType, config, decRule)
}
