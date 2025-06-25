package constant

type ElasticKeyAlgorithm string

const (
	A256GCM_A256KW    ElasticKeyAlgorithm = "A256GCM/A256KW"    // ElasticKeyAlgorithm
	A192GCM_A256KW    ElasticKeyAlgorithm = "A192GCM/A256KW"    // ElasticKeyAlgorithm
	A128GCM_A256KW    ElasticKeyAlgorithm = "A128GCM/A256KW"    // ElasticKeyAlgorithm
	A256GCM_A192KW    ElasticKeyAlgorithm = "A256GCM/A192KW"    // ElasticKeyAlgorithm
	A192GCM_A192KW    ElasticKeyAlgorithm = "A192GCM/A192KW"    // ElasticKeyAlgorithm
	A128GCM_A192KW    ElasticKeyAlgorithm = "A128GCM/A192KW"    // ElasticKeyAlgorithm
	A256GCM_A128KW    ElasticKeyAlgorithm = "A256GCM/A128KW"    // ElasticKeyAlgorithm
	A192GCM_A128KW    ElasticKeyAlgorithm = "A192GCM/A128KW"    // ElasticKeyAlgorithm
	A128GCM_A128KW    ElasticKeyAlgorithm = "A128GCM/A128KW"    // ElasticKeyAlgorithm
	A256GCM_A256GCMKW ElasticKeyAlgorithm = "A256GCM/A256GCMKW" // ElasticKeyAlgorithm
	A192GCM_A256GCMKW ElasticKeyAlgorithm = "A192GCM/A256GCMKW" // ElasticKeyAlgorithm
	A128GCM_A256GCMKW ElasticKeyAlgorithm = "A128GCM/A256GCMKW" // ElasticKeyAlgorithm
	A256GCM_A192GCMKW ElasticKeyAlgorithm = "A256GCM/A192GCMKW" // ElasticKeyAlgorithm
	A192GCM_A192GCMKW ElasticKeyAlgorithm = "A192GCM/A192GCMKW" // ElasticKeyAlgorithm
	A128GCM_A192GCMKW ElasticKeyAlgorithm = "A128GCM/A192GCMKW" // ElasticKeyAlgorithm
	A256GCM_A128GCMKW ElasticKeyAlgorithm = "A256GCM/A128GCMKW" // ElasticKeyAlgorithm
	A192GCM_A128GCMKW ElasticKeyAlgorithm = "A192GCM/A128GCMKW" // ElasticKeyAlgorithm
	A128GCM_A128GCMKW ElasticKeyAlgorithm = "A128GCM/A128GCMKW" // ElasticKeyAlgorithm
	A256GCM_dir       ElasticKeyAlgorithm = "A256GCM/dir"       // ElasticKeyAlgorithm
	A192GCM_dir       ElasticKeyAlgorithm = "A192GCM/dir"       // ElasticKeyAlgorithm
	A128GCM_dir       ElasticKeyAlgorithm = "A128GCM/dir"       // ElasticKeyAlgorithm

	A256GCM_RSAOAEP512 ElasticKeyAlgorithm = "A256GCM/RSA-OAEP-512" // ElasticKeyAlgorithm
	A192GCM_RSAOAEP512 ElasticKeyAlgorithm = "A192GCM/RSA-OAEP-512" // ElasticKeyAlgorithm
	A128GCM_RSAOAEP512 ElasticKeyAlgorithm = "A128GCM/RSA-OAEP-512" // ElasticKeyAlgorithm
	A256GCM_RSAOAEP384 ElasticKeyAlgorithm = "A256GCM/RSA-OAEP-384" // ElasticKeyAlgorithm
	A192GCM_RSAOAEP384 ElasticKeyAlgorithm = "A192GCM/RSA-OAEP-384" // ElasticKeyAlgorithm
	A128GCM_RSAOAEP384 ElasticKeyAlgorithm = "A128GCM/RSA-OAEP-384" // ElasticKeyAlgorithm
	A256GCM_RSAOAEP256 ElasticKeyAlgorithm = "A256GCM/RSA-OAEP-256" // ElasticKeyAlgorithm
	A192GCM_RSAOAEP256 ElasticKeyAlgorithm = "A192GCM/RSA-OAEP-256" // ElasticKeyAlgorithm
	A128GCM_RSAOAEP256 ElasticKeyAlgorithm = "A128GCM/RSA-OAEP-256" // ElasticKeyAlgorithm
	A256GCM_RSAOAEP    ElasticKeyAlgorithm = "A256GCM/RSA-OAEP"     // ElasticKeyAlgorithm
	A192GCM_RSAOAEP    ElasticKeyAlgorithm = "A192GCM/RSA-OAEP"     // ElasticKeyAlgorithm
	A128GCM_RSAOAEP    ElasticKeyAlgorithm = "A128GCM/RSA-OAEP"     // ElasticKeyAlgorithm
	A256GCM_RSA15      ElasticKeyAlgorithm = "A256GCM/RSA1_5"       // ElasticKeyAlgorithm
	A192GCM_RSA15      ElasticKeyAlgorithm = "A192GCM/RSA1_5"       // ElasticKeyAlgorithm
	A128GCM_RSA15      ElasticKeyAlgorithm = "A128GCM/RSA1_5"       // ElasticKeyAlgorithm

	A256GCM_ECDHESA256KW ElasticKeyAlgorithm = "A256GCM/ECDH-ES+A256KW" // ElasticKeyAlgorithm
	A192GCM_ECDHESA256KW ElasticKeyAlgorithm = "A192GCM/ECDH-ES+A256KW" // ElasticKeyAlgorithm
	A128GCM_ECDHESA256KW ElasticKeyAlgorithm = "A128GCM/ECDH-ES+A256KW" // ElasticKeyAlgorithm
	A256GCM_ECDHESA192KW ElasticKeyAlgorithm = "A256GCM/ECDH-ES+A192KW" // ElasticKeyAlgorithm
	A192GCM_ECDHESA192KW ElasticKeyAlgorithm = "A192GCM/ECDH-ES+A192KW" // ElasticKeyAlgorithm
	A128GCM_ECDHESA192KW ElasticKeyAlgorithm = "A128GCM/ECDH-ES+A192KW" // ElasticKeyAlgorithm
	A256GCM_ECDHESA128KW ElasticKeyAlgorithm = "A256GCM/ECDH-ES+A128KW" // ElasticKeyAlgorithm
	A192GCM_ECDHESA128KW ElasticKeyAlgorithm = "A192GCM/ECDH-ES+A128KW" // ElasticKeyAlgorithm
	A128GCM_ECDHESA128KW ElasticKeyAlgorithm = "A128GCM/ECDH-ES+A128KW" // ElasticKeyAlgorithm
	A256GCM_ECDHES       ElasticKeyAlgorithm = "A256GCM/ECDH-ES"        // ElasticKeyAlgorithm
	A192GCM_ECDHES       ElasticKeyAlgorithm = "A192GCM/ECDH-ES"        // ElasticKeyAlgorithm
	A128GCM_ECDHES       ElasticKeyAlgorithm = "A128GCM/ECDH-ES"        // ElasticKeyAlgorithm

	A256CBCHS512_A256KW    ElasticKeyAlgorithm = "A256CBC-HS512/A256KW"    // ElasticKeyAlgorithm
	A192CBCHS384_A256KW    ElasticKeyAlgorithm = "A192CBC-HS384/A256KW"    // ElasticKeyAlgorithm
	A128CBCHS256_A256KW    ElasticKeyAlgorithm = "A128CBC-HS256/A256KW"    // ElasticKeyAlgorithm
	A256CBCHS512_A192KW    ElasticKeyAlgorithm = "A256CBC-HS512/A192KW"    // ElasticKeyAlgorithm
	A192CBCHS384_A192KW    ElasticKeyAlgorithm = "A192CBC-HS384/A192KW"    // ElasticKeyAlgorithm
	A128CBCHS256_A192KW    ElasticKeyAlgorithm = "A128CBC-HS256/A192KW"    // ElasticKeyAlgorithm
	A256CBCHS512_A128KW    ElasticKeyAlgorithm = "A256CBC-HS512/A128KW"    // ElasticKeyAlgorithm
	A192CBCHS384_A128KW    ElasticKeyAlgorithm = "A192CBC-HS384/A128KW"    // ElasticKeyAlgorithm
	A128CBCHS256_A128KW    ElasticKeyAlgorithm = "A128CBC-HS256/A128KW"    // ElasticKeyAlgorithm
	A256CBCHS512_A256GCMKW ElasticKeyAlgorithm = "A256CBC-HS512/A256GCMKW" // ElasticKeyAlgorithm
	A192CBCHS384_A256GCMKW ElasticKeyAlgorithm = "A192CBC-HS384/A256GCMKW" // ElasticKeyAlgorithm
	A128CBCHS256_A256GCMKW ElasticKeyAlgorithm = "A128CBC-HS256/A256GCMKW" // ElasticKeyAlgorithm
	A256CBCHS512_A192GCMKW ElasticKeyAlgorithm = "A256CBC-HS512/A192GCMKW" // ElasticKeyAlgorithm
	A192CBCHS384_A192GCMKW ElasticKeyAlgorithm = "A192CBC-HS384/A192GCMKW" // ElasticKeyAlgorithm
	A128CBCHS256_A192GCMKW ElasticKeyAlgorithm = "A128CBC-HS256/A192GCMKW" // ElasticKeyAlgorithm
	A256CBCHS512_A128GCMKW ElasticKeyAlgorithm = "A256CBC-HS512/A128GCMKW" // ElasticKeyAlgorithm
	A192CBCHS384_A128GCMKW ElasticKeyAlgorithm = "A192CBC-HS384/A128GCMKW" // ElasticKeyAlgorithm
	A128CBCHS256_A128GCMKW ElasticKeyAlgorithm = "A128CBC-HS256/A128GCMKW" // ElasticKeyAlgorithm
	A256CBCHS512_dir       ElasticKeyAlgorithm = "A256CBC-HS512/dir"       // ElasticKeyAlgorithm
	A192CBCHS384_dir       ElasticKeyAlgorithm = "A192CBC-HS384/dir"       // ElasticKeyAlgorithm
	A128CBCHS256_dir       ElasticKeyAlgorithm = "A128CBC-HS256/dir"       // ElasticKeyAlgorithm

	A256CBC_HS512_RSAOAEP512 ElasticKeyAlgorithm = "A256CBC-HS512/RSA-OAEP-512" // ElasticKeyAlgorithm
	A192CBC_HS384_RSAOAEP512 ElasticKeyAlgorithm = "A192CBC-HS384/RSA-OAEP-512" // ElasticKeyAlgorithm
	A128CBC_HS256_RSAOAEP512 ElasticKeyAlgorithm = "A128CBC-HS256/RSA-OAEP-512" // ElasticKeyAlgorithm
	A256CBC_HS512_RSAOAEP384 ElasticKeyAlgorithm = "A256CBC-HS512/RSA-OAEP-384" // ElasticKeyAlgorithm
	A192CBC_HS384_RSAOAEP384 ElasticKeyAlgorithm = "A192CBC-HS384/RSA-OAEP-384" // ElasticKeyAlgorithm
	A128CBC_HS256_RSAOAEP384 ElasticKeyAlgorithm = "A128CBC-HS256/RSA-OAEP-384" // ElasticKeyAlgorithm
	A256CBC_HS512_RSAOAEP256 ElasticKeyAlgorithm = "A256CBC-HS512/RSA-OAEP-256" // ElasticKeyAlgorithm
	A192CBC_HS384_RSAOAEP256 ElasticKeyAlgorithm = "A192CBC-HS384/RSA-OAEP-256" // ElasticKeyAlgorithm
	A128CBC_HS256_RSAOAEP256 ElasticKeyAlgorithm = "A128CBC-HS256/RSA-OAEP-256" // ElasticKeyAlgorithm
	A256CBC_HS512_RSAOAEP    ElasticKeyAlgorithm = "A256CBC-HS512/RSA-OAEP"     // ElasticKeyAlgorithm
	A192CBC_HS384_RSAOAEP    ElasticKeyAlgorithm = "A192CBC-HS384/RSA-OAEP"     // ElasticKeyAlgorithm
	A128CBC_HS256_RSAOAEP    ElasticKeyAlgorithm = "A128CBC-HS256/RSA-OAEP"     // ElasticKeyAlgorithm
	A256CBC_HS512_RSA15      ElasticKeyAlgorithm = "A256CBC-HS512/RSA1_5"       // ElasticKeyAlgorithm
	A192CBC_HS384_RSA15      ElasticKeyAlgorithm = "A192CBC-HS384/RSA1_5"       // ElasticKeyAlgorithm
	A128CBC_HS256_RSA15      ElasticKeyAlgorithm = "A128CBC-HS256/RSA1_5"       // ElasticKeyAlgorithm

	A256CBC_HS512_ECDHESA256KW ElasticKeyAlgorithm = "A256CBC-HS512/ECDH-ES+A256KW" // ElasticKeyAlgorithm
	A192CBC_HS384_ECDHESA256KW ElasticKeyAlgorithm = "A192CBC-HS384/ECDH-ES+A256KW" // ElasticKeyAlgorithm
	A128CBC_HS256_ECDHESA256KW ElasticKeyAlgorithm = "A128CBC-HS256/ECDH-ES+A256KW" // ElasticKeyAlgorithm
	A192CBC_HS384_ECDHESA192KW ElasticKeyAlgorithm = "A192CBC-HS384/ECDH-ES+A192KW" // ElasticKeyAlgorithm
	A128CBC_HS256_ECDHESA192KW ElasticKeyAlgorithm = "A128CBC-HS256/ECDH-ES+A192KW" // ElasticKeyAlgorithm
	A128CBC_HS256_ECDHESA128KW ElasticKeyAlgorithm = "A128CBC-HS256/ECDH-ES+A128KW" // ElasticKeyAlgorithm
	A256CBC_HS512_ECDHES       ElasticKeyAlgorithm = "A256CBC-HS512/ECDH-ES"        // ElasticKeyAlgorithm
	A192CBC_HS384_ECDHES       ElasticKeyAlgorithm = "A192CBC-HS384/ECDH-ES"        // ElasticKeyAlgorithm
	A128CBC_HS256_ECDHES       ElasticKeyAlgorithm = "A128CBC-HS256/ECDH-ES"        // ElasticKeyAlgorithm

	RS512 ElasticKeyAlgorithm = "RS512" // ElasticKeyAlgorithm
	RS384 ElasticKeyAlgorithm = "RS384" // ElasticKeyAlgorithm
	RS256 ElasticKeyAlgorithm = "RS256" // ElasticKeyAlgorithm
	PS512 ElasticKeyAlgorithm = "PS512" // ElasticKeyAlgorithm
	PS384 ElasticKeyAlgorithm = "PS384" // ElasticKeyAlgorithm
	PS256 ElasticKeyAlgorithm = "PS256" // ElasticKeyAlgorithm
	ES512 ElasticKeyAlgorithm = "ES512" // ElasticKeyAlgorithm
	ES384 ElasticKeyAlgorithm = "ES384" // ElasticKeyAlgorithm
	ES256 ElasticKeyAlgorithm = "ES256" // ElasticKeyAlgorithm
	HS512 ElasticKeyAlgorithm = "HS512" // ElasticKeyAlgorithm
	HS384 ElasticKeyAlgorithm = "HS384" // ElasticKeyAlgorithm
	HS256 ElasticKeyAlgorithm = "HS256" // ElasticKeyAlgorithm
	EdDSA ElasticKeyAlgorithm = "EdDSA" // ElasticKeyAlgorithm
)

var (
	asymmetricElasticKeyAlgorithm = map[ElasticKeyAlgorithm]bool{
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
)

var symmetricElasticKeyAlgorithm = map[ElasticKeyAlgorithm]bool{
	A256GCM_A256KW: true, A192GCM_A256KW: true, A128GCM_A256KW: true,
	A256GCM_A192KW: true, A192GCM_A192KW: true, A128GCM_A192KW: true,
	A256GCM_A128KW: true, A192GCM_A128KW: true, A128GCM_A128KW: true,
	A256GCM_A256GCMKW: true, A192GCM_A256GCMKW: true, A128GCM_A256GCMKW: true,
	A256GCM_A192GCMKW: true, A192GCM_A192GCMKW: true, A128GCM_A192GCMKW: true,
	A256GCM_A128GCMKW: true, A192GCM_A128GCMKW: true, A128GCM_A128GCMKW: true,
	A256GCM_dir: true, A192GCM_dir: true, A128GCM_dir: true,

	A256CBCHS512_A256KW: true, A192CBCHS384_A256KW: true, A128CBCHS256_A256KW: true,
	A256CBCHS512_A192KW: true, A192CBCHS384_A192KW: true, A128CBCHS256_A192KW: true,
	A256CBCHS512_A128KW: true, A192CBCHS384_A128KW: true, A128CBCHS256_A128KW: true,
	A256CBCHS512_A256GCMKW: true, A192CBCHS384_A256GCMKW: true, A128CBCHS256_A256GCMKW: true,
	A256CBCHS512_A192GCMKW: true, A192CBCHS384_A192GCMKW: true, A128CBCHS256_A192GCMKW: true,
	A256CBCHS512_A128GCMKW: true, A192CBCHS384_A128GCMKW: true, A128CBCHS256_A128GCMKW: true,
	A256CBCHS512_dir: true, A192CBCHS384_dir: true, A128CBCHS256_dir: true,
}

func IsAsymmetric(alg *ElasticKeyAlgorithm) bool {
	return asymmetricElasticKeyAlgorithm[*alg]
}

func IsSymmetric(alg *ElasticKeyAlgorithm) bool {
	return symmetricElasticKeyAlgorithm[*alg]
}

type ElasticKeyProvider string

const (
	Internal ElasticKeyProvider = "Internal"
)

type ElasticKeyStatus string

const (
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

type (
	ElasticKeyDescription       string
	ElasticKeyId                string
	ElasticKeyExportAllowed     bool
	ElasticKeyImportAllowed     bool
	ElasticKeyVersioningAllowed bool
	ElasticKeyName              string
)
