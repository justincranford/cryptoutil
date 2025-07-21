package jose

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

var (
	testSettings         = cryptoutilConfig.Default()
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
	testJwkGenService    *JwkGenService
)

func TestMain(m *testing.M) {
	var rc int
	func() {
		testSettings.DevMode = true
		testSettings.Migrations = true
		testSettings.OTLPScope = "jwkgen_service_test"

		testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		var err error
		testJwkGenService, err = NewJwkGenService(testCtx, testTelemetryService)
		cryptoutilAppErr.RequireNoError(err, "failed to initialize NewJwkGenService")
		defer testJwkGenService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func Test_HappyPath_JwkGenService_Jwe_Jwk_EncryptDecryptBytes(t *testing.T) {
	for _, testCase := range happyPathJweTestCases {
		plaintext := fmt.Appendf(nil, "Hello world enc=%s alg=%s!", testCase.enc, testCase.alg)
		t.Run(fmt.Sprintf("%s %s", testCase.enc, testCase.alg), func(t *testing.T) {
			t.Parallel()

			actualKeyKid, nonPublicJweJwk, publicJweJwk, clearNonPublicJweJwkBytes, clearPublicJweJwkBytes, err := testJwkGenService.GenerateJweJwk(testCase.enc, testCase.alg)
			require.NoError(t, err)
			require.NotEmpty(t, actualKeyKid)
			require.NotNil(t, nonPublicJweJwk)
			// TODO Util to check AsymmetricJWK vs SymmetricJWK
			// require.NotNil(t, publicJweJwk)
			require.NotEmpty(t, clearNonPublicJweJwkBytes)
			// require.NotEmpty(t, encodedPublicJweJwk)
			log.Printf("Generated:\n%s\n%s", clearNonPublicJweJwkBytes, clearPublicJweJwkBytes)

			requireJweJwkHeaders(t, nonPublicJweJwk, OpsEncDec, &testCase)
			if publicJweJwk != nil {
				requireJweJwkHeaders(t, publicJweJwk, OpsEnc, &testCase)
			}

			jweMessage, encodedJweMessage, err := EncryptBytes([]joseJwk.Key{nonPublicJweJwk}, plaintext)
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
			// TODO Util to check AsymmetricJWK vs SymmetricJWK
			// require.NotNil(t, publicJwsJwk)
			require.NotEmpty(t, clearNonPublicJwsJwkBytes)
			// require.NotEmpty(t, encodedPublicJwsJwk)
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

			verified, err := VerifyBytes([]joseJwk.Key{nonPublicJwsJwk}, encodedJwsMessage)
			require.NoError(t, err)
			require.NotNil(t, verified)
		})
	}
}
