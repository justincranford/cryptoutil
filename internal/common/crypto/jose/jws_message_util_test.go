package jose

import (
	cryptoutilUtil "cryptoutil/internal/common/util"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

var happyPathJwsTestCases = []struct {
	alg *joseJwa.SignatureAlgorithm
}{
	{alg: &AlgRS256},
	{alg: &AlgRS384},
	{alg: &AlgRS512},
	{alg: &AlgPS256},
	{alg: &AlgPS384},
	{alg: &AlgPS512},
	{alg: &AlgES256},
	{alg: &AlgES384},
	{alg: &AlgES512},
	{alg: &AlgHS256},
	{alg: &AlgHS384},
	{alg: &AlgHS512},
	{alg: &AlgEdDSA},
}

func Test_HappyPath_Jws_Jwk(t *testing.T) {
	for _, testCase := range happyPathJwsTestCases {
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
			require.True(t, cryptoutilUtil.Contains([]*joseJwa.KeyType{&KtyRsa, &KtyEC, &KtyOkp, &KtyOct}, &actualJwkKty))

			plaintext := []byte("hello, world!")
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
