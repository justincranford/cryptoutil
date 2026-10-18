// Copyright (c) 2025-2026 Justin Cranford.
//

package apis_test

import (
	"context"
	json "encoding/json"
	"net/http/httptest"
	"net/url"
	"strings"
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
	devicePollingTestClientID = "device-polling-test-client"
)

func TestHandleDeviceCodeGrant_PollingValidationAndStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		includeDevice   bool
		includeClient   bool
		storeDevice     bool
		storedClientID  string
		requestClientID string
		status          string
		expiresIn       time.Duration
		wantStatus      int
		wantError       string
	}{
		{
			name:          "missing device code",
			includeClient: true,
			wantStatus:    fiber.StatusBadRequest,
			wantError:     cryptoutilSharedMagic.ErrorInvalidRequest,
		},
		{
			name:          "missing client ID",
			includeDevice: true,
			wantStatus:    fiber.StatusBadRequest,
			wantError:     cryptoutilSharedMagic.ErrorInvalidRequest,
		},
		{
			name:            "device code not found",
			includeDevice:   true,
			includeClient:   true,
			requestClientID: devicePollingTestClientID,
			wantStatus:      fiber.StatusBadRequest,
			wantError:       cryptoutilSharedMagic.ErrorInvalidGrant,
		},
		{
			name:            "client ID mismatch",
			includeDevice:   true,
			includeClient:   true,
			storeDevice:     true,
			storedClientID:  devicePollingTestClientID,
			requestClientID: devicePollingTestClientID + "-other",
			status:          cryptoutilIdentityDomain.DeviceAuthStatusPending,
			expiresIn:       time.Minute,
			wantStatus:      fiber.StatusBadRequest,
			wantError:       cryptoutilSharedMagic.ErrorInvalidGrant,
		},
		{
			name:            "expired device code",
			includeDevice:   true,
			includeClient:   true,
			storeDevice:     true,
			storedClientID:  devicePollingTestClientID,
			requestClientID: devicePollingTestClientID,
			status:          cryptoutilIdentityDomain.DeviceAuthStatusPending,
			expiresIn:       -time.Second,
			wantStatus:      fiber.StatusBadRequest,
			wantError:       cryptoutilSharedMagic.ErrorExpiredToken,
		},
		{
			name:            "device code already used",
			includeDevice:   true,
			includeClient:   true,
			storeDevice:     true,
			storedClientID:  devicePollingTestClientID,
			requestClientID: devicePollingTestClientID,
			status:          cryptoutilIdentityDomain.DeviceAuthStatusUsed,
			expiresIn:       time.Minute,
			wantStatus:      fiber.StatusBadRequest,
			wantError:       cryptoutilSharedMagic.ErrorInvalidGrant,
		},
		{
			name:            "unknown status",
			includeDevice:   true,
			includeClient:   true,
			storeDevice:     true,
			storedClientID:  devicePollingTestClientID,
			requestClientID: devicePollingTestClientID,
			status:          cryptoutilSharedMagic.STRATEGY_UNKNOWN,
			expiresIn:       time.Minute,
			wantStatus:      fiber.StatusInternalServerError,
			wantError:       cryptoutilSharedMagic.ErrorServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, repoFactory := createIntegrationTestDependencies(t)
			deviceCode := "device-code-" + googleUuid.Must(googleUuid.NewV7()).String()

			if tc.storeDevice {
				now := time.Now().UTC()
				deviceAuth := &cryptoutilIdentityDomain.DeviceAuthorization{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					ClientID:   tc.storedClientID,
					DeviceCode: deviceCode,
					UserCode:   deviceCode,
					Status:     tc.status,
					CreatedAt:  now,
					ExpiresAt:  now.Add(tc.expiresIn),
				}
				require.NoError(t, repoFactory.DeviceAuthorizationRepository().Create(context.Background(), deviceAuth))
			}

			service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
			app := fiber.New()
			service.RegisterRoutes(app)

			form := url.Values{
				cryptoutilSharedMagic.ParamGrantType: []string{cryptoutilSharedMagic.GrantTypeDeviceCode},
			}
			if tc.includeDevice {
				form.Set(cryptoutilSharedMagic.ParamDeviceCode, deviceCode)
			}

			if tc.includeClient {
				form.Set(cryptoutilSharedMagic.ParamClientID, tc.requestClientID)
			}

			request := httptest.NewRequest("POST", testOauth2V1Token, strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, tc.wantStatus, response.StatusCode)

			var body map[string]any
			require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
			require.Equal(t, tc.wantError, body[cryptoutilSharedMagic.StringError])
		})
	}
}
