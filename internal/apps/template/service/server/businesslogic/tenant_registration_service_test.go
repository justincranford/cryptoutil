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

	cryptoutilTemplateDomain "cryptoutil/internal/apps/template/service/server/domain"
	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

var (
	testDB         *gorm.DB
	testTenantID   googleUuid.UUID
	testRealmID    googleUuid.UUID
	testUserID     googleUuid.UUID
	testClientID   googleUuid.UUID
	postgresTestDB bool
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

		sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)
		sqlDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)
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
		&cryptoutilTemplateRepository.Tenant{},
		&cryptoutilTemplateRepository.TenantRealm{},
		&cryptoutilTemplateDomain.TenantJoinRequest{},
	); err != nil {
		panic(fmt.Sprintf("failed to run migrations: %v", err))
	}

	// Create test tenant and realm for all tests.
	testTenantID = googleUuid.New()
	testRealmID = googleUuid.New()
	testUserID = googleUuid.New()
	testClientID = googleUuid.New()

	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(testDB)
	tenant := &cryptoutilTemplateRepository.Tenant{
		ID:   testTenantID,
		Name: "Test Tenant",
	}

	if err := tenantRepo.Create(ctx, tenant); err != nil {
		panic(fmt.Sprintf("failed to create test tenant: %v", err))
	}

	realmRepo := cryptoutilTemplateRepository.NewTenantRealmRepository(testDB)
	realm := &cryptoutilTemplateRepository.TenantRealm{
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

	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(testDB)

	service := NewTenantRegistrationService(testDB, tenantRepo, userRepo, joinRequestRepo)

	require.NotNil(t, service)
	require.Equal(t, testDB, service.db)
	require.Equal(t, tenantRepo, service.tenantRepo)
	require.Equal(t, userRepo, service.userRepo)
	require.Equal(t, joinRequestRepo, service.joinRequestRepo)
}

func TestRegisterUserWithTenant_CreateTenant(t *testing.T) {
	t.Parallel()

	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(testDB)

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

	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(testDB)

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

	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(testDB)

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

	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(testDB)

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

	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(testDB)

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

	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(testDB)

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

	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(testDB)

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

	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(testDB)

	service := NewTenantRegistrationService(testDB, tenantRepo, userRepo, joinRequestRepo)

	ctx := context.Background()
	userID := googleUuid.New()

	// Test join flow (createTenant=false) - should return "not yet implemented" error.
	_, err := service.RegisterUserWithTenant(ctx, userID, "Existing Tenant", false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "join existing tenant flow not yet implemented")
}
