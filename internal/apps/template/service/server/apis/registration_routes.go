// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package apis

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
)

// RegisterRegistrationRoutes registers tenant registration endpoints on PUBLIC server.
// These endpoints are unauthenticated - they allow new users to create accounts.
// Call this function from services that want to expose registration APIs.
// requestsPerMin: Rate limit (requests per minute) per IP address (default: 10).
func RegisterRegistrationRoutes(
	app *fiber.App,
	registrationService *cryptoutilTemplateBusinessLogic.TenantRegistrationService,
	requestsPerMin int,
) {
	// Create registration handlers.
	handlers := NewRegistrationHandlers(registrationService)

	// Create rate limiter (10 requests/min per IP, burst 5).
	rateLimiter := NewRateLimiter(requestsPerMin, cryptoutilMagic.RateLimitDefaultBurstSize)

	// Rate limit middleware.
	rateLimitMiddleware := func(c *fiber.Ctx) error {
		ipAddress := c.IP()
		if !rateLimiter.Allow(ipAddress) {
			return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded. Please try again later.",
			})
		}

		return c.Next()
	}

	// User registration endpoints (no authentication required - these create accounts).
	app.Post("/browser/api/v1/auth/register", rateLimitMiddleware, handlers.HandleRegisterUser)
	app.Post("/service/api/v1/auth/register", rateLimitMiddleware, handlers.HandleRegisterUser)
}

// RegisterJoinRequestManagementRoutes registers admin join request management endpoints on ADMIN server.
// These endpoints require admin authentication (TODO: add admin middleware).
// Call this function from services that want to expose admin join-request management APIs.
func RegisterJoinRequestManagementRoutes(
	app *fiber.App,
	registrationService *cryptoutilTemplateBusinessLogic.TenantRegistrationService,
) {
	// Create registration handlers.
	handlers := NewRegistrationHandlers(registrationService)

	// Admin endpoints for managing join requests.
	// CRITICAL: These are on ADMIN server at /admin/api/v1 paths.
	app.Get("/admin/api/v1/join-requests", handlers.HandleListJoinRequests)
	app.Put("/admin/api/v1/join-requests/:id", handlers.HandleProcessJoinRequest)
}
