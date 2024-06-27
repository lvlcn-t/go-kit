package middleware

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/go-kit/apimanager/fiberutils"
	"github.com/lvlcn-t/loggerhead/logger"
	"golang.org/x/oauth2"
)

// AuthProvider represents an OIDC provider.
type AuthProvider struct {
	// verifier is used to verify ID tokens.
	verifier verifier
	// config is used to configure the OAuth2 client.
	config oauth2.Config
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

	if c.ClientID == "" {
		err = errors.Join(err, errors.New("client ID is required"))
	}

	if err != nil {
		return err
	}

	c.validated = true
	return nil
}

// NewAuthProvider initializes a new OIDC provider.
// Returns an error if the configuration is invalid or the provider cannot be initialized.
func NewAuthProvider(ctx context.Context, c *AuthConfig) (*AuthProvider, error) {
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

	return &AuthProvider{
		verifier: provider.VerifierContext(ctx, &c.Config),
		config: oauth2.Config{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  c.RedirectURL,
			Scopes:       c.Scopes,
		},
	}, nil
}

// Authenticate creates a middleware to check if the request is authenticated.
func Authenticate(provider *AuthProvider) fiber.Handler {
	return AuthenticateWithClaims[map[string]any](provider)
}

// Authorize creates a middleware to check if the request is authorized based on roles.
func Authorize(roles ...string) fiber.Handler {
	return AuthorizeWithClaims[map[string]any]("roles", roles...)
}

// AuthenticateWithClaims creates a middleware to check if the request is authenticated and extracts claims.
func AuthenticateWithClaims[T any](provider *AuthProvider) fiber.Handler {
	if provider == nil {
		panic("provider is nil")
	}

	return func(c fiber.Ctx) error {
		log := logger.FromContext(c.UserContext())
		token, err := extractToken(c)
		if err != nil {
			log.DebugContext(c.Context(), "Failed to extract token", "error", err)
			return fiberutils.UnauthorizedResponse(c, err.Error())
		}

		var idToken tokenUnmarshaler
		idToken, err = provider.verifier.Verify(c.Context(), token)
		if err != nil {
			log.DebugContext(c.Context(), "Failed to verify token", "error", err)
			return fiberutils.UnauthorizedResponse(c, "invalid token")
		}

		var claims T
		if err := idToken.Claims(&claims); err != nil {
			log.ErrorContext(c.Context(), "Failed to parse token claims", "error", err)
			return fiberutils.InternalServerErrorResponse(c, "failed to parse token claims")
		}

		c.Locals("claims", claims)
		return c.Next()
	}
}

// AuthorizeWithClaims creates a middleware to check if the request is authorized based on roles.
func AuthorizeWithClaims[T any](key string, roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if len(roles) == 0 {
			return c.Next()
		}

		log := logger.FromContext(c.UserContext())
		claims, ok := c.Locals("claims").(T)
		if !ok {
			log.DebugContext(c.Context(), "No claims found")
			return fiberutils.ForbiddenResponse(c, "no claims found")
		}

		claimRoles, err := getRolesFromClaims(claims, key)
		if err != nil {
			log.ErrorContext(c.Context(), "Failed to get roles from claims", "error", err)
			return fiberutils.InternalServerErrorResponse(c, "failed to get roles from claims")
		}

		roleMap := make(map[string]bool)
		for _, role := range claimRoles {
			roleMap[role] = true
		}

		for _, role := range roles {
			if !roleMap[role] {
				log.DebugContext(c.Context(), "Insufficient permissions", "roles", roles)
				return fiberutils.ForbiddenResponse(c, "insufficient permissions")
			}
		}

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

// getRolesFromClaims extracts roles from the provided claims.
//
// For a struct, it attempts to find a field tagged with `json` that matches the provided key. If no tagged field is found,
// it falls back to matching the field name directly with the key. The field must be a slice of strings.
//
// For a map, the key is used directly to retrieve the value, which must be a slice of strings.
//
// Returns an error if the claims are not a struct or map, the field is not found, or the field is not a slice of strings.
//
// Example with a struct:
//
//	type UserClaims struct {
//		Roles []string `json:"roles"`
//	}
//
//	claims := UserClaims{Roles: []string{"admin", "user"}}
//	roles, err := getRolesFromClaims(claims, "roles")
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(roles) // Output: [admin, user]
//
// Example with a map:
//
//	claimsMap := map[string]any{"roles": []string{"admin", "user"}}
//	roles, err := getRolesFromClaims(claimsMap, "roles")
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(roles) // Output: [admin, user]
func getRolesFromClaims(claims any, key string) ([]string, error) {
	v := reflect.ValueOf(claims)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	var field reflect.Value
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).Tag.Get("json") == key {
				field = v.Field(i)
				break
			}

			if v.Type().Field(i).Name == key {
				field = v.Field(i)
				break
			}
		}
	case reflect.Map:
		field = v.MapIndex(reflect.ValueOf(key))
	default:
		return nil, fmt.Errorf("claims must be a struct or map, got %v", v.Kind())
	}

	if !field.IsValid() {
		return nil, fmt.Errorf("field %q not found in claims", key)
	}

	if field.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field %q is not a slice, got %v", key, field.Kind())
	}

	roles := make([]string, field.Len())
	for i := 0; i < field.Len(); i++ {
		roles[i] = field.Index(i).Interface().(string)
	}

	return roles, nil
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
	Verify(ctx context.Context, token string) (*oidc.IDToken, error)
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
