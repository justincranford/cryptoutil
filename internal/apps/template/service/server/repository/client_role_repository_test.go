// Copyright 2025 Cisco Systems, Inc. and its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupClientRoleTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:test_%s?mode=memory&cache=shared", googleUuid.Must(googleUuid.NewV7()).String())

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

	err = db.AutoMigrate(&Tenant{}, &Client{}, &Role{}, &ClientRole{})
	require.NoError(t, err)

	return db
}

func TestClientRoleRepository_Assign_HappyPath(t *testing.T) {
	t.Parallel()

	db := setupClientRoleTestDB(t)
	repo := NewClientRoleRepository(db)
	ctx := context.Background()

	// Create test tenant.
	tenantID := googleUuid.New()
	tenant := &Tenant{
		ID:          tenantID,
		Name:        tenantID.String(),
		Description: "Test tenant for client role assignment",
		Active:      1,
	}
	require.NoError(t, db.Create(tenant).Error)

	// Create test client.
	client := &Client{
		ID:               googleUuid.New(),
		TenantID:         tenant.ID,
		ClientID:         googleUuid.New().String(),
		ClientSecretHash: "hashed_secret",
		Active:           1,
	}
	require.NoError(t, db.Create(client).Error)

	// Create test role.
	role := &Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "test_role",
		Description: "Test role",
	}
	require.NoError(t, db.Create(role).Error)

	// Assign role to client.
	clientRole := &ClientRole{
		ClientID: client.ID,
		RoleID:   role.ID,
		TenantID: tenant.ID,
	}

	err := repo.Assign(ctx, clientRole)
	require.NoError(t, err)

	// Verify assignment.
	var count int64

	db.Model(&ClientRole{}).Where("client_id = ? AND role_id = ?", client.ID, role.ID).Count(&count)
	require.Equal(t, int64(1), count)
}

func TestClientRoleRepository_Assign_DuplicateAssignment(t *testing.T) {
	t.Parallel()

	db := setupClientRoleTestDB(t)
	repo := NewClientRoleRepository(db)
	ctx := context.Background()

	// Create test tenant.
	tenantID := googleUuid.New()
	tenant := &Tenant{
		ID:          tenantID,
		Name:        tenantID.String(),
		Description: "Test tenant for duplicate assignment",
		Active:      1,
	}
	require.NoError(t, db.Create(tenant).Error)

	// Create test client.
	client := &Client{
		ID:               googleUuid.New(),
		TenantID:         tenant.ID,
		ClientID:         googleUuid.New().String(),
		ClientSecretHash: "hashed_secret",
		Active:           1,
	}
	require.NoError(t, db.Create(client).Error)

	// Create test role.
	role := &Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "test_role",
		Description: "Test role",
	}
	require.NoError(t, db.Create(role).Error)

	// First assignment.
	clientRole := &ClientRole{
		ClientID: client.ID,
		RoleID:   role.ID,
		TenantID: tenant.ID,
	}
	require.NoError(t, repo.Assign(ctx, clientRole))

	// Attempt duplicate assignment.
	duplicateClientRole := &ClientRole{
		ClientID: client.ID,
		RoleID:   role.ID,
		TenantID: tenant.ID,
	}

	err := repo.Assign(ctx, duplicateClientRole)
	require.Error(t, err)
	require.Contains(t, err.Error(), "CLIENT_ERROR_CONFLICT")
}

func TestClientRoleRepository_ListRolesByClient_ErrorPath(t *testing.T) {
	t.Parallel()

	db := setupClientRoleTestDB(t)
	repo := NewClientRoleRepository(db)
	ctx := context.Background()

	// Test with non-existent client ID - should return empty slice, not error.
	roles, err := repo.ListRolesByClient(ctx, googleUuid.New())
	require.NoError(t, err)
	require.Empty(t, roles)
}

func TestClientRoleRepository_ListClientsByRole_ErrorPath(t *testing.T) {
	t.Parallel()

	db := setupClientRoleTestDB(t)
	repo := NewClientRoleRepository(db)
	ctx := context.Background()

	// Test with non-existent role ID - should return empty slice, not error.
	clients, err := repo.ListClientsByRole(ctx, googleUuid.New())
	require.NoError(t, err)
	require.Empty(t, clients)
}

func TestClientRoleRepository_ListRolesByClient_HappyPath(t *testing.T) {
	t.Parallel()

	db := setupClientRoleTestDB(t)
	repo := NewClientRoleRepository(db)
	ctx := context.Background()

	// Create test tenant.
	tenantID := googleUuid.New()
	tenant := &Tenant{
		ID:          tenantID,
		Name:        tenantID.String(),
		Description: "Test tenant for listing roles",
		Active:      1,
	}
	require.NoError(t, db.Create(tenant).Error)

	// Create test client.
	clientUUID := googleUuid.New()
	client := &Client{
		ID:               clientUUID,
		TenantID:         tenant.ID,
		ClientID:         clientUUID.String(),
		ClientSecretHash: "hashed_secret",
		Active:           1,
	}
	require.NoError(t, db.Create(client).Error)

	// Create test roles.
	role1 := &Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "role1",
		Description: "First role",
	}
	require.NoError(t, db.Create(role1).Error)

	role2 := &Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "role2",
		Description: "Second role",
	}
	require.NoError(t, db.Create(role2).Error)

	// Assign roles.
	require.NoError(t, repo.Assign(ctx, &ClientRole{ClientID: client.ID, RoleID: role1.ID, TenantID: tenant.ID}))
	require.NoError(t, repo.Assign(ctx, &ClientRole{ClientID: client.ID, RoleID: role2.ID, TenantID: tenant.ID}))

	// List roles.
	roles, err := repo.ListRolesByClient(ctx, client.ID)
	require.NoError(t, err)
	require.Len(t, roles, 2)

	roleNames := []string{roles[0].Name, roles[1].Name}
	require.Contains(t, roleNames, "role1")
	require.Contains(t, roleNames, "role2")
}

func TestClientRoleRepository_ListRolesByClient_NoRoles(t *testing.T) {
	t.Parallel()

	db := setupClientRoleTestDB(t)
	repo := NewClientRoleRepository(db)
	ctx := context.Background()

	// Create test tenant.
	tenantID := googleUuid.New()
	tenant := &Tenant{
		ID:          tenantID,
		Name:        tenantID.String(),
		Description: "Test tenant for no roles",
		Active:      1,
	}
	require.NoError(t, db.Create(tenant).Error)

	// Create test client with no roles.
	clientUUID := googleUuid.New()
	client := &Client{
		ID:               clientUUID,
		TenantID:         tenant.ID,
		ClientID:         clientUUID.String(),
		ClientSecretHash: "hashed_secret",
		Active:           1,
	}
	require.NoError(t, db.Create(client).Error)

	// List roles.
	roles, err := repo.ListRolesByClient(ctx, client.ID)
	require.NoError(t, err)
	require.Empty(t, roles)
}

func TestClientRoleRepository_Revoke_HappyPath(t *testing.T) {
	t.Parallel()

	db := setupClientRoleTestDB(t)
	repo := NewClientRoleRepository(db)
	ctx := context.Background()

	// Create test tenant.
	tenantID := googleUuid.New()
	tenant := &Tenant{
		ID:          tenantID,
		Name:        tenantID.String(),
		Description: "Test tenant for revoke",
		Active:      1,
	}
	require.NoError(t, db.Create(tenant).Error)

	// Create test client.
	client := &Client{
		ID:               googleUuid.New(),
		TenantID:         tenant.ID,
		ClientID:         googleUuid.New().String(),
		ClientSecretHash: "hashed_secret",
		Active:           1,
	}
	require.NoError(t, db.Create(client).Error)

	// Create test role.
	role := &Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "test_role",
		Description: "Test role",
	}
	require.NoError(t, db.Create(role).Error)

	// Assign role.
	require.NoError(t, repo.Assign(ctx, &ClientRole{ClientID: client.ID, RoleID: role.ID, TenantID: tenant.ID}))

	// Revoke role.
	err := repo.Revoke(ctx, client.ID, role.ID)
	require.NoError(t, err)

	// Verify revocation.
	var count int64

	db.Model(&ClientRole{}).Where("client_id = ? AND role_id = ?", client.ID, role.ID).Count(&count)
	require.Equal(t, int64(0), count)
}

func TestClientRoleRepository_ListClientsByRole_HappyPath(t *testing.T) {
	t.Parallel()

	db := setupClientRoleTestDB(t)
	repo := NewClientRoleRepository(db)
	ctx := context.Background()

	// Create test tenant.
	tenantID := googleUuid.New()
	tenant := &Tenant{
		ID:          tenantID,
		Name:        tenantID.String(),
		Description: "Test tenant for listing clients by role",
		Active:      1,
	}
	require.NoError(t, db.Create(tenant).Error)

	// Create test role.
	role := &Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "shared_role",
		Description: "Role shared by multiple clients",
	}
	require.NoError(t, db.Create(role).Error)

	// Create test clients.
	client1 := &Client{
		ID:               googleUuid.New(),
		TenantID:         tenant.ID,
		ClientID:         googleUuid.New().String(),
		ClientSecretHash: "hashed_secret1",
		Active:           1,
	}
	require.NoError(t, db.Create(client1).Error)

	client2 := &Client{
		ID:               googleUuid.New(),
		TenantID:         tenant.ID,
		ClientID:         googleUuid.New().String(),
		ClientSecretHash: "hashed_secret2",
		Active:           1,
	}
	require.NoError(t, db.Create(client2).Error)

	// Assign role to both clients.
	require.NoError(t, repo.Assign(ctx, &ClientRole{ClientID: client1.ID, RoleID: role.ID, TenantID: tenant.ID}))
	require.NoError(t, repo.Assign(ctx, &ClientRole{ClientID: client2.ID, RoleID: role.ID, TenantID: tenant.ID}))

	// List clients by role.
	clients, err := repo.ListClientsByRole(ctx, role.ID)
	require.NoError(t, err)
	require.Len(t, clients, 2)

	clientIDs := []googleUuid.UUID{clients[0].ID, clients[1].ID}
	require.Contains(t, clientIDs, client1.ID)
	require.Contains(t, clientIDs, client2.ID)
}

func TestClientRoleRepository_Revoke_NotFound(t *testing.T) {
	t.Parallel()

	db := setupClientRoleTestDB(t)
	repo := NewClientRoleRepository(db)
	ctx := context.Background()

	nonExistentClientID := googleUuid.New()
	nonExistentRoleID := googleUuid.New()

	err := repo.Revoke(ctx, nonExistentClientID, nonExistentRoleID)
	// GORM Delete returns nil error when no rows affected (not an error condition)
	require.NoError(t, err, "GORM Delete returns success even when record doesn't exist")
}
