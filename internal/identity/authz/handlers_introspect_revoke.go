// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	"strings"
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

	// First, try to validate as a JWT (self-contained token).
	if isJWT(token) {
		claims, err := s.tokenSvc.ValidateAccessToken(ctx, token)
		if err == nil {
			// JWT signature is valid - check database for revocation status.
			tokenRecord, dbErr := tokenRepo.GetByTokenValue(ctx, token)
			if dbErr == nil && tokenRecord.Revoked {
				// Token was revoked in database - return inactive.
				return c.Status(fiber.StatusOK).JSON(fiber.Map{
					"active": false,
				})
			}

			// Token is a valid JWT and not revoked - return introspection response from claims.
			response := fiber.Map{
				"active":     true,
				"token_type": "Bearer",
			}

			// Extract standard claims.
			if clientID, ok := claims["client_id"].(string); ok {
				response["client_id"] = clientID
			}

			if sub, ok := claims["sub"].(string); ok {
				response["sub"] = sub
			}

			if scope, ok := claims["scope"].(string); ok {
				response["scope"] = scope
			}

			if exp, ok := claims["exp"].(float64); ok {
				response["exp"] = int64(exp)
			}

			if iat, ok := claims["iat"].(float64); ok {
				response["iat"] = int64(iat)
			}

			if iss, ok := claims["iss"].(string); ok {
				response["iss"] = iss
			}

			if aud, ok := claims["aud"]; ok {
				response["aud"] = aud
			}

			if jti, ok := claims["jti"].(string); ok {
				response["jti"] = jti
			}

			// Add token type hint if provided.
			if tokenTypeHint != "" {
				response["token_type_hint"] = tokenTypeHint
			}

			return c.Status(fiber.StatusOK).JSON(response)
		}
		// JWT validation failed - fall through to database lookup.
	}

	// Fallback: lookup token in database (for opaque tokens or JWTs with invalid signatures).
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

// isJWT checks if a token looks like a JWT (three base64url parts separated by dots).
func isJWT(token string) bool {
	parts := strings.Split(token, ".")

	return len(parts) == cryptoutilIdentityMagic.JWSPartCount
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
