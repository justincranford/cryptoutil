// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	json "encoding/json"
	"encoding/pem"
	http "net/http"
	"net/http/httptest"
	"testing"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"

	fiber "github.com/gofiber/fiber/v2"
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

const pkiProductionRolloverAuthorityID = "rollover-test-authority"

func TestNewPKIComponentsProductionRollingRolloverPreservesHistoricalAuthority(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(
		&cryptoutilCAStorage.StoredCertificate{},
		&cryptoutilCAStorage.StoredAuthority{},
		&cryptoutilCAStorage.StoredCRLSequence{},
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

	oldInstanceApp := fiber.New()
	require.NoError(t, components.RegisterPublicRevocationEndpoints(oldInstanceApp))
	require.NoError(t, RegisterPKIRoutes(
		oldInstanceApp.Group(cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath),
		oldInstanceApp.Group(cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath),
		components,
	))

	rolloverConfig := DefaultPKIAuthorityConfig()
	rolloverConfig.ID = pkiProductionRolloverAuthorityID

	rolledOverComponents, err := NewPKIComponentsWithAuthorityConfig(ctx, db, barrierService, rolloverConfig)
	require.NoError(t, err)
	require.Equal(t, pkiProductionRolloverAuthorityID, rolledOverComponents.issuer.GetCAConfig().AuthorityID)
	require.NotEqual(t, components.issuer.GetCAConfig().Certificate.Raw, rolledOverComponents.issuer.GetCAConfig().Certificate.Raw)
	require.Equal(t, DefaultPKIAuthorityConfig().ID, components.issuer.GetCAConfig().AuthorityID)
	require.Contains(t, rolledOverComponents.publicRevocationEndpoints, DefaultPKIAuthorityConfig().ID)
	require.Contains(t, rolledOverComponents.publicRevocationEndpoints, pkiProductionRolloverAuthorityID)

	authorityStore, err := cryptoutilCAStorage.NewGORMAuthorityStore(db)
	require.NoError(t, err)

	persistedAuthorities, err := authorityStore.List(ctx)
	require.NoError(t, err)
	require.Len(t, persistedAuthorities, 2)

	app := fiber.New()
	require.NoError(t, rolledOverComponents.RegisterPublicRevocationEndpoints(app))
	require.NoError(t, RegisterPKIRoutes(
		app.Group(cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath),
		app.Group(cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath),
		rolledOverComponents,
	))

	for _, authorityID := range []string{DefaultPKIAuthorityConfig().ID, pkiProductionRolloverAuthorityID} {
		request := httptest.NewRequest(http.MethodGet, pkiPublicTestAuthorityPath+authorityID+"/crl", nil)
		response, requestErr := app.Test(request, -1)
		require.NoError(t, requestErr)
		require.Equal(t, fiber.StatusOK, response.StatusCode)
		require.NoError(t, response.Body.Close())
	}

	for _, tc := range []struct {
		name              string
		app               *fiber.App
		activeCAName      string
		activeCertificate []byte
	}{
		{
			name:              "old instance remains on original authority",
			app:               oldInstanceApp,
			activeCAName:      components.issuer.GetCAConfig().Name,
			activeCertificate: components.issuer.GetCAConfig().Certificate.Raw,
		},
		{
			name:              "replacement instance serves replacement authority",
			app:               app,
			activeCAName:      rolledOverComponents.issuer.GetCAConfig().Name,
			activeCertificate: rolledOverComponents.issuer.GetCAConfig().Certificate.Raw,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, pathPrefix := range []string{
				cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath,
				cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath,
			} {
				request := httptest.NewRequest(
					http.MethodGet,
					pathPrefix+"/pki/cas/"+tc.activeCAName,
					nil,
				)
				response, requestErr := tc.app.Test(request, -1)
				require.NoError(t, requestErr)
				require.Equal(t, fiber.StatusOK, response.StatusCode)

				var ca cryptoutilApiCaServer.CAResponse
				require.NoError(t, json.NewDecoder(response.Body).Decode(&ca))
				require.NoError(t, response.Body.Close())

				certificatePEM, _ := pem.Decode([]byte(ca.CertificatePEM))
				require.NotNil(t, certificatePEM)
				require.Equal(t, tc.activeCertificate, certificatePEM.Bytes)
			}
		})
	}

	oldInstanceRequest := httptest.NewRequest(
		http.MethodGet,
		pkiPublicTestAuthorityPath+DefaultPKIAuthorityConfig().ID+"/crl",
		nil,
	)
	oldInstanceResponse, err := oldInstanceApp.Test(oldInstanceRequest, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, oldInstanceResponse.StatusCode)
	require.NoError(t, oldInstanceResponse.Body.Close())

	newAuthorityRequest := httptest.NewRequest(
		http.MethodGet,
		pkiPublicTestAuthorityPath+pkiProductionRolloverAuthorityID+"/crl",
		nil,
	)
	oldInstanceResponse, err = oldInstanceApp.Test(newAuthorityRequest, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, oldInstanceResponse.StatusCode)
	require.NoError(t, oldInstanceResponse.Body.Close())
}
