package auth

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/ory/fosite"
	fositeOAuth2 "github.com/ory/fosite/handler/oauth2"
	"ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/internal/user"
)

type Storage struct {
	allowedClients []string
	userStore      user.Store
	Clients        *internal.Map[fosite.Client]

	// access token storage
	accessTokens          map[string]*AccessToken
	accessTokenRequesters map[string]*AccessToken
	accMu                 sync.RWMutex

	// refresh token storage
	refreshTokens               map[string]*RefreshToken
	refreshTokenRequesters      map[string]*RefreshToken
	refMu                       sync.RWMutex
	refreshTokenRevokeCallback  func(ref *RefreshToken)
	refreshTokenTimeoutCallback func(ref *RefreshToken)
	refreshTokenRemovedCallback func(ref *RefreshToken)

	// cron
	cronMu   sync.Mutex
	cronCtrl *cron.Cron

	// metrics counters
	sessionExpiredTotal atomic.Int64
	sessionRevokedTotal atomic.Int64
}

func newStorage(userStore user.Store, cfg StorageConfig) (*Storage, error) {
	storage := &Storage{
		allowedClients: cfg.AllowedClients,
		userStore:      userStore,
	}

	storage.accessTokens = make(map[string]*AccessToken)
	storage.accessTokenRequesters = make(map[string]*AccessToken)

	storage.refreshTokens = make(map[string]*RefreshToken)
	storage.refreshTokenRequesters = make(map[string]*RefreshToken)

	storage.Clients = internal.NewMap[fosite.Client]()
	for _, client := range cfg.Clients {
		storage.Clients.Set(client.ID, &fosite.DefaultClient{
			ID:         client.ID,
			Secret:     []byte(client.Secret),
			GrantTypes: []string{string(fosite.GrantTypeClientCredentials)},
		})
	}

	storage.refreshTokenRevokeCallback = cfg.RefreshTokenRevokeCallback
	storage.refreshTokenTimeoutCallback = cfg.RefreshTokenTimeoutCallback
	storage.refreshTokenRemovedCallback = cfg.RefreshTokenRemovedCallback

	err := storage.cronStart()
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Storage) getDefaultClient(id string) *fosite.DefaultClient {
	return &fosite.DefaultClient{
		ID:         id,
		Public:     true,
		GrantTypes: []string{"password", "refresh_token"},
		Scopes:     []string{""},
	}
}

func (s *Storage) GetClient(_ context.Context, id string) (fosite.Client, error) {
	slog.Info("Fetching client", slog.String("client_id", id))
	if slices.Contains(s.allowedClients, id) {
		return s.getDefaultClient(id), nil
	}

	if s.Clients != nil && s.Clients.Exist(id) {
		return s.Clients.Get(id), nil
	}

	slog.Error("Client does not exist", slog.String("client_id", id))
	return nil, fosite.ErrNotFound
}

func (s *Storage) ClientAssertionJWTValid(_ context.Context, _ string) error {
	slog.Info("ClientAssertionJWTValid function is not implemented")
	return ErrNotImplemented
}

func (s *Storage) SetClientAssertionJWT(context.Context, string, time.Time) error {
	slog.Info("SetClientAssertionJWT function is not implemented")
	return ErrNotImplemented
}

func (s *Storage) CreateAuthorizeCodeSession(context.Context, string, fosite.Requester) (err error) {
	slog.Info("Authorize code is not implemented")
	return ErrNotImplemented
}
func (s *Storage) GetAuthorizeCodeSession(context.Context, string, fosite.Session) (request fosite.Requester, err error) {
	slog.Info("Authorize code is not implemented")
	return nil, ErrNotImplemented
}

func (s *Storage) InvalidateAuthorizeCodeSession(context.Context, string) (err error) {
	slog.Info("Authorize code is not implemented")
	return ErrNotImplemented
}

func (s *Storage) Authenticate(_ context.Context, name string, secret string) (subject string, err error) {
	subject = name
	err = s.userStore.Authenticate(name, secret)
	if err != nil {
		slog.Error("Authentication failed", slog.String("username", name), slog.String("error", err.Error()))
		return
	}
	slog.Info("Authentication successful", slog.String("username", name))

	return
}

func (s *Storage) CreateAccessTokenSession(_ context.Context, signature string, request fosite.Requester) (err error) {
	s.accMu.Lock()
	defer s.accMu.Unlock()

	accToken := AccessToken{
		token:     signature,
		Requester: request,
	}

	existsToken, ok := s.accessTokenRequesters[request.GetID()]
	if ok {
		slog.Debug("delete previous token", "token", existsToken.token, "request_id", existsToken.Requester.GetID())
		delete(s.accessTokenRequesters, existsToken.Requester.GetID())
		delete(s.accessTokens, existsToken.token)
	}

	slog.Debug("create token", "token", accToken.token, "request_id", request.GetID())
	s.accessTokens[signature] = &accToken
	s.accessTokenRequesters[request.GetID()] = &accToken
	return nil
}

func (s *Storage) GetAccessTokenSession(_ context.Context, signature string, _ fosite.Session) (request fosite.Requester, err error) {
	s.accMu.Lock()
	defer s.accMu.Unlock()

	if token, ok := s.accessTokens[signature]; ok {
		expiredAt := (*token).GetSession().GetExpiresAt(fosite.AccessToken).UTC()
		if expiredAt.Before(time.Now().UTC()) {
			return *token, fosite.ErrTokenExpired
		}

		return *token, nil
	}

	return nil, fosite.ErrTokenExpired
}

func (s *Storage) DeleteAccessTokenSession(_ context.Context, signature string) (err error) {
	s.accMu.Lock()
	defer s.accMu.Unlock()

	token, ok := s.accessTokens[signature]
	if !ok {
		return nil
	}

	slog.Debug("delete token", "token", token.token, "request_id", token.GetID())
	delete(s.accessTokenRequesters, token.Requester.GetID())
	delete(s.accessTokens, signature)

	return nil
}

func (s *Storage) RevokeAccessToken(_ context.Context, requestID string) error {
	s.accMu.Lock()
	defer s.accMu.Unlock()

	token, ok := s.accessTokenRequesters[requestID]
	if !ok {
		return nil
	}
	delete(s.accessTokenRequesters, token.Requester.GetID())
	delete(s.accessTokens, requestID)

	return nil
}

func (s *Storage) CreateRefreshTokenSession(_ context.Context, signature, _ string, req fosite.Requester) error {
	s.refMu.Lock()
	defer s.refMu.Unlock()

	// 정리
	for sig, ref := range s.refreshTokens {
		if !ref.isValid() {
			delete(s.refreshTokens, sig)
		}
	}

	refToken := RefreshToken{
		token:     signature,
		Requester: req,
	}

	s.refreshTokens[signature] = &refToken
	s.refreshTokenRequesters[req.GetID()] = &refToken

	return nil
}

func (s *Storage) GetRefreshTokenSession(_ context.Context, signature string, session fosite.Session) (request fosite.Requester, err error) {
	s.refMu.RLock()
	defer s.refMu.RUnlock()

	if ref, ok := s.refreshTokens[signature]; ok {
		if ref.isValid() {
			return ref, nil
		}
		// 만료된 토큰이더라도 사용자 식별을 위해 subject와 session claims를 전달
		if session != nil && ref.GetSession() != nil {
			if subject := ref.GetSession().GetSubject(); subject != "" {
				if jwtSess, ok := session.(*fositeOAuth2.JWTSession); ok && jwtSess.JWTClaims != nil {
					jwtSess.JWTClaims.Subject = subject
					// Extra claims(session_id 등) 복사 - handler에서 세션 이력 업데이트에 사용
					if refJwtSess, ok2 := ref.GetSession().(*fositeOAuth2.JWTSession); ok2 &&
						refJwtSess.JWTClaims != nil && refJwtSess.JWTClaims.Extra != nil {
						if jwtSess.JWTClaims.Extra == nil {
							jwtSess.JWTClaims.Extra = make(map[string]any, len(refJwtSess.JWTClaims.Extra))
						}
						maps.Copy(jwtSess.JWTClaims.Extra, refJwtSess.JWTClaims.Extra)
					}
				}
			}
		}
		return ref, fosite.ErrTokenExpired
	}

	return nil, fosite.ErrTokenExpired
}

func (s *Storage) DeleteRefreshTokenSession(_ context.Context, signature string) (err error) {
	s.refMu.Lock()
	defer s.refMu.Unlock()

	req, ok := s.refreshTokens[signature]
	if !ok {
		return nil
	}

	delete(s.refreshTokens, signature)
	delete(s.refreshTokenRequesters, req.GetID())

	return nil
}

func (s *Storage) RotateRefreshToken(ctx context.Context, _ string, refreshTokenSignature string) (err error) {
	return s.DeleteRefreshTokenSession(ctx, refreshTokenSignature)
}

func (s *Storage) RevokeRefreshToken(_ context.Context, requestID string) error {
	s.refMu.Lock()
	defer s.refMu.Unlock()

	ref, ok := s.refreshTokenRequesters[requestID]
	if !ok {
		return nil
	}

	if s.refreshTokenRevokeCallback != nil {
		s.refreshTokenRevokeCallback(ref)
	}

	delete(s.refreshTokens, ref.token)
	delete(s.refreshTokenRequesters, requestID)
	s.sessionRevokedTotal.Add(1)

	return nil
}

func (s *Storage) RemoveAccessTokenBySubject(subject string) (err error) {
	s.accMu.Lock()
	defer s.accMu.Unlock()

	for sig, token := range s.accessTokens {
		tokenSub := (*token).GetSession().GetSubject()
		if tokenSub == subject {
			delete(s.accessTokenRequesters, token.Requester.GetID())
			delete(s.accessTokens, sig)
		}
	}

	return
}

func (s *Storage) accessTokenGc() {
	s.accMu.Lock()
	defer s.accMu.Unlock()

	for sig, token := range s.accessTokens {
		expiredAt := (*token).GetSession().GetExpiresAt(fosite.AccessToken).UTC()
		if expiredAt.Before(time.Now().UTC()) {
			delete(s.accessTokenRequesters, token.Requester.GetID())
			delete(s.accessTokens, sig)
		}
	}

	// 버그등으로 인하여 데이터가 삭제되는 것을 방지하기 위해 한번더 돌림(장애 대응 목적)
	// 일반적으로 많은 유저가 사용하지 않을것을 가정하기 때문에 loop를 2번 돌려도 문제가 없을것으로 보임
	for sig, ref := range s.accessTokenRequesters {
		expiredAt := (*ref).GetSession().GetExpiresAt(fosite.AccessToken).UTC()
		if expiredAt.Before(time.Now().UTC()) {
			slog.Warn("access token GC triggered - token marked as invalid",
				"request_id", ref.GetID(),
				"token", ref.token)
			delete(s.accessTokenRequesters, sig)
		}
	}
}

func (s *Storage) RemoveRefreshTokenBySubject(subject string) (err error) {
	s.refMu.Lock()
	defer s.refMu.Unlock()

	for sig, ref := range s.refreshTokens {
		tokenSub := (*ref).GetSession().GetSubject()
		if tokenSub == subject {
			if s.refreshTokenRemovedCallback != nil {
				s.refreshTokenRemovedCallback(ref)
			}
			delete(s.refreshTokens, sig)
			delete(s.refreshTokenRequesters, ref.GetID())
			s.sessionRevokedTotal.Add(1)
		}
	}

	return
}

func (s *Storage) GetRefreshTokenList() (list []RefreshToken) {
	for _, ref := range s.refreshTokens {
		if ref != nil {
			list = append(list, *ref)
		}
	}

	return
}

func (s *Storage) GetActiveRefreshTokenCount() int {
	s.refMu.RLock()
	defer s.refMu.RUnlock()
	return len(s.refreshTokens)
}

func (s *Storage) GetSessionExpiredTotal() int64 {
	return s.sessionExpiredTotal.Load()
}

func (s *Storage) GetSessionRevokedTotal() int64 {
	return s.sessionRevokedTotal.Load()
}

func (s *Storage) refreshTokenGc() {
	s.refMu.Lock()
	defer s.refMu.Unlock()

	for sig, ref := range s.refreshTokens {
		if !ref.isValid() {
			if ref.GetSession() != nil {
				if s.refreshTokenTimeoutCallback != nil {
					s.refreshTokenTimeoutCallback(ref)
				}
			}
			delete(s.refreshTokens, sig)
			delete(s.refreshTokenRequesters, ref.GetID())
			s.sessionExpiredTotal.Add(1)
			continue
		}

		expiredAt := ref.GetSession().GetExpiresAt(fosite.RefreshToken).UTC()

		if expiredAt.Before(time.Now().UTC()) {
			if s.refreshTokenTimeoutCallback != nil {
				s.refreshTokenTimeoutCallback(ref)
			}

			delete(s.refreshTokens, sig)
			delete(s.refreshTokenRequesters, ref.GetID())
		}
	}

	// 버그등으로 인하여 데이터가 삭제되는 것을 방지하기 위해 한번더 돌림(장애 대응 목적)
	// 일반적으로 많은 유저가 사용하지 않을것을 가정하기 때문에 loop를 2번 돌려도 문제가 없을것으로 보임
	for sig, ref := range s.refreshTokenRequesters {
		if !ref.isValid() {
			delete(s.refreshTokenRequesters, sig)
			continue
		}

		expiredAt := ref.GetSession().GetExpiresAt(fosite.RefreshToken).UTC()
		if expiredAt.Before(time.Now().UTC()) {
			slog.Warn("refresh token GC triggered - token marked as invalid",
				"request_id", ref.GetID(),
				"token", ref.token)

			delete(s.refreshTokenRequesters, sig)
		}
	}
}

func (s *Storage) cronStart() error {
	s.cronCtrl = cron.New(cron.WithSeconds())

	_, err := s.cronCtrl.AddJob("@every 10s", cron.FuncJob(func() {
		s.accessTokenGc()
	}))
	if err != nil {
		return err
	}

	_, err = s.cronCtrl.AddJob("@every 10s", cron.FuncJob(func() {
		s.refreshTokenGc()
	}))
	if err != nil {
		return err
	}

	go s.cronCtrl.Start()

	return nil
}

func (s *Storage) cronStop() error {
	s.cronMu.Lock()
	defer s.cronMu.Unlock()

	if s.cronCtrl != nil {
		s.cronCtrl.Stop()
		s.cronCtrl = nil
	}
	return nil
}

func (s *Storage) Close() error {
	_ = s.cronStop()
	return nil
}

func (s *Storage) Run() error {
	s.cronMu.Lock()
	defer s.cronMu.Unlock()

	if s.cronCtrl == nil {
		s.cronCtrl = cron.New(cron.WithSeconds())
	}

	go s.cronCtrl.Start()

	return nil
}
