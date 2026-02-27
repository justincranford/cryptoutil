// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

package repository

import (
	"context"
	"errors"
	"fmt"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceServerDomain "cryptoutil/internal/apps/template/service/server/domain"
)

// tenantJoinRequestRepository implements TenantJoinRequestRepository using GORM.
type tenantJoinRequestRepository struct {
	db *gorm.DB
}

// NewTenantJoinRequestRepository creates a new tenant join request repository.
func NewTenantJoinRequestRepository(db *gorm.DB) TenantJoinRequestRepository {
	return &tenantJoinRequestRepository{db: db}
}

func (r *tenantJoinRequestRepository) Create(ctx context.Context, request *cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest) error {
	if err := GetDB(ctx, r.db).WithContext(ctx).Create(request).Error; err != nil {
		return fmt.Errorf("failed to create join request: %w", err)
	}

	return nil
}

func (r *tenantJoinRequestRepository) Update(ctx context.Context, request *cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest) error {
	if err := GetDB(ctx, r.db).WithContext(ctx).Save(request).Error; err != nil {
		return fmt.Errorf("failed to update join request: %w", err)
	}

	return nil
}

func (r *tenantJoinRequestRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest, error) {
	var request cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest

	if err := GetDB(ctx, r.db).WithContext(ctx).Where("id = ?", id.String()).First(&request).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("join request not found")
		}

		return nil, fmt.Errorf("failed to get join request: %w", err)
	}

	return &request, nil
}

func (r *tenantJoinRequestRepository) ListByTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest, error) {
	var requests []*cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest

	if err := GetDB(ctx, r.db).WithContext(ctx).
		Where("tenant_id = ?", tenantID.String()).
		Order("requested_at DESC").
		Find(&requests).Error; err != nil {
		return nil, fmt.Errorf("failed to list join requests by tenant: %w", err)
	}

	return requests, nil
}

func (r *tenantJoinRequestRepository) ListByStatus(ctx context.Context, status string) ([]*cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest, error) {
	var requests []*cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest

	if err := GetDB(ctx, r.db).WithContext(ctx).
		Where("status = ?", status).
		Order("requested_at DESC").
		Find(&requests).Error; err != nil {
		return nil, fmt.Errorf("failed to list join requests by status: %w", err)
	}

	return requests, nil
}

func (r *tenantJoinRequestRepository) ListByTenantAndStatus(ctx context.Context, tenantID googleUuid.UUID, status string) ([]*cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest, error) {
	var requests []*cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest

	if err := GetDB(ctx, r.db).WithContext(ctx).
		Where("tenant_id = ? AND status = ?", tenantID.String(), status).
		Order("requested_at DESC").
		Find(&requests).Error; err != nil {
		return nil, fmt.Errorf("failed to list join requests by tenant and status: %w", err)
	}

	return requests, nil
}
