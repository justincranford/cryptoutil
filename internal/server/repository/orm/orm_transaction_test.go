package orm

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilSqlRepository "cryptoutil/internal/server/repository/sqlrepository"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
	testJwkGenService    *cryptoutilJose.JwkGenService
	testSqlRepository    *cryptoutilSqlRepository.SqlRepository
	testOrmRepository    *OrmRepository
	testGivens           *Givens
	skipReadOnlyTxTests  = true                                 // true for DBTypeSQLite, false for DBTypePostgres
	testDbType           = cryptoutilSqlRepository.DBTypeSQLite // Caution: modernc.org/sqlite doesn't support read-only transactions, but PostgreSQL does
)

func TestMain(m *testing.M) {
	var rc int
	func() {
		testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "gorm_transaction_test", false, false)
		defer testTelemetryService.Shutdown()

		testJwkGenService = cryptoutilJose.RequireNewForTest(testCtx, testTelemetryService)
		defer testJwkGenService.Shutdown()

		testSqlRepository = cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testDbType)
		defer testSqlRepository.Shutdown()

		testOrmRepository = RequireNewForTest(testCtx, testTelemetryService, testJwkGenService, testSqlRepository, true)
		defer testOrmRepository.Shutdown()

		testGivens = RequireNewGivensForTest(testCtx, testTelemetryService)
		defer testGivens.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestSqlTransaction_PanicRecovery(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			require.NotNil(t, r)
		}
	}()

	panicErr := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(ormTransaction *OrmTransaction) error {
		require.NotNil(t, ormTransaction)
		panic("simulated panic during transaction")
	})
	require.Error(t, panicErr)
	require.EqualError(t, panicErr, "simulated panic during transaction")
}

func TestSqlTransaction_BeginAlreadyStartedFailure(t *testing.T) {
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(ormTransaction *OrmTransaction) error {
		require.NotNil(t, ormTransaction)
		require.Equal(t, ReadWrite, *ormTransaction.Mode())

		err := ormTransaction.begin(testCtx, ReadWrite)
		require.Error(t, err)

		return err
	})
	require.Error(t, err)
	require.EqualError(t, err, "failed to execute transaction: transaction already started")
}

func TestSqlTransaction_CommitNotStartedFailure(t *testing.T) {
	ormTransaction := &OrmTransaction{ormRepository: testOrmRepository}

	commitErr := ormTransaction.commit()
	require.Error(t, commitErr)
	require.EqualError(t, commitErr, "can't commit because transaction not active")
}

func TestSqlTransaction_RollbackNotStartedFailure(t *testing.T) {
	ormTransaction := &OrmTransaction{ormRepository: testOrmRepository}

	rollbackErr := ormTransaction.rollback()
	require.Error(t, rollbackErr)
	require.EqualError(t, rollbackErr, "can't rollback because transaction not active")
}

func TestSqlTransaction_BeginWithReadOnly(t *testing.T) {
	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(ormTransaction *OrmTransaction) error {
		require.NotNil(t, ormTransaction)
		require.Equal(t, ReadOnly, *ormTransaction.Mode())

		return nil
	})
	require.NoError(t, err)
}

func TestSqlTransaction_RollbackOnError(t *testing.T) {
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(ormTransaction *OrmTransaction) error {
		require.NotNil(t, ormTransaction)
		require.Equal(t, ReadWrite, *ormTransaction.Mode())
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
		err := testOrmRepository.WithTransaction(testCtx, testCase.txMode, func(ormTransaction *OrmTransaction) error {
			require.NotNil(t, ormTransaction)
			require.NotNil(t, ormTransaction.ID())
			require.NotNil(t, ormTransaction.Context())
			require.Equal(t, testCase.txMode, *ormTransaction.Mode())

			keyPool := testGivens.Aes256KeyPool(true, true, true)
			err := ormTransaction.AddKeyPool(keyPool)
			if err != nil {
				return fmt.Errorf("failed to add Key Pool: %w", err)
			}
			addedKeyPools = append(addedKeyPools, keyPool)

			for nextKeyId := 1; nextKeyId <= 10; nextKeyId++ {
				now := time.Now().UTC()
				key := testGivens.Aes256Key(keyPool.KeyPoolID, &now, nil, nil, nil)
				err = ormTransaction.AddKeyPoolKey(key)
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
			err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(ormTransaction *OrmTransaction) error {
				require.NotNil(t, ormTransaction)
				require.NotNil(t, ormTransaction.ID())
				require.NotNil(t, ormTransaction.Context())
				require.Equal(t, ReadOnly, *ormTransaction.Mode())

				retrievedKeyPool, err := ormTransaction.GetKeyPool(addedKeyPool.KeyPoolID)
				if err != nil {
					return fmt.Errorf("failed to get Key Pool: %w", err)
				}
				require.Equal(t, addedKeyPool, retrievedKeyPool)

				return nil
			})
			require.NoError(t, err)
		}

		for _, addedKey := range addedKeys {
			err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(ormTransaction *OrmTransaction) error {
				require.NotNil(t, ormTransaction)
				require.NotNil(t, ormTransaction.ID())
				require.NotNil(t, ormTransaction.Context())
				require.Equal(t, ReadOnly, *ormTransaction.Mode())

				retrievedKey, err := ormTransaction.GetKeyPoolKey(addedKey.KeyPoolID, addedKey.KeyID)
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
