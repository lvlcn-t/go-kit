package middleware

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestContext(t *testing.T) {
	type key struct{}

	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "success",
			ctx:  context.WithValue(context.Background(), key{}, "value"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			_ = app.Use(Context(tt.ctx))
			_ = app.Get("/", func(c fiber.Ctx) error {
				if got := c.UserContext().Value(key{}); got != "value" {
					t.Errorf("expected %v, got %v", "value", got)
				}
				return nil
			})
			req, err := http.NewRequestWithContext(tt.ctx, http.MethodGet, "/", http.NoBody)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to test app: %v", err)
			}
			defer func() {
				err := resp.Body.Close()
				if err != nil {
					t.Fatalf("failed to close response body: %v", err)
				}
			}()
		})
	}
}

func TestLogger(t *testing.T) {
	tests := []struct {
		name   string
		paths  []string
		ignore []string
	}{
		{
			name:   "success without ignore",
			paths:  []string{"/"},
			ignore: nil,
		},
		{
			name:   "success with ignoring all paths",
			paths:  []string{"/"},
			ignore: []string{"/"},
		},
		{
			name:   "success with ignoring one path",
			paths:  []string{"/", "/path"},
			ignore: []string{"/path"},
		},
		{
			name:   "success with ignoring multiple paths",
			paths:  []string{"/", "/path", "/path2"},
			ignore: []string{"/path", "/path2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			_ = app.Use(Logger(tt.ignore...))
			for _, path := range tt.paths {
				_ = app.Get(path, func(c fiber.Ctx) error {
					return nil
				})
			}
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to test app: %v", err)
			}
			defer func() {
				err := resp.Body.Close()
				if err != nil {
					t.Fatalf("failed to close response body: %v", err)
				}
			}()
		})
	}
}

func TestRecovery(t *testing.T) {
	app := fiber.New()
	_ = app.Use(Recover())
	_ = app.Get("/", func(c fiber.Ctx) error {
		panic("test")
	})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test app: %v", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Fatalf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status code %v, got %v", http.StatusInternalServerError, resp.StatusCode)
	}
}
