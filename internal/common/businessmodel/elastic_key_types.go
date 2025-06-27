package businessmodel

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

func ToElasticKeyInitialStatus(isImportAllowed bool) *ElasticKeyStatus {
	var ormElasticKeyStatus ElasticKeyStatus
	if isImportAllowed {
		ormElasticKeyStatus = PendingImport
	} else {
		ormElasticKeyStatus = PendingGenerate
	}
	return &ormElasticKeyStatus
}
