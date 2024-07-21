package main

import (
	"fmt"

	"github.com/lvlcn-t/go-kit/config"
)

// Config represents the configuration of the application.
// It implements the [config.Loadable] interface.
type Config struct {
	Host string `yaml:"host" mapstructure:"host" validate:"required"`
	Port int    `yaml:"port" mapstructure:"port" validate:"required,min=1024,max=65535"`
}

// IsEmpty returns true if the configuration is empty.
func (c *Config) IsEmpty() bool {
	return c == (&Config{})
}

func main() {
	// Load the configuration from the file
	cfg, err := config.Load[*Config]("./config.yaml")
	if err != nil {
		panic(err)
	}

	// Validate the configuration using the struct tags
	err = config.Validate(cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Host)
	fmt.Println(cfg.Port)
}
