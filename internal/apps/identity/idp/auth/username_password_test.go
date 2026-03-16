// Copyright (c) 2025 Justin Cranford
//
//

package auth_test

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityAuth "cryptoutil/internal/apps/identity/idp/auth"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
)

const testPassword = "CorrectPassword123!"

// mockUserRepo implements UserRepository for testing.
type mockUserRepo struct {
	users    map[string]*cryptoutilIdentityDomain.User
	bySubMap map[string]*cryptoutilIdentityDomain.User
	byEmail  map[string]*cryptoutilIdentityDomain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:    make(map[string]*cryptoutilIdentityDomain.User),
		bySubMap: make(map[string]*cryptoutilIdentityDomain.User),
		byEmail:  make(map[string]*cryptoutilIdentityDomain.User),
	}
}

func (m *mockUserRepo) GetByUsername(_ context.Context, username string) (*cryptoutilIdentityDomain.User, error) {
	user, ok := m.users[username]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", username)
	}

	return user, nil
}

func (m *mockUserRepo) GetBySub(_ context.Context, sub string) (*cryptoutilIdentityDomain.User, error) {
	user, ok := m.bySubMap[sub]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", sub)
	}

	return user, nil
}

func (m *mockUserRepo) Create(_ context.Context, _ *cryptoutilIdentityDomain.User) error {
	return nil
}

func (m *mockUserRepo) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*cryptoutilIdentityDomain.User, error) {
	user, ok := m.byEmail[email]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", email)
	}

	return user, nil
}

func (m *mockUserRepo) Update(_ context.Context, _ *cryptoutilIdentityDomain.User) error {
	return nil
}

func (m *mockUserRepo) Delete(_ context.Context, _ googleUuid.UUID) error {
	return nil
}

func (m *mockUserRepo) List(_ context.Context, _, _ int) ([]*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) Count(_ context.Context) (int64, error) {
	return 0, nil
}

func (m *mockUserRepo) AddUser(user *cryptoutilIdentityDomain.User) {
	m.users[user.PreferredUsername] = user
	m.bySubMap[user.Sub] = user
	m.byEmail[user.Email] = user
}

// TestUsernamePasswordProfile_NewProfile tests NewUsernamePasswordProfile.
func TestUsernamePasswordProfile_NewProfile(t *testing.T) {
	t.Parallel()

	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewUsernamePasswordProfile(userRepo)
	require.NotNil(t, profile, "NewUsernamePasswordProfile should return non-nil profile")
}

// TestUsernamePasswordProfile_Name tests Name.
func TestUsernamePasswordProfile_Name(t *testing.T) {
	t.Parallel()

	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewUsernamePasswordProfile(userRepo)
	require.Equal(t, cryptoutilSharedMagic.AuthMethodUsernamePassword, profile.Name(), "Name should return 'username_password'")
}

// TestUsernamePasswordProfile_RequiresMFA tests RequiresMFA.
func TestUsernamePasswordProfile_RequiresMFA(t *testing.T) {
	t.Parallel()

	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewUsernamePasswordProfile(userRepo)
	require.False(t, profile.RequiresMFA(), "RequiresMFA should return false")
}

// TestUsernamePasswordProfile_ValidateCredentials tests ValidateCredentials.
func TestUsernamePasswordProfile_ValidateCredentials(t *testing.T) {
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
				"username": "testuser",
				"password": "SecurePassword123!",
			},
			wantErr: false,
		},
		{
			name: "missing username",
			credentials: map[string]string{
				"password": "SecurePassword123!",
			},
			wantErr:     true,
			errContains: "missing username",
		},
		{
			name: "empty username",
			credentials: map[string]string{
				"username": "",
				"password": "SecurePassword123!",
			},
			wantErr:     true,
			errContains: "missing username",
		},
		{
			name: "missing password",
			credentials: map[string]string{
				"username": "testuser",
			},
			wantErr:     true,
			errContains: "missing password",
		},
		{
			name: "empty password",
			credentials: map[string]string{
				"username": "testuser",
				"password": "",
			},
			wantErr:     true,
			errContains: "missing password",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			profile := cryptoutilIdentityAuth.NewUsernamePasswordProfile(nil)

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

// TestUsernamePasswordProfile_AuthenticateSuccess tests successful authentication.
func TestUsernamePasswordProfile_AuthenticateSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockUserRepo()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	// Create password hash using PBKDF2 (old format for username_password).
	password := "SecurePassword123!"
	hash, err := cryptoutilSharedCryptoHash.HashSecretPBKDF2(password)
	require.NoError(t, err, "HashSecret should succeed")

	// Add user with hashed password.
	user := &cryptoutilIdentityDomain.User{
		ID:                userID,
		Sub:               userID.String(),
		PreferredUsername: "testuser",
		PasswordHash:      hash,
		Enabled:           true,
		Locked:            false,
	}
	userRepo.AddUser(user)

	profile := cryptoutilIdentityAuth.NewUsernamePasswordProfile(userRepo)

	credentials := map[string]string{
		"username": "testuser",
		"password": password,
	}

	authenticatedUser, err := profile.Authenticate(ctx, credentials)
	require.NoError(t, err, "Authenticate should succeed")
	require.NotNil(t, authenticatedUser, "User should not be nil")
	require.Equal(t, userID, authenticatedUser.ID, "User ID should match")
}

// TestUsernamePasswordProfile_AuthenticateMissingUsername tests missing username.
func TestUsernamePasswordProfile_AuthenticateMissingUsername(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewUsernamePasswordProfile(userRepo)

	credentials := map[string]string{
		"password": "SecurePassword123!",
	}

	user, err := profile.Authenticate(ctx, credentials)
	require.Error(t, err, "Authenticate should fail with missing username")
	require.Nil(t, user, "User should be nil on error")
	require.Contains(t, err.Error(), "missing username", "Error should indicate missing username")
}

// TestUsernamePasswordProfile_AuthenticateMissingPassword tests missing password.
func TestUsernamePasswordProfile_AuthenticateMissingPassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewUsernamePasswordProfile(userRepo)

	credentials := map[string]string{
		"username": "testuser",
	}

	user, err := profile.Authenticate(ctx, credentials)
	require.Error(t, err, "Authenticate should fail with missing password")
	require.Nil(t, user, "User should be nil on error")
	require.Contains(t, err.Error(), "missing password", "Error should indicate missing password")
}

// TestUsernamePasswordProfile_AuthenticateUserNotFound tests non-existent user.
func TestUsernamePasswordProfile_AuthenticateUserNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockUserRepo()
	profile := cryptoutilIdentityAuth.NewUsernamePasswordProfile(userRepo)

	credentials := map[string]string{
		"username": "nonexistent",
		"password": "SecurePassword123!",
	}

	user, err := profile.Authenticate(ctx, credentials)
	require.Error(t, err, "Authenticate should fail for non-existent user")
	require.Nil(t, user, "User should be nil on error")
	require.Contains(t, err.Error(), "user lookup failed", "Error should indicate user not found")
}

// TestUsernamePasswordProfile_AuthenticateDisabledUser tests disabled user.
func TestUsernamePasswordProfile_AuthenticateDisabledUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockUserRepo()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	hash, err := cryptoutilSharedCryptoHash.HashSecretPBKDF2("SecurePassword123!")
	require.NoError(t, err, "HashSecret should succeed")

	// Add disabled user.
	user := &cryptoutilIdentityDomain.User{
		ID:                userID,
		Sub:               userID.String(),
		PreferredUsername: "disableduser",
		PasswordHash:      hash,
		Enabled:           false,
		Locked:            false,
	}
	userRepo.AddUser(user)

	profile := cryptoutilIdentityAuth.NewUsernamePasswordProfile(userRepo)

	credentials := map[string]string{
		"username": "disableduser",
		"password": "SecurePassword123!",
	}

	authenticatedUser, err := profile.Authenticate(ctx, credentials)
	require.Error(t, err, "Authenticate should fail for disabled user")
	require.Nil(t, authenticatedUser, "User should be nil on error")
	require.Contains(t, err.Error(), "account disabled", "Error should indicate account disabled")
}

// TestUsernamePasswordProfile_AuthenticateLockedUser tests locked user.
func TestUsernamePasswordProfile_AuthenticateLockedUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockUserRepo()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	hash, err := cryptoutilSharedCryptoHash.HashSecretPBKDF2("SecurePassword123!")
	require.NoError(t, err, "HashSecret should succeed")

	// Add locked user.
	user := &cryptoutilIdentityDomain.User{
		ID:                userID,
		Sub:               userID.String(),
		PreferredUsername: "lockeduser",
		PasswordHash:      hash,
		Enabled:           true,
		Locked:            true,
	}
	userRepo.AddUser(user)

	profile := cryptoutilIdentityAuth.NewUsernamePasswordProfile(userRepo)

	credentials := map[string]string{
		"username": "lockeduser",
		"password": "SecurePassword123!",
	}

	authenticatedUser, err := profile.Authenticate(ctx, credentials)
	require.Error(t, err, "Authenticate should fail for locked user")
	require.Nil(t, authenticatedUser, "User should be nil on error")
	require.Contains(t, err.Error(), "account locked", "Error should indicate account locked")
}

// TestUsernamePasswordProfile_AuthenticateWrongPassword tests wrong password.
func TestUsernamePasswordProfile_AuthenticateWrongPassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockUserRepo()

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	hash, err := cryptoutilSharedCryptoHash.HashSecretPBKDF2(testPassword)
	require.NoError(t, err, "HashSecret should succeed")

	user := &cryptoutilIdentityDomain.User{
		ID:                userID,
		Sub:               userID.String(),
		PreferredUsername: "wrongpwduser",
		PasswordHash:      hash,
		Enabled:           true,
		Locked:            false,
	}
	userRepo.AddUser(user)

	profile := cryptoutilIdentityAuth.NewUsernamePasswordProfile(userRepo)

	credentials := map[string]string{
		"username": "wrongpwduser",
		"password": "WrongPassword123!",
	}

	authenticatedUser, err := profile.Authenticate(ctx, credentials)
	require.Error(t, err, "Authenticate should fail with wrong password")
	require.Nil(t, authenticatedUser, "User should be nil on error")
	require.Contains(t, err.Error(), "invalid password", "Error should indicate invalid password")
}

// TestEmailPasswordProfile_NewProfile tests NewEmailPasswordProfile.
