// Copyright (c) 2025 Justin Cranford
//
//

package telemetry

import (
	"context"
	"fmt"
	"testing"
	"time"

	stdoutLogExporter "log/slog"

	stdoutMetricExporter "go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	stdoutTraceExporter "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	logSdk "go.opentelemetry.io/otel/sdk/log"
	metricSdk "go.opentelemetry.io/otel/sdk/metric"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"

	"github.com/stretchr/testify/require"
)

// TestNewTelemetryService_InitMetricsError tests the error path when initMetrics fails.
func TestNewTelemetryService_InitMetricsError(t *testing.T) {
	original := initMetricsFn
	initMetricsFn = func(_ context.Context, _ *stdoutLogExporter.Logger, _ *TelemetrySettings) (*metricSdk.MeterProvider, error) {
		return nil, fmt.Errorf("injected initMetrics error")
	}

	defer func() { initMetricsFn = original }()

	settings := NewTestTelemetrySettings("test_metrics_error")

	_, err := NewTelemetryService(context.Background(), settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to init metrics")
}

// TestNewTelemetryService_InitTracesError tests the error path when initTraces fails.
func TestNewTelemetryService_InitTracesError(t *testing.T) {
	original := initTracesFn
	initTracesFn = func(_ context.Context, _ *stdoutLogExporter.Logger, _ *TelemetrySettings) (*traceSdk.TracerProvider, error) {
		return nil, fmt.Errorf("injected initTraces error")
	}

	defer func() { initTracesFn = original }()

	settings := NewTestTelemetrySettings("test_traces_error")

	_, err := NewTelemetryService(context.Background(), settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to init traces")
}

// TestShutdown_ForceFlushTracesError tests the Shutdown error path when traces ForceFlush fails.
func TestShutdown_ForceFlushTracesError(t *testing.T) {
	t.Parallel()

	// Create a TracerProvider with a batch processor backed by a failing exporter.
	exporter := &failFlushTraceExporter{}
	tp := traceSdk.NewTracerProvider(
		traceSdk.WithBatcher(exporter, traceSdk.WithBatchTimeout(time.Millisecond)),
	)

	// Create and end a span to buffer data in the batch processor.
	_, span := tp.Tracer("test").Start(context.Background(), "test-span")
	span.End()

	svc := &TelemetryService{
		StartTime:         time.Now().UTC(),
		Slogger:           testTelemetryService.Slogger,
		VerboseMode:       true,
		tracesProviderSdk: tp,
	}

	// ForceFlush will attempt to export the buffered span, triggering the exporter error.
	svc.Shutdown()
}

// TestShutdown_ForceFlushMetricsError tests the Shutdown error path when metrics shutdown fails.
func TestShutdown_ForceFlushMetricsError(t *testing.T) {
	t.Parallel()

	// Create a MeterProvider and shut it down early so subsequent Shutdown returns an error.
	mp := metricSdk.NewMeterProvider()
	_ = mp.Shutdown(context.Background())

	svc := &TelemetryService{
		StartTime:          time.Now().UTC(),
		Slogger:            testTelemetryService.Slogger,
		VerboseMode:        true,
		metricsProviderSdk: mp,
	}

	// Shutdown calls ForceFlush and Shutdown on the already-shut-down provider.
	svc.Shutdown()
}

// TestShutdown_LogsProviderShutdownError tests the Shutdown error path when logs shutdown fails.
func TestShutdown_LogsProviderShutdownError(t *testing.T) {
	t.Parallel()

	// Create a LoggerProvider and shut it down early so subsequent Shutdown returns an error.
	lp := logSdk.NewLoggerProvider()
	_ = lp.Shutdown(context.Background())

	svc := &TelemetryService{
		StartTime:       time.Now().UTC(),
		Slogger:         testTelemetryService.Slogger,
		VerboseMode:     true,
		LogsProvider:    lp,
		logsProviderSdk: lp,
	}

	// Shutdown calls ForceFlush and Shutdown on the already-shut-down provider.
	svc.Shutdown()
}

// TestInitMetrics_StdoutExporterError tests the error path when STDOUT metric exporter creation fails.
func TestInitMetrics_StdoutExporterError(t *testing.T) {
	originalStdout := stdoutMetricExporterNewFn
	stdoutMetricExporterNewFn = func(_ ...stdoutMetricExporter.Option) (metricSdk.Exporter, error) {
		return nil, fmt.Errorf("injected STDOUT metrics error")
	}

	defer func() { stdoutMetricExporterNewFn = originalStdout }()

	settings := NewTestTelemetrySettings("test_stdout_metrics_error")
	settings.OTLPConsole = true

	_, err := NewTelemetryService(context.Background(), settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "create STDOUT metrics failed")
}

// TestInitTraces_StdoutExporterError tests the error path when STDOUT trace exporter creation fails.
func TestInitTraces_StdoutExporterError(t *testing.T) {
	originalStdout := stdoutTraceExporterNewFn
	stdoutTraceExporterNewFn = func(_ ...stdoutTraceExporter.Option) (*stdoutTraceExporter.Exporter, error) {
		return nil, fmt.Errorf("injected STDOUT traces error")
	}

	defer func() { stdoutTraceExporterNewFn = originalStdout }()

	settings := NewTestTelemetrySettings("test_stdout_traces_error")
	settings.OTLPConsole = true

	_, err := NewTelemetryService(context.Background(), settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "create STDOUT traces failed")
}

// failFlushTraceExporter is a SpanExporter whose ExportSpans always returns an error.
type failFlushTraceExporter struct{}

func (e *failFlushTraceExporter) ExportSpans(_ context.Context, _ []traceSdk.ReadOnlySpan) error {
	return fmt.Errorf("injected export spans error")
}

func (e *failFlushTraceExporter) Shutdown(_ context.Context) error {
	return fmt.Errorf("injected exporter shutdown error")
}
