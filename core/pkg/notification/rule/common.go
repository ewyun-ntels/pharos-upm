package rule

import (
	"context"
	"encoding/json"
	"errors"

	"ntels.com/pharos/core/pkg/common"
	notification_common "ntels.com/pharos/core/pkg/notification/common"
	"ntels.com/pharos/core/pkg/notification/rule/snmp"
)

type Rule interface {
	Load(config common.Config, data map[string]any) error
	GetId() string
	SetId(id string)
	GetName() string
	Validate() error
	Send(value *notification_common.AlertValue) error
	Run(ctx context.Context) error
	Destroy() error
}

func GetRuleTemplate(alertType notification_common.Type) (Rule, error) {
	var rule Rule
	switch alertType {
	case notification_common.TypeSNMP:
		rule = &snmp.SNMP{}
	default:
		return nil, errors.New("unknown notification type")
	}

	return rule, nil
}

func GetRule(notificationType notification_common.Type, config common.Config, data map[string]any) (Rule, error) {
	rule, err := GetRuleTemplate(notificationType)
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

func GetRuleFromStr(notificationType notification_common.Type, config common.Config, strRule []byte) (Rule, error) {
	var decRule map[string]any
	err := json.Unmarshal(strRule, &decRule)
	if err != nil {
		return nil, err
	}

	return GetRule(notificationType, config, decRule)
}
