package authhandler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/token/jwt"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/shared/types/role"
)

func GetDecodeAuthenticationHandler() func(c *gin.Context) {
	return func(c *gin.Context) {
		_, _, err := setContext(c)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

func GetAuthenticationHandler(abortOnFail bool, roles ...string) func(c *gin.Context) {
	return func(c *gin.Context) {
		check := func() int {
			token := extractToken(c.Request)

			// 로그인 세션 존재 여부 확인
			loggedIn := false
			if cookie, err := c.Cookie("logged_in"); err == nil && cookie == "true" {
				loggedIn = true
			}

			// 토큰 없는 경우
			if len(token) == 0 {
				// 로그인 세션이 있다면 토큰 만료로 판단 → 401 반환하여 refresh 트리거
				if loggedIn {
					return http.StatusUnauthorized
				}
				// 진짜 익명 사용자
				if !abortOnFail {
					return http.StatusOK
				}
				return http.StatusUnauthorized
			}

			_, jwtClaims, err := setContext(c)
			if err != nil {
				return http.StatusUnauthorized
			}

			if common.GetRole(jwtClaims, string(role.AttrTemporaryUser)) {
				return http.StatusUnauthorized
			}

			for _, r := range roles {
				if common.GetRole(jwtClaims, r) {
					return http.StatusOK
				}
			}
			if len(roles) != 0 {
				return http.StatusForbidden
			}

			return http.StatusOK
		}

		if statusCode := check(); statusCode != http.StatusOK {
			c.AbortWithStatus(statusCode)
			return
		}

		c.Next()
	}
}

func setContext(c *gin.Context) (*oauth2.JWTSession, *jwt.JWTClaims, error) {
	token := extractToken(c.Request)
	_, accessRequester, err := oauth2Provider.IntrospectToken(c.Request.Context(), token, fosite.AccessToken, nil)
	if err != nil {
		return nil, nil, err
	} else if accessRequester.GetSession() == nil {
		return nil, nil, errors.New("no session")
	}

	jwtSession := accessRequester.GetSession().(*oauth2.JWTSession)
	c.Set(common.ContextKeyJWTSession, jwtSession)

	jwtClaims := jwtSession.GetJWTClaims().(*jwt.JWTClaims)
	c.Set(common.ContextKeyJWTClaims, jwtClaims)

	return jwtSession, jwtClaims, nil
}

func extractToken(r *http.Request) string {
	// 1. Authorization Header 우선순위
	token := fosite.AccessTokenFromRequest(r)
	if token != "" {
		return token
	}

	// 2. Cookie fallback
	cookie, err := r.Cookie("access_token")
	if err == nil {
		return cookie.Value
	}

	return ""
}
