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
	"fmt"

	googleUuid "github.com/google/uuid"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// TenantService defines operations for tenant management.
type TenantService interface {
	// CreateTenant creates a new tenant with default admin role.
	CreateTenant(ctx context.Context, name, description string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error)

	// GetTenant retrieves a tenant by ID.
	GetTenant(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error)

	// GetTenantByName retrieves a tenant by name.
	GetTenantByName(ctx context.Context, name string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error)

	// ListTenants retrieves all tenants with optional active filtering.
	ListTenants(ctx context.Context, activeOnly bool) ([]*cryptoutilAppsTemplateServiceServerRepository.Tenant, error)

	// UpdateTenant updates tenant information.
	UpdateTenant(ctx context.Context, id googleUuid.UUID, name, description *string, active *bool) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error)

	// DeleteTenant deletes a tenant (fails if users/clients exist).
	DeleteTenant(ctx context.Context, id googleUuid.UUID) error
}

// TenantServiceImpl implements TenantService.
type TenantServiceImpl struct {
	tenantRepo cryptoutilAppsTemplateServiceServerRepository.TenantRepository
	roleRepo   cryptoutilAppsTemplateServiceServerRepository.RoleRepository
}

// NewTenantService creates a new TenantService instance.
func NewTenantService(tenantRepo cryptoutilAppsTemplateServiceServerRepository.TenantRepository, roleRepo cryptoutilAppsTemplateServiceServerRepository.RoleRepository) TenantService {
	return &TenantServiceImpl{
		tenantRepo: tenantRepo,
		roleRepo:   roleRepo,
	}
}

// CreateTenant creates a new tenant with default admin role.
func (s *TenantServiceImpl) CreateTenant(ctx context.Context, name, description string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
	// Create tenant.
	tenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{
		ID:          googleUuid.New(),
		Name:        name,
		Description: description,
		Active:      1, // 1 = active.
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create default admin role for the new tenant.
	adminRole := &cryptoutilAppsTemplateServiceServerRepository.Role{
		ID:          googleUuid.New(),
		TenantID:    tenant.ID,
		Name:        "admin",
		Description: "Tenant administrator role",
	}

	if err := s.roleRepo.Create(ctx, adminRole); err != nil {
		// Rollback tenant creation would require transaction support.
		// For now, we'll leave the tenant without a role and return an error.
		return nil, fmt.Errorf("failed to create admin role for tenant: %w", err)
	}

	return tenant, nil
}

// GetTenant retrieves a tenant by ID.
func (s *TenantServiceImpl) GetTenant(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by ID: %w", err)
	}

	return tenant, nil
}

// GetTenantByName retrieves a tenant by name.
func (s *TenantServiceImpl) GetTenantByName(ctx context.Context, name string) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
	tenant, err := s.tenantRepo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by name: %w", err)
	}

	return tenant, nil
}

// ListTenants retrieves all tenants with optional active filtering.
func (s *TenantServiceImpl) ListTenants(ctx context.Context, activeOnly bool) ([]*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
	tenants, err := s.tenantRepo.List(ctx, activeOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	return tenants, nil
}

// UpdateTenant updates tenant information.
func (s *TenantServiceImpl) UpdateTenant(ctx context.Context, id googleUuid.UUID, name, description *string, active *bool) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
	// Get existing tenant.
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant for update: %w", err)
	}

	// Update fields if provided.
	if name != nil {
		tenant.Name = *name
	}

	if description != nil {
		tenant.Description = *description
	}

	if active != nil {
		tenant.SetActive(*active)
	}

	// Save changes.
	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}

// DeleteTenant deletes a tenant (fails if users/clients exist).
func (s *TenantServiceImpl) DeleteTenant(ctx context.Context, id googleUuid.UUID) error {
	if err := s.tenantRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	return nil
}
