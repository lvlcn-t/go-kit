package apimanager_test

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/go-kit/apimanager"
)

func Example() {
	server := apimanager.New(nil)
	err := server.Mount(apimanager.Route{
		Path:    "/",
		Methods: []string{fiber.MethodGet},
		Handler: func(c fiber.Ctx) error {
			return c.SendString("Hello, World!")
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := server.Run(context.Background()); err != nil {
		panic(err)
	}
}
