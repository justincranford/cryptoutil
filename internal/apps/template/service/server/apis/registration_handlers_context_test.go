package apis

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestHandleListJoinRequests_MissingTenantID tests missing tenant_id in context.
func TestHandleListJoinRequests_MissingTenantID(t *testing.T) {
	t.Parallel()

	db := &gorm.DB{}
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(db)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(db)
	registrationService := cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)
	handlers := NewRegistrationHandlers(registrationService)

	app := fiber.New()
	app.Get("/admin/join-requests", func(c *fiber.Ctx) error {
		// Do NOT inject tenant_id to trigger error path
		return handlers.HandleListJoinRequests(c)
	})

	req := httptest.NewRequest("GET", "/admin/join-requests", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 401, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Contains(t, result, "error")
	require.Equal(t, "Missing tenant_id in session context", result["error"])
}

// TestHandleListJoinRequests_InvalidTenantIDType tests invalid tenant_id type in context.
func TestHandleListJoinRequests_InvalidTenantIDType(t *testing.T) {
	t.Parallel()

	db := &gorm.DB{}
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(db)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(db)
	registrationService := cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)
	handlers := NewRegistrationHandlers(registrationService)

	app := fiber.New()
	app.Get("/admin/join-requests", func(c *fiber.Ctx) error {
		// Inject wrong type to trigger error path
		c.Locals("tenant_id", "not-a-uuid-type")
		return handlers.HandleListJoinRequests(c)
	})

	req := httptest.NewRequest("GET", "/admin/join-requests", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 500, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Contains(t, result, "error")
	require.Equal(t, "Invalid tenant_id type in context", result["error"])
}

// TestHandleProcessJoinRequest_MissingUserID tests missing user_id in context.
func TestHandleProcessJoinRequest_MissingUserID(t *testing.T) {
	t.Parallel()

	db := &gorm.DB{}
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(db)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(db)
	registrationService := cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)
	handlers := NewRegistrationHandlers(registrationService)

	app := fiber.New()
	app.Put("/admin/join-requests/:id", func(c *fiber.Ctx) error {
		// Do NOT inject user_id to trigger error path
		return handlers.HandleProcessJoinRequest(c)
	})

	validID := googleUuid.New().String()
	reqBody := ProcessJoinRequestRequest{Approved: true}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/admin/join-requests/"+validID, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 401, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Contains(t, result, "error")
	require.Equal(t, "Missing user_id in session context", result["error"])
}

// TestHandleProcessJoinRequest_InvalidUserIDType tests invalid user_id type in context.
func TestHandleProcessJoinRequest_InvalidUserIDType(t *testing.T) {
	t.Parallel()

	db := &gorm.DB{}
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(db)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(db)
	registrationService := cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)
	handlers := NewRegistrationHandlers(registrationService)

	app := fiber.New()
	app.Put("/admin/join-requests/:id", func(c *fiber.Ctx) error {
		// Inject wrong type to trigger error path
		c.Locals("user_id", "not-a-uuid-type")
		return handlers.HandleProcessJoinRequest(c)
	})

	validID := googleUuid.New().String()
	reqBody := ProcessJoinRequestRequest{Approved: true}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/admin/join-requests/"+validID, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 500, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Contains(t, result, "error")
	require.Equal(t, "Invalid user_id type in context", result["error"])
}
