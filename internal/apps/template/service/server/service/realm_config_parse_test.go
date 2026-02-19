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

)

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
