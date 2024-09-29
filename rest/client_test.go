package rest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"golang.org/x/time/rate"
)

type response struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestDefaultClient_Do(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	DefaultClient.(*restClient).client = http.DefaultClient

	httpmock.RegisterResponder(http.MethodGet, "https://example.com/resource", httpmock.NewJsonResponderOrPanic(200, map[string]any{"id": 1, "name": "Resource"}))

	type response struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	endpoint := Get("https://example.com/resource")

	resp, status, err := Do[response](context.Background(), endpoint, nil)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	if status != http.StatusOK {
		t.Errorf("Do() status = %v, want %v", status, http.StatusOK)
	}

	if resp.ID != 1 || resp.Name != "Resource" {
		t.Errorf("Do() resp = %v, want %v", resp, response{ID: 1, Name: "Resource"})
	}
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
			_, err := New(tt.baseURL, 5*time.Second)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_Do(t *testing.T) { //nolint:gocyclo // Either complexity or duplication
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
			name: "POST method with error response",
			endpoint: &Endpoint{
				Method: http.MethodPost,
				Path:   "/resource",
			},
			payload: map[string]string{
				"name": "New Resource",
			},
			options: []RequestOption{
				WithResponseHandler(func(r *http.Response) error {
					if r.StatusCode != http.StatusConflict {
						return errors.New("unexpected status code")
					}

					var er struct {
						Error string `json:"error"`
					}
					if err := json.NewDecoder(r.Body).Decode(&er); err != nil {
						return err
					}

					if er.Error == "" {
						return errors.New("error field not found in response")
					}
					return nil
				}),
			},
			want:     map[string]string{"error": "Resource already exists"},
			wantCode: http.StatusConflict,
			wantErr:  false,
		},
		{
			name: "PUT method with error response and custom error handler",
			endpoint: &Endpoint{
				Method: http.MethodPut,
				Path:   "/resource",
			},
			payload: map[string]string{
				"name": "Updated Resource",
			},
			options: []RequestOption{
				WithErrorHandler(func(resp *http.Response) error {
					if resp.StatusCode < http.StatusBadRequest {
						return errors.New("unexpected status code")
					}

					var er struct {
						Error string `json:"error"`
					}
					if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
						return err
					}

					if er.Error == "" {
						return errors.New("error field not found in response")
					}
					return nil
				}, nil),
			},
			want:     map[string]string{"error": "Resource not found"},
			wantCode: http.StatusNotFound,
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
	c := &restClient{
		baseURL: "https://example.com",
		client:  http.DefaultClient,
		limiter: rate.NewLimiter(maxRequestRate, maxRequestBurst),
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
				u, e := tt.endpoint.Compile(c.baseURL)
				if e != nil && !tt.wantErr {
					t.Fatalf("Failed to compile endpoint: %v", err)
				}

				httpmock.RegisterResponder(tt.endpoint.Method, u, func(req *http.Request) (*http.Response, error) {
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

			_, isResponse := tt.want.(response)
			if !tt.wantErr && isResponse && !reflect.DeepEqual(got, tt.want) {
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
	*httpmock.MockTransport
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &restClient{
				client: http.DefaultClient,
				wg:     sync.WaitGroup{},
			}

			httpmock.ActivateNonDefault(c.client)
			httpmock.RegisterResponder(http.MethodGet, "https://example.com/resource", httpmock.NewJsonResponderOrPanic(200, map[string]any{"id": 1, "name": "Resource"}))
			c.client.Transport = &mockTransport{MockTransport: http.DefaultClient.Transport.(*httpmock.MockTransport)}

			var wg sync.WaitGroup
			for range tt.numRequests {
				wg.Add(1)
				c.wg.Add(1)
				go func() {
					defer wg.Done()
					defer c.wg.Done()

					<-time.After(tt.reqDelay)
					req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com/resource", http.NoBody)
					if err != nil {
						t.Errorf("Failed to create request: %v", err)
					}

					resp, err := c.client.Do(req)
					if err != nil {
						t.Errorf("Failed to make request: %v", err)
					}
					defer resp.Body.Close()
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

			wg.Wait()

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
	c := &restClient{
		client: http.DefaultClient,
	}

	if c.Client() != http.DefaultClient {
		t.Errorf("Client() = %v, want %v", c.Client(), http.DefaultClient)
	}
}

func TestClient_RateLimiter(t *testing.T) {
	limiter := rate.NewLimiter(maxRequestRate, maxRequestBurst)
	c := &restClient{
		limiter: limiter,
	}

	if c.RateLimiter() != limiter {
		t.Errorf("RateLimiter() = %v, want %v", c.RateLimiter(), limiter)
	}
}

func TestWithDelay(t *testing.T) {
	delay := 5 * time.Second
	request := &Request{
		Http: &http.Request{},
	}
	WithDelay(delay)(request)

	if request.Delay != delay {
		t.Errorf("WithDelay() = %v, want %v", request.Delay, delay)
	}
}

func TestWithHeader(t *testing.T) {
	key, value := "X-Custom-Header", "CustomValue"
	request := &Request{
		Http: &http.Request{
			Header: http.Header{},
		},
	}
	WithHeader(key, value)(request)

	if request.Http.Header.Get(key) != value {
		t.Errorf("WithHeader() = %v, want %v", request.Http.Header.Get(key), value)
	}
}

func TestWithBasicAuth(t *testing.T) {
	username, password := "user", "password"
	request := &Request{
		Http: &http.Request{
			Header: http.Header{},
		},
	}
	WithBasicAuth(username, password)(request)

	auth := fmt.Sprintf("%s:%s", username, password)
	auth = base64.StdEncoding.EncodeToString([]byte(auth))
	if request.Http.Header.Get("Authorization") != fmt.Sprintf("Basic %s", auth) {
		t.Errorf("WithBasicAuth() = %v, want %v", request.Http.Header.Get("Authorization"), fmt.Sprintf("%s:%s", username, password))
	}
}

func TestWithBearer(t *testing.T) {
	token := "token"
	request := &Request{
		Http: &http.Request{
			Header: http.Header{},
		},
	}
	WithBearer(token)(request)

	if request.Http.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", token) {
		t.Errorf("WithBearer() = %v, want %v", request.Http.Header.Get("Authorization"), fmt.Sprintf("Bearer %s", token))
	}
}

func TestWithTracer(t *testing.T) {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com", http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	request := &Request{Http: req}
	c := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {},
	}
	WithTracer(c)(request)

	got := httptrace.ContextClientTrace(request.Http.Context())
	if got == nil {
		t.Fatalf("WithTracer() did not set ClientTrace in context")
	}

	if !reflect.DeepEqual(got, c) {
		t.Errorf("WithTracer() = %v, want %v", got, c)
	}
}
