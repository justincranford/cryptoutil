// Copyright 2025 Marlon Almeida. All rights reserved.

package repository

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUnverifiedUserRepository_GetByUsername(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	unverifiedUserRepo := NewUnverifiedUserRepository(db)
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

	username := "unverified_user_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength]
	email := username + "@example.com"

	unverifiedUser := &UnverifiedUser{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		Username:  username,
		Email:     email,
		ExpiresAt: time.Now().UTC().Add(72 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	err = unverifiedUserRepo.Create(ctx, unverifiedUser)
	require.NoError(t, err)

	result, err := unverifiedUserRepo.GetByUsername(ctx, username)
	require.NoError(t, err)
	require.Equal(t, unverifiedUser.ID, result.ID)
	require.Equal(t, unverifiedUser.Username, result.Username)
	require.Equal(t, unverifiedUser.Email, result.Email)
}

func TestUnverifiedUserRepository_GetByUsername_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	unverifiedUserRepo := NewUnverifiedUserRepository(db)
	ctx := context.Background()

	_, err := unverifiedUserRepo.GetByUsername(ctx, "nonexistent_user")
	require.Error(t, err)
}

func TestUnverifiedUserRepository_ListByTenant(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	unverifiedUserRepo := NewUnverifiedUserRepository(db)
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

	user1 := &UnverifiedUser{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		Username:  "user1_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Email:     "user1_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
		ExpiresAt: time.Now().UTC().Add(72 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	user2 := &UnverifiedUser{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		Username:  "user2_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Email:     "user2_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
		ExpiresAt: time.Now().UTC().Add(72 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	err = unverifiedUserRepo.Create(ctx, user1)
	require.NoError(t, err)

	err = unverifiedUserRepo.Create(ctx, user2)
	require.NoError(t, err)

	results, err := unverifiedUserRepo.ListByTenant(ctx, tenant.ID)
	require.NoError(t, err)
	require.Len(t, results, 2)
}

func TestUnverifiedUserRepository_Delete(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	unverifiedUserRepo := NewUnverifiedUserRepository(db)
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

	unverifiedUser := &UnverifiedUser{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		Username:  "user_to_delete_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Email:     "delete_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
		ExpiresAt: time.Now().UTC().Add(72 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	err = unverifiedUserRepo.Create(ctx, unverifiedUser)
	require.NoError(t, err)

	err = unverifiedUserRepo.Delete(ctx, unverifiedUser.ID)
	require.NoError(t, err)

	_, err = unverifiedUserRepo.GetByID(ctx, unverifiedUser.ID)
	require.Error(t, err)
}

func TestUnverifiedClientRepository_GetByClientID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	unverifiedClientRepo := NewUnverifiedClientRepository(db)
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

	clientID := "unverified_client_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength]

	unverifiedClient := &UnverifiedClient{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		ClientID:  clientID,
		ExpiresAt: time.Now().UTC().Add(72 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	err = unverifiedClientRepo.Create(ctx, unverifiedClient)
	require.NoError(t, err)

	result, err := unverifiedClientRepo.GetByClientID(ctx, clientID)
	require.NoError(t, err)
	require.Equal(t, unverifiedClient.ID, result.ID)
	require.Equal(t, unverifiedClient.ClientID, result.ClientID)
}

func TestUnverifiedClientRepository_GetByClientID_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	unverifiedClientRepo := NewUnverifiedClientRepository(db)
	ctx := context.Background()

	_, err := unverifiedClientRepo.GetByClientID(ctx, "nonexistent_client")
	require.Error(t, err)
}

func TestUnverifiedClientRepository_ListByTenant(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	unverifiedClientRepo := NewUnverifiedClientRepository(db)
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

	client1 := &UnverifiedClient{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		ClientID:  "client1_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		ExpiresAt: time.Now().UTC().Add(72 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	client2 := &UnverifiedClient{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		ClientID:  "client2_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		ExpiresAt: time.Now().UTC().Add(72 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	err = unverifiedClientRepo.Create(ctx, client1)
	require.NoError(t, err)

	err = unverifiedClientRepo.Create(ctx, client2)
	require.NoError(t, err)

	results, err := unverifiedClientRepo.ListByTenant(ctx, tenant.ID)
	require.NoError(t, err)
	require.Len(t, results, 2)
}

func TestUnverifiedClientRepository_Delete(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	unverifiedClientRepo := NewUnverifiedClientRepository(db)
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

	unverifiedClient := &UnverifiedClient{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		ClientID:  "client_to_delete_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		ExpiresAt: time.Now().UTC().Add(72 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	err = unverifiedClientRepo.Create(ctx, unverifiedClient)
	require.NoError(t, err)

	err = unverifiedClientRepo.Delete(ctx, unverifiedClient.ID)
	require.NoError(t, err)

	_, err = unverifiedClientRepo.GetByID(ctx, unverifiedClient.ID)
	require.Error(t, err)
}
