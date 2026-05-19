package role

import (
	"testing"

	sharedRole "ntels.com/pharos/shared/types/role"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  []sharedRole.RoleMetadata
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty config is valid",
			config:  []sharedRole.RoleMetadata{},
			wantErr: false,
		},
		{
			name: "valid config",
			config: []sharedRole.RoleMetadata{
				{
					Key:         "role:test",
					DisplayName: "Test Role",
					Description: "Test",
					Group:       "custom",
				},
			},
			wantErr: false,
		},
		{
			name: "missing key",
			config: []sharedRole.RoleMetadata{
				{
					DisplayName: "No Key",
					Group:       "test",
				},
			},
			wantErr: true,
			errMsg:  "key is required",
		},
		{
			name: "empty key",
			config: []sharedRole.RoleMetadata{
				{
					Key:         "",
					DisplayName: "Empty Key",
					Group:       "test",
				},
			},
			wantErr: true,
			errMsg:  "key is required",
		},
		{
			name: "missing displayName",
			config: []sharedRole.RoleMetadata{
				{
					Key:   "role:noname",
					Group: "test",
				},
			},
			wantErr: true,
			errMsg:  "displayName is required",
		},
		{
			name: "empty displayName",
			config: []sharedRole.RoleMetadata{
				{
					Key:         "role:test",
					DisplayName: "",
					Group:       "test",
				},
			},
			wantErr: true,
			errMsg:  "displayName is required",
		},
		{
			name: "duplicate keys",
			config: []sharedRole.RoleMetadata{
				{
					Key:         "role:dup",
					DisplayName: "Dup 1",
					Group:       "test",
				},
				{
					Key:         "role:dup",
					DisplayName: "Dup 2",
					Group:       "test",
				},
			},
			wantErr: true,
			errMsg:  "duplicate key",
		},
		{
			name: "valid composite role",
			config: []sharedRole.RoleMetadata{
				{
					Key:         "role:comp",
					DisplayName: "Composite",
					Group:       "composite",
					Roles: map[string]bool{
						"role:user_read":   true,
						"role:user_update": true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "composite role with empty roles map",
			config: []sharedRole.RoleMetadata{
				{
					Key:         "role:comp",
					DisplayName: "Empty Composite",
					Group:       "composite",
					Roles:       map[string]bool{},
				},
			},
			wantErr: true,
			errMsg:  "must have at least one role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateConfig() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateConfig() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateConfig() unexpected error = %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
