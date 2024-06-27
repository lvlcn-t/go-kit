package fiberutils

import (
	"context"
	"html/template"
	"net/http"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/google/go-cmp/cmp"
)

func TestNewErrorResponse(t *testing.T) {
	tests := []struct {
		name   string
		msg    string
		status int
		want   fiber.Map
	}{
		{
			name:   "known status",
			msg:    "message",
			status: http.StatusBadRequest,
			want:   fiber.Map{"error": fiber.Map{"message": "message", "code": http.StatusText(http.StatusBadRequest)}},
		},
		{
			name:   "unknown status",
			msg:    "message",
			status: 999,
			want:   fiber.Map{"error": fiber.Map{"message": "message", "code": "Unknown"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewErrorResponse(tt.msg, tt.status)
			if !cmp.Equal(got, tt.want) {
				t.Error(cmp.Diff(got, tt.want))
			}
		})
	}
}

func Test_Responses(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(c fiber.Ctx, msg string, ctype ...string) error
		wantStatus int
	}{
		{
			name:       "BadRequestResponse",
			fn:         BadRequestResponse,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "UnauthorizedResponse",
			fn:         UnauthorizedResponse,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "ForbiddenResponse",
			fn:         ForbiddenResponse,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "NotFoundResponse",
			fn:         NotFoundResponse,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "InternalServerErrorResponse",
			fn:         InternalServerErrorResponse,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "ServiceUnavailableResponse",
			fn:         ServiceUnavailableResponse,
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c fiber.Ctx) error {
				return tt.fn(c, "message")
			})

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to send request: %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Fatalf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("got status %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestRenderView(t *testing.T) {
	tests := []struct {
		name     string
		template *template.Template
		options  []func(*templ.ComponentHandler)
		wantCT   string
	}{
		{
			name:     "without options",
			template: template.Must(template.New("test").Parse("<h1>Hello, World!</h1>")),
			options:  nil,
			wantCT:   "text/html",
		},
		{
			name:     "with options",
			template: template.Must(template.New("test").Parse("<h1>Hello, World!</h1>")),
			options: []func(*templ.ComponentHandler){
				func(ch *templ.ComponentHandler) {
					ch.ContentType = "text/plain"
				},
			},
			wantCT: "text/plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c fiber.Ctx) error {
				return RenderView(c, templ.FromGoHTML(tt.template, nil), tt.options...)
			})

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to send request: %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Fatalf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("got status %d, want %d", resp.StatusCode, http.StatusOK)
			}

			if !strings.Contains(resp.Header.Get("Content-Type"), tt.wantCT) {
				t.Errorf("got content type %q, want %q", resp.Header.Get("Content-Type"), "text/html")
			}
		})
	}
}
