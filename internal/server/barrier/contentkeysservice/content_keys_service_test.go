package contentkeysservice

import (
	"context"
	"os"
	"testing"

	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilKeyGenPoolTest "cryptoutil/internal/common/crypto/keygenpooltest"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilIntermediateKeysService "cryptoutil/internal/server/barrier/intermediatekeysservice"
	cryptoutilRootKeysService "cryptoutil/internal/server/barrier/rootkeysservice"
	cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/server/repository/sqlrepository"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var (
	testCtx                     = context.Background()
	testTelemetryService        *cryptoutilTelemetry.TelemetryService
	testSqlRepository           *cryptoutilSqlRepository.SqlRepository
	testOrmRepository           *cryptoutilOrmRepository.OrmRepository
	testDbType                  = cryptoutilSqlRepository.DBTypeSQLite // Caution: modernc.org/sqlite doesn't support read-only transactions, but PostgreSQL does
	testUuidV7KeyGenPool        *cryptoutilPool.ValueGenPool[*googleUuid.UUID]
	testAes256KeyGenPool        *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey]
	testRootKeysService         *cryptoutilRootKeysService.RootKeysService
	testIntermediateKeysService *cryptoutilIntermediateKeysService.IntermediateKeysService
)

func TestMain(m *testing.M) {
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "content_keys_service_test", false, false)
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

	testIntermediateKeysService = cryptoutilIntermediateKeysService.RequireNewForTest(testTelemetryService, testOrmRepository, testRootKeysService, testUuidV7KeyGenPool, testAes256KeyGenPool)
	defer testIntermediateKeysService.Shutdown()

	os.Exit(m.Run())
}

func TestContentKeysService_HappyPath(t *testing.T) {
	contentKeysService, err := NewContentKeysService(testTelemetryService, testOrmRepository, testIntermediateKeysService, testUuidV7KeyGenPool, testAes256KeyGenPool)
	require.NoError(t, err)
	require.NotNil(t, contentKeysService)
	defer contentKeysService.Shutdown()

	clearBytes := []byte("Hello World")

	var encrypted []byte
	var contentKeyKidUuid *googleUuid.UUID
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		encrypted, contentKeyKidUuid, err = contentKeysService.EncryptContent(sqlTransaction, clearBytes)
		require.NoError(t, err)
		require.NotNil(t, encrypted)
		require.NotNil(t, contentKeyKidUuid)
		return err
	})
	require.NoError(t, err)
	require.NotNil(t, encrypted)
	require.NotNil(t, contentKeyKidUuid)

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
