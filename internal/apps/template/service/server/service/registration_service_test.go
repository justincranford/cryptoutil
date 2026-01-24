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
	"errors"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// Mock TenantService.
type mockTenantService struct {
	createTenantFn func(ctx context.Context, name, description string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error)
}

func (m *mockTenantService) CreateTenant(ctx context.Context, name, description string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
	if m.createTenantFn != nil {
		return m.createTenantFn(ctx, name, description)
	}

	return nil, errors.New("not implemented")
}

func (m *mockTenantService) GetTenant(_ context.Context, _ googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
	return nil, errors.New("not implemented")
}

func (m *mockTenantService) GetTenantByName(_ context.Context, _ string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
	return nil, errors.New("not implemented")
}

func (m *mockTenantService) ListTenants(_ context.Context, _ bool) ([]*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
	return nil, errors.New("not implemented")
}

func (m *mockTenantService) UpdateTenant(_ context.Context, _ googleUuid.UUID, _, _ *string, _ *bool) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
	return nil, errors.New("not implemented")
}

func (m *mockTenantService) DeleteTenant(_ context.Context, _ googleUuid.UUID) error {
	return errors.New("not implemented")
}

// Mock UserRepository.
type mockUserRepository struct {
	createFn func(ctx context.Context, user *cryptoutilAppsTemplateServiceServerRepository.User) error
}

func (m *mockUserRepository) Create(ctx context.Context, user *cryptoutilAppsTemplateServiceServerRepository.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}

	return nil
}

func (m *mockUserRepository) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserRepository) GetByUsername(_ context.Context, _ string) (*cryptoutilAppsTemplateServiceServerRepository.User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserRepository) GetByEmail(_ context.Context, _ string) (*cryptoutilAppsTemplateServiceServerRepository.User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserRepository) ListByTenant(_ context.Context, _ googleUuid.UUID, _ bool) ([]*cryptoutilAppsTemplateServiceServerRepository.User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserRepository) Update(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.User) error {
	return errors.New("not implemented")
}

func (m *mockUserRepository) Delete(_ context.Context, _ googleUuid.UUID) error {
	return errors.New("not implemented")
}

// Mock ClientRepository.
type mockClientRepository struct {
	createFn func(ctx context.Context, client *cryptoutilAppsTemplateServiceServerRepository.Client) error
}

func (m *mockClientRepository) Create(ctx context.Context, client *cryptoutilAppsTemplateServiceServerRepository.Client) error {
	if m.createFn != nil {
		return m.createFn(ctx, client)
	}

	return nil
}

func (m *mockClientRepository) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Client, error) {
	return nil, errors.New("not implemented")
}

func (m *mockClientRepository) GetByClientID(_ context.Context, _ string) (*cryptoutilAppsTemplateServiceServerRepository.Client, error) {
	return nil, errors.New("not implemented")
}

func (m *mockClientRepository) ListByTenant(_ context.Context, _ googleUuid.UUID, _ bool) ([]*cryptoutilAppsTemplateServiceServerRepository.Client, error) {
	return nil, errors.New("not implemented")
}

func (m *mockClientRepository) Update(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.Client) error {
	return errors.New("not implemented")
}

func (m *mockClientRepository) Delete(_ context.Context, _ googleUuid.UUID) error {
	return errors.New("not implemented")
}

// Mock UnverifiedUserRepository.
type mockUnverifiedUserRepository struct {
	createFn func(ctx context.Context, user *cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser) error
}

func (m *mockUnverifiedUserRepository) Create(ctx context.Context, user *cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}

	return nil
}

func (m *mockUnverifiedUserRepository) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUnverifiedUserRepository) GetByUsername(_ context.Context, _ string) (*cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUnverifiedUserRepository) ListByTenant(_ context.Context, _ googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUnverifiedUserRepository) Delete(_ context.Context, _ googleUuid.UUID) error {
	return errors.New("not implemented")
}

func (m *mockUnverifiedUserRepository) DeleteExpired(_ context.Context) (int64, error) {
	return 0, errors.New("not implemented")
}

// Mock UnverifiedClientRepository.
type mockUnverifiedClientRepository struct {
	createFn func(ctx context.Context, client *cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient) error
}

func (m *mockUnverifiedClientRepository) Create(ctx context.Context, client *cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient) error {
	if m.createFn != nil {
		return m.createFn(ctx, client)
	}

	return nil
}

func (m *mockUnverifiedClientRepository) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUnverifiedClientRepository) GetByClientID(_ context.Context, _ string) (*cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUnverifiedClientRepository) ListByTenant(_ context.Context, _ googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUnverifiedClientRepository) Delete(_ context.Context, _ googleUuid.UUID) error {
	return errors.New("not implemented")
}

func (m *mockUnverifiedClientRepository) DeleteExpired(_ context.Context) (int64, error) {
	return 0, errors.New("not implemented")
}

// Mock UserRoleRepository.
type mockUserRoleRepository struct {
	assignFn func(ctx context.Context, userRole *cryptoutilAppsTemplateServiceServerRepository.UserRole) error
}

func (m *mockUserRoleRepository) Assign(ctx context.Context, userRole *cryptoutilAppsTemplateServiceServerRepository.UserRole) error {
	if m.assignFn != nil {
		return m.assignFn(ctx, userRole)
	}

	return nil
}

func (m *mockUserRoleRepository) ListRolesByUser(_ context.Context, _ googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserRoleRepository) ListUsersByRole(_ context.Context, _ googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserRoleRepository) Revoke(_ context.Context, _, _ googleUuid.UUID) error {
	return errors.New("not implemented")
}

// Mock ClientRoleRepository.
type mockClientRoleRepository struct {
	assignFn func(ctx context.Context, clientRole *cryptoutilAppsTemplateServiceServerRepository.ClientRole) error
}

func (m *mockClientRoleRepository) Assign(ctx context.Context, clientRole *cryptoutilAppsTemplateServiceServerRepository.ClientRole) error {
	if m.assignFn != nil {
		return m.assignFn(ctx, clientRole)
	}

	return nil
}

func (m *mockClientRoleRepository) ListRolesByClient(_ context.Context, _ googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
	return nil, errors.New("not implemented")
}

func (m *mockClientRoleRepository) ListClientsByRole(_ context.Context, _ googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.Client, error) {
	return nil, errors.New("not implemented")
}

func (m *mockClientRoleRepository) Revoke(_ context.Context, _, _ googleUuid.UUID) error {
	return errors.New("not implemented")
}

func TestRegistrationService_RegisterUser_NewTenant(t *testing.T) {
	ctx := context.Background()
	tenantID := googleUuid.New()
	roleID := googleUuid.New()

	tests := []struct {
		name        string
		username    string
		email       string
		password    string
		tenantInfo  *NewTenantInfo
		setupMocks  func(*mockTenantService, *mockUserRepository, *mockRoleRepository, *mockUserRoleRepository)
		wantStatus  RegistrationStatus
		wantErr     bool
		errContains string
	}{
		{
			name:     "happy path - new tenant with user",
			username: "admin",
			email:    "admin@example.com",
			password: "hashed_password",
			tenantInfo: &NewTenantInfo{
				Name:        "Acme Corp",
				Description: "Test tenant",
			},
			setupMocks: func(tenantSvc *mockTenantService, userRepo *mockUserRepository, roleRepo *mockRoleRepository, userRoleRepo *mockUserRoleRepository) {
				tenantSvc.createTenantFn = func(_ context.Context, name, description string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
					return &cryptoutilAppsTemplateServiceServerRepository.Tenant{
						ID:          tenantID,
						Name:        name,
						Description: description,
						Active:      1,
					}, nil
				}
				userRepo.createFn = func(ctx context.Context, user *cryptoutilAppsTemplateServiceServerRepository.User) error {
					return nil
				}
				roleRepo.getByNameFn = func(ctx context.Context, tenantID googleUuid.UUID, name string) (*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
					return &cryptoutilAppsTemplateServiceServerRepository.Role{
						ID:       roleID,
						TenantID: tenantID,
						Name:     "admin",
					}, nil
				}
				userRoleRepo.assignFn = func(ctx context.Context, userRole *cryptoutilAppsTemplateServiceServerRepository.UserRole) error {
					return nil
				}
			},
			wantStatus: RegistrationStatusApproved,
			wantErr:    false,
		},
		{
			name:     "tenant creation fails",
			username: "admin",
			email:    "admin@example.com",
			password: "hashed_password",
			tenantInfo: &NewTenantInfo{
				Name:        "Acme Corp",
				Description: "Test tenant",
			},
			setupMocks: func(tenantSvc *mockTenantService, _ *mockUserRepository, _ *mockRoleRepository, _ *mockUserRoleRepository) {
				tenantSvc.createTenantFn = func(ctx context.Context, name, description string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
					return nil, errors.New("database error")
				}
			},
			wantErr:     true,
			errContains: "failed to create tenant",
		},
		{
			name:     "user creation fails",
			username: "admin",
			email:    "admin@example.com",
			password: "hashed_password",
			tenantInfo: &NewTenantInfo{
				Name:        "Acme Corp",
				Description: "Test tenant",
			},
			setupMocks: func(tenantSvc *mockTenantService, userRepo *mockUserRepository, _ *mockRoleRepository, _ *mockUserRoleRepository) {
				tenantSvc.createTenantFn = func(ctx context.Context, name, description string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
					return &cryptoutilAppsTemplateServiceServerRepository.Tenant{ID: tenantID, Name: name}, nil
				}
				userRepo.createFn = func(ctx context.Context, user *cryptoutilAppsTemplateServiceServerRepository.User) error {
					return errors.New("duplicate username")
				}
			},
			wantErr:     true,
			errContains: "failed to create user",
		},
		{
			name:     "role assignment fails",
			username: "admin",
			email:    "admin@example.com",
			password: "hashed_password",
			tenantInfo: &NewTenantInfo{
				Name:        "Acme Corp",
				Description: "Test tenant",
			},
			setupMocks: func(tenantSvc *mockTenantService, userRepo *mockUserRepository, roleRepo *mockRoleRepository, userRoleRepo *mockUserRoleRepository) {
				tenantSvc.createTenantFn = func(ctx context.Context, name, description string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
					return &cryptoutilAppsTemplateServiceServerRepository.Tenant{ID: tenantID, Name: name}, nil
				}
				userRepo.createFn = func(ctx context.Context, user *cryptoutilAppsTemplateServiceServerRepository.User) error {
					return nil
				}
				roleRepo.getByNameFn = func(ctx context.Context, tenantID googleUuid.UUID, name string) (*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
					return &cryptoutilAppsTemplateServiceServerRepository.Role{ID: roleID, TenantID: tenantID, Name: "admin"}, nil
				}
				userRoleRepo.assignFn = func(ctx context.Context, userRole *cryptoutilAppsTemplateServiceServerRepository.UserRole) error {
					return errors.New("database error")
				}
			},
			wantErr:     true,
			errContains: "failed to assign admin role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantSvc := &mockTenantService{}
			userRepo := &mockUserRepository{}
			roleRepo := &mockRoleRepository{}
			userRoleRepo := &mockUserRoleRepository{}

			if tt.setupMocks != nil {
				tt.setupMocks(tenantSvc, userRepo, roleRepo, userRoleRepo)
			}

			service := NewRegistrationService(
				tenantSvc,
				userRepo,
				nil, // clientRepo.
				nil, // unverifiedUserRepo.
				nil, // unverifiedClientRepo.
				roleRepo,
				userRoleRepo,
				nil, // clientRoleRepo.
			)

			result, err := service.RegisterUser(ctx, tt.username, tt.email, tt.password, tt.tenantInfo, nil)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tt.wantStatus, result.Status)
				require.NotNil(t, result.UserID)
			}
		})
	}
}

func TestRegistrationService_RegisterUser_ExistingTenant(t *testing.T) {
	ctx := context.Background()
	tenantID := googleUuid.New()

	tests := []struct {
		name        string
		username    string
		email       string
		password    string
		setupMocks  func(*mockUnverifiedUserRepository)
		wantStatus  RegistrationStatus
		wantErr     bool
		errContains string
	}{
		{
			name:     "happy path - pending verification",
			username: "newuser",
			email:    "newuser@example.com",
			password: "hashed_password",
			setupMocks: func(unverifiedUserRepo *mockUnverifiedUserRepository) {
				unverifiedUserRepo.createFn = func(ctx context.Context, user *cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser) error {
					return nil
				}
			},
			wantStatus: RegistrationStatusPending,
			wantErr:    false,
		},
		{
			name:     "unverified user creation fails",
			username: "newuser",
			email:    "newuser@example.com",
			password: "hashed_password",
			setupMocks: func(unverifiedUserRepo *mockUnverifiedUserRepository) {
				unverifiedUserRepo.createFn = func(ctx context.Context, user *cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser) error {
					return errors.New("duplicate username")
				}
			},
			wantErr:     true,
			errContains: "failed to create unverified user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unverifiedUserRepo := &mockUnverifiedUserRepository{}

			if tt.setupMocks != nil {
				tt.setupMocks(unverifiedUserRepo)
			}

			service := NewRegistrationService(
				nil, // tenantService.
				nil, // userRepo.
				nil, // clientRepo.
				unverifiedUserRepo,
				nil, // unverifiedClientRepo.
				nil, // roleRepo.
				nil, // userRoleRepo.
				nil, // clientRoleRepo.
			)

			result, err := service.RegisterUser(ctx, tt.username, tt.email, tt.password, nil, &tenantID)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tt.wantStatus, result.Status)
				require.Nil(t, result.UserID)
				require.NotNil(t, result.ExpiresAt)
				require.True(t, result.ExpiresAt.After(time.Now()))
			}
		})
	}
}

func TestRegistrationService_RegisterUser_ValidationErrors(t *testing.T) {
	ctx := context.Background()
	tenantID := googleUuid.New()

	service := NewRegistrationService(nil, nil, nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name             string
		newTenant        *NewTenantInfo
		existingTenantID *googleUuid.UUID
		errContains      string
	}{
		{
			name:             "neither newTenant nor existingTenantID provided",
			newTenant:        nil,
			existingTenantID: nil,
			errContains:      "exactly one of newTenant or existingTenantID must be provided",
		},
		{
			name: "both newTenant and existingTenantID provided",
			newTenant: &NewTenantInfo{
				Name:        "Test",
				Description: "Test",
			},
			existingTenantID: &tenantID,
			errContains:      "exactly one of newTenant or existingTenantID must be provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.RegisterUser(ctx, "user", "user@example.com", "hash", tt.newTenant, tt.existingTenantID)

			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errContains)
			require.Nil(t, result)
		})
	}
}

func TestRegistrationService_RegisterClient_NewTenant(t *testing.T) {
	ctx := context.Background()
	tenantID := googleUuid.New()
	roleID := googleUuid.New()

	tenantSvc := &mockTenantService{
		createTenantFn: func(ctx context.Context, name, description string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
			return &cryptoutilAppsTemplateServiceServerRepository.Tenant{ID: tenantID, Name: name, Active: 1}, nil
		},
	}
	clientRepo := &mockClientRepository{
		createFn: func(ctx context.Context, client *cryptoutilAppsTemplateServiceServerRepository.Client) error {
			return nil
		},
	}
	roleRepo := &mockRoleRepository{
		getByNameFn: func(ctx context.Context, tenantID googleUuid.UUID, name string) (*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
			return &cryptoutilAppsTemplateServiceServerRepository.Role{ID: roleID, TenantID: tenantID, Name: "admin"}, nil
		},
	}
	clientRoleRepo := &mockClientRoleRepository{
		assignFn: func(ctx context.Context, clientRole *cryptoutilAppsTemplateServiceServerRepository.ClientRole) error {
			return nil
		},
	}

	service := NewRegistrationService(
		tenantSvc,
		nil, // userRepo.
		clientRepo,
		nil, // unverifiedUserRepo.
		nil, // unverifiedClientRepo.
		roleRepo,
		nil, // userRoleRepo.
		clientRoleRepo,
	)

	result, err := service.RegisterClient(ctx, "client-id", "client-secret", &NewTenantInfo{Name: "Test Corp", Description: "Test"}, nil)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, RegistrationStatusApproved, result.Status)
	require.NotNil(t, result.ClientID)
}

func TestRegistrationService_RegisterClient_ExistingTenant(t *testing.T) {
	ctx := context.Background()
	tenantID := googleUuid.New()

	unverifiedClientRepo := &mockUnverifiedClientRepository{
		createFn: func(ctx context.Context, client *cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient) error {
			return nil
		},
	}

	service := NewRegistrationService(
		nil, // tenantService.
		nil, // userRepo.
		nil, // clientRepo.
		nil, // unverifiedUserRepo.
		unverifiedClientRepo,
		nil, // roleRepo.
		nil, // userRoleRepo.
		nil, // clientRoleRepo.
	)

	result, err := service.RegisterClient(ctx, "client-id", "client-secret", nil, &tenantID)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, RegistrationStatusPending, result.Status)
	require.Nil(t, result.ClientID)
	require.NotNil(t, result.ExpiresAt)
}
