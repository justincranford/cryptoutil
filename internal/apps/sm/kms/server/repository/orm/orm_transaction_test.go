//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford
//
// NOTE: These tests require a PostgreSQL database and are skipped in CI without the integration tag.
//

package orm

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerApplication "cryptoutil/internal/apps/template/service/server/application"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedTelemetry "cryptoutil/internal/apps/template/service/telemetry"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"

	googleUuid "github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	testSettings         = cryptoutilAppsTemplateServiceConfig.RequireNewForTest("orm_transaction_test")
	testCtx              = context.Background()
	testTelemetryService *cryptoutilSharedTelemetry.TelemetryService
	testJWKGenService    *cryptoutilSharedCryptoJose.JWKGenService
	testTemplateCore     *cryptoutilAppsTemplateServiceServerApplication.Core
	testOrmRepository    *OrmRepository
	skipReadOnlyTxTests  = true // true for DBTypeSQLite, false for DBTypePostgres
	numMaterialKeys      = 10
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		// Start template Core which provides GORM directly with proper migrations
		var err error

		testTemplateCore, err = cryptoutilAppsTemplateServiceServerApplication.StartCore(testCtx, testSettings)
		if err != nil {
			panic(fmt.Sprintf("failed to start template core: %v", err))
		}

		defer func() {
			if testTemplateCore.ShutdownDBContainer != nil {
				testTemplateCore.ShutdownDBContainer()
			}

			testTemplateCore.Basic.Shutdown()
		}()

		testTelemetryService = testTemplateCore.Basic.TelemetryService
		testJWKGenService = testTemplateCore.Basic.JWKGenService

		// Apply template migrations (1001-1005 for barrier tables, sessions, etc.)
		sqlDB, err := testTemplateCore.DB.DB()
		if err != nil {
			panic(fmt.Sprintf("failed to get sql.DB from GORM: %v", err))
		}

		err = cryptoutilAppsTemplateServiceServerRepository.ApplyMigrationsFromFS(
			sqlDB,
			cryptoutilAppsTemplateServiceServerRepository.MigrationsFS,
			"migrations",
			"sqlite",
		)
		if err != nil {
			panic(fmt.Sprintf("failed to apply template migrations: %v", err))
		}

		// Apply KMS domain tables using GORM AutoMigrate.
		// This creates elastic_keys and material_keys tables without golang-migrate.
		err = testTemplateCore.DB.AutoMigrate(&ElasticKey{}, &MaterialKey{})
		if err != nil {
			panic(fmt.Sprintf("failed to apply KMS domain tables: %v", err))
		}

		// Use GORM directly from template Core (not SQLRepository)
		testOrmRepository = RequireNewForTest(testCtx, testTelemetryService, testTemplateCore.DB, testJWKGenService, testSettings.VerboseMode)
		defer testOrmRepository.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}

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

		err := ormTransaction.begin(testCtx, ReadWrite)
		require.Error(t, err)

		return err
	})
	require.Error(t, err)
	require.EqualError(t, err, "failed to execute transaction: transaction already started")
}

func TestSQLTransaction_CommitNotStartedFailure(t *testing.T) {
	t.Parallel()

	ormTransaction := &OrmTransaction{ormRepository: testOrmRepository}

	commitErr := ormTransaction.commit()
	require.Error(t, commitErr)
	require.EqualError(t, commitErr, "can't commit because transaction not active")
}

func TestSQLTransaction_RollbackNotStartedFailure(t *testing.T) {
	t.Parallel()

	ormTransaction := &OrmTransaction{ormRepository: testOrmRepository}

	rollbackErr := ormTransaction.rollback()
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

func TestSQLTransaction_Success(t *testing.T) {
	t.Parallel()

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
			err = ormTransaction.AddElasticKey(elasticKey)
			cryptoutilSharedApperr.RequireNoError(err, "failed to add AES 256 Elastic Key")

			addedElasticKeys = append(addedElasticKeys, elasticKey)

			multipleByteSlices, err := cryptoutilSharedUtilRandom.GenerateMultipleBytes(numMaterialKeys, 32)
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

				retrievedElasticKey, err := ormTransaction.GetElasticKey(addedElasticKey.TenantID, &addedElasticKey.ElasticKeyID)
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
