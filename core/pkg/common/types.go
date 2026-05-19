package common

import (
	"encoding/json"
	"log/slog"

	"github.com/go-viper/mapstructure/v2"
)

const (
	ServerTypeHttp  = "http"
	ServerTypeHttps = "https"
	ServerTypeNats  = "nats"

	ClientTypeNats = "nats"

	SchemaHttp  = "http"
	SchemaHttps = "https"
	SchemaNats  = "nats"
	SchemaTls   = "tls"
)

const (
	ParamUsername = "username"
)

const (
	ContextKeyJWTSession = "JWTSession"
	ContextKeyJWTClaims  = "JWTClaims"

	// JWT Extra keys
	ExtraKeyRoles  = "roles"
	ExtraKeyInfo   = "info"
	ExtraKeyGroups = "groups"

	// Deprecated: Use ExtraKeyGroups instead
	AttributeKeyGroupName = "group_name"
)

// JWTExtra JWTExtra는 JWT claims의 Extra 필드 구조를 정의합니다.
// DB의 account.extra 컬럼과 동일한 구조를 사용합니다.
type JWTExtra struct {
	// Roles는 사용자의 역할 맵 (role:* 형식의 키와 bool 값)
	Roles map[string]bool `json:"roles" mapstructure:"roles"`
	// Info는 사용자의 추가 속성 맵 (임의의 키-값 쌍)
	Info map[string]any `json:"info" mapstructure:"info"`
	// Groups는 사용자가 속한 그룹 목록
	Groups []string `json:"groups,omitempty" mapstructure:"groups"`
}

// ToMap ToMap은 JWTExtra를 map[string]interface{}로 변환합니다.
// fosite의 JWTClaims.Extra는 map[string]interface{} 타입을 사용하기 때문입니다.
// json.Marshal/Unmarshal을 사용하여 안전하게 변환합니다.
func (e *JWTExtra) ToMap() (map[string]any, error) {
	// JSON으로 직렬화 후 map으로 역직렬화
	// 이 방법이 수동 변환보다 안전하고 타입 변환이 정확함
	data, err := json.Marshal(e)
	if err != nil {
		slog.Error("failed to marshal JWTExtra to JSON",
			"error", err,
			"extra", e)
		return nil, err
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		slog.Error("failed to unmarshal JSON to map",
			"error", err,
			"json", string(data))
		return nil, err
	}

	return m, nil
}

// JWTExtraFromMap JWTExtraFromMap은 map[string]interface{}에서 JWTExtra를 생성합니다.
// mapstructure v2를 사용하여 안전하게 변환합니다.
func JWTExtraFromMap(m map[string]any) (*JWTExtra, error) {
	if m == nil {
		return &JWTExtra{
			Roles: make(map[string]bool),
			Info:  make(map[string]any),
		}, nil
	}

	extra := &JWTExtra{}

	// mapstructure v2를 사용한 안전한 변환
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           extra,
		WeaklyTypedInput: true, // 타입 변환 허용 (예: interface{} -> bool)
		TagName:          "mapstructure",
	})

	if err != nil {
		slog.Error("failed to create mapstructure decoder for JWTExtra",
			"error", err,
			"input", m)
		return nil, err
	}

	if err := decoder.Decode(m); err != nil {
		slog.Error("failed to decode map to JWTExtra",
			"error", err,
			"input", m)
		return nil, err
	}

	// nil 방지
	if extra.Roles == nil {
		extra.Roles = make(map[string]bool)
	}
	if extra.Info == nil {
		extra.Info = make(map[string]any)
	}

	return extra, nil
}

// HasRole HasRole은 extra map에서 특정 role의 존재 여부를 확인합니다.
// DB의 account.extra 또는 user.GetExtra() 결과에 사용됩니다.
func HasRole(extra map[string]any, roleKey string) bool {
	if extra == nil {
		return false
	}

	// JWTExtraFromMap을 사용하여 타입 안전하게 변환
	jwtExtra, err := JWTExtraFromMap(extra)
	if err != nil {
		slog.Error("failed to convert extra to JWTExtra in HasRole",
			"roleKey", roleKey,
			"error", err,
			"extra", extra)
		return false
	}
	return jwtExtra.Roles[roleKey]
}
