// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package idp

import (
	"github.com/gofiber/fiber/v2"

	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// handleLogout handles POST /logout - Terminate user session.
func (s *Service) handleLogout(c *fiber.Ctx) error {
	ctx := c.Context()

	// Extract session cookie.
	sessionID := c.Cookies(s.config.Sessions.CookieName)

	if sessionID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "No active session found",
		})
	}

	// Retrieve session from database.
	sessionRepo := s.repoFactory.SessionRepository()

	session, err := sessionRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "Session not found",
		})
	}

	// Delete session from database.
	if err := sessionRepo.Delete(ctx, session.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to delete session",
		})
	}

	// Clear session cookie.
	c.ClearCookie(s.config.Sessions.CookieName)

	// Return success response.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}
