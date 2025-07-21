package sqlrepository

import (
	"context"
	"fmt"
	"os"
	"testing"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
	testSqlRepository    *SqlRepository
)

func TestMain(m *testing.M) {
	var rc int
	func() {
		testSettings := &cryptoutilConfig.Settings{
			LogLevel:  "ALL",
			DevMode:   true,
			OTLPScope: "sql_provider_test",
		}
		testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		testSqlRepository = RequireNewForTest(testCtx, testTelemetryService, DBTypeSQLite)
		defer testSqlRepository.Shutdown()
		testSqlRepository.logConnectionPoolSettings()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestSqlRepository_UnsupportedDatabaseType(t *testing.T) {
	_, err := NewSqlRepository(testCtx, testTelemetryService, "invalidDbType", "", ContainerModeDisabled)
	require.Error(t, err)
}

func TestSqlRepository_PingFailure(t *testing.T) {
	invalidProvider, err := NewSqlRepository(testCtx, testTelemetryService, DBTypeSQLite, "invalid:memory:", ContainerModeDisabled)
	require.Error(t, err)
	require.Nil(t, invalidProvider)
}

func TestSqlTransaction_PanicRecovery(t *testing.T) {
	defer func() {
		if recover := recover(); recover != nil {
			require.NotNil(t, recover)
		}
	}()

	err := testSqlRepository.WithTransaction(testCtx, false, func(sqlTransaction *SqlTransaction) error {
		require.NotNil(t, sqlTransaction)
		panic("simulated panic during transaction")
	})
	require.Error(t, err)
}

func TestSqlTransaction_Success(t *testing.T) {
	err := testSqlRepository.WithTransaction(testCtx, false, func(sqlTransaction *SqlTransaction) error {
		require.NotNil(t, sqlTransaction)
		require.False(t, sqlTransaction.IsReadOnly())
		return nil
	})
	require.NoError(t, err)
}

func TestSqlTransaction_BeginAlreadyStartedFailure(t *testing.T) {
	err := testSqlRepository.WithTransaction(testCtx, false, func(sqlTransaction *SqlTransaction) error {
		require.NotNil(t, sqlTransaction)
		require.False(t, sqlTransaction.IsReadOnly())

		err := sqlTransaction.begin(testCtx, false)
		require.Error(t, err)

		return err
	})
	require.Error(t, err)
}

func TestSqlTransaction_CommitNotStartedFailure(t *testing.T) {
	sqlTransaction := &SqlTransaction{sqlRepository: testSqlRepository}

	commitErr := sqlTransaction.commit()
	require.Error(t, commitErr)
	require.EqualError(t, commitErr, "can't commit because transaction not active")
}

func TestSqlTransaction_RollbackNotStartedFailure(t *testing.T) {
	sqlTransaction := &SqlTransaction{sqlRepository: testSqlRepository}

	rollbackErr := sqlTransaction.rollback()
	require.Error(t, rollbackErr)
	require.EqualError(t, rollbackErr, "can't rollback because transaction not active")
}

func TestSqlTransaction_BeginWithReadOnly(t *testing.T) {
	newVar := func(sqlTransaction *SqlTransaction) error {
		require.NotNil(t, sqlTransaction)
		require.True(t, sqlTransaction.IsReadOnly())

		return nil
	}
	err := testSqlRepository.WithTransaction(testCtx, true, newVar)
	require.Error(t, err)
	require.EqualError(t, err, "database sqlite doesn't support read-only transactions")
}

func TestSqlTransaction_RollbackOnError(t *testing.T) {
	err := testSqlRepository.WithTransaction(testCtx, false, func(sqlTransaction *SqlTransaction) error {
		require.NotNil(t, sqlTransaction)
		require.False(t, sqlTransaction.IsReadOnly())
		return fmt.Errorf("intentional failure") // Simulate an error within the transaction
	})
	require.Error(t, err)
}
