// Copyright (c) 2025-2026 Justin Cranford.
//go:build e2e

package e2e_test

import (
	"bytes"
	"context"
	json "encoding/json"
	"fmt"
	http "net/http"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
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

			ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
			defer cancel()

			// Generate unique user credentials.
			uniqueSuffix := time.Now().UTC().UnixNano()
			username := fmt.Sprintf("tenant_owner_%d", uniqueSuffix)
			email := fmt.Sprintf("tenant_owner_%d@test.local", uniqueSuffix)
			tenantName := fmt.Sprintf("tenant_%d", uniqueSuffix)
			password := generateTestPassword(t)

			// Determine API path prefix based on client type.
			pathPrefix := cryptoutilSharedMagic.PathPrefixService
			if tt.useBrowser {
				pathPrefix = cryptoutilSharedMagic.PathPrefixBrowser
			}

			// Register user with create_tenant=true (automatic tenant creation).
			registerURL := tt.publicURL + pathPrefix + cryptoutilSharedMagic.IMAPV1AuthRegister
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
			// Planned: Parse response JSON to extract tenant_id and verify it's returned.
			// For now, just verify 201 status indicates success.
		})
	}
}

// TestE2E_RegistrationFlowWithJoinRequest validates user registration with join request to existing tenant.
// This tests the Phase 0 join request authorization workflow.
//
// SKIPPED: The framework tenant registration handler does not yet accept join_tenant_id.
// RegisterUserRequest currently supports create_tenant=true and tenant_name for tenant creation.
// Re-enable when framework registration request model adds a join-by-tenant-ID path.
func TestE2E_RegistrationFlowWithJoinRequest(t *testing.T) {
	t.Skip("framework registration handler does not accept join_tenant_id yet (create_tenant flow only)")

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

			ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
			defer cancel()

			// Determine API path prefix.
			pathPrefix := cryptoutilSharedMagic.PathPrefixService
			if tt.useBrowser {
				pathPrefix = cryptoutilSharedMagic.PathPrefixBrowser
			}

			// Step 1: Create a tenant (first user).
			uniqueSuffix := time.Now().UTC().UnixNano()
			tenantOwner := fmt.Sprintf("owner_%d", uniqueSuffix)
			ownerEmail := fmt.Sprintf("owner_%d@test.local", uniqueSuffix)
			tenantName := fmt.Sprintf("tenant_%d", uniqueSuffix)
			ownerPassword := generateTestPassword(t)

			ownerRegisterURL := tt.publicURL + pathPrefix + cryptoutilSharedMagic.IMAPV1AuthRegister
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

			// Planned: Parse response to get tenant_id.
			// For this E2E test, we'll use a placeholder tenant_id and expect 400 for now.
			// In real implementation, we'd extract tenant_id from owner registration response.

			// Step 2: Second user attempts to join the tenant (creates join request).
			joinerUsername := fmt.Sprintf("joiner_%d", time.Now().UTC().UnixNano())
			joinerPassword := generateTestPassword(t)
			placeholderTenantID := googleUuid.Nil.String() // Placeholder until we parse response.

			joinerRegisterURL := tt.publicURL + pathPrefix + cryptoutilSharedMagic.IMAPV1AuthRegister
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
// SKIPPED: Admin join request routes are on the private admin listener (/admin/api/v1/join-requests).
// The admin listener is intentionally not exposed to host ports in E2E compose, so host-based
// E2E tests cannot call these endpoints directly without an in-container exec path.
// Sm-im DOES use framework admin server infrastructure; this skip is about reachability from host tests.
//
// If admin functionality is needed for sm-im:
// 1. Add an E2E helper that executes requests from inside the app container or test sidecar
// 2. Authenticate and call /admin/api/v1/join-requests via the private admin listener
// 3. Re-enable this test with private-plane assertions.
func TestE2E_AdminJoinRequestManagement(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
	defer cancel()

	// Step 1: Create a tenant owner through the public browser path.
	uniqueSuffix := time.Now().UTC().UnixNano()
	username := fmt.Sprintf("admin_join_owner_%d", uniqueSuffix)
	email := fmt.Sprintf("admin_join_owner_%d@test.local", uniqueSuffix)
	tenantName := fmt.Sprintf("admin_join_tenant_%d", uniqueSuffix)
	password := generateTestPassword(t)

	registerURL := sqlitePublicURL + cryptoutilSharedMagic.PathPrefixBrowser + cryptoutilSharedMagic.IMAPV1AuthRegister
	registerBody := fmt.Sprintf(`{
		"username": "%s",
		"email": "%s",
		"password": "%s",
		"tenant_name": "%s",
		"create_tenant": true
	}`, username, email, password, tenantName)

	registerReq, err := http.NewRequestWithContext(ctx, http.MethodPost, registerURL, bytes.NewBufferString(registerBody))
	require.NoError(t, err, "Creating registration request should succeed")
	registerReq.Header.Set("Content-Type", "application/json")

	registerResp, err := sharedHTTPClient.Do(registerReq)
	require.NoError(t, err, "Owner registration should succeed")

	defer func() { _ = registerResp.Body.Close() }()

	require.Equal(t, http.StatusCreated, registerResp.StatusCode, "Owner registration should return 201")

	var registerResult map[string]any
	require.NoError(t, json.NewDecoder(registerResp.Body).Decode(&registerResult), "Registration response JSON should decode")

	userID, ok := registerResult["user_id"].(string)
	require.True(t, ok && userID != "", "Registration response should include user_id")

	tenantID, ok := registerResult["tenant_id"].(string)
	require.True(t, ok && tenantID != "", "Registration response should include tenant_id")

	// Step 2: Issue a browser session token for admin route authentication.
	realmID := googleUuid.NewString()
	sessionIssueURL := sqlitePublicURL + "/browser/api/v1/sessions/issue"
	sessionIssueBody := fmt.Sprintf(`{
		"user_id": "%s",
		"tenant_id": "%s",
		"realm_id": "%s",
		"session_type": "browser"
	}`, userID, tenantID, realmID)

	sessionReq, err := http.NewRequestWithContext(ctx, http.MethodPost, sessionIssueURL, bytes.NewBufferString(sessionIssueBody))
	require.NoError(t, err, "Creating session issue request should succeed")
	sessionReq.Header.Set("Content-Type", "application/json")

	sessionResp, err := sharedHTTPClient.Do(sessionReq)
	require.NoError(t, err, "Session issue request should succeed")

	defer func() { _ = sessionResp.Body.Close() }()

	require.Equal(t, http.StatusOK, sessionResp.StatusCode, "Session issue should return 200")

	var sessionIssueResult map[string]any
	require.NoError(t, json.NewDecoder(sessionResp.Body).Decode(&sessionIssueResult), "Session response JSON should decode")

	token, ok := sessionIssueResult[cryptoutilSharedMagic.ParamToken].(string)
	require.True(t, ok && token != "", "Session issue response should include token")

	// Step 3: Call admin join-request route from inside app container (private admin listener).
	// BusyBox wget in the runtime image does not support client cert flags; install curl once as root for this E2E probe.
	_ = execComposeInContainerAsUser(t, sqliteContainer, "0", "apk", "add", "--no-cache", "curl")

	adminBody := execComposeInContainer(t,
		sqliteContainer,
		"curl",
		"--silent",
		"--show-error",
		"--fail",
		cryptoutilSharedMagic.CLICACertFlag,
		"/certs/issuing-ca.pem",
		cryptoutilSharedMagic.CLICertFlag,
		"/certs/sm-im/private-https-mutual-entity-sm-im-sqlite-1/private-https-mutual-entity-sm-im-sqlite-1.crt",
		cryptoutilSharedMagic.CLIKeyFlag,
		"/certs/sm-im/private-https-mutual-entity-sm-im-sqlite-1/private-https-mutual-entity-sm-im-sqlite-1.key",
		"--header",
		"Authorization: Bearer "+token,
		"https://127.0.0.1:9090/admin/api/v1/join-requests",
	)

	var adminResult map[string]any
	require.NoError(t, json.Unmarshal([]byte(adminBody), &adminResult), "Admin response should be valid JSON")

	requests, exists := adminResult["requests"]
	require.True(t, exists, "Admin response should include requests field")

	_, ok = requests.([]any)
	require.True(t, ok, "requests field should be an array")
}

func execComposeInContainer(t *testing.T, service string, command ...string) string {
	t.Helper()

	return execComposeInContainerAsUser(t, service, "", command...)
}

func execComposeInContainerAsUser(t *testing.T, service, user string, command ...string) string {
	t.Helper()

	require.NotNil(t, composeManager, "compose manager must be initialized")

	args := composeManager.BuildDockerExecArgs(service, command...)
	if user != "" {
		execIndex := slices.Index(args, "exec")
		require.NotEqual(t, -1, execIndex, "docker compose exec command must contain exec subcommand")
		args = append(args[:execIndex+1], append([]string{"-u", user}, args[execIndex+1:]...)...)
	}

	cmd := exec.Command("docker", args...)

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "docker compose exec failed: %s", strings.TrimSpace(string(output)))

	return strings.TrimSpace(string(output))
}
