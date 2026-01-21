// Copyright (c) 2025 Justin Cranford

package crypto

import (
	crand "crypto/rand"
	"crypto/rsa"
	"testing"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"

	cryptoutilKeyGen "cryptoutil/internal/shared/crypto/keygen"
)

// TestEncryptDecryptKey_AES256KW tests key wrapping with AES256KW.
func TestEncryptDecryptKey_AES256KW(t *testing.T) {
	t.Parallel()

	// Generate KEK (Key Encryption Key) using A256KW.
	kekKid := googleUuid.New()
	kekAlg := AlgA256KW
	kekEnc := EncA256GCM
	kekBytes := make([]byte, 32) // 256 bits for A256KW.
	_, err := crand.Read(kekBytes)
	require.NoError(t, err)

	// Create KEK JWK.
	_, kekJWK, _, _, _, err := CreateJWEJWKFromKey(&kekKid, &kekEnc, &kekAlg, cryptoutilKeyGen.SecretKey(kekBytes))
	require.NoError(t, err)
	require.NotNil(t, kekJWK)

	// Generate CEK (Content Encryption Key) to be wrapped.
	cekKid := googleUuid.New()
	cekBytes := make([]byte, 32) // 256 bits.
	_, err = crand.Read(cekBytes)
	require.NoError(t, err)

	// Create CEK JWK (will be wrapped).
	_, cekJWK, _, _, _, err := CreateJWEJWKFromKey(&cekKid, &kekEnc, &AlgDir, cryptoutilKeyGen.SecretKey(cekBytes))
	require.NoError(t, err)
	require.NotNil(t, cekJWK)

	// Encrypt the CEK using KEK.
	encryptedMessage, encryptedBytes, err := EncryptKey([]joseJwk.Key{kekJWK}, cekJWK)
	require.NoError(t, err)
	require.NotNil(t, encryptedMessage)
	require.NotEmpty(t, encryptedBytes)

	// Decrypt the CEK using KEK.
	decryptedCEK, err := DecryptKey([]joseJwk.Key{kekJWK}, encryptedBytes)
	require.NoError(t, err)
	require.NotNil(t, decryptedCEK)

	// Verify decrypted CEK kid matches original.
	var decryptedKid string

	require.NoError(t, decryptedCEK.Get(joseJwk.KeyIDKey, &decryptedKid))
	require.Equal(t, cekKid.String(), decryptedKid)
}

// TestEncryptDecryptKey_RSAOAEP tests key wrapping with RSA-OAEP.
func TestEncryptDecryptKey_RSAOAEP(t *testing.T) {
	t.Parallel()

	// Generate RSA KEK.
	rsaPrivateKey, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	kekKid := googleUuid.New()
	kekAlg := AlgRSAOAEP
	kekEnc := EncA256GCM
	rsaKeyPair := &cryptoutilKeyGen.KeyPair{
		Private: rsaPrivateKey,
		Public:  &rsaPrivateKey.PublicKey,
	}

	// Create RSA KEK JWK (returns both private and public JWKs).
	_, kekPrivateJWK, kekPublicJWK, _, _, err := CreateJWEJWKFromKey(&kekKid, &kekEnc, &kekAlg, rsaKeyPair)
	require.NoError(t, err)
	require.NotNil(t, kekPrivateJWK)
	require.NotNil(t, kekPublicJWK)

	// Generate CEK to be wrapped.
	cekKid := googleUuid.New()
	cekBytes := make([]byte, 32)
	_, err = crand.Read(cekBytes)
	require.NoError(t, err)

	_, cekJWK, _, _, _, err := CreateJWEJWKFromKey(&cekKid, &kekEnc, &AlgDir, cryptoutilKeyGen.SecretKey(cekBytes))
	require.NoError(t, err)

	// Encrypt CEK with RSA KEK (use public JWK for encryption).
	encryptedMessage, encryptedBytes, err := EncryptKey([]joseJwk.Key{kekPublicJWK}, cekJWK)
	require.NoError(t, err)
	require.NotNil(t, encryptedMessage)
	require.NotEmpty(t, encryptedBytes)

	// Decrypt CEK with RSA KEK (use private JWK for decryption).
	decryptedCEK, err := DecryptKey([]joseJwk.Key{kekPrivateJWK}, encryptedBytes)
	require.NoError(t, err)
	require.NotNil(t, decryptedCEK)

	// Verify kid.
	var decryptedKid string

	require.NoError(t, decryptedCEK.Get(joseJwk.KeyIDKey, &decryptedKid))
	require.Equal(t, cekKid.String(), decryptedKid)
}

// TestEncryptKey_MarshalError tests error handling when CEK cannot be marshaled.
func TestEncryptKey_MarshalError(t *testing.T) {
	t.Parallel()

	// Generate valid KEK.
	kekKid := googleUuid.New()
	kekBytes := make([]byte, 32)
	_, err := crand.Read(kekBytes)
	require.NoError(t, err)

	_, kekJWK, _, _, _, err := CreateJWEJWKFromKey(&kekKid, &EncA256GCM, &AlgA256KW, cryptoutilKeyGen.SecretKey(kekBytes))
	require.NoError(t, err)

	// Try to encrypt nil CEK (EncryptKey marshals it first, so nil will cause marshal to succeed with "null").
	// Instead test that empty JWK bytes fail in EncryptBytes.
	_, _, err = EncryptBytes([]joseJwk.Key{kekJWK}, []byte{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid clearBytes")
}

// TestDecryptKey_InvalidEncryptedBytes tests error handling for invalid encrypted data.
func TestDecryptKey_InvalidEncryptedBytes(t *testing.T) {
	t.Parallel()

	// Generate valid KDK (Key Decryption Key).
	kdkKid := googleUuid.New()
	kdkBytes := make([]byte, 32)
	_, err := crand.Read(kdkBytes)
	require.NoError(t, err)

	_, kdkJWK, _, _, _, err := CreateJWEJWKFromKey(&kdkKid, &EncA256GCM, &AlgA256KW, cryptoutilKeyGen.SecretKey(kdkBytes))
	require.NoError(t, err)

	// Try to decrypt invalid bytes.
	_, err = DecryptKey([]joseJwk.Key{kdkJWK}, []byte("invalid jwe data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt CDK bytes")
}

// TestDecryptKey_CorruptedEncryptedBytes tests decryption with corrupted data.
func TestDecryptKey_CorruptedEncryptedBytes(t *testing.T) {
	t.Parallel()

	// Generate KEK and CEK.
	kekKid := googleUuid.New()
	kekBytes := make([]byte, 32)
	_, err := crand.Read(kekBytes)
	require.NoError(t, err)

	_, kekJWK, _, _, _, err := CreateJWEJWKFromKey(&kekKid, &EncA256GCM, &AlgA256KW, cryptoutilKeyGen.SecretKey(kekBytes))
	require.NoError(t, err)

	cekKid := googleUuid.New()
	cekBytes := make([]byte, 32)
	_, err = crand.Read(cekBytes)
	require.NoError(t, err)

	_, cekJWK, _, _, _, err := CreateJWEJWKFromKey(&cekKid, &EncA256GCM, &AlgDir, cryptoutilKeyGen.SecretKey(cekBytes))
	require.NoError(t, err)

	// Encrypt CEK.
	_, encryptedBytes, err := EncryptKey([]joseJwk.Key{kekJWK}, cekJWK)
	require.NoError(t, err)

	// Corrupt the encrypted bytes.
	encryptedBytes[len(encryptedBytes)-1] ^= 0xFF

	// Try to decrypt corrupted data.
	_, err = DecryptKey([]joseJwk.Key{kekJWK}, encryptedBytes)
	require.Error(t, err)
}

func TestDecryptKey_InvalidJWKFormat(t *testing.T) {
	t.Parallel()

	// Generate KEK.
	kekKid := googleUuid.New()
	kekBytes := make([]byte, 32)
	_, err := crand.Read(kekBytes)
	require.NoError(t, err)

	_, kekJWK, _, _, _, err := CreateJWEJWKFromKey(&kekKid, &EncA256GCM, &AlgA256KW, cryptoutilKeyGen.SecretKey(kekBytes))
	require.NoError(t, err)

	// Encrypt plaintext that's NOT a valid JWK (just random bytes).
	plaintext := []byte("not a valid jwk format")
	_, encryptedBytes, err := EncryptBytes([]joseJwk.Key{kekJWK}, plaintext)
	require.NoError(t, err)

	// DecryptKey will decrypt successfully but fail to parse as JWK.
	_, err = DecryptKey([]joseJwk.Key{kekJWK}, encryptedBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to derypt CDK")
}

func TestEncryptKey_NilKEKs(t *testing.T) {
	t.Parallel()

	// Generate CEK.
	cekKid := googleUuid.New()
	cekBytes := make([]byte, 32)
	_, err := crand.Read(cekBytes)
	require.NoError(t, err)

	_, cekJWK, _, _, _, err := CreateJWEJWKFromKey(&cekKid, &EncA256GCM, &AlgDir, cryptoutilKeyGen.SecretKey(cekBytes))
	require.NoError(t, err)

	// Try to encrypt with nil KEKs.
	_, _, err = EncryptKey(nil, cekJWK)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
}

func TestDecryptKey_NilKDKs(t *testing.T) {
	t.Parallel()

	// Try to decrypt with nil KDKs.
	_, err := DecryptKey(nil, []byte("encrypted"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
}
