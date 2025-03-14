package openapi

import (
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// FiberHandlerOpenAPISpec Expose OpenAPI spec embedded inside openapi-gen.go, Swagger UI needs it to render the APIs
func FiberHandlerOpenAPISpec() func(c *fiber.Ctx) error {
	rawSpecBytes, err := rawSpec()
	if err != nil {
		log.Fatalf("Missing openapi spec: %v", err)
	}
	return func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).Send(rawSpecBytes)
	}
}
