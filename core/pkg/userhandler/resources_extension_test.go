package userhandler

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/plugins"
	"ntels.com/pharos/core/pkg/plugins/model"
	sharedRole "ntels.com/pharos/shared/types/role"
)

// fakeExtension은 테스트용 extension 구현체입니다.
type fakeExtension struct {
	name  string
	roles []sharedRole.RoleMetadata
}

func (e *fakeExtension) GetPlugin() (*model.Plugin, error)   { return nil, nil }
func (e *fakeExtension) RegisterRoutes(_ *gin.RouterGroup)   {}
func (e *fakeExtension) GetName() string                     { return e.name }
func (e *fakeExtension) GetVersion() string                  { return "1.0.0" }
func (e *fakeExtension) GetType() string                     { return "feature" }
func (e *fakeExtension) GetRoles() []sharedRole.RoleMetadata { return e.roles }

// Test_getUserDetail_ExtensionRoleDisplayName은 extension role의 displayName이
// /me 응답의 roles[].display_name에 올바르게 반영되는지 검증합니다.
//
// 버그 재현:
//   - extension roles는 GetCoreRoleMetadata()에 포함되지 않으므로
//     수정 전 코드는 metadataMap에 extension key가 없어서 fallback(key 그대로)이 반환됨.
//
// 기대 동작 (수정 후):
//   - extension registry의 roles까지 metadataMap에 포함하므로 올바른 displayName 반환.
func Test_getUserDetail_ExtensionRoleDisplayName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const extName = "Test Extension"
	ext := &fakeExtension{
		name: extName,
		roles: []sharedRole.RoleMetadata{
			{Key: "extension:test:read", DisplayName: "Test Read", Description: "Read access", Group: "permission"},
			{Key: "extension:test:write", DisplayName: "Test Write", Description: "Write access", Group: "permission"},
		},
	}
	plugins.RegisterExtension(ext)

	extra := map[string]any{
		"roles": map[string]any{
			"extension:test:read":  true,
			"extension:test:write": true,
		},
		"info": map[string]any{},
	}
	fu := &fakeUser{
		id:      "testuser",
		created: time.Now(),
		extra:   extra,
	}

	d := getUserDetail(fu, 3, 30*time.Minute, nil)

	require.NotNil(t, d.Roles)
	require.Len(t, d.Roles, 2)

	displayNames := make(map[string]string)
	for _, r := range d.Roles {
		displayNames[r.Role] = r.DisplayName
	}

	// 수정 전: displayNames["extension:test:read"] == "extension:test:read" (key 그대로)
	// 수정 후: displayNames["extension:test:read"] == "Test Read"
	require.Equal(t, "Test Read", displayNames["extension:test:read"],
		"extension role의 displayName이 key 그대로 반환됨 (버그)")
	require.Equal(t, "Test Write", displayNames["extension:test:write"],
		"extension role의 displayName이 key 그대로 반환됨 (버그)")
}
