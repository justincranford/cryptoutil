// Copyright (c) 2025-2026 Justin Cranford.
package apis_test

import (
	"bytes"
	"context"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	recoveryCodesGeneratePath   = "/oidc/v1/mfa/recovery-codes/generate"
	recoveryCodesCountPath      = "/oidc/v1/mfa/recovery-codes/count"
	recoveryCodesRegeneratePath = "/oidc/v1/mfa/recovery-codes/regenerate"
	recoveryCodeVerifyPath      = "/oidc/v1/mfa/verify-recovery-code"
	recoveryCodeFormatPattern   = `^[A-Z2-9]{4}-[A-Z2-9]{4}-[A-Z2-9]{4}-[A-Z2-9]{4}$`
)

type recoveryCodeResponse struct {
	StatusCode int
	Body       []byte
}

func TestRecoveryCodeHandler_Flows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T, app *fiber.App, user *cryptoutilIdentityDomain.User)
	}{
		{
			name: "generate codes and report count",
			run: func(t *testing.T, app *fiber.App, user *cryptoutilIdentityDomain.User) {
				generated := generateRecoveryCodes(t, app, user)
				assertGeneratedRecoveryCodes(t, generated)

				count := getRecoveryCodeCount(t, app, user)
				require.Equal(t, int64(cryptoutilSharedMagic.DefaultRecoveryCodeCount), count.Remaining)
				require.Equal(t, int64(cryptoutilSharedMagic.DefaultRecoveryCodeCount), count.Total)
			},
		},
		{
			name: "regenerate replaces existing codes",
			run: func(t *testing.T, app *fiber.App, user *cryptoutilIdentityDomain.User) {
				initial := generateRecoveryCodes(t, app, user)
				assertGeneratedRecoveryCodes(t, initial)

				response := recoveryCodeRequest(t, app, http.MethodPost, recoveryCodesRegeneratePath, cryptoutilIdentityAuthz.GenerateRecoveryCodesRequest{
					UserID: user.ID.String(),
				}, "")
				require.Equal(t, fiber.StatusCreated, response.StatusCode)

				var regenerated cryptoutilIdentityAuthz.GenerateRecoveryCodesResponse
				decodeRecoveryCodeResponse(t, response, &regenerated)
				assertGeneratedRecoveryCodes(t, regenerated)

				for _, oldCode := range initial.Codes {
					require.NotContains(t, regenerated.Codes, oldCode)
				}

				count := getRecoveryCodeCount(t, app, user)
				require.Equal(t, int64(cryptoutilSharedMagic.DefaultRecoveryCodeCount), count.Remaining)
				require.Equal(t, int64(cryptoutilSharedMagic.DefaultRecoveryCodeCount), count.Total)
			},
		},
		{
			name: "verification consumes code once",
			run: func(t *testing.T, app *fiber.App, user *cryptoutilIdentityDomain.User) {
				generated := generateRecoveryCodes(t, app, user)
				assertGeneratedRecoveryCodes(t, generated)

				verify := func() recoveryCodeResponse {
					return recoveryCodeRequest(t, app, http.MethodPost, recoveryCodeVerifyPath, cryptoutilIdentityAuthz.VerifyRecoveryCodeRequest{
						Code: generated.Codes[0],
					}, user.ID.String())
				}

				response := verify()
				require.Equal(t, fiber.StatusOK, response.StatusCode)

				var verified cryptoutilIdentityAuthz.VerifyRecoveryCodeResponse
				decodeRecoveryCodeResponse(t, response, &verified)
				require.True(t, verified.Verified)

				count := getRecoveryCodeCount(t, app, user)
				require.Equal(t, int64(cryptoutilSharedMagic.DefaultRecoveryCodeCount-1), count.Remaining)
				require.Equal(t, int64(cryptoutilSharedMagic.DefaultRecoveryCodeCount), count.Total)

				response = verify()
				require.Equal(t, fiber.StatusUnauthorized, response.StatusCode)

				var errorResponse map[string]string
				decodeRecoveryCodeResponse(t, response, &errorResponse)
				require.Equal(t, "invalid_code", errorResponse[cryptoutilSharedMagic.StringError])
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, user := newRecoveryCodeHandlerTestFixture(t)
			tc.run(t, app, user)
		})
	}
}

func newRecoveryCodeHandlerTestFixture(t *testing.T) (*fiber.App, *cryptoutilIdentityDomain.User) {
	t.Helper()

	repoFactory := createTokenTestRepoFactory(t)
	t.Cleanup(func() {
		require.NoError(t, repoFactory.Close())
	})
	require.NoError(t, repoFactory.DB().AutoMigrate(&cryptoutilIdentityDomain.RecoveryCode{}))

	userID, err := googleUuid.NewV7()
	require.NoError(t, err)

	user := &cryptoutilIdentityDomain.User{
		ID:           userID,
		Sub:          userID.String(),
		Email:        fmt.Sprintf("recovery-%s@example.com", userID),
		PasswordHash: testHash,
	}
	require.NoError(t, repoFactory.UserRepository().Create(context.Background(), user))

	service := cryptoutilIdentityAuthz.NewService(createTokenTestConfig(t), repoFactory, nil)
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	service.RegisterRoutes(app)

	return app, user
}

func generateRecoveryCodes(t *testing.T, app *fiber.App, user *cryptoutilIdentityDomain.User) cryptoutilIdentityAuthz.GenerateRecoveryCodesResponse {
	t.Helper()

	response := recoveryCodeRequest(t, app, http.MethodPost, recoveryCodesGeneratePath, cryptoutilIdentityAuthz.GenerateRecoveryCodesRequest{
		UserID: user.ID.String(),
	}, "")
	require.Equal(t, fiber.StatusCreated, response.StatusCode)

	var generated cryptoutilIdentityAuthz.GenerateRecoveryCodesResponse
	decodeRecoveryCodeResponse(t, response, &generated)

	return generated
}

func getRecoveryCodeCount(t *testing.T, app *fiber.App, user *cryptoutilIdentityDomain.User) cryptoutilIdentityAuthz.RecoveryCodeCountResponse {
	t.Helper()

	response := recoveryCodeRequest(t, app, http.MethodGet, recoveryCodesCountPath+"?user_id="+user.ID.String(), nil, "")
	require.Equal(t, fiber.StatusOK, response.StatusCode)

	var count cryptoutilIdentityAuthz.RecoveryCodeCountResponse
	decodeRecoveryCodeResponse(t, response, &count)

	return count
}

func assertGeneratedRecoveryCodes(t *testing.T, generated cryptoutilIdentityAuthz.GenerateRecoveryCodesResponse) {
	t.Helper()

	require.Len(t, generated.Codes, cryptoutilSharedMagic.DefaultRecoveryCodeCount)
	require.NotEmpty(t, generated.ExpiresAt)

	expiresAt, err := time.Parse(time.RFC3339, generated.ExpiresAt)
	require.NoError(t, err)
	require.True(t, expiresAt.After(time.Now().UTC()))

	for _, code := range generated.Codes {
		require.Regexp(t, recoveryCodeFormatPattern, code)
	}
}

func recoveryCodeRequest(t *testing.T, app *fiber.App, method, path string, body any, userID string) recoveryCodeResponse {
	t.Helper()

	var requestBody io.Reader

	if body != nil {
		encodedBody, err := json.Marshal(body)
		require.NoError(t, err)

		requestBody = bytes.NewReader(encodedBody)
	}

	request := httptest.NewRequest(method, path, requestBody)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	if userID != "" {
		request.Header.Set("X-User-ID", userID)
	}

	httpResponse, err := app.Test(request, -1)
	require.NoError(t, err)

	responseBody, readErr := io.ReadAll(httpResponse.Body)
	closeErr := httpResponse.Body.Close()

	require.NoError(t, readErr)
	require.NoError(t, closeErr)

	return recoveryCodeResponse{StatusCode: httpResponse.StatusCode, Body: responseBody}
}

func decodeRecoveryCodeResponse(t *testing.T, response recoveryCodeResponse, target any) {
	t.Helper()

	require.NoError(t, json.Unmarshal(response.Body, target))
}
