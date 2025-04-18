package contentkeysservice

import (
	"context"
	"os"
	"testing"

	cryptoutilIntermediateKeysService "cryptoutil/internal/crypto/barrier/intermediatekeysservice"
	cryptoutilRootKeysService "cryptoutil/internal/crypto/barrier/rootkeysservice"
	cryptoutilUnsealRepository "cryptoutil/internal/crypto/barrier/unsealrepository"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/repository/sqlrepository"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	"github.com/stretchr/testify/assert"
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
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "contentkeysservice_test", false, false)
	defer testTelemetryService.Shutdown()

	testSqlRepository = cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testDbType)
	defer testSqlRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSqlRepository, true)
	defer testOrmRepository.Shutdown()

	unsealRepository := cryptoutilUnsealRepository.RequireNewForTest()
	defer unsealRepository.Shutdown()

	testAes256KeyGenPool = cryptoutilKeygen.RequireNewAes256GenKeyPoolForTest(testTelemetryService)
	defer testAes256KeyGenPool.Close()

	testRootKeysService = cryptoutilRootKeysService.RequireNewForTest(testTelemetryService, testOrmRepository, unsealRepository, testAes256KeyGenPool)
	defer testRootKeysService.Shutdown()

	testIntermediateKeysService = cryptoutilIntermediateKeysService.RequireNewForTest(testTelemetryService, testOrmRepository, testRootKeysService, testAes256KeyGenPool)
	defer testIntermediateKeysService.Shutdown()

	os.Exit(m.Run())
}

func TestContentKeysService_HappyPath(t *testing.T) {
	contentKeysService, err := NewContentKeysService(testTelemetryService, testOrmRepository, testIntermediateKeysService, testAes256KeyGenPool)
	assert.NoError(t, err)
	assert.NotNil(t, contentKeysService)
	defer contentKeysService.Shutdown()

	clearBytes := []byte("Hello World")

	var encrypted []byte
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		encrypted, err = contentKeysService.EncryptContent(sqlTransaction, clearBytes)
		assert.NoError(t, err)
		assert.NotNil(t, encrypted)
		return err
	})
	assert.NoError(t, err)
	assert.NotNil(t, encrypted)

	var decrypted []byte
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		decrypted, err = contentKeysService.DecryptContent(sqlTransaction, encrypted)
		assert.NoError(t, err)
		assert.NotNil(t, decrypted)
		return err
	})
	assert.NoError(t, err)
	assert.NotNil(t, decrypted)

	assert.Equal(t, clearBytes, decrypted)
}
