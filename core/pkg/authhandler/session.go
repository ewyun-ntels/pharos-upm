package authhandler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/internal/user"
	"ntels.com/pharos/core/pkg/common"
)

type Session struct {
	Username   string         `json:"username"`
	ClientId   string         `json:"client_id"`
	ClientIp   string         `json:"client_ip"`
	UserAgent  string         `json:"user_agent"`
	Attributes map[string]any `json:"attributes"`
	Prepare    []string       `json:"prepare"`
	Blocked    bool           `json:"blocked"`
	CreatedAt  time.Time      `json:"created_at"`
	ExpiresAt  time.Time      `json:"expires_at"`
}

func sessions(c *gin.Context) {
	refreshTokenSessions := oauth2Provider.Storage.GetRefreshTokenList()

	users, err := oauth2Provider.UserStore.ListUsers()
	if err != nil {
		slog.Error("Error occurred in ListUsers", "error", err)
		c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
		return
	}
	userMap := make(map[string]user.User)
	for _, v := range users {
		userMap[v.GetID()] = v
	}

	var result []Session
	for _, session := range refreshTokenSessions {
		resultSession := Session{
			Username:  session.GetSession().GetSubject(),
			ClientId:  session.GetClient().GetID(),
			ExpiresAt: session.GetSession().GetExpiresAt(fosite.RefreshToken),
		}
		// Extract client_ip and user_agent from JWT extra claims if present
		if jwtSess, ok := session.GetSession().(*oauth2.JWTSession); ok &&
			jwtSess.JWTClaims != nil && jwtSess.JWTClaims.Extra != nil {
			if v, ok := jwtSess.JWTClaims.Extra["client_ip"].(string); ok {
				resultSession.ClientIp = v
			}
			if v, ok := jwtSess.JWTClaims.Extra["user_agent"].(string); ok {
				resultSession.UserAgent = v
			}
		}
		if findUser, ok := userMap[session.GetSession().GetSubject()]; ok {
			resultSession.Blocked = findUser.GetBlocked()
			resultSession.Prepare = findUser.GetPrepare()
			resultSession.Attributes = findUser.GetExtra()
			resultSession.CreatedAt = findUser.GetCreatedAt()
		}
		result = append(result, resultSession)
	}

	c.JSON(http.StatusOK, result)
}

func revokeSession(c *gin.Context) {
	username := c.Param(common.ParamUsername)
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	if err := oauth2Provider.Storage.RemoveAccessTokenBySubject(username); err != nil {
		slog.Error("remove access token failed", "username", username, "error", err)
	}

	if err := oauth2Provider.Storage.RemoveRefreshTokenBySubject(username); err != nil {
		slog.Error("remove refresh token failed", "username", username, "error", err)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
