// Copyright (c) 2025 Justin Cranford
//
//

package idp

import (
	"strings"
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
}

// AuthMiddleware validates session existence for protected endpoints.
func (s *Service) AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		// Extract session cookie.
		sessionID := c.Cookies(s.config.Sessions.CookieName)

		if sessionID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":             cryptoutilIdentityMagic.ErrorAccessDenied,
				"error_description": "Authentication required",
			})
		}

		// Retrieve session from database.
		sessionRepo := s.repoFactory.SessionRepository()

		session, err := sessionRepo.GetBySessionID(ctx, sessionID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":             cryptoutilIdentityMagic.ErrorAccessDenied,
				"error_description": "Invalid or expired session",
			})
		}

		// Validate session is active.
		if !session.Active {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":             cryptoutilIdentityMagic.ErrorAccessDenied,
				"error_description": "Session is no longer active",
			})
		}

		// Validate session not expired.
		if session.IsExpired() {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":             cryptoutilIdentityMagic.ErrorAccessDenied,
				"error_description": "Session has expired",
			})
		}

		// Store session in locals for downstream handlers.
		c.Locals("session", session)

		return c.Next()
	}
}

// TokenAuthMiddleware validates Bearer token for API endpoints.
func (s *Service) TokenAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		// Extract Bearer token from Authorization header.
		authHeader := c.Get("Authorization")

		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":             cryptoutilIdentityMagic.ErrorInvalidToken,
				"error_description": "Missing Authorization header",
			})
		}

		// Parse Bearer token.
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != cryptoutilIdentityMagic.AuthorizationBearer {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":             cryptoutilIdentityMagic.ErrorInvalidToken,
				"error_description": "Invalid Authorization header format",
			})
		}

		accessToken := parts[1]

		// Validate access token and extract claims.
		claims, err := s.tokenSvc.ValidateAccessToken(ctx, accessToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":             cryptoutilIdentityMagic.ErrorInvalidToken,
				"error_description": "Invalid or expired access token",
			})
		}

		// Store claims in locals for downstream handlers.
		c.Locals("claims", claims)

		return c.Next()
	}
}
