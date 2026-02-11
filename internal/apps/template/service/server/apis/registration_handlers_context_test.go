package apis

import (
	"bytes"
	json "encoding/json"
	"net/http/httptest"
	"testing"

	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestHandleListJoinRequests_ValidationErrors tests context validation for HandleListJoinRequests.
func TestHandleListJoinRequests_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setupLocals      func(*fiber.Ctx)
		wantStatusCode   int
		wantErrorMessage string
	}{
		{
			name: "missing tenant_id",
			setupLocals: func(c *fiber.Ctx) {
				// Do NOT inject tenant_id to trigger error path
			},
			wantStatusCode:   401,
			wantErrorMessage: "Missing tenant_id in session context",
		},
		{
			name: "invalid tenant_id type",
			setupLocals: func(c *fiber.Ctx) {
				// Inject wrong type to trigger error path
				c.Locals("tenant_id", "not-a-uuid-type")
			},
			wantStatusCode:   500,
			wantErrorMessage: "Invalid tenant_id type in context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := &gorm.DB{}
			tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(db)
			userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(db)
			joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(db)
			registrationService := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)
			handlers := NewRegistrationHandlers(registrationService)

			app := fiber.New()
			app.Get("/admin/join-requests", func(c *fiber.Ctx) error {
				tt.setupLocals(c)

				return handlers.HandleListJoinRequests(c)
			})

			req := httptest.NewRequest("GET", "/admin/join-requests", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, tt.wantStatusCode, resp.StatusCode)

			var result map[string]string

			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			require.Contains(t, result, "error")
			require.Equal(t, tt.wantErrorMessage, result["error"])
		})
	}
}

// TestHandleProcessJoinRequest_ValidationErrors tests context validation for HandleProcessJoinRequest.
func TestHandleProcessJoinRequest_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setupLocals      func(*fiber.Ctx)
		wantStatusCode   int
		wantErrorMessage string
	}{
		{
			name: "missing user_id",
			setupLocals: func(c *fiber.Ctx) {
				// Do NOT inject user_id to trigger error path
			},
			wantStatusCode:   401,
			wantErrorMessage: "Missing user_id in session context",
		},
		{
			name: "invalid user_id type",
			setupLocals: func(c *fiber.Ctx) {
				// Inject wrong type to trigger error path
				c.Locals("user_id", "not-a-uuid-type")
			},
			wantStatusCode:   500,
			wantErrorMessage: "Invalid user_id type in context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := &gorm.DB{}
			tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(db)
			userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(db)
			joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(db)
			registrationService := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)
			handlers := NewRegistrationHandlers(registrationService)

			app := fiber.New()
			app.Put("/admin/join-requests/:id", func(c *fiber.Ctx) error {
				tt.setupLocals(c)

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

			require.Equal(t, tt.wantStatusCode, resp.StatusCode)

			var result map[string]string

			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			require.Contains(t, result, "error")
			require.Equal(t, tt.wantErrorMessage, result["error"])
		})
	}
}
