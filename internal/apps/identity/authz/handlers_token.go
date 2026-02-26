// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	"log/slog"
	"strings"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityAuthzPkce "cryptoutil/internal/apps/identity/authz/pkce"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// handleToken handles POST /token - OAuth 2.1 token endpoint.
func (s *Service) handleToken(c *fiber.Ctx) error {
	// Extract grant type.
	grantType := c.FormValue(cryptoutilSharedMagic.ParamGrantType)

	switch grantType {
	case cryptoutilSharedMagic.GrantTypeAuthorizationCode:
		return s.handleAuthorizationCodeGrant(c)
	case cryptoutilSharedMagic.GrantTypeClientCredentials:
		return s.handleClientCredentialsGrant(c)
	case cryptoutilSharedMagic.GrantTypeRefreshToken:
		return s.handleRefreshTokenGrant(c)
	case cryptoutilSharedMagic.GrantTypeDeviceCode:
		return s.handleDeviceCodeGrant(c)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorUnsupportedGrantType,
			"error_description": "Unsupported grant type",
		})
	}
}

// handleAuthorizationCodeGrant handles authorization_code grant.
func (s *Service) handleAuthorizationCodeGrant(c *fiber.Ctx) error {
	// Extract parameters.
	code := c.FormValue(cryptoutilSharedMagic.ParamCode)
	redirectURI := c.FormValue(cryptoutilSharedMagic.ParamRedirectURI)
	clientID := c.FormValue(cryptoutilSharedMagic.ParamClientID)
	codeVerifier := c.FormValue(cryptoutilSharedMagic.ParamCodeVerifier)

	// Validate required parameters.
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description": "code is required",
		})
	}

	if redirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description": "redirect_uri is required",
		})
	}

	if clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description": "client_id is required",
		})
	}

	if codeVerifier == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description": "code_verifier is required (OAuth 2.1 requires PKCE)",
		})
	}

	ctx := c.Context()

	// Retrieve authorization request by code from database.
	authzReqRepo := s.repoFactory.AuthorizationRequestRepository()

	authRequest, err := authzReqRepo.GetByCode(ctx, code)
	if err != nil {
		slog.ErrorContext(ctx, "Authorization code not found or expired", cryptoutilSharedMagic.StringError, err, cryptoutilSharedMagic.ResponseTypeCode, code)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidGrant,
			"error_description": "Invalid or expired authorization code",
		})
	}

	// Validate code not expired.
	if authRequest.IsExpired() {
		slog.ErrorContext(ctx, "Authorization code expired", "request_id", authRequest.ID, "expires_at", authRequest.ExpiresAt)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidGrant,
			"error_description": "Authorization code has expired",
		})
	}

	// Validate code not already used (single-use enforcement).
	if authRequest.IsUsed() {
		slog.ErrorContext(ctx, "Authorization code already used", "request_id", authRequest.ID, "used_at", authRequest.UsedAt)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidGrant,
			"error_description": "Authorization code has already been used",
		})
	}

	// Validate client ID matches.
	if authRequest.ClientID != clientID {
		slog.ErrorContext(ctx, "Client ID mismatch", "expected", authRequest.ClientID, "provided", clientID)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidGrant,
			"error_description": "Client ID does not match authorization code",
		})
	}

	// Validate redirect URI matches.
	if authRequest.RedirectURI != redirectURI {
		slog.ErrorContext(ctx, "Redirect URI mismatch", "expected", authRequest.RedirectURI, "provided", redirectURI)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidGrant,
			"error_description": "Redirect URI does not match authorization request",
		})
	}

	// Validate PKCE code_verifier.
	if !cryptoutilIdentityAuthzPkce.ValidateCodeVerifier(codeVerifier, authRequest.CodeChallenge, authRequest.CodeChallengeMethod) {
		slog.ErrorContext(ctx, "PKCE validation failed", cryptoutilSharedMagic.ParamCodeChallenge, authRequest.CodeChallenge, "method", authRequest.CodeChallengeMethod)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidGrant,
			"error_description": "PKCE validation failed",
		})
	}

	// Mark authorization code as used (single-use enforcement).
	now := time.Now().UTC()
	authRequest.Used = true
	authRequest.UsedAt = &now

	if err := authzReqRepo.Update(ctx, authRequest); err != nil {
		slog.ErrorContext(ctx, "Failed to mark authorization code as used", cryptoutilSharedMagic.StringError, err, "request_id", authRequest.ID)
		// Continue anyway - token issuance is more important than cleanup.
	}

	// Get client for token configuration.
	clientRepo := s.repoFactory.ClientRepository()

	client, err := clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrClientNotFound

		slog.ErrorContext(ctx, "Client not found", cryptoutilSharedMagic.StringError, err, cryptoutilSharedMagic.ClaimClientID, clientID)

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidClient,
			"error_description": appErr.Message,
		})
	}

	// Generate access token.
	// Validate user ID from authorization request (required after login/consent).
	if !authRequest.UserID.Valid || authRequest.UserID.UUID == googleUuid.Nil {
		slog.ErrorContext(ctx, "User ID missing from authorization request", "auth_request_id", authRequest.ID)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description": "Authorization request missing user ID (login/consent not completed)",
		})
	}

	// Ensure token service is configured.
	if s.tokenSvc == nil {
		slog.ErrorContext(ctx, "Token service not configured")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorServerError,
			"error_description": "Token service not configured",
		})
	}

	accessTokenClaims := map[string]any{
		cryptoutilSharedMagic.ClaimSub:       authRequest.UserID.UUID.String(),
		cryptoutilSharedMagic.ClaimClientID: clientID,
		cryptoutilSharedMagic.ClaimScope:     authRequest.Scope,
		cryptoutilSharedMagic.ClaimExp:       time.Now().UTC().Add(time.Duration(client.AccessTokenLifetime) * time.Second).Unix(),
		cryptoutilSharedMagic.ClaimIat:       time.Now().UTC().Unix(),
	}

	accessToken, err := s.tokenSvc.IssueAccessToken(ctx, accessTokenClaims)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrTokenIssuanceFailed

		slog.ErrorContext(ctx, "Access token issuance failed", cryptoutilSharedMagic.StringError, err)

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorServerError,
			"error_description": appErr.Message,
		})
	}

	// Generate refresh token.
	refreshToken, err := s.tokenSvc.IssueRefreshToken(ctx)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrTokenIssuanceFailed

		slog.ErrorContext(ctx, "Refresh token issuance failed", cryptoutilSharedMagic.StringError, err)

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorServerError,
			"error_description": appErr.Message,
		})
	}

	slog.InfoContext(ctx, "Token exchange successful",
		cryptoutilSharedMagic.ClaimClientID, clientID,
		cryptoutilSharedMagic.ClaimScope, authRequest.Scope,
		"request_id", authRequest.ID,
	)

	// Return tokens.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		cryptoutilSharedMagic.TokenTypeAccessToken:  accessToken,
		cryptoutilSharedMagic.ParamTokenType:    cryptoutilSharedMagic.AuthorizationBearer,
		cryptoutilSharedMagic.ParamExpiresIn:    client.AccessTokenLifetime,
		cryptoutilSharedMagic.GrantTypeRefreshToken: refreshToken,
		cryptoutilSharedMagic.ClaimScope:         authRequest.Scope,
	})
}

// handleClientCredentialsGrant handles client_credentials grant.
func (s *Service) handleClientCredentialsGrant(c *fiber.Ctx) error {
	// Authenticate client.
	client, err := s.authenticateClient(c)
	if err != nil {
		slog.ErrorContext(c.Context(), "Client authentication failed in client_credentials grant",
			cryptoutilSharedMagic.StringError, err,
			cryptoutilSharedMagic.ClaimClientID, c.FormValue(cryptoutilSharedMagic.ParamClientID),
		)

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidClient,
			"error_description": "Client authentication failed",
		})
	}

	// Extract scope.
	scope := c.FormValue(cryptoutilSharedMagic.ParamScope)

	ctx := c.Context()

	// Ensure token service is configured.
	if s.tokenSvc == nil {
		slog.ErrorContext(ctx, "Token service not configured for client_credentials grant",
			cryptoutilSharedMagic.ClaimClientID, client.ClientID,
		)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorServerError,
			"error_description": "Token service not configured",
		})
	}

	// Generate access token.
	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(cryptoutilSharedMagic.AccessTokenExpirySeconds) * time.Second)

	accessTokenClaims := map[string]any{
		cryptoutilSharedMagic.ClaimClientID: client.ClientID,
		cryptoutilSharedMagic.ClaimScope:     scope,
		cryptoutilSharedMagic.ClaimExp:       expiresAt.Unix(),
		cryptoutilSharedMagic.ClaimIat:       now.Unix(),
	}

	accessToken, err := s.tokenSvc.IssueAccessToken(ctx, accessTokenClaims)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrTokenIssuanceFailed

		slog.ErrorContext(ctx, "Access token issuance failed in client_credentials grant",
			cryptoutilSharedMagic.StringError, err,
			cryptoutilSharedMagic.ClaimClientID, client.ClientID,
			cryptoutilSharedMagic.ClaimScope, scope,
		)

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorServerError,
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
			cryptoutilSharedMagic.StringError, err,
			cryptoutilSharedMagic.ClaimClientID, client.ClientID,
		)
		// Continue anyway - token was issued successfully, just won't be revokable.
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		cryptoutilSharedMagic.TokenTypeAccessToken: accessToken,
		cryptoutilSharedMagic.ParamTokenType:   cryptoutilSharedMagic.AuthorizationBearer,
		cryptoutilSharedMagic.ParamExpiresIn:   cryptoutilSharedMagic.AccessTokenExpirySeconds,
		cryptoutilSharedMagic.ClaimScope:        scope,
	})
}

// handleRefreshTokenGrant handles refresh_token grant.
func (s *Service) handleRefreshTokenGrant(c *fiber.Ctx) error {
	// Extract parameters.
	refreshToken := c.FormValue(cryptoutilSharedMagic.ParamRefreshToken)
	clientID := c.FormValue(cryptoutilSharedMagic.ParamClientID)
	scope := c.FormValue(cryptoutilSharedMagic.ParamScope)

	// Validate required parameters.
	if refreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description": "refresh_token is required",
		})
	}

	if clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
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
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidGrant,
			"error_description": "Invalid refresh token",
		})
	}

	// Validate token type.
	if token.TokenType != cryptoutilIdentityDomain.TokenTypeRefresh {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidGrant,
			"error_description": "Token is not a refresh token",
		})
	}

	// Validate token not revoked.
	if token.RevokedAt != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidGrant,
			"error_description": "Refresh token has been revoked",
		})
	}

	// Validate token not expired.
	if token.ExpiresAt.Before(time.Now().UTC()) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidGrant,
			"error_description": "Refresh token has expired",
		})
	}

	// Ensure token service is configured.
	if s.tokenSvc == nil {
		slog.ErrorContext(ctx, "Token service not configured for refresh_token grant",
			cryptoutilSharedMagic.ClaimClientID, clientID,
		)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorServerError,
			"error_description": "Token service not configured",
		})
	}

	// Generate new access token.
	accessTokenClaims := map[string]any{
		cryptoutilSharedMagic.ClaimSub:       token.UserID,
		cryptoutilSharedMagic.ClaimClientID: clientID,
		cryptoutilSharedMagic.ClaimScope:     scope,
		cryptoutilSharedMagic.ClaimExp:       time.Now().UTC().Add(time.Hour).Unix(),
		cryptoutilSharedMagic.ClaimIat:       time.Now().UTC().Unix(),
	}

	accessToken, err := s.tokenSvc.IssueAccessToken(ctx, accessTokenClaims)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrTokenIssuanceFailed

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorServerError,
			"error_description": appErr.Message,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		cryptoutilSharedMagic.TokenTypeAccessToken: accessToken,
		cryptoutilSharedMagic.ParamTokenType:   cryptoutilSharedMagic.AuthorizationBearer,
		cryptoutilSharedMagic.ParamExpiresIn:   cryptoutilSharedMagic.AccessTokenExpirySeconds,
		cryptoutilSharedMagic.ClaimScope:        scope,
	})
}

// handleDeviceCodeGrant handles device_code grant (RFC 8628 Section 3.4).
