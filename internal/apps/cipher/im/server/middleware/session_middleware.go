// Copyright (c) 2025 Justin Cranford
//

package middleware

import (
	"strings"

	googleUuid "github.com/google/uuid"
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
		ctx := c.Context()

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
				// Parse the UserID string as a UUID for user_id
				if userID, parseErr := googleUuid.Parse(*session.UserID); parseErr == nil {
					c.Locals("user_id", userID)
				}
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
				// For cipher-im, ClientID actually contains the UserID
				// because we're using service sessions for user authentication
				c.Locals("client_id", *session.ClientID)
				
				// Parse the ClientID string as a UUID for user_id
				if userID, parseErr := googleUuid.Parse(*session.ClientID); parseErr == nil {
					c.Locals("user_id", userID)
				}
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
