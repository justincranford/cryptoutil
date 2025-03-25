package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	uuid2 "github.com/google/uuid"
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

// Service Composite of OpenTelemetry providers for Logs, Metrics, and Traces
type Service struct {
	StartTime         time.Time
	StopTime          time.Time
	Slogger           *slog.Logger
	LogsProvider      *log.LoggerProvider
	MetricsProvider   *metric.MeterProvider
	TracesProvider    *trace.TracerProvider
	TextMapPropagator *propagation.TextMapPropagator
}

const (
	OtelGrpcPush   = "127.0.0.1:4317"
	LogsTimeout    = 500 * time.Millisecond
	MetricsTimeout = 500 * time.Millisecond
	TracesTimeout  = 500 * time.Millisecond
)

var (
	AttrEnv               = "dev"
	AttrHostName          = "localhost"
	AttrServiceName       = "cryptoutil"
	AttrServiceVersion    = "0.0.1"
	AttrServiceInstanceID = func() string {
		uuid, err := uuid2.NewV7()
		if err != nil {
			os.Exit(1)
		}
		return uuid.String()
	}()
)

var otelMetricsTracesAttributes = []attribute.KeyValue{
	semconv.DeploymentID(AttrEnv),                    // deployment.environment.name (e.g. local-standalone, adhoc, dev, qa, preprod, prod)
	semconv.HostName(AttrHostName),                   // service.instance.id (e.g. 12)
	semconv.ServiceName(AttrServiceName),             // service.name (e.g. cryptoutil)
	semconv.ServiceVersion(AttrServiceVersion),       // service.version (e.g. 0.0.1, 1.0.2, 2.1.0)
	semconv.ServiceInstanceID(AttrServiceInstanceID), // service.instance.id (e.g. 12, uuidV7)
}

var otelLogsAttributes = otelMetricsTracesAttributes // same (for now)

var slogStdoutAttributes = func() []slog.Attr {
	var slogAttrs []slog.Attr
	for _, otelLogAttr := range otelLogsAttributes {
		slogAttrs = append(slogAttrs, slog.String(string(otelLogAttr.Key), otelLogAttr.Value.AsString()))
	}
	return slogAttrs
}()

func NewService(ctx context.Context, scope string, enableOtel, enableStdout bool) *Service {
	startTime := time.Now().UTC()
	slogger, logsProvider := initLogger(ctx, enableOtel, scope)
	metricsProvider := initMetrics(ctx, enableOtel, enableStdout)
	tracesProvider := initTraces(ctx, enableOtel, enableStdout)
	textMapPropagator := initTextMapPropagator()
	return &Service{
		StartTime:         startTime,
		Slogger:           slogger,
		LogsProvider:      logsProvider,
		MetricsProvider:   metricsProvider,
		TracesProvider:    tracesProvider,
		TextMapPropagator: textMapPropagator,
	}
}

func (s *Service) Shutdown(ctx context.Context) {
	if s.TracesProvider != nil {
		if err := s.TracesProvider.Shutdown(ctx); err != nil {
			s.Slogger.Info("traces provider shutdown failed", "error", fmt.Errorf("traces provider shutdown error: %w", err))
		}
		s.TracesProvider = nil
	}
	if s.MetricsProvider != nil {
		if err := s.MetricsProvider.Shutdown(ctx); err != nil {
			s.Slogger.Info("metrics provider shutdown failed", "error", fmt.Errorf("metrics provider shutdown error: %w", err))
		}
		s.MetricsProvider = nil
	}
	if s.LogsProvider != nil {
		s.Slogger.Info("Stop", "uptime", time.Since(s.StartTime).Seconds())
		if err := s.LogsProvider.Shutdown(ctx); err != nil {
			s.Slogger.Info("logs provider shutdown failed", "error", fmt.Errorf("logs provider shutdown error: %w", err))
		}
		s.LogsProvider = nil
	}
	s.TextMapPropagator = nil
	s.Slogger = nil
	s.StopTime = time.Now().UTC()
}

func ifErrorLogAndExit(format string, err error) {
	if err != nil {
		fmt.Printf(format, err)
		os.Exit(-1)
	}
}

func initLogger(ctx context.Context, enableOtel bool, otelLoggerName string) (*slog.Logger, *log.LoggerProvider) {
	stdoutSlogHandler := slog.NewTextHandler(os.Stdout, nil).WithAttrs(slogStdoutAttributes)

	if enableOtel {
		otelLogsResource := resource.NewWithAttributes("", otelLogsAttributes...)

		otelExporter, err := otlploggrpc.New(ctx, otlploggrpc.WithEndpoint(OtelGrpcPush), otlploggrpc.WithInsecure())
		ifErrorLogAndExit("create Otel GRPC logger failed: %v", err)
		otelProviderOptions := []log.LoggerProviderOption{
			log.WithResource(otelLogsResource),
			log.WithProcessor(log.NewBatchProcessor(otelExporter, log.WithExportTimeout(LogsTimeout))),
		}
		otelProvider := log.NewLoggerProvider(otelProviderOptions...)
		otelSlogHandler := otelslog.NewHandler(otelLoggerName, otelslog.WithLoggerProvider(otelProvider))

		return slog.New(slogmulti.Fanout(stdoutSlogHandler, otelSlogHandler)), otelProvider
	}

	return slog.New(stdoutSlogHandler), nil
}

func initMetrics(ctx context.Context, enableOtel bool, enableStdout bool) *metric.MeterProvider {
	var metricsOptions []metric.Option

	otelMeterTracerTags, err := resource.New(ctx, resource.WithAttributes(otelMetricsTracesAttributes...))
	ifErrorLogAndExit("create Otel GRPC metrics resource failed: %v", err)
	metricsOptions = append(metricsOptions, metric.WithResource(otelMeterTracerTags))

	if enableOtel {
		otelGrpcMetrics, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(OtelGrpcPush), otlpmetricgrpc.WithInsecure())
		ifErrorLogAndExit("create Otel GRPC metrics failed: %v", err)
		metricsOptions = append(metricsOptions, metric.WithReader(metric.NewPeriodicReader(otelGrpcMetrics, metric.WithInterval(MetricsTimeout))))
	}

	if enableStdout {
		stdoutMetrics, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
		ifErrorLogAndExit("create STDOUT metrics failed: %v", err)
		metric.NewPeriodicReader(stdoutMetrics)
		metricsOptions = append(metricsOptions, metric.WithReader(metric.NewPeriodicReader(stdoutMetrics, metric.WithInterval(MetricsTimeout))))
	}

	return metric.NewMeterProvider(metricsOptions...)
}

func initTraces(ctx context.Context, enableOtel bool, enableStdout bool) *trace.TracerProvider {
	var tracesOptions []trace.TracerProviderOption

	otelMeterTracerResource, err := resource.New(ctx, resource.WithAttributes(otelMetricsTracesAttributes...))
	ifErrorLogAndExit("create Otel GRPC traces resource failed: %v", err)
	tracesOptions = append(tracesOptions, trace.WithResource(otelMeterTracerResource))

	if enableOtel {
		tracerOtelGrpc, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(OtelGrpcPush), otlptracegrpc.WithInsecure())
		ifErrorLogAndExit("create Otel GRPC traces failed: %v", err)
		tracesOptions = append(tracesOptions, trace.WithSpanProcessor(trace.NewBatchSpanProcessor(tracerOtelGrpc, trace.WithBatchTimeout(TracesTimeout))))
	}

	if enableStdout {
		stdoutTraces, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		ifErrorLogAndExit("create STDOUT traces failed: %v", err)
		tracesOptions = append(tracesOptions, trace.WithSpanProcessor(trace.NewBatchSpanProcessor(stdoutTraces, trace.WithBatchTimeout(TracesTimeout))))
	}

	return trace.NewTracerProvider(tracesOptions...)
}

func initTextMapPropagator() *propagation.TextMapPropagator {
	textMapPropagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	return &textMapPropagator
}
