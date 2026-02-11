// Copyright (c) 2025 Justin Cranford
//

package realms

import (
	"context"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// mockUserRepository implements UserRepository for testing.
type mockUserRepository struct {
	users             map[string]UserModel
	createErr         error
	findByUsernameErr error
	findByIDErr       error
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]UserModel),
	}
}

func (m *mockUserRepository) Create(_ context.Context, user UserModel) error {
	if m.createErr != nil {
		return m.createErr
	}

	if _, exists := m.users[user.GetUsername()]; exists {
		return fmt.Errorf("duplicate username")
	}

	m.users[user.GetUsername()] = user

	return nil
}

func (m *mockUserRepository) FindByUsername(_ context.Context, username string) (UserModel, error) {
	if m.findByUsernameErr != nil {
		return nil, m.findByUsernameErr
	}

	user, exists := m.users[username]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (m *mockUserRepository) FindByID(_ context.Context, id googleUuid.UUID) (UserModel, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}

	for _, user := range m.users {
		if user.GetID() == id {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

// TestNewUserService tests the NewUserService constructor.
func TestNewUserService(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }

	svc := NewUserService(repo, factory)

	require.NotNil(t, svc)
	require.NotNil(t, svc.userRepo)
	require.NotNil(t, svc.userFactory)
}

// TestRegisterUser_HappyPath tests successful user registration.
func TestRegisterUser_HappyPath(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	ctx := context.Background()
	user, err := svc.RegisterUser(ctx, "testuser", "SecurePass123!")

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "testuser", user.GetUsername())
	require.NotEmpty(t, user.GetPasswordHash())
	require.NotEqual(t, googleUuid.Nil, user.GetID())
}

// TestRegisterUser_EmptyUsername tests username validation.
func TestRegisterUser_EmptyUsername(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	ctx := context.Background()
	user, err := svc.RegisterUser(ctx, "", "SecurePass123!")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "username cannot be empty")
}

// TestRegisterUser_UsernameTooShort tests minimum username length.
func TestRegisterUser_UsernameTooShort(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	ctx := context.Background()
	user, err := svc.RegisterUser(ctx, "ab", "SecurePass123!")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "username must be at least 3 characters")
}

// TestRegisterUser_UsernameTooLong tests maximum username length.
func TestRegisterUser_UsernameTooLong(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	ctx := context.Background()
	longUsername := "a" + string(make([]byte, 50)) // 51 characters
	user, err := svc.RegisterUser(ctx, longUsername, "SecurePass123!")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "username cannot exceed 50 characters")
}

// TestRegisterUser_EmptyPassword tests password validation.
func TestRegisterUser_EmptyPassword(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	ctx := context.Background()
	user, err := svc.RegisterUser(ctx, "testuser", "")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "password cannot be empty")
}

// TestRegisterUser_PasswordTooShort tests minimum password length.
func TestRegisterUser_PasswordTooShort(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	ctx := context.Background()
	user, err := svc.RegisterUser(ctx, "testuser", "short1!")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "password must be at least 8 characters")
}

// TestRegisterUser_DuplicateUsername tests duplicate username handling.
func TestRegisterUser_DuplicateUsername(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	ctx := context.Background()

	// First registration should succeed.
	_, err := svc.RegisterUser(ctx, "testuser", "SecurePass123!")
	require.NoError(t, err)

	// Second registration with same username should fail.
	user, err := svc.RegisterUser(ctx, "testuser", "AnotherPass456!")
	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "failed to create user")
}

// TestRegisterUser_RepositoryError tests repository error handling.
func TestRegisterUser_RepositoryError(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	repo.createErr = fmt.Errorf("database connection failed")
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	ctx := context.Background()
	user, err := svc.RegisterUser(ctx, "testuser", "SecurePass123!")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "failed to create user")
}

// TestAuthenticateUser_HappyPath tests successful authentication.
func TestAuthenticateUser_HappyPath(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	ctx := context.Background()
	password := "SecurePass123!"

	// Register user first.
	_, err := svc.RegisterUser(ctx, "testuser", password)
	require.NoError(t, err)

	// Authenticate with correct credentials.
	user, err := svc.AuthenticateUser(ctx, "testuser", password)

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "testuser", user.GetUsername())
}

// TestAuthenticateUser_UserNotFound tests authentication with non-existent user.
func TestAuthenticateUser_UserNotFound(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	ctx := context.Background()
	user, err := svc.AuthenticateUser(ctx, "nonexistent", "SomePass123!")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "invalid credentials")
}

// TestAuthenticateUser_WrongPassword tests authentication with wrong password.
func TestAuthenticateUser_WrongPassword(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	ctx := context.Background()

	// Register user first.
	_, err := svc.RegisterUser(ctx, "testuser", "CorrectPass123!")
	require.NoError(t, err)

	// Authenticate with wrong password.
	user, err := svc.AuthenticateUser(ctx, "testuser", "WrongPass456!")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "invalid credentials")
}

// TestAuthenticateUser_RepositoryError tests repository error handling.
func TestAuthenticateUser_RepositoryError(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	repo.findByUsernameErr = fmt.Errorf("database connection failed")
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	ctx := context.Background()
	user, err := svc.AuthenticateUser(ctx, "testuser", "SomePass123!")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "invalid credentials")
}

// TestBasicUser_Getters tests BasicUser getter methods.
func TestBasicUser_Getters(t *testing.T) {
	t.Parallel()

	id := googleUuid.Must(googleUuid.NewV7())
	user := &BasicUser{
		ID:           id,
		Username:     "testuser",
		PasswordHash: "hashedpassword123",
	}

	require.Equal(t, id, user.GetID())
	require.Equal(t, "testuser", user.GetUsername())
	require.Equal(t, "hashedpassword123", user.GetPasswordHash())
}

// TestBasicUser_Setters tests BasicUser setter methods.
func TestBasicUser_Setters(t *testing.T) {
	t.Parallel()

	user := &BasicUser{}
	id := googleUuid.Must(googleUuid.NewV7())

	user.SetID(id)
	user.SetUsername("newuser")
	user.SetPasswordHash("newhash456")

	require.Equal(t, id, user.ID)
	require.Equal(t, "newuser", user.Username)
	require.Equal(t, "newhash456", user.PasswordHash)
}

// TestBasicUser_ImplementsUserModel verifies BasicUser implements UserModel interface.
func TestBasicUser_ImplementsUserModel(t *testing.T) {
	t.Parallel()

	// Compile-time interface check.
	var _ UserModel = (*BasicUser)(nil)

	// Runtime check.
	user := &BasicUser{}
	require.NotNil(t, user)
}
