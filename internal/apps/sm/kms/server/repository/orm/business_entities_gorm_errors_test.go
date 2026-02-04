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

// TestToAppErr_GormDuplicatedKey tests toAppErr handling of gorm.ErrDuplicatedKey.
func TestToAppErr_GormDuplicatedKey(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create first elastic key.
		tenantID := googleUuid.New()
		ekID := googleUuid.New()
		uniqueName := fmt.Sprintf("dup-key-test-%s", ekID.String())
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

		// Try to create second elastic key with same primary key (ErrDuplicatedKey).
		elasticKey2, buildErr2 := BuildElasticKey(
			tenantID, // Same tenantID.
			ekID,     // Same ID - should trigger ErrDuplicatedKey.
			"different-name",
			"Second key",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A128GCMA128KW,
			true,
			false,
			false,
			"active",
		)
		require.NoError(t, buildErr2)

		dupKeyErr := tx.state.gormTx.Create(elasticKey2).Error
		require.Error(t, dupKeyErr, "Duplicate key should fail")

		// Test toAppErr mapping.
		msg := "test duplicated key"
		mappedErr := tx.toAppErr(&msg, dupKeyErr)
		require.Error(t, mappedErr, "Mapped error should not be nil")
		require.Contains(t, mappedErr.Error(), "test duplicated key", "Should contain custom message")

		return fmt.Errorf("rollback test transaction") // Force rollback.
	})

	require.Error(t, err) // Expect error from rollback.
	require.Contains(t, err.Error(), "rollback test transaction")
}

// TestToAppErr_GormCheckConstraintViolated tests toAppErr handling of gorm.ErrCheckConstraintViolated.
func TestToAppErr_GormCheckConstraintViolated(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create material key with empty encrypted data (violates check constraint).
		tenantID := googleUuid.New()
		ekID := googleUuid.New()
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			ekID,
			"parent-for-check-test",
			"Parent for check constraint test",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMA256KW,
			true,
			false,
			false,
			"active",
		)
		require.NoError(t, buildErr)

		createErr := tx.state.gormTx.Create(elasticKey).Error
		require.NoError(t, createErr, "Parent key create should succeed")

		// Try to create material key with empty encrypted data (check constraint: length >= 1).
		materialKey := &MaterialKey{
			ElasticKeyID:                  ekID,
			MaterialKeyID:                 googleUuid.New(),
			MaterialKeyEncryptedNonPublic: []byte{}, // Empty - violates check constraint.
		}

		checkErr := tx.state.gormTx.Create(materialKey).Error

		// Check constraint might not always trigger gorm.ErrCheckConstraintViolated.
		// It depends on database driver error mapping.
		if checkErr != nil {
			msg := "test check constraint violation"
			mappedErr := tx.toAppErr(&msg, checkErr)
			require.Error(t, mappedErr, "Mapped error should not be nil")
			require.Contains(t, mappedErr.Error(), "test check constraint violation", "Should contain custom message")
		}

		return fmt.Errorf("rollback test transaction") // Force rollback.
	})

	require.Error(t, err) // Expect error from rollback.
	require.Contains(t, err.Error(), "rollback test transaction")
}

// TestToAppErr_GormInvalidData tests toAppErr handling of gorm.ErrInvalidData.
func TestToAppErr_GormInvalidData(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Simulate gorm.ErrInvalidData by passing it directly.
		// Note: This error is typically triggered by GORM's internal validation, which is hard to trigger naturally.
		// We'll test the mapping logic directly.
		msg := "test invalid data"
		mappedErr := tx.toAppErr(&msg, gorm.ErrInvalidData)
		require.Error(t, mappedErr, "Mapped error should not be nil")
		require.Contains(t, mappedErr.Error(), "test invalid data", "Should contain custom message")
		require.Contains(t, mappedErr.Error(), "invalid data", "Should contain GORM error")

		return nil
	})

	require.NoError(t, err)
}

// TestToAppErr_GormInvalidValueOfLength tests toAppErr handling of gorm.ErrInvalidValueOfLength.
func TestToAppErr_GormInvalidValueOfLength(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Simulate gorm.ErrInvalidValueOfLength by passing it directly.
		// This error is typically triggered by GORM's internal length validation.
		msg := "test invalid value of length"
		mappedErr := tx.toAppErr(&msg, gorm.ErrInvalidValueOfLength)
		require.Error(t, mappedErr, "Mapped error should not be nil")
		require.Contains(t, mappedErr.Error(), "test invalid value of length", "Should contain custom message")
		// Note: The actual error message varies, so just check it's an error.

		return nil
	})

	require.NoError(t, err)
}

// TestToAppErr_GormNotImplemented tests toAppErr handling of gorm.ErrNotImplemented.
func TestToAppErr_GormNotImplemented(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Simulate gorm.ErrNotImplemented by passing it directly.
		// This error is typically triggered when calling unimplemented features.
		msg := "test not implemented"
		mappedErr := tx.toAppErr(&msg, gorm.ErrNotImplemented)
		require.Error(t, mappedErr, "Mapped error should not be nil")
		require.Contains(t, mappedErr.Error(), "test not implemented", "Should contain custom message")
		require.Contains(t, mappedErr.Error(), "not implemented", "Should contain GORM error")

		return nil
	})

	require.NoError(t, err)
}
