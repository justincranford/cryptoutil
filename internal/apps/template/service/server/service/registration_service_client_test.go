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

func TestRegistrationService_RegisterUser_ExistingTenant(t *testing.T) {
	t.Parallel()

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
				require.True(t, result.ExpiresAt.After(time.Now().UTC()))
			}
		})
	}
}

func TestRegistrationService_RegisterUser_ValidationErrors(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
