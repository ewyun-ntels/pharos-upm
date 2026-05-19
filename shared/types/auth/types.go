// Code generated from JSON Schema using quicktype. DO NOT EDIT.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    authConfig, err := UnmarshalAuthConfig(bytes)
//    bytes, err = authConfig.Marshal()
//
//    passwordPolicy, err := UnmarshalPasswordPolicy(bytes)
//    bytes, err = passwordPolicy.Marshal()
//
//    userPolicy, err := UnmarshalUserPolicy(bytes)
//    bytes, err = userPolicy.Marshal()

package auth

import "encoding/json"

func UnmarshalAuthConfig(data []byte) (AuthConfig, error) {
	var r AuthConfig
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AuthConfig) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalPasswordPolicy(data []byte) (PasswordPolicy, error) {
	var r PasswordPolicy
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *PasswordPolicy) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalUserPolicy(data []byte) (UserPolicy, error) {
	var r UserPolicy
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *UserPolicy) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Authentication configuration from GET /auth/config
type AuthConfig struct {
	// Common authentication settings
	Common   Common         `json:"common"`
	Password PasswordPolicy `json:"password"`
	User     UserPolicy     `json:"user"`
}

// Common authentication settings
type Common struct {
	// Whether to collect and expose client IP and User-Agent
	CollectClientInfo *bool `json:"collect_client_info,omitempty"`
	// Duration in seconds for temporary account block
	LoginBlockDuration int64 `json:"login_block_duration"`
	// Time in seconds to reset login retry count
	LoginRetryResetTimeout int64 `json:"login_retry_reset_timeout"`
}

// Password policy configuration from GET /auth/config
type PasswordPolicy struct {
	// Minimum digit characters required
	MinDigits int64 `json:"min_digits"`
	// Minimum password length
	MinLength int64 `json:"min_length"`
	// Minimum lowercase characters required
	MinLowercase int64 `json:"min_lowercase"`
	// Minimum special characters required
	MinSpecial int64 `json:"min_special"`
	// Minimum uppercase characters required
	MinUppercase int64 `json:"min_uppercase"`
	// Password time-to-live in seconds
	PasswordTTL *int64 `json:"password_ttl,omitempty"`
}

// User policy configuration (username rules) from GET /auth/config
type UserPolicy struct {
	// Whether email format is allowed for username
	EmailAllowed bool `json:"email_allowed"`
	// Maximum username length
	MaxLength int64 `json:"max_length"`
	// Minimum username length
	MinLength int64 `json:"min_length"`
}
