package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/go-kit/apimanager/fiberutils"
	"github.com/lvlcn-t/loggerhead/logger"
	"golang.org/x/oauth2"
)

// Authenticator represents an OIDC authenticator.
// The provided type T is the type of the token claims that the authenticator expects.
type Authenticator[T any] interface {
	// Authenticate creates a middleware that verifies if the request is authenticated.
	// The token claims are stored in the [fiber.Ctx] locals with the key "claims" and are of the provided type T.
	Authenticate() fiber.Handler
}

// AuthConfig represents the configuration for an OIDC provider.
type AuthConfig struct {
	// Config is the OpenID Connect configuration.
	oidc.Config
	// ProviderURL is the URL of the OIDC provider.
	ProviderURL string
	// ClientSecret is the client secret used to authenticate with the OIDC provider.
	ClientSecret string
	// RedirectURL is the URL to redirect to after authentication.
	RedirectURL string
	// Scopes are the scopes to request during authentication.
	Scopes []string
	// validated is true if the configuration has been validated successfully.
	validated bool
}

// Validate validates the configuration.
func (c *AuthConfig) Validate() error {
	var err error
	if c.ProviderURL == "" {
		err = errors.New("provider URL is required")
	}

	if c.ClientID == "" && !c.SkipClientIDCheck {
		err = errors.Join(err, errors.New("you must provide a client ID, or set SkipClientIDCheck"))
	}

	if err != nil {
		return err
	}

	c.validated = true
	return nil
}

// authProvider represents an OIDC provider.
type authProvider[T any] struct {
	// verifier is used to verify ID tokens.
	verifier verifier
	// config is used to configure the OAuth2 client.
	config oauth2.Config
}

// NewDefaultAuthenticator initializes a new [Authenticator] with the expected token claims of type map[string]any.
// Returns an error if the configuration is invalid or the provider cannot be initialized.
//
// Note: This is a convenience function for when the token claims are not known.
//
// Example:
//
//	auth, err := NewDefaultAuthenticator(ctx, &AuthConfig{
//		ProviderURL:   "https://accounts.google.com",
//		ClientID:      "your-client-id",
//		ClientSecret:  "your-client-secret",
//		RedirectURL:   "http://localhost:8080/auth/callback",
//		Scopes:        []string{oidc.ScopeOpenID, "profile", "email"},
//	})
//	if err != nil {
//		// Handle error
//	}
//
//	app.Use(auth.Authenticate())
//	// The token claims are stored in the [fiber.Ctx] locals with the key "claims".
//	// [fiber.Ctx].Locals("claims") is of type map[string]any.
func NewDefaultAuthenticator(ctx context.Context, c *AuthConfig) (Authenticator[map[string]any], error) {
	return newAuthenticator[map[string]any](ctx, c)
}

// NewAuthenticator initializes a new [Authenticator] with the expected token claims of type T.
// Returns an error if the configuration is invalid or the provider cannot be initialized.
//
// Example:
//
//	type Claims struct {
//	    Issuer  string   `json:"iss"`
//	    Subject string   `json:"sub"`
//	    Roles   []string `json:"roles"`
//	}
//
//	auth, err := NewAuthenticator[Claims](ctx, &AuthConfig{
//		ProviderURL:   "https://accounts.google.com",
//		ClientID:      "your-client-id",
//		ClientSecret:  "your-client-secret",
//		RedirectURL:   "http://localhost:8080/auth/callback",
//		Scopes:        []string{oidc.ScopeOpenID, "profile", "email"},
//	})
//	if err != nil {
//		// Handle error
//	}
//
//	app.Use(auth.Authenticate())
//	// The token claims are stored in the [fiber.Ctx] locals with the key "claims".
//	// [fiber.Ctx].Locals("claims") is of type Claims.
func NewAuthenticator[T any](ctx context.Context, c *AuthConfig) (Authenticator[T], error) {
	return newAuthenticator[T](ctx, c)
}

// newAuthenticator initializes a new [authProvider] with the expected token claims of type T.
func newAuthenticator[T any](ctx context.Context, c *AuthConfig) (*authProvider[T], error) {
	if c == nil {
		return nil, errors.New("config is nil")
	}

	if !c.validated {
		if err := c.Validate(); err != nil {
			return nil, fmt.Errorf("failed to validate config: %w", err)
		}
	}

	provider, err := oidc.NewProvider(ctx, c.ProviderURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize oidc provider: %w", err)
	}

	return &authProvider[T]{
		verifier: newTokenVerifier(ctx, provider, &c.Config),
		config: oauth2.Config{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  c.RedirectURL,
			Scopes:       c.Scopes,
		},
	}, nil
}

// Authenticate creates a middleware that verifies if the request is authenticated.
// The token claims are stored in the [fiber.Ctx] locals with the key "claims" and are of the provided type T.
func (p *authProvider[T]) Authenticate() fiber.Handler {
	return func(c fiber.Ctx) error {
		log := logger.FromContext(c.Context())
		token, err := extractToken(c)
		if err != nil {
			log.DebugContext(c.RequestCtx(), "Failed to extract token", "error", err)
			return fiberutils.UnauthorizedResponse(c, err.Error())
		}

		idToken, err := p.verifier.Verify(c.RequestCtx(), token)
		if err != nil {
			log.DebugContext(c.RequestCtx(), "Failed to verify token", "error", err)
			return fiberutils.UnauthorizedResponse(c, "invalid token")
		}

		var claims T
		if err := idToken.Claims(&claims); err != nil {
			log.ErrorContext(c.RequestCtx(), "Failed to parse token claims", "error", err)
			return fiberutils.InternalServerErrorResponse(c, "failed to parse token claims")
		}

		c.Locals("claims", claims)
		return c.Next()
	}
}

// extractToken extracts the token from the Authorization header.
func extractToken(c fiber.Ctx) (string, error) {
	header := c.Get("Authorization")
	if header == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	const prefix = "Bearer "
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return header[len(prefix):], nil
}

//go:generate moq -out auth_verifier_moq.go . verifier
type verifier interface {
	// Verify parses a raw ID Token, verifies it's been signed by the provider, performs
	// any additional checks depending on the Config, and returns the payload.
	//
	// Verify does NOT do nonce validation, which is the callers responsibility.
	//
	// See: https://openid.net/specs/openid-connect-core-1_0.html#IDTokenValidation
	//
	//	oauth2Token, err := oauth2Config.Exchange(ctx, r.URL.Query().Get("code"))
	//	if err != nil {
	//	    // handle error
	//	}
	//
	//	// Extract the ID Token from oauth2 token.
	//	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	//	if !ok {
	//	    // handle error
	//	}
	//
	//	token, err := verifier.Verify(ctx, rawIDToken)
	Verify(ctx context.Context, token string) (tokenUnmarshaler, error)
}

var _ verifier = (*tokenVerifier)(nil)

type tokenVerifier struct{ *oidc.IDTokenVerifier }

func newTokenVerifier(ctx context.Context, provider *oidc.Provider, config *oidc.Config) *tokenVerifier {
	return &tokenVerifier{
		IDTokenVerifier: provider.VerifierContext(ctx, config),
	}
}

func (v *tokenVerifier) Verify(ctx context.Context, token string) (tokenUnmarshaler, error) {
	return v.IDTokenVerifier.Verify(ctx, token)
}

//go:generate moq -out auth_unmarshaler_moq.go . tokenUnmarshaler
type tokenUnmarshaler interface {
	// Claims unmarshals the raw JSON payload of the ID Token into a provided struct.
	//
	//	idToken, err := idTokenVerifier.Verify(rawIDToken)
	//	if err != nil {
	//		// handle error
	//	}
	//	var claims struct {
	//		Email         string `json:"email"`
	//		EmailVerified bool   `json:"email_verified"`
	//	}
	//	if err := idToken.Claims(&claims); err != nil {
	//		// handle error
	//	}
	Claims(claims any) error
}
