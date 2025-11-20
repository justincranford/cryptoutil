// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package idp

import (
	"github.com/gofiber/fiber/v2"

	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// handleConsent handles GET /consent - Display consent page.
func (s *Service) handleConsent(c *fiber.Ctx) error {
	// Extract parameters.
	clientID := c.Query(cryptoutilIdentityMagic.ParamClientID)
	scope := c.Query(cryptoutilIdentityMagic.ParamScope)
	state := c.Query(cryptoutilIdentityMagic.ParamState)

	// TODO: Fetch client details.
	// TODO: Render consent page with scopes and client information.

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":   "Consent page",
		"client_id": clientID,
		"scope":     scope,
		"state":     state,
	})
}

// handleConsentSubmit handles POST /consent - Process consent approval.
func (s *Service) handleConsentSubmit(c *fiber.Ctx) error {
	// Extract parameters.
	approved := c.FormValue("approved")
	scope := c.FormValue(cryptoutilIdentityMagic.ParamScope)

	// Validate approval.
	if approved != "true" {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorAccessDenied,
			"error_description": "User denied consent",
		})
	}

	// TODO: Store consent decision.
	// TODO: Generate authorization code.
	// TODO: Redirect to authorization callback.

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Consent approved",
		"scope":   scope,
	})
}
