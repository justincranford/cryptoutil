package rootkeysservice

import (
	"context"
	"os"
	"testing"

	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/server/repository/sqlrepository"

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
)

func TestMain(m *testing.M) {
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "root_keys_service_test", false, false)
	defer testTelemetryService.Shutdown()

	testAes256KeyGenPool = cryptoutilKeygen.RequireNewAes256GcmGenKeyPoolForTest(testTelemetryService)
	defer testAes256KeyGenPool.Close()

	os.Exit(m.Run())
}

func TestRootKeysService_HappyPath_OneUnsealJwks(t *testing.T) {
	unsealJwks := cryptoutilJose.GenerateJweJwksForTest(t, 1, &cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NotNil(t, unsealJwks)

	unsealKeysServiceSimple, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple(unsealJwks)
	require.NoError(t, err)
	require.NotNil(t, unsealKeysServiceSimple)
	defer unsealKeysServiceSimple.Shutdown()

	testSqlRepository = cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testDbType)
	defer testSqlRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSqlRepository, true)
	defer testOrmRepository.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testOrmRepository, unsealKeysServiceSimple, testAes256KeyGenPool)
	require.NoError(t, err)
	require.NotNil(t, rootKeysService)
	defer rootKeysService.Shutdown()
}

func TestRootKeysService_SadPath_ZeroUnsealJwks(t *testing.T) {
	unsealKeysServiceSimple := cryptoutilUnsealKeysService.RequireNewSimpleForTest([]joseJwk.Key{})
	require.NotNil(t, unsealKeysServiceSimple)
	defer unsealKeysServiceSimple.Shutdown()

	testSqlRepository = cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testDbType)
	defer testSqlRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSqlRepository, true)
	defer testOrmRepository.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testOrmRepository, unsealKeysServiceSimple, testAes256KeyGenPool)
	require.Error(t, err)
	require.Nil(t, rootKeysService)
	require.EqualError(t, err, "failed to initialize first root JWK: failed to encrypt first root JWK: failed to encrypt root JWK with unseal JWK: invalid JWKs: jwks can't be empty")
}

func TestRootKeysService_SadPath_NilUnsealJwks(t *testing.T) {
	unsealKeysServiceSimple := cryptoutilUnsealKeysService.RequireNewSimpleForTest(nil)
	require.NotNil(t, unsealKeysServiceSimple)
	defer unsealKeysServiceSimple.Shutdown()

	testSqlRepository = cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testDbType)
	defer testSqlRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSqlRepository, true)
	defer testOrmRepository.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testOrmRepository, unsealKeysServiceSimple, testAes256KeyGenPool)
	require.Error(t, err)
	require.Nil(t, rootKeysService)
	require.EqualError(t, err, "failed to initialize first root JWK: failed to encrypt first root JWK: failed to encrypt root JWK with unseal JWK: invalid JWKs: jwks can't be nil")
}
