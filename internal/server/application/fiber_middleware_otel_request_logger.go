package application

import (
	"log/slog"
	"time"

	telemetryService "cryptoutil/internal/common/telemetry"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

// commonOtelFiberRequestLoggerMiddleware provides structured HTTP request logging with OpenTelemetry correlation.
//
// Message Format Comparison: logger.New() vs commonOtelFiberRequestLoggerMiddleware
//
// OLD FORMAT (logger.New()):
//   - Format: "timestamp [level] method path status duration - ip user_agent"
//   - Example: "2023/10/17 14:30:15 [INFO] GET /api/users 200 1.234ms - 127.0.0.1 Mozilla/5.0..."
//   - Fields: timestamp, level, method, path, status, duration, ip, user_agent
//   - Limitations: No trace/span correlation, no request ID, no query params, no error details
//
// NEW FORMAT (commonOtelFiberRequestLoggerMiddleware):
//   - Format: Structured slog with OpenTelemetry correlation
//   - Fields: status, duration, reqhead, reqbody, resphead, respbody, method, path, ip, user_agent, trace_id, span_id, req_id?, query?, error?
//   - Example: "status=200 duration=1.234ms reqhead=256 reqbody=512 resphead=128 respbody=1024 method=GET path=/api/users ip=127.0.0.1 user_agent=Mozilla/5.0... trace_id=abc123 span_id=def456 req_id=req789"
//
// Request Types:
//
// 1. SUCCESS REQUESTS (status 200-299):
//   - Old: "2023/10/17 14:30:15 [INFO] GET /api/users 200 1.234ms - 127.0.0.1 Mozilla/5.0..."
//   - New: "status=200 duration=1.234ms reqhead=256 reqbody=512 resphead=128 respbody=1024 method=GET path=/api/users ip=127.0.0.1 user_agent=Mozilla/5.0... trace_id=abc123 span_id=def456 req_id=req789"
//
// 2. FAILED REQUESTS (status 400-599):
//   - Old: "2023/10/17 14:30:15 [INFO] POST /api/users 400 2.345ms - 127.0.0.1 Mozilla/5.0..."
//   - New: "status=400 duration=2.345ms reqhead=512 reqbody=256 resphead=64 respbody=128 method=POST path=/api/users ip=127.0.0.1 user_agent=Mozilla/5.0... trace_id=abc123 span_id=def456 req_id=req789 error=validation failed"
//
// 3. NO RESPONSE REQUESTS (errors preventing proper response):
//   - Old: No log entry (request fails before response)
//   - New: "status=200 duration=854.6µs reqhead=128 reqbody=0 resphead=64 respbody=0 method=GET path=/nonexistent ip=127.0.0.1 user_agent=Mozilla/5.0... trace_id=abc123 span_id=def456 req_id=req789 error='no matching operation was found'"
//   - Note: Status may be 200 (initial/default) when middleware executes, but error field captures the actual issue
func commonOtelFiberRequestLoggerMiddleware(telemetryService *telemetryService.TelemetryService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now().UTC()
		err := c.Next()

		userContext := c.UserContext()
		span := trace.SpanFromContext(userContext)
		spanContext := span.SpanContext()

		request := c.Request()
		response := c.Response()

		// Log comprehensive request details with OpenTelemetry correlation
		args := []any{
			slog.Int("status", response.StatusCode()),
			slog.Duration("duration", time.Since(start)),
			slog.Int("reqhead", len(request.Header.Header())),
			slog.Int("reqbody", len(request.Body())),
			slog.Int("resphead", len(response.Header.Header())),
			slog.Int("respbody", len(response.Body())),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.String("ip", c.IP()),
			slog.String("user_agent", c.Get("User-Agent")),
			slog.String("trace_id", spanContext.TraceID().String()),
			slog.String("span_id", spanContext.SpanID().String()),
		}

		// Add request ID if available (from requestid middleware)
		if requestID := c.Locals("requestid"); requestID != nil {
			if requestIDStr, ok := requestID.(string); ok {
				args = append(args, slog.String("req_id", requestIDStr))
			}
		}

		// Add query string if present
		if query := request.URI().QueryString(); len(query) > 0 {
			args = append(args, slog.String("query", string(query)))
		}

		// Add error information if present
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
		}

		telemetryService.Slogger.Info("http_request", args...)

		return err //nolint:wrapcheck
	}
}
