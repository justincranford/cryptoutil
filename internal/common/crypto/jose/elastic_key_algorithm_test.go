package jose

import (
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"

	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestCase struct {
	actualElasticKeyAlgorithm cryptoutilOpenapiModel.ElasticKeyAlgorithm
	expectedIsSymmetric       bool
	expectedIsAsymmetric      bool
}

var happyPathTestCases2 = []TestCase{
	{cryptoutilOpenapiModel.A256GCMA256KW, true, false}, {cryptoutilOpenapiModel.A192GCMA256KW, true, false}, {cryptoutilOpenapiModel.A128GCMA256KW, true, false},
	{cryptoutilOpenapiModel.A256GCMA192KW, true, false}, {cryptoutilOpenapiModel.A192GCMA192KW, true, false}, {cryptoutilOpenapiModel.A128GCMA192KW, true, false},
	{cryptoutilOpenapiModel.A256GCMA128KW, true, false}, {cryptoutilOpenapiModel.A192GCMA128KW, true, false}, {cryptoutilOpenapiModel.A128GCMA128KW, true, false},
	{cryptoutilOpenapiModel.A256GCMA256GCMKW, true, false}, {cryptoutilOpenapiModel.A192GCMA256GCMKW, true, false}, {cryptoutilOpenapiModel.A128GCMA256GCMKW, true, false},
	{cryptoutilOpenapiModel.A256GCMA192GCMKW, true, false}, {cryptoutilOpenapiModel.A192GCMA192GCMKW, true, false}, {cryptoutilOpenapiModel.A128GCMA192GCMKW, true, false},
	{cryptoutilOpenapiModel.A256GCMA128GCMKW, true, false}, {cryptoutilOpenapiModel.A192GCMA128GCMKW, true, false}, {cryptoutilOpenapiModel.A128GCMA128GCMKW, true, false},
	{cryptoutilOpenapiModel.A256GCMDir, true, false}, {cryptoutilOpenapiModel.A192GCMDir, true, false}, {cryptoutilOpenapiModel.A128GCMDir, true, false},

	{cryptoutilOpenapiModel.A256GCMRSAOAEP512, false, true}, {cryptoutilOpenapiModel.A192GCMRSAOAEP512, false, true}, {cryptoutilOpenapiModel.A128GCMRSAOAEP512, false, true},
	{cryptoutilOpenapiModel.A256GCMRSAOAEP384, false, true}, {cryptoutilOpenapiModel.A192GCMRSAOAEP384, false, true}, {cryptoutilOpenapiModel.A128GCMRSAOAEP384, false, true},
	{cryptoutilOpenapiModel.A256GCMRSAOAEP256, false, true}, {cryptoutilOpenapiModel.A192GCMRSAOAEP256, false, true}, {cryptoutilOpenapiModel.A128GCMRSAOAEP256, false, true},
	{cryptoutilOpenapiModel.A256GCMRSAOAEP, false, true}, {cryptoutilOpenapiModel.A192GCMRSAOAEP, false, true}, {cryptoutilOpenapiModel.A128GCMRSAOAEP, false, true},
	{cryptoutilOpenapiModel.A256GCMRSA15, false, true}, {cryptoutilOpenapiModel.A192GCMRSA15, false, true}, {cryptoutilOpenapiModel.A128GCMRSA15, false, true},

	{cryptoutilOpenapiModel.A256GCMECDHESA256KW, false, true}, {cryptoutilOpenapiModel.A192GCMECDHESA256KW, false, true}, {cryptoutilOpenapiModel.A128GCMECDHESA256KW, false, true},
	{cryptoutilOpenapiModel.A256GCMECDHESA192KW, false, true}, {cryptoutilOpenapiModel.A192GCMECDHESA192KW, false, true}, {cryptoutilOpenapiModel.A128GCMECDHESA192KW, false, true},
	{cryptoutilOpenapiModel.A256GCMECDHESA128KW, false, true}, {cryptoutilOpenapiModel.A192GCMECDHESA128KW, false, true}, {cryptoutilOpenapiModel.A128GCMECDHESA128KW, false, true},
	{cryptoutilOpenapiModel.A256GCMECDHES, false, true}, {cryptoutilOpenapiModel.A192GCMECDHES, false, true}, {cryptoutilOpenapiModel.A128GCMECDHES, false, true},

	{cryptoutilOpenapiModel.A256CBCHS512A256KW, true, false}, {cryptoutilOpenapiModel.A192CBCHS384A256KW, true, false}, {cryptoutilOpenapiModel.A128CBCHS256A256KW, true, false},
	{cryptoutilOpenapiModel.A256CBCHS512A192KW, true, false}, {cryptoutilOpenapiModel.A192CBCHS384A192KW, true, false}, {cryptoutilOpenapiModel.A128CBCHS256A192KW, true, false},
	{cryptoutilOpenapiModel.A256CBCHS512A128KW, true, false}, {cryptoutilOpenapiModel.A192CBCHS384A128KW, true, false}, {cryptoutilOpenapiModel.A128CBCHS256A128KW, true, false},
	{cryptoutilOpenapiModel.A256CBCHS512A256GCMKW, true, false}, {cryptoutilOpenapiModel.A192CBCHS384A256GCMKW, true, false}, {cryptoutilOpenapiModel.A128CBCHS256A256GCMKW, true, false},
	{cryptoutilOpenapiModel.A256CBCHS512A192GCMKW, true, false}, {cryptoutilOpenapiModel.A192CBCHS384A192GCMKW, true, false}, {cryptoutilOpenapiModel.A128CBCHS256A192GCMKW, true, false},
	{cryptoutilOpenapiModel.A256CBCHS512A128GCMKW, true, false}, {cryptoutilOpenapiModel.A192CBCHS384A128GCMKW, true, false}, {cryptoutilOpenapiModel.A128CBCHS256A128GCMKW, true, false},
	{cryptoutilOpenapiModel.A256CBCHS512Dir, true, false}, {cryptoutilOpenapiModel.A192CBCHS384Dir, true, false}, {cryptoutilOpenapiModel.A128CBCHS256Dir, true, false},

	{cryptoutilOpenapiModel.A256CBCHS512RSAOAEP512, false, true}, {cryptoutilOpenapiModel.A192CBCHS384RSAOAEP512, false, true}, {cryptoutilOpenapiModel.A128CBCHS256RSAOAEP512, false, true},
	{cryptoutilOpenapiModel.A256CBCHS512RSAOAEP384, false, true}, {cryptoutilOpenapiModel.A192CBCHS384RSAOAEP384, false, true}, {cryptoutilOpenapiModel.A128CBCHS256RSAOAEP384, false, true},
	{cryptoutilOpenapiModel.A256CBCHS512RSAOAEP256, false, true}, {cryptoutilOpenapiModel.A192CBCHS384RSAOAEP256, false, true}, {cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256, false, true},
	{cryptoutilOpenapiModel.A256CBCHS512RSAOAEP, false, true}, {cryptoutilOpenapiModel.A192CBCHS384RSAOAEP, false, true}, {cryptoutilOpenapiModel.A128CBCHS256RSAOAEP, false, true},
	{cryptoutilOpenapiModel.A256CBCHS512RSA15, false, true}, {cryptoutilOpenapiModel.A192CBCHS384RSA15, false, true}, {cryptoutilOpenapiModel.A128CBCHS256RSA15, false, true},

	{cryptoutilOpenapiModel.A256CBCHS512ECDHESA256KW, false, true}, {cryptoutilOpenapiModel.A192CBCHS384ECDHESA256KW, false, true}, {cryptoutilOpenapiModel.A128CBCHS256ECDHESA256KW, false, true},
	{cryptoutilOpenapiModel.A192CBCHS384ECDHESA192KW, false, true}, {cryptoutilOpenapiModel.A128CBCHS256ECDHESA192KW, false, true}, {cryptoutilOpenapiModel.A128CBCHS256ECDHESA128KW, false, true},
	{cryptoutilOpenapiModel.A256CBCHS512ECDHES, false, true}, {cryptoutilOpenapiModel.A192CBCHS384ECDHES, false, true}, {cryptoutilOpenapiModel.A128CBCHS256ECDHES, false, true},

	{cryptoutilOpenapiModel.RS512, false, true}, {cryptoutilOpenapiModel.RS384, false, true}, {cryptoutilOpenapiModel.RS256, false, true},
	{cryptoutilOpenapiModel.PS512, false, true}, {cryptoutilOpenapiModel.PS384, false, true}, {cryptoutilOpenapiModel.PS256, false, true},
	{cryptoutilOpenapiModel.ES512, false, true}, {cryptoutilOpenapiModel.ES384, false, true}, {cryptoutilOpenapiModel.ES256, false, true},
	{cryptoutilOpenapiModel.HS512, true, false}, {cryptoutilOpenapiModel.HS384, true, false}, {cryptoutilOpenapiModel.HS256, true, false},
	{cryptoutilOpenapiModel.EdDSA, false, true},
}

func Test_ElasticKeyAlgorithm_Symmetric(t *testing.T) {
	for _, alg := range happyPathTestCases2 {
		t.Run(strings.ReplaceAll(string(alg.actualElasticKeyAlgorithm), "/", "_"), func(t *testing.T) {
			isSymmetric, err := IsSymmetric(&alg.actualElasticKeyAlgorithm)
			require.NoError(t, err, "IsSymmetric(%q)", alg.actualElasticKeyAlgorithm)
			require.Equal(t, alg.expectedIsSymmetric, isSymmetric, "IsSymmetric(%q)", alg.actualElasticKeyAlgorithm)
		})
	}
}

func Test_ElasticKeyAlgorithmAsymmetric(t *testing.T) {
	for _, alg := range happyPathTestCases2 {
		t.Run(strings.ReplaceAll(string(alg.actualElasticKeyAlgorithm), "/", "_"), func(t *testing.T) {
			isAsymmetric, err := IsAsymmetric(&alg.actualElasticKeyAlgorithm)
			require.NoError(t, err, "IsAsymmetric(%q)", alg.actualElasticKeyAlgorithm)
			require.Equal(t, alg.expectedIsAsymmetric, isAsymmetric, "IsAsymmetric(%q)", alg.actualElasticKeyAlgorithm)
		})
	}
}
