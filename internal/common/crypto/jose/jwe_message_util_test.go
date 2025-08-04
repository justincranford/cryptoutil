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
	// {enc: &EncA256GCM, alg: &AlgA256KW, expectedType: KtyOCT},
	// {enc: &EncA192GCM, alg: &AlgA256KW, expectedType: KtyOCT},
	// {enc: &EncA128GCM, alg: &AlgA256KW, expectedType: KtyOCT},
	// {enc: &EncA256GCM, alg: &AlgA192KW, expectedType: KtyOCT},
	// {enc: &EncA192GCM, alg: &AlgA192KW, expectedType: KtyOCT},
	// {enc: &EncA128GCM, alg: &AlgA192KW, expectedType: KtyOCT},
	// {enc: &EncA256GCM, alg: &AlgA128KW, expectedType: KtyOCT},
	// {enc: &EncA192GCM, alg: &AlgA128KW, expectedType: KtyOCT},
	// {enc: &EncA128GCM, alg: &AlgA128KW, expectedType: KtyOCT},

	// {enc: &EncA256GCM, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	// {enc: &EncA192GCM, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	// {enc: &EncA128GCM, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	// {enc: &EncA256GCM, alg: &AlgA192GCMKW, expectedType: KtyOCT},
	// {enc: &EncA192GCM, alg: &AlgA192GCMKW, expectedType: KtyOCT},
	// {enc: &EncA128GCM, alg: &AlgA192GCMKW, expectedType: KtyOCT},
	// {enc: &EncA256GCM, alg: &AlgA128GCMKW, expectedType: KtyOCT},
	// {enc: &EncA192GCM, alg: &AlgA128GCMKW, expectedType: KtyOCT},
	// {enc: &EncA128GCM, alg: &AlgA128GCMKW, expectedType: KtyOCT},

	// {enc: &EncA256GCM, alg: &AlgDir, expectedType: KtyOCT},
	// {enc: &EncA192GCM, alg: &AlgDir, expectedType: KtyOCT},
	// {enc: &EncA128GCM, alg: &AlgDir, expectedType: KtyOCT},

	// {enc: &EncA256GCM, alg: &AlgRSAOAEP512, expectedType: KtyRSA},
	// {enc: &EncA192GCM, alg: &AlgRSAOAEP512, expectedType: KtyRSA},
	// {enc: &EncA128GCM, alg: &AlgRSAOAEP512, expectedType: KtyRSA},
	// {enc: &EncA256GCM, alg: &AlgRSAOAEP384, expectedType: KtyRSA},
	// {enc: &EncA192GCM, alg: &AlgRSAOAEP384, expectedType: KtyRSA},
	// {enc: &EncA128GCM, alg: &AlgRSAOAEP384, expectedType: KtyRSA},
	// {enc: &EncA256GCM, alg: &AlgRSAOAEP256, expectedType: KtyRSA},
	// {enc: &EncA192GCM, alg: &AlgRSAOAEP256, expectedType: KtyRSA},
	// {enc: &EncA128GCM, alg: &AlgRSAOAEP256, expectedType: KtyRSA},
	// {enc: &EncA256GCM, alg: &AlgRSAOAEP, expectedType: KtyRSA},
	// {enc: &EncA192GCM, alg: &AlgRSAOAEP, expectedType: KtyRSA},
	// {enc: &EncA128GCM, alg: &AlgRSAOAEP, expectedType: KtyRSA},
	// {enc: &EncA256GCM, alg: &AlgRSA15, expectedType: KtyRSA},
	// {enc: &EncA192GCM, alg: &AlgRSA15, expectedType: KtyRSA},
	// {enc: &EncA128GCM, alg: &AlgRSA15, expectedType: KtyRSA},

	{enc: &EncA256GCM, alg: &AlgECDHESA256KW, expectedType: KtyEC},
	// {enc: &EncA192GCM, alg: &AlgECDHESA256KW, expectedType: KtyEC},
	// {enc: &EncA128GCM, alg: &AlgECDHESA256KW, expectedType: KtyEC},
	// {enc: &EncA256GCM, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	// {enc: &EncA192GCM, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	// {enc: &EncA128GCM, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	// {enc: &EncA256GCM, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	// {enc: &EncA192GCM, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	// {enc: &EncA128GCM, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	// {enc: &EncA256GCM, alg: &AlgECDHES, expectedType: KtyEC},
	// {enc: &EncA192GCM, alg: &AlgECDHES, expectedType: KtyEC},
	// {enc: &EncA128GCM, alg: &AlgECDHES, expectedType: KtyEC},

	// {enc: &EncA256CBCHS512, alg: &AlgA256KW, expectedType: KtyOCT},
	// {enc: &EncA192CBCHS384, alg: &AlgA256KW, expectedType: KtyOCT},
	// {enc: &EncA128CBCHS256, alg: &AlgA256KW, expectedType: KtyOCT},
	// {enc: &EncA256CBCHS512, alg: &AlgA192KW, expectedType: KtyOCT},
	// {enc: &EncA192CBCHS384, alg: &AlgA192KW, expectedType: KtyOCT},
	// {enc: &EncA128CBCHS256, alg: &AlgA192KW, expectedType: KtyOCT},
	// {enc: &EncA256CBCHS512, alg: &AlgA128KW, expectedType: KtyOCT},
	// {enc: &EncA192CBCHS384, alg: &AlgA128KW, expectedType: KtyOCT},
	// {enc: &EncA128CBCHS256, alg: &AlgA128KW, expectedType: KtyOCT},

	// {enc: &EncA256CBCHS512, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	// {enc: &EncA192CBCHS384, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	// {enc: &EncA128CBCHS256, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	// {enc: &EncA256CBCHS512, alg: &AlgA192GCMKW, expectedType: KtyOCT},
	// {enc: &EncA192CBCHS384, alg: &AlgA192GCMKW, expectedType: KtyOCT},
	// {enc: &EncA128CBCHS256, alg: &AlgA192GCMKW, expectedType: KtyOCT},
	// {enc: &EncA256CBCHS512, alg: &AlgA128GCMKW, expectedType: KtyOCT},
	// {enc: &EncA192CBCHS384, alg: &AlgA128GCMKW, expectedType: KtyOCT},
	// {enc: &EncA128CBCHS256, alg: &AlgA128GCMKW, expectedType: KtyOCT},

	// {enc: &EncA256CBCHS512, alg: &AlgDir, expectedType: KtyOCT},
	// {enc: &EncA192CBCHS384, alg: &AlgDir, expectedType: KtyOCT},
	// {enc: &EncA128CBCHS256, alg: &AlgDir, expectedType: KtyOCT},

	// {enc: &EncA256CBCHS512, alg: &AlgRSAOAEP512, expectedType: KtyRSA},
	// {enc: &EncA192CBCHS384, alg: &AlgRSAOAEP512, expectedType: KtyRSA},
	// {enc: &EncA128CBCHS256, alg: &AlgRSAOAEP512, expectedType: KtyRSA},
	// {enc: &EncA256CBCHS512, alg: &AlgRSAOAEP384, expectedType: KtyRSA},
	// {enc: &EncA192CBCHS384, alg: &AlgRSAOAEP384, expectedType: KtyRSA},
	// {enc: &EncA128CBCHS256, alg: &AlgRSAOAEP384, expectedType: KtyRSA},
	// {enc: &EncA256CBCHS512, alg: &AlgRSAOAEP256, expectedType: KtyRSA},
	// {enc: &EncA192CBCHS384, alg: &AlgRSAOAEP256, expectedType: KtyRSA},
	// {enc: &EncA128CBCHS256, alg: &AlgRSAOAEP256, expectedType: KtyRSA},
	// {enc: &EncA256CBCHS512, alg: &AlgRSAOAEP, expectedType: KtyRSA},
	// {enc: &EncA192CBCHS384, alg: &AlgRSAOAEP, expectedType: KtyRSA},
	// {enc: &EncA128CBCHS256, alg: &AlgRSAOAEP, expectedType: KtyRSA},
	// {enc: &EncA256CBCHS512, alg: &AlgRSA15, expectedType: KtyRSA},
	// {enc: &EncA192CBCHS384, alg: &AlgRSA15, expectedType: KtyRSA},
	// {enc: &EncA128CBCHS256, alg: &AlgRSA15, expectedType: KtyRSA},

	// {enc: &EncA256CBCHS512, alg: &AlgECDHESA256KW, expectedType: KtyEC},
	// {enc: &EncA192CBCHS384, alg: &AlgECDHESA256KW, expectedType: KtyEC},
	// {enc: &EncA128CBCHS256, alg: &AlgECDHESA256KW, expectedType: KtyEC},
	// {enc: &EncA256CBCHS512, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	// {enc: &EncA192CBCHS384, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	// {enc: &EncA128CBCHS256, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	// {enc: &EncA256CBCHS512, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	// {enc: &EncA192CBCHS384, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	// {enc: &EncA128CBCHS256, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	// {enc: &EncA256CBCHS512, alg: &AlgECDHES, expectedType: KtyEC},
	// {enc: &EncA192CBCHS384, alg: &AlgECDHES, expectedType: KtyEC},
	// {enc: &EncA128CBCHS256, alg: &AlgECDHES, expectedType: KtyEC},
}

func Test_HappyPath_NonJwkGenService_Jwe_Jwk_EncryptDecryptBytes(t *testing.T) {
	for _, testCase := range happyPathJweTestCases {
		cleartext := fmt.Appendf(nil, "Hello world enc=%s alg=%s!", testCase.enc, testCase.alg)
		t.Run(fmt.Sprintf("%s %s", testCase.enc, testCase.alg), func(t *testing.T) {
			t.Parallel()

			actualKeyKid, nonPublicJweJwk, publicJweJwk, clearNonPublicJweJwkBytes, clearPublicJweJwkBytes, err := GenerateJweJwkForEncAndAlg(testCase.enc, testCase.alg)
			require.NoError(t, err)
			require.NotNil(t, nonPublicJweJwk)
			// TODO Util to check AsymmetricJWK vs SymmetricJWK
			// require.NotNil(t, publicJweJwk)
			require.NotEmpty(t, clearNonPublicJweJwkBytes)
			// require.NotEmpty(t, encodedPublicJweJwk)
			require.NotEmpty(t, actualKeyKid)
			log.Printf("Generated:\n%s\n%s", clearNonPublicJweJwkBytes, clearPublicJweJwkBytes)

			// TODO Asymmetric check
			requireJweJwkHeaders(t, nonPublicJweJwk, OpsEncDec, &testCase)
			if publicJweJwk != nil {
				requireJweJwkHeaders(t, publicJweJwk, OpsEnc, &testCase)
			}

			jweMessage, encodedJweMessage, err := EncryptBytes([]joseJwk.Key{nonPublicJweJwk}, cleartext)
			require.NoError(t, err)
			require.NotEmpty(t, encodedJweMessage)
			log.Printf("JWE Message: %s", string(encodedJweMessage))

			requireJweMessageHeaders(t, jweMessage, actualKeyKid, &testCase)

			decryptedtext, err := DecryptBytes([]joseJwk.Key{nonPublicJweJwk}, encodedJweMessage)
			require.NoError(t, err)
			require.Equal(t, cleartext, decryptedtext)
		})
	}
}

func requireJweJwkHeaders(t *testing.T, nonPublicJweJwk joseJwk.Key, expectedJweJwkOps joseJwk.KeyOperationList, testCase *happyPathJweTestCase) {
	var actualJwkAlg joseJwa.KeyEncryptionAlgorithm
	require.NoError(t, nonPublicJweJwk.Get(joseJwk.AlgorithmKey, &actualJwkAlg))
	require.Equal(t, *testCase.alg, actualJwkAlg)

	var actualJwkOps joseJwk.KeyOperationList
	require.NoError(t, nonPublicJweJwk.Get(joseJwk.KeyOpsKey, &actualJwkOps))
	require.Equal(t, expectedJweJwkOps, actualJwkOps)

	var actualJwkKty joseJwa.KeyType
	require.NoError(t, nonPublicJweJwk.Get(joseJwk.KeyTypeKey, &actualJwkKty))
	require.Equal(t, testCase.expectedType, actualJwkKty)

	var actualJwkUse string
	require.NoError(t, nonPublicJweJwk.Get(joseJwk.KeyUsageKey, &actualJwkUse))
	require.Equal(t, "enc", actualJwkUse)
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

func Test_HappyPath_NonJwkGenService_Jwe_Jwk_EncryptDecryptKey(t *testing.T) {
	for _, testCase := range happyPathJweTestCases {
		t.Run(fmt.Sprintf("%s %s", testCase.enc, testCase.alg), func(t *testing.T) {
			t.Parallel()

			_, nonPublicJwsJwkKek, _, clearNonPublicJwsKekJwkBytes, _, err := GenerateJweJwkForEncAndAlg(testCase.enc, testCase.alg)
			require.NoError(t, err)
			log.Printf("KEK: %s", string(clearNonPublicJwsKekJwkBytes))

			_, nonPublicJwsJwk, _, clearNonPublicJwsJwkBytes, _, err := GenerateJweJwkForEncAndAlg(testCase.enc, testCase.alg)
			require.NoError(t, err)
			log.Printf("Original Key: %s", string(clearNonPublicJwsJwkBytes))

			jweMessage, encodedJweMessage, err := EncryptKey([]joseJwk.Key{nonPublicJwsJwkKek}, nonPublicJwsJwk)
			require.NoError(t, err)
			jsonHeaders, err := JweHeadersString(jweMessage)
			require.NoError(t, err)
			log.Printf("JWE Message Headers: %s", jsonHeaders)

			decryptedNonPublicJwsJwk, err := DecryptKey([]joseJwk.Key{nonPublicJwsJwkKek}, encodedJweMessage)
			require.NoError(t, err)

			decryptedEncodedKey, err := json.Marshal(decryptedNonPublicJwsJwk)
			require.NoError(t, err)
			log.Printf("Decrypted Key: %s", string(decryptedEncodedKey))

			require.ElementsMatch(t, nonPublicJwsJwk.Keys(), decryptedNonPublicJwsJwk.Keys())
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
	kid, nonPublicJweJwk, _, clearNonPublicJweJwkBytes, _, err := GenerateJweJwkForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, kid)
	require.NotNil(t, nonPublicJweJwk)
	// TODO Util to check AsymmetricJWK vs SymmetricJWK
	// require.NotNil(t, publicJweJwk)
	require.NotNil(t, clearNonPublicJweJwkBytes)
	// require.NotNil(t, encodedPublicJweJwk)

	_, err = DecryptBytes([]joseJwk.Key{nonPublicJweJwk}, []byte("this-is-not-a-valid-jwe-message"))
	require.Error(t, err)
}

func Test_SadPath_GenerateJweJwk_UnsupportedEnc(t *testing.T) {
	kid, nonPublicJweJwk, publicJweJwk, clearNonPublicJweJwkBytes, clearPublicJweJwkBytes, err := GenerateJweJwkForEncAndAlg(&EncInvalid, &AlgA256KW)
	require.Error(t, err)
	require.Equal(t, "invalid JWE JWK headers: JWE JWK length error: unsupported JWE JWK enc invalid", err.Error())
	require.Nil(t, kid)
	require.Nil(t, nonPublicJweJwk)
	require.Nil(t, publicJweJwk)
	require.Nil(t, clearNonPublicJweJwkBytes)
	require.Nil(t, clearPublicJweJwkBytes)
}

func Test_SadPath_GenerateJweJwk_UnsupportedAlg(t *testing.T) {
	kid, nonPublicJweJwk, publicJweJwk, clearNonPublicJweJwkBytes, clearPublicJweJwkBytes, err := GenerateJweJwkForEncAndAlg(&EncA256GCM, &AlgEncInvalid)
	require.Error(t, err)
	require.Equal(t, "invalid JWE JWK headers: unsupported JWE JWK alg invalid", err.Error())
	require.Nil(t, kid)
	require.Nil(t, nonPublicJweJwk)
	require.Nil(t, publicJweJwk)
	require.Nil(t, clearNonPublicJweJwkBytes)
	require.Nil(t, clearPublicJweJwkBytes)
}

func Test_SadPath_ConcurrentGenerateJweJwk_UnsupportedEnc(t *testing.T) {
	nonPublicJweJwks, publicJweJwks, err := GenerateJweJwksForTest(t, 2, &EncInvalid, &AlgA256KW)
	require.Error(t, err)
	require.Equal(t, "unexpected 2 errors: invalid JWE JWK headers: JWE JWK length error: unsupported JWE JWK enc invalid\ninvalid JWE JWK headers: JWE JWK length error: unsupported JWE JWK enc invalid", err.Error())
	require.Nil(t, nonPublicJweJwks)
	require.Nil(t, publicJweJwks)
}

func Test_SadPath_ConcurrentGenerateJweJwk_UnsupportedAlg(t *testing.T) {
	nonPublicJweJwks, publicJweJwks, err := GenerateJweJwksForTest(t, 2, &EncA256GCM, &AlgEncInvalid)
	require.Error(t, err)
	require.Equal(t, "unexpected 2 errors: invalid JWE JWK headers: unsupported JWE JWK alg invalid\ninvalid JWE JWK headers: unsupported JWE JWK alg invalid", err.Error())
	require.Nil(t, nonPublicJweJwks)
	require.Nil(t, publicJweJwks)
}
