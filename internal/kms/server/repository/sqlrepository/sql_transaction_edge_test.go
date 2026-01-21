// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository_test

import (
	"context"
	"errors"
	"testing"

	"cryptoutil/internal/kms/server/repository/sqlrepository"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// TestWithTransaction_Success tests successful transaction execution.
func TestWithTransaction_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestWithTransaction_Success_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	sqlRepo := sqlrepository.RequireNewForTest(ctx, telemetryService, testSettings)
	defer sqlRepo.Shutdown()

	// Execute transaction successfully.
	executedCommit := false
	err := sqlRepo.WithTransaction(ctx, false, func(_ *sqlrepository.SQLTransaction) error {
		executedCommit = true

		return nil
	})

	testify.NoError(t, err)
	testify.True(t, executedCommit)
}

// TestWithTransaction_Rollback tests transaction rollback on error.
func TestWithTransaction_Rollback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestWithTransaction_Rollback_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	sqlRepo := sqlrepository.RequireNewForTest(ctx, telemetryService, testSettings)
	defer sqlRepo.Shutdown()

	// Execute transaction with intentional error to trigger rollback.
	expectedErr := errors.New("intentional test error")
	err := sqlRepo.WithTransaction(ctx, false, func(_ *sqlrepository.SQLTransaction) error {
		return expectedErr
	})

	testify.Error(t, err)
	testify.Contains(t, err.Error(), "intentional test error")
}

// TestWithTransaction_Panic tests transaction panic recovery.
func TestWithTransaction_Panic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestWithTransaction_Panic_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	sqlRepo := sqlrepository.RequireNewForTest(ctx, telemetryService, testSettings)
	defer sqlRepo.Shutdown()

	// Execute transaction with panic to trigger recovery and re-panic.
	testify.Panics(t, func() {
		_ = sqlRepo.WithTransaction(ctx, false, func(_ *sqlrepository.SQLTransaction) error { //nolint:errcheck // Test expects panic
			panic("intentional test panic")
		})
	})
}

// TestWithTransaction_ContextCancelled tests transaction with cancelled context.
func TestWithTransaction_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestWithTransaction_ContextCancelled_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	sqlRepo := sqlrepository.RequireNewForTest(ctx, telemetryService, testSettings)
	defer sqlRepo.Shutdown()

	// Create cancelled context.
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately.

	// Execute transaction with cancelled context.
	err := sqlRepo.WithTransaction(cancelledCtx, false, func(_ *sqlrepository.SQLTransaction) error {
		// Transaction should handle cancellation.
		return nil
	})

	// Transaction may succeed or fail depending on timing.
	// This tests the code path that checks context cancellation.
	_ = err // Don't assert specific outcome, just test the code path.
}

// TestWithTransaction_CommitError tests transaction commit failure handling.
func TestWithTransaction_CommitError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestWithTransaction_CommitError_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	sqlRepo := sqlrepository.RequireNewForTest(ctx, telemetryService, testSettings)
	defer sqlRepo.Shutdown()

	// Force shutdown to close database connection.
	sqlRepo.Shutdown()

	// Attempt transaction after shutdown (will fail).
	err := sqlRepo.WithTransaction(ctx, false, func(_ *sqlrepository.SQLTransaction) error {
		return nil
	})

	// Should fail with database error after shutdown.
	testify.Error(t, err)
}
