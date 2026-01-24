// Copyright (c) 2025 Justin Cranford
//
//

package intermediatekeysservice

import (
	"context"
	"os"
	"testing"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilRootKeysService "cryptoutil/internal/shared/barrier/rootkeysservice"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

var (
	testSettings         = cryptoutilConfig.RequireNewForTest("intermediate_keys_service_test")
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
	testJWKGenService    *cryptoutilJose.JWKGenService
	testSQLRepository    *cryptoutilSQLRepository.SQLRepository
	testOrmRepository    *cryptoutilOrmRepository.OrmRepository
	testRootKeysService  *cryptoutilRootKeysService.RootKeysService
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		testJWKGenService = cryptoutilJose.RequireNewForTest(testCtx, testTelemetryService)
		defer testJWKGenService.Shutdown()

		testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
		defer testSQLRepository.Shutdown()

		testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
		defer testOrmRepository.Shutdown()

		_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
		cryptoutilSharedApperr.RequireNoError(err, "failed to generate unseal JWK for test")

		unsealKeysService := cryptoutilUnsealKeysService.RequireNewSimpleForTest([]joseJwk.Key{unsealJWK})
		defer unsealKeysService.Shutdown()

		testRootKeysService = cryptoutilRootKeysService.RequireNewForTest(testTelemetryService, testJWKGenService, testOrmRepository, unsealKeysService)
		defer testRootKeysService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestIntermediateKeysService_HappyPath(t *testing.T) {
	intermediateKeysService, err := NewIntermediateKeysService(testTelemetryService, testJWKGenService, testOrmRepository, testRootKeysService)
	require.NoError(t, err)
	require.NotNil(t, intermediateKeysService)

	defer intermediateKeysService.Shutdown()

	_, clearContentKey, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgDir)
	require.NoError(t, err)
	require.NotNil(t, clearContentKey)

	var encryptedContentKeyBytes []byte

	var intermediateKeyKidUUID *googleUuid.UUID

	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		encryptedContentKeyBytes, intermediateKeyKidUUID, err = intermediateKeysService.EncryptKey(sqlTransaction, clearContentKey)
		require.NoError(t, err)
		require.NotNil(t, encryptedContentKeyBytes)
		require.NotNil(t, intermediateKeyKidUUID)

		return err
	})
	require.NoError(t, err)
	require.NotNil(t, encryptedContentKeyBytes)
	require.NotNil(t, intermediateKeyKidUUID)

	var decryptedContentKey joseJwk.Key

	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		decryptedContentKey, err = intermediateKeysService.DecryptKey(sqlTransaction, encryptedContentKeyBytes)
		require.NoError(t, err)
		require.NotNil(t, decryptedContentKey)

		return err
	})
	require.NoError(t, err)
	require.NotNil(t, decryptedContentKey)

	require.ElementsMatch(t, clearContentKey.Keys(), decryptedContentKey.Keys())
}
