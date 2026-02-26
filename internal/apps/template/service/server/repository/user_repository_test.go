// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// uniqueTenantName returns a unique tenant name for tests.
func uniqueUserTenantName(base string) string {
	return base + " " + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength]
}

func TestUserRepository_Create(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueUserTenantName("UserCreate"),
		Description: "Test tenant",
		Active: 1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Generate a unique username for the duplicate test.
	dupUsername := "testuser-" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength]

	tests := []struct {
		name      string
		user      *User
		wantError bool
	}{
		{
			name: "happy path - valid user",
			user: &User{
				ID:        googleUuid.New(),
				TenantID:  tenant.ID,
				Username:  "testuser-" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
				Email:     "test-" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
				Active: 1,
				CreatedAt: time.Now().UTC(),
			},
			wantError: false,
		},
		{
			name: "duplicate username",
			user: &User{
				ID:        googleUuid.New(),
				TenantID:  tenant.ID,
				Username:  dupUsername,
				Email:     "different@example.com",
				Active: 1,
				CreatedAt: time.Now().UTC(),
			},
			wantError: true,
		},
	}

	// Create first user with dupUsername for duplicate test.
	firstUser := &User{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		Username:  dupUsername,
		Email:     "first@example.com",
		Active: 1,
		CreatedAt: time.Now().UTC(),
	}
	err = userRepo.Create(ctx, firstUser)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := userRepo.Create(ctx, tt.user)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueUserTenantName("Test"),
		Description: "Test tenant",
		Active: 1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	user := &User{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		Username:  "testuser-" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Email:     "test-" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
		Active: 1,
		CreatedAt: time.Now().UTC(),
	}

	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name      string
		username  string
		wantError bool
	}{
		{
			name:      "happy path - existing user",
			username:  user.Username,
			wantError: false,
		},
		{
			name:      "user not found",
			username:  "nonexistent-" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := userRepo.GetByUsername(ctx, tt.username)

			if tt.wantError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, user.ID, result.ID)
				require.Equal(t, user.Username, result.Username)
			}
		})
	}
}

func TestUserRepository_ListByTenant(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	tenant1 := &Tenant{
		ID:          googleUuid.New(),
		Name:        "Tenant 1",
		Description: "First tenant",
		Active: 1,
		CreatedAt:   time.Now().UTC(),
	}

	tenant2 := &Tenant{
		ID:          googleUuid.New(),
		Name:        "Tenant 2",
		Description: "Second tenant",
		Active: 1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant1)
	require.NoError(t, err)

	err = tenantRepo.Create(ctx, tenant2)
	require.NoError(t, err)

	user1 := &User{
		ID:        googleUuid.New(),
		TenantID:  tenant1.ID,
		Username:  "user1",
		Email:     "user1@example.com",
		Active: 1,
		CreatedAt: time.Now().UTC(),
	}

	user2 := &User{
		ID:        googleUuid.New(),
		TenantID:  tenant1.ID,
		Username:  "user2",
		Email:     "user2@example.com",
		Active: 1,
		CreatedAt: time.Now().UTC(),
	}

	user3 := &User{
		ID:        googleUuid.New(),
		TenantID:  tenant2.ID,
		Username:  "user3",
		Email:     "user3@example.com",
		Active: 1,
		CreatedAt: time.Now().UTC(),
	}

	err = userRepo.Create(ctx, user1)
	require.NoError(t, err)

	err = userRepo.Create(ctx, user2)
	require.NoError(t, err)

	err = userRepo.Create(ctx, user3)
	require.NoError(t, err)

	result, err := userRepo.ListByTenant(ctx, tenant1.ID, true)
	require.NoError(t, err)
	require.Len(t, result, 2)

	for _, user := range result {
		require.Equal(t, tenant1.ID, user.TenantID)
	}
}

func TestClientRepository_Create(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	clientRepo := NewClientRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueUserTenantName("Test"),
		Description: "Test tenant",
		Active: 1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	tests := []struct {
		name      string
		client    *Client
		wantError bool
	}{
		{
			name: "happy path - valid client",
			client: &Client{
				ID:        googleUuid.New(),
				TenantID:  tenant.ID,
				ClientID:  "client123",
				Active: 1,
				CreatedAt: time.Now().UTC(),
			},
			wantError: false,
		},
		{
			name: "duplicate client ID",
			client: &Client{
				ID:        googleUuid.New(),
				TenantID:  tenant.ID,
				ClientID:  "client123",
				Active: 1,
				CreatedAt: time.Now().UTC(),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := clientRepo.Create(ctx, tt.client)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUnverifiedUserRepository_Create(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	unverifiedUserRepo := NewUnverifiedUserRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueUserTenantName("Test"),
		Description: "Test tenant",
		Active: 1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	unverifiedUser := &UnverifiedUser{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		Username:  "unverified1",
		Email:     "unverified1@example.com",
		ExpiresAt: time.Now().UTC().Add(72 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	err = unverifiedUserRepo.Create(ctx, unverifiedUser)
	require.NoError(t, err)

	result, err := unverifiedUserRepo.GetByID(ctx, unverifiedUser.ID)
	require.NoError(t, err)
	require.Equal(t, unverifiedUser.Username, result.Username)
}

func TestUnverifiedUserRepository_DeleteExpired(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	unverifiedUserRepo := NewUnverifiedUserRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueUserTenantName("Test"),
		Description: "Test tenant",
		Active: 1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	expiredUser := &UnverifiedUser{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		Username:  "expired_user",
		Email:     "expired@example.com",
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
		CreatedAt: time.Now().UTC().Add(-73 * time.Hour),
	}

	validUser := &UnverifiedUser{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		Username:  "valid_user",
		Email:     "valid@example.com",
		ExpiresAt: time.Now().UTC().Add(72 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	err = unverifiedUserRepo.Create(ctx, expiredUser)
	require.NoError(t, err)

	err = unverifiedUserRepo.Create(ctx, validUser)
	require.NoError(t, err)

	deletedCount, err := unverifiedUserRepo.DeleteExpired(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), deletedCount)

	_, err = unverifiedUserRepo.GetByID(ctx, expiredUser.ID)
	require.Error(t, err)

	result, err := unverifiedUserRepo.GetByID(ctx, validUser.ID)
	require.NoError(t, err)
	require.Equal(t, validUser.Username, result.Username)
}

func TestUnverifiedClientRepository_DeleteExpired(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	unverifiedClientRepo := NewUnverifiedClientRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueUserTenantName("Test"),
		Description: "Test tenant",
		Active: 1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	expiredClient := &UnverifiedClient{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		ClientID:  "expired_client",
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
		CreatedAt: time.Now().UTC().Add(-73 * time.Hour),
	}

	validClient := &UnverifiedClient{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		ClientID:  "valid_client",
		ExpiresAt: time.Now().UTC().Add(72 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	err = unverifiedClientRepo.Create(ctx, expiredClient)
	require.NoError(t, err)

	err = unverifiedClientRepo.Create(ctx, validClient)
	require.NoError(t, err)

	deletedCount, err := unverifiedClientRepo.DeleteExpired(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), deletedCount)

	_, err = unverifiedClientRepo.GetByID(ctx, expiredClient.ID)
	require.Error(t, err)

	result, err := unverifiedClientRepo.GetByID(ctx, validClient.ID)
	require.NoError(t, err)
	require.Equal(t, validClient.ClientID, result.ClientID)
}
