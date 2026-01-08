// Copyright (c) 2025 Justin Cranford
//

package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"

	"cryptoutil/internal/apps/cipher/im/server/businesslogic"
	cryptoutilAppErr "cryptoutil/internal/shared/apperr"
)

// SessionMiddleware validates session tokens for browser or service requests.
// Extracts token from Authorization header (Bearer format) and validates using SessionManager.
// Sets validated session information in context for downstream handlers.
func SessionMiddleware(sessionManager *businesslogic.SessionManagerService, isBrowser bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			summary := "Missing Authorization header"
			return cryptoutilAppErr.NewHTTP401Unauthorized(&summary, nil)
		}

		// Parse Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			summary := "Invalid Authorization header format (expected: Bearer <token>)"
			return cryptoutilAppErr.NewHTTP401Unauthorized(&summary, nil)
		}

		token := parts[1]
		if token == "" {
			summary := "Empty token in Authorization header"
			return cryptoutilAppErr.NewHTTP401Unauthorized(&summary, nil)
		}

		// Validate token using SessionManager
		ctx := context.Background()

		if isBrowser {
			// Validate browser session
			session, validateErr := sessionManager.ValidateBrowserSession(ctx, token)
			if validateErr != nil {
				summary := "Invalid or expired browser session token"
				return cryptoutilAppErr.NewHTTP401Unauthorized(&summary, validateErr)
			}

			// Store session in context for downstream handlers
			c.Locals("session", session)
			if session.UserID != nil {
				c.Locals("user_id", *session.UserID)
			}
			if session.Realm != nil {
				c.Locals("realm", *session.Realm)
			}
		} else {
			// Validate service session
			session, validateErr := sessionManager.ValidateServiceSession(ctx, token)
			if validateErr != nil {
				summary := "Invalid or expired service session token"
				return cryptoutilAppErr.NewHTTP401Unauthorized(&summary, validateErr)
			}

			// Store session in context for downstream handlers
			c.Locals("session", session)
			if session.ClientID != nil {
				c.Locals("client_id", *session.ClientID)
			}
			if session.Realm != nil {
				c.Locals("realm", *session.Realm)
			}
		}

		return c.Next()
	}
}

// BrowserSessionMiddleware creates middleware for browser session validation.
func BrowserSessionMiddleware(sessionManager *businesslogic.SessionManagerService) fiber.Handler {
	return SessionMiddleware(sessionManager, true)
}

// ServiceSessionMiddleware creates middleware for service session validation.
func ServiceSessionMiddleware(sessionManager *businesslogic.SessionManagerService) fiber.Handler {
	return SessionMiddleware(sessionManager, false)
}
