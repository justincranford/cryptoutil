// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository

import (
	"context"
	"fmt"
	"os"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	testSettings         = cryptoutilAppsTemplateServiceConfig.RequireNewForTest("sql_provider_test")
	testCtx              = context.Background()
	testTelemetryService *cryptoutilSharedTelemetry.TelemetryService
	testSQLRepository    *SQLRepository
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		testTelemetryService = cryptoutilSharedTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		testSQLRepository = RequireNewForTest(testCtx, testTelemetryService, testSettings)
		defer testSQLRepository.Shutdown()

		testSQLRepository.logConnectionPoolSettings()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestSQLTransaction_PanicRecovery(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			require.NotNil(t, r)
		}
	}()

	err := testSQLRepository.WithTransaction(testCtx, false, func(sqlTransaction *SQLTransaction) error {
		require.NotNil(t, sqlTransaction)
		panic("simulated panic during transaction")
	})
	require.Error(t, err)
}

func TestSQLTransaction_Success(t *testing.T) {
	t.Parallel()

	err := testSQLRepository.WithTransaction(testCtx, false, func(sqlTransaction *SQLTransaction) error {
		require.NotNil(t, sqlTransaction)
		require.False(t, sqlTransaction.IsReadOnly())

		return nil
	})
	require.NoError(t, err)
}

func TestSQLTransaction_BeginAlreadyStartedFailure(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	sqlTransaction := &SQLTransaction{sqlRepository: testSQLRepository}

	commitErr := sqlTransaction.commit()
	require.Error(t, commitErr)
	require.EqualError(t, commitErr, "can't commit because transaction not active")
}

func TestSQLTransaction_RollbackNotStartedFailure(t *testing.T) {
	t.Parallel()

	sqlTransaction := &SQLTransaction{sqlRepository: testSQLRepository}

	rollbackErr := sqlTransaction.rollback()
	require.Error(t, rollbackErr)
	require.EqualError(t, rollbackErr, "can't rollback because transaction not active")
}

func TestSQLTransaction_BeginWithReadOnly(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	err := testSQLRepository.WithTransaction(testCtx, false, func(sqlTransaction *SQLTransaction) error {
		require.NotNil(t, sqlTransaction)
		require.False(t, sqlTransaction.IsReadOnly())

		return fmt.Errorf("intentional failure") // Simulate an error within the transaction
	})
	require.Error(t, err)
}
