// Copyright (c) 2025 Justin Cranford

package orm

import (
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestOrmTransaction_AutoCommit_CommitFailure tests that calling commit on an AutoCommit transaction fails.
func TestOrmTransaction_AutoCommit_CommitFailure(t *testing.T) {
	err := testOrmRepository.WithTransaction(testCtx, AutoCommit, func(tx *OrmTransaction) error {
		require.NotNil(t, tx)
		require.Equal(t, AutoCommit, *tx.Mode())

		// Manually call commit on autocommit transaction - should fail.
		commitErr := tx.commit()
		require.Error(t, commitErr, "Commit should fail for AutoCommit transaction")
		require.Contains(t, commitErr.Error(), "can't commit because transaction is autocommit")

		return nil
	})

	require.NoError(t, err)
}

// TestOrmTransaction_AutoCommit_RollbackFailure tests that calling rollback on an AutoCommit transaction fails.
func TestOrmTransaction_AutoCommit_RollbackFailure(t *testing.T) {
	err := testOrmRepository.WithTransaction(testCtx, AutoCommit, func(tx *OrmTransaction) error {
		require.NotNil(t, tx)
		require.Equal(t, AutoCommit, *tx.Mode())

		// Manually call rollback on autocommit transaction - should fail.
		rollbackErr := tx.rollback()
		require.Error(t, rollbackErr, "Rollback should fail for AutoCommit transaction")
		require.Contains(t, rollbackErr.Error(), "can't rollback because transaction is autocommit")

		return nil
	})

	require.NoError(t, err)
}

// TestOrmTransaction_AutoCommit_Success tests successful AutoCommit transaction.
func TestOrmTransaction_AutoCommit_Success(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, AutoCommit, func(tx *OrmTransaction) error {
		require.NotNil(t, tx)
		require.Equal(t, AutoCommit, *tx.Mode())
		require.NotNil(t, tx.state)
		require.NotNil(t, tx.state.gormTx)

		// AutoCommit mode should allow database operations.
		tenantID := googleUuid.New()
		ekID := googleUuid.New()
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			ekID,
			"autocommit-test-key",
			"Test AutoCommit",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMA256KW,
			false,
			false,
			false,
			"active",
		)
		require.NoError(t, buildErr)

		createErr := tx.state.gormTx.Create(elasticKey).Error
		require.NoError(t, createErr, "AutoCommit should allow database operations")

		return nil
	})

	require.NoError(t, err)
}
