// Copyright (c) 2025 Justin Cranford
//
//

package auth_test

import (
	"context"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityClientAuth "cryptoutil/internal/identity/authz/clientauth"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityAuth "cryptoutil/internal/identity/idp/auth"
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
	require.Equal(t, "username_password", profile.Name(), "Name should return 'username_password'")
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
		"email": "test@example.com",
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
		"email":    "nonexistent@example.com",
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
		"email":    "testuser@example.com",
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
		"email":    "testuser@example.com",
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
			creds:   map[string]string{"email": "test@example.com", "password": "pwd123"},
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
			creds:     map[string]string{"email": "test@example.com"},
			wantErr:   true,
			errSubstr: "missing password",
		},
		{
			name:      "empty email",
			creds:     map[string]string{"email": "", "password": "pwd123"},
			wantErr:   true,
			errSubstr: "missing email",
		},
		{
			name:      "empty password",
			creds:     map[string]string{"email": "test@example.com", "password": ""},
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
func TestTOTPProfile_NewProfile(t *testing.T) {
	t.Parallel()

	profile := cryptoutilIdentityAuth.NewTOTPProfile(nil)
	require.NotNil(t, profile, "NewTOTPProfile should return non-nil profile")
}

// TestTOTPProfile_Name tests Name.
func TestTOTPProfile_Name(t *testing.T) {
	t.Parallel()

	profile := cryptoutilIdentityAuth.NewTOTPProfile(nil)
	require.Equal(t, "totp", profile.Name(), "Name should return 'totp'")
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
	require.Equal(t, "passkey", profile.Name(), "Name should return 'passkey'")
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

	retrieved, ok := registry.Get("username_password")
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
func TestNewOTPService(t *testing.T) {
	t.Parallel()

	service := cryptoutilIdentityAuth.NewOTPService()
	require.NotNil(t, service, "NewOTPService should return non-nil service")
}

// TestOTPService_GenerateOTP tests GenerateOTP (returns not implemented error).
func TestOTPService_GenerateOTP(t *testing.T) {
	t.Parallel()

	service := cryptoutilIdentityAuth.NewOTPService()
	user := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.New(),
		Sub:               "testuser",
		PreferredUsername: "testuser",
		Email:             "test@example.com",
	}

	// Test with email method.
	otp, err := service.GenerateOTP(context.Background(), user, cryptoutilIdentityAuth.OTPMethodEmail)
	require.Error(t, err, "GenerateOTP should return error (not implemented)")
	require.Empty(t, otp, "OTP should be empty on error")

	// Test with SMS method.
	otp, err = service.GenerateOTP(context.Background(), user, cryptoutilIdentityAuth.OTPMethodSMS)
	require.Error(t, err, "GenerateOTP should return error (not implemented)")
	require.Empty(t, otp, "OTP should be empty on error")
}

// TestOTPService_ValidateOTP tests ValidateOTP (returns not implemented error).
func TestOTPService_ValidateOTP(t *testing.T) {
	t.Parallel()

	service := cryptoutilIdentityAuth.NewOTPService()
	user := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.New(),
		Sub:               "testuser",
		PreferredUsername: "testuser",
		Email:             "test@example.com",
	}

	// Test with email method.
	err := service.ValidateOTP(context.Background(), user, "123456", cryptoutilIdentityAuth.OTPMethodEmail)
	require.Error(t, err, "ValidateOTP should return error (not implemented)")

	// Test with SMS method.
	err = service.ValidateOTP(context.Background(), user, "654321", cryptoutilIdentityAuth.OTPMethodSMS)
	require.Error(t, err, "ValidateOTP should return error (not implemented)")
}

// ---------------------- TOTPValidator Tests ----------------------

// mockOTPSecretStore is a mock implementation of OTPSecretStore.
type mockOTPSecretStore struct {
	totpSecret  string
	emailSecret string
	smsSecret   string
	err         error
}

func (m *mockOTPSecretStore) GetTOTPSecret(_ context.Context, _ string) (string, error) {
	if m.err != nil {
		return "", m.err
	}

	return m.totpSecret, nil
}

func (m *mockOTPSecretStore) GetEmailOTPSecret(_ context.Context, _ string) (string, error) {
	if m.err != nil {
		return "", m.err
	}

	return m.emailSecret, nil
}

func (m *mockOTPSecretStore) GetSMSOTPSecret(_ context.Context, _ string) (string, error) {
	if m.err != nil {
		return "", m.err
	}

	return m.smsSecret, nil
}

// TestNewTOTPValidator tests TOTPValidator creation.
func TestNewTOTPValidator(t *testing.T) {
	t.Parallel()

	store := &mockOTPSecretStore{}
	validator := cryptoutilIdentityAuth.NewTOTPValidator(store)
	require.NotNil(t, validator, "NewTOTPValidator should return non-nil validator")
}

// TestTOTPValidator_ValidateTOTP tests TOTP validation.
func TestTOTPValidator_ValidateTOTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		secret    string
		code      string
		storeErr  error
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "invalid code with valid secret",
			secret:    "JBSWY3DPEHPK3PXP", // Base32-encoded secret
			code:      "000000",
			wantValid: false,
			wantErr:   false,
		},
		{
			name:     "store error",
			secret:   "",
			code:     "123456",
			storeErr: fmt.Errorf("store error"),
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := &mockOTPSecretStore{
				totpSecret: tc.secret,
				err:        tc.storeErr,
			}
			validator := cryptoutilIdentityAuth.NewTOTPValidator(store)

			valid, err := validator.ValidateTOTP(context.Background(), "user123", tc.code)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantValid, valid)
			}
		})
	}
}

// TestTOTPValidator_ValidateTOTPWithWindow tests TOTP validation with window.
func TestTOTPValidator_ValidateTOTPWithWindow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		secret     string
		code       string
		windowSize uint
		storeErr   error
		wantErr    bool
	}{
		{
			name:       "invalid code with valid secret",
			secret:     "JBSWY3DPEHPK3PXP",
			code:       "000000",
			windowSize: 1,
			wantErr:    false,
		},
		{
			name:       "store error",
			secret:     "",
			code:       "123456",
			windowSize: 1,
			storeErr:   fmt.Errorf("store error"),
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := &mockOTPSecretStore{
				totpSecret: tc.secret,
				err:        tc.storeErr,
			}
			validator := cryptoutilIdentityAuth.NewTOTPValidator(store)

			_, err := validator.ValidateTOTPWithWindow(context.Background(), "user123", tc.code, tc.windowSize)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestTOTPValidator_ValidateEmailOTP tests email OTP validation.
func TestTOTPValidator_ValidateEmailOTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		secret   string
		code     string
		storeErr error
		wantErr  bool
	}{
		{
			name:    "invalid code with valid secret",
			secret:  "JBSWY3DPEHPK3PXP",
			code:    "000000",
			wantErr: false,
		},
		{
			name:     "store error",
			secret:   "",
			code:     "123456",
			storeErr: fmt.Errorf("store error"),
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := &mockOTPSecretStore{
				emailSecret: tc.secret,
				err:         tc.storeErr,
			}
			validator := cryptoutilIdentityAuth.NewTOTPValidator(store)

			_, err := validator.ValidateEmailOTP(context.Background(), "user123", tc.code)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestTOTPValidator_ValidateSMSOTP tests SMS OTP validation.
func TestTOTPValidator_ValidateSMSOTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		secret   string
		code     string
		storeErr error
		wantErr  bool
	}{
		{
			name:    "invalid code with valid secret",
			secret:  "JBSWY3DPEHPK3PXP",
			code:    "000000",
			wantErr: false,
		},
		{
			name:     "store error",
			secret:   "",
			code:     "123456",
			storeErr: fmt.Errorf("store error"),
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := &mockOTPSecretStore{
				smsSecret: tc.secret,
				err:       tc.storeErr,
			}
			validator := cryptoutilIdentityAuth.NewTOTPValidator(store)

			_, err := validator.ValidateSMSOTP(context.Background(), "user123", tc.code)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestMFAOrchestrator_NewMFAOrchestrator tests NewMFAOrchestrator.
func TestMFAOrchestrator_NewMFAOrchestrator(t *testing.T) {
	t.Parallel()

	orchestrator := cryptoutilIdentityAuth.NewMFAOrchestrator(nil, nil, nil, nil, nil)
	require.NotNil(t, orchestrator, "NewMFAOrchestrator should return non-nil orchestrator")
}
