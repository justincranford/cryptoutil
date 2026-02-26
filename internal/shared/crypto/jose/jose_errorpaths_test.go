// Copyright (c) 2025 Justin Cranford

// Package crypto — additional coverage tests for jose message utilities.
// Covers: wrong JWK type for sign/encrypt/decrypt, ExtractKidAlg malformed headers, mixed enc/alg.
package crypto

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// SignBytes with non-signing JWK (covers !isSignJWK path).
// =============================================================================

// TestSignBytes_EncryptionJWK passes an encrypt-only JWK to trigger the !isSignJWK error.
func TestSignBytes_EncryptionJWK(t *testing.T) {
	t.Parallel()

	// Generate a JWE JWK (encryption-only, no signing alg header).
	// Use 2nd return (nonPublicJWK) — 3rd return (publicJWK) is nil for symmetric.
	encAlg := joseJwa.A256GCMKW()
	enc := joseJwa.A256GCM()
	_, encJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &encAlg)
	require.NoError(t, err)

	// Attempt to sign with the encryption JWK.
	_, _, err = SignBytes([]joseJwk.Key{encJWK}, []byte("test payload"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWK must be a sign JWK")
}

// =============================================================================
// EncryptBytesWithContext with non-encryption JWK (covers !isEncryptJWK path).
// =============================================================================

// TestEncryptBytes_SigningJWK passes a signing JWK to trigger the !isEncryptJWK error.
func TestEncryptBytes_SigningJWK(t *testing.T) {
	t.Parallel()

	// Generate a JWS JWK (signing-only, no enc header).
	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgRS256)
	require.NoError(t, err)

	// Attempt to encrypt with the signing JWK.
	_, _, err = EncryptBytes([]joseJwk.Key{signingJWK}, []byte("test payload"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWK must be an encrypt JWK")
}

// =============================================================================
// DecryptBytesWithContext with non-decryption JWK (covers !isDecryptJWK path).
// =============================================================================

// TestDecryptBytes_SigningJWK passes a signing JWK to trigger the !isDecryptJWK error.
func TestDecryptBytes_SigningJWK(t *testing.T) {
	t.Parallel()

	// Generate a JWS JWK (signing-only, no enc header).
	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgRS256)
	require.NoError(t, err)

	// Attempt to decrypt with the signing JWK.
	_, err = DecryptBytes([]joseJwk.Key{signingJWK}, []byte("fake-jwe-bytes"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWK must be a decrypt JWK")
}

// =============================================================================
// DecryptBytesWithContext with mixed enc/alg JWKs (covers mixed enc/alg paths).
// =============================================================================

// TestDecryptBytes_MixedEncAlgorithms covers the mixed enc algorithm error path.
func TestDecryptBytes_MixedEncAlgorithms(t *testing.T) {
	t.Parallel()

	// Generate two JWE JWKs with different enc algorithms.
	enc1 := joseJwa.A256GCM()
	alg1 := joseJwa.A256GCMKW()
	_, nonPublicJWK1, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc1, &alg1)
	require.NoError(t, err)

	enc2 := joseJwa.A128GCM()
	alg2 := joseJwa.A128GCMKW()
	_, nonPublicJWK2, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc2, &alg2)
	require.NoError(t, err)

	// Encrypt with JWK1, then try to decrypt with both JWKs (mixed enc).
	_, jweBytes, err := EncryptBytes([]joseJwk.Key{nonPublicJWK1}, []byte("test payload"))
	require.NoError(t, err)

	_, err = DecryptBytes([]joseJwk.Key{nonPublicJWK1, nonPublicJWK2}, jweBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only one unique 'enc' attribute is allowed")
}

// =============================================================================
// ExtractKidAlgFromJWSMessage with malformed JWS (covers missing kid/alg paths).
// =============================================================================

// TestExtractKidAlgFromJWSMessage_MissingKid covers the missing kid header error.
func TestExtractKidAlgFromJWSMessage_MissingKid(t *testing.T) {
	t.Parallel()

	// Sign a valid JWS, then remove kid from parsed protected headers.
	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	_, jwsBytes, err := SignBytes([]joseJwk.Key{signingJWK}, []byte("test"))
	require.NoError(t, err)

	jwsMessage, err := joseJws.Parse(jwsBytes)
	require.NoError(t, err)

	// Remove kid from the parsed message's protected headers.
	sig := jwsMessage.Signatures()[0]
	require.NoError(t, sig.ProtectedHeaders().Remove(joseJwk.KeyIDKey))

	_, _, err = ExtractKidAlgFromJWSMessage(jwsMessage)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get kid UUID")
}

// TestExtractKidAlgFromJWSMessage_InvalidKidUUID covers the invalid kid UUID error.
func TestExtractKidAlgFromJWSMessage_InvalidKidUUID(t *testing.T) {
	t.Parallel()

	// Sign a valid JWS, then overwrite kid with a non-UUID value.
	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	_, jwsBytes, err := SignBytes([]joseJwk.Key{signingJWK}, []byte("test"))
	require.NoError(t, err)

	jwsMessage, err := joseJws.Parse(jwsBytes)
	require.NoError(t, err)

	// Overwrite kid with a non-UUID string.
	sig := jwsMessage.Signatures()[0]
	require.NoError(t, sig.ProtectedHeaders().Set(joseJwk.KeyIDKey, "not-a-valid-uuid"))

	_, _, err = ExtractKidAlgFromJWSMessage(jwsMessage)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse kid UUID")
}

// TestExtractKidAlgFromJWSMessage_MissingAlg covers the missing alg header error.
func TestExtractKidAlgFromJWSMessage_MissingAlg(t *testing.T) {
	t.Parallel()

	// Sign a valid JWS, then remove alg from parsed protected headers.
	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	_, jwsBytes, err := SignBytes([]joseJwk.Key{signingJWK}, []byte("test"))
	require.NoError(t, err)

	jwsMessage, err := joseJws.Parse(jwsBytes)
	require.NoError(t, err)

	// Remove alg from the parsed message's protected headers.
	sig := jwsMessage.Signatures()[0]
	require.NoError(t, sig.ProtectedHeaders().Remove(joseJwk.AlgorithmKey))

	_, _, err = ExtractKidAlgFromJWSMessage(jwsMessage)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get alg")
}

// =============================================================================
// DecryptBytes with mixed alg (same enc) triggers the mixed alg error path.
// =============================================================================

// TestDecryptBytes_MixedAlgAlgorithms covers the mixed key encryption alg error path.
func TestDecryptBytes_MixedAlgAlgorithms(t *testing.T) {
	t.Parallel()

	// Generate two JWE JWKs with same enc but different alg.
	enc := joseJwa.A256GCM()
	alg1 := joseJwa.A256GCMKW()
	_, nonPublicJWK1, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &alg1)
	require.NoError(t, err)

	alg2 := joseJwa.A128GCMKW()
	_, nonPublicJWK2, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &alg2)
	require.NoError(t, err)

	// Encrypt with JWK1, then try to decrypt with both JWKs (mixed alg).
	_, jweBytes, err := EncryptBytes([]joseJwk.Key{nonPublicJWK1}, []byte("test payload"))
	require.NoError(t, err)

	_, err = DecryptBytes([]joseJwk.Key{nonPublicJWK1, nonPublicJWK2}, jweBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only one unique 'alg' attribute is allowed")
}

// =============================================================================
// EncryptBytes with mixed enc/alg triggers the corresponding error paths.
// =============================================================================

// TestEncryptBytes_MixedEncAlgorithms covers the mixed enc error in EncryptBytesWithContext.
func TestEncryptBytes_MixedEncAlgorithms(t *testing.T) {
	t.Parallel()

	enc1 := joseJwa.A256GCM()
	alg1 := joseJwa.A256GCMKW()
	_, nonPublicJWK1, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc1, &alg1)
	require.NoError(t, err)

	enc2 := joseJwa.A128GCM()
	alg2 := joseJwa.A128GCMKW()
	_, nonPublicJWK2, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc2, &alg2)
	require.NoError(t, err)

	_, _, err = EncryptBytes([]joseJwk.Key{nonPublicJWK1, nonPublicJWK2}, []byte("test payload"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "only one unique 'enc' attribute is allowed")
}

// TestEncryptBytes_MixedAlgAlgorithms covers the mixed alg error in EncryptBytesWithContext.
func TestEncryptBytes_MixedAlgAlgorithms(t *testing.T) {
	t.Parallel()

	enc := joseJwa.A256GCM()
	alg1 := joseJwa.A256GCMKW()
	_, nonPublicJWK1, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &alg1)
	require.NoError(t, err)

	alg2 := joseJwa.A128GCMKW()
	_, nonPublicJWK2, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &alg2)
	require.NoError(t, err)

	_, _, err = EncryptBytes([]joseJwk.Key{nonPublicJWK1, nonPublicJWK2}, []byte("test payload"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "only one unique 'alg' attribute is allowed")
}

// =============================================================================
// VerifyBytes with mixed alg triggers the mixed alg error path.
// =============================================================================

// TestVerifyBytes_MixedAlgAlgorithms covers the mixed signature alg error path.
func TestVerifyBytes_MixedAlgAlgorithms(t *testing.T) {
	t.Parallel()

	// Generate two JWS JWKs with different alg values.
	_, sigJWK1, publicJWK1, _, _, err := GenerateJWSJWKForAlg(&AlgRS256)
	require.NoError(t, err)
	require.NotNil(t, publicJWK1)

	_, _, publicJWK2, _, _, err := GenerateJWSJWKForAlg(&AlgRS384)
	require.NoError(t, err)
	require.NotNil(t, publicJWK2)

	// Sign with JWK1 to get valid JWS bytes.
	_, jwsBytes, err := SignBytes([]joseJwk.Key{sigJWK1}, []byte("test payload"))
	require.NoError(t, err)

	// Try to verify with both public keys (mixed alg).
	_, err = VerifyBytes([]joseJwk.Key{publicJWK1, publicJWK2}, jwsBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only one unique 'alg' attribute is allowed")
}

// =============================================================================
// LogJWSInfo with exotic headers (covers typ, cty, jku, x5u, x5c, x5t, etc.).
// =============================================================================

// TestLogJWSInfo_ExoticHeaders covers LogJWSInfo branches for optional JWS headers.
func TestLogJWSInfo_ExoticHeaders(t *testing.T) {
	t.Parallel()

	// Build a JWS with exotic headers set to cover optional header branches.
	headers := joseJws.NewHeaders()
	require.NoError(t, headers.Set(joseJwk.AlgorithmKey, joseJwa.HS256()))
	require.NoError(t, headers.Set(joseJwk.KeyIDKey, "019c0000-0000-7000-8000-000000000001"))
	require.NoError(t, headers.Set("typ", "JWT"))
	require.NoError(t, headers.Set("cty", "jwt"))
	require.NoError(t, headers.Set("jku", "https://example.com/.well-known/jwks.json"))
	require.NoError(t, headers.Set("x5u", "https://example.com/cert"))
	require.NoError(t, headers.Set("x5t", "dGVzdA"))      // base64 "test".
	require.NoError(t, headers.Set("x5t#S256", "dGVzdA"))  // base64 "test".

	key, err := GenerateHMACJWK(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	require.NoError(t, err)

	jwsBytes, err := joseJws.Sign(
		[]byte("test payload"),
		joseJws.WithKey(joseJwa.HS256(), key, joseJws.WithProtectedHeaders(headers)),
	)
	require.NoError(t, err)

	jwsMessage, err := joseJws.Parse(jwsBytes)
	require.NoError(t, err)

	// LogJWSInfo should handle all headers without error.
	err = LogJWSInfo(jwsMessage)
	require.NoError(t, err)
}

// =============================================================================
// EncryptBytes/SignBytes with JWK missing kid (covers ExtractKidUUID error).
// =============================================================================

// TestEncryptBytes_MissingKid covers the ExtractKidUUID error path in EncryptBytesWithContext.
func TestEncryptBytes_MissingKid(t *testing.T) {
	t.Parallel()

	enc := joseJwa.A256GCM()
	alg := joseJwa.A256GCMKW()
	_, nonPublicJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &alg)
	require.NoError(t, err)

	// Remove kid but keep enc/alg so IsEncryptJWK still returns true.
	require.NoError(t, nonPublicJWK.Remove(joseJwk.KeyIDKey))

	_, _, err = EncryptBytes([]joseJwk.Key{nonPublicJWK}, []byte("test payload"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid")
}

// TestSignBytes_MissingKid covers the ExtractKidUUID error path in SignBytes.
func TestSignBytes_MissingKid(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)

	// Remove kid but keep signing alg so IsSignJWK still returns true.
	require.NoError(t, nonPublicJWK.Remove(joseJwk.KeyIDKey))

	_, _, err = SignBytes([]joseJwk.Key{nonPublicJWK}, []byte("test payload"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid")
}
