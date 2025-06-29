package businessmodel

import (
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"

	"fmt"
	"strings"
)

func ToElasticKeyInitialStatus(isImportAllowed bool) *cryptoutilOpenapiModel.ElasticKeyStatus {
	var ormElasticKeyStatus cryptoutilOpenapiModel.ElasticKeyStatus
	if isImportAllowed {
		ormElasticKeyStatus = cryptoutilOpenapiModel.PendingImport
	} else {
		ormElasticKeyStatus = cryptoutilOpenapiModel.PendingGenerate
	}
	return &ormElasticKeyStatus
}

func ToElasticKeyName(name string) (*cryptoutilOpenapiModel.ElasticKeyName, error) {
	if err := validateString(name); err != nil {
		return nil, fmt.Errorf("invalid elastic Key name: %w", err)
	}
	elasticKeyName := cryptoutilOpenapiModel.ElasticKeyName(name)
	return &elasticKeyName, nil
}

func ToElasticKeyDescription(description string) (*cryptoutilOpenapiModel.ElasticKeyDescription, error) {
	if err := validateString(description); err != nil {
		return nil, fmt.Errorf("invalid elastic Key description: %w", err)
	}
	elasticKeyDescription := cryptoutilOpenapiModel.ElasticKeyDescription(description)
	return &elasticKeyDescription, nil
}

func ToElasticKeyAlgorithm(algorithm string) (*cryptoutilOpenapiModel.ElasticKeyAlgorithm, error) {
	if err := validateString(algorithm); err != nil {
		return nil, fmt.Errorf("invalid elastic Key algorithm: %w", err)
	}
	if alg, exists := elasticKeyAlgorithms[algorithm]; exists {
		return &alg, nil
	}
	return nil, fmt.Errorf("invalid elastic Key algorithm: %s", algorithm)
}

func IsSymmetric(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) (bool, error) {
	isSymmetric, ok := symmetricElasticKeyAlgorithm[*elasticKeyAlgorithm]
	if ok {
		return isSymmetric, nil
	}
	return false, fmt.Errorf("unsupported ElasticKeyAlgorithm '%s'", *elasticKeyAlgorithm)
}

func IsAsymmetric(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) (bool, error) {
	isAsymmetric, ok := asymmetricElasticKeyAlgorithm[*elasticKeyAlgorithm]
	if ok {
		return isAsymmetric, nil
	}
	return false, fmt.Errorf("unsupported ElasticKeyAlgorithm '%s'", *elasticKeyAlgorithm)
}

func ToElasticKeyProvider(provider string) (*cryptoutilOpenapiModel.ElasticKeyProvider, error) {
	if err := validateString(provider); err != nil {
		return nil, fmt.Errorf("invalid elastic Key provider value: %w", err)
	}
	var elasticKeyProvider cryptoutilOpenapiModel.ElasticKeyProvider
	switch provider {
	case string(cryptoutilOpenapiModel.Internal):
		elasticKeyProvider = cryptoutilOpenapiModel.Internal
	default:
		return nil, fmt.Errorf("invalid elastic Key provider option: %s", provider)
	}
	return &elasticKeyProvider, nil
}

func ToElasticKeyImportAllowed(importAllowed bool) *cryptoutilOpenapiModel.ElasticKeyImportAllowed {
	elasticKeyElasticKeyImportAllowed := cryptoutilOpenapiModel.ElasticKeyImportAllowed(importAllowed)
	return &elasticKeyElasticKeyImportAllowed
}

func ToElasticKeyExportAllowed(exportAllowed bool) *cryptoutilOpenapiModel.ElasticKeyExportAllowed {
	elasticKeyElasticKeyExportAllowed := cryptoutilOpenapiModel.ElasticKeyExportAllowed(exportAllowed)
	return &elasticKeyElasticKeyExportAllowed
}

func ToElasticKeyVersioningAllowed(versioningAllowed bool) *cryptoutilOpenapiModel.ElasticKeyVersioningAllowed {
	elasticKeyElasticKeyVersioningAllowed := cryptoutilOpenapiModel.ElasticKeyVersioningAllowed(versioningAllowed)
	return &elasticKeyElasticKeyVersioningAllowed
}

func validateString(value string) error {
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
