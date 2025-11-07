package idp

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers all OIDC identity provider routes.
func (s *Service) RegisterRoutes(app *fiber.App) {
	// OIDC IdP endpoints.
	app.Get("/login", s.handleLogin)
	app.Post("/login", s.handleLoginSubmit)
	app.Get("/consent", s.handleConsent)
	app.Post("/consent", s.handleConsentSubmit)
	app.Get("/userinfo", s.handleUserInfo)
	app.Post("/logout", s.handleLogout)
}
