package businessmodel

import (
	"fmt"
	"strings"
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

type (
	ElasticKeyId                string
	ElasticKeyName              string
	ElasticKeyDescription       string
	ElasticKeyProvider          string
	ElasticKeyStatus            string
	ElasticKeyImportAllowed     bool
	ElasticKeyExportAllowed     bool
	ElasticKeyVersioningAllowed bool
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

func MapElasticKeyName(name string) (*ElasticKeyName, error) {
	if err := ValidateString(name); err != nil {
		return nil, fmt.Errorf("invalid elastic Key name: %w", err)
	}
	elasticKeyName := ElasticKeyName(name)
	return &elasticKeyName, nil
}

func MapElasticKeyDescription(description string) (*ElasticKeyDescription, error) {
	if err := ValidateString(description); err != nil {
		return nil, fmt.Errorf("invalid elastic Key description: %w", err)
	}
	elasticKeyDescription := ElasticKeyDescription(description)
	return &elasticKeyDescription, nil
}

func MapElasticKeyProvider(provider string) (*ElasticKeyProvider, error) {
	if err := ValidateString(provider); err != nil {
		return nil, fmt.Errorf("invalid elastic Key provider value: %w", err)
	}
	var elasticKeyProvider ElasticKeyProvider
	switch provider {
	case string(Internal):
		elasticKeyProvider = Internal
	default:
		return nil, fmt.Errorf("invalid elastic Key provider option: %s", provider)
	}
	return &elasticKeyProvider, nil
}

func MapElasticKeyImportAllowed(importAllowed bool) *ElasticKeyImportAllowed {
	elasticKeyElasticKeyImportAllowed := ElasticKeyImportAllowed(importAllowed)
	return &elasticKeyElasticKeyImportAllowed
}

func MapElasticKeyExportAllowed(exportAllowed bool) *ElasticKeyExportAllowed {
	elasticKeyElasticKeyExportAllowed := ElasticKeyExportAllowed(exportAllowed)
	return &elasticKeyElasticKeyExportAllowed
}

func MapElasticKeyVersioningAllowed(versioningAllowed bool) *ElasticKeyVersioningAllowed {
	elasticKeyElasticKeyVersioningAllowed := ElasticKeyVersioningAllowed(versioningAllowed)
	return &elasticKeyElasticKeyVersioningAllowed
}

func ValidateString(value string) error {
	length := len(value)
	trimmedLength := len(strings.TrimSpace(value))
	if length == 0 {
		return fmt.Errorf("string can't be empty")
	} else if trimmedLength == 0 {
		return fmt.Errorf("string can't contain all whitespace")
	} else if trimmedLength != length {
		return fmt.Errorf("string can't contain leading or trailing whitespace")
	}
	return nil
}
