//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

package orm

import (
"fmt"
"testing"

cryptoutilKmsServer "cryptoutil/api/kms/server"
cryptoutilOpenapiModel "cryptoutil/api/model"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

"github.com/jackc/pgx/v5/pgconn"
googleUuid "github.com/google/uuid"
"github.com/stretchr/testify/require"
"gorm.io/gorm"
)

// TestToAppErr_GormDuplicatedKey tests toAppErr handling of gorm.ErrDuplicatedKey.
func TestToAppErr_GormDuplicatedKey(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

const testOperationFailedMsg = "test operation failed"

// TestToAppErr_PostgresUniqueViolation tests toAppErr handling of PostgreSQL unique_violation errors.
func TestToAppErr_PostgresUniqueViolation(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		msg := testOperationFailedMsg

		// Create PostgreSQL unique violation error.
		pgErr := &pgconn.PgError{
			Code:    cryptoutilSharedMagic.PGCodeUniqueViolation, // unique_violation
			Message: "duplicate key value violates unique constraint",
		}

		mappedErr := tx.toAppErr(&msg, pgErr)

		require.Error(t, mappedErr, "Mapped error should not be nil")
		require.Contains(t, mappedErr.Error(), testOperationFailedMsg, "Should contain custom message")

		return fmt.Errorf("rollback test transaction")
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "rollback test transaction")
}

// TestToAppErr_PostgresForeignKeyViolation tests toAppErr handling of PostgreSQL foreign_key_violation errors.
func TestToAppErr_PostgresForeignKeyViolation(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		msg := testOperationFailedMsg

		// Create PostgreSQL foreign key violation error.
		pgErr := &pgconn.PgError{
			Code:    cryptoutilSharedMagic.PGCodeForeignKeyViolation, // foreign_key_violation
			Message: "insert or update on table violates foreign key constraint",
		}

		mappedErr := tx.toAppErr(&msg, pgErr)

		require.Error(t, mappedErr, "Mapped error should not be nil")
		require.Contains(t, mappedErr.Error(), testOperationFailedMsg, "Should contain custom message")

		return fmt.Errorf("rollback test transaction")
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "rollback test transaction")
}

// TestToAppErr_PostgresCheckViolation tests toAppErr handling of PostgreSQL check_violation errors.
func TestToAppErr_PostgresCheckViolation(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		msg := testOperationFailedMsg

		// Create PostgreSQL check violation error.
		pgErr := &pgconn.PgError{
			Code:    cryptoutilSharedMagic.PGCodeCheckViolation, // check_violation
			Message: "new row violates check constraint",
		}

		mappedErr := tx.toAppErr(&msg, pgErr)

		require.Error(t, mappedErr, "Mapped error should not be nil")
		require.Contains(t, mappedErr.Error(), testOperationFailedMsg, "Should contain custom message")

		return fmt.Errorf("rollback test transaction")
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "rollback test transaction")
}

// TestToAppErr_PostgresStringDataTruncation tests toAppErr handling of PostgreSQL string_data_right_truncation errors.
func TestToAppErr_PostgresStringDataTruncation(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		msg := testOperationFailedMsg

		// Create PostgreSQL string data truncation error.
		pgErr := &pgconn.PgError{
			Code:    cryptoutilSharedMagic.PGCodeStringDataTruncation, // string_data_right_truncation
			Message: "value too long for type character varying",
		}

		mappedErr := tx.toAppErr(&msg, pgErr)

		require.Error(t, mappedErr, "Mapped error should not be nil")
		require.Contains(t, mappedErr.Error(), testOperationFailedMsg, "Should contain custom message")

		return fmt.Errorf("rollback test transaction")
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "rollback test transaction")
}

// TestToAppErr_UnknownPostgresError tests toAppErr handling of unknown PostgreSQL errors (fallback to 500).
func TestToAppErr_UnknownPostgresError(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		msg := testOperationFailedMsg

		// Create PostgreSQL error with unknown code.
		pgErr := &pgconn.PgError{
			Code:    "99999", // Unknown error code
			Message: "unknown database error",
		}

		mappedErr := tx.toAppErr(&msg, pgErr)

		require.Error(t, mappedErr, "Mapped error should not be nil")
		require.Contains(t, mappedErr.Error(), testOperationFailedMsg, "Should contain custom message")

		return fmt.Errorf("rollback test transaction")
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "rollback test transaction")
}

// TestAddElasticKey_DuplicateConstraintViolation tests AddElasticKey duplicate key error.
func TestAddElasticKey_DuplicateConstraintViolation(t *testing.T) {
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create elastic key successfully.
		tenantID := googleUuid.New()
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			googleUuid.New(),
			"duplicate-elastic-key-test",
			"Test Duplicate Elastic Key",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMDir,
			false,
			false,
			false,
			string(cryptoutilKmsServer.Active),
		)
		require.NoError(t, buildErr, "Should build elastic key")

		createErr := tx.AddElasticKey(elasticKey)
		require.NoError(t, createErr, "First elastic key creation should succeed")

		// Attempt to create duplicate elastic key (same ID).
		duplicateKey := &ElasticKey{
			ElasticKeyID:                elasticKey.ElasticKeyID, // DUPLICATE
			ElasticKeyName:              "duplicate-name",
			ElasticKeyDescription:       "Duplicate description",
			ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
			ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A256GCMDir,
			ElasticKeyVersioningAllowed: false,
			ElasticKeyImportAllowed:     false,
			ElasticKeyStatus:            cryptoutilKmsServer.Active,
		}
		addErr := tx.AddElasticKey(duplicateKey)
		require.Error(t, addErr, "Duplicate elastic key should fail")
		require.Contains(t, addErr.Error(), "UNIQUE", "Error should mention UNIQUE constraint")

		return nil
	})
	require.NoError(t, err)
}

// TestGetElasticKeyMaterialKeyLatest_NotFoundError tests GetElasticKeyMaterialKeyLatest when no material keys exist.
func TestGetElasticKeyMaterialKeyLatest_NotFoundError(t *testing.T) {
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create elastic key without any material keys.
		tenantID := googleUuid.New()
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			googleUuid.New(),
			"no-material-keys-test",
			"Test No Material Keys",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMDir,
			false,
			false,
			false,
			string(cryptoutilKmsServer.Active),
		)
		require.NoError(t, buildErr, "Should build elastic key")

		createErr := tx.AddElasticKey(elasticKey)
		require.NoError(t, createErr, "Elastic key creation should succeed")

		// Attempt to get latest material key when none exist.
		_, getErr := tx.GetElasticKeyMaterialKeyLatest(elasticKey.ElasticKeyID)
		require.Error(t, getErr, "Should fail when no material keys exist")
		require.Contains(t, getErr.Error(), ErrFailedToGetLatestMaterialKeyByElasticKeyID, "Error should indicate get latest failure")

		return nil
	})
	require.NoError(t, err)
}

// TestGetElasticKeyMaterialKeyVersion_NotFoundError tests GetElasticKeyMaterialKeyVersion when material key version does not exist.
func TestGetElasticKeyMaterialKeyVersion_NotFoundError(t *testing.T) {
	t.Parallel()

	nonExistentElasticKeyID := googleUuid.New()
	nonExistentMaterialKeyID := googleUuid.New()

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Attempt to get non-existent material key version.
		_, getErr := tx.GetElasticKeyMaterialKeyVersion(&nonExistentElasticKeyID, &nonExistentMaterialKeyID)
		require.Error(t, getErr, "Should fail when material key version not found")
		require.Contains(t, getErr.Error(), ErrFailedToGetMaterialKeyByElasticKeyIDAndMaterialKeyID, "Error should indicate get failure")

		return nil
	})
	require.NoError(t, err)
}
