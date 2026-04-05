// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration

package apis

import (
	"bytes"
	"context"
	json "encoding/json"
	"fmt"
	http "net/http"
	"net/http/httptest"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServerDomain "cryptoutil/internal/apps/framework/service/server/domain"
	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps/framework/service/server/repository"

	// Use modernc SQLite driver (CGO-free).
	_ "modernc.org/sqlite"
)

func TestIntegration_ListJoinRequests(t *testing.T) {
	t.Skip("Join request management requires join flow to be implemented first")
	t.Parallel()

	// Create tenant.
	tenant := &cryptoutilAppsFrameworkServiceServerRepository.Tenant{
		ID:   googleUuid.New(),
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength]),
	}
	require.NoError(t, testDB.Create(tenant).Error)

	// Create join requests.
	userID1 := googleUuid.New()
	userID2 := googleUuid.New()
	jr1 := &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		UserID:   &userID1,
		Status:   "pending",
	}
	jr2 := &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		UserID:   &userID2,
		Status:   "pending",
	}

	require.NoError(t, testDB.Create(jr1).Error)
	require.NoError(t, testDB.Create(jr2).Error)

	// List join requests.
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/join-requests", nil)
	addAuthHeader(req)
	resp, err := testJoinRequestMgmtApp.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.Contains(t, result, "requests")

	requests, ok := result["requests"].([]any)
	require.True(t, ok, "requests field should be type []any")
	require.GreaterOrEqual(t, len(requests), 2, "Should have at least 2 join requests")
}

func TestIntegration_ProcessJoinRequest_Approve(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create tenant.
	tenant := &cryptoutilAppsFrameworkServiceServerRepository.Tenant{
		ID:   googleUuid.New(),
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength]),
	}
	require.NoError(t, testDB.Create(tenant).Error)

	// Create join request.
	userID := googleUuid.New()
	jr := &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		UserID:   &userID,
		Status:   "pending",
	}
	require.NoError(t, testDB.Create(jr).Error)

	// Approve join request.
	reqBody := ProcessJoinRequestRequest{
		Approved: true,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/admin/api/v1/join-requests/%s", jr.ID), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	addAuthHeader(req)

	resp, err := testJoinRequestMgmtApp.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify status updated.
	var updated cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest
	require.NoError(t, testDB.WithContext(ctx).First(&updated, "id = ?", jr.ID).Error)
	require.Equal(t, "approved", updated.Status)
}

func TestIntegration_ProcessJoinRequest_Reject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create tenant.
	tenant := &cryptoutilAppsFrameworkServiceServerRepository.Tenant{
		ID:   googleUuid.New(),
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength]),
	}
	require.NoError(t, testDB.Create(tenant).Error)

	// Create join request.
	userID := googleUuid.New()
	jr := &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		UserID:   &userID,
		Status:   "pending",
	}
	require.NoError(t, testDB.Create(jr).Error)

	// Reject join request.
	reqBody := ProcessJoinRequestRequest{
		Approved: false,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/admin/api/v1/join-requests/%s", jr.ID), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	addAuthHeader(req)

	resp, err := testJoinRequestMgmtApp.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify status updated.
	var updated cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest
	require.NoError(t, testDB.WithContext(ctx).First(&updated, "id = ?", jr.ID).Error)
	require.Equal(t, "rejected", updated.Status)
}

func TestIntegration_DuplicateUsername_SameTenant(t *testing.T) {
	t.Skip("Join existing tenant flow not yet implemented in service")
	t.Parallel()

	ctx := context.Background()

	// Create tenant.
	tenant := &cryptoutilAppsFrameworkServiceServerRepository.Tenant{
		ID:   googleUuid.New(),
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength]),
	}
	require.NoError(t, testDB.Create(tenant).Error)

	username := fmt.Sprintf("user_%s", googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	// Create first join request.
	userID1 := googleUuid.New()
	jr1 := &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		UserID:   &userID1,
		Status:   "pending",
	}
	require.NoError(t, testDB.Create(jr1).Error)

	// Try to create second join request with same username.
	reqBody := RegisterUserRequest{
		Username:     username,
		Password:     "SecurePass123!",
		Email:        fmt.Sprintf("%s@example.com", username),
		CreateTenant: false,
		TenantName:   tenant.Name,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/browser/api/v1/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testRegistrationApp.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	// Should still succeed (duplicate check happens during approval).
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify two join requests exist.
	var joinRequests []cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest
	require.NoError(t, testDB.WithContext(ctx).Where("tenant_id = ?", tenant.ID).Find(&joinRequests).Error)
	require.GreaterOrEqual(t, len(joinRequests), 2, "Should have at least 2 join requests (duplicate checking deferred to approval)")
}

// TestIntegration_RegisterUser_InvalidJSON tests HandleRegisterUser with malformed JSON.
func TestIntegration_RegisterUser_InvalidJSON(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/browser/api/v1/auth/register", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testRegistrationApp.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.Contains(t, result, cryptoutilSharedMagic.StringError)
	require.Contains(t, result[cryptoutilSharedMagic.StringError], "Invalid request body")
}

// TestIntegration_ListJoinRequests_NoRequests tests list when no requests exist.
// Note: This test creates its own isolated Fiber app to ensure no state pollution.
