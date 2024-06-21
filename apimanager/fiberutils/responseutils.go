package fiberutils

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
)

// NewErrorResponse creates a new error response.
func NewErrorResponse(msg string, status int) fiber.Map {
	code := http.StatusText(status)
	if code == "" {
		code = "Unknown"
	}

	return fiber.Map{"error": fiber.Map{"message": msg, "code": code}}
}

// BadRequestResponse sends a bad request response.
// If the ctype parameter is given, this method will set the Content-Type header equal to ctype.
// If ctype is not given, The Content-Type header will be set to application/json.
func BadRequestResponse(c fiber.Ctx, msg string, ctype ...string) error {
	return c.Status(http.StatusBadRequest).JSON(NewErrorResponse(msg, http.StatusBadRequest), ctype...)
}

// UnauthorizedResponse sends an unauthorized response.
// If the ctype parameter is given, this method will set the Content-Type header equal to ctype.
// If ctype is not given, The Content-Type header will be set to application/json.
func UnauthorizedResponse(c fiber.Ctx, msg string, ctype ...string) error {
	return c.Status(http.StatusUnauthorized).JSON(NewErrorResponse(msg, http.StatusUnauthorized), ctype...)
}

// ForbiddenResponse sends a forbidden response.
// If the ctype parameter is given, this method will set the Content-Type header equal to ctype.
// If ctype is not given, The Content-Type header will be set to application/json.
func ForbiddenResponse(c fiber.Ctx, msg string, ctype ...string) error {
	return c.Status(http.StatusForbidden).JSON(NewErrorResponse(msg, http.StatusForbidden), ctype...)
}

// NotFoundResponse sends a not found response.
// If the ctype parameter is given, this method will set the Content-Type header equal to ctype.
// If ctype is not given, The Content-Type header will be set to application/json.
func NotFoundResponse(c fiber.Ctx, msg string, ctype ...string) error {
	return c.Status(http.StatusNotFound).JSON(NewErrorResponse(msg, http.StatusNotFound), ctype...)
}

// InternalServerErrorResponse sends an internal server error response.
// If the ctype parameter is given, this method will set the Content-Type header equal to ctype.
// If ctype is not given, The Content-Type header will be set to application/json.
func InternalServerErrorResponse(c fiber.Ctx, msg string, ctype ...string) error {
	return c.Status(http.StatusInternalServerError).JSON(NewErrorResponse(msg, http.StatusInternalServerError), ctype...)
}

// ServiceUnavailableResponse sends a service unavailable response.
// If the ctype parameter is given, this method will set the Content-Type header equal to ctype.
// If ctype is not given, The Content-Type header will be set to application/json.
func ServiceUnavailableResponse(c fiber.Ctx, msg string, ctype ...string) error {
	return c.Status(http.StatusServiceUnavailable).JSON(NewErrorResponse(msg, http.StatusServiceUnavailable), ctype...)
}
