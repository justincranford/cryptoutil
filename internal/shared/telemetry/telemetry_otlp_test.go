// Copyright (c) 2025 Justin Cranford

//nolint:wrapcheck,thelper // Test code doesn't need to wrap errors or use t.Helper()
package telemetry

import (
	"context"
	"testing"


	"github.com/stretchr/testify/require"
)

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
	settings := NewTestTelemetrySettings("test_sidecar_disabled")
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
	settings := NewTestTelemetrySettings("test_sidecar_enabled")
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
	settings := NewTestTelemetrySettings("test_grpc")
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
	settings := NewTestTelemetrySettings("test_grpcs")
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
	settings := NewTestTelemetrySettings("test_https")
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
	settings := NewTestTelemetrySettings("test_console")
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
	settings := NewTestTelemetrySettings("test_invalid_endpoint")
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
	settings := NewTestTelemetrySettings("test_verbose_otlp")
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
	settings := NewTestTelemetrySettings("test_shutdown_verbose")
	settings.VerboseMode = true

	service, err := NewTelemetryService(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, service)

	// This should log verbose shutdown messages
	service.Shutdown()
}

// TestCheckSidecarHealth_HTTP tests checkSidecarHealth with HTTP protocol.
