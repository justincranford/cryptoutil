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

// PKIInitAppInstances returns the 4 canonical app instance suffixes for a PS-ID.
func PKIInitAppInstances(psID string) []string {
	return []string{
		psID + "-app-sqlite-1",
		psID + "-app-sqlite-2",
		psID + "-app-postgresql-1",
		psID + "-app-postgresql-2",
	}
}

// PKIInitDomains returns the 3 PKI domain identifiers for a PS-ID.
// sqlite-1 and sqlite-2 each get their own domain; postgresql-1 and postgresql-2 share one.
func PKIInitDomains(psID string) []string {
	return []string{
		psID + "-app-sqlite-1",
		psID + "-app-sqlite-2",
		psID + "-app-postgresql-ALL",
	}
}

// PKIInitClientRealms returns the 4 client realm names per PKI domain.
func PKIInitClientRealms() []string {
	return []string{
		"browser-realm-file",
		"browser-realm-db",
		"service-realm-file",
		"service-realm-db",
	}
}
