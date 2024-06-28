package fiberutils

import (
	"fmt"
	"net"
	"reflect"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// Params returns the parameter with the given name from the context and converts it to the given type.
func Params[T any](c fiber.Ctx, name string, parse Parser[T]) (T, error) {
	return parseParam(name, c.Params, parse)
}

// Query returns the query parameter with the given name from the context and converts it to the given type.
func Query[T any](c fiber.Ctx, name string, parse Parser[T]) (T, error) {
	return parseParam(name, c.Query, parse)
}

// Cookies returns the cookie with the given name from the context and converts it to the given type.
func Cookies[T any](c fiber.Ctx, name string, parse Parser[T]) (T, error) {
	return parseParam(name, c.Cookies, parse)
}

// Form returns the form value with the given name from the context and converts it to the given type.
func Form[T any](c fiber.Ctx, name string, parse Parser[T]) (T, error) {
	return parseParam(name, c.FormValue, parse)
}

// Header returns the request header with the given name from the context and converts it to the given type.
func Header[T any](c fiber.Ctx, name string, parse Parser[T]) (T, error) {
	return parseParam(name, c.Get, parse)
}

// Body returns the body of the request and converts it to the given type.
func Body[T any](c fiber.Ctx) (T, error) {
	var v T
	if reflect.TypeOf(v).Kind() == reflect.Map {
		v = reflect.MakeMap(reflect.TypeOf(v)).Interface().(T)
	}

	if reflect.TypeOf(v).Kind() == reflect.Pointer {
		v = reflect.New(reflect.TypeOf(v).Elem()).Interface().(T)
	}

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
func parseParam[T any](name string, get func(string, ...string) string, parse Parser[T]) (T, error) {
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
