package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestAuthorizationOptions_Authorize(t *testing.T) {
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
			name:  "Role claims is nil",
			roles: []string{"admin"},
			claims: map[string]any{
				"roles": nil,
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
				authorizer := NewDefaultAuthorizer().WithKey(tt.key).WithRoles("do not skip")
				if roles, ok := tt.roles.([]string); ok || tt.roles == nil {
					authorizer = authorizer.WithRoles(roles...)
				}
				return authorizer.Build().Authorize()(c)
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
