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
					config: Config{
						Address:  ":8080",
						BasePath: "/",
					},
					app:    app,
					router: app.Group("/"),
					routes: []Route{},
					groups: []RouteGroup{},
					middlewares: []fiber.Handler{
						middleware.Recover(),
						middleware.Logger(),
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
					config: Config{
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
					config: Config{
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
					config: Config{
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
			name:        "New with empty config",
			config:      &Config{},
			middlewares: nil,
			want: func(t *testing.T) *server {
				app := fiber.New()
				return &server{
					mu: sync.Mutex{},
					config: Config{
						Address:  ":8080",
						BasePath: "/",
					},
					app:    app,
					router: app.Group("/"),
					routes: []Route{},
					groups: []RouteGroup{},
					middlewares: []fiber.Handler{
						middleware.Recover(),
						middleware.Logger(),
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
	const defaultRoutes = (1 * 9) // 1 route * 9 methods

	tests := []struct {
		name       string
		server     Server
		routes     []Route
		wantErr    bool
		runTwice   bool
		wantRoutes int
	}{
		{
			name:       "Run without routes",
			server:     New(nil, nil),
			routes:     nil,
			wantErr:    false,
			wantRoutes: defaultRoutes,
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
			wantErr:    false,
			wantRoutes: 1 + defaultRoutes,
		},
		{
			name:   "Run with USE route",
			server: New(nil, nil),
			routes: []Route{
				{
					Methods: []string{MethodUse},
					Path:    "/",
					Handler: func(c fiber.Ctx) error {
						return c.Status(http.StatusOK).SendString("Hello, World!")
					},
				},
			},
			wantErr: false,
			// Since a USE route is the next in the chain of global middleware, it will be added to the routes
			wantRoutes: defaultRoutes,
		},
		{
			name:   "Run with USE route and middleware",
			server: New(nil, nil),
			routes: []Route{
				{
					Methods: []string{MethodUse},
					Path:    "/",
					Handler: func(c fiber.Ctx) error {
						return c.Status(http.StatusOK).SendString("Hello, World!")
					},
					Middlewares: []fiber.Handler{
						middleware.Recover(),
					},
				},
			},
			wantErr:    false,
			wantRoutes: defaultRoutes,
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
			wantErr:    true,
			wantRoutes: defaultRoutes,
		},
		{
			name:       "Run with default healthz route",
			server:     New(&Config{UseDefaultHealthz: true}, nil),
			routes:     nil,
			wantErr:    false,
			wantRoutes: defaultRoutes + 1,
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
			runTwice:   true,
			wantErr:    true,
			wantRoutes: 1 + defaultRoutes,
		},
		{
			name:   "Run with global middlewares",
			server: New(nil, middleware.Recover(), middleware.Logger()),
			routes: []Route{
				{
					Methods: []string{http.MethodGet},
					Path:    "/",
					Handler: func(c fiber.Ctx) error {
						return c.Status(http.StatusOK).SendString("Hello, World!")
					},
				},
			},
			wantErr:    false,
			wantRoutes: 1 + defaultRoutes,
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

			if len(s.app.GetRoutes()) != tt.wantRoutes {
				t.Errorf("server.app.routes = %v, want %v", len(s.app.GetRoutes()), tt.wantRoutes)
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
		{
			name:   "Mount with USE method",
			server: New(nil, nil),
			routes: []Route{
				{
					Path:    "/",
					Methods: []string{MethodUse},
					Handler: func(c fiber.Ctx) error {
						return c.SendString("Hello, World!")
					},
				},
			},
			wantErr:    false,
			wantRoutes: 1,
		},
		{
			name:   "Mount with USE method and other methods",
			server: New(nil, nil),
			routes: []Route{
				{
					Path:    "/",
					Methods: []string{MethodUse, http.MethodGet},
					Handler: func(c fiber.Ctx) error {
						return c.SendString("Hello, World!")
					},
				},
			},
			wantErr:    true,
			wantRoutes: 0,
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

func TestServer_Shutdown(t *testing.T) {
	tests := []struct {
		name    string
		server  Server
		running bool
		wantErr bool
	}{
		{
			name:    "Shutdown without running server",
			server:  New(nil, nil),
			running: false,
			wantErr: false,
		},
		{
			name:    "Shutdown with running server",
			server:  New(nil, nil),
			running: true,
			// Want false because we want indempotent behavior
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.server.(*server)
			s.running = tt.running

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			if !tt.wantErr {
				err := s.Shutdown(ctx)
				if (err != nil) != tt.wantErr {
					t.Errorf("server.Shutdown() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestServer_Restart(t *testing.T) {
	tests := []struct {
		name    string
		server  Server
		running bool
		routes  []Route
		groups  []RouteGroup
		wantErr bool
	}{
		{
			name:    "Restart without routes",
			server:  New(nil, nil),
			running: true,
			routes:  nil,
			groups:  nil,
			wantErr: false,
		},
		{
			name:    "Restart with routes",
			server:  New(nil, nil),
			running: true,
			routes: []Route{
				{
					Methods: []string{http.MethodGet},
					Path:    "/",
					Handler: func(c fiber.Ctx) error {
						return c.Status(http.StatusOK).SendString("Hello, World!")
					},
				},
			},
			groups:  nil,
			wantErr: false,
		},
		{
			name:    "Restart with invalid route",
			server:  New(nil, nil),
			running: true,
			routes: []Route{
				{
					Methods: []string{http.MethodGet},
					Path:    "",
					Handler: nil,
				},
			},
			groups:  nil,
			wantErr: true,
		},
		{
			name:    "Restart with invalid method",
			server:  New(nil, nil),
			running: true,
			routes: []Route{
				{
					Methods: []string{"INVALID"},
					Path:    "/",
					Handler: func(c fiber.Ctx) error {
						return c.SendString("Hello, World!")
					},
				},
			},
			groups:  nil,
			wantErr: true,
		},
		{
			name:    "Restart with no methods",
			server:  New(nil, nil),
			running: true,
			routes: []Route{
				{
					Methods: nil,
					Path:    "/",
					Handler: func(c fiber.Ctx) error {
						return c.SendString("Hello, World!")
					},
				},
			},
			groups:  nil,
			wantErr: true,
		},
		{
			name:    "Restart with groups",
			server:  New(nil, nil),
			running: true,
			routes:  nil,
			groups: []RouteGroup{
				{
					Path: "/api",
					App: fiber.New().Get("/", func(c fiber.Ctx) error {
						return c.SendString("Hello, World!")
					}),
				},
			},
			wantErr: false,
		},
		{
			name:    "Restart without running server",
			server:  New(nil, nil),
			running: false,
			routes:  nil,
			groups:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.server.(*server)
			s.running = tt.running

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			err := s.Restart(ctx, tt.routes, tt.groups)
			if (err != nil) != tt.wantErr {
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Errorf("server.Restart() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestServer_App(t *testing.T) {
	app := fiber.New()
	tests := []struct {
		name   string
		server Server
		want   *fiber.App
	}{
		{
			name: "Get app",
			server: &server{
				app: app,
			},
			want: app,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.server.App(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("server.App() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_Mounted(t *testing.T) {
	tests := []struct {
		name            string
		server          Server
		wantRoutes      int
		wantGroups      int
		wantMiddlewares int
	}{
		{
			name: "No routes, groups or middleware mounted",
			server: &server{
				routes:      nil,
				groups:      nil,
				middlewares: nil,
			},
		},
		{
			name: "Routes mounted",
			server: &server{
				routes: []Route{
					{
						Methods: []string{http.MethodGet},
						Path:    "/",
						Handler: func(c fiber.Ctx) error {
							return c.SendString("Hello, World!")
						},
					},
				},
			},
			wantRoutes: 1,
		},
		{
			name: "Groups mounted",
			server: &server{
				groups: []RouteGroup{
					{
						Path: "/api",
						App: fiber.New().Get("/", func(c fiber.Ctx) error {
							return c.SendString("Hello, World!")
						}),
					},
				},
			},
			wantGroups: 1,
		},
		{
			name: "Middleware mounted",
			server: &server{
				middlewares: []fiber.Handler{
					middleware.Recover(),
					middleware.Logger(),
				},
			},
			wantMiddlewares: 2,
		},
		{
			name: "Routes, groups and middleware mounted",
			server: &server{
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
				middlewares: []fiber.Handler{
					middleware.Recover(),
					middleware.Logger(),
				},
			},
			wantRoutes:      1,
			wantGroups:      1,
			wantMiddlewares: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.server.(*server)

			routes, groups, mw := s.Mounted()
			if len(routes) != tt.wantRoutes {
				t.Errorf("server.Mounted() routes = %v, want %v", routes, tt.wantRoutes)
			}

			if len(groups) != tt.wantGroups {
				t.Errorf("server.Mounted() groups = %v, want %v", groups, tt.wantGroups)
			}

			if len(mw) != tt.wantMiddlewares {
				t.Errorf("server.Mounted() middlewares = %v, want %v", mw, tt.wantMiddlewares)
			}
		})
	}
}

func TestConfig_IsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{
			name:   "Empty config",
			config: &Config{},
			want:   true,
		},
		{
			name: "Non-empty config",
			config: &Config{
				Address: ":8080",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsEmpty(); got != tt.want {
				t.Errorf("Config.IsEmpty() = %v, want %v", got, tt.want)
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
			name: "Valid config",
			config: &Config{
				Address: ":8080",
				TLS: TLSConfig{
					Enabled: false,
				},
			},
			wantErr: false,
		},
		{
			name: "Valid config with TLS",
			config: &Config{
				Address: ":8080",
				TLS: TLSConfig{
					Enabled:     true,
					CertFile:    "cert.pem",
					CertKeyFile: "key.pem",
				},
			},
			wantErr: false,
		},
		{
			name:    "Invalid config",
			config:  &Config{Address: ""},
			wantErr: true,
		},
		{
			name:    "Invalid config with TLS",
			config:  &Config{Address: ":8080", TLS: TLSConfig{Enabled: true}},
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

func TestOkHandler(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{
			name: "OK handler",
			want: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", OkHandler)

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			if err != nil {
				t.Fatalf("http.NewRequestWithContext() error = %v", err)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Fatalf("resp.Body.Close() error = %v", err)
				}
			}()

			if resp.StatusCode != tt.want {
				t.Errorf("OkHandler() = %v, want %v", resp.StatusCode, http.StatusOK)
			}
		})
	}
}
