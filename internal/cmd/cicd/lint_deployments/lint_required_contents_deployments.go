package lint_deployments

const (
	// RequiredFileStatus indicates a file MUST exist.
	RequiredFileStatus = "REQUIRED"
	// OptionalFileStatus indicates a file MAY exist.
	OptionalFileStatus = "OPTIONAL"
	// ForbiddenFileStatus indicates a file MUST NOT exist (deprecated/removed).
	ForbiddenFileStatus = "FORBIDDEN"
)

// GetDeploymentDirectories returns lists of deployment directories by level.
// See: docs/ARCHITECTURE.md Section 12.4 "Deployment Structure Validation".
func GetDeploymentDirectories() (suite []string, product []string, productService []string, infrastructure []string, template []string) {
	// SUITE-level deployment (all 9 services)
	suite = []string{
		"cryptoutil-suite",
	}

	// PRODUCT-level deployments (aggregation of services within product)
	product = []string{
		"identity", // Multi-service product (5 identity services, more later)
		"jose",     // Multi-service product (1 jose service at this time, more later)
		"pki",      // Multi-service product (1 pki service at this time, more later)
		"sm",       // Multi-service product (2 sm services: sm-kms, sm-im)
	}

	// PRODUCT-SERVICE level deployments (individual services)
	productService = []string{
		"sm-im",
		"identity-authz",
		"identity-idp",
		"identity-rp",
		"identity-rs",
		"identity-spa",
		"jose-ja",
		"pki-ca",
		"sm-kms",
	}

	// Infrastructure deployments (shared resources)
	infrastructure = []string{
		"shared-citus",     // CitusDB cluster infrastructure
		"shared-postgres",  // PostgreSQL leader/follower infrastructure
		"shared-telemetry", // Telemetry infrastructure
	}

	// Template deployment (for creating new services)
	template = []string{
		"template",
	}

	return suite, product, productService, infrastructure, template
}

// GetExpectedDeploymentsContents returns the complete expected file structure for deployments/.
// This is the single source of truth for what files MUST/MAY exist in each deployment directory.
//
// Format: map[relativePathFromDeploymentsRoot]fileStatus where fileStatus is:
//   - RequiredFileStatus - file MUST exist (missing = error)
//   - ForbiddenFileStatus - file MUST NOT exist (present = error) - used for deprecated files
//   - OptionalFileStatus - file MAY exist (used for documentation, optional configs)
//
// See: docs/ARCHITECTURE.md Section 12.3.4 "Multi-Level Deployment Hierarchy"
// See: docs/ARCHITECTURE-COMPOSE-MULTIDEPLOY.md for complete hierarchy documentation.
func GetExpectedDeploymentsContents() map[string]string {
	contents := make(map[string]string)

	// SUITE Level (cryptoutil-suite) - ALL 9 services
	// Required: compose.yml, Dockerfile, hash_pepper only
	// Forbidden (via .never): unseal keys, postgres secrets (MUST be service-specific)
	contents["cryptoutil-suite/compose.yml"] = RequiredFileStatus
	contents["cryptoutil-suite/Dockerfile"] = RequiredFileStatus
	contents["cryptoutil-suite/secrets/cryptoutil-hash_pepper.secret"] = RequiredFileStatus
	contents["cryptoutil-suite/secrets/cryptoutil-unseal_1of5.secret.never"] = RequiredFileStatus
	contents["cryptoutil-suite/secrets/cryptoutil-unseal_2of5.secret.never"] = RequiredFileStatus
	contents["cryptoutil-suite/secrets/cryptoutil-unseal_3of5.secret.never"] = RequiredFileStatus
	contents["cryptoutil-suite/secrets/cryptoutil-unseal_4of5.secret.never"] = RequiredFileStatus
	contents["cryptoutil-suite/secrets/cryptoutil-unseal_5of5.secret.never"] = RequiredFileStatus
	contents["cryptoutil-suite/secrets/cryptoutil-postgres_username.secret.never"] = RequiredFileStatus
	contents["cryptoutil-suite/secrets/cryptoutil-postgres_password.secret.never"] = RequiredFileStatus
	contents["cryptoutil-suite/secrets/cryptoutil-postgres_database.secret.never"] = RequiredFileStatus
	contents["cryptoutil-suite/secrets/cryptoutil-postgres_url.secret.never"] = RequiredFileStatus

	// PRODUCT Level - identity (multi-service product: 5 identity services)
	contents["identity/compose.yml"] = RequiredFileStatus
	contents["identity/secrets/identity-hash_pepper.secret"] = RequiredFileStatus
	contents["identity/secrets/identity-unseal_1of5.secret.never"] = RequiredFileStatus
	contents["identity/secrets/identity-unseal_2of5.secret.never"] = RequiredFileStatus
	contents["identity/secrets/identity-unseal_3of5.secret.never"] = RequiredFileStatus
	contents["identity/secrets/identity-unseal_4of5.secret.never"] = RequiredFileStatus
	contents["identity/secrets/identity-unseal_5of5.secret.never"] = RequiredFileStatus
	contents["identity/secrets/identity-postgres_username.secret.never"] = RequiredFileStatus
	contents["identity/secrets/identity-postgres_password.secret.never"] = RequiredFileStatus
	contents["identity/secrets/identity-postgres_database.secret.never"] = RequiredFileStatus
	contents["identity/secrets/identity-postgres_url.secret.never"] = RequiredFileStatus

	// PRODUCT Level - jose (currently single service: jose-ja)
	contents["jose/compose.yml"] = RequiredFileStatus
	contents["jose/secrets/jose-hash_pepper.secret"] = RequiredFileStatus
	contents["jose/secrets/jose-unseal_1of5.secret.never"] = RequiredFileStatus
	contents["jose/secrets/jose-unseal_2of5.secret.never"] = RequiredFileStatus
	contents["jose/secrets/jose-unseal_3of5.secret.never"] = RequiredFileStatus
	contents["jose/secrets/jose-unseal_4of5.secret.never"] = RequiredFileStatus
	contents["jose/secrets/jose-unseal_5of5.secret.never"] = RequiredFileStatus
	contents["jose/secrets/jose-postgres_username.secret.never"] = RequiredFileStatus
	contents["jose/secrets/jose-postgres_password.secret.never"] = RequiredFileStatus
	contents["jose/secrets/jose-postgres_database.secret.never"] = RequiredFileStatus
	contents["jose/secrets/jose-postgres_url.secret.never"] = RequiredFileStatus

	// PRODUCT Level - pki (currently single service: pki-ca)
	contents["pki/compose.yml"] = RequiredFileStatus
	contents["pki/secrets/pki-hash_pepper.secret"] = RequiredFileStatus
	contents["pki/secrets/pki-unseal_1of5.secret.never"] = RequiredFileStatus
	contents["pki/secrets/pki-unseal_2of5.secret.never"] = RequiredFileStatus
	contents["pki/secrets/pki-unseal_3of5.secret.never"] = RequiredFileStatus
	contents["pki/secrets/pki-unseal_4of5.secret.never"] = RequiredFileStatus
	contents["pki/secrets/pki-unseal_5of5.secret.never"] = RequiredFileStatus
	contents["pki/secrets/pki-postgres_username.secret.never"] = RequiredFileStatus
	contents["pki/secrets/pki-postgres_password.secret.never"] = RequiredFileStatus
	contents["pki/secrets/pki-postgres_database.secret.never"] = RequiredFileStatus
	contents["pki/secrets/pki-postgres_url.secret.never"] = RequiredFileStatus

	// PRODUCT Level - sm (services: sm-kms, sm-im)
	contents["sm/compose.yml"] = RequiredFileStatus
	contents["sm/secrets/sm-hash_pepper.secret"] = RequiredFileStatus
	contents["sm/secrets/sm-unseal_1of5.secret.never"] = RequiredFileStatus
	contents["sm/secrets/sm-unseal_2of5.secret.never"] = RequiredFileStatus
	contents["sm/secrets/sm-unseal_3of5.secret.never"] = RequiredFileStatus
	contents["sm/secrets/sm-unseal_4of5.secret.never"] = RequiredFileStatus
	contents["sm/secrets/sm-unseal_5of5.secret.never"] = RequiredFileStatus
	contents["sm/secrets/sm-postgres_username.secret.never"] = RequiredFileStatus
	contents["sm/secrets/sm-postgres_password.secret.never"] = RequiredFileStatus
	contents["sm/secrets/sm-postgres_database.secret.never"] = RequiredFileStatus
	contents["sm/secrets/sm-postgres_url.secret.never"] = RequiredFileStatus

	// PRODUCT-SERVICE Level - sm-im
	addProductServiceFiles(&contents, "sm-im")

	// PRODUCT-SERVICE Level - identity services (5 services)
	addProductServiceFiles(&contents, "identity-authz")
	addProductServiceFiles(&contents, "identity-idp")
	addProductServiceFiles(&contents, "identity-rp")
	addProductServiceFiles(&contents, "identity-rs")
	addProductServiceFiles(&contents, "identity-spa")

	// PRODUCT-SERVICE Level - jose-ja
	addProductServiceFiles(&contents, "jose-ja")

	// PRODUCT-SERVICE Level - pki-ca
	addProductServiceFiles(&contents, "pki-ca")

	// PRODUCT-SERVICE Level - sm-kms
	addProductServiceFiles(&contents, "sm-kms")

	// Infrastructure deployments
	addInfrastructureFiles(&contents, "shared-citus")
	addInfrastructureFiles(&contents, "shared-postgres")
	addInfrastructureFiles(&contents, "shared-telemetry")

	// Template deployment
	addTemplateFiles(&contents)

	return contents
}

// addProductServiceFiles adds required files for a PRODUCT-SERVICE deployment.
// Pattern: 10 secrets (5 unseal, 1 hash_pepper, 4 postgres) + compose.yml + Dockerfile + 4 config files.
func addProductServiceFiles(contents *map[string]string, serviceName string) {
	// Required root files
	(*contents)[serviceName+"/compose.yml"] = RequiredFileStatus
	(*contents)[serviceName+"/Dockerfile"] = RequiredFileStatus

	// Required secrets (10 total)
	(*contents)[serviceName+"/secrets/"+serviceName+"-hash_pepper.secret"] = RequiredFileStatus
	(*contents)[serviceName+"/secrets/"+serviceName+"-unseal_1of5.secret"] = RequiredFileStatus
	(*contents)[serviceName+"/secrets/"+serviceName+"-unseal_2of5.secret"] = RequiredFileStatus
	(*contents)[serviceName+"/secrets/"+serviceName+"-unseal_3of5.secret"] = RequiredFileStatus
	(*contents)[serviceName+"/secrets/"+serviceName+"-unseal_4of5.secret"] = RequiredFileStatus
	(*contents)[serviceName+"/secrets/"+serviceName+"-unseal_5of5.secret"] = RequiredFileStatus
	(*contents)[serviceName+"/secrets/"+serviceName+"-postgres_username.secret"] = RequiredFileStatus
	(*contents)[serviceName+"/secrets/"+serviceName+"-postgres_password.secret"] = RequiredFileStatus
	(*contents)[serviceName+"/secrets/"+serviceName+"-postgres_database.secret"] = RequiredFileStatus
	(*contents)[serviceName+"/secrets/"+serviceName+"-postgres_url.secret"] = RequiredFileStatus

	// Required config files (4 standard files per Section 12.4.5)
	(*contents)[serviceName+"/config/"+serviceName+"-app-common.yml"] = RequiredFileStatus
	(*contents)[serviceName+"/config/"+serviceName+"-app-sqlite-1.yml"] = RequiredFileStatus
	(*contents)[serviceName+"/config/"+serviceName+"-app-postgresql-1.yml"] = RequiredFileStatus
	(*contents)[serviceName+"/config/"+serviceName+"-app-postgresql-2.yml"] = RequiredFileStatus

	// FORBIDDEN deprecated config files (Section 12.4.6)
	(*contents)[serviceName+"/config/demo-seed.yml"] = ForbiddenFileStatus
	(*contents)[serviceName+"/config/integration.yml"] = ForbiddenFileStatus
	(*contents)[serviceName+"/config/"+serviceName+"-demo.yml"] = ForbiddenFileStatus
	(*contents)[serviceName+"/config/"+serviceName+"-e2e.yml"] = ForbiddenFileStatus
}

// addInfrastructureFiles adds required files for infrastructure deployments.
func addInfrastructureFiles(contents *map[string]string, infraName string) {
	(*contents)[infraName+"/compose.yml"] = RequiredFileStatus

	// Optional files for specific infrastructure types
	if infraName == "shared-postgres" {
		(*contents)[infraName+"/init-db.sql"] = OptionalFileStatus
	}

	if infraName == "shared-citus" {
		(*contents)[infraName+"/init-citus.sql"] = OptionalFileStatus
	}
}

// addTemplateFiles adds required files for template deployment.
func addTemplateFiles(contents *map[string]string) {
	(*contents)["template/compose.yml"] = RequiredFileStatus

	// Template has same secret structure as PRODUCT-SERVICE for reference
	(*contents)["template/secrets/hash_pepper_v3.secret"] = RequiredFileStatus
	(*contents)["template/secrets/unseal_1of5.secret"] = RequiredFileStatus
	(*contents)["template/secrets/unseal_2of5.secret"] = RequiredFileStatus
	(*contents)["template/secrets/unseal_3of5.secret"] = RequiredFileStatus
	(*contents)["template/secrets/unseal_4of5.secret"] = RequiredFileStatus
	(*contents)["template/secrets/unseal_5of5.secret"] = RequiredFileStatus
	(*contents)["template/secrets/postgres_username.secret"] = RequiredFileStatus
	(*contents)["template/secrets/postgres_password.secret"] = RequiredFileStatus
	(*contents)["template/secrets/postgres_database.secret"] = RequiredFileStatus
	(*contents)["template/secrets/postgres_url.secret"] = RequiredFileStatus
}
