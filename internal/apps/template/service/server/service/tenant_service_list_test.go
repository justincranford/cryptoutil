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

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

func TestTenantService_ListTenants(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	activeTenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{
		ID:     googleUuid.New(),
		Name:   "active-tenant",
		Active: 1,
	}
	inactiveTenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{
		ID:     googleUuid.New(),
		Name:   "inactive-tenant",
		Active: 0,
	}

	tests := []struct {
		name       string
		activeOnly bool
		setupMocks func(*mockTenantRepository)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "list all tenants",
			activeOnly: false,
			setupMocks: func(tenantRepo *mockTenantRepository) {
				tenantRepo.listFn = func(_ context.Context, _ bool) ([]*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
					return []*cryptoutilAppsTemplateServiceServerRepository.Tenant{activeTenant, inactiveTenant}, nil
				}
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "list active tenants only",
			activeOnly: true,
			setupMocks: func(tenantRepo *mockTenantRepository) {
				tenantRepo.listFn = func(_ context.Context, _ bool) ([]*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
					return []*cryptoutilAppsTemplateServiceServerRepository.Tenant{activeTenant}, nil
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "repository error",
			activeOnly: false,
			setupMocks: func(tenantRepo *mockTenantRepository) {
				tenantRepo.listFn = func(_ context.Context, _ bool) ([]*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
					return nil, errors.New("database error")
				}
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantRepo := &mockTenantRepository{}

			if tt.setupMocks != nil {
				tt.setupMocks(tenantRepo)
			}

			service := NewTenantService(tenantRepo, nil)

			tenants, err := service.ListTenants(ctx, tt.activeOnly)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, tenants)
			} else {
				require.NoError(t, err)
				require.Len(t, tenants, tt.wantCount)
			}
		})
	}
}

func TestTenantService_UpdateTenant_UpdateError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := googleUuid.New()
	existingTenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{
		ID:          tenantID,
		Name:        "old-name",
		Description: "old description",
		Active:      1,
	}

	tenantRepo := &mockTenantRepository{
		getByIDFn: func(_ context.Context, _ googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
			return existingTenant, nil
		},
		updateFn: func(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.Tenant) error {
			return errors.New("database update error")
		},
	}

	service := NewTenantService(tenantRepo, nil)

	newName := "new-name"
	_, err := service.UpdateTenant(ctx, tenantID, &newName, nil, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to update tenant")
}

func TestTenantService_GetTenantByName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantName := "test-tenant"
	tenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{
		ID:   googleUuid.New(),
		Name: tenantName,
	}

	tests := []struct {
		name       string
		setupMocks func(*mockTenantRepository)
		wantErr    bool
	}{
		{
			name: "happy path",
			setupMocks: func(tenantRepo *mockTenantRepository) {
				tenantRepo.getByNameFn = func(_ context.Context, _ string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
					return tenant, nil
				}
			},
			wantErr: false,
		},
		{
			name: "tenant not found",
			setupMocks: func(tenantRepo *mockTenantRepository) {
				summary := testErrSummaryTenantNotFound
				tenantRepo.getByNameFn = func(_ context.Context, _ string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
					return nil, cryptoutilSharedApperr.NewHTTP404NotFound(&summary, errors.New("not found"))
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantRepo := &mockTenantRepository{}

			if tt.setupMocks != nil {
				tt.setupMocks(tenantRepo)
			}

			service := NewTenantService(tenantRepo, nil)

			result, err := service.GetTenantByName(ctx, tenantName)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tenantName, result.Name)
			}
		})
	}
}
