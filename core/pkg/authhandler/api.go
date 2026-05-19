package authhandler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	fositeOAuth2 "github.com/ory/fosite/handler/oauth2"
	"ntels.com/pharos/core/internal/auth"
	"ntels.com/pharos/core/internal/cert"
	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/factory"
	"ntels.com/pharos/core/internal/user"
	"ntels.com/pharos/core/pkg/common"
	coreRole "ntels.com/pharos/core/pkg/role"
	roleTypes "ntels.com/pharos/shared/types/role"
)

const JwtCaCrt = "jwt_ca.crt"
const JwtCrt = "jwt.crt"
const JwtKey = "jwt.key"

var oauth2Provider *auth.Oauth2Provider
var loginHistoryRepo repositories.LoginHistoryRepository
var metadataService *coreRole.MetadataService

type Api struct {
	configPath string
	config     common.Config
}

func (a *Api) Init(configPath string, config common.Config) {
	a.configPath = configPath
	a.config = config
}

func (a *Api) Use() bool {
	return true
}

func (a *Api) Load() error {
	// Initialize MetadataService for role expansion in JWT
	configRepo, err := factory.NewRoleMetadataConfigRepository(factory.RepositoryOptions{
		DatabaseConfig: a.config.Database,
	})
	if err != nil {
		slog.Error("failed to create role metadata repository", "error", err)
		return err
	}
	metadataService = coreRole.NewMetadataService(configRepo)

	userStore, err := user.NewStore(user.GetUserStoreConfig(a.config))
	if err != nil {
		return err
	}

	jwtKeyFile, err := a.getJwtKeyFile()
	if err != nil {
		slog.Error("failed to get jwt key file", "error", err)
		return err
	}

	clients := make([]auth.ClientConfig, 0)
	for _, client := range a.config.Auth.Clients {
		clients = append(clients, auth.ClientConfig{
			ID:     client.ID,
			Secret: client.Secret,
		})
	}

	loginHistoryRepo = factory.NewLoginHistoryRepository(factory.StatisticsRepositoryOptions{
		DatabaseConfig: a.config.Statistics.Database,
	})

	oauth2Provider, err = auth.NewProvider(userStore, auth.ProviderConfig{
		KeyFile:              jwtKeyFile,
		AccessTokenLifespan:  a.config.Auth.AccessTokenLifespan,
		RefreshTokenLifespan: a.config.Auth.RefreshTokenLifespan,
		GlobalSecret:         a.config.Auth.GlobalSecret,
		JWTIssuer:            a.config.Auth.JWTIssuer,
		StorageConfig: auth.StorageConfig{
			AllowedClients: a.config.Auth.AllowedClients,
			Clients:        clients,
			RefreshTokenRevokeCallback: func(ref *auth.RefreshToken) {
				if sid, startedAt := extractSessionClaims(ref); sid != "" {
					now := time.Now().UTC().Unix()
					sendSession(context.Background(), SessionRecord{
						SessionId: sid,
						UserId:    ref.GetSession().GetSubject(),
						StartedAt: startedAt,
						EndedAt:   &now,
						EndReason: "logout",
					})
				}
			},
			RefreshTokenTimeoutCallback: func(ref *auth.RefreshToken) {
				if sid, startedAt := extractSessionClaims(ref); sid != "" {
					now := time.Now().UTC().Unix()
					sendSession(context.Background(), SessionRecord{
						SessionId: sid,
						UserId:    ref.GetSession().GetSubject(),
						StartedAt: startedAt,
						EndedAt:   &now,
						EndReason: "expired",
					})
				}
			},
			RefreshTokenRemovedCallback: func(ref *auth.RefreshToken) {
				if sid, startedAt := extractSessionClaims(ref); sid != "" {
					now := time.Now().UTC().Unix()
					sendSession(context.Background(), SessionRecord{
						SessionId: sid,
						UserId:    ref.GetSession().GetSubject(),
						StartedAt: startedAt,
						EndedAt:   &now,
						EndReason: "revoked",
					})
				}
			},
		},
	})
	if err != nil {
		return err
	}

	closeOrphanedSessions(loginHistoryRepo, context.Background())

	return nil
}

func (a *Api) Unload() {}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	// 내부적으로 gin.IRouter 인터페이스가 필요하여 변환
	// 실제 사용중인 struct가 RouterGroup 라서 변환이 가능함
	// 함수를 변경하려면 다른 모듈의 Load 함수도 변경해야 하기 때문에 임시로 내부에서 변환함
	router, ok := routes.(gin.IRouter)
	if !ok {
		slog.Error("routes is not a gin.IRouter")
		return
	}

	router.GET("/config", GetConfigInfoHandler(a.config))
	router.Any("/token", GetTokenEndpointHandler(a.config))
	router.POST("/revoke", getTokenRevokeHandler(a.config)) // 쿠키 모드: 쿠키에서 토큰 읽고 삭제, REST: body에서 토큰
	router.GET("/sessions",
		GetAuthenticationHandler(true, string(roleTypes.RoleSuperAdmin), string(roleTypes.RoleUserRead)),
		sessions)
	router.DELETE("/sessions/:username",
		GetAuthenticationHandler(true, string(roleTypes.RoleSuperAdmin), string(roleTypes.RoleUserUpdate)),
		revokeSession)

	loginHistoryRepo := factory.NewLoginHistoryRepository(factory.StatisticsRepositoryOptions{
		DatabaseConfig: a.config.Statistics.Database,
	})
	router.GET("/history",
		GetAuthenticationHandler(true, string(roleTypes.RoleSuperAdmin), string(roleTypes.RoleUserRead)),
		getLoginHistoryHandler(loginHistoryRepo))

}

func (a *Api) GetRelativePath() string {
	return "/auth"
}

// extractSessionClaims은 refresh token의 JWT extra claims에서 session_id와 session_started_at을 추출합니다.
func extractSessionClaims(ref *auth.RefreshToken) (sessionId string, startedAt int64) {
	sess, ok := ref.GetSession().(*fositeOAuth2.JWTSession)
	if !ok || sess.JWTClaims == nil || sess.JWTClaims.Extra == nil {
		return "", 0
	}
	if v, ok := sess.JWTClaims.Extra["session_id"].(string); ok {
		sessionId = v
	}
	switch v := sess.JWTClaims.Extra["session_started_at"].(type) {
	case int64:
		startedAt = v
	case float64:
		startedAt = int64(v)
	}
	return
}

func (a *Api) getJwtKeyFile() (string, error) {
	if a.config.Auth.JwtCert.JwtDir == "" && a.config.Auth.JwtCert.JwtKeyFile == "" {
		return "", errors.New("jwt cert dir or key file is empty")
	}

	if a.config.Auth.JwtCert.JwtKeyFile != "" {
		return a.config.Auth.JwtCert.JwtKeyFile, nil
	}

	if a.config.Auth.JwtCert.AutoGenerate {
		keyFilePath := path.Join(path.Clean(a.config.Auth.JwtCert.JwtDir), JwtKey)
		if _, err := os.Stat(keyFilePath); err == nil {
			return keyFilePath, nil
		}

		if err := cert.GenerateJwtCertAuto(
			cert.Config{
				CommonName:                   a.config.Auth.JwtCert.AutoGenerateInfo.CommonName,
				Organization:                 a.config.Auth.JwtCert.AutoGenerateInfo.Organization,
				SubjectAlternateNameDnsNames: a.config.Auth.JwtCert.AutoGenerateInfo.SubjectAlternateNameDnsNames,
				SubjectAlternateNameIps:      a.config.Auth.JwtCert.AutoGenerateInfo.SubjectAlternateNameIps,
				NotAfter:                     a.config.Auth.JwtCert.AutoGenerateInfo.NotAfter,
				Destination:                  a.config.Auth.JwtCert.JwtDir,
				KeyFile:                      JwtKey,
				CaCrtFile:                    JwtCaCrt,
				CertFile:                     JwtCrt,
			}); err != nil {
			return "", fmt.Errorf("jwt cert auto generate failed: %w", err)
		}
		return keyFilePath, nil
	}

	return "", errors.New("jwt cert key file is empty")
}
