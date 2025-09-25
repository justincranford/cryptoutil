package jose

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

func Test_HappyPath_JWKGenService_Jwe_JWK_EncryptDecryptBytes(t *testing.T) {
	for _, testCase := range happyPathJweTestCases {
		plaintext := fmt.Appendf(nil, "Hello world enc=%s alg=%s!", testCase.enc, testCase.alg)
		t.Run(fmt.Sprintf("%s %s", testCase.enc, testCase.alg), func(t *testing.T) {
			t.Parallel()

			actualKeyKid, nonPublicJweJWK, publicJweJWK, clearNonPublicJweJWKBytes, clearPublicJweJWKBytes, err := testJWKGenService.GenerateJweJWK(testCase.enc, testCase.alg)
			require.NoError(t, err)
			require.NotEmpty(t, actualKeyKid)
			require.NotNil(t, nonPublicJweJWK)
			require.NotEmpty(t, clearNonPublicJweJWKBytes)
			log.Printf("Generated:\n%s\n%s", clearNonPublicJweJWKBytes, clearPublicJweJWKBytes)

			var encryptJWK joseJwk.Key
			requireJweJWKHeaders(t, nonPublicJweJWK, OpsEncDec, &testCase)
			if publicJweJWK == nil {
				encryptJWK = nonPublicJweJWK
			} else {
				encryptJWK = publicJweJWK
				requireJweJWKHeaders(t, publicJweJWK, OpsEnc, &testCase)
			}
			isEncryptJWK, err := IsEncryptJWK(encryptJWK)
			require.NoError(t, err)
			require.True(t, isEncryptJWK)

			jweMessage, encodedJweMessage, err := EncryptBytes([]joseJwk.Key{encryptJWK}, plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, encodedJweMessage)
			log.Printf("JWE Message: %s", string(encodedJweMessage))

			jweHeaders := jweMessage.ProtectedHeaders()
			encodedJweHeaders, err := json.Marshal(jweHeaders)
			require.NoError(t, err)
			log.Printf("JWE Message Headers: %v", string(encodedJweHeaders))

			requireJweMessageHeaders(t, jweMessage, actualKeyKid, &testCase)

			decrypted, err := DecryptBytes([]joseJwk.Key{nonPublicJweJWK}, encodedJweMessage)
			require.NoError(t, err)
			require.Equal(t, plaintext, decrypted)
		})
	}
}

func Test_HappyPath_JWKGenService_Jws_JWK_SignVerifyBytes(t *testing.T) {
	for _, testCase := range happyPathJwsTestCases {
		plaintext := fmt.Appendf(nil, "Hello world alg=%s!", testCase.alg)
		t.Run(fmt.Sprintf("%v", testCase.alg), func(t *testing.T) {
			t.Parallel()

			jwsJWKKid, nonPublicJwsJWK, publicJwsJWK, clearNonPublicJwsJWKBytes, _, err := testJWKGenService.GenerateJwsJWK(testCase.alg)
			require.NoError(t, err)
			require.NotEmpty(t, jwsJWKKid)
			require.NotNil(t, nonPublicJwsJWK)
			isSigntJWK, err := IsSignJWK(nonPublicJwsJWK)
			require.NoError(t, err)
			require.True(t, isSigntJWK)
			require.NotEmpty(t, clearNonPublicJwsJWKBytes)
			log.Printf("Generated: %s", clearNonPublicJwsJWKBytes)

			requireJwsJWKHeaders(t, nonPublicJwsJWK, OpsSigVer, &testCase)
			if publicJwsJWK != nil {
				requireJwsJWKHeaders(t, publicJwsJWK, OpsVer, &testCase)
			}

			jwsMessage, encodedJwsMessage, err := SignBytes([]joseJwk.Key{nonPublicJwsJWK}, plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, encodedJwsMessage)
			log.Printf("JWS Message: %s", string(encodedJwsMessage))

			requireJwsMessageHeaders(t, jwsMessage, jwsJWKKid, &testCase)

			var verifyJWK joseJwk.Key
			if publicJwsJWK == nil {
				verifyJWK = nonPublicJwsJWK
			} else {
				verifyJWK = publicJwsJWK
				requireJwsJWKHeaders(t, publicJwsJWK, OpsVer, &testCase)
			}

			require.NoError(t, err)
			verified, err := VerifyBytes([]joseJwk.Key{verifyJWK}, encodedJwsMessage)
			require.NoError(t, err)
			require.NotNil(t, verified)
		})
	}
}
