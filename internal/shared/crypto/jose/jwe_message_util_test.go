// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

func TestEncryptBytesWithContext_NilJWKs(t *testing.T) {
	t.Parallel()

	clearBytes := []byte("test message")
	_, _, err := EncryptBytesWithContext(nil, clearBytes, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func TestEncryptBytesWithContext_EmptyJWKs(t *testing.T) {
	t.Parallel()

	jwks := []joseJwk.Key{}
	clearBytes := []byte("test message")
	_, _, err := EncryptBytesWithContext(jwks, clearBytes, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWKs")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeEmpty)
}

func TestEncryptBytesWithContext_NilClearBytes(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK}

	_, _, err = EncryptBytesWithContext(jwks, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid clearBytes")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func TestEncryptBytesWithContext_EmptyClearBytes(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK}

	clearBytes := []byte{}
	_, _, err = EncryptBytesWithContext(jwks, clearBytes, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid clearBytes")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeEmpty)
}

func TestEncryptBytesWithContext_NonEncryptJWK(t *testing.T) {
	t.Parallel()

	_, signingJWK, _, _, _, err := GenerateJWSJWKForAlg(&AlgRS256)
	require.NoError(t, err)

	jwks := []joseJwk.Key{signingJWK}

	clearBytes := []byte("test message")
	_, _, err = EncryptBytesWithContext(jwks, clearBytes, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWK")
}

func TestEncryptBytesWithContext_MultipleEncs(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK1, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	_, nonPublicJWK2, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA192GCM, &AlgA256KW)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK1, nonPublicJWK2}
	clearBytes := []byte("test message")
	_, _, err = EncryptBytesWithContext(jwks, clearBytes, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only one unique 'enc' attribute is allowed")
}

func TestEncryptBytesWithContext_MultipleAlgs(t *testing.T) {
	t.Parallel()

	_, nonPublicJWK1, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)

	_, nonPublicJWK2, _, _, _, err := GenerateJWEJWKForEncAndAlg(&EncA256GCM, &AlgA192KW)
	require.NoError(t, err)

	jwks := []joseJwk.Key{nonPublicJWK1, nonPublicJWK2}
	clearBytes := []byte("test message")
	_, _, err = EncryptBytesWithContext(jwks, clearBytes, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only one unique 'alg' attribute is allowed")
}

type happyPathJWETestCase struct {
	enc          *joseJwa.ContentEncryptionAlgorithm
	alg          *joseJwa.KeyEncryptionAlgorithm
	expectedType joseJwa.KeyType
}

var happyPathJWETestCases = []happyPathJWETestCase{
	{enc: &EncA256GCM, alg: &AlgA256KW, expectedType: KtyOCT},
	{enc: &EncA192GCM, alg: &AlgA256KW, expectedType: KtyOCT},
	{enc: &EncA128GCM, alg: &AlgA256KW, expectedType: KtyOCT},
	{enc: &EncA256GCM, alg: &AlgA192KW, expectedType: KtyOCT},
	{enc: &EncA192GCM, alg: &AlgA192KW, expectedType: KtyOCT},
	{enc: &EncA128GCM, alg: &AlgA192KW, expectedType: KtyOCT},
	{enc: &EncA256GCM, alg: &AlgA128KW, expectedType: KtyOCT},
	{enc: &EncA192GCM, alg: &AlgA128KW, expectedType: KtyOCT},
	{enc: &EncA128GCM, alg: &AlgA128KW, expectedType: KtyOCT},

	{enc: &EncA256GCM, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	{enc: &EncA192GCM, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	{enc: &EncA128GCM, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	{enc: &EncA256GCM, alg: &AlgA192GCMKW, expectedType: KtyOCT},
	{enc: &EncA192GCM, alg: &AlgA192GCMKW, expectedType: KtyOCT},
	{enc: &EncA128GCM, alg: &AlgA192GCMKW, expectedType: KtyOCT},
	{enc: &EncA256GCM, alg: &AlgA128GCMKW, expectedType: KtyOCT},
	{enc: &EncA192GCM, alg: &AlgA128GCMKW, expectedType: KtyOCT},
	{enc: &EncA128GCM, alg: &AlgA128GCMKW, expectedType: KtyOCT},

	{enc: &EncA256GCM, alg: &AlgDir, expectedType: KtyOCT},
	{enc: &EncA192GCM, alg: &AlgDir, expectedType: KtyOCT},
	{enc: &EncA128GCM, alg: &AlgDir, expectedType: KtyOCT},

	{enc: &EncA256GCM, alg: &AlgRSAOAEP512, expectedType: KtyRSA},
	{enc: &EncA192GCM, alg: &AlgRSAOAEP512, expectedType: KtyRSA},
	{enc: &EncA128GCM, alg: &AlgRSAOAEP512, expectedType: KtyRSA},
	{enc: &EncA256GCM, alg: &AlgRSAOAEP384, expectedType: KtyRSA},
	{enc: &EncA192GCM, alg: &AlgRSAOAEP384, expectedType: KtyRSA},
	{enc: &EncA128GCM, alg: &AlgRSAOAEP384, expectedType: KtyRSA},
	{enc: &EncA256GCM, alg: &AlgRSAOAEP256, expectedType: KtyRSA},
	{enc: &EncA192GCM, alg: &AlgRSAOAEP256, expectedType: KtyRSA},
	{enc: &EncA128GCM, alg: &AlgRSAOAEP256, expectedType: KtyRSA},
	{enc: &EncA256GCM, alg: &AlgRSAOAEP, expectedType: KtyRSA},
	{enc: &EncA192GCM, alg: &AlgRSAOAEP, expectedType: KtyRSA},
	{enc: &EncA128GCM, alg: &AlgRSAOAEP, expectedType: KtyRSA},
	{enc: &EncA256GCM, alg: &AlgRSA15, expectedType: KtyRSA},
	{enc: &EncA192GCM, alg: &AlgRSA15, expectedType: KtyRSA},
	{enc: &EncA128GCM, alg: &AlgRSA15, expectedType: KtyRSA},

	{enc: &EncA256GCM, alg: &AlgECDHESA256KW, expectedType: KtyEC},
	{enc: &EncA192GCM, alg: &AlgECDHESA256KW, expectedType: KtyEC},
	{enc: &EncA128GCM, alg: &AlgECDHESA256KW, expectedType: KtyEC},
	{enc: &EncA256GCM, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	{enc: &EncA192GCM, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	{enc: &EncA128GCM, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	{enc: &EncA256GCM, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	{enc: &EncA192GCM, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	{enc: &EncA128GCM, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	{enc: &EncA256GCM, alg: &AlgECDHES, expectedType: KtyEC},
	{enc: &EncA192GCM, alg: &AlgECDHES, expectedType: KtyEC},
	{enc: &EncA128GCM, alg: &AlgECDHES, expectedType: KtyEC},

	{enc: &EncA256CBCHS512, alg: &AlgA256KW, expectedType: KtyOCT},
	{enc: &EncA192CBCHS384, alg: &AlgA256KW, expectedType: KtyOCT},
	{enc: &EncA128CBCHS256, alg: &AlgA256KW, expectedType: KtyOCT},
	{enc: &EncA256CBCHS512, alg: &AlgA192KW, expectedType: KtyOCT},
	{enc: &EncA192CBCHS384, alg: &AlgA192KW, expectedType: KtyOCT},
	{enc: &EncA128CBCHS256, alg: &AlgA192KW, expectedType: KtyOCT},
	{enc: &EncA256CBCHS512, alg: &AlgA128KW, expectedType: KtyOCT},
	{enc: &EncA192CBCHS384, alg: &AlgA128KW, expectedType: KtyOCT},
	{enc: &EncA128CBCHS256, alg: &AlgA128KW, expectedType: KtyOCT},

	{enc: &EncA256CBCHS512, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	{enc: &EncA192CBCHS384, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	{enc: &EncA128CBCHS256, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	{enc: &EncA256CBCHS512, alg: &AlgA192GCMKW, expectedType: KtyOCT},
	{enc: &EncA192CBCHS384, alg: &AlgA192GCMKW, expectedType: KtyOCT},
	{enc: &EncA128CBCHS256, alg: &AlgA192GCMKW, expectedType: KtyOCT},
	{enc: &EncA256CBCHS512, alg: &AlgA128GCMKW, expectedType: KtyOCT},
	{enc: &EncA192CBCHS384, alg: &AlgA128GCMKW, expectedType: KtyOCT},
	{enc: &EncA128CBCHS256, alg: &AlgA128GCMKW, expectedType: KtyOCT},

	{enc: &EncA256CBCHS512, alg: &AlgDir, expectedType: KtyOCT},
	{enc: &EncA192CBCHS384, alg: &AlgDir, expectedType: KtyOCT},
	{enc: &EncA128CBCHS256, alg: &AlgDir, expectedType: KtyOCT},

	{enc: &EncA256CBCHS512, alg: &AlgRSAOAEP512, expectedType: KtyRSA},
	{enc: &EncA192CBCHS384, alg: &AlgRSAOAEP512, expectedType: KtyRSA},
	{enc: &EncA128CBCHS256, alg: &AlgRSAOAEP512, expectedType: KtyRSA},
	{enc: &EncA256CBCHS512, alg: &AlgRSAOAEP384, expectedType: KtyRSA},
	{enc: &EncA192CBCHS384, alg: &AlgRSAOAEP384, expectedType: KtyRSA},
	{enc: &EncA128CBCHS256, alg: &AlgRSAOAEP384, expectedType: KtyRSA},
	{enc: &EncA256CBCHS512, alg: &AlgRSAOAEP256, expectedType: KtyRSA},
	{enc: &EncA192CBCHS384, alg: &AlgRSAOAEP256, expectedType: KtyRSA},
	{enc: &EncA128CBCHS256, alg: &AlgRSAOAEP256, expectedType: KtyRSA},
	{enc: &EncA256CBCHS512, alg: &AlgRSAOAEP, expectedType: KtyRSA},
	{enc: &EncA192CBCHS384, alg: &AlgRSAOAEP, expectedType: KtyRSA},
	{enc: &EncA128CBCHS256, alg: &AlgRSAOAEP, expectedType: KtyRSA},
	{enc: &EncA256CBCHS512, alg: &AlgRSA15, expectedType: KtyRSA},
	{enc: &EncA192CBCHS384, alg: &AlgRSA15, expectedType: KtyRSA},
	{enc: &EncA128CBCHS256, alg: &AlgRSA15, expectedType: KtyRSA},

	{enc: &EncA256CBCHS512, alg: &AlgECDHESA256KW, expectedType: KtyEC},
	{enc: &EncA192CBCHS384, alg: &AlgECDHESA256KW, expectedType: KtyEC},
	{enc: &EncA128CBCHS256, alg: &AlgECDHESA256KW, expectedType: KtyEC},
	{enc: &EncA256CBCHS512, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	{enc: &EncA192CBCHS384, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	{enc: &EncA128CBCHS256, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	{enc: &EncA256CBCHS512, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	{enc: &EncA192CBCHS384, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	{enc: &EncA128CBCHS256, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	{enc: &EncA256CBCHS512, alg: &AlgECDHES, expectedType: KtyEC},
	{enc: &EncA192CBCHS384, alg: &AlgECDHES, expectedType: KtyEC},
	{enc: &EncA128CBCHS256, alg: &AlgECDHES, expectedType: KtyEC},
}

func Test_HappyPath_NonJWKGenService_JWE_JWK_EncryptDecryptBytes(t *testing.T) {
	t.Parallel()

	for _, testCase := range happyPathJWETestCases {
		cleartext := fmt.Appendf(nil, "Hello world enc=%s alg=%s!", testCase.enc, testCase.alg)
		t.Run(fmt.Sprintf("%s %s", testCase.enc, testCase.alg), func(t *testing.T) {
			t.Parallel()

			actualKeyKid, nonPublicJWEJWK, publicJWEJWK, clearNonPublicJWEJWKBytes, clearPublicJWEJWKBytes, err := GenerateJWEJWKForEncAndAlg(testCase.enc, testCase.alg)
			require.NoError(t, err)
			require.NotNil(t, nonPublicJWEJWK)
			require.NotEmpty(t, clearNonPublicJWEJWKBytes)
			require.NotEmpty(t, actualKeyKid)
			log.Printf("Generated:\n%s\n%s", clearNonPublicJWEJWKBytes, clearPublicJWEJWKBytes)

			var encryptJWK joseJwk.Key

			requireJWEJWKHeaders(t, nonPublicJWEJWK, OpsEncDec, &testCase)

			if publicJWEJWK == nil {
				encryptJWK = nonPublicJWEJWK
			} else {
				encryptJWK = publicJWEJWK
				requireJWEJWKHeaders(t, publicJWEJWK, OpsEnc, &testCase)
			}

			jweMessage, encodedJWEMessage, err := EncryptBytes([]joseJwk.Key{encryptJWK}, cleartext)
			require.NoError(t, err)
			require.NotEmpty(t, encodedJWEMessage)
			log.Printf("JWE Message: %s", string(encodedJWEMessage))

			requireJWEMessageHeaders(t, jweMessage, actualKeyKid, &testCase)

			decryptedtext, err := DecryptBytes([]joseJwk.Key{nonPublicJWEJWK}, encodedJWEMessage)
			require.NoError(t, err)
			require.Equal(t, cleartext, decryptedtext)
		})
	}
}

func requireJWEJWKHeaders(t *testing.T, nonPublicJWEJWK joseJwk.Key, expectedJWEJWKOps joseJwk.KeyOperationList, testCase *happyPathJWETestCase) {
	t.Helper()

	var actualJWKAlg joseJwa.KeyEncryptionAlgorithm

	require.NoError(t, nonPublicJWEJWK.Get(joseJwk.AlgorithmKey, &actualJWKAlg))
	require.Equal(t, *testCase.alg, actualJWKAlg)

	var actualJWKOps joseJwk.KeyOperationList

	require.NoError(t, nonPublicJWEJWK.Get(joseJwk.KeyOpsKey, &actualJWKOps))
	require.Equal(t, expectedJWEJWKOps, actualJWKOps)

	var actualJWKKty joseJwa.KeyType

	require.NoError(t, nonPublicJWEJWK.Get(joseJwk.KeyTypeKey, &actualJWKKty))
	require.Equal(t, testCase.expectedType, actualJWKKty)

	var actualJWKUse string

	require.NoError(t, nonPublicJWEJWK.Get(joseJwk.KeyUsageKey, &actualJWKUse))
	require.Equal(t, "enc", actualJWKUse)
}

func requireJWEMessageHeaders(t *testing.T, jweMessage *joseJwe.Message, actualKeyKid *googleUuid.UUID, testCase *happyPathJWETestCase) {
	t.Helper()

	jweHeaders := jweMessage.ProtectedHeaders()
	encodedJWEHeaders, err := json.Marshal(jweHeaders)
	require.NoError(t, err)
	log.Printf("JWE Message Headers: %v", string(encodedJWEHeaders))

	var actualJWEKid string

	require.NoError(t, jweHeaders.Get(joseJwk.KeyIDKey, &actualJWEKid))
	require.NotEmpty(t, actualJWEKid)
	require.Equal(t, actualKeyKid.String(), actualJWEKid)

	var actualJWEEnc joseJwa.ContentEncryptionAlgorithm

	require.NoError(t, jweHeaders.Get("enc", &actualJWEEnc))
	// require.Equal(t, AlgCekA256GCM, actualJWEEnc)

	var actualJWEAlg joseJwa.KeyAlgorithm

	require.NoError(t, jweHeaders.Get(joseJwk.AlgorithmKey, &actualJWEAlg))
	require.Equal(t, *testCase.alg, actualJWEAlg)
}

func Test_HappyPath_NonJWKGenService_JWE_JWK_EncryptDecryptKey(t *testing.T) {
	t.Parallel()

	for _, testCase := range happyPathJWETestCases {
		t.Run(fmt.Sprintf("%s %s", testCase.enc, testCase.alg), func(t *testing.T) {
			t.Parallel()

			_, nonPublicJWKKek, publicJWEJWKKek, clearNonPublicKekJWKBytes, _, err := GenerateJWEJWKForEncAndAlg(testCase.enc, testCase.alg)
			require.NoError(t, err)
			log.Printf("KEK: %s", string(clearNonPublicKekJWKBytes))

			_, nonPublicJWK, _, clearNonPublicJWEJWKBytes, _, err := GenerateJWEJWKForEncAndAlg(testCase.enc, testCase.alg)
			require.NoError(t, err)
			log.Printf("Original Key: %s", string(clearNonPublicJWEJWKBytes))

			var encryptJWKKek joseJwk.Key
			if publicJWEJWKKek == nil {
				encryptJWKKek = nonPublicJWKKek
			} else {
				encryptJWKKek = publicJWEJWKKek
			}

			jweMessage, encodedJWEMessage, err := EncryptKey([]joseJwk.Key{encryptJWKKek}, nonPublicJWK)
			require.NoError(t, err)
			jsonHeaders, err := JWEHeadersString(jweMessage)
			require.NoError(t, err)
			log.Printf("JWE Message Headers: %s", jsonHeaders)

			decryptedNonPublicJWK, err := DecryptKey([]joseJwk.Key{nonPublicJWKKek}, encodedJWEMessage)
			require.NoError(t, err)

			decryptedEncodedKey, err := json.Marshal(decryptedNonPublicJWK)
			require.NoError(t, err)
			log.Printf("Decrypted Key: %s", string(decryptedEncodedKey))

			require.ElementsMatch(t, nonPublicJWK.Keys(), decryptedNonPublicJWK.Keys())
		})
	}
}

func Test_SadPath_EncryptBytes_NilKey(t *testing.T) {
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
	nonPublicJWEJWKs, publicJWEJWKs, err := GenerateJWEJWKsForTest(t, 2, &EncInvalid, &AlgA256KW)
	require.Error(t, err)
	require.Equal(t, "unexpected 2 errors: invalid JWE JWK headers: JWE JWK length error: unsupported JWE JWK enc invalid\ninvalid JWE JWK headers: JWE JWK length error: unsupported JWE JWK enc invalid", err.Error())
	require.Nil(t, nonPublicJWEJWKs)
	require.Nil(t, publicJWEJWKs)
}

func Test_SadPath_ConcurrentGenerateJWEJWK_UnsupportedAlg(t *testing.T) {
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
