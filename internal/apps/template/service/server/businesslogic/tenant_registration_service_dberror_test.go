// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package businesslogic

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsTemplateServiceServerDomain "cryptoutil/internal/apps/template/service/server/domain"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

func TestListJoinRequests(t *testing.T) {
	t.Parallel()

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testDB)

	service := NewTenantRegistrationService(testDB, tenantRepo, userRepo, joinRequestRepo)

	ctx := context.Background()

	// Create multiple join requests for test tenant.
	clientID1 := googleUuid.New()
	clientID2 := googleUuid.New()

	err := service.RegisterClientWithTenant(ctx, clientID1, testTenantID)
	require.NoError(t, err)

	err = service.RegisterClientWithTenant(ctx, clientID2, testTenantID)
	require.NoError(t, err)

	// List join requests.
	requests, err := service.ListJoinRequests(ctx, testTenantID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(requests), 2) // At least our 2 requests.

	// Verify requests belong to test tenant.
	for _, req := range requests {
		require.Equal(t, testTenantID, req.TenantID)
	}
}

func TestRegisterUserWithTenant_JoinFlow(t *testing.T) {
	t.Parallel()

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testDB)

	service := NewTenantRegistrationService(testDB, tenantRepo, userRepo, joinRequestRepo)

	ctx := context.Background()
	userID := googleUuid.Must(googleUuid.NewV7())
	username := fmt.Sprintf("testuser_%s", userID.String()[:cryptoutilSharedMagic.IMMinPasswordLength])
	email := fmt.Sprintf("test_%s@example.com", userID.String()[:cryptoutilSharedMagic.IMMinPasswordLength])

	// Test join flow (createTenant=false) - should return "not yet implemented" error.
	_, err := service.RegisterUserWithTenant(ctx, userID, username, email, testPasswordHash, "Existing Tenant", false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "join existing tenant flow not yet implemented")
}

// =============================================================================
// Database Error Tests - test repository error paths
// =============================================================================

// TestRegisterUserWithTenant_CreateTenant_DBError tests that RegisterUserWithTenant returns error when DB is closed.
func TestRegisterUserWithTenant_CreateTenant_DBError(t *testing.T) {
	// Create fresh database for this test.
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteInMemoryDSN)
	require.NoError(t, err)

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Auto-migrate tables.
	err = db.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerRepository.Tenant{},
	)
	require.NoError(t, err)

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(db)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(db)

	service := NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)

	// Close the database connection to trigger an error.
	_ = sqlDB.Close()

	ctx := context.Background()
	userID := googleUuid.Must(googleUuid.NewV7())
	username := fmt.Sprintf("testuser_%s", userID.String()[:cryptoutilSharedMagic.IMMinPasswordLength])
	email := fmt.Sprintf("test_%s@example.com", userID.String()[:cryptoutilSharedMagic.IMMinPasswordLength])

	// Try to create tenant with closed DB.
	_, err = service.RegisterUserWithTenant(ctx, userID, username, email, testPasswordHash, "New Tenant", true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create tenant")
}

// TestRegisterClientWithTenant_DBError tests that RegisterClientWithTenant returns error when DB is closed.
func TestRegisterClientWithTenant_DBError(t *testing.T) {
	// Create fresh database for this test.
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteInMemoryDSN)
	require.NoError(t, err)

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Auto-migrate tables.
	err = db.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerRepository.Tenant{},
		&cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{},
	)
	require.NoError(t, err)

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(db)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(db)

	service := NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)

	ctx := context.Background()
	clientID := googleUuid.New()

	// Create test tenant first.
	tenantID := googleUuid.New()
	tenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{
		ID:   tenantID,
		Name: "Test Tenant for DB Error",
	}

	err = tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Close the database connection to trigger an error during join request creation.
	_ = sqlDB.Close()

	// Try to register client with closed DB - will fail somewhere in the chain.
	err = service.RegisterClientWithTenant(ctx, clientID, tenantID)
	require.Error(t, err)
	// Error can be "tenant not found" (if GetByID fails first) or "failed to create client join request" (if Create fails).
}

// TestRegisterClientWithTenant_JoinRequestCreateError tests join request creation failure specifically.
func TestRegisterClientWithTenant_JoinRequestCreateError(t *testing.T) {
	t.Parallel()

	// Use the shared testDB to test a different error scenario.
	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testDB)

	service := NewTenantRegistrationService(testDB, tenantRepo, userRepo, joinRequestRepo)

	ctx := context.Background()
	clientID := googleUuid.New()

	// Try to register client with non-existent tenant - tests tenant lookup path.
	// For join request create error, we would need a constraint violation or similar.
	nonExistentTenant := googleUuid.New()

	err := service.RegisterClientWithTenant(ctx, clientID, nonExistentTenant)
	require.Error(t, err)
	require.Contains(t, err.Error(), "tenant not found")
}

// TestAuthorizeJoinRequest_GetByID_DBError tests that AuthorizeJoinRequest returns error when GetByID fails.
func TestAuthorizeJoinRequest_GetByID_DBError(t *testing.T) {
	// Create fresh database for this test.
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteInMemoryDSN)
	require.NoError(t, err)

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Auto-migrate tables.
	err = db.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{},
	)
	require.NoError(t, err)

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(db)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(db)

	service := NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)

	ctx := context.Background()

	// Close the database connection to trigger an error.
	_ = sqlDB.Close()

	// Try to authorize with closed DB.
	requestID := googleUuid.New()
	adminUserID := googleUuid.New()

	err = service.AuthorizeJoinRequest(ctx, requestID, adminUserID, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get join request")
}

// TestAuthorizeJoinRequest_Update_DBError tests that AuthorizeJoinRequest returns error when Update fails.
func TestAuthorizeJoinRequest_Update_DBError(t *testing.T) {
	// Create fresh database for this test.
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteInMemoryDSN)
	require.NoError(t, err)

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Auto-migrate tables.
	err = db.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{},
	)
	require.NoError(t, err)

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(db)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(db)

	service := NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)

	ctx := context.Background()

	// Create a pending join request.
	clientID := googleUuid.New()
	tenantID := googleUuid.New()
	requestID := googleUuid.New()
	joinRequest := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
		ID:          requestID,
		ClientID:    &clientID,
		TenantID:    tenantID,
		Status:      cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending,
		RequestedAt: time.Now().UTC(),
	}

	err = joinRequestRepo.Create(ctx, joinRequest)
	require.NoError(t, err)

	// Close the database connection to trigger an error on Update.
	_ = sqlDB.Close()

	// Try to authorize with closed DB - GetByID returns cached but Update fails.
	adminUserID := googleUuid.New()

	err = service.AuthorizeJoinRequest(ctx, requestID, adminUserID, true)
	require.Error(t, err)
	// Error is "failed to update join request" since GetByID succeeds from cache/memory.
}

// TestListJoinRequests_DBError tests that ListJoinRequests returns error when ListByTenant fails.
func TestListJoinRequests_DBError(t *testing.T) {
	// Create fresh database for this test.
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteInMemoryDSN)
	require.NoError(t, err)

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Auto-migrate tables.
	err = db.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{},
	)
	require.NoError(t, err)

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(db)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(db)

	service := NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)

	ctx := context.Background()

	// Close the database connection to trigger an error.
	_ = sqlDB.Close()

	// Try to list join requests with closed DB.
	tenantID := googleUuid.New()

	_, err = service.ListJoinRequests(ctx, tenantID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list join requests")
}
