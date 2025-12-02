// Copyright (c) 2025 Justin Cranford
//
//

package realm

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/pbkdf2"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

func TestNewAuthenticator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name:    "empty config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "config with file realm",
			config: &Config{
				Realms: []RealmConfig{
					{
						ID:      "550e8400-e29b-41d4-a716-446655440000",
						Name:    "test-realm",
						Type:    RealmTypeFile,
						Enabled: true,
						Users: []UserConfig{
							{
								ID:           "660e8400-e29b-41d4-a716-446655440001",
								Username:     "testuser",
								PasswordHash: createTestPasswordHash(t, "testpass"),
								Roles:        []string{"user"},
								Enabled:      true,
							},
						},
					},
				},
				Defaults: RealmDefaults{
					PasswordPolicy: DefaultPasswordPolicy(),
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			auth, err := NewAuthenticator(tc.config)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, auth)
			} else {
				require.NoError(t, err)
				require.NotNil(t, auth)
			}
		})
	}
}

func TestAuthenticator_Authenticate(t *testing.T) {
	t.Parallel()

	config := &Config{
		Realms: []RealmConfig{
			{
				ID:      "550e8400-e29b-41d4-a716-446655440000",
				Name:    "test-realm",
				Type:    RealmTypeFile,
				Enabled: true,
				Users: []UserConfig{
					{
						ID:           "660e8400-e29b-41d4-a716-446655440001",
						Username:     "admin",
						PasswordHash: createTestPasswordHash(t, "adminpass"),
						Roles:        []string{"admin"},
						Enabled:      true,
					},
					{
						ID:           "770e8400-e29b-41d4-a716-446655440002",
						Username:     "disabled_user",
						PasswordHash: createTestPasswordHash(t, "password"),
						Roles:        []string{"user"},
						Enabled:      false,
					},
				},
				Roles: []RoleConfig{
					{
						Name:        "admin",
						Permissions: []string{"read", "write", "delete"},
					},
					{
						Name:        "user",
						Permissions: []string{"read"},
					},
				},
			},
			{
				ID:      "880e8400-e29b-41d4-a716-446655440003",
				Name:    "disabled-realm",
				Type:    RealmTypeFile,
				Enabled: false,
			},
		},
		Defaults: RealmDefaults{
			PasswordPolicy: DefaultPasswordPolicy(),
		},
	}

	auth, err := NewAuthenticator(config)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name        string
		realmID     string
		username    string
		password    string
		wantAuth    bool
		wantErrCode AuthErrorCode
	}{
		{
			name:        "successful authentication",
			realmID:     "550e8400-e29b-41d4-a716-446655440000",
			username:    "admin",
			password:    "adminpass",
			wantAuth:    true,
			wantErrCode: AuthErrorNone,
		},
		{
			name:        "wrong password",
			realmID:     "550e8400-e29b-41d4-a716-446655440000",
			username:    "admin",
			password:    "wrongpass",
			wantAuth:    false,
			wantErrCode: AuthErrorPasswordMismatch,
		},
		{
			name:        "user not found",
			realmID:     "550e8400-e29b-41d4-a716-446655440000",
			username:    "nonexistent",
			password:    "password",
			wantAuth:    false,
			wantErrCode: AuthErrorUserNotFound,
		},
		{
			name:        "disabled user",
			realmID:     "550e8400-e29b-41d4-a716-446655440000",
			username:    "disabled_user",
			password:    "password",
			wantAuth:    false,
			wantErrCode: AuthErrorUserDisabled,
		},
		{
			name:        "realm not found",
			realmID:     "990e8400-e29b-41d4-a716-446655440099",
			username:    "admin",
			password:    "adminpass",
			wantAuth:    false,
			wantErrCode: AuthErrorRealmNotFound,
		},
		{
			name:        "disabled realm",
			realmID:     "880e8400-e29b-41d4-a716-446655440003",
			username:    "admin",
			password:    "password",
			wantAuth:    false,
			wantErrCode: AuthErrorRealmDisabled,
		},
		{
			name:        "empty credentials",
			realmID:     "",
			username:    "",
			password:    "",
			wantAuth:    false,
			wantErrCode: AuthErrorInvalidCreds,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := auth.Authenticate(ctx, tc.realmID, tc.username, tc.password)
			require.NotNil(t, result)
			require.Equal(t, tc.wantAuth, result.Authenticated)
			require.Equal(t, tc.wantErrCode, result.ErrorCode)

			if tc.wantAuth {
				require.NotEmpty(t, result.UserID)
				require.Equal(t, tc.username, result.Username)
				require.NotEmpty(t, result.Roles)
			}
		})
	}
}

func TestAuthenticator_AuthenticateByRealmName(t *testing.T) {
	t.Parallel()

	config := &Config{
		Realms: []RealmConfig{
			{
				ID:      "550e8400-e29b-41d4-a716-446655440000",
				Name:    "demo-realm",
				Type:    RealmTypeFile,
				Enabled: true,
				Users: []UserConfig{
					{
						ID:           "660e8400-e29b-41d4-a716-446655440001",
						Username:     "testuser",
						PasswordHash: createTestPasswordHash(t, "testpass"),
						Roles:        []string{"user"},
						Enabled:      true,
					},
				},
			},
		},
		Defaults: RealmDefaults{
			PasswordPolicy: DefaultPasswordPolicy(),
		},
	}

	auth, err := NewAuthenticator(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Success case.
	result := auth.AuthenticateByRealmName(ctx, "demo-realm", "testuser", "testpass")
	require.True(t, result.Authenticated)
	require.Equal(t, "testuser", result.Username)

	// Realm not found.
	result = auth.AuthenticateByRealmName(ctx, "nonexistent-realm", "testuser", "testpass")
	require.False(t, result.Authenticated)
	require.Equal(t, AuthErrorRealmNotFound, result.ErrorCode)
}

func TestAuthenticator_ExpandPermissions(t *testing.T) {
	t.Parallel()

	config := &Config{
		Realms: []RealmConfig{
			{
				ID:      "550e8400-e29b-41d4-a716-446655440000",
				Name:    "test-realm",
				Type:    RealmTypeFile,
				Enabled: true,
				Users: []UserConfig{
					{
						ID:           "660e8400-e29b-41d4-a716-446655440001",
						Username:     "superadmin",
						PasswordHash: createTestPasswordHash(t, "password"),
						Roles:        []string{"superadmin"},
						Enabled:      true,
					},
				},
				Roles: []RoleConfig{
					{
						Name:        "reader",
						Permissions: []string{"read"},
					},
					{
						Name:        "writer",
						Permissions: []string{"write"},
						Inherits:    []string{"reader"},
					},
					{
						Name:        "admin",
						Permissions: []string{"delete"},
						Inherits:    []string{"writer"},
					},
					{
						Name:        "superadmin",
						Permissions: []string{"manage_users"},
						Inherits:    []string{"admin"},
					},
				},
			},
		},
		Defaults: RealmDefaults{
			PasswordPolicy: DefaultPasswordPolicy(),
		},
	}

	auth, err := NewAuthenticator(config)
	require.NoError(t, err)

	ctx := context.Background()

	result := auth.Authenticate(ctx, "550e8400-e29b-41d4-a716-446655440000", "superadmin", "password")
	require.True(t, result.Authenticated)

	// Should have all inherited permissions.
	require.Contains(t, result.Permissions, "manage_users")
	require.Contains(t, result.Permissions, "delete")
	require.Contains(t, result.Permissions, "write")
	require.Contains(t, result.Permissions, "read")
}

func TestAuthenticator_GetRealm(t *testing.T) {
	t.Parallel()

	config := &Config{
		Realms: []RealmConfig{
			{
				ID:      "550e8400-e29b-41d4-a716-446655440000",
				Name:    "test-realm",
				Type:    RealmTypeFile,
				Enabled: true,
			},
		},
	}

	auth, err := NewAuthenticator(config)
	require.NoError(t, err)

	// Found.
	realm, ok := auth.GetRealm("550e8400-e29b-41d4-a716-446655440000")
	require.True(t, ok)
	require.Equal(t, "test-realm", realm.Name)

	// Not found.
	realm, ok = auth.GetRealm("nonexistent")
	require.False(t, ok)
	require.Nil(t, realm)
}

func TestAuthenticator_ListRealms(t *testing.T) {
	t.Parallel()

	config := &Config{
		Realms: []RealmConfig{
			{
				ID:      "550e8400-e29b-41d4-a716-446655440000",
				Name:    "realm1",
				Type:    RealmTypeFile,
				Enabled: true,
			},
			{
				ID:      "660e8400-e29b-41d4-a716-446655440001",
				Name:    "realm2",
				Type:    RealmTypeFile,
				Enabled: true,
			},
		},
	}

	auth, err := NewAuthenticator(config)
	require.NoError(t, err)

	realms := auth.ListRealms()
	require.Len(t, realms, 2)
}

func TestAuthenticator_UnsupportedRealmTypes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name      string
		realmType RealmType
	}{
		{name: "database realm", realmType: RealmTypeDatabase},
		{name: "ldap realm", realmType: RealmTypeLDAP},
		{name: "oidc realm", realmType: RealmTypeOIDC},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := &Config{
				Realms: []RealmConfig{
					{
						ID:      "550e8400-e29b-41d4-a716-446655440000",
						Name:    "test-realm",
						Type:    tc.realmType,
						Enabled: true,
					},
				},
			}

			auth, err := NewAuthenticator(config)
			require.NoError(t, err)

			result := auth.Authenticate(ctx, "550e8400-e29b-41d4-a716-446655440000", "user", "pass")
			require.False(t, result.Authenticated)
			require.Equal(t, AuthErrorRealmNotFound, result.ErrorCode)
		})
	}
}

// createTestPasswordHash creates a PBKDF2-SHA256 password hash for testing.
func createTestPasswordHash(t *testing.T, password string) string {
	t.Helper()

	salt := make([]byte, cryptoutilMagic.PBKDF2DefaultSaltBytes)
	_, err := rand.Read(salt)
	require.NoError(t, err)

	hashFunc := cryptoutilMagic.PBKDF2HashFunction(cryptoutilMagic.PBKDF2DefaultAlgorithm)
	derivedKey := pbkdf2.Key(
		[]byte(password),
		salt,
		cryptoutilMagic.PBKDF2DefaultIterations,
		cryptoutilMagic.PBKDF2DefaultHashBytes,
		hashFunc,
	)

	return "$pbkdf2-sha256$" +
		"600000$" +
		base64.StdEncoding.EncodeToString(salt) + "$" +
		base64.StdEncoding.EncodeToString(derivedKey)
}
