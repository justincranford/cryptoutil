package jose

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

func Test_HappyPath_JwkGenService_Jwe_Jwk_EncryptDecryptBytes(t *testing.T) {
	for _, testCase := range happyPathJweTestCases {
		plaintext := fmt.Appendf(nil, "Hello world enc=%s alg=%s!", testCase.enc, testCase.alg)
		t.Run(fmt.Sprintf("%s %s", testCase.enc, testCase.alg), func(t *testing.T) {
			t.Parallel()

			actualKeyKid, nonPublicJweJwk, publicJweJwk, clearNonPublicJweJwkBytes, clearPublicJweJwkBytes, err := testJwkGenService.GenerateJweJwk(testCase.enc, testCase.alg)
			require.NoError(t, err)
			require.NotEmpty(t, actualKeyKid)
			require.NotNil(t, nonPublicJweJwk)
			isEncryptJwk, err := IsEncryptJwk(nonPublicJweJwk)
			require.NoError(t, err)
			require.True(t, isEncryptJwk)
			require.NotEmpty(t, clearNonPublicJweJwkBytes)
			log.Printf("Generated:\n%s\n%s", clearNonPublicJweJwkBytes, clearPublicJweJwkBytes)

			var encryptJWK joseJwk.Key
			requireJweJwkHeaders(t, nonPublicJweJwk, OpsEncDec, &testCase)
			if publicJweJwk == nil {
				encryptJWK = nonPublicJweJwk
			} else {
				encryptJWK = publicJweJwk
				requireJweJwkHeaders(t, publicJweJwk, OpsEnc, &testCase)
			}

			jweMessage, encodedJweMessage, err := EncryptBytes([]joseJwk.Key{encryptJWK}, plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, encodedJweMessage)
			log.Printf("JWE Message: %s", string(encodedJweMessage))

			jweHeaders := jweMessage.ProtectedHeaders()
			encodedJweHeaders, err := json.Marshal(jweHeaders)
			require.NoError(t, err)
			log.Printf("JWE Message Headers: %v", string(encodedJweHeaders))

			requireJweMessageHeaders(t, jweMessage, actualKeyKid, &testCase)

			decrypted, err := DecryptBytes([]joseJwk.Key{nonPublicJweJwk}, encodedJweMessage)
			require.NoError(t, err)
			require.Equal(t, plaintext, decrypted)
		})
	}
}

func Test_HappyPath_JwkGenService_Jws_Jwk_SignVerifyBytes(t *testing.T) {
	for _, testCase := range happyPathJwsTestCases {
		plaintext := fmt.Appendf(nil, "Hello world alg=%s!", testCase.alg)
		t.Run(fmt.Sprintf("%v", testCase.alg), func(t *testing.T) {
			t.Parallel()

			jwsJwkKid, nonPublicJwsJwk, publicJwsJwk, clearNonPublicJwsJwkBytes, _, err := testJwkGenService.GenerateJwsJwk(testCase.alg)
			require.NoError(t, err)
			require.NotEmpty(t, jwsJwkKid)
			require.NotNil(t, nonPublicJwsJwk)
			isSigntJwk, err := IsSignJwk(nonPublicJwsJwk)
			require.NoError(t, err)
			require.True(t, isSigntJwk)
			require.NotEmpty(t, clearNonPublicJwsJwkBytes)
			log.Printf("Generated: %s", clearNonPublicJwsJwkBytes)

			requireJwsJwkHeaders(t, nonPublicJwsJwk, OpsSigVer, &testCase)
			if publicJwsJwk != nil {
				requireJwsJwkHeaders(t, publicJwsJwk, OpsVer, &testCase)
			}

			jwsMessage, encodedJwsMessage, err := SignBytes([]joseJwk.Key{nonPublicJwsJwk}, plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, encodedJwsMessage)
			log.Printf("JWS Message: %s", string(encodedJwsMessage))

			requireJwsMessageHeaders(t, jwsMessage, jwsJwkKid, &testCase)

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
