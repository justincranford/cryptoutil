// Copyright (c) 2025 Justin Cranford
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	json "encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestRealmConfig_GetType tests all 14 config types implement GetType() correctly.
func TestRealmConfig_GetType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   RealmConfig
		expected RealmType
	}{
		// Federated realm types (4).
		{
			name:     "UsernamePasswordConfig",
			config:   &UsernamePasswordConfig{},
			expected: RealmTypeUsernamePassword,
		},
		{
			name:     "LDAPConfig",
			config:   &LDAPConfig{},
			expected: RealmTypeLDAP,
		},
		{
			name:     "OAuth2Config",
			config:   &OAuth2Config{},
			expected: RealmTypeOAuth2,
		},
		{
			name:     "SAMLConfig",
			config:   &SAMLConfig{},
			expected: RealmTypeSAML,
		},
		// Browser realm types (6).
		{
			name:     "JWESessionCookieConfig",
			config:   &JWESessionCookieConfig{},
			expected: RealmTypeJWESessionCookie,
		},
		{
			name:     "JWSSessionCookieConfig",
			config:   &JWSSessionCookieConfig{},
			expected: RealmTypeJWSSessionCookie,
		},
		{
			name:     "OpaqueSessionCookieConfig",
			config:   &OpaqueSessionCookieConfig{},
			expected: RealmTypeOpaqueSessionCookie,
		},
		{
			name:     "BasicUsernamePasswordConfig",
			config:   &BasicUsernamePasswordConfig{},
			expected: RealmTypeBasicUsernamePassword,
		},
		{
			name:     "BearerAPITokenConfig",
			config:   &BearerAPITokenConfig{},
			expected: RealmTypeBearerAPIToken,
		},
		{
			name:     "HTTPSClientCertConfig",
			config:   &HTTPSClientCertConfig{},
			expected: RealmTypeHTTPSClientCert,
		},
		// Service realm types (4).
		{
			name:     "JWESessionTokenConfig",
			config:   &JWESessionTokenConfig{},
			expected: RealmTypeJWESessionToken,
		},
		{
			name:     "JWSSessionTokenConfig",
			config:   &JWSSessionTokenConfig{},
			expected: RealmTypeJWSSessionToken,
		},
		{
			name:     "OpaqueSessionTokenConfig",
			config:   &OpaqueSessionTokenConfig{},
			expected: RealmTypeOpaqueSessionToken,
		},
		{
			name:     "BasicClientIDSecretConfig",
			config:   &BasicClientIDSecretConfig{},
			expected: RealmTypeBasicClientIDSecret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.config.GetType()
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestRealmConfig_Validate_Valid tests valid configurations pass validation.
func TestRealmConfig_Validate_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config RealmConfig
	}{
		// Federated realm types (4).
		{
			name: "UsernamePasswordConfig_Valid",
			config: &UsernamePasswordConfig{
				MinPasswordLength: 8,
				RequireUppercase:  true,
				RequireLowercase:  true,
				RequireDigit:      true,
				RequireSpecial:    false,
			},
		},
		{
			name: "LDAPConfig_Valid",
			config: &LDAPConfig{
				URL:    "ldap://ldap.example.com:389",
				BaseDN: "dc=example,dc=com",
			},
		},
		{
			name: "OAuth2Config_Valid_WithDiscovery",
			config: &OAuth2Config{
				ClientID:     "test-client-id",
				ProviderURL:  "https://auth.example.com",
				UseDiscovery: true,
			},
		},
		{
			name: "OAuth2Config_Valid_WithoutDiscovery",
			config: &OAuth2Config{
				ClientID:     "test-client-id",
				UseDiscovery: false,
				AuthorizeURL: "https://auth.example.com/authorize",
				TokenURL:     "https://auth.example.com/token",
			},
		},
		{
			name: "SAMLConfig_Valid_WithMetadataURL",
			config: &SAMLConfig{
				MetadataURL: "https://idp.example.com/metadata",
				EntityID:    "test-entity-id",
			},
		},
		{
			name: "SAMLConfig_Valid_WithMetadataXML",
			config: &SAMLConfig{
				MetadataXML: "<EntityDescriptor>...</EntityDescriptor>",
				EntityID:    "test-entity-id",
			},
		},
		// Browser realm types (6).
		{
			name: "JWESessionCookieConfig_Valid",
			config: &JWESessionCookieConfig{
				EncryptionAlgorithm:  "dir+A256GCM",
				SessionExpiryMinutes: 15,
			},
		},
		{
			name: "JWSSessionCookieConfig_Valid",
			config: &JWSSessionCookieConfig{
				SigningAlgorithm:     "RS256",
				SessionExpiryMinutes: 15,
			},
		},
		{
			name: "OpaqueSessionCookieConfig_Valid",
			config: &OpaqueSessionCookieConfig{
				TokenLengthBytes:     32,
				SessionExpiryMinutes: 15,
				StorageType:          cryptoutilSharedMagic.RealmStorageTypeDatabase,
			},
		},
		{
			name: "OpaqueSessionCookieConfig_Valid_Redis",
			config: &OpaqueSessionCookieConfig{
				TokenLengthBytes:     32,
				SessionExpiryMinutes: 15,
				StorageType:          cryptoutilSharedMagic.RealmStorageTypeRedis,
			},
		},
		{
			name: "BasicUsernamePasswordConfig_Valid",
			config: &BasicUsernamePasswordConfig{
				MinPasswordLength: 8,
				RequireUppercase:  true,
				RequireLowercase:  true,
				RequireDigit:      true,
				RequireSpecial:    false,
			},
		},
		{
			name: "BearerAPITokenConfig_Valid",
			config: &BearerAPITokenConfig{
				TokenExpiryDays:   30,
				TokenLengthBytes:  64,
				AllowRefreshToken: true,
			},
		},
		{
			name: "HTTPSClientCertConfig_Valid_WithCAs",
			config: &HTTPSClientCertConfig{
				RequireClientCert: true,
				TrustedCAs:        []string{"-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"},
				ValidateOCSP:      true,
				ValidateCRL:       true,
			},
		},
		{
			name: "HTTPSClientCertConfig_Valid_WithoutRequire",
			config: &HTTPSClientCertConfig{
				RequireClientCert: false,
				TrustedCAs:        nil, // Empty is valid when RequireClientCert is false
			},
		},
		// Service realm types (4).
		{
			name: "JWESessionTokenConfig_Valid",
			config: &JWESessionTokenConfig{
				EncryptionAlgorithm: "dir+A256GCM",
				TokenExpiryMinutes:  60,
			},
		},
		{
			name: "JWSSessionTokenConfig_Valid",
			config: &JWSSessionTokenConfig{
				SigningAlgorithm:   "ES256",
				TokenExpiryMinutes: 60,
			},
		},
		{
			name: "OpaqueSessionTokenConfig_Valid",
			config: &OpaqueSessionTokenConfig{
				TokenLengthBytes:   32,
				TokenExpiryMinutes: 60,
				StorageType:        cryptoutilSharedMagic.RealmStorageTypeDatabase,
			},
		},
		{
			name: "BasicClientIDSecretConfig_Valid",
			config: &BasicClientIDSecretConfig{
				MinSecretLength:  32,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     true,
				RequireSpecial:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			require.NoError(t, err)
		})
	}
}

// TestRealmConfig_Validate_Errors tests that invalid configurations return validation errors.
func TestRealmConfig_Validate_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		config         RealmConfig
		expectedSubstr string
	}{
		// UsernamePasswordConfig errors.
		{
			name: "UsernamePasswordConfig_MinPasswordLengthZero",
			config: &UsernamePasswordConfig{
				MinPasswordLength: 0,
			},
			expectedSubstr: "min_password_length must be at least 1",
		},
		// LDAPConfig errors.
		{
			name: "LDAPConfig_MissingURL",
			config: &LDAPConfig{
				BaseDN: "dc=example,dc=com",
			},
			expectedSubstr: "url is required",
		},
		{
			name: "LDAPConfig_MissingBaseDN",
			config: &LDAPConfig{
				URL: "ldap://ldap.example.com:389",
			},
			expectedSubstr: "base_dn is required",
		},
		// OAuth2Config errors.
		{
			name: "OAuth2Config_MissingClientID",
			config: &OAuth2Config{
				UseDiscovery: true,
				ProviderURL:  "https://auth.example.com",
			},
			expectedSubstr: "client_id is required",
		},
		{
			name: "OAuth2Config_MissingProviderURL_WithDiscovery",
			config: &OAuth2Config{
				ClientID:     "test-client-id",
				UseDiscovery: true,
			},
			expectedSubstr: "provider_url is required when use_discovery is true",
		},
		{
			name: "OAuth2Config_MissingURLs_WithoutDiscovery",
			config: &OAuth2Config{
				ClientID:     "test-client-id",
				UseDiscovery: false,
			},
			expectedSubstr: "authorize_url and token_url are required when use_discovery is false",
		},
		// SAMLConfig errors.
		{
			name: "SAMLConfig_MissingMetadata",
			config: &SAMLConfig{
				EntityID: "test-entity-id",
			},
			expectedSubstr: "either metadata_url or metadata_xml is required",
		},
		{
			name: "SAMLConfig_MissingEntityID",
			config: &SAMLConfig{
				MetadataURL: "https://idp.example.com/metadata",
			},
			expectedSubstr: "entity_id is required",
		},
		// JWESessionCookieConfig errors.
		{
			name: "JWESessionCookieConfig_SessionExpiryZero",
			config: &JWESessionCookieConfig{
				SessionExpiryMinutes: 0,
			},
			expectedSubstr: "session_expiry_minutes must be at least 1",
		},
		// JWSSessionCookieConfig errors.
		{
			name: "JWSSessionCookieConfig_SessionExpiryZero",
			config: &JWSSessionCookieConfig{
				SessionExpiryMinutes: 0,
			},
			expectedSubstr: "session_expiry_minutes must be at least 1",
		},
		// OpaqueSessionCookieConfig errors.
		{
			name: "OpaqueSessionCookieConfig_TokenLengthTooSmall",
			config: &OpaqueSessionCookieConfig{
				TokenLengthBytes:     8, // Below minimum
				SessionExpiryMinutes: 15,
				StorageType:          cryptoutilSharedMagic.RealmStorageTypeDatabase,
			},
			expectedSubstr: "token_length_bytes must be at least",
		},
		{
			name: "OpaqueSessionCookieConfig_SessionExpiryZero",
			config: &OpaqueSessionCookieConfig{
				TokenLengthBytes:     32,
				SessionExpiryMinutes: 0,
				StorageType:          cryptoutilSharedMagic.RealmStorageTypeDatabase,
			},
			expectedSubstr: "session_expiry_minutes must be at least 1",
		},
		{
			name: "OpaqueSessionCookieConfig_InvalidStorageType",
			config: &OpaqueSessionCookieConfig{
				TokenLengthBytes:     32,
				SessionExpiryMinutes: 15,
				StorageType:          "invalid",
			},
			expectedSubstr: "storage_type must be",
		},
		// BasicUsernamePasswordConfig errors.
		{
			name: "BasicUsernamePasswordConfig_MinPasswordLengthZero",
			config: &BasicUsernamePasswordConfig{
				MinPasswordLength: 0,
			},
			expectedSubstr: "min_password_length must be at least 1",
		},
		// BearerAPITokenConfig errors.
		{
			name: "BearerAPITokenConfig_TokenExpiryZero",
			config: &BearerAPITokenConfig{
				TokenExpiryDays:  0,
				TokenLengthBytes: 64,
			},
			expectedSubstr: "token_expiry_days must be at least 1",
		},
		{
			name: "BearerAPITokenConfig_TokenLengthTooSmall",
			config: &BearerAPITokenConfig{
				TokenExpiryDays:  30,
				TokenLengthBytes: 16, // Below minimum
			},
			expectedSubstr: "token_length_bytes must be at least",
		},
		// HTTPSClientCertConfig errors.
		{
			name: "HTTPSClientCertConfig_RequireClientCertWithoutCAs",
			config: &HTTPSClientCertConfig{
				RequireClientCert: true,
				TrustedCAs:        nil,
			},
			expectedSubstr: "trusted_cas is required when require_client_cert is true",
		},
		// JWESessionTokenConfig errors.
		{
			name: "JWESessionTokenConfig_TokenExpiryZero",
			config: &JWESessionTokenConfig{
				TokenExpiryMinutes: 0,
			},
			expectedSubstr: "token_expiry_minutes must be at least 1",
		},
		// JWSSessionTokenConfig errors.
		{
			name: "JWSSessionTokenConfig_TokenExpiryZero",
			config: &JWSSessionTokenConfig{
				TokenExpiryMinutes: 0,
			},
			expectedSubstr: "token_expiry_minutes must be at least 1",
		},
		// OpaqueSessionTokenConfig errors.
		{
			name: "OpaqueSessionTokenConfig_TokenLengthTooSmall",
			config: &OpaqueSessionTokenConfig{
				TokenLengthBytes:   8, // Below minimum
				TokenExpiryMinutes: 60,
				StorageType:        cryptoutilSharedMagic.RealmStorageTypeDatabase,
			},
			expectedSubstr: "token_length_bytes must be at least",
		},
		{
			name: "OpaqueSessionTokenConfig_TokenExpiryZero",
			config: &OpaqueSessionTokenConfig{
				TokenLengthBytes:   32,
				TokenExpiryMinutes: 0,
				StorageType:        cryptoutilSharedMagic.RealmStorageTypeDatabase,
			},
			expectedSubstr: "token_expiry_minutes must be at least 1",
		},
		{
			name: "OpaqueSessionTokenConfig_InvalidStorageType",
			config: &OpaqueSessionTokenConfig{
				TokenLengthBytes:   32,
				TokenExpiryMinutes: 60,
				StorageType:        "invalid",
			},
			expectedSubstr: "storage_type must be",
		},
		// BasicClientIDSecretConfig errors.
		{
			name: "BasicClientIDSecretConfig_MinSecretLengthZero",
			config: &BasicClientIDSecretConfig{
				MinSecretLength: 0,
			},
			expectedSubstr: "min_secret_length must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedSubstr)
		})
	}
}

// TestParseRealmConfig_AllTypes tests parsing JSON config for all 14 realm types.
func TestParseRealmConfig_AllTypes(t *testing.T) {
	t.Parallel()

	// Create a RealmServiceImpl for testing private methods.
	svc := &RealmServiceImpl{}

	tests := []struct {
		name         string
		realmType    string
		configJSON   string
		expectedType RealmType
	}{
		// Federated realm types (4).
		{
			name:         "UsernamePassword",
			realmType:    string(RealmTypeUsernamePassword),
			configJSON:   `{"min_password_length":8,"require_uppercase":true}`,
			expectedType: RealmTypeUsernamePassword,
		},
		{
			name:         "LDAP",
			realmType:    string(RealmTypeLDAP),
			configJSON:   `{"url":"ldap://ldap.example.com","base_dn":"dc=example,dc=com"}`,
			expectedType: RealmTypeLDAP,
		},
		{
			name:         "OAuth2",
			realmType:    string(RealmTypeOAuth2),
			configJSON:   `{"client_id":"test","provider_url":"https://auth.example.com","use_discovery":true}`,
			expectedType: RealmTypeOAuth2,
		},
		{
			name:         "SAML",
			realmType:    string(RealmTypeSAML),
			configJSON:   `{"metadata_url":"https://idp.example.com/metadata","entity_id":"test"}`,
			expectedType: RealmTypeSAML,
		},
		// Browser realm types (6).
		{
			name:         "JWESessionCookie",
			realmType:    string(RealmTypeJWESessionCookie),
			configJSON:   `{"encryption_algorithm":"dir+A256GCM","session_expiry_minutes":15}`,
			expectedType: RealmTypeJWESessionCookie,
		},
		{
			name:         "JWSSessionCookie",
			realmType:    string(RealmTypeJWSSessionCookie),
			configJSON:   `{"signing_algorithm":"RS256","session_expiry_minutes":15}`,
			expectedType: RealmTypeJWSSessionCookie,
		},
		{
			name:         "OpaqueSessionCookie",
			realmType:    string(RealmTypeOpaqueSessionCookie),
			configJSON:   `{"token_length_bytes":32,"session_expiry_minutes":15,"storage_type":"database"}`,
			expectedType: RealmTypeOpaqueSessionCookie,
		},
		{
			name:         "BasicUsernamePassword",
			realmType:    string(RealmTypeBasicUsernamePassword),
			configJSON:   `{"min_password_length":8}`,
			expectedType: RealmTypeBasicUsernamePassword,
		},
		{
			name:         "BearerAPIToken",
			realmType:    string(RealmTypeBearerAPIToken),
			configJSON:   `{"token_expiry_days":30,"token_length_bytes":64}`,
			expectedType: RealmTypeBearerAPIToken,
		},
		{
			name:         "HTTPSClientCert",
			realmType:    string(RealmTypeHTTPSClientCert),
			configJSON:   `{"require_client_cert":true,"trusted_cas":["-----BEGIN CERTIFICATE-----"]}`,
			expectedType: RealmTypeHTTPSClientCert,
		},
		// Service realm types (4).
		{
			name:         "JWESessionToken",
			realmType:    string(RealmTypeJWESessionToken),
			configJSON:   `{"encryption_algorithm":"dir+A256GCM","token_expiry_minutes":60}`,
			expectedType: RealmTypeJWESessionToken,
		},
		{
			name:         "JWSSessionToken",
			realmType:    string(RealmTypeJWSSessionToken),
			configJSON:   `{"signing_algorithm":"ES256","token_expiry_minutes":60}`,
			expectedType: RealmTypeJWSSessionToken,
		},
		{
			name:         "OpaqueSessionToken",
			realmType:    string(RealmTypeOpaqueSessionToken),
			configJSON:   `{"token_length_bytes":32,"token_expiry_minutes":60,"storage_type":"database"}`,
			expectedType: RealmTypeOpaqueSessionToken,
		},
		{
			name:         "BasicClientIDSecret",
			realmType:    string(RealmTypeBasicClientIDSecret),
			configJSON:   `{"min_secret_length":32}`,
			expectedType: RealmTypeBasicClientIDSecret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config, err := svc.parseRealmConfig(tt.realmType, tt.configJSON)
			require.NoError(t, err)
			require.NotNil(t, config)
			require.Equal(t, tt.expectedType, config.GetType())
		})
	}
}

// TestParseRealmConfig_EmptyConfig tests that empty config returns nil.
func TestParseRealmConfig_EmptyConfig(t *testing.T) {
	t.Parallel()

	svc := &RealmServiceImpl{}

	config, err := svc.parseRealmConfig(string(RealmTypeUsernamePassword), "")
	require.NoError(t, err)
	require.Nil(t, config)
}

// TestParseRealmConfig_InvalidJSON tests that invalid JSON returns error.
func TestParseRealmConfig_InvalidJSON(t *testing.T) {
	t.Parallel()

	svc := &RealmServiceImpl{}

	config, err := svc.parseRealmConfig(string(RealmTypeUsernamePassword), "not valid json")
	require.Error(t, err)
	require.Nil(t, config)
	require.Contains(t, err.Error(), "failed to parse realm configuration")
}

// TestParseRealmConfig_UnsupportedType tests that unsupported realm type returns error.
func TestParseRealmConfig_UnsupportedType(t *testing.T) {
	t.Parallel()

	svc := &RealmServiceImpl{}

	config, err := svc.parseRealmConfig("unsupported-type", `{"key":"value"}`)
	require.Error(t, err)
	require.Nil(t, config)
	require.Contains(t, err.Error(), "unsupported realm type")
}

// TestValidateRealmType_AllTypes tests validation for all 14 realm types.
func TestValidateRealmType_AllTypes(t *testing.T) {
	t.Parallel()

	svc := &RealmServiceImpl{}

	validTypes := []RealmType{
		// Federated realm types (4).
		RealmTypeUsernamePassword,
		RealmTypeLDAP,
		RealmTypeOAuth2,
		RealmTypeSAML,
		// Browser realm types (6).
		RealmTypeJWESessionCookie,
		RealmTypeJWSSessionCookie,
		RealmTypeOpaqueSessionCookie,
		RealmTypeBasicUsernamePassword,
		RealmTypeBearerAPIToken,
		RealmTypeHTTPSClientCert,
		// Service realm types (4).
		RealmTypeJWESessionToken,
		RealmTypeJWSSessionToken,
		RealmTypeOpaqueSessionToken,
		RealmTypeBasicClientIDSecret,
	}

	for _, realmType := range validTypes {
		t.Run(string(realmType), func(t *testing.T) {
			t.Parallel()

			err := svc.validateRealmType(string(realmType))
			require.NoError(t, err)
		})
	}
}

// TestValidateRealmType_Invalid tests that invalid realm type returns error.
func TestValidateRealmType_Invalid(t *testing.T) {
	t.Parallel()

	svc := &RealmServiceImpl{}

	err := svc.validateRealmType("invalid-type")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported realm type")
}

// TestParseRealmConfig_VerifyConfigValues tests that parsed config has correct values.
func TestParseRealmConfig_VerifyConfigValues(t *testing.T) {
	t.Parallel()

	svc := &RealmServiceImpl{}

	// Test UsernamePasswordConfig with specific values.
	config := &UsernamePasswordConfig{
		MinPasswordLength: 12,
		RequireUppercase:  true,
		RequireLowercase:  true,
		RequireDigit:      true,
		RequireSpecial:    true,
	}

	configJSON, err := json.Marshal(config)
	require.NoError(t, err)

	parsed, err := svc.parseRealmConfig(string(RealmTypeUsernamePassword), string(configJSON))
	require.NoError(t, err)
	require.NotNil(t, parsed)

	upConfig, ok := parsed.(*UsernamePasswordConfig)
	require.True(t, ok)
	require.Equal(t, 12, upConfig.MinPasswordLength)
	require.True(t, upConfig.RequireUppercase)
	require.True(t, upConfig.RequireLowercase)
	require.True(t, upConfig.RequireDigit)
	require.True(t, upConfig.RequireSpecial)
}
