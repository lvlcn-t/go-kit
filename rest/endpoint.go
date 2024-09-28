package rest

import (
	"net/http"
	"net/url"
)

// Endpoint represents a REST endpoint.
type Endpoint struct {
	// Method is the HTTP method to use for the request.
	Method string
	// Path is the URL path to the endpoint.
	Path string
	// Query is the URL query parameters to use for the request.
	Query url.Values
}

// Get creates a new [Endpoint] with the [http.MethodGet] method and the given path.
func Get(path string) *Endpoint {
	return &Endpoint{Method: http.MethodGet, Path: path}
}

// Post creates a new [Endpoint] with the [http.MethodPost] method and the given path.
func Post(path string) *Endpoint {
	return &Endpoint{Method: http.MethodPost, Path: path}
}

// Put creates a new [Endpoint] with the [http.MethodPut] method and the given path.
func Put(path string) *Endpoint {
	return &Endpoint{Method: http.MethodPut, Path: path}
}

// Patch creates a new [Endpoint] with the [http.MethodPatch] method and the given path.
func Patch(path string) *Endpoint {
	return &Endpoint{Method: http.MethodPatch, Path: path}
}

// Delete creates a new [Endpoint] with the [http.MethodDelete] method and the given path.
func Delete(path string) *Endpoint {
	return &Endpoint{Method: http.MethodDelete, Path: path}
}

// AddQuery adds a query parameter value to a key.
// It appends to any existing values associated with key.
func (e *Endpoint) AddQuery(key, value string) *Endpoint {
	if e.Query == nil {
		e.Query = url.Values{}
	}
	e.Query.Add(key, value)
	return e
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
	// e.g. If baseURL is "https://example.com" and path is "/resource" the full URL will be "https://example.com/resource"
	// If the path is "https://example.com/resource" it will be used as is
	u := base.ResolveReference(path)
	if e.Query != nil {
		u.RawQuery = e.Query.Encode()
	}

	return u.String(), nil
}
