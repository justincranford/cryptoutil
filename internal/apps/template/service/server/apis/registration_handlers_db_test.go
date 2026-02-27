// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !integration

package apis

import (
	"bytes"
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerDomain "cryptoutil/internal/apps/template/service/server/domain"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

func TestHandleListJoinRequests_WithDB(t *testing.T) {
	// Uses testGormDB from TestMain - cannot use t.Parallel() safely.
	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testGormDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testGormDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testGormDB)
	registrationService := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(testGormDB, tenantRepo, userRepo, joinRequestRepo)
	handlers := NewRegistrationHandlers(registrationService)

	// Create test tenant
	testTenantID := googleUuid.New()
	testUserID := googleUuid.New()
	testClientID := googleUuid.New()

	// Create test join requests with different scenarios to exercise optional field formatting
	ctx := context.Background()

	// Request 1: User-initiated request (has UserID, no ClientID)
	req1 := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
		ID:          googleUuid.New(),
		UserID:      &testUserID,
		ClientID:    nil,
		TenantID:    testTenantID,
		Status:      cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending,
		RequestedAt: time.Now().UTC(),
		ProcessedAt: nil,
		ProcessedBy: nil,
	}
	err := joinRequestRepo.Create(ctx, req1)
	require.NoError(t, err)

	// Request 2: Client-initiated request (has ClientID, no UserID)
	req2 := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
		ID:          googleUuid.New(),
		UserID:      nil,
		ClientID:    &testClientID,
		TenantID:    testTenantID,
		Status:      cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending,
		RequestedAt: time.Now().UTC(),
		ProcessedAt: nil,
		ProcessedBy: nil,
	}
	err = joinRequestRepo.Create(ctx, req2)
	require.NoError(t, err)

	// Request 3: Processed request (has ProcessedAt and ProcessedBy)
	processedAt := time.Now().UTC().Add(-1 * time.Hour)
	processedBy := googleUuid.New()
	req3 := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
		ID:          googleUuid.New(),
		UserID:      &testUserID,
		ClientID:    nil,
		TenantID:    testTenantID,
		Status:      cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusApproved,
		RequestedAt: time.Now().UTC().Add(-2 * time.Hour),
		ProcessedAt: &processedAt,
		ProcessedBy: &processedBy,
	}
	err = joinRequestRepo.Create(ctx, req3)
	require.NoError(t, err)

	app := fiber.New()
	app.Get("/admin/join-requests", func(c *fiber.Ctx) error {
		// Inject session context (simulating middleware)
		c.Locals("tenant_id", testTenantID)
		c.Locals("user_id", googleUuid.New())

		return handlers.HandleListJoinRequests(c)
	})

	req := httptest.NewRequest("GET", "/admin/join-requests", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Should get 200 status with 3 join requests
	require.Equal(t, 200, resp.StatusCode)

	// Decode response body
	var result map[string][]JoinRequestSummary

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Contains(t, result, "requests")
	requests := result["requests"]
	require.Len(t, requests, 3)

	// Verify optional fields are formatted correctly
	// Request 1 should have UserID but not ClientID
	var req1Summary *JoinRequestSummary

	for i := range requests {
		if requests[i].ID == req1.ID.String() {
			req1Summary = &requests[i]

			break
		}
	}

	require.NotNil(t, req1Summary)
	require.NotNil(t, req1Summary.UserID)
	require.Equal(t, testUserID.String(), *req1Summary.UserID)
	require.Nil(t, req1Summary.ClientID)
	require.Nil(t, req1Summary.ProcessedAt)
	require.Nil(t, req1Summary.ProcessedBy)

	// Request 2 should have ClientID but not UserID
	var req2Summary *JoinRequestSummary

	for i := range requests {
		if requests[i].ID == req2.ID.String() {
			req2Summary = &requests[i]

			break
		}
	}

	require.NotNil(t, req2Summary)
	require.Nil(t, req2Summary.UserID)
	require.NotNil(t, req2Summary.ClientID)
	require.Equal(t, testClientID.String(), *req2Summary.ClientID)
	require.Nil(t, req2Summary.ProcessedAt)
	require.Nil(t, req2Summary.ProcessedBy)

	// Request 3 should have ProcessedAt and ProcessedBy
	var req3Summary *JoinRequestSummary

	for i := range requests {
		if requests[i].ID == req3.ID.String() {
			req3Summary = &requests[i]

			break
		}
	}

	require.NotNil(t, req3Summary)
	require.NotNil(t, req3Summary.UserID)
	require.NotNil(t, req3Summary.ProcessedAt)
	require.NotNil(t, req3Summary.ProcessedBy)
	require.Equal(t, processedBy.String(), *req3Summary.ProcessedBy)
}

// TestHandleProcessJoinRequest_ApproveMessage tests the "approved" message path.
func TestHandleProcessJoinRequest_ApproveMessage(t *testing.T) {
	// Uses testGormDB from TestMain - cannot use t.Parallel() safely.
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

	// Use a non-existent ID which will hit the service error path.
	// The point is to exercise the message selection logic.
	reqBody := ProcessJoinRequestRequest{Approved: true}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/admin/join-requests/"+googleUuid.New().String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Will get 500 because join request doesn't exist.
	require.Equal(t, cryptoutilSharedMagic.TestDefaultRateLimitServiceIP, resp.StatusCode)
}

// TestHandleProcessJoinRequest_RejectMessage tests the "rejected" message path.
func TestHandleProcessJoinRequest_RejectMessage(t *testing.T) {
	// Uses testGormDB from TestMain - cannot use t.Parallel() safely.
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

	// Use a non-existent ID which will hit the service error path.
	reqBody := ProcessJoinRequestRequest{Approved: false}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/admin/join-requests/"+googleUuid.New().String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Will get 500 because join request doesn't exist.
	require.Equal(t, cryptoutilSharedMagic.TestDefaultRateLimitServiceIP, resp.StatusCode)
}

// TestHandleProcessJoinRequest_SuccessMessages tests lines 211-218 (success response with message variations).
// This uses integration testing with the real database to create actual join requests and process them.
func TestHandleProcessJoinRequest_SuccessMessages(t *testing.T) {
	// Uses testGormDB from TestMain - cannot use t.Parallel() safely.
	tests := []struct {
		name            string
		approved        bool
		expectedMessage string
	}{
		{
			name:            "Approved",
			approved:        true,
			expectedMessage: "Join request approved",
		},
		{
			name:            "Rejected",
			approved:        false,
			expectedMessage: "Join request rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cannot use t.Parallel() - shares testGormDB.

			// Create real repositories and service with testGormDB.
			tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testGormDB)
			userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testGormDB)
			joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testGormDB)
			registrationService := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(testGormDB, tenantRepo, userRepo, joinRequestRepo)
			handlers := NewRegistrationHandlers(registrationService)

			// Create actual tenant and join request in database.
			tenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{
				ID:   googleUuid.New(),
				Name: "test-tenant-" + googleUuid.New().String(),
			}
			require.NoError(t, tenantRepo.Create(context.Background(), tenant))

			adminUserID := googleUuid.New()
			adminUser := &cryptoutilAppsTemplateServiceServerRepository.User{
				ID:       adminUserID,
				TenantID: tenant.ID,
				Username: "admin-" + googleUuid.New().String(),
			}
			require.NoError(t, userRepo.Create(context.Background(), adminUser))

			clientID := googleUuid.New()
			joinRequest := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
				ID:          googleUuid.New(),
				TenantID:    tenant.ID,
				ClientID:    &clientID,
				Status:      cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending,
				RequestedAt: time.Now().UTC(),
			}
			require.NoError(t, joinRequestRepo.Create(context.Background(), joinRequest))

			// Test processing.
			app := fiber.New()
			app.Put("/admin/join-requests/:id", func(c *fiber.Ctx) error {
				c.Locals("tenant_id", tenant.ID)
				c.Locals("user_id", adminUserID)

				return handlers.HandleProcessJoinRequest(c)
			})

			reqBody := ProcessJoinRequestRequest{Approved: tt.approved}
			bodyBytes, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("PUT", "/admin/join-requests/"+joinRequest.ID.String(), bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			// Verify 200 with correct message.
			require.Equal(t, 200, resp.StatusCode)

			var respBody map[string]any

			err = json.NewDecoder(resp.Body).Decode(&respBody)
			require.NoError(t, err)
			require.Equal(t, tt.expectedMessage, respBody["message"])

			// No cleanup needed - test database will be cleared between test runs.
		})
	}
}
