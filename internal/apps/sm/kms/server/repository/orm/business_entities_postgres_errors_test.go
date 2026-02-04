//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

package orm

import (
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

const testOperationFailedMsg = "test operation failed"

// TestToAppErr_PostgresUniqueViolation tests toAppErr handling of PostgreSQL unique_violation errors.
func TestToAppErr_PostgresUniqueViolation(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		msg := testOperationFailedMsg

		// Create PostgreSQL unique violation error.
		pgErr := &pgconn.PgError{
			Code:    "23505", // unique_violation
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
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		msg := testOperationFailedMsg

		// Create PostgreSQL foreign key violation error.
		pgErr := &pgconn.PgError{
			Code:    "23503", // foreign_key_violation
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
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		msg := testOperationFailedMsg

		// Create PostgreSQL check violation error.
		pgErr := &pgconn.PgError{
			Code:    "23514", // check_violation
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
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		msg := testOperationFailedMsg

		// Create PostgreSQL string data truncation error.
		pgErr := &pgconn.PgError{
			Code:    "22001", // string_data_right_truncation
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
