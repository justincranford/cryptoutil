package businessmodel

import (
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"

	"fmt"
)

type ElasticKeyAlgorithm string

const (
	A256GCM_A256KW    ElasticKeyAlgorithm = "A256GCM/A256KW"
	A192GCM_A256KW    ElasticKeyAlgorithm = "A192GCM/A256KW"
	A128GCM_A256KW    ElasticKeyAlgorithm = "A128GCM/A256KW"
	A256GCM_A192KW    ElasticKeyAlgorithm = "A256GCM/A192KW"
	A192GCM_A192KW    ElasticKeyAlgorithm = "A192GCM/A192KW"
	A128GCM_A192KW    ElasticKeyAlgorithm = "A128GCM/A192KW"
	A256GCM_A128KW    ElasticKeyAlgorithm = "A256GCM/A128KW"
	A192GCM_A128KW    ElasticKeyAlgorithm = "A192GCM/A128KW"
	A128GCM_A128KW    ElasticKeyAlgorithm = "A128GCM/A128KW"
	A256GCM_A256GCMKW ElasticKeyAlgorithm = "A256GCM/A256GCMKW"
	A192GCM_A256GCMKW ElasticKeyAlgorithm = "A192GCM/A256GCMKW"
	A128GCM_A256GCMKW ElasticKeyAlgorithm = "A128GCM/A256GCMKW"
	A256GCM_A192GCMKW ElasticKeyAlgorithm = "A256GCM/A192GCMKW"
	A192GCM_A192GCMKW ElasticKeyAlgorithm = "A192GCM/A192GCMKW"
	A128GCM_A192GCMKW ElasticKeyAlgorithm = "A128GCM/A192GCMKW"
	A256GCM_A128GCMKW ElasticKeyAlgorithm = "A256GCM/A128GCMKW"
	A192GCM_A128GCMKW ElasticKeyAlgorithm = "A192GCM/A128GCMKW"
	A128GCM_A128GCMKW ElasticKeyAlgorithm = "A128GCM/A128GCMKW"
	A256GCM_dir       ElasticKeyAlgorithm = "A256GCM/dir"
	A192GCM_dir       ElasticKeyAlgorithm = "A192GCM/dir"
	A128GCM_dir       ElasticKeyAlgorithm = "A128GCM/dir"

	A256GCM_RSAOAEP512 ElasticKeyAlgorithm = "A256GCM/RSA-OAEP-512"
	A192GCM_RSAOAEP512 ElasticKeyAlgorithm = "A192GCM/RSA-OAEP-512"
	A128GCM_RSAOAEP512 ElasticKeyAlgorithm = "A128GCM/RSA-OAEP-512"
	A256GCM_RSAOAEP384 ElasticKeyAlgorithm = "A256GCM/RSA-OAEP-384"
	A192GCM_RSAOAEP384 ElasticKeyAlgorithm = "A192GCM/RSA-OAEP-384"
	A128GCM_RSAOAEP384 ElasticKeyAlgorithm = "A128GCM/RSA-OAEP-384"
	A256GCM_RSAOAEP256 ElasticKeyAlgorithm = "A256GCM/RSA-OAEP-256"
	A192GCM_RSAOAEP256 ElasticKeyAlgorithm = "A192GCM/RSA-OAEP-256"
	A128GCM_RSAOAEP256 ElasticKeyAlgorithm = "A128GCM/RSA-OAEP-256"
	A256GCM_RSAOAEP    ElasticKeyAlgorithm = "A256GCM/RSA-OAEP"
	A192GCM_RSAOAEP    ElasticKeyAlgorithm = "A192GCM/RSA-OAEP"
	A128GCM_RSAOAEP    ElasticKeyAlgorithm = "A128GCM/RSA-OAEP"
	A256GCM_RSA15      ElasticKeyAlgorithm = "A256GCM/RSA1_5"
	A192GCM_RSA15      ElasticKeyAlgorithm = "A192GCM/RSA1_5"
	A128GCM_RSA15      ElasticKeyAlgorithm = "A128GCM/RSA1_5"

	A256GCM_ECDHESA256KW ElasticKeyAlgorithm = "A256GCM/ECDH-ES+A256KW"
	A192GCM_ECDHESA256KW ElasticKeyAlgorithm = "A192GCM/ECDH-ES+A256KW"
	A128GCM_ECDHESA256KW ElasticKeyAlgorithm = "A128GCM/ECDH-ES+A256KW"
	A256GCM_ECDHESA192KW ElasticKeyAlgorithm = "A256GCM/ECDH-ES+A192KW"
	A192GCM_ECDHESA192KW ElasticKeyAlgorithm = "A192GCM/ECDH-ES+A192KW"
	A128GCM_ECDHESA192KW ElasticKeyAlgorithm = "A128GCM/ECDH-ES+A192KW"
	A256GCM_ECDHESA128KW ElasticKeyAlgorithm = "A256GCM/ECDH-ES+A128KW"
	A192GCM_ECDHESA128KW ElasticKeyAlgorithm = "A192GCM/ECDH-ES+A128KW"
	A128GCM_ECDHESA128KW ElasticKeyAlgorithm = "A128GCM/ECDH-ES+A128KW"
	A256GCM_ECDHES       ElasticKeyAlgorithm = "A256GCM/ECDH-ES"
	A192GCM_ECDHES       ElasticKeyAlgorithm = "A192GCM/ECDH-ES"
	A128GCM_ECDHES       ElasticKeyAlgorithm = "A128GCM/ECDH-ES"

	A256CBCHS512_A256KW    ElasticKeyAlgorithm = "A256CBC-HS512/A256KW"
	A192CBCHS384_A256KW    ElasticKeyAlgorithm = "A192CBC-HS384/A256KW"
	A128CBCHS256_A256KW    ElasticKeyAlgorithm = "A128CBC-HS256/A256KW"
	A256CBCHS512_A192KW    ElasticKeyAlgorithm = "A256CBC-HS512/A192KW"
	A192CBCHS384_A192KW    ElasticKeyAlgorithm = "A192CBC-HS384/A192KW"
	A128CBCHS256_A192KW    ElasticKeyAlgorithm = "A128CBC-HS256/A192KW"
	A256CBCHS512_A128KW    ElasticKeyAlgorithm = "A256CBC-HS512/A128KW"
	A192CBCHS384_A128KW    ElasticKeyAlgorithm = "A192CBC-HS384/A128KW"
	A128CBCHS256_A128KW    ElasticKeyAlgorithm = "A128CBC-HS256/A128KW"
	A256CBCHS512_A256GCMKW ElasticKeyAlgorithm = "A256CBC-HS512/A256GCMKW"
	A192CBCHS384_A256GCMKW ElasticKeyAlgorithm = "A192CBC-HS384/A256GCMKW"
	A128CBCHS256_A256GCMKW ElasticKeyAlgorithm = "A128CBC-HS256/A256GCMKW"
	A256CBCHS512_A192GCMKW ElasticKeyAlgorithm = "A256CBC-HS512/A192GCMKW"
	A192CBCHS384_A192GCMKW ElasticKeyAlgorithm = "A192CBC-HS384/A192GCMKW"
	A128CBCHS256_A192GCMKW ElasticKeyAlgorithm = "A128CBC-HS256/A192GCMKW"
	A256CBCHS512_A128GCMKW ElasticKeyAlgorithm = "A256CBC-HS512/A128GCMKW"
	A192CBCHS384_A128GCMKW ElasticKeyAlgorithm = "A192CBC-HS384/A128GCMKW"
	A128CBCHS256_A128GCMKW ElasticKeyAlgorithm = "A128CBC-HS256/A128GCMKW"
	A256CBCHS512_dir       ElasticKeyAlgorithm = "A256CBC-HS512/dir"
	A192CBCHS384_dir       ElasticKeyAlgorithm = "A192CBC-HS384/dir"
	A128CBCHS256_dir       ElasticKeyAlgorithm = "A128CBC-HS256/dir"

	A256CBC_HS512_RSAOAEP512 ElasticKeyAlgorithm = "A256CBC-HS512/RSA-OAEP-512"
	A192CBC_HS384_RSAOAEP512 ElasticKeyAlgorithm = "A192CBC-HS384/RSA-OAEP-512"
	A128CBC_HS256_RSAOAEP512 ElasticKeyAlgorithm = "A128CBC-HS256/RSA-OAEP-512"
	A256CBC_HS512_RSAOAEP384 ElasticKeyAlgorithm = "A256CBC-HS512/RSA-OAEP-384"
	A192CBC_HS384_RSAOAEP384 ElasticKeyAlgorithm = "A192CBC-HS384/RSA-OAEP-384"
	A128CBC_HS256_RSAOAEP384 ElasticKeyAlgorithm = "A128CBC-HS256/RSA-OAEP-384"
	A256CBC_HS512_RSAOAEP256 ElasticKeyAlgorithm = "A256CBC-HS512/RSA-OAEP-256"
	A192CBC_HS384_RSAOAEP256 ElasticKeyAlgorithm = "A192CBC-HS384/RSA-OAEP-256"
	A128CBC_HS256_RSAOAEP256 ElasticKeyAlgorithm = "A128CBC-HS256/RSA-OAEP-256"
	A256CBC_HS512_RSAOAEP    ElasticKeyAlgorithm = "A256CBC-HS512/RSA-OAEP"
	A192CBC_HS384_RSAOAEP    ElasticKeyAlgorithm = "A192CBC-HS384/RSA-OAEP"
	A128CBC_HS256_RSAOAEP    ElasticKeyAlgorithm = "A128CBC-HS256/RSA-OAEP"
	A256CBC_HS512_RSA15      ElasticKeyAlgorithm = "A256CBC-HS512/RSA1_5"
	A192CBC_HS384_RSA15      ElasticKeyAlgorithm = "A192CBC-HS384/RSA1_5"
	A128CBC_HS256_RSA15      ElasticKeyAlgorithm = "A128CBC-HS256/RSA1_5"

	A256CBC_HS512_ECDHESA256KW ElasticKeyAlgorithm = "A256CBC-HS512/ECDH-ES+A256KW"
	A192CBC_HS384_ECDHESA256KW ElasticKeyAlgorithm = "A192CBC-HS384/ECDH-ES+A256KW"
	A128CBC_HS256_ECDHESA256KW ElasticKeyAlgorithm = "A128CBC-HS256/ECDH-ES+A256KW"
	A192CBC_HS384_ECDHESA192KW ElasticKeyAlgorithm = "A192CBC-HS384/ECDH-ES+A192KW"
	A128CBC_HS256_ECDHESA192KW ElasticKeyAlgorithm = "A128CBC-HS256/ECDH-ES+A192KW"
	A128CBC_HS256_ECDHESA128KW ElasticKeyAlgorithm = "A128CBC-HS256/ECDH-ES+A128KW"
	A256CBC_HS512_ECDHES       ElasticKeyAlgorithm = "A256CBC-HS512/ECDH-ES"
	A192CBC_HS384_ECDHES       ElasticKeyAlgorithm = "A192CBC-HS384/ECDH-ES"
	A128CBC_HS256_ECDHES       ElasticKeyAlgorithm = "A128CBC-HS256/ECDH-ES"

	RS512 ElasticKeyAlgorithm = "RS512"
	RS384 ElasticKeyAlgorithm = "RS384"
	RS256 ElasticKeyAlgorithm = "RS256"
	PS512 ElasticKeyAlgorithm = "PS512"
	PS384 ElasticKeyAlgorithm = "PS384"
	PS256 ElasticKeyAlgorithm = "PS256"
	ES512 ElasticKeyAlgorithm = "ES512"
	ES384 ElasticKeyAlgorithm = "ES384"
	ES256 ElasticKeyAlgorithm = "ES256"
	HS512 ElasticKeyAlgorithm = "HS512"
	HS384 ElasticKeyAlgorithm = "HS384"
	HS256 ElasticKeyAlgorithm = "HS256"
	EdDSA ElasticKeyAlgorithm = "EdDSA"
)

var elasticKeyAlgorithms = map[string]cryptoutilOpenapiModel.ElasticKeyAlgorithm{
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
	string(cryptoutilOpenapiModel.A256CBCHS512ECDHESA128KW): cryptoutilOpenapiModel.A256CBCHS512ECDHESA128KW,
	string(cryptoutilOpenapiModel.A192CBCHS384ECDHESA128KW): cryptoutilOpenapiModel.A192CBCHS384ECDHESA128KW,
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

var asymmetricElasticKeyAlgorithm = map[ElasticKeyAlgorithm]bool{
	A256GCM_RSAOAEP512: true, A192GCM_RSAOAEP512: true, A128GCM_RSAOAEP512: true,
	A256GCM_RSAOAEP384: true, A192GCM_RSAOAEP384: true, A128GCM_RSAOAEP384: true,
	A256GCM_RSAOAEP256: true, A192GCM_RSAOAEP256: true, A128GCM_RSAOAEP256: true,
	A256GCM_RSAOAEP: true, A192GCM_RSAOAEP: true, A128GCM_RSAOAEP: true,
	A256GCM_RSA15: true, A192GCM_RSA15: true, A128GCM_RSA15: true,

	A256GCM_ECDHESA256KW: true, A192GCM_ECDHESA256KW: true, A128GCM_ECDHESA256KW: true,
	A256GCM_ECDHESA192KW: true, A192GCM_ECDHESA192KW: true, A128GCM_ECDHESA192KW: true,
	A256GCM_ECDHESA128KW: true, A192GCM_ECDHESA128KW: true, A128GCM_ECDHESA128KW: true,
	A256GCM_ECDHES: true, A192GCM_ECDHES: true, A128GCM_ECDHES: true,

	A256CBC_HS512_RSAOAEP512: true, A192CBC_HS384_RSAOAEP512: true, A128CBC_HS256_RSAOAEP512: true,
	A256CBC_HS512_RSAOAEP384: true, A192CBC_HS384_RSAOAEP384: true, A128CBC_HS256_RSAOAEP384: true,
	A256CBC_HS512_RSAOAEP256: true, A192CBC_HS384_RSAOAEP256: true, A128CBC_HS256_RSAOAEP256: true,
	A256CBC_HS512_RSAOAEP: true, A192CBC_HS384_RSAOAEP: true, A128CBC_HS256_RSAOAEP: true,
	A256CBC_HS512_RSA15: true, A192CBC_HS384_RSA15: true, A128CBC_HS256_RSA15: true,

	A256CBC_HS512_ECDHESA256KW: true, A192CBC_HS384_ECDHESA256KW: true, A128CBC_HS256_ECDHESA256KW: true,
	A192CBC_HS384_ECDHESA192KW: true, A128CBC_HS256_ECDHESA192KW: true, A128CBC_HS256_ECDHESA128KW: true,
	A256CBC_HS512_ECDHES: true, A192CBC_HS384_ECDHES: true, A128CBC_HS256_ECDHES: true,

	RS512: true, RS384: true, RS256: true, PS512: true, PS384: true, PS256: true, ES512: true, ES384: true, ES256: true, EdDSA: true,
}

var symmetricElasticKeyAlgorithm = map[ElasticKeyAlgorithm]bool{
	A256GCM_A256KW: true, A192GCM_A256KW: true, A128GCM_A256KW: true,
	A256GCM_A192KW: true, A192GCM_A192KW: true, A128GCM_A192KW: true,
	A256GCM_A128KW: true, A192GCM_A128KW: true, A128GCM_A128KW: true,
	A256GCM_A256GCMKW: true, A192GCM_A256GCMKW: true, A128GCM_A256GCMKW: true,
	A256GCM_A192GCMKW: true, A192GCM_A192GCMKW: true, A128GCM_A192GCMKW: true,
	A256GCM_A128GCMKW: true, A192GCM_A128GCMKW: true, A128GCM_A128GCMKW: true,
	A256GCM_dir: true, A192GCM_dir: true, A128GCM_dir: true,

	A256GCM_RSAOAEP512: false, A192GCM_RSAOAEP512: false, A128GCM_RSAOAEP512: false,
	A256GCM_RSAOAEP384: false, A192GCM_RSAOAEP384: false, A128GCM_RSAOAEP384: false,
	A256GCM_RSAOAEP256: false, A192GCM_RSAOAEP256: false, A128GCM_RSAOAEP256: false,
	A256GCM_RSAOAEP: false, A192GCM_RSAOAEP: false, A128GCM_RSAOAEP: false,
	A256GCM_RSA15: false, A192GCM_RSA15: false, A128GCM_RSA15: false,

	A256GCM_ECDHESA256KW: false, A192GCM_ECDHESA256KW: false, A128GCM_ECDHESA256KW: false,
	A256GCM_ECDHESA192KW: false, A192GCM_ECDHESA192KW: false, A128GCM_ECDHESA192KW: false,
	A256GCM_ECDHESA128KW: false, A192GCM_ECDHESA128KW: false, A128GCM_ECDHESA128KW: false,
	A256GCM_ECDHES: false, A192GCM_ECDHES: false, A128GCM_ECDHES: false,

	A256CBCHS512_A256KW: true, A192CBCHS384_A256KW: true, A128CBCHS256_A256KW: true,
	A256CBCHS512_A192KW: true, A192CBCHS384_A192KW: true, A128CBCHS256_A192KW: true,
	A256CBCHS512_A128KW: true, A192CBCHS384_A128KW: true, A128CBCHS256_A128KW: true,
	A256CBCHS512_A256GCMKW: true, A192CBCHS384_A256GCMKW: true, A128CBCHS256_A256GCMKW: true,
	A256CBCHS512_A192GCMKW: true, A192CBCHS384_A192GCMKW: true, A128CBCHS256_A192GCMKW: true,
	A256CBCHS512_A128GCMKW: true, A192CBCHS384_A128GCMKW: true, A128CBCHS256_A128GCMKW: true,
	A256CBCHS512_dir: true, A192CBCHS384_dir: true, A128CBCHS256_dir: true,

	A256CBC_HS512_RSAOAEP512: false, A192CBC_HS384_RSAOAEP512: false, A128CBC_HS256_RSAOAEP512: false,
	A256CBC_HS512_RSAOAEP384: false, A192CBC_HS384_RSAOAEP384: false, A128CBC_HS256_RSAOAEP384: false,
	A256CBC_HS512_RSAOAEP256: false, A192CBC_HS384_RSAOAEP256: false, A128CBC_HS256_RSAOAEP256: false,
	A256CBC_HS512_RSAOAEP: false, A192CBC_HS384_RSAOAEP: false, A128CBC_HS256_RSAOAEP: false,
	A256CBC_HS512_RSA15: false, A192CBC_HS384_RSA15: false, A128CBC_HS256_RSA15: false,

	A256CBC_HS512_ECDHESA256KW: false, A192CBC_HS384_ECDHESA256KW: false, A128CBC_HS256_ECDHESA256KW: false,
	A192CBC_HS384_ECDHESA192KW: false, A128CBC_HS256_ECDHESA192KW: false, A128CBC_HS256_ECDHESA128KW: false,
	A256CBC_HS512_ECDHES: false, A192CBC_HS384_ECDHES: false, A128CBC_HS256_ECDHES: false,

	RS512: false, RS384: false, RS256: false,
	PS512: false, PS384: false, PS256: false,
	ES512: false, ES384: false, ES256: false,
	HS512: true, HS384: true, HS256: true,
	EdDSA: false,
}

func MapElasticKeyAlgorithm(algorithm string) (*cryptoutilOpenapiModel.ElasticKeyAlgorithm, error) {
	if err := ValidateString(algorithm); err != nil {
		return nil, fmt.Errorf("invalid elastic Key algorithm: %w", err)
	}
	if alg, exists := elasticKeyAlgorithms[algorithm]; exists {
		return &alg, nil
	}
	return nil, fmt.Errorf("invalid elastic Key algorithm: %s", algorithm)
}

// func IsSymmetric(ormElasticKeyAlgorithm *constant.ElasticKeyAlgorithm) (bool, error) {
// 	isSymmetric, ok := isSymmetric[*ormElasticKeyAlgorithm]
// 	if ok {
// 		return isSymmetric, nil
// 	}
// 	return false, fmt.Errorf("unsupported ElasticKeyAlgorithm '%s'", *ormElasticKeyAlgorithm)
// }

// func isAsymmetric(ormElasticKeyAlgorithm *constant.ElasticKeyAlgorithm) (bool, error) {
// 	isSymmetric, ok := isSymmetric[*ormElasticKeyAlgorithm]
// 	if ok {
// 		return !isSymmetric, nil
// 	}
// 	return false, fmt.Errorf("unsupported ElasticKeyAlgorithm '%s'", *ormElasticKeyAlgorithm)
// }

func IsAsymmetric(alg *ElasticKeyAlgorithm) bool {
	return asymmetricElasticKeyAlgorithm[*alg]
}

func IsSymmetric(alg *ElasticKeyAlgorithm) bool {
	return symmetricElasticKeyAlgorithm[*alg]
}
