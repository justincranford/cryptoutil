// Copyright (c) 2025 Justin Cranford
//
//

package magic

// CICDTemplatesRelPath is the path to the canonical templates directory relative to the project root.
const CICDTemplatesRelPath = "api/cryptosuite-registry/templates"

// Template expansion placeholder keys detected in template paths.
const (
	// CICDTemplateExpansionKeyPSID is the placeholder for a PS-ID in template paths.
	CICDTemplateExpansionKeyPSID = "__PS_ID__"
	// CICDTemplateExpansionKeyProduct is the placeholder for a product ID in template paths.
	CICDTemplateExpansionKeyProduct = "__PRODUCT__"
	// CICDTemplateExpansionKeySuite is the placeholder for a suite ID in template paths.
	CICDTemplateExpansionKeySuite = "__SUITE__"
)

// Canonical template parameter values used when generating expected filesystem content.
const (
	// CICDTemplateBuildDate is the pinned build date embedded in template Dockerfiles.
	CICDTemplateBuildDate = "2026-02-17T00:00:00Z"
	// CICDTemplateGoVersion is the Go version used across all template Dockerfiles.
	CICDTemplateGoVersion = "1.26.1"
	// CICDTemplateAlpineVersion is the Alpine Linux version used in runtime stages.
	CICDTemplateAlpineVersion = "latest"
	// CICDTemplateCGOEnabled is the CGO_ENABLED build argument (always "0" — CGO is banned).
	CICDTemplateCGOEnabled = "0"
	// CICDTemplateContainerUID is the non-root container user ID used in all runtime images.
	CICDTemplateContainerUID = "65532"
	// CICDTemplateContainerGID is the non-root container group ID used in all runtime images.
	CICDTemplateContainerGID = "65532"
	// CICDTemplateGitHubRepoURL is the canonical GitHub repository URL embedded in image labels.
	CICDTemplateGitHubRepoURL = "https://github.com/justincranford/cryptoutil"
	// CICDTemplateAuthors is the canonical author string embedded in image labels.
	CICDTemplateAuthors = "Justin Cranford"
)

// Healthcheck template parameter values used in Docker Compose service definitions.
const (
	// CICDTemplateHealthcheckInterval is the Docker Compose healthcheck interval.
	CICDTemplateHealthcheckInterval = "30s"
	// CICDTemplateHealthcheckTimeout is the Docker Compose healthcheck timeout.
	CICDTemplateHealthcheckTimeout = "10s"
	// CICDTemplateHealthcheckStartPeriod is the Docker Compose healthcheck start period.
	CICDTemplateHealthcheckStartPeriod = "30s"
	// CICDTemplateHealthcheckRetries is the Docker Compose healthcheck retry count.
	CICDTemplateHealthcheckRetries = "3"
)

// Compose variant name constants for the four service instances per PS-ID.
const (
	// CICDTemplateVariantSQLite1 is the compose service suffix for the first SQLite instance.
	CICDTemplateVariantSQLite1 = "sqlite-1"
	// CICDTemplateVariantSQLite2 is the compose service suffix for the second SQLite instance.
	CICDTemplateVariantSQLite2 = "sqlite-2"
	// CICDTemplateVariantPostgres1 is the compose service suffix for the first PostgreSQL instance.
	CICDTemplateVariantPostgres1 = "postgresql-1"
	// CICDTemplateVariantPostgres2 is the compose service suffix for the second PostgreSQL instance.
	CICDTemplateVariantPostgres2 = "postgresql-2"
)

// CICDTemplateBase64Char43Placeholder is the placeholder used in secret template files to
// indicate a position where a base64-encoded value of at least 43 characters should appear.
// The double-underscore wrapping follows the same convention as all other template placeholders.
const CICDTemplateBase64Char43Placeholder = "__BASE64_CHAR43__"

// Template secrets path constants.
const (
	// CICDTemplateSecretsPath is the relative path to the skeleton-template secrets reference directory.
	CICDTemplateSecretsPath = "deployments/skeleton-template/secrets"
	// CICDTemplateSecretFileSuffix is the required extension for Docker secret files.
	CICDTemplateSecretFileSuffix = ".secret"
	// CICDTemplateSecretFileNeverSuffix is the extension for credential marker files that must
	// never contain real secrets at product/suite deployment levels.
	CICDTemplateSecretFileNeverSuffix = ".secret.never"
)
