//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
// Copyright (c) 2025 Justin Cranford
//
//

package idp

import (
	"github.com/gofiber/fiber/v2"
)

// handleUserInfo handles GET /userinfo - Return OIDC UserInfo claims.
func (s *Service) handleUserInfo(c *fiber.Ctx) error {
	// Extract access token from Authorization header.
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             "invalid_token",
			"error_description": "Missing access token",
		})
	}

	// TODO: Parse Bearer token.
	// TODO: Introspect/validate token.
	// TODO: Fetch user details from repository.
	// TODO: Map user claims to OIDC standard claims (sub, name, email, etc.).

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"sub":   "user123",
		"name":  "John Doe",
		"email": "john.doe@example.com",
	})
}
