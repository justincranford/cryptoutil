package authz

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers all OAuth 2.1 authorization server routes.
func (s *Service) RegisterRoutes(app *fiber.App) {
	// OAuth 2.1 endpoints with /oauth2/v1 prefix.
	oauth := app.Group("/oauth2/v1")
	oauth.Get("/authorize", s.handleAuthorizeGET)
	oauth.Post("/authorize", s.handleAuthorizePOST)
	oauth.Post("/token", s.handleToken)
	oauth.Post("/introspect", s.handleIntrospect)
	oauth.Post("/revoke", s.handleRevoke)
}
