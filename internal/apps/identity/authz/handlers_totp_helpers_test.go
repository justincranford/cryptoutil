// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck,revive // Integration test with realistic error propagation
package authz_test

import (
	"bytes"
	"context"
	json "encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

func createTOTPIntegrationTestDependencies(t *testing.T) (*cryptoutilIdentityConfig.Config, *cryptoutilIdentityRepository.RepositoryFactory) {
	t.Helper()

	config := &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  "file::memory:?cache=private",
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://localhost:8080",
		},
	}

	ctx := context.Background()
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, config.Database)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, repoFactory, "Repository factory should not be nil")

	// Auto-migrate required tables.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.User{},
		&cryptoutilIdentityMfa.TOTPSecret{},
		&cryptoutilIdentityMfa.BackupCode{},
	)
	require.NoError(t, err, "Failed to auto-migrate database tables")

	return config, repoFactory
}

// createTestUser creates a user in the database for testing.
func createTestUser(t *testing.T, db *gorm.DB, userID googleUuid.UUID, email string) {
	t.Helper()

	user := &cryptoutilIdentityDomain.User{
		ID:    userID,
		Email: email,
	}

	err := db.Create(user).Error
	require.NoError(t, err, "Failed to create test user")
}

// enrollTOTP sends POST /oidc/v1/mfa/totp/enroll request.
func enrollTOTP(t *testing.T, app *fiber.App, userID googleUuid.UUID, issuer, accountName string) map[string]any {
	t.Helper()

	reqBody := map[string]any{
		"user_id":      userID.String(),
		"issuer":       issuer,
		"account_name": accountName,
	}

	reqBytes, err := json.Marshal(reqBody)
	require.NoError(t, err, "Should marshal request body")

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/totp/enroll", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 30000) // 30-second timeout for password hashing (10 backup codes × ~150ms each)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusCreated, resp.StatusCode, "Should return 201 Created")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	return result
}

// verifyTOTP sends POST /oidc/v1/mfa/totp/verify request.
func verifyTOTP(t *testing.T, app *fiber.App, userID googleUuid.UUID, code string, expectedStatus int) map[string]any {
	t.Helper()

	reqBody := map[string]any{
		"user_id": userID.String(),
		"code":    code,
	}

	reqBytes, err := json.Marshal(reqBody)
	require.NoError(t, err, "Should marshal request body")

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/totp/verify", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 30000) // 30-second timeout for password verification
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, expectedStatus, resp.StatusCode, "Should return expected status code")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	return result
}

// checkMFAStepUp sends GET /oidc/v1/mfa/totp/step-up request.
func checkMFAStepUp(t *testing.T, app *fiber.App, userID googleUuid.UUID, expectedStatus int) map[string]any {
	t.Helper()

	req := httptest.NewRequest("GET", "/oidc/v1/mfa/totp/step-up?user_id="+userID.String(), nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, expectedStatus, resp.StatusCode, "Should return expected status code")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	return result
}

// generateBackupCodes sends POST /oidc/v1/mfa/totp/backup-codes/generate request.
func generateBackupCodes(t *testing.T, app *fiber.App, userID googleUuid.UUID, expectedStatus int) map[string]any {
	t.Helper()

	reqBody := map[string]any{
		"user_id": userID.String(),
	}

	reqBytes, err := json.Marshal(reqBody)
	require.NoError(t, err, "Should marshal request body")

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/totp/backup-codes/generate", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 30000) // 30-second timeout for generating 10 backup codes with password hashing
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, expectedStatus, resp.StatusCode, "Should return expected status code")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	return result
}

// verifyBackupCode sends POST /oidc/v1/mfa/totp/backup-codes/verify request.
func verifyBackupCode(t *testing.T, app *fiber.App, userID googleUuid.UUID, code string, expectedStatus int) map[string]any {
	t.Helper()

	reqBody := map[string]any{
		"user_id": userID.String(),
		"code":    code,
	}

	reqBytes, err := json.Marshal(reqBody)
	require.NoError(t, err, "Should marshal request body")

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/totp/backup-codes/verify", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 30000) // 30-second timeout for backup code password verification
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, expectedStatus, resp.StatusCode, "Should return expected status code")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	return result
}

// extractSecretFromQRCode extracts the secret parameter from otpauth:// URI.
func extractSecretFromQRCode(t *testing.T, qrCodeURI string) string {
	t.Helper()

	// Parse URI: otpauth://totp/Issuer:AccountName?secret=BASE32SECRET&issuer=Issuer
	require.Contains(t, qrCodeURI, "secret=", "QR code should contain secret parameter")

	// Extract secret from query string.
	parts := strings.Split(qrCodeURI, "?")
	require.Len(t, parts, 2, "QR code should have query parameters")

	queryParams := parts[1]
	params := strings.Split(queryParams, "&")

	for _, param := range params {
		if strings.HasPrefix(param, "secret=") {
			secret := strings.TrimPrefix(param, "secret=")
			require.NotEmpty(t, secret, "Secret should not be empty")

			return secret
		}
	}

	t.Fatal("Secret not found in QR code URI")

	return ""
}
