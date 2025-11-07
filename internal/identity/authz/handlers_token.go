package authz

import (
	"time"

	"github.com/gofiber/fiber/v2"

	cryptoutilIdentityApperr "cryptoutil/internal/identity/apperr"
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

	// Placeholder: generate tokens.
	accessTokenClaims := map[string]any{
		"sub":       "user123",
		"client_id": clientID,
		"scope":     "openid profile email",
		"exp":       time.Now().Add(time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}

	accessToken, err := s.tokenSvc.IssueAccessToken(ctx, accessTokenClaims)
	if err != nil {
		appErr := cryptoutilIdentityApperr.ErrTokenIssuanceFailed

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": appErr.Message,
		})
	}

	// Placeholder response.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   cryptoutilIdentityMagic.AccessTokenExpirySeconds,
		"scope":        "openid profile email",
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
		appErr := cryptoutilIdentityApperr.ErrTokenIssuanceFailed

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
		appErr := cryptoutilIdentityApperr.ErrTokenNotFound

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
		appErr := cryptoutilIdentityApperr.ErrTokenIssuanceFailed

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
