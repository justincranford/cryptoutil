// Copyright (c) 2025 Justin Cranford

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	"errors"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// GenerateRecoveryCodesRequest represents request to generate recovery codes.
type GenerateRecoveryCodesRequest struct {
	UserID string `json:"user_id"`
}

// GenerateRecoveryCodesResponse contains generated recovery codes.
type GenerateRecoveryCodesResponse struct {
	Codes     []string `json:"codes"`
	ExpiresAt string   `json:"expires_at"`
}

// RecoveryCodeCountResponse contains count of remaining recovery codes.
type RecoveryCodeCountResponse struct {
	Remaining int64 `json:"remaining"`
	Total     int64 `json:"total"`
}

// VerifyRecoveryCodeRequest represents request to verify a recovery code.
type VerifyRecoveryCodeRequest struct {
	Code string `json:"code"`
}

// VerifyRecoveryCodeResponse indicates verification result.
type VerifyRecoveryCodeResponse struct {
	Verified bool `json:"verified"`
}

// handleGenerateRecoveryCodes handles POST /oidc/v1/mfa/recovery-codes/generate.
// Generates a new batch of recovery codes for a user.
// WARNING: Codes shown ONCE - user must save them immediately.
func (s *Service) handleGenerateRecoveryCodes(c *fiber.Ctx) error {
	var req GenerateRecoveryCodesRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Invalid request body",
		})
	}

	// Validate user_id.
	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Missing user_id parameter",
		})
	}

	userID, err := googleUuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Invalid user_id format (must be UUID)",
		})
	}

	// Verify user exists.
	_, err = s.repoFactory.UserRepository().GetByID(c.Context(), userID)
	if err != nil {
		if errors.Is(err, cryptoutilIdentityAppErr.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "user_not_found",
				"error_description":               "User not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Internal server error",
		})
	}

	// Generate recovery codes.
	service := cryptoutilIdentityMfa.NewRecoveryCodeService(s.repoFactory.RecoveryCodeRepository())

	codes, err := service.GenerateForUser(c.Context(), userID, cryptoutilSharedMagic.DefaultRecoveryCodeCount)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to generate recovery codes",
		})
	}

	// Calculate expiration timestamp.
	expiresAt := time.Now().UTC().Add(cryptoutilSharedMagic.DefaultRecoveryCodeLifetime)

	return c.Status(fiber.StatusCreated).JSON(GenerateRecoveryCodesResponse{
		Codes:     codes,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

// handleGetRecoveryCodeCount handles GET /oidc/v1/mfa/recovery-codes/count.
// Returns count of unused, unexpired recovery codes for a user.
func (s *Service) handleGetRecoveryCodeCount(c *fiber.Ctx) error {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Missing user_id query parameter",
		})
	}

	userID, err := googleUuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Invalid user_id format (must be UUID)",
		})
	}

	// Get remaining count.
	service := cryptoutilIdentityMfa.NewRecoveryCodeService(s.repoFactory.RecoveryCodeRepository())

	remaining, err := service.GetRemainingCount(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to get recovery code count",
		})
	}

	return c.Status(fiber.StatusOK).JSON(RecoveryCodeCountResponse{
		Remaining: remaining,
		Total:     int64(cryptoutilSharedMagic.DefaultRecoveryCodeCount),
	})
}

// handleRegenerateRecoveryCodes handles POST /oidc/v1/mfa/recovery-codes/regenerate.
// Deletes all existing recovery codes and generates a new batch.
func (s *Service) handleRegenerateRecoveryCodes(c *fiber.Ctx) error {
	var req GenerateRecoveryCodesRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Invalid request body",
		})
	}

	// Validate user_id.
	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Missing user_id parameter",
		})
	}

	userID, err := googleUuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Invalid user_id format (must be UUID)",
		})
	}

	// Verify user exists.
	_, err = s.repoFactory.UserRepository().GetByID(c.Context(), userID)
	if err != nil {
		if errors.Is(err, cryptoutilIdentityAppErr.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "user_not_found",
				"error_description":               "User not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Internal server error",
		})
	}

	// Regenerate recovery codes.
	service := cryptoutilIdentityMfa.NewRecoveryCodeService(s.repoFactory.RecoveryCodeRepository())

	codes, err := service.RegenerateForUser(c.Context(), userID, cryptoutilSharedMagic.DefaultRecoveryCodeCount)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to regenerate recovery codes",
		})
	}

	// Calculate expiration timestamp.
	expiresAt := time.Now().UTC().Add(cryptoutilSharedMagic.DefaultRecoveryCodeLifetime)

	return c.Status(fiber.StatusCreated).JSON(GenerateRecoveryCodesResponse{
		Codes:     codes,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

// handleVerifyRecoveryCode handles POST /oidc/v1/mfa/verify-recovery-code.
// Verifies a recovery code and marks it as used on success.
func (s *Service) handleVerifyRecoveryCode(c *fiber.Ctx) error {
	var req VerifyRecoveryCodeRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Invalid request body",
		})
	}

	// Validate code.
	if req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Missing code parameter",
		})
	}

	// Get user_id from session/token context.
	// For now, use hardcoded user_id (integration with login flow deferred).
	userIDStr := c.Get("X-User-ID")
	if userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: "unauthorized",
			"error_description":               "Missing user authentication",
		})
	}

	userID, err := googleUuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Invalid user_id format",
		})
	}

	// Verify recovery code.
	service := cryptoutilIdentityMfa.NewRecoveryCodeService(s.repoFactory.RecoveryCodeRepository())

	err = service.Verify(c.Context(), userID, req.Code)
	if err != nil {
		if errors.Is(err, cryptoutilIdentityAppErr.ErrRecoveryCodeNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "invalid_code",
				"error_description":               "Invalid or expired recovery code",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to verify recovery code",
		})
	}

	return c.Status(fiber.StatusOK).JSON(VerifyRecoveryCodeResponse{
		Verified: true,
	})
}
