// Copyright (c) 2025 Justin Cranford
//
//

// Package realm provides database-backed realm authentication for KMS.
package realm

import (
	"context"
	crand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/pbkdf2"
	"gorm.io/gorm"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ErrUserNotFound is returned when a user is not found in the database.
var ErrUserNotFound = errors.New("user not found")

// DBRealmUser represents a user in the database realm.
// This is separate from Identity server users.
type DBRealmUser struct {
	// ID is the primary key (UUIDv7).
	ID string `gorm:"type:text;primaryKey"`

	// RealmID is the foreign key to the realm.
	RealmID string `gorm:"type:text;index;not null"`

	// Username is the unique username within the realm.
	Username string `gorm:"type:text;not null;uniqueIndex:idx_realm_username"`

	// PasswordHash is the PBKDF2-HMAC-SHA256 hashed password.
	PasswordHash string `gorm:"type:text;not null"`

	// Email is the optional user email.
	Email string `gorm:"type:text"`

	// Roles is the JSON-encoded list of role names.
	Roles string `gorm:"type:text;default:'[]'"` // JSON array as text.

	// Enabled indicates if the user is active.
	Enabled bool `gorm:"not null"`

	// Metadata is optional JSON metadata.
	Metadata string `gorm:"type:text"` // JSON object as text.

	// MetadataSchema is the JSON schema reference.
	MetadataSchema string `gorm:"type:text"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `gorm:"not null"`

	// UpdatedAt is the last update timestamp.
	UpdatedAt time.Time `gorm:"not null"`

	// LastLoginAt is the last successful login timestamp.
	LastLoginAt *time.Time `gorm:""`
}

// TableName returns the table name for GORM.
func (DBRealmUser) TableName() string {
	return "kms_realm_users"
}

// DBRealmRepository provides database operations for realm users.
type DBRealmRepository struct {
	db     *gorm.DB
	policy *PasswordPolicyConfig
}

// NewDBRealmRepository creates a new database realm repository.
func NewDBRealmRepository(db *gorm.DB, policy *PasswordPolicyConfig) (*DBRealmRepository, error) {
	if db == nil {
		return nil, errors.New("database connection is required")
	}

	if policy == nil {
		defaultPolicy := DefaultPasswordPolicy()
		policy = &defaultPolicy
	}

	return &DBRealmRepository{
		db:     db,
		policy: policy,
	}, nil
}

// Migrate creates or updates the database schema.
func (r *DBRealmRepository) Migrate(ctx context.Context) error {
	if err := r.db.WithContext(ctx).AutoMigrate(&DBRealmUser{}); err != nil {
		return fmt.Errorf("failed to migrate kms_realm_users table: %w", err)
	}

	return nil
}

// CreateUser creates a new user in the database realm.
func (r *DBRealmRepository) CreateUser(ctx context.Context, user *DBRealmUser, password string) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	if password == "" {
		return errors.New("password cannot be empty")
	}

	// Hash password.
	hash, err := r.hashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = hash
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUser retrieves a user by ID.
// Returns ErrUserNotFound if the user does not exist.
func (r *DBRealmRepository) GetUser(ctx context.Context, userID string) (*DBRealmUser, error) {
	var user DBRealmUser
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by realm ID and username.
// Returns ErrUserNotFound if the user does not exist.
func (r *DBRealmRepository) GetUserByUsername(ctx context.Context, realmID, username string) (*DBRealmUser, error) {
	var user DBRealmUser
	if err := r.db.WithContext(ctx).Where("realm_id = ? AND username = ?", realmID, username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// UpdateUser updates an existing user.
func (r *DBRealmRepository) UpdateUser(ctx context.Context, user *DBRealmUser) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	user.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdatePassword updates a user's password.
func (r *DBRealmRepository) UpdatePassword(ctx context.Context, userID, newPassword string) error {
	hash, err := r.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := r.db.WithContext(ctx).Model(&DBRealmUser{}).Where("id = ?", userID).Updates(map[string]any{
		"password_hash": hash,
		"updated_at":    time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// DeleteUser deletes a user by ID.
func (r *DBRealmRepository) DeleteUser(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).Delete(&DBRealmUser{}, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers lists all users in a realm.
func (r *DBRealmRepository) ListUsers(ctx context.Context, realmID string, limit, offset int) ([]*DBRealmUser, error) {
	var users []*DBRealmUser

	query := r.db.WithContext(ctx).Where("realm_id = ?", realmID)

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// Authenticate verifies user credentials and returns an auth result.
func (r *DBRealmRepository) Authenticate(ctx context.Context, realmID, username, password string) (*AuthResult, error) {
	result := &AuthResult{
		Timestamp: time.Now(),
		RealmID:   realmID,
	}

	// Get user.
	user, err := r.GetUserByUsername(ctx, realmID, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			result.Error = errMsgUserNotFound
			result.ErrorCode = AuthErrorUserNotFound

			return result, nil
		}

		result.Error = "database error"
		result.ErrorCode = AuthErrorInvalidCreds

		return result, fmt.Errorf("failed to lookup user: %w", err)
	}

	if !user.Enabled {
		result.Error = errMsgUserDisabled
		result.ErrorCode = AuthErrorUserDisabled

		return result, nil
	}

	// Verify password.
	if err := r.verifyPassword(password, user.PasswordHash); err != nil {
		result.Error = errMsgInvalidPassword
		result.ErrorCode = AuthErrorPasswordMismatch
		// Return nil error - password mismatch is expected user error, not system error.

		return result, nil //nolint:nilerr // Intentional: password mismatch is auth failure, not system error.
	}

	// Update last login.
	now := time.Now()
	user.LastLoginAt = &now

	if err := r.UpdateUser(ctx, user); err != nil {
		// Log but don't fail authentication.
		_ = err
	}

	// Success.
	result.Authenticated = true
	result.UserID = user.ID
	result.Username = user.Username

	return result, nil
}

// hashPassword creates a PBKDF2-SHA256 password hash.
func (r *DBRealmRepository) hashPassword(password string) (string, error) {
	salt := make([]byte, r.policy.SaltBytes)
	if _, err := crand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hashFunc := cryptoutilSharedMagic.PBKDF2HashFunction(r.policy.Algorithm)
	derivedKey := pbkdf2.Key(
		[]byte(password),
		salt,
		r.policy.Iterations,
		r.policy.HashBytes,
		hashFunc,
	)

	return fmt.Sprintf("$%s$%d$%s$%s",
		cryptoutilSharedMagic.PBKDF2DefaultHashName,
		r.policy.Iterations,
		base64.StdEncoding.EncodeToString(salt),
		base64.StdEncoding.EncodeToString(derivedKey),
	), nil
}

// verifyPassword verifies a password against a stored hash.
func (r *DBRealmRepository) verifyPassword(password, hashStr string) error {
	// Use the same verifyPassword from Authenticator.
	auth := &Authenticator{}

	return auth.verifyPassword(password, hashStr, r.policy)
}

// CountUsers returns the total number of users in a realm.
func (r *DBRealmRepository) CountUsers(ctx context.Context, realmID string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&DBRealmUser{}).Where("realm_id = ?", realmID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// EnableUser enables a user account.
func (r *DBRealmRepository) EnableUser(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).Model(&DBRealmUser{}).Where("id = ?", userID).Updates(map[string]any{
		"enabled":    true,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to enable user: %w", err)
	}

	return nil
}

// DisableUser disables a user account.
func (r *DBRealmRepository) DisableUser(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).Model(&DBRealmUser{}).Where("id = ?", userID).Updates(map[string]any{
		"enabled":    false,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to disable user: %w", err)
	}

	return nil
}
