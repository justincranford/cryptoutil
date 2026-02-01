// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package apis

import (
	http "net/http"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerMiddleware "cryptoutil/internal/apps/template/service/server/middleware"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// RegisterRegistrationRoutes registers tenant registration endpoints on PUBLIC server.
// These endpoints are unauthenticated - they allow new users to create accounts.
// Call this function from services that want to expose registration APIs.
// requestsPerMin: Rate limit (requests per minute) per IP address (default: 10).
func RegisterRegistrationRoutes(
	app *fiber.App,
	registrationService *cryptoutilAppsTemplateServiceServerBusinesslogic.TenantRegistrationService,
	requestsPerMin int,
) {
	// Create registration handlers.
	handlers := NewRegistrationHandlers(registrationService)

	// Create rate limiter (10 requests/min per IP, burst 5).
	rateLimiter := NewRateLimiter(requestsPerMin, cryptoutilSharedMagic.RateLimitDefaultBurstSize)

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
// These endpoints require admin authentication.
// Call this function from services that want to expose admin join-request management APIs.
//
// Parameters:
//   - app: Fiber app instance (typically admin server)
//   - registrationService: Service handling tenant registration logic
//   - sessionValidator: Session validation service for authentication
func RegisterJoinRequestManagementRoutes(
	app *fiber.App,
	registrationService *cryptoutilAppsTemplateServiceServerBusinesslogic.TenantRegistrationService,
	sessionValidator cryptoutilAppsTemplateServiceServerMiddleware.SessionValidator,
) {
	// Create registration handlers.
	handlers := NewRegistrationHandlers(registrationService)

	// Create admin authentication middleware.
	// Validates session token and ensures user is authenticated.
	// TODO: Add role-based authorization when role management system implemented.
	// For now, this only validates authentication (not admin role).
	adminAuthMiddleware := cryptoutilAppsTemplateServiceServerMiddleware.BrowserSessionMiddleware(sessionValidator)

	// Admin endpoints for managing join requests.
	// CRITICAL: These are on ADMIN server at /admin/api/v1 paths.
	// Protected by session authentication middleware.
	app.Get("/admin/api/v1/join-requests", adminAuthMiddleware, handlers.HandleListJoinRequests)
	app.Put("/admin/api/v1/join-requests/:id", adminAuthMiddleware, handlers.HandleProcessJoinRequest)
}
