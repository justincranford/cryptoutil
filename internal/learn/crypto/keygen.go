// Copyright (c) 2025 Justin Cranford

package crypto

import (
	"crypto/ecdh"
	crand "crypto/rand"
	"fmt"
)

// GenerateECDHKeyPair generates a new ECDH P-256 key pair for a user.
// Returns the private key and public key bytes (X9.62 uncompressed format).
func GenerateECDHKeyPair() (*ecdh.PrivateKey, []byte, error) {
	curve := ecdh.P256()

	privateKey, err := curve.GenerateKey(crand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate ECDH key pair: %w", err)
	}

	publicKeyBytes := privateKey.PublicKey().Bytes()

	return privateKey, publicKeyBytes, nil
}

// ParseECDHPublicKey parses a public key from bytes (X9.62 uncompressed format).
func ParseECDHPublicKey(publicKeyBytes []byte) (*ecdh.PublicKey, error) {
	curve := ecdh.P256()

	publicKey, err := curve.NewPublicKey(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ECDH public key: %w", err)
	}

	return publicKey, nil
}

// ParseECDHPrivateKey parses a private key from bytes.
func ParseECDHPrivateKey(privateKeyBytes []byte) (*ecdh.PrivateKey, error) {
	curve := ecdh.P256()

	privateKey, err := curve.NewPrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ECDH private key: %w", err)
	}

	return privateKey, nil
}
