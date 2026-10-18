// Copyright (c) 2025-2026 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !integration

package apis

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	http "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServerService "cryptoutil/internal/apps-framework/service/server/service"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type clientAuthRegistrationServiceStub struct {
	cryptoutilAppsFrameworkServiceServerService.RegistrationService
	registerClientFn func(
		context.Context,
		string,
		string,
		*cryptoutilAppsFrameworkServiceServerService.NewTenantInfo,
		*googleUuid.UUID,
	) (*cryptoutilAppsFrameworkServiceServerService.RegistrationResult, error)
}

func (s *clientAuthRegistrationServiceStub) RegisterClient(
	ctx context.Context,
	clientID string,
	clientSecretHash string,
	newTenant *cryptoutilAppsFrameworkServiceServerService.NewTenantInfo,
	existingTenantID *googleUuid.UUID,
) (*cryptoutilAppsFrameworkServiceServerService.RegistrationResult, error) {
	return s.registerClientFn(ctx, clientID, clientSecretHash, newTenant, existingTenantID)
}

func TestHandleRegisterClient_ErrorPaths(t *testing.T) {
	t.Parallel()

	validBody, err := json.Marshal(registerClientRequest{
		TenantName: strings.Repeat("t", cryptoutilSharedMagic.DefaultTenantNameMinLength),
	})
	require.NoError(t, err)

	generateErr := errors.New("injected credential generation error")
	hashErr := errors.New("injected credential hash error")
	registerErr := errors.New("injected client registration error")

	tests := []struct {
		name             string
		requestBody      []byte
		generateErr      error
		hashErr          error
		registerErr      error
		registerResult   *cryptoutilAppsFrameworkServiceServerService.RegistrationResult
		wantStatus       int
		wantErrorMessage string
	}{
		{
			name:             testInvalidJSONCaseName,
			requestBody:      []byte("{"),
			wantStatus:       fiber.StatusBadRequest,
			wantErrorMessage: "Invalid request body",
		},
		{
			name: "tenant name shorter than minimum after trimming",
			requestBody: mustMarshalClientRegisterRequest(
				t,
				strings.Repeat(" ", cryptoutilSharedMagic.DefaultTenantNameMinLength),
			),
			wantStatus:       fiber.StatusBadRequest,
			wantErrorMessage: testTenantNameLengthInvalid,
		},
		{
			name: "tenant name longer than maximum",
			requestBody: mustMarshalClientRegisterRequest(
				t,
				strings.Repeat("t", cryptoutilSharedMagic.DefaultTenantNameMaxLength+1),
			),
			wantStatus:       fiber.StatusBadRequest,
			wantErrorMessage: testTenantNameLengthInvalid,
		},
		{
			name:             "credential generation error",
			requestBody:      validBody,
			generateErr:      generateErr,
			wantStatus:       fiber.StatusInternalServerError,
			wantErrorMessage: clientAuthErrorRegistrationFailed,
		},
		{
			name:             "credential hash error",
			requestBody:      validBody,
			hashErr:          hashErr,
			wantStatus:       fiber.StatusInternalServerError,
			wantErrorMessage: clientAuthErrorRegistrationFailed,
		},
		{
			name:             "registration service error",
			requestBody:      validBody,
			registerErr:      registerErr,
			wantStatus:       fiber.StatusInternalServerError,
			wantErrorMessage: clientAuthErrorRegistrationFailed,
		},
		{
			name:             "registration service returns no result",
			requestBody:      validBody,
			wantStatus:       fiber.StatusInternalServerError,
			wantErrorMessage: clientAuthErrorRegistrationFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handlers := &clientAuthenticationHandlers{
				registrationService: &clientAuthRegistrationServiceStub{
					registerClientFn: func(
						context.Context,
						string,
						string,
						*cryptoutilAppsFrameworkServiceServerService.NewTenantInfo,
						*googleUuid.UUID,
					) (*cryptoutilAppsFrameworkServiceServerService.RegistrationResult, error) {
						return tc.registerResult, tc.registerErr
					},
				},
				generateCredentialsFn: func() (string, string, error) {
					if tc.generateErr != nil {
						return "", "", tc.generateErr
					}

					return testServiceClientToken, testServiceClientSecret, nil
				},
				hashSecretFn: func(string) (string, error) {
					if tc.hashErr != nil {
						return "", tc.hashErr
					}

					return "hashed-secret", nil
				},
			}
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Post(
				cryptoutilSharedMagic.DefaultPublicServiceAuthClientRegisterPath,
				handlers.handleRegisterClient,
			)

			request := httptest.NewRequestWithContext(
				t.Context(),
				http.MethodPost,
				cryptoutilSharedMagic.DefaultPublicServiceAuthClientRegisterPath,
				bytes.NewReader(tc.requestBody),
			)
			request.Header.Set("Content-Type", "application/json")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, tc.wantStatus, response.StatusCode)

			var body map[string]string
			require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
			require.Equal(t, tc.wantErrorMessage, body[cryptoutilSharedMagic.StringError])
		})
	}
}

func mustMarshalClientRegisterRequest(t *testing.T, tenantName string) []byte {
	t.Helper()

	body, err := json.Marshal(registerClientRequest{TenantName: tenantName})
	require.NoError(t, err)

	return body
}
