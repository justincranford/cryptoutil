// Copyright (c) 2025 Justin Cranford

package server

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"

	cryptoutilTemplateServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilSPAConfig "cryptoutil/internal/apps/identity/spa/server/config"
)

// PublicServer implements SPA-specific public endpoints.
// SPA serves static files and handles SPA routing (all routes return index.html).
type PublicServer struct {
	base *cryptoutilTemplateServer.PublicServerBase
	cfg  *cryptoutilSPAConfig.IdentitySPAServerSettings
}

// NewPublicServer creates a new SPA public server.
func NewPublicServer(base *cryptoutilTemplateServer.PublicServerBase, cfg *cryptoutilSPAConfig.IdentitySPAServerSettings) *PublicServer {
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
	app.Get("/livez", s.handleLivez)
	app.Get("/readyz", s.handleReadyz)

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
		"status":  "healthy",
		"service": "identity-spa",
	}); err != nil {
		return fmt.Errorf("failed to send health response: %w", err)
	}

	return nil
}

// handleLivez handles /livez endpoint (liveness probe).
func (s *PublicServer) handleLivez(c *fiber.Ctx) error {
	c.Status(http.StatusOK)

	if err := c.JSON(fiber.Map{
		"status": "live",
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
		"status": "ready",
	}); err != nil {
		return fmt.Errorf("failed to send readiness response: %w", err)
	}

	return nil
}

// handleConfig returns client-side configuration for the SPA.
// This allows the SPA to discover the RP (BFF) endpoint dynamically.
func (s *PublicServer) handleConfig(c *fiber.Ctx) error {
	config := fiber.Map{
		"service": "identity-spa",
		"version": "0.0.1",
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

	// Check if this is an API request (should return 404, not SPA).
	path := c.Path()
	if len(path) >= apiPrefixLen && path[:apiPrefixLen] == "/api/" {
		c.Status(http.StatusNotFound)

		if err := c.JSON(fiber.Map{
			"error": "API endpoint not found",
		}); err != nil {
			return fmt.Errorf("failed to send API not found response: %w", err)
		}

		return nil
	}

	// For non-API routes, return SPA placeholder.
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

// API path prefix length for checking API routes.
const apiPrefixLen = 5 // len("/api/")
