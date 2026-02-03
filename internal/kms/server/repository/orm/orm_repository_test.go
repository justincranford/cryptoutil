//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

package orm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestOrmRepository_Shutdown_NoOp tests shutdown (no-op implementation).
func TestOrmRepository_Shutdown_NoOp(t *testing.T) {
	// Use testOrmRepository which is created from template Core
	require.NotNil(t, testOrmRepository)

	// Shutdown should not panic (it's a no-op).
	require.NotPanics(t, func() {
		testOrmRepository.Shutdown()
	})

	// Can call multiple times without issue.
	require.NotPanics(t, func() {
		testOrmRepository.Shutdown()
	})
}

// TestNewOrmRepository_NilChecks tests nil parameter validation for GORM constructor.
func TestNewOrmRepository_NilChecks(t *testing.T) {
	// Get gormDB from testOrmRepository for success case
	testGormDB := testOrmRepository.GormDB()
	require.NotNil(t, testGormDB)

	t.Run("Nil telemetry service", func(t *testing.T) {
		repo, err := NewOrmRepository(testCtx, nil, testGormDB, testJWKGenService, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "telemetryService must be non-nil")
		require.Nil(t, repo)
	})

	t.Run("Nil gormDB", func(t *testing.T) {
		repo, err := NewOrmRepository(testCtx, testTelemetryService, nil, testJWKGenService, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "gormDB must be non-nil")
		require.Nil(t, repo)
	})

	t.Run("Nil JWK gen service", func(t *testing.T) {
		repo, err := NewOrmRepository(testCtx, testTelemetryService, testGormDB, nil, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "jwkGenService must be non-nil")
		require.Nil(t, repo)
	})

	t.Run("Success with all valid parameters", func(t *testing.T) {
		repo, err := NewOrmRepository(testCtx, testTelemetryService, testGormDB, testJWKGenService, true)
		require.NoError(t, err)
		require.NotNil(t, repo)
		require.NotNil(t, repo.GormDB())

		// Verify it can still shutdown without panic
		require.NotPanics(t, func() {
			repo.Shutdown()
		})
	})
}

// TestOrmRepository_GormDB tests the GormDB getter.
func TestOrmRepository_GormDB(t *testing.T) {
	gormDB := testOrmRepository.GormDB()
	require.NotNil(t, gormDB)
}

// TestNewOrmRepository_NilContext tests that nil context is accepted (context is not validated).
func TestNewOrmRepository_NilContext(t *testing.T) {
	testGormDB := testOrmRepository.GormDB()
	require.NotNil(t, testGormDB)

	//nolint:staticcheck // Test needs nil to verify context is not validated
	repo, err := NewOrmRepository(nil, testTelemetryService, testGormDB, testJWKGenService, false)
	require.NoError(t, err)
	require.NotNil(t, repo)
	repo.Shutdown()
}

// TestNewOrmRepository_UsableForTransactions verifies repositories work with transactions.
func TestNewOrmRepository_UsableForTransactions(t *testing.T) {
	testGormDB := testOrmRepository.GormDB()
	require.NotNil(t, testGormDB)

	repo, err := NewOrmRepository(context.Background(), testTelemetryService, testGormDB, testJWKGenService, true)
	require.NoError(t, err)

	require.NotNil(t, repo)
	defer repo.Shutdown()

	// Verify the repository can execute transactions
	err = repo.WithTransaction(context.Background(), ReadWrite, func(_ *OrmTransaction) error {
		// Simple no-op transaction to verify it works
		return nil
	})
	require.NoError(t, err)
}
