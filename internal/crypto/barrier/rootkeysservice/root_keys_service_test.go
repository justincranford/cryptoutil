package rootkeysservice

import (
	"context"
	"os"
	"testing"

	cryptoutilAppErr "cryptoutil/internal/apperr"
	cryptoutilUnsealRepository "cryptoutil/internal/crypto/barrier/unsealrepository"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/repository/sqlrepository"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	"github.com/stretchr/testify/assert"
)

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
	testSqlRepository    *cryptoutilSqlRepository.SqlRepository
	testOrmRepository    *cryptoutilOrmRepository.OrmRepository
	testDbType           = cryptoutilSqlRepository.DBTypeSQLite // Caution: modernc.org/sqlite doesn't support read-only transactions, but PostgreSQL does
	testAes256KeyGenPool *cryptoutilKeygen.KeyGenPool
)

func TestMain(m *testing.M) {
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "servicelogic_test", false, false)
	defer testTelemetryService.Shutdown()

	testSqlRepository = cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testDbType)
	defer testSqlRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSqlRepository, true)
	defer testOrmRepository.Shutdown()

	keyPoolConfig, err := cryptoutilKeygen.NewKeyGenPoolConfig(context.Background(), testTelemetryService, "Test AES-256", 3, 6, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(256))
	cryptoutilAppErr.RequireNoError(err, "failed to create AES-256 pool config")
	testAes256KeyGenPool, err = cryptoutilKeygen.NewGenKeyPool(keyPoolConfig)
	cryptoutilAppErr.RequireNoError(err, "failed to create AES-256 pool")
	defer testAes256KeyGenPool.Close()

	os.Exit(m.Run())
}

func TestRootKeysService_HappyPath_OneUnsealJwks(t *testing.T) {
	mockUnsealRepository, _, err := cryptoutilUnsealRepository.NewUnsealRepositoryMock(t, 1)
	assert.NoError(t, err)
	assert.NotNil(t, mockUnsealRepository)
	defer mockUnsealRepository.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testOrmRepository, mockUnsealRepository, testAes256KeyGenPool)
	assert.NoError(t, err)
	assert.NotNil(t, rootKeysService)
	defer rootKeysService.Shutdown()
}

func TestRootKeysService_SadPath_ZeroUnsealJwks(t *testing.T) {
	mockUnsealRepository, _, err := cryptoutilUnsealRepository.NewUnsealRepositoryMock(t, 0)
	assert.NoError(t, err)
	assert.NotNil(t, mockUnsealRepository)
	defer mockUnsealRepository.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testOrmRepository, mockUnsealRepository, testAes256KeyGenPool)
	assert.Error(t, err)
	assert.Nil(t, rootKeysService)
	assert.EqualError(t, err, "no unseal JWKs")
}

func TestRootKeysService_SadPath_NilUnsealJwks(t *testing.T) {
	mockUnsealRepository, _, err := cryptoutilUnsealRepository.NewUnsealRepositoryMock(t, 0)
	assert.NoError(t, err)
	assert.NotNil(t, mockUnsealRepository)
	mockUnsealRepository.On("UnsealJwks").Return(nil)
	defer mockUnsealRepository.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testOrmRepository, mockUnsealRepository, testAes256KeyGenPool)
	assert.Error(t, err)
	assert.Nil(t, rootKeysService)
	assert.EqualError(t, err, "no unseal JWKs")
}
