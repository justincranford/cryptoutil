package issuer

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"encoding/base64"
	"fmt"

	cryptoutilIdentityApperr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// JWEIssuer issues JWE (encrypted) tokens using AES-GCM encryption.
type JWEIssuer struct {
	encryptionKey []byte
}

// NewJWEIssuer creates a new JWE issuer with the specified encryption key.
func NewJWEIssuer(encryptionKey []byte) (*JWEIssuer, error) {
	if len(encryptionKey) != cryptoutilIdentityMagic.AES256KeySize {
		return nil, cryptoutilIdentityApperr.WrapError(
			cryptoutilIdentityApperr.ErrInvalidConfiguration,
			fmt.Errorf("encryption key must be %d bytes (AES-256), got %d bytes", cryptoutilIdentityMagic.AES256KeySize, len(encryptionKey)),
		)
	}

	return &JWEIssuer{
		encryptionKey: encryptionKey,
	}, nil
}

// EncryptToken encrypts a plaintext token (e.g., JWS) using AES-GCM.
func (i *JWEIssuer) EncryptToken(ctx context.Context, plaintext string) (string, error) {
	// Create AES cipher.
	block, err := aes.NewCipher(i.encryptionKey)
	if err != nil {
		return "", cryptoutilIdentityApperr.WrapError(
			cryptoutilIdentityApperr.ErrTokenIssuanceFailed,
			fmt.Errorf("failed to create AES cipher: %w", err),
		)
	}

	// Create GCM mode.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", cryptoutilIdentityApperr.WrapError(
			cryptoutilIdentityApperr.ErrTokenIssuanceFailed,
			fmt.Errorf("failed to create GCM mode: %w", err),
		)
	}

	// Generate nonce.
	nonce := make([]byte, gcm.NonceSize())
	if _, err := crand.Read(nonce); err != nil {
		return "", cryptoutilIdentityApperr.WrapError(
			cryptoutilIdentityApperr.ErrTokenIssuanceFailed,
			fmt.Errorf("failed to generate nonce: %w", err),
		)
	}

	// Encrypt plaintext.
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode as base64.
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

// DecryptToken decrypts a JWE token using AES-GCM.
func (i *JWEIssuer) DecryptToken(ctx context.Context, encryptedToken string) (string, error) {
	// Decode base64.
	ciphertext, err := base64.RawURLEncoding.DecodeString(encryptedToken)
	if err != nil {
		return "", cryptoutilIdentityApperr.WrapError(
			cryptoutilIdentityApperr.ErrTokenValidationFailed,
			fmt.Errorf("failed to decode base64: %w", err),
		)
	}

	// Create AES cipher.
	block, err := aes.NewCipher(i.encryptionKey)
	if err != nil {
		return "", cryptoutilIdentityApperr.WrapError(
			cryptoutilIdentityApperr.ErrTokenValidationFailed,
			fmt.Errorf("failed to create AES cipher: %w", err),
		)
	}

	// Create GCM mode.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", cryptoutilIdentityApperr.WrapError(
			cryptoutilIdentityApperr.ErrTokenValidationFailed,
			fmt.Errorf("failed to create GCM mode: %w", err),
		)
	}

	// Extract nonce.
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", cryptoutilIdentityApperr.ErrInvalidToken
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt ciphertext.
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", cryptoutilIdentityApperr.WrapError(
			cryptoutilIdentityApperr.ErrTokenValidationFailed,
			fmt.Errorf("failed to decrypt token: %w", err),
		)
	}

	return string(plaintext), nil
}
