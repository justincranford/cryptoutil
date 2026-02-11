// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

//go:build !integration

package middleware

import (
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

const testInvalidUUIDFormat = "not-a-valid-uuid-format"

// TestSessionMiddleware_BrowserSession_InvalidUserIDFormat tests UUID parse error for browser sessions.
// Targets: session.go line 100-102 (UserID parse error case).
func TestSessionMiddleware_BrowserSession_InvalidUserIDFormat(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	invalidUserID := testInvalidUUIDFormat

	validator := &mockSessionValidator{
		browserSession: &cryptoutilAppsTemplateServiceServerRepository.BrowserSession{
			Session: cryptoutilAppsTemplateServiceServerRepository.Session{
				TenantID: tenantID,
				RealmID:  realmID,
			},
			UserID: &invalidUserID, // Invalid UUID format
		},
	}

	var capturedUserID any

	app := createTestApp()

	app.Get("/test", SessionMiddleware(validator, true), func(c *fiber.Ctx) error {
		capturedUserID = c.Locals(ContextKeyUserID)

		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer validtoken")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)

	// user_id should NOT be set because UUID parsing failed
	require.Nil(t, capturedUserID, "user_id should not be set when UserID parse fails")
}

// TestSessionMiddleware_ServiceSession_InvalidClientIDFormat tests UUID parse error for service sessions.
// Targets: session.go line 125-127 (ClientID parse error case).
func TestSessionMiddleware_ServiceSession_InvalidClientIDFormat(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	invalidClientID := testInvalidUUIDFormat

	validator := &mockSessionValidator{
		serviceSession: &cryptoutilAppsTemplateServiceServerRepository.ServiceSession{
			Session: cryptoutilAppsTemplateServiceServerRepository.Session{
				TenantID: tenantID,
				RealmID:  realmID,
			},
			ClientID: &invalidClientID, // Invalid UUID format
		},
	}

	var (
		capturedUserID   any
		capturedClientID any
	)

	app := createTestApp()

	app.Get("/test", SessionMiddleware(validator, false), func(c *fiber.Ctx) error {
		capturedUserID = c.Locals(ContextKeyUserID)
		capturedClientID = c.Locals(ContextKeyClientID)

		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer validtoken")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)

	// client_id should be set (raw string value)
	require.NotNil(t, capturedClientID, "client_id should be set")
	require.Equal(t, invalidClientID, capturedClientID, "client_id should match invalid value")

	// user_id should NOT be set because UUID parsing failed
	require.Nil(t, capturedUserID, "user_id should not be set when ClientID parse fails")
}
