package businessmodel

import (
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"
)

type (
	ElasticKeyId                string
	ElasticKeyName              string
	ElasticKeyDescription       string
	ElasticKeyAlgorithm         string
	ElasticKeyProvider          string
	ElasticKeyStatus            string
	ElasticKeyImportAllowed     bool
	ElasticKeyExportAllowed     bool
	ElasticKeyVersioningAllowed bool
)

const (
	Internal ElasticKeyProvider = "Internal"

	Creating                       ElasticKeyStatus = "creating"
	ImportFailed                   ElasticKeyStatus = "import_failed"
	PendingImport                  ElasticKeyStatus = "pending_import"
	PendingGenerate                ElasticKeyStatus = "pending_generate"
	GenerateFailed                 ElasticKeyStatus = "generate_failed"
	Active                         ElasticKeyStatus = "active"
	Disabled                       ElasticKeyStatus = "disabled"
	PendingDeleteWasImportFailed   ElasticKeyStatus = "pending_delete_was_import_failed"
	PendingDeleteWasPendingImport  ElasticKeyStatus = "pending_delete_was_pending_import"
	PendingDeleteWasActive         ElasticKeyStatus = "pending_delete_was_active"
	PendingDeleteWasDisabled       ElasticKeyStatus = "pending_delete_was_disabled"
	PendingDeleteWasGenerateFailed ElasticKeyStatus = "pending_delete_was_generate_failed"
	StartedDelete                  ElasticKeyStatus = "started_delete"
	FinishedDelete                 ElasticKeyStatus = "finished_delete"
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

var asymmetricElasticKeyAlgorithm = map[cryptoutilOpenapiModel.ElasticKeyAlgorithm]bool{
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

var symmetricElasticKeyAlgorithm = map[cryptoutilOpenapiModel.ElasticKeyAlgorithm]bool{
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
