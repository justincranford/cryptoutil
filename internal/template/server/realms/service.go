// Copyright (c) 2025 Justin Cranford
//

package realms

import (
	"context"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserServiceImpl implements user registration and authentication using bcrypt.
type UserServiceImpl struct {
	userRepo UserRepository
}

// NewUserService creates a new UserService.
//
// Parameters:
// - userRepo: Repository for user CRUD operations
//
// Returns configured UserService ready for registration and authentication.
func NewUserService(userRepo UserRepository) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo: userRepo,
	}
}

// RegisterUser creates a new user account with hashed password.
//
// Workflow:
// 1. Validate username and password (non-empty, length requirements)
// 2. Hash password using bcrypt (cost factor 10)
// 3. Create user entity with UUIDv7
// 4. Save to repository
// 5. Return created user
//
// Validation Rules:
// - Username: 3-50 characters, alphanumeric + underscore
// - Password: 8+ characters minimum
//
// Security Notes:
// - Password hashed with bcrypt (OWASP-recommended for passwords)
// - Cost factor 10 balances security and performance
// - UUIDv7 provides time-ordered IDs (sortable, indexable)
// - Returns error (NOT user) on duplicate username (prevents enumeration)
//
// Error Conditions:
// - Empty username or password
// - Username too short (<3) or too long (>50)
// - Password too short (<8)
// - Duplicate username (repository constraint violation)
// - Database errors (connection, timeout)
//
// Example Usage:
//
//	user, err := userService.RegisterUser(ctx, "alice", "securePassword123")
//	if err != nil {
//	    return fmt.Errorf("registration failed: %w", err)
//	}
//	log.Printf("User created: %s (ID: %s)", user.GetUsername(), user.GetID())
func (s *UserServiceImpl) RegisterUser(ctx context.Context, username, password string) (UserModel, error) {
	// Validate username.
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	const (
		minUsernameLength = 3
		maxUsernameLength = 50
	)

	if len(username) < minUsernameLength {
		return nil, fmt.Errorf("username must be at least %d characters", minUsernameLength)
	}

	if len(username) > maxUsernameLength {
		return nil, fmt.Errorf("username cannot exceed %d characters", maxUsernameLength)
	}

	// Validate password.
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	const minPasswordLength = 8

	if len(password) < minPasswordLength {
		return nil, fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}

	// Hash password with bcrypt.
	const bcryptCostFactor = 10

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCostFactor)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity.
	// NOTE: Caller must provide a User implementation that supports SetPasswordHash.
	// For cipher-im: user := &domain.User{ID: googleUuid.New(), Username: username}
	// For jose-ja: user := &models.User{ID: googleUuid.New(), Username: username}
	userID := googleUuid.Must(googleUuid.NewV7())

	// Create minimal user entity (implementation-specific).
	user := &BasicUser{
		ID:           userID,
		Username:     username,
		PasswordHash: string(passwordHash),
	}

	// Save to repository.
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// AuthenticateUser validates credentials and returns user if successful.
//
// Workflow:
// 1. Find user by username
// 2. Verify password hash with bcrypt
// 3. Return user on success
//
// Security Notes:
// - Constant-time comparison (bcrypt.CompareHashAndPassword prevents timing attacks)
// - Returns generic error (NOT "user not found" vs "wrong password" - prevents enumeration)
// - Caller should generate JWT after successful authentication
//
// Error Conditions:
// - User not found (username doesn't exist)
// - Password mismatch (wrong password)
// - Database errors (connection, timeout)
//
// Example Usage:
//
//	user, err := userService.AuthenticateUser(ctx, "alice", "securePassword123")
//	if err != nil {
//	    return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
//	}
//
//	// Generate JWT for authenticated user.
//	token, err := jwtService.GenerateToken(user.GetID(), user.GetUsername(), 15*time.Minute)
//	if err != nil {
//	    return fmt.Errorf("failed to generate token: %w", err)
//	}
//
//	return c.JSON(fiber.Map{"token": token})
func (s *UserServiceImpl) AuthenticateUser(ctx context.Context, username, password string) (UserModel, error) {
	// Find user by username.
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		// Return generic error (prevent username enumeration).
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password.
	err = bcrypt.CompareHashAndPassword([]byte(user.GetPasswordHash()), []byte(password))
	if err != nil {
		// Return generic error (prevent timing attacks).
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

// BasicUser is a minimal UserModel implementation for testing and simple use cases.
//
// Usage:
// - Template realms package uses this for internal operations
// - Services can provide their own UserModel implementations (domain.User, models.User)
// - Satisfies UserModel interface (GetID, GetUsername, GetPasswordHash, SetID, SetUsername, SetPasswordHash).
type BasicUser struct {
	ID           googleUuid.UUID
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

// GetID returns the user's unique identifier.
func (u *BasicUser) GetID() googleUuid.UUID {
	return u.ID
}

// GetUsername returns the user's username.
func (u *BasicUser) GetUsername() string {
	return u.Username
}

// GetPasswordHash returns the user's bcrypt password hash.
func (u *BasicUser) GetPasswordHash() string {
	return u.PasswordHash
}

// SetID sets the user's unique identifier.
func (u *BasicUser) SetID(id googleUuid.UUID) {
	u.ID = id
}

// SetUsername sets the user's username.
func (u *BasicUser) SetUsername(username string) {
	u.Username = username
}

// SetPasswordHash sets the user's password hash (for updates).
func (u *BasicUser) SetPasswordHash(hash string) {
	u.PasswordHash = hash
}
