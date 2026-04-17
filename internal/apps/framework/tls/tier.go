// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"fmt"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TierType represents the level of deployment: suite, product, or service.
type TierType int

const (
	// TierSuite generates certs for all 10 PS-IDs.
	TierSuite TierType = iota
	// TierProduct generates certs for all PS-IDs within a single product.
	TierProduct
	// TierService generates certs for a single PS-ID.
	TierService
)

// ResolveTier determines the tier type and PS-IDs for a canonical tier ID.
func ResolveTier(tierID string) (TierType, []string, error) {
	return resolveTierInternal(tierID)
}

func resolveTierInternal(tierID string) (TierType, []string, error) {
	switch {
	case tierID == cryptoutilSharedMagic.DefaultOTLPServiceDefault:
		return TierSuite, cryptoutilSharedMagic.AllPSIDs, nil
	case cryptoutilSharedMagic.ProductToPSIDs[tierID] != nil:
		return TierProduct, cryptoutilSharedMagic.ProductToPSIDs[tierID], nil
	case cryptoutilSharedMagic.PSIDSet[tierID]:
		return TierService, []string{tierID}, nil
	default:
		return 0, nil, fmt.Errorf("unknown tier ID %q: must be %q, a product (sm, jose, pki, identity, skeleton), or a PS-ID", tierID, cryptoutilSharedMagic.DefaultOTLPServiceDefault)
	}
}

// PKIInitAppInstanceSuffixes returns the 4 canonical app instance suffixes.
// These are used to form directory names and Docker service DNS names.
func PKIInitAppInstanceSuffixes() []string {
	return []string{cryptoutilSharedMagic.CICDTemplateVariantSQLite1, cryptoutilSharedMagic.CICDTemplateVariantSQLite2, cryptoutilSharedMagic.PKIInitInstanceSuffixPostgres1, cryptoutilSharedMagic.PKIInitInstanceSuffixPostgres2}
}

// PKIInitClientPKIDomains returns the 3 PKI domain identifiers used for client cert generation.
// sqlite-1 and sqlite-2 each get their own domain; postgres-1 and postgres-2 share one.
func PKIInitClientPKIDomains() []string {
	return []string{cryptoutilSharedMagic.CICDTemplateVariantSQLite1, cryptoutilSharedMagic.CICDTemplateVariantSQLite2, cryptoutilSharedMagic.DockerServicePostgres}
}

// PKIInitAdminInstanceSuffixes returns the 4 instance suffixes for the private admin (mTLS) channel.
func PKIInitAdminInstanceSuffixes() []string {
	return []string{cryptoutilSharedMagic.CICDTemplateVariantSQLite1, cryptoutilSharedMagic.CICDTemplateVariantSQLite2, cryptoutilSharedMagic.PKIInitInstanceSuffixPostgres1, cryptoutilSharedMagic.PKIInitInstanceSuffixPostgres2}
}

// PKIInitPostgresAppInstanceSuffixes returns the 2 PostgreSQL app instance suffixes.
// Only postgres-1 and postgres-2 connect to PostgreSQL; sqlite-1 and sqlite-2 use
// in-memory SQLite and never connect to PostgreSQL.
func PKIInitPostgresAppInstanceSuffixes() []string {
	return []string{cryptoutilSharedMagic.PKIInitInstanceSuffixPostgres1, cryptoutilSharedMagic.PKIInitInstanceSuffixPostgres2}
}

// PKIInitUserTypes returns the 2 API path user type names for client cert generation.
func PKIInitUserTypes() []string {
	return []string{"browseruser", "serviceuser"}
}
