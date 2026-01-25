// Copyright (c) 2025 Justin Cranford
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceServerDomain "cryptoutil/internal/apps/template/service/server/domain"
)

// setupJoinRequestTestDB creates an isolated in-memory SQLite database for join request tests.
// Each call creates a unique database to prevent test interference.
func setupJoinRequestTestDB(t *testing.T) *gorm.DB {
	t.Helper(
	// Use unique database name to avoid sharing between parallel tests.
	)

	dsn := fmt.Sprintf("file:memdb_%s?mode=memory&cache=shared", googleUuid.NewString()[:8])

	sqlDB, err := sql.Open("sqlite", dsn)
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

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(0)

	// Run migrations for all required tables.
	err = db.AutoMigrate(&Tenant{}, &User{}, &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{})
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
		ID:		googleUuid.Must(googleUuid.NewV7()),
		Name:		"Test Tenant",
		Description:	"Test tenant for join request tests",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Create user for join request.
	user := &User{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		TenantID:	tenant.ID,
		Username:	"joinrequestuser",
		PasswordHash:	"hashedpassword",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	request := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		UserID:		&user.ID,
		TenantID:	tenant.ID,
		Status:		cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending,
		RequestedAt:	time.Now(),
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
		ID:		googleUuid.Must(googleUuid.NewV7()),
		Name:		"Test Tenant",
		Description:	"Test tenant for update tests",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Create user for join request.
	user := &User{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		TenantID:	tenant.ID,
		Username:	"updateuser",
		PasswordHash:	"hashedpassword",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create initial request.
	request := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		UserID:		&user.ID,
		TenantID:	tenant.ID,
		Status:		cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending,
		RequestedAt:	time.Now(),
	}
	err = repo.Create(ctx, request)
	require.NoError(t, err)

	// Update request.
	request.Status = cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusApproved
	processedAt := time.Now()
	request.ProcessedAt = &processedAt
	request.ProcessedBy = &user.ID

	err = repo.Update(ctx, request)
	require.NoError(t, err)

	// Verify update.
	retrieved, err := repo.GetByID(ctx, request.ID)
	require.NoError(t, err)
	require.Equal(t, cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusApproved, retrieved.Status)
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
		ID:		googleUuid.Must(googleUuid.NewV7()),
		Name:		"Test Tenant",
		Description:	"Test tenant for list tests",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Create user for join requests.
	user := &User{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		TenantID:	tenant.ID,
		Username:	"listbytenantuser",
		PasswordHash:	"hashedpassword",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create multiple requests for the tenant.
	for i := 0; i < 3; i++ {
		request := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
			ID:		googleUuid.Must(googleUuid.NewV7()),
			UserID:		&user.ID,
			TenantID:	tenant.ID,
			Status:		cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending,
			RequestedAt:	time.Now().Add(time.Duration(i) * time.Second),
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
		ID:		googleUuid.Must(googleUuid.NewV7()),
		Name:		"Test Tenant",
		Description:	"Test tenant for status tests",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Create user for join requests.
	user := &User{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		TenantID:	tenant.ID,
		Username:	"listbystatususer",
		PasswordHash:	"hashedpassword",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create requests with different statuses.
	statuses := []string{
		cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending,
		cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusApproved,
		cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending,
		cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusRejected,
	}
	for _, status := range statuses {
		request := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
			ID:		googleUuid.Must(googleUuid.NewV7()),
			UserID:		&user.ID,
			TenantID:	tenant.ID,
			Status:		status,
			RequestedAt:	time.Now(),
		}
		err = repo.Create(ctx, request)
		require.NoError(t, err)
	}

	// List pending requests.
	pendingResults, err := repo.ListByStatus(ctx, cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending)
	require.NoError(t, err)
	require.Len(t, pendingResults, 2)

	for _, r := range pendingResults {
		require.Equal(t, cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending, r.Status)
	}

	// List approved requests.
	approvedResults, err := repo.ListByStatus(ctx, cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusApproved)
	require.NoError(t, err)
	require.Len(t, approvedResults, 1)
}

func TestTenantJoinRequestRepository_ListByTenantAndStatus(t *testing.T) {
	t.Parallel()

	db := setupJoinRequestTestDB(t)
	repo := NewTenantJoinRequestRepository(db)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create first tenant.
	tenant1 := &Tenant{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		Name:		"Test Tenant 1",
		Description:	"First tenant for combined tests",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err := tenantRepo.Create(ctx, tenant1)
	require.NoError(t, err)

	// Create user for first tenant.
	user1 := &User{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		TenantID:	tenant1.ID,
		Username:	"tenant1user",
		PasswordHash:	"hashedpassword",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err = userRepo.Create(ctx, user1)
	require.NoError(t, err)

	// Create second tenant.
	tenant2 := &Tenant{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		Name:		"Test Tenant 2",
		Description:	"Second tenant for combined tests",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err = tenantRepo.Create(ctx, tenant2)
	require.NoError(t, err)

	// Create user for second tenant.
	user2 := &User{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		TenantID:	tenant2.ID,
		Username:	"tenant2user",
		PasswordHash:	"hashedpassword",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err = userRepo.Create(ctx, user2)
	require.NoError(t, err)

	// Create requests for tenant 1: 2 pending, 1 approved.
	for i := 0; i < 2; i++ {
		request := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
			ID:		googleUuid.Must(googleUuid.NewV7()),
			UserID:		&user1.ID,
			TenantID:	tenant1.ID,
			Status:		cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending,
			RequestedAt:	time.Now(),
		}
		err = repo.Create(ctx, request)
		require.NoError(t, err)
	}

	approvedRequest := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		UserID:		&user1.ID,
		TenantID:	tenant1.ID,
		Status:		cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusApproved,
		RequestedAt:	time.Now(),
	}
	err = repo.Create(ctx, approvedRequest)
	require.NoError(t, err)

	// Create request for tenant 2: 1 pending.
	tenant2Request := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		UserID:		&user2.ID,
		TenantID:	tenant2.ID,
		Status:		cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending,
		RequestedAt:	time.Now(),
	}
	err = repo.Create(ctx, tenant2Request)
	require.NoError(t, err)

	// List tenant 1 pending requests.
	results, err := repo.ListByTenantAndStatus(ctx, tenant1.ID, cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending)
	require.NoError(t, err)
	require.Len(t, results, 2)

	for _, r := range results {
		require.Equal(t, tenant1.ID, r.TenantID)
		require.Equal(t, cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending, r.Status)
	}

	// List tenant 1 approved requests.
	approvedResults, err := repo.ListByTenantAndStatus(ctx, tenant1.ID, cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusApproved)
	require.NoError(t, err)
	require.Len(t, approvedResults, 1)

	// List tenant 2 pending requests.
	tenant2Results, err := repo.ListByTenantAndStatus(ctx, tenant2.ID, cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending)
	require.NoError(t, err)
	require.Len(t, tenant2Results, 1)
}

func TestTenantJoinRequestRepository_CreateWithClientID(t *testing.T) {
	t.Parallel()

	db := setupJoinRequestTestDB(t)
	repo := NewTenantJoinRequestRepository(db)
	tenantRepo := NewTenantRepository(db)
	ctx := context.Background()

	// Create tenant.
	tenant := &Tenant{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		Name:		"Test Tenant",
		Description:	"Test tenant for client ID tests",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	clientID := googleUuid.Must(googleUuid.NewV7())
	request := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		ClientID:	&clientID,
		TenantID:	tenant.ID,
		Status:		cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending,
		RequestedAt:	time.Now(),
	}

	err = repo.Create(ctx, request)
	require.NoError(t, err)

	// Verify request was created with ClientID.
	retrieved, err := repo.GetByID(ctx, request.ID)
	require.NoError(t, err)
	require.Nil(t, retrieved.UserID)
	require.NotNil(t, retrieved.ClientID)
	require.Equal(t, clientID, *retrieved.ClientID)
}

func TestTenantJoinRequestRepository_ListByTenant_Empty(t *testing.T) {
	t.Parallel()

	db := setupJoinRequestTestDB(t)
	repo := NewTenantJoinRequestRepository(db)
	tenantRepo := NewTenantRepository(db)
	ctx := context.Background()

	// Create tenant.
	tenant := &Tenant{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		Name:		"Empty Tenant",
		Description:	"Tenant with no join requests",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// List requests from empty tenant.
	results, err := repo.ListByTenant(ctx, tenant.ID)
	require.NoError(t, err)
	require.Empty(t, results)
}

func TestTenantJoinRequestRepository_ListByStatus_Empty(t *testing.T) {
	t.Parallel()

	db := setupJoinRequestTestDB(t)
	repo := NewTenantJoinRequestRepository(db)
	ctx := context.Background()

	// List rejected status (none exist).
	results, err := repo.ListByStatus(ctx, cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusRejected)
	require.NoError(t, err)
	require.Empty(t, results)
}

func TestTenantJoinRequestRepository_ListByTenantAndStatus_Empty(t *testing.T) {
	t.Parallel()

	db := setupJoinRequestTestDB(t)
	repo := NewTenantJoinRequestRepository(db)
	tenantRepo := NewTenantRepository(db)
	ctx := context.Background()

	// Create tenant.
	tenant := &Tenant{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		Name:		"Empty Status Tenant",
		Description:	"Tenant with no matching status",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// List by tenant and status (none exist).
	results, err := repo.ListByTenantAndStatus(ctx, tenant.ID, cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusApproved)
	require.NoError(t, err)
	require.Empty(t, results)
}

func TestTenantJoinRequestRepository_Create_DuplicateID(t *testing.T) {
	t.Parallel()

	db := setupJoinRequestTestDB(t)
	repo := NewTenantJoinRequestRepository(db)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create tenant.
	tenant := &Tenant{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		Name:		"Test Tenant",
		Description:	"Test tenant for duplicate tests",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Create user.
	user := &User{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		TenantID:	tenant.ID,
		Username:	"duplicateuser",
		PasswordHash:	"hashedpassword",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	requestID := googleUuid.Must(googleUuid.NewV7())
	request1 := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
		ID:		requestID,
		UserID:		&user.ID,
		TenantID:	tenant.ID,
		Status:		cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending,
		RequestedAt:	time.Now(),
	}

	err = repo.Create(ctx, request1)
	require.NoError(t, err)

	// Try to create another request with same ID.
	request2 := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
		ID:		requestID,
		UserID:		&user.ID,
		TenantID:	tenant.ID,
		Status:		cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending,
		RequestedAt:	time.Now(),
	}

	err = repo.Create(ctx, request2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create join request")
}

func TestTenantJoinRequestRepository_Update_NonExistent(t *testing.T) {
	t.Parallel()

	db := setupJoinRequestTestDB(t)
	repo := NewTenantJoinRequestRepository(db)
	tenantRepo := NewTenantRepository(db)
	ctx := context.Background()

	// Create tenant.
	tenant := &Tenant{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		Name:		"Test Tenant",
		Description:	"Test tenant for update tests",
		Active:		1,
		CreatedAt:	time.Now(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Try to update non-existent request - GORM Save creates if not exists.
	userID := googleUuid.Must(googleUuid.NewV7())
	request := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
		ID:		googleUuid.Must(googleUuid.NewV7()),
		UserID:		&userID,
		TenantID:	tenant.ID,
		Status:		cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusApproved,
		RequestedAt:	time.Now(),
	}

	// GORM Save does upsert, so this will succeed.
	err = repo.Update(ctx, request)
	require.NoError(t, err)

	// Verify it was created.
	retrieved, err := repo.GetByID(ctx, request.ID)
	require.NoError(t, err)
	require.Equal(t, request.ID, retrieved.ID)
}
