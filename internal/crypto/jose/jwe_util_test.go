package jose

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

var happyPathTestCases = []struct {
	alg jwa.KeyEncryptionAlgorithm
}{
	{alg: AlgA256GCMKW},
	{alg: AlgDIRECT},
}

func Test_HappyPath_Bytes(t *testing.T) {
	for _, testCase := range happyPathTestCases {
		cek, encodedAesJwk, err := GenerateAesJWK(testCase.alg)
		require.NoError(t, err)
		require.NotNil(t, cek)
		require.NotEmpty(t, encodedAesJwk)
		log.Printf("Generated: %s", encodedAesJwk)

		var actualJwkKid string
		require.NoError(t, cek.Get(jwk.KeyIDKey, &actualJwkKid))
		require.NotEmpty(t, actualJwkKid)

		var actualJwkAlg jwa.KeyAlgorithm
		require.NoError(t, cek.Get(jwk.AlgorithmKey, &actualJwkAlg))
		require.Equal(t, testCase.alg, actualJwkAlg)

		var actualJwkUse string
		require.NoError(t, cek.Get(jwk.KeyUsageKey, &actualJwkUse))
		require.Equal(t, "enc", actualJwkUse)

		var actualJwkOps jwk.KeyOperationList
		require.NoError(t, cek.Get(jwk.KeyOpsKey, &actualJwkOps))
		require.Equal(t, OpsEncDec, actualJwkOps)

		var actualJwkKty jwa.KeyType
		require.NoError(t, cek.Get(jwk.KeyTypeKey, &actualJwkKty))
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
		require.NoError(t, jweHeaders.Get(jwk.KeyIDKey, &actualJweKid))
		require.NotEmpty(t, actualJweKid)
		require.Equal(t, actualJwkKid, actualJweKid)

		var actualJweAlg jwa.KeyAlgorithm
		require.NoError(t, jweHeaders.Get(jwk.AlgorithmKey, &actualJweAlg))
		require.Equal(t, testCase.alg, actualJweAlg)
		require.Equal(t, actualJwkAlg, actualJweAlg)

		var actualJweEnc jwa.ContentEncryptionAlgorithm
		require.NoError(t, jweHeaders.Get("enc", &actualJweEnc))
		require.Equal(t, AlgA256GCM, actualJweEnc)

		decrypted, err := DecryptBytes([]joseJwk.Key{cek}, encodedJweMessage)
		require.NoError(t, err)
		require.Equal(t, plaintext, decrypted)
	}
}

func Test_HappyPath_Key(t *testing.T) {
	for _, testCase := range happyPathTestCases {
		kek, encodedKek, _ := GenerateAesJWK(testCase.alg)
		log.Printf("KEK: %s", string(encodedKek))

		originalKey, encodedKey, _ := GenerateAesJWK(testCase.alg)
		log.Printf("Original Key: %s", string(encodedKey))

		jweMessage, encodedJweMessage, err := EncryptKey([]joseJwk.Key{kek}, originalKey)
		require.NoError(t, err)
		jsonHeaders, err := JSONHeadersString(jweMessage)
		require.NoError(t, err)
		log.Printf("JWE Message Headers: %s", jsonHeaders)

		decryptedKey, err := DecryptKey([]joseJwk.Key{kek}, encodedJweMessage)
		require.NoError(t, err)

		decryptedEncodedKey, err := json.Marshal(decryptedKey)
		require.NoError(t, err)
		log.Printf("Decrypted Key: %s", string(decryptedEncodedKey))

		require.Equal(t, originalKey, decryptedKey)
	}
}

func Test_SadPath_GenerateAesJWK_UnsupportedAlg(t *testing.T) {
	invalidAlg := jwa.RSA_OAEP()
	key, raw, err := GenerateAesJWK(invalidAlg)
	require.Error(t, err)
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
	key, _, err := GenerateAesJWK(AlgA256GCMKW)
	require.NoError(t, err)

	_, err = DecryptBytes([]joseJwk.Key{key}, []byte("this-is-not-valid-ciphertext"))
	require.Error(t, err)
}
