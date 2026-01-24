// Copyright (c) 2025 Justin Cranford
//
//

package contentkeysservice

import (
	"context"
	"os"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilIntermediateKeysService "cryptoutil/internal/shared/barrier/intermediatekeysservice"
	cryptoutilRootKeysService "cryptoutil/internal/shared/barrier/rootkeysservice"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

var (
	testSettings                = cryptoutilAppsTemplateServiceConfig.RequireNewForTest("content_keys_service_test")
	testCtx                     = context.Background()
	testTelemetryService        *cryptoutilSharedTelemetry.TelemetryService
	testJWKGenService           *cryptoutilSharedCryptoJose.JWKGenService
	testSQLRepository           *cryptoutilSQLRepository.SQLRepository
	testOrmRepository           *cryptoutilOrmRepository.OrmRepository
	testRootKeysService         *cryptoutilRootKeysService.RootKeysService
	testIntermediateKeysService *cryptoutilIntermediateKeysService.IntermediateKeysService
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		testTelemetryService = cryptoutilSharedTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		testJWKGenService = cryptoutilSharedCryptoJose.RequireNewForTest(testCtx, testTelemetryService)
		defer testJWKGenService.Shutdown()

		testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
		defer testSQLRepository.Shutdown()

		testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
		defer testOrmRepository.Shutdown()

		_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
		cryptoutilSharedApperr.RequireNoError(err, "failed to generate unseal JWK for test")

		unsealKeysService := cryptoutilUnsealKeysService.RequireNewSimpleForTest([]joseJwk.Key{unsealJWK})
		defer unsealKeysService.Shutdown()

		testRootKeysService = cryptoutilRootKeysService.RequireNewForTest(testTelemetryService, testJWKGenService, testOrmRepository, unsealKeysService)
		defer testRootKeysService.Shutdown()

		testIntermediateKeysService = cryptoutilIntermediateKeysService.RequireNewForTest(testTelemetryService, testJWKGenService, testOrmRepository, testRootKeysService)
		defer testIntermediateKeysService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestContentKeysService_HappyPath(t *testing.T) {
	contentKeysService, err := NewContentKeysService(testTelemetryService, testJWKGenService, testOrmRepository, testIntermediateKeysService)
	require.NoError(t, err)
	require.NotNil(t, contentKeysService)

	defer contentKeysService.Shutdown()

	clearBytes := []byte("Hello World")

	var encrypted []byte

	var contentKeyKidUUID *googleUuid.UUID

	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		encrypted, contentKeyKidUUID, err = contentKeysService.EncryptContent(sqlTransaction, clearBytes)
		require.NoError(t, err)
		require.NotNil(t, encrypted)
		require.NotNil(t, contentKeyKidUUID)

		return err
	})
	require.NoError(t, err)
	require.NotNil(t, encrypted)
	require.NotNil(t, contentKeyKidUUID)

	var decrypted []byte

	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		decrypted, err = contentKeysService.DecryptContent(sqlTransaction, encrypted)
		require.NoError(t, err)
		require.NotNil(t, decrypted)

		return err
	})
	require.NoError(t, err)
	require.NotNil(t, decrypted)

	require.Equal(t, clearBytes, decrypted)
}
