package jose

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

var happyPathJweTestCases = []struct {
	enc          *joseJwa.ContentEncryptionAlgorithm
	alg          *joseJwa.KeyEncryptionAlgorithm
	expectedType joseJwa.KeyType
}{
	{enc: &EncA256GCM, alg: &AlgA256KW, expectedType: KtyOCT},
	{enc: &EncA192GCM, alg: &AlgA256KW, expectedType: KtyOCT},
	{enc: &EncA128GCM, alg: &AlgA256KW, expectedType: KtyOCT},
	{enc: &EncA192GCM, alg: &AlgA192KW, expectedType: KtyOCT},
	{enc: &EncA128GCM, alg: &AlgA192KW, expectedType: KtyOCT},
	{enc: &EncA128GCM, alg: &AlgA128KW, expectedType: KtyOCT},

	{enc: &EncA256GCM, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	{enc: &EncA192GCM, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	{enc: &EncA128GCM, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	{enc: &EncA192GCM, alg: &AlgA192GCMKW, expectedType: KtyOCT},
	{enc: &EncA128GCM, alg: &AlgA192GCMKW, expectedType: KtyOCT},
	{enc: &EncA128GCM, alg: &AlgA128GCMKW, expectedType: KtyOCT},

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

	// {enc: &EncA256GCM, alg: &AlgECDHES, expectedType: KtyEC},
	// {enc: &EncA192GCM, alg: &AlgECDHES, expectedType: KtyEC},
	{enc: &EncA128GCM, alg: &AlgECDHES, expectedType: KtyEC},
	// {enc: &EncA256GCM, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	// {enc: &EncA192GCM, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	{enc: &EncA128GCM, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	// {enc: &EncA256GCM, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	{enc: &EncA192GCM, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	{enc: &EncA128GCM, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	{enc: &EncA256GCM, alg: &AlgECDHESA256KW, expectedType: KtyEC},
	{enc: &EncA192GCM, alg: &AlgECDHESA256KW, expectedType: KtyEC},
	{enc: &EncA128GCM, alg: &AlgECDHESA256KW, expectedType: KtyEC},

	{enc: &EncA256GCM, alg: &AlgDir, expectedType: KtyOCT},
	{enc: &EncA192GCM, alg: &AlgDir, expectedType: KtyOCT},
	{enc: &EncA128GCM, alg: &AlgDir, expectedType: KtyOCT},

	{enc: &EncA256CBC_HS512, alg: &AlgA256KW, expectedType: KtyOCT},
	{enc: &EncA192CBC_HS384, alg: &AlgA256KW, expectedType: KtyOCT},
	{enc: &EncA128CBC_HS256, alg: &AlgA256KW, expectedType: KtyOCT},
	{enc: &EncA192CBC_HS384, alg: &AlgA192KW, expectedType: KtyOCT},
	{enc: &EncA128CBC_HS256, alg: &AlgA192KW, expectedType: KtyOCT},
	{enc: &EncA128CBC_HS256, alg: &AlgA128KW, expectedType: KtyOCT},

	{enc: &EncA256CBC_HS512, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	{enc: &EncA192CBC_HS384, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	{enc: &EncA128CBC_HS256, alg: &AlgA256GCMKW, expectedType: KtyOCT},
	{enc: &EncA192CBC_HS384, alg: &AlgA192GCMKW, expectedType: KtyOCT},
	{enc: &EncA128CBC_HS256, alg: &AlgA192GCMKW, expectedType: KtyOCT},
	{enc: &EncA128CBC_HS256, alg: &AlgA192GCMKW, expectedType: KtyOCT},

	{enc: &EncA256CBC_HS512, alg: &AlgRSAOAEP512, expectedType: KtyRSA},
	{enc: &EncA192CBC_HS384, alg: &AlgRSAOAEP512, expectedType: KtyRSA},
	{enc: &EncA128CBC_HS256, alg: &AlgRSAOAEP512, expectedType: KtyRSA},
	{enc: &EncA256CBC_HS512, alg: &AlgRSAOAEP384, expectedType: KtyRSA},
	{enc: &EncA192CBC_HS384, alg: &AlgRSAOAEP384, expectedType: KtyRSA},
	{enc: &EncA128CBC_HS256, alg: &AlgRSAOAEP384, expectedType: KtyRSA},
	{enc: &EncA256CBC_HS512, alg: &AlgRSAOAEP256, expectedType: KtyRSA},
	{enc: &EncA192CBC_HS384, alg: &AlgRSAOAEP256, expectedType: KtyRSA},
	{enc: &EncA128CBC_HS256, alg: &AlgRSAOAEP256, expectedType: KtyRSA},
	{enc: &EncA256CBC_HS512, alg: &AlgRSAOAEP, expectedType: KtyRSA},
	{enc: &EncA192CBC_HS384, alg: &AlgRSAOAEP, expectedType: KtyRSA},
	{enc: &EncA128CBC_HS256, alg: &AlgRSAOAEP, expectedType: KtyRSA},
	{enc: &EncA256CBC_HS512, alg: &AlgRSA15, expectedType: KtyRSA},
	{enc: &EncA192CBC_HS384, alg: &AlgRSA15, expectedType: KtyRSA},
	{enc: &EncA128CBC_HS256, alg: &AlgRSA15, expectedType: KtyRSA},

	// {enc: &EncA256CBC_HS512, alg: &AlgECDHES, expectedType: KtyEC},
	// {enc: &EncA192CBC_HS384, alg: &AlgECDHES, expectedType: KtyEC},
	{enc: &EncA128CBC_HS256, alg: &AlgECDHES, expectedType: KtyEC},
	// {enc: &EncA256CBC_HS512, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	// {enc: &EncA192CBC_HS384, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	{enc: &EncA128CBC_HS256, alg: &AlgECDHESA128KW, expectedType: KtyEC},
	// {enc: &EncA256CBC_HS512, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	{enc: &EncA192CBC_HS384, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	{enc: &EncA128CBC_HS256, alg: &AlgECDHESA192KW, expectedType: KtyEC},
	{enc: &EncA256CBC_HS512, alg: &AlgECDHESA256KW, expectedType: KtyEC},
	{enc: &EncA192CBC_HS384, alg: &AlgECDHESA256KW, expectedType: KtyEC},
	{enc: &EncA128CBC_HS256, alg: &AlgECDHESA256KW, expectedType: KtyEC},

	{enc: &EncA256CBC_HS512, alg: &AlgDir, expectedType: KtyOCT},
	{enc: &EncA192CBC_HS384, alg: &AlgDir, expectedType: KtyOCT},
	{enc: &EncA128CBC_HS256, alg: &AlgDir, expectedType: KtyOCT},
}

func Test_HappyPath_NonJwkGenService_Jwe_Jwk_EncryptDecryptBytes(t *testing.T) {
	for _, testCase := range happyPathJweTestCases {
		plaintext := fmt.Appendf(nil, "Hello world enc=%s alg=%s!", testCase.enc, testCase.alg)
		t.Run(fmt.Sprintf("%s %s", testCase.enc, testCase.alg), func(t *testing.T) {
			t.Parallel()

			actualKeyKid, cek, encodedJweJwk, err := GenerateJweJwkForEncAndAlg(testCase.enc, testCase.alg)
			require.NoError(t, err)
			require.NotNil(t, cek)
			require.NotEmpty(t, encodedJweJwk)
			require.NotEmpty(t, actualKeyKid)
			log.Printf("Generated: %s", encodedJweJwk)

			var actualJwkAlg joseJwa.KeyAlgorithm
			require.NoError(t, cek.Get(joseJwk.AlgorithmKey, &actualJwkAlg))
			require.Equal(t, *testCase.alg, actualJwkAlg)

			var actualJwkUse string
			require.NoError(t, cek.Get(joseJwk.KeyUsageKey, &actualJwkUse))
			require.Equal(t, "enc", actualJwkUse)

			var actualJwkOps joseJwk.KeyOperationList
			require.NoError(t, cek.Get(joseJwk.KeyOpsKey, &actualJwkOps))
			require.Equal(t, OpsEncDec, actualJwkOps)

			var actualJwkKty joseJwa.KeyType
			require.NoError(t, cek.Get(joseJwk.KeyTypeKey, &actualJwkKty))
			require.Equal(t, testCase.expectedType, actualJwkKty)

			jweMessage, encodedJweMessage, err := EncryptBytes([]joseJwk.Key{cek}, plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, encodedJweMessage)
			log.Printf("JWE Message: %s", string(encodedJweMessage))

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
			require.Equal(t, actualJwkAlg, actualJweAlg)

			decrypted, err := DecryptBytes([]joseJwk.Key{cek}, encodedJweMessage)
			require.NoError(t, err)
			require.Equal(t, plaintext, decrypted)
		})
	}
}

func Test_HappyPath_NonJwkGenService_Jwe_Jwk_EncryptDecryptKey(t *testing.T) {
	for _, testCase := range happyPathJweTestCases {
		t.Run(fmt.Sprintf("%s %s", testCase.enc, testCase.alg), func(t *testing.T) {
			t.Parallel()

			_, kek, encodedKek, _ := GenerateJweJwkForEncAndAlg(testCase.enc, testCase.alg)
			log.Printf("KEK: %s", string(encodedKek))

			_, originalKey, encodedKey, _ := GenerateJweJwkForEncAndAlg(testCase.enc, testCase.alg)
			log.Printf("Original Key: %s", string(encodedKey))

			jweMessage, encodedJweMessage, err := EncryptKey([]joseJwk.Key{kek}, originalKey)
			require.NoError(t, err)
			jsonHeaders, err := JweHeadersString(jweMessage)
			require.NoError(t, err)
			log.Printf("JWE Message Headers: %s", jsonHeaders)

			decryptedKey, err := DecryptKey([]joseJwk.Key{kek}, encodedJweMessage)
			require.NoError(t, err)

			decryptedEncodedKey, err := json.Marshal(decryptedKey)
			require.NoError(t, err)
			log.Printf("Decrypted Key: %s", string(decryptedEncodedKey))

			require.Equal(t, originalKey.Keys(), decryptedKey.Keys())
		})
	}
}

func Test_SadPath_EncryptBytes_NilKey(t *testing.T) {
	_, _, err := EncryptBytes(nil, []byte("test"))
	require.Error(t, err)
}

func Test_SadPath_DecryptBytes_NilKey(t *testing.T) {
	_, err := DecryptBytes(nil, []byte("ciphertext"))
	require.Error(t, err)
}

func Test_SadPath_DecryptBytes_InvalidCiphertext(t *testing.T) {
	kid, key, raw, err := GenerateJweJwkForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, kid)
	require.NotNil(t, key)
	require.NotNil(t, raw)

	_, err = DecryptBytes([]joseJwk.Key{key}, []byte("this-is-not-valid-ciphertext"))
	require.Error(t, err)
}

func Test_SadPath_GenerateJweJwk_UnsupportedEnc(t *testing.T) {
	key, raw, kid, err := GenerateJweJwkForEncAndAlg(&EncInvalid, &AlgA256KW)
	require.Error(t, err)
	require.Equal(t, "invalid JWE JWK headers: JWE JWK length error: unsupported JWE JWK enc invalid", err.Error())
	require.Nil(t, kid)
	require.Nil(t, key)
	require.Nil(t, raw)
}

func Test_SadPath_GenerateJweJwk_UnsupportedAlg(t *testing.T) {
	key, raw, kid, err := GenerateJweJwkForEncAndAlg(&EncA256GCM, &AlgEncInvalid)
	require.Error(t, err)
	require.Equal(t, "invalid JWE JWK headers: unsupported JWE JWK alg invalid", err.Error())
	require.Nil(t, kid)
	require.Nil(t, key)
	require.Nil(t, raw)
}

func Test_SadPath_ConcurrentGenerateJweJwk_UnsupportedEnc(t *testing.T) {
	jwks, err := GenerateJweJwksForTest(t, 2, &EncInvalid, &AlgA256KW)
	require.Error(t, err)
	require.Equal(t, "unexpected 2 errors: invalid JWE JWK headers: JWE JWK length error: unsupported JWE JWK enc invalid\ninvalid JWE JWK headers: JWE JWK length error: unsupported JWE JWK enc invalid", err.Error())
	require.Nil(t, jwks)
}

func Test_SadPath_ConcurrentGenerateJweJwk_UnsupportedAlg(t *testing.T) {
	jwks, err := GenerateJweJwksForTest(t, 2, &EncA256GCM, &AlgEncInvalid)
	require.Error(t, err)
	require.Equal(t, "unexpected 2 errors: invalid JWE JWK headers: unsupported JWE JWK alg invalid\ninvalid JWE JWK headers: unsupported JWE JWK alg invalid", err.Error())
	require.Nil(t, jwks)
}
