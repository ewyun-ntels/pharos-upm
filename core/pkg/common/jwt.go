package common

import (
	"log/slog"

	"github.com/ory/fosite/token/jwt"
)

// GetJWTExtra GetJWTExtra는 JWT claims에서 타입이 안전한 JWTExtra 구조체를 반환합니다.
// 에러 발생 시 빈 JWTExtra를 반환합니다.
func GetJWTExtra(jwtClaims *jwt.JWTClaims) *JWTExtra {
	if jwtClaims == nil || jwtClaims.Extra == nil {
		return &JWTExtra{
			Roles: make(map[string]bool),
			Info:  make(map[string]any),
		}
	}

	extra, err := JWTExtraFromMap(jwtClaims.Extra)
	if err != nil {
		slog.Error("failed to convert JWT claims extra to JWTExtra",
			"subject", jwtClaims.Subject,
			"error", err,
			"extra", jwtClaims.Extra)
		// 변환 실패 시 빈 구조체 반환
		return &JWTExtra{
			Roles: make(map[string]bool),
			Info:  make(map[string]any),
		}
	}

	return extra
}

// GetRole checks if user has a specific role or attribute
// Structure: {"roles": {"role:*": true, "attr:*": true}, ...}
// Returns true if the role/attr exists and is true
func GetRole(jwtClaims *jwt.JWTClaims, key string) bool {
	extra := GetJWTExtra(jwtClaims)
	if extra.Roles == nil {
		return false
	}
	return extra.Roles[key]
}

// GetGroups returns the list of groups the user belongs to
// Structure: {"groups": ["group1", "group2"], ...}
func GetGroups(jwtClaims *jwt.JWTClaims) []string {
	extra := GetJWTExtra(jwtClaims)
	return extra.Groups
}
