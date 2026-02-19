//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

package orm

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

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
