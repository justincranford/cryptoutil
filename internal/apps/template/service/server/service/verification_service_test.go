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

package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// setupVerificationTestDB creates an in-memory SQLite database for testing verification service.
func setupVerificationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Create unique database name to avoid sharing between tests.
	dbName := fmt.Sprintf("file:test_%s.db?mode=memory&cache=private", strings.ReplaceAll(t.Name(), "/", "_"))
	sqlDB, err := sql.Open("sqlite", dbName)
	require.NoError(t, err)

	// Enable WAL mode for better concurrency.
	_, err = sqlDB.ExecContext(context.Background(), "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	// Set busy timeout for concurrent writes.
	_, err = sqlDB.ExecContext(context.Background(), "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	// Pass to GORM with auto-transactions disabled.
	dialector := sqlite.Dialector{Conn: sqlDB}
	db, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	// Configure connection pool.
	sqlDB, err = db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(0)

	// Auto-migrate all required tables.
	err = db.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerRepository.Tenant{},
		&cryptoutilAppsTemplateServiceServerRepository.User{},
		&cryptoutilAppsTemplateServiceServerRepository.Client{},
		&cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{},
		&cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{},
		&cryptoutilAppsTemplateServiceServerRepository.Role{},
		&cryptoutilAppsTemplateServiceServerRepository.UserRole{},
		&cryptoutilAppsTemplateServiceServerRepository.ClientRole{},
	)
	require.NoError(t, err)

	return db
}

// setupVerificationService creates a VerificationService with all dependencies for testing.
func setupVerificationService(t *testing.T) (VerificationService, *gorm.DB) {
	t.Helper()

	db := setupVerificationTestDB(t)

	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(db)
	clientRepo := cryptoutilAppsTemplateServiceServerRepository.NewClientRepository(db)
	unverifiedUserRepo := cryptoutilAppsTemplateServiceServerRepository.NewUnverifiedUserRepository(db)
	unverifiedClientRepo := cryptoutilAppsTemplateServiceServerRepository.NewUnverifiedClientRepository(db)
	roleRepo := cryptoutilAppsTemplateServiceServerRepository.NewRoleRepository(db)
	userRoleRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRoleRepository(db)
	clientRoleRepo := cryptoutilAppsTemplateServiceServerRepository.NewClientRoleRepository(db)

	svc := NewVerificationService(
		userRepo,
		clientRepo,
		unverifiedUserRepo,
		unverifiedClientRepo,
		roleRepo,
		userRoleRepo,
		clientRoleRepo,
	)

	return svc, db
}

// createTestTenantAndRole creates a tenant and role for testing.
func createTestTenantAndRole(t *testing.T, db *gorm.DB, tenantName string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, *cryptoutilAppsTemplateServiceServerRepository.Role) {
	t.Helper()

	tenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{
		ID:          googleUuid.New(),
		Name:        tenantName,
		Description: "Test tenant",
		Active:      1,
	}
	require.NoError(t, db.Create(tenant).Error)

	role := &cryptoutilAppsTemplateServiceServerRepository.Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "user_" + googleUuid.NewString()[:8],
		Description: "Test role",
	}
	require.NoError(t, db.Create(role).Error)

	return tenant, role
}

// TestVerificationService_ListPendingUsers tests listing pending user registrations.
func TestVerificationService_ListPendingUsers(t *testing.T) {
	t.Parallel()

	svc, db := setupVerificationService(t)
	ctx := context.Background()

	tenant, _ := createTestTenantAndRole(t, db, "list-pending-users-"+googleUuid.NewString()[:8])

	// Create pending users.
	for i := 0; i < 3; i++ {
		unverifiedUser := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
			ID:           googleUuid.New(),
			TenantID:     tenant.ID,
			Username:     "user" + googleUuid.NewString()[:8],
			Email:        "user" + googleUuid.NewString()[:8] + "@example.com",
			PasswordHash: "hash",
			ExpiresAt:    time.Now().UTC().Add(72 * time.Hour),
		}
		require.NoError(t, db.Create(unverifiedUser).Error)
	}

	// List pending users.
	pendingUsers, err := svc.ListPendingUsers(ctx, tenant.ID)
	require.NoError(t, err)
	require.Len(t, pendingUsers, 3)
}

// TestVerificationService_ListPendingClients tests listing pending client registrations.
func TestVerificationService_ListPendingClients(t *testing.T) {
	t.Parallel()

	svc, db := setupVerificationService(t)
	ctx := context.Background()

	tenant, _ := createTestTenantAndRole(t, db, "list-pending-clients-"+googleUuid.NewString()[:8])

	// Create pending clients.
	for i := 0; i < 2; i++ {
		unverifiedClient := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
			ID:               googleUuid.New(),
			TenantID:         tenant.ID,
			ClientID:         "client" + googleUuid.NewString()[:8],
			ClientSecretHash: "secret",
			ExpiresAt:        time.Now().UTC().Add(72 * time.Hour),
		}
		require.NoError(t, db.Create(unverifiedClient).Error)
	}

	// List pending clients.
	pendingClients, err := svc.ListPendingClients(ctx, tenant.ID)
	require.NoError(t, err)
	require.Len(t, pendingClients, 2)
}

// TestVerificationService_ApproveUser_Success tests successful user approval.
func TestVerificationService_ApproveUser_Success(t *testing.T) {
	t.Parallel()

	svc, db := setupVerificationService(t)
	ctx := context.Background()

	tenant, role := createTestTenantAndRole(t, db, "approve-user-"+googleUuid.NewString()[:8])

	// Create unverified user.
	unverifiedUser := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
		ID:           googleUuid.New(),
		TenantID:     tenant.ID,
		Username:     "pendinguser" + googleUuid.NewString()[:8],
		Email:        "pending" + googleUuid.NewString()[:8] + "@example.com",
		PasswordHash: "hashedpassword",
		ExpiresAt:    time.Now().UTC().Add(72 * time.Hour),
	}
	require.NoError(t, db.Create(unverifiedUser).Error)

	// Approve user.
	user, err := svc.ApproveUser(ctx, tenant.ID, unverifiedUser.ID, []googleUuid.UUID{role.ID})
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, tenant.ID, user.TenantID)
	require.Equal(t, unverifiedUser.Username, user.Username)
	require.Equal(t, unverifiedUser.Email, user.Email)
	require.Equal(t, 1, user.Active)

	// Verify unverified user was deleted.
	var count int64

	db.Model(&cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{}).Where("id = ?", unverifiedUser.ID).Count(&count)
	require.Equal(t, int64(0), count)

	// Verify role was assigned.
	var userRole cryptoutilAppsTemplateServiceServerRepository.UserRole

	err = db.Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&userRole).Error
	require.NoError(t, err)
}

// TestVerificationService_ApproveUser_ExpiredRegistration tests approval of expired registration.
func TestVerificationService_ApproveUser_ExpiredRegistration(t *testing.T) {
	t.Parallel()

	svc, db := setupVerificationService(t)
	ctx := context.Background()

	tenant, role := createTestTenantAndRole(t, db, "approve-expired-"+googleUuid.NewString()[:8])

	// Create expired unverified user.
	unverifiedUser := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
		ID:           googleUuid.New(),
		TenantID:     tenant.ID,
		Username:     "expireduser" + googleUuid.NewString()[:8],
		Email:        "expired" + googleUuid.NewString()[:8] + "@example.com",
		PasswordHash: "hashedpassword",
		ExpiresAt:    time.Now().UTC().Add(-1 * time.Hour), // Already expired.
	}
	require.NoError(t, db.Create(unverifiedUser).Error)

	// Try to approve - should fail.
	_, err := svc.ApproveUser(ctx, tenant.ID, unverifiedUser.ID, []googleUuid.UUID{role.ID})
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")
}

// TestVerificationService_ApproveUser_WrongTenant tests approval with mismatched tenant.
func TestVerificationService_ApproveUser_WrongTenant(t *testing.T) {
	t.Parallel()

	svc, db := setupVerificationService(t)
	ctx := context.Background()

	tenant1, role := createTestTenantAndRole(t, db, "tenant1-"+googleUuid.NewString()[:8])
	tenant2, _ := createTestTenantAndRole(t, db, "tenant2-"+googleUuid.NewString()[:8])

	// Create unverified user in tenant1.
	unverifiedUser := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
		ID:           googleUuid.New(),
		TenantID:     tenant1.ID,
		Username:     "wrongtenantuser" + googleUuid.NewString()[:8],
		Email:        "wrong" + googleUuid.NewString()[:8] + "@example.com",
		PasswordHash: "hashedpassword",
		ExpiresAt:    time.Now().UTC().Add(72 * time.Hour),
	}
	require.NoError(t, db.Create(unverifiedUser).Error)

	// Try to approve from tenant2 - should fail.
	_, err := svc.ApproveUser(ctx, tenant2.ID, unverifiedUser.ID, []googleUuid.UUID{role.ID})
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong to the specified tenant")
}

// TestVerificationService_ApproveUser_NoRoles tests approval without roles.
func TestVerificationService_ApproveUser_NoRoles(t *testing.T) {
	t.Parallel()

	svc, db := setupVerificationService(t)
	ctx := context.Background()

	tenant, _ := createTestTenantAndRole(t, db, "no-roles-"+googleUuid.NewString()[:8])

	// Create unverified user.
	unverifiedUser := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
		ID:           googleUuid.New(),
		TenantID:     tenant.ID,
		Username:     "norolesuser" + googleUuid.NewString()[:8],
		Email:        "noroles" + googleUuid.NewString()[:8] + "@example.com",
		PasswordHash: "hashedpassword",
		ExpiresAt:    time.Now().UTC().Add(72 * time.Hour),
	}
	require.NoError(t, db.Create(unverifiedUser).Error)

	// Try to approve without roles - should fail.
	_, err := svc.ApproveUser(ctx, tenant.ID, unverifiedUser.ID, []googleUuid.UUID{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "at least one role")
}

// TestVerificationService_ApproveClient_Success tests successful client approval.
func TestVerificationService_ApproveClient_Success(t *testing.T) {
	t.Parallel()

	svc, db := setupVerificationService(t)
	ctx := context.Background()

	tenant, role := createTestTenantAndRole(t, db, "approve-client-"+googleUuid.NewString()[:8])

	// Create unverified client.
	unverifiedClient := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
		ID:               googleUuid.New(),
		TenantID:         tenant.ID,
		ClientID:         "pendingclient" + googleUuid.NewString()[:8],
		ClientSecretHash: "clientsecret",
		ExpiresAt:        time.Now().UTC().Add(72 * time.Hour),
	}
	require.NoError(t, db.Create(unverifiedClient).Error)

	// Approve client.
	client, err := svc.ApproveClient(ctx, tenant.ID, unverifiedClient.ID, []googleUuid.UUID{role.ID})
	require.NoError(t, err)
	require.NotNil(t, client)
	require.Equal(t, tenant.ID, client.TenantID)
	require.Equal(t, unverifiedClient.ClientID, client.ClientID)
	require.Equal(t, 1, client.Active)

	// Verify unverified client was deleted.
	var count int64

	db.Model(&cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{}).Where("id = ?", unverifiedClient.ID).Count(&count)
	require.Equal(t, int64(0), count)
}

// TestVerificationService_RejectUser tests rejecting a pending user registration.
func TestVerificationService_RejectUser(t *testing.T) {
	t.Parallel()

	svc, db := setupVerificationService(t)
	ctx := context.Background()

	tenant, _ := createTestTenantAndRole(t, db, "reject-user-"+googleUuid.NewString()[:8])

	// Create unverified user.
	unverifiedUser := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
		ID:           googleUuid.New(),
		TenantID:     tenant.ID,
		Username:     "rejectuser" + googleUuid.NewString()[:8],
		Email:        "reject" + googleUuid.NewString()[:8] + "@example.com",
		PasswordHash: "hashedpassword",
		ExpiresAt:    time.Now().UTC().Add(72 * time.Hour),
	}
	require.NoError(t, db.Create(unverifiedUser).Error)

	// Reject user.
	err := svc.RejectUser(ctx, tenant.ID, unverifiedUser.ID)
	require.NoError(t, err)

	// Verify unverified user was deleted.
	var count int64

	db.Model(&cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{}).Where("id = ?", unverifiedUser.ID).Count(&count)
	require.Equal(t, int64(0), count)
}

// TestVerificationService_RejectClient tests rejecting a pending client registration.
func TestVerificationService_RejectClient(t *testing.T) {
	t.Parallel()

	svc, db := setupVerificationService(t)
	ctx := context.Background()

	tenant, _ := createTestTenantAndRole(t, db, "reject-client-"+googleUuid.NewString()[:8])

	// Create unverified client.
	unverifiedClient := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
		ID:               googleUuid.New(),
		TenantID:         tenant.ID,
		ClientID:         "rejectclient" + googleUuid.NewString()[:8],
		ClientSecretHash: "secret",
		ExpiresAt:        time.Now().UTC().Add(72 * time.Hour),
	}
	require.NoError(t, db.Create(unverifiedClient).Error)

	// Reject client.
	err := svc.RejectClient(ctx, tenant.ID, unverifiedClient.ID)
	require.NoError(t, err)

	// Verify unverified client was deleted.
	var count int64

	db.Model(&cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{}).Where("id = ?", unverifiedClient.ID).Count(&count)
	require.Equal(t, int64(0), count)
}

// TestVerificationService_CleanupExpiredRegistrations tests cleanup of expired registrations.
func TestVerificationService_CleanupExpiredRegistrations(t *testing.T) {
	t.Parallel()

	svc, db := setupVerificationService(t)
	ctx := context.Background()

	tenant, _ := createTestTenantAndRole(t, db, "cleanup-"+googleUuid.NewString()[:8])

	// Create expired and non-expired users.
	expiredUser := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
		ID:           googleUuid.New(),
		TenantID:     tenant.ID,
		Username:     "expireduser" + googleUuid.NewString()[:8],
		Email:        "expired" + googleUuid.NewString()[:8] + "@example.com",
		PasswordHash: "hash",
		ExpiresAt:    time.Now().UTC().Add(-1 * time.Hour),
	}
	require.NoError(t, db.Create(expiredUser).Error)

	validUser := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
		ID:           googleUuid.New(),
		TenantID:     tenant.ID,
		Username:     "validuser" + googleUuid.NewString()[:8],
		Email:        "valid" + googleUuid.NewString()[:8] + "@example.com",
		PasswordHash: "hash",
		ExpiresAt:    time.Now().UTC().Add(72 * time.Hour),
	}
	require.NoError(t, db.Create(validUser).Error)

	// Create expired client.
	expiredClient := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
		ID:               googleUuid.New(),
		TenantID:         tenant.ID,
		ClientID:         "expiredclient" + googleUuid.NewString()[:8],
		ClientSecretHash: "secret",
		ExpiresAt:        time.Now().UTC().Add(-1 * time.Hour),
	}
	require.NoError(t, db.Create(expiredClient).Error)

	// Run cleanup.
	err := svc.CleanupExpiredRegistrations(ctx)
	require.NoError(t, err)

	// Verify expired records were deleted.
	var expiredUserCount int64

	db.Model(&cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{}).Where("id = ?", expiredUser.ID).Count(&expiredUserCount)
	require.Equal(t, int64(0), expiredUserCount)

	var expiredClientCount int64

	db.Model(&cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{}).Where("id = ?", expiredClient.ID).Count(&expiredClientCount)
	require.Equal(t, int64(0), expiredClientCount)

	// Verify valid user still exists.
	var validUserCount int64

	db.Model(&cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{}).Where("id = ?", validUser.ID).Count(&validUserCount)
	require.Equal(t, int64(1), validUserCount)
}

// TestVerificationService_ApproveUser_RoleFromWrongTenant tests approval with role from different tenant.
func TestVerificationService_ApproveUser_RoleFromWrongTenant(t *testing.T) {
	t.Parallel()

	svc, db := setupVerificationService(t)
	ctx := context.Background()

	// Create two tenants with their roles.
	tenant1, _ := createTestTenantAndRole(t, db, "tenant1-"+googleUuid.NewString()[:8])
	_, role2 := createTestTenantAndRole(t, db, "tenant2-"+googleUuid.NewString()[:8])

	// Create unverified user in tenant1.
	unverifiedUser := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
		ID:           googleUuid.New(),
		TenantID:     tenant1.ID,
		Username:     "roletenantuser" + googleUuid.NewString()[:8],
		Email:        "roleuser" + googleUuid.NewString()[:8] + "@example.com",
		PasswordHash: "hashedpassword",
		ExpiresAt:    time.Now().UTC().Add(72 * time.Hour),
	}
	require.NoError(t, db.Create(unverifiedUser).Error)

	// Try to approve with role from tenant2 - should fail.
	_, err := svc.ApproveUser(ctx, tenant1.ID, unverifiedUser.ID, []googleUuid.UUID{role2.ID})
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong to the specified tenant")
}

// TestVerificationService_ApproveClient_RoleFromWrongTenant tests client approval with role from different tenant.
func TestVerificationService_ApproveClient_RoleFromWrongTenant(t *testing.T) {
	t.Parallel()

	svc, db := setupVerificationService(t)
	ctx := context.Background()

	// Create two tenants with their roles.
	tenant1, _ := createTestTenantAndRole(t, db, "tenant1-"+googleUuid.NewString()[:8])
	_, role2 := createTestTenantAndRole(t, db, "tenant2-"+googleUuid.NewString()[:8])

	// Create unverified client in tenant1.
	unverifiedClient := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
		ID:               googleUuid.New(),
		TenantID:         tenant1.ID,
		ClientID:         "roleclient" + googleUuid.NewString()[:8],
		ClientSecretHash: "hashedsecret",
		ExpiresAt:        time.Now().UTC().Add(72 * time.Hour),
	}
	require.NoError(t, db.Create(unverifiedClient).Error)

	// Try to approve with role from tenant2 - should fail.
	_, err := svc.ApproveClient(ctx, tenant1.ID, unverifiedClient.ID, []googleUuid.UUID{role2.ID})
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong to the specified tenant")
}
