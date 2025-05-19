package orm

import (
	"time"

	googleUuid "github.com/google/uuid"
)

type KeyPool struct {
	KeyPoolID                googleUuid.UUID  `gorm:"type:uuid;primaryKey"`
	KeyPoolName              string           `gorm:"size:63;not null;check:length(key_pool_name) >= 1;unique"`
	KeyPoolDescription       string           `gorm:"size:255;not null;check:length(key_pool_description) >= 1"`
	KeyPoolProvider          KeyPoolProvider  `gorm:"size:8;not null;check:key_pool_provider IN ('Internal')"`
	KeyPoolAlgorithm         KeyPoolAlgorithm `gorm:"size:23;not null;check:key_pool_algorithm IN ('A256GCM/A256KW', 'A192GCM/A256KW', 'A128GCM/A256KW', 'A192GCM/A192KW', 'A128GCM/A192KW', 'A128GCM/A128KW', 'A256GCM/A256GCMKW', 'A192GCM/A256GCMKW', 'A128GCM/A256GCMKW', 'A192GCM/A192GCMKW', 'A128GCM/A192GCMKW', 'A128GCM/A128GCMKW', 'A256GCM/dir', 'A192GCM/dir', 'A128GCM/dir', 'A256CBC-HS512/A256KW', 'A192CBC-HS384/A256KW', 'A128CBC-HS256/A256KW', 'A192CBC-HS384/A192KW', 'A128CBC-HS256/A192KW', 'A128CBC-HS256/A128KW', 'A256CBC-HS512/A256GCMKW', 'A192CBC-HS384/A256GCMKW', 'A128CBC-HS256/A256GCMKW', 'A192CBC-HS384/A192GCMKW', 'A128CBC-HS256/A192GCMKW', 'A128CBC-HS256/A128GCMKW', 'A256CBC-HS512/dir', 'A192CBC-HS384/dir', 'A128CBC-HS256/dir')"`
	KeyPoolVersioningAllowed bool             `gorm:"not null;check:key_pool_versioning_allowed IN (TRUE, FALSE)"`
	KeyPoolImportAllowed     bool             `gorm:"not null;check:key_pool_import_allowed IN (TRUE, FALSE)"`
	KeyPoolExportAllowed     bool             `gorm:"not null;check:key_pool_export_allowed IN (TRUE, FALSE)"`
	KeyPoolStatus            KeyPoolStatus    `gorm:"size:34;not null;check:key_pool_status IN ('creating', 'import_failed', 'pending_import', 'pending_generate', 'generate_failed', 'active', 'disabled', 'pending_delete_was_import_failed', 'pending_delete_was_pending_import', 'pending_delete_was_active', 'pending_delete_was_disabled', 'pending_delete_was_generate_failed', 'started_delete', 'finished_delete')"`
}

type Key struct {
	KeyPoolID         googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	KeyID             googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	KeyMaterial       []byte          `gorm:"not null;check(length(key_material) >= 1)"`
	KeyGenerateDate   *time.Time
	KeyImportDate     *time.Time
	KeyExpirationDate *time.Time
	KeyRevocationDate *time.Time
}

type KeyPoolAlgorithm string

const (
	A256GCM_A256KW             KeyPoolAlgorithm = "A256GCM/A256KW"               // KeyEncryptionAlgorithm
	A192GCM_A256KW             KeyPoolAlgorithm = "A192GCM/A256KW"               // KeyEncryptionAlgorithm
	A128GCM_A256KW             KeyPoolAlgorithm = "A128GCM/A256KW"               // KeyEncryptionAlgorithm
	A192GCM_A192KW             KeyPoolAlgorithm = "A192GCM/A192KW"               // KeyEncryptionAlgorithm
	A128GCM_A192KW             KeyPoolAlgorithm = "A128GCM/A192KW"               // KeyEncryptionAlgorithm
	A128GCM_A128KW             KeyPoolAlgorithm = "A128GCM/A128KW"               // KeyEncryptionAlgorithm
	A256GCM_A256GCMKW          KeyPoolAlgorithm = "A256GCM/A256GCMKW"            // KeyEncryptionAlgorithm
	A192GCM_A256GCMKW          KeyPoolAlgorithm = "A192GCM/A256GCMKW"            // KeyEncryptionAlgorithm
	A128GCM_A256GCMKW          KeyPoolAlgorithm = "A128GCM/A256GCMKW"            // KeyEncryptionAlgorithm
	A192GCM_A192GCMKW          KeyPoolAlgorithm = "A192GCM/A192GCMKW"            // KeyEncryptionAlgorithm
	A128GCM_A192GCMKW          KeyPoolAlgorithm = "A128GCM/A192GCMKW"            // KeyEncryptionAlgorithm
	A128GCM_A128GCMKW          KeyPoolAlgorithm = "A128GCM/A128GCMKW"            // KeyEncryptionAlgorithm
	A256GCM_RSAOAEP512         KeyPoolAlgorithm = "A256GCM/RSA-OAEP-512"         // KeyEncryptionAlgorithm
	A192GCM_RSAOAEP512         KeyPoolAlgorithm = "A192GCM/RSA-OAEP-512"         // KeyEncryptionAlgorithm
	A128GCM_RSAOAEP512         KeyPoolAlgorithm = "A128GCM/RSA-OAEP-512"         // KeyEncryptionAlgorithm
	A192GCM_RSAOAEP384         KeyPoolAlgorithm = "A192GCM/RSA-OAEP-384"         // KeyEncryptionAlgorithm
	A128GCM_RSAOAEP384         KeyPoolAlgorithm = "A128GCM/RSA-OAEP-384"         // KeyEncryptionAlgorithm
	A128GCM_RSAOAEP256         KeyPoolAlgorithm = "A128GCM/RSA-OAEP-256"         // KeyEncryptionAlgorithm
	A128GCM_RSAOAEP            KeyPoolAlgorithm = "A128GCM/RSA-OAEP"             // KeyEncryptionAlgorithm
	A128GCM_RSA15              KeyPoolAlgorithm = "A128GCM/RSA1_5"               // KeyEncryptionAlgorithm
	A256GCM_ECDHESA256KW       KeyPoolAlgorithm = "A256GCM/ECDH-ES+A256KW"       // KeyEncryptionAlgorithm
	A192GCM_ECDHESA256KW       KeyPoolAlgorithm = "A192GCM/ECDH-ES+A256KW"       // KeyEncryptionAlgorithm
	A128GCM_ECDHESA256KW       KeyPoolAlgorithm = "A128GCM/ECDH-ES+A256KW"       // KeyEncryptionAlgorithm
	A192GCM_ECDHESA192KW       KeyPoolAlgorithm = "A192GCM/ECDH-ES+A192KW"       // KeyEncryptionAlgorithm
	A128GCM_ECDHESA192KW       KeyPoolAlgorithm = "A128GCM/ECDH-ES+A192KW"       // KeyEncryptionAlgorithm
	A128GCM_ECDHESA128KW       KeyPoolAlgorithm = "A128GCM/ECDH-ES+A128KW"       // KeyEncryptionAlgorithm
	A128GCM_ECDHES             KeyPoolAlgorithm = "A128GCM/ECDH-ES"              // KeyEncryptionAlgorithm
	A256GCM_dir                KeyPoolAlgorithm = "A256GCM/dir"                  // KeyEncryptionAlgorithm
	A192GCM_dir                KeyPoolAlgorithm = "A192GCM/dir"                  // KeyEncryptionAlgorithm
	A128GCM_dir                KeyPoolAlgorithm = "A128GCM/dir"                  // KeyEncryptionAlgorithm
	A256CBCHS512_A256KW        KeyPoolAlgorithm = "A256CBC-HS512/A256KW"         // KeyEncryptionAlgorithm
	A192CBCHS384_A256KW        KeyPoolAlgorithm = "A192CBC-HS384/A256KW"         // KeyEncryptionAlgorithm
	A128CBCHS256_A256KW        KeyPoolAlgorithm = "A128CBC-HS256/A256KW"         // KeyEncryptionAlgorithm
	A192CBCHS384_A192KW        KeyPoolAlgorithm = "A192CBC-HS384/A192KW"         // KeyEncryptionAlgorithm
	A128CBCHS256_A192KW        KeyPoolAlgorithm = "A128CBC-HS256/A192KW"         // KeyEncryptionAlgorithm
	A128CBCHS256_A128KW        KeyPoolAlgorithm = "A128CBC-HS256/A128KW"         // KeyEncryptionAlgorithm
	A256CBCHS512_A256GCMKW     KeyPoolAlgorithm = "A256CBC-HS512/A256GCMKW"      // KeyEncryptionAlgorithm
	A192CBCHS384_A256GCMKW     KeyPoolAlgorithm = "A192CBC-HS384/A256GCMKW"      // KeyEncryptionAlgorithm
	A128CBCHS256_A256GCMKW     KeyPoolAlgorithm = "A128CBC-HS256/A256GCMKW"      // KeyEncryptionAlgorithm
	A192CBCHS384_A192GCMKW     KeyPoolAlgorithm = "A192CBC-HS384/A192GCMKW"      // KeyEncryptionAlgorithm
	A128CBCHS256_A192GCMKW     KeyPoolAlgorithm = "A128CBC-HS256/A192GCMKW"      // KeyEncryptionAlgorithm
	A128CBCHS256_A128GCMKW     KeyPoolAlgorithm = "A128CBC-HS256/A128GCMKW"      // KeyEncryptionAlgorithm
	A256CBC_HS512_RSAOAEP512   KeyPoolAlgorithm = "A256CBC-HS512/RSA-OAEP-512"   // KeyEncryptionAlgorithm
	A192CBC_HS384_RSAOAEP512   KeyPoolAlgorithm = "A192CBC-HS384/RSA-OAEP-512"   // KeyEncryptionAlgorithm
	A128CBC_HS256_RSAOAEP512   KeyPoolAlgorithm = "A128CBC-HS256/RSA-OAEP-512"   // KeyEncryptionAlgorithm
	A192CBC_HS384_RSAOAEP384   KeyPoolAlgorithm = "A192CBC-HS384/RSA-OAEP-384"   // KeyEncryptionAlgorithm
	A128CBC_HS256_RSAOAEP384   KeyPoolAlgorithm = "A128CBC-HS256/RSA-OAEP-384"   // KeyEncryptionAlgorithm
	A128CBC_HS256_RSAOAEP256   KeyPoolAlgorithm = "A128CBC-HS256/RSA-OAEP-256"   // KeyEncryptionAlgorithm
	A128CBC_HS256_RSAOAEP      KeyPoolAlgorithm = "A128CBC-HS256/RSA-OAEP"       // KeyEncryptionAlgorithm
	A128CBC_HS256_RSA15        KeyPoolAlgorithm = "A128CBC-HS256/RSA1_5"         // KeyEncryptionAlgorithm
	A256CBC_HS512_ECDHESA256KW KeyPoolAlgorithm = "A256CBC-HS512/ECDH-ES+A256KW" // KeyEncryptionAlgorithm
	A192CBC_HS384_ECDHESA256KW KeyPoolAlgorithm = "A192CBC-HS384/ECDH-ES+A256KW" // KeyEncryptionAlgorithm
	A128CBC_HS256_ECDHESA256KW KeyPoolAlgorithm = "A128CBC-HS256/ECDH-ES+A256KW" // KeyEncryptionAlgorithm
	A192CBC_HS384_ECDHESA192KW KeyPoolAlgorithm = "A192CBC-HS384/ECDH-ES+A192KW" // KeyEncryptionAlgorithm
	A128CBC_HS256_ECDHESA192KW KeyPoolAlgorithm = "A128CBC-HS256/ECDH-ES+A192KW" // KeyEncryptionAlgorithm
	A128CBC_HS256_ECDHESA128KW KeyPoolAlgorithm = "A128CBC-HS256/ECDH-ES+A128KW" // KeyEncryptionAlgorithm
	A128CBC_HS256_ECDHES       KeyPoolAlgorithm = "A128CBC-HS256/ECDH-ES"        // KeyEncryptionAlgorithm
	A256CBCHS512_dir           KeyPoolAlgorithm = "A256CBC-HS512/dir"            // KeyEncryptionAlgorithm
	A192CBCHS384_dir           KeyPoolAlgorithm = "A192CBC-HS384/dir"            // KeyEncryptionAlgorithm
	A128CBCHS256_dir           KeyPoolAlgorithm = "A128CBC-HS256/dir"            // KeyEncryptionAlgorithm
	RS256                      KeyPoolAlgorithm = "RS256"                        // SignatureAlgorithm
	RS384                      KeyPoolAlgorithm = "RS384"                        // SignatureAlgorithm
	RS512                      KeyPoolAlgorithm = "RS512"                        // SignatureAlgorithm
	PS256                      KeyPoolAlgorithm = "PS256"                        // SignatureAlgorithm
	PS384                      KeyPoolAlgorithm = "PS384"                        // SignatureAlgorithm
	PS512                      KeyPoolAlgorithm = "PS512"                        // SignatureAlgorithm
	ES256                      KeyPoolAlgorithm = "ES256"                        // SignatureAlgorithm
	ES384                      KeyPoolAlgorithm = "ES384"                        // SignatureAlgorithm
	ES512                      KeyPoolAlgorithm = "ES512"                        // SignatureAlgorithm
	HS256                      KeyPoolAlgorithm = "HS256"                        // SignatureAlgorithm
	HS384                      KeyPoolAlgorithm = "HS384"                        // SignatureAlgorithm
	HS512                      KeyPoolAlgorithm = "HS512"                        // SignatureAlgorithm
	EdDSA                      KeyPoolAlgorithm = "EdDSA"                        // SignatureAlgorithm
)

type KeyPoolProvider string

const (
	Internal KeyPoolProvider = "Internal"
)

type KeyPoolStatus string

const (
	Creating                       KeyPoolStatus = "creating"
	ImportFailed                   KeyPoolStatus = "import_failed"
	PendingImport                  KeyPoolStatus = "pending_import"
	PendingGenerate                KeyPoolStatus = "pending_generate"
	GenerateFailed                 KeyPoolStatus = "generate_failed"
	Active                         KeyPoolStatus = "active"
	Disabled                       KeyPoolStatus = "disabled"
	PendingDeleteWasImportFailed   KeyPoolStatus = "pending_delete_was_import_failed"
	PendingDeleteWasPendingImport  KeyPoolStatus = "pending_delete_was_pending_import"
	PendingDeleteWasActive         KeyPoolStatus = "pending_delete_was_active"
	PendingDeleteWasDisabled       KeyPoolStatus = "pending_delete_was_disabled"
	PendingDeleteWasGenerateFailed KeyPoolStatus = "pending_delete_was_generate_failed"
	StartedDelete                  KeyPoolStatus = "started_delete"
	FinishedDelete                 KeyPoolStatus = "finished_delete"
)

type (
	KeyPoolDescription       string
	KeyPoolId                string
	KeyPoolExportAllowed     bool
	KeyPoolImportAllowed     bool
	KeyPoolVersioningAllowed bool
	KeyPoolName              string
)
