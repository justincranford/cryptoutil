// Copyright (c) 2025-2026 Justin Cranford.
//
//

//nolint:wrapcheck,revive // Integration test with realistic error propagation
package apis_test

import (
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
)

// TestTOTPEnrollment covers successful enrollment and required-field validation.
func TestTOTPEnrollment(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		issuer               string
		accountName          string
		expectedStatus       int
		wantErrorDescription string
	}{
		{name: "valid enrollment", issuer: testTOTPIssuer, accountName: testTOTPAccountName, expectedStatus: fiber.StatusCreated},
		{name: "missing issuer", accountName: testTOTPAccountName, expectedStatus: fiber.StatusBadRequest, wantErrorDescription: "Missing issuer parameter"},
		{name: "missing account name", issuer: testTOTPIssuer, expectedStatus: fiber.StatusBadRequest, wantErrorDescription: "Missing account_name parameter"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, app := newTOTPIntegrationTestApp(t)
			userID := googleUuid.Must(googleUuid.NewV7())
			createTestUser(t, db, userID, testTOTPAccountName)
			response := enrollTOTP(t, app, userID, tc.issuer, tc.accountName, tc.expectedStatus)

			if tc.wantErrorDescription != "" {
				require.Equal(t, cryptoutilSharedMagic.ErrorInvalidRequest, response[cryptoutilSharedMagic.StringError])
				require.Equal(t, tc.wantErrorDescription, response["error_description"])

				var secretCount, backupCodeCount int64
				require.NoError(t, db.Model(&cryptoutilIdentityMfa.TOTPSecret{}).Where("user_id = ?", userID).Count(&secretCount).Error)
				require.NoError(t, db.Model(&cryptoutilIdentityMfa.BackupCode{}).Where("user_id = ?", userID).Count(&backupCodeCount).Error)
				require.Zero(t, secretCount)
				require.Zero(t, backupCodeCount)

				return
			}

			require.Contains(t, response, "secret_id")
			require.Contains(t, response, "qr_code_uri")
			require.Contains(t, response, "backup_codes")
			qrCodeURI, ok := response["qr_code_uri"].(string)
			require.True(t, ok, "qr_code_uri should be string")
			require.Contains(t, qrCodeURI, "otpauth://totp/")
			require.Contains(t, qrCodeURI, tc.issuer)
			require.Contains(t, qrCodeURI, tc.accountName)

			backupCodes, ok := response["backup_codes"].([]any)
			require.True(t, ok, "backup_codes should be array")
			require.Len(t, backupCodes, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)

			var secret cryptoutilIdentityMfa.TOTPSecret
			require.NoError(t, db.Where("user_id = ?", userID).First(&secret).Error)
			require.Equal(t, userID, secret.UserID)
			require.False(t, secret.Verified)
			require.Zero(t, secret.FailedAttempts)

			var storedCodes []cryptoutilIdentityMfa.BackupCode
			require.NoError(t, db.Where("user_id = ?", userID).Find(&storedCodes).Error)
			require.Len(t, storedCodes, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)

			for _, code := range storedCodes {
				require.False(t, code.Used)
			}
		})
	}
}

// TestTOTPVerification exercises successful and unsuccessful code verification.
func TestTOTPVerification(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		valid              bool
		wantFailedAttempts int
	}{
		{name: "valid code", valid: true},
		{name: "invalid code", wantFailedAttempts: 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, app := newTOTPIntegrationTestApp(t)
			userID := googleUuid.Must(googleUuid.NewV7())
			createTestUser(t, db, userID, testTOTPAccountName)
			secret := seedTOTPSecret(t, db, userID, time.Time{}, false)

			code := testInvalidTOTPCode

			if tc.valid {
				var err error

				code, err = totp.GenerateCode(secret, time.Now().UTC())
				require.NoError(t, err, "Should generate TOTP code")
			}

			response := verifyTOTP(t, app, userID, code, fiber.StatusOK)
			verified, ok := response["verified"].(bool)
			require.True(t, ok, "verified should be boolean")
			require.Equal(t, tc.valid, verified)

			var storedSecret cryptoutilIdentityMfa.TOTPSecret
			require.NoError(t, db.Where("user_id = ?", userID).First(&storedSecret).Error)
			require.Equal(t, tc.wantFailedAttempts, storedSecret.FailedAttempts)
			require.Equal(t, tc.valid, storedSecret.Verified)
			require.Equal(t, tc.valid, !storedSecret.LastUsedAt.IsZero())
		})
	}
}

// TestTOTPLockout validates behavior immediately below and at the failed-attempt limit.
func TestTOTPLockout(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		failedAttempts int
		wantLocked     bool
	}{
		{name: "below lockout threshold", failedAttempts: cryptoutilIdentityMfa.MaxFailedAttempts - 1},
		{name: "at lockout threshold", failedAttempts: cryptoutilIdentityMfa.MaxFailedAttempts, wantLocked: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, app := newTOTPIntegrationTestApp(t)
			userID := googleUuid.Must(googleUuid.NewV7())
			createTestUser(t, db, userID, testTOTPAccountName)
			seedTOTPSecret(t, db, userID, time.Time{}, false)

			for range tc.failedAttempts {
				response := verifyTOTP(t, app, userID, testInvalidTOTPCode, fiber.StatusOK)
				verified, ok := response["verified"].(bool)
				require.True(t, ok, "verified should be boolean")
				require.False(t, verified)
			}

			var storedSecret cryptoutilIdentityMfa.TOTPSecret
			require.NoError(t, db.Where("user_id = ?", userID).First(&storedSecret).Error)
			require.Equal(t, tc.failedAttempts, storedSecret.FailedAttempts)

			if tc.wantLocked {
				require.True(t, storedSecret.LockedUntil.After(time.Now().UTC()))
				response := verifyTOTP(t, app, userID, testInvalidTOTPCode, fiber.StatusForbidden)
				require.Contains(t, response, cryptoutilSharedMagic.StringError)
			} else {
				require.True(t, storedSecret.LockedUntil.IsZero())
			}
		})
	}
}

// TestMFAStepUp checks recent, expired, and never-used TOTP verification timestamps.
func TestMFAStepUp(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		lastUsedAge  time.Duration
		verified     bool
		wantRequired bool
	}{
		{name: "recent verification", lastUsedAge: cryptoutilIdentityMfa.MFAStepUpDuration - time.Minute, verified: true},
		{name: "expired verification", lastUsedAge: cryptoutilIdentityMfa.MFAStepUpDuration + time.Minute, verified: true, wantRequired: true},
		{name: "never verified", wantRequired: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, app := newTOTPIntegrationTestApp(t)
			userID := googleUuid.Must(googleUuid.NewV7())
			createTestUser(t, db, userID, testTOTPAccountName)

			lastUsedAt := time.Time{}
			if tc.lastUsedAge > 0 {
				lastUsedAt = time.Now().UTC().Add(-tc.lastUsedAge)
			}

			seedTOTPSecret(t, db, userID, lastUsedAt, tc.verified)

			response := checkMFAStepUp(t, app, userID, fiber.StatusOK)
			required, ok := response[testRequired].(bool)
			require.True(t, ok, "required should be boolean")
			require.Equal(t, tc.wantRequired, required)
		})
	}
}

// TestBackupCodes covers code generation, single-use, regeneration, and unknown-code paths.
func TestBackupCodes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                   string
		generateCodes          bool
		wantVerifiedOnFirstUse bool
	}{
		{name: "generated code is single use", generateCodes: true, wantVerifiedOnFirstUse: true},
		{name: "unknown code is rejected", wantVerifiedOnFirstUse: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, app := newTOTPIntegrationTestApp(t)
			userID := googleUuid.Must(googleUuid.NewV7())
			createTestUser(t, db, userID, testTOTPAccountName)
			seedTOTPSecret(t, db, userID, time.Time{}, false)

			code := testInvalidTOTPCode

			if tc.generateCodes {
				generatedResponse := generateBackupCodes(t, app, userID, fiber.StatusCreated)
				backupCodes, ok := generatedResponse["backup_codes"].([]any)
				require.True(t, ok, "backup_codes should be array")
				require.Len(t, backupCodes, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)

				code, ok = backupCodes[0].(string)
				require.True(t, ok, "Backup code should be string")
			}

			firstResponse := verifyBackupCode(t, app, userID, code, fiber.StatusOK)
			firstVerified, ok := firstResponse["verified"].(bool)
			require.True(t, ok, "verified should be boolean")
			require.Equal(t, tc.wantVerifiedOnFirstUse, firstVerified)

			var storedCodes []cryptoutilIdentityMfa.BackupCode
			require.NoError(t, db.Where("user_id = ?", userID).Find(&storedCodes).Error)

			if !tc.generateCodes {
				require.Empty(t, storedCodes)

				return
			}

			require.Len(t, storedCodes, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)

			usedCount := 0

			for _, storedCode := range storedCodes {
				if storedCode.Used {
					usedCount++
				}
			}

			require.Equal(t, 1, usedCount)

			reusedResponse := verifyBackupCode(t, app, userID, code, fiber.StatusOK)
			reusedVerified, ok := reusedResponse["verified"].(bool)
			require.True(t, ok, "verified should be boolean")
			require.False(t, reusedVerified)

			regeneratedResponse := generateBackupCodes(t, app, userID, fiber.StatusCreated)
			newCodes, ok := regeneratedResponse["backup_codes"].([]any)
			require.True(t, ok, "backup_codes should be array")
			require.Len(t, newCodes, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)

			storedCodes = make([]cryptoutilIdentityMfa.BackupCode, 0)
			require.NoError(t, db.Where("user_id = ?", userID).Find(&storedCodes).Error)
			require.Len(t, storedCodes, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)

			for _, storedCode := range storedCodes {
				require.False(t, storedCode.Used)
			}
		})
	}
}

// ========== Helper Functions ==========

// createTOTPIntegrationTestDependencies creates test dependencies for TOTP integration tests.
