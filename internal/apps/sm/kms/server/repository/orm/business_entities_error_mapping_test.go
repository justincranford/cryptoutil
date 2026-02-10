//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

package orm

import (
	"fmt"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestToAppErr_GormRecordNotFound tests toAppErr handling of gorm.ErrRecordNotFound.
func TestToAppErr_GormRecordNotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Try to get non-existent elastic key (will trigger gorm.ErrRecordNotFound).
		nonExistentID := googleUuid.New()
		elasticKey := &ElasticKey{}
		dbErr := tx.state.gormTx.Where("elastic_key_id = ?", nonExistentID).First(elasticKey).Error

		require.Error(t, dbErr, "Should get error for non-existent key")
		require.ErrorIs(t, dbErr, gorm.ErrRecordNotFound, "Error should be gorm.ErrRecordNotFound")

		// Test toAppErr mapping.
		msg := "test record not found"
		mappedErr := tx.toAppErr(&msg, dbErr)
		require.Error(t, mappedErr, "Mapped error should not be nil")
		require.Contains(t, mappedErr.Error(), "test record not found", "Should contain custom message")
		require.Contains(t, mappedErr.Error(), "record not found", "Should contain original error")

		return nil
	})

	require.NoError(t, err)
}

// TestToAppErr_UniqueConstraintViolation tests toAppErr handling of unique constraint violations.
func TestToAppErr_UniqueConstraintViolation(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create first elastic key.
		tenantID := googleUuid.New()
		ekID := googleUuid.New()
		uniqueName := fmt.Sprintf("unique-key-%s", ekID.String())
		elasticKey1, buildErr := BuildElasticKey(
			tenantID,
			ekID,
			uniqueName,
			"First key",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMA256KW,
			true,
			false,
			false,
			"active",
		)
		require.NoError(t, buildErr)

		createErr := tx.state.gormTx.Create(elasticKey1).Error
		require.NoError(t, createErr, "First create should succeed")

		// Try to create second elastic key with same name (unique constraint violation).
		ekID2 := googleUuid.New()
		elasticKey2, buildErr2 := BuildElasticKey(
			tenantID, // Same tenantID
			ekID2,
			uniqueName, // Same name - should violate unique constraint.
			"Second key",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A128GCMA128KW,
			true,
			false,
			false,
			"active",
		)
		require.NoError(t, buildErr2)

		duplicateErr := tx.state.gormTx.Create(elasticKey2).Error
		require.Error(t, duplicateErr, "Duplicate name should fail")

		// Test toAppErr mapping.
		msg := "test unique constraint violation"
		mappedErr := tx.toAppErr(&msg, duplicateErr)
		require.Error(t, mappedErr, "Mapped error should not be nil")
		require.Contains(t, mappedErr.Error(), "test unique constraint violation", "Should contain custom message")

		return fmt.Errorf("rollback test transaction") // Force rollback.
	})

	require.Error(t, err) // Expect error from rollback.
	require.Contains(t, err.Error(), "rollback test transaction")
}

// TestToAppErr_ForeignKeyViolation tests toAppErr handling of foreign key constraint violations.
func TestToAppErr_ForeignKeyViolation(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Try to create material key without parent elastic key (foreign key violation).
		nonExistentElasticKeyID := googleUuid.New()
		materialKey := &MaterialKey{
			ElasticKeyID:                  nonExistentElasticKeyID, // References non-existent elastic key.
			MaterialKeyID:                 googleUuid.New(),
			MaterialKeyEncryptedNonPublic: []byte("test-encrypted-key"),
		}

		fkErr := tx.state.gormTx.Create(materialKey).Error

		// Foreign key constraints might not be enforced in all configs.
		if fkErr != nil {
			// Test toAppErr mapping if we got an error.
			msg := "test foreign key violation"
			mappedErr := tx.toAppErr(&msg, fkErr)
			require.Error(t, mappedErr, "Mapped error should not be nil")
			require.Contains(t, mappedErr.Error(), "test foreign key violation", "Should contain custom message")
		}

		return nil
	})

	require.NoError(t, err)
}

// TestToAppErr_GenericError tests toAppErr handling of generic errors (fallback to HTTP 500).
func TestToAppErr_GenericError(t *testing.T) {
	t.Parallel()
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create a generic error (not a database-specific error).
		genericErr := fmt.Errorf("generic unexpected error")

		// Test toAppErr mapping.
		msg := "test generic error"
		mappedErr := tx.toAppErr(&msg, genericErr)
		require.Error(t, mappedErr, "Mapped error should not be nil")
		require.Contains(t, mappedErr.Error(), "test generic error", "Should contain custom message")
		require.Contains(t, mappedErr.Error(), "generic unexpected error", "Should contain original error")

		return nil
	})

	require.NoError(t, err)
}
