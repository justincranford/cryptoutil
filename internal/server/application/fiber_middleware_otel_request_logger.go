package application

import (
	"log/slog"
	"time"

	telemetryService "cryptoutil/internal/common/telemetry"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

func otelFiberRequestLoggerMiddleware(telemetryService *telemetryService.TelemetryService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now().UTC()
		err := c.Next()

		// Extract tracing information
		span := trace.SpanFromContext(c.Context())
		spanContext := span.SpanContext()

		// Log request details with OpenTelemetry correlation
		args := []any{
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("duration_ms", time.Since(start)),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.String("trace_id", spanContext.TraceID().String()),
			slog.String("span_id", spanContext.SpanID().String()),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
		}
		telemetryService.Slogger.Info("responded", args...)
		return err
	}
}
