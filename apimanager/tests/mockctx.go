package tests

import "github.com/gofiber/fiber/v3"

//go:generate moq -out mockctx_moq.go . Ctx
type Ctx interface {
	fiber.CustomCtx
}
