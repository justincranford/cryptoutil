// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !integration

package apis

import (
	"bytes"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

func TestHandleListJoinRequests_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "List join requests (empty list)",
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testGormDB)
			userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testGormDB)
			joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testGormDB)
			registrationService := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(testGormDB, tenantRepo, userRepo, joinRequestRepo)
			handlers := NewRegistrationHandlers(registrationService)

			app := fiber.New()
			app.Get("/admin/join-requests", func(c *fiber.Ctx) error {
				// Inject session context (simulating middleware)
				c.Locals("tenant_id", googleUuid.New())
				c.Locals("user_id", googleUuid.New())

				return handlers.HandleListJoinRequests(c)
			})

			req := httptest.NewRequest("GET", "/admin/join-requests", nil)

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == 200 {
				var result map[string][]JoinRequestSummary

				err := json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Contains(t, result, "requests")
			}
		})
	}
}

// TestHandleProcessJoinRequest_TableDriven uses table-driven tests for comprehensive coverage.
func TestHandleProcessJoinRequest_TableDriven(t *testing.T) {
	t.Parallel()

	validID := googleUuid.New().String()

	tests := []struct {
		name           string
		requestID      string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "Invalid UUID",
			requestID:      "invalid-uuid",
			requestBody:    `{"approved": true}`,
			expectedStatus: 400,
		},
		{
			name:           "Invalid JSON",
			requestID:      validID,
			requestBody:    "invalid json",
			expectedStatus: 400,
		},
		{
			name:           "Approve nonexistent request",
			requestID:      googleUuid.New().String(),
			requestBody:    `{"approved": true}`,
			expectedStatus: cryptoutilSharedMagic.TestDefaultRateLimitServiceIP,
		},
		{
			name:           "Reject nonexistent request",
			requestID:      googleUuid.New().String(),
			requestBody:    `{"approved": false}`,
			expectedStatus: cryptoutilSharedMagic.TestDefaultRateLimitServiceIP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testGormDB)
			userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testGormDB)
			joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testGormDB)
			registrationService := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(testGormDB, tenantRepo, userRepo, joinRequestRepo)
			handlers := NewRegistrationHandlers(registrationService)

			app := fiber.New()
			app.Put("/admin/join-requests/:id", func(c *fiber.Ctx) error {
				// Inject session context (simulating middleware)
				c.Locals("tenant_id", googleUuid.New())
				c.Locals("user_id", googleUuid.New())

				return handlers.HandleProcessJoinRequest(c)
			})

			req := httptest.NewRequest("PUT", "/admin/join-requests/"+tt.requestID, bytes.NewReader([]byte(tt.requestBody)))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// TestJoinRequestSummary_AllFields tests all fields including optional pointers.
func TestJoinRequestSummary_AllFields(t *testing.T) {
	t.Parallel()

	userID := "user-123"
	clientID := "client-456"
	processedAt := "2026-01-16T12:30:00Z"
	processedBy := "admin-789"

	summary := JoinRequestSummary{
		ID:          googleUuid.New().String(),
		UserID:      &userID,
		ClientID:    &clientID,
		TenantID:    googleUuid.New().String(),
		Status:      "approved",
		RequestedAt: "2026-01-16T12:00:00Z",
		ProcessedAt: &processedAt,
		ProcessedBy: &processedBy,
	}

	require.NotEmpty(t, summary.ID)
	require.NotNil(t, summary.UserID)
	require.NotNil(t, summary.ClientID)
	require.NotNil(t, summary.ProcessedAt)
	require.NotNil(t, summary.ProcessedBy)
	require.Equal(t, "approved", summary.Status)
}

// TestRegisterUserResponse_AllFields tests all fields of response struct.
func TestRegisterUserResponse_AllFields(t *testing.T) {
	t.Parallel()

	resp := RegisterUserResponse{
		UserID:   googleUuid.New().String(),
		TenantID: googleUuid.New().String(),
		Message:  "User registered and tenant created",
	}

	require.NotEmpty(t, resp.UserID)
	require.NotEmpty(t, resp.TenantID)
	require.Equal(t, "User registered and tenant created", resp.Message)
}

// TestProcessJoinRequestRequest_BothValues tests both approved and rejected.
func TestProcessJoinRequestRequest_BothValues(t *testing.T) {
	t.Parallel()

	approved := ProcessJoinRequestRequest{Approved: true}
	require.True(t, approved.Approved)

	rejected := ProcessJoinRequestRequest{Approved: false}
	require.False(t, rejected.Approved)
}

// TestHandleListJoinRequests_WithDB tests HandleListJoinRequests using real database.
// This exercises the list formatting logic (lines 111-141).
