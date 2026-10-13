// Copyright (c) 2025-2026 Justin Cranford.
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

func TestNewTelemetryService_Initializers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*telemetryServiceInitializers, *TelemetrySettings)
		wantErr   string
	}{
		{
			name:      "default initializers succeed",
			configure: func(_ *telemetryServiceInitializers, _ *TelemetrySettings) {},
		},
		{
			name: "metrics initializer error",
			configure: func(initializers *telemetryServiceInitializers, _ *TelemetrySettings) {
				initializers.initMetrics = func(context.Context, *stdoutLogExporter.Logger, *TelemetrySettings) (*metricSdk.MeterProvider, error) {
					return nil, fmt.Errorf("injected metrics initializer error")
				}
			},
			wantErr: "failed to init metrics",
		},
		{
			name: "traces initializer error",
			configure: func(initializers *telemetryServiceInitializers, _ *TelemetrySettings) {
				initializers.initTraces = func(context.Context, *stdoutLogExporter.Logger, *TelemetrySettings) (*traceSdk.TracerProvider, error) {
					return nil, fmt.Errorf("injected traces initializer error")
				}
			},
			wantErr: "failed to init traces",
		},
		{
			name: "stdout metrics exporter error",
			configure: func(initializers *telemetryServiceInitializers, settings *TelemetrySettings) {
				settings.OTLPConsole = true
				initializers.initMetrics = func(ctx context.Context, slogger *stdoutLogExporter.Logger, settings *TelemetrySettings) (*metricSdk.MeterProvider, error) {
					return initMetrics(ctx, slogger, settings, func(...stdoutMetricExporter.Option) (metricSdk.Exporter, error) {
						return nil, fmt.Errorf("injected stdout metrics exporter error")
					})
				}
			},
			wantErr: "create STDOUT metrics failed",
		},
		{
			name: "stdout traces exporter error",
			configure: func(initializers *telemetryServiceInitializers, settings *TelemetrySettings) {
				settings.OTLPConsole = true
				initializers.initTraces = func(ctx context.Context, slogger *stdoutLogExporter.Logger, settings *TelemetrySettings) (*traceSdk.TracerProvider, error) {
					return initTraces(ctx, slogger, settings, func(...stdoutTraceExporter.Option) (*stdoutTraceExporter.Exporter, error) {
						return nil, fmt.Errorf("injected stdout traces exporter error")
					})
				}
			},
			wantErr: "create STDOUT traces failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := NewTestTelemetrySettings("test_initializers")
			initializers := defaultTelemetryServiceInitializers()
			tc.configure(&initializers, settings)

			service, err := newTelemetryServiceWithInitializers(context.Background(), settings, initializers)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, service)
			service.Shutdown()
		})
	}
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

// failFlushTraceExporter is a SpanExporter whose ExportSpans always returns an error.
type failFlushTraceExporter struct{}

func (e *failFlushTraceExporter) ExportSpans(_ context.Context, _ []traceSdk.ReadOnlySpan) error {
	return fmt.Errorf("injected export spans error")
}

func (e *failFlushTraceExporter) Shutdown(_ context.Context) error {
	return fmt.Errorf("injected exporter shutdown error")
}
