// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps/framework/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/framework/service/server/barrier/unsealkeysservice"
	cryptoutilAppsFrameworkServiceServerBusinesslogic "cryptoutil/internal/apps/framework/service/server/businesslogic"
	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps/framework/service/server/repository"
	cryptoutilSharedContainer "cryptoutil/internal/shared/container"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

var errInjectFailure = errors.New("injected test failure")

// injectJWKGenFail replaces newJWKGenServiceFn with a failure stub.
func injectJWKGenFail(_ context.Context, _ *cryptoutilSharedTelemetry.TelemetryService, _ bool) (*cryptoutilSharedCryptoJose.JWKGenService, error) {
	return nil, errInjectFailure
}

// injectGormRepoFail replaces newBarrierGormRepositoryFn with a failure stub.
func injectGormRepoFail(_ *gorm.DB) (*cryptoutilAppsFrameworkServiceServerBarrier.GormRepository, error) {
	return nil, errInjectFailure
}

// injectSessionManagerFail replaces newSessionManagerServiceFn with a failure stub.
func injectSessionManagerFail(_ context.Context, _ *gorm.DB, _ *cryptoutilSharedTelemetry.TelemetryService, _ *cryptoutilSharedCryptoJose.JWKGenService, _ *cryptoutilAppsFrameworkServiceServerBarrier.Service, _ *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*cryptoutilAppsFrameworkServiceServerBusinesslogic.SessionManagerService, error) {
	return nil, errInjectFailure
}

// injectRotationFail replaces newRotationServiceFn with a failure stub.
func injectRotationFail(_ *cryptoutilSharedCryptoJose.JWKGenService, _ cryptoutilAppsFrameworkServiceServerBarrier.Repository, _ cryptoutilUnsealKeysService.UnsealKeysService) (*cryptoutilAppsFrameworkServiceServerBarrier.RotationService, error) {
	return nil, errInjectFailure
}

// injectStatusFail replaces newStatusServiceFn with a failure stub.
func injectStatusFail(_ cryptoutilAppsFrameworkServiceServerBarrier.Repository) (*cryptoutilAppsFrameworkServiceServerBarrier.StatusService, error) {
	return nil, errInjectFailure
}

// setupCoreWithMigrations creates a Core with a unique in-memory SQLite DB and applies all required migrations.
func setupCoreWithMigrations(t *testing.T) (*Core, *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) {
	t.Helper()

	ctx := context.Background()

	// Use a unique in-memory SQLite DB per test to avoid cross-test contamination during parallel execution.
	uniqueName, _ := googleUuid.NewV7()
	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.DatabaseURL = fmt.Sprintf("file:test_%s?mode=memory&cache=shared", uniqueName)

	core, err := StartCore(ctx, settings)
	require.NoError(t, err)

	t.Cleanup(func() { core.Shutdown() })

	err = core.DB.AutoMigrate(
		&cryptoutilAppsFrameworkServiceServerBarrier.RootKey{},
		&cryptoutilAppsFrameworkServiceServerBarrier.IntermediateKey{},
		&cryptoutilAppsFrameworkServiceServerBarrier.ContentKey{},
		&cryptoutilAppsFrameworkServiceServerRepository.BrowserSessionJWK{},
		&cryptoutilAppsFrameworkServiceServerRepository.ServiceSessionJWK{},
		&cryptoutilAppsFrameworkServiceServerRepository.BrowserSession{},
		&cryptoutilAppsFrameworkServiceServerRepository.ServiceSession{},
	)
	require.NoError(t, err)

	return core, settings
}

// TestStartBasic_JWKGenServiceFailure tests StartBasic JWKGenService error path.
func TestStartBasic_JWKGenServiceFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	_, err := startBasicInternal(ctx, settings, injectJWKGenFail)
	require.ErrorContains(t, err, "JWK Gen Service")
}

// TestStartCore_StartBasicViaJWKGenFailure tests StartCore error path via StartBasic failure.
func TestStartCore_StartBasicViaJWKGenFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	_, err := startCoreInternal(ctx, settings, injectJWKGenFail,
		cryptoutilSharedContainer.StartPostgres,
		sql.Open,
		func(d gorm.Dialector, c *gorm.Config) (*gorm.DB, error) { return gorm.Open(d, c) },
	)
	require.ErrorContains(t, err, "basic application")
}

// TestStartListener_StartCoreViaJWKGenFailure tests StartListener error path via StartCore failure.
func TestStartListener_StartCoreViaJWKGenFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	cfg := &ListenerConfig{Settings: settings, PublicServer: &mockPublicServer{}, AdminServer: &mockAdminServer{}}

	// startCoreInternal returns error due to JWK gen failure, StartListener returns error.
	_, err := startListenerInternal(ctx, cfg,
		injectJWKGenFail,
		cryptoutilSharedContainer.StartPostgres,
		sql.Open,
		func(d gorm.Dialector, c *gorm.Config) (*gorm.DB, error) { return gorm.Open(d, c) },
	)
	require.ErrorContains(t, err, "application core")
}

// TestInitializeServicesOnCore_NewGormRepositoryFailure tests error path when NewGormRepository fails.
func TestInitializeServicesOnCore_NewGormRepositoryFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	core, err := StartCore(ctx, settings)
	require.NoError(t, err)

	defer core.Shutdown()

	_, err = initializeServicesOnCoreInternal(ctx, core, settings, injectGormRepoFail,
		cryptoutilAppsFrameworkServiceServerBusinesslogic.NewSessionManagerService,
		cryptoutilAppsFrameworkServiceServerBarrier.NewRotationService,
		cryptoutilAppsFrameworkServiceServerBarrier.NewStatusService,
	)
	require.ErrorContains(t, err, "barrier repository")
}

// TestInitializeServicesOnCore_NewSessionManagerServiceFailure tests error when NewSessionManagerService fails.
func TestInitializeServicesOnCore_NewSessionManagerServiceFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	core, settings := setupCoreWithMigrations(t)

	_, err := initializeServicesOnCoreInternal(ctx, core, settings,
		cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository,
		injectSessionManagerFail,
		cryptoutilAppsFrameworkServiceServerBarrier.NewRotationService,
		cryptoutilAppsFrameworkServiceServerBarrier.NewStatusService,
	)
	require.ErrorContains(t, err, "session manager service")
}

// TestInitializeServicesOnCore_NewRotationServiceFailure tests error when NewRotationService fails.
func TestInitializeServicesOnCore_NewRotationServiceFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	core, settings := setupCoreWithMigrations(t)

	_, err := initializeServicesOnCoreInternal(ctx, core, settings,
		cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository,
		cryptoutilAppsFrameworkServiceServerBusinesslogic.NewSessionManagerService,
		injectRotationFail,
		cryptoutilAppsFrameworkServiceServerBarrier.NewStatusService,
	)
	require.ErrorContains(t, err, "rotation service")
}

// TestInitializeServicesOnCore_NewStatusServiceFailure tests error when NewStatusService fails.
func TestInitializeServicesOnCore_NewStatusServiceFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	core, settings := setupCoreWithMigrations(t)

	_, err := initializeServicesOnCoreInternal(ctx, core, settings,
		cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository,
		cryptoutilAppsFrameworkServiceServerBusinesslogic.NewSessionManagerService,
		cryptoutilAppsFrameworkServiceServerBarrier.NewRotationService,
		injectStatusFail,
	)
	require.ErrorContains(t, err, "status service")
}
