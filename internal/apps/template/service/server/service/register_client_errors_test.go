//go:build !integration

package service

import (
	"context"
	"errors"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// mockTenantServiceForRegisterClient implements TenantService for RegisterClient testing.
type mockTenantServiceForRegisterClient struct {
	tenant          *cryptoutilTemplateRepository.Tenant
	createTenantErr error
}

func (m *mockTenantServiceForRegisterClient) CreateTenant(ctx context.Context, name, description string) (*cryptoutilTemplateRepository.Tenant, error) {
	if m.createTenantErr != nil {
		return nil, m.createTenantErr
	}

	return m.tenant, nil
}

func (m *mockTenantServiceForRegisterClient) GetTenant(ctx context.Context, id googleUuid.UUID) (*cryptoutilTemplateRepository.Tenant, error) {
	return nil, nil
}

func (m *mockTenantServiceForRegisterClient) GetTenantByName(ctx context.Context, name string) (*cryptoutilTemplateRepository.Tenant, error) {
	return nil, nil
}

func (m *mockTenantServiceForRegisterClient) ListTenants(ctx context.Context, activeOnly bool) ([]*cryptoutilTemplateRepository.Tenant, error) {
	return nil, nil
}

func (m *mockTenantServiceForRegisterClient) UpdateTenant(ctx context.Context, id googleUuid.UUID, name, description *string, active *bool) (*cryptoutilTemplateRepository.Tenant, error) {
	return nil, nil
}

func (m *mockTenantServiceForRegisterClient) DeleteTenant(ctx context.Context, id googleUuid.UUID) error {
	return nil
}

// mockClientRepoForRegisterClient implements repository.ClientRepository for RegisterClient testing.
type mockClientRepoForRegisterClient struct {
	createErr error
}

func (m *mockClientRepoForRegisterClient) Create(ctx context.Context, client *cryptoutilTemplateRepository.Client) error {
	return m.createErr
}

func (m *mockClientRepoForRegisterClient) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilTemplateRepository.Client, error) {
	return nil, nil
}

func (m *mockClientRepoForRegisterClient) GetByClientID(ctx context.Context, clientID string) (*cryptoutilTemplateRepository.Client, error) {
	return nil, nil
}

func (m *mockClientRepoForRegisterClient) ListByTenant(ctx context.Context, tenantID googleUuid.UUID, activeOnly bool) ([]*cryptoutilTemplateRepository.Client, error) {
	return nil, nil
}

func (m *mockClientRepoForRegisterClient) Update(ctx context.Context, client *cryptoutilTemplateRepository.Client) error {
	return nil
}

func (m *mockClientRepoForRegisterClient) Delete(ctx context.Context, id googleUuid.UUID) error {
	return nil
}

// mockRoleRepoForRegisterClient implements repository.RoleRepository for RegisterClient testing.
type mockRoleRepoForRegisterClient struct {
	role         *cryptoutilTemplateRepository.Role
	getByNameErr error
}

func (m *mockRoleRepoForRegisterClient) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilTemplateRepository.Role, error) {
	return nil, nil
}

func (m *mockRoleRepoForRegisterClient) GetByName(ctx context.Context, tenantID googleUuid.UUID, name string) (*cryptoutilTemplateRepository.Role, error) {
	if m.getByNameErr != nil {
		return nil, m.getByNameErr
	}

	return m.role, nil
}

func (m *mockRoleRepoForRegisterClient) ListByTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilTemplateRepository.Role, error) {
	return nil, nil
}

func (m *mockRoleRepoForRegisterClient) Create(ctx context.Context, role *cryptoutilTemplateRepository.Role) error {
	return nil
}

func (m *mockRoleRepoForRegisterClient) Update(ctx context.Context, role *cryptoutilTemplateRepository.Role) error {
	return nil
}

func (m *mockRoleRepoForRegisterClient) Delete(ctx context.Context, id googleUuid.UUID) error {
	return nil
}

// mockClientRoleRepoForRegisterClient implements repository.ClientRoleRepository for RegisterClient testing.
type mockClientRoleRepoForRegisterClient struct {
	assignErr error
}

func (m *mockClientRoleRepoForRegisterClient) Assign(ctx context.Context, clientRole *cryptoutilTemplateRepository.ClientRole) error {
	return m.assignErr
}

func (m *mockClientRoleRepoForRegisterClient) Revoke(ctx context.Context, clientID, roleID googleUuid.UUID) error {
	return nil
}

func (m *mockClientRoleRepoForRegisterClient) ListRolesByClient(ctx context.Context, clientID googleUuid.UUID) ([]*cryptoutilTemplateRepository.Role, error) {
	return nil, nil
}

func (m *mockClientRoleRepoForRegisterClient) ListClientsByRole(ctx context.Context, roleID googleUuid.UUID) ([]*cryptoutilTemplateRepository.Client, error) {
	return nil, nil
}

// mockUnverifiedClientRepoForRegisterClient implements repository.UnverifiedClientRepository for RegisterClient testing.
type mockUnverifiedClientRepoForRegisterClient struct {
	createErr error
}

func (m *mockUnverifiedClientRepoForRegisterClient) Create(ctx context.Context, client *cryptoutilTemplateRepository.UnverifiedClient) error {
	return m.createErr
}

func (m *mockUnverifiedClientRepoForRegisterClient) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilTemplateRepository.UnverifiedClient, error) {
	return nil, nil
}

func (m *mockUnverifiedClientRepoForRegisterClient) GetByClientID(ctx context.Context, clientID string) (*cryptoutilTemplateRepository.UnverifiedClient, error) {
	return nil, nil
}

func (m *mockUnverifiedClientRepoForRegisterClient) ListByTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilTemplateRepository.UnverifiedClient, error) {
	return nil, nil
}

func (m *mockUnverifiedClientRepoForRegisterClient) Delete(ctx context.Context, id googleUuid.UUID) error {
	return nil
}

func (m *mockUnverifiedClientRepoForRegisterClient) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

// TestRegisterClient_ValidationError tests RegisterClient when neither or both parameters provided.
// Target: registration_service.go:182-184 (validation error).
func TestRegisterClient_ValidationError(t *testing.T) {
	t.Parallel()

	svc := &RegistrationServiceImpl{}

	// Neither parameter provided.
	result, err := svc.RegisterClient(context.Background(), "client-id", "hash", nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "exactly one of newTenant or existingTenantID must be provided")
	require.Nil(t, result)

	// Both parameters provided.
	tenantID := googleUuid.New()
	newTenant := &NewTenantInfo{Name: "test", Description: "test"}
	result, err = svc.RegisterClient(context.Background(), "client-id", "hash", newTenant, &tenantID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "exactly one of newTenant or existingTenantID must be provided")
	require.Nil(t, result)
}

// TestRegisterClient_CreateTenantError tests RegisterClient when CreateTenant fails.
// Target: registration_service.go:190-192 (CreateTenant error return).
func TestRegisterClient_CreateTenantError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database error")

	mockTenantSvc := &mockTenantServiceForRegisterClient{
		createTenantErr: expectedErr,
	}

	svc := &RegistrationServiceImpl{
		tenantService: mockTenantSvc,
	}

	newTenant := &NewTenantInfo{Name: "test", Description: "test"}
	result, err := svc.RegisterClient(context.Background(), "client-id", "hash", newTenant, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create tenant")
	require.Nil(t, result)
}

// TestRegisterClient_CreateClientError tests RegisterClient when Client creation fails.
// Target: registration_service.go:202-204 (Client Create error return).
func TestRegisterClient_CreateClientError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database error")

	mockTenantSvc := &mockTenantServiceForRegisterClient{
		tenant: &cryptoutilTemplateRepository.Tenant{
			ID:   googleUuid.New(),
			Name: "test-tenant",
		},
	}

	mockClientRepo := &mockClientRepoForRegisterClient{
		createErr: expectedErr,
	}

	svc := &RegistrationServiceImpl{
		tenantService: mockTenantSvc,
		clientRepo:    mockClientRepo,
	}

	newTenant := &NewTenantInfo{Name: "test", Description: "test"}
	result, err := svc.RegisterClient(context.Background(), "client-id", "hash", newTenant, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create client")
	require.Nil(t, result)
}

// TestRegisterClient_GetAdminRoleError tests RegisterClient when GetByName (admin role) fails.
// Target: registration_service.go:207-209 (GetByName error return).
func TestRegisterClient_GetAdminRoleError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("role not found")

	mockTenantSvc := &mockTenantServiceForRegisterClient{
		tenant: &cryptoutilTemplateRepository.Tenant{
			ID:   googleUuid.New(),
			Name: "test-tenant",
		},
	}

	mockClientRepo := &mockClientRepoForRegisterClient{}

	mockRoleRepo := &mockRoleRepoForRegisterClient{
		getByNameErr: expectedErr,
	}

	svc := &RegistrationServiceImpl{
		tenantService: mockTenantSvc,
		clientRepo:    mockClientRepo,
		roleRepo:      mockRoleRepo,
	}

	newTenant := &NewTenantInfo{Name: "test", Description: "test"}
	result, err := svc.RegisterClient(context.Background(), "client-id", "hash", newTenant, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get admin role")
	require.Nil(t, result)
}

// TestRegisterClient_AssignRoleError tests RegisterClient when role assignment fails.
// Target: registration_service.go:216-218 (Assign error return).
func TestRegisterClient_AssignRoleError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("assignment error")

	mockTenantSvc := &mockTenantServiceForRegisterClient{
		tenant: &cryptoutilTemplateRepository.Tenant{
			ID:   googleUuid.New(),
			Name: "test-tenant",
		},
	}

	mockClientRepo := &mockClientRepoForRegisterClient{}

	mockRoleRepo := &mockRoleRepoForRegisterClient{
		role: &cryptoutilTemplateRepository.Role{
			ID:   googleUuid.New(),
			Name: "admin",
		},
	}

	mockClientRoleRepo := &mockClientRoleRepoForRegisterClient{
		assignErr: expectedErr,
	}

	svc := &RegistrationServiceImpl{
		tenantService:  mockTenantSvc,
		clientRepo:     mockClientRepo,
		roleRepo:       mockRoleRepo,
		clientRoleRepo: mockClientRoleRepo,
	}

	newTenant := &NewTenantInfo{Name: "test", Description: "test"}
	result, err := svc.RegisterClient(context.Background(), "client-id", "hash", newTenant, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to assign admin role")
	require.Nil(t, result)
}

// TestRegisterClient_UnverifiedCreateError tests RegisterClient when UnverifiedClient creation fails.
// Target: registration_service.go:239-241 (UnverifiedClient Create error return).
func TestRegisterClient_UnverifiedCreateError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database error")

	mockUnverifiedRepo := &mockUnverifiedClientRepoForRegisterClient{
		createErr: expectedErr,
	}

	svc := &RegistrationServiceImpl{
		unverifiedClientRepo: mockUnverifiedRepo,
	}

	tenantID := googleUuid.New()
	result, err := svc.RegisterClient(context.Background(), "client-id", "hash", nil, &tenantID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create unverified client")
	require.Nil(t, result)
}
