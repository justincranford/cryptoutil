// Copyright (c) 2025-2026 Justin Cranford.
package apis_test

import (
	"bytes"
	json "encoding/json"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// handleGenerateRecoveryCodes tests.

func TestHandleGenerateRecoveryCodes_InvalidBody(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/recovery-codes/generate", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleGenerateRecoveryCodes_MissingUserID(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := cryptoutilIdentityAuthz.GenerateRecoveryCodesRequest{
		UserID: "",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/recovery-codes/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleGenerateRecoveryCodes_InvalidUserIDFormat(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := cryptoutilIdentityAuthz.GenerateRecoveryCodesRequest{
		UserID: testNotAUUID,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/recovery-codes/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestRecoveryCodeHandlers_UnknownUsers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{name: testGenerateCodesCase, path: recoveryCodesGeneratePath},
		{name: testRegenerateCodesCase, path: recoveryCodesRegeneratePath},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := createTokenTestConfig(t)
			repoFactory := createTokenTestRepoFactory(t)
			t.Cleanup(func() {
				require.NoError(t, repoFactory.Close())
			})

			service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			service.RegisterRoutes(app)

			userID := googleUuid.Must(googleUuid.NewV7())
			body, err := json.Marshal(cryptoutilIdentityAuthz.GenerateRecoveryCodesRequest{
				UserID: userID.String(),
			})
			require.NoError(t, err)

			request := httptest.NewRequest(testPost, tc.path, bytes.NewReader(body))
			request.Header.Set("Content-Type", "application/json")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, response.Body.Close())
			}()

			require.Equal(t, fiber.StatusNotFound, response.StatusCode)

			var errorResponse map[string]string
			require.NoError(t, json.NewDecoder(response.Body).Decode(&errorResponse))
			require.NotEmpty(t, errorResponse[cryptoutilSharedMagic.StringError])
		})
	}
}

func TestRecoveryCodeHandlers_StorageFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		method      string
		path        string
		userHeader  bool
		description string
		userBody    bool
		verifyBody  bool
		queryUserID bool
	}{
		{name: testGenerateCodesCase, method: testPost, path: recoveryCodesGeneratePath, description: "Failed to generate recovery codes", userBody: true},
		{name: "count codes", method: "GET", path: "/oidc/v1/mfa/recovery-codes/count?user_id=", description: "Failed to get recovery code count", queryUserID: true},
		{name: testRegenerateCodesCase, method: testPost, path: recoveryCodesRegeneratePath, description: "Failed to regenerate recovery codes", userBody: true},
		{name: "verify code", method: testPost, path: "/oidc/v1/mfa/verify-recovery-code", userHeader: true, description: "Failed to verify recovery code", verifyBody: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, repoFactory, err := newTOTPIntegrationTestFixture()

			require.NoError(t, err)
			defer func() {
				require.NoError(t, repoFactory.Close())
			}()

			userID := googleUuid.Must(googleUuid.NewV7())
			user := &cryptoutilIdentityDomain.User{
				ID:                userID,
				Sub:               userID.String(),
				PreferredUsername: userID.String(),
				Email:             userID.String() + "@example.com",
				PasswordHash:      testPasswordHash,
			}
			require.NoError(t, repoFactory.UserRepository().Create(t.Context(), user))

			service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			service.RegisterRoutes(app)

			var body []byte
			if tc.userBody {
				body, err = json.Marshal(cryptoutilIdentityAuthz.GenerateRecoveryCodesRequest{UserID: userID.String()})
				require.NoError(t, err)
			} else if tc.verifyBody {
				body, err = json.Marshal(cryptoutilIdentityAuthz.VerifyRecoveryCodeRequest{Code: "invalid-recovery-code"})
				require.NoError(t, err)
			}

			path := tc.path
			if tc.queryUserID {
				path += userID.String()
			}

			request := httptest.NewRequest(tc.method, path, bytes.NewReader(body))
			request.Header.Set("Content-Type", "application/json")

			if tc.userHeader {
				request.Header.Set("X-User-ID", userID.String())
			}

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, response.Body.Close())
			}()

			var result map[string]string
			require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
			require.Equal(t, fiber.StatusInternalServerError, response.StatusCode)
			require.Equal(t, cryptoutilSharedMagic.ErrorServerError, result[cryptoutilSharedMagic.StringError])
			require.Equal(t, tc.description, result["error_description"])
		})
	}
}

func TestRecoveryCodeHandlers_UserRepositoryFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{name: testGenerateCodesCase, path: recoveryCodesGeneratePath},
		{name: testRegenerateCodesCase, path: recoveryCodesRegeneratePath},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, repoFactory, err := newTOTPIntegrationTestFixture()

			require.NoError(t, err)
			defer func() {
				require.NoError(t, repoFactory.Close())
			}()

			require.NoError(t, repoFactory.DB().Migrator().DropTable(&cryptoutilIdentityDomain.User{}))

			service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			service.RegisterRoutes(app)

			requestBody, err := json.Marshal(cryptoutilIdentityAuthz.GenerateRecoveryCodesRequest{
				UserID: googleUuid.Must(googleUuid.NewV7()).String(),
			})
			require.NoError(t, err)

			request := httptest.NewRequest(testPost, tc.path, bytes.NewReader(requestBody))
			request.Header.Set("Content-Type", "application/json")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, response.Body.Close())
			}()

			require.Equal(t, fiber.StatusInternalServerError, response.StatusCode)
		})
	}
}

// handleGetRecoveryCodeCount tests.

func TestHandleGetRecoveryCodeCount_MissingUserIDQueryParam(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	req := httptest.NewRequest("GET", "/oidc/v1/mfa/recovery-codes/count", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleGetRecoveryCodeCount_InvalidUserIDFormat(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	req := httptest.NewRequest("GET", "/oidc/v1/mfa/recovery-codes/count?user_id=not-a-uuid", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// handleRegenerateRecoveryCodes tests.

func TestHandleRegenerateRecoveryCodes_InvalidBody(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/recovery-codes/regenerate", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleRegenerateRecoveryCodes_MissingUserID(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := cryptoutilIdentityAuthz.GenerateRecoveryCodesRequest{
		UserID: "",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/recovery-codes/regenerate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleRegenerateRecoveryCodes_InvalidUserIDFormat(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := cryptoutilIdentityAuthz.GenerateRecoveryCodesRequest{
		UserID: testNotAUUID,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/recovery-codes/regenerate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// handleVerifyRecoveryCode tests.

func TestHandleVerifyRecoveryCode_InvalidBody(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/verify-recovery-code", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleVerifyRecoveryCode_MissingCode(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := cryptoutilIdentityAuthz.VerifyRecoveryCodeRequest{
		Code: "", // Missing code
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/verify-recovery-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleVerifyRecoveryCode_MissingUserID(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := cryptoutilIdentityAuthz.VerifyRecoveryCodeRequest{
		Code: testTestCode1234,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/verify-recovery-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// X-User-ID header intentionally not set

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestHandleVerifyRecoveryCode_InvalidUserIDFormat(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := cryptoutilIdentityAuthz.VerifyRecoveryCodeRequest{
		Code: testTestCode1234,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/verify-recovery-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", testNotAUUID)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
