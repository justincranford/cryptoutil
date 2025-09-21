package jose

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	googleUuid "github.com/google/uuid"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"

	"github.com/stretchr/testify/require"
)

type happyPathJwsTestCase struct {
	alg          *joseJwa.SignatureAlgorithm
	expectedType joseJwa.KeyType
}

var happyPathJwsTestCases = []happyPathJwsTestCase{
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

			jwsJwkKid, nonPublicJwsJwk, publicJwsJwk, clearNonPublicJwsJwkBytes, _, err := GenerateJwsJwkForAlg(testCase.alg)
			require.NoError(t, err)
			require.NotEmpty(t, jwsJwkKid)
			require.NotNil(t, nonPublicJwsJwk)

			require.NotEmpty(t, clearNonPublicJwsJwkBytes)
			log.Printf("Generated: %s", clearNonPublicJwsJwkBytes)

			requireJwsJwkHeaders(t, nonPublicJwsJwk, OpsSigVer, &testCase)
			if publicJwsJwk != nil {
				requireJwsJwkHeaders(t, publicJwsJwk, OpsVer, &testCase)
			}

			// isSignJwk, err := IsSignJwk(nonPublicJwsJwk)
			// require.NoError(t, err, "failed to validate nonPublicJwsJwk")
			// require.True(t, isSignJwk, "nonPublicJwsJwk must be an sign JWK")

			jwsMessage, encodedJwsMessage, err := SignBytes([]joseJwk.Key{nonPublicJwsJwk}, plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, encodedJwsMessage)
			log.Printf("JWS Message: %s", string(encodedJwsMessage))

			requireJwsMessageHeaders(t, jwsMessage, jwsJwkKid, &testCase)

			// isVerifyJwk, err := IsVerifyJwk(publicJwsJwk)
			// require.NoError(t, err, "failed to validate publicJwsJwk")
			// require.True(t, isVerifyJwk, "publicJwsJwk must be an verify JWK")
			isSymmetric, err := IsSymmetricJwk(nonPublicJwsJwk)
			require.NoError(t, err, "failed to validate nonPublicJwsJwk")
			if isSymmetric {
				verified, err := VerifyBytes([]joseJwk.Key{nonPublicJwsJwk}, encodedJwsMessage)
				require.NoError(t, err)
				require.NotNil(t, verified)
			} else {
				verified, err := VerifyBytes([]joseJwk.Key{publicJwsJwk}, encodedJwsMessage)
				require.NoError(t, err)
				require.NotNil(t, verified)
			}
		})
	}
}

func requireJwsJwkHeaders(t *testing.T, nonPublicJwsJwk joseJwk.Key, expectedJwsJwkOps joseJwk.KeyOperationList, testCase *happyPathJwsTestCase) {
	var actualJwkAlg joseJwa.KeyAlgorithm
	require.NoError(t, nonPublicJwsJwk.Get(joseJwk.AlgorithmKey, &actualJwkAlg))
	require.Equal(t, *testCase.alg, actualJwkAlg)

	var actualJwkUse string
	require.NoError(t, nonPublicJwsJwk.Get(joseJwk.KeyUsageKey, &actualJwkUse))
	require.Equal(t, joseJwk.ForSignature.String(), actualJwkUse)

	var actualJwkOps joseJwk.KeyOperationList
	require.NoError(t, nonPublicJwsJwk.Get(joseJwk.KeyOpsKey, &actualJwkOps))
	require.Equal(t, expectedJwsJwkOps, actualJwkOps)

	var actualJwkKty joseJwa.KeyType
	require.NoError(t, nonPublicJwsJwk.Get(joseJwk.KeyTypeKey, &actualJwkKty))
	require.Equal(t, testCase.expectedType, actualJwkKty)
}

func requireJwsMessageHeaders(t *testing.T, jwsMessage *joseJws.Message, jwsJwkKid *googleUuid.UUID, testCase *happyPathJwsTestCase) {
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
}

func Test_SadPath_SignBytes_NilKey(t *testing.T) {
	_, _, err := SignBytes(nil, []byte("test"))
	require.Error(t, err)
}

func Test_SadPath_VerifyBytes_NilKey(t *testing.T) {
	_, err := VerifyBytes(nil, []byte("ciphertext"))
	require.Error(t, err)
}

func Test_SadPath_VerifyBytes_InvalidJwsMessage(t *testing.T) {
	kid, nonPublicJwsJwk, _, clearNonPublicJwsJwkBytes, _, err := GenerateJwsJwkForAlg(&AlgHS256)
	require.NoError(t, err)
	require.NotNil(t, kid)
	require.NotNil(t, nonPublicJwsJwk)
	isSigntJwk, err := IsSignJwk(nonPublicJwsJwk)
	require.NoError(t, err)
	require.True(t, isSigntJwk)
	require.NotNil(t, clearNonPublicJwsJwkBytes)

	_, err = VerifyBytes([]joseJwk.Key{nonPublicJwsJwk}, []byte("this-is-not-a-valid-jws-message"))
	require.Error(t, err)
}

func Test_SadPath_GenerateJwsJwk_UnsupportedAlg(t *testing.T) {
	kid, nonPublicJwsJwk, publicJwsJwk, clearNonPublicJwsJwkBytes, clearPublicJwsJwkBytes, err := GenerateJwsJwkForAlg(&AlgSigInvalid)
	require.Error(t, err)
	require.Equal(t, "invalid JWS JWK headers: unsupported JWS JWK alg: invalid", err.Error())
	require.Nil(t, kid)
	require.Nil(t, nonPublicJwsJwk)
	require.Nil(t, publicJwsJwk)
	require.Nil(t, clearNonPublicJwsJwkBytes)
	require.Nil(t, clearPublicJwsJwkBytes)
}

func Test_SadPath_ConcurrentGenerateJwsJwk_UnsupportedAlg(t *testing.T) {
	nonPublicJweJwks, publicJweJwks, err := GenerateJwsJwksForTest(t, 2, &AlgSigInvalid)
	require.Error(t, err)
	require.Equal(t, "unexpected 2 errors: invalid JWS JWK headers: unsupported JWS JWK alg: invalid\ninvalid JWS JWK headers: unsupported JWS JWK alg: invalid", err.Error())
	require.Nil(t, nonPublicJweJwks)
	require.Nil(t, publicJweJwks)
}
