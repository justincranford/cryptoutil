// Copyright (c) 2025-2026 Justin Cranford.
//
//

package realm

import (
	"context"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

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
						Name:    testTestRealm,
						Type:    RealmTypeFile,
						Enabled: true,
						Users: []UserConfig{
							{
								ID:           testUserID1,
								Username:     testTestuser,
								PasswordHash: createTestPasswordHash(t, "testpass"),
								Roles:        []string{testUser},
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
				Name:    testTestRealm,
				Type:    RealmTypeFile,
				Enabled: true,
				Users: []UserConfig{
					{
						ID:           testUserID1,
						Username:     testAdmin,
						PasswordHash: createTestPasswordHash(t, adminPassword),
						Roles:        []string{testAdmin},
						Enabled:      true,
					},
					{
						ID:           testUserID2,
						Username:     testDisabledUserValue8F2CDEEB,
						PasswordHash: createTestPasswordHash(t, disabledUserPassword),
						Roles:        []string{testUser},
						Enabled:      false,
					},
				},
				Roles: []RoleConfig{
					{
						Name:        testAdmin,
						Permissions: []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite, testDelete},
					},
					{
						Name:        testUser,
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
			name:        testSuccessfulAuthentication,
			realmID:     testRealmID1,
			username:    testAdmin,
			password:    adminPassword,
			wantAuth:    true,
			wantErrCode: AuthErrorNone,
		},
		{
			name:        testWrongPassword,
			realmID:     testRealmID1,
			username:    testAdmin,
			password:    wrongPassword,
			wantAuth:    false,
			wantErrCode: AuthErrorPasswordMismatch,
		},
		{
			name:        errMsgUserNotFound,
			realmID:     testRealmID1,
			username:    testNonexistent,
			password:    googleUuid.Must(googleUuid.NewV7()).String(),
			wantAuth:    false,
			wantErrCode: AuthErrorUserNotFound,
		},
		{
			name:        testDisabledUser,
			realmID:     testRealmID1,
			username:    testDisabledUserValue8F2CDEEB,
			password:    disabledUserPassword,
			wantAuth:    false,
			wantErrCode: AuthErrorUserDisabled,
		},
		{
			name:        valueRealmNotFound,
			realmID:     nonExistentID,
			username:    testAdmin,
			password:    googleUuid.Must(googleUuid.NewV7()).String(),
			wantAuth:    false,
			wantErrCode: AuthErrorRealmNotFound,
		},
		{
			name:        "disabled realm",
			realmID:     testRealmID2,
			username:    testAdmin,
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
				Name:    testTestRealm,
				Type:    RealmTypeFile,
				Enabled: true,
				Users: []UserConfig{
					{
						ID:           testUserID1,
						Username:     testTestuser,
						PasswordHash: createTestPasswordHash(t, testuserPassword),
						Roles:        []string{testUser},
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
	result := auth.AuthenticateByRealmName(ctx, testTestRealm, testTestuser, testuserPassword)
	require.True(t, result.Authenticated)
	require.Equal(t, testTestuser, result.Username)

	// Realm not found.
	result = auth.AuthenticateByRealmName(ctx, "nonexistent-realm", testTestuser, googleUuid.Must(googleUuid.NewV7()).String())
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
				Name:    testTestRealm,
				Type:    RealmTypeFile,
				Enabled: true,
				Users: []UserConfig{
					{
						ID:           testUserID1,
						Username:     testSuperadmin,
						PasswordHash: createTestPasswordHash(t, testPassword),
						Roles:        []string{testSuperadmin},
						Enabled:      true,
					},
				},
				Roles: []RoleConfig{
					{
						Name:        testReader,
						Permissions: []string{cryptoutilSharedMagic.ScopeRead},
					},
					{
						Name:        testWriter,
						Permissions: []string{cryptoutilSharedMagic.ScopeWrite},
						Inherits:    []string{testReader},
					},
					{
						Name:        testAdmin,
						Permissions: []string{testDelete},
						Inherits:    []string{testWriter},
					},
					{
						Name:        testSuperadmin,
						Permissions: []string{"manage_users"},
						Inherits:    []string{testAdmin},
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

	result := auth.Authenticate(ctx, testRealmID1, testSuperadmin, testPassword)
	require.True(t, result.Authenticated)

	// Should have all inherited permissions.
	require.Contains(t, result.Permissions, "manage_users")
	require.Contains(t, result.Permissions, testDelete)
	require.Contains(t, result.Permissions, cryptoutilSharedMagic.ScopeWrite)
	require.Contains(t, result.Permissions, cryptoutilSharedMagic.ScopeRead)
}

func TestAuthenticator_GetRealm(t *testing.T) {
	t.Parallel()

	config := &Config{
		Realms: []RealmConfig{
			{
				ID:      testRealmID1,
				Name:    testTestRealm,
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
	require.Equal(t, testTestRealm, realm.Name)

	// Not found.
	realm, ok = auth.GetRealm(testNonexistent)
	require.False(t, ok)
	require.Nil(t, realm)
}

func TestAuthenticator_ListRealms(t *testing.T) {
	t.Parallel()

	config := &Config{
		Realms: []RealmConfig{
			{
				ID:      testRealmID1,
				Name:    testRealm1Value7BF439FE,
				Type:    RealmTypeFile,
				Enabled: true,
			},
			{
				ID:      testUserID1,
				Name:    testRealm2,
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
						Name:    testTestRealm,
						Type:    tc.realmType,
						Enabled: true,
					},
				},
			}

			auth, err := NewAuthenticator(config)
			require.NoError(t, err)

			result := auth.Authenticate(ctx, testRealmID1, testUser, "pass")
			require.False(t, result.Authenticated)
			require.Equal(t, AuthErrorRealmNotFound, result.ErrorCode)
		})
	}
}
