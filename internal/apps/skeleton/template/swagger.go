// Copyright (c) 2025 Justin Cranford
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package template

import (
	cryptoutilApiSkeletonTemplateServer "cryptoutil/api/skeleton-template/server"
	cryptoutilAppsTemplateServiceServerBuilder "cryptoutil/internal/apps/template/service/server/builder"

	fiber "github.com/gofiber/fiber/v2"
)

// ServeOpenAPISpec serves the embedded OpenAPI specification for the Skeleton Template service.
func ServeOpenAPISpec() (func(c *fiber.Ctx) error, error) {
	return cryptoutilAppsTemplateServiceServerBuilder.FiberHandlerOpenAPISpec(cryptoutilApiSkeletonTemplateServer.GetSwagger)
}
