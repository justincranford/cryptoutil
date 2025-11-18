package orm

import (
	"errors"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestErrorMapping tests toAppErr error mapping functionality.
func TestErrorMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		operation     func(tx *OrmTransaction) error
		expectedError string
	}{
		{
			name: "Add elastic key with nil UUID",
			operation: func(tx *OrmTransaction) error {
				key := &ElasticKey{}
				key.ElasticKeyID = googleUuid.Nil
				return tx.AddElasticKey(key)
			},
			expectedError: "failed to add Elastic Key",
		},
		{
			name: "Get elastic key with nil UUID",
			operation: func(tx *OrmTransaction) error {
				nilUUID := googleUuid.Nil
				_, err := tx.GetElasticKey(&nilUUID)
				return err
			},
			expectedError: "failed to get Elastic Key",
		},
		{
			name: "Update elastic key status with nil UUID",
			operation: func(tx *OrmTransaction) error {
				return tx.UpdateElasticKeyStatus(googleUuid.Nil, "Active")
			},
			expectedError: "failed to update Elastic Key status",
		},
		{
			name: "Add material key with nil elastic key ID",
			operation: func(tx *OrmTransaction) error {
				key := &MaterialKey{}
				key.ElasticKeyID = googleUuid.Nil
				key.MaterialKeyID = googleUuid.New()
				return tx.AddElasticKeyMaterialKey(key)
			},
			expectedError: "failed to add Material Key",
		},
		{
			name: "Add material key with nil material key ID",
			operation: func(tx *OrmTransaction) error {
				key := &MaterialKey{}
				key.ElasticKeyID = googleUuid.New()
				key.MaterialKeyID = googleUuid.Nil
				return tx.AddElasticKeyMaterialKey(key)
			},
			expectedError: "failed to add Material Key",
		},
		{
			name: "Get material key with nil elastic key ID",
			operation: func(tx *OrmTransaction) error {
				nilUUID := googleUuid.Nil
				validUUID := googleUuid.New()
				_, err := tx.GetElasticKeyMaterialKeyVersion(&nilUUID, &validUUID)
				return err
			},
			expectedError: "failed to get Material Key by Elastic Key ID and Material Key ID",
		},
		{
			name: "Get material key with nil material key ID",
			operation: func(tx *OrmTransaction) error {
				validUUID := googleUuid.New()
				nilUUID := googleUuid.Nil
				_, err := tx.GetElasticKeyMaterialKeyVersion(&validUUID, &nilUUID)
				return err
			},
			expectedError: "failed to get Material Key by Elastic Key ID and Material Key ID",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
				return tc.operation(tx)
			})

			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

// TestTransactionErrorHandling tests transaction error scenarios.
func TestTransactionErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("Transaction already started error", func(t *testing.T) {
		t.Parallel()

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Attempt to start nested transaction - should fail.
			return testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx2 *OrmTransaction) error {
				return nil
			})
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "transaction already started")
	})

	t.Run("Intentional rollback", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("intentional failure")
		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add key.
			key := &BarrierRootKey{}
			key.SetUUID(googleUuid.New())
			key.SetEncrypted("test-data")
			key.SetKEKUUID(googleUuid.New())
			err := tx.AddRootKey(key)
			require.NoError(t, err)

			// Return error to trigger rollback.
			return expectedErr
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "intentional failure")
	})
}
