package idp

import (
	"github.com/gofiber/fiber/v2"

	cryptoutilIdentityApperr "cryptoutil/internal/identity/apperr"
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
	userRepo := s.repoFactory.UserRepository()

	// Authenticate user.
	user, err := userRepo.GetByUsername(ctx, username)
	if err != nil {
		appErr := cryptoutilIdentityApperr.ErrUserNotFound

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorAccessDenied,
			"error_description": "Invalid username or password",
		})
	}

	// TODO: Validate password hash.
	_ = user
	_ = password

	// TODO: Create user session.
	// TODO: Redirect to consent page or authorization callback.

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"user_id": user.ID.String(),
	})
}
