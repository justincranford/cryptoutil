// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

//go:build !integration

package service

import (
	"context"
	"errors"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// mockUnverifiedUserRepo with error injection.
type mockUnverifiedUserRepoWithErrors struct {
	getByIDErr       error
	listByTenantErr  error
	deleteErr        error
	deleteExpiredErr error
	user             *cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser
}

func (m *mockUnverifiedUserRepoWithErrors) Create(ctx context.Context, user *cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser) error {
	return nil
}

func (m *mockUnverifiedUserRepoWithErrors) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}

	return m.user, nil
}

func (m *mockUnverifiedUserRepoWithErrors) GetByUsername(ctx context.Context, username string) (*cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser, error) {
	return nil, nil
}

func (m *mockUnverifiedUserRepoWithErrors) ListByTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser, error) {
	return nil, m.listByTenantErr
}

func (m *mockUnverifiedUserRepoWithErrors) Delete(ctx context.Context, id googleUuid.UUID) error {
	return m.deleteErr
}

func (m *mockUnverifiedUserRepoWithErrors) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, m.deleteExpiredErr
}

// mockUnverifiedClientRepo with error injection.
type mockUnverifiedClientRepoWithErrors struct {
	getByIDErr       error
	listByTenantErr  error
	deleteErr        error
	deleteExpiredErr error
	client           *cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient
}

func (m *mockUnverifiedClientRepoWithErrors) Create(ctx context.Context, client *cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient) error {
	return nil
}

func (m *mockUnverifiedClientRepoWithErrors) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}

	return m.client, nil
}

func (m *mockUnverifiedClientRepoWithErrors) GetByClientID(ctx context.Context, clientID string) (*cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient, error) {
	return nil, nil
}

func (m *mockUnverifiedClientRepoWithErrors) ListByTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient, error) {
	return nil, m.listByTenantErr
}

func (m *mockUnverifiedClientRepoWithErrors) Delete(ctx context.Context, id googleUuid.UUID) error {
	return m.deleteErr
}

func (m *mockUnverifiedClientRepoWithErrors) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, m.deleteExpiredErr
}

// TestRejectUser_GetByIDError tests RejectUser when GetByID fails.
// Targets verification_service.go:249 (GetByID error return).
func TestRejectUser_GetByIDError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	unverifiedUserID := googleUuid.New()
	expectedErr := errors.New("database error")

	mockRepo := &mockUnverifiedUserRepoWithErrors{
		getByIDErr: expectedErr,
	}

	svc := &VerificationServiceImpl{
		unverifiedUserRepo: mockRepo,
	}

	err := svc.RejectUser(context.Background(), tenantID, unverifiedUserID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get unverified user")
}

// TestRejectUser_WrongTenant tests RejectUser when user belongs to different tenant.
// Targets verification_service.go:253-255 (tenant mismatch check).
func TestRejectUser_WrongTenant(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	differentTenantID := googleUuid.New()
	unverifiedUserID := googleUuid.New()

	mockRepo := &mockUnverifiedUserRepoWithErrors{
		user: &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
			ID:       unverifiedUserID,
			TenantID: differentTenantID, // Different tenant!
		},
	}

	svc := &VerificationServiceImpl{
		unverifiedUserRepo: mockRepo,
	}

	err := svc.RejectUser(context.Background(), tenantID, unverifiedUserID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong to the specified tenant")
}

// TestRejectUser_DeleteError tests RejectUser when Delete fails.
// Targets verification_service.go:259 (Delete error return).
func TestRejectUser_DeleteError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	unverifiedUserID := googleUuid.New()
	expectedErr := errors.New("delete failed")

	mockRepo := &mockUnverifiedUserRepoWithErrors{
		user: &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
			ID:       unverifiedUserID,
			TenantID: tenantID, // Correct tenant
		},
		deleteErr: expectedErr,
	}

	svc := &VerificationServiceImpl{
		unverifiedUserRepo: mockRepo,
	}

	err := svc.RejectUser(context.Background(), tenantID, unverifiedUserID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to reject user registration")
}

// TestRejectClient_GetByIDError tests RejectClient when GetByID fails.
// Targets verification_service.go:270 (GetByID error return).
func TestRejectClient_GetByIDError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	unverifiedClientID := googleUuid.New()
	expectedErr := errors.New("database error")

	mockRepo := &mockUnverifiedClientRepoWithErrors{
		getByIDErr: expectedErr,
	}

	svc := &VerificationServiceImpl{
		unverifiedClientRepo: mockRepo,
	}

	err := svc.RejectClient(context.Background(), tenantID, unverifiedClientID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get unverified client")
}

// TestRejectClient_WrongTenant tests RejectClient when client belongs to different tenant.
// Targets verification_service.go:274-276 (tenant mismatch check).
func TestRejectClient_WrongTenant(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	differentTenantID := googleUuid.New()
	unverifiedClientID := googleUuid.New()

	mockRepo := &mockUnverifiedClientRepoWithErrors{
		client: &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
			ID:       unverifiedClientID,
			TenantID: differentTenantID, // Different tenant!
		},
	}

	svc := &VerificationServiceImpl{
		unverifiedClientRepo: mockRepo,
	}

	err := svc.RejectClient(context.Background(), tenantID, unverifiedClientID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong to the specified tenant")
}

// TestRejectClient_DeleteError tests RejectClient when Delete fails.
// Targets verification_service.go:280 (Delete error return).
func TestRejectClient_DeleteError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	unverifiedClientID := googleUuid.New()
	expectedErr := errors.New("delete failed")

	mockRepo := &mockUnverifiedClientRepoWithErrors{
		client: &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
			ID:       unverifiedClientID,
			TenantID: tenantID, // Correct tenant
		},
		deleteErr: expectedErr,
	}

	svc := &VerificationServiceImpl{
		unverifiedClientRepo: mockRepo,
	}

	err := svc.RejectClient(context.Background(), tenantID, unverifiedClientID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to reject client registration")
}

// TestCleanupExpiredRegistrations_UserDeleteExpiredError tests cleanup when user DeleteExpired fails.
// Targets verification_service.go:289 (user DeleteExpired error return).
func TestCleanupExpiredRegistrations_UserDeleteExpiredError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("delete expired users failed")

	mockUserRepo := &mockUnverifiedUserRepoWithErrors{
		deleteExpiredErr: expectedErr,
	}

	svc := &VerificationServiceImpl{
		unverifiedUserRepo: mockUserRepo,
	}

	err := svc.CleanupExpiredRegistrations(context.Background())

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to cleanup expired user registrations")
}

// TestCleanupExpiredRegistrations_ClientDeleteExpiredError tests cleanup when client DeleteExpired fails.
// Targets verification_service.go:294 (client DeleteExpired error return).
func TestCleanupExpiredRegistrations_ClientDeleteExpiredError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("delete expired clients failed")

	mockUserRepo := &mockUnverifiedUserRepoWithErrors{
		// User deletion succeeds
	}
	mockClientRepo := &mockUnverifiedClientRepoWithErrors{
		deleteExpiredErr: expectedErr,
	}

	svc := &VerificationServiceImpl{
		unverifiedUserRepo:   mockUserRepo,
		unverifiedClientRepo: mockClientRepo,
	}

	err := svc.CleanupExpiredRegistrations(context.Background())

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to cleanup expired client registrations")
}

// TestListPendingUsers_RepoError tests ListPendingUsers when the repo returns an error.
// Covers verification_service.go:86-88 (ListByTenant error return in ListPendingUsers).
func TestListPendingUsers_RepoError(t *testing.T) {
t.Parallel()

tenantID := googleUuid.New()
expectedErr := errors.New("list pending users db error")

mockRepo := &mockUnverifiedUserRepoWithErrors{
listByTenantErr: expectedErr,
}

svc := &VerificationServiceImpl{
unverifiedUserRepo: mockRepo,
}

_, err := svc.ListPendingUsers(context.Background(), tenantID)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to list pending users")
}

// TestListPendingClients_RepoError tests ListPendingClients when the repo returns an error.
// Covers verification_service.go:96-98 (ListByTenant error return in ListPendingClients).
func TestListPendingClients_RepoError(t *testing.T) {
t.Parallel()

tenantID := googleUuid.New()
expectedErr := errors.New("list pending clients db error")

mockClientRepo := &mockUnverifiedClientRepoWithErrors{
listByTenantErr: expectedErr,
}

svc := &VerificationServiceImpl{
unverifiedClientRepo: mockClientRepo,
}

_, err := svc.ListPendingClients(context.Background(), tenantID)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to list pending clients")
}
