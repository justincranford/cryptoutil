package idp

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers all OIDC identity provider routes.
func (s *Service) RegisterRoutes(app *fiber.App) {
	// OIDC IdP endpoints with /oidc/v1 prefix.
	oidc := app.Group("/oidc/v1")
	oidc.Get("/login", s.handleLogin)
	oidc.Post("/login", s.handleLoginSubmit)
	oidc.Get("/consent", s.handleConsent)
	oidc.Post("/consent", s.handleConsentSubmit)
	oidc.Get("/userinfo", s.handleUserInfo)
	oidc.Post("/logout", s.handleLogout)
}
