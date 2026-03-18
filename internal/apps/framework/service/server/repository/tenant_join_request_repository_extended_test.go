// Copyright (c) 2025 Justin Cranford
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServerDomain "cryptoutil/internal/apps/framework/service/server/domain"
)

func TestTenantJoinRequestRepository_ListByTenantAndStatus(t *testing.T) {
	t.Parallel()

	db := setupJoinRequestTestDB(t)
	repo := NewTenantJoinRequestRepository(db)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create first tenant.
	tenant1 := &Tenant{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        "Test Tenant 1",
		Description: "First tenant for combined tests",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}
	err := tenantRepo.Create(ctx, tenant1)
	require.NoError(t, err)

	// Create user for first tenant.
	user1 := &User{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		TenantID:     tenant1.ID,
		Username:     "tenant1user",
		PasswordHash: "hashedpassword",
		Active:       1,
		CreatedAt:    time.Now().UTC(),
	}
	err = userRepo.Create(ctx, user1)
	require.NoError(t, err)

	// Create second tenant.
	tenant2 := &Tenant{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        "Test Tenant 2",
		Description: "Second tenant for combined tests",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}
	err = tenantRepo.Create(ctx, tenant2)
	require.NoError(t, err)

	// Create user for second tenant.
	user2 := &User{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		TenantID:     tenant2.ID,
		Username:     "tenant2user",
		PasswordHash: "hashedpassword",
		Active:       1,
		CreatedAt:    time.Now().UTC(),
	}
	err = userRepo.Create(ctx, user2)
	require.NoError(t, err)

	// Create requests for tenant 1: 2 pending, 1 approved.
	for i := 0; i < 2; i++ {
		request := &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
			ID:          googleUuid.Must(googleUuid.NewV7()),
			UserID:      &user1.ID,
			TenantID:    tenant1.ID,
			Status:      cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusPending,
			RequestedAt: time.Now().UTC(),
		}
		err = repo.Create(ctx, request)
		require.NoError(t, err)
	}

	approvedRequest := &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		UserID:      &user1.ID,
		TenantID:    tenant1.ID,
		Status:      cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusApproved,
		RequestedAt: time.Now().UTC(),
	}
	err = repo.Create(ctx, approvedRequest)
	require.NoError(t, err)

	// Create request for tenant 2: 1 pending.
	tenant2Request := &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		UserID:      &user2.ID,
		TenantID:    tenant2.ID,
		Status:      cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusPending,
		RequestedAt: time.Now().UTC(),
	}
	err = repo.Create(ctx, tenant2Request)
	require.NoError(t, err)

	// List tenant 1 pending requests.
	results, err := repo.ListByTenantAndStatus(ctx, tenant1.ID, cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusPending)
	require.NoError(t, err)
	require.Len(t, results, 2)

	for _, r := range results {
		require.Equal(t, tenant1.ID, r.TenantID)
		require.Equal(t, cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusPending, r.Status)
	}

	// List tenant 1 approved requests.
	approvedResults, err := repo.ListByTenantAndStatus(ctx, tenant1.ID, cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusApproved)
	require.NoError(t, err)
	require.Len(t, approvedResults, 1)

	// List tenant 2 pending requests.
	tenant2Results, err := repo.ListByTenantAndStatus(ctx, tenant2.ID, cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusPending)
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
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        "Test Tenant",
		Description: "Test tenant for client ID tests",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	clientID := googleUuid.Must(googleUuid.NewV7())
	request := &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		ClientID:    &clientID,
		TenantID:    tenant.ID,
		Status:      cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusPending,
		RequestedAt: time.Now().UTC(),
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
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        "Empty Tenant",
		Description: "Tenant with no join requests",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
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
	results, err := repo.ListByStatus(ctx, cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusRejected)
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
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        "Empty Status Tenant",
		Description: "Tenant with no matching status",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// List by tenant and status (none exist).
	results, err := repo.ListByTenantAndStatus(ctx, tenant.ID, cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusApproved)
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
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        "Test Tenant",
		Description: "Test tenant for duplicate tests",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Create user.
	user := &User{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		TenantID:     tenant.ID,
		Username:     "duplicateuser",
		PasswordHash: "hashedpassword",
		Active:       1,
		CreatedAt:    time.Now().UTC(),
	}
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	requestID := googleUuid.Must(googleUuid.NewV7())
	request1 := &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
		ID:          requestID,
		UserID:      &user.ID,
		TenantID:    tenant.ID,
		Status:      cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusPending,
		RequestedAt: time.Now().UTC(),
	}

	err = repo.Create(ctx, request1)
	require.NoError(t, err)

	// Try to create another request with same ID.
	request2 := &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
		ID:          requestID,
		UserID:      &user.ID,
		TenantID:    tenant.ID,
		Status:      cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusPending,
		RequestedAt: time.Now().UTC(),
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
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        "Test Tenant",
		Description: "Test tenant for update tests",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Try to update non-existent request - GORM Save creates if not exists.
	userID := googleUuid.Must(googleUuid.NewV7())
	request := &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		UserID:      &userID,
		TenantID:    tenant.ID,
		Status:      cryptoutilAppsFrameworkServiceServerDomain.JoinRequestStatusApproved,
		RequestedAt: time.Now().UTC(),
	}

	// GORM Save does upsert, so this will succeed.
	err = repo.Update(ctx, request)
	require.NoError(t, err)

	// Verify it was created.
	retrieved, err := repo.GetByID(ctx, request.ID)
	require.NoError(t, err)
	require.Equal(t, request.ID, retrieved.ID)
}
