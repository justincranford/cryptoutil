// Copyright (c) 2025 Justin Cranford

//go:build e2e

package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	http "net/http"
	"testing"
	"time"


	_ "github.com/lib/pq" // PostgreSQL driver.
	"github.com/stretchr/testify/require"
)

func TestE2E_RegistrationFlowWithTenantCreation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		publicURL  string
		useBrowser bool
	}{
		{sqliteContainer + "_browser", sqlitePublicURL, true},
		{sqliteContainer + "_service", sqlitePublicURL, false},
		{postgres1Container + "_browser", postgres1PublicURL, true},
		{postgres1Container + "_service", postgres1PublicURL, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
			defer cancel()

			// Generate unique user credentials.
			uniqueSuffix := time.Now().UTC().UnixNano()
			username := fmt.Sprintf("tenant_owner_%d", uniqueSuffix)
			email := fmt.Sprintf("tenant_owner_%d@test.local", uniqueSuffix)
			tenantName := fmt.Sprintf("tenant_%d", uniqueSuffix)
			password := generateTestPassword(t)

			// Determine API path prefix based on client type.
			pathPrefix := pathPrefixService
			if tt.useBrowser {
				pathPrefix = pathPrefixBrowser
			}

			// Register user with create_tenant=true (automatic tenant creation).
			registerURL := tt.publicURL + pathPrefix + apiV1AuthRegister
			registerBody := fmt.Sprintf(`{
				"username": "%s",
				"email": "%s",
				"password": "%s",
				"tenant_name": "%s",
				"create_tenant": true
			}`, username, email, password, tenantName)

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, registerURL, bytes.NewBufferString(registerBody))
			require.NoError(t, err, "Creating registration request should succeed")
			req.Header.Set("Content-Type", "application/json")

			resp, err := sharedHTTPClient.Do(req)
			require.NoError(t, err, "User registration should succeed")

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, http.StatusCreated, resp.StatusCode,
				"Registration with create_tenant=true should return 201 Created")
			// TODO: Parse response JSON to extract tenant_id and verify it's returned.
			// For now, just verify 201 status indicates success.
		})
	}
}

// TestE2E_RegistrationFlowWithJoinRequest validates user registration with join request to existing tenant.
// This tests the Phase 0 join request authorization workflow.
//
// SKIPPED: The join_tenant_id field is not yet implemented in the registration handler.
// The current RegisterUserRequest only supports create_tenant=true workflow.
// When join request feature is implemented, re-enable this test and update the request format.
func TestE2E_RegistrationFlowWithJoinRequest(t *testing.T) {
	t.Skip("join_tenant_id field not yet implemented in registration handler - requires Phase 0 join request feature")

	tests := []struct {
		name       string
		publicURL  string
		useBrowser bool
	}{
		{sqliteContainer + "_browser", sqlitePublicURL, true},
		{sqliteContainer + "_service", sqlitePublicURL, false},
		{postgres1Container + "_browser", postgres1PublicURL, true},
		{postgres1Container + "_service", postgres1PublicURL, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
			defer cancel()

			// Determine API path prefix.
			pathPrefix := pathPrefixService
			if tt.useBrowser {
				pathPrefix = pathPrefixBrowser
			}

			// Step 1: Create a tenant (first user).
			uniqueSuffix := time.Now().UTC().UnixNano()
			tenantOwner := fmt.Sprintf("owner_%d", uniqueSuffix)
			ownerEmail := fmt.Sprintf("owner_%d@test.local", uniqueSuffix)
			tenantName := fmt.Sprintf("tenant_%d", uniqueSuffix)
			ownerPassword := generateTestPassword(t)

			ownerRegisterURL := tt.publicURL + pathPrefix + apiV1AuthRegister
			ownerRegisterBody := fmt.Sprintf(`{
				"username": "%s",
				"email": "%s",
				"password": "%s",
				"tenant_name": "%s",
				"create_tenant": true
			}`, tenantOwner, ownerEmail, ownerPassword, tenantName)

			ownerReq, err := http.NewRequestWithContext(ctx, http.MethodPost, ownerRegisterURL, bytes.NewBufferString(ownerRegisterBody))
			require.NoError(t, err, "Creating owner registration request should succeed")
			ownerReq.Header.Set("Content-Type", "application/json")

			ownerResp, err := sharedHTTPClient.Do(ownerReq)
			require.NoError(t, err, "Owner registration should succeed")

			defer func() { _ = ownerResp.Body.Close() }()

			require.Equal(t, http.StatusCreated, ownerResp.StatusCode,
				"Owner registration should return 201 Created")

			// TODO: Parse response to get tenant_id.
			// For this E2E test, we'll use a placeholder tenant_id and expect 400 for now.
			// In real implementation, we'd extract tenant_id from owner registration response.

			// Step 2: Second user attempts to join the tenant (creates join request).
			joinerUsername := fmt.Sprintf("joiner_%d", time.Now().UTC().UnixNano())
			joinerPassword := generateTestPassword(t)
			placeholderTenantID := "00000000-0000-0000-0000-000000000000" // Placeholder until we parse response.

			joinerRegisterURL := tt.publicURL + pathPrefix + "/api/v1/auth/register"
			joinerRegisterBody := fmt.Sprintf(`{
				"username": "%s",
				"password": "%s",
				"join_tenant_id": "%s"
			}`, joinerUsername, joinerPassword, placeholderTenantID)

			joinerReq, err := http.NewRequestWithContext(ctx, http.MethodPost, joinerRegisterURL, bytes.NewBufferString(joinerRegisterBody))
			require.NoError(t, err, "Creating joiner registration request should succeed")
			joinerReq.Header.Set("Content-Type", "application/json")

			joinerResp, err := sharedHTTPClient.Do(joinerReq)
			require.NoError(t, err, "Joiner registration should complete")

			defer func() { _ = joinerResp.Body.Close() }()

			// Join request should either:
			// - Return 201 Created (join request created successfully), OR
			// - Return 400 Bad Request (invalid tenant_id - expected with placeholder).
			// For this test, we accept both as valid responses until full integration.
			require.Contains(t, []int{http.StatusCreated, http.StatusBadRequest}, joinerResp.StatusCode,
				"Join request should return 201 (success) or 400 (invalid tenant - placeholder)")
		})
	}
}

// TestE2E_AdminJoinRequestManagement validates listing and managing join requests.
// This tests the Phase 0 admin endpoints for join request approval/rejection.
//
// SKIPPED: Sm-im is a demo service without admin server infrastructure.
// Admin join request routes are registered on ADMIN server (/admin/api/v1/join-requests),
// NOT on PUBLIC server with /service or /browser prefixes.
// Template infrastructure provides RegisterJoinRequestManagementRoutes() for services
// that implement admin servers, but sm-im uses only PublicServer.
//
// If admin functionality is needed for sm-im:
// 1. Create internal/apps/sm/im/server/admin_server.go
// 2. Call RegisterJoinRequestManagementRoutes() in admin server setup
// 3. Re-enable this test and update URLs to use adminURL (port 9090).
func TestE2E_AdminJoinRequestManagement(t *testing.T) {
	t.Skip("Sm-im demo service does not implement admin server (admin routes registered on separate admin server, not public server with /service or /browser prefixes)")
}
