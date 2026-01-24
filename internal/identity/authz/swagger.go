// Copyright (c) 2025 Justin Cranford
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	"fmt"
	http "net/http"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilApiIdentityAuthz "cryptoutil/api/identity/authz"
)

// ServeOpenAPISpec serves the embedded OpenAPI specification for the AuthZ service.
func ServeOpenAPISpec() (func(c *fiber.Ctx) error, error) {
	rawSpecBytes, err := cryptoutilApiIdentityAuthz.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("failed to get AuthZ OpenAPI spec: %w", err)
	}

	specJSON, err := rawSpecBytes.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AuthZ OpenAPI spec: %w", err)
	}

	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")

		return c.Status(http.StatusOK).Send(specJSON)
	}, nil
}
