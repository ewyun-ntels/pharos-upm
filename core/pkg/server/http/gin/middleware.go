package gin

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/common"
)

func slogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.RequestURI
		c.Next() // 핸들러 실행

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("http request",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}

func slogRecovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic recovered",
					slog.Any("panic", rec),
					slog.String("path", c.Request.RequestURI),
				)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

// securityHeaders adds baseline security headers, especially for the embedded UI under /ui.
func securityHeaders() gin.HandlerFunc {
	// A moderate CSP that supports typical SPAs while adding protection.
	const csp = "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self' data:; connect-src 'self' ws: wss: http: https:; frame-ancestors 'self'; object-src 'none'; base-uri 'self'"
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Content-Security-Policy", csp)

		// Add HSTS only for HTTPS requests (or when behind a TLS-terminating proxy)
		// to avoid accidentally enforcing HTTPS on plain HTTP during local/dev.
		if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
			// 1 year max-age and include subdomains as a sensible default.
			// We intentionally omit preload unless explicitly requested/configured later.
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// csrfMiddleware protects against Cross-Site Request Forgery.
// It is only active if Auth.CsrfProtection is enabled and the request uses cookie authentication.
func csrfMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		config := common.GetGlobalConfig()
		if config == nil || !config.Auth.CsrfProtection {
			c.Next()
			return
		}

		// GET, HEAD, OPTIONS, TRACE는 보통 안전하다고 판단하여 제외 (Idempotent)
		method := c.Request.Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" || method == "TRACE" {
			c.Next()
			return
		}

		// Authorization 헤더가 있으면 토큰 방식이므로 CSRF 검증 제외
		if c.GetHeader("Authorization") != "" {
			c.Next()
			return
		}

		// 인증 관련 경로는 제외 (로그인 시 쿠키가 이미 있는 경우 대비)
		path := c.Request.URL.Path
		if strings.Contains(path, "/auth/token") || strings.Contains(path, "/auth/revoke") {
			c.Next()
			return
		}

		// 쿠키가 없으면 어차피 인증에 실패할 것이므로 제외
		if _, err := c.Cookie("access_token"); err != nil {
			c.Next()
			return
		}

		// CSRF 토큰 검증
		// 간단하게 구현: X-CSRF-Token 헤더가 있는지 확인 (쿠키 기반 인증 시 커스텀 헤더 존재 여부만으로도 많은 방어가 됨)
		// 실제로는 세션에 저장된 값과 비교하는 것이 정석이나, 브라우저 정책상 커스텀 헤더 추가는 Same-Origin에서만 가능하므로 1차 방어선이 됨.
		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token missing"})
			return
		}

		c.Next()
	}
}
