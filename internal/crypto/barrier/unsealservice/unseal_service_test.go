package unsealservice

import (
	"context"
	"os"
	"testing"

	cryptoutilUnsealRepository "cryptoutil/internal/crypto/barrier/unsealrepository"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilSqlProvider "cryptoutil/internal/repository/sqlprovider"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	"github.com/stretchr/testify/assert"
)

var (
	testCtx                = context.Background()
	testTelemetryService   *cryptoutilTelemetry.Service
	testSqlProvider        *cryptoutilSqlProvider.SqlProvider
	testRepositoryProvider *cryptoutilOrmRepository.RepositoryProvider
	testDbType             = cryptoutilSqlProvider.DBTypeSQLite // Caution: modernc.org/sqlite doesn't support read-only transactions, but PostgreSQL does
)

func TestMain(m *testing.M) {
	var err error

	testTelemetryService = cryptoutilTelemetry.RequireNewService(testCtx, "servicelogic_test", false, false)
	defer testTelemetryService.Shutdown()

	testSqlProvider = cryptoutilSqlProvider.RequireNewForTest(testCtx, testTelemetryService, testDbType)
	defer testSqlProvider.Shutdown()

	testRepositoryProvider, err = cryptoutilOrmRepository.NewRepositoryOrm(testCtx, testTelemetryService, testSqlProvider, true)
	if err != nil {
		testTelemetryService.Slogger.Error("failed to initailize repositoryProvider", "error", err)
		os.Exit(-1)
	}
	defer testRepositoryProvider.Shutdown()

	os.Exit(m.Run())
}

func TestUnsealService_HappyPath_OneUnsealJwks(t *testing.T) {
	mockUnsealRepository, _, err := cryptoutilUnsealRepository.NewUnsealRepositoryMock(t, 1)
	assert.NoError(t, err)
	assert.NotNil(t, mockUnsealRepository)

	service, err := NewUnsealService(testTelemetryService, testRepositoryProvider, mockUnsealRepository)
	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestUnsealService_SadPath_ZeroUnsealJwks(t *testing.T) {
	mockUnsealRepository, _, err := cryptoutilUnsealRepository.NewUnsealRepositoryMock(t, 0)
	assert.NoError(t, err)
	assert.NotNil(t, mockUnsealRepository)

	service, err := NewUnsealService(testTelemetryService, testRepositoryProvider, mockUnsealRepository)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.EqualError(t, err, "no unseal JWKs")
}

func TestUnsealService_SadPath_NilUnsealJwks(t *testing.T) {
	mockUnsealRepository, _, err := cryptoutilUnsealRepository.NewUnsealRepositoryMock(t, 0)
	assert.NoError(t, err)
	assert.NotNil(t, mockUnsealRepository)
	mockUnsealRepository.On("UnsealJwks").Return(nil)

	service, err := NewUnsealService(testTelemetryService, testRepositoryProvider, mockUnsealRepository)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.EqualError(t, err, "no unseal JWKs")
}
