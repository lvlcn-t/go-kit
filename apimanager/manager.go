package apimanager

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/go-kit/apimanager/middleware"
	"github.com/lvlcn-t/loggerhead/logger"
)

type Server interface {
	// Run runs the server.
	// Runs indefinitely until an error occurs or the server is shut down.
	//
	// If no health check route was mounted before, a health check route will be mounted.
	//
	// Example setup:
	//
	//	srv := api.NewServer(&Config{Address: ":8080"})
	//	err := srv.Mount(RouteGroup{
	//		Path: "/v1",
	//		App: fiber.New().Get("/hello", func(c fiber.Ctx) error {
	//			return c.SendString("Hello, World!")
	//		}),
	//	})
	//	if err != nil {
	//		// handle error
	//	}
	//
	//	_ = srv.Run(context.Background())
	Run(ctx context.Context) error
	// Mount adds the provided routes to the server.
	Mount(routes ...Route) error
	// MountGroup adds the provided route groups to the server.
	MountGroup(groups ...RouteGroup) error
	// Shutdown gracefully shuts down the server.
	Shutdown(ctx context.Context) error
}

// Route is a route to register to the server.
type Route struct {
	// Path is the path of the route.
	Path string
	// Methods is the HTTP method of the route.
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

type Config struct {
	// Address is the address to listen on.
	Address string `yaml:"address" mapstructure:"address"`
	// BasePath is the base path of the API.
	BasePath string `yaml:"basePath" mapstructure:"basePath"`
}

func (c *Config) Validate() error {
	if c.Address == "" {
		return errors.New("api.address is required")
	}
	return nil
}

type server struct {
	mu          sync.Mutex
	config      *Config
	app         *fiber.App
	router      fiber.Router
	routes      []Route
	groups      []RouteGroup
	middlewares []fiber.Handler
	running     bool
}

// New creates a new server with the provided configuration.
func New(c *Config, middlewares ...fiber.Handler) Server {
	if c == nil {
		c = &Config{
			Address:  ":8080",
			BasePath: "/",
		}
	}

	app := fiber.New()
	if len(middlewares) == 0 {
		middlewares = append(middlewares, middleware.Recover(), middleware.Logger("/healthz"))
	}

	return &server{
		mu:          sync.Mutex{},
		config:      c,
		app:         app,
		router:      app.Group(c.BasePath),
		routes:      []Route{},
		groups:      []RouteGroup{},
		middlewares: middlewares,
		running:     false,
	}
}

// Run runs the server.
// It will mount a health check route if no health check route was mounted before.
// Runs indefinitely until an error occurs or the server is shut down.
func (s *server) Run(ctx context.Context) error {
	err := s.attachRoutes(ctx)
	if err != nil {
		return err
	}

	return s.app.Listen(s.config.Address)
}

// Mount adds the provided routes to the server.
func (s *server) Mount(routes ...Route) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return &ErrAlreadyRunning{}
	}

	s.routes = append(s.routes, routes...)
	return nil
}

// MountGroup adds the provided route groups to the server.
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
// It will mount a health check route if no health check route was mounted before.
func (s *server) attachRoutes(ctx context.Context) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return &ErrAlreadyRunning{}
	}

	// Always inject the provided context into the request user context.
	_ = s.router.Use(middleware.Context(ctx))
	defer func() {
		if r := recover(); r != nil {
			if rErr, ok := r.(error); ok {
				err = rErr
				return
			}
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

	for _, route := range s.routes {
		logger.FromContext(ctx).InfoContext(ctx, "Mounting route", "path", route.Path, "methods", strings.Join(route.Methods, ","))
		_ = s.router.Add(route.Methods, route.Path, route.Handler, route.Middlewares...)
	}

	s.running = true
	return nil
}

func (s *server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}
