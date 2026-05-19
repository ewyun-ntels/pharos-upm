package auth

import "time"

type ProviderConfig struct {
	KeyFile              string
	AccessTokenLifespan  time.Duration
	RefreshTokenLifespan time.Duration
	GlobalSecret         string
	JWTIssuer            string
	StorageConfig        StorageConfig
}

type StorageConfig struct {
	AllowedClients []string
	Clients        []ClientConfig

	RefreshTokenRevokeCallback  func(ref *RefreshToken)
	RefreshTokenTimeoutCallback func(ref *RefreshToken)
	RefreshTokenRemovedCallback func(ref *RefreshToken)
}

type ClientConfig struct {
	ID     string
	Secret string
}
