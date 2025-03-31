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
	var err error

	testTelemetryService, err = cryptoutilTelemetry.NewService(testCtx, "sqlprovider_test", false, false)
	if err != nil {
		slog.Error("failed to initailize telemetry", "error", err)
		os.Exit(-1)
	}
	defer testTelemetryService.Shutdown()

	testSqlProvider, err = cryptoutilSqlProvider.NewSqlProvider(testCtx, testTelemetryService, cryptoutilSqlProvider.DBTypeSQLite, ":memory:", cryptoutilSqlProvider.ContainerModeDisabled)
	if err != nil {
		slog.Error("failed to initailize sqlProvider", "error", err)
		os.Exit(-1)
	}
	defer testSqlProvider.Shutdown()

	testRepositoryProvider, err = NewRepositoryOrm(testCtx, testTelemetryService, testSqlProvider, true)
	if err != nil {
		slog.Error("failed to initailize repositoryProvider", "error", err)
		os.Exit(-1)
	}
	defer testRepositoryProvider.Shutdown()

	os.Exit(m.Run())
}

func TestSqlTransaction_PanicRecovery(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			require.NotNil(t, r)
		}
	}()

	panicErr := testRepositoryProvider.WithTransaction(testCtx, false, false, func(repositoryTransaction *RepositoryTransaction) error {
		require.NotNil(t, repositoryTransaction)
		panic("simulated panic during transaction")
	})
	require.Error(t, panicErr)
	require.EqualError(t, panicErr, "simulated panic during transaction")
}

func TestSqlTransaction_Success(t *testing.T) {
	err := testRepositoryProvider.WithTransaction(testCtx, false, false, func(repositoryTransaction *RepositoryTransaction) error {
		require.NotNil(t, repositoryTransaction)
		require.False(t, repositoryTransaction.AutoCommit())
		require.False(t, repositoryTransaction.ReadOnly())
		return nil
	})
	require.NoError(t, err)
}

func TestSqlTransaction_BeginAlreadyStartedFailure(t *testing.T) {
	err := testRepositoryProvider.WithTransaction(testCtx, false, false, func(repositoryTransaction *RepositoryTransaction) error {
		require.NotNil(t, repositoryTransaction)
		require.False(t, repositoryTransaction.AutoCommit())
		require.False(t, repositoryTransaction.ReadOnly())

		err := repositoryTransaction.begin(testCtx, false, false)
		require.Error(t, err)

		return err
	})
	require.Error(t, err)
	require.EqualError(t, err, "failed to execute transaction: transaction already started")
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
	err := testRepositoryProvider.WithTransaction(testCtx, true, true, func(repositoryTransaction *RepositoryTransaction) error {
		require.NotNil(t, repositoryTransaction)
		require.True(t, repositoryTransaction.AutoCommit())
		require.True(t, repositoryTransaction.ReadOnly())

		return nil
	})
	require.NoError(t, err)
}

func TestSqlTransaction_RollbackOnError(t *testing.T) {
	err := testRepositoryProvider.WithTransaction(testCtx, false, false, func(repositoryTransaction *RepositoryTransaction) error {
		require.NotNil(t, repositoryTransaction)
		require.False(t, repositoryTransaction.AutoCommit())
		require.False(t, repositoryTransaction.ReadOnly())
		return fmt.Errorf("intentional failure")
	})
	require.Error(t, err)
	require.EqualError(t, err, "failed to execute transaction: intentional failure")
}
