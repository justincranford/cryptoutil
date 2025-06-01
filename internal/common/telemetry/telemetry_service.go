package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	googleUuid "github.com/google/uuid"
	slogMulti "github.com/samber/slog-multi"
	otelSlogBridge "go.opentelemetry.io/contrib/bridges/otelslog"

	stdoutLogExporter "log/slog"

	grpcLogExporter "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	grpcMetricExporter "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	grpcTraceExporterotlptracegrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
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

// TelemetryService Composite of OpenTelemetry providers for Logs, Metrics, and Traces
type TelemetryService struct {
	StartTime          time.Time
	Slogger            *stdoutLogExporter.Logger
	LogsProvider       logApi.LoggerProvider
	MetricsProvider    metricApi.MeterProvider
	TracesProvider     traceApi.TracerProvider
	TextMapPropagator  *propagationApi.TextMapPropagator
	logsProviderSdk    *logSdk.LoggerProvider   // Not exported, but still needed to do shutdown
	metricsProviderSdk *metricSdk.MeterProvider // Not exported, but still needed to do shutdown
	tracesProviderSdk  *traceSdk.TracerProvider // Not exported, but still needed to do shutdown
}

const (
	OtelGrpcPush      = "127.0.0.1:4317"
	LogsTimeout       = 500 * time.Millisecond
	MetricsTimeout    = 500 * time.Millisecond
	TracesTimeout     = 500 * time.Millisecond
	ForceFlushTimeout = 3 * time.Second
)

var (
	AttrEnv               = "dev"
	AttrHostName          = "localhost"
	AttrServiceName       = "cryptoutil"
	AttrServiceVersion    = "0.0.1"
	AttrServiceInstanceID = func() string {
		return googleUuid.Must(googleUuid.NewV7()).String()
	}()
)

var otelMetricsTracesAttributes = []attributeApi.KeyValue{
	oltpSemanticConventions.DeploymentID(AttrEnv),                    // deployment.environment.name (e.g. local-standalone, adhoc, dev, qa, preprod, prod)
	oltpSemanticConventions.HostName(AttrHostName),                   // service.instance.id (e.g. 12)
	oltpSemanticConventions.ServiceName(AttrServiceName),             // service.name (e.g. cryptoutil)
	oltpSemanticConventions.ServiceVersion(AttrServiceVersion),       // service.version (e.g. 0.0.1, 1.0.2, 2.1.0)
	oltpSemanticConventions.ServiceInstanceID(AttrServiceInstanceID), // service.instance.id (e.g. 12, uuidV7)
}

var otelLogsAttributes = otelMetricsTracesAttributes // same (for now)

var slogStdoutAttributes = func() []stdoutLogExporter.Attr {
	var slogAttrs []stdoutLogExporter.Attr
	for _, otelLogAttr := range otelLogsAttributes {
		slogAttrs = append(slogAttrs, stdoutLogExporter.String(string(otelLogAttr.Key), otelLogAttr.Value.AsString()))
	}
	return slogAttrs
}()

func NewTelemetryService(ctx context.Context, scope string, enableOtel, enableStdout bool) (*TelemetryService, error) {
	startTime := time.Now().UTC()
	if ctx == nil {
		return nil, fmt.Errorf("context must be non-nil")
	} else if len(scope) == 0 {
		return nil, fmt.Errorf("scope must be non-empty")
	}
	slogger, logsProvider, err := initLogger(ctx, enableOtel, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}
	metricsProvider, err := initMetrics(ctx, slogger, enableOtel, enableStdout)
	if err != nil {
		return nil, fmt.Errorf("failed to init metrics: %w", err)
	}
	tracesProvider, err := initTraces(ctx, slogger, enableOtel, enableStdout)
	if err != nil {
		return nil, fmt.Errorf("failed to init traces: %w", err)
	}
	textMapPropagator, err := initTextMapPropagator(slogger)
	if err != nil {
		return nil, fmt.Errorf("failed to init text map propagator: %w", err)
	}
	doExampleTracesSpans(ctx, tracesProvider, slogger)
	return &TelemetryService{
		StartTime:          startTime,
		Slogger:            slogger,
		LogsProvider:       logsProvider,
		MetricsProvider:    metricsProvider,
		TracesProvider:     tracesProvider,
		TextMapPropagator:  textMapPropagator,
		logsProviderSdk:    logsProvider,
		metricsProviderSdk: metricsProvider,
		tracesProviderSdk:  tracesProvider,
	}, nil
}

func (s *TelemetryService) Shutdown() {
	s.Slogger.Debug("stopping telemetry")

	forceFlushCtx, forceFlushCancelDueToTimeout := context.WithTimeout(context.Background(), ForceFlushTimeout)
	defer forceFlushCancelDueToTimeout()
	if s.tracesProviderSdk != nil {
		s.Slogger.Debug("traces provider force flushing", "uptime", time.Since(s.StartTime).Seconds())
		startTimeForceFlush := time.Now().UTC()
		if err := s.tracesProviderSdk.ForceFlush(forceFlushCtx); err != nil {
			s.Slogger.Error("traces provider force flush failed", "error", fmt.Errorf("traces provider force flush error: %w", err))
		}
		s.Slogger.Debug("traces provider force flushed", "uptime", time.Since(s.StartTime).Seconds(), "flush", time.Since(startTimeForceFlush).Seconds())
	}
	if s.metricsProviderSdk != nil {
		s.Slogger.Debug("metrics provider force flushing", "uptime", time.Since(s.StartTime).Seconds())
		startTimeForceFlush := time.Now().UTC()
		if err := s.metricsProviderSdk.ForceFlush(forceFlushCtx); err != nil {
			s.Slogger.Error("metrics provider force flush failed", "error", fmt.Errorf("metrics provider force flush error: %w", err))
		}
		s.Slogger.Info("metrics provider force flushed", "uptime", time.Since(s.StartTime).Seconds(), "flush", time.Since(startTimeForceFlush).Seconds())
	}
	if s.logsProviderSdk != nil {
		s.Slogger.Debug("logs provider force flushing", "uptime", time.Since(s.StartTime).Seconds())
		startTimeForceFlush := time.Now().UTC()
		if err := s.logsProviderSdk.ForceFlush(forceFlushCtx); err != nil {
			s.Slogger.Error("logs provider force flush failed", "error", fmt.Errorf("logs provider force flush error: %w", err))
		}
		s.Slogger.Debug("logs provider force flushed", "uptime", time.Since(s.StartTime).Seconds(), "flush", time.Since(startTimeForceFlush).Seconds())
	}

	shutdownCtx := context.Background()
	s.TextMapPropagator = nil
	if s.tracesProviderSdk != nil {
		s.Slogger.Debug("traces provider shutting down", "uptime", time.Since(s.StartTime).Seconds())
		startTimeShutdown := time.Now().UTC()
		if err := s.tracesProviderSdk.Shutdown(shutdownCtx); err != nil {
			s.Slogger.Error("traces provider shutdown failed", "error", fmt.Errorf("traces provider shutdown error: %w", err))
		}
		s.Slogger.Debug("traces provider shut down", "uptime", time.Since(s.StartTime).Seconds(), "flush", time.Since(startTimeShutdown).Seconds())
		s.tracesProviderSdk = nil
		s.TracesProvider = nil
	}
	if s.metricsProviderSdk != nil {
		s.Slogger.Debug("metrics provider shutting down", "uptime", time.Since(s.StartTime).Seconds())
		startTimeShutdown := time.Now().UTC()
		if err := s.metricsProviderSdk.Shutdown(shutdownCtx); err != nil {
			s.Slogger.Error("metrics provider shutdown failed", "error", fmt.Errorf("metrics provider shutdown error: %w", err))
		}
		s.Slogger.Debug("metrics provider shut down", "uptime", time.Since(s.StartTime).Seconds(), "flush", time.Since(startTimeShutdown).Seconds())
		s.metricsProviderSdk = nil
		s.MetricsProvider = nil
	}
	if s.logsProviderSdk != nil {
		s.Slogger.Debug("logs provider shutting down", "uptime", time.Since(s.StartTime).Seconds())
		startTimeShutdown := time.Now().UTC()
		if err := s.logsProviderSdk.Shutdown(shutdownCtx); err != nil {
			s.Slogger.Error("logs provider shutdown failed", "error", fmt.Errorf("logs provider shutdown error: %w", err))
		}
		s.Slogger.Debug("logs provider shut down", "uptime", time.Since(s.StartTime).Seconds(), "flush", time.Since(startTimeShutdown).Seconds())
		s.Slogger = nil
		s.logsProviderSdk = nil
		s.LogsProvider = nil
	}
}

func initLogger(ctx context.Context, enableOtel bool, otelLoggerName string) (*stdoutLogExporter.Logger, *logSdk.LoggerProvider, error) {
	stdoutSlogHandler := stdoutLogExporter.NewTextHandler(os.Stdout, nil).WithAttrs(slogStdoutAttributes)
	slogger := stdoutLogExporter.New(stdoutSlogHandler)
	slogger.Debug("initializing otel logs provider")

	otelLogsResource := resourceSdk.NewWithAttributes("", otelLogsAttributes...)
	otelExporter, err := grpcLogExporter.New(ctx, grpcLogExporter.WithEndpoint(OtelGrpcPush), grpcLogExporter.WithInsecure())
	if err != nil {
		slogger.Error("create Otel GRPC logger failed", "error", err)
	}
	otelProviderOptions := []logSdk.LoggerProviderOption{
		logSdk.WithResource(otelLogsResource),
		logSdk.WithProcessor(logSdk.NewBatchProcessor(otelExporter, logSdk.WithExportTimeout(LogsTimeout))),
	}
	otelProvider := logSdk.NewLoggerProvider(otelProviderOptions...)

	if enableOtel {
		otelSlogHandler := otelSlogBridge.NewHandler(otelLoggerName, otelSlogBridge.WithLoggerProvider(otelProvider))
		slogger = stdoutLogExporter.New(slogMulti.Fanout(stdoutSlogHandler, otelSlogHandler))
	}

	slogger.Debug("initialized otel logs provider")
	return slogger, otelProvider, nil
}

func initMetrics(ctx context.Context, slogger *stdoutLogExporter.Logger, enableOtel bool, enableStdout bool) (*metricSdk.MeterProvider, error) {
	slogger.Debug("initializing metrics provider")

	var metricsOptions []metricSdk.Option

	otelMeterTracerTags, err := resourceSdk.New(ctx, resourceSdk.WithAttributes(otelMetricsTracesAttributes...))
	if err != nil {
		slogger.Error("create Otel GRPC metrics resource failed", "error", err)
	}
	metricsOptions = append(metricsOptions, metricSdk.WithResource(otelMeterTracerTags))

	if enableOtel {
		otelGrpcMetrics, err := grpcMetricExporter.New(ctx, grpcMetricExporter.WithEndpoint(OtelGrpcPush), grpcMetricExporter.WithInsecure())
		if err != nil {
			slogger.Error("create Otel GRPC metrics failed", "error", err)
		}
		metricsOptions = append(metricsOptions, metricSdk.WithReader(metricSdk.NewPeriodicReader(otelGrpcMetrics, metricSdk.WithInterval(MetricsTimeout))))
	}

	if enableStdout {
		stdoutMetrics, err := stdoutMetricExporter.New(stdoutMetricExporter.WithPrettyPrint())
		if err != nil {
			slogger.Error("create STDOUT metrics failed", "error", err)
		}
		metricSdk.NewPeriodicReader(stdoutMetrics)
		metricsOptions = append(metricsOptions, metricSdk.WithReader(metricSdk.NewPeriodicReader(stdoutMetrics, metricSdk.WithInterval(MetricsTimeout))))
	}

	metricsProvider := metricSdk.NewMeterProvider(metricsOptions...)
	slogger.Debug("initialized metrics provider")
	return metricsProvider, nil
}

func initTraces(ctx context.Context, slogger *stdoutLogExporter.Logger, enableOtel bool, enableStdout bool) (*traceSdk.TracerProvider, error) {
	slogger.Debug("initializing traces provider")

	var tracesOptions []traceSdk.TracerProviderOption

	otelMeterTracerResource, err := resourceSdk.New(ctx, resourceSdk.WithAttributes(otelMetricsTracesAttributes...))
	if err != nil {
		slogger.Error("create Otel GRPC traces resource failed", "error", err)
	}
	tracesOptions = append(tracesOptions, traceSdk.WithResource(otelMeterTracerResource))

	if enableOtel {
		tracerOtelGrpc, err := grpcTraceExporterotlptracegrpc.New(ctx, grpcTraceExporterotlptracegrpc.WithEndpoint(OtelGrpcPush), grpcTraceExporterotlptracegrpc.WithInsecure())
		if err != nil {
			slogger.Error("create Otel GRPC traces failed", "error", err)
		}
		tracesOptions = append(tracesOptions, traceSdk.WithSpanProcessor(traceSdk.NewBatchSpanProcessor(tracerOtelGrpc, traceSdk.WithBatchTimeout(TracesTimeout))))
	}

	if enableStdout {
		stdoutTraces, err := stdoutTraceExporter.New(stdoutTraceExporter.WithPrettyPrint())
		if err != nil {
			slogger.Error("create STDOUT traces failed", "error", err)
		}
		tracesOptions = append(tracesOptions, traceSdk.WithSpanProcessor(traceSdk.NewBatchSpanProcessor(stdoutTraces, traceSdk.WithBatchTimeout(TracesTimeout))))
	}

	tracesProvider := traceSdk.NewTracerProvider(tracesOptions...)
	slogger.Debug("initialized traces provider")
	return tracesProvider, nil
}

func initTextMapPropagator(slogger *stdoutLogExporter.Logger) (*propagationApi.TextMapPropagator, error) {
	textMapPropagator := propagationApi.NewCompositeTextMapPropagator(
		propagationApi.TraceContext{},
		propagationApi.Baggage{},
	)
	return &textMapPropagator, nil
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
