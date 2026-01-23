// Copyright (c) 2025 Justin Cranford

package server

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"cryptoutil/internal/apps/identity/authz/server/config"
	cryptoutilTemplateServer "cryptoutil/internal/apps/template/service/server"
)

// PublicServer implements the identity-authz public server by embedding PublicServerBase.
type PublicServer struct {
	base *cryptoutilTemplateServer.PublicServerBase // Reusable server infrastructure.
	cfg  *config.IdentityAuthzServerSettings
}

// NewPublicServer creates a new identity-authz public server using builder-provided infrastructure.
// Used by ServerBuilder during route registration.
func NewPublicServer(
	base *cryptoutilTemplateServer.PublicServerBase,
	cfg *config.IdentityAuthzServerSettings,
) *PublicServer {
	return &PublicServer{
		base: base,
		cfg:  cfg,
	}
}

// registerRoutes sets up the OAuth 2.1 authorization server endpoints.
// Called by ServerBuilder after NewPublicServer returns.
func (s *PublicServer) registerRoutes() error {
	// Get underlying Fiber app from base for route registration.
	app := s.base.App()

	// Health endpoints (no auth required).
	app.Get("/health", s.handleHealth)
	app.Get("/livez", s.handleLivez)
	app.Get("/readyz", s.handleReadyz)

	// OIDC Discovery endpoints.
	if s.cfg.EnableDiscovery {
		app.Get("/.well-known/openid-configuration", s.handleOpenIDConfiguration)
		app.Get("/.well-known/jwks.json", s.handleJWKS)
	}

	// OAuth 2.1 Authorization Server endpoints.
	// TODO: Add OAuth 2.1 endpoints:
	// - /service/api/v1/oauth/authorize - Authorization endpoint.
	// - /service/api/v1/oauth/token - Token endpoint.
	// - /service/api/v1/oauth/revoke - Token revocation endpoint.
	// - /service/api/v1/oauth/introspect - Token introspection endpoint.
	// - /service/api/v1/userinfo - UserInfo endpoint.

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

// handleOpenIDConfiguration returns OIDC Discovery document.
func (s *PublicServer) handleOpenIDConfiguration(c *fiber.Ctx) error {
	// Build discovery document based on issuer.
	discovery := fiber.Map{
		"issuer":                 s.cfg.Issuer,
		"authorization_endpoint": s.cfg.Issuer + "/service/api/v1/oauth/authorize",
		"token_endpoint":         s.cfg.Issuer + "/service/api/v1/oauth/token",
		"userinfo_endpoint":      s.cfg.Issuer + "/service/api/v1/userinfo",
		"jwks_uri":               s.cfg.Issuer + "/.well-known/jwks.json",
		"revocation_endpoint":    s.cfg.Issuer + "/service/api/v1/oauth/revoke",
		"introspection_endpoint": s.cfg.Issuer + "/service/api/v1/oauth/introspect",
		"response_types_supported": []string{
			"code",
			"token",
			"id_token",
			"code token",
			"code id_token",
			"token id_token",
			"code token id_token",
		},
		"grant_types_supported": []string{
			"authorization_code",
			"refresh_token",
			"client_credentials",
		},
		"subject_types_supported": []string{"public"},
		"id_token_signing_alg_values_supported": []string{
			"RS256",
			"RS384",
			"RS512",
			"ES256",
			"ES384",
			"ES512",
		},
		"scopes_supported": []string{
			"openid",
			"profile",
			"email",
			"offline_access",
		},
		"token_endpoint_auth_methods_supported": []string{
			"client_secret_basic",
			"client_secret_post",
			"private_key_jwt",
		},
		"claims_supported": []string{
			"sub",
			"iss",
			"aud",
			"exp",
			"iat",
			"auth_time",
			"name",
			"email",
			"email_verified",
		},
	}

	if err := c.JSON(discovery); err != nil {
		return fmt.Errorf("failed to send discovery response: %w", err)
	}

	return nil
}

// handleJWKS returns the public JWKS for token verification.
func (s *PublicServer) handleJWKS(c *fiber.Ctx) error {
	// TODO: Return actual public JWKS from JWK generation service.
	// For now, return empty JWKS.
	jwks := fiber.Map{
		"keys": []fiber.Map{},
	}

	if err := c.JSON(jwks); err != nil {
		return fmt.Errorf("failed to send JWKS response: %w", err)
	}

	return nil
}
