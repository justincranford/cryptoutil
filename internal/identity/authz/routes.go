// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers all OAuth 2.1 authorization server routes.
func (s *Service) RegisterRoutes(app *fiber.App) {
	// Swagger UI OpenAPI spec endpoint.
	swaggerHandler, err := ServeOpenAPISpec()
	if err != nil {
		// Swagger UI is non-critical, skip if spec generation fails.
		// Error already includes context from ServeOpenAPISpec().
		_ = err
	} else {
		app.Get("/ui/swagger/doc.json", swaggerHandler)
	}

	// Health check endpoint (no prefix).
	app.Get("/health", s.handleHealth)

	// OAuth 2.1 endpoints with /oauth2/v1 prefix.
	oauth := app.Group("/oauth2/v1")
	oauth.Get("/authorize", s.handleAuthorizeGET)
	oauth.Post("/authorize", s.handleAuthorizePOST)
	oauth.Post("/token", s.handleToken)
	oauth.Post("/introspect", s.handleIntrospect)
	oauth.Post("/revoke", s.handleRevoke)
}
