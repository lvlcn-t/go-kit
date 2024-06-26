package fiberutils

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Params returns the parameter with the given name from the context and converts it to the given type.
func Params[T any](c fiber.Ctx, name string, parse func(string) (T, error)) (T, error) {
	return parseParam(name, c.Params, parse)
}

// Query returns the query parameter with the given name from the context and converts it to the given type.
func Query[T any](c fiber.Ctx, name string, parse func(string) (T, error)) (T, error) {
	return parseParam(name, c.Query, parse)
}

// Cookies returns the cookie with the given name from the context and converts it to the given type.
func Cookies[T any](c fiber.Ctx, name string, parse func(string) (T, error)) (T, error) {
	return parseParam(name, c.Cookies, parse)
}

// Form returns the form value with the given name from the context and converts it to the given type.
func Form[T any](c fiber.Ctx, name string, parse func(string) (T, error)) (T, error) {
	return parseParam(name, c.FormValue, parse)
}

// Header returns the request header with the given name from the context and converts it to the given type.
func Header[T any](c fiber.Ctx, name string, parse func(string) (T, error)) (T, error) {
	return parseParam(name, c.Get, parse)
}

// Body returns the body of the request and converts it to the given type.
func Body[T any](c fiber.Ctx) (T, error) {
	var v T
	if err := c.Bind().Body(&v); err != nil {
		return v, err
	}
	return v, nil
}

// Validator is an interface that can be implemented by types that need to be validated.
type Validator interface {
	// Validate validates the type and returns an error if it is invalid.
	Validate() error
}

// BodyWithValidation returns the body of the request and converts it to the given type, then validates it.
func BodyWithValidation[T Validator](c fiber.Ctx) (T, error) {
	data, err := Body[T](c)
	if err != nil {
		return data, err
	}

	if err := data.Validate(); err != nil {
		return data, err
	}

	return data, nil
}

// parseParam parses a parameter from a context using the given getter and parser.
func parseParam[T any](name string, get func(string, ...string) string, parse func(string) (T, error)) (T, error) {
	var empty T
	v := get(name)
	if v == "" {
		return empty, &ErrParameterNotFound{name: name}
	}
	if parse == nil {
		if _, ok := any(empty).(string); ok {
			return any(v).(T), nil
		}
		return empty, fmt.Errorf("no parser provided for parameter %q", name)
	}

	return parse(v)
}

// Client represents a client that made a request.
type Client struct {
	ip   string
	port string
}

// NewClient creates a new client from the given context.
func NewClient(c fiber.Ctx) Client {
	return Client{
		ip:   c.IP(),
		port: c.Port(),
	}
}

// String returns the client as a string in the format "ip:port".
func (b *Client) String() string {
	return fmt.Sprintf("%s:%s", b.ip, b.port)
}

// IP returns the IP of the client.
func (b *Client) IP() IP {
	return IP(net.ParseIP(b.ip))
}

// Port returns the port of the client.
func (b *Client) Port() Port {
	v, err := strconv.Atoi(b.port)
	if err != nil {
		return Port(0)
	}
	return Port(v)
}

// IP represents an IP address.
type IP net.IP

// String returns the IP as a string.
func (i IP) String() string {
	return net.IP(i).String()
}

// Get returns the IP as a [net.IP].
func (i IP) Get() net.IP {
	return net.IP(i)
}

// Port represents a port number.
type Port uint16

// String returns the port as a string.
func (p Port) String() string {
	return strconv.Itoa(int(p))
}

// Get returns the port as a uint16.
func (p Port) Get() uint16 {
	return uint16(p)
}

// ParseDate returns a parser that parses a date string into a time.Time using the given formats.
// The first format that successfully parses the date will be used.
// If no formats are provided, the default format [time.DateOnly] will be used.
func ParseDate(format ...string) func(string) (time.Time, error) {
	if len(format) == 0 {
		format = append(format, time.DateOnly)
	}

	return func(s string) (t time.Time, err error) {
		for _, f := range format {
			t, err = time.Parse(f, s)
			if err == nil {
				return t, nil
			}
		}
		return t, err
	}
}

// ParseTime returns a parser that parses a time string into a time.Time using the given formats.
// The first format that successfully parses the time will be used.
// If no formats are provided, the default format [time.TimeOnly] will be used.
func ParseTime(format ...string) func(string) (time.Time, error) {
	if len(format) == 0 {
		format = append(format, time.TimeOnly)
	}

	return func(s string) (t time.Time, err error) {
		for _, f := range format {
			t, err = time.Parse(f, s)
			if err == nil {
				return t, nil
			}
		}
		return t, err
	}
}

// ParseDateTime returns a parser that parses a date and time string into a time.Time using the given formats.
// The first format that successfully parses the date and time will be used.
// If no formats are provided, the default format [time.RFC3339] will be used.
func ParseDateTime(format ...string) func(string) (time.Time, error) {
	if len(format) == 0 {
		format = append(format, time.RFC3339)
	}

	return func(s string) (t time.Time, err error) {
		for _, f := range format {
			t, err = time.Parse(f, s)
			if err == nil {
				return t, nil
			}
		}
		return t, err
	}
}
