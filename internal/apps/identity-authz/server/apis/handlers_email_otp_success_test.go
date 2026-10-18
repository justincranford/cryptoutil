// Copyright (c) 2025-2026 Justin Cranford.
package apis_test

import (
	"bytes"
	"context"
	json "encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedCryptoPassword "cryptoutil/internal/shared/crypto/password"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestHandleSendEmailOTP_UserLookupAndSuccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		createUser bool
		wantStatus int
		wantSent   bool
	}{
		{name: testUserNotFoundCase, wantStatus: fiber.StatusNotFound},
		{name: "otp sent", createUser: true, wantStatus: fiber.StatusOK, wantSent: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := createTokenTestConfig(t)
			repoFactory := createTokenTestRepoFactory(t)
			require.NoError(t, repoFactory.DB().AutoMigrate(&cryptoutilIdentityDomain.EmailOTP{}))

			userID, err := googleUuid.NewV7()
			require.NoError(t, err)

			if tc.createUser {
				user := createIntegrationTestUser(context.Background(), t, repoFactory)
				userID = user.ID
			}

			requestBody, err := json.Marshal(cryptoutilIdentityAuthz.SendEmailOTPRequest{
				UserID: userID.String(),
				Email:  testTestExampleCom,
			})
			require.NoError(t, err)

			service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			service.RegisterRoutes(app)

			request := httptest.NewRequest(testPost, "/oidc/v1/mfa/email-otp/send", bytes.NewReader(requestBody))
			request.Header.Set("Content-Type", "application/json")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, tc.wantStatus, response.StatusCode)

			if tc.wantSent {
				var body cryptoutilIdentityAuthz.SendEmailOTPResponse
				require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
				require.True(t, body.Sent)
			}
		})
	}
}

func TestHandleVerifyEmailOTP_ValidCode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)
	require.NoError(t, repoFactory.DB().AutoMigrate(&cryptoutilIdentityDomain.EmailOTP{}))

	userID, err := googleUuid.NewV7()
	require.NoError(t, err)

	code := "123456"
	codeHash, err := cryptoutilSharedCryptoPassword.HashPassword(code)
	require.NoError(t, err)

	now := time.Now().UTC()
	otp := &cryptoutilIdentityDomain.EmailOTP{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		UserID:    userID,
		CodeHash:  codeHash,
		CreatedAt: now,
		ExpiresAt: now.Add(cryptoutilSharedMagic.DefaultEmailOTPLifetime),
	}
	require.NoError(t, repoFactory.EmailOTPRepository().Create(ctx, otp))

	requestBody, err := json.Marshal(cryptoutilIdentityAuthz.VerifyEmailOTPRequest{Code: code})
	require.NoError(t, err)

	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	service.RegisterRoutes(app)

	request := httptest.NewRequest(testPost, "/oidc/v1/mfa/email-otp/verify", bytes.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-User-ID", userID.String())

	response, err := app.Test(request, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()

	require.Equal(t, fiber.StatusOK, response.StatusCode)

	var body cryptoutilIdentityAuthz.VerifyEmailOTPResponse
	require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
	require.True(t, body.Verified)
}
