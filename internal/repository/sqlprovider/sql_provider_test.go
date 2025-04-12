package sqlprovider

import (
	"context"
	"fmt"
	"os"
	"testing"

	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.Service
	testSqlProvider      *SqlProvider
)

func TestMain(m *testing.M) {
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "sqlprovider_test", false, false)
	defer testTelemetryService.Shutdown()

	testSqlProvider = RequireNewForTest(testCtx, testTelemetryService, DBTypeSQLite)
	defer testSqlProvider.Shutdown()

	testSqlProvider.logConnectionPoolSettings()

	os.Exit(m.Run())
}

func TestSqlProvider_UnsupportedDatabaseType(t *testing.T) {
	_, err := NewSqlProvider(testCtx, testTelemetryService, "invalidDbType", "", ContainerModeDisabled)
	require.Error(t, err)
}

func TestSqlProvider_PingFailure(t *testing.T) {
	invalidProvider, err := NewSqlProvider(testCtx, testTelemetryService, DBTypeSQLite, "invalid:memory:", ContainerModeDisabled)
	require.Error(t, err)
	require.Nil(t, invalidProvider)
}

func TestSqlTransaction_PanicRecovery(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			require.NotNil(t, r)
		}
	}()

	err := testSqlProvider.WithTransaction(testCtx, false, func(sqlTransaction *SqlTransaction) error {
		require.NotNil(t, sqlTransaction)
		panic("simulated panic during transaction")
	})
	require.Error(t, err)
}

func TestSqlTransaction_Success(t *testing.T) {
	err := testSqlProvider.WithTransaction(testCtx, false, func(sqlTransaction *SqlTransaction) error {
		require.NotNil(t, sqlTransaction)
		require.False(t, sqlTransaction.IsReadOnly())
		return nil
	})
	require.NoError(t, err)
}

func TestSqlTransaction_BeginAlreadyStartedFailure(t *testing.T) {
	err := testSqlProvider.WithTransaction(testCtx, false, func(sqlTransaction *SqlTransaction) error {
		require.NotNil(t, sqlTransaction)
		require.False(t, sqlTransaction.IsReadOnly())

		err := sqlTransaction.begin(testCtx, false)
		require.Error(t, err)

		return err
	})
	require.Error(t, err)
}

func TestSqlTransaction_CommitNotStartedFailure(t *testing.T) {
	sqlTransaction := &SqlTransaction{sqlProvider: testSqlProvider}

	commitErr := sqlTransaction.commit()
	require.Error(t, commitErr)
	require.EqualError(t, commitErr, "can't commit because transaction not active")
}

func TestSqlTransaction_RollbackNotStartedFailure(t *testing.T) {
	sqlTransaction := &SqlTransaction{sqlProvider: testSqlProvider}

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
	err := testSqlProvider.WithTransaction(testCtx, true, newVar)
	require.Error(t, err)
	require.EqualError(t, err, "database sqlite doesn't support read-only transactions")
}

func TestSqlTransaction_RollbackOnError(t *testing.T) {
	err := testSqlProvider.WithTransaction(testCtx, false, func(sqlTransaction *SqlTransaction) error {
		require.NotNil(t, sqlTransaction)
		require.False(t, sqlTransaction.IsReadOnly())
		return fmt.Errorf("intentional failure") // Simulate an error within the transaction
	})
	require.Error(t, err)
}
