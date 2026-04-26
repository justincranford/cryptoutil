// Copyright (c) 2025 Justin Cranford
//
//

// Package realm provides realm-based authentication for KMS.
package realm

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/pbkdf2"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Authenticator provides authentication for realm users.
type Authenticator struct {
	config    *Config
	realmMap  map[string]*RealmConfig // keyed by realm ID.
	userIndex map[string]*userEntry   // keyed by "realmID:username".
	mu        sync.RWMutex
}

// userEntry holds indexed user data.
type userEntry struct {
	user   *UserConfig
	realm  *RealmConfig
	policy *PasswordPolicyConfig
}

// AuthResult represents the result of an authentication attempt.
type AuthResult struct {
	// Authenticated indicates if authentication succeeded.
	Authenticated bool `json:"authenticated"`

	// UserID is the authenticated user's ID (if successful).
	UserID string `json:"user_id,omitempty"`

	// Username is the authenticated username (if successful).
	Username string `json:"username,omitempty"`

	// RealmID is the realm that authenticated the user.
	RealmID string `json:"realm_id,omitempty"`

	// RealmName is the human-readable realm name.
	RealmName string `json:"realm_name,omitempty"`

	// Roles is the list of roles assigned to the user.
	Roles []string `json:"roles,omitempty"`

	// Permissions is the expanded list of permissions.
	Permissions []string `json:"permissions,omitempty"`

	// Error provides details if authentication failed.
	Error string `json:"error,omitempty"`

	// ErrorCode provides a machine-readable error code.
	ErrorCode AuthErrorCode `json:"error_code,omitempty"`

	// Timestamp is when the authentication occurred.
	Timestamp time.Time `json:"timestamp"`
}

// AuthErrorCode represents authentication error codes.
type AuthErrorCode string

// Authentication error codes.
const (
	AuthErrorNone             AuthErrorCode = ""
	AuthErrorInvalidCreds     AuthErrorCode = "invalid_credentials"
	AuthErrorUserDisabled     AuthErrorCode = "user_disabled"
	AuthErrorRealmDisabled    AuthErrorCode = "realm_disabled"
	AuthErrorRealmNotFound    AuthErrorCode = "realm_not_found"
	AuthErrorUserNotFound     AuthErrorCode = "user_not_found"
	AuthErrorInvalidHashFmt   AuthErrorCode = "invalid_hash_format"
	AuthErrorPasswordMismatch AuthErrorCode = "password_mismatch"
)

// Authentication error messages.
const (
	errMsgUserNotFound    = "user not found"
	errMsgUserDisabled    = "user is disabled"
	errMsgInvalidPassword = "invalid password"
)

// NewAuthenticator creates a new realm authenticator from configuration.
func NewAuthenticator(config *Config) (*Authenticator, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	auth := &Authenticator{
		config:    config,
		realmMap:  make(map[string]*RealmConfig),
		userIndex: make(map[string]*userEntry),
	}

	if err := auth.buildIndex(); err != nil {
		return nil, fmt.Errorf("failed to build authentication index: %w", err)
	}

	return auth, nil
}

// buildIndex creates lookup indexes for realms and users.
func (a *Authenticator) buildIndex() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Clear existing indexes.
	a.realmMap = make(map[string]*RealmConfig)
	a.userIndex = make(map[string]*userEntry)

	for i := range a.config.Realms {
		realm := &a.config.Realms[i]
		a.realmMap[realm.ID] = realm

		// Determine password policy.
		policy := &realm.PasswordPolicy
		if policy.Algorithm == "" {
			policy = &a.config.Defaults.PasswordPolicy
		}

		// Index users for file-based realms.
		if realm.Type == RealmTypeFile {
			for j := range realm.Users {
				user := &realm.Users[j]
				key := fmt.Sprintf("%s:%s", realm.ID, user.Username)
				a.userIndex[key] = &userEntry{
					user:   user,
					realm:  realm,
					policy: policy,
				}
			}
		}
	}

	return nil
}

// Authenticate authenticates a user against a specific realm.
func (a *Authenticator) Authenticate(ctx context.Context, realmID, username, password string) *AuthResult {
	result := &AuthResult{
		Timestamp: time.Now().UTC(),
	}

	// Validate inputs.
	if realmID == "" || username == "" || password == "" {
		result.Error = "missing required credentials"
		result.ErrorCode = AuthErrorInvalidCreds

		return result
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	// Find realm.
	realm, ok := a.realmMap[realmID]
	if !ok {
		result.Error = "realm not found"
		result.ErrorCode = AuthErrorRealmNotFound

		return result
	}

	if !realm.Enabled {
		result.Error = "realm is disabled"
		result.ErrorCode = AuthErrorRealmDisabled

		return result
	}

	// Route to appropriate authentication method.
	switch realm.Type {
	case RealmTypeFile:
		return a.authenticateFileRealm(ctx, realm, username, password)
	case RealmTypeDatabase:
		result.Error = "database realm not implemented"
		result.ErrorCode = AuthErrorRealmNotFound

		return result
	case RealmTypeLDAP:
		result.Error = "LDAP realm not implemented"
		result.ErrorCode = AuthErrorRealmNotFound

		return result
	case RealmTypeOIDC:
		result.Error = "OIDC realm not implemented"
		result.ErrorCode = AuthErrorRealmNotFound

		return result
	default:
		result.Error = "unsupported realm type"
		result.ErrorCode = AuthErrorRealmNotFound

		return result
	}
}

// authenticateFileRealm authenticates against a file-based realm.
func (a *Authenticator) authenticateFileRealm(_ context.Context, realm *RealmConfig, username, password string) *AuthResult {
	result := &AuthResult{
		Timestamp: time.Now().UTC(),
		RealmID:   realm.ID,
		RealmName: realm.Name,
	}

	// Look up user.
	key := fmt.Sprintf("%s:%s", realm.ID, username)

	entry, ok := a.userIndex[key]
	if !ok {
		result.Error = errMsgUserNotFound
		result.ErrorCode = AuthErrorUserNotFound

		return result
	}

	if !entry.user.Enabled {
		result.Error = errMsgUserDisabled
		result.ErrorCode = AuthErrorUserDisabled

		return result
	}

	// Verify password.
	if err := a.verifyPassword(password, entry.user.PasswordHash, entry.policy); err != nil {
		result.Error = errMsgInvalidPassword
		result.ErrorCode = AuthErrorPasswordMismatch

		return result
	}

	// Success.
	result.Authenticated = true
	result.UserID = entry.user.ID
	result.Username = entry.user.Username
	result.Roles = entry.user.Roles
	result.Permissions = a.expandPermissions(realm, entry.user.Roles)

	return result
}

// verifyPassword verifies a password against a PBKDF2 hash.
// Hash format: $pbkdf2-sha256$iterations$salt$hash.
func (a *Authenticator) verifyPassword(password, hashStr string, policy *PasswordPolicyConfig) error {
	// Parse the hash string.
	parts := strings.Split(hashStr, "$")

	// Expected format: ["", "pbkdf2-sha256", "iterations", "salt", "hash"].
	const expectedParts = 5
	if len(parts) != expectedParts {
		return fmt.Errorf("invalid hash format: expected %d parts, got %d", expectedParts, len(parts))
	}

	// Verify algorithm matches.
	if parts[1] != cryptoutilSharedMagic.PBKDF2DefaultHashName {
		return fmt.Errorf("unsupported hash algorithm: %s", parts[1])
	}

	// Parse iterations.
	var iterations int
	if _, err := fmt.Sscanf(parts[2], "%d", &iterations); err != nil {
		return fmt.Errorf("invalid iterations: %w", err)
	}

	// Decode salt.
	salt, err := base64.StdEncoding.DecodeString(parts[3])
	if err != nil {
		return fmt.Errorf("invalid salt encoding: %w", err)
	}

	// Decode expected hash.
	expectedHash, err := base64.StdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("invalid hash encoding: %w", err)
	}

	// Compute hash with same parameters.
	hashFunc := cryptoutilSharedMagic.PBKDF2HashFunction(policy.Algorithm)
	computedHash := pbkdf2.Key([]byte(password), salt, iterations, len(expectedHash), hashFunc)

	// Constant-time comparison.
	if subtle.ConstantTimeCompare(expectedHash, computedHash) != 1 {
		return errors.New("password mismatch")
	}

	return nil
}

// expandPermissions expands role names into permissions.
func (a *Authenticator) expandPermissions(realm *RealmConfig, roleNames []string) []string {
	permSet := make(map[string]bool)

	// Build role lookup.
	roleMap := make(map[string]*RoleConfig)
	for i := range realm.Roles {
		roleMap[realm.Roles[i].Name] = &realm.Roles[i]
	}

	// Recursively expand roles.
	var expand func(roleName string)

	expand = func(roleName string) {
		role, ok := roleMap[roleName]
		if !ok {
			return
		}

		// Add direct permissions.
		for _, perm := range role.Permissions {
			permSet[perm] = true
		}

		// Expand inherited roles.
		for _, parent := range role.Inherits {
			expand(parent)
		}
	}

	for _, roleName := range roleNames {
		expand(roleName)
	}

	// Convert to sorted slice.
	permissions := make([]string, 0, len(permSet))
	for perm := range permSet {
		permissions = append(permissions, perm)
	}

	return permissions
}

// AuthenticateByRealmName authenticates using realm name instead of ID.
func (a *Authenticator) AuthenticateByRealmName(ctx context.Context, realmName, username, password string) *AuthResult {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Find realm by name.
	for _, realm := range a.realmMap {
		if realm.Name == realmName {
			return a.Authenticate(ctx, realm.ID, username, password)
		}
	}

	return &AuthResult{
		Timestamp: time.Now().UTC(),
		Error:     "realm not found",
		ErrorCode: AuthErrorRealmNotFound,
	}
}

// GetRealm returns realm configuration by ID.
func (a *Authenticator) GetRealm(realmID string) (*RealmConfig, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	realm, ok := a.realmMap[realmID]

	return realm, ok
}

// ListRealms returns all configured realms.
func (a *Authenticator) ListRealms() []RealmConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()

	realms := make([]RealmConfig, 0, len(a.realmMap))
	for _, realm := range a.realmMap {
		realms = append(realms, *realm)
	}

	return realms
}

// Reload reloads configuration from disk.
func (a *Authenticator) Reload(configDir string) error {
	newConfig, err := LoadConfig(configDir)
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	a.mu.Lock()
	a.config = newConfig
	a.mu.Unlock()

	return a.buildIndex()
}
