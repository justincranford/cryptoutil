package jose

import (
	"cryptoutil/internal/common/businessmodel"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var happyPathTestCases = []struct {
	actual   businessmodel.ElasticKeyAlgorithm
	expected string
}{
	{actual: businessmodel.A256GCM_A256KW, expected: EncA256GCM.String() + "/" + AlgA256KW.String()},
	{actual: businessmodel.A192GCM_A256KW, expected: EncA192GCM.String() + "/" + AlgA256KW.String()},
	{actual: businessmodel.A128GCM_A256KW, expected: EncA128GCM.String() + "/" + AlgA256KW.String()},
	{actual: businessmodel.A256GCM_A192KW, expected: EncA256GCM.String() + "/" + AlgA192KW.String()},
	{actual: businessmodel.A192GCM_A192KW, expected: EncA192GCM.String() + "/" + AlgA192KW.String()},
	{actual: businessmodel.A128GCM_A192KW, expected: EncA128GCM.String() + "/" + AlgA192KW.String()},
	{actual: businessmodel.A256GCM_A128KW, expected: EncA256GCM.String() + "/" + AlgA128KW.String()},
	{actual: businessmodel.A192GCM_A128KW, expected: EncA192GCM.String() + "/" + AlgA128KW.String()},
	{actual: businessmodel.A128GCM_A128KW, expected: EncA128GCM.String() + "/" + AlgA128KW.String()},
	{actual: businessmodel.A256GCM_A256GCMKW, expected: EncA256GCM.String() + "/" + AlgA256GCMKW.String()},
	{actual: businessmodel.A192GCM_A256GCMKW, expected: EncA192GCM.String() + "/" + AlgA256GCMKW.String()},
	{actual: businessmodel.A128GCM_A256GCMKW, expected: EncA128GCM.String() + "/" + AlgA256GCMKW.String()},
	{actual: businessmodel.A256GCM_A192GCMKW, expected: EncA256GCM.String() + "/" + AlgA192GCMKW.String()},
	{actual: businessmodel.A192GCM_A192GCMKW, expected: EncA192GCM.String() + "/" + AlgA192GCMKW.String()},
	{actual: businessmodel.A128GCM_A192GCMKW, expected: EncA128GCM.String() + "/" + AlgA192GCMKW.String()},
	{actual: businessmodel.A256GCM_A128GCMKW, expected: EncA256GCM.String() + "/" + AlgA128GCMKW.String()},
	{actual: businessmodel.A192GCM_A128GCMKW, expected: EncA192GCM.String() + "/" + AlgA128GCMKW.String()},
	{actual: businessmodel.A128GCM_A128GCMKW, expected: EncA128GCM.String() + "/" + AlgA128GCMKW.String()},
	{actual: businessmodel.A256GCM_dir, expected: EncA256GCM.String() + "/" + AlgDir.String()},
	{actual: businessmodel.A192GCM_dir, expected: EncA192GCM.String() + "/" + AlgDir.String()},
	{actual: businessmodel.A128GCM_dir, expected: EncA128GCM.String() + "/" + AlgDir.String()},

	{actual: businessmodel.A256GCM_RSAOAEP512, expected: EncA256GCM.String() + "/" + AlgRSAOAEP512.String()},
	{actual: businessmodel.A192GCM_RSAOAEP512, expected: EncA192GCM.String() + "/" + AlgRSAOAEP512.String()},
	{actual: businessmodel.A128GCM_RSAOAEP512, expected: EncA128GCM.String() + "/" + AlgRSAOAEP512.String()},
	{actual: businessmodel.A256GCM_RSAOAEP384, expected: EncA256GCM.String() + "/" + AlgRSAOAEP384.String()},
	{actual: businessmodel.A192GCM_RSAOAEP384, expected: EncA192GCM.String() + "/" + AlgRSAOAEP384.String()},
	{actual: businessmodel.A128GCM_RSAOAEP384, expected: EncA128GCM.String() + "/" + AlgRSAOAEP384.String()},
	{actual: businessmodel.A256GCM_RSAOAEP256, expected: EncA256GCM.String() + "/" + AlgRSAOAEP256.String()},
	{actual: businessmodel.A192GCM_RSAOAEP256, expected: EncA192GCM.String() + "/" + AlgRSAOAEP256.String()},
	{actual: businessmodel.A128GCM_RSAOAEP256, expected: EncA128GCM.String() + "/" + AlgRSAOAEP256.String()},
	{actual: businessmodel.A256GCM_RSAOAEP, expected: EncA256GCM.String() + "/" + AlgRSAOAEP.String()},
	{actual: businessmodel.A192GCM_RSAOAEP, expected: EncA192GCM.String() + "/" + AlgRSAOAEP.String()},
	{actual: businessmodel.A128GCM_RSAOAEP, expected: EncA128GCM.String() + "/" + AlgRSAOAEP.String()},
	{actual: businessmodel.A256GCM_RSA15, expected: EncA256GCM.String() + "/" + AlgRSA15.String()},
	{actual: businessmodel.A192GCM_RSA15, expected: EncA192GCM.String() + "/" + AlgRSA15.String()},
	{actual: businessmodel.A128GCM_RSA15, expected: EncA128GCM.String() + "/" + AlgRSA15.String()},

	{actual: businessmodel.A256GCM_ECDHESA256KW, expected: EncA256GCM.String() + "/" + AlgECDHESA256KW.String()},
	{actual: businessmodel.A192GCM_ECDHESA256KW, expected: EncA192GCM.String() + "/" + AlgECDHESA256KW.String()},
	{actual: businessmodel.A128GCM_ECDHESA256KW, expected: EncA128GCM.String() + "/" + AlgECDHESA256KW.String()},
	{actual: businessmodel.A256GCM_ECDHESA192KW, expected: EncA256GCM.String() + "/" + AlgECDHESA192KW.String()},
	{actual: businessmodel.A192GCM_ECDHESA192KW, expected: EncA192GCM.String() + "/" + AlgECDHESA192KW.String()},
	{actual: businessmodel.A128GCM_ECDHESA192KW, expected: EncA128GCM.String() + "/" + AlgECDHESA192KW.String()},
	{actual: businessmodel.A256GCM_ECDHESA128KW, expected: EncA256GCM.String() + "/" + AlgECDHESA128KW.String()},
	{actual: businessmodel.A192GCM_ECDHESA128KW, expected: EncA192GCM.String() + "/" + AlgECDHESA128KW.String()},
	{actual: businessmodel.A128GCM_ECDHESA128KW, expected: EncA128GCM.String() + "/" + AlgECDHESA128KW.String()},
	{actual: businessmodel.A256GCM_ECDHES, expected: EncA256GCM.String() + "/" + AlgECDHES.String()},
	{actual: businessmodel.A192GCM_ECDHES, expected: EncA192GCM.String() + "/" + AlgECDHES.String()},
	{actual: businessmodel.A128GCM_ECDHES, expected: EncA128GCM.String() + "/" + AlgECDHES.String()},

	{actual: businessmodel.A256CBCHS512_A256KW, expected: EncA256CBC_HS512.String() + "/" + AlgA256KW.String()},
	{actual: businessmodel.A192CBCHS384_A256KW, expected: EncA192CBC_HS384.String() + "/" + AlgA256KW.String()},
	{actual: businessmodel.A128CBCHS256_A256KW, expected: EncA128CBC_HS256.String() + "/" + AlgA256KW.String()},
	{actual: businessmodel.A256CBCHS512_A192KW, expected: EncA256CBC_HS512.String() + "/" + AlgA192KW.String()},
	{actual: businessmodel.A192CBCHS384_A192KW, expected: EncA192CBC_HS384.String() + "/" + AlgA192KW.String()},
	{actual: businessmodel.A128CBCHS256_A192KW, expected: EncA128CBC_HS256.String() + "/" + AlgA192KW.String()},
	{actual: businessmodel.A256CBCHS512_A128KW, expected: EncA256CBC_HS512.String() + "/" + AlgA128KW.String()},
	{actual: businessmodel.A192CBCHS384_A128KW, expected: EncA192CBC_HS384.String() + "/" + AlgA128KW.String()},
	{actual: businessmodel.A128CBCHS256_A128KW, expected: EncA128CBC_HS256.String() + "/" + AlgA128KW.String()},
	{actual: businessmodel.A256CBCHS512_A256GCMKW, expected: EncA256CBC_HS512.String() + "/" + AlgA256GCMKW.String()},
	{actual: businessmodel.A192CBCHS384_A256GCMKW, expected: EncA192CBC_HS384.String() + "/" + AlgA256GCMKW.String()},
	{actual: businessmodel.A128CBCHS256_A256GCMKW, expected: EncA128CBC_HS256.String() + "/" + AlgA256GCMKW.String()},
	{actual: businessmodel.A256CBCHS512_A192GCMKW, expected: EncA256CBC_HS512.String() + "/" + AlgA192GCMKW.String()},
	{actual: businessmodel.A192CBCHS384_A192GCMKW, expected: EncA192CBC_HS384.String() + "/" + AlgA192GCMKW.String()},
	{actual: businessmodel.A128CBCHS256_A192GCMKW, expected: EncA128CBC_HS256.String() + "/" + AlgA192GCMKW.String()},
	{actual: businessmodel.A256CBCHS512_A128GCMKW, expected: EncA256CBC_HS512.String() + "/" + AlgA128GCMKW.String()},
	{actual: businessmodel.A192CBCHS384_A128GCMKW, expected: EncA192CBC_HS384.String() + "/" + AlgA128GCMKW.String()},
	{actual: businessmodel.A128CBCHS256_A128GCMKW, expected: EncA128CBC_HS256.String() + "/" + AlgA128GCMKW.String()},
	{actual: businessmodel.A256CBCHS512_dir, expected: EncA256CBC_HS512.String() + "/" + AlgDir.String()},
	{actual: businessmodel.A192CBCHS384_dir, expected: EncA192CBC_HS384.String() + "/" + AlgDir.String()},
	{actual: businessmodel.A128CBCHS256_dir, expected: EncA128CBC_HS256.String() + "/" + AlgDir.String()},

	{actual: businessmodel.A256CBC_HS512_RSAOAEP512, expected: EncA256CBC_HS512.String() + "/" + AlgRSAOAEP512.String()},
	{actual: businessmodel.A192CBC_HS384_RSAOAEP512, expected: EncA192CBC_HS384.String() + "/" + AlgRSAOAEP512.String()},
	{actual: businessmodel.A128CBC_HS256_RSAOAEP512, expected: EncA128CBC_HS256.String() + "/" + AlgRSAOAEP512.String()},
	{actual: businessmodel.A256CBC_HS512_RSAOAEP384, expected: EncA256CBC_HS512.String() + "/" + AlgRSAOAEP384.String()},
	{actual: businessmodel.A192CBC_HS384_RSAOAEP384, expected: EncA192CBC_HS384.String() + "/" + AlgRSAOAEP384.String()},
	{actual: businessmodel.A128CBC_HS256_RSAOAEP384, expected: EncA128CBC_HS256.String() + "/" + AlgRSAOAEP384.String()},
	{actual: businessmodel.A256CBC_HS512_RSAOAEP256, expected: EncA256CBC_HS512.String() + "/" + AlgRSAOAEP256.String()},
	{actual: businessmodel.A192CBC_HS384_RSAOAEP256, expected: EncA192CBC_HS384.String() + "/" + AlgRSAOAEP256.String()},
	{actual: businessmodel.A128CBC_HS256_RSAOAEP256, expected: EncA128CBC_HS256.String() + "/" + AlgRSAOAEP256.String()},
	{actual: businessmodel.A256CBC_HS512_RSAOAEP, expected: EncA256CBC_HS512.String() + "/" + AlgRSAOAEP.String()},
	{actual: businessmodel.A192CBC_HS384_RSAOAEP, expected: EncA192CBC_HS384.String() + "/" + AlgRSAOAEP.String()},
	{actual: businessmodel.A128CBC_HS256_RSAOAEP, expected: EncA128CBC_HS256.String() + "/" + AlgRSAOAEP.String()},
	{actual: businessmodel.A256CBC_HS512_RSA15, expected: EncA256CBC_HS512.String() + "/" + AlgRSA15.String()},
	{actual: businessmodel.A192CBC_HS384_RSA15, expected: EncA192CBC_HS384.String() + "/" + AlgRSA15.String()},
	{actual: businessmodel.A128CBC_HS256_RSA15, expected: EncA128CBC_HS256.String() + "/" + AlgRSA15.String()},

	{actual: businessmodel.A256CBC_HS512_ECDHESA256KW, expected: EncA256CBC_HS512.String() + "/" + AlgECDHESA256KW.String()},
	{actual: businessmodel.A192CBC_HS384_ECDHESA256KW, expected: EncA192CBC_HS384.String() + "/" + AlgECDHESA256KW.String()},
	{actual: businessmodel.A128CBC_HS256_ECDHESA256KW, expected: EncA128CBC_HS256.String() + "/" + AlgECDHESA256KW.String()},
	{actual: businessmodel.A192CBC_HS384_ECDHESA192KW, expected: EncA192CBC_HS384.String() + "/" + AlgECDHESA192KW.String()},
	{actual: businessmodel.A128CBC_HS256_ECDHESA192KW, expected: EncA128CBC_HS256.String() + "/" + AlgECDHESA192KW.String()},
	{actual: businessmodel.A128CBC_HS256_ECDHESA128KW, expected: EncA128CBC_HS256.String() + "/" + AlgECDHESA128KW.String()},
	{actual: businessmodel.A256CBC_HS512_ECDHES, expected: EncA256CBC_HS512.String() + "/" + AlgECDHES.String()},
	{actual: businessmodel.A192CBC_HS384_ECDHES, expected: EncA192CBC_HS384.String() + "/" + AlgECDHES.String()},
	{actual: businessmodel.A128CBC_HS256_ECDHES, expected: EncA128CBC_HS256.String() + "/" + AlgECDHES.String()},

	{actual: businessmodel.RS512, expected: AlgRS512.String()},
	{actual: businessmodel.RS384, expected: AlgRS384.String()},
	{actual: businessmodel.RS256, expected: AlgRS256.String()},
	{actual: businessmodel.PS512, expected: AlgPS512.String()},
	{actual: businessmodel.PS384, expected: AlgPS384.String()},
	{actual: businessmodel.PS256, expected: AlgPS256.String()},
	{actual: businessmodel.ES512, expected: AlgES512.String()},
	{actual: businessmodel.ES384, expected: AlgES384.String()},
	{actual: businessmodel.ES256, expected: AlgES256.String()},
	{actual: businessmodel.HS512, expected: AlgHS512.String()},
	{actual: businessmodel.HS384, expected: AlgHS384.String()},
	{actual: businessmodel.HS256, expected: AlgHS256.String()},
	{actual: businessmodel.EdDSA, expected: AlgEdDSA.String()},
}

func Test_HappyPath_Match(t *testing.T) {
	for _, testCase := range happyPathTestCases {
		actualAndExpected := fmt.Sprintf("%s  %s", string(testCase.actual), testCase.expected)
		t.Run(strings.ReplaceAll(actualAndExpected, "/", "_"), func(t *testing.T) {
			require.Equal(t, string(testCase.actual), testCase.expected)
		})
	}
}
