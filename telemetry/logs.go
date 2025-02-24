package telemetry

import (
	"context"
	"log/slog"
	"os"
	"time"

	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

func InitLogger(ctx context.Context, enableOtel bool, otelLoggerName string) (*slog.Logger, *log.LoggerProvider) {
	stdoutSlogHandler := slog.NewTextHandler(os.Stdout, nil).WithAttrs(stdoutLogAttributes)

	if enableOtel {
		otelLogsResource := resource.NewWithAttributes("", otelLogsAttributes...)

		otelExporter, err := otlploggrpc.New(ctx, otlploggrpc.WithEndpoint(OtelGrpcPush), otlploggrpc.WithInsecure())
		logInitErrorAndExit("create Otel GRPC logger failed: %v", err)
		otelProviderOptions := []log.LoggerProviderOption{
			log.WithResource(otelLogsResource),
			log.WithProcessor(log.NewBatchProcessor(otelExporter, log.WithExportTimeout(500*time.Millisecond))),
		}
		otelProvider := log.NewLoggerProvider(otelProviderOptions...)
		otelSlogHandler := otelslog.NewHandler(otelLoggerName, otelslog.WithLoggerProvider(otelProvider))

		return slog.New(slogmulti.Fanout(stdoutSlogHandler, otelSlogHandler)), otelProvider
	}

	return slog.New(stdoutSlogHandler), nil
}
