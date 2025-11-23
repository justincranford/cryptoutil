// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package idp

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// handleConsent handles GET /consent - Display consent page.
func (s *Service) handleConsent(c *fiber.Ctx) error {
	// Extract request_id parameter.
	requestIDStr := c.Query("request_id")

	if requestIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "request_id is required",
		})
	}

	ctx := c.Context()

	// Parse request_id.
	requestID, err := googleUuid.Parse(requestIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "Invalid request_id format",
		})
	}

	// Retrieve authorization request from database.
	authzReqRepo := s.repoFactory.AuthorizationRequestRepository()

	authRequest, err := authzReqRepo.GetByID(ctx, requestID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "Authorization request not found or expired",
		})
	}

	// Validate request not expired.
	if authRequest.IsExpired() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "Authorization request has expired",
		})
	}

	// Validate user ID was set during login.
	if !authRequest.UserID.Valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "User not authenticated",
		})
	}

	// Retrieve client details.
	clientRepo := s.repoFactory.ClientRepository()

	_, err = clientRepo.GetByClientID(ctx, authRequest.ClientID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "Client not found",
		})
	}

	// Check if user has already consented to this client/scope combination.
	consentRepo := s.repoFactory.ConsentDecisionRepository()

	existingConsent, err := consentRepo.GetByUserClientScope(ctx, authRequest.UserID.UUID, authRequest.ClientID, authRequest.Scope)
	if err == nil && existingConsent != nil && !existingConsent.IsRevoked() && !existingConsent.IsExpired() {
		// Consent exists and is valid - skip consent page, generate code immediately.
		authCode := generateRandomString(cryptoutilIdentityMagic.DefaultAuthCodeLength)
		authRequest.Code = authCode
		authRequest.ExpiresAt = time.Now().Add(cryptoutilIdentityMagic.DefaultCodeLifetime)
		authRequest.ConsentGranted = true

		if err := authzReqRepo.Update(ctx, authRequest); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":             cryptoutilIdentityMagic.ErrorServerError,
				"error_description": "Failed to update authorization request with code",
			})
		}

		// Build redirect URI with authorization code and state.
		redirectURI := fmt.Sprintf("%s?code=%s&state=%s", authRequest.RedirectURI, authCode, authRequest.State)

		return c.Redirect(redirectURI, fiber.StatusFound)
	}

	// TODO: Render HTML consent page with request_id, client name, scopes.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":    "Consent page",
		"request_id": requestID.String(),
		"client_id":  authRequest.ClientID,
		"scope":      authRequest.Scope,
	})
}

// handleConsentSubmit handles POST /consent - Process consent approval.
func (s *Service) handleConsentSubmit(c *fiber.Ctx) error {
	// Extract parameters.
	requestIDStr := c.FormValue("request_id")
	approved := c.FormValue("approved")

	// Validate required parameters.
	if requestIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "request_id is required",
		})
	}

	// Validate approval decision.
	if approved != "true" {
		// User denied consent - redirect back to client with error.
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorAccessDenied,
			"error_description": "User denied consent",
		})
	}

	ctx := c.Context()

	// Parse request_id.
	requestID, err := googleUuid.Parse(requestIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "Invalid request_id format",
		})
	}

	// Retrieve authorization request from database.
	authzReqRepo := s.repoFactory.AuthorizationRequestRepository()

	authRequest, err := authzReqRepo.GetByID(ctx, requestID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "Authorization request not found or expired",
		})
	}

	// Validate request not expired.
	if authRequest.IsExpired() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "Authorization request has expired",
		})
	}

	// Validate user ID was set during login.
	if !authRequest.UserID.Valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "User not authenticated",
		})
	}

	// Store consent decision in database.
	consentRepo := s.repoFactory.ConsentDecisionRepository()
	consentDecision := &cryptoutilIdentityDomain.ConsentDecision{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		UserID:    authRequest.UserID.UUID,
		ClientID:  authRequest.ClientID,
		Scope:     authRequest.Scope,
		GrantedAt: time.Now(),
		ExpiresAt: time.Now().Add(cryptoutilIdentityMagic.DefaultRefreshTokenLifetime), // Consent lasts as long as refresh token.
	}

	if err := consentRepo.Create(ctx, consentDecision); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to store consent decision",
		})
	}

	// Generate authorization code.
	authCode := generateRandomString(cryptoutilIdentityMagic.DefaultAuthCodeLength)
	authRequest.Code = authCode
	authRequest.ExpiresAt = time.Now().Add(cryptoutilIdentityMagic.DefaultCodeLifetime)
	authRequest.ConsentGranted = true

	if err := authzReqRepo.Update(ctx, authRequest); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to update authorization request with code",
		})
	}

	// Build redirect URI with authorization code and state.
	redirectURI := fmt.Sprintf("%s?code=%s&state=%s", authRequest.RedirectURI, authCode, authRequest.State)

	return c.Redirect(redirectURI, fiber.StatusFound)
}
