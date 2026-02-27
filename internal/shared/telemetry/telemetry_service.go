// Copyright (c) 2025 Justin Cranford
//
//

package telemetry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	slogMulti "github.com/samber/slog-multi"
	otelSlogBridge "go.opentelemetry.io/contrib/bridges/otelslog"

	stdoutLogExporter "log/slog"

	grpcLogExporter "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	httpLogExporter "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	grpcMetricExporter "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	httpMetricExporter "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	grpcTraceExporterotlptracegrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	httpTraceExporterotlptracehttp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	stdoutMetricExporter "go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	stdoutTraceExporter "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"

	logApi "go.opentelemetry.io/otel/log"
	metricApi "go.opentelemetry.io/otel/metric"
	propagationApi "go.opentelemetry.io/otel/propagation"
	traceApi "go.opentelemetry.io/otel/trace"

	logSdk "go.opentelemetry.io/otel/sdk/log"
	metricSdk "go.opentelemetry.io/otel/sdk/metric"
	resourceSdk "go.opentelemetry.io/otel/sdk/resource"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
)

var (
	initMetricsFn             = initMetrics              // injectable for testing error paths.
	initTracesFn              = initTraces               // injectable for testing error paths.
	stdoutMetricExporterNewFn = stdoutMetricExporter.New // injectable for testing error paths.
	stdoutTraceExporterNewFn  = stdoutTraceExporter.New  // injectable for testing error paths.
)

// TelemetryService is a composite of OpenTelemetry providers for Logs, Metrics, and Traces.
type TelemetryService struct {
	StartTime          time.Time
	shutdownOnce       sync.Once
	Slogger            *stdoutLogExporter.Logger
	VerboseMode        bool
	LogsProvider       logApi.LoggerProvider
	MetricsProvider    metricApi.MeterProvider
	TracesProvider     traceApi.TracerProvider
	TextMapPropagator  *propagationApi.TextMapPropagator
	logsProviderSdk    *logSdk.LoggerProvider   // Not exported, but still needed to do shutdown
	metricsProviderSdk *metricSdk.MeterProvider // Not exported, but still needed to do shutdown
	tracesProviderSdk  *traceSdk.TracerProvider // Not exported, but still needed to do shutdown
	settings           *TelemetrySettings       // Store settings for health checks
}

// Timeout constants for telemetry operations.
const (
	LogsTimeout       = cryptoutilSharedMagic.DefaultLogsTimeout
	MetricsTimeout    = cryptoutilSharedMagic.DefaultMetricsTimeout
	TracesTimeout     = cryptoutilSharedMagic.DefaultTracesTimeout
	ForceFlushTimeout = cryptoutilSharedMagic.DefaultForceFlushTimeout

	// MaxLogsBatchSize is the conservative batch size for log processing.
	MaxLogsBatchSize    = cryptoutilSharedMagic.DefaultLogsBatchSize    // Conservative for logs
	MaxMetricsBatchSize = cryptoutilSharedMagic.DefaultMetricsBatchSize // Moderate for metrics
	MaxTracesBatchSize  = cryptoutilSharedMagic.DefaultTracesBatchSize  // Conservative for traces to prevent memory issues
)

// NewTelemetryService creates and initializes a TelemetryService with OTLP exporters.
func NewTelemetryService(ctx context.Context, settings *TelemetrySettings) (*TelemetryService, error) {
	startTime := time.Now().UTC()

	if ctx == nil {
		return nil, fmt.Errorf("context must be non-nil")
	} else if len(settings.OTLPService) == 0 {
		return nil, fmt.Errorf("service name must be non-empty")
	}

	// Check sidecar connectivity during startup if OTLP is enabled
	var retryErrors []error

	var overallErr error

	if settings.OTLPEnabled {
		retryErrors, overallErr = checkSidecarHealthWithRetry(ctx, settings)
	}

	slogger, logsProvider, err := initLogger(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}

	// logsProvider is initialized, now we can log the sidecar health checks done just before logsProvider initialization
	if overallErr == nil {
		slogger.Info("sidecar health check succeeded", "attempts", len(retryErrors), "errors", errors.Join(retryErrors...))
	}

	metricsProvider, err := initMetricsFn(ctx, slogger, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to init metrics: %w", err)
	}

	tracesProvider, err := initTracesFn(ctx, slogger, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to init traces: %w", err)
	}

	textMapPropagator := initTextMapPropagator()

	if settings.VerboseMode {
		doExampleTracesSpans(ctx, tracesProvider, slogger)
	}

	return &TelemetryService{
		StartTime:          startTime,
		Slogger:            slogger,
		VerboseMode:        settings.VerboseMode,
		LogsProvider:       logsProvider,
		MetricsProvider:    metricsProvider,
		TracesProvider:     tracesProvider,
		TextMapPropagator:  textMapPropagator,
		logsProviderSdk:    logsProvider,
		metricsProviderSdk: metricsProvider,
		tracesProviderSdk:  tracesProvider,
		settings:           settings,
	}, nil
}

// Shutdown gracefully shuts down all telemetry providers with force flush and timeout handling.
func (s *TelemetryService) Shutdown() {
	s.shutdownOnce.Do(func() {
		if s.VerboseMode {
			s.Slogger.Debug("shutting down telemetry providers", "uptime", time.Since(s.StartTime).Seconds())
		}

		if s.tracesProviderSdk != nil || s.metricsProviderSdk != nil || s.logsProviderSdk != nil {
			startTimeForceFlush := time.Now().UTC()

			forceFlushCtx, forceFlushCancelDueToTimeout := context.WithTimeout(context.Background(), ForceFlushTimeout)
			defer forceFlushCancelDueToTimeout()

			if s.tracesProviderSdk != nil {
				if s.VerboseMode {
					s.Slogger.Debug("traces provider force flushing", "uptime", time.Since(s.StartTime).Seconds())
				}

				startTimeForceFlushTracesProvider := time.Now().UTC()

				if err := s.tracesProviderSdk.ForceFlush(forceFlushCtx); err != nil {
					s.Slogger.Error("traces provider force flush failed", cryptoutilSharedMagic.StringError, fmt.Errorf("traces provider force flush error: %w", err))
				} else if s.VerboseMode {
					s.Slogger.Debug("traces provider force flushed", "uptime", time.Since(s.StartTime).Seconds(), "flush", time.Since(startTimeForceFlushTracesProvider).Seconds())
				}
			}

			if s.metricsProviderSdk != nil {
				if s.VerboseMode {
					s.Slogger.Debug("metrics provider force flushing", "uptime", time.Since(s.StartTime).Seconds())
				}

				startTimeForceFlushMetricsProvider := time.Now().UTC()

				if err := s.metricsProviderSdk.ForceFlush(forceFlushCtx); err != nil {
					s.Slogger.Error("metrics provider force flush failed", cryptoutilSharedMagic.StringError, fmt.Errorf("metrics provider force flush error: %w", err))
				} else if s.VerboseMode {
					s.Slogger.Debug("metrics provider force flushed", "uptime", time.Since(s.StartTime).Seconds(), "flush", time.Since(startTimeForceFlushMetricsProvider).Seconds())
				}
			}

			if s.logsProviderSdk != nil {
				if s.VerboseMode {
					s.Slogger.Debug("logs provider force flushing", "uptime", time.Since(s.StartTime).Seconds())
				}

				startTimeForceFlushLogsProvider := time.Now().UTC()

				if err := s.logsProviderSdk.ForceFlush(forceFlushCtx); err != nil {
					s.Slogger.Error("logs provider force flush failed", cryptoutilSharedMagic.StringError, fmt.Errorf("logs provider force flush error: %w", err))
				} else if s.VerboseMode {
					s.Slogger.Debug("logs provider force flushed", "uptime", time.Since(s.StartTime).Seconds(), "flush", time.Since(startTimeForceFlushLogsProvider).Seconds())
				}
			}

			s.Slogger.Debug("telemetry providers force flushed", "uptime", time.Since(s.StartTime).Seconds(), "flush", time.Since(startTimeForceFlush).Seconds())

			startTimeShutdown := time.Now().UTC()
			shutdownCtx := context.Background()

			if s.tracesProviderSdk != nil {
				if s.VerboseMode {
					s.Slogger.Debug("traces provider shutting down", "uptime", time.Since(s.StartTime).Seconds())
				}

				startTimeShutdownTracesProvider := time.Now().UTC()

				if err := s.tracesProviderSdk.Shutdown(shutdownCtx); err != nil {
					s.Slogger.Error("traces provider shutdown failed", cryptoutilSharedMagic.StringError, fmt.Errorf("traces provider shutdown error: %w", err))
				} else if s.VerboseMode {
					s.Slogger.Debug("traces provider shut down", "uptime", time.Since(s.StartTime).Seconds(), "flush", time.Since(startTimeShutdownTracesProvider).Seconds())
				}
			}

			if s.metricsProviderSdk != nil {
				if s.VerboseMode {
					s.Slogger.Debug("metrics provider shutting down", "uptime", time.Since(s.StartTime).Seconds())
				}

				startTimeShutdownMetricsProvider := time.Now().UTC()

				if err := s.metricsProviderSdk.Shutdown(shutdownCtx); err != nil {
					s.Slogger.Error("metrics provider shutdown failed", cryptoutilSharedMagic.StringError, fmt.Errorf("metrics provider shutdown error: %w", err))
				} else if s.VerboseMode {
					s.Slogger.Debug("metrics provider shut down", "uptime", time.Since(s.StartTime).Seconds(), "flush", time.Since(startTimeShutdownMetricsProvider).Seconds())
				}
			}

			if s.logsProviderSdk != nil {
				if s.VerboseMode {
					s.Slogger.Debug("logs provider shutting down", "uptime", time.Since(s.StartTime).Seconds())
				}

				startTimeShutdownLogsProvider := time.Now().UTC()

				if err := s.logsProviderSdk.Shutdown(shutdownCtx); err != nil {
					s.Slogger.Error("logs provider shutdown failed", cryptoutilSharedMagic.StringError, fmt.Errorf("logs provider shutdown error: %w", err))
				} else if s.VerboseMode {
					s.Slogger.Debug("logs provider shut down", "uptime", time.Since(s.StartTime).Seconds(), "flush", time.Since(startTimeShutdownLogsProvider).Seconds())
				}
			}

			s.Slogger.Debug("telemetry providers shut down", "uptime", time.Since(s.StartTime).Seconds(), "flush", time.Since(startTimeShutdown).Seconds())
		}
	})
}

func initLogger(ctx context.Context, settings *TelemetrySettings) (*stdoutLogExporter.Logger, *logSdk.LoggerProvider, error) {
	slogLevel, err := ParseLogLevel(settings.LogLevel)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse log level: %w", err)
	}

	handlerOptions := &stdoutLogExporter.HandlerOptions{
		Level: slogLevel,
	}
	stdoutSlogHandler := stdoutLogExporter.NewTextHandler(os.Stdout, handlerOptions).WithAttrs(getSlogStdoutAttributes(settings))

	slogger := stdoutLogExporter.New(stdoutSlogHandler)
	if settings.VerboseMode {
		slogger.Debug("initializing otel logs provider")
	}

	otelLogsResource := resourceSdk.NewWithAttributes("", getOtelLogsAttributes(settings)...)

	isHTTP, isHTTPS, isGRPC, isGRPCS, endpoint, err := parseProtocolAndEndpoint(&settings.OTLPEndpoint)
	if err != nil {
		slogger.Error("parse protocol and endpoint failed", cryptoutilSharedMagic.StringError, err)

		return nil, nil, fmt.Errorf("parse protocol and endpoint failed: %w", err)
	}

	var otelExporter logSdk.Exporter
	if isHTTP {
		otelExporter, _ = httpLogExporter.New(ctx, httpLogExporter.WithEndpoint(*endpoint), httpLogExporter.WithInsecure())
	} else if isHTTPS {
		otelExporter, _ = httpLogExporter.New(ctx, httpLogExporter.WithEndpoint(*endpoint))
	} else if isGRPC {
		otelExporter, _ = grpcLogExporter.New(ctx, grpcLogExporter.WithEndpoint(*endpoint), grpcLogExporter.WithInsecure())
	} else if isGRPCS {
		otelExporter, _ = grpcLogExporter.New(ctx, grpcLogExporter.WithEndpoint(*endpoint))
	}

	otelProviderOptions := []logSdk.LoggerProviderOption{
		logSdk.WithResource(otelLogsResource),
		logSdk.WithProcessor(logSdk.NewBatchProcessor(otelExporter,
			logSdk.WithExportTimeout(LogsTimeout),
			logSdk.WithExportMaxBatchSize(MaxLogsBatchSize))),
	}
	otelProvider := logSdk.NewLoggerProvider(otelProviderOptions...)

	if settings.OTLPEnabled {
		otelSlogHandler := otelSlogBridge.NewHandler(settings.OTLPService, otelSlogBridge.WithLoggerProvider(otelProvider))
		slogger = stdoutLogExporter.New(slogMulti.Fanout(stdoutSlogHandler, otelSlogHandler))
	}

	slogger.Debug("initialized otel logs provider")

	return slogger, otelProvider, nil
}

func initMetrics(ctx context.Context, slogger *stdoutLogExporter.Logger, settings *TelemetrySettings) (*metricSdk.MeterProvider, error) {
	if settings.VerboseMode {
		slogger.Debug("initializing metrics provider")
	}

	var metricsOptions []metricSdk.Option

	otelMeterTracerTags, _ := resourceSdk.New(ctx, resourceSdk.WithAttributes(getOtelMetricsTracesAttributes(settings)...))

	metricsOptions = append(metricsOptions, metricSdk.WithResource(otelMeterTracerTags))

	if settings.OTLPEnabled {
		isHTTP, isHTTPS, isGRPC, isGRPCS, endpoint, _ := parseProtocolAndEndpoint(&settings.OTLPEndpoint)

		var httpMetricsExporter *httpMetricExporter.Exporter

		var grpcMetricsExporter *grpcMetricExporter.Exporter

		var err error

		if isHTTP {
			httpMetricsExporter, err = httpMetricExporter.New(ctx, httpMetricExporter.WithEndpoint(*endpoint), httpMetricExporter.WithInsecure())
		} else if isHTTPS {
			httpMetricsExporter, err = httpMetricExporter.New(ctx, httpMetricExporter.WithEndpoint(*endpoint))
		} else if isGRPC {
			grpcMetricsExporter, err = grpcMetricExporter.New(ctx, grpcMetricExporter.WithEndpoint(*endpoint), grpcMetricExporter.WithInsecure())
		} else if isGRPCS {
			grpcMetricsExporter, err = grpcMetricExporter.New(ctx, grpcMetricExporter.WithEndpoint(*endpoint))
		}

		if err != nil {
			slogger.Error("create Otel metrics exporter failed", cryptoutilSharedMagic.StringError, err)

			return nil, fmt.Errorf("create Otel metrics exporter failed: %w", err)
		}

		var metricsReader metricSdk.Reader
		if httpMetricsExporter != nil {
			metricsReader = metricSdk.NewPeriodicReader(httpMetricsExporter, metricSdk.WithInterval(MetricsTimeout))
		} else if grpcMetricsExporter != nil {
			metricsReader = metricSdk.NewPeriodicReader(grpcMetricsExporter, metricSdk.WithInterval(MetricsTimeout))
		}

		metricsOptions = append(metricsOptions, metricSdk.WithReader(metricsReader))
	}

	if settings.OTLPConsole {
		stdoutMetrics, err := stdoutMetricExporterNewFn(stdoutMetricExporter.WithPrettyPrint())
		if err != nil {
			slogger.Error("create STDOUT metrics failed", cryptoutilSharedMagic.StringError, err)

			return nil, fmt.Errorf("create STDOUT metrics failed: %w", err)
		}

		metricSdk.NewPeriodicReader(stdoutMetrics)
		metricsOptions = append(metricsOptions, metricSdk.WithReader(metricSdk.NewPeriodicReader(stdoutMetrics, metricSdk.WithInterval(MetricsTimeout))))
	}

	metricsProvider := metricSdk.NewMeterProvider(metricsOptions...)

	slogger.Debug("initialized metrics provider")

	return metricsProvider, nil
}

func initTraces(ctx context.Context, slogger *stdoutLogExporter.Logger, settings *TelemetrySettings) (*traceSdk.TracerProvider, error) {
	if settings.VerboseMode {
		slogger.Debug("initializing traces provider")
	}

	var tracesOptions []traceSdk.TracerProviderOption

	otelMeterTracerResource, _ := resourceSdk.New(ctx, resourceSdk.WithAttributes(getOtelMetricsTracesAttributes(settings)...))

	tracesOptions = append(tracesOptions, traceSdk.WithResource(otelMeterTracerResource))

	if settings.OTLPEnabled {
		isHTTP, isHTTPS, isGRPC, isGRPCS, endpoint, _ := parseProtocolAndEndpoint(&settings.OTLPEndpoint)

		var tracesSpanExporter traceSdk.SpanExporter

		var err error
		if isHTTP {
			tracesSpanExporter, err = httpTraceExporterotlptracehttp.New(ctx, httpTraceExporterotlptracehttp.WithEndpoint(*endpoint), httpTraceExporterotlptracehttp.WithInsecure())
		} else if isHTTPS {
			tracesSpanExporter, err = httpTraceExporterotlptracehttp.New(ctx, httpTraceExporterotlptracehttp.WithEndpoint(*endpoint))
		} else if isGRPC {
			tracesSpanExporter, err = grpcTraceExporterotlptracegrpc.New(ctx, grpcTraceExporterotlptracegrpc.WithEndpoint(*endpoint), grpcTraceExporterotlptracegrpc.WithInsecure())
		} else if isGRPCS {
			tracesSpanExporter, err = grpcTraceExporterotlptracegrpc.New(ctx, grpcTraceExporterotlptracegrpc.WithEndpoint(*endpoint))
		}

		if err != nil {
			slogger.Error("create Otel traces exporter failed", cryptoutilSharedMagic.StringError, err)

			return nil, fmt.Errorf("create Otel traces exporter failed: %w", err)
		}

		tracesOptions = append(tracesOptions, traceSdk.WithSpanProcessor(traceSdk.NewBatchSpanProcessor(tracesSpanExporter,
			traceSdk.WithBatchTimeout(TracesTimeout),
			traceSdk.WithMaxExportBatchSize(MaxTracesBatchSize))))
	}

	if settings.OTLPConsole {
		stdoutTraces, err := stdoutTraceExporterNewFn(stdoutTraceExporter.WithPrettyPrint())
		if err != nil {
			slogger.Error("create STDOUT traces failed", cryptoutilSharedMagic.StringError, err)

			return nil, fmt.Errorf("create STDOUT traces failed: %w", err)
		}

		tracesOptions = append(tracesOptions, traceSdk.WithSpanProcessor(traceSdk.NewBatchSpanProcessor(stdoutTraces,
			traceSdk.WithBatchTimeout(TracesTimeout),
			traceSdk.WithMaxExportBatchSize(MaxTracesBatchSize))))
	}

	tracesProvider := traceSdk.NewTracerProvider(tracesOptions...)

	slogger.Debug("initialized traces provider")

	return tracesProvider, nil
}
