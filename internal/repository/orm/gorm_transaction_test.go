package orm

import (
	"context"
	cryptoutilSqlProvider "cryptoutil/internal/repository/sqlprovider"
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
	testCtx                = context.Background()
	testTelemetryService   *cryptoutilTelemetry.Service
	testSqlProvider        *cryptoutilSqlProvider.SqlProvider
	testRepositoryProvider *RepositoryProvider
)

func TestMain(m *testing.M) {
	telemetryService, err := cryptoutilTelemetry.NewService(testCtx, "sqlprovider_test", false, false)
	if err != nil {
		slog.Error("failed to initailize telemetry", "error", err)
		os.Exit(-1)
	}
	testTelemetryService = telemetryService
	defer testTelemetryService.Shutdown()

	sqlProvider, err := cryptoutilSqlProvider.NewSqlProvider(testCtx, testTelemetryService, cryptoutilSqlProvider.DBTypeSQLite, ":memory:", cryptoutilSqlProvider.ContainerModeDisabled)
	if err != nil {
		slog.Error("failed to initailize sqlProvider", "error", err)
		os.Exit(-1)
	}
	testSqlProvider = sqlProvider
	defer sqlProvider.Shutdown()

	repositoryProvider, err := NewRepositoryOrm(testCtx, testTelemetryService, sqlProvider, true)
	if err != nil {
		slog.Error("failed to initailize repositoryProvider", "error", err)
		os.Exit(-1)
	}
	testRepositoryProvider = repositoryProvider
	defer repositoryProvider.Shutdown()

	os.Exit(m.Run())
}

func TestSqlTransaction_PanicRecovery(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			require.NotNil(t, r)
		}
	}()

	err := testRepositoryProvider.WithTransaction(testCtx, false, func(repositoryTransaction *RepositoryTransaction) error {
		require.NotNil(t, repositoryTransaction)
		panic("simulated panic during transaction")
	})
	require.Error(t, err)
}

func TestSqlTransaction_Success(t *testing.T) {
	err := testRepositoryProvider.WithTransaction(testCtx, false, func(repositoryTransaction *RepositoryTransaction) error {
		require.NotNil(t, repositoryTransaction)
		require.False(t, repositoryTransaction.IsReadOnly())
		return nil
	})
	require.NoError(t, err)
}

func TestSqlTransaction_BeginAlreadyStartedFailure(t *testing.T) {
	err := testRepositoryProvider.WithTransaction(testCtx, false, func(repositoryTransaction *RepositoryTransaction) error {
		require.NotNil(t, repositoryTransaction)
		require.False(t, repositoryTransaction.IsReadOnly())

		err := repositoryTransaction.begin(testCtx, false)
		require.Error(t, err)

		return err
	})
	require.Error(t, err)
}

func TestSqlTransaction_CommitNotStartedFailure(t *testing.T) {
	repositoryTransaction := &RepositoryTransaction{repositoryProvider: testRepositoryProvider}

	commitErr := repositoryTransaction.commit()
	require.Error(t, commitErr)
	require.EqualError(t, commitErr, "can't commit because transaction not active")
}

func TestSqlTransaction_RollbackNotStartedFailure(t *testing.T) {
	repositoryTransaction := &RepositoryTransaction{repositoryProvider: testRepositoryProvider}

	rollbackErr := repositoryTransaction.rollback()
	require.Error(t, rollbackErr)
	require.EqualError(t, rollbackErr, "can't rollback because transaction not active")
}

func TestSqlTransaction_BeginWithReadOnly(t *testing.T) {
	err := testRepositoryProvider.WithTransaction(testCtx, true, func(repositoryTransaction *RepositoryTransaction) error {
		require.NotNil(t, repositoryTransaction)
		require.True(t, repositoryTransaction.IsReadOnly())

		return nil
	})
	require.NoError(t, err)
}

func TestSqlTransaction_RollbackOnError(t *testing.T) {
	err := testRepositoryProvider.WithTransaction(testCtx, false, func(repositoryTransaction *RepositoryTransaction) error {
		require.NotNil(t, repositoryTransaction)
		require.False(t, repositoryTransaction.IsReadOnly())
		return fmt.Errorf("intentional failure") // Simulate an error within the transaction
	})
	require.Error(t, err)
}
