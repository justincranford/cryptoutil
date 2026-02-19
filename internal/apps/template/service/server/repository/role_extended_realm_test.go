// Copyright 2025 Marlon Almeida. All rights reserved.

package repository

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestClientRoleRepository_ListClientsByRole(t *testing.T) {
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

	client1 := &Client{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		ClientID: "client1_" + googleUuid.New().String()[:8],
		Active:   1,
	}

	client2 := &Client{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		ClientID: "client2_" + googleUuid.New().String()[:8],
		Active:   1,
	}

	err = clientRepo.Create(ctx, client1)
	require.NoError(t, err)

	err = clientRepo.Create(ctx, client2)
	require.NoError(t, err)

	role := &Role{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		Name:     "api_access_" + googleUuid.New().String()[:8],
	}

	err = roleRepo.Create(ctx, role)
	require.NoError(t, err)

	err = clientRoleRepo.Assign(ctx, &ClientRole{
		ClientID: client1.ID,
		RoleID:   role.ID,
		TenantID: tenant.ID,
	})
	require.NoError(t, err)

	err = clientRoleRepo.Assign(ctx, &ClientRole{
		ClientID: client2.ID,
		RoleID:   role.ID,
		TenantID: tenant.ID,
	})
	require.NoError(t, err)

	clients, err := clientRoleRepo.ListClientsByRole(ctx, role.ID)
	require.NoError(t, err)
	require.Len(t, clients, 2)
}

func TestTenantRealmRepository_GetByID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	realmRepo := NewTenantRealmRepository(db)
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

	realm := &TenantRealm{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		RealmID:  googleUuid.New(),
		Type:     "OIDC",
		Active:   true,
		Source:   "external",
	}

	err = realmRepo.Create(ctx, realm)
	require.NoError(t, err)

	result, err := realmRepo.GetByID(ctx, realm.ID)
	require.NoError(t, err)
	require.Equal(t, realm.ID, result.ID)
	require.Equal(t, realm.Type, result.Type)
}

func TestTenantRealmRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	realmRepo := NewTenantRealmRepository(db)
	ctx := context.Background()

	_, err := realmRepo.GetByID(ctx, googleUuid.New())
	require.Error(t, err)
}

func TestTenantRealmRepository_GetByRealmID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	realmRepo := NewTenantRealmRepository(db)
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

	realmID := googleUuid.New()

	realm := &TenantRealm{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		RealmID:  realmID,
		Type:     "SAML",
		Active:   true,
		Source:   "external",
	}

	err = realmRepo.Create(ctx, realm)
	require.NoError(t, err)

	result, err := realmRepo.GetByRealmID(ctx, tenant.ID, realmID)
	require.NoError(t, err)
	require.Equal(t, realmID, result.RealmID)
}

func TestTenantRealmRepository_GetByRealmID_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	realmRepo := NewTenantRealmRepository(db)
	ctx := context.Background()

	_, err := realmRepo.GetByRealmID(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
}

func TestTenantRealmRepository_Update(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	realmRepo := NewTenantRealmRepository(db)
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

	realm := &TenantRealm{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		RealmID:  googleUuid.New(),
		Type:     "LDAP",
		Active:   true,
		Source:   "external",
	}

	err = realmRepo.Create(ctx, realm)
	require.NoError(t, err)

	realm.Active = false

	err = realmRepo.Update(ctx, realm)
	require.NoError(t, err)

	result, err := realmRepo.GetByID(ctx, realm.ID)
	require.NoError(t, err)
	require.False(t, result.Active)
}

func TestTenantRealmRepository_Delete(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	realmRepo := NewTenantRealmRepository(db)
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

	realm := &TenantRealm{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		RealmID:  googleUuid.New(),
		Type:     "DB",
		Active:   true,
		Source:   "local",
	}

	err = realmRepo.Create(ctx, realm)
	require.NoError(t, err)

	err = realmRepo.Delete(ctx, realm.ID)
	require.NoError(t, err)

	_, err = realmRepo.GetByID(ctx, realm.ID)
	require.Error(t, err)
}
