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
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

func TestVerificationService_ApproveUser_NoRoles(t *testing.T) {
	t.Parallel()

	svc, db := setupVerificationService(t)
	ctx := context.Background()

	tenant, _ := createTestTenantAndRole(t, db, "no-roles-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	// Create unverified user.
	unverifiedUser := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
		ID:           googleUuid.New(),
		TenantID:     tenant.ID,
		Username:     "norolesuser" + googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Email:        "noroles" + googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
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

	tenant, role := createTestTenantAndRole(t, db, "approve-client-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	// Create unverified client.
	unverifiedClient := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
		ID:               googleUuid.New(),
		TenantID:         tenant.ID,
		ClientID:         "pendingclient" + googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength],
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

	tenant, _ := createTestTenantAndRole(t, db, "reject-user-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	// Create unverified user.
	unverifiedUser := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
		ID:           googleUuid.New(),
		TenantID:     tenant.ID,
		Username:     "rejectuser" + googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Email:        "reject" + googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
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

	tenant, _ := createTestTenantAndRole(t, db, "reject-client-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	// Create unverified client.
	unverifiedClient := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
		ID:               googleUuid.New(),
		TenantID:         tenant.ID,
		ClientID:         "rejectclient" + googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength],
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

	tenant, _ := createTestTenantAndRole(t, db, "cleanup-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	// Create expired and non-expired users.
	expiredUser := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
		ID:           googleUuid.New(),
		TenantID:     tenant.ID,
		Username:     "expireduser" + googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Email:        "expired" + googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
		PasswordHash: "hash",
		ExpiresAt:    time.Now().UTC().Add(-1 * time.Hour),
	}
	require.NoError(t, db.Create(expiredUser).Error)

	validUser := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
		ID:           googleUuid.New(),
		TenantID:     tenant.ID,
		Username:     "validuser" + googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Email:        "valid" + googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
		PasswordHash: "hash",
		ExpiresAt:    time.Now().UTC().Add(72 * time.Hour),
	}
	require.NoError(t, db.Create(validUser).Error)

	// Create expired client.
	expiredClient := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
		ID:               googleUuid.New(),
		TenantID:         tenant.ID,
		ClientID:         "expiredclient" + googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength],
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
	tenant1, _ := createTestTenantAndRole(t, db, "tenant1-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])
	_, role2 := createTestTenantAndRole(t, db, "tenant2-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	// Create unverified user in tenant1.
	unverifiedUser := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
		ID:           googleUuid.New(),
		TenantID:     tenant1.ID,
		Username:     "roletenantuser" + googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength],
		Email:        "roleuser" + googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength] + "@example.com",
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
	tenant1, _ := createTestTenantAndRole(t, db, "tenant1-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])
	_, role2 := createTestTenantAndRole(t, db, "tenant2-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	// Create unverified client in tenant1.
	unverifiedClient := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
		ID:               googleUuid.New(),
		TenantID:         tenant1.ID,
		ClientID:         "roleclient" + googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength],
		ClientSecretHash: "hashedsecret",
		ExpiresAt:        time.Now().UTC().Add(72 * time.Hour),
	}
	require.NoError(t, db.Create(unverifiedClient).Error)

	// Try to approve with role from tenant2 - should fail.
	_, err := svc.ApproveClient(ctx, tenant1.ID, unverifiedClient.ID, []googleUuid.UUID{role2.ID})
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong to the specified tenant")
}
