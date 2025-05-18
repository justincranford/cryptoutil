package orm

import (
	"fmt"
	"strings"
	"testing"

	cryptoutilJose "cryptoutil/internal/common/crypto/jose"

	"github.com/stretchr/testify/require"
)

var happyPathTestCases = []struct {
	actual   KeyPoolAlgorithm
	expected string
}{
	{actual: A256GCM_A256KW, expected: cryptoutilJose.EncA256GCM.String() + "/" + cryptoutilJose.AlgA256KW.String()},
	{actual: A192GCM_A256KW, expected: cryptoutilJose.EncA192GCM.String() + "/" + cryptoutilJose.AlgA256KW.String()},
	{actual: A128GCM_A256KW, expected: cryptoutilJose.EncA128GCM.String() + "/" + cryptoutilJose.AlgA256KW.String()},
	{actual: A192GCM_A192KW, expected: cryptoutilJose.EncA192GCM.String() + "/" + cryptoutilJose.AlgA192KW.String()},
	{actual: A128GCM_A192KW, expected: cryptoutilJose.EncA128GCM.String() + "/" + cryptoutilJose.AlgA192KW.String()},
	{actual: A128GCM_A128KW, expected: cryptoutilJose.EncA128GCM.String() + "/" + cryptoutilJose.AlgA128KW.String()},

	{actual: A256GCM_A256GCMKW, expected: cryptoutilJose.EncA256GCM.String() + "/" + cryptoutilJose.AlgA256GCMKW.String()},
	{actual: A192GCM_A256GCMKW, expected: cryptoutilJose.EncA192GCM.String() + "/" + cryptoutilJose.AlgA256GCMKW.String()},
	{actual: A128GCM_A256GCMKW, expected: cryptoutilJose.EncA128GCM.String() + "/" + cryptoutilJose.AlgA256GCMKW.String()},
	{actual: A192GCM_A192GCMKW, expected: cryptoutilJose.EncA192GCM.String() + "/" + cryptoutilJose.AlgA192GCMKW.String()},
	{actual: A128GCM_A192GCMKW, expected: cryptoutilJose.EncA128GCM.String() + "/" + cryptoutilJose.AlgA192GCMKW.String()},
	{actual: A128GCM_A128GCMKW, expected: cryptoutilJose.EncA128GCM.String() + "/" + cryptoutilJose.AlgA128GCMKW.String()},

	{actual: A256GCM_dir, expected: cryptoutilJose.EncA256GCM.String() + "/" + cryptoutilJose.AlgDir.String()},
	{actual: A192GCM_dir, expected: cryptoutilJose.EncA192GCM.String() + "/" + cryptoutilJose.AlgDir.String()},
	{actual: A128GCM_dir, expected: cryptoutilJose.EncA128GCM.String() + "/" + cryptoutilJose.AlgDir.String()},

	{actual: A256CBCHS512_A256KW, expected: cryptoutilJose.EncA256CBC_HS512.String() + "/" + cryptoutilJose.AlgA256KW.String()},
	{actual: A192CBCHS384_A256KW, expected: cryptoutilJose.EncA192CBC_HS384.String() + "/" + cryptoutilJose.AlgA256KW.String()},
	{actual: A128CBCHS256_A256KW, expected: cryptoutilJose.EncA128CBC_HS256.String() + "/" + cryptoutilJose.AlgA256KW.String()},
	{actual: A192CBCHS384_A192KW, expected: cryptoutilJose.EncA192CBC_HS384.String() + "/" + cryptoutilJose.AlgA192KW.String()},
	{actual: A128CBCHS256_A192KW, expected: cryptoutilJose.EncA128CBC_HS256.String() + "/" + cryptoutilJose.AlgA192KW.String()},
	{actual: A128CBCHS256_A128KW, expected: cryptoutilJose.EncA128CBC_HS256.String() + "/" + cryptoutilJose.AlgA128KW.String()},

	{actual: A256CBCHS512_A256GCMKW, expected: cryptoutilJose.EncA256CBC_HS512.String() + "/" + cryptoutilJose.AlgA256GCMKW.String()},
	{actual: A192CBCHS384_A256GCMKW, expected: cryptoutilJose.EncA192CBC_HS384.String() + "/" + cryptoutilJose.AlgA256GCMKW.String()},
	{actual: A128CBCHS256_A256GCMKW, expected: cryptoutilJose.EncA128CBC_HS256.String() + "/" + cryptoutilJose.AlgA256GCMKW.String()},
	{actual: A192CBCHS384_A192GCMKW, expected: cryptoutilJose.EncA192CBC_HS384.String() + "/" + cryptoutilJose.AlgA192GCMKW.String()},
	{actual: A128CBCHS256_A192GCMKW, expected: cryptoutilJose.EncA128CBC_HS256.String() + "/" + cryptoutilJose.AlgA192GCMKW.String()},
	{actual: A128CBCHS256_A128GCMKW, expected: cryptoutilJose.EncA128CBC_HS256.String() + "/" + cryptoutilJose.AlgA128GCMKW.String()},

	{actual: A256CBCHS512_dir, expected: cryptoutilJose.EncA256CBC_HS512.String() + "/" + cryptoutilJose.AlgDir.String()},
	{actual: A192CBCHS384_dir, expected: cryptoutilJose.EncA192CBC_HS384.String() + "/" + cryptoutilJose.AlgDir.String()},
	{actual: A128CBCHS256_dir, expected: cryptoutilJose.EncA128CBC_HS256.String() + "/" + cryptoutilJose.AlgDir.String()},

	{actual: A256GCM_RSAOAEP512, expected: cryptoutilJose.EncA256GCM.String() + "/" + cryptoutilJose.AlgRSAOAEP512.String()},
	{actual: A192GCM_RSAOAEP512, expected: cryptoutilJose.EncA192GCM.String() + "/" + cryptoutilJose.AlgRSAOAEP512.String()},
	{actual: A128GCM_RSAOAEP512, expected: cryptoutilJose.EncA128GCM.String() + "/" + cryptoutilJose.AlgRSAOAEP512.String()},
	{actual: A192GCM_RSAOAEP384, expected: cryptoutilJose.EncA192GCM.String() + "/" + cryptoutilJose.AlgRSAOAEP384.String()},
	{actual: A128GCM_RSAOAEP384, expected: cryptoutilJose.EncA128GCM.String() + "/" + cryptoutilJose.AlgRSAOAEP384.String()},
	{actual: A128GCM_RSAOAEP256, expected: cryptoutilJose.EncA128GCM.String() + "/" + cryptoutilJose.AlgRSAOAEP256.String()},
	{actual: A128GCM_RSAOAEP, expected: cryptoutilJose.EncA128GCM.String() + "/" + cryptoutilJose.AlgRSAOAEP.String()},
	{actual: A128GCM_RSA15, expected: cryptoutilJose.EncA128GCM.String() + "/" + cryptoutilJose.AlgRSA15.String()},

	{actual: A256GCM_ECDHESA256KW, expected: cryptoutilJose.EncA256GCM.String() + "/" + cryptoutilJose.AlgECDHESA256KW.String()},
	{actual: A192GCM_ECDHESA256KW, expected: cryptoutilJose.EncA192GCM.String() + "/" + cryptoutilJose.AlgECDHESA256KW.String()},
	{actual: A128GCM_ECDHESA256KW, expected: cryptoutilJose.EncA128GCM.String() + "/" + cryptoutilJose.AlgECDHESA256KW.String()},
	{actual: A192GCM_ECDHESA192KW, expected: cryptoutilJose.EncA192GCM.String() + "/" + cryptoutilJose.AlgECDHESA192KW.String()},
	{actual: A128GCM_ECDHESA192KW, expected: cryptoutilJose.EncA128GCM.String() + "/" + cryptoutilJose.AlgECDHESA192KW.String()},
	{actual: A128GCM_ECDHESA128KW, expected: cryptoutilJose.EncA128GCM.String() + "/" + cryptoutilJose.AlgECDHESA128KW.String()},
	{actual: A128GCM_ECDHES, expected: cryptoutilJose.EncA128GCM.String() + "/" + cryptoutilJose.AlgECDHES.String()},

	{actual: A256CBC_HS512_RSAOAEP512, expected: cryptoutilJose.EncA256CBC_HS512.String() + "/" + cryptoutilJose.AlgRSAOAEP512.String()},
	{actual: A192CBC_HS384_RSAOAEP512, expected: cryptoutilJose.EncA192CBC_HS384.String() + "/" + cryptoutilJose.AlgRSAOAEP512.String()},
	{actual: A128CBC_HS256_RSAOAEP512, expected: cryptoutilJose.EncA128CBC_HS256.String() + "/" + cryptoutilJose.AlgRSAOAEP512.String()},
	{actual: A192CBC_HS384_RSAOAEP384, expected: cryptoutilJose.EncA192CBC_HS384.String() + "/" + cryptoutilJose.AlgRSAOAEP384.String()},
	{actual: A128CBC_HS256_RSAOAEP384, expected: cryptoutilJose.EncA128CBC_HS256.String() + "/" + cryptoutilJose.AlgRSAOAEP384.String()},
	{actual: A128CBC_HS256_RSAOAEP256, expected: cryptoutilJose.EncA128CBC_HS256.String() + "/" + cryptoutilJose.AlgRSAOAEP256.String()},
	{actual: A128CBC_HS256_RSAOAEP, expected: cryptoutilJose.EncA128CBC_HS256.String() + "/" + cryptoutilJose.AlgRSAOAEP.String()},
	{actual: A128CBC_HS256_RSA15, expected: cryptoutilJose.EncA128CBC_HS256.String() + "/" + cryptoutilJose.AlgRSA15.String()},

	{actual: A256CBC_HS512_ECDHESA256KW, expected: cryptoutilJose.EncA256CBC_HS512.String() + "/" + cryptoutilJose.AlgECDHESA256KW.String()},
	{actual: A192CBC_HS384_ECDHESA256KW, expected: cryptoutilJose.EncA192CBC_HS384.String() + "/" + cryptoutilJose.AlgECDHESA256KW.String()},
	{actual: A128CBC_HS256_ECDHESA256KW, expected: cryptoutilJose.EncA128CBC_HS256.String() + "/" + cryptoutilJose.AlgECDHESA256KW.String()},
	{actual: A192CBC_HS384_ECDHESA192KW, expected: cryptoutilJose.EncA192CBC_HS384.String() + "/" + cryptoutilJose.AlgECDHESA192KW.String()},
	{actual: A128CBC_HS256_ECDHESA192KW, expected: cryptoutilJose.EncA128CBC_HS256.String() + "/" + cryptoutilJose.AlgECDHESA192KW.String()},
	{actual: A128CBC_HS256_ECDHESA128KW, expected: cryptoutilJose.EncA128CBC_HS256.String() + "/" + cryptoutilJose.AlgECDHESA128KW.String()},
	{actual: A128CBC_HS256_ECDHES, expected: cryptoutilJose.EncA128CBC_HS256.String() + "/" + cryptoutilJose.AlgECDHES.String()},
}

func Test_HappyPath_Match(t *testing.T) {
	for _, testCase := range happyPathTestCases {
		actualAndExpected := fmt.Sprintf("%s  %s", string(testCase.actual), testCase.expected)
		t.Run(strings.ReplaceAll(actualAndExpected, "/", "_"), func(t *testing.T) {
			require.Equal(t, string(testCase.actual), testCase.expected)
		})
	}
}
