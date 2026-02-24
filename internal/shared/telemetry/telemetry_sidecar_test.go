// Copyright (c) 2025 Justin Cranford

package telemetry

import (
	"context"
	"testing"
	"time"


	"github.com/stretchr/testify/require"
)

func TestCheckSidecarHealth_HTTP(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_sidecar_http")
	settings.OTLPEndpoint = testHTTPEndpoint

	// This should not panic, may return error since no sidecar is running
	_ = checkSidecarHealth(ctx, settings)
}

// TestCheckSidecarHealth_HTTPS tests checkSidecarHealth with HTTPS protocol.
func TestCheckSidecarHealth_HTTPS(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_sidecar_https")
	settings.OTLPEndpoint = testHTTPSEndpoint

	// This should not panic, may return error since no sidecar is running
	_ = checkSidecarHealth(ctx, settings)
}

// TestCheckSidecarHealth_GRPC tests checkSidecarHealth with gRPC protocol.
func TestCheckSidecarHealth_GRPC(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_sidecar_grpc")
	settings.OTLPEndpoint = testGRPCEndpoint

	// This should not panic, may return error since no sidecar is running
	_ = checkSidecarHealth(ctx, settings)
}

// TestCheckSidecarHealth_GRPCS tests checkSidecarHealth with secure gRPC protocol.
func TestCheckSidecarHealth_GRPCS(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_sidecar_grpcs")
	settings.OTLPEndpoint = testGRPCSEndpoint

	// This should not panic, may return error since no sidecar is running
	_ = checkSidecarHealth(ctx, settings)
}

// TestCheckSidecarHealth_InvalidProtocol tests checkSidecarHealth with invalid protocol.
func TestCheckSidecarHealth_InvalidProtocol(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_sidecar_invalid")
	settings.OTLPEndpoint = "ftp://localhost:4317"

	err := checkSidecarHealth(ctx, settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse OTLP endpoint")
}

// TestCheckSidecarHealthWithRetry_ContextCancellation tests retry with context cancellation.
func TestCheckSidecarHealthWithRetry_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	settings := NewTestTelemetrySettings("test_sidecar_cancel")
	settings.OTLPEndpoint = testHTTPEndpoint

	// Cancel context immediately
	cancel()

	// With cancelled context, should fail quickly
	_, err := checkSidecarHealthWithRetry(ctx, settings)
	// May succeed (lazy connection) or fail with context cancelled
	if err != nil {
		// Should contain context cancelled error if it failed
		require.Contains(t, err.Error(), "context")
	}
}

// TestCheckSidecarHealthWithRetry_AllRetriesFail tests retry when all attempts fail.
func TestCheckSidecarHealthWithRetry_AllRetriesFail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_sidecar_all_fail")
	settings.OTLPEndpoint = testInvalidFTPEndpoint // Invalid protocol will fail on all retries

	intermediateErrs, err := checkSidecarHealthWithRetry(ctx, settings)
	require.Error(t, err)
	require.NotEmpty(t, intermediateErrs)
	require.Contains(t, err.Error(), "sidecar health check failed after")
}

// TestTelemetryService_ExampleTracesSpans tests the example traces spans helper.
func TestTelemetryService_ExampleTracesSpans(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_example_spans")
	settings.VerboseMode = true

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()

	// Service was created with verbose mode, which triggers doExampleTracesSpans
}

// TestLevelConstants tests that level constants have expected values.
func TestLevelConstants(t *testing.T) {
	t.Parallel()

	// Test relationships between levels
	require.Less(t, int(LevelAll), int(LevelTrace))
	require.Less(t, int(LevelTrace), int(LevelDebug))
	require.Less(t, int(LevelDebug), int(LevelConfig))
	require.Less(t, int(LevelConfig), int(LevelInfo))
	require.Less(t, int(LevelInfo), int(LevelNotice))
	require.Less(t, int(LevelNotice), int(LevelWarn))
	require.Less(t, int(LevelWarn), int(LevelError))
	require.Less(t, int(LevelError), int(LevelFatal))
	require.Less(t, int(LevelFatal), int(LevelMax))
}

// TestTelemetryService_UptimeCalculation tests uptime calculation.
func TestTelemetryService_UptimeCalculation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_uptime")

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Small delay
	time.Sleep(10 * time.Millisecond)

	// Calculate uptime
	uptime := time.Since(service.StartTime)
	require.Greater(t, uptime, 10*time.Millisecond)

	defer service.Shutdown()
}

// TestTelemetryService_OTLPEnabledWithGRPC tests service with OTLP enabled using gRPC.
func TestTelemetryService_OTLPEnabledWithGRPC(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_otlp_grpc")
	settings.OTLPEnabled = true
	settings.OTLPEndpoint = testGRPCEndpoint

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()

	// Use the service
	service.Slogger.Info("test with gRPC endpoint")
}

// TestTelemetryService_OTLPEnabledWithGRPCS tests service with OTLP enabled using secure gRPC.
func TestTelemetryService_OTLPEnabledWithGRPCS(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_otlp_grpcs")
	settings.OTLPEnabled = true
	settings.OTLPEndpoint = testGRPCSEndpoint

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()

	// Use the service
	service.Slogger.Info("test with gRPCS endpoint")
}

// TestTelemetryService_OTLPEnabledWithHTTPS tests service with OTLP enabled using HTTPS.
func TestTelemetryService_OTLPEnabledWithHTTPS(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_otlp_https")
	settings.OTLPEnabled = true
	settings.OTLPEndpoint = testHTTPSEndpoint

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()

	// Use the service
	service.Slogger.Info("test with HTTPS endpoint")
}

// TestTelemetryService_OTLPConsoleWithOTLP tests service with both console and OTLP enabled.
func TestTelemetryService_OTLPConsoleWithOTLP(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_console_otlp")
	settings.OTLPEnabled = true
	settings.OTLPEndpoint = testHTTPEndpoint
	settings.OTLPConsole = true

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()

	// Use the service with both outputs
	service.Slogger.Info("test with console and OTLP")
}

// TestCheckSidecarHealth_ErrorPropagation tests CheckSidecarHealth propagates errors.
func TestCheckSidecarHealth_ErrorPropagation(t *testing.T) {
	t.Parallel()

	// Create service with HTTP endpoint and OTLPEnabled
	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_sidecar_err_propagation")
	settings.OTLPEnabled = true
	settings.OTLPEndpoint = testHTTPEndpoint

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()

	// Use a cancelled context to force checkSidecarHealth to fail via HTTP Start(ctx) check
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err = service.CheckSidecarHealth(cancelledCtx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "sidecar health check failed")
}
