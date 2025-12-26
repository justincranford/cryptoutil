// Copyright (c) 2025 Justin Cranford

package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateECDHKeyPair(t *testing.T) {
	t.Parallel()

	privateKey, publicKeyBytes, err := GenerateECDHKeyPair()
	require.NoError(t, err)
	require.NotNil(t, privateKey)
	require.NotNil(t, publicKeyBytes)
	require.Equal(t, 65, len(publicKeyBytes), "P-256 public key should be 65 bytes (X9.62 uncompressed format)")
}

func TestParseECDHPublicKey(t *testing.T) {
	t.Parallel()

	// Generate a key pair first.
	_, publicKeyBytes, err := GenerateECDHKeyPair()
	require.NoError(t, err)

	// Parse the public key bytes.
	publicKey, err := ParseECDHPublicKey(publicKeyBytes)
	require.NoError(t, err)
	require.NotNil(t, publicKey)

	// Verify the parsed key matches original bytes.
	parsedBytes := publicKey.Bytes()
	require.True(t, bytes.Equal(publicKeyBytes, parsedBytes))
}

func TestParseECDHPublicKey_InvalidBytes(t *testing.T) {
	t.Parallel()

	// Test with invalid key bytes.
	invalidBytes := []byte{1, 2, 3}
	_, err := ParseECDHPublicKey(invalidBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse ECDH public key")
}

func TestParseECDHPrivateKey(t *testing.T) {
	t.Parallel()

	// Generate a key pair first.
	privateKey, _, err := GenerateECDHKeyPair()
	require.NoError(t, err)

	privateKeyBytes := privateKey.Bytes()

	// Parse the private key bytes.
	parsedPrivateKey, err := ParseECDHPrivateKey(privateKeyBytes)
	require.NoError(t, err)
	require.NotNil(t, parsedPrivateKey)

	// Verify the parsed key matches original bytes.
	parsedBytes := parsedPrivateKey.Bytes()
	require.True(t, bytes.Equal(privateKeyBytes, parsedBytes))
}

func TestParseECDHPrivateKey_InvalidBytes(t *testing.T) {
	t.Parallel()

	// Test with invalid key bytes.
	invalidBytes := []byte{1, 2, 3}
	_, err := ParseECDHPrivateKey(invalidBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse ECDH private key")
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	t.Parallel()

	// Generate receiver's key pair.
	receiverPrivateKey, receiverPublicKeyBytes, err := GenerateECDHKeyPair()
	require.NoError(t, err)

	receiverPublicKey, err := ParseECDHPublicKey(receiverPublicKeyBytes)
	require.NoError(t, err)

	// Encrypt a message.
	plaintext := []byte("Hello, Learn-IM!")
	ephemeralPublicKeyBytes, ciphertext, nonce, err := EncryptMessage(plaintext, receiverPublicKey)
	require.NoError(t, err)
	require.NotNil(t, ephemeralPublicKeyBytes)
	require.NotNil(t, ciphertext)
	require.NotNil(t, nonce)
	require.NotEqual(t, plaintext, ciphertext, "ciphertext should differ from plaintext")

	// Decrypt the message.
	decrypted, err := DecryptMessage(ciphertext, nonce, ephemeralPublicKeyBytes, receiverPrivateKey)
	require.NoError(t, err)
	require.True(t, bytes.Equal(plaintext, decrypted), "decrypted plaintext should match original")
}

func TestEncryptMessage_DifferentEphemeralKeys(t *testing.T) {
	t.Parallel()

	// Generate receiver's key pair.
	_, receiverPublicKeyBytes, err := GenerateECDHKeyPair()
	require.NoError(t, err)

	receiverPublicKey, err := ParseECDHPublicKey(receiverPublicKeyBytes)
	require.NoError(t, err)

	// Encrypt the same message twice.
	plaintext := []byte("Test message")
	ephem1, cipher1, nonce1, err := EncryptMessage(plaintext, receiverPublicKey)
	require.NoError(t, err)

	ephem2, cipher2, nonce2, err := EncryptMessage(plaintext, receiverPublicKey)
	require.NoError(t, err)

	// Ephemeral keys should be different.
	require.False(t, bytes.Equal(ephem1, ephem2), "ephemeral keys should differ")

	// Ciphertexts should be different (different ephemeral keys).
	require.False(t, bytes.Equal(cipher1, cipher2), "ciphertexts should differ")

	// Nonces should be different.
	require.False(t, bytes.Equal(nonce1, nonce2), "nonces should differ")
}

func TestDecryptMessage_WrongPrivateKey(t *testing.T) {
	t.Parallel()

	// Generate receiver's key pair.
	_, receiverPublicKeyBytes, err := GenerateECDHKeyPair()
	require.NoError(t, err)

	receiverPublicKey, err := ParseECDHPublicKey(receiverPublicKeyBytes)
	require.NoError(t, err)

	// Encrypt a message.
	plaintext := []byte("Secret message")
	ephemeralPublicKeyBytes, ciphertext, nonce, err := EncryptMessage(plaintext, receiverPublicKey)
	require.NoError(t, err)

	// Try to decrypt with a different (wrong) private key.
	wrongPrivateKey, _, err := GenerateECDHKeyPair()
	require.NoError(t, err)

	decrypted, err := DecryptMessage(ciphertext, nonce, ephemeralPublicKeyBytes, wrongPrivateKey)
	require.Error(t, err)
	require.Nil(t, decrypted)
	require.Contains(t, err.Error(), "failed to decrypt message")
}

func TestDecryptMessage_TamperedCiphertext(t *testing.T) {
	t.Parallel()

	// Generate receiver's key pair.
	receiverPrivateKey, receiverPublicKeyBytes, err := GenerateECDHKeyPair()
	require.NoError(t, err)

	receiverPublicKey, err := ParseECDHPublicKey(receiverPublicKeyBytes)
	require.NoError(t, err)

	// Encrypt a message.
	plaintext := []byte("Authenticated message")
	ephemeralPublicKeyBytes, ciphertext, nonce, err := EncryptMessage(plaintext, receiverPublicKey)
	require.NoError(t, err)

	// Tamper with the ciphertext.
	tamperedCiphertext := make([]byte, len(ciphertext))
	copy(tamperedCiphertext, ciphertext)
	tamperedCiphertext[0] ^= 1 // Flip one bit.

	// Decryption should fail due to authentication tag mismatch.
	decrypted, err := DecryptMessage(tamperedCiphertext, nonce, ephemeralPublicKeyBytes, receiverPrivateKey)
	require.Error(t, err)
	require.Nil(t, decrypted)
	require.Contains(t, err.Error(), "failed to decrypt message")
}

func TestEncryptDecrypt_MultipleReceivers(t *testing.T) {
	t.Parallel()

	// Simulate encrypting for multiple receivers.
	receiver1PrivateKey, receiver1PublicKeyBytes, err := GenerateECDHKeyPair()
	require.NoError(t, err)

	receiver2PrivateKey, receiver2PublicKeyBytes, err := GenerateECDHKeyPair()
	require.NoError(t, err)

	receiver1PublicKey, err := ParseECDHPublicKey(receiver1PublicKeyBytes)
	require.NoError(t, err)

	receiver2PublicKey, err := ParseECDHPublicKey(receiver2PublicKeyBytes)
	require.NoError(t, err)

	// Encrypt for receiver 1.
	plaintext := []byte("Multi-receiver message")
	ephem1, cipher1, nonce1, err := EncryptMessage(plaintext, receiver1PublicKey)
	require.NoError(t, err)

	// Encrypt for receiver 2 (same plaintext, different receiver).
	ephem2, cipher2, nonce2, err := EncryptMessage(plaintext, receiver2PublicKey)
	require.NoError(t, err)

	// Receiver 1 should decrypt successfully.
	decrypted1, err := DecryptMessage(cipher1, nonce1, ephem1, receiver1PrivateKey)
	require.NoError(t, err)
	require.True(t, bytes.Equal(plaintext, decrypted1))

	// Receiver 2 should decrypt successfully.
	decrypted2, err := DecryptMessage(cipher2, nonce2, ephem2, receiver2PrivateKey)
	require.NoError(t, err)
	require.True(t, bytes.Equal(plaintext, decrypted2))

	// Receiver 1 cannot decrypt receiver 2's message.
	_, err = DecryptMessage(cipher2, nonce2, ephem2, receiver1PrivateKey)
	require.Error(t, err)

	// Receiver 2 cannot decrypt receiver 1's message.
	_, err = DecryptMessage(cipher1, nonce1, ephem1, receiver2PrivateKey)
	require.Error(t, err)
}

func TestDecryptMessage_InvalidEphemeralPublicKey(t *testing.T) {
	t.Parallel()

	// Generate receiver's key pair.
	receiverPrivateKey, receiverPublicKeyBytes, err := GenerateECDHKeyPair()
	require.NoError(t, err)

	receiverPublicKey, err := ParseECDHPublicKey(receiverPublicKeyBytes)
	require.NoError(t, err)

	// Encrypt a message.
	plaintext := []byte("Test message")
	_, ciphertext, nonce, err := EncryptMessage(plaintext, receiverPublicKey)
	require.NoError(t, err)

	// Try to decrypt with invalid ephemeral public key bytes.
	invalidEphemeralKey := []byte{1, 2, 3} // Too short, invalid format.
	decrypted, err := DecryptMessage(ciphertext, nonce, invalidEphemeralKey, receiverPrivateKey)
	require.Error(t, err)
	require.Nil(t, decrypted)
	require.Contains(t, err.Error(), "failed to parse ephemeral public key")
}

func TestEncryptDecrypt_EmptyMessage(t *testing.T) {
	t.Parallel()

	// Generate receiver's key pair.
	receiverPrivateKey, receiverPublicKeyBytes, err := GenerateECDHKeyPair()
	require.NoError(t, err)

	receiverPublicKey, err := ParseECDHPublicKey(receiverPublicKeyBytes)
	require.NoError(t, err)

	// Encrypt an empty message.
	plaintext := []byte("")
	ephemeralPublicKeyBytes, ciphertext, nonce, err := EncryptMessage(plaintext, receiverPublicKey)
	require.NoError(t, err)
	require.NotNil(t, ephemeralPublicKeyBytes)
	require.NotNil(t, ciphertext)
	require.NotNil(t, nonce)

	// Decrypt the empty message.
	decrypted, err := DecryptMessage(ciphertext, nonce, ephemeralPublicKeyBytes, receiverPrivateKey)
	require.NoError(t, err)
	require.True(t, bytes.Equal(plaintext, decrypted), "empty message should decrypt correctly")
	require.Equal(t, 0, len(decrypted), "decrypted message should be empty")
}

func TestEncryptDecrypt_LargeMessage(t *testing.T) {
	t.Parallel()

	// Generate receiver's key pair.
	receiverPrivateKey, receiverPublicKeyBytes, err := GenerateECDHKeyPair()
	require.NoError(t, err)

	receiverPublicKey, err := ParseECDHPublicKey(receiverPublicKeyBytes)
	require.NoError(t, err)

	// Encrypt a large message (1 MB).
	plaintext := make([]byte, 1024*1024)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	ephemeralPublicKeyBytes, ciphertext, nonce, err := EncryptMessage(plaintext, receiverPublicKey)
	require.NoError(t, err)
	require.NotNil(t, ephemeralPublicKeyBytes)
	require.NotNil(t, ciphertext)
	require.NotNil(t, nonce)
	require.Greater(t, len(ciphertext), len(plaintext), "ciphertext includes GCM tag")

	// Decrypt the large message.
	decrypted, err := DecryptMessage(ciphertext, nonce, ephemeralPublicKeyBytes, receiverPrivateKey)
	require.NoError(t, err)
	require.True(t, bytes.Equal(plaintext, decrypted), "large message should decrypt correctly")
	require.Equal(t, len(plaintext), len(decrypted))
}
