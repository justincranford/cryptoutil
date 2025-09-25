package jose

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
)

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

func Test_SadPath_DecryptBytes_NilKey(t *testing.T) {
	_, err := DecryptBytes(nil, []byte("cleartext"))
	require.Error(t, err)
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
