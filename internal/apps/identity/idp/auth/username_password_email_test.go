// Copyright (c) 2025 Justin Cranford
//
//

package auth_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityClientAuth "cryptoutil/internal/apps/identity/authz/clientauth"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityAuth "cryptoutil/internal/apps/identity/idp/auth"
)

func TestEmailPasswordProfile_NewProfile(t *testing.T) {
	t.Parallel()

	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewEmailPasswordProfile(userRepo)
	require.NotNil(t, profile, "NewEmailPasswordProfile should return non-nil profile")
}

// TestEmailPasswordProfile_Name tests Name.
func TestEmailPasswordProfile_Name(t *testing.T) {
	t.Parallel()

	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewEmailPasswordProfile(userRepo)
	require.Equal(t, "email_password", profile.Name(), "Name should return 'email_password'")
}

// TestEmailPasswordProfile_RequiresMFA tests RequiresMFA.
func TestEmailPasswordProfile_RequiresMFA(t *testing.T) {
	t.Parallel()

	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewEmailPasswordProfile(userRepo)
	require.True(t, profile.RequiresMFA(), "RequiresMFA should return true")
}

// TestEmailPasswordProfile_AuthenticateMissingEmail tests missing email.
func TestEmailPasswordProfile_AuthenticateMissingEmail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewEmailPasswordProfile(userRepo)

	credentials := map[string]string{
		"password": "SecurePassword123!",
	}

	user, err := profile.Authenticate(ctx, credentials)
	require.Error(t, err, "Authenticate should fail with missing email")
	require.Nil(t, user, "User should be nil on error")
	require.Contains(t, err.Error(), "missing email", "Error should indicate missing email")
}

// TestEmailPasswordProfile_AuthenticateMissingPassword tests missing password.
func TestEmailPasswordProfile_AuthenticateMissingPassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewEmailPasswordProfile(userRepo)

	credentials := map[string]string{
		cryptoutilSharedMagic.ClaimEmail: "test@example.com",
	}

	user, err := profile.Authenticate(ctx, credentials)
	require.Error(t, err, "Authenticate should fail with missing password")
	require.Nil(t, user, "User should be nil on error")
	require.Contains(t, err.Error(), "missing password", "Error should indicate missing password")
}

// TestEmailPasswordProfile_AuthenticateUserNotFound tests user lookup failure.
func TestEmailPasswordProfile_AuthenticateUserNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewEmailPasswordProfile(userRepo)

	credentials := map[string]string{
		cryptoutilSharedMagic.ClaimEmail:    "nonexistent@example.com",
		"password": "SecurePassword123!",
	}

	user, err := profile.Authenticate(ctx, credentials)
	require.Error(t, err, "Authenticate should fail with non-existent user")
	require.Nil(t, user, "User should be nil on error")
	require.Contains(t, err.Error(), "user lookup failed", "Error should indicate user lookup failed")
}

// TestEmailPasswordProfile_AuthenticateInvalidPassword tests password hash compare failure.
func TestEmailPasswordProfile_AuthenticateInvalidPassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockUserRepo()

	// Create user with hashed password.
	hasher := cryptoutilIdentityClientAuth.NewPBKDF2Hasher()
	correctPassword := testPassword
	passwordHash, err := hasher.HashLowEntropyNonDeterministic(correctPassword)
	require.NoError(t, err, "Failed to hash password")

	testUser := &cryptoutilIdentityDomain.User{
		Sub:               googleUuid.NewString(),
		Email:             "testuser@example.com",
		PreferredUsername: "testuser",
		PasswordHash:      passwordHash,
	}
	userRepo.AddUser(testUser)

	profile := cryptoutilIdentityAuth.NewEmailPasswordProfile(userRepo)

	credentials := map[string]string{
		cryptoutilSharedMagic.ClaimEmail:    "testuser@example.com",
		"password": "WrongPassword456!",
	}

	user, err := profile.Authenticate(ctx, credentials)
	require.Error(t, err, "Authenticate should fail with wrong password")
	require.Nil(t, user, "User should be nil on error")
	require.Contains(t, err.Error(), "invalid credentials", "Error should indicate invalid credentials")
}

// TestEmailPasswordProfile_AuthenticateSuccess tests successful authentication.
func TestEmailPasswordProfile_AuthenticateSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockUserRepo()

	// Create user with hashed password.
	hasher := cryptoutilIdentityClientAuth.NewPBKDF2Hasher()
	correctPassword := testPassword
	passwordHash, err := hasher.HashLowEntropyNonDeterministic(correctPassword)
	require.NoError(t, err, "Failed to hash password")

	testUser := &cryptoutilIdentityDomain.User{
		Sub:               googleUuid.NewString(),
		Email:             "testuser@example.com",
		PreferredUsername: "testuser",
		PasswordHash:      passwordHash,
	}
	userRepo.AddUser(testUser)

	profile := cryptoutilIdentityAuth.NewEmailPasswordProfile(userRepo)

	credentials := map[string]string{
		cryptoutilSharedMagic.ClaimEmail:    "testuser@example.com",
		"password": correctPassword,
	}

	user, err := profile.Authenticate(ctx, credentials)
	require.NoError(t, err, "Authenticate should succeed with correct credentials")
	require.NotNil(t, user, "User should not be nil on success")
	require.Equal(t, testUser.Email, user.Email, "Returned user should match test user")
	require.Equal(t, testUser.Sub, user.Sub, "Returned user sub should match")
}

// TestEmailPasswordProfile_ValidateCredentials tests ValidateCredentials.
func TestEmailPasswordProfile_ValidateCredentials(t *testing.T) {
	t.Parallel()

	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewEmailPasswordProfile(userRepo)

	tests := []struct {
		name      string
		creds     map[string]string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid credentials",
			creds:   map[string]string{cryptoutilSharedMagic.ClaimEmail: "test@example.com", "password": "pwd123"},
			wantErr: false,
		},
		{
			name:      "missing email",
			creds:     map[string]string{"password": "pwd123"},
			wantErr:   true,
			errSubstr: "missing email",
		},
		{
			name:      "missing password",
			creds:     map[string]string{cryptoutilSharedMagic.ClaimEmail: "test@example.com"},
			wantErr:   true,
			errSubstr: "missing password",
		},
		{
			name:      "empty email",
			creds:     map[string]string{cryptoutilSharedMagic.ClaimEmail: "", "password": "pwd123"},
			wantErr:   true,
			errSubstr: "missing email",
		},
		{
			name:      "empty password",
			creds:     map[string]string{cryptoutilSharedMagic.ClaimEmail: "test@example.com", "password": ""},
			wantErr:   true,
			errSubstr: "missing password",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := profile.ValidateCredentials(tc.creds)
			if tc.wantErr {
				require.Error(t, err, "ValidateCredentials should fail")
				require.Contains(t, err.Error(), tc.errSubstr, "Error should contain expected substring")
			} else {
				require.NoError(t, err, "ValidateCredentials should succeed")
			}
		})
	}
}

// TestTOTPProfile_NewProfile tests NewTOTPProfile.
