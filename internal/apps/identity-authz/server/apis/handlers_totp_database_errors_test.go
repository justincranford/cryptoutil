// Copyright (c) 2025-2026 Justin Cranford.
//

package apis_test

import (
	"bytes"
	crand "crypto/rand"
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
)

func TestTOTPHandlers_DatabaseErrors(t *testing.T) {
	t.Parallel()

	config, repoFactory, err := newTOTPIntegrationTestFixture()
	require.NoError(t, err)
	require.NoError(t, repoFactory.Close())

	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	userID := googleUuid.Must(googleUuid.NewV7()).String()

	tests := []struct {
		name             string
		method           string
		path             string
		body             string
		status           int
		verify           *bool
		errorDescription string
	}{
		{
			name:             "enroll TOTP",
			method:           http.MethodPost,
			path:             testTOTPEnrollPath,
			body:             `{"user_id":"` + userID + `","issuer":"` + testTOTPIssuer + `","account_name":"` + testTestExampleCom + `"}`,
			status:           fiber.StatusInternalServerError,
			errorDescription: testInternalServerErrorDescription,
		},
		{
			name:   "verify TOTP",
			method: http.MethodPost,
			path:   testTOTPVerifyPath,
			body:   `{"user_id":"` + userID + `","code":"123456"}`,
			status: fiber.StatusOK,
			verify: boolPointer(false),
		},
		{
			name:             "check MFA step-up",
			method:           http.MethodGet,
			path:             "/oidc/v1/mfa/totp/step-up?user_id=" + userID,
			status:           fiber.StatusInternalServerError,
			errorDescription: "Failed to check MFA step-up requirement",
		},
		{
			name:             testTOTPGenerateBackupCodesCase,
			method:           http.MethodPost,
			path:             testTOTPBackupCodesGeneratePath,
			body:             `{"user_id":"` + userID + `"}`,
			status:           fiber.StatusInternalServerError,
			errorDescription: "Internal server error",
		},
		{
			name:   "verify backup code",
			method: http.MethodPost,
			path:   testTOTPBackupCodesVerifyPath,
			body:   `{"user_id":"` + userID + `","code":"` + testTestCode1234 + `"}`,
			status: fiber.StatusOK,
			verify: boolPointer(false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			service.RegisterRoutes(app)

			request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			request.Header.Set("Content-Type", "application/json")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, tc.status, response.StatusCode)

			if tc.errorDescription != "" {
				var result map[string]string
				require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
				require.Equal(t, tc.errorDescription, result["error_description"])
			}

			if tc.verify != nil {
				var result cryptoutilIdentityAuthz.VerifyTOTPResponse
				require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
				require.Equal(t, *tc.verify, result.Verified)
			}
		})
	}
}

func boolPointer(value bool) *bool {
	return &value
}

// Sequential: replaces crypto/rand.Reader to test TOTP random-source failures.
func TestTOTPHandlers_RandomSourceFailures(t *testing.T) {
	originalReader := crand.Reader
	defer func() { crand.Reader = originalReader }()

	crand.Reader = bytes.NewReader(nil)

	db, app := newTOTPIntegrationTestApp(t)
	userID := googleUuid.Must(googleUuid.NewV7())
	createTestUser(t, db, userID, "totp-random-source-failure")

	tests := []struct {
		name                 string
		path                 string
		body                 any
		wantErrorDescription string
	}{
		{
			name: "enrollment random source",
			path: testTOTPEnrollPath,
			body: cryptoutilIdentityAuthz.EnrollTOTPRequest{
				UserID:      userID.String(),
				Issuer:      testTOTPIssuer,
				AccountName: testTOTPAccountName,
			},
			wantErrorDescription: "Failed to enroll in TOTP",
		},
		{
			name:                 "backup code random source",
			path:                 "/oidc/v1/mfa/totp/backup-codes/generate",
			body:                 map[string]string{testUserIDValueF89D6B69: userID.String()},
			wantErrorDescription: "Failed to generate backup codes",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requestBody, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request := httptest.NewRequest(http.MethodPost, tc.path, bytes.NewReader(requestBody))
			request.Header.Set("Content-Type", "application/json")
			response, err := appTestTOTPHashRequest(app, request)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			var result map[string]string
			require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
			require.Equal(t, fiber.StatusInternalServerError, response.StatusCode)
			require.Equal(t, tc.wantErrorDescription, result["error_description"])
		})
	}
}
