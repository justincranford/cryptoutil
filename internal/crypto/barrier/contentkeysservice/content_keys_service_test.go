package contentkeysservice

import (
	"context"
	"os"
	"testing"

	cryptoutilIntermediateKeysService "cryptoutil/internal/crypto/barrier/intermediatekeysservice"
	cryptoutilRootKeysService "cryptoutil/internal/crypto/barrier/rootkeysservice"
	cryptoutilUnsealKeysService "cryptoutil/internal/crypto/barrier/unsealkeysservice"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/repository/sqlrepository"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var (
	testCtx                     = context.Background()
	testTelemetryService        *cryptoutilTelemetry.TelemetryService
	testSqlRepository           *cryptoutilSqlRepository.SqlRepository
	testOrmRepository           *cryptoutilOrmRepository.OrmRepository
	testDbType                  = cryptoutilSqlRepository.DBTypeSQLite // Caution: modernc.org/sqlite doesn't support read-only transactions, but PostgreSQL does
	testAes256KeyGenPool        *cryptoutilKeygen.KeyGenPool
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

	testAes256KeyGenPool = cryptoutilKeygen.RequireNewAes256GenKeyPoolForTest(testTelemetryService)
	defer testAes256KeyGenPool.Close()

	testRootKeysService = cryptoutilRootKeysService.RequireNewForTest(testTelemetryService, testOrmRepository, unsealKeysService, testAes256KeyGenPool)
	defer testRootKeysService.Shutdown()

	testIntermediateKeysService = cryptoutilIntermediateKeysService.RequireNewForTest(testTelemetryService, testOrmRepository, testRootKeysService, testAes256KeyGenPool)
	defer testIntermediateKeysService.Shutdown()

	os.Exit(m.Run())
}

func TestContentKeysService_HappyPath(t *testing.T) {
	contentKeysService, err := NewContentKeysService(testTelemetryService, testOrmRepository, testIntermediateKeysService, testAes256KeyGenPool)
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
