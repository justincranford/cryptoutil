//go:build integration
// +build integration

// Copyright (c) 2025-2026 Justin Cranford.
//
// ORM transaction tests using SQLite for integration testing.
//

package orm

import (
	"fmt"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilAppsFrameworkServiceServerRepositoryOrm "cryptoutil/internal/apps-framework/service/server/repository/orm"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"
	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestSQLTransaction_PanicRecovery(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(ormTransaction *OrmTransaction) error {
		require.NotNil(t, ormTransaction)
		require.Equal(t, ReadWrite, *ormTransaction.Mode())

		err := ormTransaction.Begin(testCtx, ReadWrite)
		require.Error(t, err)

		return err
	})
	require.Error(t, err)
	require.EqualError(t, err, "failed to execute transaction: transaction already started")
}

func TestSQLTransaction_CommitNotStartedFailure(t *testing.T) {
	t.Parallel()

	ormTransaction := cryptoutilAppsFrameworkServiceServerRepositoryOrm.NewOrmTransactionWithRepository(testOrmRepository)

	commitErr := ormTransaction.Commit()
	require.Error(t, commitErr)
	require.EqualError(t, commitErr, "can't commit because transaction not active")
}

func TestSQLTransaction_RollbackNotStartedFailure(t *testing.T) {
	t.Parallel()

	ormTransaction := cryptoutilAppsFrameworkServiceServerRepositoryOrm.NewOrmTransactionWithRepository(testOrmRepository)

	rollbackErr := ormTransaction.Rollback()
	require.Error(t, rollbackErr)
	require.EqualError(t, rollbackErr, "can't rollback because transaction not active")
}

func TestSQLTransaction_BeginWithReadOnly(t *testing.T) {
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(ormTransaction *OrmTransaction) error {
		require.NotNil(t, ormTransaction)
		require.Equal(t, ReadOnly, *ormTransaction.Mode())

		return nil
	})
	require.NoError(t, err)
}

func TestSQLTransaction_RollbackOnError(t *testing.T) {
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(ormTransaction *OrmTransaction) error {
		require.NotNil(t, ormTransaction)
		require.Equal(t, ReadWrite, *ormTransaction.Mode())

		return fmt.Errorf("intentional failure")
	})
	require.Error(t, err)
	require.EqualError(t, err, "failed to execute transaction: intentional failure")
}

// Sequential: uses shared package-level SQLite fixture state via CleanupDatabase.
func TestSQLTransaction_Success(t *testing.T) {
	CleanupDatabase(t, testOrmRepository, KMSCleanupTables)

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
			tenantID := googleUuid.New()

			elasticKey, err := BuildElasticKey(tenantID, *uuidV7, "Elastic Key Name "+uuidV7.String(), "Elastic Key Description "+uuidV7.String(), cryptoutilOpenapiModel.Internal, cryptoutilOpenapiModel.A256GCMDir, true, true, true, string(cryptoutilKmsServer.Active))
			cryptoutilSharedApperr.RequireNoError(err, "failed to create AES 256 Elastic Key")
			err = AddElasticKey(ormTransaction.GormTx(), testTelemetryService.Slogger, elasticKey)
			cryptoutilSharedApperr.RequireNoError(err, "failed to add AES 256 Elastic Key")

			addedElasticKeys = append(addedElasticKeys, elasticKey)

			multipleByteSlices, err := cryptoutilSharedUtilRandom.GenerateMultipleBytes(numMaterialKeys, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
			cryptoutilSharedApperr.RequireNoError(err, "failed to generate AES 256 key materials")

			for nextKeyID := 1; nextKeyID <= numMaterialKeys; nextKeyID++ {
				now := time.Now().UTC().UnixMilli()
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

				if err = AddElasticKeyMaterialKey(ormTransaction.GormTx(), testTelemetryService.Slogger, &key); err != nil {
					return fmt.Errorf("failed to add Key: %w", err)
				}
			}

			return nil
		})

		testTelemetryService.Slogger.Info("Happy path test case result", "mode", testCase.txMode, "expectError", testCase.expectError, cryptoutilSharedMagic.StringError, err)

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

				retrievedElasticKey, err := GetElasticKey(ormTransaction.GormTx(), testTelemetryService.Slogger, addedElasticKey.TenantID, &addedElasticKey.ElasticKeyID)
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

				retrievedKey, err := GetElasticKeyMaterialKeyVersion(ormTransaction.GormTx(), testTelemetryService.Slogger, &addedKey.ElasticKeyID, &addedKey.MaterialKeyID)
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
