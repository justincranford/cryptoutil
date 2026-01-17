// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package apis

import (
	"github.com/gofiber/fiber/v2"

	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
)

// RegisterRegistrationRoutes registers tenant registration endpoints.
// Call this function from services that want to expose registration/join-request APIs.
func RegisterRegistrationRoutes(
	app *fiber.App,
	registrationService *cryptoutilTemplateBusinessLogic.TenantRegistrationService,
) {
	// Create registration handlers.
	handlers := NewRegistrationHandlers(registrationService)

	// User registration endpoints (no authentication required - these create accounts).
	app.Post("/browser/api/v1/auth/register", handlers.HandleRegisterUser)
	app.Post("/service/api/v1/auth/register", handlers.HandleRegisterUser)

	// Admin endpoints for managing join requests (TODO: add admin middleware).
	app.Get("/browser/api/v1/admin/join-requests", handlers.HandleListJoinRequests)
	app.Post("/browser/api/v1/admin/join-requests/:id/approve", handlers.HandleApproveJoinRequest)
	app.Post("/browser/api/v1/admin/join-requests/:id/reject", handlers.HandleRejectJoinRequest)

	// Service endpoints for managing join requests (TODO: add service auth middleware).
	app.Get("/service/api/v1/admin/join-requests", handlers.HandleListJoinRequests)
	app.Post("/service/api/v1/admin/join-requests/:id/approve", handlers.HandleApproveJoinRequest)
	app.Post("/service/api/v1/admin/join-requests/:id/reject", handlers.HandleRejectJoinRequest)
}
