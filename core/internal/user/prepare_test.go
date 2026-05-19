package user

import (
	"slices"
	"testing"

	"ntels.com/pharos/shared/types/role"
)

func TestGetActionRoleDefinition_PasswordChange(t *testing.T) {
	def, ok := GetActionRoleDefinition(PreparePasswordChange)
	if !ok {
		t.Fatalf("expected definition for %s", PreparePasswordChange)
	}

	if def.PrepareKey != PreparePasswordChange {
		t.Fatalf("PrepareKey mismatch: got %s, want %s", def.PrepareKey, PreparePasswordChange)
	}
	if def.DisplayRole != role.AttrPasswordChange {
		t.Fatalf("DisplayRole mismatch: got %s, want %s", def.DisplayRole, role.AttrPasswordChange)
	}
	if def.DisplayLabel(nil) != "Password change required" {
		t.Fatalf("DisplayLabel mismatch: got %s", def.DisplayLabel(nil))
	}
	if !slices.Contains(def.RuntimeRoles, role.AttrTemporaryUser) {
		t.Fatalf("RuntimeRoles should contain %s", role.AttrTemporaryUser)
	}
	if !slices.Contains(def.RuntimeRoles, role.AttrPasswordChange) {
		t.Fatalf("RuntimeRoles should contain %s", role.AttrPasswordChange)
	}
}

func TestGetActionRoleDefinition_Unknown(t *testing.T) {
	if _, ok := GetActionRoleDefinition("unknown"); ok {
		t.Fatalf("unknown action role key should not have a definition")
	}
}

func TestActionRoleDefinition_DisplayLabelFallback(t *testing.T) {
	def := ActionRoleDefinition{PrepareKey: "custom_action"}
	if got := def.DisplayLabel(nil); got != "custom_action" {
		t.Fatalf("DisplayLabel fallback mismatch: got %s, want custom_action", got)
	}
}
