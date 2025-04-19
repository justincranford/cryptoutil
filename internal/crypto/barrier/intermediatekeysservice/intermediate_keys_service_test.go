package intermediatekeysservice

import (
	"context"
	"os"
	"testing"

	cryptoutilRootKeysService "cryptoutil/internal/crypto/barrier/rootkeysservice"
	cryptoutilUnsealKeysService "cryptoutil/internal/crypto/barrier/unsealkeysservice"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/repository/sqlrepository"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

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
	testAes256KeyGenPool *cryptoutilKeygen.KeyGenPool
	testRootKeysService  *cryptoutilRootKeysService.RootKeysService
)

func TestMain(m *testing.M) {
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "intermediate_keys_service_test", false, false)
	defer testTelemetryService.Shutdown()

	testSqlRepository = cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testDbType)
	defer testSqlRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSqlRepository, true)
	defer testOrmRepository.Shutdown()

	unsealKeysService := cryptoutilUnsealKeysService.RequireNewForTest()
	defer unsealKeysService.Shutdown()

	testAes256KeyGenPool = cryptoutilKeygen.RequireNewAes256GenKeyPoolForTest(testTelemetryService)
	defer testAes256KeyGenPool.Close()

	testRootKeysService = cryptoutilRootKeysService.RequireNewForTest(testTelemetryService, testOrmRepository, unsealKeysService, testAes256KeyGenPool)
	defer testRootKeysService.Shutdown()

	os.Exit(m.Run())
}

func TestIntermediateKeysService_HappyPath(t *testing.T) {
	intermediateKeysService, err := NewIntermediateKeysService(testTelemetryService, testOrmRepository, testRootKeysService, testAes256KeyGenPool)
	require.NoError(t, err)
	require.NotNil(t, intermediateKeysService)
	defer intermediateKeysService.Shutdown()

	clearContentKey, _, _, err := cryptoutilJose.GenerateAesJWKFromPool(cryptoutilJose.AlgDIRECT, testAes256KeyGenPool)
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
