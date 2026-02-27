// Copyright 2025 Marlon Almeida. All rights reserved.

package repository

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_GetByID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueUserTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	user := &User{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		Username: "testuser_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Email:    "testuser_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
		Active:   1,
	}

	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	result, err := userRepo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, user.ID, result.ID)
	require.Equal(t, user.Username, result.Username)
	require.Equal(t, user.Email, result.Email)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	_, err := userRepo.GetByID(ctx, googleUuid.New())
	require.Error(t, err)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueUserTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	email := "unique_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com"

	user := &User{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		Username: "testuser_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Email:    email,
		Active:   1,
	}

	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	result, err := userRepo.GetByEmail(ctx, email)
	require.NoError(t, err)
	require.Equal(t, user.ID, result.ID)
	require.Equal(t, user.Email, result.Email)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	_, err := userRepo.GetByEmail(ctx, "nonexistent@example.com")
	require.Error(t, err)
}

func TestUserRepository_Update(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueUserTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	user := &User{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		Username: "testuser_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Email:    "original_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
		Active:   1,
	}

	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	newEmail := "updated_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com"
	user.Email = newEmail

	err = userRepo.Update(ctx, user)
	require.NoError(t, err)

	result, err := userRepo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, newEmail, result.Email)
}

func TestUserRepository_Delete(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueUserTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	user := &User{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		Username: "testuser_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Email:    "todelete_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
		Active:   1,
	}

	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	err = userRepo.Delete(ctx, user.ID)
	require.NoError(t, err)

	_, err = userRepo.GetByID(ctx, user.ID)
	require.Error(t, err)
}

func TestClientRepository_GetByID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	clientRepo := NewClientRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueUserTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	client := &Client{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		ClientID: "client_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Active:   1,
	}

	err = clientRepo.Create(ctx, client)
	require.NoError(t, err)

	result, err := clientRepo.GetByID(ctx, client.ID)
	require.NoError(t, err)
	require.Equal(t, client.ID, result.ID)
	require.Equal(t, client.ClientID, result.ClientID)
}

func TestClientRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	clientRepo := NewClientRepository(db)
	ctx := context.Background()

	_, err := clientRepo.GetByID(ctx, googleUuid.New())
	require.Error(t, err)
}

func TestClientRepository_GetByClientID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	clientRepo := NewClientRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueUserTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	clientID := "unique_client_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength]

	client := &Client{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		ClientID: clientID,
		Active:   1,
	}

	err = clientRepo.Create(ctx, client)
	require.NoError(t, err)

	result, err := clientRepo.GetByClientID(ctx, clientID)
	require.NoError(t, err)
	require.Equal(t, client.ID, result.ID)
	require.Equal(t, clientID, result.ClientID)
}

func TestClientRepository_GetByClientID_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	clientRepo := NewClientRepository(db)
	ctx := context.Background()

	_, err := clientRepo.GetByClientID(ctx, "nonexistent_client")
	require.Error(t, err)
}

func TestClientRepository_ListByTenant(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	clientRepo := NewClientRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueUserTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	client1 := &Client{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		ClientID: "client1_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Active:   1,
	}

	client2 := &Client{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		ClientID: "client2_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Active:   0,
	}

	err = clientRepo.Create(ctx, client1)
	require.NoError(t, err)

	err = clientRepo.Create(ctx, client2)
	require.NoError(t, err)

	allResults, err := clientRepo.ListByTenant(ctx, tenant.ID, false)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(allResults), 2)

	activeResults, err := clientRepo.ListByTenant(ctx, tenant.ID, true)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(activeResults), 1)
}

func TestClientRepository_Update(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	clientRepo := NewClientRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueUserTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	client := &Client{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		ClientID: "original_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Name:     "Original Name",
		Active:   1,
	}

	err = clientRepo.Create(ctx, client)
	require.NoError(t, err)

	client.Name = "Updated Name"

	err = clientRepo.Update(ctx, client)
	require.NoError(t, err)

	result, err := clientRepo.GetByID(ctx, client.ID)
	require.NoError(t, err)
	require.Equal(t, "Updated Name", result.Name)
}

func TestClientRepository_Delete(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	clientRepo := NewClientRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueUserTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	client := &Client{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		ClientID: "todelete_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Active:   1,
	}

	err = clientRepo.Create(ctx, client)
	require.NoError(t, err)

	err = clientRepo.Delete(ctx, client.ID)
	require.NoError(t, err)

	_, err = clientRepo.GetByID(ctx, client.ID)
	require.Error(t, err)
}
