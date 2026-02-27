// Copyright (c) 2025 Justin Cranford

package telemetry

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Test constants for telemetry endpoint tests.
const (
	testHTTPEndpoint       = "http://localhost:4318"
	testHTTPSEndpoint      = "https://localhost:4318"
	testGRPCEndpoint       = "grpc://localhost:4317"
	testGRPCSEndpoint      = "grpcs://localhost:4317"
	testInvalidFTPEndpoint = "ftp://invalid:1234"
)

// TestNewTelemetryService_NilContext tests that NewTelemetryService fails with nil context.
func TestNewTelemetryService_NilContext(t *testing.T) {
	t.Parallel()

	settings := NewTestTelemetrySettings("test_nil_ctx")

	_, err := NewTelemetryService(nil, settings) //nolint:staticcheck // Testing nil context error handling
	require.Error(t, err)
	require.Contains(t, err.Error(), "context must be non-nil")
}

// TestNewTelemetryService_EmptyServiceName tests that NewTelemetryService fails with empty service name.
func TestNewTelemetryService_EmptyServiceName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_empty_service")
	settings.OTLPService = "" // Empty service name

	_, err := NewTelemetryService(ctx, settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "service name must be non-empty")
}

// TestNewTelemetryService_Success tests successful telemetry service creation.
func TestNewTelemetryService_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_success")

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)
	require.NotNil(t, service.Slogger)
	require.NotNil(t, service.LogsProvider)
	require.NotNil(t, service.MetricsProvider)
	require.NotNil(t, service.TracesProvider)
	require.NotNil(t, service.TextMapPropagator)
	require.False(t, service.StartTime.IsZero())

	// Cleanup
	defer service.Shutdown()
}

// TestNewTelemetryService_VerboseMode tests telemetry service creation with verbose mode.
func TestNewTelemetryService_VerboseMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_verbose")
	settings.VerboseMode = true

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)
	require.True(t, service.VerboseMode)

	// Cleanup
	defer service.Shutdown()
}

// TestShutdown_CallsOnce tests that Shutdown can be called multiple times safely.
func TestShutdown_CallsOnce(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_shutdown_once")

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Call Shutdown multiple times - should be safe due to sync.Once
	service.Shutdown()
	service.Shutdown()
	service.Shutdown()
}

// TestShutdown_NonVerboseMode tests Shutdown in non-verbose mode.
func TestShutdown_NonVerboseMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_shutdown_non_verbose")
	settings.VerboseMode = false

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Shutdown should work without verbose logging
	service.Shutdown()
}

// TestShutdown_VerboseMode tests Shutdown in verbose mode.
func TestShutdown_VerboseMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_shutdown_verbose")
	settings.VerboseMode = true

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Shutdown should work with verbose logging
	service.Shutdown()
}

// TestShutdown_WithTracesMetricsLogs tests Shutdown with all providers initialized.
func TestShutdown_WithTracesMetricsLogs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_shutdown_full")
	settings.VerboseMode = true

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Use the providers before shutdown
	tracer := service.TracesProvider.Tracer("test-tracer")
	_, span := tracer.Start(ctx, "test-span")
	span.End()

	meter := service.MetricsProvider.Meter("test-meter")
	counter, err := meter.Float64Counter("test-counter")
	require.NoError(t, err)
	counter.Add(ctx, cryptoutilSharedMagic.TestProbAlways)

	service.Slogger.Info("test log message")

	// Allow time for async operations
	time.Sleep(cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Millisecond)

	// Shutdown should flush and shutdown all providers
	service.Shutdown()
}

// TestShutdown_NilProviders tests Shutdown with nil providers.
func TestShutdown_NilProviders(t *testing.T) {
	t.Parallel()

	// Create a service with nil SDK providers (simulate partial initialization)
	service := &TelemetryService{
		StartTime:          time.Now().UTC(),
		VerboseMode:        false,
		logsProviderSdk:    nil,
		metricsProviderSdk: nil,
		tracesProviderSdk:  nil,
	}

	// Create a minimal logger for Shutdown to use
	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_nil_providers")
	slogger, _, err := initLogger(ctx, settings)
	require.NoError(t, err)

	service.Slogger = slogger

	// Shutdown should handle nil providers gracefully
	service.Shutdown()
}

// TestShutdown_AfterUptime tests Shutdown after some uptime.
func TestShutdown_AfterUptime(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_uptime")
	settings.VerboseMode = true

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Wait a bit to accumulate uptime
	time.Sleep(cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond)

	// Verify uptime is non-zero
	uptime := time.Since(service.StartTime)
	require.Greater(t, uptime, time.Duration(0))

	// Shutdown should log uptime
	service.Shutdown()
}

// TestTelemetryService_ProvidersNotNil tests that all providers are initialized.
func TestTelemetryService_ProvidersNotNil(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_providers")

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Verify all providers are non-nil
	require.NotNil(t, service.LogsProvider, "LogsProvider should not be nil")
	require.NotNil(t, service.MetricsProvider, "MetricsProvider should not be nil")
	require.NotNil(t, service.TracesProvider, "TracesProvider should not be nil")
	require.NotNil(t, service.TextMapPropagator, "TextMapPropagator should not be nil")

	defer service.Shutdown()
}

// TestTelemetryService_Timeouts tests timeout constants.
func TestTelemetryService_Timeouts(t *testing.T) {
	t.Parallel()

	// Verify timeout constants are reasonable
	require.Greater(t, LogsTimeout, time.Duration(0))
	require.Greater(t, MetricsTimeout, time.Duration(0))
	require.Greater(t, TracesTimeout, time.Duration(0))
	require.Greater(t, ForceFlushTimeout, time.Duration(0))

	// Verify batch size constants are reasonable
	require.Greater(t, MaxLogsBatchSize, 0)
	require.Greater(t, MaxMetricsBatchSize, 0)
	require.Greater(t, MaxTracesBatchSize, 0)
}

// TestTelemetryService_OTLPDisabled tests service creation with OTLP disabled.
func TestTelemetryService_OTLPDisabled(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_otlp_disabled")
	settings.OTLPEnabled = false

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Service should still work with OTLP disabled
	service.Slogger.Info("test message with OTLP disabled")

	defer service.Shutdown()
}

// TestTelemetryService_OTLPEnabled tests service creation with OTLP enabled.
func TestTelemetryService_OTLPEnabled(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_otlp_enabled")
	settings.OTLPEnabled = true
	settings.OTLPEndpoint = testHTTPEndpoint // HTTP endpoint

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Service should work with OTLP enabled (even if sidecar not available)
	service.Slogger.Info("test message with OTLP enabled")

	defer service.Shutdown()
}

// TestTelemetryService_DifferentLogLevels tests service creation with different log levels.
func TestTelemetryService_DifferentLogLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		logLevel string
	}{
		{"DEBUG", "DEBUG"},
		{cryptoutilSharedMagic.DefaultLogLevelInfo, cryptoutilSharedMagic.DefaultLogLevelInfo},
		{"WARN", "WARN"},
		{"ERROR", "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			settings := NewTestTelemetrySettings("test_loglevel_" + tt.logLevel)
			settings.LogLevel = tt.logLevel

			service, err := NewTelemetryService(ctx, settings)
			require.NoError(t, err)
			require.NotNil(t, service)

			// Test logging at the configured level
			service.Slogger.Info("test message at " + tt.logLevel)

			defer service.Shutdown()
		})
	}
}

// TestTelemetryService_SettingsStored tests that settings are stored in the service.
func TestTelemetryService_SettingsStored(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_settings_stored")
	settings.OTLPService = "test-service"

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)
	require.NotNil(t, service.settings)
	require.Equal(t, "test-service", service.settings.OTLPService)

	defer service.Shutdown()
}

// TestParseLogLevel_ValidLevels tests ParseLogLevel with valid log levels.
func TestParseLogLevel_ValidLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"DEBUG", "DEBUG"},
		{cryptoutilSharedMagic.DefaultLogLevelInfo, cryptoutilSharedMagic.DefaultLogLevelInfo},
		{"WARN", "WARN"},
		{"ERROR", "ERROR"},
		{"lowercase", "info"},
		{"mixed case", "WaRn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			level, err := ParseLogLevel(tt.input)
			require.NoError(t, err)
			require.NotNil(t, level)
		})
	}
}

// TestParseLogLevel_InvalidLevel tests ParseLogLevel with invalid log level.
func TestParseLogLevel_InvalidLevel(t *testing.T) {
	t.Parallel()

	_, err := ParseLogLevel("INVALID")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid log level")
}

// TestTelemetryService_Concurrent tests concurrent access to telemetry service.
func TestTelemetryService_Concurrent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_concurrent")

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Launch concurrent goroutines using the service
	done := make(chan bool)

	for i := 0; i < cryptoutilSharedMagic.JoseJADefaultMaxMaterials; i++ {
		go func(id int) {
			// Log
			service.Slogger.Info("concurrent log", "goroutine", id)

			// Trace
			tracer := service.TracesProvider.Tracer("concurrent-tracer")
			_, span := tracer.Start(ctx, "concurrent-span")
			span.End()

			// Metric
			meter := service.MetricsProvider.Meter("concurrent-meter")

			counter, err := meter.Int64Counter("concurrent-counter")
			if err == nil {
				counter.Add(ctx, int64(id))
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < cryptoutilSharedMagic.JoseJADefaultMaxMaterials; i++ {
		<-done
	}

	defer service.Shutdown()
}

// TestTelemetryService_ShutdownIdempotent tests that multiple shutdowns are safe.
func TestTelemetryService_ShutdownIdempotent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_shutdown_idempotent")

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Multiple shutdowns should be safe
	for i := 0; i < cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries; i++ {
		service.Shutdown()
	}
}

// TestParseProtocolAndEndpoint_AllProtocols tests parseProtocolAndEndpoint with all supported protocols.

// TestNewTelemetryService_InvalidLogLevel tests that NewTelemetryService fails with invalid log level.
func TestNewTelemetryService_InvalidLogLevel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := NewTestTelemetrySettings("test_invalid_log_level")
	settings.LogLevel = "INVALID_LEVEL"

	_, err := NewTelemetryService(ctx, settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to init logger")
}
