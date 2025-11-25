// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package idp

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// handleUserInfo handles GET /userinfo - Return OIDC UserInfo claims.
func (s *Service) handleUserInfo(c *fiber.Ctx) error {
	ctx := c.Context()

	// Extract Bearer token from Authorization header.
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidToken,
			"error_description": "Missing Authorization header",
		})
	}

	// Parse Bearer token.
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != cryptoutilIdentityMagic.AuthorizationBearer {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidToken,
			"error_description": "Invalid Authorization header format",
		})
	}

	accessToken := parts[1]

	// Validate access token and extract claims.
	claims, err := s.tokenSvc.ValidateAccessToken(ctx, accessToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidToken,
			"error_description": "Invalid or expired access token",
		})
	}

	// For UUID tokens, fetch from database to check expiration and get claims.
	// JWT tokens have expiration checked during validation above.
	tokenRepo := s.repoFactory.TokenRepository()

	dbToken, err := tokenRepo.GetByTokenValue(ctx, accessToken)
	if err != nil {
		// Token not found in database (might be JWT - continue with claims from validation).
		// JWT tokens don't exist in database, so this is expected for JWT format.
		if len(claims) == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":             cryptoutilIdentityMagic.ErrorInvalidToken,
				"error_description": "Token not found",
			})
		}
	} else {
		// Check token expiration for UUID tokens.
		if time.Now().After(dbToken.ExpiresAt) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":             cryptoutilIdentityMagic.ErrorInvalidToken,
				"error_description": "Token has expired",
			})
		}

		// Populate claims from database token for UUID format.
		if len(claims) == 0 {
			claims = map[string]any{
				"sub":   dbToken.UserID.UUID.String(),
				"scope": dbToken.Scopes,
			}
		}
	}

	// Extract sub (subject) claim to identify user.
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidToken,
			"error_description": "Token missing sub claim",
		})
	}

	// Fetch user from database.
	userRepo := s.repoFactory.UserRepository()

	user, err := userRepo.GetBySub(ctx, sub)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidToken,
			"error_description": "User not found",
		})
	}

	// Map user to OIDC standard claims.
	userInfo := fiber.Map{
		"sub": user.Sub,
	}

	// Add optional claims based on scopes (extract from token claims).
	scopesAny, scopesExist := claims["scope"]
	if !scopesExist {
		return c.Status(fiber.StatusOK).JSON(userInfo)
	}

	scopes, ok := scopesAny.(string)
	if !ok {
		return c.Status(fiber.StatusOK).JSON(userInfo)
	}

	scopeList := strings.Split(scopes, " ")

	for _, scope := range scopeList {
		switch scope {
		case "profile":
			userInfo["name"] = user.Name
			userInfo["given_name"] = user.GivenName
			userInfo["family_name"] = user.FamilyName
			userInfo["middle_name"] = user.MiddleName
			userInfo["nickname"] = user.Nickname
			userInfo["preferred_username"] = user.PreferredUsername
			userInfo["profile"] = user.Profile
			userInfo["picture"] = user.Picture
			userInfo["website"] = user.Website
			userInfo["gender"] = user.Gender
			userInfo["birthdate"] = user.Birthdate
			userInfo["zoneinfo"] = user.Zoneinfo
			userInfo["locale"] = user.Locale
			userInfo["updated_at"] = user.UpdatedAt.Unix()

		case "email":
			userInfo["email"] = user.Email
			userInfo["email_verified"] = user.EmailVerified

		case "address":
			if user.Address != nil {
				userInfo["address"] = fiber.Map{
					"formatted":      user.Address.Formatted,
					"street_address": user.Address.StreetAddress,
					"locality":       user.Address.Locality,
					"region":         user.Address.Region,
					"postal_code":    user.Address.PostalCode,
					"country":        user.Address.Country,
				}
			}

		case "phone":
			userInfo["phone_number"] = user.PhoneNumber
			userInfo["phone_number_verified"] = user.PhoneVerified
		}
	}

	return c.Status(fiber.StatusOK).JSON(userInfo)
}
