// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"fmt"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilMagic "cryptoutil/internal/shared/magic"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
)

var (
	generateAlgorithms = map[string]cryptoutilOpenapiModel.GenerateAlgorithm{
		string(cryptoutilOpenapiModel.RSA4096):    cryptoutilOpenapiModel.RSA4096,
		string(cryptoutilOpenapiModel.RSA3072):    cryptoutilOpenapiModel.RSA3072,
		string(cryptoutilOpenapiModel.RSA2048):    cryptoutilOpenapiModel.RSA2048,
		string(cryptoutilOpenapiModel.ECP521):     cryptoutilOpenapiModel.ECP521,
		string(cryptoutilOpenapiModel.ECP384):     cryptoutilOpenapiModel.ECP384,
		string(cryptoutilOpenapiModel.ECP256):     cryptoutilOpenapiModel.ECP256,
		string(cryptoutilOpenapiModel.OKPEd25519): cryptoutilOpenapiModel.OKPEd25519,
		string(cryptoutilOpenapiModel.Oct512):     cryptoutilOpenapiModel.Oct512,
		string(cryptoutilOpenapiModel.Oct384):     cryptoutilOpenapiModel.Oct384,
		string(cryptoutilOpenapiModel.Oct256):     cryptoutilOpenapiModel.Oct256,
		string(cryptoutilOpenapiModel.Oct192):     cryptoutilOpenapiModel.Oct192,
		string(cryptoutilOpenapiModel.Oct128):     cryptoutilOpenapiModel.Oct128,
	}

	elasticKeyAlgorithms = map[string]cryptoutilOpenapiModel.ElasticKeyAlgorithm{
		string(cryptoutilOpenapiModel.A256GCMA256KW): cryptoutilOpenapiModel.A256GCMA256KW,
		string(cryptoutilOpenapiModel.A192GCMA256KW): cryptoutilOpenapiModel.A192GCMA256KW,
		string(cryptoutilOpenapiModel.A128GCMA256KW): cryptoutilOpenapiModel.A128GCMA256KW,
		string(cryptoutilOpenapiModel.A256GCMA192KW): cryptoutilOpenapiModel.A256GCMA192KW,
		string(cryptoutilOpenapiModel.A192GCMA192KW): cryptoutilOpenapiModel.A192GCMA192KW,
		string(cryptoutilOpenapiModel.A128GCMA192KW): cryptoutilOpenapiModel.A128GCMA192KW,
		string(cryptoutilOpenapiModel.A256GCMA128KW): cryptoutilOpenapiModel.A256GCMA128KW,
		string(cryptoutilOpenapiModel.A192GCMA128KW): cryptoutilOpenapiModel.A192GCMA128KW,
		string(cryptoutilOpenapiModel.A128GCMA128KW): cryptoutilOpenapiModel.A128GCMA128KW,

		string(cryptoutilOpenapiModel.A256GCMA256GCMKW): cryptoutilOpenapiModel.A256GCMA256GCMKW,
		string(cryptoutilOpenapiModel.A192GCMA256GCMKW): cryptoutilOpenapiModel.A192GCMA256GCMKW,
		string(cryptoutilOpenapiModel.A128GCMA256GCMKW): cryptoutilOpenapiModel.A128GCMA256GCMKW,
		string(cryptoutilOpenapiModel.A256GCMA192GCMKW): cryptoutilOpenapiModel.A256GCMA192GCMKW,
		string(cryptoutilOpenapiModel.A192GCMA192GCMKW): cryptoutilOpenapiModel.A192GCMA192GCMKW,
		string(cryptoutilOpenapiModel.A128GCMA192GCMKW): cryptoutilOpenapiModel.A128GCMA192GCMKW,
		string(cryptoutilOpenapiModel.A256GCMA128GCMKW): cryptoutilOpenapiModel.A256GCMA128GCMKW,
		string(cryptoutilOpenapiModel.A192GCMA128GCMKW): cryptoutilOpenapiModel.A192GCMA128GCMKW,
		string(cryptoutilOpenapiModel.A128GCMA128GCMKW): cryptoutilOpenapiModel.A128GCMA128GCMKW,

		string(cryptoutilOpenapiModel.A256GCMDir): cryptoutilOpenapiModel.A256GCMDir,
		string(cryptoutilOpenapiModel.A192GCMDir): cryptoutilOpenapiModel.A192GCMDir,
		string(cryptoutilOpenapiModel.A128GCMDir): cryptoutilOpenapiModel.A128GCMDir,

		string(cryptoutilOpenapiModel.A256GCMRSAOAEP512): cryptoutilOpenapiModel.A256GCMRSAOAEP512,
		string(cryptoutilOpenapiModel.A192GCMRSAOAEP512): cryptoutilOpenapiModel.A192GCMRSAOAEP512,
		string(cryptoutilOpenapiModel.A128GCMRSAOAEP512): cryptoutilOpenapiModel.A128GCMRSAOAEP512,
		string(cryptoutilOpenapiModel.A256GCMRSAOAEP384): cryptoutilOpenapiModel.A256GCMRSAOAEP384,
		string(cryptoutilOpenapiModel.A192GCMRSAOAEP384): cryptoutilOpenapiModel.A192GCMRSAOAEP384,
		string(cryptoutilOpenapiModel.A128GCMRSAOAEP384): cryptoutilOpenapiModel.A128GCMRSAOAEP384,
		string(cryptoutilOpenapiModel.A256GCMRSAOAEP256): cryptoutilOpenapiModel.A256GCMRSAOAEP256,
		string(cryptoutilOpenapiModel.A192GCMRSAOAEP256): cryptoutilOpenapiModel.A192GCMRSAOAEP256,
		string(cryptoutilOpenapiModel.A128GCMRSAOAEP256): cryptoutilOpenapiModel.A128GCMRSAOAEP256,
		string(cryptoutilOpenapiModel.A256GCMRSAOAEP):    cryptoutilOpenapiModel.A256GCMRSAOAEP,
		string(cryptoutilOpenapiModel.A192GCMRSAOAEP):    cryptoutilOpenapiModel.A192GCMRSAOAEP,
		string(cryptoutilOpenapiModel.A128GCMRSAOAEP):    cryptoutilOpenapiModel.A128GCMRSAOAEP,
		string(cryptoutilOpenapiModel.A256GCMRSA15):      cryptoutilOpenapiModel.A256GCMRSA15,
		string(cryptoutilOpenapiModel.A192GCMRSA15):      cryptoutilOpenapiModel.A192GCMRSA15,
		string(cryptoutilOpenapiModel.A128GCMRSA15):      cryptoutilOpenapiModel.A128GCMRSA15,

		string(cryptoutilOpenapiModel.A256GCMECDHESA256KW): cryptoutilOpenapiModel.A256GCMECDHESA256KW,
		string(cryptoutilOpenapiModel.A192GCMECDHESA256KW): cryptoutilOpenapiModel.A192GCMECDHESA256KW,
		string(cryptoutilOpenapiModel.A128GCMECDHESA256KW): cryptoutilOpenapiModel.A128GCMECDHESA256KW,
		string(cryptoutilOpenapiModel.A256GCMECDHESA192KW): cryptoutilOpenapiModel.A256GCMECDHESA192KW,
		string(cryptoutilOpenapiModel.A192GCMECDHESA192KW): cryptoutilOpenapiModel.A192GCMECDHESA192KW,
		string(cryptoutilOpenapiModel.A128GCMECDHESA192KW): cryptoutilOpenapiModel.A128GCMECDHESA192KW,
		string(cryptoutilOpenapiModel.A256GCMECDHESA128KW): cryptoutilOpenapiModel.A256GCMECDHESA128KW,
		string(cryptoutilOpenapiModel.A192GCMECDHESA128KW): cryptoutilOpenapiModel.A192GCMECDHESA128KW,
		string(cryptoutilOpenapiModel.A128GCMECDHESA128KW): cryptoutilOpenapiModel.A128GCMECDHESA128KW,
		string(cryptoutilOpenapiModel.A256GCMECDHES):       cryptoutilOpenapiModel.A256GCMECDHES,
		string(cryptoutilOpenapiModel.A192GCMECDHES):       cryptoutilOpenapiModel.A192GCMECDHES,
		string(cryptoutilOpenapiModel.A128GCMECDHES):       cryptoutilOpenapiModel.A128GCMECDHES,

		string(cryptoutilOpenapiModel.A256CBCHS512A256KW): cryptoutilOpenapiModel.A256CBCHS512A256KW,
		string(cryptoutilOpenapiModel.A192CBCHS384A256KW): cryptoutilOpenapiModel.A192CBCHS384A256KW,
		string(cryptoutilOpenapiModel.A128CBCHS256A256KW): cryptoutilOpenapiModel.A128CBCHS256A256KW,
		string(cryptoutilOpenapiModel.A256CBCHS512A192KW): cryptoutilOpenapiModel.A256CBCHS512A192KW,
		string(cryptoutilOpenapiModel.A192CBCHS384A192KW): cryptoutilOpenapiModel.A192CBCHS384A192KW,
		string(cryptoutilOpenapiModel.A128CBCHS256A192KW): cryptoutilOpenapiModel.A128CBCHS256A192KW,
		string(cryptoutilOpenapiModel.A256CBCHS512A128KW): cryptoutilOpenapiModel.A256CBCHS512A128KW,
		string(cryptoutilOpenapiModel.A192CBCHS384A128KW): cryptoutilOpenapiModel.A192CBCHS384A128KW,
		string(cryptoutilOpenapiModel.A128CBCHS256A128KW): cryptoutilOpenapiModel.A128CBCHS256A128KW,

		string(cryptoutilOpenapiModel.A256CBCHS512A256GCMKW): cryptoutilOpenapiModel.A256CBCHS512A256GCMKW,
		string(cryptoutilOpenapiModel.A192CBCHS384A256GCMKW): cryptoutilOpenapiModel.A192CBCHS384A256GCMKW,
		string(cryptoutilOpenapiModel.A128CBCHS256A256GCMKW): cryptoutilOpenapiModel.A128CBCHS256A256GCMKW,
		string(cryptoutilOpenapiModel.A256CBCHS512A192GCMKW): cryptoutilOpenapiModel.A256CBCHS512A192GCMKW,
		string(cryptoutilOpenapiModel.A192CBCHS384A192GCMKW): cryptoutilOpenapiModel.A192CBCHS384A192GCMKW,
		string(cryptoutilOpenapiModel.A128CBCHS256A192GCMKW): cryptoutilOpenapiModel.A128CBCHS256A192GCMKW,
		string(cryptoutilOpenapiModel.A256CBCHS512A128GCMKW): cryptoutilOpenapiModel.A256CBCHS512A128GCMKW,
		string(cryptoutilOpenapiModel.A192CBCHS384A128GCMKW): cryptoutilOpenapiModel.A192CBCHS384A128GCMKW,
		string(cryptoutilOpenapiModel.A128CBCHS256A128GCMKW): cryptoutilOpenapiModel.A128CBCHS256A128GCMKW,

		string(cryptoutilOpenapiModel.A256CBCHS512Dir): cryptoutilOpenapiModel.A256CBCHS512Dir,
		string(cryptoutilOpenapiModel.A192CBCHS384Dir): cryptoutilOpenapiModel.A192CBCHS384Dir,
		string(cryptoutilOpenapiModel.A128CBCHS256Dir): cryptoutilOpenapiModel.A128CBCHS256Dir,

		string(cryptoutilOpenapiModel.A256CBCHS512RSAOAEP512): cryptoutilOpenapiModel.A256CBCHS512RSAOAEP512,
		string(cryptoutilOpenapiModel.A192CBCHS384RSAOAEP512): cryptoutilOpenapiModel.A192CBCHS384RSAOAEP512,
		string(cryptoutilOpenapiModel.A128CBCHS256RSAOAEP512): cryptoutilOpenapiModel.A128CBCHS256RSAOAEP512,
		string(cryptoutilOpenapiModel.A256CBCHS512RSAOAEP384): cryptoutilOpenapiModel.A256CBCHS512RSAOAEP384,
		string(cryptoutilOpenapiModel.A192CBCHS384RSAOAEP384): cryptoutilOpenapiModel.A192CBCHS384RSAOAEP384,
		string(cryptoutilOpenapiModel.A128CBCHS256RSAOAEP384): cryptoutilOpenapiModel.A128CBCHS256RSAOAEP384,
		string(cryptoutilOpenapiModel.A256CBCHS512RSAOAEP256): cryptoutilOpenapiModel.A256CBCHS512RSAOAEP256,
		string(cryptoutilOpenapiModel.A192CBCHS384RSAOAEP256): cryptoutilOpenapiModel.A192CBCHS384RSAOAEP256,
		string(cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256): cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256,
		string(cryptoutilOpenapiModel.A256CBCHS512RSAOAEP):    cryptoutilOpenapiModel.A256CBCHS512RSAOAEP,
		string(cryptoutilOpenapiModel.A192CBCHS384RSAOAEP):    cryptoutilOpenapiModel.A192CBCHS384RSAOAEP,
		string(cryptoutilOpenapiModel.A128CBCHS256RSAOAEP):    cryptoutilOpenapiModel.A128CBCHS256RSAOAEP,
		string(cryptoutilOpenapiModel.A256CBCHS512RSA15):      cryptoutilOpenapiModel.A256CBCHS512RSA15,
		string(cryptoutilOpenapiModel.A192CBCHS384RSA15):      cryptoutilOpenapiModel.A192CBCHS384RSA15,
		string(cryptoutilOpenapiModel.A128CBCHS256RSA15):      cryptoutilOpenapiModel.A128CBCHS256RSA15,

		string(cryptoutilOpenapiModel.A256CBCHS512ECDHESA256KW): cryptoutilOpenapiModel.A256CBCHS512ECDHESA256KW,
		string(cryptoutilOpenapiModel.A192CBCHS384ECDHESA256KW): cryptoutilOpenapiModel.A192CBCHS384ECDHESA256KW,
		string(cryptoutilOpenapiModel.A128CBCHS256ECDHESA256KW): cryptoutilOpenapiModel.A128CBCHS256ECDHESA256KW,
		string(cryptoutilOpenapiModel.A256CBCHS512ECDHESA192KW): cryptoutilOpenapiModel.A256CBCHS512ECDHESA192KW,
		string(cryptoutilOpenapiModel.A192CBCHS384ECDHESA192KW): cryptoutilOpenapiModel.A192CBCHS384ECDHESA192KW,
		string(cryptoutilOpenapiModel.A128CBCHS256ECDHESA192KW): cryptoutilOpenapiModel.A128CBCHS256ECDHESA192KW,
		string(cryptoutilOpenapiModel.A128CBCHS256ECDHESA128KW): cryptoutilOpenapiModel.A128CBCHS256ECDHESA128KW,
		string(cryptoutilOpenapiModel.A256CBCHS512ECDHES):       cryptoutilOpenapiModel.A256CBCHS512ECDHES,
		string(cryptoutilOpenapiModel.A192CBCHS384ECDHES):       cryptoutilOpenapiModel.A192CBCHS384ECDHES,
		string(cryptoutilOpenapiModel.A128CBCHS256ECDHES):       cryptoutilOpenapiModel.A128CBCHS256ECDHES,

		string(cryptoutilOpenapiModel.RS256): cryptoutilOpenapiModel.RS256,
		string(cryptoutilOpenapiModel.RS384): cryptoutilOpenapiModel.RS384,
		string(cryptoutilOpenapiModel.RS512): cryptoutilOpenapiModel.RS512,
		string(cryptoutilOpenapiModel.PS256): cryptoutilOpenapiModel.PS256,
		string(cryptoutilOpenapiModel.PS384): cryptoutilOpenapiModel.PS384,
		string(cryptoutilOpenapiModel.PS512): cryptoutilOpenapiModel.PS512,
		string(cryptoutilOpenapiModel.ES256): cryptoutilOpenapiModel.ES256,
		string(cryptoutilOpenapiModel.ES384): cryptoutilOpenapiModel.ES384,
		string(cryptoutilOpenapiModel.ES512): cryptoutilOpenapiModel.ES512,
		string(cryptoutilOpenapiModel.HS256): cryptoutilOpenapiModel.HS256,
		string(cryptoutilOpenapiModel.HS384): cryptoutilOpenapiModel.HS384,
		string(cryptoutilOpenapiModel.HS512): cryptoutilOpenapiModel.HS512,
		string(cryptoutilOpenapiModel.EdDSA): cryptoutilOpenapiModel.EdDSA,
	}

	asymmetricElasticKeyAlgorithm = map[cryptoutilOpenapiModel.ElasticKeyAlgorithm]bool{
		cryptoutilOpenapiModel.A256GCMA256KW: false, cryptoutilOpenapiModel.A192GCMA256KW: false, cryptoutilOpenapiModel.A128GCMA256KW: false,
		cryptoutilOpenapiModel.A256GCMA192KW: false, cryptoutilOpenapiModel.A192GCMA192KW: false, cryptoutilOpenapiModel.A128GCMA192KW: false,
		cryptoutilOpenapiModel.A256GCMA128KW: false, cryptoutilOpenapiModel.A192GCMA128KW: false, cryptoutilOpenapiModel.A128GCMA128KW: false,
		cryptoutilOpenapiModel.A256GCMA256GCMKW: false, cryptoutilOpenapiModel.A192GCMA256GCMKW: false, cryptoutilOpenapiModel.A128GCMA256GCMKW: false,
		cryptoutilOpenapiModel.A256GCMA192GCMKW: false, cryptoutilOpenapiModel.A192GCMA192GCMKW: false, cryptoutilOpenapiModel.A128GCMA192GCMKW: false,
		cryptoutilOpenapiModel.A256GCMA128GCMKW: false, cryptoutilOpenapiModel.A192GCMA128GCMKW: false, cryptoutilOpenapiModel.A128GCMA128GCMKW: false,
		cryptoutilOpenapiModel.A256GCMDir: false, cryptoutilOpenapiModel.A192GCMDir: false, cryptoutilOpenapiModel.A128GCMDir: false,

		cryptoutilOpenapiModel.A256GCMRSAOAEP512: true, cryptoutilOpenapiModel.A192GCMRSAOAEP512: true, cryptoutilOpenapiModel.A128GCMRSAOAEP512: true,
		cryptoutilOpenapiModel.A256GCMRSAOAEP384: true, cryptoutilOpenapiModel.A192GCMRSAOAEP384: true, cryptoutilOpenapiModel.A128GCMRSAOAEP384: true,
		cryptoutilOpenapiModel.A256GCMRSAOAEP256: true, cryptoutilOpenapiModel.A192GCMRSAOAEP256: true, cryptoutilOpenapiModel.A128GCMRSAOAEP256: true,
		cryptoutilOpenapiModel.A256GCMRSAOAEP: true, cryptoutilOpenapiModel.A192GCMRSAOAEP: true, cryptoutilOpenapiModel.A128GCMRSAOAEP: true,
		cryptoutilOpenapiModel.A256GCMRSA15: true, cryptoutilOpenapiModel.A192GCMRSA15: true, cryptoutilOpenapiModel.A128GCMRSA15: true,

		cryptoutilOpenapiModel.A256GCMECDHESA256KW: true, cryptoutilOpenapiModel.A192GCMECDHESA256KW: true, cryptoutilOpenapiModel.A128GCMECDHESA256KW: true,
		cryptoutilOpenapiModel.A256GCMECDHESA192KW: true, cryptoutilOpenapiModel.A192GCMECDHESA192KW: true, cryptoutilOpenapiModel.A128GCMECDHESA192KW: true,
		cryptoutilOpenapiModel.A256GCMECDHESA128KW: true, cryptoutilOpenapiModel.A192GCMECDHESA128KW: true, cryptoutilOpenapiModel.A128GCMECDHESA128KW: true,
		cryptoutilOpenapiModel.A256GCMECDHES: true, cryptoutilOpenapiModel.A192GCMECDHES: true, cryptoutilOpenapiModel.A128GCMECDHES: true,

		cryptoutilOpenapiModel.A256CBCHS512A256KW: false, cryptoutilOpenapiModel.A192CBCHS384A256KW: false, cryptoutilOpenapiModel.A128CBCHS256A256KW: false,
		cryptoutilOpenapiModel.A256CBCHS512A192KW: false, cryptoutilOpenapiModel.A192CBCHS384A192KW: false, cryptoutilOpenapiModel.A128CBCHS256A192KW: false,
		cryptoutilOpenapiModel.A256CBCHS512A128KW: false, cryptoutilOpenapiModel.A192CBCHS384A128KW: false, cryptoutilOpenapiModel.A128CBCHS256A128KW: false,
		cryptoutilOpenapiModel.A256CBCHS512A256GCMKW: false, cryptoutilOpenapiModel.A192CBCHS384A256GCMKW: false, cryptoutilOpenapiModel.A128CBCHS256A256GCMKW: false,
		cryptoutilOpenapiModel.A256CBCHS512A192GCMKW: false, cryptoutilOpenapiModel.A192CBCHS384A192GCMKW: false, cryptoutilOpenapiModel.A128CBCHS256A192GCMKW: false,
		cryptoutilOpenapiModel.A256CBCHS512A128GCMKW: false, cryptoutilOpenapiModel.A192CBCHS384A128GCMKW: false, cryptoutilOpenapiModel.A128CBCHS256A128GCMKW: false,
		cryptoutilOpenapiModel.A256CBCHS512Dir: false, cryptoutilOpenapiModel.A192CBCHS384Dir: false, cryptoutilOpenapiModel.A128CBCHS256Dir: false,

		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP512: true, cryptoutilOpenapiModel.A192CBCHS384RSAOAEP512: true, cryptoutilOpenapiModel.A128CBCHS256RSAOAEP512: true,
		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP384: true, cryptoutilOpenapiModel.A192CBCHS384RSAOAEP384: true, cryptoutilOpenapiModel.A128CBCHS256RSAOAEP384: true,
		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP256: true, cryptoutilOpenapiModel.A192CBCHS384RSAOAEP256: true, cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256: true,
		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP: true, cryptoutilOpenapiModel.A192CBCHS384RSAOAEP: true, cryptoutilOpenapiModel.A128CBCHS256RSAOAEP: true,
		cryptoutilOpenapiModel.A256CBCHS512RSA15: true, cryptoutilOpenapiModel.A192CBCHS384RSA15: true, cryptoutilOpenapiModel.A128CBCHS256RSA15: true,

		cryptoutilOpenapiModel.A256CBCHS512ECDHESA256KW: true, cryptoutilOpenapiModel.A192CBCHS384ECDHESA256KW: true, cryptoutilOpenapiModel.A128CBCHS256ECDHESA256KW: true,
		cryptoutilOpenapiModel.A256CBCHS512ECDHESA192KW: true, cryptoutilOpenapiModel.A192CBCHS384ECDHESA192KW: true, cryptoutilOpenapiModel.A128CBCHS256ECDHESA192KW: true,
		cryptoutilOpenapiModel.A256CBCHS512ECDHESA128KW: true, cryptoutilOpenapiModel.A192CBCHS384ECDHESA128KW: true, cryptoutilOpenapiModel.A128CBCHS256ECDHESA128KW: true,
		cryptoutilOpenapiModel.A256CBCHS512ECDHES: true, cryptoutilOpenapiModel.A192CBCHS384ECDHES: true, cryptoutilOpenapiModel.A128CBCHS256ECDHES: true,

		cryptoutilOpenapiModel.RS512: true, cryptoutilOpenapiModel.RS384: true, cryptoutilOpenapiModel.RS256: true,
		cryptoutilOpenapiModel.PS512: true, cryptoutilOpenapiModel.PS384: true, cryptoutilOpenapiModel.PS256: true,
		cryptoutilOpenapiModel.ES512: true, cryptoutilOpenapiModel.ES384: true, cryptoutilOpenapiModel.ES256: true,
		cryptoutilOpenapiModel.HS512: false, cryptoutilOpenapiModel.HS384: false, cryptoutilOpenapiModel.HS256: false,
		cryptoutilOpenapiModel.EdDSA: true,
	}

	symmetricElasticKeyAlgorithm = map[cryptoutilOpenapiModel.ElasticKeyAlgorithm]bool{
		cryptoutilOpenapiModel.A256GCMA256KW: true, cryptoutilOpenapiModel.A192GCMA256KW: true, cryptoutilOpenapiModel.A128GCMA256KW: true,
		cryptoutilOpenapiModel.A256GCMA192KW: true, cryptoutilOpenapiModel.A192GCMA192KW: true, cryptoutilOpenapiModel.A128GCMA192KW: true,
		cryptoutilOpenapiModel.A256GCMA128KW: true, cryptoutilOpenapiModel.A192GCMA128KW: true, cryptoutilOpenapiModel.A128GCMA128KW: true,
		cryptoutilOpenapiModel.A256GCMA256GCMKW: true, cryptoutilOpenapiModel.A192GCMA256GCMKW: true, cryptoutilOpenapiModel.A128GCMA256GCMKW: true,
		cryptoutilOpenapiModel.A256GCMA192GCMKW: true, cryptoutilOpenapiModel.A192GCMA192GCMKW: true, cryptoutilOpenapiModel.A128GCMA192GCMKW: true,
		cryptoutilOpenapiModel.A256GCMA128GCMKW: true, cryptoutilOpenapiModel.A192GCMA128GCMKW: true, cryptoutilOpenapiModel.A128GCMA128GCMKW: true,
		cryptoutilOpenapiModel.A256GCMDir: true, cryptoutilOpenapiModel.A192GCMDir: true, cryptoutilOpenapiModel.A128GCMDir: true,

		cryptoutilOpenapiModel.A256GCMRSAOAEP512: false, cryptoutilOpenapiModel.A192GCMRSAOAEP512: false, cryptoutilOpenapiModel.A128GCMRSAOAEP512: false,
		cryptoutilOpenapiModel.A256GCMRSAOAEP384: false, cryptoutilOpenapiModel.A192GCMRSAOAEP384: false, cryptoutilOpenapiModel.A128GCMRSAOAEP384: false,
		cryptoutilOpenapiModel.A256GCMRSAOAEP256: false, cryptoutilOpenapiModel.A192GCMRSAOAEP256: false, cryptoutilOpenapiModel.A128GCMRSAOAEP256: false,
		cryptoutilOpenapiModel.A256GCMRSAOAEP: false, cryptoutilOpenapiModel.A192GCMRSAOAEP: false, cryptoutilOpenapiModel.A128GCMRSAOAEP: false,
		cryptoutilOpenapiModel.A256GCMRSA15: false, cryptoutilOpenapiModel.A192GCMRSA15: false, cryptoutilOpenapiModel.A128GCMRSA15: false,

		cryptoutilOpenapiModel.A256GCMECDHESA256KW: false, cryptoutilOpenapiModel.A192GCMECDHESA256KW: false, cryptoutilOpenapiModel.A128GCMECDHESA256KW: false,
		cryptoutilOpenapiModel.A256GCMECDHESA192KW: false, cryptoutilOpenapiModel.A192GCMECDHESA192KW: false, cryptoutilOpenapiModel.A128GCMECDHESA192KW: false,
		cryptoutilOpenapiModel.A256GCMECDHESA128KW: false, cryptoutilOpenapiModel.A192GCMECDHESA128KW: false, cryptoutilOpenapiModel.A128GCMECDHESA128KW: false,
		cryptoutilOpenapiModel.A256GCMECDHES: false, cryptoutilOpenapiModel.A192GCMECDHES: false, cryptoutilOpenapiModel.A128GCMECDHES: false,

		cryptoutilOpenapiModel.A256CBCHS512A256KW: true, cryptoutilOpenapiModel.A192CBCHS384A256KW: true, cryptoutilOpenapiModel.A128CBCHS256A256KW: true,
		cryptoutilOpenapiModel.A256CBCHS512A192KW: true, cryptoutilOpenapiModel.A192CBCHS384A192KW: true, cryptoutilOpenapiModel.A128CBCHS256A192KW: true,
		cryptoutilOpenapiModel.A256CBCHS512A128KW: true, cryptoutilOpenapiModel.A192CBCHS384A128KW: true, cryptoutilOpenapiModel.A128CBCHS256A128KW: true,
		cryptoutilOpenapiModel.A256CBCHS512A256GCMKW: true, cryptoutilOpenapiModel.A192CBCHS384A256GCMKW: true, cryptoutilOpenapiModel.A128CBCHS256A256GCMKW: true,
		cryptoutilOpenapiModel.A256CBCHS512A192GCMKW: true, cryptoutilOpenapiModel.A192CBCHS384A192GCMKW: true, cryptoutilOpenapiModel.A128CBCHS256A192GCMKW: true,
		cryptoutilOpenapiModel.A256CBCHS512A128GCMKW: true, cryptoutilOpenapiModel.A192CBCHS384A128GCMKW: true, cryptoutilOpenapiModel.A128CBCHS256A128GCMKW: true,
		cryptoutilOpenapiModel.A256CBCHS512Dir: true, cryptoutilOpenapiModel.A192CBCHS384Dir: true, cryptoutilOpenapiModel.A128CBCHS256Dir: true,

		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP512: false, cryptoutilOpenapiModel.A192CBCHS384RSAOAEP512: false, cryptoutilOpenapiModel.A128CBCHS256RSAOAEP512: false,
		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP384: false, cryptoutilOpenapiModel.A192CBCHS384RSAOAEP384: false, cryptoutilOpenapiModel.A128CBCHS256RSAOAEP384: false,
		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP256: false, cryptoutilOpenapiModel.A192CBCHS384RSAOAEP256: false, cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256: false,
		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP: false, cryptoutilOpenapiModel.A192CBCHS384RSAOAEP: false, cryptoutilOpenapiModel.A128CBCHS256RSAOAEP: false,
		cryptoutilOpenapiModel.A256CBCHS512RSA15: false, cryptoutilOpenapiModel.A192CBCHS384RSA15: false, cryptoutilOpenapiModel.A128CBCHS256RSA15: false,

		cryptoutilOpenapiModel.A256CBCHS512ECDHESA256KW: false, cryptoutilOpenapiModel.A192CBCHS384ECDHESA256KW: false, cryptoutilOpenapiModel.A128CBCHS256ECDHESA256KW: false,
		cryptoutilOpenapiModel.A256CBCHS512ECDHESA192KW: false, cryptoutilOpenapiModel.A192CBCHS384ECDHESA192KW: false, cryptoutilOpenapiModel.A128CBCHS256ECDHESA192KW: false,
		cryptoutilOpenapiModel.A256CBCHS512ECDHESA128KW: false, cryptoutilOpenapiModel.A192CBCHS384ECDHESA128KW: false, cryptoutilOpenapiModel.A128CBCHS256ECDHESA128KW: false,
		cryptoutilOpenapiModel.A256CBCHS512ECDHES: false, cryptoutilOpenapiModel.A192CBCHS384ECDHES: false, cryptoutilOpenapiModel.A128CBCHS256ECDHES: false,

		cryptoutilOpenapiModel.RS512: false, cryptoutilOpenapiModel.RS384: false, cryptoutilOpenapiModel.RS256: false,
		cryptoutilOpenapiModel.PS512: false, cryptoutilOpenapiModel.PS384: false, cryptoutilOpenapiModel.PS256: false,
		cryptoutilOpenapiModel.ES512: false, cryptoutilOpenapiModel.ES384: false, cryptoutilOpenapiModel.ES256: false,
		cryptoutilOpenapiModel.HS512: true, cryptoutilOpenapiModel.HS384: true, cryptoutilOpenapiModel.HS256: true,
		cryptoutilOpenapiModel.EdDSA: false,
	}

	elasticKeyAlgorithmToJoseEncAndAlg = map[cryptoutilOpenapiModel.ElasticKeyAlgorithm]struct {
		enc *joseJwa.ContentEncryptionAlgorithm
		alg *joseJwa.KeyEncryptionAlgorithm
	}{
		cryptoutilOpenapiModel.A256GCMA256KW:    {enc: &EncA256GCM, alg: &AlgA256KW},
		cryptoutilOpenapiModel.A192GCMA256KW:    {enc: &EncA192GCM, alg: &AlgA256KW},
		cryptoutilOpenapiModel.A128GCMA256KW:    {enc: &EncA128GCM, alg: &AlgA256KW},
		cryptoutilOpenapiModel.A256GCMA192KW:    {enc: &EncA256GCM, alg: &AlgA192KW},
		cryptoutilOpenapiModel.A192GCMA192KW:    {enc: &EncA192GCM, alg: &AlgA192KW},
		cryptoutilOpenapiModel.A128GCMA192KW:    {enc: &EncA128GCM, alg: &AlgA192KW},
		cryptoutilOpenapiModel.A256GCMA128KW:    {enc: &EncA256GCM, alg: &AlgA128KW},
		cryptoutilOpenapiModel.A192GCMA128KW:    {enc: &EncA192GCM, alg: &AlgA128KW},
		cryptoutilOpenapiModel.A128GCMA128KW:    {enc: &EncA128GCM, alg: &AlgA128KW},
		cryptoutilOpenapiModel.A256GCMA256GCMKW: {enc: &EncA256GCM, alg: &AlgA256GCMKW},
		cryptoutilOpenapiModel.A192GCMA256GCMKW: {enc: &EncA192GCM, alg: &AlgA256GCMKW},
		cryptoutilOpenapiModel.A128GCMA256GCMKW: {enc: &EncA128GCM, alg: &AlgA256GCMKW},
		cryptoutilOpenapiModel.A256GCMA192GCMKW: {enc: &EncA256GCM, alg: &AlgA192GCMKW},
		cryptoutilOpenapiModel.A192GCMA192GCMKW: {enc: &EncA192GCM, alg: &AlgA192GCMKW},
		cryptoutilOpenapiModel.A128GCMA192GCMKW: {enc: &EncA128GCM, alg: &AlgA192GCMKW},
		cryptoutilOpenapiModel.A256GCMA128GCMKW: {enc: &EncA256GCM, alg: &AlgA128GCMKW},
		cryptoutilOpenapiModel.A192GCMA128GCMKW: {enc: &EncA192GCM, alg: &AlgA128GCMKW},
		cryptoutilOpenapiModel.A128GCMA128GCMKW: {enc: &EncA128GCM, alg: &AlgA128GCMKW},
		cryptoutilOpenapiModel.A256GCMDir:       {enc: &EncA256GCM, alg: &AlgDir},
		cryptoutilOpenapiModel.A192GCMDir:       {enc: &EncA192GCM, alg: &AlgDir},
		cryptoutilOpenapiModel.A128GCMDir:       {enc: &EncA128GCM, alg: &AlgDir},

		cryptoutilOpenapiModel.A256GCMRSAOAEP512: {enc: &EncA256GCM, alg: &AlgRSAOAEP512},
		cryptoutilOpenapiModel.A192GCMRSAOAEP512: {enc: &EncA192GCM, alg: &AlgRSAOAEP512},
		cryptoutilOpenapiModel.A128GCMRSAOAEP512: {enc: &EncA128GCM, alg: &AlgRSAOAEP512},
		cryptoutilOpenapiModel.A256GCMRSAOAEP384: {enc: &EncA256GCM, alg: &AlgRSAOAEP384},
		cryptoutilOpenapiModel.A192GCMRSAOAEP384: {enc: &EncA192GCM, alg: &AlgRSAOAEP384},
		cryptoutilOpenapiModel.A128GCMRSAOAEP384: {enc: &EncA128GCM, alg: &AlgRSAOAEP384},
		cryptoutilOpenapiModel.A256GCMRSAOAEP256: {enc: &EncA256GCM, alg: &AlgRSAOAEP256},
		cryptoutilOpenapiModel.A192GCMRSAOAEP256: {enc: &EncA192GCM, alg: &AlgRSAOAEP256},
		cryptoutilOpenapiModel.A128GCMRSAOAEP256: {enc: &EncA128GCM, alg: &AlgRSAOAEP256},
		cryptoutilOpenapiModel.A256GCMRSAOAEP:    {enc: &EncA256GCM, alg: &AlgRSAOAEP},
		cryptoutilOpenapiModel.A192GCMRSAOAEP:    {enc: &EncA192GCM, alg: &AlgRSAOAEP},
		cryptoutilOpenapiModel.A128GCMRSAOAEP:    {enc: &EncA128GCM, alg: &AlgRSAOAEP},
		cryptoutilOpenapiModel.A256GCMRSA15:      {enc: &EncA256GCM, alg: &AlgRSA15},
		cryptoutilOpenapiModel.A192GCMRSA15:      {enc: &EncA192GCM, alg: &AlgRSA15},
		cryptoutilOpenapiModel.A128GCMRSA15:      {enc: &EncA128GCM, alg: &AlgRSA15},

		cryptoutilOpenapiModel.A256GCMECDHESA256KW: {enc: &EncA256GCM, alg: &AlgECDHESA256KW},
		cryptoutilOpenapiModel.A192GCMECDHESA256KW: {enc: &EncA192GCM, alg: &AlgECDHESA256KW},
		cryptoutilOpenapiModel.A128GCMECDHESA256KW: {enc: &EncA128GCM, alg: &AlgECDHESA256KW},
		cryptoutilOpenapiModel.A256GCMECDHESA192KW: {enc: &EncA256GCM, alg: &AlgECDHESA192KW},
		cryptoutilOpenapiModel.A192GCMECDHESA192KW: {enc: &EncA192GCM, alg: &AlgECDHESA192KW},
		cryptoutilOpenapiModel.A128GCMECDHESA192KW: {enc: &EncA128GCM, alg: &AlgECDHESA192KW},
		cryptoutilOpenapiModel.A256GCMECDHESA128KW: {enc: &EncA256GCM, alg: &AlgECDHESA128KW},
		cryptoutilOpenapiModel.A192GCMECDHESA128KW: {enc: &EncA192GCM, alg: &AlgECDHESA128KW},
		cryptoutilOpenapiModel.A128GCMECDHESA128KW: {enc: &EncA128GCM, alg: &AlgECDHESA128KW},
		cryptoutilOpenapiModel.A256GCMECDHES:       {enc: &EncA256GCM, alg: &AlgECDHES},
		cryptoutilOpenapiModel.A192GCMECDHES:       {enc: &EncA192GCM, alg: &AlgECDHES},
		cryptoutilOpenapiModel.A128GCMECDHES:       {enc: &EncA128GCM, alg: &AlgECDHES},

		cryptoutilOpenapiModel.A256CBCHS512A256KW:    {enc: &EncA256CBCHS512, alg: &AlgA256KW},
		cryptoutilOpenapiModel.A192CBCHS384A256KW:    {enc: &EncA192CBCHS384, alg: &AlgA256KW},
		cryptoutilOpenapiModel.A128CBCHS256A256KW:    {enc: &EncA128CBCHS256, alg: &AlgA256KW},
		cryptoutilOpenapiModel.A256CBCHS512A192KW:    {enc: &EncA256CBCHS512, alg: &AlgA192KW},
		cryptoutilOpenapiModel.A192CBCHS384A192KW:    {enc: &EncA192CBCHS384, alg: &AlgA192KW},
		cryptoutilOpenapiModel.A128CBCHS256A192KW:    {enc: &EncA128CBCHS256, alg: &AlgA192KW},
		cryptoutilOpenapiModel.A256CBCHS512A128KW:    {enc: &EncA256CBCHS512, alg: &AlgA128KW},
		cryptoutilOpenapiModel.A192CBCHS384A128KW:    {enc: &EncA192CBCHS384, alg: &AlgA128KW},
		cryptoutilOpenapiModel.A128CBCHS256A128KW:    {enc: &EncA128CBCHS256, alg: &AlgA128KW},
		cryptoutilOpenapiModel.A256CBCHS512A256GCMKW: {enc: &EncA256CBCHS512, alg: &AlgA256GCMKW},
		cryptoutilOpenapiModel.A192CBCHS384A256GCMKW: {enc: &EncA192CBCHS384, alg: &AlgA256GCMKW},
		cryptoutilOpenapiModel.A128CBCHS256A256GCMKW: {enc: &EncA128CBCHS256, alg: &AlgA256GCMKW},
		cryptoutilOpenapiModel.A256CBCHS512A192GCMKW: {enc: &EncA256CBCHS512, alg: &AlgA192GCMKW},
		cryptoutilOpenapiModel.A192CBCHS384A192GCMKW: {enc: &EncA192CBCHS384, alg: &AlgA192GCMKW},
		cryptoutilOpenapiModel.A128CBCHS256A192GCMKW: {enc: &EncA128CBCHS256, alg: &AlgA192GCMKW},
		cryptoutilOpenapiModel.A256CBCHS512A128GCMKW: {enc: &EncA256CBCHS512, alg: &AlgA128GCMKW},
		cryptoutilOpenapiModel.A192CBCHS384A128GCMKW: {enc: &EncA192CBCHS384, alg: &AlgA128GCMKW},
		cryptoutilOpenapiModel.A128CBCHS256A128GCMKW: {enc: &EncA128CBCHS256, alg: &AlgA128GCMKW},
		cryptoutilOpenapiModel.A256CBCHS512Dir:       {enc: &EncA256CBCHS512, alg: &AlgDir},
		cryptoutilOpenapiModel.A192CBCHS384Dir:       {enc: &EncA192CBCHS384, alg: &AlgDir},
		cryptoutilOpenapiModel.A128CBCHS256Dir:       {enc: &EncA128CBCHS256, alg: &AlgDir},

		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP512: {enc: &EncA256CBCHS512, alg: &AlgRSAOAEP512},
		cryptoutilOpenapiModel.A192CBCHS384RSAOAEP512: {enc: &EncA192CBCHS384, alg: &AlgRSAOAEP512},
		cryptoutilOpenapiModel.A128CBCHS256RSAOAEP512: {enc: &EncA128CBCHS256, alg: &AlgRSAOAEP512},
		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP384: {enc: &EncA256CBCHS512, alg: &AlgRSAOAEP384},
		cryptoutilOpenapiModel.A192CBCHS384RSAOAEP384: {enc: &EncA192CBCHS384, alg: &AlgRSAOAEP384},
		cryptoutilOpenapiModel.A128CBCHS256RSAOAEP384: {enc: &EncA128CBCHS256, alg: &AlgRSAOAEP384},
		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP256: {enc: &EncA256CBCHS512, alg: &AlgRSAOAEP256},
		cryptoutilOpenapiModel.A192CBCHS384RSAOAEP256: {enc: &EncA192CBCHS384, alg: &AlgRSAOAEP256},
		cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256: {enc: &EncA128CBCHS256, alg: &AlgRSAOAEP256},
		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP:    {enc: &EncA256CBCHS512, alg: &AlgRSAOAEP},
		cryptoutilOpenapiModel.A192CBCHS384RSAOAEP:    {enc: &EncA192CBCHS384, alg: &AlgRSAOAEP},
		cryptoutilOpenapiModel.A128CBCHS256RSAOAEP:    {enc: &EncA128CBCHS256, alg: &AlgRSAOAEP},
		cryptoutilOpenapiModel.A256CBCHS512RSA15:      {enc: &EncA256CBCHS512, alg: &AlgRSA15},
		cryptoutilOpenapiModel.A192CBCHS384RSA15:      {enc: &EncA192CBCHS384, alg: &AlgRSA15},
		cryptoutilOpenapiModel.A128CBCHS256RSA15:      {enc: &EncA128CBCHS256, alg: &AlgRSA15},

		cryptoutilOpenapiModel.A256CBCHS512ECDHESA256KW: {enc: &EncA256CBCHS512, alg: &AlgECDHESA256KW},
		cryptoutilOpenapiModel.A192CBCHS384ECDHESA256KW: {enc: &EncA192CBCHS384, alg: &AlgECDHESA256KW},
		cryptoutilOpenapiModel.A128CBCHS256ECDHESA256KW: {enc: &EncA128CBCHS256, alg: &AlgECDHESA256KW},
		cryptoutilOpenapiModel.A192CBCHS384ECDHESA192KW: {enc: &EncA192CBCHS384, alg: &AlgECDHESA192KW},
		cryptoutilOpenapiModel.A128CBCHS256ECDHESA192KW: {enc: &EncA128CBCHS256, alg: &AlgECDHESA192KW},
		cryptoutilOpenapiModel.A128CBCHS256ECDHESA128KW: {enc: &EncA128CBCHS256, alg: &AlgECDHESA128KW},
		cryptoutilOpenapiModel.A256CBCHS512ECDHES:       {enc: &EncA256CBCHS512, alg: &AlgECDHES},
		cryptoutilOpenapiModel.A192CBCHS384ECDHES:       {enc: &EncA192CBCHS384, alg: &AlgECDHES},
		cryptoutilOpenapiModel.A128CBCHS256ECDHES:       {enc: &EncA128CBCHS256, alg: &AlgECDHES},
	}

	elasticKeyAlgorithmToJoseAlg = map[cryptoutilOpenapiModel.ElasticKeyAlgorithm]*joseJwa.SignatureAlgorithm{
		cryptoutilOpenapiModel.RS512: &AlgRS512,
		cryptoutilOpenapiModel.RS384: &AlgRS384,
		cryptoutilOpenapiModel.RS256: &AlgRS256,
		cryptoutilOpenapiModel.PS512: &AlgPS512,
		cryptoutilOpenapiModel.PS384: &AlgPS384,
		cryptoutilOpenapiModel.PS256: &AlgPS256,
		cryptoutilOpenapiModel.ES512: &AlgES512,
		cryptoutilOpenapiModel.ES384: &AlgES384,
		cryptoutilOpenapiModel.ES256: &AlgES256,
		cryptoutilOpenapiModel.HS512: &AlgHS512,
		cryptoutilOpenapiModel.HS384: &AlgHS384,
		cryptoutilOpenapiModel.HS256: &AlgHS256,
		cryptoutilOpenapiModel.EdDSA: &AlgEdDSA,
	}
)

// ToJWEEncAndAlg converts an ElasticKeyAlgorithm to JWE encryption and key algorithms.
func ToJWEEncAndAlg(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) (*joseJwa.ContentEncryptionAlgorithm, *joseJwa.KeyEncryptionAlgorithm, error) {
	if encAndAlg, ok := elasticKeyAlgorithmToJoseEncAndAlg[*elasticKeyAlgorithm]; ok {
		return encAndAlg.enc, encAndAlg.alg, nil
	}

	return nil, nil, fmt.Errorf("unsupported JWE ElasticKeyAlgorithm '%s'", *elasticKeyAlgorithm)
}

// ToJWSAlg converts an ElasticKeyAlgorithm to a JWS signature algorithm.
func ToJWSAlg(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) (*joseJwa.SignatureAlgorithm, error) {
	if alg, ok := elasticKeyAlgorithmToJoseAlg[*elasticKeyAlgorithm]; ok {
		return alg, nil
	}

	return nil, fmt.Errorf("unsupported JWS ElasticKeyAlgorithm '%s'", *elasticKeyAlgorithm)
}

// IsJWE returns true if the algorithm is a JWE encryption algorithm.
func IsJWE(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) bool {
	_, ok := elasticKeyAlgorithmToJoseEncAndAlg[*elasticKeyAlgorithm]

	return ok
}

// IsJWS returns true if the algorithm is a JWS signature algorithm.
func IsJWS(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) bool {
	_, ok := elasticKeyAlgorithmToJoseAlg[*elasticKeyAlgorithm]

	return ok
}

// IsSymmetric returns true if the algorithm uses symmetric keys.
func IsSymmetric(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) (bool, error) {
	isSymmetric, ok := symmetricElasticKeyAlgorithm[*elasticKeyAlgorithm]
	if ok {
		return isSymmetric, nil
	}

	return false, fmt.Errorf("unsupported ElasticKeyAlgorithm '%s'", *elasticKeyAlgorithm)
}

// IsAsymmetric returns true if the algorithm uses asymmetric keys.
func IsAsymmetric(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) (bool, error) {
	isAsymmetric, ok := asymmetricElasticKeyAlgorithm[*elasticKeyAlgorithm]
	if ok {
		return isAsymmetric, nil
	}

	return false, fmt.Errorf("unsupported ElasticKeyAlgorithm '%s'", *elasticKeyAlgorithm)
}

// ToElasticKeyAlgorithm converts a string to an ElasticKeyAlgorithm.
func ToElasticKeyAlgorithm(algorithm *string) (*cryptoutilOpenapiModel.ElasticKeyAlgorithm, error) {
	if alg, exists := elasticKeyAlgorithms[*algorithm]; exists {
		return &alg, nil
	}

	return nil, fmt.Errorf("invalid elastic Key algorithm: %v", algorithm)
}

// ToGenerateAlgorithm converts a string to a GenerateAlgorithm.
func ToGenerateAlgorithm(algorithm *string) (*cryptoutilOpenapiModel.GenerateAlgorithm, error) {
	if alg, exists := generateAlgorithms[*algorithm]; exists {
		return &alg, nil
	}

	return nil, fmt.Errorf("invalid generate algorithm: %v", algorithm)
}

// GetGenerateAlgorithmTestProbability returns the execution probability for table-driven tests.
// Different key sizes of the same algorithm type can use lower probabilities to reduce test time.
// Base algorithms (e.g., RSA2048, ECP256, Oct256) use TestProbAlways for comprehensive coverage.
// Larger variants (e.g., RSA4096, ECP521, Oct512) use TestProbThird for sampling coverage.
func GetGenerateAlgorithmTestProbability(alg cryptoutilOpenapiModel.GenerateAlgorithm) float64 {
	switch alg {
	// Base RSA size - always test.
	case cryptoutilOpenapiModel.RSA2048:
		return cryptoutilMagic.TestProbAlways
	// Larger RSA sizes - sample testing.
	case cryptoutilOpenapiModel.RSA3072, cryptoutilOpenapiModel.RSA4096:
		return cryptoutilMagic.TestProbThird
	// Base EC size - always test.
	case cryptoutilOpenapiModel.ECP256:
		return cryptoutilMagic.TestProbAlways
	// Larger EC sizes - sample testing.
	case cryptoutilOpenapiModel.ECP384, cryptoutilOpenapiModel.ECP521:
		return cryptoutilMagic.TestProbThird
	// EdDSA - always test (only one size).
	case cryptoutilOpenapiModel.OKPEd25519:
		return cryptoutilMagic.TestProbAlways
	// Base symmetric key size - always test.
	case cryptoutilOpenapiModel.Oct256:
		return cryptoutilMagic.TestProbAlways
	// Other symmetric sizes - sample testing.
	case cryptoutilOpenapiModel.Oct128, cryptoutilOpenapiModel.Oct192, cryptoutilOpenapiModel.Oct384, cryptoutilOpenapiModel.Oct512:
		return cryptoutilMagic.TestProbThird
	default:
		return cryptoutilMagic.TestProbAlways
	}
}

// GetElasticKeyAlgorithmTestProbability returns the execution probability for table-driven tests.
// Encryption/signing algorithms with multiple key sizes use lower probabilities for variants.
// Base algorithms (e.g., A256GCM, RS256, ES256) use TestProbAlways for comprehensive coverage.
// Variants with different key sizes use TestProbQuarter for sampling coverage.
func GetElasticKeyAlgorithmTestProbability(alg cryptoutilOpenapiModel.ElasticKeyAlgorithm) float64 {
	switch alg {
	// Base AES-GCM + Key Wrap combinations - always test 256-bit.
	case cryptoutilOpenapiModel.A256GCMA256KW, cryptoutilOpenapiModel.A256GCMA256GCMKW, cryptoutilOpenapiModel.A256GCMDir:
		return cryptoutilMagic.TestProbAlways
	// Other AES-GCM + Key Wrap - sample testing.
	case cryptoutilOpenapiModel.A192GCMA256KW, cryptoutilOpenapiModel.A128GCMA256KW,
		cryptoutilOpenapiModel.A256GCMA192KW, cryptoutilOpenapiModel.A192GCMA192KW, cryptoutilOpenapiModel.A128GCMA192KW,
		cryptoutilOpenapiModel.A256GCMA128KW, cryptoutilOpenapiModel.A192GCMA128KW, cryptoutilOpenapiModel.A128GCMA128KW,
		cryptoutilOpenapiModel.A192GCMA256GCMKW, cryptoutilOpenapiModel.A128GCMA256GCMKW,
		cryptoutilOpenapiModel.A256GCMA192GCMKW, cryptoutilOpenapiModel.A192GCMA192GCMKW, cryptoutilOpenapiModel.A128GCMA192GCMKW,
		cryptoutilOpenapiModel.A256GCMA128GCMKW, cryptoutilOpenapiModel.A192GCMA128GCMKW, cryptoutilOpenapiModel.A128GCMA128GCMKW,
		cryptoutilOpenapiModel.A192GCMDir, cryptoutilOpenapiModel.A128GCMDir:
		return cryptoutilMagic.TestProbQuarter
	// Base RSA OAEP - always test.
	case cryptoutilOpenapiModel.A256GCMRSAOAEP256, cryptoutilOpenapiModel.A256CBCHS512RSAOAEP256:
		return cryptoutilMagic.TestProbAlways
	// Other RSA OAEP variants - sample testing.
	case cryptoutilOpenapiModel.A192GCMRSAOAEP512, cryptoutilOpenapiModel.A128GCMRSAOAEP512,
		cryptoutilOpenapiModel.A256GCMRSAOAEP384, cryptoutilOpenapiModel.A192GCMRSAOAEP384, cryptoutilOpenapiModel.A128GCMRSAOAEP384,
		cryptoutilOpenapiModel.A192GCMRSAOAEP256, cryptoutilOpenapiModel.A128GCMRSAOAEP256,
		cryptoutilOpenapiModel.A256GCMRSAOAEP, cryptoutilOpenapiModel.A192GCMRSAOAEP, cryptoutilOpenapiModel.A128GCMRSAOAEP,
		cryptoutilOpenapiModel.A256GCMRSA15, cryptoutilOpenapiModel.A192GCMRSA15, cryptoutilOpenapiModel.A128GCMRSA15,
		cryptoutilOpenapiModel.A192CBCHS384RSAOAEP512, cryptoutilOpenapiModel.A128CBCHS256RSAOAEP512,
		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP384, cryptoutilOpenapiModel.A192CBCHS384RSAOAEP384, cryptoutilOpenapiModel.A128CBCHS256RSAOAEP384,
		cryptoutilOpenapiModel.A192CBCHS384RSAOAEP256, cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256,
		cryptoutilOpenapiModel.A256CBCHS512RSAOAEP, cryptoutilOpenapiModel.A192CBCHS384RSAOAEP, cryptoutilOpenapiModel.A128CBCHS256RSAOAEP,
		cryptoutilOpenapiModel.A256CBCHS512RSA15, cryptoutilOpenapiModel.A192CBCHS384RSA15, cryptoutilOpenapiModel.A128CBCHS256RSA15:
		return cryptoutilMagic.TestProbQuarter
	// Base ECDH-ES - always test.
	case cryptoutilOpenapiModel.A256GCMECDHESA256KW, cryptoutilOpenapiModel.A256CBCHS512ECDHESA256KW:
		return cryptoutilMagic.TestProbAlways
	// Other ECDH-ES variants - sample testing.
	case cryptoutilOpenapiModel.A192GCMECDHESA256KW, cryptoutilOpenapiModel.A128GCMECDHESA256KW,
		cryptoutilOpenapiModel.A256GCMECDHESA192KW, cryptoutilOpenapiModel.A192GCMECDHESA192KW, cryptoutilOpenapiModel.A128GCMECDHESA192KW,
		cryptoutilOpenapiModel.A256GCMECDHESA128KW, cryptoutilOpenapiModel.A192GCMECDHESA128KW, cryptoutilOpenapiModel.A128GCMECDHESA128KW,
		cryptoutilOpenapiModel.A256GCMECDHES, cryptoutilOpenapiModel.A192GCMECDHES, cryptoutilOpenapiModel.A128GCMECDHES,
		cryptoutilOpenapiModel.A192CBCHS384ECDHESA256KW, cryptoutilOpenapiModel.A128CBCHS256ECDHESA256KW,
		cryptoutilOpenapiModel.A256CBCHS512ECDHESA192KW, cryptoutilOpenapiModel.A192CBCHS384ECDHESA192KW, cryptoutilOpenapiModel.A128CBCHS256ECDHESA192KW,
		cryptoutilOpenapiModel.A128CBCHS256ECDHESA128KW,
		cryptoutilOpenapiModel.A256CBCHS512ECDHES, cryptoutilOpenapiModel.A192CBCHS384ECDHES, cryptoutilOpenapiModel.A128CBCHS256ECDHES:
		return cryptoutilMagic.TestProbQuarter
	// Base AES-CBC-HMAC + Key Wrap - always test 256-bit.
	case cryptoutilOpenapiModel.A256CBCHS512A256KW, cryptoutilOpenapiModel.A256CBCHS512A256GCMKW, cryptoutilOpenapiModel.A256CBCHS512Dir:
		return cryptoutilMagic.TestProbAlways
	// Other AES-CBC-HMAC + Key Wrap - sample testing.
	case cryptoutilOpenapiModel.A192CBCHS384A256KW, cryptoutilOpenapiModel.A128CBCHS256A256KW,
		cryptoutilOpenapiModel.A256CBCHS512A192KW, cryptoutilOpenapiModel.A192CBCHS384A192KW, cryptoutilOpenapiModel.A128CBCHS256A192KW,
		cryptoutilOpenapiModel.A256CBCHS512A128KW, cryptoutilOpenapiModel.A192CBCHS384A128KW, cryptoutilOpenapiModel.A128CBCHS256A128KW,
		cryptoutilOpenapiModel.A192CBCHS384A256GCMKW, cryptoutilOpenapiModel.A128CBCHS256A256GCMKW,
		cryptoutilOpenapiModel.A256CBCHS512A192GCMKW, cryptoutilOpenapiModel.A192CBCHS384A192GCMKW, cryptoutilOpenapiModel.A128CBCHS256A192GCMKW,
		cryptoutilOpenapiModel.A256CBCHS512A128GCMKW, cryptoutilOpenapiModel.A192CBCHS384A128GCMKW, cryptoutilOpenapiModel.A128CBCHS256A128GCMKW,
		cryptoutilOpenapiModel.A192CBCHS384Dir, cryptoutilOpenapiModel.A128CBCHS256Dir:
		return cryptoutilMagic.TestProbQuarter
	// Base signature algorithms - always test.
	case cryptoutilOpenapiModel.RS256, cryptoutilOpenapiModel.PS256, cryptoutilOpenapiModel.ES256, cryptoutilOpenapiModel.HS256, cryptoutilOpenapiModel.EdDSA:
		return cryptoutilMagic.TestProbAlways
	// Other signature algorithm sizes - sample testing.
	case cryptoutilOpenapiModel.RS384, cryptoutilOpenapiModel.RS512,
		cryptoutilOpenapiModel.PS384, cryptoutilOpenapiModel.PS512,
		cryptoutilOpenapiModel.ES384, cryptoutilOpenapiModel.ES512,
		cryptoutilOpenapiModel.HS384, cryptoutilOpenapiModel.HS512:
		return cryptoutilMagic.TestProbThird
	default:
		return cryptoutilMagic.TestProbAlways
	}
}
