// Copyright (c) 2025 Justin Cranford
//
//

package idp

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers all OIDC identity provider routes.
func (s *Service) RegisterRoutes(app *fiber.App) {
	// Swagger UI OpenAPI spec endpoint.
	swaggerHandler, err := ServeOpenAPISpec()
	if err != nil {
		// Log error but continue - Swagger UI is non-critical.
		// TODO: Add structured logging when logger available in Service.
		_ = err
	} else {
		app.Get("/ui/swagger/doc.json", swaggerHandler)
	}

	// Health check endpoint (no prefix).
	app.Get("/health", s.handleHealth)

	// OIDC Discovery endpoints with /.well-known prefix.
	_ = app.Group("/.well-known")
	// JWKS endpoint will be added here when handler implementation complete.

	// OIDC IdP endpoints with /oidc/v1 prefix.
	oidc := app.Group("/oidc/v1")
	oidc.Get("/login", s.handleLogin)
	oidc.Post("/login", s.handleLoginSubmit)
	oidc.Get("/consent", s.AuthMiddleware(), s.handleConsent)
	oidc.Post("/consent", s.AuthMiddleware(), s.handleConsentSubmit)
	oidc.Get("/userinfo", s.TokenAuthMiddleware(), s.handleUserInfo)
	oidc.Post("/logout", s.AuthMiddleware(), s.handleLogout)
}
