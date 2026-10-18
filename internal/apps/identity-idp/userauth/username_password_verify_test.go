// Copyright (c) 2025-2026 Justin Cranford.
//
//

package userauth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity-idp/userauth"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedCryptoPassword "cryptoutil/internal/shared/crypto/password"
)

func TestUsernamePasswordAuthenticator_VerifyAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	credStore := newMockPasswordCredentialStore()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	userStore := newMockUserStore()

	// Create a test user.
	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	testUser := &cryptoutilIdentityDomain.User{
		ID:      userID,
		Enabled: true,
		Locked:  false,
	}
	userStore.AddUser(testUser)

	auth := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credStore, challengeStore, userStore, nil, false)

	// Hash password with PBKDF2 (FIPS-compliant).
	password := "SecurePassword123!"
	hash, err := cryptoutilSharedCryptoPassword.HashPassword(password)
	require.NoError(t, err, "HashPassword should succeed")

	err = credStore.StoreCredential(ctx, userID.String(), []byte(hash))
	require.NoError(t, err, "StoreCredential should succeed")

	// Initiate auth.
	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.NoError(t, err, "InitiateAuth should succeed")

	// Verify auth with correct password.
	user, err := auth.VerifyAuth(ctx, challenge.ID, password)
	require.NoError(t, err, "VerifyAuth should succeed with correct password")
	require.NotNil(t, user, "User should not be nil")
	require.Equal(t, userID, user.ID, "User ID should match")
}

// TestUsernamePasswordAuthenticator_VerifyAuthWrongPassword tests VerifyAuth with wrong password.
func TestUsernamePasswordAuthenticator_VerifyAuthWrongPassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	credStore := newMockPasswordCredentialStore()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	userStore := newMockUserStore()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	testUser := &cryptoutilIdentityDomain.User{
		ID:      userID,
		Enabled: true,
		Locked:  false,
	}
	userStore.AddUser(testUser)

	auth := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credStore, challengeStore, userStore, nil, false)

	// Hash password with PBKDF2 (FIPS-compliant).
	hash, err := cryptoutilSharedCryptoPassword.HashPassword(testCorrectPassword)
	require.NoError(t, err, "HashPassword should succeed")

	err = credStore.StoreCredential(ctx, userID.String(), []byte(hash))
	require.NoError(t, err, "StoreCredential should succeed")

	// Initiate auth.
	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.NoError(t, err, "InitiateAuth should succeed")

	// Verify auth with wrong password.
	user, err := auth.VerifyAuth(ctx, challenge.ID, "WrongPassword123!")
	require.Error(t, err, "VerifyAuth should fail with wrong password")
	require.Nil(t, user, "User should be nil on error")
	require.Contains(t, err.Error(), "invalid password", "Error should indicate invalid password")
}

// TestUsernamePasswordAuthenticator_VerifyAuthEmptyPassword tests VerifyAuth with empty password.
func TestUsernamePasswordAuthenticator_VerifyAuthEmptyPassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	credStore := newMockPasswordCredentialStore()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	userStore := newMockUserStore()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	testUser := &cryptoutilIdentityDomain.User{
		ID:      userID,
		Enabled: true,
		Locked:  false,
	}
	userStore.AddUser(testUser)

	auth := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credStore, challengeStore, userStore, nil, false)

	// Hash and store password.
	hash, err := auth.HashPassword(testValidPassword123)
	require.NoError(t, err, "HashPassword should succeed")

	err = credStore.StoreCredential(ctx, userID.String(), hash)
	require.NoError(t, err, "StoreCredential should succeed")

	// Initiate auth.
	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.NoError(t, err, "InitiateAuth should succeed")

	// Verify auth with empty password.
	user, err := auth.VerifyAuth(ctx, challenge.ID, "")
	require.Error(t, err, "VerifyAuth should fail with empty password")
	require.Nil(t, user, "User should be nil on error")
	require.Contains(t, err.Error(), "password is required", "Error should indicate password required")
}

func TestUsernamePasswordAuthenticator_VerifyAuthErrorPaths(t *testing.T) {
	t.Parallel()

	validHash, err := cryptoutilSharedCryptoPassword.HashPassword(testOldPassword)
	require.NoError(t, err)

	tests := []struct {
		name               string
		expiration         time.Time
		retrieveErr        error
		userErr            error
		userEnabled        bool
		userLocked         bool
		credentialHash     []byte
		credentialErr      error
		challengeDeleteErr error
	}{
		{name: "challenge retrieval fails", retrieveErr: errors.New("injected challenge retrieval failure")},
		{name: "challenge is expired", expiration: time.Now().UTC().Add(-time.Hour)},
		{name: "user retrieval fails", userErr: errors.New("injected user retrieval failure")},
		{name: "user is disabled"},
		{name: "user is locked", userEnabled: true, userLocked: true},
		{name: "credential retrieval fails", userEnabled: true, credentialErr: errors.New("injected credential retrieval failure")},
		{name: "VerifyAuth credential hash is malformed", userEnabled: true, credentialHash: []byte(testOldPassword)},
		{name: "challenge deletion fails", userEnabled: true, challengeDeleteErr: errors.New("injected challenge deletion failure")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			userID := googleUuid.Must(googleUuid.NewV7())
			challengeID := googleUuid.Must(googleUuid.NewV7())

			expiresAt := tc.expiration
			if expiresAt.IsZero() {
				expiresAt = time.Now().UTC().Add(time.Hour)
			}

			challengeStore := &userauthTestChallengeStoreStub{
				challenge: &cryptoutilIdentityIdpUserauth.AuthChallenge{
					ID:        challengeID,
					UserID:    userID.String(),
					ExpiresAt: expiresAt,
				},
				retrieveErr: tc.retrieveErr,
				deleteErr:   tc.challengeDeleteErr,
			}
			userStore := newMockUserStore()

			userStore.getErr = tc.userErr
			if tc.userErr == nil {
				userStore.AddUser(&cryptoutilIdentityDomain.User{
					ID:      userID,
					Enabled: tc.userEnabled,
					Locked:  tc.userLocked,
				})
			}

			credentialStore := newMockPasswordCredentialStore()
			credentialStore.getErr = tc.credentialErr

			credentialHash := tc.credentialHash
			if credentialHash == nil {
				credentialHash = []byte(validHash)
			}

			require.NoError(t, credentialStore.StoreCredential(ctx, userID.String(), credentialHash))

			authenticator := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credentialStore, challengeStore, userStore, nil, false)
			user, err := authenticator.VerifyAuth(ctx, challengeID, testOldPassword)
			require.Error(t, err)
			require.Nil(t, user)
		})
	}
}

// TestUsernamePasswordAuthenticator_UpdatePassword tests UpdatePassword.
func TestUsernamePasswordAuthenticator_UpdatePassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	credStore := newMockPasswordCredentialStore()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	userStore := newMockUserStore()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	auth := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credStore, challengeStore, userStore, nil, false)

	// Hash and store initial password using PBKDF2 (FIPS-compliant).
	oldPassword := testOldPassword
	hash, err := cryptoutilSharedCryptoPassword.HashPassword(oldPassword)
	require.NoError(t, err, "HashPassword should succeed")

	err = credStore.StoreCredential(ctx, userID.String(), []byte(hash))
	require.NoError(t, err, "StoreCredential should succeed")

	// Update password.
	newPassword := testNewPassword
	err = auth.UpdatePassword(ctx, userID.String(), oldPassword, newPassword)
	require.NoError(t, err, "UpdatePassword should succeed")
}

func TestUsernamePasswordAuthenticator_UpdatePasswordSamePassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	credStore := newMockPasswordCredentialStore()
	userID := googleUuid.Must(googleUuid.NewV7())
	auth := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credStore, nil, nil, nil, false)

	hash, err := cryptoutilSharedCryptoPassword.HashPassword(testOldPassword)
	require.NoError(t, err)
	require.NoError(t, credStore.StoreCredential(ctx, userID.String(), []byte(hash)))

	err = auth.UpdatePassword(ctx, userID.String(), testOldPassword, testOldPassword)
	require.ErrorContains(t, err, "new password must be different from current password")

	storedHash, err := credStore.GetCredential(ctx, userID.String())
	require.NoError(t, err)
	require.Equal(t, []byte(hash), storedHash)
}

// TestUsernamePasswordAuthenticator_UpdatePasswordWrongOld tests UpdatePassword with wrong old password.
func TestUsernamePasswordAuthenticator_UpdatePasswordWrongOld(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	credStore := newMockPasswordCredentialStore()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	userStore := newMockUserStore()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	auth := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credStore, challengeStore, userStore, nil, false)

	// Hash and store initial password using PBKDF2 (FIPS-compliant).
	hash, err := cryptoutilSharedCryptoPassword.HashPassword("CorrectOldPassword123!")
	require.NoError(t, err, "HashPassword should succeed")

	err = credStore.StoreCredential(ctx, userID.String(), []byte(hash))
	require.NoError(t, err, "StoreCredential should succeed")

	// Try to update password with wrong old password.
	err = auth.UpdatePassword(ctx, userID.String(), "WrongOldPassword123!", testNewPassword)
	require.Error(t, err, "UpdatePassword should fail with wrong old password")
	require.Contains(t, err.Error(), "invalid current password", "Error should indicate invalid current password")
}

// TestUsernamePasswordAuthenticator_UpdatePasswordInvalidNew tests UpdatePassword with invalid new password.
func TestUsernamePasswordAuthenticator_UpdatePasswordInvalidNew(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	credStore := newMockPasswordCredentialStore()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	userStore := newMockUserStore()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	auth := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credStore, challengeStore, userStore, nil, false)

	// Hash and store initial password using PBKDF2 (FIPS-compliant).
	oldPassword := testOldPassword
	hash, err := cryptoutilSharedCryptoPassword.HashPassword(oldPassword)
	require.NoError(t, err, "HashPassword should succeed")

	err = credStore.StoreCredential(ctx, userID.String(), []byte(hash))
	require.NoError(t, err, "StoreCredential should succeed")

	// Try to update password with too short new password.
	err = auth.UpdatePassword(ctx, userID.String(), oldPassword, testShort)
	require.Error(t, err, "UpdatePassword should fail with too short new password")
	require.Contains(t, err.Error(), testPasswordTooShort, "Error should indicate password too short")
}

func TestUsernamePasswordAuthenticator_UpdatePasswordErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		credentialGet error
		credential    []byte
		credentialPut error
	}{
		{name: "credential lookup fails", credentialGet: errors.New("injected credential lookup failure")},
		{name: "UpdatePassword credential hash is malformed", credential: []byte(testOldPassword)},
		{name: "credential update fails", credentialPut: errors.New("injected credential update failure")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			credentialStore := newMockPasswordCredentialStore()
			challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
			userStore := newMockUserStore()
			userID := googleUuid.Must(googleUuid.NewV7())

			storedHash := tc.credential
			if storedHash == nil {
				hash, err := cryptoutilSharedCryptoPassword.HashPassword(testOldPassword)
				require.NoError(t, err)

				storedHash = []byte(hash)
				err = credentialStore.StoreCredential(ctx, userID.String(), storedHash)
				require.NoError(t, err)
			} else {
				require.NoError(t, credentialStore.StoreCredential(ctx, userID.String(), storedHash))
			}

			credentialStore.getErr = tc.credentialGet
			credentialStore.updateErr = tc.credentialPut

			authenticator := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credentialStore, challengeStore, userStore, nil, false)
			err := authenticator.UpdatePassword(ctx, userID.String(), testOldPassword, testNewPassword)
			require.Error(t, err)
		})
	}
}
