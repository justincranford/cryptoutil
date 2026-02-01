// Copyright (c) 2025 Justin Cranford
//
// Unit tests for JOSE-JA public server validation.
package server

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsJoseJaRepository "cryptoutil/internal/apps/jose/ja/repository"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilAppsTemplateServiceServerService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
)

// mockRealmService implements cryptoutilTemplateService.RealmService for testing.
type mockRealmService struct{}

func (m *mockRealmService) CreateRealm(_ context.Context, _ googleUuid.UUID, _ string, _ cryptoutilAppsTemplateServiceServerService.RealmConfig) (*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error) {
	return nil, nil
}

func (m *mockRealmService) GetRealm(_ context.Context, _, _ googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error) {
	return nil, nil
}

func (m *mockRealmService) ListRealms(_ context.Context, _ googleUuid.UUID, _ bool) ([]*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error) {
	return nil, nil
}

func (m *mockRealmService) UpdateRealm(_ context.Context, _, _ googleUuid.UUID, _ cryptoutilAppsTemplateServiceServerService.RealmConfig, _ *bool) (*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error) {
	return nil, nil
}

func (m *mockRealmService) DeleteRealm(_ context.Context, _, _ googleUuid.UUID) error {
	return nil
}

func (m *mockRealmService) GetRealmConfig(_ context.Context, _, _ googleUuid.UUID) (cryptoutilAppsTemplateServiceServerService.RealmConfig, error) {
	return nil, nil
}

func (m *mockRealmService) GetFirstActiveRealm(_ context.Context, _ googleUuid.UUID) (*cryptoutilAppsTemplateServiceServerRepository.TenantRealm, error) {
	return nil, nil
}

// newMockRealmService creates a mock realm service for testing.
func newMockRealmService() cryptoutilAppsTemplateServiceServerService.RealmService {
	return &mockRealmService{}
}

func TestNewPublicServer_NilBase(t *testing.T) {
	t.Parallel()

	// Call with nil base - only testing the first nil check.
	_, err := NewPublicServer(
		nil, // base is nil
		&cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService{},
		newMockRealmService(),
		cryptoutilAppsJoseJaRepository.NewElasticJWKRepository(nil),
		cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(nil),
		cryptoutilAppsJoseJaRepository.NewAuditConfigRepository(nil),
		cryptoutilAppsJoseJaRepository.NewAuditLogRepository(nil),
		&cryptoutilSharedCryptoJose.JWKGenService{},
		&cryptoutilAppsTemplateServiceServerBarrier.Service{},
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "public server base cannot be nil")
}

func TestNewPublicServer_NilSessionManager(t *testing.T) {
	t.Parallel()

	// Create minimal base.
	base := &cryptoutilAppsTemplateServiceServer.PublicServerBase{}

	// Call with nil session manager - testing second nil check.
	_, err := NewPublicServer(
		base,
		nil, // session manager is nil
		newMockRealmService(),
		cryptoutilAppsJoseJaRepository.NewElasticJWKRepository(nil),
		cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(nil),
		cryptoutilAppsJoseJaRepository.NewAuditConfigRepository(nil),
		cryptoutilAppsJoseJaRepository.NewAuditLogRepository(nil),
		&cryptoutilSharedCryptoJose.JWKGenService{},
		&cryptoutilAppsTemplateServiceServerBarrier.Service{},
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "session manager service cannot be nil")
}

func TestNewPublicServer_NilRealmService(t *testing.T) {
	t.Parallel()

	// Create minimal base.
	base := &cryptoutilAppsTemplateServiceServer.PublicServerBase{}

	// Call with nil realm service - testing third nil check.
	_, err := NewPublicServer(
		base,
		&cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService{},
		nil, // realm service is nil
		cryptoutilAppsJoseJaRepository.NewElasticJWKRepository(nil),
		cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(nil),
		cryptoutilAppsJoseJaRepository.NewAuditConfigRepository(nil),
		cryptoutilAppsJoseJaRepository.NewAuditLogRepository(nil),
		&cryptoutilSharedCryptoJose.JWKGenService{},
		&cryptoutilAppsTemplateServiceServerBarrier.Service{},
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "realm service cannot be nil")
}

func TestNewPublicServer_NilElasticJWKRepo(t *testing.T) {
	t.Parallel()

	// Create minimal base.
	base := &cryptoutilAppsTemplateServiceServer.PublicServerBase{}

	// Call with nil elastic JWK repository.
	_, err := NewPublicServer(
		base,
		&cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService{},
		newMockRealmService(),
		nil, // elastic JWK repo is nil
		cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(nil),
		cryptoutilAppsJoseJaRepository.NewAuditConfigRepository(nil),
		cryptoutilAppsJoseJaRepository.NewAuditLogRepository(nil),
		&cryptoutilSharedCryptoJose.JWKGenService{},
		&cryptoutilAppsTemplateServiceServerBarrier.Service{},
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "elastic JWK repository cannot be nil")
}

func TestNewPublicServer_NilMaterialJWKRepo(t *testing.T) {
	t.Parallel()

	// Create minimal base.
	base := &cryptoutilAppsTemplateServiceServer.PublicServerBase{}

	// Call with nil material JWK repository.
	_, err := NewPublicServer(
		base,
		&cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService{},
		newMockRealmService(),
		cryptoutilAppsJoseJaRepository.NewElasticJWKRepository(nil),
		nil, // material JWK repo is nil
		cryptoutilAppsJoseJaRepository.NewAuditConfigRepository(nil),
		cryptoutilAppsJoseJaRepository.NewAuditLogRepository(nil),
		&cryptoutilSharedCryptoJose.JWKGenService{},
		&cryptoutilAppsTemplateServiceServerBarrier.Service{},
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "material JWK repository cannot be nil")
}

func TestNewPublicServer_NilAuditConfigRepo(t *testing.T) {
	t.Parallel()

	// Create minimal base.
	base := &cryptoutilAppsTemplateServiceServer.PublicServerBase{}

	// Call with nil audit config repository.
	_, err := NewPublicServer(
		base,
		&cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService{},
		newMockRealmService(),
		cryptoutilAppsJoseJaRepository.NewElasticJWKRepository(nil),
		cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(nil),
		nil, // audit config repo is nil
		cryptoutilAppsJoseJaRepository.NewAuditLogRepository(nil),
		&cryptoutilSharedCryptoJose.JWKGenService{},
		&cryptoutilAppsTemplateServiceServerBarrier.Service{},
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "audit config repository cannot be nil")
}

func TestNewPublicServer_NilAuditLogRepo(t *testing.T) {
	t.Parallel()

	// Create minimal base.
	base := &cryptoutilAppsTemplateServiceServer.PublicServerBase{}

	// Call with nil audit log repository.
	_, err := NewPublicServer(
		base,
		&cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService{},
		newMockRealmService(),
		cryptoutilAppsJoseJaRepository.NewElasticJWKRepository(nil),
		cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(nil),
		cryptoutilAppsJoseJaRepository.NewAuditConfigRepository(nil),
		nil, // audit log repo is nil
		&cryptoutilSharedCryptoJose.JWKGenService{},
		&cryptoutilAppsTemplateServiceServerBarrier.Service{},
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "audit log repository cannot be nil")
}

func TestNewPublicServer_NilJWKGenService(t *testing.T) {
	t.Parallel()

	// Create minimal base.
	base := &cryptoutilAppsTemplateServiceServer.PublicServerBase{}

	// Call with nil JWK generation service.
	_, err := NewPublicServer(
		base,
		&cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService{},
		newMockRealmService(),
		cryptoutilAppsJoseJaRepository.NewElasticJWKRepository(nil),
		cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(nil),
		cryptoutilAppsJoseJaRepository.NewAuditConfigRepository(nil),
		cryptoutilAppsJoseJaRepository.NewAuditLogRepository(nil),
		nil, // JWK gen service is nil
		&cryptoutilAppsTemplateServiceServerBarrier.Service{},
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "JWK generation service cannot be nil")
}

func TestNewPublicServer_NilBarrierService(t *testing.T) {
	t.Parallel()

	// Create minimal base.
	base := &cryptoutilAppsTemplateServiceServer.PublicServerBase{}

	// Call with nil barrier service.
	_, err := NewPublicServer(
		base,
		&cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService{},
		newMockRealmService(),
		cryptoutilAppsJoseJaRepository.NewElasticJWKRepository(nil),
		cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(nil),
		cryptoutilAppsJoseJaRepository.NewAuditConfigRepository(nil),
		cryptoutilAppsJoseJaRepository.NewAuditLogRepository(nil),
		&cryptoutilSharedCryptoJose.JWKGenService{},
		nil, // barrier service is nil
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "barrier service cannot be nil")
}
