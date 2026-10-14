// Copyright (c) 2025-2026 Justin Cranford.
//
//

package apis

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	testify "github.com/stretchr/testify/require"
)

func TestParseScopeDescriptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		scopeStr       string
		expectedCount  int
		expectedScopes []string
	}{
		{
			name:           "single scope",
			scopeStr:       cryptoutilSharedMagic.ScopeOpenID,
			expectedCount:  1,
			expectedScopes: []string{cryptoutilSharedMagic.ScopeOpenID},
		},
		{
			name:           testMultipleScopes,
			scopeStr:       "openid profile email",
			expectedCount:  3,
			expectedScopes: []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimEmail},
		},
		{
			name:           "all standard scopes",
			scopeStr:       "openid profile email address phone offline_access",
			expectedCount:  cryptoutilSharedMagic.DefaultEmailOTPLength,
			expectedScopes: []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ClaimAddress, cryptoutilSharedMagic.ScopePhone, cryptoutilSharedMagic.ScopeOfflineAccess},
		},
		{
			name:           "empty string",
			scopeStr:       "",
			expectedCount:  0,
			expectedScopes: []string{},
		},
		{
			name:           "extra spaces",
			scopeStr:       "openid  profile",
			expectedCount:  2,
			expectedScopes: []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
		},
		{
			name:           testCustomScope,
			scopeStr:       "openid custom_scope",
			expectedCount:  2,
			expectedScopes: []string{cryptoutilSharedMagic.ScopeOpenID, testCustomScopeValue2BF082C3},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			descriptions := parseScopeDescriptions(tc.scopeStr)
			testify.Len(t, descriptions, tc.expectedCount)

			for i, scope := range tc.expectedScopes {
				if i < len(descriptions) {
					testify.Equal(t, scope, descriptions[i].Name)
					testify.NotEmpty(t, descriptions[i].Description)
				}
			}
		})
	}
}

func TestGetScopeDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		scope               string
		expectedDescription string
	}{
		{
			name:                "openid scope",
			scope:               cryptoutilSharedMagic.ScopeOpenID,
			expectedDescription: scopeDescriptionOpenID,
		},
		{
			name:                testProfileScope,
			scope:               cryptoutilSharedMagic.ClaimProfile,
			expectedDescription: "Access your profile information (name, picture, etc.)",
		},
		{
			name:                testEmailScope,
			scope:               cryptoutilSharedMagic.ClaimEmail,
			expectedDescription: scopeDescriptionEmail,
		},
		{
			name:                testAddressScope,
			scope:               cryptoutilSharedMagic.ClaimAddress,
			expectedDescription: scopeDescriptionAddress,
		},
		{
			name:                testPhoneScope,
			scope:               cryptoutilSharedMagic.ScopePhone,
			expectedDescription: "Access your phone number",
		},
		{
			name:                "offline_access scope",
			scope:               cryptoutilSharedMagic.ScopeOfflineAccess,
			expectedDescription: "Maintain access when you're offline (refresh token)",
		},
		{
			name:                testCustomScope,
			scope:               testCustomScopeValue2BF082C3,
			expectedDescription: "Access custom_scope data",
		},
		{
			name:                "unknown scope",
			scope:               "unknown_xyz",
			expectedDescription: "Access unknown_xyz data",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			description := getScopeDescription(tc.scope)
			testify.Equal(t, tc.expectedDescription, description)
		})
	}
}
