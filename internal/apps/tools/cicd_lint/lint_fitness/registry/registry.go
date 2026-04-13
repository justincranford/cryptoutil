// Copyright (c) 2025 Justin Cranford

// Package registry defines the canonical entity registry for all cryptoutil products,
// product-services, and suites. This is the single source of truth for structural
// conventions. Fitness linters use this registry to detect drift.
//
// Data is loaded from api/cryptosuite-registry/registry.yaml at package initialization.
// The YAML file is the machine-readable SSOT; this package exposes Go-typed accessors.
package registry

// Product represents a cryptoutil product (e.g. "sm", "jose").
type Product struct {
	// ID is the canonical product identifier (e.g. "sm").
	ID string
	// DisplayName is the human-readable name (e.g. "Secrets Manager").
	DisplayName string
	// InternalAppsDir is the path under internal/apps/ (e.g. "sm/").
	InternalAppsDir string
	// CmdDir is the sub-path under cmd/ (e.g. "sm/").
	CmdDir string
	// InitPSID is the PS-ID used for pki-init at product level (e.g. "sm-kms").
	InitPSID string
}

// ProductService represents a product-service pair (e.g. "sm-kms").
type ProductService struct {
	// PSID is the canonical PS identifier (e.g. "sm-kms").
	PSID string
	// Product is the product name component (e.g. "sm").
	Product string
	// Service is the service name component (e.g. "kms").
	Service string
	// DisplayName is the human-readable name (e.g. "Secrets Manager Key Management").
	DisplayName string
	// InternalAppsDir is the path under internal/apps/ (e.g. "sm-kms/").
	InternalAppsDir string
	// MagicFile is the filename of the primary magic constants file
	// under internal/shared/magic/ (e.g. "magic_sm.go").
	MagicFile string
}

// Suite represents the cryptoutil top-level suite deployment.
type Suite struct {
	// ID is the canonical suite identifier (e.g. "cryptoutil").
	ID string
	// DisplayName is the human-readable name.
	DisplayName string
	// CmdDir is the sub-path under cmd/ (e.g. "cryptoutil/").
	CmdDir string
	// InitPSID is the PS-ID used for pki-init at suite level (e.g. "sm-kms").
	InitPSID string
}

// allRegistryFile holds the parsed registry YAML loaded once at init time.
var allRegistryFile *RegistryFile

// allProducts is the canonical registry of all 5 cryptoutil products.
// Canonical order: sm, jose, pki, identity, skeleton.
// Populated by init() from api/cryptosuite-registry/registry.yaml.
var allProducts []Product

// allProductServices is the canonical registry of all 10 cryptoutil product-services.
// Canonical order: sm-kms, sm-im, jose-ja, pki-ca, identity-authz, identity-idp,
// identity-rs, identity-rp, identity-spa, skeleton-template.
// Populated by init() from api/cryptosuite-registry/registry.yaml.
var allProductServices []ProductService

// allSuites is the canonical registry of the cryptoutil suite.
// Populated by init() from api/cryptosuite-registry/registry.yaml.
var allSuites []Suite

func init() {
	path, err := findRegistryYAMLPath()
	if err != nil {
		panic("registry: cannot locate registry YAML: " + err.Error())
	}

	r, err := LoadRegistry(path)
	if err != nil {
		panic("registry: failed to load " + path + ": " + err.Error())
	}

	allRegistryFile = r
	allProducts = r.ToProducts()
	allProductServices = r.ToProductServices()
	allSuites = r.ToSuites()
}

// AllProducts returns the canonical list of all 5 products.
func AllProducts() []Product {
	result := make([]Product, len(allProducts))
	copy(result, allProducts)

	return result
}

// AllProductServices returns the canonical list of all 10 product-services.
func AllProductServices() []ProductService {
	result := make([]ProductService, len(allProductServices))
	copy(result, allProductServices)

	return result
}

// AllSuites returns the canonical list of all suites (currently 1).
func AllSuites() []Suite {
	result := make([]Suite, len(allSuites))
	copy(result, allSuites)

	return result
}

// AllPorts returns port information for every product-service.
// Each entry provides the base public port and the host-side PostgreSQL port.
func AllPorts() []PortInfo {
	result := make([]PortInfo, len(allRegistryFile.ProductServices))

	for i, ps := range allRegistryFile.ProductServices {
		result[i] = PortInfo{
			PSID:       ps.PSID,
			BasePort:   ps.BasePort,
			PGHostPort: ps.PGHostPort,
		}
	}

	return result
}

// AllMigrationRanges returns the assigned migration version range per product-service.
func AllMigrationRanges() []MigrationRangeInfo {
	result := make([]MigrationRangeInfo, len(allRegistryFile.ProductServices))

	for i, ps := range allRegistryFile.ProductServices {
		result[i] = MigrationRangeInfo{
			PSID:  ps.PSID,
			Start: ps.MigrationRangeStart,
			End:   ps.MigrationRangeEnd,
		}
	}

	return result
}

// AllAPIResources returns the declared OpenAPI path list per product-service.
func AllAPIResources() []APIResourceInfo {
	result := make([]APIResourceInfo, len(allRegistryFile.ProductServices))

	for i, ps := range allRegistryFile.ProductServices {
		resources := make([]string, len(ps.APIResources))
		copy(resources, ps.APIResources)
		result[i] = APIResourceInfo{
			PSID:      ps.PSID,
			Resources: resources,
		}
	}

	return result
}
