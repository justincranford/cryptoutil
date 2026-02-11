// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

//go:build !integration

package service

import (
	"context"
	"errors"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// mockRealmRepoWithErrors for error injection.
type mockRealmRepoWithErrors struct {
	getByRealmIDErr error
	createErr       error
	updateErr       error
	deleteErr       error
	realm           *cryptoutilAppsTemplateServiceServerRepository.TenantRealm
}

func (m *mockRealmRepoWithErrors) Create(ctx context.Context, realm *cryptoutilAppsTemplateServiceServerRepository.TenantRealm) error {
	return m.createErr
}

func (m *mockRealmRepoWithErrors) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error) {
	return nil, nil
}

func (m *mockRealmRepoWithErrors) GetByRealmID(ctx context.Context, tenantID, realmID googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error) {
	if m.getByRealmIDErr != nil {
		return nil, m.getByRealmIDErr
	}

	return m.realm, nil
}

func (m *mockRealmRepoWithErrors) GetByName(ctx context.Context, tenantID googleUuid.UUID, name string) (*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error) {
	return nil, nil
}

func (m *mockRealmRepoWithErrors) ListByTenant(ctx context.Context, tenantID googleUuid.UUID, activeOnly bool) ([]*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error) {
	return nil, nil
}

func (m *mockRealmRepoWithErrors) Update(ctx context.Context, realm *cryptoutilAppsTemplateServiceServerRepository.TenantRealm) error {
	return m.updateErr
}

func (m *mockRealmRepoWithErrors) Delete(ctx context.Context, id googleUuid.UUID) error {
	return m.deleteErr
}

// TestDeleteRealm_GetByRealmIDError tests DeleteRealm when GetByRealmID fails.
// Targets realm_service.go:585 (GetByRealmID error return).
func TestDeleteRealm_GetByRealmIDError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	expectedErr := errors.New("database error")

	mockRepo := &mockRealmRepoWithErrors{
		getByRealmIDErr: expectedErr,
	}

	svc := &RealmServiceImpl{
		realmRepo: mockRepo,
	}

	err := svc.DeleteRealm(context.Background(), tenantID, realmID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get realm")
}

// TestDeleteRealm_WrongTenant tests DeleteRealm when realm belongs to different tenant.
// Targets realm_service.go:589-591 (tenant mismatch check).
func TestDeleteRealm_WrongTenant(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	differentTenantID := googleUuid.New()
	realmID := googleUuid.New()

	mockRepo := &mockRealmRepoWithErrors{
		realm: &cryptoutilAppsTemplateServiceServerRepository.TenantRealm{
			TenantID: differentTenantID, // Different tenant!
			RealmID:  realmID,
		},
	}

	svc := &RealmServiceImpl{
		realmRepo: mockRepo,
	}

	err := svc.DeleteRealm(context.Background(), tenantID, realmID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong to the specified tenant")
}

// TestDeleteRealm_UpdateError tests DeleteRealm when Update fails.
// Targets realm_service.go:596 (Update error return).
func TestDeleteRealm_UpdateError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	expectedErr := errors.New("update failed")

	mockRepo := &mockRealmRepoWithErrors{
		realm: &cryptoutilAppsTemplateServiceServerRepository.TenantRealm{
			TenantID: tenantID, // Correct tenant
			RealmID:  realmID,
		},
		updateErr: expectedErr,
	}

	svc := &RealmServiceImpl{
		realmRepo: mockRepo,
	}

	err := svc.DeleteRealm(context.Background(), tenantID, realmID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to deactivate realm")
}

// TestGetRealmConfig_GetByRealmIDError tests GetRealmConfig when GetByRealmID fails.
// Targets realm_service.go:605 (GetByRealmID error return).
func TestGetRealmConfig_GetByRealmIDError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	expectedErr := errors.New("database error")

	mockRepo := &mockRealmRepoWithErrors{
		getByRealmIDErr: expectedErr,
	}

	svc := &RealmServiceImpl{
		realmRepo: mockRepo,
	}

	config, err := svc.GetRealmConfig(context.Background(), tenantID, realmID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get realm")
	require.Nil(t, config)
}

// TestGetRealmConfig_WrongTenant tests GetRealmConfig when realm belongs to different tenant.
// Targets realm_service.go:609-611 (tenant mismatch check).
func TestGetRealmConfig_WrongTenant(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	differentTenantID := googleUuid.New()
	realmID := googleUuid.New()

	mockRepo := &mockRealmRepoWithErrors{
		realm: &cryptoutilAppsTemplateServiceServerRepository.TenantRealm{
			TenantID: differentTenantID, // Different tenant!
			RealmID:  realmID,
		},
	}

	svc := &RealmServiceImpl{
		realmRepo: mockRepo,
	}

	config, err := svc.GetRealmConfig(context.Background(), tenantID, realmID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong to the specified tenant")
	require.Nil(t, config)
}

// TestGetRealm_GetByRealmIDError tests GetRealm when GetByRealmID fails.
// Targets realm_service.go:523 (GetByRealmID error return).
func TestGetRealm_GetByRealmIDError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	expectedErr := errors.New("database error")

	mockRepo := &mockRealmRepoWithErrors{
		getByRealmIDErr: expectedErr,
	}

	svc := &RealmServiceImpl{
		realmRepo: mockRepo,
	}

	realm, err := svc.GetRealm(context.Background(), tenantID, realmID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get realm")
	require.Nil(t, realm)
}

// TestGetRealm_WrongTenant tests GetRealm when realm belongs to different tenant.
// Targets realm_service.go:527-529 (tenant mismatch check).
func TestGetRealm_WrongTenant(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	differentTenantID := googleUuid.New()
	realmID := googleUuid.New()

	mockRepo := &mockRealmRepoWithErrors{
		realm: &cryptoutilAppsTemplateServiceServerRepository.TenantRealm{
			TenantID: differentTenantID, // Different tenant!
			RealmID:  realmID,
		},
	}

	svc := &RealmServiceImpl{
		realmRepo: mockRepo,
	}

	realm, err := svc.GetRealm(context.Background(), tenantID, realmID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong to the specified tenant")
	require.Nil(t, realm)
}
