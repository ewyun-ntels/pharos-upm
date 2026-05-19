package gin

import (
	"bytes"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/pkg/common"
)

func TestSlogMiddleware_ConcurrentRequests(t *testing.T) {
	t.Parallel()

	// 테스트용 버퍼와 로거 생성
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Gin 엔진 설정
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(slogMiddleware(logger))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	const numRequests = 20
	var wg sync.WaitGroup
	results := make([]int, numRequests)

	wg.Add(numRequests)

	for i := range numRequests {
		go func(id int) {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			results[id] = w.Code
		}(i)
	}

	wg.Wait()

	// 모든 요청이 성공했는지 확인
	for i, code := range results {
		assert.Equal(t, http.StatusOK, code, "Request %d should return 200", i)
	}
}

func TestSlogRecovery_ConcurrentPanics(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(slogRecovery(logger))

	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	const numRequests = 10
	var wg sync.WaitGroup
	results := make([]int, numRequests)

	wg.Add(numRequests)

	for i := range numRequests {
		go func(id int) {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/panic", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			results[id] = w.Code
		}(i)
	}

	wg.Wait()

	// 모든 패닉이 적절히 처리되었는지 확인
	for i, code := range results {
		assert.Equal(t, http.StatusInternalServerError, code, "Request %d should return 500 after panic", i)
	}
}

func TestSecurityHeaders_ConcurrentRequests(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(securityHeaders())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	const numRequests = 15
	var wg sync.WaitGroup
	headers := make([]map[string][]string, numRequests)

	wg.Add(numRequests)

	for i := range numRequests {
		go func(id int) {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			headers[id] = make(map[string][]string)
			maps.Copy(headers[id], w.Header())
		}(i)
	}

	wg.Wait()

	// 모든 응답에서 보안 헤더가 설정되었는지 확인
	for i, header := range headers {
		assert.Contains(t, header, "X-Content-Type-Options", "Request %d should have X-Content-Type-Options header", i)
		assert.Contains(t, header, "Content-Security-Policy", "Request %d should have Content-Security-Policy header", i)
		assert.Equal(t, []string{"nosniff"}, header["X-Content-Type-Options"], "Request %d should have correct X-Content-Type-Options value", i)
	}
}

func TestSecurityHeaders_HTTPSDetection(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(securityHeaders())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	tests := []struct {
		name           string
		forwardedProto string
		expectHSTS     bool
	}{
		{"HTTP request", "", false},
		{"HTTPS via proxy", "https", true},
		{"HTTPS via proxy (case insensitive)", "HTTPS", true},
		{"HTTP via proxy", "http", false},
	}

	var wg sync.WaitGroup
	results := make([]map[string]bool, len(tests))

	wg.Add(len(tests))

	for i, tt := range tests {
		go func(id int, test struct {
			name           string
			forwardedProto string
			expectHSTS     bool
		}) {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/test", nil)
			if test.forwardedProto != "" {
				req.Header.Set("X-Forwarded-Proto", test.forwardedProto)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			results[id] = make(map[string]bool)
			results[id]["hasHSTS"] = w.Header().Get("Strict-Transport-Security") != ""
			results[id]["expectHSTS"] = test.expectHSTS
		}(i, tt)
	}

	wg.Wait()

	for i, result := range results {
		assert.Equal(t, result["expectHSTS"], result["hasHSTS"],
			"Test %d: HSTS header presence should match expectation", i)
	}
}

func TestMiddleware_ChainedExecution(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 미들웨어 체인 설정
	router.Use(slogMiddleware(logger))
	router.Use(slogRecovery(logger))
	router.Use(securityHeaders())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	const numRequests = 12
	var wg sync.WaitGroup
	responses := make([]struct {
		code    int
		headers map[string]string
	}, numRequests)

	wg.Add(numRequests)

	for i := range numRequests {
		go func(id int) {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			responses[id].code = w.Code
			responses[id].headers = make(map[string]string)
			responses[id].headers["X-Content-Type-Options"] = w.Header().Get("X-Content-Type-Options")
			responses[id].headers["Content-Security-Policy"] = w.Header().Get("Content-Security-Policy")
		}(i)
	}

	wg.Wait()

	// 모든 응답이 일관되게 처리되었는지 확인
	for i, resp := range responses {
		assert.Equal(t, http.StatusOK, resp.code, "Request %d should return 200", i)
		assert.Equal(t, "nosniff", resp.headers["X-Content-Type-Options"], "Request %d should have correct X-Content-Type-Options", i)
		assert.NotEmpty(t, resp.headers["Content-Security-Policy"], "Request %d should have Content-Security-Policy", i)
	}
}

func TestCsrfMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		authHeader     string
		cookie         *http.Cookie
		csrfHeader     string
		csrfProtection bool
		expectedStatus int
	}{
		{
			name:           "should allow GET without CSRF token",
			method:         "GET",
			csrfProtection: true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "should allow POST with Authorization header",
			method:         "POST",
			authHeader:     "Bearer token",
			csrfProtection: true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "should deny POST with cookie but without CSRF token",
			method:         "POST",
			cookie:         &http.Cookie{Name: "access_token", Value: "token"},
			csrfProtection: true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "should allow POST with cookie and CSRF token",
			method:         "POST",
			cookie:         &http.Cookie{Name: "access_token", Value: "token"},
			csrfHeader:     "valid-token",
			csrfProtection: true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "should allow POST when CSRF protection is disabled",
			method:         "POST",
			cookie:         &http.Cookie{Name: "access_token", Value: "token"},
			csrfProtection: false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			common.SetGlobalConfig(common.Config{
				Auth: common.AuthConfig{
					CsrfProtection: tt.csrfProtection,
				},
			})

			router := gin.New()
			router.Use(csrfMiddleware())
			router.Any("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req, _ := http.NewRequest(tt.method, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			if tt.csrfHeader != "" {
				req.Header.Set("X-CSRF-Token", tt.csrfHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
