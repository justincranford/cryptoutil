package idp

import (
	"time"

	"github.com/gofiber/fiber/v2"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// handleLogin handles GET /login - Display login page.
func (s *Service) handleLogin(c *fiber.Ctx) error {
	// Extract parameters.
	clientID := c.Query(cryptoutilIdentityMagic.ParamClientID)
	redirectURI := c.Query(cryptoutilIdentityMagic.ParamRedirectURI)
	state := c.Query(cryptoutilIdentityMagic.ParamState)
	scope := c.Query(cryptoutilIdentityMagic.ParamScope)

	// TODO: Render login page with parameters.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Login page",
		"client_id":    clientID,
		"redirect_uri": redirectURI,
		"state":        state,
		"scope":        scope,
	})
}

// handleLoginSubmit handles POST /login - Process login form.
func (s *Service) handleLoginSubmit(c *fiber.Ctx) error {
	// Extract credentials.
	username := c.FormValue("username")
	password := c.FormValue("password")

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

	ctx := c.Context()

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

	// Create user session.
	session := &cryptoutilIdentityDomain.Session{
		UserID:               user.ID,
		IPAddress:            c.IP(),
		UserAgent:            c.Get("User-Agent"),
		IssuedAt:             time.Now(),
		ExpiresAt:            time.Now().Add(s.config.Sessions.SessionLifetime),
		LastSeenAt:           time.Now(),
		Active:               true,
		AuthenticationMethods: []string{"username_password"},
		AuthenticationTime:   time.Now(),
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

	// TODO: Redirect to consent page or authorization callback based on original request.
	// For now, return success.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":    "Login successful",
		"user_id":    user.ID.String(),
		"session_id": session.SessionID,
	})
}
