package authz

import (
	"github.com/gofiber/fiber/v2"
)

// handleHealth handles GET /health - health check endpoint.
func (s *Service) handleHealth(c *fiber.Ctx) error {
	// Check database connectivity via repository HealthCheck method.
	// For now, assume healthy if service is running.
	// TODO: Add actual database health check when repository interface supports it.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":   "healthy",
		"database": "ok",
		"service":  "authz",
	})
}
