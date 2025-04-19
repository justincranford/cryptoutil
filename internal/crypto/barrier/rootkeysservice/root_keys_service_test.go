package rootkeysservice

import (
	"context"
	"os"
	"testing"

	cryptoutilUnsealKeysService "cryptoutil/internal/crypto/barrier/unsealkeysservice"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/repository/sqlrepository"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	"github.com/stretchr/testify/require"
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
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "root_keys_service_test", false, false)
	defer testTelemetryService.Shutdown()

	testAes256KeyGenPool = cryptoutilKeygen.RequireNewAes256GenKeyPoolForTest(testTelemetryService)
	defer testAes256KeyGenPool.Close()

	os.Exit(m.Run())
}

func TestRootKeysService_HappyPath_OneUnsealJwks(t *testing.T) {
	mockUnsealKeysService, _, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceMock(t, 1)
	require.NoError(t, err)
	require.NotNil(t, mockUnsealKeysService)
	defer mockUnsealKeysService.Shutdown()

	testSqlRepository = cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testDbType)
	defer testSqlRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSqlRepository, true)
	defer testOrmRepository.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testOrmRepository, mockUnsealKeysService, testAes256KeyGenPool)
	require.NoError(t, err)
	require.NotNil(t, rootKeysService)
	defer rootKeysService.Shutdown()
}

func TestRootKeysService_SadPath_ZeroUnsealJwks(t *testing.T) {
	mockUnsealKeysService, _, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceMock(t, 0)
	require.NoError(t, err)
	require.NotNil(t, mockUnsealKeysService)
	defer mockUnsealKeysService.Shutdown()

	testSqlRepository = cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testDbType)
	defer testSqlRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSqlRepository, true)
	defer testOrmRepository.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testOrmRepository, mockUnsealKeysService, testAes256KeyGenPool)
	require.Error(t, err)
	require.Nil(t, rootKeysService)
	require.EqualError(t, err, "failed to initialize first root JWK: failed to encrypt first root JWK: failed to encrypt root JWK with unseal JWK")
}

func TestRootKeysService_SadPath_NilUnsealJwks(t *testing.T) {
	mockUnsealKeysService, _, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceMock(t, 0)
	require.NoError(t, err)
	require.NotNil(t, mockUnsealKeysService)
	mockUnsealKeysService.On("unsealJwks").Return(nil)
	defer mockUnsealKeysService.Shutdown()

	testSqlRepository = cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testDbType)
	defer testSqlRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSqlRepository, true)
	defer testOrmRepository.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testOrmRepository, mockUnsealKeysService, testAes256KeyGenPool)
	require.Error(t, err)
	require.Nil(t, rootKeysService)
	require.EqualError(t, err, "failed to initialize first root JWK: failed to encrypt first root JWK: failed to encrypt root JWK with unseal JWK")
}
