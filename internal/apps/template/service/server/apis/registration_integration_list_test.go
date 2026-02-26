// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration

package apis

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerDomain "cryptoutil/internal/apps/template/service/server/domain"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"

	// Use modernc SQLite driver (CGO-free).
	_ "modernc.org/sqlite"
)

func TestIntegration_ListJoinRequests_NoRequests(t *testing.T) {
	t.Parallel()

	// Create empty database for this test only
	emptyDB, err := gorm.Open(sqlite.Dialector{
		DriverName: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:        cryptoutilSharedMagic.SQLiteInMemoryDSN,
	}, &gorm.Config{})
	require.NoError(t, err)

	// Run migrations on empty DB
	require.NoError(t, emptyDB.AutoMigrate(&cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{}))

	// Create repositories with empty DB
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(emptyDB)
	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(emptyDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(emptyDB)

	// Create service with empty DB
	svc := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(emptyDB, tenantRepo, userRepo, joinRequestRepo)

	// Create isolated Fiber app with auth middleware
	app := fiber.New()
	testTenantID := googleUuid.New()
	authMiddleware := func(c *fiber.Ctx) error {
		c.Locals("tenant_id", testTenantID)

		return c.Next()
	}
	app.Use(authMiddleware)

	// Create a mock SessionValidator for this isolated test.
	// It bypasses actual session validation and returns a valid session.
	mockValidator := &mockSessionValidatorIntegration{
		tenantID: testTenantID,
		realmID:  googleUuid.New(),
		userID:   googleUuid.NewString(),
	}

	RegisterJoinRequestManagementRoutes(app, svc, mockValidator)

	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/join-requests", nil)
	addAuthHeader(req)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.Contains(t, result, "requests")

	requests, ok := result["requests"].([]any)
	require.True(t, ok)
	require.Equal(t, 0, len(requests))
}

// TestIntegration_ListJoinRequests_WithData tests list with existing requests.
func TestIntegration_ListJoinRequests_WithData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use testTenantID since the auth middleware sets tenant_id from this global.
	// Create tenant if it doesn't exist (FirstOrCreate handles both cases).
	tenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{
		ID:   testTenantID,
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength]),
	}
	require.NoError(t, testDB.WithContext(ctx).Where("id = ?", testTenantID).FirstOrCreate(tenant).Error)

	// Create join request with unique ID.
	userID := googleUuid.New()
	joinReq := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
		ID:       googleUuid.New(),
		TenantID: testTenantID,
		UserID:   &userID,
		Status:   "pending",
	}
	require.NoError(t, testDB.WithContext(ctx).Create(joinReq).Error)

	// Verify join request was created in database.
	var dbCount int64
	require.NoError(t, testDB.Model(&cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{}).Count(&dbCount).Error)
	t.Logf("Total join requests in DB: %d", dbCount)

	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/join-requests", nil)
	addAuthHeader(req)

	resp, err := testJoinRequestMgmtApp.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	t.Logf("Response body: %s", string(bodyBytes))

	var result map[string]any
	require.NoError(t, json.Unmarshal(bodyBytes, &result))
	require.Contains(t, result, "requests")

	requests, ok := result["requests"].([]any)
	require.True(t, ok)
	t.Logf("Returned requests count: %d", len(requests))
	// Note: There might be multiple requests from parallel tests, so just check >= 1
	require.GreaterOrEqual(t, len(requests), 1)

	// Find our specific request in the list
	foundOurRequest := false

	for _, reqItem := range requests {
		reqMap, ok := reqItem.(map[string]any)
		if !ok {
			continue
		}

		if reqMap["id"] == joinReq.ID.String() {
			foundOurRequest = true

			require.Equal(t, "pending", reqMap[cryptoutilSharedMagic.StringStatus])
			require.Equal(t, tenant.ID.String(), reqMap["tenant_id"])

			break
		}
	}

	require.True(t, foundOurRequest, "Our join request should be in the list")
}

// TestIntegration_ListJoinRequests_AllOptionalFields tests list with join request containing all optional fields.
func TestIntegration_ListJoinRequests_AllOptionalFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create or find existing tenant using the same tenant ID from auth middleware.
	tenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{
		ID:   testTenantID,
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength]),
	}
	dbResult := testDB.Where("id = ?", testTenantID).FirstOrCreate(tenant)
	require.NoError(t, dbResult.Error)

	// Create join request with ALL optional fields populated.
	userID := googleUuid.New()
	clientID := googleUuid.New()
	processedBy := googleUuid.New()
	processedAt := time.Now().UTC()

	joinReq := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
		ID:          googleUuid.New(),
		TenantID:    testTenantID,
		UserID:      &userID,
		ClientID:    &clientID, // Optional field 1
		Status:      "approved",
		ProcessedAt: &processedAt, // Optional field 2
		ProcessedBy: &processedBy, // Optional field 3
		RequestedAt: time.Now().UTC().Add(-1 * time.Hour),
	}
	require.NoError(t, testDB.WithContext(ctx).Create(joinReq).Error)

	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/join-requests", nil)
	addAuthHeader(req)

	resp, err := testJoinRequestMgmtApp.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any

	bodyBytes, _ := io.ReadAll(resp.Body)
	require.NoError(t, json.Unmarshal(bodyBytes, &result))
	require.Contains(t, result, "requests")

	requests, ok := result["requests"].([]any)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(requests), 1)

	// Find our specific request and verify optional fields are included.
	foundOurRequest := false

	for _, reqItem := range requests {
		reqMap, ok := reqItem.(map[string]any)
		if !ok {
			continue
		}

		if reqMap["id"] == joinReq.ID.String() {
			foundOurRequest = true

			require.Equal(t, "approved", reqMap[cryptoutilSharedMagic.StringStatus])

			// Verify optional fields are present in response.
			require.Contains(t, reqMap, cryptoutilSharedMagic.ClaimClientID, "ClientID should be present")
			require.Equal(t, clientID.String(), reqMap[cryptoutilSharedMagic.ClaimClientID])

			require.Contains(t, reqMap, "processed_at", "ProcessedAt should be present")
			require.NotNil(t, reqMap["processed_at"])

			require.Contains(t, reqMap, "processed_by", "ProcessedBy should be present")
			require.Equal(t, processedBy.String(), reqMap["processed_by"])

			break
		}
	}

	require.True(t, foundOurRequest, "Our join request with all optional fields should be in the list")
}

// TestIntegration_ProcessJoinRequest_InvalidID tests handling of invalid request IDs.
func TestIntegration_ProcessJoinRequest_InvalidID(t *testing.T) {
	t.Parallel()

	reqBody := `{"approved": true}`
	req := httptest.NewRequest(http.MethodPut, "/admin/api/v1/join-requests/invalid-uuid", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	addAuthHeader(req)

	resp, err := testJoinRequestMgmtApp.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.Contains(t, result, cryptoutilSharedMagic.StringError)
	require.Equal(t, "Invalid request ID", result[cryptoutilSharedMagic.StringError])
}

// TestIntegration_ProcessJoinRequest_InvalidJSON tests handling of malformed JSON in request body.
func TestIntegration_ProcessJoinRequest_InvalidJSON(t *testing.T) {
	t.Parallel()

	validID := googleUuid.New().String()
	req := httptest.NewRequest(http.MethodPut, "/admin/api/v1/join-requests/"+validID, strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	addAuthHeader(req)

	resp, err := testJoinRequestMgmtApp.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.Contains(t, result, cryptoutilSharedMagic.StringError)
	require.Equal(t, "Invalid request body", result[cryptoutilSharedMagic.StringError])
}

// TestIntegration_JoinRequestManagement_Unauthenticated tests that admin endpoints
// return 401 when no Authorization header is provided.
func TestIntegration_JoinRequestManagement_Unauthenticated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "GET /admin/api/v1/join-requests without auth",
			method: http.MethodGet,
			path:   "/admin/api/v1/join-requests",
			body:   "",
		},
		{
			name:   "PUT /admin/api/v1/join-requests/:id without auth",
			method: http.MethodPut,
			path:   fmt.Sprintf("/admin/api/v1/join-requests/%s", googleUuid.NewString()),
			body:   `{"approved":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			// NOTE: Intentionally NOT calling addAuthHeader(req) to test unauthenticated access

			resp, err := testJoinRequestMgmtApp.Test(req, -1)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			// Verify 401 Unauthorized response
			require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Expected 401 Unauthorized for unauthenticated request to %s", tt.path)
		})
	}
}

// TestRegistrationRoutes_MethodNotAllowed tests unsupported HTTP methods return 405.
func TestRegistrationRoutes_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"GET on register endpoint", http.MethodGet, "/browser/api/v1/auth/register"},
		{"DELETE on register endpoint", http.MethodDelete, "/browser/api/v1/auth/register"},
		{"PATCH on register endpoint", http.MethodPatch, "/browser/api/v1/auth/register"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)

			resp, err := testRegistrationApp.Test(req, -1)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		})
	}
}
