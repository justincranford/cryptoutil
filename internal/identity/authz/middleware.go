// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// RegisterMiddleware registers all middleware for the AuthZ server.
func (s *Service) RegisterMiddleware(app *fiber.App) {
	// Recover from panics.
	app.Use(recover.New())

	// Logging middleware.
	app.Use(logger.New(logger.Config{
		Format: "${time} ${status} ${method} ${path} ${latency}\n",
	}))

	// CORS middleware - skip for OAuth endpoints (machine-to-machine).
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		Next: func(c *fiber.Ctx) bool {
			// Skip CORS for OAuth 2.1 endpoints (machine-to-machine, not browser-based).
			url := c.OriginalURL()
			return strings.HasPrefix(url, "/oauth2/v1/") || strings.HasPrefix(url, "/openid/v1/")
		},
	}))

	// CSRF middleware - skip for OAuth endpoints (machine-to-machine).
	app.Use(csrf.New(csrf.Config{
		Next: func(c *fiber.Ctx) bool {
			// Skip CSRF for OAuth 2.1 endpoints (machine-to-machine, never browser-based).
			url := c.OriginalURL()
			return strings.HasPrefix(url, "/oauth2/v1/") || strings.HasPrefix(url, "/openid/v1/")
		},
	}))

	// Rate limiting middleware.
	app.Use(limiter.New(limiter.Config{
		Max:        cryptoutilIdentityMagic.DefaultRateLimitRequests,
		Expiration: cryptoutilIdentityMagic.DefaultRateLimitWindow,
	}))
}
