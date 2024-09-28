package config

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

type config struct {
	Host string
	Port int
}

func (c config) IsEmpty() bool {
	return reflect.DeepEqual(c, config{})
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		fallbacks []Fallback
		want      Loadable
		wantErr   bool
		errType   reflect.Type
	}{
		{
			name: "success",
			path: "testdata/config.yaml",
			want: config{
				Host: "localhost",
				Port: 8080,
			},
			wantErr: false,
		},
		{
			name: "fallback",
			path: "",
			fallbacks: []Fallback{func() (string, error) {
				return "testdata/config.yaml", nil
			}},
			want:    config{Host: "localhost", Port: 8080},
			wantErr: false,
		},
		{
			name: "fallback error",
			path: "",
			fallbacks: []Fallback{func() (string, error) {
				return "", errors.New("error")
			}},
			wantErr: true,
			errType: reflect.TypeOf(fmt.Errorf("%w", errors.New("error"))),
		},
		{
			name:    "empty path",
			path:    "",
			want:    config{Host: "localhost", Port: 8080},
			wantErr: false,
		},
		{
			name:    "empty config",
			path:    "testdata/empty.yaml",
			want:    config{},
			wantErr: true,
		},
		{
			name:    "invalid config format",
			path:    "testdata/invalid.yaml",
			wantErr: true,
			errType: reflect.TypeOf(fmt.Errorf("%w", errors.New("yaml: unmarshal errors"))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t, tt.path, tt.want, tt.fallbacks...)

			got, err := Load[config](tt.path, tt.fallbacks...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil && reflect.TypeOf(err) != tt.errType {
				t.Errorf("Load() error type = %v, want %v", reflect.TypeOf(err), tt.errType)
			}

			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() = %v, want %v", got, tt.want)
			}
		})
	}
}

type invalid int

func (i invalid) IsEmpty() bool {
	return false
}

func TestLoad_InvalidType(t *testing.T) {
	_, err := Load[invalid]("testdata/config.yaml")
	if err == nil {
		t.Error("Load() error = nil, want error")
	}
}

func TestLoad_Pointer(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    Loadable
		wantErr bool
	}{
		{
			name: "pointer",
			path: "testdata/config.yaml",
			want: &config{
				Host: "localhost",
				Port: 8080,
			},
			wantErr: false,
		},
		{
			name:    "nil pointer",
			path:    "testdata/config.yaml",
			want:    (*config)(nil),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t, tt.path, tt.want)
			if reflect.ValueOf(tt.want).IsNil() {
				tt.want = &config{}
			}

			got, err := Load[*config](tt.path)
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

func TestSetName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want string
	}{
		{
			name: "success",
			want: "test",
		},
		{
			name: "empty",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetName(tt.want)
			if appName != tt.want {
				t.Errorf("SetName() = %v, want %v", appName, tt.want)
			}
		})
	}
}

func TestSetFs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want afero.Fs
	}{
		{
			name: "success",
			want: afero.NewMemMapFs(),
		},
		{
			name: "nil",
			want: nil,
		},
		{
			name: "default",
			want: afero.NewOsFs(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetFs(tt.want)
			if fsys != tt.want {
				t.Errorf("SetFs() = %v, want %v", fsys, tt.want)
			}
		})
	}
}

func setup(t *testing.T, path string, cfg Loadable, fallbacks ...Fallback) {
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

	var err error
	data := append([]byte{}, []byte("in: va: lid\n")...)
	if cfg != nil {
		data, err = yaml.Marshal(cfg)
		if err != nil {
			t.Fatalf("yaml.Marshal() error = %v", err)
		}
	}

	if err := afero.WriteFile(fsys, path, data, 0o644); err != nil {
		t.Fatalf("afero.WriteFile() error = %v", err)
	}
}
