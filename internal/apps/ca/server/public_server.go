// Copyright (c) 2025 Justin Cranford

package server

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	cryptoutilCAServer "cryptoutil/api/ca/server"
	"cryptoutil/internal/apps/ca/server/config"
	cryptoutilTemplateServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilCAHandler "cryptoutil/internal/ca/api/handler"
	cryptoutilCAServiceRevocation "cryptoutil/internal/ca/service/revocation"
)

// PublicServer implements the pki-ca public server by embedding PublicServerBase.
type PublicServer struct {
	base *cryptoutilTemplateServer.PublicServerBase // Reusable server infrastructure

	handler     *cryptoutilCAHandler.Handler
	crlService  *cryptoutilCAServiceRevocation.CRLService
	ocspService *cryptoutilCAServiceRevocation.OCSPService
	config      *config.CAServerSettings
}

// NewPublicServer creates a new pki-ca public server using builder-provided infrastructure.
// Used by ServerBuilder during route registration.
func NewPublicServer(
	base *cryptoutilTemplateServer.PublicServerBase,
	handler *cryptoutilCAHandler.Handler,
	crlService *cryptoutilCAServiceRevocation.CRLService,
	ocspService *cryptoutilCAServiceRevocation.OCSPService,
	cfg *config.CAServerSettings,
) *PublicServer {
	return &PublicServer{
		base:        base,
		handler:     handler,
		crlService:  crlService,
		ocspService: ocspService,
		config:      cfg,
	}
}

// registerRoutes sets up the CA API endpoints.
// Called by ServerBuilder after NewPublicServer returns.
func (s *PublicServer) registerRoutes() error {
	// Get underlying Fiber app from base for route registration.
	app := s.base.App()

	// Health endpoints (no auth required).
	app.Get("/health", s.handleHealth)
	app.Get("/livez", s.handleLivez)
	app.Get("/readyz", s.handleReadyz)

	// Register CA API handlers using oapi-codegen generated code.
	// Routes registered at /service/api/v1/ca/* path prefix.
	cryptoutilCAServer.RegisterHandlersWithOptions(app, s.handler, cryptoutilCAServer.FiberServerOptions{
		BaseURL: "/service/api/v1/ca",
	})

	// Register browser paths with same handlers.
	cryptoutilCAServer.RegisterHandlersWithOptions(app, s.handler, cryptoutilCAServer.FiberServerOptions{
		BaseURL: "/browser/api/v1/ca",
	})

	// CRL distribution point (typically public, no auth).
	if s.config.EnableCRL {
		app.Get("/service/api/v1/crl", s.handleCRLDistribution)
		app.Get("/browser/api/v1/crl", s.handleCRLDistribution)
		app.Get("/.well-known/pki-ca/crl", s.handleCRLDistribution)
	}

	// OCSP responder endpoint (typically public, no auth).
	if s.config.EnableOCSP {
		app.Post("/service/api/v1/ocsp", s.handleOCSP)
		app.Post("/browser/api/v1/ocsp", s.handleOCSP)
		app.Post("/.well-known/pki-ca/ocsp", s.handleOCSP)
	}

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

// handleCRLDistribution returns the current CRL.
func (s *PublicServer) handleCRLDistribution(c *fiber.Ctx) error {
	crl, err := s.crlService.GenerateCRL()
	if err != nil {
		if jsonErr := c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate CRL",
		}); jsonErr != nil {
			return fmt.Errorf("failed to send error response: %w", jsonErr)
		}

		return nil
	}

	c.Set("Content-Type", "application/pkix-crl")

	if sendErr := c.Send(crl); sendErr != nil {
		return fmt.Errorf("failed to send CRL: %w", sendErr)
	}

	return nil
}

// handleOCSP handles OCSP requests.
// Note: This is a simplified implementation. A full implementation would look up
// the certificate by serial number from storage.
func (s *PublicServer) handleOCSP(c *fiber.Ctx) error {
	body := c.Body()

	// Parse the OCSP request first.
	_, err := s.ocspService.ParseRequest(body)
	if err != nil {
		if jsonErr := c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to parse OCSP request",
		}); jsonErr != nil {
			return fmt.Errorf("failed to send error response: %w", jsonErr)
		}

		return nil
	}

	// For now, respond with a basic response without certificate lookup.
	// A full implementation would look up the certificate and call RespondToRequest.
	// The existing CA API handler has more complete OCSP handling.
	if jsonErr := c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "OCSP endpoint uses /api/v1/ca/ocsp handler",
	}); jsonErr != nil {
		return fmt.Errorf("failed to send response: %w", jsonErr)
	}

	return nil
}
