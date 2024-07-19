// Package config provides a way to load and validate configurations.
//
// You can load configurations from a file and environment variables.
// You can validate configurations either by implementing the [Validator] interface or using the "validate" struct tag.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

var (
	// fsys is the filesystem used to load the configuration
	fsys afero.Fs = afero.NewOsFs()
	// appName is the name of the application (default: binary name)
	appName string = filepath.Base(os.Args[0])
)

// Loadable is an interface that must be implemented by a configuration struct to be loaded.
type Loadable interface {
	// IsEmpty returns true if the configuration is empty
	IsEmpty() bool
}

// Fallback is a function that returns the path to the configuration file if an empty path is provided
type Fallback func() (string, error)

// Load loads the configuration from the provided path or fallback path.
// Returns an error if the configuration cannot be loaded or unmarshalled into the provided struct.
//
// You can provide a slice of fallback functions that will be used to get the configuration path if an empty path is provided.
// The first fallback function that returns a path is used.
// If no fallback functions are provided, the default fallback is used (~/.config/<appName>/config.yaml).
//
// All environment variables with the scheme "<appName>_<field-name>(_<recursive-field-name>)" will be considered.
//
// The configuration is unmarshalled into the provided struct.
// Its IsEmpty method is called to check if the loaded configuration is empty.
//
// Example:
//
//	type Config struct {
//		Host string
//		Port int
//	}
//
//	func (c Config) IsEmpty() bool {
//		return c == (Config{})
//	}
//
//	cfg, err := config.Load[Config]("config.yaml")
//	if err != nil {
//		// Handle error
//	}
func Load[T Loadable](path string, fallbacks ...Fallback) (cfg T, err error) {
	cfg, err = ensureStruct(cfg)
	if err != nil {
		return cfg, fmt.Errorf("given type is not a struct: %w", err)
	}

	v := viper.NewWithOptions(viper.ExperimentalBindStruct())
	v.SetFs(fsys)
	if path == "" {
		if len(fallbacks) == 0 {
			fallbacks = append(fallbacks, defaultFallback)
		}

		for i, f := range fallbacks {
			if path, err = f(); err == nil {
				break
			}
			if i == len(fallbacks)-1 {
				return cfg, fmt.Errorf("failed to get fallback path: %w", err)
			}
		}
	}
	v.SetConfigFile(path)

	v.SetEnvPrefix(strings.ToUpper(strings.ReplaceAll(appName, "-", "_")))
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(*fs.PathError); !ok {
			return cfg, fmt.Errorf("failed to read configuration file: %w", err)
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if cfg.IsEmpty() {
		return cfg, &ErrConfigEmpty{}
	}

	return cfg, nil
}

// SetName replaces the default application name with the provided one.
// This function is not safe for concurrent use.
func SetName(name string) {
	appName = name
}

// SetFs replaces the default filesystem with the provided one.
// This may only be used for testing purposes.
// This function is not safe for concurrent use.
func SetFs(filesystem afero.Fs) {
	fsys = filesystem
}

// defaultFallback returns the default fallback path for the configuration file.
func defaultFallback() (string, error) {
	home, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, appName, "config.yaml"), nil
}

// ensureStruct ensures that the provided value is a struct or a pointer to a struct.
func ensureStruct[T any](value T) (T, error) {
	var empty T
	t := reflect.TypeOf(value)

	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return empty, errors.New("value must be a struct or a pointer to a struct")
	}

	if reflect.TypeOf(value).Kind() == reflect.Pointer && reflect.ValueOf(value).IsNil() {
		return reflect.New(t).Interface().(T), nil
	}

	return value, nil
}
