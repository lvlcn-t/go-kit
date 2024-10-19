package middleware

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofiber/fiber/v3"
)

func TestNewAuthenticator(t *testing.T) {
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
			provider, err := newAuthenticator[map[string]any](context.Background(), tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("newAuthenticator() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && provider == nil {
				t.Errorf("newAuthenticator() returned nil provider")
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

func TestAuthenticator_Authenticate(t *testing.T) {
	validUnmarshaler := &tokenUnmarshalerMock{
		ClaimsFunc: func(claims any) error {
			if claims == nil || reflect.ValueOf(claims).Kind() != reflect.Pointer {
				t.Fatal("claims is either nil or not a pointer")
			}

			return nil
		},
	}

	tests := []struct {
		name          string
		authenticator Authenticator[struct{}]
		header        http.Header
		wantStatus    int
	}{
		{
			name: "Valid token",
			authenticator: &authProvider[struct{}]{
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
			authenticator: &authProvider[struct{}]{
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
			authenticator: &authProvider[struct{}]{
				verifier: &verifierMock{},
			},
			header:     http.Header{"Authorization": []string{""}},
			wantStatus: fiber.StatusUnauthorized,
		},
		{
			name: "No authorization header",
			authenticator: &authProvider[struct{}]{
				verifier: &verifierMock{},
			},
			header:     http.Header{},
			wantStatus: fiber.StatusUnauthorized,
		},
		{
			name: "Invalid authorization header",
			authenticator: &authProvider[struct{}]{
				verifier: &verifierMock{},
			},
			header:     http.Header{"Authorization": []string{"invalid"}},
			wantStatus: fiber.StatusUnauthorized,
		},
		{
			name: "Failed to unmarshal token",
			authenticator: &authProvider[struct{}]{
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
			app := fiber.New()
			app.Use(tt.authenticator.Authenticate())
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
