// Copyright (c) 2025 Justin Cranford
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package idp

import (
	cryptoutilApiIdentityIdp "cryptoutil/api/identity-idp"
	cryptoutilAppsTemplateServiceServerBuilder "cryptoutil/internal/apps/template/service/server/builder"

	fiber "github.com/gofiber/fiber/v2"
)

// ServeOpenAPISpec serves the embedded OpenAPI specification for the IdP service.
func ServeOpenAPISpec() (func(c *fiber.Ctx) error, error) {
	return cryptoutilAppsTemplateServiceServerBuilder.FiberHandlerOpenAPISpec(cryptoutilApiIdentityIdp.GetSwagger)
}
