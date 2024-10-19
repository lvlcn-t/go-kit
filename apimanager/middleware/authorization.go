package middleware

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/go-kit/apimanager/fiberutils"
	"github.com/lvlcn-t/loggerhead/logger"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Authorizer represents an authorizer that checks if the request is authorized based on roles.
// The provided type T is the type of the claims.
type Authorizer[T any] interface {
	// Authorize creates a middleware that checks if the request is authorized based on roles.
	// The token are extracted from the [fiber.Ctx].Locals("claims") of type T using the provided key.
	//
	// Note:
	//   - If the roles claim is nested, use a period as a separator.
	//   - If no key is provided, it defaults to "roles".
	//   - If the [Authorizer] has no ExtractUserRoles function provided, it tries to extract the roles from the claims directly.
	//
	// Example:
	//
	//	type Claims struct {
	//	    Issuer        string `json:"iss"`
	//	    Subject       string `json:"sub"`
	//	    Permissions   struct {
	//	        Roles []string `json:"roles"`
	//	    } `json:"permissions"`
	//	}
	//
	//	authorizer := middleware.NewAuthorizer[Claims]().
	//		WithKey("permissions.roles").
	//		WithRoles("admin").
	//		Build()
	//
	//	app.Use(authorizer.Authorize())
	Authorize() fiber.Handler
}

// AuthorizerBuilder is used to build an [Authorizer].
type AuthorizerBuilder[T any] interface {
	// WithKey sets the key used to extract the roles from the claims.
	WithKey(key string) AuthorizerBuilder[T]
	// WithRoles sets the roles required to access the route.
	WithRoles(roles ...string) AuthorizerBuilder[T]
	// WithRoleExtractor sets the function used to extract roles from the claims.
	// The function must return the roles and an error if the roles cannot be extracted from the provided token.
	WithRoleExtractor(f func(token T) ([]string, error)) AuthorizerBuilder[T]
	// Build assembles the [Authorizer] with the provided options.
	Build() Authorizer[T]
}

// authorizer is used to authorize requests based on roles.
// The provided type T is the type of the claims.
type authorizer[T any] struct {
	// key is the key used to extract the roles from the claims with the default [AuthorizationOptions].RoleExtractor.
	// If you are not using a custom role extraction function, use periods to indicate nested fields.
	key string
	// roles are the roles required to access the route.
	roles []string
	// extractUserRoles is used to extract roles from the claims.
	// Returns the roles and an error if the roles cannot be extracted.
	extractUserRoles func(token T) ([]string, error)
}

// NewDefaultAuthorizer initializes a new [Authorizer] expecting claims of type map[string]any.
//
// Note: This is a convenience function for when the stored token claims are not known.
// You may want to use this if you are using the default [Authenticator] with the default claims type (map[string]any).
//
// Example:
//
//	authorizer := middleware.NewDefaultAuthorizer().
//		WithKey("roles").
//		WithRoles("admin").
//		Build()
//
//	app.Use(authorizer.Authorize())
func NewDefaultAuthorizer() AuthorizerBuilder[map[string]any] {
	return authorizer[map[string]any]{}
}

// NewAuthorizer initializes a new [Authorizer] expecting claims of type T.
//
// Example:
//
//	type Claims struct {
//	    Issuer  string   `json:"iss"`
//	    Subject string   `json:"sub"`
//	    Roles   []string `json:"roles"`
//	}
//
//	authorizer := middleware.NewAuthorizer[Claims]().
//		WithRoles("admin").
//		WithRoleExtractor(func(token Claims) ([]string, error) {
//			return token.Roles, nil
//		}).
//		Build()
//
//	app.Use(authorizer.Authorize())
func NewAuthorizer[T any]() AuthorizerBuilder[T] {
	return authorizer[T]{}
}

// WithKey sets the key used to extract the roles from the claims.
func (a authorizer[T]) WithKey(key string) AuthorizerBuilder[T] {
	a.key = key
	return a
}

// WithRoles sets the roles required to access the route.
func (a authorizer[T]) WithRoles(roles ...string) AuthorizerBuilder[T] {
	a.roles = roles
	return a
}

// WithRoleExtractor sets the function used to extract roles from the claims.
func (a authorizer[T]) WithRoleExtractor(f func(token T) ([]string, error)) AuthorizerBuilder[T] {
	a.extractUserRoles = f
	return a
}

// Build assembles the [Authorizer] with the provided options.
func (a authorizer[T]) Build() Authorizer[T] {
	return a
}

// withDefaults sets the default values for the options if they are not provided.
func (a authorizer[T]) withDefaults() authorizer[T] {
	if a.key == "" {
		a.key = "roles"
	}

	if a.extractUserRoles == nil {
		a.extractUserRoles = func(token T) ([]string, error) {
			return getRolesFromClaims(token, a.key)
		}
	}

	return a
}

// Authorize creates a middleware that checks if the request is authorized based on roles.
func (a authorizer[T]) Authorize() fiber.Handler {
	a = a.withDefaults()
	return func(c fiber.Ctx) error {
		if len(a.roles) == 0 {
			return c.Next()
		}

		log := logger.FromContext(c.UserContext())
		claims, ok := c.Locals("claims").(T)
		if !ok {
			log.WarnContext(c.Context(), "No claims found or invalid type", "claims", claims)
			return fiberutils.ForbiddenResponse(c, "no claims found")
		}

		roles, err := a.extractUserRoles(claims)
		if err != nil {
			log.ErrorContext(c.Context(), "Failed to get roles from claims", "error", err)
			return fiberutils.InternalServerErrorResponse(c, "failed to get roles from claims")
		}

		userRoles := map[string]bool{}
		for _, role := range roles {
			userRoles[role] = true
		}

		for _, role := range a.roles {
			if !userRoles[role] {
				log.DebugContext(c.Context(), "Insufficient permissions", "roles", a.roles)
				return fiberutils.ForbiddenResponse(c, "insufficient permissions")
			}
		}

		return c.Next()
	}
}

// getRolesFromClaims extracts roles from the provided claims.
//
// For a struct, it attempts to find a field tagged with `json` that matches the provided key. If no tagged field is found,
// it falls back to matching the field name directly with the key. The field must be a slice of strings.
//
// For a map, the key is used directly to retrieve the value, which must be a slice of strings.
//
// Returns an error if the claims are not a struct or map, the field is not found, or the field is not a slice of strings.
func getRolesFromClaims[T any](claims T, key string) ([]string, error) {
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
		default:
			return reflect.Value{}, fmt.Errorf("field %q is neither a struct nor a map", part)
		}
		if !val.IsValid() {
			return reflect.Value{}, fmt.Errorf("field %q not found", part)
		}
		if val.Kind() == reflect.Interface {
			val = val.Elem()
			if !val.IsValid() {
				return reflect.Value{}, fmt.Errorf("field %q is nil", part)
			}
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
