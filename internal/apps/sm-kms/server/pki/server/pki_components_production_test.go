// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"testing"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilAppsFrameworkServiceUnsealKeys "cryptoutil/internal/apps-framework/service/server/barrier/unsealkeysservice"
	cryptoutilAppsFrameworkTestDB "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

func TestNewPKIComponentsProductionDependencies(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(
		&cryptoutilCAStorage.StoredCertificate{},
		&cryptoutilCAStorage.StoredAuthority{},
		&cryptoutilAppsFrameworkServiceServerBarrier.RootKey{},
		&cryptoutilAppsFrameworkServiceServerBarrier.IntermediateKey{},
		&cryptoutilAppsFrameworkServiceServerBarrier.ContentKey{},
	))

	telemetrySettings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, telemetrySettings.ToTelemetrySettings())
	require.NoError(t, err)
	t.Cleanup(telemetryService.Shutdown)

	jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, false)
	require.NoError(t, err)
	t.Cleanup(jwkGenService.Shutdown)

	_, unsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(
		&cryptoutilSharedCryptoJose.EncA256GCM,
		&cryptoutilSharedCryptoJose.AlgA256KW,
	)
	require.NoError(t, err)

	unsealKeysService, err := cryptoutilAppsFrameworkServiceUnsealKeys.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(unsealKeysService.Shutdown)

	barrierRepository, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(barrierRepository.Shutdown)

	barrierService, err := cryptoutilAppsFrameworkServiceServerBarrier.NewService(
		ctx,
		telemetryService,
		jwkGenService,
		barrierRepository,
		unsealKeysService,
	)
	require.NoError(t, err)
	t.Cleanup(barrierService.Shutdown)

	components, err := NewPKIComponents(ctx, db, barrierService)
	require.NoError(t, err)
	require.NotNil(t, components)
	require.NotNil(t, components.issuer)
	require.NotNil(t, components.storage)
	require.NotNil(t, components.handler)
	require.Equal(t, DefaultPKIAuthorityConfig().ID, components.issuer.GetCAConfig().AuthorityID)
}
