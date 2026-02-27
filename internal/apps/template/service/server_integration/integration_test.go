// Package server_integration provides integration tests for service-template server.
// These tests validate end-to-end flows including tenant registration and join requests.
//
//go:build integration

package server_integration

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// TestIntegration_TenantRegistration_CreateTenant tests creating a new tenant.
func TestIntegration_TenantRegistration_CreateTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create in-memory SQLite database with migrations.
	gormDB, err := cryptoutilAppsTemplateServiceServerRepository.InitSQLite(ctx, cryptoutilSharedMagic.SQLiteInMemoryDSN, cryptoutilAppsTemplateServiceServerRepository.MigrationsFS)
	require.NoError(t, err, "Failed to initialize in-memory database")

	sqlDB, err := gormDB.DB()
	require.NoError(t, err, "Failed to get sql.DB")

	defer func() { _ = sqlDB.Close() }()

	// Create repositories.
	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(gormDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(gormDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(gormDB)

	// Create registration service.
	registrationSvc := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(
		gormDB,
		tenantRepo,
		userRepo,
		joinRequestRepo,
	)

	// Test: Create new tenant.
	userID := googleUuid.New()
	tenantName := "TestTenant"

	tenant, err := registrationSvc.RegisterUserWithTenant(ctx, userID, "test-user", "test@example.com", "hashed_password_123", tenantName, true)
	require.NoError(t, err, "Creating tenant should succeed")

	// Assertions
	require.NotNil(t, tenant, "Tenant should be returned")
	require.Equal(t, tenantName, tenant.Name, "Tenant name should match")
	require.NotEqual(t, googleUuid.Nil, tenant.ID, "Tenant ID should be set")

	// Verify tenant exists in database
	retrieved, err := tenantRepo.GetByID(ctx, tenant.ID)
	require.NoError(t, err, "Should retrieve created tenant")
	require.Equal(t, tenant.ID, retrieved.ID, "Retrieved tenant ID should match")
	require.Equal(t, tenantName, retrieved.Name, "Retrieved tenant name should match")
}

// TestIntegration_TenantRegistration_ClientJoinRequest tests client join request creation.
func TestIntegration_TenantRegistration_ClientJoinRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create in-memory SQLite database with migrations.
	gormDB, err := cryptoutilAppsTemplateServiceServerRepository.InitSQLite(ctx, cryptoutilSharedMagic.SQLiteInMemoryDSN, cryptoutilAppsTemplateServiceServerRepository.MigrationsFS)
	require.NoError(t, err, "Failed to initialize in-memory database")

	sqlDB, err := gormDB.DB()
	require.NoError(t, err, "Failed to get sql.DB")

	defer func() { _ = sqlDB.Close() }()

	// Create repositories.
	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(gormDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(gormDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(gormDB)

	// Create registration service.
	registrationSvc := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(
		gormDB,
		tenantRepo,
		userRepo,
		joinRequestRepo,
	)

	// Step 1: Create tenant first
	userID := googleUuid.New()
	tenant, err := registrationSvc.RegisterUserWithTenant(ctx, userID, "test-user", "test@example.com", "hashed_password_123", "TestTenant", true)
	require.NoError(t, err, "Creating tenant should succeed")

	// Step 2: Register client with join request
	clientID := googleUuid.New()
	err = registrationSvc.RegisterClientWithTenant(ctx, clientID, tenant.ID)
	require.NoError(t, err, "Creating client join request should succeed")

	// Step 3: Verify join request created
	joinRequests, err := joinRequestRepo.ListByTenant(ctx, tenant.ID)
	require.NoError(t, err, "Should list join requests")
	require.Len(t, joinRequests, 1, "Should have one join request")
	require.NotNil(t, joinRequests[0].ClientID, "Join request should have client ID")
	require.Equal(t, clientID, *joinRequests[0].ClientID, "Join request should be for the correct client")
	require.Equal(t, tenant.ID, joinRequests[0].TenantID, "Join request should be for the correct tenant")
	require.Equal(t, "pending", joinRequests[0].Status, "Join request should be pending")
}

// TestIntegration_JoinRequest_Approve tests approving a join request.
func TestIntegration_JoinRequest_Approve(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create in-memory SQLite database with migrations.
	gormDB, err := cryptoutilAppsTemplateServiceServerRepository.InitSQLite(ctx, cryptoutilSharedMagic.SQLiteInMemoryDSN, cryptoutilAppsTemplateServiceServerRepository.MigrationsFS)
	require.NoError(t, err, "Failed to initialize in-memory database")

	sqlDB, err := gormDB.DB()
	require.NoError(t, err, "Failed to get sql.DB")

	defer func() { _ = sqlDB.Close() }()

	// Create repositories.
	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(gormDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(gormDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(gormDB)

	// Create registration service.
	registrationSvc := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(
		gormDB,
		tenantRepo,
		userRepo,
		joinRequestRepo,
	)

	// Step 1: Create tenant
	userID := googleUuid.New()
	tenant, err := registrationSvc.RegisterUserWithTenant(ctx, userID, "test-user", "test@example.com", "hashed_password_123", "TestTenant", true)
	require.NoError(t, err, "Creating tenant should succeed")

	// Step 2: Create client join request
	clientID := googleUuid.New()
	err = registrationSvc.RegisterClientWithTenant(ctx, clientID, tenant.ID)
	require.NoError(t, err, "Creating client join request should succeed")

	// Step 3: Get join request
	joinRequests, err := joinRequestRepo.ListByTenant(ctx, tenant.ID)
	require.NoError(t, err, "Should list join requests")
	require.Len(t, joinRequests, 1, "Should have one join request")
	joinRequestID := joinRequests[0].ID

	// Step 4: Approve join request
	adminUserID := googleUuid.New()
	err = registrationSvc.AuthorizeJoinRequest(ctx, joinRequestID, adminUserID, true)
	require.NoError(t, err, "Approving join request should succeed")

	// Step 5: Verify join request status updated
	approved, err := joinRequestRepo.GetByID(ctx, joinRequestID)
	require.NoError(t, err, "Should retrieve join request")
	require.Equal(t, "approved", approved.Status, "Join request should be approved")
	require.NotNil(t, approved.ProcessedAt, "ProcessedAt should be set")
	require.NotNil(t, approved.ProcessedBy, "ProcessedBy should be set")
	require.Equal(t, adminUserID, *approved.ProcessedBy, "ProcessedBy should be admin user")
}

// TestIntegration_JoinRequest_Reject tests rejecting a join request.
func TestIntegration_JoinRequest_Reject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create in-memory SQLite database with migrations.
	gormDB, err := cryptoutilAppsTemplateServiceServerRepository.InitSQLite(ctx, cryptoutilSharedMagic.SQLiteInMemoryDSN, cryptoutilAppsTemplateServiceServerRepository.MigrationsFS)
	require.NoError(t, err, "Failed to initialize in-memory database")

	sqlDB, err := gormDB.DB()
	require.NoError(t, err, "Failed to get sql.DB")

	defer func() { _ = sqlDB.Close() }()

	// Create repositories.
	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(gormDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(gormDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(gormDB)

	// Create registration service.
	registrationSvc := cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(
		gormDB,
		tenantRepo,
		userRepo,
		joinRequestRepo,
	)

	// Step 1: Create tenant
	userID := googleUuid.New()
	tenant, err := registrationSvc.RegisterUserWithTenant(ctx, userID, "test-user", "test@example.com", "hashed_password_123", "TestTenant", true)
	require.NoError(t, err, "Creating tenant should succeed")

	// Step 2: Create client join request
	clientID := googleUuid.New()
	err = registrationSvc.RegisterClientWithTenant(ctx, clientID, tenant.ID)
	require.NoError(t, err, "Creating client join request should succeed")

	// Step 3: Get join request
	joinRequests, err := joinRequestRepo.ListByTenant(ctx, tenant.ID)
	require.NoError(t, err, "Should list join requests")
	require.Len(t, joinRequests, 1, "Should have one join request")
	joinRequestID := joinRequests[0].ID

	// Step 4: Reject join request
	adminUserID := googleUuid.New()
	err = registrationSvc.AuthorizeJoinRequest(ctx, joinRequestID, adminUserID, false)
	require.NoError(t, err, "Rejecting join request should succeed")

	// Step 5: Verify join request status updated
	rejected, err := joinRequestRepo.GetByID(ctx, joinRequestID)
	require.NoError(t, err, "Should retrieve join request")
	require.Equal(t, "rejected", rejected.Status, "Join request should be rejected")
	require.NotNil(t, rejected.ProcessedAt, "ProcessedAt should be set")
	require.NotNil(t, rejected.ProcessedBy, "ProcessedBy should be set")
	require.Equal(t, adminUserID, *rejected.ProcessedBy, "ProcessedBy should be admin user")
}
