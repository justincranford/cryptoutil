package lint_deployments

// Deployment type constants.
const (
	DeploymentTypeSuite          = "SUITE"
	DeploymentTypeProduct        = "PRODUCT"
	DeploymentTypeProductService = "PRODUCT-SERVICE"
	DeploymentTypeInfrastructure = "infrastructure"
	DeploymentTypeTemplate       = "template"
)

// Product count for products that currently have one service (sm, pki, jose).
// Note: These products may have multiple services in the future.
const productsWithOneServiceCount = 3

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
// See: docs/ARCHITECTURE.md Section 12.4 "Deployment Structure Validation".
func GetExpectedStructures() map[string]DeploymentStructure {
	return map[string]DeploymentStructure{
		"template": {
			Name:          "Template deployment (deployments/template/)",
			RequiredDirs:  []string{"secrets"},
			RequiredFiles: []string{"compose.yml"},
			RequiredSecrets: []string{
				"hash_pepper_v3.secret", "unseal_1of5.secret", "unseal_2of5.secret",
				"unseal_3of5.secret", "unseal_4of5.secret", "unseal_5of5.secret",
				"postgres_username.secret", "postgres_password.secret",
				"postgres_database.secret", "postgres_url.secret",
			},
			AllowedExtensions: []string{".yml", ".yaml", ".secret", ".md"},
		},
		DeploymentTypeProductService: {
			Name:          "PRODUCT-SERVICE deployment (e.g., jose-ja, sm-im)",
			RequiredDirs:  []string{"secrets", "config"},
			RequiredFiles: []string{"compose.yml", "Dockerfile"},
			OptionalFiles: []string{}, // no optional files
			RequiredSecrets: []string{
				"hash_pepper_v3.secret", "unseal_1of5.secret", "unseal_2of5.secret",
				"unseal_3of5.secret", "unseal_4of5.secret", "unseal_5of5.secret",
				"postgres_username.secret", "postgres_password.secret",
				"postgres_database.secret", "postgres_url.secret",
			},
			AllowedExtensions: []string{".yml", ".yaml", ".secret", ".md"},
		},
		DeploymentTypeProduct: {
			Name:              "PRODUCT-level deployment (e.g., identity, sm, pki, jose)",
			RequiredDirs:      []string{"secrets"},
			RequiredFiles:     []string{"compose.yml"},
			OptionalFiles:     []string{}, // no optional files
			RequiredSecrets:   []string{}, // Validated by validateProductSecrets() with product-specific prefixes
			AllowedExtensions: []string{".yml", ".yaml", ".secret", ".never", ".md"},
		},
		DeploymentTypeSuite: {
			Name:              "SUITE-level deployment (cryptoutil-suite - all 10 services)",
			RequiredDirs:      []string{"secrets"},
			RequiredFiles:     []string{"compose.yml", "Dockerfile"},
			OptionalFiles:     []string{}, // no optional files
			RequiredSecrets:   []string{}, // Validated by validateSuiteSecrets() with suite-specific prefixes
			AllowedExtensions: []string{".yml", ".yaml", ".secret", ".never", ".md"},
		},
		"infrastructure": {
			Name:              "Infrastructure deployment (postgres, citus, telemetry)",
			RequiredDirs:      []string{},
			RequiredFiles:     []string{"compose.yml"},
			OptionalFiles:     []string{"init-db.sql", "init-citus.sql", "README.md"},
			RequiredSecrets:   []string{}, // Infrastructure secrets are optional
			AllowedExtensions: []string{".yml", ".yaml", ".sql", ".md"},
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
