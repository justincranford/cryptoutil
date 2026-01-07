// Copyright (c) 2025 Justin Cranford

package orm

import (
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
