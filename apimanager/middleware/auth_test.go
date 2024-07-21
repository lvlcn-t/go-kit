package middleware

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofiber/fiber/v3"
)

func TestNewAuthProvider(t *testing.T) {
	tests := []struct {
		name    string
		config  *AuthConfig
		wantErr bool
	}{
		{
			name: "Valid configuration",
			config: &AuthConfig{
				Config: oidc.Config{
					ClientID: "client-id",
				},
				ProviderURL:  "https://accounts.google.com",
				ClientSecret: "client-secret",
				RedirectURL:  "https://example.com/auth/callback",
				Scopes:       []string{"openid", "profile"},
				validated:    true,
			},
			wantErr: false,
		},
		{
			name: "Invalid configuration",
			config: &AuthConfig{
				Config: oidc.Config{
					ClientID: "client-id",
				},
				ProviderURL:  "",
				ClientSecret: "client-secret",
				RedirectURL:  "https://example.com/auth/callback",
				Scopes:       []string{"openid", "profile"},
				validated:    false,
			},
			wantErr: true,
		},
		{
			name:    "Nil configuration",
			config:  nil,
			wantErr: true,
		},
		{
			name: "Provider not found",
			config: &AuthConfig{
				Config: oidc.Config{
					ClientID: "client-id",
				},
				ProviderURL:  "https://example.com",
				ClientSecret: "client-secret",
				RedirectURL:  "https://example.com/auth/callback",
				Scopes:       []string{"openid", "profile"},
				validated:    true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewAuthProvider(context.Background(), tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAuthProvider() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && provider == nil {
				t.Errorf("NewAuthProvider() returned nil provider")
			}
		})
	}
}

func TestAuthConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *AuthConfig
		wantErr bool
	}{
		{
			name: "Valid configuration",
			config: &AuthConfig{
				Config: oidc.Config{
					ClientID: "client-id",
				},
				ProviderURL:  "https://accounts.google.com",
				ClientSecret: "client-secret",
				RedirectURL:  "https://example.com/auth/callback",
				Scopes:       []string{"openid", "profile"},
			},
			wantErr: false,
		},
		{
			name: "No provider URL",
			config: &AuthConfig{
				Config: oidc.Config{
					ClientID: "client-id",
				},
				ClientSecret: "client-secret",
				RedirectURL:  "https://example.com/auth/callback",
				Scopes:       []string{"openid", "profile"},
			},
			wantErr: true,
		},
		{
			name: "No client ID",
			config: &AuthConfig{
				Config:       oidc.Config{},
				ProviderURL:  "https://accounts.google.com",
				ClientSecret: "client-secret",
				RedirectURL:  "https://example.com/auth/callback",
				Scopes:       []string{"openid", "profile"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthenticateWithClaims(t *testing.T) {
	validUnmarshaler := &tokenUnmarshalerMock{
		ClaimsFunc: func(claims any) error {
			if claims == nil || reflect.ValueOf(claims).Kind() != reflect.Pointer {
				t.Fatal("claims is either nil or not a pointer")
			}

			return nil
		},
	}

	tests := []struct {
		name       string
		provider   *authProvider
		header     http.Header
		wantStatus int
		wantPanic  bool
	}{
		{
			name: "Valid token",
			provider: &authProvider{
				verifier: &verifierMock{
					VerifyFunc: func(ctx context.Context, token string) (tokenUnmarshaler, error) {
						return validUnmarshaler, nil
					},
				},
			},
			header:     http.Header{"Authorization": []string{"Bearer valid-token"}},
			wantStatus: fiber.StatusOK,
		},
		{
			name: "Invalid token",
			provider: &authProvider{
				verifier: &verifierMock{
					VerifyFunc: func(ctx context.Context, token string) (tokenUnmarshaler, error) {
						return nil, fmt.Errorf("invalid token")
					},
				},
			},
			header:     http.Header{"Authorization": []string{"Bearer invalid-token"}},
			wantStatus: fiber.StatusUnauthorized,
		},
		{
			name: "No token",
			provider: &authProvider{
				verifier: &verifierMock{},
			},
			header:     http.Header{"Authorization": []string{""}},
			wantStatus: fiber.StatusUnauthorized,
		},
		{
			name: "No authorization header",
			provider: &authProvider{
				verifier: &verifierMock{},
			},
			header:     http.Header{},
			wantStatus: fiber.StatusUnauthorized,
		},
		{
			name: "Invalid authorization header",
			provider: &authProvider{
				verifier: &verifierMock{},
			},
			header:     http.Header{"Authorization": []string{"invalid"}},
			wantStatus: fiber.StatusUnauthorized,
		},
		{
			name:      "Nil provider",
			provider:  nil,
			header:    http.Header{"Authorization": []string{"Bearer valid-token"}},
			wantPanic: true,
		},
		{
			name: "Nil verifier",
			provider: &authProvider{
				verifier: nil,
			},
			header:    http.Header{"Authorization": []string{"Bearer valid-token"}},
			wantPanic: true,
		},
		{
			name: "Failed to unmarshal token",
			provider: &authProvider{
				verifier: &verifierMock{
					VerifyFunc: func(ctx context.Context, token string) (tokenUnmarshaler, error) {
						return &tokenUnmarshalerMock{
							ClaimsFunc: func(claims any) error {
								return fmt.Errorf("failed to unmarshal token")
							},
						}, nil
					},
				},
			},
			header:     http.Header{"Authorization": []string{"Bearer valid-token"}},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("AuthenticateWithClaims() did not panic")
					}
				}()
			}

			app := fiber.New()
			app.Use(AuthenticateWithClaims[struct{}](tt.provider))
			app.Use(func(c fiber.Ctx) error {
				return c.SendString("next handler")
			})

			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			req.Header = tt.header

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to send request: %v", err)
			}
			defer func() {
				cErr := resp.Body.Close()
				if cErr != nil {
					t.Fatalf("failed to close response body: %v", cErr)
				}
			}()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			t.Logf("Got response: %s", body)
			if tt.wantStatus != resp.StatusCode {
				t.Errorf("want status code %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
			}
		})
	}
}

func TestAuthorizeWithClaims(t *testing.T) {
	tests := []struct {
		name         string
		roles        any
		claims       any
		key          string
		wantStatus   int
		wantResponse string
	}{
		{
			name:         "Authorized with correct roles",
			roles:        []string{"admin"},
			claims:       map[string]any{"roles": []string{"admin", "user"}},
			key:          "roles",
			wantStatus:   fiber.StatusOK,
			wantResponse: "next handler",
		},
		{
			name:         "Unauthorized with incorrect roles",
			roles:        []string{"admin"},
			claims:       map[string]any{"roles": []string{"user"}},
			key:          "roles",
			wantStatus:   fiber.StatusForbidden,
			wantResponse: "insufficient permissions",
		},
		{
			name:         "No roles found in claims",
			roles:        []string{"admin"},
			claims:       nil,
			key:          "roles",
			wantStatus:   fiber.StatusForbidden,
			wantResponse: "no claims found",
		},
		{
			name:       "Claims does not contain key",
			roles:      []string{"admin"},
			claims:     map[string]any{"roles": []string{"admin"}},
			key:        "invalid-key",
			wantStatus: fiber.StatusInternalServerError,
		},
		{
			name:         "No roles provided",
			roles:        nil,
			claims:       map[string]any{"roles": []string{"admin"}},
			key:          "roles",
			wantStatus:   fiber.StatusOK,
			wantResponse: "next handler",
		},
		{
			name:       "No roles provided and no roles in claims",
			roles:      nil,
			claims:     nil,
			key:        "roles",
			wantStatus: fiber.StatusOK,
		},
		{
			name:       "No roles provided and no claims",
			roles:      nil,
			claims:     nil,
			key:        "roles",
			wantStatus: fiber.StatusOK,
		},
		{
			name:  "Role claims is not a slice",
			roles: []string{"admin"},
			claims: map[string]any{
				"roles": "admin",
			},
			key:        "roles",
			wantStatus: fiber.StatusInternalServerError,
		},
		{
			name:  "Role claims is not a slice of strings",
			roles: []string{"admin"},
			claims: map[string]any{
				"roles": []int{1, 2, 3},
			},
			key:        "roles",
			wantStatus: fiber.StatusInternalServerError,
		},
		{
			name:  "No key provided",
			roles: []string{"admin"},
			claims: map[string]any{
				"roles": []string{"admin"},
			},
			key:        "",
			wantStatus: fiber.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c fiber.Ctx) error {
				if tt.claims != nil {
					c.Locals("claims", tt.claims)
				}
				if roles, ok := tt.roles.([]string); ok || tt.roles == nil {
					return AuthorizeWithClaims[map[string]any](AuthorizationOptions{Key: tt.key, Roles: roles})(c)
				}
				return AuthorizeWithClaims[map[string]any](AuthorizationOptions{Key: tt.key, Roles: []string{"do not skip"}})(c)
			})

			app.Use(func(c fiber.Ctx) error {
				return c.SendString("next handler")
			})

			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to send request: %v", err)
			}
			defer func() {
				cErr := resp.Body.Close()
				if cErr != nil {
					t.Fatalf("failed to close response body: %v", cErr)
				}
			}()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			t.Logf("Got response: %s", body)
			if tt.wantStatus != resp.StatusCode {
				t.Errorf("want status code %d, got %d", tt.wantStatus, resp.StatusCode)
			}

			if tt.wantResponse != "" && !strings.Contains(string(body), tt.wantResponse) {
				t.Errorf("response body does not contain wanted response: %s", tt.wantResponse)
			}
		})
	}
}

func TestGetRolesFromClaims(t *testing.T) {
	tests := []struct {
		name      string
		claims    any
		key       string
		wantRoles []string
		wantErr   bool
	}{
		{
			name:      "Valid roles in map",
			claims:    map[string]any{"roles": []string{"admin", "user"}},
			key:       "roles",
			wantRoles: []string{"admin", "user"},
			wantErr:   false,
		},
		{
			name: "Valid roles in struct with tags",
			claims: struct {
				Roles []string `json:"roles"`
			}{Roles: []string{"admin", "user"}},
			key:       "roles",
			wantRoles: []string{"admin", "user"},
			wantErr:   false,
		},
		{
			name:      "Valid roles in struct without tags",
			claims:    struct{ Roles []string }{Roles: []string{"admin", "user"}},
			key:       "roles",
			wantRoles: []string{"admin", "user"},
			wantErr:   false,
		},
		{
			name:      "No roles found",
			claims:    map[string]any{"roles": []string{}},
			key:       "roles",
			wantRoles: []string{},
			wantErr:   false,
		},
		{
			name:      "No roles key found",
			claims:    map[string]any{"invalid-key": []string{"admin", "user"}},
			key:       "roles",
			wantRoles: nil,
			wantErr:   true,
		},
		{
			name:      "Roles is not a slice",
			claims:    map[string]any{"roles": "admin"},
			key:       "roles",
			wantRoles: nil,
			wantErr:   true,
		},
		{
			name:      "Roles is not a slice of strings",
			claims:    map[string]any{"roles": []int{1, 2, 3}},
			key:       "roles",
			wantRoles: nil,
			wantErr:   true,
		},
		{
			name:      "Claims is neither a struct nor a map",
			claims:    "invalid-claims",
			key:       "roles",
			wantRoles: nil,
			wantErr:   true,
		},
		{
			name:      "Claims is nil",
			claims:    nil,
			key:       "roles",
			wantRoles: nil,
			wantErr:   true,
		},
		{
			name:      "Nested roles in map",
			claims:    map[string]any{"nested": map[string]any{"roles": []string{"admin", "user"}}},
			key:       "nested.roles",
			wantRoles: []string{"admin", "user"},
			wantErr:   false,
		},
		{
			name: "Nested roles in struct with tags",
			claims: struct {
				Nested struct {
					Roles []string `json:"roles"`
				}
			}{
				Nested: struct {
					Roles []string "json:\"roles\""
				}{
					Roles: []string{"admin", "user"},
				},
			},
			key:       "nested.roles",
			wantRoles: []string{"admin", "user"},
			wantErr:   false,
		},
		{
			name:      "Nested roles in struct without tags",
			claims:    struct{ Nested struct{ Roles []string } }{Nested: struct{ Roles []string }{Roles: []string{"admin", "user"}}},
			key:       "nested.roles",
			wantRoles: []string{"admin", "user"},
			wantErr:   false,
		},
		{
			name: "Field not found in struct",
			claims: struct {
				Nested struct {
					Roles []string `json:"roles"`
				}
			}{
				Nested: struct {
					Roles []string "json:\"roles\""
				}{
					Roles: []string{"admin", "user"},
				},
			},
			key:       "nested.invalid",
			wantRoles: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roles, err := getRolesFromClaims(tt.claims, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRolesFromClaims() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && !reflect.DeepEqual(roles, tt.wantRoles) {
				t.Errorf("want roles %v, got %v", tt.wantRoles, roles)
			}
		})
	}
}
