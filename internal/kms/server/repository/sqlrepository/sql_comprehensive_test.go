// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const requestIDKey contextKey = "requestID"

// TestSQLRepository_HealthCheck tests the health check functionality with parameterized test cases.
func TestSQLRepository_HealthCheck(t *testing.T) {
	tests := []struct {
		name           string
		ctx            context.Context
		expectError    bool
		checkFields    []string
		expectedStatus string
	}{
		{
			name:           "valid_health_check",
			ctx:            context.Background(),
			expectError:    false,
			checkFields:    []string{"status", "db_type", "open_connections"},
			expectedStatus: "ok",
		},
		{
			name:           "health_check_with_request_id",
			ctx:            context.WithValue(context.Background(), requestIDKey, "test-request-id"),
			expectError:    false,
			checkFields:    []string{"status", "db_type"},
			expectedStatus: "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := testSQLRepository.HealthCheck(tt.ctx)

			if tt.expectError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			// Check expected fields exist
			for _, field := range tt.checkFields {
				require.Contains(t, result, field, "Health check result should contain field: %s", field)
			}

			// Verify status
			if tt.expectedStatus != "" {
				require.Equal(t, tt.expectedStatus, result["status"])
			}

			// Verify db_type is set
			dbType, ok := result["db_type"].(string)
			require.True(t, ok, "db_type should be a string")
			require.NotEmpty(t, dbType)
		})
	}
}

// TestSQLRepository_GetDBType tests the database type retrieval.
func TestSQLRepository_GetDBType(t *testing.T) {
	dbType := testSQLRepository.GetDBType()
	require.NotEmpty(t, dbType)
	require.Contains(t, []SupportedDBType{DBTypeSQLite}, dbType)
}

// TestSQLTransaction_ParameterizedScenarios tests various transaction scenarios.
func TestSQLTransaction_ParameterizedScenarios(t *testing.T) {
	tests := []struct {
		name        string
		readOnly    bool
		operation   func(*SQLTransaction) error
		expectError bool
		errorMsg    string
	}{
		{
			name:     "read_write_transaction_success",
			readOnly: false,
			operation: func(tx *SQLTransaction) error {
				require.NotNil(t, tx)
				require.False(t, tx.IsReadOnly())

				return nil
			},
			expectError: false,
		},
		{
			name:     "read_write_transaction_error",
			readOnly: false,
			operation: func(tx *SQLTransaction) error {
				require.NotNil(t, tx)

				return errors.New("intentional test error")
			},
			expectError: true,
			errorMsg:    "intentional test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testSQLRepository.WithTransaction(testCtx, tt.readOnly, tt.operation)

			if tt.expectError {
				require.Error(t, err)

				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestSQLTransaction_ConcurrentOperations tests concurrent transaction handling.
func TestSQLTransaction_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent operations test in short mode")
	}

	// Test that multiple concurrent transactions can execute
	numTransactions := 10
	errChan := make(chan error, numTransactions)

	for i := 0; i < numTransactions; i++ {
		go func(_ int) {
			err := testSQLRepository.WithTransaction(testCtx, false, func(tx *SQLTransaction) error {
				require.NotNil(t, tx)
				require.False(t, tx.IsReadOnly())
				// Simulate some work
				return nil
			})
			errChan <- err
		}(i)
	}

	// Collect results
	for i := 0; i < numTransactions; i++ {
		err := <-errChan
		require.NoError(t, err, "Transaction %d should succeed", i)
	}
}

// TestSQLRepository_MultipleHealthChecks tests multiple health checks in sequence.
func TestSQLRepository_MultipleHealthChecks(t *testing.T) {
	numChecks := 5
	for i := 0; i < numChecks; i++ {
		result, err := testSQLRepository.HealthCheck(testCtx)
		require.NoError(t, err, "Health check %d should succeed", i)
		require.NotNil(t, result)
		require.Equal(t, "ok", result["status"])
	}
}

// TestSQLTransaction_PanicHandling tests panic recovery in various scenarios.
func TestSQLTransaction_PanicHandling(t *testing.T) {
	tests := []struct {
		name      string
		panicMsg  string
		operation func(*SQLTransaction)
	}{
		{
			name:     "panic_with_string",
			panicMsg: "test panic",
			operation: func(_ *SQLTransaction) {
				panic("test panic")
			},
		},
		{
			name:     "panic_with_error",
			panicMsg: "error panic",
			operation: func(_ *SQLTransaction) {
				panic(errors.New("error panic"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var panicErr error

			func() {
				defer func() {
					if r := recover(); r != nil {
						panicErr = errors.New("panic recovered")
					}
				}()

				err := testSQLRepository.WithTransaction(testCtx, false, func(tx *SQLTransaction) error {
					tt.operation(tx)

					return nil
				})
				panicErr = err
			}()
			require.Error(t, panicErr)
		})
	}
}
