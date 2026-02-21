// Copyright (c) 2025 Justin Cranford
//

// Package middleware provides HTTP middleware for session management and authentication.
package middleware

import (
	"context"
	"strings"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

// ContextKeySession is the Fiber context key for storing session information.
const ContextKeySession = "session"

// ContextKeyUserID is the Fiber context key for storing the authenticated user ID.
const ContextKeyUserID = "user_id"

// ContextKeyClientID is the Fiber context key for storing the authenticated client ID.
const ContextKeyClientID = "client_id"

// ContextKeyTenantID is the Fiber context key for storing the tenant ID.
const ContextKeyTenantID = "tenant_id"

// ContextKeyRealmID is the Fiber context key for storing the realm ID.
const ContextKeyRealmID = "realm_id"

// SessionValidator defines the interface for session validation services.
// Implementations must provide ValidateBrowserSession and ValidateServiceSession methods.
type SessionValidator interface {
	ValidateBrowserSession(ctx context.Context, token string) (*cryptoutilAppsTemplateServiceServerRepository.BrowserSession, error)
	ValidateServiceSession(ctx context.Context, token string) (*cryptoutilAppsTemplateServiceServerRepository.ServiceSession, error)
}

// SessionMiddleware validates session tokens for browser or service requests.
// Extracts token from Authorization header (Bearer format) and validates using SessionValidator.
// Sets validated session information in context for downstream handlers.
//
// Context keys set:
// - ContextKeySession: BrowserSession or ServiceSession struct
// - ContextKeyUserID: googleUuid.UUID (if BrowserSession.UserID != nil, parsed as UUID)
// - ContextKeyClientID: string (if ServiceSession.ClientID != nil)
// - ContextKeyTenantID: googleUuid.UUID (from session.TenantID)
// - ContextKeyRealmID: googleUuid.UUID (from session.RealmID)
//
// Example Usage:
//
//	sessionValidator := businesslogic.NewSessionManager(...)
//	app.Get("/browser/api/v1/messages", middleware.BrowserSessionMiddleware(sessionValidator), handleMessages)
//	app.Get("/service/api/v1/messages", middleware.ServiceSessionMiddleware(sessionValidator), handleMessages)
// sessionMiddlewareStringsSplitNFn allows overriding strings.SplitN for testing.
var sessionMiddlewareStringsSplitNFn = strings.SplitN

func SessionMiddleware(sessionValidator SessionValidator, isBrowser bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			summary := "Missing Authorization header"

			return cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, nil)
		}

		// Parse Bearer token format
		parts := sessionMiddlewareStringsSplitNFn(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			summary := "Invalid Authorization header format (expected: Bearer <token>)"

			return cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, nil)
		}

		token := parts[1]
		if token == "" {
			summary := "Empty token in Authorization header"

			return cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, nil)
		}

		// Validate token using SessionValidator
		ctx := c.Context()

		if isBrowser {
			// Validate browser session
			session, validateErr := sessionValidator.ValidateBrowserSession(ctx, token)
			if validateErr != nil {
				summary := "Invalid or expired browser session token"

				return cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, validateErr)
			}

			// Store session in context for downstream handlers
			c.Locals(ContextKeySession, session)

			if session.UserID != nil {
				// Parse the UserID string as a UUID for user_id
				if userID, parseErr := googleUuid.Parse(*session.UserID); parseErr == nil {
					c.Locals(ContextKeyUserID, userID)
				}
			}

			c.Locals(ContextKeyTenantID, session.TenantID)
			c.Locals(ContextKeyRealmID, session.RealmID)
		} else {
			// Validate service session
			session, validateErr := sessionValidator.ValidateServiceSession(ctx, token)
			if validateErr != nil {
				summary := "Invalid or expired service session token"

				return cryptoutilSharedApperr.NewHTTP401Unauthorized(&summary, validateErr)
			}

			// Store session in context for downstream handlers
			c.Locals(ContextKeySession, session)

			if session.ClientID != nil {
				// For cipher-im, ClientID actually contains the UserID
				// because we're using service sessions for user authentication
				c.Locals(ContextKeyClientID, *session.ClientID)

				// Parse the ClientID string as a UUID for user_id
				if userID, parseErr := googleUuid.Parse(*session.ClientID); parseErr == nil {
					c.Locals(ContextKeyUserID, userID)
				}
			}

			c.Locals(ContextKeyTenantID, session.TenantID)
			c.Locals(ContextKeyRealmID, session.RealmID)
		}

		return c.Next()
	}
}

// BrowserSessionMiddleware validates browser session tokens.
// Convenience wrapper for SessionMiddleware with isBrowser=true.
func BrowserSessionMiddleware(sessionValidator SessionValidator) fiber.Handler {
	return SessionMiddleware(sessionValidator, true)
}

// ServiceSessionMiddleware validates service session tokens.
// Convenience wrapper for SessionMiddleware with isBrowser=false.
func ServiceSessionMiddleware(sessionValidator SessionValidator) fiber.Handler {
	return SessionMiddleware(sessionValidator, false)
}
