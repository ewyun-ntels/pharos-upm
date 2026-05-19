package casbin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/external"
)

// Test constants
func TestConstants(t *testing.T) {
	t.Run("SubjectPublic", func(t *testing.T) {
		assert.Equal(t, "__public", SubjectPublic)
	})

	t.Run("Actions", func(t *testing.T) {
		assert.Equal(t, "admin", ActionAdmin)
		assert.Equal(t, "member", ActionMember)
		assert.Equal(t, "owner", ActionOwner)
		assert.Equal(t, "editor", ActionEditor)
		assert.Equal(t, "viewer", ActionViewer)
	})

	t.Run("EnforcerTypes", func(t *testing.T) {
		assert.Equal(t, "group", EnforcerTypeGroup)
		assert.Equal(t, "dashboard", EnforcerTypeDashboard)
	})
}

// Test error variables
func TestErrors(t *testing.T) {
	t.Run("ErrorNoSuchPolicy", func(t *testing.T) {
		assert.Error(t, external.ErrorNoSuchPolicy)
		assert.Equal(t, "no such policy", external.ErrorNoSuchPolicy.Error())
		assert.True(t, errors.Is(external.ErrorNoSuchPolicy, external.ErrorNoSuchPolicy))
	})

	t.Run("ErrorNoPermission", func(t *testing.T) {
		assert.Error(t, external.ErrorNoSuchPermission)
		assert.Equal(t, "no such permission", external.ErrorNoSuchPermission.Error())
		assert.True(t, errors.Is(external.ErrorNoSuchPermission, external.ErrorNoSuchPermission))
	})
}

// Test EnforcerType methods
func TestEnforcerType_String(t *testing.T) {
	tests := []struct {
		name     string
		input    EnforcerType
		expected string
	}{
		{
			name:     "group enforcer type",
			input:    EnforcerTypeGroup,
			expected: "group",
		},
		{
			name:     "dashboard enforcer type",
			input:    EnforcerTypeDashboard,
			expected: "dashboard",
		},
		{
			name:     "custom enforcer type",
			input:    EnforcerType("custom"),
			expected: "custom",
		},
		{
			name:     "empty enforcer type",
			input:    EnforcerType(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test Policy methods
func TestPolicy_getAny(t *testing.T) {
	tests := []struct {
		name     string
		policy   Policy
		expected []any
	}{
		{
			name:     "standard policy",
			policy:   Policy{"sub", "obj", "act"},
			expected: []any{"sub", "obj", "act"},
		},
		{
			name:     "dashboard policy",
			policy:   Policy{"user1", "dashboard1", "viewer", "private"},
			expected: []any{"user1", "dashboard1", "viewer", "private"},
		},
		{
			name:     "empty policy",
			policy:   Policy{},
			expected: []any{},
		},
		{
			name:     "single element policy",
			policy:   Policy{"element"},
			expected: []any{"element"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.policy.getAny()
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, len(tt.policy), len(result))
		})
	}
}

// Test table names mapping
func TestTableNames(t *testing.T) {
	tests := []struct {
		name         string
		enforcerType EnforcerType
		expected     string
	}{
		{
			name:         "group enforcer type",
			enforcerType: EnforcerTypeGroup,
			expected:     "casbin_policy_groups",
		},
		{
			name:         "dashboard enforcer type",
			enforcerType: EnforcerTypeDashboard,
			expected:     "casbin_policy_dashboard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName, exists := tableNames[tt.enforcerType]
			assert.True(t, exists, "table name should exist for enforcer type")
			assert.Equal(t, tt.expected, tableName)
		})
	}

	t.Run("non-existent enforcer type", func(t *testing.T) {
		_, exists := tableNames[EnforcerType("non-existent")]
		assert.False(t, exists, "table name should not exist for non-existent enforcer type")
	})
}

// Test model texts
func TestModelTexts(t *testing.T) {
	t.Run("basic model text", func(t *testing.T) {
		assert.NotEmpty(t, basicModelText)
		assert.Contains(t, basicModelText, "[request_definition]")
		assert.Contains(t, basicModelText, "[policy_definition]")
		assert.Contains(t, basicModelText, "[policy_effect]")
		assert.Contains(t, basicModelText, "[matchers]")
	})

	t.Run("group enforcer model", func(t *testing.T) {
		modelText, exists := modelTexts[EnforcerTypeGroup]
		assert.True(t, exists)
		assert.Equal(t, basicModelText, modelText)
	})

	t.Run("dashboard enforcer model", func(t *testing.T) {
		modelText, exists := modelTexts[EnforcerTypeDashboard]
		assert.True(t, exists)
		assert.NotEmpty(t, modelText)
		assert.Contains(t, modelText, "r = sub, obj, act, kind")
		assert.Contains(t, modelText, "p = sub, obj, act, kind")
		assert.Contains(t, modelText, "m = r.sub == p.sub && r.obj == p.obj && r.act == p.act && r.kind == p.kind")
	})

	t.Run("non-existent enforcer type model", func(t *testing.T) {
		_, exists := modelTexts[EnforcerType("non-existent")]
		assert.False(t, exists, "model text should not exist for non-existent enforcer type")
	})
}

// Test global variables initialization
func TestGlobalVariables(t *testing.T) {
	t.Run("Enforcers map initialization", func(t *testing.T) {
		assert.NotNil(t, Enforcers, "Enforcers map should be initialized")
	})
}
