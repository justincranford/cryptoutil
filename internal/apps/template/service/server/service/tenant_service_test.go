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
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

// Test error summary constants.
const (
	testErrSummaryTenantNotFound = "tenant not found"
)

// mockTenantRepository implements repository.TenantRepository for testing.
type mockTenantRepository struct {
	createFn               func(ctx context.Context, tenant *cryptoutilAppsTemplateServiceServerRepository.Tenant) error
	getByIDFn              func(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error)
	getByNameFn            func(ctx context.Context, name string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error)
	listFn                 func(ctx context.Context, activeOnly bool) ([]*cryptoutilAppsTemplateServiceServerRepository.Tenant, error)
	updateFn               func(ctx context.Context, tenant *cryptoutilAppsTemplateServiceServerRepository.Tenant) error
	deleteFn               func(ctx context.Context, id googleUuid.UUID) error
	countUsersAndClientsFn func(ctx context.Context, tenantID googleUuid.UUID) (int64, int64, error)
}

func (m *mockTenantRepository) Create(ctx context.Context, tenant *cryptoutilAppsTemplateServiceServerRepository.Tenant) error {
	if m.createFn != nil {
		return m.createFn(ctx, tenant)
	}

	return nil
}

func (m *mockTenantRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}

	return nil, errors.New("not implemented")
}

func (m *mockTenantRepository) GetByName(ctx context.Context, name string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
	if m.getByNameFn != nil {
		return m.getByNameFn(ctx, name)
	}

	return nil, errors.New("not implemented")
}

func (m *mockTenantRepository) List(ctx context.Context, activeOnly bool) ([]*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
	if m.listFn != nil {
		return m.listFn(ctx, activeOnly)
	}

	return nil, errors.New("not implemented")
}

func (m *mockTenantRepository) Update(ctx context.Context, tenant *cryptoutilAppsTemplateServiceServerRepository.Tenant) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, tenant)
	}

	return nil
}

func (m *mockTenantRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}

	return nil
}

func (m *mockTenantRepository) CountUsersAndClients(ctx context.Context, tenantID googleUuid.UUID) (int64, int64, error) {
	if m.countUsersAndClientsFn != nil {
		return m.countUsersAndClientsFn(ctx, tenantID)
	}

	return 0, 0, nil
}

// mockRoleRepository implements repository.RoleRepository for testing.
type mockRoleRepository struct {
	createFn       func(ctx context.Context, role *cryptoutilAppsTemplateServiceServerRepository.Role) error
	getByNameFn    func(ctx context.Context, tenantID googleUuid.UUID, name string) (*cryptoutilAppsTemplateServiceServerRepository.Role, error)
	getByIDFn      func(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Role, error)
	listByTenantFn func(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.Role, error)
	deleteFn       func(ctx context.Context, id googleUuid.UUID) error
}

func (m *mockRoleRepository) Create(ctx context.Context, role *cryptoutilAppsTemplateServiceServerRepository.Role) error {
	if m.createFn != nil {
		return m.createFn(ctx, role)
	}

	return nil
}

func (m *mockRoleRepository) GetByName(ctx context.Context, tenantID googleUuid.UUID, name string) (*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
	if m.getByNameFn != nil {
		return m.getByNameFn(ctx, tenantID, name)
	}

	return nil, errors.New("not implemented")
}

func (m *mockRoleRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}

	return nil, errors.New("not implemented")
}

func (m *mockRoleRepository) ListByTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.Role, error) {
	if m.listByTenantFn != nil {
		return m.listByTenantFn(ctx, tenantID)
	}

	return nil, errors.New("not implemented")
}

func (m *mockRoleRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}

	return errors.New("not implemented")
}

func TestTenantService_CreateTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		tenantName  string
		description string
		setupMocks  func(*mockTenantRepository, *mockRoleRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:        "happy path",
			tenantName:  "Test Tenant",
			description: "A test tenant",
			setupMocks: func(tenantRepo *mockTenantRepository, roleRepo *mockRoleRepository) {
				tenantRepo.createFn = func(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.Tenant) error {
					return nil
				}
				roleRepo.createFn = func(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.Role) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:        "duplicate tenant name",
			tenantName:  "Duplicate Tenant",
			description: "A duplicate tenant",
			setupMocks: func(tenantRepo *mockTenantRepository, _ *mockRoleRepository) {
				summary := "duplicate key violation"
				tenantRepo.createFn = func(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.Tenant) error {
					return cryptoutilSharedApperr.NewHTTP409Conflict(&summary, errors.New("UNIQUE constraint failed"))
				}
			},
			wantErr:     true,
			errContains: "failed to create tenant",
		},
		{
			name:        "role creation fails",
			tenantName:  "Test Tenant",
			description: "Tenant with role failure",
			setupMocks: func(tenantRepo *mockTenantRepository, roleRepo *mockRoleRepository) {
				tenantRepo.createFn = func(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.Tenant) error {
					return nil
				}
				summary := "database error"
				roleRepo.createFn = func(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.Role) error {
					return cryptoutilSharedApperr.NewHTTP500InternalServerError(&summary, errors.New("database error"))
				}
			},
			wantErr:     true,
			errContains: "failed to create admin role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantRepo := &mockTenantRepository{}
			roleRepo := &mockRoleRepository{}

			if tt.setupMocks != nil {
				tt.setupMocks(tenantRepo, roleRepo)
			}

			service := NewTenantService(tenantRepo, roleRepo)

			tenant, err := service.CreateTenant(ctx, tt.tenantName, tt.description)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
				require.Nil(t, tenant)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tenant)
				require.Equal(t, tt.tenantName, tenant.Name)
				require.Equal(t, tt.description, tenant.Description)
				require.Equal(t, 1, tenant.Active)
			}
		})
	}
}

func TestTenantService_GetTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := googleUuid.New()

	tests := []struct {
		name        string
		tenantID    googleUuid.UUID
		setupMocks  func(*mockTenantRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:     "happy path",
			tenantID: tenantID,
			setupMocks: func(tenantRepo *mockTenantRepository) {
				tenantRepo.getByIDFn = func(_ context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
					return &cryptoutilAppsTemplateServiceServerRepository.Tenant{ID: id, Name: "Test Tenant", Active: 1}, nil
				}
			},
			wantErr: false,
		},
		{
			name:     "tenant not found",
			tenantID: googleUuid.New(),
			setupMocks: func(tenantRepo *mockTenantRepository) {
				summary := testErrSummaryTenantNotFound
				tenantRepo.getByIDFn = func(_ context.Context, _ googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
					return nil, cryptoutilSharedApperr.NewHTTP404NotFound(&summary, errors.New("not found"))
				}
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantRepo := &mockTenantRepository{}

			if tt.setupMocks != nil {
				tt.setupMocks(tenantRepo)
			}

			service := NewTenantService(tenantRepo, nil)

			tenant, err := service.GetTenant(ctx, tt.tenantID)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
				require.Nil(t, tenant)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tenant)
				require.Equal(t, tt.tenantID, tenant.ID)
			}
		})
	}
}

func TestTenantService_UpdateTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := googleUuid.New()

	tests := []struct {
		name         string
		tenantID     googleUuid.UUID
		nameUpdate   *string
		descUpdate   *string
		activeUpdate *bool
		setupMocks   func(*mockTenantRepository)
		wantErr      bool
		errContains  string
	}{
		{
			name:         "update name only",
			tenantID:     tenantID,
			nameUpdate:   stringPtr("Updated Name"),
			descUpdate:   nil,
			activeUpdate: nil,
			setupMocks: func(tenantRepo *mockTenantRepository) {
				tenantRepo.getByIDFn = func(_ context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
					return &cryptoutilAppsTemplateServiceServerRepository.Tenant{
						ID:          id,
						Name:        "Original Name",
						Description: "Original Description",
						Active:      1,
						CreatedAt:   time.Now().UTC(),
					}, nil
				}
				tenantRepo.updateFn = func(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.Tenant) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:         "update all fields",
			tenantID:     tenantID,
			nameUpdate:   stringPtr("New Name"),
			descUpdate:   stringPtr("New Description"),
			activeUpdate: boolPtr(false),
			setupMocks: func(tenantRepo *mockTenantRepository) {
				tenantRepo.getByIDFn = func(_ context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
					return &cryptoutilAppsTemplateServiceServerRepository.Tenant{
						ID:          id,
						Name:        "Old Name",
						Description: "Old Description",
						Active:      1,
						CreatedAt:   time.Now().UTC(),
					}, nil
				}
				tenantRepo.updateFn = func(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.Tenant) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:       "tenant not found",
			tenantID:   googleUuid.New(),
			nameUpdate: stringPtr("Test"),
			setupMocks: func(tenantRepo *mockTenantRepository) {
				summary := testErrSummaryTenantNotFound
				tenantRepo.getByIDFn = func(_ context.Context, _ googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
					return nil, cryptoutilSharedApperr.NewHTTP404NotFound(&summary, errors.New("not found"))
				}
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantRepo := &mockTenantRepository{}

			if tt.setupMocks != nil {
				tt.setupMocks(tenantRepo)
			}

			service := NewTenantService(tenantRepo, nil)

			tenant, err := service.UpdateTenant(ctx, tt.tenantID, tt.nameUpdate, tt.descUpdate, tt.activeUpdate)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
				require.Nil(t, tenant)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tenant)

				if tt.nameUpdate != nil {
					require.Equal(t, *tt.nameUpdate, tenant.Name)
				}

				if tt.descUpdate != nil {
					require.Equal(t, *tt.descUpdate, tenant.Description)
				}

				if tt.activeUpdate != nil {
					expectedActive := 0
					if *tt.activeUpdate {
						expectedActive = 1
					}

					require.Equal(t, expectedActive, tenant.Active)
				}
			}
		})
	}
}

func TestTenantService_DeleteTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := googleUuid.New()

	tests := []struct {
		name        string
		tenantID    googleUuid.UUID
		setupMocks  func(*mockTenantRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:     "happy path",
			tenantID: tenantID,
			setupMocks: func(tenantRepo *mockTenantRepository) {
				tenantRepo.deleteFn = func(_ context.Context, _ googleUuid.UUID) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:     "tenant has users or clients",
			tenantID: tenantID,
			setupMocks: func(tenantRepo *mockTenantRepository) {
				summary := "cannot delete tenant: has 1 users and 0 clients"
				tenantRepo.deleteFn = func(_ context.Context, _ googleUuid.UUID) error {
					return cryptoutilSharedApperr.NewHTTP409Conflict(&summary, errors.New("tenant has users"))
				}
			},
			wantErr:     true,
			errContains: "cannot delete tenant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantRepo := &mockTenantRepository{}

			if tt.setupMocks != nil {
				tt.setupMocks(tenantRepo)
			}

			service := NewTenantService(tenantRepo, nil)

			err := service.DeleteTenant(ctx, tt.tenantID)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Helper functions.

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
