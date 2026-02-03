//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

package orm

import (
	"context"
	"testing"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestOrmRepository_VerboseMode tests verbose logging during transactions.
func TestOrmRepository_VerboseMode(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	// Enable verbose mode.
	testOrmRepository.verboseMode = true

	defer func() { testOrmRepository.verboseMode = false }() // Restore after test

	// Test verbose logging in begin/commit/rollback.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create an elastic key to trigger verbose logging.
		tenantID := googleUuid.New()
		ekID := googleUuid.New()
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			ekID,
			"verbose-test-key",
			"Test Verbose Mode",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMDir,
			false,
			false,
			false,
			string(cryptoutilKmsServer.Active),
		)
		require.NoError(t, buildErr)

		createErr := tx.AddElasticKey(elasticKey)
		require.NoError(t, createErr)

		// Rollback transaction by returning error (triggers verbose rollback logging).
		return context.Canceled
	})

	require.Error(t, err, "Transaction should fail")
	require.Contains(t, err.Error(), "failed to execute transaction", "Should contain transaction error")
}

// TestOrmTransaction_Begin_AlreadyStarted tests begin() error when transaction already started.
func TestOrmTransaction_Begin_AlreadyStarted(t *testing.T) {
	tx := &OrmTransaction{ormRepository: testOrmRepository}

	// Start transaction first time.
	err := tx.begin(testCtx, ReadWrite)
	require.NoError(t, err, "First begin should succeed")

	// Try to start again (should fail).
	err = tx.begin(testCtx, ReadWrite)
	require.Error(t, err, "Second begin should fail")
	require.Contains(t, err.Error(), "transaction already started", "Should contain specific error")

	// Cleanup: rollback transaction.
	rollbackErr := tx.rollback()
	require.NoError(t, rollbackErr)
}

// TestOrmTransaction_Commit_NotActive tests commit() error when transaction not active.
func TestOrmTransaction_Commit_NotActive(t *testing.T) {
	tx := &OrmTransaction{ormRepository: testOrmRepository}

	// Try to commit without starting transaction.
	err := tx.commit()
	require.Error(t, err, "Commit should fail when transaction not active")
	require.Contains(t, err.Error(), "can't commit because transaction not active", "Should contain specific error")
}

// TestOrmTransaction_Rollback_NotActive tests rollback() error when transaction not active.
func TestOrmTransaction_Rollback_NotActive(t *testing.T) {
	tx := &OrmTransaction{ormRepository: testOrmRepository}

	// Try to rollback without starting transaction.
	err := tx.rollback()
	require.Error(t, err, "Rollback should fail when transaction not active")
	require.Contains(t, err.Error(), "can't rollback because transaction not active", "Should contain specific error")
}

// TestOrmTransaction_DeferredRollback_OnFunctionError tests deferred rollback when transaction function fails.
func TestOrmTransaction_DeferredRollback_OnFunctionError(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	// Create transaction that fails in function.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create an elastic key.
		tenantID := googleUuid.New()
		ekID := googleUuid.New()
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			ekID,
			"deferred-rollback-test",
			"Test Deferred Rollback",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMDir,
			false,
			false,
			false,
			string(cryptoutilKmsServer.Active),
		)
		require.NoError(t, buildErr)

		createErr := tx.AddElasticKey(elasticKey)
		require.NoError(t, createErr)

		// Fail transaction - should trigger deferred rollback.
		return context.Canceled
	})

	require.Error(t, err, "Transaction should fail")
	require.Contains(t, err.Error(), "failed to execute transaction", "Should contain transaction error")
	require.Contains(t, err.Error(), "context canceled", "Should contain context error")

	// Verify elastic key was rolled back (not persisted).
	err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		filters := GetElasticKeysFilters{
			Name: []string{"deferred-rollback-test"},
		}
		elasticKeys, getErr := tx.GetElasticKeys(&filters)
		require.NoError(t, getErr)
		require.Empty(t, elasticKeys, "Elastic key should not exist after rollback")

		return nil
	})
	require.NoError(t, err)
}
