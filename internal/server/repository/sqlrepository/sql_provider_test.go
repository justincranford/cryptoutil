package sqlrepository

import (
	"context"
	"fmt"
	"os"
	"testing"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	testSettings         = cryptoutilConfig.RequireNewForTest("sql_provider_test")
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
	testSQLRepository    *SQLRepository
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		testSQLRepository = RequireNewForTest(testCtx, testTelemetryService, testSettings)
		defer testSQLRepository.Shutdown()
		testSQLRepository.logConnectionPoolSettings()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestSQLTransaction_PanicRecovery(t *testing.T) {
	defer func() {
		if recover := recover(); recover != nil {
			require.NotNil(t, recover)
		}
	}()

	err := testSQLRepository.WithTransaction(testCtx, false, func(sqlTransaction *SQLTransaction) error {
		require.NotNil(t, sqlTransaction)
		panic("simulated panic during transaction")
	})
	require.Error(t, err)
}

func TestSQLTransaction_Success(t *testing.T) {
	err := testSQLRepository.WithTransaction(testCtx, false, func(sqlTransaction *SQLTransaction) error {
		require.NotNil(t, sqlTransaction)
		require.False(t, sqlTransaction.IsReadOnly())

		return nil
	})
	require.NoError(t, err)
}

func TestSQLTransaction_BeginAlreadyStartedFailure(t *testing.T) {
	err := testSQLRepository.WithTransaction(testCtx, false, func(sqlTransaction *SQLTransaction) error {
		require.NotNil(t, sqlTransaction)
		require.False(t, sqlTransaction.IsReadOnly())

		err := sqlTransaction.begin(testCtx, false)
		require.Error(t, err)

		return err
	})
	require.Error(t, err)
}

func TestSQLTransaction_CommitNotStartedFailure(t *testing.T) {
	sqlTransaction := &SQLTransaction{sqlRepository: testSQLRepository}

	commitErr := sqlTransaction.commit()
	require.Error(t, commitErr)
	require.EqualError(t, commitErr, "can't commit because transaction not active")
}

func TestSQLTransaction_RollbackNotStartedFailure(t *testing.T) {
	sqlTransaction := &SQLTransaction{sqlRepository: testSQLRepository}

	rollbackErr := sqlTransaction.rollback()
	require.Error(t, rollbackErr)
	require.EqualError(t, rollbackErr, "can't rollback because transaction not active")
}

func TestSQLTransaction_BeginWithReadOnly(t *testing.T) {
	newVar := func(sqlTransaction *SQLTransaction) error {
		require.NotNil(t, sqlTransaction)
		require.True(t, sqlTransaction.IsReadOnly())

		return nil
	}
	err := testSQLRepository.WithTransaction(testCtx, true, newVar)
	require.Error(t, err)
	require.EqualError(t, err, "database sqlite doesn't support read-only transactions")
}

func TestSQLTransaction_RollbackOnError(t *testing.T) {
	err := testSQLRepository.WithTransaction(testCtx, false, func(sqlTransaction *SQLTransaction) error {
		require.NotNil(t, sqlTransaction)
		require.False(t, sqlTransaction.IsReadOnly())

		return fmt.Errorf("intentional failure") // Simulate an error within the transaction
	})
	require.Error(t, err)
}
