package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"

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
	LogAttrSourceValue  = "golang"
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

func logInitErrorAndExit(format string, err error) {
	if err != nil {
		fmt.Printf(format, err)
		os.Exit(-1)
	}
}

func Init(ctx context.Context) (*slog.Logger, *log.LoggerProvider, *metric.MeterProvider, *trace.TracerProvider) {
	slogger, logsProvider := InitLogger(ctx, false, "main")
	metricsProvider := InitMetrics(ctx, false, true)
	tracesProvider := InitTraces(ctx, false, true)
	return slogger, logsProvider, metricsProvider, tracesProvider
}

func Shutdown(slogger *slog.Logger, tracesProvider *trace.TracerProvider, metricsProvider *metric.MeterProvider, logsProvider *log.LoggerProvider) {
	func() {
		if tracesProvider != nil {
			if err := tracesProvider.Shutdown(context.Background()); err != nil {
				slogger.Info("traces provider shutdown failed", "error", fmt.Errorf("traces provider shutdown failed: %w", err))
			}
		}
		if metricsProvider != nil {
			if err := metricsProvider.Shutdown(context.Background()); err != nil {
				slogger.Info("metrics provider shutdown failed", "error", fmt.Errorf("metrics provider shutdown failed: %w", err))
			}
		}
		if logsProvider != nil {
			if err := logsProvider.Shutdown(context.Background()); err != nil {
				slogger.Info("logs provider shutdown failed", "error", fmt.Errorf("logs provider shutdown failed: %w", err))
			}
		}
	}()
}
