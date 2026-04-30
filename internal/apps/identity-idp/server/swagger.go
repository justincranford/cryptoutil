// Copyright (c) 2025-2026 Justin Cranford.
//
//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package server

import (
	cryptoutilApiIdentityIdp "cryptoutil/api/identity-idp"
	cryptoutilAppsFrameworkServiceServerBuilder "cryptoutil/internal/apps-framework/service/server/builder"

	fiber "github.com/gofiber/fiber/v2"
)

// ServeOpenAPISpec serves the embedded OpenAPI specification for the identity-idp service.
func ServeOpenAPISpec() (func(c *fiber.Ctx) error, error) {
	return cryptoutilAppsFrameworkServiceServerBuilder.FiberHandlerOpenAPISpec(cryptoutilApiIdentityIdp.GetSwagger)
}
