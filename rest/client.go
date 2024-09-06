package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var _ Client = (*client)(nil)

// Client allows doing requests to different endpoints of one host.
// It provides a simple way to make requests with rate limiting and request options.
// The client is safe for concurrent use.
//
//go:generate moq -out client_moq.go . Client
type Client interface {
	// Do makes a request to the given [Endpoint], with the given payload and response objects. It applies the given options.
	// Returns the status code of the response and an error if the request fails.
	Do(ctx context.Context, endpoint *Endpoint, payload, response any, opts ...RequestOption) (int, error)
	// Close closes the rest client and awaits all pending requests to finish. You can use a canceling context to abort the waiting.
	Close(ctx context.Context)
	// Client returns the [http.Client] the rest client uses.
	Client() *http.Client
	// RateLimiter returns the [rate.Limiter] the rest client uses.
	RateLimiter() *rate.Limiter
}

// Endpoint represents a REST endpoint.
type Endpoint struct {
	// Method is the HTTP method to use for the request.
	Method string
	// Path is the URL path to the endpoint.
	Path string
	// Query is the URL query parameters to use for the request.
	Query url.Values
}

// Compile compiles the endpoint into a full URL.
func (e *Endpoint) Compile(baseURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	// Create a relative URL for the endpoint path
	path, err := url.Parse(e.Path)
	if err != nil {
		return "", err
	}

	// Resolve the base URL with the relative path. This will give us the full URL for the request.
	// e.g. If baseURL is "http://example.com" and path is "/resource" the full URL will be "http://example.com/resource"
	// If the path is "http://example.com/resource" it will be used as is
	u := base.ResolveReference(path)
	if e.Query != nil {
		u.RawQuery = e.Query.Encode()
	}

	return u.String(), nil
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

const (
	// maxIdleConns controls the maximum number of idle (keep-alive) connections across all hosts.
	maxIdleConns = 100
	// maxIdleConnsPerHost controls the maximum number of idle (keep-alive) connections to keep per-host.
	maxIdleConnsPerHost = 100
	// idleConnTimeout controls the maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.
	idleConnTimeout = 90 * time.Second
)

var (
	// defaultRateLimiter is the default rate limiter for the rest client.
	// It allows 10 requests per second with a burst of 10 (burst is the maximum number of requests that can be made in a single moment).
	defaultRateLimiter = rate.NewLimiter(rate.Limit(10), 10) //nolint:mnd // No need for another constant.
	// ErrRateLimitExceeded is the error returned when the rate limit is exceeded.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

// client is the default implementation of the Client interface.
// The client is used for making requests to different endpoints of one base URL.
type client struct {
	// baseURL is the base URL for all requests.
	baseURL string
	// client is the HTTP client used for requests.
	client *http.Client
	// rateLimiter is the rate limiter used for requests.
	rateLimiter *rate.Limiter
	// wg is the wait group used to wait for all requests to finish.
	wg sync.WaitGroup
}

// NewClient creates a new rest client with the given base URL and timeout.
func NewClient(baseURL string, timeout time.Duration) (Client, error) {
	if _, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	return &client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        maxIdleConns,
				MaxIdleConnsPerHost: maxIdleConnsPerHost,
				IdleConnTimeout:     idleConnTimeout,
			},
		},
		rateLimiter: defaultRateLimiter,
	}, nil
}

// Client returns the HTTP client the rest client uses.
func (r *client) Client() *http.Client {
	return r.client
}

// RateLimiter returns the rate limiter the rest client uses.
func (r *client) RateLimiter() *rate.Limiter {
	return r.rateLimiter
}

// Do makes a request to the given endpoint with the given payload and response objects.
// It applies the given options and returns an error if the request fails.
func (r *client) Do(ctx context.Context, endpoint *Endpoint, payload, response any, opts ...RequestOption) (int, error) {
	if ctx == nil || endpoint == nil {
		return 0, errors.New("context and endpoint must not be nil")
	}

	if err := r.rateLimiter.Wait(ctx); err != nil {
		return 0, ErrRateLimitExceeded
	}

	body := io.Reader(http.NoBody)
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return 0, err
		}
		body = bytes.NewBuffer(data)
	}

	u, err := endpoint.Compile(r.baseURL)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, endpoint.Method, u, body)
	if err != nil {
		return 0, err
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
		return 0, err
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return resp.StatusCode, err
		}
	}

	return resp.StatusCode, nil
}

// Close closes the rest client and awaits all pending requests to finish.
// You can use a canceling context to abort the waiting.
// This also ensures that any ongoing connections are closed gracefully or forced if the context is canceled.
func (r *client) Close(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		r.wg.Wait()
	}()

	select {
	case <-ctx.Done():
		if transport, ok := r.client.Transport.(interface{ CloseIdleConnections() }); ok {
			transport.CloseIdleConnections()
		}
	case <-done:
	}

	if transport, ok := r.client.Transport.(interface{ CloseIdleConnections() }); ok {
		transport.CloseIdleConnections()
	}
}

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
