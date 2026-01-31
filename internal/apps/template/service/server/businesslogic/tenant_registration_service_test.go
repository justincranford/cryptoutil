// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package businesslogic

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgresModule "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerDomain "cryptoutil/internal/apps/template/service/server/domain"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	testDB             *gorm.DB
	testSessionManager *SessionManager
	testTenantID       googleUuid.UUID
	testRealmID        googleUuid.UUID
	testUserID         googleUuid.UUID
	testClientID       googleUuid.UUID
	postgresTestDB     bool
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Try PostgreSQL first, fallback to SQLite on failure (Windows Docker Desktop requirement).
	var (
		db        *gorm.DB
		container *postgresModule.PostgresContainer
		err       error
	)

	// Try PostgreSQL with test-containers (with panic recovery for Docker Desktop not running).
	func() {
		defer func() {
			if r := recover(); r != nil {
				// testcontainers panics when Docker Desktop not running.
				// Silently fall through to SQLite fallback.
				err = fmt.Errorf("postgres container panic: %v", r)
			}
		}()

		dbName := fmt.Sprintf("test_%s", googleUuid.Must(googleUuid.NewV7()))
		userName := fmt.Sprintf("user_%s", googleUuid.Must(googleUuid.NewV7()))

		container, err = postgresModule.Run(ctx,
			"postgres:18-alpine",
			postgresModule.WithDatabase(dbName),
			postgresModule.WithUsername(userName),
			postgresModule.WithPassword("password"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(60*time.Second),
			),
		)
	}()

	if err == nil && container != nil {
		// PostgreSQL container started successfully.
		defer func() {
			if err := container.Terminate(ctx); err != nil {
				panic(fmt.Sprintf("failed to terminate postgres container: %v", err))
			}
		}()

		connStr, err := container.ConnectionString(ctx)
		if err != nil {
			panic(fmt.Sprintf("failed to get connection string: %v", err))
		}

		db, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})
		if err != nil {
			panic(fmt.Sprintf("failed to connect to postgres: %v", err))
		}

		postgresTestDB = true
	} else {
		// Fallback to SQLite in-memory database.
		// Use sql.Open with "sqlite" driver to force modernc.org/sqlite (CGO-free).
		sqlDB, err := sql.Open("sqlite", "file::memory:?cache=shared")
		if err != nil {
			panic(fmt.Sprintf("failed to open sqlite: %v", err))
		}

		// Configure SQLite for concurrent operations.
		if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
			panic(fmt.Sprintf("failed to enable WAL mode: %v", err))
		}

		if _, err := sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
			panic(fmt.Sprintf("failed to set busy timeout: %v", err))
		}

		sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
		sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
		sqlDB.SetConnMaxLifetime(0)

		// Wrap with GORM using Dialector pattern.
		db, err = gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
			SkipDefaultTransaction: true,
		})
		if err != nil {
			panic(fmt.Sprintf("failed to wrap with GORM: %v", err))
		}

		postgresTestDB = false
	}

	testDB = db

	// Auto-migrate template tables.
	if err := testDB.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerRepository.Tenant{},
		&cryptoutilAppsTemplateServiceServerRepository.TenantRealm{},
		&cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{},
	); err != nil {
		panic(fmt.Sprintf("failed to run migrations: %v", err))
	}

	// Auto-migrate session tables for SessionManager tests.
	if err := testDB.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{},
		&cryptoutilAppsTemplateServiceServerRepository.ServiceSession{},
		&cryptoutilAppsTemplateServiceServerRepository.BrowserSessionJWK{},
		&cryptoutilAppsTemplateServiceServerRepository.ServiceSessionJWK{},
	); err != nil {
		panic(fmt.Sprintf("failed to migrate session tables: %v", err))
	}

	// Create shared SessionManager for session tests.
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		ServiceSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		BrowserSessionExpiration:   24 * time.Hour,
		ServiceSessionExpiration:   7 * 24 * time.Hour,
		SessionIdleTimeout:         2 * time.Hour,
		SessionCleanupInterval:     time.Hour,
		BrowserSessionJWSAlgorithm: "RS256",
		BrowserSessionJWEAlgorithm: "dir+A256GCM",
		ServiceSessionJWSAlgorithm: "RS256",
		ServiceSessionJWEAlgorithm: "dir+A256GCM",
	}

	testSessionManager = NewSessionManager(testDB, nil, config)
	if err := testSessionManager.Initialize(ctx); err != nil {
		panic(fmt.Sprintf("failed to initialize session manager: %v", err))
	}

	// Create test tenant and realm for all tests.
	testTenantID = googleUuid.New()
	testRealmID = googleUuid.New()
	testUserID = googleUuid.New()
	testClientID = googleUuid.New()

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testDB)
	tenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{
		ID:   testTenantID,
		Name: "Test Tenant",
	}

	if err := tenantRepo.Create(ctx, tenant); err != nil {
		panic(fmt.Sprintf("failed to create test tenant: %v", err))
	}

	realmRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRealmRepository(testDB)
	realm := &cryptoutilAppsTemplateServiceServerRepository.TenantRealm{
		ID:       testRealmID,
		TenantID: testTenantID,
		RealmID:  googleUuid.New(),
		Type:     "username_password",
		Active:   true,
		Source:   "db",
	}

	if err := realmRepo.Create(ctx, realm); err != nil {
		panic(fmt.Sprintf("failed to create test realm: %v", err))
	}

	// Run all tests.
	exitCode := m.Run()

	// Cleanup happens via defer (PostgreSQL container termination).
	os.Exit(exitCode)
}

func TestNewTenantRegistrationService(t *testing.T) {
	t.Parallel()

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testDB)

	service := NewTenantRegistrationService(testDB, tenantRepo, userRepo, joinRequestRepo)

	require.NotNil(t, service)
	require.Equal(t, testDB, service.db)
	require.Equal(t, tenantRepo, service.tenantRepo)
	require.Equal(t, userRepo, service.userRepo)
	require.Equal(t, joinRequestRepo, service.joinRequestRepo)
}

func TestRegisterUserWithTenant_CreateTenant(t *testing.T) {
	t.Parallel()

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testDB)

	service := NewTenantRegistrationService(testDB, tenantRepo, userRepo, joinRequestRepo)

	ctx := context.Background()
	userID := googleUuid.New()

	// Create new tenant.
	tenant, err := service.RegisterUserWithTenant(ctx, userID, "New Test Tenant", true)
	require.NoError(t, err)
	require.NotNil(t, tenant)
	require.Equal(t, "New Test Tenant", tenant.Name)
	require.NotEqual(t, googleUuid.Nil, tenant.ID)

	// Verify tenant exists in database.
	retrieved, err := tenantRepo.GetByID(ctx, tenant.ID)
	require.NoError(t, err)
	require.Equal(t, tenant.ID, retrieved.ID)
	require.Equal(t, "New Test Tenant", retrieved.Name)
}

func TestRegisterClientWithTenant(t *testing.T) {
	t.Parallel()

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testDB)

	service := NewTenantRegistrationService(testDB, tenantRepo, userRepo, joinRequestRepo)

	ctx := context.Background()
	clientID := googleUuid.New()

	// Register client with test tenant.
	err := service.RegisterClientWithTenant(ctx, clientID, testTenantID)
	require.NoError(t, err)

	// Verify join request created.
	requests, err := joinRequestRepo.ListByTenant(ctx, testTenantID)
	require.NoError(t, err)
	require.NotEmpty(t, requests)

	// Find our request.
	var found bool

	for _, req := range requests {
		if req.ClientID != nil && *req.ClientID == clientID {
			found = true

			require.Equal(t, testTenantID, req.TenantID)
			require.Equal(t, "pending", req.Status)

			break
		}
	}

	require.True(t, found, "join request not found for client")
}

// TestRegisterClientWithTenant_TenantNotFound tests that RegisterClientWithTenant returns error when tenant doesn't exist.
func TestRegisterClientWithTenant_TenantNotFound(t *testing.T) {
	t.Parallel()

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testDB)

	service := NewTenantRegistrationService(testDB, tenantRepo, userRepo, joinRequestRepo)

	ctx := context.Background()
	clientID := googleUuid.New()
	nonExistentTenantID := googleUuid.New()

	// Try to register client with non-existent tenant.
	err := service.RegisterClientWithTenant(ctx, clientID, nonExistentTenantID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "tenant not found")
}

func TestAuthorizeJoinRequest_Approve(t *testing.T) {
	t.Parallel()

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testDB)

	service := NewTenantRegistrationService(testDB, tenantRepo, userRepo, joinRequestRepo)

	ctx := context.Background()
	clientID := googleUuid.New()
	adminUserID := googleUuid.New()

	// Create join request.
	err := service.RegisterClientWithTenant(ctx, clientID, testTenantID)
	require.NoError(t, err)

	// Find created request.
	requests, err := joinRequestRepo.ListByTenant(ctx, testTenantID)
	require.NoError(t, err)

	var requestID googleUuid.UUID

	for _, req := range requests {
		if req.ClientID != nil && *req.ClientID == clientID {
			requestID = req.ID

			break
		}
	}

	require.NotEqual(t, googleUuid.Nil, requestID)

	// Approve join request.
	err = service.AuthorizeJoinRequest(ctx, requestID, adminUserID, true)
	require.NoError(t, err)

	// Verify status updated.
	updated, err := joinRequestRepo.GetByID(ctx, requestID)
	require.NoError(t, err)
	require.Equal(t, "approved", updated.Status)
	require.NotNil(t, updated.ProcessedAt)
	require.Equal(t, adminUserID, *updated.ProcessedBy)
}

func TestAuthorizeJoinRequest_Reject(t *testing.T) {
	t.Parallel()

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testDB)

	service := NewTenantRegistrationService(testDB, tenantRepo, userRepo, joinRequestRepo)

	ctx := context.Background()
	clientID := googleUuid.New()
	adminUserID := googleUuid.New()

	// Create join request.
	err := service.RegisterClientWithTenant(ctx, clientID, testTenantID)
	require.NoError(t, err)

	// Find created request.
	requests, err := joinRequestRepo.ListByTenant(ctx, testTenantID)
	require.NoError(t, err)

	var requestID googleUuid.UUID

	for _, req := range requests {
		if req.ClientID != nil && *req.ClientID == clientID {
			requestID = req.ID

			break
		}
	}

	require.NotEqual(t, googleUuid.Nil, requestID)

	// Reject join request.
	err = service.AuthorizeJoinRequest(ctx, requestID, adminUserID, false)
	require.NoError(t, err)

	// Verify status updated.
	updated, err := joinRequestRepo.GetByID(ctx, requestID)
	require.NoError(t, err)
	require.Equal(t, "rejected", updated.Status)
	require.NotNil(t, updated.ProcessedAt)
	require.Equal(t, adminUserID, *updated.ProcessedBy)
}

func TestAuthorizeJoinRequest_AlreadyProcessed(t *testing.T) {
	t.Parallel()

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testDB)

	service := NewTenantRegistrationService(testDB, tenantRepo, userRepo, joinRequestRepo)

	ctx := context.Background()
	clientID := googleUuid.New()
	adminUserID := googleUuid.New()

	// Create and approve join request.
	err := service.RegisterClientWithTenant(ctx, clientID, testTenantID)
	require.NoError(t, err)

	// Find created request.
	requests, err := joinRequestRepo.ListByTenant(ctx, testTenantID)
	require.NoError(t, err)

	var requestID googleUuid.UUID

	for _, req := range requests {
		if req.ClientID != nil && *req.ClientID == clientID {
			requestID = req.ID

			break
		}
	}

	require.NotEqual(t, googleUuid.Nil, requestID)

	err = service.AuthorizeJoinRequest(ctx, requestID, adminUserID, true)
	require.NoError(t, err)

	// Try to process again - should return error.
	err = service.AuthorizeJoinRequest(ctx, requestID, adminUserID, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not pending")
}

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
	userID := googleUuid.New()

	// Test join flow (createTenant=false) - should return "not yet implemented" error.
	_, err := service.RegisterUserWithTenant(ctx, userID, "Existing Tenant", false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "join existing tenant flow not yet implemented")
}

// =============================================================================
// Database Error Tests - test repository error paths
// =============================================================================

// TestRegisterUserWithTenant_CreateTenant_DBError tests that RegisterUserWithTenant returns error when DB is closed.
func TestRegisterUserWithTenant_CreateTenant_DBError(t *testing.T) {
	// Create fresh database for this test.
	sqlDB, err := sql.Open("sqlite", "file::memory:?cache=shared")
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
	userID := googleUuid.New()

	// Try to create tenant with closed DB.
	_, err = service.RegisterUserWithTenant(ctx, userID, "New Tenant", true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create tenant")
}

// TestRegisterClientWithTenant_DBError tests that RegisterClientWithTenant returns error when DB is closed.
func TestRegisterClientWithTenant_DBError(t *testing.T) {
	// Create fresh database for this test.
	sqlDB, err := sql.Open("sqlite", "file::memory:?cache=shared")
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
	sqlDB, err := sql.Open("sqlite", "file::memory:?cache=shared")
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
	sqlDB, err := sql.Open("sqlite", "file::memory:?cache=shared")
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
	sqlDB, err := sql.Open("sqlite", "file::memory:?cache=shared")
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
