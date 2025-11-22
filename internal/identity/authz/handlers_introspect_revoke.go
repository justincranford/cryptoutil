// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	"time"

	"github.com/gofiber/fiber/v2"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// handleIntrospect handles POST /introspect - OAuth 2.1 token introspection endpoint.
func (s *Service) handleIntrospect(c *fiber.Ctx) error {
	// Extract parameters.
	token := c.FormValue(cryptoutilIdentityMagic.ParamToken)
	tokenTypeHint := c.FormValue(cryptoutilIdentityMagic.ParamTokenTypeHint)

	// Validate required parameters.
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "token is required",
		})
	}

	ctx := c.Context()
	tokenRepo := s.repoFactory.TokenRepository()

	// Lookup token in database.
	tokenRecord, err := tokenRepo.GetByTokenValue(ctx, token)
	if err != nil {
		// Token not found - return inactive response.
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"active": false,
		})
	}

	// Check if token is revoked.
	if tokenRecord.RevokedAt != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"active": false,
		})
	}

	// Check if token is expired.
	if tokenRecord.ExpiresAt.Before(time.Now()) {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"active": false,
		})
	}

	// Token is active - return introspection response.
	response := fiber.Map{
		"active":     true,
		"client_id":  tokenRecord.ClientID.String(),
		"token_type": "Bearer",
		"exp":        tokenRecord.ExpiresAt.Unix(),
		"iat":        tokenRecord.IssuedAt.Unix(),
	}

	// Add scopes if present.
	if len(tokenRecord.Scopes) > 0 {
		response["scope"] = tokenRecord.Scopes
	}

	// Add user ID if present (not present for client_credentials).
	if tokenRecord.UserID.Valid {
		response["sub"] = tokenRecord.UserID.UUID.String()
	}

	// Add token type hint if provided.
	if tokenTypeHint != "" {
		response["token_type_hint"] = tokenTypeHint
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// handleRevoke handles POST /revoke - OAuth 2.1 token revocation endpoint.
func (s *Service) handleRevoke(c *fiber.Ctx) error {
	// Extract parameters.
	token := c.FormValue(cryptoutilIdentityMagic.ParamToken)
	tokenTypeHint := c.FormValue(cryptoutilIdentityMagic.ParamTokenTypeHint)

	// Validate required parameters.
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "token is required",
		})
	}

	ctx := c.Context()
	tokenRepo := s.repoFactory.TokenRepository()

	// Lookup token in database.
	tokenRecord, err := tokenRepo.GetByTokenValue(ctx, token)
	if err != nil {
		// Token not found - return success (per RFC 7009 section 2.2).
		return c.SendStatus(fiber.StatusOK)
	}

	// Validate token type hint if provided.
	if tokenTypeHint != "" {
		switch tokenTypeHint {
		case cryptoutilIdentityMagic.TokenTypeAccessToken:
			if tokenRecord.TokenType != cryptoutilIdentityDomain.TokenTypeAccess {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
					"error_description": "Token type hint does not match actual token type",
				})
			}
		case cryptoutilIdentityMagic.TokenTypeRefreshToken:
			if tokenRecord.TokenType != cryptoutilIdentityDomain.TokenTypeRefresh {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
					"error_description": "Token type hint does not match actual token type",
				})
			}
		}
	}

	// Revoke token.
	err = tokenRepo.RevokeByTokenValue(ctx, token)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrTokenRevocationFailed

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": appErr.Message,
		})
	}

	return c.SendStatus(fiber.StatusOK)
}
