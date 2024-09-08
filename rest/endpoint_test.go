package rest

import (
	"net/http"
	"net/url"
	"testing"
)

func TestEndpoint_Compile(t *testing.T) {
	tests := []struct {
		name    string
		e       *Endpoint
		baseURL string
		want    string
		wantErr bool
	}{
		{
			name:    "empty path",
			e:       Get(""),
			baseURL: "http://localhost",
			want:    "http://localhost",
			wantErr: false,
		},
		{
			name:    "empty base URL",
			e:       Post("/path"),
			baseURL: "",
			want:    "/path",
			wantErr: false,
		},
		{
			name:    "invalid base URL",
			e:       Put("/path"),
			baseURL: "://?http://localhost",
			want:    "",
			wantErr: true,
		},
		{
			name:    "valid path",
			e:       Patch("/path"),
			baseURL: "http://localhost",
			want:    "http://localhost/path",
			wantErr: false,
		},
		{
			name:    "valid path with query",
			e:       Delete("/path").AddQuery("key", "value"),
			baseURL: "http://localhost",
			want:    "http://localhost/path?key=value",
			wantErr: false,
		},
		{
			name:    "valid path with multiple queries",
			e:       Get("/path", url.Values{"key": {"value"}}, url.Values{"key2": {"value2"}}),
			baseURL: "http://localhost",
			want:    "http://localhost/path?key=value&key2=value2",
			wantErr: false,
		},
		{
			name:    "valid path with invalid query",
			e:       Post("/path", url.Values{"key": {"value"}}, url.Values{"key2": {}}),
			baseURL: "http://localhost",
			want:    "http://localhost/path?key=value",
			wantErr: false,
		},
		{
			name:    "valid path as URL",
			e:       Put("http://localhost/path"),
			baseURL: "http://localhost",
			want:    "http://localhost/path",
			wantErr: false,
		},
		{
			name:    "valid path as URL without base URL",
			e:       Patch("http://localhost/path"),
			baseURL: "",
			want:    "http://localhost/path",
			wantErr: false,
		},
		{
			name:    "valid path as URL with another base URL",
			e:       Delete("http://localhost/path"),
			baseURL: "https://example.com",
			want:    "http://localhost/path",
			wantErr: false,
		},
		{
			name:    "invalid path as invalid URL",
			e:       Get("://?http://localhost/path"),
			baseURL: "http://localhost",
			want:    "",
			wantErr: true,
		},
		{
			name:    "valid path as URL substituting base URL",
			e:       Post("http://localhost/path"),
			baseURL: "http://localhost",
			want:    "http://localhost/path",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.Compile(tt.baseURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Endpoint.Compile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Endpoint.Compile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewEndpoint(t *testing.T) {
	tests := []struct {
		name    string
		fun     func(path string, queries ...url.Values) *Endpoint
		path    string
		queries []url.Values
		want    *Endpoint
	}{
		{
			name:    "empty path",
			fun:     Get,
			path:    "",
			queries: nil,
			want:    &Endpoint{Method: http.MethodGet, Path: "", Query: nil},
		},
		{
			name:    "valid path",
			fun:     Post,
			path:    "/path",
			queries: nil,
			want:    &Endpoint{Method: http.MethodPost, Path: "/path", Query: nil},
		},
		{
			name:    "valid path with query",
			fun:     Put,
			path:    "/path",
			queries: []url.Values{{"key": {"value"}}},
			want:    &Endpoint{Method: http.MethodPut, Path: "/path", Query: url.Values{"key": {"value"}}},
		},
		{
			name:    "valid path with multiple queries",
			fun:     Patch,
			path:    "/path",
			queries: []url.Values{{"key": {"value"}}, {"key2": {"value2"}}},
			want:    &Endpoint{Method: http.MethodPatch, Path: "/path", Query: url.Values{"key": {"value"}, "key2": {"value2"}}},
		},
		{
			name:    "valid path with invalid query",
			fun:     Delete,
			path:    "/path",
			queries: []url.Values{{"key": {"value"}}, {"key2": {}}},
			want:    &Endpoint{Method: http.MethodDelete, Path: "/path", Query: url.Values{"key": {"value"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fun(tt.path, tt.queries...)
			if got.Method != tt.want.Method {
				t.Errorf("Get() Method = %v, want %v", got.Method, tt.want.Method)
			}
			if got.Path != tt.want.Path {
				t.Errorf("Get() Path = %v, want %v", got.Path, tt.want.Path)
			}
			if got.Query.Encode() != tt.want.Query.Encode() {
				t.Errorf("Get() Query = %v, want %v", got.Query, tt.want.Query)
			}
		})
	}
}
