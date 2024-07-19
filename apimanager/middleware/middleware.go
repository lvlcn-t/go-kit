// middleware package provides common middleware as [fiber.Handler].
package middleware

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/go-kit/apimanager/fiberutils"
	"github.com/lvlcn-t/loggerhead/logger"
)

// Context sets the user context on the request.
// The user context then contains the logger and open telemetry span.
func Context(ctx context.Context) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.SetUserContext(ctx)
		return c.Next()
	}
}

// Logger logs the request if the path is not ignored.
func Logger(ignore ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		log := logger.FromContext(c.UserContext()).With("method", c.Method(), "path", c.Path())
		c.SetUserContext(logger.IntoContext(c.UserContext(), log))
		if !slices.Contains(ignore, c.Path()) {
			log.InfoContext(c.Context(), "Request received", "ip", c.IP(), "method", c.Method(), "path", c.Path())
		}
		return c.Next()
	}
}

// Recover recovers from panics and logs the error.
func Recover() fiber.Handler {
	return func(c fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log := logger.FromContext(c.UserContext())
				log.ErrorContext(c.Context(), "Panic recovered", "error", r)
				err = errors.Join(err, fiberutils.InternalServerErrorResponse(c, fmt.Sprintf("panic: %v", r)))
			}
		}()
		return c.Next()
	}
}
