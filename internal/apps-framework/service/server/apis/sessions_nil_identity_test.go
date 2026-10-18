// Copyright (c) 2025-2026 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !integration

package apis

import (
	"bytes"
	"context"
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"
)

func TestValidateSession_NilOptionalUserIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		sessionType string
		setup       func(*mockSessionManager, googleUuid.UUID, googleUuid.UUID)
	}{
		{
			name:        "browser session with no user ID",
			sessionType: sessionTypeBrowser,
			setup: func(manager *mockSessionManager, tenantID, realmID googleUuid.UUID) {
				manager.validateBrowserFunc = func(
					context.Context,
					string,
				) (*cryptoutilAppsFrameworkServiceServerRepository.BrowserSession, error) {
					return &cryptoutilAppsFrameworkServiceServerRepository.BrowserSession{
						Session: cryptoutilAppsFrameworkServiceServerRepository.Session{
							TenantID: tenantID,
							RealmID:  realmID,
						},
					}, nil
				}
			},
		},
		{
			name:        "service session with no client ID",
			sessionType: testSessionTypeService,
			setup: func(manager *mockSessionManager, tenantID, realmID googleUuid.UUID) {
				manager.validateServiceFunc = func(
					context.Context,
					string,
				) (*cryptoutilAppsFrameworkServiceServerRepository.ServiceSession, error) {
					return &cryptoutilAppsFrameworkServiceServerRepository.ServiceSession{
						Session: cryptoutilAppsFrameworkServiceServerRepository.Session{
							TenantID: tenantID,
							RealmID:  realmID,
						},
					}, nil
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tenantID, err := googleUuid.NewV7()
			require.NoError(t, err)
			realmID, err := googleUuid.NewV7()
			require.NoError(t, err)

			sessionManager := newMockSessionManagerSuccess()
			tc.setup(sessionManager, tenantID, realmID)

			body, err := json.Marshal(SessionValidateRequest{
				Token:       testServiceSessionToken,
				SessionType: tc.sessionType,
			})
			require.NoError(t, err)

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Post("/session/validate", NewSessionHandler(sessionManager).ValidateSession)

			request := httptest.NewRequestWithContext(
				t.Context(),
				http.MethodPost,
				"/session/validate",
				bytes.NewReader(body),
			)
			request.Header.Set("Content-Type", "application/json")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, fiber.StatusOK, response.StatusCode)

			var actual SessionValidateResponse
			require.NoError(t, json.NewDecoder(response.Body).Decode(&actual))
			require.True(t, actual.Valid)
			require.Empty(t, actual.UserID)
			require.Equal(t, tenantID.String(), actual.TenantID)
			require.Equal(t, realmID.String(), actual.RealmID)
		})
	}
}
