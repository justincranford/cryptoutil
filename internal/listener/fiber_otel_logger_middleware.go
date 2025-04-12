package listener

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

func fiberOtelLoggerMiddleware(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		// Extract tracing information
		span := trace.SpanFromContext(c.Context())
		spanContext := span.SpanContext()

		// Log request details with OpenTelemetry correlation
		logger.Info("Responded",
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("duration_ms", time.Since(start)),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.String("trace_id", spanContext.TraceID().String()),
			slog.String("span_id", spanContext.SpanID().String()),
		)
		return err
	}
}
