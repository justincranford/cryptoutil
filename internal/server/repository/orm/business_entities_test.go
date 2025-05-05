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
}

func Test_HappyPath_Match(t *testing.T) {
	for _, testCase := range happyPathTestCases {
		actualAndExpected := fmt.Sprintf("%s  %s", string(testCase.actual), testCase.expected)
		t.Run(strings.ReplaceAll(actualAndExpected, "/", "_"), func(t *testing.T) {
			require.Equal(t, string(testCase.actual), testCase.expected)
		})
	}
}
