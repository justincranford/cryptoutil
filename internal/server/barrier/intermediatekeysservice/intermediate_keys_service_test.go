package intermediatekeysservice

import (
	"context"
	"os"
	"testing"

	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilKeyGenPoolTest "cryptoutil/internal/common/crypto/keygenpooltest"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilRootKeysService "cryptoutil/internal/server/barrier/rootkeysservice"
	cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/server/repository/sqlrepository"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
	testSqlRepository    *cryptoutilSqlRepository.SqlRepository
	testOrmRepository    *cryptoutilOrmRepository.OrmRepository
	testDbType           = cryptoutilSqlRepository.DBTypeSQLite // Caution: modernc.org/sqlite doesn't support read-only transactions, but PostgreSQL does
	testUuidV7KeyGenPool *cryptoutilPool.ValueGenPool[*googleUuid.UUID]
	testAes256KeyGenPool *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey]
	testRootKeysService  *cryptoutilRootKeysService.RootKeysService
)

func TestMain(m *testing.M) {
	var rc int
	func() {
		testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "intermediate_keys_service_test", false, false)
		defer testTelemetryService.Shutdown()

		testSqlRepository = cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testDbType)
		defer testSqlRepository.Shutdown()

		testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSqlRepository, true)
		defer testOrmRepository.Shutdown()

		unsealKeysService := cryptoutilUnsealKeysService.RequireNewFromSysInfoForTest()
		defer unsealKeysService.Shutdown()

		testUuidV7KeyGenPool = cryptoutilKeyGenPoolTest.RequireNewUuidV7GenKeyPoolForTest(testTelemetryService)
		defer testUuidV7KeyGenPool.Cancel()

		testAes256KeyGenPool = cryptoutilKeyGenPoolTest.RequireNewAes256GcmGenKeyPoolForTest(testTelemetryService)
		defer testAes256KeyGenPool.Cancel()

		testRootKeysService = cryptoutilRootKeysService.RequireNewForTest(testTelemetryService, testOrmRepository, unsealKeysService, testUuidV7KeyGenPool, testAes256KeyGenPool)
		defer testRootKeysService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestIntermediateKeysService_HappyPath(t *testing.T) {
	intermediateKeysService, err := NewIntermediateKeysService(testTelemetryService, testOrmRepository, testRootKeysService, testUuidV7KeyGenPool, testAes256KeyGenPool)
	require.NoError(t, err)
	require.NotNil(t, intermediateKeysService)
	defer intermediateKeysService.Shutdown()

	_, clearContentKey, _, err := cryptoutilJose.GenerateJweJwkFromSecretKeyPool(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgDir, testUuidV7KeyGenPool, testAes256KeyGenPool)
	require.NoError(t, err)
	require.NotNil(t, clearContentKey)

	var encryptedContentKeyBytes []byte
	var intermediateKeyKidUuid *googleUuid.UUID
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		encryptedContentKeyBytes, intermediateKeyKidUuid, err = intermediateKeysService.EncryptKey(sqlTransaction, clearContentKey)
		require.NoError(t, err)
		require.NotNil(t, encryptedContentKeyBytes)
		require.NotNil(t, intermediateKeyKidUuid)
		return err
	})
	require.NoError(t, err)
	require.NotNil(t, encryptedContentKeyBytes)
	require.NotNil(t, intermediateKeyKidUuid)

	var decryptedContentKey joseJwk.Key
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		decryptedContentKey, err = intermediateKeysService.DecryptKey(sqlTransaction, encryptedContentKeyBytes)
		require.NoError(t, err)
		require.NotNil(t, decryptedContentKey)
		return err
	})
	require.NoError(t, err)
	require.NotNil(t, decryptedContentKey)

	// TODO Why does Equal not work on clearContentKey <=> decryptedContentKey?
	require.Equal(t, clearContentKey.Keys(), decryptedContentKey.Keys())
}
