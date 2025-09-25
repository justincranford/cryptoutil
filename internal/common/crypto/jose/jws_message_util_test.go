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

func Test_HappyPath_NonJWKGenService_Jws_JWK_SignVerifyBytes(t *testing.T) {
	for _, testCase := range happyPathJwsTestCases {
		plaintext := fmt.Appendf(nil, "Hello world alg=%s!", testCase.alg)
		t.Run(fmt.Sprintf("%v", testCase.alg), func(t *testing.T) {
			t.Parallel()

			jwsJWKKid, nonPublicJwsJWK, publicJwsJWK, clearNonPublicJwsJWKBytes, _, err := GenerateJwsJWKForAlg(testCase.alg)
			require.NoError(t, err)
			require.NotEmpty(t, jwsJWKKid)
			require.NotNil(t, nonPublicJwsJWK)

			require.NotEmpty(t, clearNonPublicJwsJWKBytes)
			log.Printf("Generated: %s", clearNonPublicJwsJWKBytes)

			requireJwsJWKHeaders(t, nonPublicJwsJWK, OpsSigVer, &testCase)
			if publicJwsJWK != nil {
				requireJwsJWKHeaders(t, publicJwsJWK, OpsVer, &testCase)
			}

			// isSignJWK, err := IsSignJWK(nonPublicJwsJWK)
			// require.NoError(t, err, "failed to validate nonPublicJwsJWK")
			// require.True(t, isSignJWK, "nonPublicJwsJWK must be an sign JWK")

			jwsMessage, encodedJwsMessage, err := SignBytes([]joseJwk.Key{nonPublicJwsJWK}, plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, encodedJwsMessage)
			log.Printf("JWS Message: %s", string(encodedJwsMessage))

			requireJwsMessageHeaders(t, jwsMessage, jwsJWKKid, &testCase)

			// isVerifyJWK, err := IsVerifyJWK(publicJwsJWK)
			// require.NoError(t, err, "failed to validate publicJwsJWK")
			// require.True(t, isVerifyJWK, "publicJwsJWK must be an verify JWK")
			isSymmetric, err := IsSymmetricJWK(nonPublicJwsJWK)
			require.NoError(t, err, "failed to validate nonPublicJwsJWK")
			if isSymmetric {
				verified, err := VerifyBytes([]joseJwk.Key{nonPublicJwsJWK}, encodedJwsMessage)
				require.NoError(t, err)
				require.NotNil(t, verified)
			} else {
				verified, err := VerifyBytes([]joseJwk.Key{publicJwsJWK}, encodedJwsMessage)
				require.NoError(t, err)
				require.NotNil(t, verified)
			}
		})
	}
}

func requireJwsJWKHeaders(t *testing.T, nonPublicJwsJWK joseJwk.Key, expectedJwsJWKOps joseJwk.KeyOperationList, testCase *happyPathJwsTestCase) {
	var actualJWKAlg joseJwa.KeyAlgorithm
	require.NoError(t, nonPublicJwsJWK.Get(joseJwk.AlgorithmKey, &actualJWKAlg))
	require.Equal(t, *testCase.alg, actualJWKAlg)

	var actualJWKUse string
	require.NoError(t, nonPublicJwsJWK.Get(joseJwk.KeyUsageKey, &actualJWKUse))
	require.Equal(t, joseJwk.ForSignature.String(), actualJWKUse)

	var actualJWKOps joseJwk.KeyOperationList
	require.NoError(t, nonPublicJwsJWK.Get(joseJwk.KeyOpsKey, &actualJWKOps))
	require.Equal(t, expectedJwsJWKOps, actualJWKOps)

	var actualJWKKty joseJwa.KeyType
	require.NoError(t, nonPublicJwsJWK.Get(joseJwk.KeyTypeKey, &actualJWKKty))
	require.Equal(t, testCase.expectedType, actualJWKKty)
}

func requireJwsMessageHeaders(t *testing.T, jwsMessage *joseJws.Message, jwsJWKKid *googleUuid.UUID, testCase *happyPathJwsTestCase) {
	jwsHeaders := jwsMessage.Signatures()[0].ProtectedHeaders()
	encodedJweHeaders, err := json.Marshal(jwsHeaders)
	require.NoError(t, err)
	log.Printf("JWS Message Headers: %v", string(encodedJweHeaders))

	var actualJweKid string
	require.NoError(t, jwsHeaders.Get(joseJwk.KeyIDKey, &actualJweKid))
	require.NotEmpty(t, actualJweKid)
	require.Equal(t, jwsJWKKid.String(), actualJweKid)

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
	kid, nonPublicJwsJWK, _, clearNonPublicJwsJWKBytes, _, err := GenerateJwsJWKForAlg(&AlgHS256)
	require.NoError(t, err)
	require.NotNil(t, kid)
	require.NotNil(t, nonPublicJwsJWK)
	isSigntJWK, err := IsSignJWK(nonPublicJwsJWK)
	require.NoError(t, err)
	require.True(t, isSigntJWK)
	require.NotNil(t, clearNonPublicJwsJWKBytes)

	_, err = VerifyBytes([]joseJwk.Key{nonPublicJwsJWK}, []byte("this-is-not-a-valid-jws-message"))
	require.Error(t, err)
}

func Test_SadPath_GenerateJwsJWK_UnsupportedAlg(t *testing.T) {
	kid, nonPublicJwsJWK, publicJwsJWK, clearNonPublicJwsJWKBytes, clearPublicJwsJWKBytes, err := GenerateJwsJWKForAlg(&AlgSigInvalid)
	require.Error(t, err)
	require.Equal(t, "invalid JWS JWK headers: unsupported JWS JWK alg: invalid", err.Error())
	require.Nil(t, kid)
	require.Nil(t, nonPublicJwsJWK)
	require.Nil(t, publicJwsJWK)
	require.Nil(t, clearNonPublicJwsJWKBytes)
	require.Nil(t, clearPublicJwsJWKBytes)
}

func Test_SadPath_ConcurrentGenerateJwsJWK_UnsupportedAlg(t *testing.T) {
	nonPublicJweJWKs, publicJweJWKs, err := GenerateJwsJWKsForTest(t, 2, &AlgSigInvalid)
	require.Error(t, err)
	require.Equal(t, "unexpected 2 errors: invalid JWS JWK headers: unsupported JWS JWK alg: invalid\ninvalid JWS JWK headers: unsupported JWS JWK alg: invalid", err.Error())
	require.Nil(t, nonPublicJweJWKs)
	require.Nil(t, publicJweJWKs)
}
