// Copyright (c) 2025 Justin Cranford
//
//

package issuer

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	"github.com/stretchr/testify/require"
)

// FuzzJWEEncryptionDecryption tests JWE encryption and decryption with various inputs.
func FuzzJWEEncryptionDecryption(f *testing.F) {
	// Seed corpus with various plaintext values.
	f.Add("short")
	f.Add("")
	f.Add("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature")
	f.Add(string(make([]byte, cryptoutilSharedMagic.JoseJADefaultListLimit)))
	f.Add("special-chars-!@#$%^&*()[]{}|\\:;\"'<>,.?/~`")
	f.Add("unicode-测试-test-🔐")

	// Create legacy JWE issuer.
	encryptionKey := []byte("01234567890123456789012345678901") // 32 bytes.

	issuer, err := NewJWEIssuerLegacy(encryptionKey)
	if err != nil {
		f.Fatalf("failed to create issuer: %v", err)
	}

	ctx := context.Background()

	f.Fuzz(func(t *testing.T, plaintext string) {
		// Encrypt token - should not panic.
		encrypted, err := issuer.EncryptToken(ctx, plaintext)
		if err != nil {
			return
		}

		// Decrypt token - should not panic.
		decrypted, err := issuer.DecryptToken(ctx, encrypted)
		if err != nil {
			return
		}

		// Verify roundtrip if successful.
		require.Equal(t, plaintext, decrypted, "roundtrip failed")
	})
}

// FuzzJWEDecryptionInvalidInputs tests JWE decryption with invalid inputs.
func FuzzJWEDecryptionInvalidInputs(f *testing.F) {
	// Seed corpus with invalid encrypted tokens.
	f.Add("invalid-base64")
	f.Add("")
	f.Add("YWJjZGVm") // Valid base64 but invalid ciphertext.
	f.Add("!!!!")
	f.Add(string(make([]byte, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)))

	// Create legacy JWE issuer.
	encryptionKey := []byte("01234567890123456789012345678901") // 32 bytes.

	issuer, err := NewJWEIssuerLegacy(encryptionKey)
	if err != nil {
		f.Fatalf("failed to create issuer: %v", err)
	}

	ctx := context.Background()

	f.Fuzz(func(_ *testing.T, encryptedToken string) {
		// Decrypt token - should not panic.
		_, err := issuer.DecryptToken(ctx, encryptedToken)

		// We don't care about the error, just that it doesn't panic.
		_ = err
	})
}

// FuzzJWEKeyIDHandling tests JWE key ID extraction with various inputs.
func FuzzJWEKeyIDHandling(f *testing.F) {
	// Seed corpus with various key ID formats.
	f.Add(uint8(0), uint8(0), "plaintext")
	f.Add(uint8(1), uint8(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries), "test")
	f.Add(uint8(cryptoutilSharedMagic.HKDFMaxMultiplier), uint8(cryptoutilSharedMagic.HKDFMaxMultiplier), "data")
	f.Add(uint8(0), uint8(cryptoutilSharedMagic.HKDFMaxMultiplier), "content")

	// Create legacy JWE issuer.
	encryptionKey := []byte("01234567890123456789012345678901") // 32 bytes.

	issuer, err := NewJWEIssuerLegacy(encryptionKey)
	if err != nil {
		f.Fatalf("failed to create issuer: %v", err)
	}

	ctx := context.Background()

	f.Fuzz(func(_ *testing.T, byte1, byte2 uint8, plaintext string) {
		// Encrypt token first.
		encrypted, err := issuer.EncryptToken(ctx, plaintext)
		if err != nil {
			return
		}

		// Tamper with key ID bytes.
		if len(encrypted) >= 2 {
			tamperedBytes := []byte(encrypted)
			tamperedBytes[0] = byte1
			tamperedBytes[1] = byte2
			tamperedEncrypted := string(tamperedBytes)

			// Attempt decryption - should not panic.
			_, err = issuer.DecryptToken(ctx, tamperedEncrypted)

			// We don't care about the error, just that it doesn't panic.
			_ = err
		}
	})
}
