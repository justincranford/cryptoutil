package unsealservice

import (
	"context"
	"os"
	"testing"

	cryptoutilUnsealRepository "cryptoutil/internal/crypto/barrier/unsealrepository"
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
)

func TestMain(m *testing.M) {
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "servicelogic_test", false, false)
	defer testTelemetryService.Shutdown()

	testSqlRepository = cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testDbType)
	defer testSqlRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSqlRepository, true)
	defer testOrmRepository.Shutdown()

	os.Exit(m.Run())
}

func TestUnsealService_HappyPath_OneUnsealJwks(t *testing.T) {
	mockUnsealRepository, _, err := cryptoutilUnsealRepository.NewUnsealRepositoryMock(t, 1)
	assert.NoError(t, err)
	assert.NotNil(t, mockUnsealRepository)

	service, err := NewUnsealService(testTelemetryService, testOrmRepository, mockUnsealRepository)
	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestUnsealService_SadPath_ZeroUnsealJwks(t *testing.T) {
	mockUnsealRepository, _, err := cryptoutilUnsealRepository.NewUnsealRepositoryMock(t, 0)
	assert.NoError(t, err)
	assert.NotNil(t, mockUnsealRepository)

	service, err := NewUnsealService(testTelemetryService, testOrmRepository, mockUnsealRepository)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.EqualError(t, err, "no unseal JWKs")
}

func TestUnsealService_SadPath_NilUnsealJwks(t *testing.T) {
	mockUnsealRepository, _, err := cryptoutilUnsealRepository.NewUnsealRepositoryMock(t, 0)
	assert.NoError(t, err)
	assert.NotNil(t, mockUnsealRepository)
	mockUnsealRepository.On("UnsealJwks").Return(nil)

	service, err := NewUnsealService(testTelemetryService, testOrmRepository, mockUnsealRepository)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.EqualError(t, err, "no unseal JWKs")
}
