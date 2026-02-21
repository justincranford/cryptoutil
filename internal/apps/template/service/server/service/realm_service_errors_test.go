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
	getByRealmIDErr    error
	createErr          error
	updateErr          error
	deleteErr          error
	listByTenantErr    error
	listByTenantRealms []*cryptoutilAppsTemplateServiceServerRepository.TenantRealm
	realm              *cryptoutilAppsTemplateServiceServerRepository.TenantRealm
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

func (m *mockRealmRepoWithErrors) ListByTenant(_ context.Context, _ googleUuid.UUID, _ bool) ([]*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error) {
	if m.listByTenantErr != nil {
		return nil, m.listByTenantErr
	}

	return m.listByTenantRealms, nil
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

// --- Helper types for realm service additional coverage ---

// badJSONRealmConfig implements RealmConfig but fails json.Marshal.
type badJSONRealmConfig struct{}

func (b *badJSONRealmConfig) GetType() RealmType { return RealmTypeBasicClientIDSecret }

func (b *badJSONRealmConfig) Validate() error { return nil }

func (b *badJSONRealmConfig) MarshalJSON() ([]byte, error) {
	return nil, errBadJSONMarshal
}

var errBadJSONMarshal = errors.New("test: marshal error for realm config")

// TestCreateRealm_RepoError tests CreateRealm when the repo Create fails.
// Covers realm_service_impl.go:192-194.
func TestCreateRealm_RepoError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("create realm db error")
	tenantID := googleUuid.New()

	mockRepo := &mockRealmRepoWithErrors{createErr: expectedErr}
	svc := NewRealmService(mockRepo)

	_, err := svc.CreateRealm(context.Background(), tenantID, string(RealmTypeBasicClientIDSecret), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create realm")
}

// TestCreateRealm_JsonMarshalError tests CreateRealm when json.Marshal fails.
// Covers realm_service_impl.go:175-177.
func TestCreateRealm_JsonMarshalError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	mockRepo := &mockRealmRepoWithErrors{}
	svc := NewRealmService(mockRepo)

	_, err := svc.CreateRealm(context.Background(), tenantID, string(RealmTypeBasicClientIDSecret), &badJSONRealmConfig{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to serialize realm configuration")
}

// TestListRealms_RepoError tests ListRealms when the repo ListByTenant fails.
// Covers realm_service_impl.go:217-219.
func TestListRealms_RepoError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("list realms db error")
	tenantID := googleUuid.New()

	mockRepo := &mockRealmRepoWithErrors{listByTenantErr: expectedErr}
	svc := NewRealmService(mockRepo)

	_, err := svc.ListRealms(context.Background(), tenantID, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list realms")
}

// TestGetFirstActiveRealm_ListError tests GetFirstActiveRealm when ListByTenant fails.
// Covers realm_service_impl.go:229-231.
func TestGetFirstActiveRealm_ListError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("list active realms db error")
	tenantID := googleUuid.New()

	mockRepo := &mockRealmRepoWithErrors{listByTenantErr: expectedErr}
	svc := NewRealmService(mockRepo)

	_, err := svc.GetFirstActiveRealm(context.Background(), tenantID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list realms")
}

// TestGetFirstActiveRealm_EmptyList tests GetFirstActiveRealm when no active realms exist.
// Covers realm_service_impl.go:234-236.
func TestGetFirstActiveRealm_EmptyList(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	mockRepo := &mockRealmRepoWithErrors{listByTenantRealms: []*cryptoutilAppsTemplateServiceServerRepository.TenantRealm{}}
	svc := NewRealmService(mockRepo)

	realm, err := svc.GetFirstActiveRealm(context.Background(), tenantID)
	require.Nil(t, realm)
	require.ErrorIs(t, err, ErrNoActiveRealm)
}

// TestGetFirstActiveRealm_Success tests GetFirstActiveRealm when an active realm exists.
// Covers realm_service_impl.go:239.
func TestGetFirstActiveRealm_Success(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	activeRealm := &cryptoutilAppsTemplateServiceServerRepository.TenantRealm{
		ID:       googleUuid.New(),
		TenantID: tenantID,
		RealmID:  googleUuid.New(),
		Active:   true,
	}

	mockRepo := &mockRealmRepoWithErrors{
		listByTenantRealms: []*cryptoutilAppsTemplateServiceServerRepository.TenantRealm{activeRealm},
	}
	svc := NewRealmService(mockRepo)

	result, err := svc.GetFirstActiveRealm(context.Background(), tenantID)
	require.NoError(t, err)
	require.Equal(t, activeRealm, result)
}

// TestUpdateRealm_JsonMarshalError tests UpdateRealm when json.Marshal on config fails.
// Covers realm_service_impl.go:261-263.
func TestUpdateRealm_JsonMarshalError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	existingRealm := &cryptoutilAppsTemplateServiceServerRepository.TenantRealm{
		ID:       googleUuid.New(),
		TenantID: tenantID,
		RealmID:  realmID,
		Active:   true,
	}

	mockRepo := &mockRealmRepoWithErrors{realm: existingRealm}
	svc := NewRealmService(mockRepo)

	_, err := svc.UpdateRealm(context.Background(), tenantID, realmID, &badJSONRealmConfig{}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to serialize realm configuration")
}
