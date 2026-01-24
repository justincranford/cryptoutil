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
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// VerificationService defines operations for verifying (approving) pending user and client registrations.
type VerificationService interface {
	// ListPendingUsers lists all unverified users for a tenant.
	ListPendingUsers(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser, error)

	// ListPendingClients lists all unverified clients for a tenant.
	ListPendingClients(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient, error)

	// ApproveUser approves a pending user registration and assigns roles.
	ApproveUser(ctx context.Context, tenantID, unverifiedUserID googleUuid.UUID, roleIDs []googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.User, error)

	// ApproveClient approves a pending client registration and assigns roles.
	ApproveClient(ctx context.Context, tenantID, unverifiedClientID googleUuid.UUID, roleIDs []googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Client, error)

	// RejectUser rejects a pending user registration (removes from unverified table).
	RejectUser(ctx context.Context, tenantID, unverifiedUserID googleUuid.UUID) error

	// RejectClient rejects a pending client registration (removes from unverified table).
	RejectClient(ctx context.Context, tenantID, unverifiedClientID googleUuid.UUID) error

	// CleanupExpiredRegistrations removes all expired unverified user and client registrations.
	CleanupExpiredRegistrations(ctx context.Context) error
}

// VerificationServiceImpl implements VerificationService.
type VerificationServiceImpl struct {
	userRepo             cryptoutilAppsTemplateServiceServerRepository.UserRepository
	clientRepo           cryptoutilAppsTemplateServiceServerRepository.ClientRepository
	unverifiedUserRepo   cryptoutilAppsTemplateServiceServerRepository.UnverifiedUserRepository
	unverifiedClientRepo cryptoutilAppsTemplateServiceServerRepository.UnverifiedClientRepository
	roleRepo             cryptoutilAppsTemplateServiceServerRepository.RoleRepository
	userRoleRepo         cryptoutilAppsTemplateServiceServerRepository.UserRoleRepository
	clientRoleRepo       cryptoutilAppsTemplateServiceServerRepository.ClientRoleRepository
}

// NewVerificationService creates a new VerificationService instance.
func NewVerificationService(
	userRepo cryptoutilAppsTemplateServiceServerRepository.UserRepository,
	clientRepo cryptoutilAppsTemplateServiceServerRepository.ClientRepository,
	unverifiedUserRepo cryptoutilAppsTemplateServiceServerRepository.UnverifiedUserRepository,
	unverifiedClientRepo cryptoutilAppsTemplateServiceServerRepository.UnverifiedClientRepository,
	roleRepo cryptoutilAppsTemplateServiceServerRepository.RoleRepository,
	userRoleRepo cryptoutilAppsTemplateServiceServerRepository.UserRoleRepository,
	clientRoleRepo cryptoutilAppsTemplateServiceServerRepository.ClientRoleRepository,
) VerificationService {
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

// ListPendingUsers lists all unverified users for a tenant.
func (s *VerificationServiceImpl) ListPendingUsers(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser, error) {
	users, err := s.unverifiedUserRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending users: %w", err)
	}

	return users, nil
}

// ListPendingClients lists all unverified clients for a tenant.
func (s *VerificationServiceImpl) ListPendingClients(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient, error) {
	clients, err := s.unverifiedClientRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending clients: %w", err)
	}

	return clients, nil
}

// ApproveUser approves a pending user registration and assigns roles.
func (s *VerificationServiceImpl) ApproveUser(ctx context.Context, tenantID, unverifiedUserID googleUuid.UUID, roleIDs []googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.User, error) {
	// Get the unverified user.
	unverifiedUser, err := s.unverifiedUserRepo.GetByID(ctx, unverifiedUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unverified user: %w", err)
	}

	// Verify the user belongs to the specified tenant.
	if unverifiedUser.TenantID != tenantID {
		return nil, fmt.Errorf("unverified user does not belong to the specified tenant")
	}

	// Check if the registration has expired.
	if time.Now().After(unverifiedUser.ExpiresAt) {
		return nil, fmt.Errorf("user registration has expired")
	}

	// At least one role must be assigned.
	if len(roleIDs) == 0 {
		return nil, fmt.Errorf("at least one role must be assigned when approving a user")
	}

	// Verify all role IDs belong to the tenant.
	for _, roleID := range roleIDs {
		role, err := s.roleRepo.GetByID(ctx, roleID)
		if err != nil {
			return nil, fmt.Errorf("failed to get role %s: %w", roleID, err)
		}

		if role.TenantID != tenantID {
			return nil, fmt.Errorf("role %s does not belong to the specified tenant", roleID)
		}
	}

	// Create the verified user.
	user := &cryptoutilAppsTemplateServiceServerRepository.User{
		ID:           googleUuid.New(),
		TenantID:     unverifiedUser.TenantID,
		Username:     unverifiedUser.Username,
		Email:        unverifiedUser.Email,
		PasswordHash: unverifiedUser.PasswordHash,
		Active:       1,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create verified user: %w", err)
	}

	// Assign roles.
	for _, roleID := range roleIDs {
		userRole := &cryptoutilAppsTemplateServiceServerRepository.UserRole{
			TenantID: tenantID,
			UserID:   user.ID,
			RoleID:   roleID,
		}
		if err := s.userRoleRepo.Assign(ctx, userRole); err != nil {
			return nil, fmt.Errorf("failed to assign role %s: %w", roleID, err)
		}
	}

	// Delete the unverified user record.
	if err := s.unverifiedUserRepo.Delete(ctx, unverifiedUserID); err != nil {
		// Log but don't fail - user is already created.
		// In production, this should use a transaction.
		return user, fmt.Errorf("user approved but failed to delete unverified record: %w", err)
	}

	return user, nil
}

// ApproveClient approves a pending client registration and assigns roles.
func (s *VerificationServiceImpl) ApproveClient(ctx context.Context, tenantID, unverifiedClientID googleUuid.UUID, roleIDs []googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.Client, error) {
	// Get the unverified client.
	unverifiedClient, err := s.unverifiedClientRepo.GetByID(ctx, unverifiedClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unverified client: %w", err)
	}

	// Verify the client belongs to the specified tenant.
	if unverifiedClient.TenantID != tenantID {
		return nil, fmt.Errorf("unverified client does not belong to the specified tenant")
	}

	// Check if the registration has expired.
	if time.Now().After(unverifiedClient.ExpiresAt) {
		return nil, fmt.Errorf("client registration has expired")
	}

	// At least one role must be assigned.
	if len(roleIDs) == 0 {
		return nil, fmt.Errorf("at least one role must be assigned when approving a client")
	}

	// Verify all role IDs belong to the tenant.
	for _, roleID := range roleIDs {
		role, err := s.roleRepo.GetByID(ctx, roleID)
		if err != nil {
			return nil, fmt.Errorf("failed to get role %s: %w", roleID, err)
		}

		if role.TenantID != tenantID {
			return nil, fmt.Errorf("role %s does not belong to the specified tenant", roleID)
		}
	}

	// Create the verified client.
	client := &cryptoutilAppsTemplateServiceServerRepository.Client{
		ID:               googleUuid.New(),
		TenantID:         unverifiedClient.TenantID,
		ClientID:         unverifiedClient.ClientID,
		ClientSecretHash: unverifiedClient.ClientSecretHash,
		Active:           1,
	}
	if err := s.clientRepo.Create(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to create verified client: %w", err)
	}

	// Assign roles.
	for _, roleID := range roleIDs {
		clientRole := &cryptoutilAppsTemplateServiceServerRepository.ClientRole{
			TenantID: tenantID,
			ClientID: client.ID,
			RoleID:   roleID,
		}
		if err := s.clientRoleRepo.Assign(ctx, clientRole); err != nil {
			return nil, fmt.Errorf("failed to assign role %s: %w", roleID, err)
		}
	}

	// Delete the unverified client record.
	if err := s.unverifiedClientRepo.Delete(ctx, unverifiedClientID); err != nil {
		// Log but don't fail - client is already created.
		// In production, this should use a transaction.
		return client, fmt.Errorf("client approved but failed to delete unverified record: %w", err)
	}

	return client, nil
}

// RejectUser rejects a pending user registration (removes from unverified table).
func (s *VerificationServiceImpl) RejectUser(ctx context.Context, tenantID, unverifiedUserID googleUuid.UUID) error {
	// Get the unverified user to verify tenant ownership.
	unverifiedUser, err := s.unverifiedUserRepo.GetByID(ctx, unverifiedUserID)
	if err != nil {
		return fmt.Errorf("failed to get unverified user: %w", err)
	}

	// Verify the user belongs to the specified tenant.
	if unverifiedUser.TenantID != tenantID {
		return fmt.Errorf("unverified user does not belong to the specified tenant")
	}

	// Delete the unverified user record.
	if err := s.unverifiedUserRepo.Delete(ctx, unverifiedUserID); err != nil {
		return fmt.Errorf("failed to reject user registration: %w", err)
	}

	return nil
}

// RejectClient rejects a pending client registration (removes from unverified table).
func (s *VerificationServiceImpl) RejectClient(ctx context.Context, tenantID, unverifiedClientID googleUuid.UUID) error {
	// Get the unverified client to verify tenant ownership.
	unverifiedClient, err := s.unverifiedClientRepo.GetByID(ctx, unverifiedClientID)
	if err != nil {
		return fmt.Errorf("failed to get unverified client: %w", err)
	}

	// Verify the client belongs to the specified tenant.
	if unverifiedClient.TenantID != tenantID {
		return fmt.Errorf("unverified client does not belong to the specified tenant")
	}

	// Delete the unverified client record.
	if err := s.unverifiedClientRepo.Delete(ctx, unverifiedClientID); err != nil {
		return fmt.Errorf("failed to reject client registration: %w", err)
	}

	return nil
}

// CleanupExpiredRegistrations removes all expired unverified user and client registrations.
func (s *VerificationServiceImpl) CleanupExpiredRegistrations(ctx context.Context) error {
	// Cleanup expired user registrations.
	if _, err := s.unverifiedUserRepo.DeleteExpired(ctx); err != nil {
		return fmt.Errorf("failed to cleanup expired user registrations: %w", err)
	}

	// Cleanup expired client registrations.
	if _, err := s.unverifiedClientRepo.DeleteExpired(ctx); err != nil {
		return fmt.Errorf("failed to cleanup expired client registrations: %w", err)
	}

	return nil
}
