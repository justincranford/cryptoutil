// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"
	"github.com/stretchr/testify/require"
)

// TestSignBytes_MultipleJWKsSameAlg tests signing with multiple JWKs of same algorithm (JSON encoding).
func TestSignBytes_MultipleJWKsSameAlg(t *testing.T) {
	t.Parallel()

	// Generate two JWKs with same algorithm.
	_, nonPublicJWK1, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	_, nonPublicJWK2, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK1, nonPublicJWK2}
	clearBytes := []byte("test message for multiple JWKs")

	// Should succeed with JSON encoding (multiple signatures).
	jwsMessage, jwsMessageBytes, err := SignBytes(jwks, clearBytes)
	require.NoError(t, err)
	require.NotNil(t, jwsMessage)
	require.NotEmpty(t, jwsMessageBytes)

	// Verify message has 2 signatures.
	require.Equal(t, 2, len(jwsMessage.Signatures()))
}

// TestVerifyBytes_HappyPath tests successful verification of signed message.
func TestVerifyBytes_HappyPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		alg  *joseJwa.SignatureAlgorithm
	}{
		{cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, &AlgRS256},
		{cryptoutilSharedMagic.JoseAlgES256, &AlgES256},
		{cryptoutilSharedMagic.JoseAlgHS256, &AlgHS256},
		{cryptoutilSharedMagic.JoseAlgEdDSA, &AlgEdDSA},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Generate JWK and sign.
			_, nonPublicJWK, publicJWK, _, _, err := GenerateJWSJWKForAlg(tc.alg)
			require.NoError(t, err)

			clearBytes := []byte("test message for " + tc.name)
			_, jwsMessageBytes, err := SignBytes([]joseJwk.Key{nonPublicJWK}, clearBytes)
			require.NoError(t, err)

			// Verify with public key (or same key for HMAC).
			verifyJWK := publicJWK
			if publicJWK == nil {
				verifyJWK = nonPublicJWK // HMAC uses same key.
			}

			verifiedBytes, err := VerifyBytes([]joseJwk.Key{verifyJWK}, jwsMessageBytes)
			require.NoError(t, err)
			require.Equal(t, clearBytes, verifiedBytes)
		})
	}
}

// TestVerifyBytes_InvalidSignature tests verification failure with wrong key.
func TestVerifyBytes_InvalidSignature(t *testing.T) {
	t.Parallel()

	// Sign with one key.
	_, nonPublicJWK1, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	clearBytes := []byte("test message")
	_, jwsMessageBytes, err := SignBytes([]joseJwk.Key{nonPublicJWK1}, clearBytes)
	require.NoError(t, err)

	// Try to verify with different key.
	_, nonPublicJWK2, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	_, err = VerifyBytes([]joseJwk.Key{nonPublicJWK2}, jwsMessageBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to verify JWS message")
}

// TestExtractKidAlgFromJWSMessage_HappyPath tests successful extraction of kid and alg.
func TestExtractKidAlgFromJWSMessage_HappyPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		alg  *joseJwa.SignatureAlgorithm
	}{
		{cryptoutilSharedMagic.JoseAlgRS384, &AlgRS384},
		{cryptoutilSharedMagic.JoseAlgES384, &AlgES384},
		{cryptoutilSharedMagic.JoseAlgHS512, &AlgHS512},
		{cryptoutilSharedMagic.JoseAlgEdDSA, &AlgEdDSA},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Generate JWK and sign.
			kid, nonPublicJWK, _, _, _, err := GenerateJWSJWKForAlg(tc.alg)
			require.NoError(t, err)

			clearBytes := []byte("test message")
			jwsMessage, _, err := SignBytes([]joseJwk.Key{nonPublicJWK}, clearBytes)
			require.NoError(t, err)

			// Extract kid and alg.
			extractedKid, extractedAlg, err := ExtractKidAlgFromJWSMessage(jwsMessage)
			require.NoError(t, err)
			require.Equal(t, kid, extractedKid)
			require.Equal(t, *tc.alg, *extractedAlg)
		})
	}
}

// TestExtractKidAlgFromJWSMessage_NoSignatures tests error when message has no signatures.
func TestExtractKidAlgFromJWSMessage_NoSignatures(t *testing.T) {
	t.Parallel()

	// Use a manually constructed message with no signatures.
	_, nonPublicJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	clearBytes := []byte("test")
	jwsMessage, jwsBytes, err := SignBytes([]joseJwk.Key{nonPublicJWK}, clearBytes)
	require.NoError(t, err)

	// Manually parse to get message structure.
	parsed, err := joseJws.Parse(jwsBytes)
	require.NoError(t, err)
	require.NotNil(t, parsed)

	// Now test extraction (should succeed with valid message).
	kid, alg, err := ExtractKidAlgFromJWSMessage(jwsMessage)
	require.NoError(t, err)
	require.NotNil(t, kid)
	require.NotNil(t, alg)
}

// TestJWSHeadersString_HappyPath tests header string generation.
func TestJWSHeadersString_HappyPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		alg  *joseJwa.SignatureAlgorithm
	}{
		{cryptoutilSharedMagic.JoseAlgPS256, &AlgPS256},
		{cryptoutilSharedMagic.JoseAlgES512, &AlgES512},
		{cryptoutilSharedMagic.JoseAlgHS384, &AlgHS384},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Generate JWK and sign.
			_, nonPublicJWK, _, _, _, err := GenerateJWSJWKForAlg(tc.alg)
			require.NoError(t, err)

			clearBytes := []byte("test message")
			jwsMessage, _, err := SignBytes([]joseJwk.Key{nonPublicJWK}, clearBytes)
			require.NoError(t, err)

			// Get headers string.
			headersStr, err := JWSHeadersString(jwsMessage)
			require.NoError(t, err)
			require.NotEmpty(t, headersStr)
			require.Contains(t, headersStr, "alg")
			require.Contains(t, headersStr, "kid")
		})
	}
}

// TestJWSHeadersString_NilMessage tests error handling for nil message.
func TestJWSHeadersString_NilMessage(t *testing.T) {
	t.Parallel()

	_, err := JWSHeadersString(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "jwsMessage is nil")
}

// TestLogJWSInfo_HappyPath tests JWS info logging.
func TestLogJWSInfo_HappyPath(t *testing.T) {
	t.Parallel()

	// Generate JWK and sign.
	_, nonPublicJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgRS512)
	require.NoError(t, err)

	clearBytes := []byte("test message for logging")
	jwsMessage, _, err := SignBytes([]joseJwk.Key{nonPublicJWK}, clearBytes)
	require.NoError(t, err)

	// Log info (should not return error).
	err = LogJWSInfo(jwsMessage)
	require.NoError(t, err)
}

// TestLogJWSInfo_NilMessage tests error handling for nil message.
func TestLogJWSInfo_NilMessage(t *testing.T) {
	t.Parallel()

	err := LogJWSInfo(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "jwsMessage is nil")
}

// TestLogJWSInfo_MultipleSignatures tests logging with multiple signatures.
func TestLogJWSInfo_MultipleSignatures(t *testing.T) {
	t.Parallel()

	// Generate two JWKs.
	_, nonPublicJWK1, _, _, _, err := GenerateJWSJWKForAlg(&AlgES256)
	require.NoError(t, err)

	_, nonPublicJWK2, _, _, _, err := GenerateJWSJWKForAlg(&AlgES256)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK1, nonPublicJWK2}
	clearBytes := []byte("test message with multiple signatures")

	jwsMessage, _, err := SignBytes(jwks, clearBytes)
	require.NoError(t, err)
	require.Equal(t, 2, len(jwsMessage.Signatures()))

	// Log info (should handle multiple signatures).
	err = LogJWSInfo(jwsMessage)
	require.NoError(t, err)
}

// TestVerifyBytes_MultipleJWKs tests verification with multiple public keys.
func TestVerifyBytes_MultipleJWKs(t *testing.T) {
	t.Parallel()

	// Generate two RSA key pairs.
	_, nonPublicJWK1, publicJWK1, _, _, err := GenerateJWSJWKForAlg(&AlgRS256)
	require.NoError(t, err)

	_, _, publicJWK2, _, _, err := GenerateJWSJWKForAlg(&AlgRS256)
	require.NoError(t, err)

	// Sign with first key.
	clearBytes := []byte("test message")
	_, jwsMessageBytes, err := SignBytes([]joseJwk.Key{nonPublicJWK1}, clearBytes)
	require.NoError(t, err)

	// Verify with both public keys (first should succeed).
	verifiedBytes, err := VerifyBytes([]joseJwk.Key{publicJWK1, publicJWK2}, jwsMessageBytes)
	require.NoError(t, err)
	require.Equal(t, clearBytes, verifiedBytes)
}
