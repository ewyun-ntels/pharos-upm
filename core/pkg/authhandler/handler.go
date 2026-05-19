package authhandler

import (
	"context"
	"errors"
	"log/slog"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/token/jwt"
	"ntels.com/pharos/core/internal/auth"
	"ntels.com/pharos/core/internal/casbin"
	"ntels.com/pharos/core/internal/user"
	"ntels.com/pharos/core/pkg/authhandler/adapters"
	"ntels.com/pharos/core/pkg/common"
)

// CreateUser creates a user using the globally initialized userStore.
// It no longer returns a cleanup function; deletion should be performed by calling DeleteUser.
func CreateUser(username, password string, attrs map[string]any, prepare *[]string) error {
	if oauth2Provider.UserStore == nil {
		return errors.New("auth: user store not initialized")
	}

	// Try to create. If it already exists (e.g., previous run crashed before cleanup),
	// attempt to delete and retry once.
	if err := oauth2Provider.UserStore.CreateUser(username, password, attrs, prepare); err != nil {
		if errors.Is(err, user.ErrUserAlreadyExists) {
			slog.Info("auth: user already exists, attempting to delete and recreate", "user", username)
			if derr := oauth2Provider.UserStore.DeleteUser(username); derr != nil {
				slog.Info("auth: delete pre-existing user failed (will retry create anyway)", "user", username, "error", derr)
			}
			if err2 := oauth2Provider.UserStore.CreateUser(username, password, attrs, prepare); err2 != nil {
				return err2
			}
		} else {
			return err
		}
	}
	return nil
}

// DeleteUser deletes a user using the globally initialized userStore.
func DeleteUser(username string) error {
	if oauth2Provider.UserStore == nil {
		return errors.New("auth: user store not initialized")
	}
	if err := oauth2Provider.UserStore.DeleteUser(username); err != nil {
		return err
	}
	return nil
}

// logTokenError logs token-related errors with consistent context information
func logTokenError(c *gin.Context, accessRequest fosite.AccessRequester, grantType, stage string, err error) {
	username := ""

	// Try to extract username from form first (password grant)
	if u, exists := c.GetPostForm("username"); exists && u != "" {
		username = u
	} else if accessRequest != nil && accessRequest.GetSession() != nil {
		// Extract from session (refresh_token grant)
		username = accessRequest.GetSession().GetSubject()
	}

	slog.Error(stage,
		"error", err,
		"username", username,
		"client_ip", c.ClientIP(),
		"grant_type", grantType)
}

func getTokenRevokeHandler(config common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 쿠키 모드: 쿠키에서 토큰 읽어서 무효화 + 쿠키 삭제
		if config.Auth.CookieMode {
			if refreshToken, err := c.Cookie("refresh_token"); err == nil && refreshToken != "" {
				// fosite의 NewRevocationRequest는 client 인증이 필요하지만
				// cookie 모드에서는 client_id/secret 없이 요청하므로 스토리지에서 직접 처리한다.
				// HMAC 토큰 형식: "{payload}.{signature}" — 마지막 점 이후가 스토리지 키
				if idx := strings.LastIndex(refreshToken, "."); idx >= 0 {
					sig := refreshToken[idx+1:]
					tmpSess := &oauth2.JWTSession{JWTClaims: &jwt.JWTClaims{}}
					if req, sessErr := oauth2Provider.Storage.GetRefreshTokenSession(ctx, sig, tmpSess); sessErr == nil {
						_ = oauth2Provider.Storage.RevokeRefreshToken(ctx, req.GetID())
					}
				}
			}

			// 쿠키 삭제
			c.SetCookie("access_token", "", -1, "/", "", config.Auth.SecureCookies, true)
			c.SetCookie("refresh_token", "", -1, "/auth", "", config.Auth.SecureCookies, true)
			c.SetCookie("logged_in", "", -1, "/", "", config.Auth.SecureCookies, false)

			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}

		// REST API 모드: 기존 방식
		err := oauth2Provider.NewRevocationRequest(ctx, c.Request)
		if err != nil && !errors.Is(err, fosite.ErrInvalidRequest) &&
			!errors.Is(err, fosite.ErrInvalidClient) {
			slog.Error("revoke request error", "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		oauth2Provider.WriteRevocationResponse(ctx, c.Writer, err)
	}
}

// clearLoggedInCookieOnRefreshFailure clears the logged_in cookie when refresh_token request fails
func clearLoggedInCookieOnRefreshFailure(c *gin.Context, config common.Config, authType, grantType string) {
	if config.Auth.CookieMode && authType == "cookie" && grantType == "refresh_token" {
		c.SetCookie("logged_in", "", -1, "/", "", config.Auth.SecureCookies, false)
	}
}

func GetTokenEndpointHandler(config common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This context will be passed to all methods.
		ctx := c.Request.Context()

		// Cookie 모드에서 refresh_token 요청 시 쿠키에서 토큰 읽기
		authType := c.GetHeader("X-Auth-Type")
		if authType == "" {
			authType = c.PostForm("X-Auth-Type")
		}
		grantType := c.PostForm("grant_type")

		if config.Auth.CookieMode && authType == "cookie" && grantType == "refresh_token" {
			// 쿠키에서 refresh_token 읽어서 form에 주입
			if refreshToken, err := c.Cookie("refresh_token"); err == nil && refreshToken != "" {
				c.Request.PostForm.Set("refresh_token", refreshToken)
			}
		}

		// Create an empty session object which will be passed to the request handlers
		mySessionData := &oauth2.JWTSession{
			JWTClaims: &jwt.JWTClaims{},
		}

		// This will create an access request object and iterate through the registered TokenEndpointHandlers to validate the request.
		accessRequest, err := oauth2Provider.NewAccessRequest(ctx, c.Request, mySessionData)

		// Catch any errors, e.g.:
		// * unknown client
		// * invalid redirect
		// * ...
		if err != nil {
			logTokenError(c, accessRequest, grantType, "Error occurred in NewAccessRequest", err)

			// refresh_token 요청 실패 시 logged_in 쿠키 삭제 (세션 만료)
			clearLoggedInCookieOnRefreshFailure(c, config, authType, grantType)

			var fositeErr *fosite.RFC6749Error
			if errors.As(err, &fositeErr) {
				resultErr := fositeErr.Cause()
				if resultErr == nil {
					resultErr = fositeErr
				}

				// grant_type=password 인 경우에만 hint 적용
				if accessRequest != nil && accessRequest.GetGrantTypes().Has(string(fosite.GrantTypePassword)) {
					hintLevel := config.Auth.AuthenticateHint
					if hintLevel < common.AuthenticateHintAll {
						var rfcErr *fosite.RFC6749Error
						if errors.As(resultErr, &rfcErr) {
							if hintLevel == common.AuthenticateHintBlock && strings.Contains(rfcErr.DescriptionField, "blocked") {
								resultErr = auth.ErrUserBlocked
							} else {
								resultErr = auth.ErrInvalidPassword
							}
						}
					}
				}

				oauth2Provider.WriteAccessError(ctx, c.Writer, accessRequest, resultErr)

				var sendErr *fosite.RFC6749Error
				if errors.As(fositeErr.Cause(), &sendErr) {
					username, _ := c.GetPostForm("username")
					// refresh_token grant: form에 username이 없으므로 파싱된 세션에서 subject 추출
					if username == "" && accessRequest != nil && accessRequest.GetSession() != nil {
						username = accessRequest.GetSession().GetSubject()
					}
					// accessRequest 세션에서도 못 가져온 경우 mySessionData에서 직접 추출
					if username == "" && mySessionData.JWTClaims != nil {
						username = mySessionData.JWTClaims.Subject
					}

					now := time.Now().UTC().Unix()
					if grantType == "refresh_token" && errors.Is(fositeErr.Cause(), fosite.ErrTokenExpired) {
						// refresh_token 만료: mySessionData에 복사된 session_id로 세션 종료 처리
						if mySessionData.JWTClaims != nil && mySessionData.JWTClaims.Extra != nil {
							if sid, ok := mySessionData.JWTClaims.Extra["session_id"].(string); ok && sid != "" {
								startedAt, _ := mySessionData.JWTClaims.Extra["session_started_at"].(int64)
								sendSession(ctx, SessionRecord{
									SessionId: sid,
									UserId:    username,
									StartedAt: startedAt,
									EndedAt:   &now,
									EndReason: "expired",
								})
							}
						}
					} else {
						// 로그인 실패(비밀번호 오류, 계정 잠금 등): 즉시 종료된 기록으로 저장
						sendSession(ctx, SessionRecord{
							SessionId: uuid.NewString(),
							UserId:    username,
							ClientIp:  c.ClientIP(),
							UserAgent: c.GetHeader("User-Agent"),
							StartedAt: now,
							EndedAt:   &now,
							EndReason: "failed: " + sendErr.DescriptionField,
						})
					}
				}
				return
			}
		}

		if err := performByGrantType(accessRequest, c.ClientIP(), c.GetHeader("User-Agent"), config.Auth.CollectClientInfo); err != nil {
			logTokenError(c, accessRequest, grantType, "Error occurred in performByGrantType", err)
			// refresh_token 요청 실패 시 logged_in 쿠키 삭제
			clearLoggedInCookieOnRefreshFailure(c, config, authType, grantType)
			oauth2Provider.WriteAccessError(ctx, c.Writer, accessRequest, fosite.ErrServerError.WithWrap(err))
			return
		}

		// Next, we create a response for the access request. Again, we iterate through the TokenEndpointHandlers
		// and aggregate the result in response.
		response, err := oauth2Provider.NewAccessResponse(ctx, accessRequest)
		if err != nil {
			logTokenError(c, accessRequest, grantType, "Error occurred in NewAccessResponse", err)
			// refresh_token 요청 실패 시 logged_in 쿠키 삭제
			clearLoggedInCookieOnRefreshFailure(c, config, authType, grantType)
			oauth2Provider.WriteAccessError(ctx, c.Writer, accessRequest, err)
			return
		}

		// All done, send the response.

		if config.Auth.CookieMode && authType == "cookie" {
			setTokenCookie(c, config, response)
			// Cookie 모드: 토큰은 쿠키로만 전달, body에는 성공 여부만
			c.JSON(http.StatusOK, gin.H{
				"success":    true,
				"expires_in": int(config.Auth.AccessTokenLifespan.Seconds()),
			})
		} else {
			// REST API 모드: 기존대로 body에 토큰 전달
			oauth2Provider.WriteAccessResponse(ctx, c.Writer, accessRequest, response)
		}

		if accessRequest.GetGrantTypes().ExactOne(string(fosite.GrantTypePassword)) {
			subject := accessRequest.GetSession().GetSubject()
			if jwtSess, ok := accessRequest.GetSession().(*oauth2.JWTSession); ok &&
				jwtSess.JWTClaims != nil && jwtSess.JWTClaims.Extra != nil {
				if sid, ok := jwtSess.JWTClaims.Extra["session_id"].(string); ok && sid != "" {
					startedAt, _ := jwtSess.JWTClaims.Extra["session_started_at"].(int64)
					sendSession(ctx, SessionRecord{
						SessionId: sid,
						UserId:    subject,
						ClientIp:  c.ClientIP(),
						UserAgent: c.GetHeader("User-Agent"),
						StartedAt: startedAt,
					})
				}
			}
		}
	}
}

func setTokenCookie(c *gin.Context, config common.Config, response fosite.AccessResponder) {
	accessToken := response.GetAccessToken()
	if accessToken == "" {
		return
	}

	// Calculate max age from lifespan
	maxAge := int(config.Auth.AccessTokenLifespan.Seconds())

	c.SetCookie(
		"access_token",
		accessToken,
		maxAge,
		"/",
		"", // domain
		config.Auth.SecureCookies,
		true, // httpOnly
	)

	// Refresh token도 쿠키에 저장 (필요 시)
	refreshToken := response.ToMap()["refresh_token"]
	if refreshToken != nil {
		rfTokenStr, ok := refreshToken.(string)
		if ok && rfTokenStr != "" {
			rfMaxAge := int(config.Auth.RefreshTokenLifespan.Seconds())
			c.SetCookie(
				"refresh_token",
				rfTokenStr,
				rfMaxAge,
				"/auth", // /auth/* 경로에서 쿠키 전송 (/auth/token, /auth/revoke 모두 포함)
				"",
				config.Auth.SecureCookies,
				true,
			)

			// 세션 존재 표시 쿠키 (access_token 만료 시에도 로그인 상태 판별용)
			// refresh_token과 같은 lifespan, path=/ (모든 API에서 확인 가능)
			c.SetCookie(
				"logged_in",
				"true",
				rfMaxAge,
				"/",
				"",
				config.Auth.SecureCookies,
				false, // httpOnly=false (CSRF 토큰이 보호)
			)
		}
	}
}

func performByGrantType(accessRequester fosite.AccessRequester, clientIp, userAgent string, collectClientInfo bool) error {
	grantTypes := []fosite.GrantType{
		fosite.GrantTypePassword,
		fosite.GrantTypeClientCredentials,
	}
	for _, grantType := range grantTypes {
		if !accessRequester.GetGrantTypes().ExactOne(string(grantType)) {
			continue
		}

		switch grantType {
		case fosite.GrantTypePassword:
			if err := setExtraForUser(accessRequester, clientIp, userAgent, collectClientInfo); err != nil {
				return err
			}
		case fosite.GrantTypeClientCredentials:
			// If this is a client_credentials grant, grant all requested scopes
			// NewAccessRequest validated that all requested scopes the client is allowed to perform
			// based on configured scope matching strategy.
			for _, scope := range accessRequester.GetRequestedScopes() {
				accessRequester.GrantScope(scope)
			}
		}
	}

	return nil
}

func setExtraForUser(accessRequester fosite.AccessRequester, clientIp, userAgent string, collectClientInfo bool) error {
	if _, ok := accessRequester.GetSession().(*oauth2.JWTSession); !ok {
		return nil
	}

	claims := accessRequester.GetSession().(*oauth2.JWTSession).JWTClaims
	claims.Subject = accessRequester.GetSession().GetSubject()
	if claims.Extra == nil {
		claims.Extra = make(map[string]any)
	}

	u, err := oauth2Provider.UserStore.FindByUsername(claims.Subject)
	if err != nil {
		slog.Error("Error occurred in FindByUsername", "error", err, "username", claims.Subject)
		return err
	}

	attrs := u.GetExtra()
	attrs, err = oauth2Provider.UserStore.DecodeExtra(attrs)
	if err != nil {
		slog.Error("Error occurred in DecodeExtra", "error", err, "username", claims.Subject)
		return err
	}

	// Expand composite roles using MetadataService for JWT
	if metadataService != nil {
		if rolesMap, ok := attrs[common.ExtraKeyRoles].(map[string]any); ok {
			// Convert to map[string]bool for expansion
			rolesBool := make(map[string]bool)
			for k, v := range rolesMap {
				if boolVal, ok := v.(bool); ok {
					rolesBool[k] = boolVal
				}
			}

			// Expand composite roles
			expandedRoles, err := metadataService.ExpandCompositeRoles(context.Background(), rolesBool)
			if err != nil {
				slog.Warn("Failed to expand composite roles for JWT", "error", err, "username", claims.Subject)
				// Continue with unexpanded roles on error
			} else {
				// Convert back to map[string]interface{}
				expandedMap := make(map[string]any)
				for k, v := range expandedRoles {
					expandedMap[k] = v
				}
				attrs[common.ExtraKeyRoles] = expandedMap
			}
		}
	}

	prepare, err := oauth2Provider.UserStore.GetPrepare(u)
	if err != nil {
		return err
	}

	// Handle prepare flags - add to roles
	if len(prepare) > 0 {
		if rolesMap, ok := attrs[common.ExtraKeyRoles].(map[string]any); ok {
			if def, ok := user.GetActionRoleDefinition(prepare[0]); ok {
				for _, r := range def.RuntimeRoles {
					rolesMap[string(r)] = true
				}
			}
		}
	}

	// Add groups array to top level (not in roles or info)
	casbin.SetAttributesAboutGroups(claims.Subject, attrs)

	// Copy all attrs to Extra (attrs already has structure: {roles, info, groups, ...})
	maps.Copy(claims.Extra, attrs)

	if u.GetPasswordExpiresAt() != nil {
		claims.Extra["password_expired_at"] = u.GetPasswordExpiresAt().Unix()
	}

	// 최초 로그인 시 session_id를 생성하여 JWT에 포함
	// refresh_token 갱신 시 fosite가 기존 session을 그대로 유지하므로 session_id가 보존됨
	claims.Extra["session_id"] = uuid.NewString()
	claims.Extra["session_started_at"] = time.Now().UTC().Unix()

	if collectClientInfo {
		claims.Extra["client_ip"] = clientIp
		claims.Extra["user_agent"] = userAgent
	}

	return nil
}

func ClearOwnToken(c *gin.Context) {
	var jwtClaims *jwt.JWTClaims
	if value, exists := c.Get(common.ContextKeyJWTClaims); exists {
		jwtClaims = value.(*jwt.JWTClaims)
	} else {
		jwtClaims = &jwt.JWTClaims{}
	}

	sub := jwtClaims.Subject
	if sub != "" {
		err := oauth2Provider.Storage.RemoveAccessTokenBySubject(sub)
		if err != nil {
			slog.Error("remove access token failed", "error", err)
		}

		err = oauth2Provider.Storage.RemoveRefreshTokenBySubject(sub)
		if err != nil {
			slog.Error("remove refresh token failed", "error", err)
		}
	}
}

func ClearParamToken(c *gin.Context) {
	username := c.Param(common.ParamUsername)
	if username == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	err := oauth2Provider.Storage.RemoveAccessTokenBySubject(username)
	if err != nil {
		slog.Error("remove access token failed", "error", err)
	}

	err = oauth2Provider.Storage.RemoveRefreshTokenBySubject(username)
	if err != nil {
		slog.Error("remove refresh token failed", "error", err)
	}
}

func GetConfigInfoHandler(config common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		adapter := &adapters.AuthAdapter{}
		result := adapter.ToAPIAuthConfig(config)

		c.JSON(http.StatusOK, result)
	}
}
