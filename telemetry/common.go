package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

const (
	OtelGrpcPush        = "127.0.0.1:4317"
	LogAttrEnvKey       = "env"
	LogAttrEnvValue     = "dev"
	LogAttrHostKey      = "host"
	LogAttrHostValue    = "localhost"
	LogAttrServiceKey   = "service"
	LogAttrServiceValue = "cryptoutil"
	LogAttrSourceKey    = "source"
	LogAttrSourceValue  = "go"
	LogAttrVersionKey   = "version"
	LogAttrVersionValue = "0.0.1"
)

var stdoutLogAttributes = []slog.Attr{
	{Key: LogAttrEnvKey, Value: slog.StringValue(LogAttrEnvValue)},
	{Key: LogAttrHostKey, Value: slog.StringValue(LogAttrHostValue)},
	{Key: LogAttrServiceKey, Value: slog.StringValue(LogAttrServiceValue)},
	{Key: LogAttrSourceKey, Value: slog.StringValue(LogAttrSourceValue)},
}

var otelLogsAttributes = []attribute.KeyValue{
	{Key: LogAttrEnvKey, Value: attribute.StringValue(LogAttrEnvValue)},
	{Key: LogAttrHostKey, Value: attribute.StringValue(LogAttrHostValue)},
	{Key: LogAttrServiceKey, Value: attribute.StringValue(LogAttrServiceValue)},
	{Key: LogAttrSourceKey, Value: attribute.StringValue(LogAttrSourceValue)},
}

var otelMetricsTracesAttributes = []attribute.KeyValue{
	{Key: LogAttrEnvKey, Value: attribute.StringValue(LogAttrEnvValue)},
	{Key: LogAttrHostKey, Value: attribute.StringValue(LogAttrHostValue)},
	{Key: LogAttrServiceKey, Value: attribute.StringValue(LogAttrServiceValue)},
	{Key: LogAttrSourceKey, Value: attribute.StringValue(LogAttrSourceValue)},
	{Key: LogAttrVersionKey, Value: attribute.StringValue(LogAttrVersionValue)},
}

type Service struct {
	startTime       time.Time
	Slogger         *slog.Logger
	LogsProvider    *log.LoggerProvider
	MetricsProvider *metric.MeterProvider
	TracesProvider  *trace.TracerProvider
}

func Init(ctx context.Context, startTime time.Time, scope string) *Service {
	slogger, logsProvider := InitLogger(ctx, false, scope)
	slogger.Info("Start", "uptime", time.Since(startTime).Seconds())
	metricsProvider := InitMetrics(ctx, false, true)
	tracesProvider := InitTraces(ctx, false, true)
	return &Service{startTime: startTime, Slogger: slogger, LogsProvider: logsProvider, MetricsProvider: metricsProvider, TracesProvider: tracesProvider}
}

func Shutdown(service *Service) {
	func() {
		if service.TracesProvider != nil {
			if err := service.TracesProvider.Shutdown(context.Background()); err != nil {
				service.Slogger.Info("traces provider shutdown failed", "error", fmt.Errorf("traces provider shutdown failed: %w", err))
			}
		}
		if service.MetricsProvider != nil {
			if err := service.MetricsProvider.Shutdown(context.Background()); err != nil {
				service.Slogger.Info("metrics provider shutdown failed", "error", fmt.Errorf("metrics provider shutdown failed: %w", err))
			}
		}
		if service.LogsProvider != nil {
			service.Slogger.Info("Stop", "uptime", time.Since(service.startTime).Seconds())
			if err := service.LogsProvider.Shutdown(context.Background()); err != nil {
				service.Slogger.Info("logs provider shutdown failed", "error", fmt.Errorf("logs provider shutdown failed: %w", err))
			}
		}
	}()
}

func ifErrorLogAndExit(format string, err error) {
	if err != nil {
		fmt.Printf(format, err)
		os.Exit(-1)
	}
}
