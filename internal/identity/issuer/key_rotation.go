// Copyright (c) 2025 Justin Cranford
//
//

package issuer

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	rsa "crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"sync"
	"time"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// KeyRotationPolicy defines key rotation behavior.
type KeyRotationPolicy struct {
	RotationInterval    time.Duration // How often to rotate keys.
	GracePeriod         time.Duration // Grace period for old keys (overlap period).
	MaxActiveKeys       int           // Maximum number of active keys.
	AutoRotationEnabled bool          // Whether automatic rotation is enabled.
}

// DefaultKeyRotationPolicy returns the default key rotation policy.
func DefaultKeyRotationPolicy() *KeyRotationPolicy {
	return &KeyRotationPolicy{
		RotationInterval:    cryptoutilIdentityMagic.DefaultKeyRotationInterval,
		GracePeriod:         cryptoutilIdentityMagic.DefaultKeyGracePeriod,
		MaxActiveKeys:       cryptoutilIdentityMagic.DefaultMaxActiveKeys,
		AutoRotationEnabled: false,
	}
}

// StrictKeyRotationPolicy returns a strict key rotation policy for production.
func StrictKeyRotationPolicy() *KeyRotationPolicy {
	return &KeyRotationPolicy{
		RotationInterval:    cryptoutilIdentityMagic.StrictKeyRotationInterval,
		GracePeriod:         cryptoutilIdentityMagic.StrictKeyGracePeriod,
		MaxActiveKeys:       cryptoutilIdentityMagic.StrictMaxActiveKeys,
		AutoRotationEnabled: true,
	}
}

// DevelopmentKeyRotationPolicy returns a relaxed policy for development.
func DevelopmentKeyRotationPolicy() *KeyRotationPolicy {
	return &KeyRotationPolicy{
		RotationInterval:    cryptoutilIdentityMagic.DevelopmentKeyRotationInterval,
		GracePeriod:         cryptoutilIdentityMagic.DevelopmentKeyGracePeriod,
		MaxActiveKeys:       cryptoutilIdentityMagic.DevelopmentMaxActiveKeys,
		AutoRotationEnabled: false,
	}
}

// SigningKey represents a versioned signing key.
type SigningKey struct {
	KeyID         string    // Unique key identifier.
	Key           any       // Actual key material (RSA, ECDSA, HMAC).
	Algorithm     string    // Signing algorithm (RS256, ES256, HS256, etc.).
	CreatedAt     time.Time // When key was created.
	ExpiresAt     time.Time // When key expires (after rotation + grace period).
	Active        bool      // Whether key is active for signing new tokens.
	ValidForVerif bool      // Whether key is valid for verification.
}

// EncryptionKey represents a versioned encryption key.
type EncryptionKey struct {
	KeyID        string    // Unique key identifier.
	Key          []byte    // AES key material.
	CreatedAt    time.Time // When key was created.
	ExpiresAt    time.Time // When key expires.
	Active       bool      // Whether key is active for encrypting new tokens.
	ValidForDecr bool      // Whether key is valid for decryption.
}

// KeyRotationManager manages key rotation for signing and encryption.
type KeyRotationManager struct {
	mu               sync.RWMutex
	signingKeys      []*SigningKey
	encryptionKeys   []*EncryptionKey
	policy           *KeyRotationPolicy
	keyGenerator     KeyGenerator
	rotationCallback func(keyID string) // Callback when key rotates.
}

// KeyGenerator generates new signing and encryption keys.
type KeyGenerator interface {
	GenerateSigningKey(ctx context.Context, algorithm string) (*SigningKey, error)
	GenerateEncryptionKey(ctx context.Context) (*EncryptionKey, error)
}

// NewKeyRotationManager creates a new key rotation manager.
func NewKeyRotationManager(
	policy *KeyRotationPolicy,
	generator KeyGenerator,
	rotationCallback func(keyID string),
) (*KeyRotationManager, error) {
	if policy == nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrInvalidConfiguration,
			fmt.Errorf("rotation policy cannot be nil"),
		)
	}

	if generator == nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrInvalidConfiguration,
			fmt.Errorf("key generator cannot be nil"),
		)
	}

	return &KeyRotationManager{
		signingKeys:      make([]*SigningKey, 0),
		encryptionKeys:   make([]*EncryptionKey, 0),
		policy:           policy,
		keyGenerator:     generator,
		rotationCallback: rotationCallback,
	}, nil
}

// GetActiveSigningKey returns the currently active signing key.
func (m *KeyRotationManager) GetActiveSigningKey() (*SigningKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, key := range m.signingKeys {
		if key.Active && time.Now().UTC().Before(key.ExpiresAt) {
			return key, nil
		}
	}

	return nil, cryptoutilIdentityAppErr.WrapError(
		cryptoutilIdentityAppErr.ErrKeyNotFound,
		fmt.Errorf("no active signing key available"),
	)
}

// GetSigningKeyCount returns the number of signing keys (thread-safe).
func (m *KeyRotationManager) GetSigningKeyCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.signingKeys)
}

// GetPublicKeys returns all valid signing keys in JWK format for JWKS endpoint.
func (m *KeyRotationManager) GetPublicKeys() []map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]map[string]any, 0)

	for _, key := range m.signingKeys {
		if !key.ValidForVerif || time.Now().UTC().After(key.ExpiresAt) {
			continue
		}

		jwk := convertToJWK(key)
		if jwk != nil {
			keys = append(keys, jwk)
		}
	}

	return keys
}

// GetSigningKeyByID retrieves a signing key by its ID for verification.
func (m *KeyRotationManager) GetSigningKeyByID(keyID string) (*SigningKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, key := range m.signingKeys {
		if key.KeyID == keyID && key.ValidForVerif {
			return key, nil
		}
	}

	return nil, cryptoutilIdentityAppErr.WrapError(
		cryptoutilIdentityAppErr.ErrKeyNotFound,
		fmt.Errorf("signing key not found: %s", keyID),
	)
}

// GetAllValidVerificationKeys returns all signing keys that are valid for verification.
func (m *KeyRotationManager) GetAllValidVerificationKeys() []*SigningKey {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]*SigningKey, 0)

	for _, key := range m.signingKeys {
		if key.ValidForVerif && (key.ExpiresAt.IsZero() || time.Now().UTC().Before(key.ExpiresAt)) {
			keys = append(keys, key)
		}
	}

	return keys
}

// GetActiveEncryptionKey returns the currently active encryption key.
func (m *KeyRotationManager) GetActiveEncryptionKey() (*EncryptionKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, key := range m.encryptionKeys {
		if key.Active && time.Now().UTC().Before(key.ExpiresAt) {
			return key, nil
		}
	}

	return nil, cryptoutilIdentityAppErr.WrapError(
		cryptoutilIdentityAppErr.ErrKeyNotFound,
		fmt.Errorf("no active encryption key available"),
	)
}

// GetEncryptionKeyByID retrieves an encryption key by its ID for decryption.
func (m *KeyRotationManager) GetEncryptionKeyByID(keyID string) (*EncryptionKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, key := range m.encryptionKeys {
		if key.KeyID == keyID && key.ValidForDecr {
			return key, nil
		}
	}

	return nil, cryptoutilIdentityAppErr.WrapError(
		cryptoutilIdentityAppErr.ErrKeyNotFound,
		fmt.Errorf("encryption key not found: %s", keyID),
	)
}

// RotateSigningKey creates a new signing key and marks old keys for grace period.
func (m *KeyRotationManager) RotateSigningKey(ctx context.Context, algorithm string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate new signing key.
	newKey, err := m.keyGenerator.GenerateSigningKey(ctx, algorithm)
	if err != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrKeyRotationFailed,
			fmt.Errorf("failed to generate signing key: %w", err),
		)
	}

	// Set expiration based on rotation interval + grace period.
	newKey.ExpiresAt = time.Now().UTC().Add(m.policy.RotationInterval + m.policy.GracePeriod)
	newKey.Active = true
	newKey.ValidForVerif = true

	// Mark current active key as inactive (but still valid for verification during grace period).
	for _, key := range m.signingKeys {
		if key.Active {
			key.Active = false
			// Extend verification validity through grace period.
			if time.Now().UTC().Add(m.policy.GracePeriod).After(key.ExpiresAt) {
				key.ExpiresAt = time.Now().UTC().Add(m.policy.GracePeriod)
			}
		}
	}

	// Add new key.
	m.signingKeys = append(m.signingKeys, newKey)

	// Prune expired keys.
	m.pruneSigningKeys()

	// Enforce max active keys limit.
	if len(m.signingKeys) > m.policy.MaxActiveKeys {
		// Mark oldest keys as invalid for verification.
		excess := len(m.signingKeys) - m.policy.MaxActiveKeys
		for i := 0; i < excess; i++ {
			m.signingKeys[i].ValidForVerif = false
		}

		// Remove them from the list.
		m.signingKeys = m.signingKeys[excess:]
	}

	// Trigger rotation callback.
	if m.rotationCallback != nil {
		m.rotationCallback(newKey.KeyID)
	}

	return nil
}

// RotateEncryptionKey creates a new encryption key and marks old keys for grace period.
func (m *KeyRotationManager) RotateEncryptionKey(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate new encryption key.
	newKey, err := m.keyGenerator.GenerateEncryptionKey(ctx)
	if err != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrKeyRotationFailed,
			fmt.Errorf("failed to generate encryption key: %w", err),
		)
	}

	// Set expiration.
	newKey.ExpiresAt = time.Now().UTC().Add(m.policy.RotationInterval + m.policy.GracePeriod)
	newKey.Active = true
	newKey.ValidForDecr = true

	// Mark current active key as inactive.
	for _, key := range m.encryptionKeys {
		if key.Active {
			key.Active = false
			if time.Now().UTC().Add(m.policy.GracePeriod).After(key.ExpiresAt) {
				key.ExpiresAt = time.Now().UTC().Add(m.policy.GracePeriod)
			}
		}
	}

	// Add new key.
	m.encryptionKeys = append(m.encryptionKeys, newKey)

	// Prune expired keys.
	m.pruneEncryptionKeys()

	// Enforce max active keys limit.
	if len(m.encryptionKeys) > m.policy.MaxActiveKeys {
		excess := len(m.encryptionKeys) - m.policy.MaxActiveKeys
		for i := 0; i < excess; i++ {
			m.encryptionKeys[i].ValidForDecr = false
		}

		m.encryptionKeys = m.encryptionKeys[excess:]
	}

	// Trigger rotation callback.
	if m.rotationCallback != nil {
		m.rotationCallback(newKey.KeyID)
	}

	return nil
}

// pruneSigningKeys removes expired signing keys.
func (m *KeyRotationManager) pruneSigningKeys() {
	now := time.Now().UTC()
	validKeys := make([]*SigningKey, 0)

	for _, key := range m.signingKeys {
		if now.Before(key.ExpiresAt) {
			validKeys = append(validKeys, key)
		}
	}

	m.signingKeys = validKeys
}

// pruneEncryptionKeys removes expired encryption keys.
func (m *KeyRotationManager) pruneEncryptionKeys() {
	now := time.Now().UTC()
	validKeys := make([]*EncryptionKey, 0)

	for _, key := range m.encryptionKeys {
		if now.Before(key.ExpiresAt) {
			validKeys = append(validKeys, key)
		}
	}

	m.encryptionKeys = validKeys
}

// StartAutoRotation starts automatic key rotation based on policy.
func (m *KeyRotationManager) StartAutoRotation(ctx context.Context, algorithm string) {
	if !m.policy.AutoRotationEnabled {
		return
	}

	ticker := time.NewTicker(m.policy.RotationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Rotate signing key.
			if err := m.RotateSigningKey(ctx, algorithm); err != nil {
				// Log error (would integrate with telemetry).
				continue
			}

			// Rotate encryption key.
			if err := m.RotateEncryptionKey(ctx); err != nil {
				// Log error (would integrate with telemetry).
				continue
			}
		}
	}
}

// convertToJWK converts a SigningKey to JWK format for JWKS endpoint.
func convertToJWK(key *SigningKey) map[string]any {
	if key == nil || key.Key == nil {
		return nil
	}

	jwk := map[string]any{
		"kid": key.KeyID,
		"use": "sig",
		"alg": key.Algorithm,
	}

	switch k := key.Key.(type) {
	case *rsa.PrivateKey:
		jwk["kty"] = cryptoutilIdentityMagic.KeyTypeRSA
		jwk["n"] = base64URLEncode(k.N.Bytes())
		jwk["e"] = base64URLEncode(big.NewInt(int64(k.E)).Bytes())
	case *rsa.PublicKey:
		jwk["kty"] = cryptoutilIdentityMagic.KeyTypeRSA
		jwk["n"] = base64URLEncode(k.N.Bytes())
		jwk["e"] = base64URLEncode(big.NewInt(int64(k.E)).Bytes())
	case *ecdsa.PrivateKey:
		jwk["kty"] = cryptoutilIdentityMagic.KeyTypeEC
		jwk["crv"] = ecdsaCurveName(k.Curve)
		jwk["x"] = base64URLEncode(k.X.Bytes())
		jwk["y"] = base64URLEncode(k.Y.Bytes())
	case *ecdsa.PublicKey:
		jwk["kty"] = cryptoutilIdentityMagic.KeyTypeEC
		jwk["crv"] = ecdsaCurveName(k.Curve)
		jwk["x"] = base64URLEncode(k.X.Bytes())
		jwk["y"] = base64URLEncode(k.Y.Bytes())
	case []byte:
		// HMAC keys - don't expose in JWKS (symmetric)
		return nil
	default:
		return nil
	}

	return jwk
}

// base64URLEncode encodes bytes to base64url without padding.
func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// ecdsaCurveName returns the curve name for JWK.
func ecdsaCurveName(curve elliptic.Curve) string {
	switch curve {
	case elliptic.P256():
		return "P-256"
	case elliptic.P384():
		return "P-384"
	case elliptic.P521():
		return "P-521"
	default:
		return ""
	}
}
