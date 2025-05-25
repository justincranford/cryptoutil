package jose

import (
	"context"
	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
	testJwkGenService    *JwkGenService
)

func TestMain(m *testing.M) {
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "asn1_test", false, false)
	defer testTelemetryService.Shutdown()

	var err error
	testJwkGenService, err = NewJwkGenService(testCtx, testTelemetryService)
	cryptoutilAppErr.RequireNoError(err, "failed to initialize NewJwkGenService")
	defer testJwkGenService.Shutdown()

	os.Exit(m.Run())
}

func Test_HappyPath_JwkGenService_Jwe_Jwk_EncryptDecryptBytes(t *testing.T) {
	for _, testCase := range happyPathJweTestCases {
		plaintext := fmt.Appendf(nil, "Hello world enc=%s alg=%s!", testCase.enc, testCase.alg)
		t.Run(fmt.Sprintf("%s %s", testCase.enc, testCase.alg), func(t *testing.T) {
			t.Parallel()

			actualKeyKid, cek, encodedJweJwk, err := testJwkGenService.GenerateJweJwk(testCase.enc, testCase.alg)
			require.NoError(t, err)
			require.NotNil(t, cek)
			require.NotEmpty(t, encodedJweJwk)
			require.NotEmpty(t, actualKeyKid)
			log.Printf("Generated: %s", encodedJweJwk)

			var actualJwkAlg joseJwa.KeyAlgorithm
			require.NoError(t, cek.Get(joseJwk.AlgorithmKey, &actualJwkAlg))
			require.Equal(t, *testCase.alg, actualJwkAlg)

			var actualJwkUse string
			require.NoError(t, cek.Get(joseJwk.KeyUsageKey, &actualJwkUse))
			require.Equal(t, "enc", actualJwkUse)

			var actualJwkOps joseJwk.KeyOperationList
			require.NoError(t, cek.Get(joseJwk.KeyOpsKey, &actualJwkOps))
			require.Equal(t, OpsEncDec, actualJwkOps)

			var actualJwkKty joseJwa.KeyType
			require.NoError(t, cek.Get(joseJwk.KeyTypeKey, &actualJwkKty))
			require.Equal(t, testCase.expectedType, actualJwkKty)

			jweMessage, encodedJweMessage, err := EncryptBytes([]joseJwk.Key{cek}, plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, encodedJweMessage)
			log.Printf("JWE Message: %s", string(encodedJweMessage))

			jweHeaders := jweMessage.ProtectedHeaders()
			encodedJweHeaders, err := json.Marshal(jweHeaders)
			require.NoError(t, err)
			log.Printf("JWE Message Headers: %v", string(encodedJweHeaders))

			var actualJweKid string
			require.NoError(t, jweHeaders.Get(joseJwk.KeyIDKey, &actualJweKid))
			require.NotEmpty(t, actualJweKid)
			require.Equal(t, actualKeyKid.String(), actualJweKid)

			var actualJweEnc joseJwa.ContentEncryptionAlgorithm
			require.NoError(t, jweHeaders.Get("enc", &actualJweEnc))
			// require.Equal(t, AlgCekA256GCM, actualJweEnc)

			var actualJweAlg joseJwa.KeyAlgorithm
			require.NoError(t, jweHeaders.Get(joseJwk.AlgorithmKey, &actualJweAlg))
			require.Equal(t, *testCase.alg, actualJweAlg)
			require.Equal(t, actualJwkAlg, actualJweAlg)

			decrypted, err := DecryptBytes([]joseJwk.Key{cek}, encodedJweMessage)
			require.NoError(t, err)
			require.Equal(t, plaintext, decrypted)
		})
	}
}
