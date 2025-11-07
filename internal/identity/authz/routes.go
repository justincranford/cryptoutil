package authz

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers all OAuth 2.1 authorization server routes.
func (s *Service) RegisterRoutes(app *fiber.App) {
	// OAuth 2.1 endpoints.
	app.Get("/authorize", s.handleAuthorizeGET)
	app.Post("/authorize", s.handleAuthorizePOST)
	app.Post("/token", s.handleToken)
	app.Post("/introspect", s.handleIntrospect)
	app.Post("/revoke", s.handleRevoke)
}
