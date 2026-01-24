// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package idp

import (
	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
)

// DiscoveryMetadata represents OIDC discovery metadata per https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata.
type DiscoveryMetadata struct {
	Issuer                                     string   `json:"issuer"`
	AuthorizationEndpoint                      string   `json:"authorization_endpoint"`
	TokenEndpoint                              string   `json:"token_endpoint"`
	UserInfoEndpoint                           string   `json:"userinfo_endpoint"`
	JWKSUri                                    string   `json:"jwks_uri"`
	RegistrationEndpoint                       string   `json:"registration_endpoint,omitempty"`
	ScopesSupported                            []string `json:"scopes_supported"`
	ResponseTypesSupported                     []string `json:"response_types_supported"`
	ResponseModesSupported                     []string `json:"response_modes_supported,omitempty"`
	GrantTypesSupported                        []string `json:"grant_types_supported"`
	ACRValuesSupported                         []string `json:"acr_values_supported,omitempty"`
	SubjectTypesSupported                      []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported           []string `json:"id_token_signing_alg_values_supported"`
	IDTokenEncryptionAlgValuesSupported        []string `json:"id_token_encryption_alg_values_supported,omitempty"`
	IDTokenEncryptionEncValuesSupported        []string `json:"id_token_encryption_enc_values_supported,omitempty"`
	UserInfoSigningAlgValuesSupported          []string `json:"userinfo_signing_alg_values_supported,omitempty"`
	UserInfoEncryptionAlgValuesSupported       []string `json:"userinfo_encryption_alg_values_supported,omitempty"`
	UserInfoEncryptionEncValuesSupported       []string `json:"userinfo_encryption_enc_values_supported,omitempty"`
	RequestObjectSigningAlgValuesSupported     []string `json:"request_object_signing_alg_values_supported,omitempty"`
	RequestObjectEncryptionAlgValuesSupported  []string `json:"request_object_encryption_alg_values_supported,omitempty"`
	RequestObjectEncryptionEncValuesSupported  []string `json:"request_object_encryption_enc_values_supported,omitempty"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported"`
	TokenEndpointAuthSigningAlgValuesSupported []string `json:"token_endpoint_auth_signing_alg_values_supported,omitempty"`
	DisplayValuesSupported                     []string `json:"display_values_supported,omitempty"`
	ClaimTypesSupported                        []string `json:"claim_types_supported,omitempty"`
	ClaimsSupported                            []string `json:"claims_supported,omitempty"`
	ServiceDocumentation                       string   `json:"service_documentation,omitempty"`
	ClaimsLocalesSupported                     []string `json:"claims_locales_supported,omitempty"`
	UILocalesSupported                         []string `json:"ui_locales_supported,omitempty"`
	ClaimsParameterSupported                   bool     `json:"claims_parameter_supported,omitempty"`
	RequestParameterSupported                  bool     `json:"request_parameter_supported,omitempty"`
	RequestURIParameterSupported               bool     `json:"request_uri_parameter_supported,omitempty"`
	RequireRequestURIRegistration              bool     `json:"require_request_uri_registration,omitempty"`
	OPPolicyURI                                string   `json:"op_policy_uri,omitempty"`
	OPTermsOfServiceURI                        string   `json:"op_tos_uri,omitempty"`
	RevocationEndpoint                         string   `json:"revocation_endpoint,omitempty"`
	RevocationEndpointAuthMethodsSupported     []string `json:"revocation_endpoint_auth_methods_supported,omitempty"`
	IntrospectionEndpoint                      string   `json:"introspection_endpoint,omitempty"`
	IntrospectionEndpointAuthMethodsSupported  []string `json:"introspection_endpoint_auth_methods_supported,omitempty"`
	CodeChallengeMethodsSupported              []string `json:"code_challenge_methods_supported,omitempty"`
}

// handleDiscovery handles GET /.well-known/openid-configuration - OIDC discovery endpoint.
func (s *Service) handleDiscovery(c *fiber.Ctx) error {
	const (
		schemeHTTP  = "http"
		schemeHTTPS = "https"
	)

	// Generate base URL from request (protocol + host).
	scheme := schemeHTTPS

	// Check X-Forwarded-Proto header first (standard proxy header).
	forwardedProto := c.Get("X-Forwarded-Proto")
	if forwardedProto == schemeHTTP {
		scheme = schemeHTTP
	} else if c.Protocol() == schemeHTTP {
		scheme = schemeHTTP
	}

	baseURL := scheme + "://" + c.Hostname()

	// Generate unique issuer ID (instance-specific UUID v7).
	issuerID := googleUuid.Must(googleUuid.NewV7()).String()
	issuer := baseURL + "/oidc/v1/" + issuerID

	// Construct OIDC discovery metadata.
	metadata := DiscoveryMetadata{
		// Required fields per OIDC spec.
		Issuer:                issuer,
		AuthorizationEndpoint: baseURL + "/authz/v1/authorize",
		TokenEndpoint:         baseURL + "/authz/v1/token",
		UserInfoEndpoint:      baseURL + "/oidc/v1/userinfo",
		JWKSUri:               baseURL + "/.well-known/jwks.json",

		// Supported scopes (OIDC + custom).
		ScopesSupported: []string{
			"openid",
			"profile",
			"email",
			"address",
			"phone",
			"offline_access",
		},

		// Response types (OAuth 2.1 + OIDC).
		ResponseTypesSupported: []string{
			"code",
		},

		// Response modes.
		ResponseModesSupported: []string{
			"query",
		},

		// Grant types (OAuth 2.1).
		GrantTypesSupported: []string{
			"authorization_code",
			"refresh_token",
			"client_credentials",
		},

		// Subject types (OIDC).
		SubjectTypesSupported: []string{
			"public",
		},

		// ID token signing algorithms (FIPS 140-3 approved).
		IDTokenSigningAlgValuesSupported: []string{
			"RS256",
			"RS384",
			"RS512",
			"ES256",
			"ES384",
			"ES512",
			"EdDSA",
		},

		// Token endpoint authentication methods (OAuth 2.1 + OIDC).
		TokenEndpointAuthMethodsSupported: []string{
			"client_secret_basic",
			"client_secret_post",
			"client_secret_jwt",
			"private_key_jwt",
			"tls_client_auth",
			"self_signed_tls_client_auth",
		},

		// Token endpoint auth signing algorithms (FIPS 140-3 approved).
		TokenEndpointAuthSigningAlgValuesSupported: []string{
			"RS256",
			"RS384",
			"RS512",
			"ES256",
			"ES384",
			"ES512",
			"EdDSA",
		},

		// Claims supported (OIDC standard claims).
		ClaimsSupported: []string{
			"sub",
			"iss",
			"aud",
			"exp",
			"iat",
			"auth_time",
			"nonce",
			"acr",
			"amr",
			"azp",
			"name",
			"given_name",
			"family_name",
			"middle_name",
			"nickname",
			"preferred_username",
			"profile",
			"picture",
			"website",
			"email",
			"email_verified",
			"gender",
			"birthdate",
			"zoneinfo",
			"locale",
			"phone_number",
			"phone_number_verified",
			"address",
			"updated_at",
		},

		// Revocation endpoint (OAuth 2.1).
		RevocationEndpoint: baseURL + "/authz/v1/revoke",
		RevocationEndpointAuthMethodsSupported: []string{
			"client_secret_basic",
			"client_secret_post",
			"client_secret_jwt",
			"private_key_jwt",
			"tls_client_auth",
			"self_signed_tls_client_auth",
		},

		// Introspection endpoint (OAuth 2.1).
		IntrospectionEndpoint: baseURL + "/authz/v1/introspect",
		IntrospectionEndpointAuthMethodsSupported: []string{
			"client_secret_basic",
			"client_secret_post",
			"client_secret_jwt",
			"private_key_jwt",
			"tls_client_auth",
			"self_signed_tls_client_auth",
		},

		// PKCE support (OAuth 2.1 required).
		CodeChallengeMethodsSupported: []string{
			"S256",
		},

		// Optional fields.
		ClaimTypesSupported:          []string{"normal"},
		RequestParameterSupported:    false,
		RequestURIParameterSupported: false,
	}

	return c.JSON(metadata)
}
