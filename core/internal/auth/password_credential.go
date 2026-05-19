package auth

import (
	"context"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	pkgerrors "github.com/pkg/errors"
)

// ResourceOwnerPasswordCredentialsGrantHandlerLocal is a local, non-deprecated
// copy of the upstream handler sufficient for our needs. It implements
// fosite.TokenEndpointHandler.
type ResourceOwnerPasswordCredentialsGrantHandlerLocal struct {
	*oauth2.HandleHelper
	// Storage is used to persist session data across requests.
	Storage              oauth2.ResourceOwnerPasswordCredentialsGrantStorage
	RefreshTokenStrategy oauth2.RefreshTokenStrategy
	Config               interface {
		fosite.ScopeStrategyProvider
		fosite.AudienceStrategyProvider
		fosite.RefreshTokenScopesProvider
		fosite.RefreshTokenLifespanProvider
		fosite.AccessTokenLifespanProvider
	}
}

// Ensure interface implementation
var _ fosite.TokenEndpointHandler = (*ResourceOwnerPasswordCredentialsGrantHandlerLocal)(nil)

func (c *ResourceOwnerPasswordCredentialsGrantHandlerLocal) HandleTokenEndpointRequest(ctx context.Context, request fosite.AccessRequester) error {
	if !c.CanHandleTokenEndpointRequest(ctx, request) {
		return fosite.ErrUnknownRequest
	}
	if !request.GetClient().GetGrantTypes().Has("password") {
		return fosite.ErrUnauthorizedClient.WithHint("The client is not allowed to use authorization grant 'password'.")
	}
	client := request.GetClient()
	for _, scope := range request.GetRequestedScopes() {
		if !c.Config.GetScopeStrategy(ctx)(client.GetScopes(), scope) {
			return fosite.ErrInvalidScope.WithHintf("The OAuth 2.0 Client is not allowed to request scope '%s'.", scope)
		}
	}
	if err := c.Config.GetAudienceStrategy(ctx)(client.GetAudience(), request.GetRequestedAudience()); err != nil {
		return err
	}
	username := request.GetRequestForm().Get("username")
	password := request.GetRequestForm().Get("password")
	if username == "" || password == "" {
		return fosite.ErrInvalidRequest.WithHint("Username or password are missing from the POST body.")
	} else if sub, err := c.Storage.Authenticate(ctx, username, password); pkgerrors.Is(err, fosite.ErrNotFound) {
		return fosite.ErrInvalidGrant.WithHint("Unable to authenticate the provided username and password credentials.").WithWrap(err).WithDebug(err.Error())
	} else if err != nil {
		return fosite.ErrServerError.WithWrap(err).WithDebug(err.Error())
	} else {
		if sess, ok := request.GetSession().(interface{ SetSubject(string) }); ok {
			sess.SetSubject(sub)
		}
	}
	delete(request.GetRequestForm(), "password")
	atLifespan := fosite.GetEffectiveLifespan(request.GetClient(), fosite.GrantTypePassword, fosite.AccessToken, c.Config.GetAccessTokenLifespan(ctx))
	request.GetSession().SetExpiresAt(fosite.AccessToken, time.Now().UTC().Add(atLifespan).Round(time.Second))

	rtLifespan := fosite.GetEffectiveLifespan(request.GetClient(), fosite.GrantTypePassword, fosite.RefreshToken, c.Config.GetRefreshTokenLifespan(ctx))
	if rtLifespan > -1 {
		request.GetSession().SetExpiresAt(fosite.RefreshToken, time.Now().UTC().Add(rtLifespan).Round(time.Second))
	}
	return nil
}

func (c *ResourceOwnerPasswordCredentialsGrantHandlerLocal) PopulateTokenEndpointResponse(ctx context.Context, requester fosite.AccessRequester, responder fosite.AccessResponder) error {
	if !c.CanHandleTokenEndpointRequest(ctx, requester) {
		return fosite.ErrUnknownRequest
	}
	atLifespan := fosite.GetEffectiveLifespan(requester.GetClient(), fosite.GrantTypePassword, fosite.AccessToken, c.Config.GetAccessTokenLifespan(ctx))
	accessTokenSignature, err := c.IssueAccessToken(ctx, atLifespan, requester, responder)
	if err != nil {
		return err
	}
	var refresh, refreshSignature string
	if len(c.Config.GetRefreshTokenScopes(ctx)) == 0 || requester.GetGrantedScopes().HasOneOf(c.Config.GetRefreshTokenScopes(ctx)...) {
		var err error
		refresh, refreshSignature, err = c.RefreshTokenStrategy.GenerateRefreshToken(ctx, requester)
		if err != nil {
			return fosite.ErrServerError.WithWrap(err).WithDebug(err.Error())
		} else if err := c.Storage.CreateRefreshTokenSession(ctx, refreshSignature, accessTokenSignature, requester.Sanitize([]string{})); err != nil {
			return fosite.ErrServerError.WithWrap(err).WithDebug(err.Error())
		}
	}
	if refresh != "" {
		responder.SetExtra("refresh_token", refresh)
	}
	return nil
}

func (c *ResourceOwnerPasswordCredentialsGrantHandlerLocal) CanSkipClientAuth(_ context.Context, _ fosite.AccessRequester) bool {
	return false
}

func (c *ResourceOwnerPasswordCredentialsGrantHandlerLocal) CanHandleTokenEndpointRequest(_ context.Context, requester fosite.AccessRequester) bool {
	return requester.GetGrantTypes().ExactOne("password")
}

// OAuth2ResourceOwnerPasswordCredentialsFactoryLocal Factory that wires dependencies from fosite compose into our local handler
func OAuth2ResourceOwnerPasswordCredentialsFactoryLocal(config fosite.Configurator, storage any, strategy any) any {
	return &ResourceOwnerPasswordCredentialsGrantHandlerLocal{
		HandleHelper: &oauth2.HandleHelper{
			AccessTokenStrategy: strategy.(oauth2.AccessTokenStrategy),
			AccessTokenStorage:  storage.(oauth2.AccessTokenStorage),
			Config:              config,
		},
		Storage:              storage.(oauth2.ResourceOwnerPasswordCredentialsGrantStorage),
		RefreshTokenStrategy: strategy.(oauth2.RefreshTokenStrategy),
		Config:               config,
	}
}
