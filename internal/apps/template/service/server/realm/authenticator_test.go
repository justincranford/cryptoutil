// Copyright (c) 2025 Justin Cranford
//
//

package realm

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Test UUIDs generated once per test run for consistency.
var (
	testRealmID1  = googleUuid.Must(googleUuid.NewV7()).String()
	testRealmID2  = googleUuid.Must(googleUuid.NewV7()).String()
	testUserID1   = googleUuid.Must(googleUuid.NewV7()).String()
	testUserID2   = googleUuid.Must(googleUuid.NewV7()).String()
	nonExistentID = googleUuid.Must(googleUuid.NewV7()).String()
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
						ID:      testRealmID1,
						Name:    "test-realm",
						Type:    RealmTypeFile,
						Enabled: true,
						Users: []UserConfig{
							{
								ID:           testUserID1,
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

	// Generate unique passwords for this test run.
	adminPassword := googleUuid.Must(googleUuid.NewV7()).String()
	disabledUserPassword := googleUuid.Must(googleUuid.NewV7()).String()
	wrongPassword := googleUuid.Must(googleUuid.NewV7()).String()

	config := &Config{
		Realms: []RealmConfig{
			{
				ID:      testRealmID1,
				Name:    "test-realm",
				Type:    RealmTypeFile,
				Enabled: true,
				Users: []UserConfig{
					{
						ID:           testUserID1,
						Username:     "admin",
						PasswordHash: createTestPasswordHash(t, adminPassword),
						Roles:        []string{"admin"},
						Enabled:      true,
					},
					{
						ID:           testUserID2,
						Username:     "disabled_user",
						PasswordHash: createTestPasswordHash(t, disabledUserPassword),
						Roles:        []string{"user"},
						Enabled:      false,
					},
				},
				Roles: []RoleConfig{
					{
						Name:        "admin",
						Permissions: []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite, "delete"},
					},
					{
						Name:        "user",
						Permissions: []string{cryptoutilSharedMagic.ScopeRead},
					},
				},
			},
			{
				ID:      testRealmID2,
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
			realmID:     testRealmID1,
			username:    "admin",
			password:    adminPassword,
			wantAuth:    true,
			wantErrCode: AuthErrorNone,
		},
		{
			name:        "wrong password",
			realmID:     testRealmID1,
			username:    "admin",
			password:    wrongPassword,
			wantAuth:    false,
			wantErrCode: AuthErrorPasswordMismatch,
		},
		{
			name:        "user not found",
			realmID:     testRealmID1,
			username:    "nonexistent",
			password:    googleUuid.Must(googleUuid.NewV7()).String(),
			wantAuth:    false,
			wantErrCode: AuthErrorUserNotFound,
		},
		{
			name:        "disabled user",
			realmID:     testRealmID1,
			username:    "disabled_user",
			password:    disabledUserPassword,
			wantAuth:    false,
			wantErrCode: AuthErrorUserDisabled,
		},
		{
			name:        "realm not found",
			realmID:     nonExistentID,
			username:    "admin",
			password:    googleUuid.Must(googleUuid.NewV7()).String(),
			wantAuth:    false,
			wantErrCode: AuthErrorRealmNotFound,
		},
		{
			name:        "disabled realm",
			realmID:     testRealmID2,
			username:    "admin",
			password:    googleUuid.Must(googleUuid.NewV7()).String(),
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

	// Generate unique password for this test run.
	testuserPassword := googleUuid.Must(googleUuid.NewV7()).String()

	config := &Config{
		Realms: []RealmConfig{
			{
				ID:      testRealmID1,
				Name:    "demo-realm",
				Type:    RealmTypeFile,
				Enabled: true,
				Users: []UserConfig{
					{
						ID:           testUserID1,
						Username:     "testuser",
						PasswordHash: createTestPasswordHash(t, testuserPassword),
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
	result := auth.AuthenticateByRealmName(ctx, "demo-realm", "testuser", testuserPassword)
	require.True(t, result.Authenticated)
	require.Equal(t, "testuser", result.Username)

	// Realm not found.
	result = auth.AuthenticateByRealmName(ctx, "nonexistent-realm", "testuser", googleUuid.Must(googleUuid.NewV7()).String())
	require.False(t, result.Authenticated)
	require.Equal(t, AuthErrorRealmNotFound, result.ErrorCode)
}

func TestAuthenticator_ExpandPermissions(t *testing.T) {
	t.Parallel()

	// Generate unique password for this test run.
	testPassword := googleUuid.Must(googleUuid.NewV7()).String()

	config := &Config{
		Realms: []RealmConfig{
			{
				ID:      testRealmID1,
				Name:    "test-realm",
				Type:    RealmTypeFile,
				Enabled: true,
				Users: []UserConfig{
					{
						ID:           testUserID1,
						Username:     "superadmin",
						PasswordHash: createTestPasswordHash(t, testPassword),
						Roles:        []string{"superadmin"},
						Enabled:      true,
					},
				},
				Roles: []RoleConfig{
					{
						Name:        "reader",
						Permissions: []string{cryptoutilSharedMagic.ScopeRead},
					},
					{
						Name:        "writer",
						Permissions: []string{cryptoutilSharedMagic.ScopeWrite},
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

	result := auth.Authenticate(ctx, testRealmID1, "superadmin", testPassword)
	require.True(t, result.Authenticated)

	// Should have all inherited permissions.
	require.Contains(t, result.Permissions, "manage_users")
	require.Contains(t, result.Permissions, "delete")
	require.Contains(t, result.Permissions, cryptoutilSharedMagic.ScopeWrite)
	require.Contains(t, result.Permissions, cryptoutilSharedMagic.ScopeRead)
}

func TestAuthenticator_GetRealm(t *testing.T) {
	t.Parallel()

	config := &Config{
		Realms: []RealmConfig{
			{
				ID:      testRealmID1,
				Name:    "test-realm",
				Type:    RealmTypeFile,
				Enabled: true,
			},
		},
	}

	auth, err := NewAuthenticator(config)
	require.NoError(t, err)

	// Found.
	realm, ok := auth.GetRealm(testRealmID1)
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
				ID:      testRealmID1,
				Name:    "realm1",
				Type:    RealmTypeFile,
				Enabled: true,
			},
			{
				ID:      testUserID1,
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
						ID:      testRealmID1,
						Name:    "test-realm",
						Type:    tc.realmType,
						Enabled: true,
					},
				},
			}

			auth, err := NewAuthenticator(config)
			require.NoError(t, err)

			result := auth.Authenticate(ctx, testRealmID1, "user", "pass")
			require.False(t, result.Authenticated)
			require.Equal(t, AuthErrorRealmNotFound, result.ErrorCode)
		})
	}
}
