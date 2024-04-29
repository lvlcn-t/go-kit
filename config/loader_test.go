package config

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

type Config struct {
	Host string
	Port int
}

func (c Config) IsEmpty() bool {
	return c == (Config{})
}

type Invalid int

func (i Invalid) IsEmpty() bool {
	return false
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		fallbacks []Fallback
		want      Settings
		wantErr   bool
	}{
		{
			name: "success",
			path: "testdata/config.yaml",
			want: Config{
				Host: "localhost",
				Port: 8080,
			},
		},
		{
			name: "fallback",
			fallbacks: []Fallback{
				func() (string, error) {
					return "testdata/config.yaml", nil
				},
			},
			want: Config{
				Host: "localhost",
				Port: 8080,
			},
		},
		{
			name: "fallback error",
			fallbacks: []Fallback{
				func() (string, error) {
					return "", errors.New("error")
				},
			},
			wantErr: true,
		},
		{
			name: "empty path",
			want: Config{
				Host: "localhost",
				Port: 8080,
			},
		},
		{
			name:    "empty config",
			path:    "testdata/empty.yaml",
			wantErr: true,
		},
		{
			name:    "invalid type",
			path:    "testdata/config.yaml",
			want:    Invalid(0),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t, tt.path, tt.want, tt.fallbacks...)

			got, err := Load[Config](tt.path, tt.fallbacks...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() = %v, want %v", got, tt.want)
			}
		})
	}
}

func setup(t *testing.T, path string, cfg Settings, fallbacks ...Fallback) {
	t.Helper()
	if path == "" {
		var err error
		if len(fallbacks) == 0 {
			path, err = defaultFallback()
			if err != nil {
				t.Fatalf("defaultFallback() error = %v", err)
			}
		} else {
			for _, f := range fallbacks {
				path, err = f()
				if err == nil {
					break
				}
			}
		}
	}

	fsys = afero.NewMemMapFs()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("yaml.Marshal() error = %v", err)
	}

	if err := afero.WriteFile(fsys, path, data, 0o644); err != nil {
		t.Fatalf("afero.WriteFile() error = %v", err)
	}
}
