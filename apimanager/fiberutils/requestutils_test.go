package fiberutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/go-cmp/cmp"
)

const testVal = "value"

type ctxValueGetter[T any] func(fiber.Ctx, string, func(string) (T, error)) (T, error)

type url struct {
	Path  string
	Param *string
	Query *string
}

func Test_ContextValueGetter(t *testing.T) {
	tests := []struct {
		name    string
		fun     ctxValueGetter[string]
		parse   func(string) (string, error)
		url     url
		body    io.Reader
		header  http.Header
		cookies []*http.Cookie
		wantErr bool
	}{
		{
			name: "params - success",
			fun:  Params[string],
			parse: func(string) (string, error) {
				return testVal, nil
			},
			url: url{
				Path:  "/:name",
				Param: toPtr(testVal),
			},
			body:    http.NoBody,
			wantErr: false,
		},
		{
			name: "params - success - no parser with string",
			fun:  Params[string],
			url: url{
				Path:  "/:name",
				Param: toPtr(testVal),
			},
			body:    http.NoBody,
			parse:   nil,
			wantErr: false,
		},
		{
			name: "params - fail - parse error",
			fun:  Params[string],
			parse: func(string) (string, error) {
				return "", fmt.Errorf("error")
			},
			url: url{
				Path:  "/:name",
				Param: toPtr(testVal),
			},
			body:    http.NoBody,
			wantErr: true,
		},
		{
			name: "query - success",
			fun:  Query[string],
			parse: func(string) (string, error) {
				return testVal, nil
			},
			url: url{
				Path:  "/:name",
				Query: toPtr(fmt.Sprintf("name=%s", testVal)),
			},
			body: http.NoBody,
		},
		{
			name:  "query - success - no parser with string",
			fun:   Query[string],
			parse: nil,
			url: url{
				Path:  "/",
				Query: toPtr(fmt.Sprintf("name=%s", testVal)),
			},
			body:    http.NoBody,
			wantErr: false,
		},
		{
			name: "query - fail - parse error",
			fun:  Query[string],
			parse: func(string) (string, error) {
				return "", fmt.Errorf("error")
			},
			url: url{
				Path:  "/",
				Query: toPtr(fmt.Sprintf("name=%s", testVal)),
			},
			body:    http.NoBody,
			wantErr: true,
		},
		{
			name: "cookies - success",
			fun:  Cookies[string],
			parse: func(string) (string, error) {
				return testVal, nil
			},
			url:     url{Path: "/"},
			body:    http.NoBody,
			cookies: []*http.Cookie{{Name: "name", Value: testVal}},
		},
		{
			name:    "cookies - success - no parser with string",
			fun:     Cookies[string],
			parse:   nil,
			url:     url{Path: "/"},
			body:    http.NoBody,
			cookies: []*http.Cookie{{Name: "name", Value: testVal}},
			wantErr: false,
		},
		{
			name: "cookies - fail - parse error",
			fun:  Cookies[string],
			parse: func(string) (string, error) {
				return "", fmt.Errorf("error")
			},
			url:     url{Path: "/"},
			body:    http.NoBody,
			cookies: []*http.Cookie{{Name: "name", Value: testVal}},
			wantErr: true,
		},
		{
			name: "form - success",
			fun:  Form[string],
			parse: func(string) (string, error) {
				return testVal, nil
			},
			url:  url{Path: "/"},
			body: strings.NewReader(fmt.Sprintf("name=%s", testVal)),
		},
		{
			name:    "form - success - no parser with string",
			fun:     Form[string],
			parse:   nil,
			url:     url{Path: "/"},
			body:    strings.NewReader(fmt.Sprintf("name=%s", testVal)),
			wantErr: false,
		},
		{
			name: "form - fail - parse error",
			fun:  Form[string],
			parse: func(string) (string, error) {
				return "", fmt.Errorf("error")
			},
			url:     url{Path: "/"},
			body:    strings.NewReader(fmt.Sprintf("name=%s", testVal)),
			wantErr: true,
		},
		{
			name: "header - success",
			fun:  Header[string],
			parse: func(string) (string, error) {
				return testVal, nil
			},
			url:    url{Path: "/"},
			body:   http.NoBody,
			header: http.Header{"name": []string{testVal}},
		},
		{
			name:    "header - success - no parser with string",
			fun:     Header[string],
			parse:   nil,
			url:     url{Path: "/"},
			body:    http.NoBody,
			header:  http.Header{"name": []string{testVal}},
			wantErr: false,
		},
		{
			name: "header - fail - parse error",
			fun:  Header[string],
			parse: func(string) (string, error) {
				return "", fmt.Errorf("error")
			},
			url:     url{Path: "/"},
			body:    http.NoBody,
			header:  http.Header{"name": []string{testVal}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Post(tt.url.Path, func(c fiber.Ctx) error {
				v, err := tt.fun(c, "name", tt.parse)
				if (err != nil) != tt.wantErr {
					t.Errorf("want error %v, got %v", tt.wantErr, err)
					return c.SendStatus(http.StatusInternalServerError)
				}

				if !tt.wantErr && v != testVal {
					t.Errorf("want %v, got %v", testVal, v)
					return c.SendStatus(http.StatusInternalServerError)
				}
				return c.SendStatus(http.StatusOK)
			})

			req := newTestRequest(t, tt.url, tt.header, tt.body, tt.cookies)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to test app: %v", err)
			}
			defer func() {
				err := resp.Body.Close()
				if err != nil {
					t.Fatalf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("want status code %v, got %v", http.StatusOK, resp.StatusCode)
			}
		})
	}
}

func Test_parseParam(t *testing.T) {
	tests := []struct {
		name    string
		get     func(string, ...string) string
		parse   func(string) (string, error)
		wantErr bool
	}{
		{
			name: "success",
			get: func(string, ...string) string {
				return testVal
			},
			parse: func(string) (string, error) {
				return testVal, nil
			},
			wantErr: false,
		},
		{
			name: "success - no parser with string",
			get: func(string, ...string) string {
				return testVal
			},
			parse:   nil,
			wantErr: false,
		},
		{
			name: "fail - empty",
			get: func(string, ...string) string {
				return ""
			},
			parse: func(string) (string, error) {
				return "", nil
			},
			wantErr: true,
		},
		{
			name: "fail - parse error",
			get: func(string, ...string) string {
				return testVal
			},
			parse: func(string) (string, error) {
				return "", fmt.Errorf("error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := parseParam("name", tt.get, tt.parse)
			if (err != nil) != tt.wantErr {
				t.Errorf("want error %v, got %v", tt.wantErr, err)
			}

			if !tt.wantErr && v != testVal {
				t.Errorf("want %v, got %v", testVal, v)
			}
		})
	}
}

func Test_parseParam_notString(t *testing.T) {
	tests := []struct {
		name    string
		get     func(string, ...string) string
		parse   func(string) (int, error)
		wantErr bool
	}{
		{
			name: "success",
			get: func(string, ...string) string {
				return "1"
			},
			parse: func(string) (int, error) {
				return 1, nil
			},
			wantErr: false,
		},
		{
			name: "fail - no parser",
			get: func(string, ...string) string {
				return "1"
			},
			parse:   nil,
			wantErr: true,
		},
		{
			name: "fail - parse error",
			get: func(string, ...string) string {
				return "1"
			},
			parse: func(string) (int, error) {
				return 0, fmt.Errorf("error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := parseParam("name", tt.get, tt.parse)
			if (err != nil) != tt.wantErr {
				t.Errorf("want error %v, got %v", tt.wantErr, err)
			}

			if !tt.wantErr && v != 1 {
				t.Errorf("want %v, got %v", 1, v)
			}
		})
	}
}

func Test_Body(t *testing.T) {
	tests := []struct {
		name    string
		body    func(t *testing.T) io.Reader
		wantErr bool
	}{
		{
			name: "success",
			body: func(t *testing.T) io.Reader {
				b, err := json.Marshal(map[string]string{"name": testVal})
				if err != nil {
					t.Fatalf("failed to marshal json: %v", err)
				}
				return bytes.NewReader(b)
			},
			wantErr: false,
		},
		{
			name: "fail - invalid json",
			body: func(t *testing.T) io.Reader {
				return bytes.NewReader([]byte("invalid"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/", func(c fiber.Ctx) error {
				v, err := Body[map[string]string](c)
				if (err != nil) != tt.wantErr {
					t.Errorf("want error %v, got %v", tt.wantErr, err)
					return c.SendStatus(http.StatusInternalServerError)
				}

				if !tt.wantErr && v["name"] != testVal {
					t.Errorf("want %v, got %v", testVal, v)
					return c.SendStatus(http.StatusInternalServerError)
				}
				return c.SendStatus(http.StatusOK)
			})

			req := newTestRequest(t, url{Path: "/"}, nil, tt.body(t), nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to test app: %v", err)
			}
			defer func() {
				err := resp.Body.Close()
				if err != nil {
					t.Fatalf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("want status code %v, got %v", http.StatusOK, resp.StatusCode)
			}
		})
	}
}

func TestClient_NewClient(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		port string
	}{
		{
			name: "success",
			ip:   "0.0.0.0",
			port: "8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/", func(c fiber.Ctx) error {
				client := NewClient(c)
				if client.ip != c.IP() {
					t.Errorf("want %v, got %v", c.IP(), client.ip)
					return c.SendStatus(http.StatusInternalServerError)
				}

				if client.port != c.Port() {
					t.Errorf("want %v, got %v", c.Port(), client.port)
					return c.SendStatus(http.StatusInternalServerError)
				}
				return c.SendStatus(http.StatusOK)
			})

			req := newTestRequest(t, url{Path: "/"}, nil, http.NoBody, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to test app: %v", err)
			}
			defer func() {
				err := resp.Body.Close()
				if err != nil {
					t.Fatalf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("want status code %v, got %v", http.StatusOK, resp.StatusCode)
			}
		})
	}
}

func TestClient_String(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		port string
	}{
		{
			name: "success",
			ip:   "0.0.0.0",
			port: "8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				ip:   tt.ip,
				port: tt.port,
			}

			if got := client.String(); got != fmt.Sprintf("%s:%s", tt.ip, tt.port) {
				t.Errorf("want %v, got %v", tt.ip+":"+tt.port, got)
			}
		})
	}
}

func TestClient_Builders(t *testing.T) {
	tests := []struct {
		name     string
		ip       *string
		wantIP   net.IP
		port     *string
		wantPort uint16
	}{
		{
			name:   "success with ip",
			ip:     toPtr("0.0.0.0"),
			wantIP: net.ParseIP("0.0.0.0"),
		},
		{
			name:     "success with port",
			port:     toPtr("8080"),
			wantPort: 8080,
		},
		{
			name:     "success with both",
			ip:       toPtr("0.0.0.0"),
			wantIP:   net.ParseIP("0.0.0.0"),
			port:     toPtr("8080"),
			wantPort: 8080,
		},
		{
			name:     "fail - port is not int",
			port:     toPtr("notint"),
			wantPort: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{}
			if tt.ip != nil {
				client.ip = *tt.ip
				got := client.IP()
				if !cmp.Equal(got.Get(), tt.wantIP) {
					t.Errorf("want %v, got %v", IP(tt.wantIP), got)
				}

				if got.String() != tt.wantIP.String() {
					t.Errorf("want %v, got %v", tt.wantIP.String(), got.String())
				}
				return
			}

			if tt.port != nil {
				client.port = *tt.port
				got := client.Port()
				if got.Get() != tt.wantPort {
					t.Errorf("want %v, got %v", tt.wantPort, got)
				}

				if got.String() != strconv.Itoa(int(tt.wantPort)) {
					t.Errorf("want %v, got %v", strconv.Itoa(int(tt.wantPort)), got.String())
				}
				return
			}
		})
	}
}

// newTestRequest creates a new test request with the given parameters.
func newTestRequest(t *testing.T, url url, header http.Header, body io.Reader, cookies []*http.Cookie) *http.Request {
	t.Helper()

	if url.Param != nil {
		url.Path = strings.Replace(url.Path, ":name", *url.Param, 1)
	}

	if url.Query != nil {
		url.Path = fmt.Sprintf("%s?%s", url.Path, *url.Query)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url.Path, body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if header != nil {
		req.Header = header
	}

	if body != nil && body != http.NoBody {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	for _, c := range cookies {
		req.AddCookie(c)
	}

	t.Logf("Request: url=%s, body=%v, header=%v, cookies=%v", req.URL, body, header, cookies)
	return req
}

// toPtr is a helper function to convert a value to a pointer.
func toPtr[T any](s T) *T {
	return &s
}
