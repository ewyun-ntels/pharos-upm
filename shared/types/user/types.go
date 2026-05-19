// Code generated from JSON Schema using quicktype. DO NOT EDIT.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    attributes, err := UnmarshalAttributes(bytes)
//    bytes, err = attributes.Marshal()
//
//    blockUserRequest, err := UnmarshalBlockUserRequest(bytes)
//    bytes, err = blockUserRequest.Marshal()
//
//    changeMyPasswordRequest, err := UnmarshalChangeMyPasswordRequest(bytes)
//    bytes, err = changeMyPasswordRequest.Marshal()
//
//    changePasswordRequest, err := UnmarshalChangePasswordRequest(bytes)
//    bytes, err = changePasswordRequest.Marshal()
//
//    createUserRequest, err := UnmarshalCreateUserRequest(bytes)
//    bytes, err = createUserRequest.Marshal()
//
//    getUserResponse, err := UnmarshalGetUserResponse(bytes)
//    bytes, err = getUserResponse.Marshal()
//
//    listUserResponse, err := UnmarshalListUserResponse(bytes)
//    bytes, err = listUserResponse.Marshal()
//
//    setAttributesRequest, err := UnmarshalSetAttributesRequest(bytes)
//    bytes, err = setAttributesRequest.Marshal()
//
//    updatePasswordExpirationRequest, err := UnmarshalUpdatePasswordExpirationRequest(bytes)
//    bytes, err = updatePasswordExpirationRequest.Marshal()
//
//    user, err := UnmarshalUser(bytes)
//    bytes, err = user.Marshal()
//
//    userMetadataConfig, err := UnmarshalUserMetadataConfig(bytes)
//    bytes, err = userMetadataConfig.Marshal()

package user

import "time"

import "encoding/json"

func UnmarshalAttributes(data []byte) (Attributes, error) {
	var r Attributes
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Attributes) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalBlockUserRequest(data []byte) (BlockUserRequest, error) {
	var r BlockUserRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *BlockUserRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalChangeMyPasswordRequest(data []byte) (ChangeMyPasswordRequest, error) {
	var r ChangeMyPasswordRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ChangeMyPasswordRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalChangePasswordRequest(data []byte) (ChangePasswordRequest, error) {
	var r ChangePasswordRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ChangePasswordRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalCreateUserRequest(data []byte) (CreateUserRequest, error) {
	var r CreateUserRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *CreateUserRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalGetUserResponse(data []byte) (GetUserResponse, error) {
	var r GetUserResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *GetUserResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalListUserResponse(data []byte) (ListUserResponse, error) {
	var r ListUserResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ListUserResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalSetAttributesRequest(data []byte) (SetAttributesRequest, error) {
	var r SetAttributesRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *SetAttributesRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalUpdatePasswordExpirationRequest(data []byte) (UpdatePasswordExpirationRequest, error) {
	var r UpdatePasswordExpirationRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *UpdatePasswordExpirationRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalUser(data []byte) (User, error) {
	var r User
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *User) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalUserMetadataConfig(data []byte) (UserMetadataConfig, error) {
	var r UserMetadataConfig
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *UserMetadataConfig) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Request body for blocking/unblocking user (PUT /user/{username}/block)
type BlockUserRequest struct {
	// Whether to block (true) or unblock (false) the user
	Block bool `json:"block"`
}

// Request body for changing current user's password (PUT /me/password)
type ChangeMyPasswordRequest struct {
	// Current password for verification
	CurrentPassword string `json:"current_password"`
	// New password
	NewPassword string `json:"new_password"`
}

// Request body for changing user password (PUT /user/{username}/password)
type ChangePasswordRequest struct {
	// New password for the user
	Password string `json:"password"`
}

// Request body for creating a new user (POST /user)
type CreateUserRequest struct {
	Attributes *Attributes `json:"attributes,omitempty"`
	// User's password
	Password string `json:"password"`
	// User's unique identifier (username)
	Username string `json:"username"`
}

// Structured user attributes - roles and info (shared/types/user.Attributes)
type Attributes struct {
	// User information (flexible data)
	Info map[string]interface{} `json:"info,omitempty"`
	// User roles with prefixes (role:*, attr:*). All values must be boolean.
	Roles map[string]bool `json:"roles,omitempty"`
}

// Response body for getting a single user (GET /user/{username})
type GetUserResponse struct {
	User User `json:"user"`
}

// User information from Backend GET /me API (core/pkg/userhandler/resources.go - Details
// struct)
type User struct {
	Attributes *Attributes `json:"attributes,omitempty"`
	// Whether the user is blocked
	Blocked *bool `json:"blocked,omitempty"`
	// User creation timestamp
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// Whether this user is the currently authenticated user (only in user list)
	IsMe *bool `json:"is_me,omitempty"`
	// User's unique identifier (username)
	Name string `json:"name"`
	// Password expiration timestamp
	PasswordExpiredAt *time.Time `json:"password_expired_at,omitempty"`
	// Preparation steps required (e.g., password_change)
	Prepare []string `json:"prepare,omitempty"`
	// Display labels for each prepare step, computed by the backend
	PrepareLabels []string `json:"prepare_labels,omitempty"`
	// List of user's roles (both atomic and composite)
	Roles []Role `json:"roles,omitempty"`
	// Temporary block expiration timestamp
	TemporaryBlockedExpiresAt *time.Time `json:"temporary_blocked_expires_at,omitempty"`
}

type Role struct {
	// Description of what this role allows
	Description *string `json:"description,omitempty"`
	// Display name of the composite role (e.g., 'Admin')
	DisplayName string `json:"display_name"`
	// Role group type for categorization (e.g., 'composite', 'role', 'permission')
	Group *string `json:"group,omitempty"`
	// Composite role key (e.g., 'role:tarzan_admin_role')
	Role string `json:"role"`
}

// Response body for listing users (GET /user)
type ListUserResponse struct {
	// List of users
	Users []User `json:"users,omitempty"`
}

// Request body for updating user attributes (PUT /user/{username}/attributes)
type SetAttributesRequest struct {
	Attributes Attributes `json:"attributes"`
}

// Request body for updating password expiration date (PUT
// /user/{username}/password-expire-date)
type UpdatePasswordExpirationRequest struct {
	// Password expiration timestamp (null to remove expiration)
	ExpiredAt *time.Time `json:"expired_at"`
}

// JSON Schema configuration for user additional info fields (runtime configurable)
type UserMetadataConfig struct {
	// JSONSchema definition for user additional fields
	Schema map[string]interface{} `json:"schema"`
	// React JSONSchema Form UI hints and customization
	UISchema map[string]interface{} `json:"uiSchema,omitempty"`
}
