// Copyright (c) 2025 Justin Cranford
//

package server

import (
	"encoding/hex"
	"time"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilCrypto "cryptoutil/internal/learn/crypto"
	cryptoutilDomain "cryptoutil/internal/learn/domain"
)

// RegisterUserRequest represents the request to register a new user.
type RegisterUserRequest struct {
	Username string `json:"username"` // Username (3-50 characters).
	Password string `json:"password"` // Password (minimum 8 characters).
}

// RegisterUserResponse represents the response after successful registration.
type RegisterUserResponse struct {
	UserID     string `json:"user_id"`     // Created user ID.
	PublicKey  string `json:"public_key"`  // User's ECDH public key (hex-encoded).
	PrivateKey string `json:"private_key"` // User's ECDH private key (hex-encoded, for testing only).
}

// LoginUserRequest represents the request to login.
type LoginUserRequest struct {
	Username string `json:"username"` // Username.
	Password string `json:"password"` // Password.
}

// LoginUserResponse represents the response after successful login.
type LoginUserResponse struct {
	Token     string `json:"token"`      // JWT authentication token.
	ExpiresAt string `json:"expires_at"` // Token expiration (RFC3339).
}

// handleRegisterUser implements POST /users/register.
func (s *PublicServer) handleRegisterUser(c *fiber.Ctx) error {
	// Parse request.
	var req RegisterUserRequest
	if err := c.BodyParser(&req); err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate request.
	const (
		minUsernameLength = 3
		maxUsernameLength = 50
		minPasswordLength = 8
	)

	if len(req.Username) < minUsernameLength || len(req.Username) > maxUsernameLength {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username must be 3-50 characters",
		})
	}

	if len(req.Password) < minPasswordLength {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password must be at least 8 characters",
		})
	}

	// Check username uniqueness.
	existing, err := s.userRepo.FindByUsername(c.Context(), req.Username)
	if err == nil && existing != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "username already exists",
		})
	}

	// Generate ECDH key pair for message encryption.
	privateKey, publicKeyBytes, err := cryptoutilCrypto.GenerateECDHKeyPair()
	if err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate encryption keys",
		})
	}

	// Extract private key bytes.
	privateKeyBytes := privateKey.Bytes()

	// Hash password using PBKDF2-HMAC-SHA256.
	passwordHash, err := cryptoutilCrypto.HashPassword(req.Password)
	if err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to hash password",
		})
	}

	// Encode password hash as hex for string storage.
	passwordHashHex := hex.EncodeToString(passwordHash)

	// Create user.
	// NOTE: ECDH keys are ephemeral per-message in current implementation.
	// Phase 5 will implement 3-table design with JWE multi-recipient encryption.
	user := &cryptoutilDomain.User{
		ID:           googleUuid.New(),
		Username:     req.Username,
		PasswordHash: passwordHashHex,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// NOTE: ECDH key generation removed - ephemeral per-message (Phase 5 3-table design).
	_ = publicKeyBytes  // Will be removed in Phase 5.
	_ = privateKeyBytes // Will be removed in Phase 5.

	if err := s.userRepo.Create(c.Context(), user); err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create user",
		})
	}

	// Return response.
	//nolint:wrapcheck // Fiber framework error, wrapping not needed.
	return c.Status(fiber.StatusCreated).JSON(RegisterUserResponse{
		UserID:     user.ID.String(),
		PublicKey:  hex.EncodeToString(publicKeyBytes),
		PrivateKey: hex.EncodeToString(privateKeyBytes),
	})
}

// handleLoginUser implements POST /users/login.
func (s *PublicServer) handleLoginUser(c *fiber.Ctx) error {
	// Parse request.
	var req LoginUserRequest
	if err := c.BodyParser(&req); err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate request.
	if req.Username == "" || req.Password == "" {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username and password are required",
		})
	}

	// Find user by username.
	user, err := s.userRepo.FindByUsername(c.Context(), req.Username)
	if err != nil || user == nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	}

	// Decode hex-encoded password hash from database.
	storedPasswordHash, err := hex.DecodeString(user.PasswordHash)
	if err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to decode password hash",
		})
	}

	// Verify password.
	verified, err := cryptoutilCrypto.VerifyPassword(req.Password, storedPasswordHash)
	if err != nil || !verified {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	}

	// Generate JWT token.
	token, expiresAt, err := GenerateJWT(user.ID, user.Username, s.jwtSecret)
	if err != nil {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate authentication token",
		})
	}

	//nolint:wrapcheck // Fiber framework error, wrapping not needed.
	return c.Status(fiber.StatusOK).JSON(LoginUserResponse{
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}
