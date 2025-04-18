package intermediatekeysservice

import (
	"context"
	"os"
	"testing"

	cryptoutilRootKeysService "cryptoutil/internal/crypto/barrier/rootkeysservice"
	cryptoutilUnsealRepository "cryptoutil/internal/crypto/barrier/unsealrepository"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/repository/sqlrepository"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/assert"
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
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "intermediatekeysservice_test", false, false)
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

	os.Exit(m.Run())
}

func TestIntermediateKeysService_HappyPath(t *testing.T) {
	intermediateKeysService, err := NewIntermediateKeysService(testTelemetryService, testOrmRepository, testRootKeysService, testAes256KeyGenPool)
	assert.NoError(t, err)
	assert.NotNil(t, intermediateKeysService)
	defer intermediateKeysService.Shutdown()

	var latest joseJwk.Key
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		latest, err = intermediateKeysService.GetLatest(sqlTransaction)
		assert.NoError(t, err)
		assert.NotNil(t, latest)
		return err
	})
	assert.NoError(t, err)
	assert.NotNil(t, latest)

	var all []joseJwk.Key
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		all, err = intermediateKeysService.GetAll(sqlTransaction)
		assert.NoError(t, err)
		assert.NotNil(t, latest)
		assert.Equal(t, 1, len(all))
		return err
	})
	assert.NoError(t, err)
	assert.NotNil(t, all)
	assert.Equal(t, 1, len(all))
}
