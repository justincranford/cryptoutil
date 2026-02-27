// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	fiber "github.com/gofiber/fiber/v2"
)

// handleHealth handles GET /health - health check endpoint.
func (s *Service) handleHealth(c *fiber.Ctx) error {
	ctx := c.Context()

	// Check database connectivity via Ping().
	db := s.repoFactory.DB()

	sqlDB, err := db.DB()
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			cryptoutilSharedMagic.StringStatus:             "unhealthy",
			cryptoutilSharedMagic.RealmStorageTypeDatabase: "unavailable",
			"service":                         cryptoutilSharedMagic.AuthzServiceName,
			cryptoutilSharedMagic.StringError: err.Error(),
		})
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			cryptoutilSharedMagic.StringStatus:             "unhealthy",
			cryptoutilSharedMagic.RealmStorageTypeDatabase: "unreachable",
			"service":                         cryptoutilSharedMagic.AuthzServiceName,
			cryptoutilSharedMagic.StringError: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		cryptoutilSharedMagic.StringStatus:             cryptoutilSharedMagic.DockerServiceHealthHealthy,
		cryptoutilSharedMagic.RealmStorageTypeDatabase: "ok",
		"service": cryptoutilSharedMagic.AuthzServiceName,
	})
}
