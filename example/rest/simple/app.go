package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/go-kit/apimanager"
	"github.com/lvlcn-t/go-kit/rest"
)

// Define the response type
type response struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	ctx := context.Background()
	// Create a new API server
	srv := apimanager.New(&apimanager.Config{
		Address: ":8080",
	})

	// Mount a new route
	err := srv.Mount(apimanager.Route{
		Methods: []string{http.MethodGet},
		Path:    "/",
		Handler: func(c fiber.Ctx) error {
			return c.Status(http.StatusOK).JSON(response{
				ID:   1,
				Name: "Jane Doe",
			})
		},
	})
	if err != nil {
		panic(err)
	}

	// Run the server in a goroutine
	go func() {
		if err = srv.Run(ctx); err != nil {
			panic(err)
		}
	}()

	// Wait for the server to start
	time.Sleep(time.Second)

	// Make an API call to the server
	callAPI(ctx)

	// Shutdown the server
	err = srv.Shutdown(ctx)
	if err != nil {
		panic(err)
	}
}

// callAPI makes an API call to the server
func callAPI(ctx context.Context) {
	// Create a new REST endpoint
	endpoint := rest.Get("http://localhost:8080")

	// Make a request to the server
	resp, status, err := rest.Do[response](ctx, endpoint, nil)
	if err != nil {
		panic(err)
	}
	// Close the rest client gracefully (optional)
	defer rest.Close(ctx)

	// Check the status code
	if status != http.StatusOK {
		panic(fmt.Errorf("unexpected status code: %d", status))
	}

	// Print the response
	fmt.Printf("\nResponse from the server: %+v\n", resp)
}
