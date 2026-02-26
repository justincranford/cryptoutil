// Copyright (c) 2025 Justin Cranford
//
//

package rs

import (
	"context"
	"log/slog"
	"strings"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TokenService defines the interface for token validation operations.
type TokenService interface {
	ValidateAccessToken(ctx context.Context, token string) (map[string]any, error)
	IsTokenActive(claims map[string]any) bool
	IntrospectToken(ctx context.Context, token string) (*cryptoutilIdentityIssuer.TokenMetadata, error)
}

// TokenValidator validates access tokens for resource server.
type TokenValidator struct {
	tokenSvc TokenService
	logger   *slog.Logger
}

// NewTokenValidator creates a new token validator.
func NewTokenValidator(
	tokenSvc TokenService,
	logger *slog.Logger,
) *TokenValidator {
	return &TokenValidator{
		tokenSvc: tokenSvc,
		logger:   logger,
	}
}

// ValidateToken creates middleware that validates access tokens.
func (v *TokenValidator) ValidateToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()

		// Extract Authorization header.
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			v.logger.Warn("Missing Authorization header",
				"path", c.Path(),
				"method", c.Method())

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidToken,
				"error_description": "Missing Authorization header",
			})
		}

		// Parse Bearer token.
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != cryptoutilSharedMagic.AuthorizationBearer {
			v.logger.Warn("Invalid Authorization header format",
				"path", c.Path(),
				"auth_header", authHeader)

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidToken,
				"error_description": "Invalid Authorization header format",
			})
		}

		token := parts[1]

		// Validate token using token service.
		claims, err := v.tokenSvc.ValidateAccessToken(ctx, token)
		if err != nil {
			v.logger.Warn("Token validation failed",
				"path", c.Path(),
				cryptoutilSharedMagic.StringError, err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidToken,
				"error_description": "Token validation failed",
			})
		}

		// Check token expiration.
		if !v.tokenSvc.IsTokenActive(claims) {
			v.logger.Warn("Token expired or not yet valid",
				"path", c.Path(),
				cryptoutilSharedMagic.ClaimClientID, claims[cryptoutilSharedMagic.ClaimClientID])

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidToken,
				"error_description": "Token expired",
			})
		}

		v.logger.Debug("Token validated successfully",
			cryptoutilSharedMagic.ClaimClientID, claims[cryptoutilSharedMagic.ClaimClientID],
			cryptoutilSharedMagic.ClaimScope, claims[cryptoutilSharedMagic.ClaimScope],
			"path", c.Path())

		// Store claims in context for downstream handlers.
		c.Locals("token_claims", claims)

		return c.Next()
	}
}
