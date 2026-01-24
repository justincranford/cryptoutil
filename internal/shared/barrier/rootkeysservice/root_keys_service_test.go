// Copyright (c) 2025 Justin Cranford
//
//

package rootkeysservice

import (
	"context"
	"os"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

var (
	testSettings         = cryptoutilAppsTemplateServiceConfig.RequireNewForTest("root_keys_service_test")
	testCtx              = context.Background()
	testTelemetryService *cryptoutilSharedTelemetry.TelemetryService
	testJWKGenService    *cryptoutilSharedCryptoJose.JWKGenService
	testSQLRepository    *cryptoutilSQLRepository.SQLRepository
	testOrmRepository    *cryptoutilOrmRepository.OrmRepository
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		testTelemetryService = cryptoutilSharedTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		testJWKGenService = cryptoutilSharedCryptoJose.RequireNewForTest(testCtx, testTelemetryService)
		defer testJWKGenService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestRootKeysService_HappyPath_OneUnsealJWKs(t *testing.T) {
	testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	defer testOrmRepository.Shutdown()

	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, unsealJWK)

	unsealKeysServiceSimple, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	require.NotNil(t, unsealKeysServiceSimple)

	defer unsealKeysServiceSimple.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, testOrmRepository, unsealKeysServiceSimple)
	require.NoError(t, err)
	require.NotNil(t, rootKeysService)

	defer rootKeysService.Shutdown()
}

func TestRootKeysService_SadPath_ZeroUnsealJWKs(t *testing.T) {
	unsealKeysServiceSimple := cryptoutilUnsealKeysService.RequireNewSimpleForTest([]joseJwk.Key{})

	require.NotNil(t, unsealKeysServiceSimple)
	defer unsealKeysServiceSimple.Shutdown()

	testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	defer testOrmRepository.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, testOrmRepository, unsealKeysServiceSimple)
	require.Error(t, err)
	require.Nil(t, rootKeysService)
	require.EqualError(t, err, "failed to initialize first root JWK: failed to encrypt first root JWK: failed to encrypt root JWK with unseal JWK: invalid JWKs: jwks can't be empty")
}

func TestRootKeysService_SadPath_NilUnsealJWKs(t *testing.T) {
	unsealKeysServiceSimple := cryptoutilUnsealKeysService.RequireNewSimpleForTest(nil)

	require.NotNil(t, unsealKeysServiceSimple)
	defer unsealKeysServiceSimple.Shutdown()

	testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	defer testOrmRepository.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, testOrmRepository, unsealKeysServiceSimple)
	require.Error(t, err)
	require.Nil(t, rootKeysService)
	require.EqualError(t, err, "failed to initialize first root JWK: failed to encrypt first root JWK: failed to encrypt root JWK with unseal JWK: invalid JWKs: jwks can't be nil")
}
