// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
)

var errInjectFailure = errors.New("injected test failure")

// injectJWKGenFail replaces newJWKGenServiceFn with a failure stub.
func injectJWKGenFail(_ context.Context, _ *cryptoutilSharedTelemetry.TelemetryService, _ bool) (*cryptoutilSharedCryptoJose.JWKGenService, error) {
	return nil, errInjectFailure
}

// injectGormRepoFail replaces newBarrierGormRepositoryFn with a failure stub.
func injectGormRepoFail(_ *gorm.DB) (*cryptoutilAppsTemplateServiceServerBarrier.GormRepository, error) {
	return nil, errInjectFailure
}

// injectSessionManagerFail replaces newSessionManagerServiceFn with a failure stub.
func injectSessionManagerFail(_ context.Context, _ *gorm.DB, _ *cryptoutilSharedTelemetry.TelemetryService, _ *cryptoutilSharedCryptoJose.JWKGenService, _ *cryptoutilAppsTemplateServiceServerBarrier.Service, _ *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) (*cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService, error) {
	return nil, errInjectFailure
}

// injectRotationFail replaces newRotationServiceFn with a failure stub.
func injectRotationFail(_ *cryptoutilSharedCryptoJose.JWKGenService, _ cryptoutilAppsTemplateServiceServerBarrier.Repository, _ cryptoutilUnsealKeysService.UnsealKeysService) (*cryptoutilAppsTemplateServiceServerBarrier.RotationService, error) {
	return nil, errInjectFailure
}

// injectStatusFail replaces newStatusServiceFn with a failure stub.
func injectStatusFail(_ cryptoutilAppsTemplateServiceServerBarrier.Repository) (*cryptoutilAppsTemplateServiceServerBarrier.StatusService, error) {
	return nil, errInjectFailure
}

// setupCoreWithMigrations creates a Core with in-memory SQLite and applies all required migrations.
func setupCoreWithMigrations(t *testing.T) (*Core, *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) {
	t.Helper()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	core, err := StartCore(ctx, settings)
	require.NoError(t, err)

	t.Cleanup(func() { core.Shutdown() })

	err = core.DB.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerBarrier.RootKey{},
		&cryptoutilAppsTemplateServiceServerBarrier.IntermediateKey{},
		&cryptoutilAppsTemplateServiceServerBarrier.ContentKey{},
		&cryptoutilAppsTemplateServiceServerRepository.BrowserSessionJWK{},
		&cryptoutilAppsTemplateServiceServerRepository.ServiceSessionJWK{},
		&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{},
		&cryptoutilAppsTemplateServiceServerRepository.ServiceSession{},
	)
	require.NoError(t, err)

	return core, settings
}

// TestStartBasic_JWKGenServiceFailure tests StartBasic JWKGenService error path.
func TestStartBasic_JWKGenServiceFailure(t *testing.T) {
	orig := newJWKGenServiceFn
	newJWKGenServiceFn = injectJWKGenFail

	defer func() { newJWKGenServiceFn = orig }()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	_, err := StartBasic(ctx, settings)
	require.ErrorContains(t, err, "JWK Gen Service")
}

// TestStartCore_StartBasicViaJWKGenFailure tests StartCore error path via StartBasic failure.
func TestStartCore_StartBasicViaJWKGenFailure(t *testing.T) {
	orig := newJWKGenServiceFn
	newJWKGenServiceFn = injectJWKGenFail

	defer func() { newJWKGenServiceFn = orig }()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	_, err := StartCore(ctx, settings)
	require.ErrorContains(t, err, "basic application")
}

// TestStartListener_StartCoreViaJWKGenFailure tests StartListener error path via StartCore failure.
func TestStartListener_StartCoreViaJWKGenFailure(t *testing.T) {
	orig := newJWKGenServiceFn
	newJWKGenServiceFn = injectJWKGenFail

	defer func() { newJWKGenServiceFn = orig }()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	cfg := &ListenerConfig{Settings: settings, PublicServer: &mockPublicServer{}, AdminServer: &mockAdminServer{}}

	_, err := StartListener(ctx, cfg)
	require.ErrorContains(t, err, "application core")
}

// TestInitializeServicesOnCore_NewGormRepositoryFailure tests error path when NewGormRepository fails.
func TestInitializeServicesOnCore_NewGormRepositoryFailure(t *testing.T) {
	orig := newBarrierGormRepositoryFn
	newBarrierGormRepositoryFn = injectGormRepoFail

	defer func() { newBarrierGormRepositoryFn = orig }()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	core, err := StartCore(ctx, settings)
	require.NoError(t, err)

	defer core.Shutdown()

	_, err = InitializeServicesOnCore(ctx, core, settings)
	require.ErrorContains(t, err, "barrier repository")
}

// TestInitializeServicesOnCore_NewSessionManagerServiceFailure tests error when NewSessionManagerService fails.
func TestInitializeServicesOnCore_NewSessionManagerServiceFailure(t *testing.T) {
	orig := newSessionManagerServiceFn
	newSessionManagerServiceFn = injectSessionManagerFail

	defer func() { newSessionManagerServiceFn = orig }()

	ctx := context.Background()
	core, settings := setupCoreWithMigrations(t)

	_, err := InitializeServicesOnCore(ctx, core, settings)
	require.ErrorContains(t, err, "session manager service")
}

// TestInitializeServicesOnCore_NewRotationServiceFailure tests error when NewRotationService fails.
func TestInitializeServicesOnCore_NewRotationServiceFailure(t *testing.T) {
	orig := newRotationServiceFn
	newRotationServiceFn = injectRotationFail

	defer func() { newRotationServiceFn = orig }()

	ctx := context.Background()
	core, settings := setupCoreWithMigrations(t)

	_, err := InitializeServicesOnCore(ctx, core, settings)
	require.ErrorContains(t, err, "rotation service")
}

// TestInitializeServicesOnCore_NewStatusServiceFailure tests error when NewStatusService fails.
func TestInitializeServicesOnCore_NewStatusServiceFailure(t *testing.T) {
	orig := newStatusServiceFn
	newStatusServiceFn = injectStatusFail

	defer func() { newStatusServiceFn = orig }()

	ctx := context.Background()
	core, settings := setupCoreWithMigrations(t)

	_, err := InitializeServicesOnCore(ctx, core, settings)
	require.ErrorContains(t, err, "status service")
}
