// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository

import (
	"context"
	"database/sql"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	"github.com/stretchr/testify/require"
)

// TestSQLRepository_Shutdown tests the shutdown functionality.
func TestSQLRepository_Shutdown(t *testing.T) {
	// Note: Cannot easily test shutdown without breaking other tests
	// since testSQLRepository is shared. This would require creating
	// a separate instance just for this test.
	// The shutdown functionality is already tested in TestMain's defer.
	t.Skip("Shutdown is tested in TestMain defer; testing here would break shared testSQLRepository")
}

// TestExtractSchemaFromURL tests schema extraction from database URLs.
func TestExtractSchemaFromURL(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedSchema string
	}{
		{
			name:           "postgres_with_schema",
			url:            "postgres://user:pass@localhost:5432/dbname?sslmode=disable&search_path=myschema",
			expectedSchema: "myschema",
		},
		{
			name:           "postgres_without_schema",
			url:            "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
			expectedSchema: "",
		},
		{
			name:           "sqlite_memory",
			url:            ":memory:",
			expectedSchema: "",
		},
		{
			name:           "malformed_url",
			url:            "not a valid url",
			expectedSchema: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			schema := extractSchemaFromURL(tc.url)
			require.Equal(t, tc.expectedSchema, schema)
		})
	}
}

// TestNewSQLRepository_NilChecks tests nil parameter validation.
func TestNewSQLRepository_NilChecks(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		telemetry *cryptoutilSharedTelemetry.TelemetryService
		settings  *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
		errorMsg  string
	}{
		{
			name:      "nil_context",
			ctx:       nil,
			telemetry: testTelemetryService,
			settings:  testSettings,
			errorMsg:  "ctx must be non-nil",
		},
		{
			name:      "nil_telemetry",
			ctx:       testCtx,
			telemetry: nil,
			settings:  testSettings,
			errorMsg:  "telemetryService must be non-nil",
		},
		{
			name:      "nil_settings",
			ctx:       testCtx,
			telemetry: testTelemetryService,
			settings:  nil,
			errorMsg:  "settings must be non-nil",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo, err := NewSQLRepository(tc.ctx, tc.telemetry, tc.settings)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errorMsg)
			require.Nil(t, repo)
		})
	}
}

// TestSQLTransaction_Context tests the Context method.
func TestSQLTransaction_Context(t *testing.T) {
	err := testSQLRepository.WithTransaction(testCtx, false, func(tx *SQLTransaction) error {
		txCtx := tx.Context()
		require.NotNil(t, txCtx)
		require.Equal(t, testCtx, txCtx)

		return nil
	})
	require.NoError(t, err)
}

// TestLogSchema tests schema logging functionality.
func TestLogSchema(t *testing.T) {
	// This test verifies that LogSchema doesn't panic or error.
	err := LogSchema(testSQLRepository)
	require.NoError(t, err)
}

// TestApplyEmbeddedSQLMigrations_UnsupportedDBType tests unsupported database type handling.
func TestApplyEmbeddedSQLMigrations_UnsupportedDBType(t *testing.T) {
	// Create a new in-memory SQLite database for testing migrations.
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	// Test with unsupported DB type (should fail).
	err = ApplyEmbeddedSQLMigrations(testTelemetryService, db, "unsupported")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported database driver")
}
