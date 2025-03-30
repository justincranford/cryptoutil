package sqlprovider

import (
	"context"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	"fmt"
	"log/slog"
	"os"
	"testing"

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
	telemetryService, err := cryptoutilTelemetry.NewService(testCtx, "sqlprovider_test", false, false)
	if err != nil {
		slog.Error("failed to initailize telemetry", "error", err)
		os.Exit(-1)
	}
	testTelemetryService = telemetryService
	defer testTelemetryService.Shutdown()

	sqlProvider, err := NewSqlProvider(testCtx, testTelemetryService, DBTypeSQLite, ":memory:", ContainerModeDisabled)
	if err != nil {
		slog.Error("failed to initailize sqlProvider", "error", err)
		os.Exit(-1)
	}
	testSqlProvider = sqlProvider
	defer sqlProvider.Shutdown()

	sqlProvider.logConnectionPoolSettings()

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

	err := testSqlProvider.WithTransaction(testCtx, false, func(tx *SqlTransaction) error {
		require.NotNil(t, tx)
		panic("simulated panic during transaction")
	})
	require.Error(t, err)
}

func TestSqlTransaction_Success(t *testing.T) {
	err := testSqlProvider.WithTransaction(testCtx, false, func(tx *SqlTransaction) error {
		require.NotNil(t, tx)
		require.False(t, tx.IsReadOnly())
		return nil
	})
	require.NoError(t, err)
}

func TestSqlTransaction_BeginAlreadyStartedFailure(t *testing.T) {
	err := testSqlProvider.WithTransaction(testCtx, false, func(tx *SqlTransaction) error {
		require.NotNil(t, tx)
		require.False(t, tx.IsReadOnly())

		err := tx.Begin(testCtx, false)
		require.Error(t, err)

		return err
	})
	require.Error(t, err)
}

func TestSqlTransaction_CommitNotStartedFailure(t *testing.T) {
	tx := &SqlTransaction{sqlProvider: testSqlProvider}

	commitErr := tx.Commit()
	require.Error(t, commitErr)
	require.EqualError(t, commitErr, "can't commit because transaction not active")
}

func TestSqlTransaction_RollbackNotStartedFailure(t *testing.T) {
	tx := &SqlTransaction{sqlProvider: testSqlProvider}

	rollbackErr := tx.Rollback()
	require.Error(t, rollbackErr)
	require.EqualError(t, rollbackErr, "can't rollback because transaction not active")
}

func TestSqlTransaction_BeginWithReadOnly(t *testing.T) {
	err := testSqlProvider.WithTransaction(testCtx, true, func(tx *SqlTransaction) error {
		require.NotNil(t, tx)
		require.True(t, tx.IsReadOnly())

		return nil
	})
	require.NoError(t, err)
}

func TestSqlTransaction_RollbackOnError(t *testing.T) {
	err := testSqlProvider.WithTransaction(testCtx, false, func(tx *SqlTransaction) error {
		require.NotNil(t, tx)
		require.False(t, tx.IsReadOnly())
		return fmt.Errorf("intentional failure") // Simulate an error within the transaction
	})
	require.Error(t, err)
}
