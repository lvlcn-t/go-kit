package rest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

type response struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		wantErr bool
	}{
		{
			name:    "valid base URL",
			baseURL: "https://example.com",
		},
		{
			name:    "invalid base URL",
			baseURL: "://:?https://example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.baseURL, 5*time.Second)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_Do(t *testing.T) {
	tests := []struct {
		name        string
		endpoint    *Endpoint
		payload     any
		options     []RequestOption
		want        any
		wantErr     bool
		wantCode    int
		wantHeaders map[string]string
		delay       time.Duration
		cancelCtx   bool
		invalidURL  bool
	}{
		{
			name: "success case with GET method",
			endpoint: &Endpoint{
				Method: http.MethodGet,
				Path:   "/resource",
			},
			want: response{
				ID:   1,
				Name: "Test Resource",
			},
			wantCode: http.StatusOK,
			wantErr:  false,
		},
		{
			name: "POST method with payload",
			endpoint: &Endpoint{
				Method: http.MethodPost,
				Path:   "/resource",
			},
			payload: map[string]string{
				"name": "New Resource",
			},
			want: response{
				ID:   2,
				Name: "New Resource",
			},
			wantCode: http.StatusCreated,
			wantErr:  false,
		},
		{
			name: "failure case with 404 error",
			endpoint: &Endpoint{
				Method: http.MethodGet,
				Path:   "/invalid-resource",
			},
			wantCode: http.StatusNotFound,
			want:     response{},
			wantErr:  false,
		},
		{
			name: "request with custom headers",
			endpoint: &Endpoint{
				Method: http.MethodGet,
				Path:   "/resource",
			},
			wantCode: http.StatusOK,
			payload:  nil,
			options: []RequestOption{
				WithHeader("X-Custom-Header", "CustomValue"),
			},
			want: response{
				ID:   3,
				Name: "Resource with Header",
			},
			wantErr: false,
			wantHeaders: map[string]string{
				"X-Custom-Header": "CustomValue",
			},
		},
		{
			name: "request with query parameters",
			endpoint: &Endpoint{
				Method: http.MethodGet,
				Path:   "/resource",
				Query:  newQuery("id", "1"),
			},
			wantCode: http.StatusOK,
			want: response{
				ID:   1,
				Name: "Resource with Query",
			},
			wantErr: false,
		},
		{
			name: "invalid endpoint URL compilation",
			endpoint: &Endpoint{
				Method: http.MethodGet,
				Path:   "://invalid-url",
			},
			wantCode: 0,
			wantErr:  true,
		},
		{
			name: "request with delay",
			endpoint: &Endpoint{
				Method: http.MethodGet,
				Path:   "/resource",
			},
			want: response{
				ID:   4,
				Name: "Delayed Resource",
			},
			wantCode: http.StatusOK,
			wantErr:  false,
			delay:    2 * time.Second,
		},
		{
			name: "context cancellation",
			endpoint: &Endpoint{
				Method: http.MethodGet,
				Path:   "/resource",
			},
			wantErr:   true,
			delay:     5 * time.Second,
			cancelCtx: true,
		},
		{
			name: "invalid payload",
			endpoint: &Endpoint{
				Method: http.MethodPost,
				Path:   "/resource",
			},
			payload:  make(chan int),
			wantCode: 0,
			wantErr:  true,
		},
		{
			name: "invalid JSON response",
			endpoint: &Endpoint{
				Method: http.MethodGet,
				Path:   "/resource",
			},
			want:     1,
			wantCode: http.StatusOK,
			wantErr:  true,
		},
		{
			name:       "nil endpoint",
			endpoint:   nil,
			wantErr:    true,
			invalidURL: true,
		},
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	c := &client{
		baseURL:     "https://example.com",
		client:      http.DefaultClient,
		rateLimiter: defaultRateLimiter,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if tt.cancelCtx {
				go func() {
					time.Sleep(1 * time.Second)
					cancel()
				}()
			}

			b, err := json.Marshal(tt.want)
			if err != nil {
				t.Fatalf("Failed to marshal expected response: %v", err)
			}

			if !tt.invalidURL {
				url, err := tt.endpoint.Compile(c.baseURL)
				if err != nil && !tt.wantErr {
					t.Fatalf("Failed to compile endpoint: %v", err)
				}

				httpmock.RegisterResponder(tt.endpoint.Method, url, func(req *http.Request) (*http.Response, error) {
					if tt.wantHeaders != nil {
						for key, value := range tt.wantHeaders {
							if req.Header.Get(key) != value {
								t.Errorf("Expected header %s to have value %s, got %s", key, value, req.Header.Get(key))
							}
						}
					}
					return httpmock.NewBytesResponse(tt.wantCode, b), nil
				})
			}

			if tt.delay > 0 {
				tt.options = append(tt.options, WithDelay(tt.delay))
			}

			var got response
			code, err := c.Do(ctx, tt.endpoint, tt.payload, &got, tt.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Do() error = %v, wantErr %v", err, tt.wantErr)
			}

			if code != tt.wantCode && tt.wantCode != 0 {
				t.Errorf("Client.Do() code = %v, want %v", code, tt.wantCode)
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("Client.Do() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func newQuery(key, value string) url.Values {
	query := url.Values{}
	query.Add(key, value)
	return query
}

type mockTransport struct {
	httpmock.MockTransport
	numClosed int
}

func (t *mockTransport) CloseIdleConnections() {
	t.numClosed += 1
}

func (t *mockTransport) NumClosed() int {
	return t.numClosed
}

func TestClient_Close(t *testing.T) {
	tests := []struct {
		name        string
		numRequests int
		reqDelay    time.Duration
		cancelCtx   bool
		wantTimeout bool
	}{
		{
			name:        "Close without pending requests",
			numRequests: 0,
			reqDelay:    0,
			cancelCtx:   false,
			wantTimeout: false,
		},
		{
			name:        "Close with pending requests (no context cancel)",
			numRequests: 2,
			reqDelay:    500 * time.Millisecond,
			cancelCtx:   false,
			wantTimeout: false,
		},
		{
			name:        "Close with pending requests and context cancellation",
			numRequests: 2,
			reqDelay:    500 * time.Millisecond,
			cancelCtx:   true,
			wantTimeout: true,
		},
	}
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				client: &http.Client{
					Transport: &mockTransport{},
				},
				wg: sync.WaitGroup{},
			}

			httpmock.RegisterResponder(http.MethodGet, "https://example.com/resource", httpmock.NewJsonResponderOrPanic(200, map[string]any{"id": 1, "name": "Resource"}))

			for range tt.numRequests {
				c.wg.Add(1)
				go func() {
					defer c.wg.Done()

					time.Sleep(tt.reqDelay)
					req, _ := http.NewRequest(http.MethodGet, "https://example.com/resource", http.NoBody)
					_, _ = c.client.Do(req)
				}()
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			if tt.cancelCtx {
				cancel()
			}
			defer cancel()

			start := time.Now()
			c.Close(ctx)
			elapsed := time.Since(start)

			if !tt.wantTimeout && elapsed < tt.reqDelay {
				t.Errorf("Close() returned too early, expected it to wait for pending requests")
			}

			if transport, ok := c.client.Transport.(*mockTransport); ok {
				if transport.NumClosed() == 0 {
					t.Errorf("Close() did not call CloseIdleConnections on the transport")
				}
			}
		})
	}
}

func TestClient_Client(t *testing.T) {
	c := &client{
		client: http.DefaultClient,
	}

	if c.Client() != http.DefaultClient {
		t.Errorf("Client() = %v, want %v", c.Client(), http.DefaultClient)
	}
}

func TestClient_RateLimiter(t *testing.T) {
	c := &client{
		rateLimiter: defaultRateLimiter,
	}

	if c.RateLimiter() != defaultRateLimiter {
		t.Errorf("RateLimiter() = %v, want %v", c.RateLimiter(), defaultRateLimiter)
	}
}

func TestWithDelay(t *testing.T) {
	delay := 5 * time.Second
	request := &Request{
		Request: &http.Request{},
	}
	WithDelay(delay)(request)

	if request.Delay != delay {
		t.Errorf("WithDelay() = %v, want %v", request.Delay, delay)
	}
}

func TestWithHeader(t *testing.T) {
	key, value := "X-Custom-Header", "CustomValue"
	request := &Request{
		Request: &http.Request{
			Header: http.Header{},
		},
	}
	WithHeader(key, value)(request)

	if request.Request.Header.Get(key) != value {
		t.Errorf("WithHeader() = %v, want %v", request.Request.Header.Get(key), value)
	}
}

func TestWithBasicAuth(t *testing.T) {
	username, password := "user", "password"
	request := &Request{
		Request: &http.Request{
			Header: http.Header{},
		},
	}
	WithBasicAuth(username, password)(request)

	auth := fmt.Sprintf("%s:%s", username, password)
	auth = base64.StdEncoding.EncodeToString([]byte(auth))
	if request.Request.Header.Get("Authorization") != fmt.Sprintf("Basic %s", auth) {
		t.Errorf("WithBasicAuth() = %v, want %v", request.Request.Header.Get("Authorization"), fmt.Sprintf("%s:%s", username, password))
	}
}

func TestWithBearer(t *testing.T) {
	token := "token"
	request := &Request{
		Request: &http.Request{
			Header: http.Header{},
		},
	}
	WithBearer(token)(request)

	if request.Request.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", token) {
		t.Errorf("WithBearer() = %v, want %v", request.Request.Header.Get("Authorization"), fmt.Sprintf("Bearer %s", token))
	}
}
