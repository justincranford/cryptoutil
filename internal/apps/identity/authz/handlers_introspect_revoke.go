// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	"strings"
	"time"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// handleIntrospect handles POST /introspect - OAuth 2.1 token introspection endpoint.
func (s *Service) handleIntrospect(c *fiber.Ctx) error {
	// Extract parameters.
	token := c.FormValue(cryptoutilSharedMagic.ParamToken)
	tokenTypeHint := c.FormValue(cryptoutilSharedMagic.ParamTokenTypeHint)

	// Validate required parameters.
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
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
				cryptoutilSharedMagic.ParamTokenType: cryptoutilSharedMagic.AuthorizationBearer,
			}

			// Extract standard claims.
			if clientID, ok := claims[cryptoutilSharedMagic.ClaimClientID].(string); ok {
				response[cryptoutilSharedMagic.ClaimClientID] = clientID
			}

			if sub, ok := claims[cryptoutilSharedMagic.ClaimSub].(string); ok {
				response[cryptoutilSharedMagic.ClaimSub] = sub
			}

			if scope, ok := claims[cryptoutilSharedMagic.ClaimScope].(string); ok {
				response[cryptoutilSharedMagic.ClaimScope] = scope
			}

			if exp, ok := claims[cryptoutilSharedMagic.ClaimExp].(float64); ok {
				response[cryptoutilSharedMagic.ClaimExp] = int64(exp)
			}

			if iat, ok := claims[cryptoutilSharedMagic.ClaimIat].(float64); ok {
				response[cryptoutilSharedMagic.ClaimIat] = int64(iat)
			}

			if iss, ok := claims[cryptoutilSharedMagic.ClaimIss].(string); ok {
				response[cryptoutilSharedMagic.ClaimIss] = iss
			}

			if aud, ok := claims[cryptoutilSharedMagic.ClaimAud]; ok {
				response[cryptoutilSharedMagic.ClaimAud] = aud
			}

			if jti, ok := claims[cryptoutilSharedMagic.ClaimJti].(string); ok {
				response[cryptoutilSharedMagic.ClaimJti] = jti
			}

			// Add token type hint if provided.
			if tokenTypeHint != "" {
				response[cryptoutilSharedMagic.ParamTokenTypeHint] = tokenTypeHint
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
	if tokenRecord.ExpiresAt.Before(time.Now().UTC()) {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"active": false,
		})
	}

	// Token is active - return introspection response.
	response := fiber.Map{
		"active":     true,
		cryptoutilSharedMagic.ClaimClientID:  tokenRecord.ClientID.String(),
		cryptoutilSharedMagic.ParamTokenType: cryptoutilSharedMagic.AuthorizationBearer,
		cryptoutilSharedMagic.ClaimExp:        tokenRecord.ExpiresAt.Unix(),
		cryptoutilSharedMagic.ClaimIat:        tokenRecord.IssuedAt.Unix(),
	}

	// Add scopes if present.
	if len(tokenRecord.Scopes) > 0 {
		response[cryptoutilSharedMagic.ClaimScope] = tokenRecord.Scopes
	}

	// Add user ID if present (not present for client_credentials).
	if tokenRecord.UserID.Valid {
		response[cryptoutilSharedMagic.ClaimSub] = tokenRecord.UserID.UUID.String()
	}

	// Add token type hint if provided.
	if tokenTypeHint != "" {
		response[cryptoutilSharedMagic.ParamTokenTypeHint] = tokenTypeHint
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// isJWT checks if a token looks like a JWT (three base64url parts separated by dots).
func isJWT(token string) bool {
	parts := strings.Split(token, ".")

	return len(parts) == cryptoutilSharedMagic.JWSPartCount
}

// handleRevoke handles POST /revoke - OAuth 2.1 token revocation endpoint.
func (s *Service) handleRevoke(c *fiber.Ctx) error {
	// Extract parameters.
	token := c.FormValue(cryptoutilSharedMagic.ParamToken)
	tokenTypeHint := c.FormValue(cryptoutilSharedMagic.ParamTokenTypeHint)

	// Validate required parameters.
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
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
		case cryptoutilSharedMagic.TokenTypeAccessToken:
			if tokenRecord.TokenType != cryptoutilIdentityDomain.TokenTypeAccess {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
					"error_description": "Token type hint does not match actual token type",
				})
			}
		case cryptoutilSharedMagic.TokenTypeRefreshToken:
			if tokenRecord.TokenType != cryptoutilIdentityDomain.TokenTypeRefresh {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
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
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorServerError,
			"error_description": appErr.Message,
		})
	}

	return c.SendStatus(fiber.StatusOK)
}
