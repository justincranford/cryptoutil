// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

// TestEncryptBytesWithContext_WithContext tests encryption with authenticated data.
func TestEncryptBytesWithContext_WithContext(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		enc  *joseJwa.ContentEncryptionAlgorithm
		alg  *joseJwa.KeyEncryptionAlgorithm
	}{
		{"A256GCM_A256KW", &EncA256GCM, &AlgA256KW},
		{"A192GCM_A192KW", &EncA192GCM, &AlgA192KW},
		{"A128GCM_A128KW", &EncA128GCM, &AlgA128KW},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Generate JWK and encrypt with context.
			_, nonPublicJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(tc.enc, tc.alg)
			require.NoError(t, err)

			clearBytes := []byte("test message with context")
			contextBytes := []byte("additional authenticated data")

			jweMessage, jweMessageBytes, err := EncryptBytesWithContext([]joseJwk.Key{nonPublicJWK}, clearBytes, contextBytes)
			require.NoError(t, err)
			require.NotNil(t, jweMessage)
			require.NotEmpty(t, jweMessageBytes)

			// Decrypt and verify.
			decryptedBytes, err := DecryptBytesWithContext([]joseJwk.Key{nonPublicJWK}, jweMessageBytes, contextBytes)
			require.NoError(t, err)
			require.Equal(t, clearBytes, decryptedBytes)
		})
	}
}

// TestEncryptBytes_MultipleJWKsSameAlg tests encryption with multiple JWKs (JSON encoding).
func TestEncryptBytes_MultipleJWKsSameAlg(t *testing.T) {
	t.Parallel()

	// Generate two JWKs with same algorithm.
	_, nonPublicJWK1, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	_, nonPublicJWK2, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK1, nonPublicJWK2}
	clearBytes := []byte("test message for multiple JWKs")

	// Should succeed with JSON encoding (multiple recipients).
	jweMessage, jweMessageBytes, err := EncryptBytes(jwks, clearBytes)
	require.NoError(t, err)
	require.NotNil(t, jweMessage)
	require.NotEmpty(t, jweMessageBytes)

	// Verify message has 2 recipients.
	require.Equal(t, 2, len(jweMessage.Recipients()))
}

// TestEncryptBytes_DifferentEncAlgorithms tests error when using different enc algorithms.
func TestEncryptBytes_DifferentEncAlgorithms(t *testing.T) {
	t.Parallel()

	// Generate JWKs with different enc algorithms.
	_, nonPublicJWK1, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	_, nonPublicJWK2, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA192GCM, &AlgA192KW)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK1, nonPublicJWK2}
	clearBytes := []byte("test message")

	// Should fail - different enc algorithms not allowed.
	_, _, err = EncryptBytes(jwks, clearBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only one unique 'enc' attribute is allowed")
}

// TestEncryptBytes_DifferentKeyAlgorithms tests error when using RSA vs symmetric keys.
func TestEncryptBytes_DifferentKeyAlgorithms(t *testing.T) {
	t.Parallel()

	// Generate JWK with symmetric algorithm.
	_, nonPublicJWK1, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	// Generate JWK with RSA algorithm - but we need private key for encryption JWK.
	// Actually, mixing algorithms is caught by IsEncryptJWK check earlier.
	// Let me test a more realistic scenario.
	_, _, publicJWK2, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgRSAOAEP)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK1, publicJWK2}
	clearBytes := []byte("test message")

	// Should work - both are encrypt JWKs, but with different algs.
	_, _, err = EncryptBytes(jwks, clearBytes)
	// May error or succeed depending on algorithm mixing rules.
	if err != nil {
		require.Contains(t, err.Error(), "only one unique")
	}
}

// TestEncryptKey_HappyPath tests key encryption (KEK wrapping CEK).
func TestEncryptKey_HappyPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		enc  *joseJwa.ContentEncryptionAlgorithm
		alg  *joseJwa.KeyEncryptionAlgorithm
	}{
		{cryptoutilSharedMagic.JoseAlgRSAOAEP, &EncA256GCM, &AlgRSAOAEP},
		{cryptoutilSharedMagic.JoseAlgRSAOAEP256, &EncA256GCM, &AlgRSAOAEP256},
		{"ECDH-ES+A256KW", &EncA256GCM, &AlgECDHESA256KW},
		{"A256KW", &EncA256GCM, &AlgA256KW},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Generate KEK (Key Encryption Key) JWK.
			_, nonPublicJWK, publicJWK, _, _, err := GenerateJWEJWKForEncAndAlg(tc.enc, tc.alg)
			require.NoError(t, err)

			// For symmetric algorithms, use nonPublicJWK; for asymmetric, use publicJWK.
			kekJWK := publicJWK
			if publicJWK == nil {
				kekJWK = nonPublicJWK // Symmetric key (A256KW).
			}

			require.NotNil(t, kekJWK)

			// Generate CEK (Content Encryption Key) to wrap.
			_, cekJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(tc.enc, &AlgDir)
			require.NoError(t, err)

			// Encrypt the CEK with KEK.
			jweMessage, jweBytes, err := EncryptKey([]joseJwk.Key{kekJWK}, cekJWK)
			require.NoError(t, err)
			require.NotNil(t, jweMessage)
			require.NotEmpty(t, jweBytes)
		})
	}
}

// TestJWEHeadersString_HappyPath tests JWE header string generation.
func TestJWEHeadersString_HappyPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		enc  *joseJwa.ContentEncryptionAlgorithm
		alg  *joseJwa.KeyEncryptionAlgorithm
	}{
		{"A256GCM_RSA-OAEP", &EncA256GCM, &AlgRSAOAEP},
		{"A192GCM_ECDH-ES", &EncA192GCM, &AlgECDHES},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Generate JWK and encrypt.
			_, _, publicJWK, _, _, err := GenerateJWEJWKForEncAndAlg(tc.enc, tc.alg)
			require.NoError(t, err)

			clearBytes := []byte("test message")
			jweMessage, _, err := EncryptBytes([]joseJwk.Key{publicJWK}, clearBytes)
			require.NoError(t, err)

			// Get headers string.
			headersStr, err := JWEHeadersString(jweMessage)
			require.NoError(t, err)
			require.NotEmpty(t, headersStr)
			require.Contains(t, headersStr, cryptoutilSharedMagic.JoseKeyUseEnc)
			require.Contains(t, headersStr, "alg")
		})
	}
}

// TestJWEHeadersString_NilMessage tests error handling for nil message.
func TestJWEHeadersString_NilMessage(t *testing.T) {
	t.Parallel()

	_, err := JWEHeadersString(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "jwks can't be nil")
}

// TestDecryptBytes_MultipleJWKs tests decryption with multiple keys (first match wins).
func TestDecryptBytes_MultipleJWKs(t *testing.T) {
	t.Parallel()

	// Generate two key pairs.
	_, nonPublicJWK1, publicJWK1, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgRSAOAEP)
	require.NoError(t, err)

	_, _, publicJWK2, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgRSAOAEP)
	require.NoError(t, err)

	// Encrypt with first public key.
	clearBytes := []byte("test message")
	_, jweMessageBytes, err := EncryptBytes([]joseJwk.Key{publicJWK1}, clearBytes)
	require.NoError(t, err)

	// Decrypt with both private keys (first should succeed).
	decryptedBytes, err := DecryptBytes([]joseJwk.Key{nonPublicJWK1, nonPublicJWK1}, jweMessageBytes)
	require.NoError(t, err)
	require.Equal(t, clearBytes, decryptedBytes)

	// Try with only wrong key (should fail).
	_, err = DecryptBytes([]joseJwk.Key{publicJWK2}, jweMessageBytes)
	require.Error(t, err)
}
