// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"fmt"
	"strings"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"

	"github.com/stretchr/testify/require"
)

const testInvalidAlgorithm = "INVALID"

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
	{cryptoutilOpenapiModel.A256GCMA256KW, true, false},
	{cryptoutilOpenapiModel.A192GCMA256KW, true, false},
	{cryptoutilOpenapiModel.A128GCMA256KW, true, false},
	{cryptoutilOpenapiModel.A256GCMA192KW, true, false},
	{cryptoutilOpenapiModel.A192GCMA192KW, true, false},
	{cryptoutilOpenapiModel.A128GCMA192KW, true, false},
	{cryptoutilOpenapiModel.A256GCMA128KW, true, false},
	{cryptoutilOpenapiModel.A192GCMA128KW, true, false},
	{cryptoutilOpenapiModel.A128GCMA128KW, true, false},
	{cryptoutilOpenapiModel.A256GCMA256GCMKW, true, false},
	{cryptoutilOpenapiModel.A192GCMA256GCMKW, true, false},
	{cryptoutilOpenapiModel.A128GCMA256GCMKW, true, false},
	{cryptoutilOpenapiModel.A256GCMA192GCMKW, true, false},
	{cryptoutilOpenapiModel.A192GCMA192GCMKW, true, false},
	{cryptoutilOpenapiModel.A128GCMA192GCMKW, true, false},
	{cryptoutilOpenapiModel.A256GCMA128GCMKW, true, false},
	{cryptoutilOpenapiModel.A192GCMA128GCMKW, true, false},
	{cryptoutilOpenapiModel.A128GCMA128GCMKW, true, false},
	{cryptoutilOpenapiModel.A256GCMDir, true, false},
	{cryptoutilOpenapiModel.A192GCMDir, true, false},
	{cryptoutilOpenapiModel.A128GCMDir, true, false},

	{cryptoutilOpenapiModel.A256GCMRSAOAEP512, false, true},
	{cryptoutilOpenapiModel.A192GCMRSAOAEP512, false, true},
	{cryptoutilOpenapiModel.A128GCMRSAOAEP512, false, true},
	{cryptoutilOpenapiModel.A256GCMRSAOAEP384, false, true},
	{cryptoutilOpenapiModel.A192GCMRSAOAEP384, false, true},
	{cryptoutilOpenapiModel.A128GCMRSAOAEP384, false, true},
	{cryptoutilOpenapiModel.A256GCMRSAOAEP256, false, true},
	{cryptoutilOpenapiModel.A192GCMRSAOAEP256, false, true},
	{cryptoutilOpenapiModel.A128GCMRSAOAEP256, false, true},
	{cryptoutilOpenapiModel.A256GCMRSAOAEP, false, true},
	{cryptoutilOpenapiModel.A192GCMRSAOAEP, false, true},
	{cryptoutilOpenapiModel.A128GCMRSAOAEP, false, true},
	{cryptoutilOpenapiModel.A256GCMRSA15, false, true},
	{cryptoutilOpenapiModel.A192GCMRSA15, false, true},
	{cryptoutilOpenapiModel.A128GCMRSA15, false, true},

	{cryptoutilOpenapiModel.A256GCMECDHESA256KW, false, true},
	{cryptoutilOpenapiModel.A192GCMECDHESA256KW, false, true},
	{cryptoutilOpenapiModel.A128GCMECDHESA256KW, false, true},
	{cryptoutilOpenapiModel.A256GCMECDHESA192KW, false, true},
	{cryptoutilOpenapiModel.A192GCMECDHESA192KW, false, true},
	{cryptoutilOpenapiModel.A128GCMECDHESA192KW, false, true},
	{cryptoutilOpenapiModel.A256GCMECDHESA128KW, false, true},
	{cryptoutilOpenapiModel.A192GCMECDHESA128KW, false, true},
	{cryptoutilOpenapiModel.A128GCMECDHESA128KW, false, true},
	{cryptoutilOpenapiModel.A256GCMECDHES, false, true},
	{cryptoutilOpenapiModel.A192GCMECDHES, false, true},
	{cryptoutilOpenapiModel.A128GCMECDHES, false, true},

	{cryptoutilOpenapiModel.A256CBCHS512A256KW, true, false},
	{cryptoutilOpenapiModel.A192CBCHS384A256KW, true, false},
	{cryptoutilOpenapiModel.A128CBCHS256A256KW, true, false},
	{cryptoutilOpenapiModel.A256CBCHS512A192KW, true, false},
	{cryptoutilOpenapiModel.A192CBCHS384A192KW, true, false},
	{cryptoutilOpenapiModel.A128CBCHS256A192KW, true, false},
	{cryptoutilOpenapiModel.A256CBCHS512A128KW, true, false},
	{cryptoutilOpenapiModel.A192CBCHS384A128KW, true, false},
	{cryptoutilOpenapiModel.A128CBCHS256A128KW, true, false},
	{cryptoutilOpenapiModel.A256CBCHS512A256GCMKW, true, false},
	{cryptoutilOpenapiModel.A192CBCHS384A256GCMKW, true, false},
	{cryptoutilOpenapiModel.A128CBCHS256A256GCMKW, true, false},
	{cryptoutilOpenapiModel.A256CBCHS512A192GCMKW, true, false},
	{cryptoutilOpenapiModel.A192CBCHS384A192GCMKW, true, false},
	{cryptoutilOpenapiModel.A128CBCHS256A192GCMKW, true, false},
	{cryptoutilOpenapiModel.A256CBCHS512A128GCMKW, true, false},
	{cryptoutilOpenapiModel.A192CBCHS384A128GCMKW, true, false},
	{cryptoutilOpenapiModel.A128CBCHS256A128GCMKW, true, false},
	{cryptoutilOpenapiModel.A256CBCHS512Dir, true, false},
	{cryptoutilOpenapiModel.A192CBCHS384Dir, true, false},
	{cryptoutilOpenapiModel.A128CBCHS256Dir, true, false},

	{cryptoutilOpenapiModel.A256CBCHS512RSAOAEP512, false, true},
	{cryptoutilOpenapiModel.A192CBCHS384RSAOAEP512, false, true},
	{cryptoutilOpenapiModel.A128CBCHS256RSAOAEP512, false, true},
	{cryptoutilOpenapiModel.A256CBCHS512RSAOAEP384, false, true},
	{cryptoutilOpenapiModel.A192CBCHS384RSAOAEP384, false, true},
	{cryptoutilOpenapiModel.A128CBCHS256RSAOAEP384, false, true},
	{cryptoutilOpenapiModel.A256CBCHS512RSAOAEP256, false, true},
	{cryptoutilOpenapiModel.A192CBCHS384RSAOAEP256, false, true},
	{cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256, false, true},
	{cryptoutilOpenapiModel.A256CBCHS512RSAOAEP, false, true},
	{cryptoutilOpenapiModel.A192CBCHS384RSAOAEP, false, true},
	{cryptoutilOpenapiModel.A128CBCHS256RSAOAEP, false, true},
	{cryptoutilOpenapiModel.A256CBCHS512RSA15, false, true},
	{cryptoutilOpenapiModel.A192CBCHS384RSA15, false, true},
	{cryptoutilOpenapiModel.A128CBCHS256RSA15, false, true},

	{cryptoutilOpenapiModel.A256CBCHS512ECDHESA256KW, false, true},
	{cryptoutilOpenapiModel.A192CBCHS384ECDHESA256KW, false, true},
	{cryptoutilOpenapiModel.A128CBCHS256ECDHESA256KW, false, true},
	{cryptoutilOpenapiModel.A192CBCHS384ECDHESA192KW, false, true},
	{cryptoutilOpenapiModel.A128CBCHS256ECDHESA192KW, false, true},
	{cryptoutilOpenapiModel.A128CBCHS256ECDHESA128KW, false, true},
	{cryptoutilOpenapiModel.A256CBCHS512ECDHES, false, true},
	{cryptoutilOpenapiModel.A192CBCHS384ECDHES, false, true},
	{cryptoutilOpenapiModel.A128CBCHS256ECDHES, false, true},

	{cryptoutilOpenapiModel.RS512, false, true},
	{cryptoutilOpenapiModel.RS384, false, true},
	{cryptoutilOpenapiModel.RS256, false, true},
	{cryptoutilOpenapiModel.PS512, false, true},
	{cryptoutilOpenapiModel.PS384, false, true},
	{cryptoutilOpenapiModel.PS256, false, true},
	{cryptoutilOpenapiModel.ES512, false, true},
	{cryptoutilOpenapiModel.ES384, false, true},
	{cryptoutilOpenapiModel.ES256, false, true},
	{cryptoutilOpenapiModel.HS512, true, false},
	{cryptoutilOpenapiModel.HS384, true, false},
	{cryptoutilOpenapiModel.HS256, true, false},
	{cryptoutilOpenapiModel.EdDSA, false, true},
}

func Test_HappyPath_Split(t *testing.T) {
	t.Parallel()

	for _, testCase := range happyPathSplitTestCases {
		actualAndExpected := fmt.Sprintf("%s  %s", string(testCase.actualElasticKeyAlgorithm), testCase.expectedSplitString)
		t.Run(strings.ReplaceAll(actualAndExpected, "/", "_"), func(t *testing.T) {
			require.Equal(t, string(testCase.actualElasticKeyAlgorithm), testCase.expectedSplitString)
		})
	}
}

func Test_ElasticKeyAlgorithm_Symmetric(t *testing.T) {
	t.Parallel()

	for _, alg := range happyPathSymmetricTestCases {
		t.Run(strings.ReplaceAll(string(alg.actualElasticKeyAlgorithm), "/", "_"), func(t *testing.T) {
			isSymmetric, err := IsSymmetric(&alg.actualElasticKeyAlgorithm)
			require.NoError(t, err, "IsSymmetric(%q)", alg.actualElasticKeyAlgorithm)
			require.Equal(t, alg.expectedIsSymmetric, isSymmetric, "IsSymmetric(%q)", alg.actualElasticKeyAlgorithm)
		})
	}
}

func Test_ElasticKeyAlgorithmAsymmetric(t *testing.T) {
	t.Parallel()

	for _, alg := range happyPathSymmetricTestCases {
		t.Run(strings.ReplaceAll(string(alg.actualElasticKeyAlgorithm), "/", "_"), func(t *testing.T) {
			isAsymmetric, err := IsAsymmetric(&alg.actualElasticKeyAlgorithm)
			require.NoError(t, err, "IsAsymmetric(%q)", alg.actualElasticKeyAlgorithm)
			require.Equal(t, alg.expectedIsAsymmetric, isAsymmetric, "IsAsymmetric(%q)", alg.actualElasticKeyAlgorithm)
		})
	}
}

func Test_ToJWEEncAndAlg_ValidAlgorithms(t *testing.T) {
	t.Parallel()

	tests := []cryptoutilOpenapiModel.ElasticKeyAlgorithm{
		cryptoutilOpenapiModel.A256GCMA256KW,
		cryptoutilOpenapiModel.A192GCMA256KW,
		cryptoutilOpenapiModel.A128GCMA256KW,
		cryptoutilOpenapiModel.A256GCMRSAOAEP512,
		cryptoutilOpenapiModel.A256GCMECDHESA256KW,
	}

	for _, alg := range tests {
		t.Run(string(alg), func(t *testing.T) {
			t.Parallel()

			enc, keyAlg, err := ToJWEEncAndAlg(&alg)
			require.NoError(t, err)
			require.NotNil(t, enc)
			require.NotNil(t, keyAlg)
		})
	}
}

func Test_ToJWEEncAndAlg_InvalidAlgorithm(t *testing.T) {
	t.Parallel()

	// JWS algorithms should fail when used with ToJWEEncAndAlg.
	alg := cryptoutilOpenapiModel.RS256

	enc, keyAlg, err := ToJWEEncAndAlg(&alg)
	require.Error(t, err)
	require.Nil(t, enc)
	require.Nil(t, keyAlg)
	require.Contains(t, err.Error(), "unsupported JWE ElasticKeyAlgorithm")
}

func Test_ToJWSAlg_ValidAlgorithms(t *testing.T) {
	t.Parallel()

	tests := []cryptoutilOpenapiModel.ElasticKeyAlgorithm{
		cryptoutilOpenapiModel.RS256,
		cryptoutilOpenapiModel.RS384,
		cryptoutilOpenapiModel.RS512,
		cryptoutilOpenapiModel.PS256,
		cryptoutilOpenapiModel.PS384,
		cryptoutilOpenapiModel.PS512,
		cryptoutilOpenapiModel.ES256,
		cryptoutilOpenapiModel.ES384,
		cryptoutilOpenapiModel.ES512,
		cryptoutilOpenapiModel.HS256,
		cryptoutilOpenapiModel.HS384,
		cryptoutilOpenapiModel.HS512,
		cryptoutilOpenapiModel.EdDSA,
	}

	for _, alg := range tests {
		t.Run(string(alg), func(t *testing.T) {
			t.Parallel()

			sigAlg, err := ToJWSAlg(&alg)
			require.NoError(t, err)
			require.NotNil(t, sigAlg)
		})
	}
}

func Test_ToJWSAlg_InvalidAlgorithm(t *testing.T) {
	t.Parallel()

	// JWE algorithms should fail when used with ToJWSAlg.
	alg := cryptoutilOpenapiModel.A256GCMA256KW

	sigAlg, err := ToJWSAlg(&alg)
	require.Error(t, err)
	require.Nil(t, sigAlg)
	require.Contains(t, err.Error(), "unsupported JWS ElasticKeyAlgorithm")
}

func Test_IsJWE(t *testing.T) {
	t.Parallel()

	// JWE algorithms should return true.
	jweAlg := cryptoutilOpenapiModel.A256GCMA256KW
	require.True(t, IsJWE(&jweAlg))

	// JWS algorithms should return false.
	jwsAlg := cryptoutilOpenapiModel.RS256
	require.False(t, IsJWE(&jwsAlg))
}

func Test_IsJWS(t *testing.T) {
	t.Parallel()

	// JWS algorithms should return true.
	jwsAlg := cryptoutilOpenapiModel.RS256
	require.True(t, IsJWS(&jwsAlg))

	// JWE algorithms should return false.
	jweAlg := cryptoutilOpenapiModel.A256GCMA256KW
	require.False(t, IsJWS(&jweAlg))
}

func Test_ToElasticKeyAlgorithm_ValidAlgorithms(t *testing.T) {
	t.Parallel()

	tests := []string{
		string(cryptoutilOpenapiModel.A256GCMA256KW),
		string(cryptoutilOpenapiModel.RS256),
		string(cryptoutilOpenapiModel.EdDSA),
	}

	for _, alg := range tests {
		t.Run(alg, func(t *testing.T) {
			t.Parallel()

			result, err := ToElasticKeyAlgorithm(&alg)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, alg, string(*result))
		})
	}
}

func Test_ToElasticKeyAlgorithm_InvalidAlgorithm(t *testing.T) {
	t.Parallel()

	alg := testInvalidAlgorithm

	result, err := ToElasticKeyAlgorithm(&alg)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "invalid elastic Key algorithm")
}

func Test_ToGenerateAlgorithm_ValidAlgorithms(t *testing.T) {
	t.Parallel()

	tests := []string{
		string(cryptoutilOpenapiModel.RSA2048),
		string(cryptoutilOpenapiModel.ECP256),
		string(cryptoutilOpenapiModel.OKPEd25519),
		string(cryptoutilOpenapiModel.Oct128),
		string(cryptoutilOpenapiModel.Oct256),
	}

	for _, alg := range tests {
		t.Run(alg, func(t *testing.T) {
			t.Parallel()

			result, err := ToGenerateAlgorithm(&alg)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, alg, string(*result))
		})
	}
}

func Test_ToGenerateAlgorithm_InvalidAlgorithm(t *testing.T) {
	t.Parallel()

	alg := testInvalidAlgorithm

	result, err := ToGenerateAlgorithm(&alg)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "invalid generate algorithm")
}

func Test_IsSymmetric_InvalidAlgorithm(t *testing.T) {
	t.Parallel()

	alg := cryptoutilOpenapiModel.ElasticKeyAlgorithm("INVALID")

	result, err := IsSymmetric(&alg)
	require.Error(t, err)
	require.False(t, result)
	require.Contains(t, err.Error(), "unsupported ElasticKeyAlgorithm")
}

func Test_IsAsymmetric_InvalidAlgorithm(t *testing.T) {
	t.Parallel()

	alg := cryptoutilOpenapiModel.ElasticKeyAlgorithm("INVALID")

	result, err := IsAsymmetric(&alg)
	require.Error(t, err)
	require.False(t, result)
	require.Contains(t, err.Error(), "unsupported ElasticKeyAlgorithm")
}
