package middleware

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/oauth2"
)

// AuthProvider represents an OIDC provider.
//
// Deprecated: Use the [Authenticator] interface instead.
// Will be removed with v0.5.0.
type AuthProvider struct {
	// verifier is used to verify ID tokens.
	verifier verifier
	// config is used to configure the OAuth2 client.
	config oauth2.Config
}

// NewAuthProvider initializes a new OIDC provider.
// Returns an error if the configuration is invalid or the provider cannot be initialized.
//
// Deprecated: Use [NewDefaultAuthenticator] or [NewAuthenticator] instead.
// Will be removed with v0.5.0.
func NewAuthProvider(ctx context.Context, c *AuthConfig) (*AuthProvider, error) {
	// TODO: Remove this function in v0.5.0.
	a, err := newAuthenticator[map[string]any](ctx, c)
	return &AuthProvider{verifier: a.verifier, config: a.config}, err
}

// Authenticate creates a middleware that verifies if the request is authenticated.
// The token claims are extracted and stored in the [fiber.Ctx] locals with the key "claims" of type map[string]any.
//
// Panics if the provider is nil or the provider's verifier is nil.
//
// Deprecated: Use [Authenticator.Authenticate] instead.
// Will be removed with v0.5.0.
func Authenticate(provider *AuthProvider) fiber.Handler {
	// TODO: Remove this function in v0.5.0.
	return AuthenticateWithClaims[map[string]any](provider)
}

// AuthenticateWithClaims creates a middleware that verifies if the request is authenticated.
// The token claims are stored in the [fiber.Ctx] locals with the key "claims" and are of the provided type T.
//
// Panics if the provider is nil or the provider's verifier is nil.
//
// Deprecated: Use [Authenticator.Authenticate] instead.
// Will be removed with v0.5.0.
func AuthenticateWithClaims[T any](provider *AuthProvider) fiber.Handler {
	// TODO: Remove this function in v0.5.0.
	if provider == nil || provider.verifier == nil {
		panic("provider or verifier is nil")
	}
	p := &authProvider[T]{verifier: provider.verifier, config: provider.config}
	return p.Authenticate()
}

// AuthorizationOptions represents the options for the authorization middleware.
//
// Deprecated: Use the [Authorizer] interface instead.
// Will be removed with v0.5.0.
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
// The token are extracted from the [fiber.Ctx].Locals("claims") of type map[string]any using the provided key.
//
// Note:
//   - When using this function, the type is assumed to be map[string]any. To use a different type, use [AuthorizeWithClaims].
//   - If the roles claim is nested, use a period as a separator.
//   - If no key is provided, it defaults to "roles".
//   - If the [AuthorizationOptions].GetRolesFromToken function is not provided, it tries to extract the roles from the claims directly.
//
// Deprecated: Use [Authorizer.Authorize] instead.
// Will be removed with v0.5.0.
func Authorize(options AuthorizationOptions) fiber.Handler {
	// TODO: Remove this function in v0.5.0.
	return AuthorizeWithClaims[map[string]any](options)
}

// AuthorizeWithClaims creates a middleware that checks if the request is authorized based on roles.
// The roles are extracted from the local claims of type T using the provided key.
//
// Note:
//   - If the roles claim is nested, use a period as a separator.
//   - If no key is provided, it defaults to "roles".
//   - If the [AuthorizationOptions].GetRolesFromToken function is not provided, it tries to extract the roles from the claims directly.
//
// Deprecated: Use [Authorizer.Authorize] instead.
// Will be removed with v0.5.0.
func AuthorizeWithClaims[T any](options AuthorizationOptions) fiber.Handler {
	// TODO: Remove this function in v0.5.0.
	if options.GetRolesFromToken == nil {
		options.GetRolesFromToken = getRolesFromClaims
	}
	return NewAuthorizer[T]().
		WithKey(options.Key).
		WithRoles(options.Roles...).
		WithRoleExtractor(func(token T) ([]string, error) {
			return options.GetRolesFromToken(token, options.Key)
		}).Build().Authorize()
}
