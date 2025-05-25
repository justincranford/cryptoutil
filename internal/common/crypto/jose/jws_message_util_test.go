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

var happyPathJwsTestCases = []struct {
	alg          *joseJwa.SignatureAlgorithm
	expectedType joseJwa.KeyType
}{
	{alg: &AlgRS256, expectedType: KtyRSA}, // RSA 1.5 & SHA-256
	{alg: &AlgRS384, expectedType: KtyRSA}, // RSA 1.5 & SHA-384
	{alg: &AlgRS512, expectedType: KtyRSA}, // RSA 1.5 & SHA-512
	{alg: &AlgPS256, expectedType: KtyRSA}, // RSA 2.0 & SHA-256
	{alg: &AlgPS384, expectedType: KtyRSA}, // RSA 2.0 & SHA-384
	{alg: &AlgPS512, expectedType: KtyRSA}, // RSA 2.0 & SHA-512
	{alg: &AlgES256, expectedType: KtyEC},  // EC P-256 & SHA-256
	{alg: &AlgES384, expectedType: KtyEC},  // EC P-394 & SHA-384
	{alg: &AlgES512, expectedType: KtyEC},  // EC P-521 & SHA-512
	{alg: &AlgHS256, expectedType: KtyOCT}, // HMAC with SHA-256 & SHA-512
	{alg: &AlgHS384, expectedType: KtyOCT}, // HMAC with SHA-384 & SHA-512
	{alg: &AlgHS512, expectedType: KtyOCT}, // HMAC with SHA-512 & SHA-512
	{alg: &AlgEdDSA, expectedType: KtyOKP}, // ED25519 & SHA-256
}

func Test_HappyPath_NonJwkGenService_Jws_Jwk_SignVerifyBytes(t *testing.T) {
	for _, testCase := range happyPathJwsTestCases {
		plaintext := fmt.Appendf(nil, "Hello world alg=%s!", testCase.alg)
		t.Run(fmt.Sprintf("%v", testCase.alg), func(t *testing.T) {
			t.Parallel()

			jwsJwkKid, jwsJwk, encodedJwsJwk, err := GenerateJwsJwkForAlg(testCase.alg)
			require.NoError(t, err)
			require.NotEmpty(t, jwsJwkKid)
			require.NotNil(t, jwsJwk)
			require.NotEmpty(t, encodedJwsJwk)
			log.Printf("Generated: %s", encodedJwsJwk)

			var actualJwkAlg joseJwa.KeyAlgorithm
			require.NoError(t, jwsJwk.Get(joseJwk.AlgorithmKey, &actualJwkAlg))
			require.Equal(t, *testCase.alg, actualJwkAlg)

			var actualJwkUse string
			require.NoError(t, jwsJwk.Get(joseJwk.KeyUsageKey, &actualJwkUse))
			require.Equal(t, "sig", actualJwkUse)

			var actualJwkOps joseJwk.KeyOperationList
			require.NoError(t, jwsJwk.Get(joseJwk.KeyOpsKey, &actualJwkOps))
			require.Equal(t, OpsSigVer, actualJwkOps)

			var actualJwkKty joseJwa.KeyType
			require.NoError(t, jwsJwk.Get(joseJwk.KeyTypeKey, &actualJwkKty))
			require.Equal(t, testCase.expectedType, actualJwkKty)

			jwsMessage, encodedJwsMessage, err := SignBytes([]joseJwk.Key{jwsJwk}, plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, encodedJwsMessage)
			log.Printf("JWS Message: %s", string(encodedJwsMessage))

			jwsHeaders := jwsMessage.Signatures()[0].ProtectedHeaders()
			encodedJweHeaders, err := json.Marshal(jwsHeaders)
			require.NoError(t, err)
			log.Printf("JWS Message Headers: %v", string(encodedJweHeaders))

			var actualJweKid string
			require.NoError(t, jwsHeaders.Get(joseJwk.KeyIDKey, &actualJweKid))
			require.NotEmpty(t, actualJweKid)
			require.Equal(t, jwsJwkKid.String(), actualJweKid)

			var actualJwsAlg joseJwa.KeyAlgorithm
			require.NoError(t, jwsHeaders.Get(joseJwk.AlgorithmKey, &actualJwsAlg))
			require.Equal(t, *testCase.alg, actualJwsAlg)
			require.Equal(t, actualJwkAlg, actualJwsAlg)

			verified, err := VerifyBytes([]joseJwk.Key{jwsJwk}, encodedJwsMessage)
			require.NoError(t, err)
			require.NotNil(t, verified)
		})
	}
}

func Test_SadPath_SignBytes_NilKey(t *testing.T) {
	_, _, err := SignBytes(nil, []byte("test"))
	require.Error(t, err)
}

func Test_SadPath_VerifyBytes_NilKey(t *testing.T) {
	_, err := VerifyBytes(nil, []byte("ciphertext"))
	require.Error(t, err)
}

func Test_SadPath_GenerateJwsJwk_UnsupportedAlg(t *testing.T) {
	key, raw, kid, err := GenerateJwsJwkForAlg(&AlgSigInvalid)
	require.Error(t, err)
	require.Equal(t, "invalid JWS JWK headers: unsupported JWS JWK alg: invalid", err.Error())
	require.Nil(t, kid)
	require.Nil(t, key)
	require.Nil(t, raw)
}

func Test_SadPath_ConcurrentGenerateJwsJwk_UnsupportedAlg(t *testing.T) {
	jwks, err := GenerateJwsJwksForTest(t, 2, &AlgSigInvalid)
	require.Error(t, err)
	require.Equal(t, "unexpected 2 errors: invalid JWS JWK headers: unsupported JWS JWK alg: invalid\ninvalid JWS JWK headers: unsupported JWS JWK alg: invalid", err.Error())
	require.Nil(t, jwks)
}
