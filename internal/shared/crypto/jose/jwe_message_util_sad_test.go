// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"testing"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

func Test_SadPath_EncryptBytes_NilKey(t *testing.T) {
	t.Parallel()

	_, _, err := EncryptBytes(nil, []byte("cleartext"))
	require.Error(t, err)
}

func Test_EncryptBytesWithContext_NilJWKS(t *testing.T) {
	t.Parallel()

	_, _, err := EncryptBytesWithContext(nil, []byte("cleartext"), []byte("context"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func Test_EncryptBytesWithContext_EmptyJWKS(t *testing.T) {
	t.Parallel()

	_, _, err := EncryptBytesWithContext([]joseJwk.Key{}, []byte("cleartext"), []byte("context"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeEmpty)
}

func Test_EncryptBytesWithContext_NilClearBytes(t *testing.T) {
	t.Parallel()

	_, encryptJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	_, _, err = EncryptBytesWithContext([]joseJwk.Key{encryptJWK}, nil, []byte("context"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid clearBytes")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func Test_EncryptBytesWithContext_EmptyClearBytes(t *testing.T) {
	t.Parallel()

	_, encryptJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	_, _, err = EncryptBytesWithContext([]joseJwk.Key{encryptJWK}, []byte{}, []byte("context"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid clearBytes")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeEmpty)
}

func Test_EncryptBytesWithContext_NonEncryptJWK(t *testing.T) {
	t.Parallel()

	// Generate JWS signing JWK (not encrypt JWK).
	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgRS256)
	require.NoError(t, err)

	// Should error on non-encrypt JWK validation.
	_, _, err = EncryptBytesWithContext([]joseJwk.Key{signingJWK}, []byte("cleartext"), []byte("context"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWK")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrJWKMustBeEncryptJWK)
}

// Test_EncryptBytesWithContext_MultipleEnc tests error for multiple encryption algorithms.
func Test_EncryptBytesWithContext_MultipleEnc(t *testing.T) {
	t.Parallel()

	// Generate two JWKs with different enc values: A128GCM and A256GCM.
	_, jwkA128, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA128GCM, &AlgDir)
	require.NoError(t, err)

	_, jwkA256, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgDir)
	require.NoError(t, err)

	// Encrypt with multiple enc algorithms should error.
	_, _, err = EncryptBytesWithContext([]joseJwk.Key{jwkA128, jwkA256}, []byte("cleartext"), []byte("context"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "only one unique 'enc' attribute is allowed")
}

func Test_SadPath_DecryptBytes_NilKey(t *testing.T) {
	t.Parallel()

	_, err := DecryptBytes(nil, []byte("cleartext"))
	require.Error(t, err)
}

func Test_DecryptBytesWithContext_NilJWKS(t *testing.T) {
	t.Parallel()

	_, err := DecryptBytesWithContext(nil, []byte("cipher"), []byte("context"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func Test_DecryptBytesWithContext_EmptyJWKS(t *testing.T) {
	t.Parallel()

	_, err := DecryptBytesWithContext([]joseJwk.Key{}, []byte("cipher"), []byte("context"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeEmpty)
}

func Test_DecryptBytesWithContext_NilJWEMessageBytes(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	_, err = DecryptBytesWithContext([]joseJwk.Key{nonPublicJWK}, nil, []byte("context"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid jweMessageBytes")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func Test_DecryptBytesWithContext_EmptyJWEMessageBytes(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	_, err = DecryptBytesWithContext([]joseJwk.Key{nonPublicJWK}, []byte{}, []byte("context"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid jweMessageBytes")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeEmpty)
}

func Test_SadPath_DecryptBytes_InvalidJWEMessage(t *testing.T) {
	t.Parallel()

	kid, nonPublicJWEJWK, _, clearNonPublicJWEJWKBytes, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, kid)
	require.NotNil(t, nonPublicJWEJWK)
	isEncryptJWK, err := IsEncryptJWK(nonPublicJWEJWK)
	require.NoError(t, err)
	require.True(t, isEncryptJWK)
	require.NotNil(t, clearNonPublicJWEJWKBytes)

	_, err = DecryptBytes([]joseJwk.Key{nonPublicJWEJWK}, []byte("this-is-not-a-valid-jwe-message"))
	require.Error(t, err)
}

func Test_SadPath_GenerateJWEJWK_UnsupportedEnc(t *testing.T) {
	t.Parallel()

	kid, nonPublicJWEJWK, publicJWEJWK, clearNonPublicJWEJWKBytes, clearPublicJWEJWKBytes, err := GenerateJWEJWKForEncAndAlg(&EncInvalid, &AlgA256KW)
	require.Error(t, err)
	require.Equal(t, "invalid JWE JWK headers: JWE JWK length error: unsupported JWE JWK enc invalid", err.Error())
	require.Nil(t, kid)
	require.Nil(t, nonPublicJWEJWK)
	require.Nil(t, publicJWEJWK)
	require.Nil(t, clearNonPublicJWEJWKBytes)
	require.Nil(t, clearPublicJWEJWKBytes)
}

func Test_SadPath_GenerateJWEJWK_UnsupportedAlg(t *testing.T) {
	t.Parallel()

	kid, nonPublicJWEJWK, publicJWEJWK, clearNonPublicJWEJWKBytes, clearPublicJWEJWKBytes, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgEncInvalid)
	require.Error(t, err)
	require.Equal(t, "invalid JWE JWK headers: unsupported JWE JWK alg invalid", err.Error())
	require.Nil(t, kid)
	require.Nil(t, nonPublicJWEJWK)
	require.Nil(t, publicJWEJWK)
	require.Nil(t, clearNonPublicJWEJWKBytes)
	require.Nil(t, clearPublicJWEJWKBytes)
}

func Test_SadPath_ConcurrentGenerateJWEJWK_UnsupportedEnc(t *testing.T) {
	t.Parallel()
	nonPublicJWEJWKs, publicJWEJWKs, err := GenerateJWEJWKsForTest(t, 2, &EncInvalid, &AlgA256KW)
	require.Error(t, err)
	require.Equal(t, "unexpected 2 errors: invalid JWE JWK headers: JWE JWK length error: unsupported JWE JWK enc invalid\ninvalid JWE JWK headers: JWE JWK length error: unsupported JWE JWK enc invalid", err.Error())
	require.Nil(t, nonPublicJWEJWKs)
	require.Nil(t, publicJWEJWKs)
}

func Test_SadPath_ConcurrentGenerateJWEJWK_UnsupportedAlg(t *testing.T) {
	t.Parallel()
	nonPublicJWEJWKs, publicJWEJWKs, err := GenerateJWEJWKsForTest(t, 2, &EncA256GCM, &AlgEncInvalid)
	require.Error(t, err)
	require.Equal(t, "unexpected 2 errors: invalid JWE JWK headers: unsupported JWE JWK alg invalid\ninvalid JWE JWK headers: unsupported JWE JWK alg invalid", err.Error())
	require.Nil(t, nonPublicJWEJWKs)
	require.Nil(t, publicJWEJWKs)
}

func Test_ExtractKidFromJWEMessage_HappyPath(t *testing.T) {
	t.Parallel()

	// Generate JWK for encryption.
	jweJWKs, _, err := GenerateJWEJWKsForTest(t, 1, &EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	// Encrypt test data.
	plaintext := []byte("test data")
	jweMessage, _, err := EncryptBytes(jweJWKs, plaintext)
	require.NoError(t, err)

	// Test extraction.
	kid, err := ExtractKidFromJWEMessage(jweMessage)
	require.NoError(t, err)
	require.NotNil(t, kid)

	// Verify KID matches JWK.
	expectedKid, err := ExtractKidUUID(jweJWKs[0])
	require.NoError(t, err)
	require.Equal(t, expectedKid, kid)
}

func Test_ExtractKidEncAlgFromJWEMessage_HappyPath(t *testing.T) {
	t.Parallel()

	// Generate JWK for encryption.
	jweJWKs, _, err := GenerateJWEJWKsForTest(t, 1, &EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	// Encrypt test data.
	plaintext := []byte("test data")
	jweMessage, _, err := EncryptBytes(jweJWKs, plaintext)
	require.NoError(t, err)

	// Test extraction.
	kid, enc, alg, err := ExtractKidEncAlgFromJWEMessage(jweMessage)
	require.NoError(t, err)
	require.NotNil(t, kid)
	require.NotNil(t, enc)
	require.NotNil(t, alg)

	// Verify values.
	expectedKid, err := ExtractKidUUID(jweJWKs[0])
	require.NoError(t, err)
	require.Equal(t, expectedKid, kid)
	require.Equal(t, EncA256GCM, *enc)
	require.Equal(t, AlgA256KW, *alg)
}

func Test_ExtractKidFromJWEMessage_InvalidMessage(t *testing.T) {
	t.Parallel()

	// Create JWE message without proper headers will fail during extraction.
	// Parse a minimal JWE message (this will be invalid but not nil).
	jweCompact := "eyJhbGciOiJkaXIiLCJlbmMiOiJBMTI4R0NNIn0..invalid.invalid.invalid"

	jweMessage, err := joseJwe.Parse([]byte(jweCompact))
	if err == nil {
		// If parse succeeds, extraction should still fail due to missing KID.
		kid, extractErr := ExtractKidFromJWEMessage(jweMessage)
		require.Error(t, extractErr)
		require.Nil(t, kid)
		require.Contains(t, extractErr.Error(), "failed to get kid UUID")
	} else {
		// If parse fails, that's also acceptable for this invalid message test.
		require.Error(t, err)
	}
}
