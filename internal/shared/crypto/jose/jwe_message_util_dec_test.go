// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"testing"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

func TestDecryptBytesWithContext_NilJWKs(t *testing.T) {
	t.Parallel()

	jweMessageBytes := []byte("dummy")
	_, err := DecryptBytesWithContext(nil, jweMessageBytes, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func TestDecryptBytesWithContext_EmptyJWKs(t *testing.T) {
	t.Parallel()

	jwks := []joseJwk.Key{}
	jweMessageBytes := []byte("dummy")
	_, err := DecryptBytesWithContext(jwks, jweMessageBytes, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeEmpty)
}

func TestDecryptBytesWithContext_NilMessageBytes(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK}

	_, err = DecryptBytesWithContext(jwks, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid jweMessageBytes")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func TestDecryptBytesWithContext_EmptyMessageBytes(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK}

	jweMessageBytes := []byte{}
	_, err = DecryptBytesWithContext(jwks, jweMessageBytes, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid jweMessageBytes")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeEmpty)
}

func TestDecryptBytesWithContext_NonDecryptJWK(t *testing.T) {
	t.Parallel()

	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgRS256)
	require.NoError(t, err)

	jwks := []joseJwk.Key{signingJWK}

	jweMessageBytes := []byte("dummy")
	_, err = DecryptBytesWithContext(jwks, jweMessageBytes, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWK")
}

func TestDecryptBytesWithContext_InvalidMessageBytes(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK}

	jweMessageBytes := []byte("invalid-jwe-message")
	_, err = DecryptBytesWithContext(jwks, jweMessageBytes, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse JWE message bytes")
}

func Test_JWEHeadersString_NilMessage(t *testing.T) {
	t.Parallel()

	// Test nil JWE message should return error.
	headers, err := JWEHeadersString(nil)
	require.Error(t, err)
	require.Empty(t, headers)
	require.Contains(t, err.Error(), "invalid jweMessage")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func Test_ExtractKidFromJWEMessage_NilMessage(t *testing.T) {
	t.Parallel()

	// Test nil JWE message should return error.
	kid, err := ExtractKidFromJWEMessage(nil)
	require.Error(t, err)
	require.Nil(t, kid)
	require.Contains(t, err.Error(), "invalid jweMessage")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func Test_ExtractKidFromJWEMessage_InvalidUUID(t *testing.T) {
	t.Parallel()

	// Generate JWK and encrypt data to create valid JWE message.
	jweJWKs, _, err := GenerateJWEJWKsForTest(t, 1, &EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	plaintext := []byte("test data")
	jweMessage, _, err := EncryptBytes(jweJWKs, plaintext)
	require.NoError(t, err)

	// Manually set kid to invalid UUID format.
	err = jweMessage.ProtectedHeaders().Set(joseJwk.KeyIDKey, "not-a-valid-uuid")
	require.NoError(t, err)

	// Test extraction should fail with UUID parse error.
	kid, err := ExtractKidFromJWEMessage(jweMessage)
	require.Error(t, err)
	require.Nil(t, kid)
	require.Contains(t, err.Error(), "failed to parse kid UUID")
}

func Test_ExtractKidEncAlgFromJWEMessage_NilMessage(t *testing.T) {
	t.Parallel()

	// Test nil JWE message should return error.
	kid, enc, alg, err := ExtractKidEncAlgFromJWEMessage(nil)
	require.Error(t, err)
	require.Nil(t, kid)
	require.Nil(t, enc)
	require.Nil(t, alg)
	require.Contains(t, err.Error(), "failed to get kid UUID")
}

func Test_ExtractKidEncAlgFromJWEMessage_MissingEnc(t *testing.T) {
	t.Parallel()

	// Generate JWK and encrypt data to create valid JWE message.
	jweJWKs, _, err := GenerateJWEJWKsForTest(t, 1, &EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	plaintext := []byte("test data")
	jweMessage, _, err := EncryptBytes(jweJWKs, plaintext)
	require.NoError(t, err)

	// Remove enc header.
	err = jweMessage.ProtectedHeaders().Remove("enc")
	require.NoError(t, err)

	// Test extraction should fail with missing enc error.
	kid, enc, alg, err := ExtractKidEncAlgFromJWEMessage(jweMessage)
	require.Error(t, err)
	require.Nil(t, kid)
	require.Nil(t, enc)
	require.Nil(t, alg)
	require.Contains(t, err.Error(), "failed to get enc")
}

func Test_ExtractKidEncAlgFromJWEMessage_MissingAlg(t *testing.T) {
	t.Parallel()

	// Generate JWK and encrypt data to create valid JWE message.
	jweJWKs, _, err := GenerateJWEJWKsForTest(t, 1, &EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	plaintext := []byte("test data")
	jweMessage, _, err := EncryptBytes(jweJWKs, plaintext)
	require.NoError(t, err)

	// Remove alg header.
	err = jweMessage.ProtectedHeaders().Remove(joseJwk.AlgorithmKey)
	require.NoError(t, err)

	// Test extraction should fail with missing alg error.
	kid, enc, alg, err := ExtractKidEncAlgFromJWEMessage(jweMessage)
	require.Error(t, err)
	require.Nil(t, kid)
	require.Nil(t, enc)
	require.Nil(t, alg)
	require.Contains(t, err.Error(), "failed to get alg")
}

func Test_EncryptKey_HappyPath(t *testing.T) {
	t.Parallel()

	// Generate KEK for key encryption and CEK to encrypt.
	enc := joseJwa.A256GCM()
	alg := joseJwa.A256KW()
	_, keks, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &alg)
	require.NoError(t, err)

	dirAlg := joseJwa.DIRECT()
	_, cek, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &dirAlg)
	require.NoError(t, err)

	// Encrypt CEK.
	_, encryptedCEKBytes, err := EncryptKey([]joseJwk.Key{keks}, cek)
	require.NoError(t, err)
	require.NotEmpty(t, encryptedCEKBytes)
}

func Test_DecryptKey_InvalidEncryptedBytes(t *testing.T) {
	t.Parallel()

	// Generate KDK for key decryption.
	enc := joseJwa.A256GCM()
	alg := joseJwa.A256KW()
	_, kdks, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &alg)
	require.NoError(t, err)

	// Try to decrypt invalid bytes.
	invalidBytes := []byte("not-valid-jwe-message")

	_, err = DecryptKey([]joseJwk.Key{kdks}, invalidBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt CDK bytes")
}

// Test_DecryptKey_InvalidKeyBytes tests error for malformed decrypted key bytes.
func Test_DecryptKey_InvalidKeyBytes(t *testing.T) {
	t.Parallel()

	// Generate KEK and encrypt invalid CEK bytes.
	enc := joseJwa.A256GCM()
	alg := joseJwa.A256KW()
	_, keks, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &alg)
	require.NoError(t, err)

	// Encrypt malformed key bytes (not valid JWK).
	invalidCEKBytes := []byte(`{"invalid":"jwk"}`)
	_, encryptedBytes, err := EncryptBytes([]joseJwk.Key{keks}, invalidCEKBytes)
	require.NoError(t, err)

	// Try to decrypt - should fail on ParseKey.
	_, err = DecryptKey([]joseJwk.Key{keks}, encryptedBytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to derypt CDK")
}
