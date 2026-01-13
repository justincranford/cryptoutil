// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"strings"
	"testing"

	cryptoutilTemplateServerRealms "cryptoutil/internal/apps/template/service/server/realms"

	"github.com/stretchr/testify/require"
)

func TestValidatePasswordForRealm_ValidPasswords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
		realm    *cryptoutilTemplateServerRealms.RealmConfig
	}{
		{
			name:     "default realm - valid password with all character types",
			password: "Abc123!@#xyz", // pragma: allowlist secret - Test vector for realm validation.
			realm:    cryptoutilTemplateServerRealms.DefaultRealm(),
		},
		{
			name:     "default realm - minimum length with variety",
			password: "Aa1!Bb2@Cc3#", // pragma: allowlist secret - Test vector for realm validation.
			realm:    cryptoutilTemplateServerRealms.DefaultRealm(),
		},
		{
			name:     "enterprise realm - strong password",
			password: "Enterprise2025!SecurePass", // pragma: allowlist secret - Test vector for realm validation.
			realm:    cryptoutilTemplateServerRealms.EnterpriseRealm(),
		},
		{
			name:     "enterprise realm - exactly 16 chars with variety",
			password: "Entr1se!2025Pasx", // pragma: allowlist secret - Test vector for realm validation.
			realm:    cryptoutilTemplateServerRealms.EnterpriseRealm(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := cryptoutilTemplateServerRealms.ValidatePasswordForRealm(tt.password, tt.realm)
			require.NoError(t, err)
		})
	}
}

func TestValidatePasswordForRealm_InvalidPasswords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		password    string
		realm       *cryptoutilTemplateServerRealms.RealmConfig
		expectedErr string
	}{
		{
			name:        "nil realm",
			password:    "ValidPass123!", // pragma: allowlist secret - Test vector for realm validation.
			realm:       nil,
			expectedErr: "realm configuration is nil",
		},
		{
			name:        "too short for default realm",
			password:    "Abc1!", // pragma: allowlist secret - Test vector for realm validation.
			realm:       cryptoutilTemplateServerRealms.DefaultRealm(),
			expectedErr: "password must be at least 12 characters long",
		},
		{
			name:        "too short for enterprise realm",
			password:    "Abc123!@#xyz", // pragma: allowlist secret - Test vector for realm validation.
			realm:       cryptoutilTemplateServerRealms.EnterpriseRealm(),
			expectedErr: "password must be at least 16 characters long",
		},
		{
			name:        "missing uppercase",
			password:    "abc123!@#xyz", // pragma: allowlist secret - Test vector for realm validation.
			realm:       cryptoutilTemplateServerRealms.DefaultRealm(),
			expectedErr: "password must contain at least one uppercase letter",
		},
		{
			name:        "missing lowercase",
			password:    "ABC123!@#XYZ", // pragma: allowlist secret - Test vector for realm validation.
			realm:       cryptoutilTemplateServerRealms.DefaultRealm(),
			expectedErr: "password must contain at least one lowercase letter",
		},
		{
			name:        "missing digit",
			password:    "Abcdefg!@#xy", // pragma: allowlist secret - Test vector for realm validation.
			realm:       cryptoutilTemplateServerRealms.DefaultRealm(),
			expectedErr: "password must contain at least one digit",
		},
		{
			name:        "missing special character",
			password:    "Abc123456xyz", // pragma: allowlist secret - Test vector for realm validation.
			realm:       cryptoutilTemplateServerRealms.DefaultRealm(),
			expectedErr: "password must contain at least one special character",
		},
		{
			name:        "insufficient unique characters",
			password:    "Aaaa1111!!!!", // pragma: allowlist secret - Test vector for realm validation.
			realm:       cryptoutilTemplateServerRealms.DefaultRealm(),
			expectedErr: "password must contain at least 8 unique characters",
		},
		{
			name:        "too many consecutive repeated characters (default)",
			password:    "Abc1aaaa23!@", // pragma: allowlist secret - Test vector for realm validation.
			realm:       cryptoutilTemplateServerRealms.DefaultRealm(),
			expectedErr: "password must not contain more than 3 consecutive repeated characters",
		},
		{
			name:        "too many consecutive repeated characters (enterprise)",
			password:    "Enterprise2025!aaa", // pragma: allowlist secret - Test vector for realm validation.
			realm:       cryptoutilTemplateServerRealms.EnterpriseRealm(),
			expectedErr: "password must not contain more than 2 consecutive repeated characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := cryptoutilTemplateServerRealms.ValidatePasswordForRealm(tt.password, tt.realm)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestValidateUsernameForRealm(t *testing.T) {
	t.Parallel()

	realm := cryptoutilTemplateServerRealms.DefaultRealm()

	tests := []struct {
		name        string
		username    string
		wantErr     bool
		expectedErr string
	}{
		{
			name:     "valid username - minimum length",
			username: "abc",
			wantErr:  false,
		},
		{
			name:     "valid username - typical length",
			username: "john_doe123",
			wantErr:  false,
		},
		{
			name:     "valid username - maximum length",
			username: strings.Repeat("a", 64),
			wantErr:  false,
		},
		{
			name:        "invalid - too short",
			username:    "ab",
			wantErr:     true,
			expectedErr: "username must be at least 3 characters long",
		},
		{
			name:        "invalid - too long",
			username:    strings.Repeat("a", 65),
			wantErr:     true,
			expectedErr: "username must not exceed 64 characters",
		},
		{
			name:        "invalid - only whitespace",
			username:    "   ",
			wantErr:     true,
			expectedErr: "username must be at least 3 characters long",
		},
		{
			name:        "invalid - empty string",
			username:    "",
			wantErr:     true,
			expectedErr: "username must be at least 3 characters long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := cryptoutilTemplateServerRealms.ValidateUsernameForRealm(tt.username, realm)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultAndEnterpriseRealmDefaults(t *testing.T) {
	t.Parallel()

	defaultRealm := cryptoutilTemplateServerRealms.DefaultRealm()
	enterpriseRealm := cryptoutilTemplateServerRealms.EnterpriseRealm()

	// Verify default realm settings.
	require.Equal(t, 12, defaultRealm.PasswordMinLength, "default realm password min length")
	require.Equal(t, 3, defaultRealm.PasswordMaxRepeatedChars, "default realm max repeated chars")

	// Verify enterprise realm has stricter settings.
	require.Equal(t, 16, enterpriseRealm.PasswordMinLength, "enterprise realm password min length")
	require.Equal(t, 2, enterpriseRealm.PasswordMaxRepeatedChars, "enterprise realm max repeated chars")

	// Enterprise should be stricter than default.
	require.Greater(t, enterpriseRealm.PasswordMinLength, defaultRealm.PasswordMinLength, "enterprise should require longer passwords")
	require.Less(t, enterpriseRealm.PasswordMaxRepeatedChars, defaultRealm.PasswordMaxRepeatedChars, "enterprise should allow fewer repeated chars")
}
