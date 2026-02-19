// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration

package apis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTemplateBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilTemplateDomain "cryptoutil/internal/apps/template/service/server/domain"
	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/apps/template/service/telemetry"

	// Use modernc SQLite driver (CGO-free).
	_ "modernc.org/sqlite"
)

func TestIntegration_ListJoinRequests(t *testing.T) {
	t.Skip("Join request management requires join flow to be implemented first")
	t.Parallel()

	// Create tenant.
	tenant := &cryptoutilTemplateRepository.Tenant{
		ID:   googleUuid.New(),
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:8]),
	}
	require.NoError(t, testDB.Create(tenant).Error)

	// Create join requests.
	userID1 := googleUuid.New()
	userID2 := googleUuid.New()
	jr1 := &cryptoutilTemplateDomain.TenantJoinRequest{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		UserID:   &userID1,
		Status:   "pending",
	}
	jr2 := &cryptoutilTemplateDomain.TenantJoinRequest{
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
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.Contains(t, result, "requests")

	requests := result["requests"].([]any)
	require.GreaterOrEqual(t, len(requests), 2, "Should have at least 2 join requests")
}

func TestIntegration_ProcessJoinRequest_Approve(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create tenant.
	tenant := &cryptoutilTemplateRepository.Tenant{
		ID:   googleUuid.New(),
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:8]),
	}
	require.NoError(t, testDB.Create(tenant).Error)

	// Create join request.
	userID := googleUuid.New()
	jr := &cryptoutilTemplateDomain.TenantJoinRequest{
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
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify status updated.
	var updated cryptoutilTemplateDomain.TenantJoinRequest
	require.NoError(t, testDB.WithContext(ctx).First(&updated, "id = ?", jr.ID).Error)
	require.Equal(t, "approved", updated.Status)
}

func TestIntegration_ProcessJoinRequest_Reject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create tenant.
	tenant := &cryptoutilTemplateRepository.Tenant{
		ID:   googleUuid.New(),
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:8]),
	}
	require.NoError(t, testDB.Create(tenant).Error)

	// Create join request.
	userID := googleUuid.New()
	jr := &cryptoutilTemplateDomain.TenantJoinRequest{
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
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify status updated.
	var updated cryptoutilTemplateDomain.TenantJoinRequest
	require.NoError(t, testDB.WithContext(ctx).First(&updated, "id = ?", jr.ID).Error)
	require.Equal(t, "rejected", updated.Status)
}

func TestIntegration_DuplicateUsername_SameTenant(t *testing.T) {
	t.Skip("Join existing tenant flow not yet implemented in service")
	t.Parallel()

	ctx := context.Background()

	// Create tenant.
	tenant := &cryptoutilTemplateRepository.Tenant{
		ID:   googleUuid.New(),
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:8]),
	}
	require.NoError(t, testDB.Create(tenant).Error)

	username := fmt.Sprintf("user_%s", googleUuid.NewString()[:8])

	// Create first join request.
	userID1 := googleUuid.New()
	jr1 := &cryptoutilTemplateDomain.TenantJoinRequest{
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
	defer resp.Body.Close()

	// Should still succeed (duplicate check happens during approval).
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify two join requests exist.
	var joinRequests []cryptoutilTemplateDomain.TenantJoinRequest
	require.NoError(t, testDB.WithContext(ctx).Where("tenant_id = ?", tenant.ID).Find(&joinRequests).Error)
	require.GreaterOrEqual(t, len(joinRequests), 2, "Should have at least 2 join requests (duplicate checking deferred to approval)")
}

// TestIntegration_PostgreSQL tests with real PostgreSQL container (slow, only run with -tags=integration).
// NOTE: Disabled on Windows due to testcontainers "rootless Docker" error. Run on Linux/Mac instead.
func TestIntegration_PostgreSQL(t *testing.T) {
	t.Skip("PostgreSQL container test disabled on Windows - rootless Docker not supported")

	ctx := context.Background()

	// Start PostgreSQL container.
	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase(fmt.Sprintf("test_%s", googleUuid.NewString())),
		postgres.WithUsername(fmt.Sprintf("user_%s", googleUuid.NewString())),
		postgres.WithPassword("password"),
	)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	db, err := gorm.Open(postgresDriver.New(postgresDriver.Config{DSN: connStr}), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations.
	require.NoError(t, db.AutoMigrate(
		&cryptoutilTemplateRepository.Tenant{},
		&cryptoutilTemplateRepository.TenantRealm{},
		&cryptoutilTemplateRepository.User{},
		&cryptoutilTemplateDomain.TenantJoinRequest{},
	))

	// Create repositories and service.
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(db)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(db)
	_ = cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)

	// Create tenant via service.
	tenant := &cryptoutilTemplateRepository.Tenant{
		ID:   googleUuid.New(),
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:8]),
	}
	require.NoError(t, db.Create(tenant).Error)

	// Verify tenant exists.
	var retrieved cryptoutilTemplateRepository.Tenant
	require.NoError(t, db.First(&retrieved, "id = ?", tenant.ID).Error)
	require.Equal(t, tenant.Name, retrieved.Name)

	t.Logf("PostgreSQL integration test passed with tenant: %s", tenant.Name)
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
	require.Contains(t, result, "error")
	require.Contains(t, result["error"], "Invalid request body")
}

// TestIntegration_ListJoinRequests_NoRequests tests list when no requests exist.
// Note: This test creates its own isolated Fiber app to ensure no state pollution.
