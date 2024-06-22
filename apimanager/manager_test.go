package apimanager

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/go-kit/apimanager/middleware"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		middlewares []fiber.Handler
		want        func(t *testing.T) *server
	}{
		{
			name:        "New without middleware",
			middlewares: nil,
			want: func(t *testing.T) *server {
				app := fiber.New()
				return &server{
					mu: sync.Mutex{},
					config: &Config{
						Address:  ":8080",
						BasePath: "/",
					},
					app:    app,
					router: app.Group("/"),
					routes: []Route{},
					groups: []RouteGroup{},
					middlewares: []fiber.Handler{
						middleware.Recover(),
						middleware.Logger("/healthz"),
					},
				}
			},
		},
		{
			name: "New with middleware",
			middlewares: []fiber.Handler{
				middleware.Recover(),
				middleware.Logger("/healthz"),
			},
			want: func(t *testing.T) *server {
				app := fiber.New()
				return &server{
					mu: sync.Mutex{},
					config: &Config{
						Address:  ":8080",
						BasePath: "/",
					},
					app:    app,
					router: app.Group("/"),
					routes: []Route{},
					groups: []RouteGroup{},
					middlewares: []fiber.Handler{
						middleware.Recover(),
						middleware.Logger("/healthz"),
					},
				}
			},
		},
		{
			name: "New with custom config",
			config: &Config{
				Address:  ":8081",
				BasePath: "/api",
			},
			middlewares: nil,
			want: func(t *testing.T) *server {
				app := fiber.New()
				return &server{
					mu: sync.Mutex{},
					config: &Config{
						Address:  ":8081",
						BasePath: "/api",
					},
					app:    fiber.New(),
					router: app.Group("/api"),
					routes: []Route{},
					groups: []RouteGroup{},
					middlewares: []fiber.Handler{
						middleware.Recover(),
						middleware.Logger("/healthz"),
					},
				}
			},
		},
		{
			name: "New with custom config and middleware",
			config: &Config{
				Address:  ":8081",
				BasePath: "/api",
			},
			middlewares: []fiber.Handler{
				middleware.Recover(),
				middleware.Logger("/healthz"),
			},
			want: func(t *testing.T) *server {
				app := fiber.New()
				return &server{
					mu: sync.Mutex{},
					config: &Config{
						Address:  ":8081",
						BasePath: "/api",
					},
					app:    fiber.New(),
					router: app.Group("/api"),
					routes: []Route{},
					groups: []RouteGroup{},
					middlewares: []fiber.Handler{
						middleware.Recover(),
						middleware.Logger("/healthz"),
					},
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.config, tt.middlewares...)
			want := tt.want(t)
			got := s.(*server)

			if !reflect.DeepEqual(want.config, got.config) {
				t.Errorf("server.config = %v, want %v", got.config, want.config)
			}

			if len(want.middlewares) != len(got.middlewares) {
				t.Errorf("server.middlewares = %v, want %v", got.middlewares, want.middlewares)
			}

			if got.running {
				t.Errorf("server.running = %v, want %v", got.running, false)
			}
		})
	}
}

func TestServer_Run(t *testing.T) {
	tests := []struct {
		name     string
		server   Server
		routes   []Route
		wantErr  bool
		runTwice bool
	}{
		{
			name:    "Run without routes",
			server:  New(nil, nil),
			routes:  nil,
			wantErr: false,
		},
		{
			name:   "Run with routes",
			server: New(nil, nil),
			routes: []Route{
				{
					Methods: []string{http.MethodGet},
					Path:    "/",
					Handler: func(c fiber.Ctx) error {
						return c.Status(http.StatusOK).SendString("Hello, World!")
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "Run with invalid route",
			server: New(nil, nil),
			routes: []Route{
				{
					Methods: []string{http.MethodGet},
					Path:    "/",
					Handler: nil,
				},
			},
			wantErr: true,
		},
		{
			name:   "Try to run server twice",
			server: New(nil, nil),
			routes: []Route{
				{
					Methods: []string{http.MethodGet},
					Path:    "/",
					Handler: func(c fiber.Ctx) error {
						return c.Status(http.StatusOK).SendString("Hello, World!")
					},
				},
			},
			runTwice: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.server.(*server)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			err := s.Mount(tt.routes...)
			if err != nil {
				t.Fatalf("server.Mount() error = %v", err)
			}

			cErr := make(chan error, 1)
			go func() {
				cErr <- s.Run(ctx)
			}()

			time.Sleep(100 * time.Millisecond)

			if tt.runTwice {
				go func() {
					cErr <- s.Run(ctx)
				}()
			}

			s.mu.Lock()
			if !s.running && !tt.wantErr {
				t.Errorf("server.running = %v, want %v", s.running, true)
			}

			routes := len(tt.routes) + 9
			for _, route := range tt.routes {
				if route.Handler == nil {
					routes--
				}
			}

			if len(s.app.GetRoutes()) != routes {
				t.Errorf("server.app.routes = %v, want %v", len(s.app.GetRoutes()), routes)
			}
			s.mu.Unlock()

			err = <-cErr
			if (err != nil) != tt.wantErr {
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Errorf("server.Run() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestServer_Mount(t *testing.T) {
	tests := []struct {
		name       string
		server     Server
		routes     []Route
		groups     []RouteGroup
		running    bool
		wantErr    bool
		wantRoutes int
	}{
		{
			name:    "Mount without routes",
			server:  New(nil, nil),
			routes:  nil,
			groups:  nil,
			wantErr: false,
		},
		{
			name:   "Mount with routes",
			server: New(nil, nil),
			routes: []Route{
				{
					Methods: []string{http.MethodGet},
					Path:    "/",
					Handler: func(c fiber.Ctx) error {
						return c.Status(http.StatusOK).SendString("Hello, World!")
					},
				},
			},
			groups:     nil,
			wantErr:    false,
			wantRoutes: 1,
		},
		{
			name:   "Mount with invalid route",
			server: New(nil, nil),
			routes: []Route{
				{
					Methods: []string{http.MethodGet},
					Path:    "",
					Handler: nil,
				},
			},
			groups:     nil,
			wantErr:    true,
			wantRoutes: 0,
		},
		{
			name:   "Mount with invalid method",
			server: New(nil, nil),
			routes: []Route{
				{
					Methods: []string{"INVALID"},
					Path:    "/",
					Handler: func(c fiber.Ctx) error {
						return c.SendString("Hello, World!")
					},
				},
			},
			groups:     nil,
			wantErr:    true,
			wantRoutes: 0,
		},
		{
			name:   "Mount with no methods",
			server: New(nil, nil),
			routes: []Route{
				{
					Methods: nil,
					Path:    "/",
					Handler: func(c fiber.Ctx) error {
						return c.SendString("Hello, World!")
					},
				},
			},
			groups:     nil,
			wantErr:    true,
			wantRoutes: 0,
		},
		{
			name:   "Mount with server running",
			server: New(nil, nil),
			routes: []Route{
				{
					Methods: []string{http.MethodGet},
					Path:    "/",
					Handler: func(c fiber.Ctx) error {
						return c.SendString("Hello, World!")
					},
				},
			},
			groups: []RouteGroup{
				{
					Path: "/api",
					App: fiber.New().Get("/", func(c fiber.Ctx) error {
						return c.SendString("Hello, World!")
					}),
				},
			},
			running:    true,
			wantErr:    true,
			wantRoutes: 0,
		},
		{
			name:   "Mount with groups",
			server: New(nil, nil),
			routes: nil,
			groups: []RouteGroup{
				{
					Path: "/api",
					App: fiber.New().Get("/", func(c fiber.Ctx) error {
						return c.SendString("Hello, World!")
					}),
				},
			},
			wantErr:    false,
			wantRoutes: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.server.(*server)
			if tt.running {
				s.running = true
			}

			var err error
			if tt.routes != nil {
				for i := range tt.routes {
					t.Logf("Mounting route: method=%v, path=%s", tt.routes[i].Methods, tt.routes[i].Path)
				}
				err = s.Mount(tt.routes...)
			}
			if tt.groups != nil {
				for i := range tt.groups {
					t.Logf("Mounting group: path=%s", tt.groups[i].Path)
				}
				err = s.MountGroup(tt.groups...)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("server.Mount() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.routes != nil && len(s.routes) != tt.wantRoutes {
				t.Errorf("server.routes = %v, want %v", len(s.routes), tt.wantRoutes)
			}

			if tt.groups != nil && len(s.groups) != tt.wantRoutes {
				t.Errorf("server.groups = %v, want %v", len(s.groups), tt.wantRoutes)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "Valid config",
			config:  &Config{Address: ":8080"},
			wantErr: false,
		},
		{
			name:    "Invalid config",
			config:  &Config{Address: ""},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
