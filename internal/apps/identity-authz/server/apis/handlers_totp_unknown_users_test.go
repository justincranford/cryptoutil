// Copyright (c) 2025-2026 Justin Cranford.
package apis_test

import (
	"bytes"
	json "encoding/json"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestTOTPHandlers_UnknownUsers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		method        string
		path          func(string) string
		body          func(string) map[string]string
		wantStatus    int
		responseField string
		wantBool      bool
	}{
		{
			name:   "enroll",
			method: testPost,
			path:   func(_ string) string { return testTOTPEnrollPath },
			body: func(userID string) map[string]string {
				return map[string]string{testUserIDValueF89D6B69: userID, testTOTPIssuerField: cryptoutilSharedMagic.DefaultOTLPServiceDefault, testTOTPAccountNameField: "test"}
			},
			wantStatus: fiber.StatusNotFound,
		},
		{
			name:   testTOTPGenerateBackupCodesCase,
			method: testPost,
			path:   func(_ string) string { return testTOTPBackupCodesGeneratePath },
			body: func(userID string) map[string]string {
				return map[string]string{"user_id": userID}
			},
			wantStatus: fiber.StatusNotFound,
		},
		{
			name:   "verify unenrolled TOTP",
			method: testPost,
			path:   func(_ string) string { return testTOTPVerifyPath },
			body: func(userID string) map[string]string {
				return map[string]string{testUserIDValueF89D6B69: userID, cryptoutilSharedMagic.ResponseTypeCode: "000000"}
			},
			wantStatus:    fiber.StatusOK,
			responseField: testTOTPVerifiedField,
		},
		{
			name:   "step up without TOTP enrollment",
			method: http.MethodGet,
			path: func(userID string) string {
				return "/oidc/v1/mfa/totp/step-up?user_id=" + userID
			},
			wantStatus:    fiber.StatusOK,
			responseField: "required",
		},
		{
			name:   "verify absent backup code",
			method: testPost,
			path:   func(_ string) string { return testTOTPBackupCodesVerifyPath },
			body: func(userID string) map[string]string {
				return map[string]string{testUserIDValueF89D6B69: userID, cryptoutilSharedMagic.ResponseTypeCode: testTestCode1234}
			},
			wantStatus:    fiber.StatusOK,
			responseField: testTOTPVerifiedField,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userID := googleUuid.Must(googleUuid.NewV7()).String()

			var requestBody io.Reader

			if tc.body != nil {
				encodedBody, err := json.Marshal(tc.body(userID))
				require.NoError(t, err)

				requestBody = bytes.NewReader(encodedBody)
			}

			request := httptest.NewRequest(tc.method, tc.path(userID), requestBody)
			if tc.body != nil {
				request.Header.Set("Content-Type", "application/json")
			}

			service := cryptoutilIdentityAuthz.NewService(totpTestConfig, totpTestRepoFactory, nil)
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			service.RegisterRoutes(app)

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, response.Body.Close())
			}()

			require.Equal(t, tc.wantStatus, response.StatusCode)

			if tc.responseField != "" {
				var result map[string]any
				require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
				require.Equal(t, tc.wantBool, result[tc.responseField])
			}
		})
	}
}
