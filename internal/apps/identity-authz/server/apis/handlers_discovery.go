// Copyright (c) 2025-2026 Justin Cranford.
//
//

package apis

import (
	"fmt"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
)

// handleOAuthMetadata handles GET /.well-known/oauth-authorization-server.
// Returns OAuth 2.1 Authorization Server Metadata per RFC 8414.
func (s *Service) handleOAuthMetadata(c *fiber.Ctx) error {
	baseURL := s.config.Tokens.Issuer

	metadata := fiber.Map{
		valueIssuer:                            baseURL,
		valueAuthorizationEndpoint:             fmt.Sprintf("%s/oauth2/v1/authorize", baseURL),
		valueTokenEndpoint:                     fmt.Sprintf("%s/oauth2/v1/token", baseURL),
		valueIntrospectionEndpoint:             fmt.Sprintf("%s/oauth2/v1/introspect", baseURL),
		valueRevocationEndpoint:                fmt.Sprintf("%s/oauth2/v1/revoke", baseURL),
		valueJwksURI:                           fmt.Sprintf("%s/oauth2/v1/jwks", baseURL),
		valueGrantTypesSupported:               []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode, cryptoutilSharedMagic.GrantTypeRefreshToken, cryptoutilSharedMagic.GrantTypeClientCredentials},
		valueResponseTypesSupported:            []string{cryptoutilSharedMagic.ResponseTypeCode},
		valueResponseModesSupported:            []string{cryptoutilSharedMagic.ResponseModeQuery, cryptoutilSharedMagic.ResponseModeFragment},
		valueCodeChallengeMethodsSupported:     []string{cryptoutilSharedMagic.PKCEMethodS256},
		valueTokenEndpointAuthMethodsSupported: []string{cryptoutilSharedMagic.ClientAuthMethodSecretPost, cryptoutilSharedMagic.ClientAuthMethodSecretBasic},
		valueScopesSupported:                   []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite},
		valueServiceDocumentation:              fmt.Sprintf("%s/docs", baseURL),
	}

	return c.JSON(metadata) //nolint:wrapcheck // Fiber JSON returns internal error
}

// handleOIDCDiscovery handles GET /.well-known/openid-configuration.
// Returns OpenID Connect Discovery 1.0 metadata.
func (s *Service) handleOIDCDiscovery(c *fiber.Ctx) error {
	baseURL := s.config.Tokens.Issuer

	metadata := fiber.Map{
		valueIssuer:                             baseURL,
		valueAuthorizationEndpoint:              fmt.Sprintf("%s/oauth2/v1/authorize", baseURL),
		valueTokenEndpoint:                      fmt.Sprintf("%s/oauth2/v1/token", baseURL),
		"userinfo_endpoint":                     fmt.Sprintf("%s/oauth2/v1/userinfo", baseURL),
		valueIntrospectionEndpoint:              fmt.Sprintf("%s/oauth2/v1/introspect", baseURL),
		valueRevocationEndpoint:                 fmt.Sprintf("%s/oauth2/v1/revoke", baseURL),
		valueJwksURI:                            fmt.Sprintf("%s/oauth2/v1/jwks", baseURL),
		valueGrantTypesSupported:                []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode, cryptoutilSharedMagic.GrantTypeRefreshToken, cryptoutilSharedMagic.GrantTypeClientCredentials},
		valueResponseTypesSupported:             []string{cryptoutilSharedMagic.ResponseTypeCode, cryptoutilSharedMagic.ParamIDToken, "token id_token"},
		valueResponseModesSupported:             []string{cryptoutilSharedMagic.ResponseModeQuery, cryptoutilSharedMagic.ResponseModeFragment},
		"subject_types_supported":               []string{cryptoutilSharedMagic.SubjectTypePublic},
		"id_token_signing_alg_values_supported": []string{cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilSharedMagic.JoseAlgES256},
		valueCodeChallengeMethodsSupported:      []string{cryptoutilSharedMagic.PKCEMethodS256},
		valueTokenEndpointAuthMethodsSupported:  []string{cryptoutilSharedMagic.ClientAuthMethodSecretPost, cryptoutilSharedMagic.ClientAuthMethodSecretBasic},
		valueScopesSupported:                    []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ClaimAddress, cryptoutilSharedMagic.ScopePhone},
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
		valueServiceDocumentation:          fmt.Sprintf("%s/docs", baseURL),
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
