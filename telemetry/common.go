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
	OtelGrpcPush     = "127.0.0.1:4317"
	LogsTimeout      = 500 * time.Millisecond
	MetricsTimeout   = 500 * time.Millisecond
	TracesTimeout    = 500 * time.Millisecond
	AttrEnvKey       = "env"
	AttrEnvValue     = "dev"
	AttrHostKey      = "host"
	AttrHostValue    = "localhost"
	AttrServiceKey   = "service"
	AttrServiceValue = "cryptoutil"
	AttrSourceKey    = "source"
	AttrSourceValue  = "go"
	AttrVersionKey   = "version"
	AttrVersionValue = "0.0.1"
)

var stdoutAttributes = []slog.Attr{
	{Key: AttrEnvKey, Value: slog.StringValue(AttrEnvValue)},
	{Key: AttrHostKey, Value: slog.StringValue(AttrHostValue)},
	{Key: AttrServiceKey, Value: slog.StringValue(AttrServiceValue)},
	{Key: AttrSourceKey, Value: slog.StringValue(AttrSourceValue)},
}

var otelLogsAttributes = []attribute.KeyValue{
	{Key: AttrEnvKey, Value: attribute.StringValue(AttrEnvValue)},
	{Key: AttrHostKey, Value: attribute.StringValue(AttrHostValue)},
	{Key: AttrServiceKey, Value: attribute.StringValue(AttrServiceValue)},
	{Key: AttrSourceKey, Value: attribute.StringValue(AttrSourceValue)},
}

var otelMetricsTracesAttributes = []attribute.KeyValue{
	{Key: AttrEnvKey, Value: attribute.StringValue(AttrEnvValue)},
	{Key: AttrHostKey, Value: attribute.StringValue(AttrHostValue)},
	{Key: AttrServiceKey, Value: attribute.StringValue(AttrServiceValue)},
	{Key: AttrSourceKey, Value: attribute.StringValue(AttrSourceValue)},
	{Key: AttrVersionKey, Value: attribute.StringValue(AttrVersionValue)},
}

type Service struct {
	startTime       time.Time
	Slogger         *slog.Logger
	LogsProvider    *log.LoggerProvider
	MetricsProvider *metric.MeterProvider
	TracesProvider  *trace.TracerProvider
}

func Init(ctx context.Context, startTime time.Time, scope string, enableOtel, enableStdout bool) *Service {
	slogger, logsProvider := InitLogger(ctx, enableOtel, scope)
	slogger.Info("Start", "uptime", time.Since(startTime).Seconds())
	metricsProvider := InitMetrics(ctx, enableOtel, enableStdout)
	tracesProvider := InitTraces(ctx, enableOtel, enableStdout)
	return &Service{startTime: startTime, Slogger: slogger, LogsProvider: logsProvider, MetricsProvider: metricsProvider, TracesProvider: tracesProvider}
}

func Shutdown(service *Service) {
	func() {
		ctx := context.Background()
		if service.TracesProvider != nil {
			if err := service.TracesProvider.Shutdown(ctx); err != nil {
				service.Slogger.Info("traces provider shutdown failed", "error", fmt.Errorf("traces provider shutdown error: %w", err))
			}
		}
		if service.MetricsProvider != nil {
			if err := service.MetricsProvider.Shutdown(ctx); err != nil {
				service.Slogger.Info("metrics provider shutdown failed", "error", fmt.Errorf("metrics provider shutdown error: %w", err))
			}
		}
		if service.LogsProvider != nil {
			service.Slogger.Info("Stop", "uptime", time.Since(service.startTime).Seconds())
			if err := service.LogsProvider.Shutdown(ctx); err != nil {
				service.Slogger.Info("logs provider shutdown failed", "error", fmt.Errorf("logs provider shutdown error: %w", err))
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
