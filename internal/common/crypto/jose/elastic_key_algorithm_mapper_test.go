package jose

import (
	cryptoutilBusinessModel "cryptoutil/internal/common/businessmodel"

	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var happyPathTestCases = []struct {
	actual   cryptoutilBusinessModel.ElasticKeyAlgorithm
	expected string
}{
	{actual: cryptoutilBusinessModel.A256GCM_A256KW, expected: EncA256GCM.String() + "/" + AlgA256KW.String()},
	{actual: cryptoutilBusinessModel.A192GCM_A256KW, expected: EncA192GCM.String() + "/" + AlgA256KW.String()},
	{actual: cryptoutilBusinessModel.A128GCM_A256KW, expected: EncA128GCM.String() + "/" + AlgA256KW.String()},
	{actual: cryptoutilBusinessModel.A256GCM_A192KW, expected: EncA256GCM.String() + "/" + AlgA192KW.String()},
	{actual: cryptoutilBusinessModel.A192GCM_A192KW, expected: EncA192GCM.String() + "/" + AlgA192KW.String()},
	{actual: cryptoutilBusinessModel.A128GCM_A192KW, expected: EncA128GCM.String() + "/" + AlgA192KW.String()},
	{actual: cryptoutilBusinessModel.A256GCM_A128KW, expected: EncA256GCM.String() + "/" + AlgA128KW.String()},
	{actual: cryptoutilBusinessModel.A192GCM_A128KW, expected: EncA192GCM.String() + "/" + AlgA128KW.String()},
	{actual: cryptoutilBusinessModel.A128GCM_A128KW, expected: EncA128GCM.String() + "/" + AlgA128KW.String()},
	{actual: cryptoutilBusinessModel.A256GCM_A256GCMKW, expected: EncA256GCM.String() + "/" + AlgA256GCMKW.String()},
	{actual: cryptoutilBusinessModel.A192GCM_A256GCMKW, expected: EncA192GCM.String() + "/" + AlgA256GCMKW.String()},
	{actual: cryptoutilBusinessModel.A128GCM_A256GCMKW, expected: EncA128GCM.String() + "/" + AlgA256GCMKW.String()},
	{actual: cryptoutilBusinessModel.A256GCM_A192GCMKW, expected: EncA256GCM.String() + "/" + AlgA192GCMKW.String()},
	{actual: cryptoutilBusinessModel.A192GCM_A192GCMKW, expected: EncA192GCM.String() + "/" + AlgA192GCMKW.String()},
	{actual: cryptoutilBusinessModel.A128GCM_A192GCMKW, expected: EncA128GCM.String() + "/" + AlgA192GCMKW.String()},
	{actual: cryptoutilBusinessModel.A256GCM_A128GCMKW, expected: EncA256GCM.String() + "/" + AlgA128GCMKW.String()},
	{actual: cryptoutilBusinessModel.A192GCM_A128GCMKW, expected: EncA192GCM.String() + "/" + AlgA128GCMKW.String()},
	{actual: cryptoutilBusinessModel.A128GCM_A128GCMKW, expected: EncA128GCM.String() + "/" + AlgA128GCMKW.String()},
	{actual: cryptoutilBusinessModel.A256GCM_dir, expected: EncA256GCM.String() + "/" + AlgDir.String()},
	{actual: cryptoutilBusinessModel.A192GCM_dir, expected: EncA192GCM.String() + "/" + AlgDir.String()},
	{actual: cryptoutilBusinessModel.A128GCM_dir, expected: EncA128GCM.String() + "/" + AlgDir.String()},

	{actual: cryptoutilBusinessModel.A256GCM_RSAOAEP512, expected: EncA256GCM.String() + "/" + AlgRSAOAEP512.String()},
	{actual: cryptoutilBusinessModel.A192GCM_RSAOAEP512, expected: EncA192GCM.String() + "/" + AlgRSAOAEP512.String()},
	{actual: cryptoutilBusinessModel.A128GCM_RSAOAEP512, expected: EncA128GCM.String() + "/" + AlgRSAOAEP512.String()},
	{actual: cryptoutilBusinessModel.A256GCM_RSAOAEP384, expected: EncA256GCM.String() + "/" + AlgRSAOAEP384.String()},
	{actual: cryptoutilBusinessModel.A192GCM_RSAOAEP384, expected: EncA192GCM.String() + "/" + AlgRSAOAEP384.String()},
	{actual: cryptoutilBusinessModel.A128GCM_RSAOAEP384, expected: EncA128GCM.String() + "/" + AlgRSAOAEP384.String()},
	{actual: cryptoutilBusinessModel.A256GCM_RSAOAEP256, expected: EncA256GCM.String() + "/" + AlgRSAOAEP256.String()},
	{actual: cryptoutilBusinessModel.A192GCM_RSAOAEP256, expected: EncA192GCM.String() + "/" + AlgRSAOAEP256.String()},
	{actual: cryptoutilBusinessModel.A128GCM_RSAOAEP256, expected: EncA128GCM.String() + "/" + AlgRSAOAEP256.String()},
	{actual: cryptoutilBusinessModel.A256GCM_RSAOAEP, expected: EncA256GCM.String() + "/" + AlgRSAOAEP.String()},
	{actual: cryptoutilBusinessModel.A192GCM_RSAOAEP, expected: EncA192GCM.String() + "/" + AlgRSAOAEP.String()},
	{actual: cryptoutilBusinessModel.A128GCM_RSAOAEP, expected: EncA128GCM.String() + "/" + AlgRSAOAEP.String()},
	{actual: cryptoutilBusinessModel.A256GCM_RSA15, expected: EncA256GCM.String() + "/" + AlgRSA15.String()},
	{actual: cryptoutilBusinessModel.A192GCM_RSA15, expected: EncA192GCM.String() + "/" + AlgRSA15.String()},
	{actual: cryptoutilBusinessModel.A128GCM_RSA15, expected: EncA128GCM.String() + "/" + AlgRSA15.String()},

	{actual: cryptoutilBusinessModel.A256GCM_ECDHESA256KW, expected: EncA256GCM.String() + "/" + AlgECDHESA256KW.String()},
	{actual: cryptoutilBusinessModel.A192GCM_ECDHESA256KW, expected: EncA192GCM.String() + "/" + AlgECDHESA256KW.String()},
	{actual: cryptoutilBusinessModel.A128GCM_ECDHESA256KW, expected: EncA128GCM.String() + "/" + AlgECDHESA256KW.String()},
	{actual: cryptoutilBusinessModel.A256GCM_ECDHESA192KW, expected: EncA256GCM.String() + "/" + AlgECDHESA192KW.String()},
	{actual: cryptoutilBusinessModel.A192GCM_ECDHESA192KW, expected: EncA192GCM.String() + "/" + AlgECDHESA192KW.String()},
	{actual: cryptoutilBusinessModel.A128GCM_ECDHESA192KW, expected: EncA128GCM.String() + "/" + AlgECDHESA192KW.String()},
	{actual: cryptoutilBusinessModel.A256GCM_ECDHESA128KW, expected: EncA256GCM.String() + "/" + AlgECDHESA128KW.String()},
	{actual: cryptoutilBusinessModel.A192GCM_ECDHESA128KW, expected: EncA192GCM.String() + "/" + AlgECDHESA128KW.String()},
	{actual: cryptoutilBusinessModel.A128GCM_ECDHESA128KW, expected: EncA128GCM.String() + "/" + AlgECDHESA128KW.String()},
	{actual: cryptoutilBusinessModel.A256GCM_ECDHES, expected: EncA256GCM.String() + "/" + AlgECDHES.String()},
	{actual: cryptoutilBusinessModel.A192GCM_ECDHES, expected: EncA192GCM.String() + "/" + AlgECDHES.String()},
	{actual: cryptoutilBusinessModel.A128GCM_ECDHES, expected: EncA128GCM.String() + "/" + AlgECDHES.String()},

	{actual: cryptoutilBusinessModel.A256CBCHS512_A256KW, expected: EncA256CBCHS512.String() + "/" + AlgA256KW.String()},
	{actual: cryptoutilBusinessModel.A192CBCHS384_A256KW, expected: EncA192CBCHS384.String() + "/" + AlgA256KW.String()},
	{actual: cryptoutilBusinessModel.A128CBCHS256_A256KW, expected: EncA128CBCHS256.String() + "/" + AlgA256KW.String()},
	{actual: cryptoutilBusinessModel.A256CBCHS512_A192KW, expected: EncA256CBCHS512.String() + "/" + AlgA192KW.String()},
	{actual: cryptoutilBusinessModel.A192CBCHS384_A192KW, expected: EncA192CBCHS384.String() + "/" + AlgA192KW.String()},
	{actual: cryptoutilBusinessModel.A128CBCHS256_A192KW, expected: EncA128CBCHS256.String() + "/" + AlgA192KW.String()},
	{actual: cryptoutilBusinessModel.A256CBCHS512_A128KW, expected: EncA256CBCHS512.String() + "/" + AlgA128KW.String()},
	{actual: cryptoutilBusinessModel.A192CBCHS384_A128KW, expected: EncA192CBCHS384.String() + "/" + AlgA128KW.String()},
	{actual: cryptoutilBusinessModel.A128CBCHS256_A128KW, expected: EncA128CBCHS256.String() + "/" + AlgA128KW.String()},
	{actual: cryptoutilBusinessModel.A256CBCHS512_A256GCMKW, expected: EncA256CBCHS512.String() + "/" + AlgA256GCMKW.String()},
	{actual: cryptoutilBusinessModel.A192CBCHS384_A256GCMKW, expected: EncA192CBCHS384.String() + "/" + AlgA256GCMKW.String()},
	{actual: cryptoutilBusinessModel.A128CBCHS256_A256GCMKW, expected: EncA128CBCHS256.String() + "/" + AlgA256GCMKW.String()},
	{actual: cryptoutilBusinessModel.A256CBCHS512_A192GCMKW, expected: EncA256CBCHS512.String() + "/" + AlgA192GCMKW.String()},
	{actual: cryptoutilBusinessModel.A192CBCHS384_A192GCMKW, expected: EncA192CBCHS384.String() + "/" + AlgA192GCMKW.String()},
	{actual: cryptoutilBusinessModel.A128CBCHS256_A192GCMKW, expected: EncA128CBCHS256.String() + "/" + AlgA192GCMKW.String()},
	{actual: cryptoutilBusinessModel.A256CBCHS512_A128GCMKW, expected: EncA256CBCHS512.String() + "/" + AlgA128GCMKW.String()},
	{actual: cryptoutilBusinessModel.A192CBCHS384_A128GCMKW, expected: EncA192CBCHS384.String() + "/" + AlgA128GCMKW.String()},
	{actual: cryptoutilBusinessModel.A128CBCHS256_A128GCMKW, expected: EncA128CBCHS256.String() + "/" + AlgA128GCMKW.String()},
	{actual: cryptoutilBusinessModel.A256CBCHS512_dir, expected: EncA256CBCHS512.String() + "/" + AlgDir.String()},
	{actual: cryptoutilBusinessModel.A192CBCHS384_dir, expected: EncA192CBCHS384.String() + "/" + AlgDir.String()},
	{actual: cryptoutilBusinessModel.A128CBCHS256_dir, expected: EncA128CBCHS256.String() + "/" + AlgDir.String()},

	{actual: cryptoutilBusinessModel.A256CBCHS512_RSAOAEP512, expected: EncA256CBCHS512.String() + "/" + AlgRSAOAEP512.String()},
	{actual: cryptoutilBusinessModel.A192CBCHS384_RSAOAEP512, expected: EncA192CBCHS384.String() + "/" + AlgRSAOAEP512.String()},
	{actual: cryptoutilBusinessModel.A128CBCHS256_RSAOAEP512, expected: EncA128CBCHS256.String() + "/" + AlgRSAOAEP512.String()},
	{actual: cryptoutilBusinessModel.A256CBCHS512_RSAOAEP384, expected: EncA256CBCHS512.String() + "/" + AlgRSAOAEP384.String()},
	{actual: cryptoutilBusinessModel.A192CBCHS384_RSAOAEP384, expected: EncA192CBCHS384.String() + "/" + AlgRSAOAEP384.String()},
	{actual: cryptoutilBusinessModel.A128CBCHS256_RSAOAEP384, expected: EncA128CBCHS256.String() + "/" + AlgRSAOAEP384.String()},
	{actual: cryptoutilBusinessModel.A256CBCHS512_RSAOAEP256, expected: EncA256CBCHS512.String() + "/" + AlgRSAOAEP256.String()},
	{actual: cryptoutilBusinessModel.A192CBCHS384_RSAOAEP256, expected: EncA192CBCHS384.String() + "/" + AlgRSAOAEP256.String()},
	{actual: cryptoutilBusinessModel.A128CBCHS256_RSAOAEP256, expected: EncA128CBCHS256.String() + "/" + AlgRSAOAEP256.String()},
	{actual: cryptoutilBusinessModel.A256CBCHS512_RSAOAEP, expected: EncA256CBCHS512.String() + "/" + AlgRSAOAEP.String()},
	{actual: cryptoutilBusinessModel.A192CBCHS384_RSAOAEP, expected: EncA192CBCHS384.String() + "/" + AlgRSAOAEP.String()},
	{actual: cryptoutilBusinessModel.A128CBCHS256_RSAOAEP, expected: EncA128CBCHS256.String() + "/" + AlgRSAOAEP.String()},
	{actual: cryptoutilBusinessModel.A256CBCHS512_RSA15, expected: EncA256CBCHS512.String() + "/" + AlgRSA15.String()},
	{actual: cryptoutilBusinessModel.A192CBCHS384_RSA15, expected: EncA192CBCHS384.String() + "/" + AlgRSA15.String()},
	{actual: cryptoutilBusinessModel.A128CBCHS256_RSA15, expected: EncA128CBCHS256.String() + "/" + AlgRSA15.String()},

	{actual: cryptoutilBusinessModel.A256CBCHS512_ECDHESA256KW, expected: EncA256CBCHS512.String() + "/" + AlgECDHESA256KW.String()},
	{actual: cryptoutilBusinessModel.A192CBCHS384_ECDHESA256KW, expected: EncA192CBCHS384.String() + "/" + AlgECDHESA256KW.String()},
	{actual: cryptoutilBusinessModel.A128CBCHS256_ECDHESA256KW, expected: EncA128CBCHS256.String() + "/" + AlgECDHESA256KW.String()},
	{actual: cryptoutilBusinessModel.A192CBCHS384_ECDHESA192KW, expected: EncA192CBCHS384.String() + "/" + AlgECDHESA192KW.String()},
	{actual: cryptoutilBusinessModel.A128CBCHS256_ECDHESA192KW, expected: EncA128CBCHS256.String() + "/" + AlgECDHESA192KW.String()},
	{actual: cryptoutilBusinessModel.A128CBCHS256_ECDHESA128KW, expected: EncA128CBCHS256.String() + "/" + AlgECDHESA128KW.String()},
	{actual: cryptoutilBusinessModel.A256CBCHS512_ECDHES, expected: EncA256CBCHS512.String() + "/" + AlgECDHES.String()},
	{actual: cryptoutilBusinessModel.A192CBCHS384_ECDHES, expected: EncA192CBCHS384.String() + "/" + AlgECDHES.String()},
	{actual: cryptoutilBusinessModel.A128CBCHS256_ECDHES, expected: EncA128CBCHS256.String() + "/" + AlgECDHES.String()},

	{actual: cryptoutilBusinessModel.RS512, expected: AlgRS512.String()},
	{actual: cryptoutilBusinessModel.RS384, expected: AlgRS384.String()},
	{actual: cryptoutilBusinessModel.RS256, expected: AlgRS256.String()},
	{actual: cryptoutilBusinessModel.PS512, expected: AlgPS512.String()},
	{actual: cryptoutilBusinessModel.PS384, expected: AlgPS384.String()},
	{actual: cryptoutilBusinessModel.PS256, expected: AlgPS256.String()},
	{actual: cryptoutilBusinessModel.ES512, expected: AlgES512.String()},
	{actual: cryptoutilBusinessModel.ES384, expected: AlgES384.String()},
	{actual: cryptoutilBusinessModel.ES256, expected: AlgES256.String()},
	{actual: cryptoutilBusinessModel.HS512, expected: AlgHS512.String()},
	{actual: cryptoutilBusinessModel.HS384, expected: AlgHS384.String()},
	{actual: cryptoutilBusinessModel.HS256, expected: AlgHS256.String()},
	{actual: cryptoutilBusinessModel.EdDSA, expected: AlgEdDSA.String()},
}

func Test_HappyPath_Match(t *testing.T) {
	for _, testCase := range happyPathTestCases {
		actualAndExpected := fmt.Sprintf("%s  %s", string(testCase.actual), testCase.expected)
		t.Run(strings.ReplaceAll(actualAndExpected, "/", "_"), func(t *testing.T) {
			require.Equal(t, string(testCase.actual), testCase.expected)
		})
	}
}
