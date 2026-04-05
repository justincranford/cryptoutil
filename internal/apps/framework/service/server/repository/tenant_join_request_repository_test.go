// Copyright (c) 2025 Justin Cranford
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cryptoutilAppsFrameworkServiceServerDomain "cryptoutil/internal/apps/framework/service/server/domain"
)

// setupJoinRequestTestDB creates an isolated in-memory SQLite database for join request tests.
// Each call creates a unique database to prevent test interference.
func setupJoinRequestTestDB(t *testing.T) *gorm.DB {
	t.Helper(
	// Use unique database name to avoid sharing between parallel tests.
	)

	dsn := fmt.Sprintf("file:memdb_%s?mode=memory&cache=shared", googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(context.Background(), "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(context.Background(), "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	dialector := sqlite.Dialector{Conn: sqlDB}
	db, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	sqlDB, err = db.DB()
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	sqlDB.SetConnMaxLifetime(0)

	// Run migrations for all required tables.
	err = db.AutoMigrate(&Tenant{}, &User{}, &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{})
	require.NoError(t, err)

	return db
}

func TestNewTenantJoinRequestRepository(t *testing.T) {
	t.Parallel()

	db := setupJoinRequestTestDB(t)
	repo := NewTenantJoinRequestRepository(db)

	require.NotNil(t, repo)
}

func TestTenantJoinRequestRepository_Create(t *testing.T) {
	t.Parallel()

	db := setupJoinRequestTestDB(t)
	repo := NewTenantJoinRequestRepository(db)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create tenant.
	tenant := &Tenant{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        "Test Tenant",
		Description: "Test tenant for join request tests",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Create user for join request.
	user := &User{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		TenantID:     tenant.ID,
		Username:     "joinrequestuser",
		PasswordHash: "hashedpassword",
		Active:       1,
		CreatedAt:    time.Now().UTC(),
	}
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	request := &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		UserID:      &user.ID,
		TenantID:    tenant.ID,
		Status:      cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusPending,
		RequestedAt: time.Now().UTC(),
	}

	err = repo.Create(ctx, request)
	require.NoError(t, err)

	// Verify request was created.
	retrieved, err := repo.GetByID(ctx, request.ID)
	require.NoError(t, err)
	require.Equal(t, request.ID, retrieved.ID)
	require.Equal(t, request.TenantID, retrieved.TenantID)
	require.Equal(t, request.Status, retrieved.Status)
}

func TestTenantJoinRequestRepository_Update(t *testing.T) {
	t.Parallel()

	db := setupJoinRequestTestDB(t)
	repo := NewTenantJoinRequestRepository(db)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create tenant.
	tenant := &Tenant{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        "Test Tenant",
		Description: "Test tenant for update tests",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Create user for join request.
	user := &User{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		TenantID:     tenant.ID,
		Username:     "updateuser",
		PasswordHash: "hashedpassword",
		Active:       1,
		CreatedAt:    time.Now().UTC(),
	}
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create initial request.
	request := &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		UserID:      &user.ID,
		TenantID:    tenant.ID,
		Status:      cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusPending,
		RequestedAt: time.Now().UTC(),
	}
	err = repo.Create(ctx, request)
	require.NoError(t, err)

	// Update request.
	request.Status = cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusApproved
	processedAt := time.Now().UTC()
	request.ProcessedAt = &processedAt
	request.ProcessedBy = &user.ID

	err = repo.Update(ctx, request)
	require.NoError(t, err)

	// Verify update.
	retrieved, err := repo.GetByID(ctx, request.ID)
	require.NoError(t, err)
	require.Equal(t, cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusApproved, retrieved.Status)
	require.NotNil(t, retrieved.ProcessedAt)
}

func TestTenantJoinRequestRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	db := setupJoinRequestTestDB(t)
	repo := NewTenantJoinRequestRepository(db)
	ctx := context.Background()

	// Try to get non-existent request.
	result, err := repo.GetByID(ctx, googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "join request not found")
	require.Nil(t, result)
}

func TestTenantJoinRequestRepository_ListByTenant(t *testing.T) {
	t.Parallel()

	db := setupJoinRequestTestDB(t)
	repo := NewTenantJoinRequestRepository(db)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create tenant.
	tenant := &Tenant{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        "Test Tenant",
		Description: "Test tenant for list tests",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Create user for join requests.
	user := &User{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		TenantID:     tenant.ID,
		Username:     "listbytenantuser",
		PasswordHash: "hashedpassword",
		Active:       1,
		CreatedAt:    time.Now().UTC(),
	}
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create multiple requests for the tenant.
	for i := 0; i < 3; i++ {
		request := &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
			ID:          googleUuid.Must(googleUuid.NewV7()),
			UserID:      &user.ID,
			TenantID:    tenant.ID,
			Status:      cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusPending,
			RequestedAt: time.Now().UTC().Add(time.Duration(i) * time.Second),
		}
		err = repo.Create(ctx, request)
		require.NoError(t, err)
	}

	// List requests by tenant.
	results, err := repo.ListByTenant(ctx, tenant.ID)
	require.NoError(t, err)
	require.Len(t, results, 3)
}

func TestTenantJoinRequestRepository_ListByStatus(t *testing.T) {
	t.Parallel()

	db := setupJoinRequestTestDB(t)
	repo := NewTenantJoinRequestRepository(db)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create tenant.
	tenant := &Tenant{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        "Test Tenant",
		Description: "Test tenant for status tests",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Create user for join requests.
	user := &User{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		TenantID:     tenant.ID,
		Username:     "listbystatususer",
		PasswordHash: "hashedpassword",
		Active:       1,
		CreatedAt:    time.Now().UTC(),
	}
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create requests with different statuses.
	statuses := []string{
		cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusPending,
		cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusApproved,
		cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusPending,
		cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusRejected,
	}
	for _, status := range statuses {
		request := &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
			ID:          googleUuid.Must(googleUuid.NewV7()),
			UserID:      &user.ID,
			TenantID:    tenant.ID,
			Status:      status,
			RequestedAt: time.Now().UTC(),
		}
		err = repo.Create(ctx, request)
		require.NoError(t, err)
	}

	// List pending requests.
	pendingResults, err := repo.ListByStatus(ctx, cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusPending)
	require.NoError(t, err)
	require.Len(t, pendingResults, 2)

	for _, r := range pendingResults {
		require.Equal(t, cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusPending, r.Status)
	}

	// List approved requests.
	approvedResults, err := repo.ListByStatus(ctx, cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusApproved)
	require.NoError(t, err)
	require.Len(t, approvedResults, 1)
}
