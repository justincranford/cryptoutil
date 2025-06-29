package jose

import (
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"

	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var happyPathTestCases = []struct {
	actual   cryptoutilOpenapiModel.ElasticKeyAlgorithm
	expected string
}{
	{actual: cryptoutilOpenapiModel.A256GCMA256KW, expected: EncA256GCM.String() + "/" + AlgA256KW.String()},
	{actual: cryptoutilOpenapiModel.A192GCMA256KW, expected: EncA192GCM.String() + "/" + AlgA256KW.String()},
	{actual: cryptoutilOpenapiModel.A128GCMA256KW, expected: EncA128GCM.String() + "/" + AlgA256KW.String()},
	{actual: cryptoutilOpenapiModel.A256GCMA192KW, expected: EncA256GCM.String() + "/" + AlgA192KW.String()},
	{actual: cryptoutilOpenapiModel.A192GCMA192KW, expected: EncA192GCM.String() + "/" + AlgA192KW.String()},
	{actual: cryptoutilOpenapiModel.A128GCMA192KW, expected: EncA128GCM.String() + "/" + AlgA192KW.String()},
	{actual: cryptoutilOpenapiModel.A256GCMA128KW, expected: EncA256GCM.String() + "/" + AlgA128KW.String()},
	{actual: cryptoutilOpenapiModel.A192GCMA128KW, expected: EncA192GCM.String() + "/" + AlgA128KW.String()},
	{actual: cryptoutilOpenapiModel.A128GCMA128KW, expected: EncA128GCM.String() + "/" + AlgA128KW.String()},
	{actual: cryptoutilOpenapiModel.A256GCMA256GCMKW, expected: EncA256GCM.String() + "/" + AlgA256GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A192GCMA256GCMKW, expected: EncA192GCM.String() + "/" + AlgA256GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A128GCMA256GCMKW, expected: EncA128GCM.String() + "/" + AlgA256GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A256GCMA192GCMKW, expected: EncA256GCM.String() + "/" + AlgA192GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A192GCMA192GCMKW, expected: EncA192GCM.String() + "/" + AlgA192GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A128GCMA192GCMKW, expected: EncA128GCM.String() + "/" + AlgA192GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A256GCMA128GCMKW, expected: EncA256GCM.String() + "/" + AlgA128GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A192GCMA128GCMKW, expected: EncA192GCM.String() + "/" + AlgA128GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A128GCMA128GCMKW, expected: EncA128GCM.String() + "/" + AlgA128GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A256GCMDir, expected: EncA256GCM.String() + "/" + AlgDir.String()},
	{actual: cryptoutilOpenapiModel.A192GCMDir, expected: EncA192GCM.String() + "/" + AlgDir.String()},
	{actual: cryptoutilOpenapiModel.A128GCMDir, expected: EncA128GCM.String() + "/" + AlgDir.String()},

	{actual: cryptoutilOpenapiModel.A256GCMRSAOAEP512, expected: EncA256GCM.String() + "/" + AlgRSAOAEP512.String()},
	{actual: cryptoutilOpenapiModel.A192GCMRSAOAEP512, expected: EncA192GCM.String() + "/" + AlgRSAOAEP512.String()},
	{actual: cryptoutilOpenapiModel.A128GCMRSAOAEP512, expected: EncA128GCM.String() + "/" + AlgRSAOAEP512.String()},
	{actual: cryptoutilOpenapiModel.A256GCMRSAOAEP384, expected: EncA256GCM.String() + "/" + AlgRSAOAEP384.String()},
	{actual: cryptoutilOpenapiModel.A192GCMRSAOAEP384, expected: EncA192GCM.String() + "/" + AlgRSAOAEP384.String()},
	{actual: cryptoutilOpenapiModel.A128GCMRSAOAEP384, expected: EncA128GCM.String() + "/" + AlgRSAOAEP384.String()},
	{actual: cryptoutilOpenapiModel.A256GCMRSAOAEP256, expected: EncA256GCM.String() + "/" + AlgRSAOAEP256.String()},
	{actual: cryptoutilOpenapiModel.A192GCMRSAOAEP256, expected: EncA192GCM.String() + "/" + AlgRSAOAEP256.String()},
	{actual: cryptoutilOpenapiModel.A128GCMRSAOAEP256, expected: EncA128GCM.String() + "/" + AlgRSAOAEP256.String()},
	{actual: cryptoutilOpenapiModel.A256GCMRSAOAEP, expected: EncA256GCM.String() + "/" + AlgRSAOAEP.String()},
	{actual: cryptoutilOpenapiModel.A192GCMRSAOAEP, expected: EncA192GCM.String() + "/" + AlgRSAOAEP.String()},
	{actual: cryptoutilOpenapiModel.A128GCMRSAOAEP, expected: EncA128GCM.String() + "/" + AlgRSAOAEP.String()},
	{actual: cryptoutilOpenapiModel.A256GCMRSA15, expected: EncA256GCM.String() + "/" + AlgRSA15.String()},
	{actual: cryptoutilOpenapiModel.A192GCMRSA15, expected: EncA192GCM.String() + "/" + AlgRSA15.String()},
	{actual: cryptoutilOpenapiModel.A128GCMRSA15, expected: EncA128GCM.String() + "/" + AlgRSA15.String()},

	{actual: cryptoutilOpenapiModel.A256GCMECDHESA256KW, expected: EncA256GCM.String() + "/" + AlgECDHESA256KW.String()},
	{actual: cryptoutilOpenapiModel.A192GCMECDHESA256KW, expected: EncA192GCM.String() + "/" + AlgECDHESA256KW.String()},
	{actual: cryptoutilOpenapiModel.A128GCMECDHESA256KW, expected: EncA128GCM.String() + "/" + AlgECDHESA256KW.String()},
	{actual: cryptoutilOpenapiModel.A256GCMECDHESA192KW, expected: EncA256GCM.String() + "/" + AlgECDHESA192KW.String()},
	{actual: cryptoutilOpenapiModel.A192GCMECDHESA192KW, expected: EncA192GCM.String() + "/" + AlgECDHESA192KW.String()},
	{actual: cryptoutilOpenapiModel.A128GCMECDHESA192KW, expected: EncA128GCM.String() + "/" + AlgECDHESA192KW.String()},
	{actual: cryptoutilOpenapiModel.A256GCMECDHESA128KW, expected: EncA256GCM.String() + "/" + AlgECDHESA128KW.String()},
	{actual: cryptoutilOpenapiModel.A192GCMECDHESA128KW, expected: EncA192GCM.String() + "/" + AlgECDHESA128KW.String()},
	{actual: cryptoutilOpenapiModel.A128GCMECDHESA128KW, expected: EncA128GCM.String() + "/" + AlgECDHESA128KW.String()},
	{actual: cryptoutilOpenapiModel.A256GCMECDHES, expected: EncA256GCM.String() + "/" + AlgECDHES.String()},
	{actual: cryptoutilOpenapiModel.A192GCMECDHES, expected: EncA192GCM.String() + "/" + AlgECDHES.String()},
	{actual: cryptoutilOpenapiModel.A128GCMECDHES, expected: EncA128GCM.String() + "/" + AlgECDHES.String()},

	{actual: cryptoutilOpenapiModel.A256CBCHS512A256KW, expected: EncA256CBCHS512.String() + "/" + AlgA256KW.String()},
	{actual: cryptoutilOpenapiModel.A192CBCHS384A256KW, expected: EncA192CBCHS384.String() + "/" + AlgA256KW.String()},
	{actual: cryptoutilOpenapiModel.A128CBCHS256A256KW, expected: EncA128CBCHS256.String() + "/" + AlgA256KW.String()},
	{actual: cryptoutilOpenapiModel.A256CBCHS512A192KW, expected: EncA256CBCHS512.String() + "/" + AlgA192KW.String()},
	{actual: cryptoutilOpenapiModel.A192CBCHS384A192KW, expected: EncA192CBCHS384.String() + "/" + AlgA192KW.String()},
	{actual: cryptoutilOpenapiModel.A128CBCHS256A192KW, expected: EncA128CBCHS256.String() + "/" + AlgA192KW.String()},
	{actual: cryptoutilOpenapiModel.A256CBCHS512A128KW, expected: EncA256CBCHS512.String() + "/" + AlgA128KW.String()},
	{actual: cryptoutilOpenapiModel.A192CBCHS384A128KW, expected: EncA192CBCHS384.String() + "/" + AlgA128KW.String()},
	{actual: cryptoutilOpenapiModel.A128CBCHS256A128KW, expected: EncA128CBCHS256.String() + "/" + AlgA128KW.String()},
	{actual: cryptoutilOpenapiModel.A256CBCHS512A256GCMKW, expected: EncA256CBCHS512.String() + "/" + AlgA256GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A192CBCHS384A256GCMKW, expected: EncA192CBCHS384.String() + "/" + AlgA256GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A128CBCHS256A256GCMKW, expected: EncA128CBCHS256.String() + "/" + AlgA256GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A256CBCHS512A192GCMKW, expected: EncA256CBCHS512.String() + "/" + AlgA192GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A192CBCHS384A192GCMKW, expected: EncA192CBCHS384.String() + "/" + AlgA192GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A128CBCHS256A192GCMKW, expected: EncA128CBCHS256.String() + "/" + AlgA192GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A256CBCHS512A128GCMKW, expected: EncA256CBCHS512.String() + "/" + AlgA128GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A192CBCHS384A128GCMKW, expected: EncA192CBCHS384.String() + "/" + AlgA128GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A128CBCHS256A128GCMKW, expected: EncA128CBCHS256.String() + "/" + AlgA128GCMKW.String()},
	{actual: cryptoutilOpenapiModel.A256CBCHS512Dir, expected: EncA256CBCHS512.String() + "/" + AlgDir.String()},
	{actual: cryptoutilOpenapiModel.A192CBCHS384Dir, expected: EncA192CBCHS384.String() + "/" + AlgDir.String()},
	{actual: cryptoutilOpenapiModel.A128CBCHS256Dir, expected: EncA128CBCHS256.String() + "/" + AlgDir.String()},

	{actual: cryptoutilOpenapiModel.A256CBCHS512RSAOAEP512, expected: EncA256CBCHS512.String() + "/" + AlgRSAOAEP512.String()},
	{actual: cryptoutilOpenapiModel.A192CBCHS384RSAOAEP512, expected: EncA192CBCHS384.String() + "/" + AlgRSAOAEP512.String()},
	{actual: cryptoutilOpenapiModel.A128CBCHS256RSAOAEP512, expected: EncA128CBCHS256.String() + "/" + AlgRSAOAEP512.String()},
	{actual: cryptoutilOpenapiModel.A256CBCHS512RSAOAEP384, expected: EncA256CBCHS512.String() + "/" + AlgRSAOAEP384.String()},
	{actual: cryptoutilOpenapiModel.A192CBCHS384RSAOAEP384, expected: EncA192CBCHS384.String() + "/" + AlgRSAOAEP384.String()},
	{actual: cryptoutilOpenapiModel.A128CBCHS256RSAOAEP384, expected: EncA128CBCHS256.String() + "/" + AlgRSAOAEP384.String()},
	{actual: cryptoutilOpenapiModel.A256CBCHS512RSAOAEP256, expected: EncA256CBCHS512.String() + "/" + AlgRSAOAEP256.String()},
	{actual: cryptoutilOpenapiModel.A192CBCHS384RSAOAEP256, expected: EncA192CBCHS384.String() + "/" + AlgRSAOAEP256.String()},
	{actual: cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256, expected: EncA128CBCHS256.String() + "/" + AlgRSAOAEP256.String()},
	{actual: cryptoutilOpenapiModel.A256CBCHS512RSAOAEP, expected: EncA256CBCHS512.String() + "/" + AlgRSAOAEP.String()},
	{actual: cryptoutilOpenapiModel.A192CBCHS384RSAOAEP, expected: EncA192CBCHS384.String() + "/" + AlgRSAOAEP.String()},
	{actual: cryptoutilOpenapiModel.A128CBCHS256RSAOAEP, expected: EncA128CBCHS256.String() + "/" + AlgRSAOAEP.String()},
	{actual: cryptoutilOpenapiModel.A256CBCHS512RSA15, expected: EncA256CBCHS512.String() + "/" + AlgRSA15.String()},
	{actual: cryptoutilOpenapiModel.A192CBCHS384RSA15, expected: EncA192CBCHS384.String() + "/" + AlgRSA15.String()},
	{actual: cryptoutilOpenapiModel.A128CBCHS256RSA15, expected: EncA128CBCHS256.String() + "/" + AlgRSA15.String()},

	{actual: cryptoutilOpenapiModel.A256CBCHS512ECDHESA256KW, expected: EncA256CBCHS512.String() + "/" + AlgECDHESA256KW.String()},
	{actual: cryptoutilOpenapiModel.A192CBCHS384ECDHESA256KW, expected: EncA192CBCHS384.String() + "/" + AlgECDHESA256KW.String()},
	{actual: cryptoutilOpenapiModel.A128CBCHS256ECDHESA256KW, expected: EncA128CBCHS256.String() + "/" + AlgECDHESA256KW.String()},
	{actual: cryptoutilOpenapiModel.A192CBCHS384ECDHESA192KW, expected: EncA192CBCHS384.String() + "/" + AlgECDHESA192KW.String()},
	{actual: cryptoutilOpenapiModel.A128CBCHS256ECDHESA192KW, expected: EncA128CBCHS256.String() + "/" + AlgECDHESA192KW.String()},
	{actual: cryptoutilOpenapiModel.A128CBCHS256ECDHESA128KW, expected: EncA128CBCHS256.String() + "/" + AlgECDHESA128KW.String()},
	{actual: cryptoutilOpenapiModel.A256CBCHS512ECDHES, expected: EncA256CBCHS512.String() + "/" + AlgECDHES.String()},
	{actual: cryptoutilOpenapiModel.A192CBCHS384ECDHES, expected: EncA192CBCHS384.String() + "/" + AlgECDHES.String()},
	{actual: cryptoutilOpenapiModel.A128CBCHS256ECDHES, expected: EncA128CBCHS256.String() + "/" + AlgECDHES.String()},

	{actual: cryptoutilOpenapiModel.RS512, expected: AlgRS512.String()},
	{actual: cryptoutilOpenapiModel.RS384, expected: AlgRS384.String()},
	{actual: cryptoutilOpenapiModel.RS256, expected: AlgRS256.String()},
	{actual: cryptoutilOpenapiModel.PS512, expected: AlgPS512.String()},
	{actual: cryptoutilOpenapiModel.PS384, expected: AlgPS384.String()},
	{actual: cryptoutilOpenapiModel.PS256, expected: AlgPS256.String()},
	{actual: cryptoutilOpenapiModel.ES512, expected: AlgES512.String()},
	{actual: cryptoutilOpenapiModel.ES384, expected: AlgES384.String()},
	{actual: cryptoutilOpenapiModel.ES256, expected: AlgES256.String()},
	{actual: cryptoutilOpenapiModel.HS512, expected: AlgHS512.String()},
	{actual: cryptoutilOpenapiModel.HS384, expected: AlgHS384.String()},
	{actual: cryptoutilOpenapiModel.HS256, expected: AlgHS256.String()},
	{actual: cryptoutilOpenapiModel.EdDSA, expected: AlgEdDSA.String()},
}

func Test_HappyPath_Match(t *testing.T) {
	for _, testCase := range happyPathTestCases {
		actualAndExpected := fmt.Sprintf("%s  %s", string(testCase.actual), testCase.expected)
		t.Run(strings.ReplaceAll(actualAndExpected, "/", "_"), func(t *testing.T) {
			require.Equal(t, string(testCase.actual), testCase.expected)
		})
	}
}
