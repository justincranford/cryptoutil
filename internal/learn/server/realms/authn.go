// Copyright (c) 2025 Justin Cranford

// Package realms provides authentication and authorization handlers.
package realms

import (
	"time"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	cryptoutilLearnDomain "cryptoutil/internal/learn/domain"
	cryptoutilLearnRepo "cryptoutil/internal/learn/repository"
	cryptoutilLearnServerUtil "cryptoutil/internal/learn/server/util"
)

// AuthnHandler handles authentication operations (login, register).
type AuthnHandler struct {
	userRepo  *cryptoutilLearnRepo.UserRepository
	jwtSecret string
}

// NewAuthnHandler creates a new authentication handler.
func NewAuthnHandler(
	userRepo *cryptoutilLearnRepo.UserRepository,
	jwtSecret string,
) *AuthnHandler {
	return &AuthnHandler{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// HandleRegisterUser returns a Fiber handler for user registration.
func (h *AuthnHandler) HandleRegisterUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.Username == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Username and password are required",
			})
		}

		// Validate username format (alphanumeric, 3-20 characters).
		if !isValidUsername(req.Username) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Username must be alphanumeric and 3-20 characters",
			})
		}

		// Check if user already exists.
		existingUser, err := h.userRepo.FindByUsername(c.Context(), req.Username)
		if err == nil && existingUser != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Username already exists",
			})
		}

		// Hash password.
		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(req.Password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to process password",
			})
		}

		// Create user.
		user := &cryptoutilLearnDomain.User{
			ID:           googleUuid.New(),
			Username:     req.Username,
			PasswordHash: string(hashedPassword),
			CreatedAt:    time.Now().UTC(),
		}

		if err := h.userRepo.Create(c.Context(), user); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create user",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id":       user.ID.String(),
			"username": user.Username,
		})
	}
}

// HandleLoginUser returns a Fiber handler for user login.
func (h *AuthnHandler) HandleLoginUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.Username == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Username and password are required",
			})
		}

		// Get user.
		user, err := h.userRepo.FindByUsername(c.Context(), req.Username)
		if err != nil || user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		// Verify password.
		if err := bcrypt.CompareHashAndPassword(
			[]byte(user.PasswordHash),
			[]byte(req.Password),
		); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		// Generate JWT.
		token, expiresAt, err := cryptoutilLearnServerUtil.GenerateJWT(
			user.ID,
			user.Username,
			h.jwtSecret,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate token",
			})
		}

		return c.JSON(fiber.Map{
			"token":      token,
			"expires_at": expiresAt.Format(time.RFC3339),
		})
	}
}

// isValidUsername checks if username is alphanumeric and 3-20 characters.
func isValidUsername(username string) bool {
	const (
		minUsernameLength = 3
		maxUsernameLength = 32
	)
	if len(username) < minUsernameLength ||
		len(username) > maxUsernameLength {
		return false
	}

	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9')) {
			return false
		}
	}

	return true
}
