// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	"log/slog"
	"strings"
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
	case cryptoutilIdentityMagic.GrantTypeDeviceCode:
		return s.handleDeviceCodeGrant(c)
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

	ctx := c.Context()

	// Retrieve authorization request by code from database.
	authzReqRepo := s.repoFactory.AuthorizationRequestRepository()

	authRequest, err := authzReqRepo.GetByCode(ctx, code)
	if err != nil {
		slog.ErrorContext(ctx, "Authorization code not found or expired", "error", err, "code", code)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "Invalid or expired authorization code",
		})
	}

	// Validate code not expired.
	if authRequest.IsExpired() {
		slog.ErrorContext(ctx, "Authorization code expired", "request_id", authRequest.ID, "expires_at", authRequest.ExpiresAt)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "Authorization code has expired",
		})
	}

	// Validate code not already used (single-use enforcement).
	if authRequest.IsUsed() {
		slog.ErrorContext(ctx, "Authorization code already used", "request_id", authRequest.ID, "used_at", authRequest.UsedAt)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "Authorization code has already been used",
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

	// Mark authorization code as used (single-use enforcement).
	now := time.Now()
	authRequest.Used = true
	authRequest.UsedAt = &now

	if err := authzReqRepo.Update(ctx, authRequest); err != nil {
		slog.ErrorContext(ctx, "Failed to mark authorization code as used", "error", err, "request_id", authRequest.ID)
		// Continue anyway - token issuance is more important than cleanup.
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
	// Validate user ID from authorization request (required after login/consent).
	if !authRequest.UserID.Valid || authRequest.UserID.UUID == googleUuid.Nil {
		slog.ErrorContext(ctx, "User ID missing from authorization request", "auth_request_id", authRequest.ID)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "Authorization request missing user ID (login/consent not completed)",
		})
	}

	// Ensure token service is configured.
	if s.tokenSvc == nil {
		slog.ErrorContext(ctx, "Token service not configured")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Token service not configured",
		})
	}

	accessTokenClaims := map[string]any{
		"sub":       authRequest.UserID.UUID.String(),
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
		"request_id", authRequest.ID,
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
		slog.ErrorContext(c.Context(), "Client authentication failed in client_credentials grant",
			"error", err,
			"client_id", c.FormValue(cryptoutilIdentityMagic.ParamClientID),
		)

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidClient,
			"error_description": "Client authentication failed",
		})
	}

	// Extract scope.
	scope := c.FormValue(cryptoutilIdentityMagic.ParamScope)

	ctx := c.Context()

	// Ensure token service is configured.
	if s.tokenSvc == nil {
		slog.ErrorContext(ctx, "Token service not configured for client_credentials grant",
			"client_id", client.ClientID,
		)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Token service not configured",
		})
	}

	// Generate access token.
	now := time.Now()
	expiresAt := now.Add(time.Duration(cryptoutilIdentityMagic.AccessTokenExpirySeconds) * time.Second)

	accessTokenClaims := map[string]any{
		"client_id": client.ClientID,
		"scope":     scope,
		"exp":       expiresAt.Unix(),
		"iat":       now.Unix(),
	}

	accessToken, err := s.tokenSvc.IssueAccessToken(ctx, accessTokenClaims)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrTokenIssuanceFailed

		slog.ErrorContext(ctx, "Access token issuance failed in client_credentials grant",
			"error", err,
			"client_id", client.ClientID,
			"scope", scope,
		)

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": appErr.Message,
		})
	}

	// Store token in database for revocation support.
	var scopes []string
	if scope != "" {
		scopes = strings.Split(scope, " ")
	}

	tokenRecord := &cryptoutilIdentityDomain.Token{
		TokenValue:  accessToken,
		TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat: cryptoutilIdentityDomain.TokenFormatJWS,
		ClientID:    client.ID,
		Scopes:      scopes,
		IssuedAt:    now,
		ExpiresAt:   expiresAt,
	}

	tokenRepo := s.repoFactory.TokenRepository()
	if err := tokenRepo.Create(ctx, tokenRecord); err != nil {
		slog.WarnContext(ctx, "Failed to store access token for revocation tracking",
			"error", err,
			"client_id", client.ClientID,
		)
		// Continue anyway - token was issued successfully, just won't be revokable.
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

	// Ensure token service is configured.
	if s.tokenSvc == nil {
		slog.ErrorContext(ctx, "Token service not configured for refresh_token grant",
			"client_id", clientID,
		)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Token service not configured",
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

// handleDeviceCodeGrant handles device_code grant (RFC 8628 Section 3.4).
func (s *Service) handleDeviceCodeGrant(c *fiber.Ctx) error {
	// Extract device_code and client_id parameters.
	deviceCode := c.FormValue(cryptoutilIdentityMagic.ParamDeviceCode)
	clientID := c.FormValue(cryptoutilIdentityMagic.ParamClientID)

	// Validate required parameters.
	if deviceCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "device_code is required",
		})
	}

	if clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "client_id is required",
		})
	}

	ctx := c.Context()

	// Retrieve device authorization from repository.
	deviceAuthRepo := s.repoFactory.DeviceAuthorizationRepository()

	deviceAuth, err := deviceAuthRepo.GetByDeviceCode(ctx, deviceCode)
	if err != nil {
		slog.ErrorContext(ctx, "Device authorization not found", "device_code", deviceCode, "error", err)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "Invalid or expired device_code",
		})
	}

	// Validate client_id matches.
	if deviceAuth.ClientID != clientID {
		slog.ErrorContext(ctx, "Client ID mismatch for device authorization", "expected", deviceAuth.ClientID, "actual", clientID)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "client_id does not match device_code",
		})
	}

	// Check if device code is expired (RFC 8628 Section 3.5).
	if deviceAuth.IsExpired() {
		slog.InfoContext(ctx, "Device code expired", "device_code", deviceCode, "expires_at", deviceAuth.ExpiresAt)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorExpiredToken,
			"error_description": "device_code has expired",
		})
	}

	// Check if device code has already been used (RFC 8628 Section 3.5).
	if deviceAuth.IsUsed() {
		slog.WarnContext(ctx, "Device code already used", "device_code", deviceCode, "used_at", deviceAuth.UsedAt)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "device_code has already been used",
		})
	}

	// Enforce polling rate limiting (RFC 8628 Section 3.5).
	if deviceAuth.LastPolledAt != nil {
		minPollTime := deviceAuth.LastPolledAt.Add(cryptoutilIdentityMagic.DefaultPollingInterval)
		if time.Now().Before(minPollTime) {
			slog.DebugContext(ctx, "Polling too fast", "device_code", deviceCode, "last_polled_at", deviceAuth.LastPolledAt)

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":             cryptoutilIdentityMagic.ErrorSlowDown,
				"error_description": "Polling too fast, slow down",
			})
		}
	}

	// Update last polled timestamp.
	now := time.Now()
	deviceAuth.LastPolledAt = &now

	if err := deviceAuthRepo.Update(ctx, deviceAuth); err != nil {
		slog.ErrorContext(ctx, "Failed to update device authorization polling timestamp", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to update polling timestamp",
		})
	}

	// Check authorization status (RFC 8628 Section 3.5).
	switch {
	case deviceAuth.IsPending():
		// User has not yet authorized - client should continue polling.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorAuthorizationPending,
			"error_description": "Authorization pending, user has not yet authorized",
		})
	case deviceAuth.IsDenied():
		// User explicitly denied authorization.
		slog.InfoContext(ctx, "Device authorization denied by user", "device_code", deviceCode)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorAccessDenied,
			"error_description": "User denied authorization",
		})
	case deviceAuth.IsAuthorized():
		// User has authorized - issue tokens.
		return s.issueDeviceCodeTokens(c, deviceAuth)
	default:
		// Unknown status - should never happen.
		slog.ErrorContext(ctx, "Invalid device authorization status", "status", deviceAuth.Status)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Invalid device authorization status",
		})
	}
}

// issueDeviceCodeTokens issues access and refresh tokens for authorized device code.
func (s *Service) issueDeviceCodeTokens(c *fiber.Ctx, deviceAuth *cryptoutilIdentityDomain.DeviceAuthorization) error {
	ctx := c.Context()

	// Mark device code as used.
	now := time.Now()
	deviceAuth.Status = cryptoutilIdentityDomain.DeviceAuthStatusUsed
	deviceAuth.UsedAt = &now

	deviceAuthRepo := s.repoFactory.DeviceAuthorizationRepository()
	if err := deviceAuthRepo.Update(ctx, deviceAuth); err != nil {
		slog.ErrorContext(ctx, "Failed to mark device code as used", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to update device authorization status",
		})
	}

	// Validate UserID is set (device must be authorized by a user).
	if !deviceAuth.UserID.Valid {
		slog.ErrorContext(ctx, "Device authorization missing user ID", "device_code", deviceAuth.DeviceCode)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Device authorization missing user ID",
		})
	}

	// Retrieve client for token generation.
	clientRepo := s.repoFactory.ClientRepository()

	client, err := clientRepo.GetByClientID(ctx, deviceAuth.ClientID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to retrieve client for device code token", "client_id", deviceAuth.ClientID, "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to retrieve client",
		})
	}

	// Ensure token service is configured.
	if s.tokenSvc == nil {
		slog.ErrorContext(ctx, "Token service not configured")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Token service not configured",
		})
	}

	// Generate access token with standard OAuth 2.1 claims.
	accessTokenClaims := map[string]any{
		"sub":       deviceAuth.UserID.UUID.String(),
		"client_id": deviceAuth.ClientID,
		"scope":     deviceAuth.Scope,
		"exp":       time.Now().Add(time.Duration(client.AccessTokenLifetime) * time.Second).Unix(),
		"iat":       time.Now().Unix(),
	}

	accessToken, err := s.tokenSvc.IssueAccessToken(ctx, accessTokenClaims)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrTokenIssuanceFailed

		slog.ErrorContext(ctx, "Access token issuance failed for device code", "error", err)

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": appErr.Message,
		})
	}

	// Generate refresh token (optional, based on client configuration).
	var refreshToken string

	if client.RefreshTokenLifetime > 0 {
		refreshToken, err = s.tokenSvc.IssueRefreshToken(ctx)
		if err != nil {
			appErr := cryptoutilIdentityAppErr.ErrTokenIssuanceFailed

			slog.ErrorContext(ctx, "Refresh token issuance failed for device code", "error", err)

			return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
				"error":             cryptoutilIdentityMagic.ErrorServerError,
				"error_description": appErr.Message,
			})
		}
	}

	slog.InfoContext(ctx, "Device code token exchange successful",
		"client_id", deviceAuth.ClientID,
		"user_id", deviceAuth.UserID.UUID.String(),
		"scope", deviceAuth.Scope,
	)

	// Build token response (RFC 8628 Section 3.5).
	tokenResponse := fiber.Map{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   client.AccessTokenLifetime,
		"scope":        deviceAuth.Scope,
	}

	if refreshToken != "" {
		tokenResponse["refresh_token"] = refreshToken
	}

	return c.Status(fiber.StatusOK).JSON(tokenResponse)
}
