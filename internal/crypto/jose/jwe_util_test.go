package jose

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

var happyPathTestCases = []struct {
	alg jwa.KeyEncryptionAlgorithm
}{
	{alg: AlgA256GCMKW},
	{alg: AlgDIRECT},
}

func Test_HappyPath(t *testing.T) {
	for _, testCase := range happyPathTestCases {
		aesJwk, encodedAesJwk, err := GenerateAesJWK(testCase.alg)
		require.NoError(t, err)
		require.NotNil(t, aesJwk)
		require.NotEmpty(t, encodedAesJwk)
		log.Printf("Generated: %s", encodedAesJwk)

		var actualJwkKid string
		require.NoError(t, aesJwk.Get(jwk.KeyIDKey, &actualJwkKid))
		require.NotEmpty(t, actualJwkKid)

		var actualJwkAlg jwa.KeyAlgorithm
		require.NoError(t, aesJwk.Get(jwk.AlgorithmKey, &actualJwkAlg))
		require.Equal(t, testCase.alg, actualJwkAlg)

		var actualJwkUse string
		require.NoError(t, aesJwk.Get(jwk.KeyUsageKey, &actualJwkUse))
		require.Equal(t, "enc", actualJwkUse)

		var actualJwkOps jwk.KeyOperationList
		require.NoError(t, aesJwk.Get(jwk.KeyOpsKey, &actualJwkOps))
		require.Equal(t, OpsEncDec, actualJwkOps)

		var actualJwkKty jwa.KeyType
		require.NoError(t, aesJwk.Get(jwk.KeyTypeKey, &actualJwkKty))
		require.Equal(t, KtyOct, actualJwkKty)

		plaintext := []byte("hello, world!")
		jweMessage, encodedJweMessage, err := Encrypt(aesJwk, plaintext)
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

		decrypted, err := Decrypt(aesJwk, encodedJweMessage)
		require.NoError(t, err)
		require.Equal(t, plaintext, decrypted)
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
	_, _, err := Encrypt(nil, []byte("test"))
	require.Error(t, err)
}

func Test_SadPath_DecryptJWE_NilKey(t *testing.T) {
	_, err := Decrypt(nil, []byte("ciphertext"))
	require.Error(t, err)
}

func Test_SadPath_DecryptJWE_InvalidCiphertext(t *testing.T) {
	key, _, err := GenerateAesJWK(AlgA256GCMKW)
	require.NoError(t, err)

	_, err = Decrypt(key, []byte("this-is-not-valid-ciphertext"))
	require.Error(t, err)
}
