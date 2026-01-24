// Copyright (c) 2025 Justin Cranford

package server

import (
	"context"
	"crypto/tls"
	"fmt"
	http "net/http"
	"time"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilAppsIdentityRpServerConfig "cryptoutil/internal/apps/identity/rp/server/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerService "cryptoutil/internal/apps/template/service/server/service"
)

// AuthZ server check timeout.
const authzCheckTimeout = 5 * time.Second

// PublicServer implements the identity-rp public server by embedding PublicServerBase.
type PublicServer struct {
	base *cryptoutilAppsTemplateServiceServer.PublicServerBase // Reusable server infrastructure.

	cfg            *cryptoutilAppsIdentityRpServerConfig.IdentityRPServerSettings
	sessionManager *cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService
	realmService   cryptoutilAppsTemplateServiceServerService.RealmService
}

// NewPublicServer creates a new identity-rp public server using builder-provided infrastructure.
// Used by ServerBuilder during route registration.
func NewPublicServer(
	base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
	cfg *cryptoutilAppsIdentityRpServerConfig.IdentityRPServerSettings,
	sessionManager *cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService,
	realmService cryptoutilAppsTemplateServiceServerService.RealmService,
) *PublicServer {
	return &PublicServer{
		base:           base,
		cfg:            cfg,
		sessionManager: sessionManager,
		realmService:   realmService,
	}
}

// registerRoutes sets up the RP BFF endpoints.
// Called by ServerBuilder after NewPublicServer returns.
func (s *PublicServer) registerRoutes() error {
	// Get underlying Fiber app from base for route registration.
	app := s.base.App()

	// Health endpoints (no auth required).
	app.Get("/health", s.handleHealth)
	app.Get("/livez", s.handleLivez)
	app.Get("/readyz", s.handleReadyz)

	// BFF OAuth 2.1 proxy endpoints.
	// TODO: Add OAuth 2.1 authorization code flow endpoints:
	// - /service/api/v1/auth/login - Initiates authorization code flow.
	// - /service/api/v1/auth/callback - Handles authorization code callback.
	// - /service/api/v1/auth/logout - Handles logout (revoke tokens, clear session).
	// - /service/api/v1/auth/userinfo - Proxies userinfo endpoint with session token.
	// - /service/api/v1/auth/refresh - Handles token refresh using refresh token from session.

	// BFF browser endpoints (with CORS, CSRF protection).
	// TODO: Add browser-specific endpoints:
	// - /browser/api/v1/auth/login - Same as service but with CORS/CSRF.
	// - /browser/api/v1/auth/callback - Same as service but with CORS/CSRF.
	// - /browser/api/v1/auth/logout - Same as service but with CORS/CSRF.
	// - /browser/api/v1/auth/userinfo - Same as service but with CORS/CSRF.
	// - /browser/api/v1/auth/refresh - Same as service but with CORS/CSRF.

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
// For RP, this includes checking OAuth 2.1 provider (AuthZ server) connectivity.
func (s *PublicServer) handleReadyz(c *fiber.Ctx) error {
	// Check OAuth 2.1 provider (AuthZ server) connectivity.
	if s.cfg.AuthzServerURL != "" {
		if err := s.checkAuthZServer(); err != nil {
			c.Status(fiber.StatusServiceUnavailable)

			if jsonErr := c.JSON(fiber.Map{
				"status": "not ready",
				"reason": "authz server unavailable",
				"error":  err.Error(),
			}); jsonErr != nil {
				return fmt.Errorf("failed to send not ready response: %w", jsonErr)
			}

			return nil
		}
	}

	if err := c.SendString("OK"); err != nil {
		return fmt.Errorf("failed to send readiness response: %w", err)
	}

	return nil
}

// checkAuthZServer verifies OAuth 2.1 provider connectivity.
// This checks the /livez endpoint on the authorization server.
func (s *PublicServer) checkAuthZServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), authzCheckTimeout)
	defer cancel()

	// Create HTTP client with TLS (skip verify for dev mode).
	client := &http.Client{
		Timeout: authzCheckTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: s.cfg.DevMode, //nolint:gosec // Dev mode only
			},
		},
	}

	// Build livez URL.
	livezURL := s.cfg.AuthzServerURL + "/livez"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, livezURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to authz server: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authz server returned status %d", resp.StatusCode)
	}

	return nil
}
