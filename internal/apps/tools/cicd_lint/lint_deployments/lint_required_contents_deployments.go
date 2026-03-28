package lint_deployments

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

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
	// SUITE-level deployment (all 10 services)
	suite = []string{
		"cryptoutil-suite",
	}

	// PRODUCT-level deployments (aggregation of services within product)
	product = []string{
		cryptoutilSharedMagic.IdentityProductName, // Multi-service product (5 identity services, more later)
		cryptoutilSharedMagic.JoseProductName,     // Multi-service product (1 jose service at this time, more later)
		cryptoutilSharedMagic.PKIProductName,      // Multi-service product (1 pki service at this time, more later)
		cryptoutilSharedMagic.SMProductName,       // Multi-service product (2 sm services: sm-kms, sm-im)
		cryptoutilSharedMagic.SkeletonProductName, // Single-service product (1 skeleton service: skeleton-template)
	}

	// PRODUCT-SERVICE level deployments (individual services)
	productService = []string{
		cryptoutilSharedMagic.OTLPServiceSMIM,
		cryptoutilSharedMagic.OTLPServiceIdentityAuthz,
		cryptoutilSharedMagic.OTLPServiceIdentityIDP,
		cryptoutilSharedMagic.OTLPServiceIdentityRP,
		cryptoutilSharedMagic.OTLPServiceIdentityRS,
		cryptoutilSharedMagic.OTLPServiceIdentitySPA,
		cryptoutilSharedMagic.OTLPServiceJoseJA,
		cryptoutilSharedMagic.OTLPServicePKICA,
		cryptoutilSharedMagic.OTLPServiceSMKMS,
		cryptoutilSharedMagic.OTLPServiceSkeletonTemplate,
	}

	// Infrastructure deployments (shared resources)
	infrastructure = []string{
		"shared-postgres",  // PostgreSQL leader/follower infrastructure
		"shared-telemetry", // Telemetry infrastructure
	}

	// Template deployment (for creating new services)
	template = []string{
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
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
// See: docs/ARCHITECTURE.md Section 12.3.4 "Multi-Level Deployment Hierarchy".
func GetExpectedDeploymentsContents() map[string]string {
	contents := make(map[string]string)

	// SUITE Level (cryptoutil-suite) - ALL 10 services.
	contents["cryptoutil-suite/compose.yml"] = RequiredFileStatus
	contents["cryptoutil-suite/Dockerfile"] = RequiredFileStatus
	addSuiteProductSecrets(&contents, "cryptoutil-suite")

	// PRODUCT Level - identity (multi-service product: 5 identity services).
	contents[cryptoutilSharedMagic.IdentityProductName+"/compose.yml"] = RequiredFileStatus
	addSuiteProductSecrets(&contents, cryptoutilSharedMagic.IdentityProductName)

	// PRODUCT Level - jose (currently single service: jose-ja).
	contents[cryptoutilSharedMagic.JoseProductName+"/compose.yml"] = RequiredFileStatus
	addSuiteProductSecrets(&contents, cryptoutilSharedMagic.JoseProductName)

	// PRODUCT Level - pki (currently single service: pki-ca).
	contents[cryptoutilSharedMagic.PKIProductName+"/compose.yml"] = RequiredFileStatus
	addSuiteProductSecrets(&contents, cryptoutilSharedMagic.PKIProductName)

	// PRODUCT Level - sm (services: sm-kms, sm-im).
	contents[cryptoutilSharedMagic.SMProductName+"/compose.yml"] = RequiredFileStatus
	addSuiteProductSecrets(&contents, cryptoutilSharedMagic.SMProductName)

	// PRODUCT Level - skeleton (single service: skeleton-template).
	contents[cryptoutilSharedMagic.SkeletonProductName+"/compose.yml"] = RequiredFileStatus
	addSuiteProductSecrets(&contents, cryptoutilSharedMagic.SkeletonProductName)

	// PRODUCT-SERVICE Level - sm-im
	addProductServiceFiles(&contents, cryptoutilSharedMagic.OTLPServiceSMIM)

	// PRODUCT-SERVICE Level - identity services (5 services)
	addProductServiceFiles(&contents, cryptoutilSharedMagic.OTLPServiceIdentityAuthz)
	addProductServiceFiles(&contents, cryptoutilSharedMagic.OTLPServiceIdentityIDP)
	addProductServiceFiles(&contents, cryptoutilSharedMagic.OTLPServiceIdentityRP)
	addProductServiceFiles(&contents, cryptoutilSharedMagic.OTLPServiceIdentityRS)
	addProductServiceFiles(&contents, cryptoutilSharedMagic.OTLPServiceIdentitySPA)

	// PRODUCT-SERVICE Level - jose-ja
	addProductServiceFiles(&contents, cryptoutilSharedMagic.OTLPServiceJoseJA)

	// PRODUCT-SERVICE Level - pki-ca
	addProductServiceFiles(&contents, cryptoutilSharedMagic.OTLPServicePKICA)

	// PRODUCT-SERVICE Level - sm-kms
	addProductServiceFiles(&contents, cryptoutilSharedMagic.OTLPServiceSMKMS)

	// PRODUCT-SERVICE Level - skeleton-template
	addProductServiceFiles(&contents, cryptoutilSharedMagic.OTLPServiceSkeletonTemplate)

	// Infrastructure deployments
	addInfrastructureFiles(&contents, "shared-postgres")
	addInfrastructureFiles(&contents, "shared-telemetry")

	// Template deployment
	addTemplateFiles(&contents)

	return contents
}

// addSuiteProductSecrets adds required secrets for SUITE and PRODUCT level deployments.
// Secret filenames use hyphens without deployment-name prefixes.
// Browser/service credentials use .secret.never markers (NEVER real secrets at suite/product level).
func addSuiteProductSecrets(contents *map[string]string, dirName string) {
	prefix := dirName + "/secrets/"

	(*contents)[prefix+"hash-pepper-v3.secret"] = RequiredFileStatus
	(*contents)[prefix+"unseal-1of5.secret"] = RequiredFileStatus
	(*contents)[prefix+"unseal-2of5.secret"] = RequiredFileStatus
	(*contents)[prefix+"unseal-3of5.secret"] = RequiredFileStatus
	(*contents)[prefix+"unseal-4of5.secret"] = RequiredFileStatus
	(*contents)[prefix+"unseal-5of5.secret"] = RequiredFileStatus
	(*contents)[prefix+"postgres-username.secret"] = RequiredFileStatus
	(*contents)[prefix+"postgres-password.secret"] = RequiredFileStatus
	(*contents)[prefix+"postgres-database.secret"] = RequiredFileStatus
	(*contents)[prefix+"postgres-url.secret"] = RequiredFileStatus
	(*contents)[prefix+"browser-password.secret.never"] = RequiredFileStatus
	(*contents)[prefix+"browser-username.secret.never"] = RequiredFileStatus
	(*contents)[prefix+"service-password.secret.never"] = RequiredFileStatus
	(*contents)[prefix+"service-username.secret.never"] = RequiredFileStatus
}

// addProductServiceFiles adds required files for a PRODUCT-SERVICE deployment.
// Secret filenames use hyphens without service-name prefixes.
func addProductServiceFiles(contents *map[string]string, serviceName string) {
	// Required root files.
	(*contents)[serviceName+"/compose.yml"] = RequiredFileStatus
	(*contents)[serviceName+"/Dockerfile"] = RequiredFileStatus

	// Required secrets.
	prefix := serviceName + "/secrets/"

	(*contents)[prefix+"hash-pepper-v3.secret"] = RequiredFileStatus
	(*contents)[prefix+"unseal-1of5.secret"] = RequiredFileStatus
	(*contents)[prefix+"unseal-2of5.secret"] = RequiredFileStatus
	(*contents)[prefix+"unseal-3of5.secret"] = RequiredFileStatus
	(*contents)[prefix+"unseal-4of5.secret"] = RequiredFileStatus
	(*contents)[prefix+"unseal-5of5.secret"] = RequiredFileStatus
	(*contents)[prefix+"postgres-username.secret"] = RequiredFileStatus
	(*contents)[prefix+"postgres-password.secret"] = RequiredFileStatus
	(*contents)[prefix+"postgres-database.secret"] = RequiredFileStatus
	(*contents)[prefix+"postgres-url.secret"] = RequiredFileStatus
	(*contents)[prefix+"browser-password.secret"] = RequiredFileStatus
	(*contents)[prefix+"browser-username.secret"] = RequiredFileStatus
	(*contents)[prefix+"service-password.secret"] = RequiredFileStatus
	(*contents)[prefix+"service-username.secret"] = RequiredFileStatus

	// Required config files (5 standard files per F.1).
	(*contents)[serviceName+"/config/"+serviceName+"-app-common.yml"] = RequiredFileStatus
	(*contents)[serviceName+"/config/"+serviceName+"-app-sqlite-1.yml"] = RequiredFileStatus
	(*contents)[serviceName+"/config/"+serviceName+"-app-sqlite-2.yml"] = RequiredFileStatus
	(*contents)[serviceName+"/config/"+serviceName+"-app-postgresql-1.yml"] = RequiredFileStatus
	(*contents)[serviceName+"/config/"+serviceName+"-app-postgresql-2.yml"] = RequiredFileStatus

	// FORBIDDEN deprecated config files (Section 12.4.6).
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
}

// addTemplateFiles adds required files for template deployment.
// Uses same hyphenated naming convention as product-service secrets.
func addTemplateFiles(contents *map[string]string) {
	(*contents)["template/compose.yml"] = RequiredFileStatus

	// Template has same secret structure as PRODUCT-SERVICE for reference.
	prefix := "template/secrets/"

	(*contents)[prefix+"hash-pepper-v3.secret"] = RequiredFileStatus
	(*contents)[prefix+"unseal-1of5.secret"] = RequiredFileStatus
	(*contents)[prefix+"unseal-2of5.secret"] = RequiredFileStatus
	(*contents)[prefix+"unseal-3of5.secret"] = RequiredFileStatus
	(*contents)[prefix+"unseal-4of5.secret"] = RequiredFileStatus
	(*contents)[prefix+"unseal-5of5.secret"] = RequiredFileStatus
	(*contents)[prefix+"postgres-username.secret"] = RequiredFileStatus
	(*contents)[prefix+"postgres-password.secret"] = RequiredFileStatus
	(*contents)[prefix+"postgres-database.secret"] = RequiredFileStatus
	(*contents)[prefix+"postgres-url.secret"] = RequiredFileStatus
	(*contents)[prefix+"browser-password.secret"] = RequiredFileStatus
	(*contents)[prefix+"browser-username.secret"] = RequiredFileStatus
	(*contents)[prefix+"service-password.secret"] = RequiredFileStatus
	(*contents)[prefix+"service-username.secret"] = RequiredFileStatus
}
