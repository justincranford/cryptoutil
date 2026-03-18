// Copyright (c) 2025 Justin Cranford
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package ja

import (
	cryptoutilApiJoseJaServer "cryptoutil/api/jose-ja/server"
	cryptoutilAppsTemplateServiceServerBuilder "cryptoutil/internal/apps/template/service/server/builder"

	fiber "github.com/gofiber/fiber/v2"
)

// ServeOpenAPISpec serves the embedded OpenAPI specification for the JWK Authority service.
func ServeOpenAPISpec() (func(c *fiber.Ctx) error, error) {
	return cryptoutilAppsTemplateServiceServerBuilder.FiberHandlerOpenAPISpec(cryptoutilApiJoseJaServer.GetSwagger)
}
