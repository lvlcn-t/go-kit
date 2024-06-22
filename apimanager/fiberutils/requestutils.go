package fiberutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

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
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	err := decoder.Decode(&v)
	if err != nil {
		return v, fmt.Errorf("failed to parse request body into %T: %w", v, err)
	}
	return v, nil
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

type Client struct {
	ip   string
	port string
}

func NewClient(c fiber.Ctx) Client {
	return Client{
		ip:   c.IP(),
		port: c.Port(),
	}
}

func (b *Client) String() string {
	return fmt.Sprintf("%s:%s", b.ip, b.port)
}

func (b *Client) IP() IP {
	return IP(net.ParseIP(b.ip))
}

func (b *Client) Port() Port {
	v, err := strconv.Atoi(b.port)
	if err != nil {
		return Port(0)
	}
	return Port(v)
}

type IP net.IP

func (i IP) String() string {
	return net.IP(i).String()
}

func (i IP) Get() net.IP {
	return net.IP(i)
}

type Port uint16

func (p Port) String() string {
	return strconv.Itoa(int(p))
}

func (p Port) Get() uint16 {
	return uint16(p)
}
