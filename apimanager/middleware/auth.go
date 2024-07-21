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
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// authProvider represents an OIDC provider.
type authProvider struct {
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
func NewAuthProvider(ctx context.Context, c *AuthConfig) (*authProvider, error) {
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

	return &authProvider{
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
// The token claims are extracted and stored in the context's locals with the key "claims".
//
// Panics if the provider is nil.
func Authenticate(provider *authProvider) fiber.Handler {
	return AuthenticateWithClaims[map[string]any](provider)
}

// AuthenticateWithClaims creates a middleware that verifies if the request is authenticated.
// The token claims are stored in the context's locals with the key "claims" and are of the provided type T.
//
// Panics if the provider is nil or the provider's verifier is nil.
func AuthenticateWithClaims[T any](provider *authProvider) fiber.Handler {
	if provider == nil {
		panic("provider is nil")
	}
	if provider.verifier == nil {
		panic("the provider's verifier is nil")
	}

	return func(c fiber.Ctx) error {
		log := logger.FromContext(c.UserContext())
		token, err := extractToken(c)
		if err != nil {
			log.DebugContext(c.Context(), "Failed to extract token", "error", err)
			return fiberutils.UnauthorizedResponse(c, err.Error())
		}

		idToken, err := provider.verifier.Verify(c.Context(), token)
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

type AuthorizationOptions struct {
	// Key is the key used to extract the roles from the claims.
	// If you are not using a custom role extraction function, use periods to indicate nested fields.
	Key string
	// Roles are the roles required to access the route.
	Roles []string
	// GetRolesFromToken is used to extract roles from the claims.
	// Returns the roles and an error if the roles cannot be extracted.
	GetRolesFromToken func(token any, key string) ([]string, error)
}

// Authorize creates a middleware that checks if the request is authorized based on roles.
// The roles are extracted from the claims stored in the context's locals as map[string]any using the provided key.
// If the roles claim is nested, use a period as a separator.
// If no key is provided, it defaults to "roles". To use a different type, use [AuthorizeWithClaims].
// If no custom GetRolesFromToken function is provided, it defaults to [getRolesFromClaims].
func Authorize(options AuthorizationOptions) fiber.Handler {
	return AuthorizeWithClaims[map[string]any](options)
}

// AuthorizeWithClaims creates a middleware that checks if the request is authorized based on roles.
// The roles are extracted from the local claims of type T using the provided key.
// If the roles claim is nested, use a period as a separator.
// If no key is provided, it defaults to "roles".
func AuthorizeWithClaims[T any](options AuthorizationOptions) fiber.Handler {
	if options.Key == "" {
		options.Key = "roles"
	}

	if options.GetRolesFromToken == nil {
		options.GetRolesFromToken = getRolesFromClaims
	}

	return func(c fiber.Ctx) error {
		if len(options.Roles) == 0 {
			return c.Next()
		}

		log := logger.FromContext(c.UserContext())
		claims, ok := c.Locals("claims").(T)
		if !ok {
			log.WarnContext(c.Context(), "No claims found or invalid type", "claims", claims)
			return fiberutils.ForbiddenResponse(c, "no claims found")
		}

		roles, err := options.GetRolesFromToken(claims, options.Key)
		if err != nil {
			log.ErrorContext(c.Context(), "Failed to get roles from claims", "error", err)
			return fiberutils.InternalServerErrorResponse(c, "failed to get roles from claims")
		}

		roleMap := make(map[string]bool)
		for _, role := range roles {
			roleMap[role] = true
		}

		for _, role := range options.Roles {
			if !roleMap[role] {
				log.DebugContext(c.Context(), "Insufficient permissions", "roles", options.Roles)
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
func getRolesFromClaims(claims any, key string) ([]string, error) {
	field, err := getRolesField(reflect.ValueOf(claims), key)
	if err != nil {
		return nil, err
	}

	if field.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field %q is not a slice, got %v", key, field.Kind())
	}

	roles := make([]string, field.Len())
	for i := 0; i < field.Len(); i++ {
		f := field.Index(i)
		if f.Kind() != reflect.String {
			return nil, fmt.Errorf("field %q is not a slice of strings, got %v", key, f.Kind())
		}
		roles[i] = f.Interface().(string)
	}

	return roles, nil
}

// getRolesField retrieves the roles field from the claims using the provided key.
// To indicate a nested field, use a period as a separator.
func getRolesField(val reflect.Value, key string) (reflect.Value, error) {
	parts := strings.Split(key, ".")
	for _, part := range parts {
		val = reflect.Indirect(val)
		switch val.Kind() {
		case reflect.Struct:
			val = getStructField(val, part)
		case reflect.Map:
			val = val.MapIndex(reflect.ValueOf(part))
			if val.Kind() == reflect.Interface {
				val = val.Elem()
			}
		default:
			return reflect.Value{}, fmt.Errorf("field %q is neither a struct nor a map", part)
		}
		if !val.IsValid() {
			return reflect.Value{}, fmt.Errorf("field %q not found", part)
		}
	}
	return val, nil
}

// titler is used to title an english string.
var titler = cases.Title(language.AmericanEnglish)

// getStructField retrieves a field from a struct using the provided name.
func getStructField(val reflect.Value, name string) reflect.Value {
	if val.Kind() != reflect.Struct {
		return reflect.Value{}
	}

	for i := 0; i < val.NumField(); i++ {
		if val.Type().Field(i).Tag.Get("json") == name {
			return val.Field(i)
		}
	}

	for i := 0; i < val.NumField(); i++ {
		// The name is capitalized to match an exported field.
		// Otherwise, an unexposed field could be extracted which would always be invalid.
		if val.Type().Field(i).Name == titler.String(name) {
			return val.Field(i)
		}
	}

	return reflect.Value{}
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
