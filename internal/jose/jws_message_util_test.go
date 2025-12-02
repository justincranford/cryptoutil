// Copyright (c) 2025 Justin Cranford
//
//

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

type happyPathJWSTestCase struct {
	alg          *joseJwa.SignatureAlgorithm
	expectedType joseJwa.KeyType
}

var happyPathJWSTestCases = []happyPathJWSTestCase{
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

func Test_HappyPath_NonJWKGenService_JWS_JWK_SignVerifyBytes(t *testing.T) {
	t.Parallel()

	for _, testCase := range happyPathJWSTestCases {
		plaintext := fmt.Appendf(nil, "Hello world alg=%s!", testCase.alg)
		t.Run(fmt.Sprintf("%v", testCase.alg), func(t *testing.T) {
			t.Parallel()

			jwsJWKKid, nonPublicJWSJWK, publicJWSJWK, clearNonPublicJWSJWKBytes, _, err := GenerateJWSJWKForAlg(testCase.alg)
			require.NoError(t, err)
			require.NotEmpty(t, jwsJWKKid)
			require.NotNil(t, nonPublicJWSJWK)

			require.NotEmpty(t, clearNonPublicJWSJWKBytes)
			log.Printf("Generated: %s", clearNonPublicJWSJWKBytes)

			requireJWSJWKHeaders(t, nonPublicJWSJWK, OpsSigVer, &testCase)

			if publicJWSJWK != nil {
				requireJWSJWKHeaders(t, publicJWSJWK, OpsVer, &testCase)
			}

			// isSignJWK, err := IsSignJWK(nonPublicJWSJWK)
			// require.NoError(t, err, "failed to validate nonPublicJWSJWK")
			// require.True(t, isSignJWK, "nonPublicJWSJWK must be an sign JWK")

			jwsMessage, encodedJWSMessage, err := SignBytes([]joseJwk.Key{nonPublicJWSJWK}, plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, encodedJWSMessage)
			log.Printf("JWS Message: %s", string(encodedJWSMessage))

			requireJWSMessageHeaders(t, jwsMessage, jwsJWKKid, &testCase)

			// isVerifyJWK, err := IsVerifyJWK(publicJWSJWK)
			// require.NoError(t, err, "failed to validate publicJWSJWK")
			// require.True(t, isVerifyJWK, "publicJWSJWK must be an verify JWK")
			isSymmetric, err := IsSymmetricJWK(nonPublicJWSJWK)
			require.NoError(t, err, "failed to validate nonPublicJWSJWK")

			if isSymmetric {
				verified, err := VerifyBytes([]joseJwk.Key{nonPublicJWSJWK}, encodedJWSMessage)
				require.NoError(t, err)
				require.NotNil(t, verified)
			} else {
				verified, err := VerifyBytes([]joseJwk.Key{publicJWSJWK}, encodedJWSMessage)
				require.NoError(t, err)
				require.NotNil(t, verified)
			}
		})
	}
}

func requireJWSJWKHeaders(t *testing.T, nonPublicJWSJWK joseJwk.Key, expectedJWSJWKOps joseJwk.KeyOperationList, testCase *happyPathJWSTestCase) {
	t.Helper()

	var actualJWKAlg joseJwa.KeyAlgorithm

	require.NoError(t, nonPublicJWSJWK.Get(joseJwk.AlgorithmKey, &actualJWKAlg))
	require.Equal(t, *testCase.alg, actualJWKAlg)

	var actualJWKUse string

	require.NoError(t, nonPublicJWSJWK.Get(joseJwk.KeyUsageKey, &actualJWKUse))
	require.Equal(t, joseJwk.ForSignature.String(), actualJWKUse)

	var actualJWKOps joseJwk.KeyOperationList

	require.NoError(t, nonPublicJWSJWK.Get(joseJwk.KeyOpsKey, &actualJWKOps))
	require.Equal(t, expectedJWSJWKOps, actualJWKOps)

	var actualJWKKty joseJwa.KeyType

	require.NoError(t, nonPublicJWSJWK.Get(joseJwk.KeyTypeKey, &actualJWKKty))
	require.Equal(t, testCase.expectedType, actualJWKKty)
}

func requireJWSMessageHeaders(t *testing.T, jwsMessage *joseJws.Message, jwsJWKKid *googleUuid.UUID, testCase *happyPathJWSTestCase) {
	t.Helper()

	jwsHeaders := jwsMessage.Signatures()[0].ProtectedHeaders()
	encodedJWEHeaders, err := json.Marshal(jwsHeaders)
	require.NoError(t, err)
	log.Printf("JWS Message Headers: %v", string(encodedJWEHeaders))

	var actualJWEKid string

	require.NoError(t, jwsHeaders.Get(joseJwk.KeyIDKey, &actualJWEKid))
	require.NotEmpty(t, actualJWEKid)
	require.Equal(t, jwsJWKKid.String(), actualJWEKid)

	var actualJWSAlg joseJwa.KeyAlgorithm

	require.NoError(t, jwsHeaders.Get(joseJwk.AlgorithmKey, &actualJWSAlg))
	require.Equal(t, *testCase.alg, actualJWSAlg)
}

func Test_SadPath_SignBytes_NilKey(t *testing.T) {
	_, _, err := SignBytes(nil, []byte("test"))
	require.Error(t, err)
}

func Test_SadPath_VerifyBytes_NilKey(t *testing.T) {
	_, err := VerifyBytes(nil, []byte("ciphertext"))
	require.Error(t, err)
}

func Test_SadPath_VerifyBytes_InvalidJWSMessage(t *testing.T) {
	kid, nonPublicJWSJWK, _, clearNonPublicJWSJWKBytes, _, err := GenerateJWSJWKForAlg(&AlgHS256)
	require.NoError(t, err)
	require.NotNil(t, kid)
	require.NotNil(t, nonPublicJWSJWK)
	isSigntJWK, err := IsSignJWK(nonPublicJWSJWK)
	require.NoError(t, err)
	require.True(t, isSigntJWK)
	require.NotNil(t, clearNonPublicJWSJWKBytes)

	_, err = VerifyBytes([]joseJwk.Key{nonPublicJWSJWK}, []byte("this-is-not-a-valid-jws-message"))
	require.Error(t, err)
}

func Test_SadPath_GenerateJWSJWK_UnsupportedAlg(t *testing.T) {
	kid, nonPublicJWSJWK, publicJWSJWK, clearNonPublicJWSJWKBytes, clearPublicJWSJWKBytes, err := GenerateJWSJWKForAlg(&AlgSigInvalid)
	require.Error(t, err)
	require.Equal(t, "invalid JWS JWK headers: unsupported JWS JWK alg: invalid", err.Error())
	require.Nil(t, kid)
	require.Nil(t, nonPublicJWSJWK)
	require.Nil(t, publicJWSJWK)
	require.Nil(t, clearNonPublicJWSJWKBytes)
	require.Nil(t, clearPublicJWSJWKBytes)
}

func Test_SadPath_ConcurrentGenerateJWSJWK_UnsupportedAlg(t *testing.T) {
	nonPublicJWEJWKs, publicJWEJWKs, err := GenerateJWSJWKsForTest(t, 2, &AlgSigInvalid)
	require.Error(t, err)
	require.Equal(t, "unexpected 2 errors: invalid JWS JWK headers: unsupported JWS JWK alg: invalid\ninvalid JWS JWK headers: unsupported JWS JWK alg: invalid", err.Error())
	require.Nil(t, nonPublicJWEJWKs)
	require.Nil(t, publicJWEJWKs)
}
