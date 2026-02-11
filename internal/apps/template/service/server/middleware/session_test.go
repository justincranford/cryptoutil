// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

//go:build !integration

package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

// mockSessionValidator implements SessionValidator for testing.
type mockSessionValidator struct {
	browserSession    *cryptoutilAppsTemplateServiceServerRepository.BrowserSession
	browserSessionErr error
	serviceSession    *cryptoutilAppsTemplateServiceServerRepository.ServiceSession
	serviceSessionErr error
}

func (m *mockSessionValidator) ValidateBrowserSession(_ context.Context, _ string) (*cryptoutilAppsTemplateServiceServerRepository.BrowserSession, error) {
	return m.browserSession, m.browserSessionErr
}

func (m *mockSessionValidator) ValidateServiceSession(_ context.Context, _ string) (*cryptoutilAppsTemplateServiceServerRepository.ServiceSession, error) {
	return m.serviceSession, m.serviceSessionErr
}

// createTestApp creates a Fiber app with custom error handler that recognizes apperr.Error.
func createTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var appErr *cryptoutilSharedApperr.Error
			if errors.As(err, &appErr) {
				return c.Status(int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode)).JSON(fiber.Map{
					"error": appErr.Summary,
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})
}

func TestSessionMiddleware_MissingAuthHeader(t *testing.T) {
	t.Parallel()

	validator := &mockSessionValidator{}
	app := createTestApp()

	app.Get("/test", SessionMiddleware(validator, true), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 401, resp.StatusCode)
}

func TestSessionMiddleware_InvalidAuthHeaderFormat(t *testing.T) {
	t.Parallel()

	validator := &mockSessionValidator{}
	app := createTestApp()

	app.Get("/test", SessionMiddleware(validator, true), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 401, resp.StatusCode)
}

func TestSessionMiddleware_EmptyToken(t *testing.T) {
	t.Parallel()

	validator := &mockSessionValidator{}
	app := createTestApp()

	app.Get("/test", SessionMiddleware(validator, true), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// Note: "Bearer " (Bearer + space) gets trimmed by Fiber to "Bearer" (no space),
	// which triggers the "invalid format" check before reaching the "empty token" check.
	// The empty token check (session.go:75-79) is defensive dead code that can't be
	// triggered due to HTTP library header trimming behavior.
	req.Header.Set("Authorization", "Bearer ")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Returns 401 from "invalid format" check (not "empty token" check)
	// because Fiber trims "Bearer " to "Bearer" (1 part, not 2 parts)
	require.Equal(t, 401, resp.StatusCode)
}

func TestSessionMiddleware_BrowserSession_ValidationError(t *testing.T) {
	t.Parallel()

	validator := &mockSessionValidator{
		browserSessionErr: errors.New("invalid token"),
	}
	app := createTestApp()

	app.Get("/test", SessionMiddleware(validator, true), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer validtoken")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 401, resp.StatusCode)
}

func TestSessionMiddleware_BrowserSession_Success(t *testing.T) {
	t.Parallel()

	userID := googleUuid.New().String()
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	validator := &mockSessionValidator{
		browserSession: &cryptoutilAppsTemplateServiceServerRepository.BrowserSession{
			Session: cryptoutilAppsTemplateServiceServerRepository.Session{
				TenantID: tenantID,
				RealmID:  realmID,
			},
			UserID: &userID,
		},
	}

	var (
		capturedUserID   any
		capturedTenantID any
		capturedRealmID  any
	)

	app := createTestApp()

	app.Get("/test", SessionMiddleware(validator, true), func(c *fiber.Ctx) error {
		capturedUserID = c.Locals(ContextKeyUserID)
		capturedTenantID = c.Locals(ContextKeyTenantID)
		capturedRealmID = c.Locals(ContextKeyRealmID)

		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer validtoken")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)
	require.NotNil(t, capturedUserID)
	require.Equal(t, tenantID, capturedTenantID)
	require.Equal(t, realmID, capturedRealmID)
}

func TestSessionMiddleware_BrowserSession_NilUserID(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	validator := &mockSessionValidator{
		browserSession: &cryptoutilAppsTemplateServiceServerRepository.BrowserSession{
			Session: cryptoutilAppsTemplateServiceServerRepository.Session{
				TenantID: tenantID,
				RealmID:  realmID,
			},
			UserID: nil,
		},
	}

	var (
		capturedSession  any
		capturedTenantID any
		capturedRealmID  any
	)

	app := createTestApp()

	app.Get("/test", SessionMiddleware(validator, true), func(c *fiber.Ctx) error {
		capturedSession = c.Locals(ContextKeySession)
		capturedTenantID = c.Locals(ContextKeyTenantID)
		capturedRealmID = c.Locals(ContextKeyRealmID)

		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer validtoken")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)
	require.NotNil(t, capturedSession)
	require.Equal(t, tenantID, capturedTenantID, "tenant_id should be set even when UserID is nil")
	require.Equal(t, realmID, capturedRealmID, "realm_id should be set even when UserID is nil")
}

func TestSessionMiddleware_ServiceSession_ValidationError(t *testing.T) {
	t.Parallel()

	validator := &mockSessionValidator{
		serviceSessionErr: errors.New("invalid token"),
	}
	app := createTestApp()

	app.Get("/test", SessionMiddleware(validator, false), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer validtoken")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 401, resp.StatusCode)
}

func TestSessionMiddleware_ServiceSession_Success(t *testing.T) {
	t.Parallel()

	clientID := googleUuid.New().String()
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	validator := &mockSessionValidator{
		serviceSession: &cryptoutilAppsTemplateServiceServerRepository.ServiceSession{
			Session: cryptoutilAppsTemplateServiceServerRepository.Session{
				TenantID: tenantID,
				RealmID:  realmID,
			},
			ClientID: &clientID,
		},
	}

	var (
		capturedUserID   any
		capturedClientID any
		capturedTenantID any
		capturedRealmID  any
	)

	app := createTestApp()

	app.Get("/test", SessionMiddleware(validator, false), func(c *fiber.Ctx) error {
		capturedUserID = c.Locals(ContextKeyUserID)
		capturedClientID = c.Locals(ContextKeyClientID)
		capturedTenantID = c.Locals(ContextKeyTenantID)
		capturedRealmID = c.Locals(ContextKeyRealmID)

		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer validtoken")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)
	require.NotNil(t, capturedUserID, "user_id should be set when ClientID is valid UUID")
	require.Equal(t, clientID, capturedClientID)
	require.Equal(t, tenantID, capturedTenantID)
	require.Equal(t, realmID, capturedRealmID)
}

func TestSessionMiddleware_ServiceSession_NilClientID(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	validator := &mockSessionValidator{
		serviceSession: &cryptoutilAppsTemplateServiceServerRepository.ServiceSession{
			Session: cryptoutilAppsTemplateServiceServerRepository.Session{
				TenantID: tenantID,
				RealmID:  realmID,
			},
			ClientID: nil,
		},
	}

	var (
		capturedSession  any
		capturedTenantID any
		capturedRealmID  any
	)

	app := createTestApp()

	app.Get("/test", SessionMiddleware(validator, false), func(c *fiber.Ctx) error {
		capturedSession = c.Locals(ContextKeySession)
		capturedTenantID = c.Locals(ContextKeyTenantID)
		capturedRealmID = c.Locals(ContextKeyRealmID)

		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer validtoken")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)
	require.NotNil(t, capturedSession)
	require.Equal(t, tenantID, capturedTenantID, "tenant_id should be set even when ClientID is nil")
	require.Equal(t, realmID, capturedRealmID, "realm_id should be set even when ClientID is nil")
}

func TestBrowserSessionMiddleware(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	validator := &mockSessionValidator{
		browserSession: &cryptoutilAppsTemplateServiceServerRepository.BrowserSession{
			Session: cryptoutilAppsTemplateServiceServerRepository.Session{
				TenantID: tenantID,
				RealmID:  realmID,
			},
		},
	}
	app := createTestApp()

	app.Get("/test", BrowserSessionMiddleware(validator), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer validtoken")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)
}

func TestServiceSessionMiddleware(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	validator := &mockSessionValidator{
		serviceSession: &cryptoutilAppsTemplateServiceServerRepository.ServiceSession{
			Session: cryptoutilAppsTemplateServiceServerRepository.Session{
				TenantID: tenantID,
				RealmID:  realmID,
			},
		},
	}
	app := createTestApp()

	app.Get("/test", ServiceSessionMiddleware(validator), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer validtoken")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)
}

func TestContextKeyConstants(t *testing.T) {
	t.Parallel()

	// Verify context key constants are defined correctly.
	require.Equal(t, "session", ContextKeySession)
	require.Equal(t, "user_id", ContextKeyUserID)
	require.Equal(t, "client_id", ContextKeyClientID)
	require.Equal(t, "tenant_id", ContextKeyTenantID)
	require.Equal(t, "realm_id", ContextKeyRealmID)
}
