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

func TestHandleSendEmailOTP_InvalidBody(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/email-otp/send", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleSendEmailOTP_InvalidUserIDFormat(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := cryptoutilIdentityAuthz.SendEmailOTPRequest{
		UserID: testNotAUUID,
		Email:  testTestExampleCom,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/email-otp/send", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleSendEmailOTP_UserLookupAndDelivery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		createUser bool
		wantStatus int
		wantError  string
	}{
		{name: testUserNotFoundCase, wantStatus: fiber.StatusNotFound},
		{name: "OTP persistence failure", createUser: true, wantStatus: fiber.StatusInternalServerError, wantError: cryptoutilSharedMagic.ErrorServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := createTokenTestConfig(t)
			repoFactory := createTokenTestRepoFactory(t)

			userID := googleUuid.Must(googleUuid.NewV7())
			if tc.createUser {
				user := &cryptoutilIdentityDomain.User{
					ID:                userID,
					Sub:               userID.String(),
					PreferredUsername: userID.String(),
					Email:             userID.String() + "@example.com",
					PasswordHash:      testPasswordHash,
				}
				require.NoError(t, repoFactory.UserRepository().Create(t.Context(), user))
			}

			service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			service.RegisterRoutes(app)

			requestBody, err := json.Marshal(cryptoutilIdentityAuthz.SendEmailOTPRequest{
				UserID: userID.String(),
				Email:  userID.String() + "@example.com",
			})
			require.NoError(t, err)

			request := httptest.NewRequest(testPost, "/oidc/v1/mfa/email-otp/send", bytes.NewReader(requestBody))
			request.Header.Set("Content-Type", "application/json")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, response.Body.Close())
			}()

			require.Equal(t, tc.wantStatus, response.StatusCode)

			if tc.wantError != "" {
				var result map[string]string
				require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
				require.Equal(t, tc.wantError, result[cryptoutilSharedMagic.StringError])
			}
		})
	}
}

func TestHandleVerifyEmailOTP_InvalidBody(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/email-otp/verify", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleVerifyEmailOTP_MissingUserIDHeader(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := cryptoutilIdentityAuthz.VerifyEmailOTPRequest{
		Code: "123456",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/email-otp/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Missing X-User-ID header.

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleVerifyEmailOTP_InvalidUserIDHeader(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := cryptoutilIdentityAuthz.VerifyEmailOTPRequest{
		Code: "123456",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/email-otp/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", testNotAUUID)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleVerifyEmailOTP_InvalidOTP(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	userID := googleUuid.New()

	reqBody := cryptoutilIdentityAuthz.VerifyEmailOTPRequest{
		Code: "999999",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(testPost, "/oidc/v1/mfa/email-otp/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID.String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
