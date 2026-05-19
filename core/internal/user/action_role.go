package user

import "ntels.com/pharos/shared/types/role"

type ActionRoleDefinition struct {
	PrepareKey      string
	RuntimeRoles    []role.PermissionKeys
	DisplayRole     role.PermissionKeys
	GetDisplayLabel func(u User) string
}

var actionRoleDefinitions = map[string]ActionRoleDefinition{
	PreparePasswordChange: {
		PrepareKey: PreparePasswordChange,
		RuntimeRoles: []role.PermissionKeys{
			role.AttrTemporaryUser,
			role.AttrPasswordChange,
		},
		DisplayRole: role.AttrPasswordChange,
		GetDisplayLabel: func(u User) string {
			return "Password change required"
		},
	},
}

func GetActionRoleDefinition(key string) (ActionRoleDefinition, bool) {
	def, ok := actionRoleDefinitions[key]
	return def, ok
}

func (def ActionRoleDefinition) DisplayLabel(u User) string {
	if def.GetDisplayLabel != nil {
		return def.GetDisplayLabel(u)
	}
	return def.PrepareKey
}
