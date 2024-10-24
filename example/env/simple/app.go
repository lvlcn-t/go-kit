package main

import (
	"fmt"
	"time"

	"github.com/lvlcn-t/go-kit/env"
)

//nolint:mnd // This is an example.
func main() {
	// Read a couple of environment variables with fallback values.
	fmt.Println(env.GetWithFallback("MY_STRING", "default"))
	fmt.Println(env.GetWithFallback("MY_INT", 42))
	fmt.Println(env.GetWithFallback("MY_DURATION", 5*time.Second, time.ParseDuration))
	fmt.Println(env.GetWithFallback("MY_BOOL", false))

	// Read a couple of environment variables that are necessary and panic if they are not set.
	fmt.Println(env.MustGet[string]("MY_STRING"))
	fmt.Println(env.MustGet[int]("MY_INT"))
	fmt.Println(env.MustGet("MY_DURATION", time.ParseDuration))

	// Read a couple of environment variables that are necessary but you want to handle the errors gracefully.
	value, err := env.Get[bool]("MY_BOOL").NoFallback().Value()
	if err != nil {
		fmt.Println("Failed to get MY_BOOL:", err)
	}
	fmt.Println(value)
}
