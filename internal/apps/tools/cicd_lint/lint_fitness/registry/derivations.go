// Copyright (c) 2025 Justin Cranford

package registry

import (
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// PortTierOffset defines the port offset constants for each deployment tier.
const (
	// PortTierOffsetService is the port offset for the SERVICE deployment tier (base port itself).
	PortTierOffsetService = 0
	// PortTierOffsetProduct is the port offset for the PRODUCT deployment tier.
	PortTierOffsetProduct = 10_000
	// PortTierOffsetSuite is the port offset for the SUITE deployment tier.
	PortTierOffsetSuite = 20_000
)

// ComposeVariantOffset defines the variant-level port offset within a deployment tier.
// Formula: host_port = base_port + tier_offset + variant_offset.
const (
	// ComposeVariantOffsetSQLite1 is the variant offset for the SQLite instance 1 service (+0).
	ComposeVariantOffsetSQLite1 = 0
	// ComposeVariantOffsetSQLite2 is the variant offset for the SQLite instance 2 service (+1).
	ComposeVariantOffsetSQLite2 = 1
	// ComposeVariantOffsetPostgres1 is the variant offset for the PostgreSQL instance 1 service (+2).
	ComposeVariantOffsetPostgres1 = 2
	// ComposeVariantOffsetPostgres2 is the variant offset for the PostgreSQL instance 2 service (+3).
	ComposeVariantOffsetPostgres2 = 3
)

// OTLPServicePrefix is the canonical OTLP service name prefix for all cryptoutil services.
const OTLPServicePrefix = "cryptoutil-"

// ComposeAppSuffix is the separator between PS-ID and variant in compose service names.
const ComposeAppSuffix = "-app-"

// DatabaseSuffix is the suffix appended to a SQL identifier to form the database name.
const DatabaseSuffix = "_database"

// DatabaseUserSuffix is the suffix appended to a SQL identifier to form the database user name.
const DatabaseUserSuffix = "_database_user"

// PostgresServiceSuffix is the suffix appended to a PS-ID to form the PostgreSQL service name in compose.
const PostgresServiceSuffix = "-postgres"

// DBServiceSuffix is the suffix appended to a PS-ID to form the DB compose service name.
const DBServiceSuffix = "-db-postgres-1"

// ComposeVariantSQLite1 is the SQLite instance 1 variant used in compose service names.
const ComposeVariantSQLite1 = "sqlite-1"

// ComposeVariantSQLite2 is the SQLite instance 2 variant used in deployment config overlay names.
const ComposeVariantSQLite2 = "sqlite-2"

// ComposeVariantPostgres1 is the PostgreSQL instance 1 variant used in compose service names.
const ComposeVariantPostgres1 = "postgresql-1"

// ComposeVariantPostgres2 is the PostgreSQL instance 2 variant used in compose service names.
const ComposeVariantPostgres2 = "postgresql-2"

// DeploymentConfigSuffixSQLite1 is the deployment config file suffix for the SQLite instance 1 overlay.
// File: {psid}-app-sqlite-1.yml → OTLP suffix: ComposeVariantSQLite1.
const DeploymentConfigSuffixSQLite1 = "-app-sqlite-1.yml"

// DeploymentConfigSuffixSQLite2 is the deployment config file suffix for the SQLite instance 2 overlay.
// File: {psid}-app-sqlite-2.yml → OTLP suffix: ComposeVariantSQLite2.
const DeploymentConfigSuffixSQLite2 = "-app-sqlite-2.yml"

// DeploymentConfigSuffixPostgresql1 is the deployment config file suffix for the PostgreSQL instance 1 overlay.
// File: {psid}-app-postgresql-1.yml → OTLP suffix: ComposeVariantPostgres1.
const DeploymentConfigSuffixPostgresql1 = "-app-postgresql-1.yml"

// DeploymentConfigSuffixPostgresql2 is the deployment config file suffix for the PostgreSQL instance 2 overlay.
// File: {psid}-app-postgresql-2.yml → OTLP suffix: ComposeVariantPostgres2.
const DeploymentConfigSuffixPostgresql2 = "-app-postgresql-2.yml"

// -----------------------------------------------------------------------
// Port derivation functions (Task 3.1)
// -----------------------------------------------------------------------

// PublicPort returns the public API port for the given PS-ID at the SERVICE tier.
// Returns 0 if the PS-ID is not found in the registry.
func PublicPort(psID string) int {
	for _, ps := range allRegistryFile.ProductServices {
		if ps.PSID == psID {
			return ps.BasePort + PortTierOffsetService
		}
	}

	return 0
}

// AdminPort returns the standard admin/health port for the given PS-ID.
// This is always the same for all services (standard admin port).
func AdminPort(psID string) int {
	_ = psID // admin port is universal, PS-ID ignored

	return int(cryptoutilSharedMagic.DefaultPrivatePortCryptoutil)
}

// PostgresPort returns the host-side PostgreSQL port for the given PS-ID (used in E2E tests).
// Returns 0 if the PS-ID is not found in the registry.
func PostgresPort(psID string) int {
	for _, ps := range allRegistryFile.ProductServices {
		if ps.PSID == psID {
			return ps.PGHostPort
		}
	}

	return 0
}

// ProductPublicPort returns the public API port for the given PS-ID at the PRODUCT deployment tier.
// Returns 0 if the PS-ID is not found in the registry.
func ProductPublicPort(psID string) int {
	for _, ps := range allRegistryFile.ProductServices {
		if ps.PSID == psID {
			return ps.BasePort + PortTierOffsetProduct
		}
	}

	return 0
}

// SuitePublicPort returns the public API port for the given PS-ID at the SUITE deployment tier.
// Returns 0 if the PS-ID is not found in the registry.
func SuitePublicPort(psID string) int {
	for _, ps := range allRegistryFile.ProductServices {
		if ps.PSID == psID {
			return ps.BasePort + PortTierOffsetSuite
		}
	}

	return 0
}

// -----------------------------------------------------------------------
// SQL identifier derivation functions (Task 3.2)
// -----------------------------------------------------------------------

// PSIDToSQLID converts a PS-ID (kebab-case) to a SQL identifier (snake_case).
// Example: "jose-ja" → "jose_ja".
func PSIDToSQLID(psID string) string {
	return strings.ReplaceAll(psID, "-", "_")
}

// DatabaseName returns the PostgreSQL database name for the given PS-ID.
// Example: "jose-ja" → "jose_ja_database".
func DatabaseName(psID string) string {
	return PSIDToSQLID(psID) + DatabaseSuffix
}

// DatabaseUser returns the PostgreSQL database user for the given PS-ID.
// Example: "jose-ja" → "jose_ja_database_user".
func DatabaseUser(psID string) string {
	return PSIDToSQLID(psID) + DatabaseUserSuffix
}

// PostgresServiceName returns the compose PostgreSQL service/container name for the given PS-ID.
// Example: "jose-ja" → "jose-ja-postgres".
func PostgresServiceName(psID string) string {
	return psID + PostgresServiceSuffix
}

// DBServiceName returns the compose database service name for the given PS-ID.
// Example: "jose-ja" → "jose-ja-db-postgres-1".
func DBServiceName(psID string) string {
	return psID + DBServiceSuffix
}

// -----------------------------------------------------------------------
// Service name derivation functions (Task 3.3)
// -----------------------------------------------------------------------

// OTLPServiceName returns the canonical OTLP service name for the given PS-ID.
// Example: "sm-kms" → "cryptoutil-sm-kms".
func OTLPServiceName(psID string) string {
	return OTLPServicePrefix + psID
}

// ComposeServiceName returns the canonical compose service name for the given PS-ID and variant.
// Example: "sm-kms", "sqlite" → "sm-kms-app-sqlite".
func ComposeServiceName(psID, variant string) string {
	return psID + ComposeAppSuffix + variant
}

// ValidOTLPServiceNames returns the computed OTLP service names for all product-services.
// Each name has the form "cryptoutil-{PS-ID}".
func ValidOTLPServiceNames() []string {
	pss := AllProductServices()
	names := make([]string, len(pss))

	for i, ps := range pss {
		names[i] = OTLPServiceName(ps.PSID)
	}

	return names
}

// ValidComposeServiceNames returns the computed compose app service names for all product-services.
// Each name has the form "{PS-ID}-app-{variant}" for the variants: sqlite-1, sqlite-2, postgresql-1, postgresql-2.
func ValidComposeServiceNames() []string {
	variants := []string{ComposeVariantSQLite1, ComposeVariantSQLite2, ComposeVariantPostgres1, ComposeVariantPostgres2}
	pss := AllProductServices()
	names := make([]string, 0, len(pss)*len(variants))

	for _, ps := range pss {
		for _, v := range variants {
			names = append(names, ComposeServiceName(ps.PSID, v))
		}
	}

	return names
}

// -----------------------------------------------------------------------
// Dockerfile derivation functions (Task 5.2)
// -----------------------------------------------------------------------

// DockerfileEntrypoint returns the canonical ENTRYPOINT arguments for the given PS-ID.
// The entrypoint is defined in registry.yaml per PS-ID and reflects the actual
// binary used to run the service (own PS-ID binary or suite binary with subcommand).
// Returns nil if the PS-ID is not found in the registry.
func DockerfileEntrypoint(psID string) []string {
	for _, ps := range allRegistryFile.ProductServices {
		if ps.PSID == psID {
			return ps.Entrypoint
		}
	}

	return nil
}
