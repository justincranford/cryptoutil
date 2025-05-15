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

var happyPathTestCases = []struct {
	enc *joseJwa.ContentEncryptionAlgorithm
	alg *joseJwa.KeyEncryptionAlgorithm
}{
	{enc: &EncA256GCM, alg: &AlgA256KW},
	{enc: &EncA192GCM, alg: &AlgA256KW},
	{enc: &EncA128GCM, alg: &AlgA256KW},
	{enc: &EncA192GCM, alg: &AlgA192KW},
	{enc: &EncA128GCM, alg: &AlgA192KW},
	{enc: &EncA128GCM, alg: &AlgA128KW},

	{enc: &EncA256GCM, alg: &AlgA256GCMKW},
	{enc: &EncA192GCM, alg: &AlgA256GCMKW},
	{enc: &EncA128GCM, alg: &AlgA256GCMKW},
	{enc: &EncA192GCM, alg: &AlgA192GCMKW},
	{enc: &EncA128GCM, alg: &AlgA192GCMKW},
	{enc: &EncA128GCM, alg: &AlgA128GCMKW},

	{enc: &EncA256GCM, alg: &AlgDir},
	{enc: &EncA192GCM, alg: &AlgDir},
	{enc: &EncA128GCM, alg: &AlgDir},

	{enc: &EncA256CBC_HS512, alg: &AlgA256KW},
	{enc: &EncA192CBC_HS384, alg: &AlgA256KW},
	{enc: &EncA128CBC_HS256, alg: &AlgA256KW},
	{enc: &EncA192CBC_HS384, alg: &AlgA192KW},
	{enc: &EncA128CBC_HS256, alg: &AlgA192KW},
	{enc: &EncA128CBC_HS256, alg: &AlgA128KW},

	{enc: &EncA256CBC_HS512, alg: &AlgA256GCMKW},
	{enc: &EncA192CBC_HS384, alg: &AlgA256GCMKW},
	{enc: &EncA128CBC_HS256, alg: &AlgA256GCMKW},
	{enc: &EncA192CBC_HS384, alg: &AlgA192GCMKW},
	{enc: &EncA128CBC_HS256, alg: &AlgA192GCMKW},
	{enc: &EncA128CBC_HS256, alg: &AlgA192GCMKW},

	{enc: &EncA256CBC_HS512, alg: &AlgDir},
	{enc: &EncA192CBC_HS384, alg: &AlgDir},
	{enc: &EncA128CBC_HS256, alg: &AlgDir},
}

func Test_HappyPath_Bytes(t *testing.T) {
	for _, testCase := range happyPathTestCases {
		t.Run(fmt.Sprintf("%s %s", testCase.enc, testCase.alg), func(t *testing.T) {
			t.Parallel()
			actualKeyKid, cek, encodedAesJwk, err := GenerateEncryptionJweJwkForEncAndAlg(testCase.enc, testCase.alg)
			require.NoError(t, err)
			require.NotNil(t, cek)
			require.NotEmpty(t, encodedAesJwk)
			require.NotEmpty(t, actualKeyKid)
			log.Printf("Generated: %s", encodedAesJwk)

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
			require.Equal(t, KtyOct, actualJwkKty)

			plaintext := []byte("hello, world!")
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

func Test_HappyPath_Key(t *testing.T) {
	for _, testCase := range happyPathTestCases {
		t.Run(fmt.Sprintf("%s %s", testCase.enc, testCase.alg), func(t *testing.T) {
			_, kek, encodedKek, _ := GenerateEncryptionJweJwkForEncAndAlg(testCase.enc, testCase.alg)
			log.Printf("KEK: %s", string(encodedKek))

			_, originalKey, encodedKey, _ := GenerateEncryptionJweJwkForEncAndAlg(testCase.enc, testCase.alg)
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

func Test_SadPath_GenerateAesJWK_UnsupportedAlg(t *testing.T) {
	invalidAlg := joseJwa.RSA_OAEP()
	key, raw, kid, err := GenerateEncryptionJweJwkForEncAndAlg(&EncA256GCM, &invalidAlg)
	require.Error(t, err)
	require.Nil(t, kid)
	require.Nil(t, key)
	require.Nil(t, raw)
}

func Test_SadPath_EncryptJWE_NilKey(t *testing.T) {
	_, _, err := EncryptBytes(nil, []byte("test"))
	require.Error(t, err)
}

func Test_SadPath_DecryptJWE_NilKey(t *testing.T) {
	_, err := DecryptBytes(nil, []byte("ciphertext"))
	require.Error(t, err)
}

func Test_SadPath_DecryptJWE_InvalidCiphertext(t *testing.T) {
	kid, key, raw, err := GenerateEncryptionJweJwkForEncAndAlg(&EncA256GCM, &AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, kid)
	require.NotNil(t, key)
	require.NotNil(t, raw)

	_, err = DecryptBytes([]joseJwk.Key{key}, []byte("this-is-not-valid-ciphertext"))
	require.Error(t, err)
}
