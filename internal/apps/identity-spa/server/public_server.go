// Copyright (c) 2025 Justin Cranford

package server

import (
	"fmt"
	http "net/http"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilAppsIdentitySpaServerConfig "cryptoutil/internal/apps/identity-spa/server/config"
	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps/framework/service/server"
)

// PublicServer implements SPA-specific public endpoints.
// SPA serves static files and handles SPA routing (all routes return index.html).
type PublicServer struct {
	base *cryptoutilAppsFrameworkServiceServer.PublicServerBase
	cfg  *cryptoutilAppsIdentitySpaServerConfig.IdentitySPAServerSettings
}

// NewPublicServer creates a new SPA public server.
func NewPublicServer(base *cryptoutilAppsFrameworkServiceServer.PublicServerBase, cfg *cryptoutilAppsIdentitySpaServerConfig.IdentitySPAServerSettings) *PublicServer {
	return &PublicServer{
		base: base,
		cfg:  cfg,
	}
}

// RegisterRoutes registers all SPA public routes.
func (s *PublicServer) RegisterRoutes() {
	app := s.base.App()

	// Health check endpoints.
	app.Get("/health", s.handleHealth)
	app.Get(cryptoutilSharedMagic.PrivateAdminLivezRequestPath, s.handleLivez)
	app.Get(cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, s.handleReadyz)

	// SPA configuration endpoint (client-side config).
	app.Get("/config.json", s.handleConfig)

	// Static file serving would be configured here.
	// In production, static files are served via CDN or nginx.
	// This endpoint provides a fallback and configuration endpoint.

	// SPA fallback: all other routes return index.html for client-side routing.
	// This is typically handled by a reverse proxy in production.
	app.Get("/*", s.handleSPAFallback)
}

// handleHealth handles /health endpoint.
func (s *PublicServer) handleHealth(c *fiber.Ctx) error {
	c.Status(http.StatusOK)

	if err := c.JSON(fiber.Map{
		cryptoutilSharedMagic.StringStatus: cryptoutilSharedMagic.DockerServiceHealthHealthy,
		"service":                          cryptoutilSharedMagic.OTLPServiceIdentitySPA,
	}); err != nil {
		return fmt.Errorf("failed to send health response: %w", err)
	}

	return nil
}

// handleLivez handles /livez endpoint (liveness probe).
func (s *PublicServer) handleLivez(c *fiber.Ctx) error {
	c.Status(http.StatusOK)

	if err := c.JSON(fiber.Map{
		cryptoutilSharedMagic.StringStatus: "live",
	}); err != nil {
		return fmt.Errorf("failed to send liveness response: %w", err)
	}

	return nil
}

// handleReadyz handles /readyz endpoint (readiness probe).
func (s *PublicServer) handleReadyz(c *fiber.Ctx) error {
	// SPA is always ready if server is running.
	// Static file serving doesn't have external dependencies.
	c.Status(http.StatusOK)

	if err := c.JSON(fiber.Map{
		cryptoutilSharedMagic.StringStatus: "ready",
	}); err != nil {
		return fmt.Errorf("failed to send readiness response: %w", err)
	}

	return nil
}

// handleConfig returns client-side configuration for the SPA.
// This allows the SPA to discover the RP (BFF) endpoint dynamically.
func (s *PublicServer) handleConfig(c *fiber.Ctx) error {
	config := fiber.Map{
		"service": cryptoutilSharedMagic.OTLPServiceIdentitySPA,
		cryptoutilSharedMagic.CLIVersionCommand: cryptoutilSharedMagic.DefaultOTLPVersionDefault,
	}

	// Include RP origin if configured.
	if s.cfg.RPOrigin != "" {
		config["rp_origin"] = s.cfg.RPOrigin
		config["api_base_url"] = s.cfg.RPOrigin + "/api"
	}

	c.Status(http.StatusOK)

	if err := c.JSON(config); err != nil {
		return fmt.Errorf("failed to send config response: %w", err)
	}

	return nil
}

// handleSPAFallback serves index.html for all unmatched routes.
// This enables client-side routing for the Single Page Application.
func (s *PublicServer) handleSPAFallback(c *fiber.Ctx) error {
	// In production, this would serve the actual index.html file.
	// For the reference implementation, we return a placeholder response.
	// The actual static files would be mounted via Fiber's static middleware.

	// Framework-reserved path prefixes must return 404, not the SPA fallback.
	// These prefixes are used by the service template for API and admin routes.
	path := c.Path()
	for _, prefix := range reservedPathPrefixes {
		if strings.HasPrefix(path, prefix) {
			c.Status(http.StatusNotFound)

			if err := c.JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "endpoint not found",
			}); err != nil {
				return fmt.Errorf("failed to send not found response: %w", err)
			}

			return nil
		}
	}

	// For non-reserved routes, return SPA placeholder.
	// In production, use: c.SendFile(s.cfg.StaticFilesPath + "/" + s.cfg.IndexFile)
	c.Status(http.StatusOK)

	if err := c.JSON(fiber.Map{
		"message":     "SPA placeholder - static files not configured",
		"index_file":  s.cfg.IndexFile,
		"static_path": s.cfg.StaticFilesPath,
		"hint":        "Configure static file serving or use CDN/reverse proxy in production",
	}); err != nil {
		return fmt.Errorf("failed to send SPA fallback response: %w", err)
	}

	return nil
}

// reservedPathPrefixes are framework-reserved path prefixes that must not be
// caught by the SPA fallback handler. These return 404 for unregistered routes.
var reservedPathPrefixes = []string{"/admin/", "/service/", "/browser/", "/api/"}
