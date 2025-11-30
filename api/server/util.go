// Copyright (c) 2025 Justin Cranford

package server

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// FiberHandlerOpenAPISpec Expose OpenAPI spec embedded inside openapi-gen.go, Swagger UI needs it to render the APIs.
func FiberHandlerOpenAPISpec() (func(c *fiber.Ctx) error, error) {
	rawSpecBytes, err := rawSpec()
	if err != nil {
		return nil, fmt.Errorf("missing openapi spec: %w", err)
	}

	return func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).Send(rawSpecBytes)
	}, nil
}
