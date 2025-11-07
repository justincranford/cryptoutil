package idp

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// RegisterMiddleware sets up Fiber middleware for the IdP server.
func (s *Service) RegisterMiddleware(app *fiber.App) {
	// Recover from panics.
	app.Use(recover.New())

	// Structured logging.
	app.Use(logger.New(logger.Config{
		Format:     "${time} ${method} ${path} - ${status} - ${latency}\n",
		TimeFormat: time.RFC3339,
		TimeZone:   "UTC",
	}))

	// CORS configuration.
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Rate limiting.
	app.Use(limiter.New(limiter.Config{
		Max:        cryptoutilIdentityMagic.RateLimitRequestsPerWindow,
		Expiration: time.Duration(cryptoutilIdentityMagic.RateLimitWindowSeconds) * time.Second,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":             "rate_limit_exceeded",
				"error_description": "Too many requests",
			})
		},
	}))
	// TODO: Add authentication middleware for protected endpoints (/userinfo, /logout).
	// TODO: Add session validation middleware.
}
