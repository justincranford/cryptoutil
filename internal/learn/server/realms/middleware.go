// Copyright (c) 2025 Justin Cranford
//
//

package realms

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	googleUuid "github.com/google/uuid"

	cryptoutilLearnServerUtil "cryptoutil/internal/learn/server/util"
)

const (
	// ContextKeyUserID is the context key for storing user ID from JWT.
	ContextKeyUserID = cryptoutilLearnServerUtil.ContextKeyUserID
)

// Claims represents JWT claims for learn-im authentication.
type Claims = cryptoutilLearnServerUtil.Claims

// JWTMiddleware validates JWT tokens and extracts user ID.
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
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			//nolint:wrapcheck // Fiber framework error, wrapping not needed.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format (expected: Bearer <token>)",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)

		// Parse and validate token.
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			// Verify signing method.
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
