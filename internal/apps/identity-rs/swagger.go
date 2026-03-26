// Copyright (c) 2025 Justin Cranford
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package rs

import (
	cryptoutilApiIdentityRs "cryptoutil/api/identity-rs"
	cryptoutilAppsFrameworkServiceServerBuilder "cryptoutil/internal/apps/framework/service/server/builder"

	fiber "github.com/gofiber/fiber/v2"
)

// ServeOpenAPISpec serves the embedded OpenAPI specification for the Resource Server.
func ServeOpenAPISpec() (func(c *fiber.Ctx) error, error) {
	return cryptoutilAppsFrameworkServiceServerBuilder.FiberHandlerOpenAPISpec(cryptoutilApiIdentityRs.GetSwagger)
}
