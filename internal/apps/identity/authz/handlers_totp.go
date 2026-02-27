// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"strings"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
)

// EnrollTOTPRequest represents request to enroll in TOTP MFA.
type EnrollTOTPRequest struct {
	UserID      string `json:"user_id"`
	Issuer      string `json:"issuer"`
	AccountName string `json:"account_name"`
}

// EnrollTOTPResponse contains TOTP enrollment data.
type EnrollTOTPResponse struct {
	SecretID    string   `json:"secret_id"`
	QRCodeURI   string   `json:"qr_code_uri"`
	BackupCodes []string `json:"backup_codes"`
}

// VerifyTOTPRequest represents request to verify a TOTP code.
type VerifyTOTPRequest struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

// VerifyTOTPResponse indicates verification result.
type VerifyTOTPResponse struct {
	Verified bool `json:"verified"`
}

// MFAStepUpRequest represents request to check MFA step-up requirement.
type MFAStepUpRequest struct {
	UserID string `json:"user_id"`
}

// MFAStepUpResponse indicates if MFA step-up is required.
type MFAStepUpResponse struct {
	Required bool `json:"required"`
}

// VerifyBackupCodeRequest represents request to verify a backup code.
type VerifyBackupCodeRequest struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

// VerifyBackupCodeResponse indicates verification result.
type VerifyBackupCodeResponse struct {
	Verified bool `json:"verified"`
}

// handleEnrollTOTP handles POST /oidc/v1/mfa/totp/enroll.
// Enrolls a user in TOTP MFA and returns QR code URI plus backup codes.
// WARNING: Secret and backup codes shown ONCE - user must save them immediately.
func (s *Service) handleEnrollTOTP(c *fiber.Ctx) error {
	var req EnrollTOTPRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Invalid request body",
		})
	}

	// Validate required fields.
	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Missing user_id parameter",
		})
	}

	if req.Issuer == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Missing issuer parameter",
		})
	}

	if req.AccountName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Missing account_name parameter",
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

	// Enroll in TOTP.
	db := s.repoFactory.DB()
	service := cryptoutilIdentityMfa.NewTOTPService(db)

	totpSecret, qrURI, backupCodes, err := service.EnrollTOTP(c.Context(), userID, req.Issuer, req.AccountName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to enroll in TOTP",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(EnrollTOTPResponse{
		SecretID:    totpSecret.ID.String(),
		QRCodeURI:   qrURI,
		BackupCodes: backupCodes,
	})
}

// handleVerifyTOTP handles POST /oidc/v1/mfa/totp/verify.
// Verifies a TOTP code for a user.
func (s *Service) handleVerifyTOTP(c *fiber.Ctx) error {
	var req VerifyTOTPRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Invalid request body",
		})
	}

	// Validate required fields.
	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Missing user_id parameter",
		})
	}

	if req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Missing code parameter",
		})
	}

	userID, err := googleUuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Invalid user_id format (must be UUID)",
		})
	}

	// Verify TOTP code.
	db := s.repoFactory.DB()
	service := cryptoutilIdentityMfa.NewTOTPService(db)

	err = service.VerifyTOTP(c.Context(), userID, req.Code)
	if err != nil {
		// Check for specific errors.
		if errors.Is(err, cryptoutilIdentityAppErr.ErrTOTPSecretNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "totp_not_enrolled",
				"error_description":               "User not enrolled in TOTP MFA",
			})
		}

		if errors.Is(err, cryptoutilIdentityAppErr.ErrTOTPAccountLocked) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "account_locked",
				"error_description":               "Account temporarily locked due to too many failed attempts",
			})
		}

		// Invalid code returns false, not error.
		return c.Status(fiber.StatusOK).JSON(VerifyTOTPResponse{
			Verified: false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(VerifyTOTPResponse{
		Verified: true,
	})
}

// handleCheckMFAStepUp handles GET /oidc/v1/mfa/totp/step-up.
// Checks if MFA step-up is required (>30 minutes since last verification).
func (s *Service) handleCheckMFAStepUp(c *fiber.Ctx) error {
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

	// Check MFA step-up requirement.
	db := s.repoFactory.DB()
	service := cryptoutilIdentityMfa.NewTOTPService(db)

	required, err := service.RequiresMFAStepUp(c.Context(), userID)
	if err != nil {
		if errors.Is(err, cryptoutilIdentityAppErr.ErrTOTPSecretNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "totp_not_enrolled",
				"error_description":               "User not enrolled in TOTP MFA",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to check MFA step-up requirement",
		})
	}

	return c.Status(fiber.StatusOK).JSON(MFAStepUpResponse{
		Required: required,
	})
}

// handleGenerateTOTPBackupCodes handles POST /oidc/v1/mfa/totp/backup-codes/generate.
// Generates new backup codes for TOTP MFA.
// WARNING: Codes shown ONCE - user must save them immediately.
func (s *Service) handleGenerateTOTPBackupCodes(c *fiber.Ctx) error {
	var req struct {
		UserID string `json:"user_id"`
	}

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

	// Generate backup codes.
	db := s.repoFactory.DB()
	service := cryptoutilIdentityMfa.NewTOTPService(db)

	codes, err := service.GenerateBackupCodes(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to generate backup codes",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"backup_codes": codes,
	})
}

// handleVerifyTOTPBackupCode handles POST /oidc/v1/mfa/totp/backup-codes/verify.
// Verifies a backup code for TOTP MFA.
func (s *Service) handleVerifyTOTPBackupCode(c *fiber.Ctx) error {
	var req VerifyBackupCodeRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Invalid request body",
		})
	}

	// Validate required fields.
	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Missing user_id parameter",
		})
	}

	if req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Missing code parameter",
		})
	}

	userID, err := googleUuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Invalid user_id format (must be UUID)",
		})
	}

	// Verify backup code.
	db := s.repoFactory.DB()
	service := cryptoutilIdentityMfa.NewTOTPService(db)

	err = service.VerifyBackupCode(c.Context(), userID, req.Code)
	if err != nil {
		// Check for specific errors.
		if errors.Is(err, cryptoutilIdentityAppErr.ErrBackupCodeNotFound) {
			return c.Status(fiber.StatusOK).JSON(VerifyBackupCodeResponse{
				Verified: false,
			})
		}

		// Generic errors (invalid code, already used, etc.) return verified=false, not 500.
		// Only database/internal errors should return 500.
		errMsg := err.Error()
		if strings.Contains(errMsg, "invalid backup code") || strings.Contains(errMsg, "failed to get backup codes") {
			return c.Status(fiber.StatusOK).JSON(VerifyBackupCodeResponse{
				Verified: false,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to verify backup code",
		})
	}

	return c.Status(fiber.StatusOK).JSON(VerifyBackupCodeResponse{
		Verified: true,
	})
}
