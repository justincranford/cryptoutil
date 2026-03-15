// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck,revive // Integration test with realistic error propagation
package authz_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
)

// TestTOTPEnrollment_HappyPath validates complete TOTP enrollment flow.
func TestTOTPEnrollment_HappyPath(t *testing.T) {
	t.Parallel()

	config, repoFactory := createTOTPIntegrationTestDependencies(t)

	db := repoFactory.DB()

	// Create TOTP service.
	totpSvc := cryptoutilIdentityMfa.NewTOTPService(db)
	require.NotNil(t, totpSvc, "TOTP service should not be nil")

	// Create AuthZ service.
	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// ========== STEP 1: Enroll User in TOTP ==========
	userID := googleUuid.Must(googleUuid.NewV7())
	issuer := "CryptoUtil"
	accountName := "testuser@example.com"

	// Create test user in database.
	createTestUser(t, db, userID, accountName)

	enrollResp := enrollTOTP(t, app, userID, issuer, accountName)

	// Validate response fields.
	require.Contains(t, enrollResp, "secret_id", "Response should include secret_id")
	require.Contains(t, enrollResp, "qr_code_uri", "Response should include qr_code_uri")
	require.Contains(t, enrollResp, "backup_codes", "Response should include backup_codes")

	qrCodeURI, ok := enrollResp["qr_code_uri"].(string)
	require.True(t, ok, "qr_code_uri should be string")
	require.Contains(t, qrCodeURI, "otpauth://totp/", "QR code should be otpauth URI")
	require.Contains(t, qrCodeURI, issuer, "QR code should contain issuer")
	require.Contains(t, qrCodeURI, accountName, "QR code should contain account name")

	backupCodes, ok := enrollResp["backup_codes"].([]any)
	require.True(t, ok, "backup_codes should be array")
	require.Len(t, backupCodes, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, "Should generate 10 backup codes")

	// ========== STEP 2: Verify Secret Stored in Database ==========
	var secret cryptoutilIdentityMfa.TOTPSecret

	err := db.Where("user_id = ?", userID).First(&secret).Error
	require.NoError(t, err, "Should retrieve TOTP secret")
	require.Equal(t, userID, secret.UserID, "User ID should match")
	require.False(t, secret.Verified, "Secret should not be verified yet")
	require.Equal(t, 0, secret.FailedAttempts, "Should have zero failed attempts")

	// ========== STEP 3: Verify Backup Codes Stored in Database ==========
	var storedCodes []cryptoutilIdentityMfa.BackupCode

	err = db.Where("user_id = ?", userID).Find(&storedCodes).Error
	require.NoError(t, err, "Should retrieve backup codes")
	require.Len(t, storedCodes, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, "Should store 10 backup codes")

	for _, code := range storedCodes {
		require.False(t, code.Used, "Backup codes should not be used yet")
	}
}

// TestTOTPVerification_ValidCode validates TOTP code verification.
func TestTOTPVerification_ValidCode(t *testing.T) {
	t.Parallel()

	config, repoFactory := createTOTPIntegrationTestDependencies(t)

	db := repoFactory.DB()

	// Create TOTP service.
	totpSvc := cryptoutilIdentityMfa.NewTOTPService(db)
	require.NotNil(t, totpSvc, "TOTP service should not be nil")

	// Create AuthZ service.
	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Enroll user.
	userID := googleUuid.Must(googleUuid.NewV7())

	// Create test user in database.
	createTestUser(t, db, userID, "testuser@example.com")

	enrollResp := enrollTOTP(t, app, userID, "CryptoUtil", "testuser@example.com")

	qrCodeURI, ok := enrollResp["qr_code_uri"].(string)
	require.True(t, ok, "qr_code_uri should be string")

	// Extract secret from QR code URI.
	secret := extractSecretFromQRCode(t, qrCodeURI)

	// Generate valid TOTP code.
	code, err := totp.GenerateCode(secret, time.Now().UTC())
	require.NoError(t, err, "Should generate TOTP code")

	// Verify TOTP code.
	verifyResp := verifyTOTP(t, app, userID, code, 200)

	verified, ok := verifyResp["verified"].(bool)
	require.True(t, ok, "verified should be boolean")
	require.True(t, verified, "Code should be verified")

	// Verify last_used_at updated in database.
	var totpSecret cryptoutilIdentityMfa.TOTPSecret

	err = db.Where("user_id = ?", userID).First(&totpSecret).Error
	require.NoError(t, err, "Should retrieve TOTP secret")
	require.NotNil(t, totpSecret.LastUsedAt, "last_used_at should be set")
	require.True(t, totpSecret.Verified, "Secret should be marked as verified")
}

// TestTOTPVerification_InvalidCode validates invalid TOTP code handling.
func TestTOTPVerification_InvalidCode(t *testing.T) {
	t.Parallel()

	config, repoFactory := createTOTPIntegrationTestDependencies(t)

	db := repoFactory.DB()

	// Create AuthZ service.
	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Enroll user.
	userID := googleUuid.Must(googleUuid.NewV7())

	// Create test user in database.
	createTestUser(t, db, userID, "testuser@example.com")

	_ = enrollTOTP(t, app, userID, "CryptoUtil", "testuser@example.com")

	// Verify with invalid code.
	invalidCode := "000000"
	verifyResp := verifyTOTP(t, app, userID, invalidCode, 200)

	verified, ok := verifyResp["verified"].(bool)
	require.True(t, ok, "verified should be boolean")
	require.False(t, verified, "Invalid code should not be verified")
}

// TestTOTPLockout_FiveFailedAttempts validates account lockout after 5 failed attempts.
func TestTOTPLockout_FiveFailedAttempts(t *testing.T) {
	t.Parallel()

	config, repoFactory := createTOTPIntegrationTestDependencies(t)

	db := repoFactory.DB()

	// Create AuthZ service.
	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Enroll user.
	userID := googleUuid.Must(googleUuid.NewV7())

	// Create test user in database.
	createTestUser(t, db, userID, "testuser@example.com")

	_ = enrollTOTP(t, app, userID, "CryptoUtil", "testuser@example.com")

	// Submit 5 invalid codes.
	invalidCode := "000000"
	for i := 0; i < cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries; i++ {
		verifyResp := verifyTOTP(t, app, userID, invalidCode, 200)
		verified, ok := verifyResp["verified"].(bool)
		require.True(t, ok, "verified should be boolean")
		require.False(t, verified, "Invalid code should not be verified")
	}

	// Verify account locked.
	var totpSecret cryptoutilIdentityMfa.TOTPSecret

	err := db.Where("user_id = ?", userID).First(&totpSecret).Error
	require.NoError(t, err, "Should retrieve TOTP secret")
	require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, totpSecret.FailedAttempts, "Should have 5 failed attempts")
	require.NotNil(t, totpSecret.LockedUntil, "Account should be locked")
	require.True(t, totpSecret.LockedUntil.After(time.Now().UTC()), "Lockout should be in future")

	// 6th attempt should return 403 Forbidden.
	verifyResp := verifyTOTP(t, app, userID, invalidCode, 403)
	require.Contains(t, verifyResp, cryptoutilSharedMagic.StringError, "Response should include error")
}

// TestMFAStepUp_ThirtyMinuteRequirement validates 30-minute step-up requirement.
func TestMFAStepUp_ThirtyMinuteRequirement(t *testing.T) {
	t.Parallel()

	config, repoFactory := createTOTPIntegrationTestDependencies(t)

	db := repoFactory.DB()

	// Create AuthZ service.
	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Enroll and verify user.
	userID := googleUuid.Must(googleUuid.NewV7())

	// Create test user in database.
	createTestUser(t, db, userID, "testuser@example.com")

	enrollResp := enrollTOTP(t, app, userID, "CryptoUtil", "testuser@example.com")

	qrCodeURI, ok := enrollResp["qr_code_uri"].(string)
	require.True(t, ok, "qr_code_uri should be string")

	secret := extractSecretFromQRCode(t, qrCodeURI)

	code, err := totp.GenerateCode(secret, time.Now().UTC())
	require.NoError(t, err, "Should generate TOTP code")

	verifyResp := verifyTOTP(t, app, userID, code, 200)

	verified, ok := verifyResp["verified"].(bool)
	require.True(t, ok, "verified should be boolean")
	require.True(t, verified, "Code should be verified")

	// Check step-up immediately (should not be required).
	stepUpResp1 := checkMFAStepUp(t, app, userID, 200)

	required1, ok := stepUpResp1["required"].(bool)
	require.True(t, ok, "required should be boolean")
	require.False(t, required1, "Step-up should not be required immediately after verification")

	// Manually set last_used_at to 31 minutes ago.
	var totpSecret cryptoutilIdentityMfa.TOTPSecret

	err = db.Where("user_id = ?", userID).First(&totpSecret).Error
	require.NoError(t, err, "Should retrieve TOTP secret")

	totpSecret.LastUsedAt = time.Now().UTC().Add(-31 * time.Minute)
	err = db.Save(&totpSecret).Error
	require.NoError(t, err, "Should update TOTP secret")

	// Check step-up again (should be required).
	stepUpResp2 := checkMFAStepUp(t, app, userID, 200)

	required2, ok := stepUpResp2["required"].(bool)
	require.True(t, ok, "required should be boolean")
	require.True(t, required2, "Step-up should be required after 30 minutes")
}

// TestBackupCodes_GenerateAndVerify validates backup code generation and verification.
func TestBackupCodes_GenerateAndVerify(t *testing.T) {
	t.Parallel()

	config, repoFactory := createTOTPIntegrationTestDependencies(t)

	db := repoFactory.DB()

	// Create AuthZ service.
	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Enroll user.
	userID := googleUuid.Must(googleUuid.NewV7())

	// Create test user in database.
	createTestUser(t, db, userID, "testuser@example.com")

	enrollResp := enrollTOTP(t, app, userID, "CryptoUtil", "testuser@example.com")

	backupCodes, ok := enrollResp["backup_codes"].([]any)
	require.True(t, ok, "backup_codes should be array")
	require.Len(t, backupCodes, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, "Should generate 10 backup codes")

	// Store first code for testing.
	code1, ok := backupCodes[0].(string)
	require.True(t, ok, "Backup code should be string")

	// Verify backup code (should succeed).
	verifyResp1 := verifyBackupCode(t, app, userID, code1, 200)

	verified1, ok := verifyResp1["verified"].(bool)
	require.True(t, ok, "verified should be boolean")
	require.True(t, verified1, "Backup code should be verified")

	// Verify code #1 marked as used in database.
	var storedCodes []cryptoutilIdentityMfa.BackupCode

	err := db.Where("user_id = ?", userID).Find(&storedCodes).Error
	require.NoError(t, err, "Should retrieve backup codes")

	usedCount := 0

	for _, code := range storedCodes {
		if code.Used {
			usedCount++
		}
	}

	require.Equal(t, 1, usedCount, "Should have 1 used backup code")

	// Verify same code again (should fail - already used).
	verifyResp2 := verifyBackupCode(t, app, userID, code1, 200)

	verified2, ok := verifyResp2["verified"].(bool)
	require.True(t, ok, "verified should be boolean")
	require.False(t, verified2, "Used backup code should not be verified again")

	// Generate new backup codes.
	regenerateResp := generateBackupCodes(t, app, userID, 201)

	newCodes, ok := regenerateResp["backup_codes"].([]any)
	require.True(t, ok, "backup_codes should be array")
	require.Len(t, newCodes, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, "Should generate 10 new backup codes")

	// Verify old codes invalidated.
	var storedCodesAfterRegenerate []cryptoutilIdentityMfa.BackupCode

	err = db.Where("user_id = ?", userID).Find(&storedCodesAfterRegenerate).Error
	require.NoError(t, err, "Should retrieve backup codes after regeneration")
	require.Len(t, storedCodesAfterRegenerate, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, "Should have 10 backup codes after regeneration")

	for _, code := range storedCodesAfterRegenerate {
		require.False(t, code.Used, "New backup codes should not be used")
	}
}

// ========== Helper Functions ==========

// createTOTPIntegrationTestDependencies creates test dependencies for TOTP integration tests.
