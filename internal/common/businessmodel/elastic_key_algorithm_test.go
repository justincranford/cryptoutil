package businessmodel

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type TestCase struct {
	actualElasticKeyAlgorithm ElasticKeyAlgorithm
	expectedIsSymmetric       bool
	expectedIsAsymmetric      bool
}

var happyPathTestCases = []TestCase{
	{A256GCM_A256KW, true, false}, {A192GCM_A256KW, true, false}, {A128GCM_A256KW, true, false},
	{A256GCM_A192KW, true, false}, {A192GCM_A192KW, true, false}, {A128GCM_A192KW, true, false},
	{A256GCM_A128KW, true, false}, {A192GCM_A128KW, true, false}, {A128GCM_A128KW, true, false},
	{A256GCM_A256GCMKW, true, false}, {A192GCM_A256GCMKW, true, false}, {A128GCM_A256GCMKW, true, false},
	{A256GCM_A192GCMKW, true, false}, {A192GCM_A192GCMKW, true, false}, {A128GCM_A192GCMKW, true, false},
	{A256GCM_A128GCMKW, true, false}, {A192GCM_A128GCMKW, true, false}, {A128GCM_A128GCMKW, true, false},
	{A256GCM_dir, true, false}, {A192GCM_dir, true, false}, {A128GCM_dir, true, false},

	{A256GCM_RSAOAEP512, false, true}, {A192GCM_RSAOAEP512, false, true}, {A128GCM_RSAOAEP512, false, true},
	{A256GCM_RSAOAEP384, false, true}, {A192GCM_RSAOAEP384, false, true}, {A128GCM_RSAOAEP384, false, true},
	{A256GCM_RSAOAEP256, false, true}, {A192GCM_RSAOAEP256, false, true}, {A128GCM_RSAOAEP256, false, true},
	{A256GCM_RSAOAEP, false, true}, {A192GCM_RSAOAEP, false, true}, {A128GCM_RSAOAEP, false, true},
	{A256GCM_RSA15, false, true}, {A192GCM_RSA15, false, true}, {A128GCM_RSA15, false, true},

	{A256GCM_ECDHESA256KW, false, true}, {A192GCM_ECDHESA256KW, false, true}, {A128GCM_ECDHESA256KW, false, true},
	{A256GCM_ECDHESA192KW, false, true}, {A192GCM_ECDHESA192KW, false, true}, {A128GCM_ECDHESA192KW, false, true},
	{A256GCM_ECDHESA128KW, false, true}, {A192GCM_ECDHESA128KW, false, true}, {A128GCM_ECDHESA128KW, false, true},
	{A256GCM_ECDHES, false, true}, {A192GCM_ECDHES, false, true}, {A128GCM_ECDHES, false, true},

	{A256CBCHS512_A256KW, true, false}, {A192CBCHS384_A256KW, true, false}, {A128CBCHS256_A256KW, true, false},
	{A256CBCHS512_A192KW, true, false}, {A192CBCHS384_A192KW, true, false}, {A128CBCHS256_A192KW, true, false},
	{A256CBCHS512_A128KW, true, false}, {A192CBCHS384_A128KW, true, false}, {A128CBCHS256_A128KW, true, false},
	{A256CBCHS512_A256GCMKW, true, false}, {A192CBCHS384_A256GCMKW, true, false}, {A128CBCHS256_A256GCMKW, true, false},
	{A256CBCHS512_A192GCMKW, true, false}, {A192CBCHS384_A192GCMKW, true, false}, {A128CBCHS256_A192GCMKW, true, false},
	{A256CBCHS512_A128GCMKW, true, false}, {A192CBCHS384_A128GCMKW, true, false}, {A128CBCHS256_A128GCMKW, true, false},
	{A256CBCHS512_dir, true, false}, {A192CBCHS384_dir, true, false}, {A128CBCHS256_dir, true, false},

	{A256CBC_HS512_RSAOAEP512, false, true}, {A192CBC_HS384_RSAOAEP512, false, true}, {A128CBC_HS256_RSAOAEP512, false, true},
	{A256CBC_HS512_RSAOAEP384, false, true}, {A192CBC_HS384_RSAOAEP384, false, true}, {A128CBC_HS256_RSAOAEP384, false, true},
	{A256CBC_HS512_RSAOAEP256, false, true}, {A192CBC_HS384_RSAOAEP256, false, true}, {A128CBC_HS256_RSAOAEP256, false, true},
	{A256CBC_HS512_RSAOAEP, false, true}, {A192CBC_HS384_RSAOAEP, false, true}, {A128CBC_HS256_RSAOAEP, false, true},
	{A256CBC_HS512_RSA15, false, true}, {A192CBC_HS384_RSA15, false, true}, {A128CBC_HS256_RSA15, false, true},

	{A256CBC_HS512_ECDHESA256KW, false, true}, {A192CBC_HS384_ECDHESA256KW, false, true}, {A128CBC_HS256_ECDHESA256KW, false, true},
	{A192CBC_HS384_ECDHESA192KW, false, true}, {A128CBC_HS256_ECDHESA192KW, false, true}, {A128CBC_HS256_ECDHESA128KW, false, true},
	{A256CBC_HS512_ECDHES, false, true}, {A192CBC_HS384_ECDHES, false, true}, {A128CBC_HS256_ECDHES, false, true},

	{RS512, false, true}, {RS384, false, true}, {RS256, false, true},
	{PS512, false, true}, {PS384, false, true}, {PS256, false, true},
	{ES512, false, true}, {ES384, false, true}, {ES256, false, true},
	{HS512, true, false}, {HS384, true, false}, {HS256, true, false},
	{EdDSA, false, true},
}

func Test_SymmetricElasticKeyAlgorithm_MapAndFuncs(t *testing.T) {
	for _, alg := range happyPathTestCases {
		t.Run(string(alg.actualElasticKeyAlgorithm), func(t *testing.T) {
			require.Equal(t, alg.expectedIsSymmetric, IsSymmetric(&alg.actualElasticKeyAlgorithm), "IsSymmetric(%q)", alg.actualElasticKeyAlgorithm)
			require.Equal(t, alg.expectedIsAsymmetric, IsAsymmetric(&alg.actualElasticKeyAlgorithm), "IsAsymmetric(%q)", alg.actualElasticKeyAlgorithm)
		})
	}
}
