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

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity/idp/userauth"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
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
