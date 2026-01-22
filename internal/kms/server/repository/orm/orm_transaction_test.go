// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"
	cryptoutilAppErr "cryptoutil/internal/shared/apperr"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilRandom "cryptoutil/internal/shared/util/random"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	testSettings         = cryptoutilConfig.RequireNewForTest("orm_transaction_test")
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
	testJWKGenService    *cryptoutilJose.JWKGenService
	testSQLRepository    *cryptoutilSQLRepository.SQLRepository
	testOrmRepository    *OrmRepository
	skipReadOnlyTxTests  = true // true for DBTypeSQLite, false for DBTypePostgres
	numMaterialKeys      = 10
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		testJWKGenService = cryptoutilJose.RequireNewForTest(testCtx, testTelemetryService)
		defer testJWKGenService.Shutdown()

		testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
		defer testSQLRepository.Shutdown()

		testOrmRepository = RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
		defer testOrmRepository.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestSQLTransaction_PanicRecovery(t *testing.T) {
	defer func() {
		if recoverValue := recover(); recoverValue != nil {
			require.NotNil(t, recoverValue)
		}
	}()

	panicErr := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(ormTransaction *OrmTransaction) error {
		require.NotNil(t, ormTransaction)
		panic("simulated panic during transaction")
	})
	require.Error(t, panicErr)
	require.EqualError(t, panicErr, "simulated panic during transaction")
}

func TestSQLTransaction_BeginAlreadyStartedFailure(t *testing.T) {
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

func TestSQLTransaction_CommitNotStartedFailure(t *testing.T) {
	ormTransaction := &OrmTransaction{ormRepository: testOrmRepository}

	commitErr := ormTransaction.commit()
	require.Error(t, commitErr)
	require.EqualError(t, commitErr, "can't commit because transaction not active")
}

func TestSQLTransaction_RollbackNotStartedFailure(t *testing.T) {
	ormTransaction := &OrmTransaction{ormRepository: testOrmRepository}

	rollbackErr := ormTransaction.rollback()
	require.Error(t, rollbackErr)
	require.EqualError(t, rollbackErr, "can't rollback because transaction not active")
}

func TestSQLTransaction_BeginWithReadOnly(t *testing.T) {
	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(ormTransaction *OrmTransaction) error {
		require.NotNil(t, ormTransaction)
		require.Equal(t, ReadOnly, *ormTransaction.Mode())

		return nil
	})
	require.NoError(t, err)
}

func TestSQLTransaction_RollbackOnError(t *testing.T) {
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(ormTransaction *OrmTransaction) error {
		require.NotNil(t, ormTransaction)
		require.Equal(t, ReadWrite, *ormTransaction.Mode())

		return fmt.Errorf("intentional failure")
	})
	require.Error(t, err)
	require.EqualError(t, err, "failed to execute transaction: intentional failure")
}

func TestSQLTransaction_Success(t *testing.T) {
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

		addedElasticKeys := []*ElasticKey{}
		addedKeys := []*MaterialKey{}
		err := testOrmRepository.WithTransaction(testCtx, testCase.txMode, func(ormTransaction *OrmTransaction) error {
			require.NotNil(t, ormTransaction)
			require.NotNil(t, ormTransaction.ID())
			require.NotNil(t, ormTransaction.Context())
			require.Equal(t, testCase.txMode, *ormTransaction.Mode())

			uuidV7 := testJWKGenService.GenerateUUIDv7()

			elasticKey, err := BuildElasticKey(*uuidV7, "Elastic Key Name "+uuidV7.String(), "Elastic Key Description "+uuidV7.String(), cryptoutilOpenapiModel.Internal, cryptoutilOpenapiModel.A256GCMDir, true, true, true, string(cryptoutilOpenapiModel.Creating))
			cryptoutilAppErr.RequireNoError(err, "failed to create AES 256 Elastic Key")
			err = ormTransaction.AddElasticKey(elasticKey)
			cryptoutilAppErr.RequireNoError(err, "failed to add AES 256 Elastic Key")

			addedElasticKeys = append(addedElasticKeys, elasticKey)

			multipleByteSlices, err := cryptoutilRandom.GenerateMultipleBytes(numMaterialKeys, 32)
			cryptoutilAppErr.RequireNoError(err, "failed to generate AES 256 key materials")

			for nextKeyID := 1; nextKeyID <= numMaterialKeys; nextKeyID++ {
				now := time.Now().UTC()
				materialKeyID := testJWKGenService.GenerateUUIDv7()
				key := MaterialKey{
					ElasticKeyID:                  elasticKey.ElasticKeyID,
					MaterialKeyID:                 *materialKeyID,
					MaterialKeyClearPublic:        nil,
					MaterialKeyEncryptedNonPublic: multipleByteSlices[nextKeyID-1],
					MaterialKeyGenerateDate:       &now,
					MaterialKeyImportDate:         nil,
					MaterialKeyExpirationDate:     nil,
					MaterialKeyRevocationDate:     nil,
				}

				err = ormTransaction.AddElasticKeyMaterialKey(&key)
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

		for _, addedElasticKey := range addedElasticKeys {
			err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(ormTransaction *OrmTransaction) error {
				require.NotNil(t, ormTransaction)
				require.NotNil(t, ormTransaction.ID())
				require.NotNil(t, ormTransaction.Context())
				require.Equal(t, ReadOnly, *ormTransaction.Mode())

				retrievedElasticKey, err := ormTransaction.GetElasticKey(&addedElasticKey.ElasticKeyID)
				if err != nil {
					return fmt.Errorf("failed to get Elastic Key: %w", err)
				}

				require.Equal(t, addedElasticKey, retrievedElasticKey)

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

				retrievedKey, err := ormTransaction.GetElasticKeyMaterialKeyVersion(&addedKey.ElasticKeyID, &addedKey.MaterialKeyID)
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
