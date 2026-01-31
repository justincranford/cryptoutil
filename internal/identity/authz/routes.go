// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	"log/slog"

	fiber "github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers all OAuth 2.1 authorization server routes.
func (s *Service) RegisterRoutes(app *fiber.App) {
	// Swagger UI OpenAPI spec endpoint.
	swaggerHandler, err := ServeOpenAPISpec()
	if err != nil {
		// Swagger UI is non-critical, but log error for diagnostics.
		slog.Warn("Failed to generate OpenAPI spec for Swagger UI", "error", err)
	} else {
		app.Get("/ui/swagger/doc.json", swaggerHandler)
	}

	// Health check endpoint (no prefix).
	app.Get("/health", s.handleHealth)

	// OAuth 2.1 Discovery endpoints (RFC 8414).
	app.Get("/.well-known/oauth-authorization-server", s.handleOAuthMetadata)
	app.Get("/.well-known/openid-configuration", s.handleOIDCDiscovery)

	// OAuth 2.1 endpoints with /oauth2/v1 prefix.
	oauth := app.Group("/oauth2/v1")
	oauth.Get("/authorize", s.handleAuthorizeGET)
	oauth.Post("/authorize", s.handleAuthorizePOST)
	oauth.Post("/token", s.handleToken)
	oauth.Post("/introspect", s.handleIntrospect)
	oauth.Post("/revoke", s.handleRevoke)
	oauth.Get("/jwks", s.handleJWKS)
	oauth.Post("/device_authorization", s.handleDeviceAuthorization)
	oauth.Post("/par", s.handlePAR)

	// Client management endpoints.
	oauth.Post("/clients/:id/rotate-secret", s.handleClientSecretRotation)

	// MFA recovery code endpoints (OIDC /oidc/v1 prefix).
	oidc := app.Group("/oidc/v1")
	oidc.Post("/mfa/recovery-codes/generate", s.handleGenerateRecoveryCodes)
	oidc.Get("/mfa/recovery-codes/count", s.handleGetRecoveryCodeCount)
	oidc.Post("/mfa/recovery-codes/regenerate", s.handleRegenerateRecoveryCodes)
	oidc.Post("/mfa/verify-recovery-code", s.handleVerifyRecoveryCode)

	// MFA email OTP endpoints.
	oidc.Post("/mfa/email-otp/send", s.handleSendEmailOTP)
	oidc.Post("/mfa/email-otp/verify", s.handleVerifyEmailOTP)

	// MFA TOTP endpoints.
	oidc.Post("/mfa/totp/enroll", s.handleEnrollTOTP)
	oidc.Post("/mfa/totp/verify", s.handleVerifyTOTP)
	oidc.Get("/mfa/totp/step-up", s.handleCheckMFAStepUp)
	oidc.Post("/mfa/totp/backup-codes/generate", s.handleGenerateTOTPBackupCodes)
	oidc.Post("/mfa/totp/backup-codes/verify", s.handleVerifyTOTPBackupCode)

	// MFA admin endpoints.
	oidc.Post("/mfa/enroll", s.handleEnrollMFA)
	oidc.Get("/mfa/factors", s.handleListMFAFactors)
	oidc.Delete("/mfa/factors/:id", s.handleDeleteMFAFactor)
}
