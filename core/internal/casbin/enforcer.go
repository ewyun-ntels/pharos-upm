package casbin

import (
	"errors"

	sqlxadapter "github.com/Blank-Xu/sqlx-adapter"
	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

func NewEnforcer(enforcerType EnforcerType, databaseConfig orm.DatabaseConfig) (*Enforcer, error) {
	modelText, exist := modelTexts[enforcerType]
	if !exist {
		return nil, external.ErrorNotExistModelText
	}

	modelFromString, err := model.NewModelFromString(modelText)
	if err != nil {
		return nil, err
	}

	db, err := orm.DatabasePool.GetDB(databaseConfig)
	if err != nil {
		return nil, err
	}

	if _, exist := tableNames[enforcerType]; !exist {
		return nil, external.ErrorNotExistTableName
	}

	adapter, err := sqlxadapter.NewAdapter(db, tableNames[enforcerType])
	if err != nil {
		return nil, err
	}

	syncedEnforcer, err := casbin.NewSyncedEnforcer(modelFromString, adapter)
	if err != nil {
		return nil, err
	}

	Enforcers.Set(enforcerType.String(), &Enforcer{se: syncedEnforcer})

	return Enforcers.Get(enforcerType.String()), nil
}

type Enforcer struct {
	se *casbin.SyncedEnforcer
}

func (enforcer *Enforcer) Enforce(policies ...Policy) (Policy, error) {
	for _, policy := range policies {
		if exist, err := enforcer.se.Enforce(policy.getAny()...); err != nil {
			return Policy{}, err
		} else if exist {
			return policy, nil
		}
	}

	return Policy{}, external.ErrorNoSuchPolicy
}

func (enforcer *Enforcer) GetFilteredPolicy(fieldIndex int, fieldValues ...string) ([]Policy, error) {
	policies, err := enforcer.se.GetFilteredPolicy(fieldIndex, fieldValues...)
	if err != nil {
		return nil, err
	}

	result := []Policy{}
	for _, policy := range policies {
		result = append(result, policy)
	}

	return result, nil
}

func (enforcer *Enforcer) AddPolicy(policy Policy) error {
	_, err := enforcer.se.AddPolicy(policy.getAny()...)
	return err
}

func (enforcer *Enforcer) RemovePolicy(policy Policy) error {
	_, err := enforcer.se.RemovePolicy(policy.getAny()...)
	return err
}

func (enforcer *Enforcer) RemovePolicyFromField(fieldIndex int, fieldValues ...string) error {
	policies, err := enforcer.GetFilteredPolicy(fieldIndex, fieldValues...)
	if err != nil {
		return err
	}

	for _, policy := range policies {
		if err := enforcer.RemovePolicy(policy); err != nil {
			return err
		}
	}

	return nil
}

func SetAttributesAboutGroups(subject string, attributes map[string]any) {
	enforcer := Enforcers.Get(EnforcerTypeGroup)
	if enforcer != nil {
		policies, err := enforcer.GetFilteredPolicy(0, subject)
		if errors.Is(err, nil) && len(policies) != 0 {
			// Collect all group names into an array
			groups := make([]string, 0, len(policies))
			for _, policy := range policies {
				if len(policy) > 1 {
					groups = append(groups, policy[1])
				}
			}
			if len(groups) > 0 {
				attributes[common.ExtraKeyGroups] = groups
			}
		}
	}

}
