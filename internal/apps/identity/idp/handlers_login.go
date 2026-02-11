// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package idp

import (
	"fmt"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
)

// handleLogin handles GET /login - Display login page.
func (s *Service) handleLogin(c *fiber.Ctx) error {
	// Extract request_id parameter.
	requestID := c.Query("request_id")

	if requestID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "request_id is required",
		})
	}

	// Render HTML login page with request_id parameter.
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)

	return s.templates.ExecuteTemplate(c.Response().BodyWriter(), "login.html", fiber.Map{
		"RequestID": requestID,
		"Error":     "",
	})
}

// handleLoginSubmit handles POST /login - Process login form.
func (s *Service) handleLoginSubmit(c *fiber.Ctx) error {
	// Extract credentials and request_id.
	username := c.FormValue("username")
	password := c.FormValue("password")
	requestIDStr := c.FormValue("request_id")

	// Validate required parameters.
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "username is required",
		})
	}

	if password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "password is required",
		})
	}

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

	// Use the default username/password authentication profile.
	profile, exists := s.authProfiles.Get("username_password")
	if !exists {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Authentication profile not available",
		})
	}

	// Authenticate user.
	user, err := profile.Authenticate(ctx, map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorAccessDenied,
			"error_description": "Invalid username or password",
		})
	}

	// Update authorization request with user ID.
	authRequest.UserID = cryptoutilIdentityDomain.NullableUUID{
		UUID:  user.ID,
		Valid: true,
	}

	if err := authzReqRepo.Update(ctx, authRequest); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to update authorization request",
		})
	}

	// Create user session.
	active := true
	session := &cryptoutilIdentityDomain.Session{
		UserID:                user.ID,
		IPAddress:             c.IP(),
		UserAgent:             c.Get("User-Agent"),
		IssuedAt:              time.Now().UTC(),
		ExpiresAt:             time.Now().UTC().Add(s.config.Sessions.SessionLifetime),
		LastSeenAt:            time.Now().UTC(),
		Active:                &active,
		AuthenticationMethods: []string{"username_password"},
		AuthenticationTime:    time.Now().UTC(),
	}

	sessionRepo := s.repoFactory.SessionRepository()
	if err := sessionRepo.Create(ctx, session); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to create session",
		})
	}

	// Set session cookie.
	c.Cookie(&fiber.Cookie{
		Name:     s.config.Sessions.CookieName,
		Value:    session.SessionID,
		Expires:  session.ExpiresAt,
		HTTPOnly: s.config.Sessions.CookieHTTPOnly,
		Secure:   s.config.IDP.TLSEnabled,
		SameSite: s.config.Sessions.CookieSameSite,
	})

	// Redirect to consent page with request_id parameter.
	consentURL := fmt.Sprintf("/oidc/v1/consent?request_id=%s", requestID.String())

	return c.Redirect(consentURL, fiber.StatusFound)
}
