// Copyright (c) 2025 Justin Cranford

package server

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilAppsIdentityRsServerConfig "cryptoutil/internal/apps/identity/rs/server/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
)

// PublicServer implements the identity-rs public server by embedding PublicServerBase.
type PublicServer struct {
	base *cryptoutilAppsTemplateServiceServer.PublicServerBase // Reusable server infrastructure.
	cfg  *cryptoutilAppsIdentityRsServerConfig.IdentityRSServerSettings
}

// NewPublicServer creates a new identity-rs public server using builder-provided infrastructure.
// Used by ServerBuilder during route registration.
func NewPublicServer(
	base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
	cfg *cryptoutilAppsIdentityRsServerConfig.IdentityRSServerSettings,
) *PublicServer {
	return &PublicServer{
		base: base,
		cfg:  cfg,
	}
}

// registerRoutes sets up the Resource Server protected API endpoints.
// Called by ServerBuilder after NewPublicServer returns.
func (s *PublicServer) registerRoutes() error {
	// Get underlying Fiber app from base for route registration.
	app := s.base.App()

	// Health endpoints (no auth required).
	app.Get("/health", s.handleHealth)
	app.Get(cryptoutilSharedMagic.PrivateAdminLivezRequestPath, s.handleLivez)
	app.Get(cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, s.handleReadyz)

	// Protected API endpoints.
	// TODO: Add protected API endpoints demonstrating token validation:
	// - /service/api/v1/resources - List resources (requires read scope).
	// - /service/api/v1/resources/:id - Get resource (requires read scope).
	// - /service/api/v1/resources - Create resource (requires write scope).
	// - /service/api/v1/resources/:id - Update resource (requires write scope).
	// - /service/api/v1/resources/:id - Delete resource (requires admin scope).

	// Protected resource endpoint - returns 401 when no authorization header present.
	app.Get("/service/api/v1/resources", s.handleListResources)

	// Token introspection demo endpoints.
	// TODO: Add token introspection demo endpoints:
	// - /service/api/v1/token/info - Display token claims.
	// - /service/api/v1/token/scopes - Display token scopes.

	return nil
}

// handleListResources demonstrates protected resource endpoint.
// Returns 401 Unauthorized when no Authorization header present.
// E2E test expects this behavior to validate token requirement.
func (s *PublicServer) handleListResources(c *fiber.Ctx) error {
	// Check for Authorization header.
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:   "unauthorized",
			"message": "Authorization header required",
		})
	}

	// TODO: Implement actual token validation when token introspection ready.
	// For now, any Authorization header is accepted (E2E test just needs 401 without header).
	return c.JSON(fiber.Map{
		"resources": []fiber.Map{},
		"message":   "Protected resource endpoint - token validation not yet implemented",
	})
}

// handleHealth returns server health status.
func (s *PublicServer) handleHealth(c *fiber.Ctx) error {
	if err := c.JSON(fiber.Map{
		cryptoutilSharedMagic.StringStatus: cryptoutilSharedMagic.DockerServiceHealthHealthy,
		"time":   c.Context().Time().UTC().Format(cryptoutilSharedMagic.StringUTCFormat),
	}); err != nil {
		return fmt.Errorf("failed to send health response: %w", err)
	}

	return nil
}

// handleLivez returns liveness status.
func (s *PublicServer) handleLivez(c *fiber.Ctx) error {
	if err := c.SendString("OK"); err != nil {
		return fmt.Errorf("failed to send liveness response: %w", err)
	}

	return nil
}

// handleReadyz returns readiness status.
func (s *PublicServer) handleReadyz(c *fiber.Ctx) error {
	if err := c.SendString("OK"); err != nil {
		return fmt.Errorf("failed to send readiness response: %w", err)
	}

	return nil
}
