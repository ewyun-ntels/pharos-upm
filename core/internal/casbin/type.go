package casbin

import (
	"ntels.com/pharos/core/internal"
)

const (
	SubjectPublic = "__public"

	ActionAdmin  = "admin"
	ActionMember = "member"

	ActionOwner  = "owner"
	ActionEditor = "editor"
	ActionViewer = "viewer"
)

type EnforcerType string

func (enforcerType EnforcerType) String() string {
	return string(enforcerType)
}

const (
	EnforcerTypeGroup     = "group"
	EnforcerTypeDashboard = "dashboard"
)

var Enforcers = internal.NewMap[*Enforcer]()

var tableNames = map[EnforcerType]string{
	EnforcerTypeGroup:     "casbin_policy_groups",
	EnforcerTypeDashboard: "casbin_policy_dashboard",
}

var basicModelText = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`

var modelTexts = map[EnforcerType]string{
	EnforcerTypeGroup: basicModelText,
	EnforcerTypeDashboard: `
[request_definition]
r = sub, obj, act, kind

[policy_definition]
p = sub, obj, act, kind

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act && r.kind == p.kind`,
}

type Policy []string

func (policy Policy) getAny() []any {
	result := []any{}

	for _, value := range policy {
		result = append(result, value)
	}

	return result
}
