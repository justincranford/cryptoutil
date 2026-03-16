// Copyright (c) 2025 Justin Cranford

package authz_test

import (
	"bytes"
	json "encoding/json"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
)

// handleEnrollTOTP tests.

func TestHandleEnrollTOTP_InvalidBody(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/totp/enroll", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleEnrollTOTP_MissingUserID(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := cryptoutilIdentityAuthz.EnrollTOTPRequest{
		UserID:      "",
		Issuer:      "CryptoUtil",
		AccountName: "test@example.com",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/totp/enroll", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleEnrollTOTP_InvalidUserIDFormat(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := cryptoutilIdentityAuthz.EnrollTOTPRequest{
		UserID:      "not-a-uuid",
		Issuer:      "CryptoUtil",
		AccountName: "test@example.com",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/totp/enroll", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleEnrollTOTP_MissingIssuer(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	testUserID := googleUuid.Must(googleUuid.NewV7()).String()
	reqBody := cryptoutilIdentityAuthz.EnrollTOTPRequest{
		UserID:      testUserID,
		Issuer:      "",
		AccountName: "test@example.com",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/totp/enroll", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleEnrollTOTP_MissingAccountName(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	testUserID := googleUuid.Must(googleUuid.NewV7()).String()
	reqBody := cryptoutilIdentityAuthz.EnrollTOTPRequest{
		UserID:      testUserID,
		Issuer:      "CryptoUtil",
		AccountName: "",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/totp/enroll", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// handleVerifyTOTP tests.

func TestHandleVerifyTOTP_InvalidBody(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/totp/verify", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleVerifyTOTP_MissingUserID(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := cryptoutilIdentityAuthz.VerifyTOTPRequest{
		UserID: "",
		Code:   "123456",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/totp/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleVerifyTOTP_InvalidUserIDFormat(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := cryptoutilIdentityAuthz.VerifyTOTPRequest{
		UserID: "not-a-uuid",
		Code:   "123456",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/totp/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleVerifyTOTP_MissingCode(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	testUserID := googleUuid.Must(googleUuid.NewV7()).String()
	reqBody := cryptoutilIdentityAuthz.VerifyTOTPRequest{
		UserID: testUserID,
		Code:   "",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/totp/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// handleCheckMFAStepUp tests.

func TestHandleCheckMFAStepUp_MissingUserIDQueryParam(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	req := httptest.NewRequest("GET", "/oidc/v1/mfa/totp/step-up", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleCheckMFAStepUp_InvalidUserIDFormat(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	req := httptest.NewRequest("GET", "/oidc/v1/mfa/totp/step-up?user_id=not-a-uuid", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// handleGenerateTOTPBackupCodes tests.
