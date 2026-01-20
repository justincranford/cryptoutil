// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package apis

import (
	"github.com/gofiber/fiber/v2"

	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
)

// RegisterRegistrationRoutes registers tenant registration endpoints on PUBLIC server.
// These endpoints are unauthenticated - they allow new users to create accounts.
// Call this function from services that want to expose registration APIs.
func RegisterRegistrationRoutes(
	app *fiber.App,
	registrationService *cryptoutilTemplateBusinessLogic.TenantRegistrationService,
) {
	// Create registration handlers.
	handlers := NewRegistrationHandlers(registrationService)

	// User registration endpoints (no authentication required - these create accounts).
	app.Post("/browser/api/v1/auth/register", handlers.HandleRegisterUser)
	app.Post("/service/api/v1/auth/register", handlers.HandleRegisterUser)
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
