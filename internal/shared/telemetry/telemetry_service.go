// Copyright (c) 2025 Justin Cranford
//
//

package telemetry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
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

	attributeApi "go.opentelemetry.io/otel/attribute"
	logApi "go.opentelemetry.io/otel/log"
	metricApi "go.opentelemetry.io/otel/metric"
	propagationApi "go.opentelemetry.io/otel/propagation"
	traceApi "go.opentelemetry.io/otel/trace"

	logSdk "go.opentelemetry.io/otel/sdk/log"
	metricSdk "go.opentelemetry.io/otel/sdk/metric"
	resourceSdk "go.opentelemetry.io/otel/sdk/resource"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"

	oltpSemanticConventions "go.opentelemetry.io/otel/semconv/v1.30.0"
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
	logsProviderSdk    *logSdk.LoggerProvider                                             // Not exported, but still needed to do shutdown
	metricsProviderSdk *metricSdk.MeterProvider                                           // Not exported, but still needed to do shutdown
	tracesProviderSdk  *traceSdk.TracerProvider                                           // Not exported, but still needed to do shutdown
	settings           *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings // Store settings for health checks
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
func NewTelemetryService(ctx context.Context, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) (*TelemetryService, error) {
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
	} else {
		// log the sidecar health check errors, and proceed anyway; if sidecar becomes healthy later, buffered telemetry exports can still go through later
		if settings.VerboseMode {
			slogger.Info("sidecar health check failed", "attempts", len(retryErrors), "errors", errors.Join(retryErrors...))
		} else {
			slogger.Info("DEBUG health check failed", "attempts", len(retryErrors), "errors", errors.Join(retryErrors...))
		}
	}

	metricsProvider, err := initMetrics(ctx, slogger, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to init metrics: %w", err)
	}

	tracesProvider, err := initTraces(ctx, slogger, settings)
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
					s.Slogger.Error("traces provider force flush failed", "error", fmt.Errorf("traces provider force flush error: %w", err))
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
					s.Slogger.Error("metrics provider force flush failed", "error", fmt.Errorf("metrics provider force flush error: %w", err))
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
					s.Slogger.Error("logs provider force flush failed", "error", fmt.Errorf("logs provider force flush error: %w", err))
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
					s.Slogger.Error("traces provider shutdown failed", "error", fmt.Errorf("traces provider shutdown error: %w", err))
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
					s.Slogger.Error("metrics provider shutdown failed", "error", fmt.Errorf("metrics provider shutdown error: %w", err))
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
					s.Slogger.Error("logs provider shutdown failed", "error", fmt.Errorf("logs provider shutdown error: %w", err))
				} else if s.VerboseMode {
					s.Slogger.Debug("logs provider shut down", "uptime", time.Since(s.StartTime).Seconds(), "flush", time.Since(startTimeShutdownLogsProvider).Seconds())
				}
			}

			s.Slogger.Debug("telemetry providers shut down", "uptime", time.Since(s.StartTime).Seconds(), "flush", time.Since(startTimeShutdown).Seconds())
		}
	})
}

func initLogger(ctx context.Context, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) (*stdoutLogExporter.Logger, *logSdk.LoggerProvider, error) {
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
		slogger.Error("parse protocol and endpoint failed", "error", err)

		return nil, nil, fmt.Errorf("parse protocol and endpoint failed: %w", err)
	}

	var otelExporter logSdk.Exporter
	if isHTTP {
		otelExporter, err = httpLogExporter.New(ctx, httpLogExporter.WithEndpoint(*endpoint), httpLogExporter.WithInsecure())
	} else if isHTTPS {
		otelExporter, err = httpLogExporter.New(ctx, httpLogExporter.WithEndpoint(*endpoint))
	} else if isGRPC {
		otelExporter, err = grpcLogExporter.New(ctx, grpcLogExporter.WithEndpoint(*endpoint), grpcLogExporter.WithInsecure())
	} else if isGRPCS {
		otelExporter, err = grpcLogExporter.New(ctx, grpcLogExporter.WithEndpoint(*endpoint))
	} else {
		return nil, nil, fmt.Errorf("unsupported protocol for endpoint: %s", settings.OTLPEndpoint)
	}

	if err != nil {
		slogger.Error("create Otel logger exporter failed", "error", err)

		return nil, nil, fmt.Errorf("create Otel logger exporter failed: %w", err)
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

func initMetrics(ctx context.Context, slogger *stdoutLogExporter.Logger, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) (*metricSdk.MeterProvider, error) {
	if settings.VerboseMode {
		slogger.Debug("initializing metrics provider")
	}

	var metricsOptions []metricSdk.Option

	otelMeterTracerTags, err := resourceSdk.New(ctx, resourceSdk.WithAttributes(getOtelMetricsTracesAttributes(settings)...))
	if err != nil {
		slogger.Error("create Otel GRPC metrics resource failed", "error", err)

		return nil, fmt.Errorf("create Otel GRPC metrics resource failed: %w", err)
	}

	metricsOptions = append(metricsOptions, metricSdk.WithResource(otelMeterTracerTags))

	if settings.OTLPEnabled {
		isHTTP, isHTTPS, isGRPC, isGRPCS, endpoint, err := parseProtocolAndEndpoint(&settings.OTLPEndpoint)
		if err != nil {
			slogger.Error("parse protocol and endpoint failed", "error", err)

			return nil, fmt.Errorf("parse protocol and endpoint failed: %w", err)
		}

		var httpMetricsExporter *httpMetricExporter.Exporter

		var grpcMetricsExporter *grpcMetricExporter.Exporter

		if isHTTP {
			httpMetricsExporter, err = httpMetricExporter.New(ctx, httpMetricExporter.WithEndpoint(*endpoint), httpMetricExporter.WithInsecure())
		} else if isHTTPS {
			httpMetricsExporter, err = httpMetricExporter.New(ctx, httpMetricExporter.WithEndpoint(*endpoint))
		} else if isGRPC {
			grpcMetricsExporter, err = grpcMetricExporter.New(ctx, grpcMetricExporter.WithEndpoint(*endpoint), grpcMetricExporter.WithInsecure())
		} else if isGRPCS {
			grpcMetricsExporter, err = grpcMetricExporter.New(ctx, grpcMetricExporter.WithEndpoint(*endpoint))
		} else {
			slogger.Error("unsupported protocol for endpoint", "endpoint", settings.OTLPEndpoint)

			return nil, fmt.Errorf("unsupported protocol for endpoint: %s", settings.OTLPEndpoint)
		}

		if err != nil {
			slogger.Error("create Otel metrics exporter failed", "error", err)

			return nil, fmt.Errorf("create Otel metrics exporter failed: %w", err)
		}

		var metricsReader metricSdk.Reader
		if httpMetricsExporter != nil {
			metricsReader = metricSdk.NewPeriodicReader(httpMetricsExporter, metricSdk.WithInterval(MetricsTimeout))
		} else if grpcMetricsExporter != nil {
			metricsReader = metricSdk.NewPeriodicReader(grpcMetricsExporter, metricSdk.WithInterval(MetricsTimeout))
		} else {
			return nil, fmt.Errorf("no valid metrics exporter created for endpoint: %s", settings.OTLPEndpoint)
		}

		metricsOptions = append(metricsOptions, metricSdk.WithReader(metricsReader))
	}

	if settings.OTLPConsole {
		stdoutMetrics, err := stdoutMetricExporter.New(stdoutMetricExporter.WithPrettyPrint())
		if err != nil {
			slogger.Error("create STDOUT metrics failed", "error", err)

			return nil, fmt.Errorf("create STDOUT metrics failed: %w", err)
		}

		metricSdk.NewPeriodicReader(stdoutMetrics)
		metricsOptions = append(metricsOptions, metricSdk.WithReader(metricSdk.NewPeriodicReader(stdoutMetrics, metricSdk.WithInterval(MetricsTimeout))))
	}

	metricsProvider := metricSdk.NewMeterProvider(metricsOptions...)

	slogger.Debug("initialized metrics provider")

	return metricsProvider, nil
}

func initTraces(ctx context.Context, slogger *stdoutLogExporter.Logger, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) (*traceSdk.TracerProvider, error) {
	if settings.VerboseMode {
		slogger.Debug("initializing traces provider")
	}

	var tracesOptions []traceSdk.TracerProviderOption

	otelMeterTracerResource, err := resourceSdk.New(ctx, resourceSdk.WithAttributes(getOtelMetricsTracesAttributes(settings)...))
	if err != nil {
		slogger.Error("create Otel GRPC traces resource failed", "error", err)

		return nil, fmt.Errorf("create Otel GRPC traces resource failed: %w", err)
	}

	tracesOptions = append(tracesOptions, traceSdk.WithResource(otelMeterTracerResource))

	if settings.OTLPEnabled {
		isHTTP, isHTTPS, isGRPC, isGRPCS, endpoint, err := parseProtocolAndEndpoint(&settings.OTLPEndpoint)
		if err != nil {
			slogger.Error("parse protocol and endpoint failed", "error", err)

			return nil, fmt.Errorf("parse protocol and endpoint failed: %w", err)
		}

		var tracesSpanExporter traceSdk.SpanExporter
		if isHTTP {
			tracesSpanExporter, err = httpTraceExporterotlptracehttp.New(ctx, httpTraceExporterotlptracehttp.WithEndpoint(*endpoint), httpTraceExporterotlptracehttp.WithInsecure())
		} else if isHTTPS {
			tracesSpanExporter, err = httpTraceExporterotlptracehttp.New(ctx, httpTraceExporterotlptracehttp.WithEndpoint(*endpoint))
		} else if isGRPC {
			tracesSpanExporter, err = grpcTraceExporterotlptracegrpc.New(ctx, grpcTraceExporterotlptracegrpc.WithEndpoint(*endpoint), grpcTraceExporterotlptracegrpc.WithInsecure())
		} else if isGRPCS {
			tracesSpanExporter, err = grpcTraceExporterotlptracegrpc.New(ctx, grpcTraceExporterotlptracegrpc.WithEndpoint(*endpoint))
		} else {
			slogger.Error("unsupported protocol for endpoint", "endpoint", settings.OTLPEndpoint)

			return nil, fmt.Errorf("unsupported protocol for endpoint: %s", settings.OTLPEndpoint)
		}

		if err != nil {
			slogger.Error("create Otel traces exporter failed", "error", err)

			return nil, fmt.Errorf("create Otel traces exporter failed: %w", err)
		}

		tracesOptions = append(tracesOptions, traceSdk.WithSpanProcessor(traceSdk.NewBatchSpanProcessor(tracesSpanExporter,
			traceSdk.WithBatchTimeout(TracesTimeout),
			traceSdk.WithMaxExportBatchSize(MaxTracesBatchSize))))
	}

	if settings.OTLPConsole {
		stdoutTraces, err := stdoutTraceExporter.New(stdoutTraceExporter.WithPrettyPrint())
		if err != nil {
			slogger.Error("create STDOUT traces failed", "error", err)

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

func initTextMapPropagator() *propagationApi.TextMapPropagator {
	textMapPropagator := propagationApi.NewCompositeTextMapPropagator(
		propagationApi.TraceContext{},
		propagationApi.Baggage{},
	)

	return &textMapPropagator
}

func doExampleTracesSpans(ctx context.Context, tracesProvider *traceSdk.TracerProvider, slogger *stdoutLogExporter.Logger) {
	tracer := tracesProvider.Tracer("fiber-tracer")
	spanCtx := doExampleTraceSpan(ctx, tracer, slogger, "sample parent trace and span")

	doExampleTraceSpan(spanCtx, tracer, slogger, "sample child trace and span")
}

func doExampleTraceSpan(ctx context.Context, tracer traceApi.Tracer, slogger *stdoutLogExporter.Logger, message string) context.Context {
	spanCtx, span := tracer.Start(ctx, "test-span")
	slogger.Debug(message, "traceid", span.SpanContext().TraceID(), "traceid", span.SpanContext().SpanID())

	return spanCtx
}

func getOtelMetricsTracesAttributes(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) []attributeApi.KeyValue {
	return []attributeApi.KeyValue{
		oltpSemanticConventions.DeploymentID(settings.OTLPEnvironment),   // deployment.environment.name (e.g. local-standalone, adhoc, dev, qa, preprod, prod)
		oltpSemanticConventions.HostName(settings.OTLPHostname),          // service.instance.id (e.g. 12)
		oltpSemanticConventions.ServiceName(settings.OTLPService),        // service.name (e.g. cryptoutil)
		oltpSemanticConventions.ServiceVersion(settings.OTLPVersion),     // service.version (e.g. 0.0.1, 1.0.2, 2.1.0)
		oltpSemanticConventions.ServiceInstanceID(settings.OTLPInstance), // service.instance.id (e.g. 12, uuidV7)
	}
}

func getOtelLogsAttributes(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) []attributeApi.KeyValue {
	return getOtelMetricsTracesAttributes(settings) // same (for now)
}

func getSlogStdoutAttributes(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) []stdoutLogExporter.Attr {
	otelAttrs := getOtelLogsAttributes(settings)
	slogAttrs := make([]stdoutLogExporter.Attr, 0, len(otelAttrs))

	for _, otelLogAttr := range otelAttrs {
		slogAttrs = append(slogAttrs, stdoutLogExporter.String(string(otelLogAttr.Key), otelLogAttr.Value.AsString()))
	}

	return slogAttrs
}

func parseProtocolAndEndpoint(otlpEndpoint *string) (bool, bool, bool, bool, *string, error) {
	if after, ok := strings.CutPrefix(*otlpEndpoint, "http://"); ok {
		return true, false, false, false, &after, nil
	} else if after, ok := strings.CutPrefix(*otlpEndpoint, "https://"); ok {
		return false, true, false, false, &after, nil
	} else if after, ok := strings.CutPrefix(*otlpEndpoint, "grpc://"); ok {
		return false, false, true, false, &after, nil
	} else if after, ok := strings.CutPrefix(*otlpEndpoint, "grpcs://"); ok {
		return false, false, false, true, &after, nil
	}

	return false, false, false, false, nil, fmt.Errorf("invalid OTLP endpoint protocol, must start with https://, grpcs://, http://, or grpc://")
}

// checkSidecarHealth performs a connectivity check to the OTLP sidecar during startup.
func checkSidecarHealth(ctx context.Context, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) error {
	// Parse the endpoint to determine protocol and address
	isHTTP, isHTTPS, isGRPC, isGRPCS, endpoint, err := parseProtocolAndEndpoint(&settings.OTLPEndpoint)
	if err != nil {
		return fmt.Errorf("failed to parse OTLP endpoint: %w", err)
	}

	// For now, we do a basic connectivity check by attempting to create an exporter
	// This will fail if the sidecar is not reachable
	if isGRPC {
		_, err = grpcTraceExporterotlptracegrpc.New(ctx,
			grpcTraceExporterotlptracegrpc.WithEndpoint(*endpoint),
			grpcTraceExporterotlptracegrpc.WithInsecure())
		if err != nil {
			return fmt.Errorf("gRPC sidecar connectivity check failed: %w", err)
		}
	} else if isGRPCS {
		_, err = grpcTraceExporterotlptracegrpc.New(ctx,
			grpcTraceExporterotlptracegrpc.WithEndpoint(*endpoint))
		if err != nil {
			return fmt.Errorf("gRPCS sidecar connectivity check failed: %w", err)
		}
	} else if isHTTP {
		_, err = httpTraceExporterotlptracehttp.New(ctx,
			httpTraceExporterotlptracehttp.WithEndpoint(*endpoint),
			httpTraceExporterotlptracehttp.WithInsecure())
		if err != nil {
			return fmt.Errorf("HTTP sidecar connectivity check failed: %w", err)
		}
	} else if isHTTPS {
		_, err = httpTraceExporterotlptracehttp.New(ctx,
			httpTraceExporterotlptracehttp.WithEndpoint(*endpoint))
		if err != nil {
			return fmt.Errorf("HTTPS sidecar connectivity check failed: %w", err)
		}
	}

	return nil
}

// checkSidecarHealthWithRetry performs connectivity check to OTLP sidecar with retry logic, before init logsProvider; caller must log results.
func checkSidecarHealthWithRetry(ctx context.Context, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) ([]error, error) {
	maxRetries := cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries
	retryDelay := cryptoutilSharedMagic.DefaultSidecarHealthCheckRetryDelay

	var intermediateErrs []error

	//nolint:wsl // wsl requires blocks not to end with blank lines, but this structure improves readability
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return intermediateErrs, fmt.Errorf("context cancelled during sidecar health check retry: %w", ctx.Err())
			case <-time.After(retryDelay):
				// Continue with retry
			}
		}

		err := checkSidecarHealth(ctx, settings)
		if err == nil {
			return intermediateErrs, nil
		}

		intermediateErrs = append(intermediateErrs, err)
	}

	return intermediateErrs, fmt.Errorf("sidecar health check failed after %d attempts (%d failures): %w", maxRetries+1, len(intermediateErrs), errors.Join(intermediateErrs...))
}

// CheckSidecarHealth performs a connectivity check to the OTLP sidecar.
func (s *TelemetryService) CheckSidecarHealth(ctx context.Context) error {
	//nolint:wsl // wsl requires blocks not to end with blank lines, but this structure improves readability
	if s.settings.OTLPEnabled {
		err := checkSidecarHealth(ctx, s.settings)
		if err != nil {
			return fmt.Errorf("sidecar health check failed: %w", err)
		}
	}

	return nil
}
