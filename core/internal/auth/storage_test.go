package auth

import (
	"context"
	"testing"
	"time"

	"github.com/ory/fosite"
	fositeOAuth2 "github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/token/jwt"
)

// newTestStorage creates a Storage without starting the cron scheduler.
func newTestStorage() *Storage {
	return &Storage{
		refreshTokens:          make(map[string]*RefreshToken),
		refreshTokenRequesters: make(map[string]*RefreshToken),
	}
}

// newTestRequester creates a fosite.Requester with the given id and session.
func newTestRequester(id string, session fosite.Session) fosite.Requester {
	return &fosite.Request{
		ID:      id,
		Session: session,
	}
}

// newTestSession creates a fosite session with the given subject and refresh token expiry.
func newTestSession(subject string, expiresAt time.Time) fosite.Session {
	s := &fosite.DefaultSession{Subject: subject}
	s.SetExpiresAt(fosite.RefreshToken, expiresAt)
	return s
}

// newJWTSession creates an empty JWTSession to pass as the session parameter to GetRefreshTokenSession.
func newJWTSession() *fositeOAuth2.JWTSession {
	return &fositeOAuth2.JWTSession{
		JWTClaims: &jwt.JWTClaims{},
	}
}

// TestGetRefreshTokenSession_ValidToken: 유효한 토큰 조회 시 에러 없이 반환
func TestGetRefreshTokenSession_ValidToken(t *testing.T) {
	s := newTestStorage()
	sig := "sig-valid"
	subject := "alice"

	req := newTestRequester("req-valid", newTestSession(subject, time.Now().Add(1*time.Hour)))
	s.refreshTokens[sig] = &RefreshToken{token: sig, Requester: req}

	session := newJWTSession()
	result, err := s.GetRefreshTokenSession(context.Background(), sig, session)

	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil requester")
	}
}

// TestGetRefreshTokenSession_ExpiredInMap: map에 있지만 만료된 토큰 → ErrTokenExpired + subject 복구
func TestGetRefreshTokenSession_ExpiredInMap(t *testing.T) {
	s := newTestStorage()
	sig := "sig-expired"
	subject := "bob"

	req := newTestRequester("req-expired", newTestSession(subject, time.Now().Add(-1*time.Hour)))
	s.refreshTokens[sig] = &RefreshToken{token: sig, Requester: req}

	session := newJWTSession()
	_, err := s.GetRefreshTokenSession(context.Background(), sig, session)

	if err != fosite.ErrTokenExpired {
		t.Fatalf("expected ErrTokenExpired, got: %v", err)
	}
	if session.JWTClaims.Subject != subject {
		t.Errorf("expected subject %q, got %q", subject, session.JWTClaims.Subject)
	}
}

// TestGetRefreshTokenSession_UnknownToken: 완전히 모르는 토큰 → ErrTokenExpired + subject 빈 값
func TestGetRefreshTokenSession_UnknownToken(t *testing.T) {
	s := newTestStorage()

	session := newJWTSession()
	result, err := s.GetRefreshTokenSession(context.Background(), "unknown-sig", session)

	if err != fosite.ErrTokenExpired {
		t.Fatalf("expected ErrTokenExpired, got: %v", err)
	}
	if result != nil {
		t.Error("expected nil requester for unknown token")
	}
	if session.JWTClaims.Subject != "" {
		t.Errorf("expected empty subject, got %q", session.JWTClaims.Subject)
	}
}

// TestRefreshTokenGc_ExpiredTokenRemoved: GC 실행 시 만료된 토큰이 두 map 모두에서 삭제됨
func TestRefreshTokenGc_ExpiredTokenRemoved(t *testing.T) {
	s := newTestStorage()
	sig := "sig-gc-expired"

	req := newTestRequester("req-gc-expired", newTestSession("dave", time.Now().Add(-1*time.Hour)))
	s.refreshTokens[sig] = &RefreshToken{token: sig, Requester: req}
	s.refreshTokenRequesters["req-gc-expired"] = s.refreshTokens[sig]

	s.refreshTokenGc()

	if _, exists := s.refreshTokens[sig]; exists {
		t.Error("expected token to be removed from refreshTokens after GC")
	}
	if _, exists := s.refreshTokenRequesters["req-gc-expired"]; exists {
		t.Error("expected token to be removed from refreshTokenRequesters after GC")
	}
}

// TestRefreshTokenGc_ValidTokenNotDeleted: GC 실행 시 유효한 토큰은 삭제되지 않아야 함
func TestRefreshTokenGc_ValidTokenNotDeleted(t *testing.T) {
	s := newTestStorage()
	sig := "sig-gc-valid"
	subject := "eve"

	req := newTestRequester("req-gc-valid", newTestSession(subject, time.Now().Add(1*time.Hour)))
	s.refreshTokens[sig] = &RefreshToken{token: sig, Requester: req}

	s.refreshTokenGc()

	if _, exists := s.refreshTokens[sig]; !exists {
		t.Error("expected valid token to remain in refreshTokens after GC")
	}
}

// TestRefreshTokenGc_ExpiredTokenCallsTimeoutCallback: GC 시 만료 토큰에서 timeout 콜백 호출
func TestRefreshTokenGc_ExpiredTokenCallsTimeoutCallback(t *testing.T) {
	s := newTestStorage()
	sig := "sig-gc-cb"
	subject := "frank"

	var callbackSubject string
	s.refreshTokenTimeoutCallback = func(ref *RefreshToken) {
		callbackSubject = ref.GetSession().GetSubject()
	}

	req := newTestRequester("req-gc-cb", newTestSession(subject, time.Now().Add(-1*time.Hour)))
	s.refreshTokens[sig] = &RefreshToken{token: sig, Requester: req}
	s.refreshTokenRequesters["req-gc-cb"] = s.refreshTokens[sig]

	s.refreshTokenGc()

	if callbackSubject != subject {
		t.Errorf("expected timeout callback with subject %q, got %q", subject, callbackSubject)
	}
}

// TestRemoveRefreshTokenBySubject_CallsRemovedCallback: subject로 직접 삭제 시 removed 콜백 호출
func TestRemoveRefreshTokenBySubject_CallsRemovedCallback(t *testing.T) {
	s := newTestStorage()
	sig := "sig-rm-cb"
	subject := "grace"

	var callbackSubject string
	s.refreshTokenRemovedCallback = func(ref *RefreshToken) {
		callbackSubject = ref.GetSession().GetSubject()
	}

	req := newTestRequester("req-rm-cb", newTestSession(subject, time.Now().Add(1*time.Hour)))
	s.refreshTokens[sig] = &RefreshToken{token: sig, Requester: req}
	s.refreshTokenRequesters["req-rm-cb"] = s.refreshTokens[sig]

	if err := s.RemoveRefreshTokenBySubject(subject); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callbackSubject != subject {
		t.Errorf("expected removed callback with subject %q, got %q", subject, callbackSubject)
	}
	if _, exists := s.refreshTokens[sig]; exists {
		t.Error("expected token to be removed from refreshTokens")
	}
}
