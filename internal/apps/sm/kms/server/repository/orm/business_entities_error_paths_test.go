//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

package orm

import (
	"errors"
	"fmt"
	"testing"

	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"

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

// toAppErr Error Mapping Tests

func TestOrmTransaction_toAppErr_GormRecordNotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		msg := "record not found test"

		mappedErr := tx.toAppErr(&msg, gorm.ErrRecordNotFound)

		require.Error(t, mappedErr)
		require.Contains(t, mappedErr.Error(), "record not found test")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_toAppErr_GormDuplicatedKey(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		msg := "duplicated key test"

		mappedErr := tx.toAppErr(&msg, gorm.ErrDuplicatedKey)

		require.Error(t, mappedErr)
		require.Contains(t, mappedErr.Error(), "duplicated key test")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_toAppErr_GormForeignKeyViolated(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		msg := "foreign key violated test"

		mappedErr := tx.toAppErr(&msg, gorm.ErrForeignKeyViolated)

		require.Error(t, mappedErr)
		require.Contains(t, mappedErr.Error(), "foreign key violated test")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_toAppErr_GormCheckConstraintViolated(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		msg := "check constraint violated test"

		mappedErr := tx.toAppErr(&msg, gorm.ErrCheckConstraintViolated)

		require.Error(t, mappedErr)
		require.Contains(t, mappedErr.Error(), "check constraint violated test")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_toAppErr_GormInvalidData(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		msg := "invalid data test"

		mappedErr := tx.toAppErr(&msg, gorm.ErrInvalidData)

		require.Error(t, mappedErr)
		require.Contains(t, mappedErr.Error(), "invalid data test")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_toAppErr_GormInvalidValueOfLength(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		msg := "invalid value of length test"

		mappedErr := tx.toAppErr(&msg, gorm.ErrInvalidValueOfLength)

		require.Error(t, mappedErr)
		require.Contains(t, mappedErr.Error(), "invalid value of length test")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_toAppErr_GormNotImplemented(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		msg := "not implemented test"

		mappedErr := tx.toAppErr(&msg, gorm.ErrNotImplemented)

		require.Error(t, mappedErr)
		require.Contains(t, mappedErr.Error(), "not implemented test")

		return nil
	})

	require.NoError(t, err)
}

// SQLite-specific error tests (trigger via actual database operations).

func TestOrmTransaction_toAppErr_SQLiteUniqueConstraint(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	// Create elastic key.
	tenantID := googleUuid.New()
	ekID := googleUuid.New()
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			ekID,
			"unique-constraint-test",
			"Test Unique Constraint",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMDir,
			false, false, false,
			string(cryptoutilKmsServer.Active),
		)
		require.NoError(t, buildErr)

		return tx.AddElasticKey(elasticKey)
	})
	require.NoError(t, err)

	// Try to create duplicate (should trigger UNIQUE constraint).
	err = testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		duplicateKey, buildErr := BuildElasticKey(
			tenantID, // Same tenantID.
			ekID,     // Same UUID - violates UNIQUE constraint.
			"unique-constraint-test-duplicate",
			"Test Duplicate",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMDir,
			false, false, false,
			string(cryptoutilKmsServer.Active),
		)
		require.NoError(t, buildErr)

		createErr := tx.AddElasticKey(duplicateKey)
		require.Error(t, createErr, "Should fail with UNIQUE constraint")
		require.Contains(t, createErr.Error(), ErrFailedToAddElasticKey)

		return nil
	})

	require.NoError(t, err)
}

// Default error fallback test.

func TestOrmTransaction_toAppErr_UnknownError(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		unknownErr := errors.New("unknown database error")
		msg := "unknown error test"

		mappedErr := tx.toAppErr(&msg, unknownErr)

		require.Error(t, mappedErr)
		require.Contains(t, mappedErr.Error(), "unknown error test")

		return nil
	})

	require.NoError(t, err)
}

// TestGetElasticKey_NotFoundError tests GetElasticKey when record does not exist.
func TestGetElasticKey_NotFoundError(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	nonExistentID := googleUuid.New()

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Attempt to get non-existent elastic key.
		_, getErr := tx.GetElasticKey(tenantID, &nonExistentID)
		require.Error(t, getErr, "Should fail when elastic key not found")
		require.Contains(t, getErr.Error(), ErrFailedToGetElasticKeyByElasticKeyID, "Error should indicate get failure")

		return nil
	})
	require.NoError(t, err)
}
