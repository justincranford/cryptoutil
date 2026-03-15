// Copyright (c) 2025 Justin Cranford
//
//

package idp

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	fiber "github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers all OIDC identity provider routes.
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

	// OIDC Discovery endpoints with /.well-known prefix.
	wellKnown := app.Group("/.well-known")
	wellKnown.Get("/openid-configuration", s.handleDiscovery)
	wellKnown.Get("/jwks.json", s.handleJWKS)

	// OIDC IdP endpoints with /oidc/v1 prefix.
	oidc := app.Group("/oidc/v1")
	oidc.Get("/login", s.handleLogin)
	oidc.Post("/login", s.handleLoginSubmit)
	oidc.Get("/consent", s.AuthMiddleware(), s.handleConsent)
	oidc.Post("/consent", s.AuthMiddleware(), s.handleConsentSubmit)
	oidc.Get(cryptoutilSharedMagic.PathUserInfo, s.TokenAuthMiddleware(), s.handleUserInfo)
	oidc.Post(cryptoutilSharedMagic.PathLogout, s.AuthMiddleware(), s.handleLogout)
	oidc.Get(cryptoutilSharedMagic.PathEndSession, s.handleEndSession) // OIDC RP-Initiated Logout (no auth required).
}
