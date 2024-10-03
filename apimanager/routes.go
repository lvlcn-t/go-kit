package apimanager

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
)

// MethodUse is a custom HTTP method to indicate that a route should be registered for all HTTP methods.
const MethodUse = "USE"

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

// Get creates a new [Route] with the provided path, handler, and middlewares for the [http.MethodGet] method.
func Get(path string, handler fiber.Handler, middlewares ...fiber.Handler) Route {
	return Route{
		Path:        path,
		Methods:     []string{http.MethodGet},
		Handler:     handler,
		Middlewares: middlewares,
	}
}

// Post creates a new [Route] with the provided path, handler, and middlewares for the [http.MethodPost] method.
func Post(path string, handler fiber.Handler, middlewares ...fiber.Handler) Route {
	return Route{
		Path:        path,
		Methods:     []string{http.MethodPost},
		Handler:     handler,
		Middlewares: middlewares,
	}
}

// Put creates a new [Route] with the provided path, handler, and middlewares for the [http.MethodPut] method.
func Put(path string, handler fiber.Handler, middlewares ...fiber.Handler) Route {
	return Route{
		Path:        path,
		Methods:     []string{http.MethodPut},
		Handler:     handler,
		Middlewares: middlewares,
	}
}

// Patch creates a new [Route] with the provided path, handler, and middlewares for the [http.MethodPatch] method.
func Patch(path string, handler fiber.Handler, middlewares ...fiber.Handler) Route {
	return Route{
		Path:        path,
		Methods:     []string{http.MethodPatch},
		Handler:     handler,
		Middlewares: middlewares,
	}
}

// Delete creates a new [Route] with the provided path, handler, and middlewares for the [http.MethodDelete] method.
func Delete(path string, handler fiber.Handler, middlewares ...fiber.Handler) Route {
	return Route{
		Path:        path,
		Methods:     []string{http.MethodDelete},
		Handler:     handler,
		Middlewares: middlewares,
	}
}

// Use creates a new [Route] with the provided path, handler, and middlewares for all HTTP methods.
func Use(path string, handler fiber.Handler, middlewares ...fiber.Handler) Route {
	return Route{
		Path:        path,
		Methods:     []string{MethodUse},
		Handler:     handler,
		Middlewares: middlewares,
	}
}

// RouteGroup is a route to register a sub-app to.
type RouteGroup struct {
	// Path is the Path of the route.
	Path string
	// App is the fiber sub-App to use.
	App fiber.Router
}

// NewRouteGroup creates a new [RouteGroup] with the provided path and sub-app.
func NewRouteGroup(path string, app fiber.Router) RouteGroup {
	return RouteGroup{Path: path, App: app}
}

// isMethodValid checks if the provided method is a valid HTTP method.
func isMethodValid(method string) bool {
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
