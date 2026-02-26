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

func TestRoleRepository_GetByID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	roleRepo := NewRoleRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	role := &Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "admin",
		Description: "Administrator role",
		CreatedAt:   time.Now().UTC(),
	}

	err = roleRepo.Create(ctx, role)
	require.NoError(t, err)

	result, err := roleRepo.GetByID(ctx, role.ID)
	require.NoError(t, err)
	require.Equal(t, role.ID, result.ID)
	require.Equal(t, role.Name, result.Name)
}

func TestRoleRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	roleRepo := NewRoleRepository(db)
	ctx := context.Background()

	_, err := roleRepo.GetByID(ctx, googleUuid.New())
	require.Error(t, err)
}

func TestRoleRepository_ListByTenant(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	roleRepo := NewRoleRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	role1 := &Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "admin",
		Description: "Administrator",
		CreatedAt:   time.Now().UTC(),
	}

	role2 := &Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "user",
		Description: "Regular user",
		CreatedAt:   time.Now().UTC(),
	}

	err = roleRepo.Create(ctx, role1)
	require.NoError(t, err)

	err = roleRepo.Create(ctx, role2)
	require.NoError(t, err)

	results, err := roleRepo.ListByTenant(ctx, tenant.ID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(results), 2)
}

func TestRoleRepository_Delete(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	roleRepo := NewRoleRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	role := &Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "deletable",
		Description: "Role to delete",
		CreatedAt:   time.Now().UTC(),
	}

	err = roleRepo.Create(ctx, role)
	require.NoError(t, err)

	err = roleRepo.Delete(ctx, role.ID)
	require.NoError(t, err)

	_, err = roleRepo.GetByID(ctx, role.ID)
	require.Error(t, err)
}

func TestUserRoleRepository_ListRolesByUser(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	roleRepo := NewRoleRepository(db)
	userRoleRepo := NewUserRoleRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueTenantName("Test"),
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
		Email:    "user_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
		Active:   1,
	}

	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	role1 := &Role{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		Name:     "role1_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
	}

	role2 := &Role{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		Name:     "role2_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
	}

	err = roleRepo.Create(ctx, role1)
	require.NoError(t, err)

	err = roleRepo.Create(ctx, role2)
	require.NoError(t, err)

	err = userRoleRepo.Assign(ctx, &UserRole{
		UserID:   user.ID,
		RoleID:   role1.ID,
		TenantID: tenant.ID,
	})
	require.NoError(t, err)

	err = userRoleRepo.Assign(ctx, &UserRole{
		UserID:   user.ID,
		RoleID:   role2.ID,
		TenantID: tenant.ID,
	})
	require.NoError(t, err)

	roles, err := userRoleRepo.ListRolesByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, roles, 2)
}

func TestUserRoleRepository_ListUsersByRole(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	roleRepo := NewRoleRepository(db)
	userRoleRepo := NewUserRoleRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	user1 := &User{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		Username: "user1_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Email:    "user1_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
		Active:   1,
	}

	user2 := &User{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		Username: "user2_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Email:    "user2_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
		Active:   1,
	}

	err = userRepo.Create(ctx, user1)
	require.NoError(t, err)

	err = userRepo.Create(ctx, user2)
	require.NoError(t, err)

	role := &Role{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		Name:     "admin_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
	}

	err = roleRepo.Create(ctx, role)
	require.NoError(t, err)

	err = userRoleRepo.Assign(ctx, &UserRole{
		UserID:   user1.ID,
		RoleID:   role.ID,
		TenantID: tenant.ID,
	})
	require.NoError(t, err)

	err = userRoleRepo.Assign(ctx, &UserRole{
		UserID:   user2.ID,
		RoleID:   role.ID,
		TenantID: tenant.ID,
	})
	require.NoError(t, err)

	users, err := userRoleRepo.ListUsersByRole(ctx, role.ID)
	require.NoError(t, err)
	require.Len(t, users, 2)
}

func TestClientRoleRepository_Revoke(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	clientRepo := NewClientRepository(db)
	roleRepo := NewRoleRepository(db)
	clientRoleRepo := NewClientRoleRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueTenantName("Test"),
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

	role := &Role{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		Name:     "service_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
	}

	err = roleRepo.Create(ctx, role)
	require.NoError(t, err)

	err = clientRoleRepo.Assign(ctx, &ClientRole{
		ClientID: client.ID,
		RoleID:   role.ID,
		TenantID: tenant.ID,
	})
	require.NoError(t, err)

	err = clientRoleRepo.Revoke(ctx, client.ID, role.ID)
	require.NoError(t, err)

	roles, err := clientRoleRepo.ListRolesByClient(ctx, client.ID)
	require.NoError(t, err)
	require.Len(t, roles, 0)
}

func TestClientRoleRepository_ListRolesByClient(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	clientRepo := NewClientRepository(db)
	roleRepo := NewRoleRepository(db)
	clientRoleRepo := NewClientRoleRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueTenantName("Test"),
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

	role1 := &Role{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		Name:     "read_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
	}

	role2 := &Role{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		Name:     "write_" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
	}

	err = roleRepo.Create(ctx, role1)
	require.NoError(t, err)

	err = roleRepo.Create(ctx, role2)
	require.NoError(t, err)

	err = clientRoleRepo.Assign(ctx, &ClientRole{
		ClientID: client.ID,
		RoleID:   role1.ID,
		TenantID: tenant.ID,
	})
	require.NoError(t, err)

	err = clientRoleRepo.Assign(ctx, &ClientRole{
		ClientID: client.ID,
		RoleID:   role2.ID,
		TenantID: tenant.ID,
	})
	require.NoError(t, err)

	roles, err := clientRoleRepo.ListRolesByClient(ctx, client.ID)
	require.NoError(t, err)
	require.Len(t, roles, 2)
}
