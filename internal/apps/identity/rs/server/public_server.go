// Copyright (c) 2025 Justin Cranford

package server

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"cryptoutil/internal/apps/identity/rs/server/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
)

// PublicServer implements the identity-rs public server by embedding PublicServerBase.
type PublicServer struct {
	base *cryptoutilAppsTemplateServiceServer.PublicServerBase // Reusable server infrastructure.
	cfg  *config.IdentityRSServerSettings
}

// NewPublicServer creates a new identity-rs public server using builder-provided infrastructure.
// Used by ServerBuilder during route registration.
func NewPublicServer(
	base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
	cfg *config.IdentityRSServerSettings,
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
	app.Get("/livez", s.handleLivez)
	app.Get("/readyz", s.handleReadyz)

	// Protected API endpoints.
	// TODO: Add protected API endpoints demonstrating token validation:
	// - /service/api/v1/resources - List resources (requires read scope).
	// - /service/api/v1/resources/:id - Get resource (requires read scope).
	// - /service/api/v1/resources - Create resource (requires write scope).
	// - /service/api/v1/resources/:id - Update resource (requires write scope).
	// - /service/api/v1/resources/:id - Delete resource (requires admin scope).

	// Token introspection demo endpoints.
	// TODO: Add token introspection demo endpoints:
	// - /service/api/v1/token/info - Display token claims.
	// - /service/api/v1/token/scopes - Display token scopes.

	return nil
}

// handleHealth returns server health status.
func (s *PublicServer) handleHealth(c *fiber.Ctx) error {
	if err := c.JSON(fiber.Map{
		"status": "healthy",
		"time":   c.Context().Time().UTC().Format("2006-01-02T15:04:05Z"),
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
