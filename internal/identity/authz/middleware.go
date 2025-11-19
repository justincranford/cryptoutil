// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

	// CORS middleware.
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Rate limiting middleware.
	app.Use(limiter.New(limiter.Config{
		Max:        cryptoutilIdentityMagic.DefaultRateLimitRequests,
		Expiration: cryptoutilIdentityMagic.DefaultRateLimitWindow,
	}))
}
