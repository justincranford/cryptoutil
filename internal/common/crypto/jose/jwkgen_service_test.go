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

func Test_HappyPath_JWKGenService_JWS_JWK_SignVerifyBytes(t *testing.T) {
	for _, testCase := range happyPathJWSTestCases {
		plaintext := fmt.Appendf(nil, "Hello world alg=%s!", testCase.alg)
		t.Run(fmt.Sprintf("%v", testCase.alg), func(t *testing.T) {
			t.Parallel()

			jwsJWKKid, nonPublicJWSJWK, publicJWSJWK, clearNonPublicJWSJWKBytes, _, err := testJWKGenService.GenerateJWSJWK(testCase.alg)
			require.NoError(t, err)
			require.NotEmpty(t, jwsJWKKid)
			require.NotNil(t, nonPublicJWSJWK)
			isSigntJWK, err := IsSignJWK(nonPublicJWSJWK)
			require.NoError(t, err)
			require.True(t, isSigntJWK)
			require.NotEmpty(t, clearNonPublicJWSJWKBytes)
			log.Printf("Generated: %s", clearNonPublicJWSJWKBytes)

			requireJWSJWKHeaders(t, nonPublicJWSJWK, OpsSigVer, &testCase)
			if publicJWSJWK != nil {
				requireJWSJWKHeaders(t, publicJWSJWK, OpsVer, &testCase)
			}

			jwsMessage, encodedJWSMessage, err := SignBytes([]joseJwk.Key{nonPublicJWSJWK}, plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, encodedJWSMessage)
			log.Printf("JWS Message: %s", string(encodedJWSMessage))

			requireJWSMessageHeaders(t, jwsMessage, jwsJWKKid, &testCase)

			var verifyJWK joseJwk.Key
			if publicJWSJWK == nil {
				verifyJWK = nonPublicJWSJWK
			} else {
				verifyJWK = publicJWSJWK
				requireJWSJWKHeaders(t, publicJWSJWK, OpsVer, &testCase)
			}

			require.NoError(t, err)
			verified, err := VerifyBytes([]joseJwk.Key{verifyJWK}, encodedJWSMessage)
			require.NoError(t, err)
			require.NotNil(t, verified)
		})
	}
}
