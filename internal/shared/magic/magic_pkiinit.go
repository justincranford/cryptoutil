// Copyright (c) 2025 Justin Cranford
//
//

package magic

import (
	"os"
	"time"
)

// PKI Init constants for the pki-init Docker Compose init job.
const (
	// PSIDPKIInit is the product-service ID for the pki-init job ("pki-init").
	// Used as the pki-init subcommand name in suite and product CLI routers.
	PSIDPKIInit = "pki-init"

	// PKIInitCertValidityDays is the validity period for PKI init certificates in days.
	// Deprecated: use PKIInitValidityLeaf, PKIInitValidityIssuingCA, or PKIInitValidityRootCA.
	PKIInitCertValidityDays = 365

	// PKIInitCertFileMode is the file permission mode for certificate files.
	// Deprecated: use PKIInitPublicCertFileMode or PKIInitPrivateKeyFileMode.
	PKIInitCertFileMode = os.FileMode(0o644)

	// PKIInitCertsDirMode is the directory permission mode for the certs directory.
	PKIInitCertsDirMode = os.FileMode(0o755)

	// PKIInitPublicCertFileMode is the file permission mode for public cert files (.crt,
	// truststore .p12).
	PKIInitPublicCertFileMode = os.FileMode(0o444)

	// PKIInitPrivateKeyFileMode is the file permission mode for private key files (.key,
	// keystore .p12 with private key).
	PKIInitPrivateKeyFileMode = os.FileMode(0o440)

	// PKIInitValidityRootCA is the validity period for root CA certificates (20 years per
	// CA/Browser Forum Baseline Requirements).
	PKIInitValidityRootCA = 20 * 365 * 24 * time.Hour

	// PKIInitValidityIssuingCA is the validity period for issuing CA certificates (5 years
	// per CA/Browser Forum Baseline Requirements).
	PKIInitValidityIssuingCA = 5 * 365 * 24 * time.Hour

	// PKIInitValidityLeaf is the validity period for end-entity (leaf) certificates (397
	// days — one day below the CA/Browser Forum 398-day hard limit).
	PKIInitValidityLeaf = 397 * 24 * time.Hour

	// PKIInitOtelCollectorContrib is the directory name component used for the OTel
	// Collector Contrib service in pki-init cert directory names.
	PKIInitOtelCollectorContrib = "otel-collector-contrib"

	// PKIInitPostgresLeader is the role name for the primary (leader) PostgreSQL instance,
	// used in pki-init cert directory names and DNS SANs.
	PKIInitPostgresLeader = "leader"

	// PKIInitPostgresFollower is the role name for the replica (follower) PostgreSQL
	// instance, used in pki-init cert directory names and DNS SANs.
	PKIInitPostgresFollower = "follower"

	// PKIInitPostgresLeaderService is the Docker service DNS name for the leader PostgreSQL
	// instance. Used as a DNS SAN in server leaf certificates.
	PKIInitPostgresLeaderService = "postgres-leader"

	// PKIInitPostgresFollowerService is the Docker service DNS name for the follower
	// PostgreSQL instance. Used as a DNS SAN in server leaf certificates.
	PKIInitPostgresFollowerService = "postgres-follower"

	// PKIInitInstanceSuffixPostgres1 is the suffix for the first PostgreSQL app instance.
	// Used to form directory names and Docker service DNS names in pki-init cert generation.
	PKIInitInstanceSuffixPostgres1 = "postgres-1"

	// PKIInitInstanceSuffixPostgres2 is the suffix for the second PostgreSQL app instance.
	// Used to form directory names and Docker service DNS names in pki-init cert generation.
	PKIInitInstanceSuffixPostgres2 = "postgres-2"

	// PKIInitEntityInfra is the entity type name for service-to-service infrastructure
	// connections in pki-init cert directory names (e.g., OTel Collector -> Grafana LGTM
	// OTLP forwarding). Distinguished from PKIInitEntityAdmin (human operator access).
	PKIInitEntityInfra = "infra"
)
