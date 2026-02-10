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

func setupUserRoleTestDB(t *testing.T) *gorm.DB {
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

	err = db.AutoMigrate(&Tenant{}, &User{}, &Role{}, &UserRole{})
	require.NoError(t, err)

	return db
}

func TestUserRoleRepository_Assign_HappyPath(t *testing.T) {
	t.Parallel()

	db := setupUserRoleTestDB(t)
	repo := NewUserRoleRepository(db)
	ctx := context.Background()

	// Create test tenant.
	tenantID := googleUuid.New()
	tenant := &Tenant{
		ID:          tenantID,
		Name:        tenantID.String(),
		Description: "Test tenant for user role assignment",
		Active:      1,
	}
	require.NoError(t, db.Create(tenant).Error)

	// Create test user.
	user := &User{
		ID:           googleUuid.New(),
		TenantID:     tenant.ID,
		Username:     googleUuid.New().String(),
		Email:        googleUuid.New().String() + "@example.com",
		PasswordHash: "hashed_password",
		Active:       1,
	}
	require.NoError(t, db.Create(user).Error)

	// Create test role.
	role := &Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "test_role",
		Description: "Test role",
	}
	require.NoError(t, db.Create(role).Error)

	// Assign role to user.
	userRole := &UserRole{
		UserID:   user.ID,
		RoleID:   role.ID,
		TenantID: tenant.ID,
	}

	err := repo.Assign(ctx, userRole)
	require.NoError(t, err)

	// Verify assignment.
	var count int64

	db.Model(&UserRole{}).Where("user_id = ? AND role_id = ?", user.ID, role.ID).Count(&count)
	require.Equal(t, int64(1), count)
}

func TestUserRoleRepository_Assign_DuplicateAssignment(t *testing.T) {
	t.Parallel()

	db := setupUserRoleTestDB(t)
	repo := NewUserRoleRepository(db)
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

	// Create test user.
	user := &User{
		ID:           googleUuid.New(),
		TenantID:     tenant.ID,
		Username:     googleUuid.New().String(),
		Email:        googleUuid.New().String() + "@example.com",
		PasswordHash: "hashed_password",
		Active:       1,
	}
	require.NoError(t, db.Create(user).Error)

	// Create test role.
	role := &Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "test_role",
		Description: "Test role",
	}
	require.NoError(t, db.Create(role).Error)

	// First assignment.
	userRole := &UserRole{
		UserID:   user.ID,
		RoleID:   role.ID,
		TenantID: tenant.ID,
	}
	require.NoError(t, repo.Assign(ctx, userRole))

	// Attempt duplicate assignment.
	duplicateUserRole := &UserRole{
		UserID:   user.ID,
		RoleID:   role.ID,
		TenantID: tenant.ID,
	}

	err := repo.Assign(ctx, duplicateUserRole)
	require.Error(t, err)
	require.Contains(t, err.Error(), "CLIENT_ERROR_CONFLICT")
}

func TestUserRoleRepository_ListRolesByUser_HappyPath_ErrorPath(t *testing.T) {
	t.Parallel()

	db := setupUserRoleTestDB(t)
	repo := NewUserRoleRepository(db)
	ctx := context.Background()

	// Test with non-existent user ID - should return empty slice, not error.
	roles, err := repo.ListRolesByUser(ctx, googleUuid.New())
	require.NoError(t, err)
	require.Empty(t, roles)
}

func TestUserRoleRepository_ListUsersByRole_ErrorPath(t *testing.T) {
	t.Parallel()

	db := setupUserRoleTestDB(t)
	repo := NewUserRoleRepository(db)
	ctx := context.Background()

	// Test with non-existent role ID - should return empty slice, not error.
	users, err := repo.ListUsersByRole(ctx, googleUuid.New())
	require.NoError(t, err)
	require.Empty(t, users)
}

func TestUserRoleRepository_ListRolesByUser_HappyPath(t *testing.T) {
	t.Parallel()

	db := setupUserRoleTestDB(t)
	repo := NewUserRoleRepository(db)
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

	// Create test user.
	userID := googleUuid.New()
	user := &User{
		ID:           userID,
		TenantID:     tenant.ID,
		Username:     userID.String(),
		Email:        googleUuid.New().String() + "@example.com",
		PasswordHash: "hashed_password",
		Active:       1,
	}
	require.NoError(t, db.Create(user).Error)

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
	require.NoError(t, repo.Assign(ctx, &UserRole{UserID: user.ID, RoleID: role1.ID, TenantID: tenant.ID}))
	require.NoError(t, repo.Assign(ctx, &UserRole{UserID: user.ID, RoleID: role2.ID, TenantID: tenant.ID}))

	// List roles.
	roles, err := repo.ListRolesByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, roles, 2)

	roleNames := []string{roles[0].Name, roles[1].Name}
	require.Contains(t, roleNames, "role1")
	require.Contains(t, roleNames, "role2")
}

func TestUserRoleRepository_ListRolesByUser_NoRoles(t *testing.T) {
	t.Parallel()

	db := setupUserRoleTestDB(t)
	repo := NewUserRoleRepository(db)
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

	// Create test user with no roles.
	userID := googleUuid.New()
	user := &User{
		ID:           userID,
		TenantID:     tenant.ID,
		Username:     userID.String(),
		Email:        googleUuid.New().String() + "@example.com",
		PasswordHash: "hashed_password",
		Active:       1,
	}
	require.NoError(t, db.Create(user).Error)

	// List roles.
	roles, err := repo.ListRolesByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Empty(t, roles)
}

func TestUserRoleRepository_Revoke_HappyPath(t *testing.T) {
	t.Parallel()

	db := setupUserRoleTestDB(t)
	repo := NewUserRoleRepository(db)
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

	// Create test user.
	user := &User{
		ID:           googleUuid.New(),
		TenantID:     tenant.ID,
		Username:     googleUuid.New().String(),
		Email:        googleUuid.New().String() + "@example.com",
		PasswordHash: "hashed_password",
		Active:       1,
	}
	require.NoError(t, db.Create(user).Error)

	// Create test role.
	role := &Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "test_role",
		Description: "Test role",
	}
	require.NoError(t, db.Create(role).Error)

	// Assign role.
	require.NoError(t, repo.Assign(ctx, &UserRole{UserID: user.ID, RoleID: role.ID, TenantID: tenant.ID}))

	// Revoke role.
	err := repo.Revoke(ctx, user.ID, role.ID)
	require.NoError(t, err)

	// Verify revocation.
	var count int64

	db.Model(&UserRole{}).Where("user_id = ? AND role_id = ?", user.ID, role.ID).Count(&count)
	require.Equal(t, int64(0), count)
}

func TestUserRoleRepository_ListUsersByRole_HappyPath(t *testing.T) {
	t.Parallel()

	db := setupUserRoleTestDB(t)
	repo := NewUserRoleRepository(db)
	ctx := context.Background()

	// Create test tenant.
	tenantID := googleUuid.New()
	tenant := &Tenant{
		ID:          tenantID,
		Name:        tenantID.String(),
		Description: "Test tenant for listing users by role",
		Active:      1,
	}
	require.NoError(t, db.Create(tenant).Error)

	// Create test role.
	role := &Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "shared_role",
		Description: "Role shared by multiple users",
	}
	require.NoError(t, db.Create(role).Error)

	// Create test users.
	user1 := &User{
		ID:           googleUuid.New(),
		TenantID:     tenant.ID,
		Username:     googleUuid.New().String(),
		Email:        googleUuid.New().String() + "@example.com",
		PasswordHash: "hashed_password",
		Active:       1,
	}
	require.NoError(t, db.Create(user1).Error)

	user2 := &User{
		ID:           googleUuid.New(),
		TenantID:     tenant.ID,
		Username:     googleUuid.New().String(),
		Email:        googleUuid.New().String() + "@example.com",
		PasswordHash: "hashed_password",
		Active:       1,
	}
	require.NoError(t, db.Create(user2).Error)

	// Assign role to both users.
	require.NoError(t, repo.Assign(ctx, &UserRole{UserID: user1.ID, RoleID: role.ID, TenantID: tenant.ID}))
	require.NoError(t, repo.Assign(ctx, &UserRole{UserID: user2.ID, RoleID: role.ID, TenantID: tenant.ID}))

	// List users by role.
	users, err := repo.ListUsersByRole(ctx, role.ID)
	require.NoError(t, err)
	require.Len(t, users, 2)

	userIDs := []googleUuid.UUID{users[0].ID, users[1].ID}
	require.Contains(t, userIDs, user1.ID)
	require.Contains(t, userIDs, user2.ID)
}

func TestUserRoleRepository_Revoke_NotFound(t *testing.T) {
	t.Parallel()

	db := setupUserRoleTestDB(t)
	repo := NewUserRoleRepository(db)
	ctx := context.Background()

	nonExistentUserID := googleUuid.New()
	nonExistentRoleID := googleUuid.New()

	err := repo.Revoke(ctx, nonExistentUserID, nonExistentRoleID)
	// GORM Delete returns nil error when no rows affected (not an error condition)
	require.NoError(t, err, "GORM Delete returns success even when record doesn't exist")
}
