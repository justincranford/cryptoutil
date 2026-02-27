// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

//go:build !integration

// Package service tests ApproveUser and ApproveClient error paths.
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// --- Minimal mock repos for approve error tests ---

// mockRoleRepoForApprove implements RoleRepository for approve error testing.
type mockRoleRepoForApprove struct {
	getByIDErr error
	role       *cryptoutilAppsTemplateServiceServerRepository.Role
}

func (m *mockRoleRepoForApprove) Create(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.Role) error {
	return nil
}

func (m *mockRoleRepoForApprove) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}

	return m.role, nil
}

func (m *mockRoleRepoForApprove) GetByName(_ context.Context, _ googleUuid.UUID, _ string) (*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
	return nil, nil
}

func (m *mockRoleRepoForApprove) ListByTenant(_ context.Context, _ googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
	return nil, nil
}

func (m *mockRoleRepoForApprove) Delete(_ context.Context, _ googleUuid.UUID) error {
	return nil
}

// mockUserRepoForApprove implements UserRepository for approve error testing.
type mockUserRepoForApprove struct {
	createErr error
}

func (m *mockUserRepoForApprove) Create(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.User) error {
	return m.createErr
}

func (m *mockUserRepoForApprove) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.User, error) {
	return nil, nil
}

func (m *mockUserRepoForApprove) GetByUsername(_ context.Context, _ string) (*cryptoutilAppsTemplateServiceServerRepository.User, error) {
	return nil, nil
}

func (m *mockUserRepoForApprove) GetByEmail(_ context.Context, _ string) (*cryptoutilAppsTemplateServiceServerRepository.User, error) {
	return nil, nil
}

func (m *mockUserRepoForApprove) ListByTenant(_ context.Context, _ googleUuid.UUID, _ bool) ([]*cryptoutilAppsTemplateServiceServerRepository.User, error) {
	return nil, nil
}

func (m *mockUserRepoForApprove) Update(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.User) error {
	return nil
}

func (m *mockUserRepoForApprove) Delete(_ context.Context, _ googleUuid.UUID) error {
	return nil
}

// mockUserRoleRepoForApprove implements UserRoleRepository for approve error testing.
type mockUserRoleRepoForApprove struct {
	assignErr error
}

func (m *mockUserRoleRepoForApprove) Assign(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.UserRole) error {
	return m.assignErr
}

func (m *mockUserRoleRepoForApprove) Revoke(_ context.Context, _, _ googleUuid.UUID) error {
	return nil
}

func (m *mockUserRoleRepoForApprove) ListRolesByUser(_ context.Context, _ googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
	return nil, nil
}

func (m *mockUserRoleRepoForApprove) ListUsersByRole(_ context.Context, _ googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.User, error) {
	return nil, nil
}

// mockClientRepoForApprove implements ClientRepository for approve error testing.
type mockClientRepoForApprove struct {
	createErr error
}

func (m *mockClientRepoForApprove) Create(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.Client) error {
	return m.createErr
}

func (m *mockClientRepoForApprove) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Client, error) {
	return nil, nil
}

func (m *mockClientRepoForApprove) GetByClientID(_ context.Context, _ string) (*cryptoutilAppsTemplateServiceServerRepository.Client, error) {
	return nil, nil
}

func (m *mockClientRepoForApprove) ListByTenant(_ context.Context, _ googleUuid.UUID, _ bool) ([]*cryptoutilAppsTemplateServiceServerRepository.Client, error) {
	return nil, nil
}

func (m *mockClientRepoForApprove) Update(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.Client) error {
	return nil
}

func (m *mockClientRepoForApprove) Delete(_ context.Context, _ googleUuid.UUID) error {
	return nil
}

// mockClientRoleRepoForApprove implements ClientRoleRepository for approve error testing.
type mockClientRoleRepoForApprove struct {
	assignErr error
}

func (m *mockClientRoleRepoForApprove) Assign(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.ClientRole) error {
	return m.assignErr
}

func (m *mockClientRoleRepoForApprove) Revoke(_ context.Context, _, _ googleUuid.UUID) error {
	return nil
}

func (m *mockClientRoleRepoForApprove) ListRolesByClient(_ context.Context, _ googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
	return nil, nil
}

func (m *mockClientRoleRepoForApprove) ListClientsByRole(_ context.Context, _ googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.Client, error) {
	return nil, nil
}

// --- Helper to create a valid unverified user with matching tenant ---

func newApproveTestUnverifiedUser(tenantID googleUuid.UUID) *cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser {
	return &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
		ID:           googleUuid.New(),
		TenantID:     tenantID,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hash",
		ExpiresAt:    time.Now().UTC().Add(72 * time.Hour),
	}
}

// stableRoleID is a fixed role ID used across approve tests.
var stableRoleID = googleUuid.New()

// newApproveVerificationService creates a VerificationServiceImpl with given repo overrides.
func newApproveVerificationService(
	userRepo cryptoutilAppsTemplateServiceServerRepository.UserRepository,
	clientRepo cryptoutilAppsTemplateServiceServerRepository.ClientRepository,
	unverifiedUserRepo cryptoutilAppsTemplateServiceServerRepository.UnverifiedUserRepository,
	unverifiedClientRepo cryptoutilAppsTemplateServiceServerRepository.UnverifiedClientRepository,
	roleRepo cryptoutilAppsTemplateServiceServerRepository.RoleRepository,
	userRoleRepo cryptoutilAppsTemplateServiceServerRepository.UserRoleRepository,
	clientRoleRepo cryptoutilAppsTemplateServiceServerRepository.ClientRoleRepository,
) *VerificationServiceImpl {
	return &VerificationServiceImpl{
		userRepo:             userRepo,
		clientRepo:           clientRepo,
		unverifiedUserRepo:   unverifiedUserRepo,
		unverifiedClientRepo: unverifiedClientRepo,
		roleRepo:             roleRepo,
		userRoleRepo:         userRoleRepo,
		clientRoleRepo:       clientRoleRepo,
	}
}

// --- Tests for ApproveUser error paths ---

// TestApproveUser_GetByIDError covers verification_service.go:107-109.
func TestApproveUser_GetByIDError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	expectedErr := errors.New("getbyid error for approve user")

	svc := newApproveVerificationService(
		&mockUserRepoForApprove{},
		&mockClientRepoForApprove{},
		&mockUnverifiedUserRepoWithErrors{getByIDErr: expectedErr},
		&mockUnverifiedClientRepoWithErrors{},
		&mockRoleRepoForApprove{},
		&mockUserRoleRepoForApprove{},
		&mockClientRoleRepoForApprove{},
	)

	_, err := svc.ApproveUser(context.Background(), tenantID, googleUuid.New(), []googleUuid.UUID{stableRoleID})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get unverified user")
}

// TestApproveUser_RoleGetByIDError covers verification_service.go:129-131.
func TestApproveUser_RoleGetByIDError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	unverifiedUser := newApproveTestUnverifiedUser(tenantID)
	expectedErr := errors.New("role not found")

	svc := newApproveVerificationService(
		&mockUserRepoForApprove{},
		&mockClientRepoForApprove{},
		&mockUnverifiedUserRepoWithErrors{user: unverifiedUser},
		&mockUnverifiedClientRepoWithErrors{},
		&mockRoleRepoForApprove{getByIDErr: expectedErr},
		&mockUserRoleRepoForApprove{},
		&mockClientRoleRepoForApprove{},
	)

	_, err := svc.ApproveUser(context.Background(), tenantID, unverifiedUser.ID, []googleUuid.UUID{stableRoleID})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get role")
}

// TestApproveUser_UserCreateError covers verification_service.go:148-150.
func TestApproveUser_UserCreateError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	unverifiedUser := newApproveTestUnverifiedUser(tenantID)
	roleID := googleUuid.New()
	expectedErr := errors.New("user create db error")

	svc := newApproveVerificationService(
		&mockUserRepoForApprove{createErr: expectedErr},
		&mockClientRepoForApprove{},
		&mockUnverifiedUserRepoWithErrors{user: unverifiedUser},
		&mockUnverifiedClientRepoWithErrors{},
		&mockRoleRepoForApprove{role: &cryptoutilAppsTemplateServiceServerRepository.Role{ID: roleID, TenantID: tenantID}},
		&mockUserRoleRepoForApprove{},
		&mockClientRoleRepoForApprove{},
	)

	_, err := svc.ApproveUser(context.Background(), tenantID, unverifiedUser.ID, []googleUuid.UUID{roleID})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create verified user")
}

// TestApproveUser_UserRoleAssignError covers verification_service.go:159-161.
func TestApproveUser_UserRoleAssignError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	unverifiedUser := newApproveTestUnverifiedUser(tenantID)
	roleID := googleUuid.New()
	expectedErr := errors.New("role assign error")

	svc := newApproveVerificationService(
		&mockUserRepoForApprove{},
		&mockClientRepoForApprove{},
		&mockUnverifiedUserRepoWithErrors{user: unverifiedUser},
		&mockUnverifiedClientRepoWithErrors{},
		&mockRoleRepoForApprove{role: &cryptoutilAppsTemplateServiceServerRepository.Role{ID: roleID, TenantID: tenantID}},
		&mockUserRoleRepoForApprove{assignErr: expectedErr},
		&mockClientRoleRepoForApprove{},
	)

	_, err := svc.ApproveUser(context.Background(), tenantID, unverifiedUser.ID, []googleUuid.UUID{roleID})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to assign role")
}

// TestApproveUser_DeleteUnverifiedError covers verification_service.go:165-169.
// Note: In this special case, the function returns (user, error) not (nil, error).
func TestApproveUser_DeleteUnverifiedError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	unverifiedUser := newApproveTestUnverifiedUser(tenantID)
	roleID := googleUuid.New()
	expectedErr := errors.New("delete unverified user error")

	svc := newApproveVerificationService(
		&mockUserRepoForApprove{},
		&mockClientRepoForApprove{},
		&mockUnverifiedUserRepoWithErrors{user: unverifiedUser, deleteErr: expectedErr},
		&mockUnverifiedClientRepoWithErrors{},
		&mockRoleRepoForApprove{role: &cryptoutilAppsTemplateServiceServerRepository.Role{ID: roleID, TenantID: tenantID}},
		&mockUserRoleRepoForApprove{},
		&mockClientRoleRepoForApprove{},
	)

	result, err := svc.ApproveUser(context.Background(), tenantID, unverifiedUser.ID, []googleUuid.UUID{roleID})
	// approve succeeds (returns user) but also returns an error about failed delete
	require.NotNil(t, result)
	require.Error(t, err)
	require.Contains(t, err.Error(), "user approved but failed to delete unverified record")
}

// TestApproveClient_DeleteUnverifiedError covers verification_service.go:234-238.
func TestApproveClient_DeleteUnverifiedError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	roleID := googleUuid.New()
	expectedDeleteErr := errors.New("delete unverified client error")

	unverifiedClient := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
		ID:               googleUuid.New(),
		TenantID:         tenantID,
		ClientID:         "testclient",
		ClientSecretHash: "hash",
		ExpiresAt:        time.Now().UTC().Add(72 * time.Hour),
	}

	svc := newApproveVerificationService(
		&mockUserRepoForApprove{},
		&mockClientRepoForApprove{},
		&mockUnverifiedUserRepoWithErrors{},
		&mockUnverifiedClientRepoWithErrors{client: unverifiedClient, deleteErr: expectedDeleteErr},
		&mockRoleRepoForApprove{role: &cryptoutilAppsTemplateServiceServerRepository.Role{ID: roleID, TenantID: tenantID}},
		&mockUserRoleRepoForApprove{},
		&mockClientRoleRepoForApprove{},
	)

	result, err := svc.ApproveClient(context.Background(), tenantID, unverifiedClient.ID, []googleUuid.UUID{roleID})
	// approve succeeds (returns client) but also returns an error about failed delete
	require.NotNil(t, result)
	require.Error(t, err)
	require.Contains(t, err.Error(), "client approved but failed to delete unverified record")
}
