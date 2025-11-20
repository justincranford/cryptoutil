// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityPKCE "cryptoutil/internal/identity/authz/pkce"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// handleToken handles POST /token - OAuth 2.1 token endpoint.
func (s *Service) handleToken(c *fiber.Ctx) error {
	// Extract grant type.
	grantType := c.FormValue(cryptoutilIdentityMagic.ParamGrantType)

	switch grantType {
	case cryptoutilIdentityMagic.GrantTypeAuthorizationCode:
		return s.handleAuthorizationCodeGrant(c)
	case cryptoutilIdentityMagic.GrantTypeClientCredentials:
		return s.handleClientCredentialsGrant(c)
	case cryptoutilIdentityMagic.GrantTypeRefreshToken:
		return s.handleRefreshTokenGrant(c)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorUnsupportedGrantType,
			"error_description": "Unsupported grant type",
		})
	}
}

// handleAuthorizationCodeGrant handles authorization_code grant.
func (s *Service) handleAuthorizationCodeGrant(c *fiber.Ctx) error {
	// Extract parameters.
	code := c.FormValue(cryptoutilIdentityMagic.ParamCode)
	redirectURI := c.FormValue(cryptoutilIdentityMagic.ParamRedirectURI)
	clientID := c.FormValue(cryptoutilIdentityMagic.ParamClientID)
	codeVerifier := c.FormValue(cryptoutilIdentityMagic.ParamCodeVerifier)

	// Validate required parameters.
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "code is required",
		})
	}

	if redirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "redirect_uri is required",
		})
	}

	if clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "client_id is required",
		})
	}

	if codeVerifier == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "code_verifier is required (OAuth 2.1 requires PKCE)",
		})
	}

	// TODO: Validate authorization code.
	// TODO: Validate PKCE code_verifier against stored code_challenge.
	// TODO: Validate client credentials.
	// TODO: Generate access token and refresh token.

	ctx := c.Context()

	// Retrieve authorization request by code.
	authRequest, err := s.authReqStore.GetByCode(ctx, code)
	if err != nil {
		slog.ErrorContext(ctx, "Authorization code not found or expired", "error", err, "code", code)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "Invalid or expired authorization code",
		})
	}

	// Validate client ID matches.
	if authRequest.ClientID != clientID {
		slog.ErrorContext(ctx, "Client ID mismatch", "expected", authRequest.ClientID, "provided", clientID)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "Client ID does not match authorization code",
		})
	}

	// Validate redirect URI matches.
	if authRequest.RedirectURI != redirectURI {
		slog.ErrorContext(ctx, "Redirect URI mismatch", "expected", authRequest.RedirectURI, "provided", redirectURI)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "Redirect URI does not match authorization request",
		})
	}

	// Validate PKCE code_verifier.
	if !cryptoutilIdentityPKCE.ValidateCodeVerifier(codeVerifier, authRequest.CodeChallenge, authRequest.CodeChallengeMethod) {
		slog.ErrorContext(ctx, "PKCE validation failed", "code_challenge", authRequest.CodeChallenge, "method", authRequest.CodeChallengeMethod)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "PKCE validation failed",
		})
	}

	// Delete authorization code (single-use).
	if err := s.authReqStore.Delete(ctx, authRequest.RequestID); err != nil {
		// Continue anyway - token issuance is more important.
		slog.ErrorContext(ctx, "Failed to delete authorization code", "error", err, "request_id", authRequest.RequestID)
	}

	// Get client for token configuration.
	clientRepo := s.repoFactory.ClientRepository()

	client, err := clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrClientNotFound

		slog.ErrorContext(ctx, "Client not found", "error", err, "client_id", clientID)

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidClient,
			"error_description": appErr.Message,
		})
	}

	// Generate access token.
	// TODO: In future tasks, populate with real user ID from authRequest.UserID after login/consent integration.
	userIDPlaceholder := googleUuid.Must(googleUuid.NewV7())

	accessTokenClaims := map[string]any{
		"sub":       userIDPlaceholder.String(),
		"client_id": clientID,
		"scope":     authRequest.Scope,
		"exp":       time.Now().Add(time.Duration(client.AccessTokenLifetime) * time.Second).Unix(),
		"iat":       time.Now().Unix(),
	}

	accessToken, err := s.tokenSvc.IssueAccessToken(ctx, accessTokenClaims)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrTokenIssuanceFailed

		slog.ErrorContext(ctx, "Access token issuance failed", "error", err)

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": appErr.Message,
		})
	}

	// Generate refresh token.
	refreshToken, err := s.tokenSvc.IssueRefreshToken(ctx)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrTokenIssuanceFailed

		slog.ErrorContext(ctx, "Refresh token issuance failed", "error", err)

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": appErr.Message,
		})
	}

	slog.InfoContext(ctx, "Token exchange successful",
		"client_id", clientID,
		"scope", authRequest.Scope,
		"request_id", authRequest.RequestID,
	)

	// Return tokens.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    client.AccessTokenLifetime,
		"refresh_token": refreshToken,
		"scope":         authRequest.Scope,
	})
}

// handleClientCredentialsGrant handles client_credentials grant.
func (s *Service) handleClientCredentialsGrant(c *fiber.Ctx) error {
	// Authenticate client.
	client, err := s.authenticateClient(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidClient,
			"error_description": "Client authentication failed",
		})
	}

	// Extract scope.
	scope := c.FormValue(cryptoutilIdentityMagic.ParamScope)

	ctx := c.Context()

	// Generate access token.
	accessTokenClaims := map[string]any{
		"client_id": client.ClientID,
		"scope":     scope,
		"exp":       time.Now().Add(time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}

	accessToken, err := s.tokenSvc.IssueAccessToken(ctx, accessTokenClaims)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrTokenIssuanceFailed

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": appErr.Message,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   cryptoutilIdentityMagic.AccessTokenExpirySeconds,
		"scope":        scope,
	})
}

// handleRefreshTokenGrant handles refresh_token grant.
func (s *Service) handleRefreshTokenGrant(c *fiber.Ctx) error {
	// Extract parameters.
	refreshToken := c.FormValue(cryptoutilIdentityMagic.ParamRefreshToken)
	clientID := c.FormValue(cryptoutilIdentityMagic.ParamClientID)
	scope := c.FormValue(cryptoutilIdentityMagic.ParamScope)

	// Validate required parameters.
	if refreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "refresh_token is required",
		})
	}

	if clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "client_id is required",
		})
	}

	ctx := c.Context()
	tokenRepo := s.repoFactory.TokenRepository()

	// Validate refresh token.
	token, err := tokenRepo.GetByTokenValue(ctx, refreshToken)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrTokenNotFound

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "Invalid refresh token",
		})
	}

	// Validate token type.
	if token.TokenType != cryptoutilIdentityDomain.TokenTypeRefresh {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "Token is not a refresh token",
		})
	}

	// Validate token not revoked.
	if token.RevokedAt != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "Refresh token has been revoked",
		})
	}

	// Validate token not expired.
	if token.ExpiresAt.Before(time.Now()) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "Refresh token has expired",
		})
	}

	// Generate new access token.
	accessTokenClaims := map[string]any{
		"sub":       token.UserID,
		"client_id": clientID,
		"scope":     scope,
		"exp":       time.Now().Add(time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}

	accessToken, err := s.tokenSvc.IssueAccessToken(ctx, accessTokenClaims)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrTokenIssuanceFailed

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": appErr.Message,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   cryptoutilIdentityMagic.AccessTokenExpirySeconds,
		"scope":        scope,
	})
}
