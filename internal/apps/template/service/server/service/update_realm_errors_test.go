//go:build !integration

package service

import (
	"context"
	"errors"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// mockRealmRepoForUpdate implements repository.TenantRealmRepository for UpdateRealm testing.
type mockRealmRepoForUpdate struct {
	realm      *cryptoutilTemplateRepository.TenantRealm
	getByIDErr error
	updateErr  error
}

func (m *mockRealmRepoForUpdate) Create(ctx context.Context, realm *cryptoutilTemplateRepository.TenantRealm) error {
	return nil
}

func (m *mockRealmRepoForUpdate) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilTemplateRepository.TenantRealm, error) {
	return nil, nil
}

func (m *mockRealmRepoForUpdate) GetByRealmID(ctx context.Context, tenantID, realmID googleUuid.UUID) (*cryptoutilTemplateRepository.TenantRealm, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}

	return m.realm, nil
}

func (m *mockRealmRepoForUpdate) ListByTenant(ctx context.Context, tenantID googleUuid.UUID, activeOnly bool) ([]*cryptoutilTemplateRepository.TenantRealm, error) {
	return nil, nil
}

func (m *mockRealmRepoForUpdate) Update(ctx context.Context, realm *cryptoutilTemplateRepository.TenantRealm) error {
	return m.updateErr
}

func (m *mockRealmRepoForUpdate) Delete(ctx context.Context, id googleUuid.UUID) error {
	return nil
}

// TestUpdateRealm_GetByRealmIDError tests UpdateRealm when GetByRealmID fails.
// Target: realm_service.go:547-549 (GetByRealmID error return).
func TestUpdateRealm_GetByRealmIDError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	expectedErr := errors.New("database error")

	mockRepo := &mockRealmRepoForUpdate{
		getByIDErr: expectedErr,
	}

	svc := &RealmServiceImpl{
		realmRepo: mockRepo,
	}

	result, err := svc.UpdateRealm(context.Background(), tenantID, realmID, nil, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get realm")
	require.Nil(t, result)
}

// TestUpdateRealm_WrongTenant tests UpdateRealm when realm belongs to different tenant.
// Target: realm_service.go:552-554 (tenant mismatch check).
func TestUpdateRealm_WrongTenant(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	differentTenantID := googleUuid.New()
	realmID := googleUuid.New()

	mockRepo := &mockRealmRepoForUpdate{
		realm: &cryptoutilTemplateRepository.TenantRealm{
			TenantID: differentTenantID, // Different tenant!
			RealmID:  realmID,
		},
	}

	svc := &RealmServiceImpl{
		realmRepo: mockRepo,
	}

	result, err := svc.UpdateRealm(context.Background(), tenantID, realmID, nil, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "realm does not belong to the specified tenant")
	require.Nil(t, result)
}

// TestUpdateRealm_InvalidConfig tests UpdateRealm when config validation fails.
// Target: realm_service.go:557-559 (config validation error).
func TestUpdateRealm_InvalidConfig(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	mockRepo := &mockRealmRepoForUpdate{
		realm: &cryptoutilTemplateRepository.TenantRealm{
			TenantID: tenantID,
			RealmID:  realmID,
		},
	}

	svc := &RealmServiceImpl{
		realmRepo: mockRepo,
	}

	// Create invalid config (MinPasswordLength < 1)
	invalidConfig := &UsernamePasswordConfig{
		MinPasswordLength: 0, // Invalid - must be at least 1!
	}

	result, err := svc.UpdateRealm(context.Background(), tenantID, realmID, invalidConfig, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid realm configuration")
	require.Nil(t, result)
}

// TestUpdateRealm_UpdateError tests UpdateRealm when Update operation fails.
// Target: realm_service.go:572-574 (Update error return).
func TestUpdateRealm_UpdateError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	expectedErr := errors.New("database update error")

	mockRepo := &mockRealmRepoForUpdate{
		realm: &cryptoutilTemplateRepository.TenantRealm{
			TenantID: tenantID,
			RealmID:  realmID,
		},
		updateErr: expectedErr,
	}

	svc := &RealmServiceImpl{
		realmRepo: mockRepo,
	}

	// Simple update with active flag
	active := false
	result, err := svc.UpdateRealm(context.Background(), tenantID, realmID, nil, &active)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to update realm")
	require.Nil(t, result)
}
