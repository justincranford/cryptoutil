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

type happyPathJweTestCase struct {
	enc          *joseJwa.ContentEncryptionAlgorithm
	alg          *joseJwa.KeyEncryptionAlgorithm
	expectedType joseJwa.KeyType
}

var happyPathJweTestCases = []happyPathJweTestCase{
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

func Test_HappyPath_NonJWKGenService_Jwe_JWK_EncryptDecryptBytes(t *testing.T) {
	for _, testCase := range happyPathJweTestCases {
		cleartext := fmt.Appendf(nil, "Hello world enc=%s alg=%s!", testCase.enc, testCase.alg)
		t.Run(fmt.Sprintf("%s %s", testCase.enc, testCase.alg), func(t *testing.T) {
			t.Parallel()

			actualKeyKid, nonPublicJweJWK, publicJweJWK, clearNonPublicJweJWKBytes, clearPublicJweJWKBytes, err := GenerateJweJWKForEncAndAlg(testCase.enc, testCase.alg)
			require.NoError(t, err)
			require.NotNil(t, nonPublicJweJWK)
			require.NotEmpty(t, clearNonPublicJweJWKBytes)
			require.NotEmpty(t, actualKeyKid)
			log.Printf("Generated:\n%s\n%s", clearNonPublicJweJWKBytes, clearPublicJweJWKBytes)

			var encryptJWK joseJwk.Key
			requireJweJWKHeaders(t, nonPublicJweJWK, OpsEncDec, &testCase)
			if publicJweJWK == nil {
				encryptJWK = nonPublicJweJWK
			} else {
				encryptJWK = publicJweJWK
				requireJweJWKHeaders(t, publicJweJWK, OpsEnc, &testCase)
			}

			jweMessage, encodedJweMessage, err := EncryptBytes([]joseJwk.Key{encryptJWK}, cleartext)
			require.NoError(t, err)
			require.NotEmpty(t, encodedJweMessage)
			log.Printf("JWE Message: %s", string(encodedJweMessage))

			requireJweMessageHeaders(t, jweMessage, actualKeyKid, &testCase)

			decryptedtext, err := DecryptBytes([]joseJwk.Key{nonPublicJweJWK}, encodedJweMessage)
			require.NoError(t, err)
			require.Equal(t, cleartext, decryptedtext)
		})
	}
}

func requireJweJWKHeaders(t *testing.T, nonPublicJweJWK joseJwk.Key, expectedJweJWKOps joseJwk.KeyOperationList, testCase *happyPathJweTestCase) {
	var actualJWKAlg joseJwa.KeyEncryptionAlgorithm
	require.NoError(t, nonPublicJweJWK.Get(joseJwk.AlgorithmKey, &actualJWKAlg))
	require.Equal(t, *testCase.alg, actualJWKAlg)

	var actualJWKOps joseJwk.KeyOperationList
	require.NoError(t, nonPublicJweJWK.Get(joseJwk.KeyOpsKey, &actualJWKOps))
	require.Equal(t, expectedJweJWKOps, actualJWKOps)

	var actualJWKKty joseJwa.KeyType
	require.NoError(t, nonPublicJweJWK.Get(joseJwk.KeyTypeKey, &actualJWKKty))
	require.Equal(t, testCase.expectedType, actualJWKKty)

	var actualJWKUse string
	require.NoError(t, nonPublicJweJWK.Get(joseJwk.KeyUsageKey, &actualJWKUse))
	require.Equal(t, "enc", actualJWKUse)
}

func requireJweMessageHeaders(t *testing.T, jweMessage *joseJwe.Message, actualKeyKid *googleUuid.UUID, testCase *happyPathJweTestCase) {
	jweHeaders := jweMessage.ProtectedHeaders()
	encodedJweHeaders, err := json.Marshal(jweHeaders)
	require.NoError(t, err)
	log.Printf("JWE Message Headers: %v", string(encodedJweHeaders))

	var actualJweKid string
	require.NoError(t, jweHeaders.Get(joseJwk.KeyIDKey, &actualJweKid))
	require.NotEmpty(t, actualJweKid)
	require.Equal(t, actualKeyKid.String(), actualJweKid)

	var actualJweEnc joseJwa.ContentEncryptionAlgorithm
	require.NoError(t, jweHeaders.Get("enc", &actualJweEnc))
	// require.Equal(t, AlgCekA256GCM, actualJweEnc)

	var actualJweAlg joseJwa.KeyAlgorithm
	require.NoError(t, jweHeaders.Get(joseJwk.AlgorithmKey, &actualJweAlg))
	require.Equal(t, *testCase.alg, actualJweAlg)
}

func Test_HappyPath_NonJWKGenService_Jwe_JWK_EncryptDecryptKey(t *testing.T) {
	for _, testCase := range happyPathJweTestCases {
		t.Run(fmt.Sprintf("%s %s", testCase.enc, testCase.alg), func(t *testing.T) {
			t.Parallel()

			_, nonPublicJWKKek, publicJweJWKKek, clearNonPublicKekJWKBytes, _, err := GenerateJweJWKForEncAndAlg(testCase.enc, testCase.alg)
			require.NoError(t, err)
			log.Printf("KEK: %s", string(clearNonPublicKekJWKBytes))

			_, nonPublicJWK, _, clearNonPublicJweJWKBytes, _, err := GenerateJweJWKForEncAndAlg(testCase.enc, testCase.alg)
			require.NoError(t, err)
			log.Printf("Original Key: %s", string(clearNonPublicJweJWKBytes))

			var encryptJWKKek joseJwk.Key
			if publicJweJWKKek == nil {
				encryptJWKKek = nonPublicJWKKek
			} else {
				encryptJWKKek = publicJweJWKKek
			}

			jweMessage, encodedJweMessage, err := EncryptKey([]joseJwk.Key{encryptJWKKek}, nonPublicJWK)
			require.NoError(t, err)
			jsonHeaders, err := JweHeadersString(jweMessage)
			require.NoError(t, err)
			log.Printf("JWE Message Headers: %s", jsonHeaders)

			decryptedNonPublicJWK, err := DecryptKey([]joseJwk.Key{nonPublicJWKKek}, encodedJweMessage)
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

func Test_SadPath_DecryptBytes_InvalidJweMessage(t *testing.T) {
	kid, nonPublicJweJWK, _, clearNonPublicJweJWKBytes, _, err := GenerateJweJWKForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, kid)
	require.NotNil(t, nonPublicJweJWK)
	isEncryptJWK, err := IsEncryptJWK(nonPublicJweJWK)
	require.NoError(t, err)
	require.True(t, isEncryptJWK)
	require.NotNil(t, clearNonPublicJweJWKBytes)

	_, err = DecryptBytes([]joseJwk.Key{nonPublicJweJWK}, []byte("this-is-not-a-valid-jwe-message"))
	require.Error(t, err)
}

func Test_SadPath_GenerateJweJWK_UnsupportedEnc(t *testing.T) {
	kid, nonPublicJweJWK, publicJweJWK, clearNonPublicJweJWKBytes, clearPublicJweJWKBytes, err := GenerateJweJWKForEncAndAlg(&EncInvalid, &AlgA256KW)
	require.Error(t, err)
	require.Equal(t, "invalid JWE JWK headers: JWE JWK length error: unsupported JWE JWK enc invalid", err.Error())
	require.Nil(t, kid)
	require.Nil(t, nonPublicJweJWK)
	require.Nil(t, publicJweJWK)
	require.Nil(t, clearNonPublicJweJWKBytes)
	require.Nil(t, clearPublicJweJWKBytes)
}

func Test_SadPath_GenerateJweJWK_UnsupportedAlg(t *testing.T) {
	kid, nonPublicJweJWK, publicJweJWK, clearNonPublicJweJWKBytes, clearPublicJweJWKBytes, err := GenerateJweJWKForEncAndAlg(&EncA256GCM, &AlgEncInvalid)
	require.Error(t, err)
	require.Equal(t, "invalid JWE JWK headers: unsupported JWE JWK alg invalid", err.Error())
	require.Nil(t, kid)
	require.Nil(t, nonPublicJweJWK)
	require.Nil(t, publicJweJWK)
	require.Nil(t, clearNonPublicJweJWKBytes)
	require.Nil(t, clearPublicJweJWKBytes)
}

func Test_SadPath_ConcurrentGenerateJweJWK_UnsupportedEnc(t *testing.T) {
	nonPublicJweJWKs, publicJweJWKs, err := GenerateJweJWKsForTest(t, 2, &EncInvalid, &AlgA256KW)
	require.Error(t, err)
	require.Equal(t, "unexpected 2 errors: invalid JWE JWK headers: JWE JWK length error: unsupported JWE JWK enc invalid\ninvalid JWE JWK headers: JWE JWK length error: unsupported JWE JWK enc invalid", err.Error())
	require.Nil(t, nonPublicJweJWKs)
	require.Nil(t, publicJweJWKs)
}

func Test_SadPath_ConcurrentGenerateJweJWK_UnsupportedAlg(t *testing.T) {
	nonPublicJweJWKs, publicJweJWKs, err := GenerateJweJWKsForTest(t, 2, &EncA256GCM, &AlgEncInvalid)
	require.Error(t, err)
	require.Equal(t, "unexpected 2 errors: invalid JWE JWK headers: unsupported JWE JWK alg invalid\ninvalid JWE JWK headers: unsupported JWE JWK alg invalid", err.Error())
	require.Nil(t, nonPublicJweJWKs)
	require.Nil(t, publicJweJWKs)
}
