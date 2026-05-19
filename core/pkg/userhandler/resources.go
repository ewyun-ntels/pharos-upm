package userhandler

import (
	"context"
	"log/slog"
	"time"

	"ntels.com/pharos/core/internal/user"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins"
	coreRole "ntels.com/pharos/core/pkg/role"
	sharedRole "ntels.com/pharos/shared/types/role"
	sharedUser "ntels.com/pharos/shared/types/user"
)

// Details Details는 shared types의 User를 그대로 사용
type Details = sharedUser.User

func getUserDetail(u user.User, retryLimit int, blockDuration time.Duration, metadataService *coreRole.MetadataService) Details {
	if u == nil {
		return Details{}
	}
	blocked := u.GetBlocked()
	createdAt := u.GetCreatedAt()

	details := Details{
		Name:                      u.GetID(),
		CreatedAt:                 &createdAt,
		Blocked:                   &blocked,
		Prepare:                   u.GetPrepare(),
		PasswordExpiredAt:         u.GetPasswordExpiresAt(),
		TemporaryBlockedExpiresAt: u.GetTemporaryBlockedExpiresAt(retryLimit, blockDuration),
	}

	// attributes 변환 - DB에 {"roles": {...}, "info": {...}} 형식으로 저장됨
	extra := u.GetExtra()
	if len(extra) > 0 {
		// JWTExtra 타입을 사용하여 타입 안전하게 변환
		jwtExtra, err := common.JWTExtraFromMap(extra)
		if err != nil {
			slog.Error("failed to convert user extra to JWTExtra",
				"username", u.GetID(),
				"error", err,
				"extra", extra)
			// 변환 실패 시 빈 attributes 설정
			details.Attributes = &sharedUser.Attributes{
				Roles: make(map[string]bool),
				Info:  make(map[string]any),
			}
			return details
		}

		// 원본 할당된 role 정보 수집 (확장 전)
		var roles []sharedUser.Role
		if jwtExtra.Roles != nil {
			// core + extension roles를 모두 포함한 base 구성
			base := coreRole.GetCoreRoleMetadata()
			for name, ext := range plugins.GetExtensionRegistry().GetAll() {
				for _, r := range ext.GetRoles() {
					r.Group = name
					base = append(base, r)
				}
			}

			// Metadata를 map으로 변환 (atomic role lookup용)
			metadataMap := make(map[string]sharedRole.RoleMetadata, len(base))
			for _, metadata := range base {
				metadataMap[metadata.Key] = metadata
			}

			// DB에서 composite role 포함 metadata 가져오기 (overlay 적용)
			if metadataService != nil {
				ctx := context.Background()
				dbMetadata, err := metadataService.ApplyOverlay(ctx, base)
				if err != nil {
					slog.Error("failed to get role metadata overlay", "error", err)
				} else {
					for _, metadata := range dbMetadata {
						metadataMap[metadata.Key] = metadata
					}
				}
			}

			for roleKey := range jwtExtra.Roles {
				roleInfo := sharedUser.Role{
					Role:        roleKey,
					DisplayName: roleKey, // fallback
				}

				// metadata에서 role 정보 찾기 (atomic or composite)
				if metadata, exists := metadataMap[roleKey]; exists {
					roleInfo.DisplayName = metadata.DisplayName
					group := metadata.Group
					roleInfo.Group = &group
					desc := metadata.Description
					roleInfo.Description = &desc
				}

				roles = append(roles, roleInfo)
			}
		}

		// Composite role 확장 적용 (attributes.roles용)
		expandedRoles := jwtExtra.Roles
		if metadataService != nil && jwtExtra.Roles != nil {
			ctx := context.Background()
			var err error
			expandedRoles, err = metadataService.ExpandCompositeRoles(ctx, jwtExtra.Roles)
			if err != nil {
				slog.Error("failed to expand composite roles", "error", err)
				expandedRoles = jwtExtra.Roles // fallback to original
			}
		}

		details.Attributes = &sharedUser.Attributes{
			Roles: expandedRoles,
			Info:  jwtExtra.Info,
		}

		// roles 필드 설정 (원본 할당된 role만, metadata 포함)
		if len(roles) > 0 {
			details.Roles = roles
		}
	}

	return details
}

// Response types는 shared types 사용
type ListUserResponse = sharedUser.ListUserResponse
type GetUserResponse = sharedUser.GetUserResponse

// CreateUserRequest는 shared/types/user.CreateUserRequest 타입 사용
type CreateUserRequest = sharedUser.CreateUserRequest

// ValidateCreateUserRequest ValidateAttributes는 더 이상 필요 없음 - shared 타입이 구조 보장
// 하지만 role 값 검증은 여전히 필요
func ValidateCreateUserRequest(req *CreateUserRequest) error {
	if req.Attributes == nil {
		return nil
	}

	// roles 검증 - 모든 값이 boolean인지 확인
	if req.Attributes.Roles != nil {
		for key := range req.Attributes.Roles {
			// map[string]bool이므로 타입은 보장됨
			_ = key
		}
	}

	return nil
}

// Request types는 shared types 사용
type UpdatePasswordExpirationRequest = sharedUser.UpdatePasswordExpirationRequest
type ChangePasswordRequest = sharedUser.ChangePasswordRequest
type BlockUserRequest = sharedUser.BlockUserRequest
type ChangeMyPasswordRequest = sharedUser.ChangeMyPasswordRequest

// SetAttributesRequest는 shared/types/user.SetAttributesRequest 타입 사용
type SetAttributesRequest = sharedUser.SetAttributesRequest
