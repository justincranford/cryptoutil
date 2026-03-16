// Copyright (c) 2025 Justin Cranford
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package idp

import (
	"fmt"
	http "net/http"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilApiIdentityIdp "cryptoutil/api/identity/idp"
)

// ServeOpenAPISpec serves the embedded OpenAPI specification for the IdP service.
func ServeOpenAPISpec() (func(c *fiber.Ctx) error, error) {
	rawSpecBytes, err := cryptoutilApiIdentityIdp.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("failed to get IdP OpenAPI spec: %w", err)
	}

	specJSON, err := rawSpecBytes.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal IdP OpenAPI spec: %w", err)
	}

	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")

		return c.Status(http.StatusOK).Send(specJSON)
	}, nil
}
