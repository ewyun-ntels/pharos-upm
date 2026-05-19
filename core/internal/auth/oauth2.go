package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/token/jwt"
	"ntels.com/pharos/core/internal/user"
)

func getKeyFromFile(filename string) (*rsa.PrivateKey, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(f)

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("invalid key type")
	}

	return key, nil
}

func NewProvider(userStore user.Store, cfg ProviderConfig) (*Oauth2Provider, error) {
	privateKey, err := getKeyFromFile(cfg.KeyFile)
	if err != nil {
		return nil, err
	}

	fositeConfig := &fosite.Config{
		AccessTokenLifespan:  cfg.AccessTokenLifespan,
		RefreshTokenLifespan: cfg.RefreshTokenLifespan,
		RefreshTokenScopes:   []string{},
		GlobalSecret:         []byte(cfg.GlobalSecret),
		AccessTokenIssuer:    cfg.JWTIssuer,
	}

	storage, err := newStorage(userStore, cfg.StorageConfig)
	if err != nil {
		return nil, err
	}

	provider := compose.Compose(
		fositeConfig,
		storage,
		&oauth2.DefaultJWTStrategy{
			Signer: &jwt.DefaultSigner{
				GetPrivateKey: func(context.Context) (any, error) {
					return privateKey, nil
				},
			},
			HMACSHAStrategy: compose.NewOAuth2HMACStrategy(fositeConfig),
			Config:          fositeConfig,
		},
		compose.OAuth2TokenIntrospectionFactory,
		OAuth2ResourceOwnerPasswordCredentialsFactoryLocal, // password 인증 (local replica)
		compose.OAuth2RefreshTokenGrantFactory,             // refresh 활성
		compose.OAuth2StatelessJWTIntrospectionFactory,     // 토큰 인증 활성(stateless): strategy implement jwt.Signer
		compose.OAuth2TokenRevocationFactory,
		compose.OAuth2ClientCredentialsGrantFactory,
	)

	return &Oauth2Provider{
		UserStore:      userStore,
		Storage:        storage,
		OAuth2Provider: provider,
	}, nil
}

type Oauth2Provider struct {
	UserStore user.Store
	Storage   *Storage
	fosite.OAuth2Provider
}
