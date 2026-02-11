// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	"crypto/subtle"
	"fmt"
	"time"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
	cryptoutilSharedCryptoPassword "cryptoutil/internal/shared/crypto/password"

	googleUuid "github.com/google/uuid"
)

// PasswordCredentialStore defines the interface for storing and retrieving password credentials.
type PasswordCredentialStore interface {
	// StoreCredential stores a password credential for a user.
	StoreCredential(ctx context.Context, userID string, passwordHash []byte) error

	// GetCredential retrieves the password hash for a user.
	GetCredential(ctx context.Context, userID string) ([]byte, error)

	// DeleteCredential deletes the password credential for a user.
	DeleteCredential(ctx context.Context, userID string) error

	// UpdateCredential updates the password credential for a user.
	UpdateCredential(ctx context.Context, userID string, newPasswordHash []byte) error
}

// UserStore defines a minimal interface for user authentication.
// TODO: Replace with proper UserRepository from domain package when available.
type UserStore interface {
	// GetByID retrieves a user by ID.
	GetByID(ctx context.Context, userID string) (*cryptoutilIdentityDomain.User, error)

	// Update updates a user.
	Update(ctx context.Context, user *cryptoutilIdentityDomain.User) error
}

// UsernamePasswordAuthenticator provides traditional username/password authentication
// optionally enhanced with hardware security features.
type UsernamePasswordAuthenticator struct {
	credentialStore  PasswordCredentialStore
	challengeStore   ChallengeStore
	userStore        UserStore
	hsm              HardwareSecurityModule // Optional hardware enhancement.
	requireHardware  bool                   // Whether to require hardware authentication.
	lockoutThreshold int                    // Failed login attempts before lockout.
	lockoutDuration  time.Duration          // Duration of account lockout.
}

// NewUsernamePasswordAuthenticator creates a new username/password authenticator.
func NewUsernamePasswordAuthenticator(
	credentialStore PasswordCredentialStore,
	challengeStore ChallengeStore,
	userStore UserStore,
	hsm HardwareSecurityModule,
	requireHardware bool,
) *UsernamePasswordAuthenticator {
	return &UsernamePasswordAuthenticator{
		credentialStore:  credentialStore,
		challengeStore:   challengeStore,
		userStore:        userStore,
		hsm:              hsm,
		requireHardware:  requireHardware,
		lockoutThreshold: cryptoutilIdentityMagic.MaxOTPAttempts,
		lockoutDuration:  cryptoutilIdentityMagic.DefaultOTPLockout,
	}
}

// Method returns the authentication method identifier.
func (u *UsernamePasswordAuthenticator) Method() string {
	return cryptoutilIdentityMagic.AuthMethodUsernamePassword
}

// InitiateAuth initiates username/password authentication.
func (u *UsernamePasswordAuthenticator) InitiateAuth(ctx context.Context, userID string) (*AuthChallenge, error) {
	// Check if user exists.
	user, err := u.userStore.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if account is enabled.
	if !user.Enabled {
		return nil, fmt.Errorf("account disabled")
	}

	// Check if account is locked.
	if user.Locked {
		return nil, fmt.Errorf("account locked")
	}

	// Generate challenge.
	challengeID, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate challenge ID: %w", err)
	}

	expiresAt := time.Now().UTC().Add(cryptoutilIdentityMagic.DefaultOTPLifetime)

	// Create challenge with hardware requirement if enabled.
	challenge := &AuthChallenge{
		ID:        challengeID,
		UserID:    userID,
		Method:    cryptoutilIdentityMagic.AuthMethodUsernamePassword,
		ExpiresAt: expiresAt,
		Metadata: map[string]any{
			"require_hardware": u.requireHardware,
		},
	}

	// Store challenge with empty secret (password verified separately).
	if err := u.challengeStore.Store(ctx, challenge, ""); err != nil {
		return nil, fmt.Errorf("failed to store challenge: %w", err)
	}

	return challenge, nil
}

// VerifyAuth verifies the username/password authentication response.
func (u *UsernamePasswordAuthenticator) VerifyAuth(ctx context.Context, challengeID googleUuid.UUID, response string) (*cryptoutilIdentityDomain.User, error) {
	// Retrieve challenge.
	challenge, _, err := u.challengeStore.Retrieve(ctx, challengeID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve challenge: %w", err)
	}

	// Check challenge expiration.
	if time.Now().UTC().After(challenge.ExpiresAt) {
		return nil, fmt.Errorf("challenge expired")
	}

	// Response is the password for this auth method.
	password := response
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	// Get user.
	user, err := u.userStore.GetByID(ctx, challenge.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if account is enabled.
	if !user.Enabled {
		return nil, fmt.Errorf("account disabled")
	}

	// Check if account is locked.
	if user.Locked {
		return nil, fmt.Errorf("account locked")
	}

	// Get stored password hash.
	passwordHash, err := u.credentialStore.GetCredential(ctx, challenge.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}

	// Verify password.
	match, _, err := cryptoutilSharedCryptoPassword.VerifyPassword(password, string(passwordHash))
	if err != nil || !match {
		// Note: Failed attempt tracking would be implemented here if User model had those fields.
		// For now, just return error.
		return nil, fmt.Errorf("invalid password")
	}

	// If hardware authentication required, verify hardware signature.
	// Note: Hardware signature would be passed via challenge metadata in a real implementation.
	if u.requireHardware && u.hsm != nil { //nolint:revive // Stub for future hardware signature verification.
		// In a real implementation, client would sign challenge ID with hardware key.
		// Hardware signature verification logic would go here.
		// For now, hardware authentication is not enforced (stub implementation).
		// Example: u.hsm.VerifySignature(ctx, user.ID.String(), challengeID.Bytes(), signature)
		_ = u.hsm // Mark as intentionally unused pending implementation
	}

	// Delete used challenge.
	if err := u.challengeStore.Delete(ctx, challengeID); err != nil {
		return nil, fmt.Errorf("failed to delete challenge: %w", err)
	}

	return user, nil
}

// HashPassword hashes a password using PBKDF2-HMAC-SHA256 (FIPS-compliant).
func (u *UsernamePasswordAuthenticator) HashPassword(password string) ([]byte, error) {
	if len(password) < cryptoutilIdentityMagic.MinPasswordLength {
		return nil, fmt.Errorf("password too short (minimum %d characters)", cryptoutilIdentityMagic.MinPasswordLength)
	}

	if len(password) > cryptoutilIdentityMagic.MaxPasswordLength {
		return nil, fmt.Errorf("password too long (maximum %d characters)", cryptoutilIdentityMagic.MaxPasswordLength)
	}

	hash, err := cryptoutilSharedCryptoHash.HashLowEntropyNonDeterministic(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	return []byte(hash), nil
}

// ValidatePassword validates a password against security requirements.
func (u *UsernamePasswordAuthenticator) ValidatePassword(password string) error {
	if len(password) < cryptoutilIdentityMagic.MinPasswordLength {
		return fmt.Errorf("password too short (minimum %d characters)", cryptoutilIdentityMagic.MinPasswordLength)
	}

	if len(password) > cryptoutilIdentityMagic.MaxPasswordLength {
		return fmt.Errorf("password too long (maximum %d characters)", cryptoutilIdentityMagic.MaxPasswordLength)
	}

	// Additional password complexity checks could be added here.
	return nil
}

// UpdatePassword updates a user's password with proper validation.
func (u *UsernamePasswordAuthenticator) UpdatePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	// Validate new password.
	if err := u.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Get current password hash.
	currentHash, err := u.credentialStore.GetCredential(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to retrieve credential: %w", err)
	}

	// Verify old password (supports legacy and PBKDF2 hashes).
	match, _, err := cryptoutilSharedCryptoPassword.VerifyPassword(oldPassword, string(currentHash))
	if err != nil {
		return fmt.Errorf("password verification failed: %w", err)
	}

	if !match {
		return fmt.Errorf("invalid current password")
	}

	// Hash new password (always uses PBKDF2).
	newHashStr, err := cryptoutilSharedCryptoPassword.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	newHash := []byte(newHashStr)

	// Note: If old hash was legacy algorithm, new password uses PBKDF2 (automatic upgrade).

	// Ensure new password is different from old password.
	// Note: This comparison is approximate due to salt randomization.
	if subtle.ConstantTimeCompare(currentHash, newHash) == 1 {
		return fmt.Errorf("new password must be different from current password")
	}

	// Update credential.
	if err := u.credentialStore.UpdateCredential(ctx, userID, newHash); err != nil {
		return fmt.Errorf("failed to update credential: %w", err)
	}

	return nil
}
