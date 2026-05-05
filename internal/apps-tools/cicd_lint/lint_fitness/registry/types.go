// Copyright (c) 2025-2026 Justin Cranford.
package registry

// RegistryFile is the top-level structure parsed from registry.yaml.
// It is the canonical machine-readable representation of all suites, products,
// and product-services. Use AllSuites(), AllProducts(), AllProductServices() for
// the API-compatible accessors, and AllPorts(), AllMigrationRanges(),
// AllAPIResources() for the richer fields.
type RegistryFile struct {
	Suites          []RegistrySuite          `yaml:"suites"`
	Products        []RegistryProduct        `yaml:"products"`
	ProductServices []RegistryProductService `yaml:"product_services"`
}

// RegistrySuite is a single suite entry in registry.yaml.
type RegistrySuite struct {
	// ID is the canonical suite identifier (e.g. "cryptoutil").
	ID string `yaml:"id"`
	// DisplayName is the human-readable name (e.g. "Cryptoutil").
	DisplayName string `yaml:"display_name"`
	// CmdDir is the sub-path under cmd/ (e.g. "cryptoutil/").
	CmdDir string `yaml:"cmd_dir"`
	// InitPSID is the PS-ID used for pki-init at suite level (e.g. "sm-kms").
	InitPSID string `yaml:"init_ps_id"`
}

// RegistryProduct is a single product entry in registry.yaml.
type RegistryProduct struct {
	// ID is the canonical product identifier (e.g. "sm").
	ID string `yaml:"id"`
	// DisplayName is the human-readable name (e.g. "Secrets Manager").
	DisplayName string `yaml:"display_name"`
	// InternalAppsDir is the path under internal/apps/ (e.g. "sm/").
	InternalAppsDir string `yaml:"internal_apps_dir"`
	// CmdDir is the sub-path under cmd/ (e.g. "sm/").
	CmdDir string `yaml:"cmd_dir"`
	// InitPSID is the PS-ID used for pki-init at product level (e.g. "sm-kms").
	InitPSID string `yaml:"init_ps_id"`
}

// RegistryProductService is a single product-service entry in registry.yaml.
type RegistryProductService struct {
	// PSID is the canonical PS identifier (e.g. "sm-kms").
	PSID string `yaml:"ps_id"`
	// Product is the product name component (e.g. "sm").
	Product string `yaml:"product"`
	// Service is the service name component (e.g. "kms").
	Service string `yaml:"service"`
	// DisplayName is the human-readable name.
	DisplayName string `yaml:"display_name"`
	// InternalAppsDir is the path under internal/apps/ (e.g. "sm-kms/").
	InternalAppsDir string `yaml:"internal_apps_dir"`
	// MagicFile is the primary magic constants filename under internal/shared/magic/.
	MagicFile string `yaml:"magic_file"`
	// BasePort is the public API port at the SERVICE deployment tier.
	// PRODUCT tier = BasePort + 10000; SUITE tier = BasePort + 20000.
	BasePort int `yaml:"base_port"`
	// PGHostPort is the host-side PostgreSQL port exposed for E2E tests.
	PGHostPort int `yaml:"pg_host_port"`
	// MigrationRangeStart is the inclusive lower bound of this PS-ID's migration versions.
	MigrationRangeStart int `yaml:"migration_range_start"`
	// MigrationRangeEnd is the inclusive upper bound of this PS-ID's migration versions.
	MigrationRangeEnd int `yaml:"migration_range_end"`
	// APIResources lists the canonical OpenAPI path strings for this service's API.
	APIResources []string `yaml:"api_resources"`
	// Entrypoint is the canonical Dockerfile ENTRYPOINT arguments for this PS-ID.
	// Example: ["/app/jose-ja"] or ["/sbin/tini", "--", "/app/cryptoutil", "identity-authz", "start"].
	Entrypoint []string `yaml:"entrypoint"`
	// GoTemplateParams holds Go source-code-specific placeholder values used to compare
	// internal/apps/__PS_ID__/ source files against their canonical templates.
	GoTemplateParams RegistryGoTemplateParams `yaml:"go_template_params"`
}

// RegistryGoTemplateParams holds the Go-specific placeholder values for a PS-ID's
// source template comparison. These map directly to __PLACEHOLDER__ tokens in the
// canonical template files under api/cryptosuite-registry/templates/internal/apps/__PS_ID__/.
type RegistryGoTemplateParams struct {
	// UsagePrefix is the Go identifier prefix for usage variables (e.g. "KMS", "Template").
	UsagePrefix string `yaml:"usage_prefix"`
	// ProductNameConst is the magic constant name for the product name (e.g. "SMProductName").
	ProductNameConst string `yaml:"product_name_const"`
	// ServiceNameConst is the magic constant name for the service name (e.g. "KMSServiceName").
	ServiceNameConst string `yaml:"service_name_const"`
	// ServiceIDConst is the magic constant name for the service ID (e.g. "KMSServiceID").
	ServiceIDConst string `yaml:"service_id_const"`
	// ServicePortConst is the magic constant name for the service port (e.g. "KMSServicePort").
	ServicePortConst string `yaml:"service_port_const"`
	// ServiceDisplayNameConst is the magic constant name for the service's human-readable
	// display name (e.g. "KMSDisplayName"). The constant must exist in the magic package.
	ServiceDisplayNameConst string `yaml:"service_display_name_const"`
}

// PortInfo holds port information derived from the registry for a product-service.
type PortInfo struct {
	// PSID is the canonical PS identifier.
	PSID string
	// BasePort is the public service port at the SERVICE tier.
	BasePort int
	// PGHostPort is the host-side PostgreSQL port.
	PGHostPort int
}

// MigrationRangeInfo holds the assigned migration version range for a product-service.
type MigrationRangeInfo struct {
	// PSID is the canonical PS identifier.
	PSID string
	// Start is the inclusive lower bound.
	Start int
	// End is the inclusive upper bound.
	End int
}

// APIResourceInfo holds the declared OpenAPI path list for a product-service.
type APIResourceInfo struct {
	// PSID is the canonical PS identifier.
	PSID string
	// Resources is the list of canonical path strings (e.g. "/elastickey").
	Resources []string
}
