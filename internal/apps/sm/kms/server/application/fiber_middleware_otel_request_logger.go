// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"fmt"
	"log/slog"
	"time"

	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	fiber "github.com/gofiber/fiber/v2"
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
func commonOtelFiberRequestLoggerMiddleware(telemetryService *cryptoutilSharedTelemetry.TelemetryService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// PHASE 1: PRE-REQUEST PROCESSING
		// Capture timing and request details that are available before processing
		start := time.Now().UTC()

		// Extract tracing information (available at request start)
		userContext := c.UserContext()
		span := trace.SpanFromContext(userContext)
		spanContext := span.SpanContext()

		// Get request details (available at request start)
		request := c.Request()
		clientIP := c.IP()
		requestHeaderSize := len(request.Header.Header())
		requestBodySize := len(request.Body())
		httpMethod := c.Method()
		requestPath := c.Path()
		queryString := request.URI().QueryString()
		userAgent := c.Get("User-Agent")

		// PHASE 2: REQUEST PROCESSING
		// Execute the request through subsequent middleware and handlers
		// This populates the response object with status, headers, and body
		err := c.Next()

		// PHASE 3: POST-REQUEST PROCESSING
		// Now response details are available for logging
		response := c.Response()
		statusCode := response.StatusCode()
		responseHeaderSize := len(response.Header.Header())
		responseBodySize := len(response.Body())
		duration := time.Since(start)

		// Build comprehensive logging arguments with all available data
		traceID := spanContext.TraceID().String()
		spanID := spanContext.SpanID().String()
		args := []any{
			// Response details (now available after processing)
			slog.Int("status", statusCode),
			slog.Duration("duration", duration),

			// Request size details
			slog.Int("reqhead", requestHeaderSize),
			slog.Int("reqbody", requestBodySize),

			// Response size details
			slog.Int("resphead", responseHeaderSize),
			slog.Int("respbody", responseBodySize),

			// Request metadata
			slog.String("method", httpMethod),
			slog.String("path", requestPath),
			slog.String("ip", clientIP),
			slog.String("user_agent", userAgent),

			// Tracing correlation
			slog.String("trace_id", traceID),
			slog.String("span_id", spanID),
		}

		// Add request ID always (from requestid middleware)
		// Convert to string representation even if nil or non-string
		requestID := c.Locals("requestid")

		var requestIDStr string

		if requestID != nil {
			if str, ok := requestID.(string); ok {
				requestIDStr = str
			} else {
				// Convert non-string values to string representation
				requestIDStr = fmt.Sprintf("%v", requestID)
			}
		}
		// Always include req_id field, even if empty
		args = append(args, slog.String("req_id", requestIDStr))

		// Add query string if present
		if len(queryString) > 0 {
			args = append(args, slog.String("query", string(queryString)))
		}

		// Add error information if request processing failed
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
		}

		// Log the complete request/response details
		if telemetryService.VerboseMode {
			telemetryService.Slogger.Info("http_request", args...)
		} else {
			telemetryService.Slogger.Debug("http_request", args...)
		}

		// Return any error from request processing to maintain normal flow

		return err //nolint:wrapcheck
	}
}
