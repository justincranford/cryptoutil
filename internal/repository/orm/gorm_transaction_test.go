package orm

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	cryptoutilSqlProvider "cryptoutil/internal/repository/sqlprovider"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.Service
	testSqlProvider      *cryptoutilSqlProvider.SqlProvider
	testOrmRepository    *RepositoryProvider
	testGivens           *Givens
	skipReadOnlyTxTests  = true
	testDbType           = cryptoutilSqlProvider.DBTypeSQLite // Caution: modernc.org/sqlite doesn't support read-only transactions, but PostgreSQL does
	// testDbType = cryptoutilSqlProvider.DBTypePostgres
)

func TestMain(m *testing.M) {
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "orm_transaction_test", false, false)
	defer testTelemetryService.Shutdown()

	testSqlProvider = cryptoutilSqlProvider.RequireNewForTest(testCtx, testTelemetryService, testDbType)
	defer testSqlProvider.Shutdown()

	testOrmRepository = RequireNewForTest(testCtx, testTelemetryService, testSqlProvider, true)
	defer testOrmRepository.Shutdown()

	testGivens = RequireNewGivensForTest(testCtx, testTelemetryService)
	defer testGivens.Shutdown()

	os.Exit(m.Run())
}

func TestSqlTransaction_PanicRecovery(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			require.NotNil(t, r)
		}
	}()

	panicErr := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(repositoryTransaction *RepositoryTransaction) error {
		require.NotNil(t, repositoryTransaction)
		panic("simulated panic during transaction")
	})
	require.Error(t, panicErr)
	require.EqualError(t, panicErr, "simulated panic during transaction")
}

func TestSqlTransaction_BeginAlreadyStartedFailure(t *testing.T) {
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(repositoryTransaction *RepositoryTransaction) error {
		require.NotNil(t, repositoryTransaction)
		require.Equal(t, ReadWrite, *repositoryTransaction.Mode())

		err := repositoryTransaction.begin(testCtx, ReadWrite)
		require.Error(t, err)

		return err
	})
	require.Error(t, err)
	require.EqualError(t, err, "failed to execute transaction: transaction already started")
}

func TestSqlTransaction_CommitNotStartedFailure(t *testing.T) {
	repositoryTransaction := &RepositoryTransaction{repositoryProvider: testOrmRepository}

	commitErr := repositoryTransaction.commit()
	require.Error(t, commitErr)
	require.EqualError(t, commitErr, "can't commit because transaction not active")
}

func TestSqlTransaction_RollbackNotStartedFailure(t *testing.T) {
	repositoryTransaction := &RepositoryTransaction{repositoryProvider: testOrmRepository}

	rollbackErr := repositoryTransaction.rollback()
	require.Error(t, rollbackErr)
	require.EqualError(t, rollbackErr, "can't rollback because transaction not active")
}

func TestSqlTransaction_BeginWithReadOnly(t *testing.T) {
	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(repositoryTransaction *RepositoryTransaction) error {
		require.NotNil(t, repositoryTransaction)
		require.Equal(t, ReadOnly, *repositoryTransaction.Mode())

		return nil
	})
	require.NoError(t, err)
}

func TestSqlTransaction_RollbackOnError(t *testing.T) {
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(repositoryTransaction *RepositoryTransaction) error {
		require.NotNil(t, repositoryTransaction)
		require.Equal(t, ReadWrite, *repositoryTransaction.Mode())
		return fmt.Errorf("intentional failure")
	})
	require.Error(t, err)
	require.EqualError(t, err, "failed to execute transaction: intentional failure")
}

func TestSqlTransaction_Success(t *testing.T) {
	type happyPathTestCase struct {
		txMode      TransactionMode
		expectError bool
	}

	tests := []happyPathTestCase{}
	tests = append(tests, happyPathTestCase{txMode: AutoCommit, expectError: false})
	tests = append(tests, happyPathTestCase{txMode: ReadWrite, expectError: false})
	if !skipReadOnlyTxTests {
		tests = append(tests, happyPathTestCase{txMode: ReadOnly, expectError: true})
	}

	for _, testCase := range tests {
		testTelemetryService.Slogger.Info("Executing test case", "mode", testCase.txMode, "expectError", testCase.expectError)

		addedKeyPools := []*KeyPool{}
		addedKeys := []*Key{}
		err := testOrmRepository.WithTransaction(testCtx, testCase.txMode, func(repositoryTransaction *RepositoryTransaction) error {
			require.NotNil(t, repositoryTransaction)
			require.NotNil(t, repositoryTransaction.ID())
			require.NotNil(t, repositoryTransaction.Context())
			require.Equal(t, testCase.txMode, *repositoryTransaction.Mode())

			keyPool := testGivens.Aes256KeyPool(true, true, true)
			err := repositoryTransaction.AddKeyPool(keyPool)
			if err != nil {
				return fmt.Errorf("failed to add Key Pool: %w", err)
			}
			addedKeyPools = append(addedKeyPools, keyPool)

			for nextKeyId := 1; nextKeyId <= 10; nextKeyId++ {
				now := time.Now().UTC()
				key := testGivens.Aes256Key(keyPool.KeyPoolID, &now, nil, nil, nil)
				err = repositoryTransaction.AddKeyPoolKey(key)
				if err != nil {
					return fmt.Errorf("failed to add Key: %w", err)
				}
			}

			return nil
		})

		testTelemetryService.Slogger.Info("Happy path test case result", "mode", testCase.txMode, "expectError", testCase.expectError, "error", err)
		if testCase.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}

		for _, addedKeyPool := range addedKeyPools {
			err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(repositoryTransaction *RepositoryTransaction) error {
				require.NotNil(t, repositoryTransaction)
				require.NotNil(t, repositoryTransaction.ID())
				require.NotNil(t, repositoryTransaction.Context())
				require.Equal(t, ReadOnly, *repositoryTransaction.Mode())

				retrievedKeyPool, err := repositoryTransaction.GetKeyPool(addedKeyPool.KeyPoolID)
				if err != nil {
					return fmt.Errorf("failed to get Key Pool: %w", err)
				}
				require.Equal(t, addedKeyPool, retrievedKeyPool)

				return nil
			})
			require.NoError(t, err)
		}

		for _, addedKey := range addedKeys {
			err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(repositoryTransaction *RepositoryTransaction) error {
				require.NotNil(t, repositoryTransaction)
				require.NotNil(t, repositoryTransaction.ID())
				require.NotNil(t, repositoryTransaction.Context())
				require.Equal(t, ReadOnly, *repositoryTransaction.Mode())

				retrievedKey, err := repositoryTransaction.GetKeyPoolKey(addedKey.KeyPoolID, addedKey.KeyID)
				if err != nil {
					return fmt.Errorf("failed to get Key: %w", err)
				}
				require.Equal(t, addedKey, retrievedKey)

				return nil
			})
			require.NoError(t, err)
		}
	}
}
