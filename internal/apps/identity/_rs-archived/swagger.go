// Copyright (c) 2025 Justin Cranford
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package rs

import (
	"fmt"
	http "net/http"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilApiIdentityRs "cryptoutil/api/identity/rs"
)

// ServeOpenAPISpec serves the embedded OpenAPI specification for the Resource Server.
func ServeOpenAPISpec() (func(c *fiber.Ctx) error, error) {
	rawSpecBytes, err := cryptoutilApiIdentityRs.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("failed to get RS OpenAPI spec: %w", err)
	}

	specJSON, err := rawSpecBytes.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RS OpenAPI spec: %w", err)
	}

	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")

		return c.Status(http.StatusOK).Send(specJSON)
	}, nil
}
