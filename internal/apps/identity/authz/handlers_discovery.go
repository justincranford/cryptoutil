// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
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
		"grant_types_supported":                 []string{"authorization_code", "refresh_token", "client_credentials"},
		"response_types_supported":              []string{"code"},
		"response_modes_supported":              []string{"query", "fragment"},
		"code_challenge_methods_supported":      []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
		"scopes_supported":                      []string{"openid", "profile", "email", "read", "write"},
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
		"grant_types_supported":                 []string{"authorization_code", "refresh_token", "client_credentials"},
		"response_types_supported":              []string{"code", "id_token", "token id_token"},
		"response_modes_supported":              []string{"query", "fragment"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256", "ES256"},
		"code_challenge_methods_supported":      []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
		"scopes_supported":                      []string{"openid", "profile", "email", "address", "phone"},
		"claims_supported": []string{
			"sub", "iss", "aud", "exp", "iat", "auth_time", "nonce", "acr", "amr", "azp",
			"name", "given_name", "family_name", "middle_name", "nickname", "preferred_username",
			"profile", "picture", "website", "email", "email_verified", "gender", "birthdate",
			"zoneinfo", "locale", "phone_number", "phone_number_verified", "address", "updated_at",
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
