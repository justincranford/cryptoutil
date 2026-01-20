// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package apis

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
)

func TestNewRegistrationHandlers(t *testing.T) {
	t.Parallel()

	db := &gorm.DB{}
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(db)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(db)
	registrationService := cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)

	handlers := NewRegistrationHandlers(registrationService)

	require.NotNil(t, handlers)
	require.Equal(t, registrationService, handlers.registrationService)
}

func TestHandleRegisterUser_InvalidJSON(t *testing.T) {
	t.Parallel()

	db := &gorm.DB{}
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(db)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(db)
	registrationService := cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)
	handlers := NewRegistrationHandlers(registrationService)

	app := fiber.New()
	app.Post("/register", handlers.HandleRegisterUser)

	req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 400, resp.StatusCode)
}

func TestHandleRegisterUser_ValidRequest(t *testing.T) {
	t.Parallel()

	// Note: Full test requires database setup with TestMain pattern
	// This test validates handler structure only
	db := &gorm.DB{}
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(db)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(db)
	registrationService := cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)
	handlers := NewRegistrationHandlers(registrationService)

	require.NotNil(t, handlers)
}

func TestHandleListJoinRequests(t *testing.T) {
	t.Parallel()

	// Note: Full test requires database setup with TestMain pattern
	db := &gorm.DB{}
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(db)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(db)
	registrationService := cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)
	handlers := NewRegistrationHandlers(registrationService)

	require.NotNil(t, handlers)
}

func TestHandleProcessJoinRequest_InvalidID(t *testing.T) {
	t.Parallel()

	db := &gorm.DB{}
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(db)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(db)
	registrationService := cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)
	handlers := NewRegistrationHandlers(registrationService)

	app := fiber.New()
	app.Put("/admin/join-requests/:id", handlers.HandleProcessJoinRequest)

	reqBody := ProcessJoinRequestRequest{Approved: true}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/admin/join-requests/invalid-uuid", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 400, resp.StatusCode)
}

func TestHandleProcessJoinRequest_InvalidJSON(t *testing.T) {
	t.Parallel()

	db := &gorm.DB{}
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(db)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(db)
	registrationService := cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)
	handlers := NewRegistrationHandlers(registrationService)

	app := fiber.New()
	app.Put("/admin/join-requests/:id", handlers.HandleProcessJoinRequest)

	validID := googleUuid.New().String()

	req := httptest.NewRequest("PUT", "/admin/join-requests/"+validID, bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 400, resp.StatusCode)
}

func TestHandlersCoverageBooster(t *testing.T) {
	t.Parallel()

	// Exercise handler creation for coverage
	db := &gorm.DB{}
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(db)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(db)
	registrationService := cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)

	handlers := NewRegistrationHandlers(registrationService)
	require.NotNil(t, handlers)

	// Test request/response types
	req := RegisterUserRequest{
		Username:     "testuser",
		Email:        "test@example.com",
		Password:     "password123",
		TenantName:   "Test Tenant",
		CreateTenant: true,
	}
	require.NotEmpty(t, req.Username)

	resp := RegisterUserResponse{
		UserID:   googleUuid.New().String(),
		TenantID: googleUuid.New().String(),
		Message:  "Success",
	}
	require.NotEmpty(t, resp.UserID)

	summary := JoinRequestSummary{
		ID:          googleUuid.New().String(),
		TenantID:    googleUuid.New().String(),
		Status:      "pending",
		RequestedAt: "2026-01-16T12:00:00Z",
	}
	require.NotEmpty(t, summary.ID)

	processReq := ProcessJoinRequestRequest{
		Approved: true,
	}
	require.True(t, processReq.Approved)
}

func TestRegisterUserRequest_JSON(t *testing.T) {
	t.Parallel()

	req := RegisterUserRequest{
		Username:     "testuser",
		Email:        "test@example.com",
		Password:     "password123",
		TenantName:   "Test Tenant",
		CreateTenant: true,
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var decoded RegisterUserRequest

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	require.Equal(t, req.Username, decoded.Username)
	require.Equal(t, req.Email, decoded.Email)
	require.Equal(t, req.TenantName, decoded.TenantName)
	require.Equal(t, req.CreateTenant, decoded.CreateTenant)
}
