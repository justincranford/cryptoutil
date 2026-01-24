// Copyright (c) 2025 Justin Cranford
//

// Package middleware provides HTTP middleware for the JOSE server.
package middleware

import (
	"time"

	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

const (
	// DefaultRateLimit is the default maximum requests per second per IP.
	DefaultRateLimit = 100

	// DefaultRateLimitExpiration is the default rate limit window.
	DefaultRateLimitExpiration = time.Second
)

// RateLimitConfig holds configuration for rate limiting middleware.
type RateLimitConfig struct {
	// Max is the maximum number of requests allowed per expiration period.
	Max int

	// Expiration is the time window for rate limiting.
	Expiration time.Duration

	// TelemetryService for logging rate limit violations.
	TelemetryService *cryptoutilSharedTelemetry.TelemetryService
}

// NewRateLimiter creates a new rate limiting middleware.
// Returns HTTP 429 Too Many Requests when limit is exceeded.
func NewRateLimiter(cfg *RateLimitConfig) fiber.Handler {
	maxRequests := DefaultRateLimit
	expiration := DefaultRateLimitExpiration

	if cfg != nil {
		if cfg.Max > 0 {
			maxRequests = cfg.Max
		}

		if cfg.Expiration > 0 {
			expiration = cfg.Expiration
		}
	}

	return limiter.New(limiter.Config{
		Max:        maxRequests,
		Expiration: expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Throttle by IP address.
			// Can be improved in the future to include tenant ID or user ID.
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			if cfg != nil && cfg.TelemetryService != nil {
				cfg.TelemetryService.Slogger.Warn("Rate limit exceeded",
					"requestid", c.Locals("requestid"),
					"method", c.Method(),
					"IP", c.IP(),
					"URL", c.OriginalURL(),
				)
			}

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
			})
		},
	})
}
