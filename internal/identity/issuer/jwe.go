// Copyright (c) 2025 Justin Cranford
//
//

// Package issuer provides token issuance services for OAuth 2.0 and OIDC.
package issuer

import (
	"context"
	aes "crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"encoding/base64"
	"fmt"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// JWEIssuer issues JWE (encrypted) tokens using versioned encryption keys.
type JWEIssuer struct {
	keyRotationMgr      *KeyRotationManager
	legacyEncryptionKey []byte // Deprecated: for backward compatibility.
}

// NewJWEIssuer creates a new JWE issuer with the specified key rotation manager.
func NewJWEIssuer(keyRotationMgr *KeyRotationManager) (*JWEIssuer, error) {
	// Key rotation manager is optional for backward compatibility.
	if keyRotationMgr == nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrInvalidConfiguration,
			fmt.Errorf("key rotation manager is required; use NewJWEIssuerLegacy for backward compatibility"),
		)
	}

	return &JWEIssuer{
		keyRotationMgr: keyRotationMgr,
	}, nil
}

// NewJWEIssuerLegacy creates a new JWE issuer with a single encryption key (deprecated).
func NewJWEIssuerLegacy(encryptionKey []byte) (*JWEIssuer, error) {
	if len(encryptionKey) != cryptoutilIdentityMagic.AES256KeySize {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrInvalidConfiguration,
			fmt.Errorf("encryption key must be %d bytes (AES-256), got %d bytes", cryptoutilIdentityMagic.AES256KeySize, len(encryptionKey)),
		)
	}

	return &JWEIssuer{
		legacyEncryptionKey: encryptionKey,
	}, nil
}

// EncryptToken encrypts a plaintext token (e.g., JWS) using AES-GCM with active encryption key.
func (i *JWEIssuer) EncryptToken(_ context.Context, plaintext string) (string, error) {
	var encryptionKey []byte

	var keyID string

	// Get active encryption key (or use legacy key).
	if i.keyRotationMgr != nil {
		activeKey, err := i.keyRotationMgr.GetActiveEncryptionKey()
		if err != nil {
			return "", cryptoutilIdentityAppErr.WrapError(
				cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
				fmt.Errorf("failed to get active encryption key: %w", err),
			)
		}

		encryptionKey = activeKey.Key
		keyID = activeKey.KeyID
	} else {
		// Legacy mode: use single encryption key.
		encryptionKey = i.legacyEncryptionKey
		keyID = "" // No key ID in legacy mode.
	}

	// Create AES cipher.
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
			fmt.Errorf("failed to create AES cipher: %w", err),
		)
	}

	// Create GCM mode.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
			fmt.Errorf("failed to create GCM mode: %w", err),
		)
	}

	// Generate nonce.
	nonce := make([]byte, gcm.NonceSize())
	if _, err := crand.Read(nonce); err != nil {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
			fmt.Errorf("failed to generate nonce: %w", err),
		)
	}

	// Encrypt plaintext.
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Prepend key ID if available (format: keyID||ciphertext).
	if keyID != "" {
		keyIDBytes := []byte(keyID)
		keyIDLen := len(keyIDBytes)
		result := make([]byte, 2+keyIDLen+len(ciphertext))
		result[0] = byte(keyIDLen >> cryptoutilIdentityMagic.ByteShift)
		result[1] = byte(keyIDLen)
		copy(result[2:], keyIDBytes)
		copy(result[2+keyIDLen:], ciphertext)
		ciphertext = result
	}

	// Encode as base64.
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

// DecryptToken decrypts a JWE token using AES-GCM with key ID lookup.
func (i *JWEIssuer) DecryptToken(_ context.Context, encryptedToken string) (string, error) {
	// Decode base64.
	ciphertext, err := base64.RawURLEncoding.DecodeString(encryptedToken)
	if err != nil {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenValidationFailed,
			fmt.Errorf("failed to decode base64: %w", err),
		)
	}

	var encryptionKey []byte

	var actualCiphertext []byte

	// Extract key ID if present (format: keyID||ciphertext).
	if i.keyRotationMgr != nil && len(ciphertext) > 2 {
		keyIDLen := int(ciphertext[0])<<cryptoutilIdentityMagic.ByteShift | int(ciphertext[1])
		if keyIDLen > 0 && len(ciphertext) > 2+keyIDLen {
			keyID := string(ciphertext[2 : 2+keyIDLen])
			actualCiphertext = ciphertext[2+keyIDLen:]

			// Get encryption key by ID.
			key, err := i.keyRotationMgr.GetEncryptionKeyByID(keyID)
			if err != nil {
				return "", cryptoutilIdentityAppErr.WrapError(
					cryptoutilIdentityAppErr.ErrTokenValidationFailed,
					fmt.Errorf("failed to get encryption key: %w", err),
				)
			}

			encryptionKey = key.Key
		} else {
			// No key ID, use active key.
			activeKey, err := i.keyRotationMgr.GetActiveEncryptionKey()
			if err != nil {
				return "", cryptoutilIdentityAppErr.WrapError(
					cryptoutilIdentityAppErr.ErrTokenValidationFailed,
					fmt.Errorf("failed to get active encryption key: %w", err),
				)
			}

			encryptionKey = activeKey.Key
			actualCiphertext = ciphertext
		}
	} else {
		// Legacy mode: use single encryption key.
		encryptionKey = i.legacyEncryptionKey
		actualCiphertext = ciphertext
	}

	// Create AES cipher.
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenValidationFailed,
			fmt.Errorf("failed to create AES cipher: %w", err),
		)
	}

	// Create GCM mode.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenValidationFailed,
			fmt.Errorf("failed to create GCM mode: %w", err),
		)
	}

	// Extract nonce.
	nonceSize := gcm.NonceSize()
	if len(actualCiphertext) < nonceSize {
		return "", cryptoutilIdentityAppErr.ErrInvalidToken
	}

	nonce, encryptedData := actualCiphertext[:nonceSize], actualCiphertext[nonceSize:]

	// Decrypt ciphertext.
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenValidationFailed,
			fmt.Errorf("failed to decrypt token: %w", err),
		)
	}

	return string(plaintext), nil
}
