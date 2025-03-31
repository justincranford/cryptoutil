package orm

import (
	"context"
	cryptoutilSqlProvider "cryptoutil/internal/repository/sqlprovider"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	testCtx                = context.Background()
	testTelemetryService   *cryptoutilTelemetry.Service
	testSqlProvider        *cryptoutilSqlProvider.SqlProvider
	testRepositoryProvider *RepositoryProvider
	testGivens             *Givens
	skipReadOnlyTxTests    bool
	testDbType             = cryptoutilSqlProvider.DBTypeSQLite // Caution: modernc.org/sqlite doesn't seem to support read-only transactions, but PostgreSQL does
	// testDbType = cryptoutilSqlProvider.DBTypePostgres
)

func TestMain(m *testing.M) {
	var err error

	testTelemetryService, err = cryptoutilTelemetry.NewService(testCtx, "orm_transaction_test", false, false)
	if err != nil {
		slog.Error("failed to initailize telemetry", "error", err)
		os.Exit(-1)
	}
	defer testTelemetryService.Shutdown()

	switch testDbType {
	case cryptoutilSqlProvider.DBTypeSQLite:
		skipReadOnlyTxTests = true
		testSqlProvider, err = cryptoutilSqlProvider.NewSqlProvider(testCtx, testTelemetryService, cryptoutilSqlProvider.DBTypeSQLite, ":memory:", cryptoutilSqlProvider.ContainerModeDisabled)
	case cryptoutilSqlProvider.DBTypePostgres:
		skipReadOnlyTxTests = false
		testSqlProvider, err = cryptoutilSqlProvider.NewSqlProvider(testCtx, testTelemetryService, cryptoutilSqlProvider.DBTypePostgres, "", cryptoutilSqlProvider.ContainerModeRequired)
	default:
		testTelemetryService.Slogger.Error("unsupported dbType", "dbType", string(testDbType))
		os.Exit(-1)
	}
	if err != nil {
		testTelemetryService.Slogger.Error("failed to initailize sqlProvider", "error", err)
		os.Exit(-1)
	}
	defer testSqlProvider.Shutdown()

	testRepositoryProvider, err = NewRepositoryOrm(testCtx, testTelemetryService, testSqlProvider, true)
	if err != nil {
		testTelemetryService.Slogger.Error("failed to initailize repositoryProvider", "error", err)
		os.Exit(-1)
	}
	defer testRepositoryProvider.Shutdown()

	testGivens = NewGivens(testCtx, testTelemetryService)
	defer testGivens.Shutdown()

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

func TestSqlTransaction_Success(t *testing.T) {
	type happyPathTestCase struct {
		readOnly    bool
		autoCommit  bool
		expectError bool
	}

	tests := []happyPathTestCase{}
	tests = append(tests, happyPathTestCase{readOnly: false, autoCommit: false, expectError: false})
	tests = append(tests, happyPathTestCase{readOnly: false, autoCommit: true, expectError: false})
	if !skipReadOnlyTxTests {
		tests = append(tests, happyPathTestCase{readOnly: true, autoCommit: false, expectError: true})
		// tests = append(tests, happyPathTestCase{readOnly: true, autoCommit: true, expectError: true})
	}

	for _, testCase := range tests {
		testTelemetryService.Slogger.Info("Executing test case", "autoCommit", testCase.autoCommit, "readOnly", testCase.readOnly, "expectError", testCase.expectError)
		err := testRepositoryProvider.WithTransaction(testCtx, testCase.autoCommit, testCase.readOnly, func(repositoryTransaction *RepositoryTransaction) error {
			require.NotNil(t, repositoryTransaction)
			require.NotNil(t, repositoryTransaction.ID())
			require.NotNil(t, repositoryTransaction.Context())
			require.Equal(t, testCase.autoCommit, repositoryTransaction.AutoCommit())
			require.Equal(t, testCase.readOnly, repositoryTransaction.ReadOnly())

			keyPool, err := testGivens.KeyPoolForAdd(true, false, false)
			if err != nil {
				return fmt.Errorf("failed to generate given Key Pool for insert: %w", err)
			}
			err = repositoryTransaction.AddKeyPool(keyPool)
			if err != nil {
				return fmt.Errorf("failed to add Key Pool: %w", err)
			}

			for nextKeyId := 1; nextKeyId <= 10; nextKeyId++ {
				now := time.Now().UTC()
				key := testGivens.Key(keyPool.KeyPoolID, nextKeyId, &now, nil, nil, nil)
				err = repositoryTransaction.AddKey(key)
				if err != nil {
					return fmt.Errorf("failed to add Key: %w", err)
				}
			}

			return nil
		})
		if testCase.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
