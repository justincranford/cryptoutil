// Copyright (c) 2025 Justin Cranford
//
//

package userauth_test

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity/idp/userauth"
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

	// Hash password with bcrypt directly (production code uses bcrypt in VerifyAuth).
	password := "SecurePassword123!"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "bcrypt.GenerateFromPassword should succeed")

	err = credStore.StoreCredential(ctx, userID.String(), hash)
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

	// Hash password with bcrypt directly.
	hash, err := bcrypt.GenerateFromPassword([]byte(testCorrectPassword), bcrypt.DefaultCost)
	require.NoError(t, err, "bcrypt.GenerateFromPassword should succeed")

	err = credStore.StoreCredential(ctx, userID.String(), hash)
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
	hash, err := auth.HashPassword("ValidPassword123!")
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

// TestUsernamePasswordAuthenticator_UpdatePassword tests UpdatePassword.
// NOTE: Uses bcrypt directly since production code uses bcrypt in verification.
func TestUsernamePasswordAuthenticator_UpdatePassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	credStore := newMockPasswordCredentialStore()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	userStore := newMockUserStore()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	auth := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credStore, challengeStore, userStore, nil, false)

	// Hash and store initial password using bcrypt.
	oldPassword := testOldPassword
	hash, err := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)
	require.NoError(t, err, "bcrypt.GenerateFromPassword should succeed")

	err = credStore.StoreCredential(ctx, userID.String(), hash)
	require.NoError(t, err, "StoreCredential should succeed")

	// Update password.
	newPassword := testNewPassword
	err = auth.UpdatePassword(ctx, userID.String(), oldPassword, newPassword)
	require.NoError(t, err, "UpdatePassword should succeed")
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

	// Hash and store initial password using bcrypt.
	hash, err := bcrypt.GenerateFromPassword([]byte("CorrectOldPassword123!"), bcrypt.DefaultCost)
	require.NoError(t, err, "bcrypt.GenerateFromPassword should succeed")

	err = credStore.StoreCredential(ctx, userID.String(), hash)
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

	// Hash and store initial password using bcrypt.
	oldPassword := testOldPassword
	hash, err := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)
	require.NoError(t, err, "bcrypt.GenerateFromPassword should succeed")

	err = credStore.StoreCredential(ctx, userID.String(), hash)
	require.NoError(t, err, "StoreCredential should succeed")

	// Try to update password with too short new password.
	err = auth.UpdatePassword(ctx, userID.String(), oldPassword, "short")
	require.Error(t, err, "UpdatePassword should fail with too short new password")
	require.Contains(t, err.Error(), "password too short", "Error should indicate password too short")
}
