// Copyright (c) 2025 Justin Cranford

package server

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"cryptoutil/internal/apps/identity/idp/server/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
)

// PublicServer implements the identity-idp public server by embedding PublicServerBase.
type PublicServer struct {
	base *cryptoutilAppsTemplateServiceServer.PublicServerBase // Reusable server infrastructure.
	cfg  *config.IdentityIDPServerSettings
}

// NewPublicServer creates a new identity-idp public server using builder-provided infrastructure.
// Used by ServerBuilder during route registration.
func NewPublicServer(
	base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
	cfg *config.IdentityIDPServerSettings,
) *PublicServer {
	return &PublicServer{
		base: base,
		cfg:  cfg,
	}
}

// registerRoutes sets up the IdP authentication endpoints.
// Called by ServerBuilder after NewPublicServer returns.
func (s *PublicServer) registerRoutes() error {
	// Get underlying Fiber app from base for route registration.
	app := s.base.App()

	// Health endpoints (no auth required).
	app.Get("/health", s.handleHealth)
	app.Get("/livez", s.handleLivez)
	app.Get("/readyz", s.handleReadyz)

	// IdP browser endpoints (login/consent UI).
	// TODO: Add IdP endpoints:
	// - /browser/login - Login page (username/password form).
	// - /browser/consent - Consent page (scope approval).
	// - /browser/logout - Logout page.
	// - /browser/mfa/enroll - MFA enrollment page.
	// - /browser/mfa/verify - MFA verification page.

	// IdP API endpoints.
	// TODO: Add IdP API endpoints:
	// - /service/api/v1/auth/login - Login submission endpoint.
	// - /service/api/v1/auth/consent - Consent submission endpoint.
	// - /service/api/v1/auth/logout - Logout endpoint.
	// - /service/api/v1/mfa/enroll - MFA enrollment endpoint.
	// - /service/api/v1/mfa/verify - MFA verification endpoint.

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
