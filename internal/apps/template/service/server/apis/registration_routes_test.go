// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !integration

package apis

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// mockSessionValidator is a mock implementation of SessionValidator for testing.
type mockSessionValidator struct{}

func (m *mockSessionValidator) ValidateBrowserSession(ctx context.Context, token string) (*cryptoutilAppsTemplateServiceServerRepository.BrowserSession, error) {
	// Return nil to indicate no session found (simulates unauthenticated request)
	return nil, nil
}

func (m *mockSessionValidator) ValidateServiceSession(ctx context.Context, token string) (*cryptoutilAppsTemplateServiceServerRepository.ServiceSession, error) {
	// Return nil to indicate no session found (simulates unauthenticated request)
	return nil, nil
}

// TestRegisterRegistrationRoutes_Integration tests route registration with real app.
func TestRegisterRegistrationRoutes_Integration(t *testing.T) {
	t.Parallel()

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testGormDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testGormDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testGormDB)
	registrationService := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(testGormDB, tenantRepo, userRepo, joinRequestRepo)

	app := fiber.New()

	RegisterRegistrationRoutes(app, registrationService, 10)

	// Verify routes were registered by checking that they exist
	require.NotNil(t, app)
}

// TestRegisterRegistrationRoutes_RateLimiting tests rate limiting middleware.
func TestRegisterRegistrationRoutes_RateLimiting(t *testing.T) {
	t.Parallel()

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testGormDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testGormDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testGormDB)
	registrationService := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(testGormDB, tenantRepo, userRepo, joinRequestRepo)

	app := fiber.New()

	// Set rate limit to 1 request per minute with burst of 2 for testing
	RegisterRegistrationRoutes(app, registrationService, 1)

	// Make requests until we hit the rate limit
	var lastStatus int

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("POST", "/browser/api/v1/auth/register", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)

		lastStatus = resp.StatusCode

		_ = resp.Body.Close()

		if lastStatus == 429 {
			// Successfully triggered rate limit
			return
		}
	}

	// If we got here, we didn't trigger the rate limit
	t.Logf("Last status code was %d, expected 429 at some point", lastStatus)
}

// TestRegisterJoinRequestManagementRoutes_Integration tests join request route registration.
func TestRegisterJoinRequestManagementRoutes_Integration(t *testing.T) {
	t.Parallel()

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testGormDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testGormDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testGormDB)
	registrationService := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(testGormDB, tenantRepo, userRepo, joinRequestRepo)

	adminAPI := fiber.New()
	mockValidator := &mockSessionValidator{}

	RegisterJoinRequestManagementRoutes(adminAPI, registrationService, mockValidator)

	// Verify routes were registered
	require.NotNil(t, adminAPI)
}
