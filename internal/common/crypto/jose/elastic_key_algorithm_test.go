package jose

import (
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"

	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type SplitTestCase struct {
	actualElasticKeyAlgorithm cryptoutilOpenapiModel.ElasticKeyAlgorithm
	expectedSplitString       string
}

type SymmetricTestCase struct {
	actualElasticKeyAlgorithm cryptoutilOpenapiModel.ElasticKeyAlgorithm
	expectedIsSymmetric       bool
	expectedIsAsymmetric      bool
}

var happyPathSplitTestCases = []SplitTestCase{
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256GCMA256KW, expectedSplitString: EncA256GCM.String() + "/" + AlgA256KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192GCMA256KW, expectedSplitString: EncA192GCM.String() + "/" + AlgA256KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128GCMA256KW, expectedSplitString: EncA128GCM.String() + "/" + AlgA256KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256GCMA192KW, expectedSplitString: EncA256GCM.String() + "/" + AlgA192KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192GCMA192KW, expectedSplitString: EncA192GCM.String() + "/" + AlgA192KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128GCMA192KW, expectedSplitString: EncA128GCM.String() + "/" + AlgA192KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256GCMA128KW, expectedSplitString: EncA256GCM.String() + "/" + AlgA128KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192GCMA128KW, expectedSplitString: EncA192GCM.String() + "/" + AlgA128KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128GCMA128KW, expectedSplitString: EncA128GCM.String() + "/" + AlgA128KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256GCMA256GCMKW, expectedSplitString: EncA256GCM.String() + "/" + AlgA256GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192GCMA256GCMKW, expectedSplitString: EncA192GCM.String() + "/" + AlgA256GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128GCMA256GCMKW, expectedSplitString: EncA128GCM.String() + "/" + AlgA256GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256GCMA192GCMKW, expectedSplitString: EncA256GCM.String() + "/" + AlgA192GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192GCMA192GCMKW, expectedSplitString: EncA192GCM.String() + "/" + AlgA192GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128GCMA192GCMKW, expectedSplitString: EncA128GCM.String() + "/" + AlgA192GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256GCMA128GCMKW, expectedSplitString: EncA256GCM.String() + "/" + AlgA128GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192GCMA128GCMKW, expectedSplitString: EncA192GCM.String() + "/" + AlgA128GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128GCMA128GCMKW, expectedSplitString: EncA128GCM.String() + "/" + AlgA128GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256GCMDir, expectedSplitString: EncA256GCM.String() + "/" + AlgDir.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192GCMDir, expectedSplitString: EncA192GCM.String() + "/" + AlgDir.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128GCMDir, expectedSplitString: EncA128GCM.String() + "/" + AlgDir.String()},

	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256GCMRSAOAEP512, expectedSplitString: EncA256GCM.String() + "/" + AlgRSAOAEP512.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192GCMRSAOAEP512, expectedSplitString: EncA192GCM.String() + "/" + AlgRSAOAEP512.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128GCMRSAOAEP512, expectedSplitString: EncA128GCM.String() + "/" + AlgRSAOAEP512.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256GCMRSAOAEP384, expectedSplitString: EncA256GCM.String() + "/" + AlgRSAOAEP384.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192GCMRSAOAEP384, expectedSplitString: EncA192GCM.String() + "/" + AlgRSAOAEP384.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128GCMRSAOAEP384, expectedSplitString: EncA128GCM.String() + "/" + AlgRSAOAEP384.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256GCMRSAOAEP256, expectedSplitString: EncA256GCM.String() + "/" + AlgRSAOAEP256.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192GCMRSAOAEP256, expectedSplitString: EncA192GCM.String() + "/" + AlgRSAOAEP256.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128GCMRSAOAEP256, expectedSplitString: EncA128GCM.String() + "/" + AlgRSAOAEP256.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256GCMRSAOAEP, expectedSplitString: EncA256GCM.String() + "/" + AlgRSAOAEP.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192GCMRSAOAEP, expectedSplitString: EncA192GCM.String() + "/" + AlgRSAOAEP.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128GCMRSAOAEP, expectedSplitString: EncA128GCM.String() + "/" + AlgRSAOAEP.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256GCMRSA15, expectedSplitString: EncA256GCM.String() + "/" + AlgRSA15.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192GCMRSA15, expectedSplitString: EncA192GCM.String() + "/" + AlgRSA15.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128GCMRSA15, expectedSplitString: EncA128GCM.String() + "/" + AlgRSA15.String()},

	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256GCMECDHESA256KW, expectedSplitString: EncA256GCM.String() + "/" + AlgECDHESA256KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192GCMECDHESA256KW, expectedSplitString: EncA192GCM.String() + "/" + AlgECDHESA256KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128GCMECDHESA256KW, expectedSplitString: EncA128GCM.String() + "/" + AlgECDHESA256KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256GCMECDHESA192KW, expectedSplitString: EncA256GCM.String() + "/" + AlgECDHESA192KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192GCMECDHESA192KW, expectedSplitString: EncA192GCM.String() + "/" + AlgECDHESA192KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128GCMECDHESA192KW, expectedSplitString: EncA128GCM.String() + "/" + AlgECDHESA192KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256GCMECDHESA128KW, expectedSplitString: EncA256GCM.String() + "/" + AlgECDHESA128KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192GCMECDHESA128KW, expectedSplitString: EncA192GCM.String() + "/" + AlgECDHESA128KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128GCMECDHESA128KW, expectedSplitString: EncA128GCM.String() + "/" + AlgECDHESA128KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256GCMECDHES, expectedSplitString: EncA256GCM.String() + "/" + AlgECDHES.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192GCMECDHES, expectedSplitString: EncA192GCM.String() + "/" + AlgECDHES.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128GCMECDHES, expectedSplitString: EncA128GCM.String() + "/" + AlgECDHES.String()},

	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256CBCHS512A256KW, expectedSplitString: EncA256CBCHS512.String() + "/" + AlgA256KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192CBCHS384A256KW, expectedSplitString: EncA192CBCHS384.String() + "/" + AlgA256KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128CBCHS256A256KW, expectedSplitString: EncA128CBCHS256.String() + "/" + AlgA256KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256CBCHS512A192KW, expectedSplitString: EncA256CBCHS512.String() + "/" + AlgA192KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192CBCHS384A192KW, expectedSplitString: EncA192CBCHS384.String() + "/" + AlgA192KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128CBCHS256A192KW, expectedSplitString: EncA128CBCHS256.String() + "/" + AlgA192KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256CBCHS512A128KW, expectedSplitString: EncA256CBCHS512.String() + "/" + AlgA128KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192CBCHS384A128KW, expectedSplitString: EncA192CBCHS384.String() + "/" + AlgA128KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128CBCHS256A128KW, expectedSplitString: EncA128CBCHS256.String() + "/" + AlgA128KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256CBCHS512A256GCMKW, expectedSplitString: EncA256CBCHS512.String() + "/" + AlgA256GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192CBCHS384A256GCMKW, expectedSplitString: EncA192CBCHS384.String() + "/" + AlgA256GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128CBCHS256A256GCMKW, expectedSplitString: EncA128CBCHS256.String() + "/" + AlgA256GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256CBCHS512A192GCMKW, expectedSplitString: EncA256CBCHS512.String() + "/" + AlgA192GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192CBCHS384A192GCMKW, expectedSplitString: EncA192CBCHS384.String() + "/" + AlgA192GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128CBCHS256A192GCMKW, expectedSplitString: EncA128CBCHS256.String() + "/" + AlgA192GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256CBCHS512A128GCMKW, expectedSplitString: EncA256CBCHS512.String() + "/" + AlgA128GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192CBCHS384A128GCMKW, expectedSplitString: EncA192CBCHS384.String() + "/" + AlgA128GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128CBCHS256A128GCMKW, expectedSplitString: EncA128CBCHS256.String() + "/" + AlgA128GCMKW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256CBCHS512Dir, expectedSplitString: EncA256CBCHS512.String() + "/" + AlgDir.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192CBCHS384Dir, expectedSplitString: EncA192CBCHS384.String() + "/" + AlgDir.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128CBCHS256Dir, expectedSplitString: EncA128CBCHS256.String() + "/" + AlgDir.String()},

	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256CBCHS512RSAOAEP512, expectedSplitString: EncA256CBCHS512.String() + "/" + AlgRSAOAEP512.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192CBCHS384RSAOAEP512, expectedSplitString: EncA192CBCHS384.String() + "/" + AlgRSAOAEP512.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128CBCHS256RSAOAEP512, expectedSplitString: EncA128CBCHS256.String() + "/" + AlgRSAOAEP512.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256CBCHS512RSAOAEP384, expectedSplitString: EncA256CBCHS512.String() + "/" + AlgRSAOAEP384.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192CBCHS384RSAOAEP384, expectedSplitString: EncA192CBCHS384.String() + "/" + AlgRSAOAEP384.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128CBCHS256RSAOAEP384, expectedSplitString: EncA128CBCHS256.String() + "/" + AlgRSAOAEP384.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256CBCHS512RSAOAEP256, expectedSplitString: EncA256CBCHS512.String() + "/" + AlgRSAOAEP256.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192CBCHS384RSAOAEP256, expectedSplitString: EncA192CBCHS384.String() + "/" + AlgRSAOAEP256.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256, expectedSplitString: EncA128CBCHS256.String() + "/" + AlgRSAOAEP256.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256CBCHS512RSAOAEP, expectedSplitString: EncA256CBCHS512.String() + "/" + AlgRSAOAEP.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192CBCHS384RSAOAEP, expectedSplitString: EncA192CBCHS384.String() + "/" + AlgRSAOAEP.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128CBCHS256RSAOAEP, expectedSplitString: EncA128CBCHS256.String() + "/" + AlgRSAOAEP.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256CBCHS512RSA15, expectedSplitString: EncA256CBCHS512.String() + "/" + AlgRSA15.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192CBCHS384RSA15, expectedSplitString: EncA192CBCHS384.String() + "/" + AlgRSA15.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128CBCHS256RSA15, expectedSplitString: EncA128CBCHS256.String() + "/" + AlgRSA15.String()},

	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256CBCHS512ECDHESA256KW, expectedSplitString: EncA256CBCHS512.String() + "/" + AlgECDHESA256KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192CBCHS384ECDHESA256KW, expectedSplitString: EncA192CBCHS384.String() + "/" + AlgECDHESA256KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128CBCHS256ECDHESA256KW, expectedSplitString: EncA128CBCHS256.String() + "/" + AlgECDHESA256KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192CBCHS384ECDHESA192KW, expectedSplitString: EncA192CBCHS384.String() + "/" + AlgECDHESA192KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128CBCHS256ECDHESA192KW, expectedSplitString: EncA128CBCHS256.String() + "/" + AlgECDHESA192KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128CBCHS256ECDHESA128KW, expectedSplitString: EncA128CBCHS256.String() + "/" + AlgECDHESA128KW.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A256CBCHS512ECDHES, expectedSplitString: EncA256CBCHS512.String() + "/" + AlgECDHES.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A192CBCHS384ECDHES, expectedSplitString: EncA192CBCHS384.String() + "/" + AlgECDHES.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.A128CBCHS256ECDHES, expectedSplitString: EncA128CBCHS256.String() + "/" + AlgECDHES.String()},

	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.RS512, expectedSplitString: AlgRS512.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.RS384, expectedSplitString: AlgRS384.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.RS256, expectedSplitString: AlgRS256.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.PS512, expectedSplitString: AlgPS512.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.PS384, expectedSplitString: AlgPS384.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.PS256, expectedSplitString: AlgPS256.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.ES512, expectedSplitString: AlgES512.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.ES384, expectedSplitString: AlgES384.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.ES256, expectedSplitString: AlgES256.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.HS512, expectedSplitString: AlgHS512.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.HS384, expectedSplitString: AlgHS384.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.HS256, expectedSplitString: AlgHS256.String()},
	{actualElasticKeyAlgorithm: cryptoutilOpenapiModel.EdDSA, expectedSplitString: AlgEdDSA.String()},
}

var happyPathSymmetricTestCases = []SymmetricTestCase{
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

func Test_HappyPath_Split(t *testing.T) {
	for _, testCase := range happyPathSplitTestCases {
		actualAndExpected := fmt.Sprintf("%s  %s", string(testCase.actualElasticKeyAlgorithm), testCase.expectedSplitString)
		t.Run(strings.ReplaceAll(actualAndExpected, "/", "_"), func(t *testing.T) {
			require.Equal(t, string(testCase.actualElasticKeyAlgorithm), testCase.expectedSplitString)
		})
	}
}

func Test_ElasticKeyAlgorithm_Symmetric(t *testing.T) {
	for _, alg := range happyPathSymmetricTestCases {
		t.Run(strings.ReplaceAll(string(alg.actualElasticKeyAlgorithm), "/", "_"), func(t *testing.T) {
			isSymmetric, err := IsSymmetric(&alg.actualElasticKeyAlgorithm)
			require.NoError(t, err, "IsSymmetric(%q)", alg.actualElasticKeyAlgorithm)
			require.Equal(t, alg.expectedIsSymmetric, isSymmetric, "IsSymmetric(%q)", alg.actualElasticKeyAlgorithm)
		})
	}
}

func Test_ElasticKeyAlgorithmAsymmetric(t *testing.T) {
	for _, alg := range happyPathSymmetricTestCases {
		t.Run(strings.ReplaceAll(string(alg.actualElasticKeyAlgorithm), "/", "_"), func(t *testing.T) {
			isAsymmetric, err := IsAsymmetric(&alg.actualElasticKeyAlgorithm)
			require.NoError(t, err, "IsAsymmetric(%q)", alg.actualElasticKeyAlgorithm)
			require.Equal(t, alg.expectedIsAsymmetric, isAsymmetric, "IsAsymmetric(%q)", alg.actualElasticKeyAlgorithm)
		})
	}
}
