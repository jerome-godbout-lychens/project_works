package auth

import (
	"context"
	"errors"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCProvider handles OIDC authentication flow
type OIDCProvider struct {
	provider   *oidc.Provider
	oauth2Config *oauth2.Config
	verifier   *oidc.IDTokenVerifier
}

// NewOIDCProvider creates a new OIDCProvider for the given issuer URL
func NewOIDCProvider(ctx context.Context, issuerURL string, clientId string, clientSecret string, redirectURL string) (*OIDCProvider, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, err
	}

	oauth2Config := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: clientId,
	})

	return &OIDCProvider{
		provider:     provider,
		oauth2Config: oauth2Config,
		verifier:     verifier,
	}, nil
}

// GetAuthURL returns the OAuth2 authorization URL
func (oidcProvider *OIDCProvider) GetAuthURL(state string) string {
	return oidcProvider.oauth2Config.AuthCodeURL(state)
}

// HandleCallback exchanges the authorization code for an ID token and extracts claims
// Returns provider (issuer URL), subject, email, displayName, and error
func (oidcProvider *OIDCProvider) HandleCallback(ctx context.Context, code string) (string, string, string, string, error) {
	oauth2Token, err := oidcProvider.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return "", "", "", "", err
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return "", "", "", "", errors.New("missing id_token in response")
	}

	idToken, err := oidcProvider.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "", "", "", "", err
	}

	var claims struct {
		Subject     string `json:"sub"`
		Email       string `json:"email"`
		DisplayName string `json:"name"`
	}

	err = idToken.Claims(&claims)
	if err != nil {
		return "", "", "", "", err
	}

	provider := idToken.Issuer

	return provider, claims.Subject, claims.Email, claims.DisplayName, nil
}
