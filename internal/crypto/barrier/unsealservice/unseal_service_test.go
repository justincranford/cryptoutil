package unsealservice

import (
	"context"
	cryptoutilUnsealRepository "cryptoutil/internal/crypto/barrier/unsealrepository"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilSqlProvider "cryptoutil/internal/repository/sqlprovider"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	"log/slog"
	"os"
	"testing"

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

	testTelemetryService, err = cryptoutilTelemetry.NewService(testCtx, "servicelogic_test", false, false)
	if err != nil {
		slog.Error("failed to initailize telemetry", "error", err)
		os.Exit(-1)
	}
	defer testTelemetryService.Shutdown()

	testSqlProvider, err = cryptoutilSqlProvider.NewSqlProviderForTest(testCtx, testTelemetryService, testDbType)
	if err != nil {
		testTelemetryService.Slogger.Error("failed to initailize sqlProvider", "error", err)
		os.Exit(-1)
	}
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
	mockUnsealKeyRepository, err := cryptoutilUnsealRepository.NewUnsealKeyRepositoryMock(t, 1)
	assert.NoError(t, err)
	assert.NotNil(t, mockUnsealKeyRepository)

	service, err := NewUnsealService(testTelemetryService, testRepositoryProvider, mockUnsealKeyRepository)
	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestUnsealService_SadPath_ZeroUnsealJwks(t *testing.T) {
	mockUnsealKeyRepository, err := cryptoutilUnsealRepository.NewUnsealKeyRepositoryMock(t, 0)
	assert.NoError(t, err)
	assert.NotNil(t, mockUnsealKeyRepository)

	service, err := NewUnsealService(testTelemetryService, testRepositoryProvider, mockUnsealKeyRepository)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.EqualError(t, err, "no unseal JWKs")
}

func TestUnsealService_SadPath_NilUnsealJwks(t *testing.T) {
	mockUnsealKeyRepository, err := cryptoutilUnsealRepository.NewUnsealKeyRepositoryMock(t, 0)
	assert.NoError(t, err)
	assert.NotNil(t, mockUnsealKeyRepository)
	mockUnsealKeyRepository.On("UnsealJwks").Return(nil)

	service, err := NewUnsealService(testTelemetryService, testRepositoryProvider, mockUnsealKeyRepository)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.EqualError(t, err, "no unseal JWKs")
}
