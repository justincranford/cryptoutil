// Copyright (c) 2025 Justin Cranford

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	crand "crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// EncryptMessage encrypts a plaintext message for a receiver using hybrid encryption.
// Steps:
// 1. Generate ephemeral ECDH key pair
// 2. Perform ECDH with receiver's public key to get shared secret
// 3. Derive AES-256 key using HKDF-SHA256
// 4. Encrypt message with AES-256-GCM
// Returns: ephemeral public key bytes, ciphertext, nonce, error.
func EncryptMessage(plaintext []byte, receiverPublicKey *ecdh.PublicKey) ([]byte, []byte, []byte, error) {
	// Generate ephemeral ECDH key pair.
	curve := ecdh.P256()

	ephemeralPrivateKey, err := curve.GenerateKey(crand.Reader)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	// Perform ECDH to get shared secret.
	sharedSecret, err := ephemeralPrivateKey.ECDH(receiverPublicKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to perform ECDH: %w", err)
	}

	// Derive AES-256 key using HKDF-SHA256.
	kdf := hkdf.New(sha256.New, sharedSecret, nil, []byte("learn-im-message-encryption"))

	aesKey := make([]byte, 32) // 256 bits.
	if _, err := io.ReadFull(kdf, aesKey); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to derive AES key: %w", err)
	}

	// Create AES-256-GCM cipher.
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce.
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(crand.Reader, nonce); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt plaintext.
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	ephemeralPublicKeyBytes := ephemeralPrivateKey.PublicKey().Bytes()

	return ephemeralPublicKeyBytes, ciphertext, nonce, nil
}

// DecryptMessage decrypts a ciphertext message using the receiver's private key.
// Steps:
// 1. Parse ephemeral public key
// 2. Perform ECDH with receiver's private key to get shared secret
// 3. Derive AES-256 key using HKDF-SHA256
// 4. Decrypt ciphertext with AES-256-GCM
// Returns: plaintext, error.
func DecryptMessage(ciphertext []byte, nonce []byte, ephemeralPublicKeyBytes []byte, receiverPrivateKey *ecdh.PrivateKey) ([]byte, error) {
	// Parse ephemeral public key.
	curve := ecdh.P256()

	ephemeralPublicKey, err := curve.NewPublicKey(ephemeralPublicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ephemeral public key: %w", err)
	}

	// Perform ECDH to get shared secret.
	sharedSecret, err := receiverPrivateKey.ECDH(ephemeralPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to perform ECDH: %w", err)
	}

	// Derive AES-256 key using HKDF-SHA256.
	kdf := hkdf.New(sha256.New, sharedSecret, nil, []byte("learn-im-message-encryption"))

	aesKey := make([]byte, 32) // 256 bits.
	if _, err := io.ReadFull(kdf, aesKey); err != nil {
		return nil, fmt.Errorf("failed to derive AES key: %w", err)
	}

	// Create AES-256-GCM cipher.
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt ciphertext.
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt message: %w", err)
	}

	return plaintext, nil
}
