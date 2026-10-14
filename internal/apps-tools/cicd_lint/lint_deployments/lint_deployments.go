package lint_deployments

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Deployment type constants.
const (
	DeploymentTypeSuite          = "SUITE"
	DeploymentTypeProduct        = "PRODUCT"
	DeploymentTypeProductService = "PRODUCT-SERVICE"
	DeploymentTypeInfrastructure = "infrastructure"
	DeploymentTypeTemplate       = cryptoutilSharedMagic.SkeletonTemplateServiceName
	deploymentDockerfileName     = "Dockerfile"
)

// DeploymentStructure defines expected directory structure for each deployment type.
type DeploymentStructure struct {
	Name              string
	RequiredDirs      []string
	RequiredFiles     []string
	OptionalFiles     []string
	RequiredSecrets   []string
	AllowedExtensions []string
}

// GetExpectedStructures returns validation rules for different deployment types.
// See: docs/ENG-HANDBOOK.md Section 12.4 "Deployment Structure Validation".
func GetExpectedStructures() map[string]DeploymentStructure {
	return map[string]DeploymentStructure{
		cryptoutilSharedMagic.SkeletonTemplateServiceName: {
			Name:          "Template deployment (deployments/template/)",
			RequiredDirs:  []string{validatorNameSecrets},
			RequiredFiles: []string{cryptoutilSharedMagic.COMPOSE_YML},
			RequiredSecrets: []string{
				valueHashPepperV3Secret, valueUnseal1of5Secret, valueUnseal2of5Secret,
				valueUnseal3of5Secret, valueUnseal4of5Secret, valueUnseal5of5Secret,
				valuePostgresUsernameSecret, valuePostgresPasswordSecret,
				valuePostgresDatabaseSecret, valuePostgresURLSecret,
			},
			AllowedExtensions: []string{valueYml, valueYAML, cryptoutilSharedMagic.CICDTemplateSecretFileSuffix, valueMd},
		},
		DeploymentTypeProductService: {
			Name:          "PRODUCT-SERVICE deployment (e.g., sm-kms, sm-kms)",
			RequiredDirs:  []string{validatorNameSecrets, fileTypeConfig},
			RequiredFiles: []string{cryptoutilSharedMagic.COMPOSE_YML, deploymentDockerfileName},
			OptionalFiles: []string{}, // no optional files
			RequiredSecrets: []string{
				valueHashPepperV3Secret, valueUnseal1of5Secret, valueUnseal2of5Secret,
				valueUnseal3of5Secret, valueUnseal4of5Secret, valueUnseal5of5Secret,
				valuePostgresUsernameSecret, valuePostgresPasswordSecret,
				valuePostgresDatabaseSecret, valuePostgresURLSecret,
			},
			AllowedExtensions: []string{valueYml, valueYAML, cryptoutilSharedMagic.CICDTemplateSecretFileSuffix, valueMd},
		},
		DeploymentTypeProduct: {
			Name:              "PRODUCT-level deployment (e.g., identity, sm, pki, jose, skeleton)",
			RequiredDirs:      []string{validatorNameSecrets},
			RequiredFiles:     []string{cryptoutilSharedMagic.COMPOSE_YML},
			OptionalFiles:     []string{}, // no optional files
			RequiredSecrets:   []string{}, // Validated by validateProductSecrets() with product-specific prefixes
			AllowedExtensions: []string{valueYml, valueYAML, cryptoutilSharedMagic.CICDTemplateSecretFileSuffix, cryptoutilSharedMagic.CICDTemplateSecretFileNeverSuffix, valueMd},
		},
		DeploymentTypeSuite: {
			Name:              "SUITE-level deployment (cryptoutil - all 10 services)",
			RequiredDirs:      []string{validatorNameSecrets},
			RequiredFiles:     []string{cryptoutilSharedMagic.COMPOSE_YML, deploymentDockerfileName},
			OptionalFiles:     []string{}, // no optional files
			RequiredSecrets:   []string{}, // Validated by validateSuiteSecrets() with suite-specific prefixes
			AllowedExtensions: []string{valueYml, valueYAML, cryptoutilSharedMagic.CICDTemplateSecretFileSuffix, cryptoutilSharedMagic.CICDTemplateSecretFileNeverSuffix, valueMd},
		},
		"infrastructure": {
			Name:              "Infrastructure deployment (postgres, telemetry)",
			RequiredDirs:      []string{},
			RequiredFiles:     []string{cryptoutilSharedMagic.COMPOSE_YML},
			OptionalFiles:     []string{"init-db.sql", valueReadmeMd},
			RequiredSecrets:   []string{}, // Infrastructure secrets are optional
			AllowedExtensions: []string{valueYml, valueYAML, ".sql", valueMd},
		},
	}
}

// ValidationResult holds validation outcome for a directory.
type ValidationResult struct {
	Path           string
	Type           string
	Valid          bool
	MissingDirs    []string
	MissingFiles   []string
	MissingSecrets []string
	Errors         []string
	Warnings       []string
}
