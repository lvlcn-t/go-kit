package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// DefaultClient is the default rest client used for making requests.
var DefaultClient = defaultClient()

// DefaultTransport is the default transport used for the default client.
var DefaultTransport = defaultTransport()

// Do makes a request to the given endpoint with the given payload and response type.
// It applies the given options and returns an error if the request fails.
//
// Example:
//
//	// Define the request endpoint
//	ctx := context.Background()
//	endpoint := rest.Get("https://api.example.com/resource")
//
//	// Define the response type
//	type response struct {
//		ID   int    `json:"id"`
//		Name string `json:"name"`
//	}
//
//	// Make the request
//	resp, status, err := rest.Do[response](ctx, endpoint, nil)
//	if err != nil {
//		// Handle error
//	}
//
// The request will be made to "https://api.example.com/resource" with the payload marshaled to JSON
// and the response unmarshaled into a response object with the given type.
func Do[T any](ctx context.Context, endpoint *Endpoint, payload any, opts ...RequestOption) (resp T, code int, err error) {
	status, err := DefaultClient.Do(ctx, endpoint, payload, &resp, opts...)
	return resp, status, err
}

// Close closes the default rest client and gracefully awaits all pending requests to finish.
// If the context is canceled, it will close the idle connections immediately.
func Close(ctx context.Context) {
	DefaultClient.Close(ctx)
}

// Client allows doing requests to different endpoints.
// It provides a simple way to make requests with rate limiting and request options.
// The client is safe for concurrent use.
//
//go:generate moq -out client_moq.go . Client
type Client interface {
	// Do makes a request to the given [Endpoint], with the given payload and response objects. It applies the given options.
	// Returns the status code of the response and an error if the request fails.
	//
	// Example:
	//	ctx := context.Background()
	// 	client := rest.NewClient("https://api.example.com", 5*time.Second)
	// 	defer client.Close(ctx)
	//
	// 	endpoint := rest.Post("/resource")
	// 	payload := map[string]string{"key": "value"}
	// 	var response map[string]any
	// 	status, err := client.Do(ctx, endpoint, payload, &response)
	//	if err != nil {
	// 		// Handle error
	// 	}
	//
	// The request will be made to "https://api.example.com/resource" with the payload marshaled to JSON
	// and the response unmarshaled into the response object.
	Do(ctx context.Context, endpoint *Endpoint, payload, response any, opts ...RequestOption) (int, error)
	// Close closes the rest client and gracefully awaits all pending requests to finish.
	// If the context is canceled, it will close the idle connections immediately.
	Close(ctx context.Context)
	// Client returns the [http.Client] the rest client uses.
	Client() *http.Client
	// RateLimiter returns the [rate.Limiter] the rest client uses.
	RateLimiter() *rate.Limiter
}

// Request represents a request to be made by the rest client.
type Request struct {
	// Request is the HTTP request to be made.
	Request *http.Request
	// Delay is the amount of time to wait before executing the request.
	Delay time.Duration
}

// RequestOption is a function that modifies a request.
type RequestOption func(*Request)

var _ Client = (*restClient)(nil)

const (
	// DefaultTimeout is the default timeout for requests.
	DefaultTimeout = 60 * time.Second
	// maxIdleConns controls the maximum number of idle (keep-alive) connections across all hosts.
	maxIdleConns = 100
	// maxIdleConnsPerHost controls the maximum number of idle (keep-alive) connections to keep per-host.
	maxIdleConnsPerHost = 100
	// idleConnTimeout controls the maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.
	idleConnTimeout = 90 * time.Second
	// maxRequestRate is the maximum number of requests that can be made in a single second.
	maxRequestRate rate.Limit = 10
	// maxRequestBurst is the maximum number of requests that can be made in a single moment.
	maxRequestBurst = 10
)

// restClient is the default implementation of the Client interface.
// It is used for making requests to different endpoints.
type restClient struct {
	// baseURL is the base URL for all requests.
	baseURL string
	// client is the HTTP client used for requests.
	client *http.Client
	// limiter is the rate limiter used for requests.
	limiter *rate.Limiter
	// wg is the wait group used to track pending requests.
	wg sync.WaitGroup
}

// New creates a new rest client with the given base URL.
// You can optionally provide a timeout for requests. If no timeout is provided, the [DefaultTimeout] will be used.
func New(baseURL string, timeout ...time.Duration) (Client, error) {
	if len(timeout) == 0 {
		return NewWithClient(baseURL, nil)
	}

	return NewWithClient(baseURL, &http.Client{Transport: defaultTransport(), Timeout: timeout[0]})
}

// NewWithClient creates a new rest client with the given base URL and [http.Client].
// If the client is nil, it will create a new client with the [DefaultTransport] and [DefaultTimeout].
func NewWithClient(baseURL string, client *http.Client) (Client, error) {
	if _, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	if client == nil {
		client = &http.Client{Transport: defaultTransport(), Timeout: DefaultTimeout}
	}

	return &restClient{
		baseURL: baseURL,
		client:  client,
		limiter: rate.NewLimiter(maxRequestRate, maxRequestBurst),
	}, nil
}

// Client returns the HTTP client the rest client uses.
func (r *restClient) Client() *http.Client {
	return r.client
}

// RateLimiter returns the rate limiter the rest client uses.
func (r *restClient) RateLimiter() *rate.Limiter {
	return r.limiter
}

// Do makes a request to the given endpoint with the given payload and response objects.
// It applies the given options and returns an error if the request fails.
// If the response cannot be unmarshalled into the response object, it returns an [ErrDecodingResponse].
func (r *restClient) Do(ctx context.Context, endpoint *Endpoint, payload, response any, opts ...RequestOption) (int, error) {
	if endpoint == nil {
		return 0, errors.New("endpoint is nil")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return r.do(ctx, endpoint, payload, response, opts)
}

// do is the implementation of the [Client].Do method that makes the request to the given endpoint.
func (r *restClient) do(ctx context.Context, endpoint *Endpoint, payload, response any, opts []RequestOption) (int, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return 0, ErrRateLimitExceeded
	}

	body := io.Reader(http.NoBody)
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return 0, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(data)
	}

	u, err := endpoint.Compile(r.baseURL)
	if err != nil {
		return 0, fmt.Errorf("failed to compile endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, endpoint.Method, u, body)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	request := &Request{Request: req, Delay: 0}
	for _, opt := range opts {
		opt(request)
	}

	if request.Delay > 0 {
		select {
		case <-time.After(request.Delay):
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}

	r.wg.Add(1)
	defer r.wg.Done()
	resp, err := r.client.Do(request.Request)
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return resp.StatusCode, &ErrDecodingResponse{err: err}
		}
	}

	return resp.StatusCode, nil
}

// Close closes the rest client and gracefully awaits all pending requests to finish.
// If the context is canceled, it will close the idle connections immediately.
func (r *restClient) Close(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		r.wg.Wait()
	}()

	select {
	case <-ctx.Done():
		r.client.CloseIdleConnections()
	case <-done:
	}

	// Ensure all idle connections are closed even if all requests should be done.
	r.client.CloseIdleConnections()
}

// ErrRateLimitExceeded is the error returned when the rate limit is exceeded.
var ErrRateLimitExceeded = errors.New("rate limit exceeded")

// ErrDecodingResponse is the error returned when the response cannot be unmarshalled into the response object.
type ErrDecodingResponse struct{ err error }

// Error returns the error message.
func (e *ErrDecodingResponse) Error() string {
	return fmt.Sprintf("failed to decode response: %v", e.err)
}

// Is checks if the target error is an [ErrDecodingResponse].
func (e *ErrDecodingResponse) Is(target error) bool {
	_, ok := target.(*ErrDecodingResponse)
	return ok
}

// Unwrap returns the wrapped error.
func (e *ErrDecodingResponse) Unwrap() error { return e.err }

// WithDelay is a request option that adds a delay before executing the request
func WithDelay(d time.Duration) RequestOption {
	return func(r *Request) {
		r.Delay = d
	}
}

// WithHeader is a request option that sets custom headers for the request
func WithHeader(key, value string) RequestOption {
	return func(r *Request) {
		r.Request.Header.Set(key, value)
	}
}

// WithBearer is a request option that sets a bearer token for the request
func WithBearer(token string) RequestOption {
	return WithHeader("Authorization", fmt.Sprintf("Bearer %s", token))
}

// WithBasicAuth is a request option that sets basic auth for the request
func WithBasicAuth(username, password string) RequestOption {
	return func(r *Request) {
		r.Request.SetBasicAuth(username, password)
	}
}

// WithTracer is a request option that sets a [httptrace.ClientTrace] for the request.
func WithTracer(c *httptrace.ClientTrace) RequestOption {
	return func(r *Request) {
		r.Request = r.Request.WithContext(httptrace.WithClientTrace(r.Request.Context(), c))
	}
}

// defaultClient creates a new [restClient] without a base URL.
// Panics if the client cannot be created.
func defaultClient() Client {
	c, err := New("")
	if err != nil {
		panic(fmt.Sprintf("failed to create default client: %v", err))
	}
	return c
}

// defaultTransport creates a new [http.Transport] with custom settings.
func defaultTransport() *http.Transport {
	tp := http.DefaultTransport.(*http.Transport).Clone()
	tp.MaxIdleConns = maxIdleConns
	tp.MaxIdleConnsPerHost = maxIdleConnsPerHost
	tp.IdleConnTimeout = idleConnTimeout
	return tp
}
