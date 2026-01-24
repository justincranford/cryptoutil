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

const (
	// DefaultRegistrationExpiryHours is the default expiry duration for unverified registrations.
	DefaultRegistrationExpiryHours = 72
)

// RegistrationService defines operations for user and client registration.
type RegistrationService interface {
	// RegisterUser registers a new user (creates tenant if newTenant, or creates unverified user if existing tenant).
	RegisterUser(ctx context.Context, username, email, passwordHash string, newTenant *NewTenantInfo, existingTenantID *googleUuid.UUID) (*RegistrationResult, error)

	// RegisterClient registers a new client (creates tenant if newTenant, or creates unverified client if existing tenant).
	RegisterClient(ctx context.Context, clientID, clientSecretHash string, newTenant *NewTenantInfo, existingTenantID *googleUuid.UUID) (*RegistrationResult, error)
}

// NewTenantInfo contains information for creating a new tenant during registration.
type NewTenantInfo struct {
	Name        string
	Description string
}

// RegistrationResult contains the result of a registration operation.
type RegistrationResult struct {
	Status    RegistrationStatus
	TenantID  googleUuid.UUID
	UserID    *googleUuid.UUID // Set if user was created immediately.
	ClientID  *googleUuid.UUID // Set if client was created immediately.
	Message   string
	ExpiresAt *time.Time // Set if registration is pending verification.
}

// RegistrationStatus represents the status of a registration.
type RegistrationStatus string

const (
	// RegistrationStatusApproved indicates the registration was approved immediately.
	RegistrationStatusApproved RegistrationStatus = "approved"

	// RegistrationStatusPending indicates the registration is pending verification.
	RegistrationStatusPending RegistrationStatus = "pending"
)

// RegistrationServiceImpl implements RegistrationService.
type RegistrationServiceImpl struct {
	tenantService        TenantService
	userRepo             cryptoutilAppsTemplateServiceServerRepository.UserRepository
	clientRepo           cryptoutilAppsTemplateServiceServerRepository.ClientRepository
	unverifiedUserRepo   cryptoutilAppsTemplateServiceServerRepository.UnverifiedUserRepository
	unverifiedClientRepo cryptoutilAppsTemplateServiceServerRepository.UnverifiedClientRepository
	roleRepo             cryptoutilAppsTemplateServiceServerRepository.RoleRepository
	userRoleRepo         cryptoutilAppsTemplateServiceServerRepository.UserRoleRepository
	clientRoleRepo       cryptoutilAppsTemplateServiceServerRepository.ClientRoleRepository
}

// NewRegistrationService creates a new RegistrationService instance.
func NewRegistrationService(
	tenantService TenantService,
	userRepo cryptoutilAppsTemplateServiceServerRepository.UserRepository,
	clientRepo cryptoutilAppsTemplateServiceServerRepository.ClientRepository,
	unverifiedUserRepo cryptoutilAppsTemplateServiceServerRepository.UnverifiedUserRepository,
	unverifiedClientRepo cryptoutilAppsTemplateServiceServerRepository.UnverifiedClientRepository,
	roleRepo cryptoutilAppsTemplateServiceServerRepository.RoleRepository,
	userRoleRepo cryptoutilAppsTemplateServiceServerRepository.UserRoleRepository,
	clientRoleRepo cryptoutilAppsTemplateServiceServerRepository.ClientRoleRepository,
) RegistrationService {
	return &RegistrationServiceImpl{
		tenantService:        tenantService,
		userRepo:             userRepo,
		clientRepo:           clientRepo,
		unverifiedUserRepo:   unverifiedUserRepo,
		unverifiedClientRepo: unverifiedClientRepo,
		roleRepo:             roleRepo,
		userRoleRepo:         userRoleRepo,
		clientRoleRepo:       clientRoleRepo,
	}
}

// RegisterUser registers a new user (creates tenant if newTenant, or creates unverified user if existing tenant).
func (s *RegistrationServiceImpl) RegisterUser(ctx context.Context, username, email, passwordHash string, newTenant *NewTenantInfo, existingTenantID *googleUuid.UUID) (*RegistrationResult, error) {
	// Validate input: exactly one of newTenant or existingTenantID must be provided.
	if (newTenant == nil && existingTenantID == nil) || (newTenant != nil && existingTenantID != nil) {
		return nil, fmt.Errorf("exactly one of newTenant or existingTenantID must be provided")
	}

	// Case 1: New tenant registration (immediate approval).
	if newTenant != nil {
		// Create tenant.
		tenant, err := s.tenantService.CreateTenant(ctx, newTenant.Name, newTenant.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to create tenant: %w", err)
		}

		// Create user.
		user := &cryptoutilAppsTemplateServiceServerRepository.User{
			ID:           googleUuid.New(),
			TenantID:     tenant.ID,
			Username:     username,
			Email:        email,
			PasswordHash: passwordHash,
			Active:       1, // 1 = active.
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		// Assign admin role to user.
		adminRole, err := s.roleRepo.GetByName(ctx, tenant.ID, "admin")
		if err != nil {
			return nil, fmt.Errorf("failed to get admin role: %w", err)
		}

		userRole := &cryptoutilAppsTemplateServiceServerRepository.UserRole{
			TenantID: tenant.ID,
			UserID:   user.ID,
			RoleID:   adminRole.ID,
		}

		if err := s.userRoleRepo.Assign(ctx, userRole); err != nil {
			return nil, fmt.Errorf("failed to assign admin role: %w", err)
		}

		return &RegistrationResult{
			Status:   RegistrationStatusApproved,
			TenantID: tenant.ID,
			UserID:   &user.ID,
			Message:  "User registered successfully as tenant administrator",
		}, nil
	}

	// Case 2: Existing tenant registration (pending verification).
	expiresAt := time.Now().Add(DefaultRegistrationExpiryHours * time.Hour)
	unverifiedUser := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedUser{
		ID:           googleUuid.New(),
		TenantID:     *existingTenantID,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		ExpiresAt:    expiresAt,
	}

	if err := s.unverifiedUserRepo.Create(ctx, unverifiedUser); err != nil {
		return nil, fmt.Errorf("failed to create unverified user: %w", err)
	}

	return &RegistrationResult{
		Status:    RegistrationStatusPending,
		TenantID:  *existingTenantID,
		Message:   "User registration pending tenant administrator approval",
		ExpiresAt: &expiresAt,
	}, nil
}

// RegisterClient registers a new client (creates tenant if newTenant, or creates unverified client if existing tenant).
func (s *RegistrationServiceImpl) RegisterClient(ctx context.Context, clientID, clientSecretHash string, newTenant *NewTenantInfo, existingTenantID *googleUuid.UUID) (*RegistrationResult, error) {
	// Validate input: exactly one of newTenant or existingTenantID must be provided.
	if (newTenant == nil && existingTenantID == nil) || (newTenant != nil && existingTenantID != nil) {
		return nil, fmt.Errorf("exactly one of newTenant or existingTenantID must be provided")
	}

	// Case 1: New tenant registration (immediate approval).
	if newTenant != nil {
		// Create tenant.
		tenant, err := s.tenantService.CreateTenant(ctx, newTenant.Name, newTenant.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to create tenant: %w", err)
		}

		// Create client.
		client := &cryptoutilAppsTemplateServiceServerRepository.Client{
			ID:               googleUuid.New(),
			TenantID:         tenant.ID,
			ClientID:         clientID,
			ClientSecretHash: clientSecretHash,
			Active:           1, // 1 = active.
		}

		if err := s.clientRepo.Create(ctx, client); err != nil {
			return nil, fmt.Errorf("failed to create client: %w", err)
		}

		// Assign admin role to client.
		adminRole, err := s.roleRepo.GetByName(ctx, tenant.ID, "admin")
		if err != nil {
			return nil, fmt.Errorf("failed to get admin role: %w", err)
		}

		clientRole := &cryptoutilAppsTemplateServiceServerRepository.ClientRole{
			TenantID: tenant.ID,
			ClientID: client.ID,
			RoleID:   adminRole.ID,
		}

		if err := s.clientRoleRepo.Assign(ctx, clientRole); err != nil {
			return nil, fmt.Errorf("failed to assign admin role: %w", err)
		}

		return &RegistrationResult{
			Status:   RegistrationStatusApproved,
			TenantID: tenant.ID,
			ClientID: &client.ID,
			Message:  "Client registered successfully as tenant administrator",
		}, nil
	}

	// Case 2: Existing tenant registration (pending verification).
	expiresAt := time.Now().Add(DefaultRegistrationExpiryHours * time.Hour)
	unverifiedClient := &cryptoutilAppsTemplateServiceServerRepository.UnverifiedClient{
		ID:               googleUuid.New(),
		TenantID:         *existingTenantID,
		ClientID:         clientID,
		ClientSecretHash: clientSecretHash,
		ExpiresAt:        expiresAt,
	}

	if err := s.unverifiedClientRepo.Create(ctx, unverifiedClient); err != nil {
		return nil, fmt.Errorf("failed to create unverified client: %w", err)
	}

	return &RegistrationResult{
		Status:    RegistrationStatusPending,
		TenantID:  *existingTenantID,
		Message:   "Client registration pending tenant administrator approval",
		ExpiresAt: &expiresAt,
	}, nil
}
