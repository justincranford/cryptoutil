// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
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
	require.Equal(t, cryptoutilSharedMagic.JoseKeyUseEnc, actualJWKUse)
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

	require.NoError(t, jweHeaders.Get(cryptoutilSharedMagic.JoseKeyUseEnc, &actualJWEEnc))
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
