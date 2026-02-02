// Copyright (c) 2025 Justin Cranford

package orm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNewOrmRepository_NilChecks tests nil parameter validation.
func TestNewOrmRepository_NilChecks(t *testing.T) {
	t.Run("Nil context", func(t *testing.T) {
		//nolint:staticcheck // Test needs nil to validate error handling
		repo, err := NewOrmRepository(nil, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
		require.Error(t, err)
		require.Contains(t, err.Error(), "ctx must be non-nil")
		require.Nil(t, repo)
	})

	t.Run("Nil telemetry service", func(t *testing.T) {
		repo, err := NewOrmRepository(testCtx, nil, testSQLRepository, testJWKGenService, testSettings)
		require.Error(t, err)
		require.Contains(t, err.Error(), "telemetryService must be non-nil")
		require.Nil(t, repo)
	})

	t.Run("Nil SQL repository", func(t *testing.T) {
		repo, err := NewOrmRepository(testCtx, testTelemetryService, nil, testJWKGenService, testSettings)
		require.Error(t, err)
		require.Contains(t, err.Error(), "sqlRepository must be non-nil")
		require.Nil(t, repo)
	})

	t.Run("Nil JWK gen service", func(t *testing.T) {
		repo, err := NewOrmRepository(testCtx, testTelemetryService, testSQLRepository, nil, testSettings)
		require.Error(t, err)
		require.Contains(t, err.Error(), "jwkGenService must be non-nil")
		require.Nil(t, repo)
	})
}

// TestOrmRepository_Shutdown_NoOp tests shutdown (no-op implementation).
func TestOrmRepository_Shutdown_NoOp(t *testing.T) {
	// Create a fresh repository just for shutdown testing.
	repo := RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	require.NotNil(t, repo)

	// Shutdown should not panic (it's a no-op).
	require.NotPanics(t, func() {
		repo.Shutdown()
	})

	// Can call multiple times without issue.
	require.NotPanics(t, func() {
		repo.Shutdown()
	})
}

// TestNewOrmRepositoryFromGORM_NilChecks tests nil parameter validation for GORM constructor.
func TestNewOrmRepositoryFromGORM_NilChecks(t *testing.T) {
	// Get gormDB from testOrmRepository for success case
	testGormDB := testOrmRepository.GormDB()
	require.NotNil(t, testGormDB)

	t.Run("Nil telemetry service", func(t *testing.T) {
		repo, err := NewOrmRepositoryFromGORM(testCtx, nil, testGormDB, testJWKGenService, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "telemetryService must be non-nil")
		require.Nil(t, repo)
	})

	t.Run("Nil gormDB", func(t *testing.T) {
		repo, err := NewOrmRepositoryFromGORM(testCtx, testTelemetryService, nil, testJWKGenService, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "gormDB must be non-nil")
		require.Nil(t, repo)
	})

	t.Run("Nil JWK gen service", func(t *testing.T) {
		repo, err := NewOrmRepositoryFromGORM(testCtx, testTelemetryService, testGormDB, nil, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "jwkGenService must be non-nil")
		require.Nil(t, repo)
	})

	t.Run("Success with all valid parameters", func(t *testing.T) {
		repo, err := NewOrmRepositoryFromGORM(testCtx, testTelemetryService, testGormDB, testJWKGenService, true)
		require.NoError(t, err)
		require.NotNil(t, repo)
		require.NotNil(t, repo.GormDB())

		// sqlRepository should be nil when using FromGORM constructor
		require.Nil(t, repo.sqlRepository)

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

// TestNewOrmRepositoryFromGORM_NilContext tests that nil context is accepted (context is not validated).
func TestNewOrmRepositoryFromGORM_NilContext(t *testing.T) {
	testGormDB := testOrmRepository.GormDB()
	require.NotNil(t, testGormDB)

	//nolint:staticcheck // Test needs nil to verify context is not validated
	repo, err := NewOrmRepositoryFromGORM(nil, testTelemetryService, testGormDB, testJWKGenService, false)
	require.NoError(t, err)
	require.NotNil(t, repo)
	repo.Shutdown()
}

// TestNewOrmRepositoryFromGORM_UsableForTransactions verifies repositories from GORM constructor work with transactions.
func TestNewOrmRepositoryFromGORM_UsableForTransactions(t *testing.T) {
	testGormDB := testOrmRepository.GormDB()
	require.NotNil(t, testGormDB)

	repo, err := NewOrmRepositoryFromGORM(context.Background(), testTelemetryService, testGormDB, testJWKGenService, true)
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
