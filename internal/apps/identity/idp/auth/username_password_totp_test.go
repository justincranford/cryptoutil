// Copyright (c) 2025 Justin Cranford
//
//

package auth_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuth "cryptoutil/internal/apps/identity/idp/auth"
)

func TestTOTPProfile_NewProfile(t *testing.T) {
	t.Parallel()

	profile := cryptoutilIdentityAuth.NewTOTPProfile(nil)
	require.NotNil(t, profile, "NewTOTPProfile should return non-nil profile")
}

// TestTOTPProfile_Name tests Name.
func TestTOTPProfile_Name(t *testing.T) {
	t.Parallel()

	profile := cryptoutilIdentityAuth.NewTOTPProfile(nil)
	require.Equal(t, cryptoutilSharedMagic.MFATypeTOTP, profile.Name(), "Name should return 'totp'")
}

// TestTOTPProfile_RequiresMFA tests RequiresMFA.
func TestTOTPProfile_RequiresMFA(t *testing.T) {
	t.Parallel()

	profile := cryptoutilIdentityAuth.NewTOTPProfile(nil)
	require.False(t, profile.RequiresMFA(), "RequiresMFA should return false (TOTP is the MFA)")
}

// TestTOTPProfile_Authenticate tests Authenticate.
func TestTOTPProfile_Authenticate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		credentials map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name: "missing user_id",
			credentials: map[string]string{
				"otp_code": "123456",
			},
			wantErr:     true,
			errContains: "missing user_id",
		},
		{
			name: "missing otp_code",
			credentials: map[string]string{
				"user_id": "test-user-id",
			},
			wantErr:     true,
			errContains: "missing otp_code",
		},
		{
			name: "not implemented",
			credentials: map[string]string{
				"user_id":  "test-user-id",
				"otp_code": "123456",
			},
			wantErr:     true,
			errContains: "TOTP validation not implemented",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			profile := cryptoutilIdentityAuth.NewTOTPProfile(nil)

			user, err := profile.Authenticate(context.Background(), tc.credentials)
			require.Error(t, err)
			require.Nil(t, user)
			require.Contains(t, err.Error(), tc.errContains)
		})
	}
}

// TestTOTPProfile_ValidateCredentials tests ValidateCredentials.
func TestTOTPProfile_ValidateCredentials(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		credentials map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid credentials",
			credentials: map[string]string{
				"user_id":  "test-user-id",
				"otp_code": "123456",
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			credentials: map[string]string{
				"otp_code": "123456",
			},
			wantErr:     true,
			errContains: "missing user_id",
		},
		{
			name: "missing otp_code",
			credentials: map[string]string{
				"user_id": "test-user-id",
			},
			wantErr:     true,
			errContains: "missing otp_code",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			profile := cryptoutilIdentityAuth.NewTOTPProfile(nil)

			err := profile.ValidateCredentials(tc.credentials)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestPasskeyProfile_NewProfile tests NewPasskeyProfile.
func TestPasskeyProfile_NewProfile(t *testing.T) {
	t.Parallel()

	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewPasskeyProfile(userRepo, nil)
	require.NotNil(t, profile, "NewPasskeyProfile should return non-nil profile")
}

// TestPasskeyProfile_Name tests Name.
func TestPasskeyProfile_Name(t *testing.T) {
	t.Parallel()

	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewPasskeyProfile(userRepo, nil)
	require.Equal(t, cryptoutilSharedMagic.AMRPasskey, profile.Name(), "Name should return 'passkey'")
}

// TestPasskeyProfile_RequiresMFA tests RequiresMFA.
func TestPasskeyProfile_RequiresMFA(t *testing.T) {
	t.Parallel()

	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewPasskeyProfile(userRepo, nil)
	require.False(t, profile.RequiresMFA(), "RequiresMFA should return false (Passkey replaces MFA)")
}

// TestPasskeyProfile_Authenticate tests Authenticate.
func TestPasskeyProfile_Authenticate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		credentials map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name: "missing credential_id",
			credentials: map[string]string{
				"assertion": "test-assertion",
			},
			wantErr:     true,
			errContains: "missing credential_id",
		},
		{
			name: "missing assertion",
			credentials: map[string]string{
				"credential_id": "test-credential-id",
			},
			wantErr:     true,
			errContains: "missing assertion",
		},
		{
			name: "not implemented",
			credentials: map[string]string{
				"credential_id": "test-credential-id",
				"assertion":     "test-assertion",
			},
			wantErr:     true,
			errContains: "passkey validation not implemented",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			profile := cryptoutilIdentityAuth.NewPasskeyProfile(nil, nil)

			user, err := profile.Authenticate(context.Background(), tc.credentials)
			require.Error(t, err)
			require.Nil(t, user)
			require.Contains(t, err.Error(), tc.errContains)
		})
	}
}

// TestPasskeyProfile_ValidateCredentials tests ValidateCredentials.
func TestPasskeyProfile_ValidateCredentials(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		credentials map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid credentials",
			credentials: map[string]string{
				"credential_id": "test-credential-id",
				"assertion":     "test-assertion",
			},
			wantErr: false,
		},
		{
			name: "missing credential_id",
			credentials: map[string]string{
				"assertion": "test-assertion",
			},
			wantErr:     true,
			errContains: "missing credential_id",
		},
		{
			name: "missing assertion",
			credentials: map[string]string{
				"credential_id": "test-credential-id",
			},
			wantErr:     true,
			errContains: "missing assertion",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			profile := cryptoutilIdentityAuth.NewPasskeyProfile(nil, nil)

			err := profile.ValidateCredentials(tc.credentials)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestProfileRegistry_NewRegistry tests NewProfileRegistry.
func TestProfileRegistry_NewRegistry(t *testing.T) {
	t.Parallel()

	registry := cryptoutilIdentityAuth.NewProfileRegistry()
	require.NotNil(t, registry, "NewProfileRegistry should return non-nil registry")
}

// TestProfileRegistry_RegisterAndGet tests Register and Get.
func TestProfileRegistry_RegisterAndGet(t *testing.T) {
	t.Parallel()

	registry := cryptoutilIdentityAuth.NewProfileRegistry()
	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewUsernamePasswordProfile(userRepo)

	registry.Register(profile)

	retrieved, ok := registry.Get(cryptoutilSharedMagic.AuthMethodUsernamePassword)
	require.True(t, ok, "Get should return true for registered profile")
	require.NotNil(t, retrieved, "Get should return registered profile")
	require.Equal(t, profile.Name(), retrieved.Name(), "Retrieved profile should match")
}

// TestProfileRegistry_GetNotFound tests Get for non-existent profile.
func TestProfileRegistry_GetNotFound(t *testing.T) {
	t.Parallel()

	registry := cryptoutilIdentityAuth.NewProfileRegistry()

	retrieved, ok := registry.Get("nonexistent")
	require.False(t, ok, "Get should return false for non-existent profile")
	require.Nil(t, retrieved, "Get should return nil for non-existent profile")
}

// TestProfileRegistry_List tests List.
func TestProfileRegistry_List(t *testing.T) {
	t.Parallel()

	registry := cryptoutilIdentityAuth.NewProfileRegistry()
	userRepo := newMockUserRepo()

	// Register multiple profiles.
	profile1 := cryptoutilIdentityAuth.NewUsernamePasswordProfile(userRepo)
	profile2 := cryptoutilIdentityAuth.NewEmailPasswordProfile(userRepo)

	registry.Register(profile1)
	registry.Register(profile2)

	list := registry.List()
	require.Len(t, list, 2, "List should return all registered profiles")
}

// ---------------------- OTPService Tests ----------------------

// TestNewOTPService tests OTPService creation.
