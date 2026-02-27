// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package idp

import (
	"strings"
	"time"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// MIMEApplicationJWT is the MIME type for JWT responses.
const MIMEApplicationJWT = "application/jwt"

// handleUserInfo handles GET /userinfo - Return OIDC UserInfo claims.
// Per OAuth 2.1, supports both JSON and JWT-signed responses based on Accept header.
// - Accept: application/json → returns JSON object (default)
// - Accept: application/jwt → returns signed JWT.
func (s *Service) handleUserInfo(c *fiber.Ctx) error {
	ctx := c.Context()

	// Extract Bearer token from Authorization header.
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidToken,
			"error_description":               "Missing Authorization header",
		})
	}

	// Parse Bearer token.
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != cryptoutilSharedMagic.AuthorizationBearer {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidToken,
			"error_description":               "Invalid Authorization header format",
		})
	}

	accessToken := parts[1]

	// Validate access token and extract claims.
	claims, err := s.tokenSvc.ValidateAccessToken(ctx, accessToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidToken,
			"error_description":               "Invalid or expired access token",
		})
	}

	// For UUID tokens, fetch from database to check expiration and get claims.
	// JWT tokens have expiration checked during validation above.
	tokenRepo := s.repoFactory.TokenRepository()

	var clientID string

	dbToken, err := tokenRepo.GetByTokenValue(ctx, accessToken)
	if err != nil {
		// Token not found in database (might be JWT - continue with claims from validation).
		// JWT tokens don't exist in database, so this is expected for JWT format.
		if len(claims) == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidToken,
				"error_description":               "Token not found",
			})
		}

		// Extract client_id from JWT claims for JWT-signed response.
		if aud, ok := claims[cryptoutilSharedMagic.ClaimAud].(string); ok {
			clientID = aud
		}
	} else {
		// Check token expiration for UUID tokens.
		if time.Now().UTC().After(dbToken.ExpiresAt) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidToken,
				"error_description":               "Token has expired",
			})
		}

		// Populate claims from database token for UUID format.
		if len(claims) == 0 {
			claims = map[string]any{
				cryptoutilSharedMagic.ClaimSub:   dbToken.UserID.UUID.String(),
				cryptoutilSharedMagic.ClaimScope: dbToken.Scopes,
			}
		}

		// Get client_id from database token for JWT-signed response.
		clientID = dbToken.ClientID.String()
	}

	// Extract sub (subject) claim to identify user.
	sub, ok := claims[cryptoutilSharedMagic.ClaimSub].(string)
	if !ok || sub == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidToken,
			"error_description":               "Token missing sub claim",
		})
	}

	// Fetch user from database.
	userRepo := s.repoFactory.UserRepository()

	user, err := userRepo.GetBySub(ctx, sub)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidToken,
			"error_description":               "User not found",
		})
	}

	// Map user to OIDC standard claims.
	userInfo := make(map[string]any)
	userInfo[cryptoutilSharedMagic.ClaimSub] = user.Sub

	// Add optional claims based on scopes (extract from token claims).
	scopesAny, scopesExist := claims[cryptoutilSharedMagic.ClaimScope]
	if scopesExist {
		scopes, scopesOk := scopesAny.(string)
		if scopesOk {
			scopeList := strings.Split(scopes, " ")
			addScopeBasedClaims(userInfo, scopeList, user)
		}
	}

	// Check Accept header for JWT-signed response (OAuth 2.1 compliance).
	acceptHeader := c.Get("Accept")
	if strings.Contains(acceptHeader, MIMEApplicationJWT) && clientID != "" {
		// Return JWT-signed userinfo response.
		jwtResponse, jwtErr := s.tokenSvc.IssueUserInfoJWT(ctx, clientID, userInfo)
		if jwtErr != nil {
			// Fallback to JSON on JWT signing error.
			return c.Status(fiber.StatusOK).JSON(userInfo)
		}

		c.Set(fiber.HeaderContentType, MIMEApplicationJWT)

		return c.Status(fiber.StatusOK).SendString(jwtResponse)
	}

	// Default: return JSON response.
	return c.Status(fiber.StatusOK).JSON(userInfo)
}

// addScopeBasedClaims adds optional claims to userInfo based on the granted scopes.
func addScopeBasedClaims(userInfo map[string]any, scopeList []string, user *cryptoutilIdentityDomain.User) {
	for _, scope := range scopeList {
		switch scope {
		case cryptoutilSharedMagic.ClaimProfile:
			userInfo[cryptoutilSharedMagic.ClaimName] = user.Name
			userInfo[cryptoutilSharedMagic.ClaimGivenName] = user.GivenName
			userInfo[cryptoutilSharedMagic.ClaimFamilyName] = user.FamilyName
			userInfo[cryptoutilSharedMagic.ClaimMiddleName] = user.MiddleName
			userInfo[cryptoutilSharedMagic.ClaimNickname] = user.Nickname
			userInfo[cryptoutilSharedMagic.ClaimPreferredUsername] = user.PreferredUsername
			userInfo[cryptoutilSharedMagic.ClaimProfile] = user.Profile
			userInfo[cryptoutilSharedMagic.ClaimPicture] = user.Picture
			userInfo[cryptoutilSharedMagic.ClaimWebsite] = user.Website
			userInfo[cryptoutilSharedMagic.ClaimGender] = user.Gender
			userInfo[cryptoutilSharedMagic.ClaimBirthdate] = user.Birthdate
			userInfo[cryptoutilSharedMagic.ClaimZoneinfo] = user.Zoneinfo
			userInfo[cryptoutilSharedMagic.ClaimLocale] = user.Locale
			userInfo[cryptoutilSharedMagic.ClaimUpdatedAt] = user.UpdatedAt.Unix()

		case cryptoutilSharedMagic.ClaimEmail:
			userInfo[cryptoutilSharedMagic.ClaimEmail] = user.Email
			userInfo[cryptoutilSharedMagic.ClaimEmailVerified] = user.EmailVerified

		case cryptoutilSharedMagic.ClaimAddress:
			if user.Address != nil {
				userInfo[cryptoutilSharedMagic.ClaimAddress] = map[string]any{
					cryptoutilSharedMagic.AddressFormatted:     user.Address.Formatted,
					cryptoutilSharedMagic.AddressStreetAddress: user.Address.StreetAddress,
					cryptoutilSharedMagic.AddressLocality:      user.Address.Locality,
					cryptoutilSharedMagic.AddressRegion:        user.Address.Region,
					cryptoutilSharedMagic.AddressPostalCode:    user.Address.PostalCode,
					cryptoutilSharedMagic.AddressCountry:       user.Address.Country,
				}
			}

		case cryptoutilSharedMagic.ScopePhone:
			userInfo[cryptoutilSharedMagic.ClaimPhoneNumber] = user.PhoneNumber
			userInfo[cryptoutilSharedMagic.ClaimPhoneVerified] = user.PhoneVerified
		}
	}
}
