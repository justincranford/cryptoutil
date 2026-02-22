// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !integration

package apis

import (
	"bytes"
	json "encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestHandleRegisterUser_ValidationError tests that validateRegistrationRequest
// error propagates as 400 (covers registration_handlers.go:71-75).
//
// NOTE: NOT parallel — modifies package-level injectable var.
func TestHandleRegisterUser_ValidationError(t *testing.T) {
	db := &gorm.DB{}
	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(db)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(db)
	joinRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(db)
	registrationService := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRepo)
	handlers := NewRegistrationHandlers(registrationService)

	app := fiber.New()
	app.Post("/register", handlers.HandleRegisterUser)

	// username too short — triggers validateRegistrationRequest error → 400
	body := RegisterUserRequest{
		Username:   strings.Repeat("a", cryptoutilSharedMagic.CipherMinUsernameLength-1),
		Email:      "user@example.com",
		Password:   strings.Repeat("p", cryptoutilSharedMagic.CipherMinPasswordLength),
		TenantName: strings.Repeat("t", cryptoutilSharedMagic.CipherMinUsernameLength),
	}
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, respErr := app.Test(req, -1)
	require.NoError(t, respErr)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 400, resp.StatusCode)
}

// TestHandleRegisterUser_HashError covers the hash failure path at
// registration_handlers.go:90-93.
//
// NOTE: NOT parallel — modifies package-level injectable var.
func TestHandleRegisterUser_HashError(t *testing.T) {
	orig := registrationHandlersHashSecretPBKDF2Fn
	registrationHandlersHashSecretPBKDF2Fn = func(_ string) (string, error) {
		return "", errors.New("hash failure injected for test")
	}

	defer func() { registrationHandlersHashSecretPBKDF2Fn = orig }()

	db := &gorm.DB{}
	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(db)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(db)
	joinRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(db)
	registrationService := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRepo)
	handlers := NewRegistrationHandlers(registrationService)

	app := fiber.New()
	app.Post("/register", handlers.HandleRegisterUser)

	body := RegisterUserRequest{
		Username:   strings.Repeat("a", cryptoutilSharedMagic.CipherMinUsernameLength),
		Email:      "user@example.com",
		Password:   strings.Repeat("p", cryptoutilSharedMagic.CipherMinPasswordLength),
		TenantName: strings.Repeat("t", cryptoutilSharedMagic.CipherMinUsernameLength),
	}
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, respErr := app.Test(req, -1)
	require.NoError(t, respErr)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 500, resp.StatusCode)
}

// TestHandleListJoinRequests_InvalidTenantIDType covers the type assertion
// failure path at registration_handlers.go:179-183.
func TestHandleListJoinRequests_InvalidTenantIDType(t *testing.T) {
	t.Parallel()

	db := &gorm.DB{}
	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(db)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(db)
	joinRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(db)
	registrationService := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRepo)
	handlers := NewRegistrationHandlers(registrationService)

	app := fiber.New()
	app.Get("/join-requests", func(c *fiber.Ctx) error {
		// Set tenant_id as a string (wrong type — should be googleUuid.UUID).
		c.Locals("tenant_id", "not-a-uuid-type")

		return c.Next()
	}, handlers.HandleListJoinRequests)

	req := httptest.NewRequest("GET", "/join-requests", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 500, resp.StatusCode)
}

// TestHandleRegisterUser_ServiceError covers the service error path at
// registration_handlers.go:93-97 (RegisterUserWithTenant returns error).
//
// Uses a closed database to force the service to return a database error.
func TestHandleRegisterUser_ServiceError(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	// Create a real SQLite DB, migrate schema, then close it to force errors.
	closedDB, initErr := cryptoutilAppsTemplateServiceServerRepository.InitSQLite(
		ctx,
		"file:test-reg-svc-err?mode=memory&cache=private",
		cryptoutilAppsTemplateServiceServerRepository.MigrationsFS,
	)
	require.NoError(t, initErr)

	sqlDB, sqlErr := closedDB.DB()
	require.NoError(t, sqlErr)
	require.NoError(t, sqlDB.Close()) // Close DB to force all queries to fail.

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(closedDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(closedDB)
	joinRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(closedDB)
	registrationService := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(
		closedDB, tenantRepo, userRepo, joinRepo,
	)
	handlers := NewRegistrationHandlers(registrationService)

	app := fiber.New()
	app.Post("/register", handlers.HandleRegisterUser)

	body := RegisterUserRequest{
		Username:     strings.Repeat("a", cryptoutilSharedMagic.CipherMinUsernameLength),
		Email:        "user@example.com",
		Password:     strings.Repeat("p", cryptoutilSharedMagic.CipherMinPasswordLength),
		TenantName:   strings.Repeat("t", cryptoutilSharedMagic.CipherMinUsernameLength),
		CreateTenant: true,
	}

	bodyBytes, marshalErr := json.Marshal(body)
	require.NoError(t, marshalErr)

	req := httptest.NewRequest("POST", "/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, respErr := app.Test(req, -1)
	require.NoError(t, respErr)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 500, resp.StatusCode)
}

// TestHandleListJoinRequests_ServiceError covers the service error path at
// registration_handlers.go:182-186 (ListJoinRequests returns error).
//
// Uses a closed database to force the service to return a database error.
func TestHandleListJoinRequests_ServiceError(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	// Create a real SQLite DB, migrate schema, then close it to force errors.
	closedDB, initErr := cryptoutilAppsTemplateServiceServerRepository.InitSQLite(
		ctx,
		"file:test-list-svc-err?mode=memory&cache=private",
		cryptoutilAppsTemplateServiceServerRepository.MigrationsFS,
	)
	require.NoError(t, initErr)

	sqlDB, sqlErr := closedDB.DB()
	require.NoError(t, sqlErr)
	require.NoError(t, sqlDB.Close())

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(closedDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(closedDB)
	joinRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(closedDB)
	registrationService := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(
		closedDB, tenantRepo, userRepo, joinRepo,
	)
	handlers := NewRegistrationHandlers(registrationService)

	tenantID := googleUuid.Must(googleUuid.NewV7())

	app := fiber.New()
	app.Get("/join-requests", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handlers.HandleListJoinRequests)

	req := httptest.NewRequest("GET", "/join-requests", nil)
	resp, respErr := app.Test(req, -1)
	require.NoError(t, respErr)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 500, resp.StatusCode)
}
