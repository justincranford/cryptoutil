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

// uniqueRoleTenantName returns a unique tenant name for tests.
func uniqueRoleTenantName(base string) string {
	return base + " " + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength]
}

func TestRoleRepository_Create(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	roleRepo := NewRoleRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueRoleTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Create first role for duplicate test.
	firstRole := &Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "admin",
		Description: "Administrator role",
		CreatedAt:   time.Now().UTC(),
	}
	err = roleRepo.Create(ctx, firstRole)
	require.NoError(t, err)

	tests := []struct {
		name      string
		role      *Role
		wantError bool
	}{
		{
			name: "happy path - valid role",
			role: &Role{
				ID:          googleUuid.New(),
				TenantID:    tenant.ID,
				Name:        "user",
				Description: "User role",
				CreatedAt:   time.Now().UTC(),
			},
			wantError: false,
		},
		{
			name: "duplicate role name for tenant",
			role: &Role{
				ID:          googleUuid.New(),
				TenantID:    tenant.ID,
				Name:        "admin",
				Description: "Duplicate admin role",
				CreatedAt:   time.Now().UTC(),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := roleRepo.Create(ctx, tt.role)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRoleRepository_GetByName(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	roleRepo := NewRoleRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueRoleTenantName("Test"),
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

	tests := []struct {
		name      string
		tenantID  googleUuid.UUID
		roleName  string
		wantError bool
	}{
		{
			name:      "happy path - existing role",
			tenantID:  tenant.ID,
			roleName:  "admin",
			wantError: false,
		},
		{
			name:      "not found - wrong tenant",
			tenantID:  googleUuid.New(),
			roleName:  "admin",
			wantError: true,
		},
		{
			name:      "not found - wrong role name",
			tenantID:  tenant.ID,
			roleName:  "nonexistent",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := roleRepo.GetByName(ctx, tt.tenantID, tt.roleName)

			if tt.wantError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, role.ID, result.ID)
				require.Equal(t, role.Name, result.Name)
			}
		})
	}
}

func TestUserRoleRepository_Assign(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	roleRepo := NewRoleRepository(db)
	userRoleRepo := NewUserRoleRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueRoleTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	user := &User{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		Username:  "testuser2-" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Email:     "test-" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
		Active:    1,
		CreatedAt: time.Now().UTC(),
	}

	err = userRepo.Create(ctx, user)
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

	userRole := &UserRole{
		UserID:    user.ID,
		RoleID:    role.ID,
		TenantID:  tenant.ID,
		CreatedAt: time.Now().UTC(),
	}

	err = userRoleRepo.Assign(ctx, userRole)
	require.NoError(t, err)

	roles, err := userRoleRepo.ListRolesByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, roles, 1)
	require.Equal(t, role.ID, roles[0].ID)
}

func TestUserRoleRepository_Revoke(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	roleRepo := NewRoleRepository(db)
	userRoleRepo := NewUserRoleRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueRoleTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	user := &User{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		Username:  "testuser",
		Email:     "test@example.com",
		Active:    1,
		CreatedAt: time.Now().UTC(),
	}

	err = userRepo.Create(ctx, user)
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

	userRole := &UserRole{
		UserID:    user.ID,
		RoleID:    role.ID,
		TenantID:  tenant.ID,
		CreatedAt: time.Now().UTC(),
	}

	err = userRoleRepo.Assign(ctx, userRole)
	require.NoError(t, err)

	err = userRoleRepo.Revoke(ctx, user.ID, role.ID)
	require.NoError(t, err)

	roles, err := userRoleRepo.ListRolesByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, roles, 0)
}

func TestClientRoleRepository_Assign(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	clientRepo := NewClientRepository(db)
	roleRepo := NewRoleRepository(db)
	clientRoleRepo := NewClientRoleRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueRoleTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	client := &Client{
		ID:        googleUuid.New(),
		TenantID:  tenant.ID,
		ClientID:  "client-" + googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Active:    1,
		CreatedAt: time.Now().UTC(),
	}

	err = clientRepo.Create(ctx, client)
	require.NoError(t, err)

	role := &Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "service",
		Description: "Service role",
		CreatedAt:   time.Now().UTC(),
	}

	err = roleRepo.Create(ctx, role)
	require.NoError(t, err)

	clientRole := &ClientRole{
		ClientID:  client.ID,
		RoleID:    role.ID,
		TenantID:  tenant.ID,
		CreatedAt: time.Now().UTC(),
	}

	err = clientRoleRepo.Assign(ctx, clientRole)
	require.NoError(t, err)

	roles, err := clientRoleRepo.ListRolesByClient(ctx, client.ID)
	require.NoError(t, err)
	require.Len(t, roles, 1)
	require.Equal(t, role.ID, roles[0].ID)
}

func TestTenantRealmRepository_Create(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	tenantRepo := NewTenantRepository(db)
	realmRepo := NewTenantRealmRepository(db)
	ctx := context.Background()

	tenant := &Tenant{
		ID:          googleUuid.New(),
		Name:        uniqueRoleTenantName("Test"),
		Description: "Test tenant",
		Active:      1,
		CreatedAt:   time.Now().UTC(),
	}

	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	tests := []struct {
		name      string
		realm     *TenantRealm
		wantError bool
	}{
		{
			name: "happy path - DB realm",
			realm: &TenantRealm{
				ID:        googleUuid.New(),
				TenantID:  tenant.ID,
				RealmID:   googleUuid.New(),
				Type:      "DB",
				Active:    true,
				CreatedAt: time.Now().UTC(),
			},
			wantError: false,
		},
		{
			name: "duplicate realm ID for tenant",
			realm: func() *TenantRealm {
				existingRealm := &TenantRealm{
					ID:        googleUuid.New(),
					TenantID:  tenant.ID,
					RealmID:   googleUuid.New(),
					Type:      "DB",
					Active:    true,
					CreatedAt: time.Now().UTC(),
				}
				err := realmRepo.Create(ctx, existingRealm)
				require.NoError(t, err)

				return &TenantRealm{
					ID:        googleUuid.New(),
					TenantID:  tenant.ID,
					RealmID:   existingRealm.RealmID,
					Type:      "DB",
					Active:    true,
					CreatedAt: time.Now().UTC(),
				}
			}(),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := realmRepo.Create(ctx, tt.realm)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
