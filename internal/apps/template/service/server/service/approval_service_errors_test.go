//go:build !integration

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

// Mock repositories for testing ApproveClient error paths.
type mockUnverifiedClientRepoForApproval struct {
	unverifiedClient *cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient
	getByIDErr       error
	deleteErr        error
}

// UnverifiedClientRepository methods.
func (m *mockUnverifiedClientRepoForApproval) Create(_ context.Context, client *cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient) error {
	return nil
}

func (m *mockUnverifiedClientRepoForApproval) GetByID(_ context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}

	return m.unverifiedClient, nil
}

func (m *mockUnverifiedClientRepoForApproval) GetByClientID(_ context.Context, clientID string) (*cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient, error) {
	return nil, nil
}

func (m *mockUnverifiedClientRepoForApproval) ListByTenant(_ context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient, error) {
	return nil, nil
}

func (m *mockUnverifiedClientRepoForApproval) Delete(_ context.Context, id googleUuid.UUID) error {
	return m.deleteErr
}

func (m *mockUnverifiedClientRepoForApproval) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

// Mock role repository.
type mockRoleRepoForApproval struct {
	role       *cryptoutilAppsTemplateServiceServerRepository.Role
	getByIDErr error
}

// RoleRepository methods.
func (m *mockRoleRepoForApproval) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}

	return m.role, nil
}

func (m *mockRoleRepoForApproval) GetByName(ctx context.Context, tenantID googleUuid.UUID, name string) (*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
	return nil, nil
}

func (m *mockRoleRepoForApproval) ListByTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
	return nil, nil
}

func (m *mockRoleRepoForApproval) Create(ctx context.Context, role *cryptoutilAppsTemplateServiceServerRepository.Role) error {
	return nil
}

func (m *mockRoleRepoForApproval) Update(ctx context.Context, role *cryptoutilAppsTemplateServiceServerRepository.Role) error {
	return nil
}

func (m *mockRoleRepoForApproval) Delete(ctx context.Context, id googleUuid.UUID) error {
	return nil
}

// ClientRepository methods.
type mockClientRepo struct {
	createErr error
}

func (m *mockClientRepo) Create(ctx context.Context, client *cryptoutilAppsTemplateServiceServerRepository.Client) error {
	return m.createErr
}

func (m *mockClientRepo) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Client, error) {
	return nil, nil
}

func (m *mockClientRepo) GetByClientID(ctx context.Context, clientID string) (*cryptoutilAppsTemplateServiceServerRepository.Client, error) {
	return nil, nil
}

func (m *mockClientRepo) ListByTenant(ctx context.Context, tenantID googleUuid.UUID, activeOnly bool) ([]*cryptoutilAppsTemplateServiceServerRepository.Client, error) {
	return nil, nil
}

func (m *mockClientRepo) Update(ctx context.Context, client *cryptoutilAppsTemplateServiceServerRepository.Client) error {
	return nil
}

func (m *mockClientRepo) Delete(ctx context.Context, id googleUuid.UUID) error {
	return nil
}

// ClientRoleRepository methods.
type mockClientRoleRepo struct {
	assignErr error
}

func (m *mockClientRoleRepo) Assign(ctx context.Context, clientRole *cryptoutilAppsTemplateServiceServerRepository.ClientRole) error {
	return m.assignErr
}

func (m *mockClientRoleRepo) Revoke(ctx context.Context, clientID, roleID googleUuid.UUID) error {
	return nil
}

func (m *mockClientRoleRepo) ListRolesByClient(ctx context.Context, clientID googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
	return nil, nil
}

func (m *mockClientRoleRepo) ListClientsByRole(ctx context.Context, roleID googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.Client, error) {
	return nil, nil
}

// TestApproveClient_GetByIDError tests ApproveClient when GetByID fails.
// Targets verification_service.go:178-180 (GetByID error return).
func TestApproveClient_GetByIDError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	unverifiedClientID := googleUuid.New()
	roleIDs := []googleUuid.UUID{googleUuid.New()}
	expectedErr := errors.New("database error")

	mockUnverifiedRepo := &mockUnverifiedClientRepoForApproval{
		getByIDErr: expectedErr,
	}

	svc := &VerificationServiceImpl{
		unverifiedClientRepo: mockUnverifiedRepo,
	}

	client, err := svc.ApproveClient(context.Background(), tenantID, unverifiedClientID, roleIDs)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get unverified client")
	require.Nil(t, client)
}

// TestApproveClient_WrongTenant tests ApproveClient when tenant mismatch.
// Targets verification_service.go:183-185 (tenant check).
func TestApproveClient_WrongTenant(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	differentTenantID := googleUuid.New()
	unverifiedClientID := googleUuid.New()
	roleIDs := []googleUuid.UUID{googleUuid.New()}

	mockUnverifiedRepo := &mockUnverifiedClientRepoForApproval{
		unverifiedClient: &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
			ID:       unverifiedClientID,
			TenantID: differentTenantID, // Wrong tenant!
		},
	}

	svc := &VerificationServiceImpl{
		unverifiedClientRepo: mockUnverifiedRepo,
	}

	client, err := svc.ApproveClient(context.Background(), tenantID, unverifiedClientID, roleIDs)

	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong to the specified tenant")
	require.Nil(t, client)
}

// TestApproveClient_Expired tests ApproveClient when registration expired.
// Targets verification_service.go:188-190 (expiration check).
func TestApproveClient_Expired(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	unverifiedClientID := googleUuid.New()
	roleIDs := []googleUuid.UUID{googleUuid.New()}

	mockUnverifiedRepo := &mockUnverifiedClientRepoForApproval{
		unverifiedClient: &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
			ID:        unverifiedClientID,
			TenantID:  tenantID,
			ExpiresAt: time.Now().UTC().Add(-1 * time.Hour), // Expired 1 hour ago!
		},
	}

	svc := &VerificationServiceImpl{
		unverifiedClientRepo: mockUnverifiedRepo,
	}

	client, err := svc.ApproveClient(context.Background(), tenantID, unverifiedClientID, roleIDs)

	require.Error(t, err)
	require.Contains(t, err.Error(), "client registration has expired")
	require.Nil(t, client)
}

// TestApproveClient_NoRoles tests ApproveClient when no roles provided.
// Targets verification_service.go:193-195 (empty role list check).
func TestApproveClient_NoRoles(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	unverifiedClientID := googleUuid.New()
	roleIDs := []googleUuid.UUID{} // Empty!

	mockUnverifiedRepo := &mockUnverifiedClientRepoForApproval{
		unverifiedClient: &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
			ID:        unverifiedClientID,
			TenantID:  tenantID,
			ExpiresAt: time.Now().UTC().Add(1 * time.Hour), // Not expired
		},
	}

	svc := &VerificationServiceImpl{
		unverifiedClientRepo: mockUnverifiedRepo,
	}

	client, err := svc.ApproveClient(context.Background(), tenantID, unverifiedClientID, roleIDs)

	require.Error(t, err)
	require.Contains(t, err.Error(), "at least one role must be assigned")
	require.Nil(t, client)
}

// TestApproveClient_RoleGetByIDError tests ApproveClient when role GetByID fails.
// Targets verification_service.go:199-201 (role GetByID error).
func TestApproveClient_RoleGetByIDError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	unverifiedClientID := googleUuid.New()
	roleID := googleUuid.New()
	roleIDs := []googleUuid.UUID{roleID}
	expectedErr := errors.New("role database error")

	mockUnverifiedRepo := &mockUnverifiedClientRepoForApproval{
		unverifiedClient: &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
			ID:        unverifiedClientID,
			TenantID:  tenantID,
			ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
		},
	}

	mockRoleRepo := &mockRoleRepoForApproval{
		getByIDErr: expectedErr,
	}

	svc := &VerificationServiceImpl{
		unverifiedClientRepo: mockUnverifiedRepo,
		roleRepo:             mockRoleRepo,
	}

	client, err := svc.ApproveClient(context.Background(), tenantID, unverifiedClientID, roleIDs)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get role")
	require.Nil(t, client)
}

// TestApproveClient_RoleWrongTenant tests ApproveClient when role belongs to different tenant.
// Targets verification_service.go:203-205 (role tenant check).
func TestApproveClient_RoleWrongTenant(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	differentTenantID := googleUuid.New()
	unverifiedClientID := googleUuid.New()
	roleID := googleUuid.New()
	roleIDs := []googleUuid.UUID{roleID}

	mockUnverifiedRepo := &mockUnverifiedClientRepoForApproval{
		unverifiedClient: &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
			ID:        unverifiedClientID,
			TenantID:  tenantID,
			ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
		},
	}

	mockRoleRepo := &mockRoleRepoForApproval{
		role: &cryptoutilAppsTemplateServiceServerRepository.Role{
			ID:       roleID,
			TenantID: differentTenantID, // Different tenant!
		},
	}

	svc := &VerificationServiceImpl{
		unverifiedClientRepo: mockUnverifiedRepo,
		roleRepo:             mockRoleRepo,
	}

	client, err := svc.ApproveClient(context.Background(), tenantID, unverifiedClientID, roleIDs)

	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong to the specified tenant")
	require.Nil(t, client)
}

// TestApproveClient_CreateClientError tests ApproveClient when Client Create fails.
// Targets verification_service.go:215-217 (Client creation error).
func TestApproveClient_CreateClientError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	unverifiedClientID := googleUuid.New()
	roleID := googleUuid.New()
	roleIDs := []googleUuid.UUID{roleID}
	expectedErr := errors.New("client create failed")

	mockUnverifiedRepo := &mockUnverifiedClientRepoForApproval{
		unverifiedClient: &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
			ID:        unverifiedClientID,
			TenantID:  tenantID,
			ClientID:  "test-client",
			ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
		},
	}

	mockRoleRepo := &mockRoleRepoForApproval{
		role: &cryptoutilAppsTemplateServiceServerRepository.Role{
			ID:       roleID,
			TenantID: tenantID,
		},
	}

	mockClientRepo := &mockClientRepo{
		createErr: expectedErr,
	}

	svc := &VerificationServiceImpl{
		unverifiedClientRepo: mockUnverifiedRepo,
		roleRepo:             mockRoleRepo,
		clientRepo:           mockClientRepo,
	}

	client, err := svc.ApproveClient(context.Background(), tenantID, unverifiedClientID, roleIDs)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create verified client")
	require.Nil(t, client)
}

// TestApproveClient_AssignRoleError tests ApproveClient when ClientRole Assign fails.
// Targets verification_service.go:223-225 (role assignment error).
func TestApproveClient_AssignRoleError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	unverifiedClientID := googleUuid.New()
	roleID := googleUuid.New()
	roleIDs := []googleUuid.UUID{roleID}
	expectedErr := errors.New("role assign failed")

	mockUnverifiedRepo := &mockUnverifiedClientRepoForApproval{
		unverifiedClient: &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
			ID:        unverifiedClientID,
			TenantID:  tenantID,
			ClientID:  "test-client",
			ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
		},
	}

	mockRoleRepo := &mockRoleRepoForApproval{
		role: &cryptoutilAppsTemplateServiceServerRepository.Role{
			ID:       roleID,
			TenantID: tenantID,
		},
	}

	mockClientRepo := &mockClientRepo{}

	mockClientRoleRepo := &mockClientRoleRepo{
		assignErr: expectedErr,
	}

	svc := &VerificationServiceImpl{
		unverifiedClientRepo: mockUnverifiedRepo,
		roleRepo:             mockRoleRepo,
		clientRepo:           mockClientRepo,
		clientRoleRepo:       mockClientRoleRepo,
	}

	client, err := svc.ApproveClient(context.Background(), tenantID, unverifiedClientID, roleIDs)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to assign role")
	require.Nil(t, client)
}
