// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"

	fiber "github.com/gofiber/fiber/v2"
)

// handleOAuthMetadata handles GET /.well-known/oauth-authorization-server.
// Returns OAuth 2.1 Authorization Server Metadata per RFC 8414.
func (s *Service) handleOAuthMetadata(c *fiber.Ctx) error {
	baseURL := s.config.Tokens.Issuer

	metadata := fiber.Map{
		"issuer":                                baseURL,
		"authorization_endpoint":                fmt.Sprintf("%s/oauth2/v1/authorize", baseURL),
		"token_endpoint":                        fmt.Sprintf("%s/oauth2/v1/token", baseURL),
		"introspection_endpoint":                fmt.Sprintf("%s/oauth2/v1/introspect", baseURL),
		"revocation_endpoint":                   fmt.Sprintf("%s/oauth2/v1/revoke", baseURL),
		"jwks_uri":                              fmt.Sprintf("%s/oauth2/v1/jwks", baseURL),
		"grant_types_supported":                 []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode, cryptoutilSharedMagic.GrantTypeRefreshToken, cryptoutilSharedMagic.GrantTypeClientCredentials},
		"response_types_supported":              []string{cryptoutilSharedMagic.ResponseTypeCode},
		"response_modes_supported":              []string{cryptoutilSharedMagic.ResponseModeQuery, cryptoutilSharedMagic.ResponseModeFragment},
		"code_challenge_methods_supported":      []string{cryptoutilSharedMagic.PKCEMethodS256},
		"token_endpoint_auth_methods_supported": []string{cryptoutilSharedMagic.ClientAuthMethodSecretPost, cryptoutilSharedMagic.ClientAuthMethodSecretBasic},
		"scopes_supported":                      []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite},
		"service_documentation":                 fmt.Sprintf("%s/docs", baseURL),
	}

	return c.JSON(metadata) //nolint:wrapcheck // Fiber JSON returns internal error
}

// handleOIDCDiscovery handles GET /.well-known/openid-configuration.
// Returns OpenID Connect Discovery 1.0 metadata.
func (s *Service) handleOIDCDiscovery(c *fiber.Ctx) error {
	baseURL := s.config.Tokens.Issuer

	metadata := fiber.Map{
		"issuer":                                baseURL,
		"authorization_endpoint":                fmt.Sprintf("%s/oauth2/v1/authorize", baseURL),
		"token_endpoint":                        fmt.Sprintf("%s/oauth2/v1/token", baseURL),
		"userinfo_endpoint":                     fmt.Sprintf("%s/oauth2/v1/userinfo", baseURL),
		"introspection_endpoint":                fmt.Sprintf("%s/oauth2/v1/introspect", baseURL),
		"revocation_endpoint":                   fmt.Sprintf("%s/oauth2/v1/revoke", baseURL),
		"jwks_uri":                              fmt.Sprintf("%s/oauth2/v1/jwks", baseURL),
		"grant_types_supported":                 []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode, cryptoutilSharedMagic.GrantTypeRefreshToken, cryptoutilSharedMagic.GrantTypeClientCredentials},
		"response_types_supported":              []string{cryptoutilSharedMagic.ResponseTypeCode, cryptoutilSharedMagic.ParamIDToken, "token id_token"},
		"response_modes_supported":              []string{cryptoutilSharedMagic.ResponseModeQuery, cryptoutilSharedMagic.ResponseModeFragment},
		"subject_types_supported":               []string{cryptoutilSharedMagic.SubjectTypePublic},
		"id_token_signing_alg_values_supported": []string{cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilSharedMagic.JoseAlgES256},
		"code_challenge_methods_supported":      []string{cryptoutilSharedMagic.PKCEMethodS256},
		"token_endpoint_auth_methods_supported": []string{cryptoutilSharedMagic.ClientAuthMethodSecretPost, cryptoutilSharedMagic.ClientAuthMethodSecretBasic},
		"scopes_supported":                      []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ClaimAddress, cryptoutilSharedMagic.ScopePhone},
		"claims_supported": []string{
			cryptoutilSharedMagic.ClaimSub, cryptoutilSharedMagic.ClaimIss, cryptoutilSharedMagic.ClaimAud, cryptoutilSharedMagic.ClaimExp, cryptoutilSharedMagic.ClaimIat, cryptoutilSharedMagic.ClaimAuthTime, cryptoutilSharedMagic.ClaimNonce, cryptoutilSharedMagic.ClaimAcr, cryptoutilSharedMagic.ClaimAmr, cryptoutilSharedMagic.ClaimAzp,
			cryptoutilSharedMagic.ClaimName, cryptoutilSharedMagic.ClaimGivenName, cryptoutilSharedMagic.ClaimFamilyName, cryptoutilSharedMagic.ClaimMiddleName, cryptoutilSharedMagic.ClaimNickname, cryptoutilSharedMagic.ClaimPreferredUsername,
			cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimPicture, cryptoutilSharedMagic.ClaimWebsite, cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ClaimEmailVerified, cryptoutilSharedMagic.ClaimGender, cryptoutilSharedMagic.ClaimBirthdate,
			cryptoutilSharedMagic.ClaimZoneinfo, cryptoutilSharedMagic.ClaimLocale, cryptoutilSharedMagic.ClaimPhoneNumber, cryptoutilSharedMagic.ClaimPhoneVerified, cryptoutilSharedMagic.ClaimAddress, cryptoutilSharedMagic.ClaimUpdatedAt,
		},
		"request_parameter_supported":      false,
		"request_uri_parameter_supported":  false,
		"require_request_uri_registration": false,
		"claims_parameter_supported":       false,
		"service_documentation":            fmt.Sprintf("%s/docs", baseURL),
	}

	return c.JSON(metadata) //nolint:wrapcheck // Fiber JSON returns internal error
}

// handleJWKS handles GET /oauth2/v1/jwks.
// Returns JSON Web Key Set containing public signing keys.
func (s *Service) handleJWKS(c *fiber.Ctx) error {
	// Get public keys from token service if available.
	publicKeys := make([]map[string]any, 0)
	if s.tokenSvc != nil {
		publicKeys = s.tokenSvc.GetPublicKeys()
	}

	jwks := fiber.Map{
		"keys": publicKeys,
	}

	return c.JSON(jwks) //nolint:wrapcheck // Fiber JSON returns internal error
}
