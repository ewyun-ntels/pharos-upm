package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// JWTExtra Tests
// ============================================================================

func TestJWTExtraFromMap_NilInput(t *testing.T) {
	extra, err := JWTExtraFromMap(nil)

	require.NoError(t, err)
	assert.NotNil(t, extra)
	assert.NotNil(t, extra.Roles)
	assert.NotNil(t, extra.Info)
	assert.Empty(t, extra.Roles)
	assert.Empty(t, extra.Info)
	assert.Empty(t, extra.Groups)
}

func TestJWTExtraFromMap_EmptyMap(t *testing.T) {
	input := map[string]any{}
	extra, err := JWTExtraFromMap(input)

	require.NoError(t, err)
	assert.NotNil(t, extra)
	assert.NotNil(t, extra.Roles)
	assert.NotNil(t, extra.Info)
}

func TestJWTExtraFromMap_ValidStructure(t *testing.T) {
	// DB 구조와 동일한 형식
	input := map[string]any{
		"roles": map[string]any{
			"role:super_admin": true,
			"role:user":        true,
		},
		"info": map[string]any{
			"department": "engineering",
			"level":      5,
		},
		"groups": []any{"group1", "group2"},
	}

	extra, err := JWTExtraFromMap(input)

	require.NoError(t, err)
	assert.NotNil(t, extra)

	// Roles 검증
	assert.True(t, extra.Roles["role:super_admin"])
	assert.True(t, extra.Roles["role:user"])
	assert.False(t, extra.Roles["role:admin"]) // 존재하지 않는 역할

	// Info 검증
	assert.Equal(t, "engineering", extra.Info["department"])
	// mapstructure는 int를 그대로 유지함 (JSON 변환과 다름)
	assert.Equal(t, 5, extra.Info["level"])

	// Groups 검증
	assert.Len(t, extra.Groups, 2)
	assert.Contains(t, extra.Groups, "group1")
	assert.Contains(t, extra.Groups, "group2")
}

func TestJWTExtraFromMap_TypeConversion(t *testing.T) {
	// interface{}에서 bool로 변환되는지 확인
	input := map[string]any{
		"roles": map[string]any{
			"role:admin": any(true), // interface{} 타입
		},
		"info": map[string]any{},
	}

	extra, err := JWTExtraFromMap(input)

	require.NoError(t, err)
	assert.True(t, extra.Roles["role:admin"])
}

func TestJWTExtraFromMap_RolesOnly(t *testing.T) {
	input := map[string]any{
		"roles": map[string]any{
			"role:viewer": true,
		},
	}

	extra, err := JWTExtraFromMap(input)

	require.NoError(t, err)
	assert.True(t, extra.Roles["role:viewer"])
	assert.NotNil(t, extra.Info) // nil 방지 확인
	assert.Empty(t, extra.Info)
}

func TestJWTExtraFromMap_InfoOnly(t *testing.T) {
	input := map[string]any{
		"info": map[string]any{
			"custom_field": "value",
		},
	}

	extra, err := JWTExtraFromMap(input)

	require.NoError(t, err)
	assert.Equal(t, "value", extra.Info["custom_field"])
	assert.NotNil(t, extra.Roles) // nil 방지 확인
	assert.Empty(t, extra.Roles)
}

func TestJWTExtra_ToMap(t *testing.T) {
	extra := &JWTExtra{
		Roles: map[string]bool{
			"role:admin": true,
			"role:user":  true,
		},
		Info: map[string]any{
			"name": "test",
			"age":  30,
		},
		Groups: []string{"group1", "group2"},
	}

	result, err := extra.ToMap()

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Roles 검증
	roles, ok := result["roles"].(map[string]any)
	require.True(t, ok, "roles should be a map")
	assert.Equal(t, true, roles["role:admin"])
	assert.Equal(t, true, roles["role:user"])

	// Info 검증
	info, ok := result["info"].(map[string]any)
	require.True(t, ok, "info should be a map")
	assert.Equal(t, "test", info["name"])
	assert.Equal(t, float64(30), info["age"]) // JSON 변환 시 숫자는 float64

	// Groups 검증
	groups, ok := result["groups"].([]any)
	require.True(t, ok, "groups should be an array")
	assert.Len(t, groups, 2)
}

func TestJWTExtra_ToMap_EmptyFields(t *testing.T) {
	extra := &JWTExtra{
		Roles: make(map[string]bool),
		Info:  make(map[string]any),
	}

	result, err := extra.ToMap()

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2, "Should have roles and info fields")

	// 키 존재 확인
	_, hasRoles := result["roles"]
	_, hasInfo := result["info"]
	assert.True(t, hasRoles, "Should contain 'roles' key")
	assert.True(t, hasInfo, "Should contain 'info' key")
}

func TestJWTExtra_RoundTrip(t *testing.T) {
	// Map → JWTExtra → Map 양방향 변환 일관성 테스트
	original := map[string]any{
		"roles": map[string]any{
			"role:admin":       true,
			"role:super_admin": false,
		},
		"info": map[string]any{
			"department": "IT",
			"level":      3,
		},
		"groups": []any{"dev", "ops"},
	}

	// Map → JWTExtra
	extra, err := JWTExtraFromMap(original)
	require.NoError(t, err)

	// JWTExtra → Map
	result, err := extra.ToMap()
	require.NoError(t, err)

	// 검증
	roles := result["roles"].(map[string]any)
	assert.Equal(t, true, roles["role:admin"])
	assert.Equal(t, false, roles["role:super_admin"])

	info := result["info"].(map[string]any)
	assert.Equal(t, "IT", info["department"])
	assert.Equal(t, float64(3), info["level"])

	groups := result["groups"].([]any)
	assert.Len(t, groups, 2)
}

func TestHasRole_ValidRole(t *testing.T) {
	extra := map[string]any{
		"roles": map[string]any{
			"role:admin": true,
			"role:user":  false,
		},
		"info": map[string]any{},
	}

	assert.True(t, HasRole(extra, "role:admin"))
	assert.False(t, HasRole(extra, "role:user"))
	assert.False(t, HasRole(extra, "role:nonexistent"))
}

func TestHasRole_NilExtra(t *testing.T) {
	assert.False(t, HasRole(nil, "role:admin"))
}

func TestHasRole_EmptyExtra(t *testing.T) {
	extra := map[string]any{}
	assert.False(t, HasRole(extra, "role:admin"))
}

func TestHasRole_MissingRolesField(t *testing.T) {
	extra := map[string]any{
		"info": map[string]any{
			"key": "value",
		},
	}
	assert.False(t, HasRole(extra, "role:admin"))
}

func TestHasRole_DBStructureFormat(t *testing.T) {
	// 실제 DB에 저장되는 형식과 동일
	extra := map[string]any{
		"roles": map[string]any{
			"role:super_admin": true,
		},
		"info": map[string]any{},
	}

	assert.True(t, HasRole(extra, "role:super_admin"))
	assert.False(t, HasRole(extra, "role:manage_user"))
}

func TestJWTExtraFromMap_GroupsTypeConversion(t *testing.T) {
	// []interface{} → []string 변환 테스트
	input := map[string]any{
		"roles":  map[string]any{},
		"info":   map[string]any{},
		"groups": []any{"group1", "group2", "group3"},
	}

	extra, err := JWTExtraFromMap(input)

	require.NoError(t, err)
	assert.Len(t, extra.Groups, 3)
	assert.Equal(t, []string{"group1", "group2", "group3"}, extra.Groups)
}

func TestJWTExtra_NilSafety(t *testing.T) {
	// nil 필드가 없도록 보장
	tests := []struct {
		name  string
		input map[string]any
	}{
		{
			name:  "completely empty",
			input: map[string]any{},
		},
		{
			name: "only groups",
			input: map[string]any{
				"groups": []any{"g1"},
			},
		},
		{
			name: "nil roles",
			input: map[string]any{
				"roles": nil,
				"info":  map[string]any{},
			},
		},
		{
			name: "nil info",
			input: map[string]any{
				"roles": map[string]any{},
				"info":  nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extra, err := JWTExtraFromMap(tt.input)
			require.NoError(t, err)
			assert.NotNil(t, extra.Roles, "Roles should never be nil")
			assert.NotNil(t, extra.Info, "Info should never be nil")
		})
	}
}

func TestJWTExtra_MultipleRoles(t *testing.T) {
	// 여러 역할을 가진 사용자
	input := map[string]any{
		"roles": map[string]any{
			"role:super_admin":    true,
			"role:manage_user":    true,
			"role:view_logs":      true,
			"attr:temporary_user": false,
		},
		"info": map[string]any{},
	}

	extra, err := JWTExtraFromMap(input)
	require.NoError(t, err)

	assert.True(t, extra.Roles["role:super_admin"])
	assert.True(t, extra.Roles["role:manage_user"])
	assert.True(t, extra.Roles["role:view_logs"])
	assert.False(t, extra.Roles["attr:temporary_user"])
	assert.Len(t, extra.Roles, 4)
}

// Benchmark tests
func BenchmarkJWTExtraFromMap(b *testing.B) {
	input := map[string]any{
		"roles": map[string]any{
			"role:admin": true,
			"role:user":  true,
		},
		"info": map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
		"groups": []any{"group1", "group2"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = JWTExtraFromMap(input)
	}
}

func BenchmarkJWTExtra_ToMap(b *testing.B) {
	extra := &JWTExtra{
		Roles: map[string]bool{
			"role:admin": true,
			"role:user":  true,
		},
		Info: map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
		Groups: []string{"group1", "group2"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extra.ToMap()
	}
}

func BenchmarkHasRole(b *testing.B) {
	extra := map[string]any{
		"roles": map[string]any{
			"role:admin": true,
		},
		"info": map[string]any{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HasRole(extra, "role:admin")
	}
}
