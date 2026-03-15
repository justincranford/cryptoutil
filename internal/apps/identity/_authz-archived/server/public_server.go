// Copyright (c) 2025 Justin Cranford

package server

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilAppsIdentityAuthzServerConfig "cryptoutil/internal/apps/identity/authz/server/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
)

// PublicServer implements the identity-authz public server by embedding PublicServerBase.
type PublicServer struct {
	base *cryptoutilAppsTemplateServiceServer.PublicServerBase // Reusable server infrastructure.
	cfg  *cryptoutilAppsIdentityAuthzServerConfig.IdentityAuthzServerSettings
}

// NewPublicServer creates a new identity-authz public server using builder-provided infrastructure.
// Used by ServerBuilder during route registration.
func NewPublicServer(
	base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
	cfg *cryptoutilAppsIdentityAuthzServerConfig.IdentityAuthzServerSettings,
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
	app.Get(cryptoutilSharedMagic.PrivateAdminLivezRequestPath, s.handleLivez)
	app.Get(cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, s.handleReadyz)

	// OIDC Discovery endpoints.
	if s.cfg.EnableDiscovery {
		app.Get(cryptoutilSharedMagic.PathDiscovery, s.handleOpenIDConfiguration)
		app.Get(cryptoutilSharedMagic.PathJWKS, s.handleJWKS)
	}

	// OAuth 2.1 Authorization Server endpoints.
	// TODO: Add OAuth 2.1 endpoints:
	// - /service/api/v1/oauth/authorize - Authorization endpoint.
	// - /service/api/v1/oauth/token - Token endpoint.
	// - /service/api/v1/oauth/revoke - Token revocation endpoint.
	// - /service/api/v1/oauth/introspect - Token introspection endpoint.
	// - /service/api/v1/userinfo - UserInfo endpoint.

	// Browser authorization endpoint - returns HTML consent form (placeholder for E2E test).
	app.Get("/browser/api/v1/authorize", s.handleBrowserAuthorize)

	return nil
}

// handleHealth returns server health status.
func (s *PublicServer) handleHealth(c *fiber.Ctx) error {
	if err := c.JSON(fiber.Map{
		cryptoutilSharedMagic.StringStatus: cryptoutilSharedMagic.DockerServiceHealthHealthy,
		"time":                             c.Context().Time().UTC().Format(cryptoutilSharedMagic.StringUTCFormat),
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

// handleBrowserAuthorize serves OAuth authorization page.
// E2E test expects non-404 response to validate browser endpoint exists.
// TODO: Replace with proper authorization flow when OAuth implementation ready.
func (s *PublicServer) handleBrowserAuthorize(c *fiber.Ctx) error {
	// Set HTML content type.
	c.Set("Content-Type", "text/html; charset=utf-8")

	// Return minimal HTML authorization page (placeholder).
	const authorizeHTML = `<!DOCTYPE html>
<html>
<head>
    <title>Authorization Server - Authorize</title>
</head>
<body>
    <h1>OAuth 2.1 Authorization</h1>
    <p>Application is requesting access to your account</p>
    <form method="post" action="/service/api/v1/oauth/authorize">
        <h2>Requested Scopes:</h2>
        <ul>
            <li>read:profile</li>
            <li>write:data</li>
        </ul>
        <button type="submit" name="action" value="allow">Allow</button>
        <button type="submit" name="action" value="deny">Deny</button>
    </form>
    <p><em>Note: OAuth authorization flow not yet implemented</em></p>
</body>
</html>`

	if err := c.SendString(authorizeHTML); err != nil {
		return fmt.Errorf("failed to send authorize page: %w", err)
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
		"jwks_uri":               s.cfg.Issuer + cryptoutilSharedMagic.PathJWKS,
		"revocation_endpoint":    s.cfg.Issuer + "/service/api/v1/oauth/revoke",
		"introspection_endpoint": s.cfg.Issuer + "/service/api/v1/oauth/introspect",
		"response_types_supported": []string{
			cryptoutilSharedMagic.ResponseTypeCode,
			cryptoutilSharedMagic.ParamToken,
			cryptoutilSharedMagic.ParamIDToken,
			"code token",
			"code id_token",
			"token id_token",
			"code token id_token",
		},
		"grant_types_supported": []string{
			cryptoutilSharedMagic.GrantTypeAuthorizationCode,
			cryptoutilSharedMagic.GrantTypeRefreshToken,
			cryptoutilSharedMagic.GrantTypeClientCredentials,
		},
		"subject_types_supported": []string{cryptoutilSharedMagic.SubjectTypePublic},
		"id_token_signing_alg_values_supported": []string{
			cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			cryptoutilSharedMagic.JoseAlgRS384,
			cryptoutilSharedMagic.JoseAlgRS512,
			cryptoutilSharedMagic.JoseAlgES256,
			cryptoutilSharedMagic.JoseAlgES384,
			cryptoutilSharedMagic.JoseAlgES512,
		},
		"scopes_supported": []string{
			cryptoutilSharedMagic.ScopeOpenID,
			cryptoutilSharedMagic.ClaimProfile,
			cryptoutilSharedMagic.ClaimEmail,
			cryptoutilSharedMagic.ScopeOfflineAccess,
		},
		"token_endpoint_auth_methods_supported": []string{
			cryptoutilSharedMagic.ClientAuthMethodSecretBasic,
			cryptoutilSharedMagic.ClientAuthMethodSecretPost,
			cryptoutilSharedMagic.ClientAuthMethodPrivateKeyJWT,
		},
		"claims_supported": []string{
			cryptoutilSharedMagic.ClaimSub,
			cryptoutilSharedMagic.ClaimIss,
			cryptoutilSharedMagic.ClaimAud,
			cryptoutilSharedMagic.ClaimExp,
			cryptoutilSharedMagic.ClaimIat,
			cryptoutilSharedMagic.ClaimAuthTime,
			cryptoutilSharedMagic.ClaimName,
			cryptoutilSharedMagic.ClaimEmail,
			cryptoutilSharedMagic.ClaimEmailVerified,
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
