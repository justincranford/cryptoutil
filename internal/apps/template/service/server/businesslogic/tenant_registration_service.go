// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package businesslogic

import (
	"context"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceServerDomain "cryptoutil/internal/apps/template/service/server/domain"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// TenantRegistrationService handles tenant creation and join request workflows.
type TenantRegistrationService struct {
	db              *gorm.DB
	tenantRepo      cryptoutilAppsTemplateServiceServerRepository.TenantRepository
	userRepo        cryptoutilAppsTemplateServiceServerRepository.UserRepository
	joinRequestRepo cryptoutilAppsTemplateServiceServerRepository.TenantJoinRequestRepository
}

// NewTenantRegistrationService creates a new tenant registration service.
func NewTenantRegistrationService(
	db *gorm.DB,
	tenantRepo cryptoutilAppsTemplateServiceServerRepository.TenantRepository,
	userRepo cryptoutilAppsTemplateServiceServerRepository.UserRepository,
	joinRequestRepo cryptoutilAppsTemplateServiceServerRepository.TenantJoinRequestRepository,
) *TenantRegistrationService {
	return &TenantRegistrationService{
		db:              db,
		tenantRepo:      tenantRepo,
		userRepo:        userRepo,
		joinRequestRepo: joinRequestRepo,
	}
}

// RegisterUserWithTenant registers a user with a tenant (create or join).
// Parameters:
// - userID: Pre-generated UUIDv7 for the new user
// - username: Validated username (3-50 characters)
// - email: Validated email address
// - passwordHash: PBKDF2-HMAC-SHA256 hashed password
// - tenantName: Name of tenant to create or join
// - createTenant: If true, create new tenant; if false, request to join existing.
func (s *TenantRegistrationService) RegisterUserWithTenant(
	ctx context.Context,
	userID googleUuid.UUID,
	username string,
	email string,
	passwordHash string,
	tenantName string,
	createTenant bool,
) (*cryptoutilAppsTemplateServiceServerRepository.Tenant, error) {
	if createTenant {
		// Create new tenant with user as admin.
		tenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{
			ID:   googleUuid.Must(googleUuid.NewV7()),
			Name: tenantName,
		}

		if err := s.tenantRepo.Create(ctx, tenant); err != nil {
			return nil, fmt.Errorf("failed to create tenant: %w", err)
		}

		// Create user with tenant association.
		user := &cryptoutilAppsTemplateServiceServerRepository.User{
			ID:           userID,
			TenantID:     tenant.ID,
			Username:     username,
			Email:        email,
			PasswordHash: passwordHash,
			Active:       1, // Active by default.
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		// Note: Admin role assignment requires role management system.
		// For now, first user of a tenant is implicitly admin.

		return tenant, nil
	}

	// Join existing tenant workflow - create join request.
	// User must provide existing tenant ID - find by name.
	// This is a simplified flow; production would have tenant discovery.
	return nil, fmt.Errorf("join existing tenant flow not yet implemented")
}

// RegisterClientWithTenant registers a client with a tenant.
func (s *TenantRegistrationService) RegisterClientWithTenant(
	ctx context.Context,
	clientID googleUuid.UUID,
	tenantID googleUuid.UUID,
) error {
	// Verify tenant exists
	_, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	// Create join request for client
	joinRequest := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
		ID:          googleUuid.New(),
		ClientID:    &clientID,
		TenantID:    tenantID,
		Status:      cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending,
		RequestedAt: time.Now().UTC(),
	}

	if err := s.joinRequestRepo.Create(ctx, joinRequest); err != nil {
		return fmt.Errorf("failed to create client join request: %w", err)
	}

	return nil
}

// AuthorizeJoinRequest approves or rejects a join request.
func (s *TenantRegistrationService) AuthorizeJoinRequest(
	ctx context.Context,
	requestID googleUuid.UUID,
	adminUserID googleUuid.UUID,
	approved bool,
) error {
	// Get join request
	joinRequest, err := s.joinRequestRepo.GetByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get join request: %w", err)
	}

	// Verify request is pending
	if joinRequest.Status != cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending {
		return fmt.Errorf("join request is not pending (status: %s)", joinRequest.Status)
	}

	// TODO: Verify admin has permission for this tenant

	// Update request status
	now := time.Now().UTC()
	joinRequest.ProcessedAt = &now
	joinRequest.ProcessedBy = &adminUserID

	if approved {
		joinRequest.Status = cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusApproved
		// TODO: Assign user/client to tenant with appropriate role
	} else {
		joinRequest.Status = cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusRejected
	}

	if err := s.joinRequestRepo.Update(ctx, joinRequest); err != nil {
		return fmt.Errorf("failed to update join request: %w", err)
	}

	return nil
}

// ListJoinRequests lists join requests for a tenant.
func (s *TenantRegistrationService) ListJoinRequests(
	ctx context.Context,
	tenantID googleUuid.UUID,
) ([]*cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest, error) {
	// TODO: Verify caller has admin permission for this tenant
	requests, err := s.joinRequestRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list join requests: %w", err)
	}

	return requests, nil
}
