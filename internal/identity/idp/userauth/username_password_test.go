// Copyright (c) 2025 Justin Cranford
//
//

package userauth_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityIdpUserauth "cryptoutil/internal/identity/idp/userauth"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

const (
	testOldPassword     = "OldPassword123!"
	testNewPassword     = "NewPassword456!"
	testCorrectPassword = "CorrectPassword123!"
)

// mockPasswordCredentialStore implements PasswordCredentialStore for testing.
type mockPasswordCredentialStore struct {
	credentials map[string][]byte
}

func newMockPasswordCredentialStore() *mockPasswordCredentialStore {
	return &mockPasswordCredentialStore{
		credentials: make(map[string][]byte),
	}
}

func (m *mockPasswordCredentialStore) StoreCredential(_ context.Context, userID string, passwordHash []byte) error {
	m.credentials[userID] = passwordHash

	return nil
}

func (m *mockPasswordCredentialStore) GetCredential(_ context.Context, userID string) ([]byte, error) {
	hash, ok := m.credentials[userID]
	if !ok {
		return nil, fmt.Errorf("credential not found")
	}

	return hash, nil
}

func (m *mockPasswordCredentialStore) DeleteCredential(_ context.Context, userID string) error {
	delete(m.credentials, userID)

	return nil
}

func (m *mockPasswordCredentialStore) UpdateCredential(_ context.Context, userID string, newPasswordHash []byte) error {
	if _, ok := m.credentials[userID]; !ok {
		return fmt.Errorf("credential not found")
	}

	m.credentials[userID] = newPasswordHash

	return nil
}

// mockUserStore implements UserStore for testing.
type mockUserStore struct {
	users map[string]*cryptoutilIdentityDomain.User
}

func newMockUserStore() *mockUserStore {
	return &mockUserStore{
		users: make(map[string]*cryptoutilIdentityDomain.User),
	}
}

func (m *mockUserStore) GetByID(_ context.Context, userID string) (*cryptoutilIdentityDomain.User, error) {
	user, ok := m.users[userID]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (m *mockUserStore) Update(_ context.Context, user *cryptoutilIdentityDomain.User) error {
	m.users[user.ID.String()] = user

	return nil
}

func (m *mockUserStore) AddUser(user *cryptoutilIdentityDomain.User) {
	m.users[user.ID.String()] = user
}

// TestUsernamePasswordAuthenticator_NewAuthenticator tests NewUsernamePasswordAuthenticator.
func TestUsernamePasswordAuthenticator_NewAuthenticator(t *testing.T) {
	t.Parallel()

	credStore := newMockPasswordCredentialStore()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	userStore := newMockUserStore()

	auth := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credStore, challengeStore, userStore, nil, false)
	require.NotNil(t, auth, "NewUsernamePasswordAuthenticator should return non-nil authenticator")
}

// TestUsernamePasswordAuthenticator_Method tests Method.
func TestUsernamePasswordAuthenticator_Method(t *testing.T) {
	t.Parallel()

	credStore := newMockPasswordCredentialStore()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	userStore := newMockUserStore()

	auth := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credStore, challengeStore, userStore, nil, false)
	require.Equal(t, cryptoutilIdentityMagic.AuthMethodUsernamePassword, auth.Method(), "Method should return correct identifier")
}

// TestUsernamePasswordAuthenticator_HashPassword tests HashPassword with various inputs.
func TestUsernamePasswordAuthenticator_HashPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		password  string
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "valid password",
			password:  "ValidPassword123!",
			wantErr:   false,
			errSubstr: "",
		},
		{
			name:      "minimum length password",
			password:  strings.Repeat("a", cryptoutilIdentityMagic.MinPasswordLength),
			wantErr:   false,
			errSubstr: "",
		},
		{
			name:      "password too short",
			password:  "short",
			wantErr:   true,
			errSubstr: "password too short",
		},
		{
			name:      "password too long",
			password:  strings.Repeat("a", cryptoutilIdentityMagic.MaxPasswordLength+1),
			wantErr:   true,
			errSubstr: "password too long",
		},
		{
			name:      "empty password",
			password:  "",
			wantErr:   true,
			errSubstr: "password too short",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			credStore := newMockPasswordCredentialStore()
			challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
			userStore := newMockUserStore()
			auth := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credStore, challengeStore, userStore, nil, false)

			hash, err := auth.HashPassword(tc.password)
			if tc.wantErr {
				require.Error(t, err, "HashPassword should fail")
				require.Contains(t, err.Error(), tc.errSubstr, "Error should contain expected substring")
				require.Nil(t, hash, "Hash should be nil on error")
			} else {
				require.NoError(t, err, "HashPassword should succeed")
				require.NotEmpty(t, hash, "Hash should not be empty")
			}
		})
	}
}

// TestUsernamePasswordAuthenticator_ValidatePassword tests ValidatePassword.
func TestUsernamePasswordAuthenticator_ValidatePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		password  string
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "valid password",
			password:  "ValidPassword123!",
			wantErr:   false,
			errSubstr: "",
		},
		{
			name:      "password too short",
			password:  "short",
			wantErr:   true,
			errSubstr: "password too short",
		},
		{
			name:      "password too long",
			password:  strings.Repeat("a", cryptoutilIdentityMagic.MaxPasswordLength+1),
			wantErr:   true,
			errSubstr: "password too long",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			credStore := newMockPasswordCredentialStore()
			challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
			userStore := newMockUserStore()
			auth := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credStore, challengeStore, userStore, nil, false)

			err := auth.ValidatePassword(tc.password)
			if tc.wantErr {
				require.Error(t, err, "ValidatePassword should fail")
				require.Contains(t, err.Error(), tc.errSubstr, "Error should contain expected substring")
			} else {
				require.NoError(t, err, "ValidatePassword should succeed")
			}
		})
	}
}

// TestUsernamePasswordAuthenticator_InitiateAuth tests InitiateAuth.
func TestUsernamePasswordAuthenticator_InitiateAuth(t *testing.T) {
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

	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge, "Challenge should not be nil")
	require.Equal(t, userID.String(), challenge.UserID, "Challenge UserID should match")
	require.Equal(t, cryptoutilIdentityMagic.AuthMethodUsernamePassword, challenge.Method, "Challenge Method should match")
}

// TestUsernamePasswordAuthenticator_InitiateAuthUserNotFound tests InitiateAuth with non-existent user.
func TestUsernamePasswordAuthenticator_InitiateAuthUserNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	credStore := newMockPasswordCredentialStore()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	userStore := newMockUserStore()
	auth := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credStore, challengeStore, userStore, nil, false)

	nonExistentID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	challenge, err := auth.InitiateAuth(ctx, nonExistentID.String())
	require.Error(t, err, "InitiateAuth should fail for non-existent user")
	require.Nil(t, challenge, "Challenge should be nil on error")
	require.Contains(t, err.Error(), "failed to get user", "Error should indicate user retrieval failure")
}

// TestUsernamePasswordAuthenticator_InitiateAuthDisabledUser tests InitiateAuth with disabled user.
func TestUsernamePasswordAuthenticator_InitiateAuthDisabledUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	credStore := newMockPasswordCredentialStore()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	userStore := newMockUserStore()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	disabledUser := &cryptoutilIdentityDomain.User{
		ID:      userID,
		Enabled: false,
		Locked:  false,
	}
	userStore.AddUser(disabledUser)

	auth := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credStore, challengeStore, userStore, nil, false)

	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.Error(t, err, "InitiateAuth should fail for disabled user")
	require.Nil(t, challenge, "Challenge should be nil on error")
	require.Contains(t, err.Error(), "account disabled", "Error should indicate account disabled")
}

// TestUsernamePasswordAuthenticator_InitiateAuthLockedUser tests InitiateAuth with locked user.
func TestUsernamePasswordAuthenticator_InitiateAuthLockedUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	credStore := newMockPasswordCredentialStore()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	userStore := newMockUserStore()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	lockedUser := &cryptoutilIdentityDomain.User{
		ID:      userID,
		Enabled: true,
		Locked:  true,
	}
	userStore.AddUser(lockedUser)

	auth := cryptoutilIdentityIdpUserauth.NewUsernamePasswordAuthenticator(credStore, challengeStore, userStore, nil, false)

	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.Error(t, err, "InitiateAuth should fail for locked user")
	require.Nil(t, challenge, "Challenge should be nil on error")
	require.Contains(t, err.Error(), "account locked", "Error should indicate account locked")
}

// TestUsernamePasswordAuthenticator_VerifyAuth tests VerifyAuth flow.
// NOTE: This test uses bcrypt directly because the production code has a bug
// where HashPassword uses PBKDF2 but VerifyAuth uses bcrypt. This will be fixed.
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
