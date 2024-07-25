// Package apimanager provides a way to create and manage API servers.
package apimanager

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/go-kit/apimanager/middleware"
	"github.com/lvlcn-t/loggerhead/logger"
)

const (
	// MethodUse is a custom HTTP method to indicate that a route should be registered for all HTTP methods.
	MethodUse = "USE"
	// shutdownTimeout is the timeout for the server to shut down.
	shutdownTimeout = 15 * time.Second
)

type Server interface {
	// Run attaches all previously mounted routes and starts the server.
	// Runs indefinitely until an error occurs, the server shuts down, or the provided context is done.
	//
	// Example setup:
	// 	server := apimanager.New(nil)
	// 	server.Mount(apimanager.Route{
	// 		Path:    "/",
	// 		Methods: []string{http.MethodGet},
	// 		Handler: func(c fiber.Ctx) error {
	// 			return c.SendString("Hello, World!")
	// 		},
	// 	})
	// 	// The server will listen on the default address ":8080" and respond with "Hello, World!" on a GET request to "/".
	// 	server.Run(context.Background())
	Run(ctx context.Context) error
	// Restart restarts the server by shutting it down and starting it again.
	// If any routes or groups are provided, they will be added to the server.
	// All existing routes and groups will be preserved.
	Restart(ctx context.Context, routes []Route, groups []RouteGroup) error
	// Shutdown gracefully shuts down the server.
	Shutdown(ctx context.Context) error
	// Mount adds the provided routes to the server.
	Mount(routes ...Route) error
	// MountGroup adds the provided route groups to the server.
	MountGroup(groups ...RouteGroup) error
	// App returns the fiber app of the server.
	App() *fiber.App
	// Mounted returns all mounted routes, groups, and global middlewares.
	Mounted() (routes []Route, groups []RouteGroup, middlewares []fiber.Handler)
}

// Route is a route to register to the server.
type Route struct {
	// Path is the path of the route.
	Path string
	// Methods is the HTTP method of the route.
	// To register the route to all http methods, use [MethodUse].
	// [MethodUse] is mutually exclusive with other methods.
	Methods []string
	// Handler is the handler function of the route.
	Handler fiber.Handler
	// Middlewares are the middlewares to use for the route.
	Middlewares []fiber.Handler
}

// RouteGroup is a route to register a sub-app to.
type RouteGroup struct {
	// Path is the path of the route.
	Path string
	// App is the fiber sub-app to use.
	App fiber.Router
}

// Config is the configuration of the server.
type Config struct {
	// Address is the address to listen on.
	Address string `yaml:"address" mapstructure:"address"`
	// BasePath is the base path of the API.
	BasePath string `yaml:"basePath" mapstructure:"basePath"`
	// UseDefaultHealthz indicates if the default healthz handler should be used.
	UseDefaultHealthz bool `yaml:"useDefaultHealthz" mapstructure:"useDefaultHealthz"`
	// TLS is the TLS configuration.
	TLS TLSConfig `yaml:"tls" mapstructure:"tls"`
}

// TLSConfig is the TLS configuration.
type TLSConfig struct {
	// Enabled indicates if TLS is enabled.
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
	// CertFile is the path to the certificate file.
	CertFile string `yaml:"certPath" mapstructure:"certPath"`
	// CertKeyFile is the path to the certificate key file.
	CertKeyFile string `yaml:"keyPath" mapstructure:"keyPath"`
}

// IsEmpty checks if the configuration is empty.
func (c Config) IsEmpty() bool { //nolint:gocritic // To ensure compatibility with viper, no pointer receiver is used.
	return reflect.DeepEqual(c, Config{})
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	var err error
	if c.Address == "" {
		err = errors.New("api.address is required")
	}

	if c.TLS.Enabled {
		if c.TLS.CertFile == "" {
			err = errors.Join(err, errors.New("api.tls.certPath is required"))
		}

		if c.TLS.CertKeyFile == "" {
			err = errors.Join(err, errors.New("api.tls.keyPath is required"))
		}
	}

	return err
}

// server is the server implementation.
type server struct {
	// mu is the mutex to synchronize access to the server.
	mu sync.Mutex
	// config is the configuration of the server.
	config Config
	// app is the fiber app of the server.
	app *fiber.App
	// router is the fiber root router of the server.
	router fiber.Router
	// routes are the routes to mount to the server on startup.
	routes []Route
	// groups are the route groups to mount to the server on startup.
	groups []RouteGroup
	// middlewares are the global middlewares to use for the server.
	middlewares []fiber.Handler
	// running indicates if the server is running.
	running bool
}

// New creates a new server with the provided configuration and middlewares.
// If no configuration is provided, a default configuration will be used.
// If no middlewares are provided, a default set of middlewares will be used.
func New(c *Config, middlewares ...fiber.Handler) Server {
	if c == nil {
		c = &Config{
			Address:           ":8080",
			BasePath:          "/",
			UseDefaultHealthz: false,
		}
	}

	if c.Address == "" {
		c.Address = ":8080"
	}

	if c.BasePath == "" {
		c.BasePath = "/"
	}

	app := fiber.New()
	if len(middlewares) == 0 {
		middlewares = append(middlewares, middleware.Recover(), middleware.Logger())
	}

	return &server{
		mu:          sync.Mutex{},
		config:      *c,
		app:         app,
		router:      app.Group(c.BasePath),
		routes:      []Route{},
		groups:      []RouteGroup{},
		middlewares: middlewares,
		running:     false,
	}
}

// Run attaches all previously mounted routes and starts the server.
// Runs indefinitely until an error occurs, the server shuts down, or the provided context is done.
// The provided context will also be injected into the (fiber.Ctx).UserContext() of the request.
func (s *server) Run(ctx context.Context) error {
	err := s.attachRoutes(ctx)
	if err != nil {
		return err
	}

	var cfg []fiber.ListenConfig
	if s.config.TLS.Enabled {
		cfg = append(cfg, fiber.ListenConfig{
			CertFile:    s.config.TLS.CertFile,
			CertKeyFile: s.config.TLS.CertKeyFile,
		})
	}

	cErr := make(chan error, 1)
	go func() {
		cErr <- s.app.Listen(s.config.Address, cfg...)
	}()

	select {
	case <-ctx.Done():
		return s.Shutdown(ctx)
	case err := <-cErr:
		return err
	}
}

// Mount adds the provided routes to the server.
//
// Note that mounting routes after the server has started will have no effect and will return an error.
func (s *server) Mount(routes ...Route) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return &ErrAlreadyRunning{}
	}

	for i := range routes {
		if len(routes[i].Methods) == 0 {
			return fmt.Errorf("route %q has no methods", routes[i].Path)
		}

		hasUse := false
		for _, method := range routes[i].Methods {
			if method == MethodUse {
				hasUse = true
				break
			}

			if !isValid(strings.ToUpper(method)) {
				return fmt.Errorf("route %q has invalid method %q", routes[i].Path, method)
			}
		}

		if hasUse && len(routes[i].Methods) > 1 {
			return fmt.Errorf("route %q has method %q and other methods: %v", routes[i].Path, MethodUse, routes[i].Methods)
		}

		if routes[i].Path == "" {
			return fmt.Errorf("route %q has no path", routes[i].Path)
		}
	}

	s.routes = append(s.routes, routes...)
	return nil
}

// MountGroup adds the provided route groups to the server.
//
// Note that mounting route groups after the server has started will have no effect and will return an error.
func (s *server) MountGroup(groups ...RouteGroup) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return &ErrAlreadyRunning{}
	}

	s.groups = append(s.groups, groups...)
	return nil
}

// attachRoutes attaches the routes to the server.
func (s *server) attachRoutes(ctx context.Context) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return &ErrAlreadyRunning{}
	}

	// Always inject the provided context into the request user context.
	// To ensure all routes have access to the same logger a new logger instance is created and
	// injected into the context if not already present.
	_ = s.router.Use(middleware.Context(logger.IntoContext(ctx, logger.FromContext(ctx))))
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to mount routes: %v", r)
		}
	}()

	for _, group := range s.groups {
		if app, ok := group.App.(*fiber.App); ok {
			routes := app.GetRoutes()
			for i := range routes {
				logger.FromContext(ctx).InfoContext(ctx, "Mounting route", "path", routes[i].Path, "method", routes[i].Method)
			}
		}
		_ = s.router.Use(group.Path, group.App)
	}

	if s.config.UseDefaultHealthz {
		s.routes = append(s.routes, Route{
			Path:    "/healthz",
			Methods: []string{http.MethodGet},
			Handler: OkHandler,
		})
	}

	for _, route := range s.routes {
		logger.FromContext(ctx).InfoContext(ctx, "Mounting route", "path", route.Path, "methods", strings.Join(route.Methods, ","))
		_ = s.router.Add(route.Methods, route.Path, route.Handler, route.Middlewares...)
	}

	s.running = true
	return nil
}

// Shutdown gracefully shuts down the server.
func (s *server) Shutdown(ctx context.Context) error {
	c, cancel := newContextWithTimeout(ctx)
	defer cancel()

	return errors.Join(ctx.Err(), s.app.ShutdownWithContext(c))
}

// Restart restarts the server by shutting it down and starting it again.
// If any routes or groups are provided, they will be added to the server.
// All existing routes and groups will be preserved.
// Runs indefinitely until an error occurs, the server shuts down, or the provided context is done.
func (s *server) Restart(ctx context.Context, routes []Route, groups []RouteGroup) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return &ErrNotRunning{}
	}
	s.mu.Unlock()

	s.toggleRunning()
	defer s.toggleRunning()
	if len(routes) > 0 {
		err := s.Mount(routes...)
		if err != nil {
			return err
		}
	}

	if len(groups) > 0 {
		err := s.MountGroup(groups...)
		if err != nil {
			return err
		}
	}

	err := s.Shutdown(ctx)
	if err != nil {
		return err
	}

	return s.Run(ctx)
}

// App returns the fiber app of the server.
func (s *server) App() *fiber.App {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.app
}

// Mounted returns all mounted routes, groups, and global middlewares.
func (s *server) Mounted() (routes []Route, groups []RouteGroup, middlewares []fiber.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.routes, s.groups, s.middlewares
}

// OkHandler is a handler that returns an HTTP 200 OK response.
func OkHandler(c fiber.Ctx) error {
	return c.Status(http.StatusOK).SendString("OK")
}

// newContextWithTimeout returns a new context with a timeout.
func newContextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if deadline, ok := ctx.Deadline(); ok {
		if time.Until(deadline) < shutdownTimeout {
			return context.WithDeadline(ctx, deadline)
		}
	}
	return context.WithTimeout(ctx, shutdownTimeout)
}

// isValid checks if the provided method is a valid HTTP method.
func isValid(method string) bool {
	switch method {
	case
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace:
		return true
	default:
		return false
	}
}

// toggleRunning toggles the running state of the server.
func (s *server) toggleRunning() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = !s.running
}
