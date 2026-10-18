// Copyright (c) 2025-2026 Justin Cranford.
//
//

//nolint:wrapcheck,revive // Integration test with realistic error propagation
package apis_test

import (
	"bytes"
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func createTOTPIntegrationTestDependencies(t *testing.T) (*cryptoutilIdentityConfig.Config, *cryptoutilIdentityRepository.RepositoryFactory) {
	t.Helper()

	require.NotNil(t, totpTestConfig, "TOTP test configuration should be initialized by TestMain")
	require.NotNil(t, totpTestRepoFactory, "TOTP test repository factory should be initialized by TestMain")

	return totpTestConfig, totpTestRepoFactory
}

const (
	testTOTPSecret      = "JBSWY3DPEHPK3PXP"
	testInvalidTOTPCode = "not-a-valid-code"
	testTOTPIssuer      = "CryptoUtil"
	testTOTPAccountName = "testuser@example.com"
)

// Serialize TOTP requests that perform multiple CPU-heavy PBKDF2 operations.
var totpHashRequestMu sync.Mutex

func appTestTOTPHashRequest(app *fiber.App, req *http.Request) (*http.Response, error) {
	totpHashRequestMu.Lock()
	defer totpHashRequestMu.Unlock()

	return app.Test(req, cryptoutilSharedMagic.FiberTestPBKDF2TimeoutMs)
}

func newTOTPIntegrationTestApp(t *testing.T) (*gorm.DB, *fiber.App) {
	t.Helper()

	config, repoFactory := createTOTPIntegrationTestDependencies(t)
	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, service, "AuthZ service should not be nil")

	app := fiber.New()
	service.RegisterRoutes(app)

	return repoFactory.DB(), app
}

func seedTOTPSecret(t *testing.T, db *gorm.DB, userID googleUuid.UUID, lastUsedAt time.Time, verified bool) string {
	t.Helper()

	secretID, err := googleUuid.NewV7()
	require.NoError(t, err, "Should generate a TOTP secret ID")

	now := time.Now().UTC()
	secret := &cryptoutilIdentityMfa.TOTPSecret{
		ID:              secretID,
		UserID:          userID,
		Secret:          testTOTPSecret,
		Algorithm:       cryptoutilIdentityMfa.DefaultTOTPAlgorithm,
		Digits:          cryptoutilIdentityMfa.DefaultTOTPDigits,
		Period:          cryptoutilIdentityMfa.DefaultTOTPPeriod,
		Verified:        verified,
		RecoveryEnabled: true,
		LastUsedAt:      lastUsedAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	require.NoError(t, db.Create(secret).Error, "Should create TOTP secret fixture")

	return testTOTPSecret
}

// createTestUser creates a user in the database for testing.
func createTestUser(t *testing.T, db *gorm.DB, userID googleUuid.UUID, email string) {
	t.Helper()

	user := &cryptoutilIdentityDomain.User{
		ID:                userID,
		Sub:               userID.String(),
		PreferredUsername: userID.String(),
		Email:             email + "." + userID.String(),
	}

	err := db.Create(user).Error
	require.NoError(t, err, "Failed to create test user")
}

// enrollTOTP sends POST /oidc/v1/mfa/totp/enroll request.
func enrollTOTP(t *testing.T, app *fiber.App, userID googleUuid.UUID, issuer, accountName string, expectedStatus int) map[string]any {
	t.Helper()

	reqBody := map[string]any{
		testUserIDValueF89D6B69:  userID.String(),
		testTOTPIssuerField:      issuer,
		testTOTPAccountNameField: accountName,
	}

	reqBytes, err := json.Marshal(reqBody)
	require.NoError(t, err, "Should marshal request body")

	req := httptest.NewRequest(testPost, testTOTPEnrollPath, bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := appTestTOTPHashRequest(app, req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, expectedStatus, resp.StatusCode, "Should return expected status code")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	return result
}

// verifyTOTP sends POST /oidc/v1/mfa/totp/verify request.
func verifyTOTP(t *testing.T, app *fiber.App, userID googleUuid.UUID, code string, expectedStatus int) map[string]any {
	t.Helper()

	reqBody := map[string]any{
		testUserIDValueF89D6B69:                userID.String(),
		cryptoutilSharedMagic.ResponseTypeCode: code,
	}

	reqBytes, err := json.Marshal(reqBody)
	require.NoError(t, err, "Should marshal request body")

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/totp/verify", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, cryptoutilSharedMagic.FiberTestTimeoutMs) // 30-second timeout for password verification
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
		testUserIDValueF89D6B69: userID.String(),
	}

	reqBytes, err := json.Marshal(reqBody)
	require.NoError(t, err, "Should marshal request body")

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/totp/backup-codes/generate", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := appTestTOTPHashRequest(app, req)
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
		testUserIDValueF89D6B69:                userID.String(),
		cryptoutilSharedMagic.ResponseTypeCode: code,
	}

	reqBytes, err := json.Marshal(reqBody)
	require.NoError(t, err, "Should marshal request body")

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/totp/backup-codes/verify", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := appTestTOTPHashRequest(app, req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, expectedStatus, resp.StatusCode, "Should return expected status code")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	return result
}
