// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

package repository

import (
	"context"

	googleUuid "github.com/google/uuid"

	cryptoutilAppsTemplateServiceServerDomain "cryptoutil/internal/apps/template/service/server/domain"
)

// TenantJoinRequestRepository defines operations for managing tenant join requests.
type TenantJoinRequestRepository interface {
	// Create creates a new join request.
	Create(ctx context.Context, request *cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest) error

	// Update updates an existing join request.
	Update(ctx context.Context, request *cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest) error

	// GetByID retrieves a join request by ID.
	GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest, error)

	// ListByTenant lists all join requests for a specific tenant.
	ListByTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest, error)

	// ListByStatus lists all join requests with a specific status.
	ListByStatus(ctx context.Context, status string) ([]*cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest, error)

	// ListByTenantAndStatus lists join requests for a tenant with a specific status.
	ListByTenantAndStatus(ctx context.Context, tenantID googleUuid.UUID, status string) ([]*cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest, error)
}
