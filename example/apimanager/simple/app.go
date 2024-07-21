package main

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/go-kit/apimanager"
	"github.com/lvlcn-t/loggerhead/logger"
)

func main() {
	ctx, cancel := logger.NewContextWithLogger(context.Background())
	defer cancel()
	log := logger.FromContext(ctx)

	server := apimanager.New(nil)
	err := server.Mount(apimanager.Route{
		Path:    "/",
		Methods: []string{http.MethodGet},
		Handler: func(c fiber.Ctx) error {
			return c.Status(http.StatusOK).JSON(fiber.Map{
				"message": c.Locals("middleware"),
			})
		},
		Middlewares: []fiber.Handler{
			func(c fiber.Ctx) error {
				_ = c.Locals("middleware", "Hello, World!")
				return c.Next()
			},
		},
	})
	if err != nil {
		log.FatalContext(ctx, "Failed to mount route", err)
	}

	if err = server.Run(ctx); err != nil {
		log.FatalContext(ctx, "Failed to run server", err)
	}
}
