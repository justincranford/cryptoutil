package telemetry

import (
	"context"
	"fmt"
	uuid2 "github.com/google/uuid"
	"go.opentelemetry.io/otel/propagation"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

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

// Service Composite of OpenTelemetry providers for Logs, Metrics, and Traces
type Service struct {
	startTime         time.Time
	stopTime          time.Time
	Slogger           *slog.Logger
	LogsProvider      *log.LoggerProvider
	MetricsProvider   *metric.MeterProvider
	TracesProvider    *trace.TracerProvider
	TextMapPropagator *propagation.TextMapPropagator
}

func Init(ctx context.Context, startTime time.Time, scope string, enableOtel, enableStdout bool) *Service {
	slogger, logsProvider := InitLogger(ctx, enableOtel, scope)
	slogger.Info("Start", "uptime", time.Since(startTime).Seconds())
	metricsProvider := InitMetrics(ctx, enableOtel, enableStdout)
	tracesProvider := InitTraces(ctx, enableOtel, enableStdout)
	textMapPropagator := InitTextMapPropagator()
	return &Service{
		startTime:         startTime,
		Slogger:           slogger,
		LogsProvider:      logsProvider,
		MetricsProvider:   metricsProvider,
		TracesProvider:    tracesProvider,
		TextMapPropagator: textMapPropagator,
	}
}

func Shutdown(service *Service) {
	func() {
		ctx := context.Background()
		if service.TracesProvider != nil {
			if err := service.TracesProvider.Shutdown(ctx); err != nil {
				service.Slogger.Info("traces provider shutdown failed", "error", fmt.Errorf("traces provider shutdown error: %w", err))
			}
			service.TracesProvider = nil
		}
		if service.MetricsProvider != nil {
			if err := service.MetricsProvider.Shutdown(ctx); err != nil {
				service.Slogger.Info("metrics provider shutdown failed", "error", fmt.Errorf("metrics provider shutdown error: %w", err))
			}
			service.MetricsProvider = nil
		}
		if service.LogsProvider != nil {
			service.Slogger.Info("Stop", "uptime", time.Since(service.startTime).Seconds())
			if err := service.LogsProvider.Shutdown(ctx); err != nil {
				service.Slogger.Info("logs provider shutdown failed", "error", fmt.Errorf("logs provider shutdown error: %w", err))
			}
			service.LogsProvider = nil
		}
		service.TextMapPropagator = nil
		service.Slogger = nil
		service.stopTime = time.Now().UTC()
	}()
}

func ifErrorLogAndExit(format string, err error) {
	if err != nil {
		fmt.Printf(format, err)
		os.Exit(-1)
	}
}
