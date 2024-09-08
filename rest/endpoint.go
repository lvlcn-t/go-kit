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

// Get creates a new [Endpoint] with the [http.MethodGet] method and the given path and queries.
func Get(path string, queries ...url.Values) *Endpoint {
	return &Endpoint{Method: http.MethodGet, Path: path, Query: combineQueries(queries...)}
}

// Post creates a new [Endpoint] with the [http.MethodPost] method and the given path and queries.
func Post(path string, queries ...url.Values) *Endpoint {
	return &Endpoint{Method: http.MethodPost, Path: path, Query: combineQueries(queries...)}
}

// Put creates a new [Endpoint] with the [http.MethodPut] method and the given path and queries.
func Put(path string, queries ...url.Values) *Endpoint {
	return &Endpoint{Method: http.MethodPut, Path: path, Query: combineQueries(queries...)}
}

// Patch creates a new [Endpoint] with the [http.MethodPatch] method and the given path and queries.
func Patch(path string, queries ...url.Values) *Endpoint {
	return &Endpoint{Method: http.MethodPatch, Path: path, Query: combineQueries(queries...)}
}

// Delete creates a new [Endpoint] with the [http.MethodDelete] method and the given path and queries.
func Delete(path string, queries ...url.Values) *Endpoint {
	return &Endpoint{Method: http.MethodDelete, Path: path, Query: combineQueries(queries...)}
}

// Query adds the given key and value to the endpoint query.
func (e *Endpoint) AddQuery(key, value string) *Endpoint {
	query := url.Values{}
	query.Add(key, value)
	e.Query = combineQueries(e.Query, query)
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

// combineQueries combines multiple queries into one.
func combineQueries(queries ...url.Values) url.Values {
	if len(queries) == 0 {
		return nil
	}

	q := url.Values{}
	for _, query := range queries {
		for key, values := range query {
			q[key] = values
		}
	}
	return q
}
