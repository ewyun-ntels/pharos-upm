package internal

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Helper to reset the global token cache between tests
func resetClientTokens() {
	mutex.Lock()
	defer mutex.Unlock()
	clientTokens = NewMap[*Token]()
}

func Test_getKey_Format(t *testing.T) {
	resetClientTokens()
	got := getKey("http://x", "id", "sec")
	require.Equal(t, "http://x:id:sec", got)
}

func Test_getToken_Success_SendsProperRequest_ParsesJSON(t *testing.T) {
	resetClientTokens()

	var gotPath string
	var gotAuth string
	var gotContentType string
	var gotBody string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"abc","expires_in":2,"refresh_token":"r","scope":"s","token_type":"Bearer"}`))
	}))
	defer srv.Close()

	tok, err := getToken(srv.URL, "myid", "mysec")
	require.NoError(t, err)
	require.NotNil(t, tok)
	require.Equal(t, "/auth/token", gotPath)
	require.Equal(t, "Bearer your_token", gotAuth)
	// resty may set charset; ensure it contains the expected media type
	require.Contains(t, strings.ToLower(gotContentType), "application/x-www-form-urlencoded")
	require.Contains(t, gotBody, "grant_type=client_credentials")
	require.Contains(t, gotBody, "client_id=myid")
	require.Contains(t, gotBody, "client_secret=mysec")

	require.Equal(t, "abc", tok.AccessToken)
	require.Equal(t, 2, tok.ExpiresIn)
	require.Equal(t, "r", tok.RefreshToken)
	require.Equal(t, "s", tok.Scope)
	require.Equal(t, "Bearer", tok.TokenType)
}

func Test_getToken_Non200_ReturnsError(t *testing.T) {
	resetClientTokens()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("nope"))
	}))
	defer srv.Close()

	tok, err := getToken(srv.URL, "id", "sec")
	require.Nil(t, tok)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Unauthorized")
	require.Contains(t, err.Error(), "nope")
}

func Test_getToken_InvalidJSON_ReturnsErrorAndTokenPtr(t *testing.T) {
	resetClientTokens()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{invalid-json}"))
	}))
	defer srv.Close()

	tok, err := getToken(srv.URL, "id", "sec")
	require.NotNil(t, tok) // function returns &Token{} alongside error
	require.Error(t, err)
}

func Test_GetClientToken_CacheAndConcurrentSingleFetch(t *testing.T) {
	resetClientTokens()

	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		// small delay to amplify race if any
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"cached","expires_in":30}`)) // Longer expiry to avoid race
	}))
	defer srv.Close()

	const N = 10
	var wg sync.WaitGroup
	wg.Add(N)

	results := make([]*Token, N)
	errs := make([]error, N)

	for i := range N {
		go func() {
			defer wg.Done()
			tok, err := GetClientToken(srv.URL, "id", "sec")
			results[i] = tok
			errs[i] = err
		}()
	}

	wg.Wait()

	// All should succeed and refer to the same token value
	for i := range N {
		require.NoError(t, errs[i])
		require.NotNil(t, results[i])
		require.Equal(t, "cached", results[i].AccessToken)
	}

	// Should have hit the server only once thanks to cache + locking
	require.Equal(t, int32(1), atomic.LoadInt32(&hits))

	// Subsequent call should be served from cache without additional hit
	tok2, err := GetClientToken(srv.URL, "id", "sec")
	require.NoError(t, err)
	require.Equal(t, "cached", tok2.AccessToken)
	require.Equal(t, int32(1), atomic.LoadInt32(&hits))
}

func Test_setToken_ExpiresRemovesEntry(t *testing.T) {
	// Ensure test isolation - wait a bit for any existing goroutines to finish
	time.Sleep(50 * time.Millisecond)
	resetClientTokens()

	key := getKey("addr", "id", "sec")
	tkn := &Token{AccessToken: "x", ExpiresIn: 2}

	setToken(key, tkn)

	// Immediately present
	mutex.Lock()
	exists := clientTokens.Exist(key)
	mutex.Unlock()
	require.True(t, exists)

	// After ~0.9*ExpiresIn seconds, it should be removed.
	// With ExpiresIn=2, 0.9*2 -> 1 (after truncation), so wait a bit over 1s
	time.Sleep(1200 * time.Millisecond)

	mutex.Lock()
	exists = clientTokens.Exist(key)
	mutex.Unlock()
	require.False(t, exists)
}
