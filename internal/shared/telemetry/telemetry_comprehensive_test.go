// Copyright (c) 2025 Justin Cranford

//nolint:wrapcheck,thelper // Test code doesn't need to wrap errors or use t.Helper()
package telemetry

import (
	"context"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"

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

	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_nil_ctx")

	_, err := NewTelemetryService(nil, settings) //nolint:staticcheck // Testing nil context error handling
	require.Error(t, err)
	require.Contains(t, err.Error(), "context must be non-nil")
}

// TestNewTelemetryService_EmptyServiceName tests that NewTelemetryService fails with empty service name.
func TestNewTelemetryService_EmptyServiceName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_empty_service")
	settings.OTLPService = "" // Empty service name

	_, err := NewTelemetryService(ctx, settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "service name must be non-empty")
}

// TestNewTelemetryService_Success tests successful telemetry service creation.
func TestNewTelemetryService_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_success")

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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_verbose")
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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_shutdown_once")

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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_shutdown_non_verbose")
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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_shutdown_verbose")
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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_shutdown_full")
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
	counter.Add(ctx, 1.0)

	service.Slogger.Info("test log message")

	// Allow time for async operations
	time.Sleep(10 * time.Millisecond)

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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_nil_providers")
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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_uptime")
	settings.VerboseMode = true

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Wait a bit to accumulate uptime
	time.Sleep(50 * time.Millisecond)

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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_providers")

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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_otlp_disabled")
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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_otlp_enabled")
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
		{"INFO", "INFO"},
		{"WARN", "WARN"},
		{"ERROR", "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_loglevel_" + tt.logLevel)
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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_settings_stored")
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
		{"INFO", "INFO"},
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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_concurrent")

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Launch concurrent goroutines using the service
	done := make(chan bool)

	for i := 0; i < 10; i++ {
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
	for i := 0; i < 10; i++ {
		<-done
	}

	defer service.Shutdown()
}

// TestTelemetryService_ShutdownIdempotent tests that multiple shutdowns are safe.
func TestTelemetryService_ShutdownIdempotent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_shutdown_idempotent")

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Multiple shutdowns should be safe
	for i := 0; i < 5; i++ {
		service.Shutdown()
	}
}

// TestParseProtocolAndEndpoint_AllProtocols tests parseProtocolAndEndpoint with all supported protocols.
func TestParseProtocolAndEndpoint_AllProtocols(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		endpoint  string
		wantHTTP  bool
		wantHTTPS bool
		wantGRPC  bool
		wantGRPCS bool
		wantAddr  string
		wantErr   bool
	}{
		{"HTTP", testHTTPEndpoint, true, false, false, false, "localhost:4318", false},
		{"HTTPS", testHTTPSEndpoint, false, true, false, false, "localhost:4318", false},
		{"gRPC", testGRPCEndpoint, false, false, true, false, "localhost:4317", false},
		{"gRPCS", testGRPCSEndpoint, false, false, false, true, "localhost:4317", false},
		{"Invalid", "ftp://localhost:4318", false, false, false, false, "", true},
		{"NoProtocol", "localhost:4318", false, false, false, false, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			endpoint := tt.endpoint
			isHTTP, isHTTPS, isGRPC, isGRPCS, addr, err := parseProtocolAndEndpoint(&endpoint)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "invalid OTLP endpoint protocol")
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantHTTP, isHTTP)
				require.Equal(t, tt.wantHTTPS, isHTTPS)
				require.Equal(t, tt.wantGRPC, isGRPC)
				require.Equal(t, tt.wantGRPCS, isGRPCS)
				require.NotNil(t, addr)
				require.Equal(t, tt.wantAddr, *addr)
			}
		})
	}
}

// TestParseLogLevel_AllLevels tests ParseLogLevel with all supported log levels.
func TestParseLogLevel_AllLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"ALL", "ALL", false},
		{"TRACE", "TRACE", false},
		{"DEBUG", "DEBUG", false},
		{"CONFIG", "CONFIG", false},
		{"INFO", "INFO", false},
		{"NOTICE", "NOTICE", false},
		{"WARN", "WARN", false},
		{"ERROR", "ERROR", false},
		{"FATAL", "FATAL", false},
		{"OFF", "OFF", false},
		{"lowercase_all", "all", false},
		{"lowercase_trace", "trace", false},
		{"lowercase_config", "config", false},
		{"lowercase_notice", "notice", false},
		{"lowercase_fatal", "fatal", false},
		{"lowercase_off", "off", false},
		{"mixed_case", "DeBuG", false},
		{"invalid", "INVALID_LEVEL", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			level, err := ParseLogLevel(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "invalid log level")
			} else {
				require.NoError(t, err)
				require.NotNil(t, level)
			}
		})
	}
}

// TestTelemetryService_CheckSidecarHealth_OTLPDisabled tests CheckSidecarHealth when OTLP is disabled.
func TestTelemetryService_CheckSidecarHealth_OTLPDisabled(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_sidecar_disabled")
	settings.OTLPEnabled = false

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()

	// With OTLP disabled, CheckSidecarHealth should return nil
	err = service.CheckSidecarHealth(ctx)
	require.NoError(t, err)
}

// TestTelemetryService_CheckSidecarHealth_OTLPEnabled tests CheckSidecarHealth when OTLP is enabled.
func TestTelemetryService_CheckSidecarHealth_OTLPEnabled(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_sidecar_enabled")
	settings.OTLPEnabled = true
	settings.OTLPEndpoint = testHTTPEndpoint // Non-existent endpoint

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()

	// With OTLP enabled but no sidecar, CheckSidecarHealth may succeed (connection is lazy)
	// or fail depending on implementation - we just test it doesn't panic
	_ = service.CheckSidecarHealth(ctx)
}

// TestTelemetryService_GRPCEndpoint tests service creation with gRPC endpoint.
func TestTelemetryService_GRPCEndpoint(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_grpc")
	settings.OTLPEnabled = true
	settings.OTLPEndpoint = testGRPCEndpoint

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()
}

// TestTelemetryService_GRPCSEndpoint tests service creation with secure gRPC endpoint.
func TestTelemetryService_GRPCSEndpoint(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_grpcs")
	settings.OTLPEnabled = true
	settings.OTLPEndpoint = testGRPCSEndpoint

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()
}

// TestTelemetryService_HTTPSEndpoint tests service creation with HTTPS endpoint.
func TestTelemetryService_HTTPSEndpoint(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_https")
	settings.OTLPEnabled = true
	settings.OTLPEndpoint = testHTTPSEndpoint

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()
}

// TestTelemetryService_OTLPConsole tests service creation with console output enabled.
func TestTelemetryService_OTLPConsole(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_console")
	settings.OTLPEnabled = false
	settings.OTLPConsole = true

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()
}

// TestTelemetryService_InvalidEndpoint tests service creation with invalid endpoint.
func TestTelemetryService_InvalidEndpoint(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_invalid_endpoint")
	settings.OTLPEnabled = true
	settings.OTLPEndpoint = testInvalidFTPEndpoint // Invalid protocol

	_, err := NewTelemetryService(ctx, settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid OTLP endpoint protocol")
}

// TestTelemetryService_VerboseModeWithOTLP tests verbose mode with OTLP enabled.
func TestTelemetryService_VerboseModeWithOTLP(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_verbose_otlp")
	settings.VerboseMode = true
	settings.OTLPEnabled = true
	settings.OTLPEndpoint = testHTTPEndpoint

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)
	require.True(t, service.VerboseMode)

	defer service.Shutdown()
}

// TestTelemetryService_ShutdownWithVerboseMode tests shutdown with verbose mode.
func TestTelemetryService_ShutdownWithVerboseMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_shutdown_verbose")
	settings.VerboseMode = true

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	// This should log verbose shutdown messages
	service.Shutdown()
}

// TestCheckSidecarHealth_HTTP tests checkSidecarHealth with HTTP protocol.
func TestCheckSidecarHealth_HTTP(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_sidecar_http")
	settings.OTLPEndpoint = testHTTPEndpoint

	// This should not panic, may return error since no sidecar is running
	_ = checkSidecarHealth(ctx, settings)
}

// TestCheckSidecarHealth_HTTPS tests checkSidecarHealth with HTTPS protocol.
func TestCheckSidecarHealth_HTTPS(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_sidecar_https")
	settings.OTLPEndpoint = testHTTPSEndpoint

	// This should not panic, may return error since no sidecar is running
	_ = checkSidecarHealth(ctx, settings)
}

// TestCheckSidecarHealth_GRPC tests checkSidecarHealth with gRPC protocol.
func TestCheckSidecarHealth_GRPC(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_sidecar_grpc")
	settings.OTLPEndpoint = testGRPCEndpoint

	// This should not panic, may return error since no sidecar is running
	_ = checkSidecarHealth(ctx, settings)
}

// TestCheckSidecarHealth_GRPCS tests checkSidecarHealth with secure gRPC protocol.
func TestCheckSidecarHealth_GRPCS(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_sidecar_grpcs")
	settings.OTLPEndpoint = testGRPCSEndpoint

	// This should not panic, may return error since no sidecar is running
	_ = checkSidecarHealth(ctx, settings)
}

// TestCheckSidecarHealth_InvalidProtocol tests checkSidecarHealth with invalid protocol.
func TestCheckSidecarHealth_InvalidProtocol(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_sidecar_invalid")
	settings.OTLPEndpoint = "ftp://localhost:4317"

	err := checkSidecarHealth(ctx, settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse OTLP endpoint")
}

// TestCheckSidecarHealthWithRetry_ContextCancellation tests retry with context cancellation.
func TestCheckSidecarHealthWithRetry_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_sidecar_cancel")
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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_sidecar_all_fail")
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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_example_spans")
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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_uptime")

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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_otlp_grpc")
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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_otlp_grpcs")
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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_otlp_https")
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
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("test_console_otlp")
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
