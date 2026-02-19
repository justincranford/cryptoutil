// Copyright (c) 2025 Justin Cranford
//

package realms

import (
	"fmt"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	googleUuid "github.com/google/uuid"
)

// ContextKeyUserID is the Fiber context key for storing the authenticated user's ID.
//
// Usage in handlers:
//
//	func protectedHandler(c *fiber.Ctx) error {
//	    userID := c.Locals(ContextKeyUserID).(googleUuid.UUID)
//	    // Use userID for authorization checks
//	}
const ContextKeyUserID = "user_id"

// JWTMiddleware validates JWT tokens and extracts the user ID into Fiber context.
//
// Workflow:
// 1. Extract "Authorization" header
// 2. Validate "Bearer <token>" format
// 3. Parse JWT token with HMAC-SHA256 signature verification
// 4. Validate token signature, expiration, and claims
// 5. Extract user ID from claims
// 6. Store user ID in Fiber context (c.Locals)
// 7. Call next handler
//
// Failure Conditions (returns 401 Unauthorized):
// - Missing Authorization header
// - Invalid Authorization header format (not "Bearer <token>")
// - Invalid JWT signature (secret mismatch)
// - Expired token (ExpiresAt < Now)
// - Invalid token structure (malformed claims)
// - Invalid user ID in claims (not a valid UUID)
//
// Security Notes:
// - ONLY validates HMAC-SHA256 signed tokens (rejects other algorithms)
// - Short expiration (15 min) mitigates token theft risk
// - User ID stored in context enables authorization checks in handlers
//
// Example Usage:
//
//	// Protect routes with JWT middleware
//	app.Get("/service/api/v1/messages", realms.JWTMiddleware(jwtSecret), handleListMessages)
//	app.Post("/service/api/v1/messages", realms.JWTMiddleware(jwtSecret), handleSendMessage)
//
//	// Handler accesses authenticated user ID
//	func handleListMessages(c *fiber.Ctx) error {
//	    userID := c.Locals(ContextKeyUserID).(googleUuid.UUID)
//	    messages, err := messageRepo.FindByRecipient(c.Context(), userID)
//	    // ...
//	}
func JWTMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract Authorization header.
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		// Extract token from "Bearer <token>" format.
		if !strings.HasPrefix(authHeader, cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix) {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format (expected: Bearer <token>)",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix)
		// Parse and validate token.
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			// Verify signing method is HMAC-SHA256 (FIPS-approved).
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(secret), nil
		})
		if err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fmt.Sprintf("invalid token: %v", err),
			})
		}

		if !token.Valid {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "token is not valid",
			})
		}

		// Parse user ID from claims.
		userID, err := googleUuid.Parse(claims.UserID)
		if err != nil {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid user ID in token",
			})
		}

		// Store user ID in context for handlers.
		c.Locals(ContextKeyUserID, userID)

		return c.Next()
	}
}
