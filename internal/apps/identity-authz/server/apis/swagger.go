// Copyright (c) 2025-2026 Justin Cranford.
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package apis

import (
	cryptoutilApiIdentityAuthz "cryptoutil/api/identity-authz"
	cryptoutilAppsFrameworkServiceServerBuilder "cryptoutil/internal/apps-framework/service/server/builder"

	fiber "github.com/gofiber/fiber/v2"
)

// ServeOpenAPISpec serves the embedded OpenAPI specification for the AuthZ service.
func ServeOpenAPISpec() (func(c *fiber.Ctx) error, error) {
	return cryptoutilAppsFrameworkServiceServerBuilder.FiberHandlerOpenAPISpec(cryptoutilApiIdentityAuthz.GetSwagger)
}
