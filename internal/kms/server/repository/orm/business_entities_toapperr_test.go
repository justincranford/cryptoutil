// Copyright (c) 2025 Justin Cranford

package orm

import (
	"errors"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// toAppErr Error Mapping Tests

func TestOrmTransaction_toAppErr_GormRecordNotFound(t *testing.T) {
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
			string(cryptoutilOpenapiModel.Creating),
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
			string(cryptoutilOpenapiModel.Creating),
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
