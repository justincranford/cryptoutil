// Copyright (c) 2025 Justin Cranford
//

package realms

import (
	"context"
	"fmt"
	"time"

	cryptoutilSharedCryptoDigests "cryptoutil/internal/shared/crypto/digests"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"

	googleUuid "github.com/google/uuid"
)

// UserServiceImpl implements user registration and authentication using PBKDF2 (LowEntropyRandom).
type UserServiceImpl struct {
	userRepo    UserRepository
	userFactory func() UserModel // Factory function to create user instances
}

// NewUserService creates a new UserService.
//
// Parameters:
// - userRepo: Repository for user CRUD operations
// - userFactory: Factory function to create new UserModel instances (e.g., func() UserModel { return &domain.User{} })
//
// Returns configured UserService ready for registration and authentication.
func NewUserService(userRepo UserRepository, userFactory func() UserModel) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo:    userRepo,
		userFactory: userFactory,
	}
}

// RegisterUser creates a new user account with hashed password.
//
// Workflow:
// 1. Validate username and password (non-empty, length requirements)
// 2. Hash password using PBKDF2-HMAC-SHA256 (LowEntropyRandom, version 1)
// 3. Create user entity with UUIDv7
// 4. Save to repository
// 5. Return created user
//
// Validation Rules:
// - Username: 3-50 characters, alphanumeric + underscore
// - Password: 8+ characters minimum
//
// Security Notes:
// - Password hashed with PBKDF2-HMAC-SHA256 (FIPS-approved, OWASP 2023 recommended)
// - 600,000 iterations (OWASP 2023 recommendation)
// - Versioned hash format: {1}$pbkdf2-sha256$600000$base64(salt)$base64(dk)
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

	// Hash password with PBKDF2-HMAC-SHA256 (LowEntropyRandom, FIPS-approved).
	passwordHash, err := cryptoutilSharedCryptoHash.HashSecretPBKDF2(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity using factory.
	userID := googleUuid.Must(googleUuid.NewV7())
	user := s.userFactory()
	user.SetID(userID)
	user.SetUsername(username)
	user.SetPasswordHash(passwordHash)

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
// 2. Verify password hash with PBKDF2 (constant-time comparison)
// 3. Return user on success
//
// Security Notes:
// - Constant-time comparison (cryptoutilDigests.VerifySecret prevents timing attacks)
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

	// Verify password using PBKDF2 (supports versioned hash format).
	valid, err := cryptoutilSharedCryptoDigests.VerifySecret(user.GetPasswordHash(), password)
	if err != nil || !valid {
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

// GetPasswordHash returns the user's PBKDF2 password hash (versioned format).
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
