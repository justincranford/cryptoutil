// Copyright (c) 2025-2026 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !integration

package apis

import (
	"bytes"
	json "encoding/json"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

const testSessionTypeService = "service"

func TestValidateSession_SessionManagerErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		sessionType string
	}{
		{name: "browser validation error", sessionType: sessionTypeBrowser},
		{name: "service validation error", sessionType: testSessionTypeService},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := NewSessionHandler(newMockSessionManagerError(errMockSessionError))
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Post("/session/validate", handler.ValidateSession)

			requestBody, err := json.Marshal(SessionValidateRequest{
				Token:       testServiceSessionToken,
				SessionType: tc.sessionType,
			})
			require.NoError(t, err)

			request := httptest.NewRequest("POST", "/session/validate", bytes.NewReader(requestBody))
			request.Header.Set("Content-Type", "application/json")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, fiber.StatusOK, response.StatusCode)

			var actual SessionValidateResponse
			require.NoError(t, json.NewDecoder(response.Body).Decode(&actual))
			require.False(t, actual.Valid)
			require.Empty(t, actual.UserID)
			require.Empty(t, actual.TenantID)
			require.Empty(t, actual.RealmID)
		})
	}
}
