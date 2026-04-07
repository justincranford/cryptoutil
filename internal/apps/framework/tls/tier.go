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

// suiteID is the canonical suite identifier.
const suiteID = cryptoutilSharedMagic.DefaultOTLPServiceDefault

// productToPSIDs maps product IDs to their constituent PS-IDs.
var productToPSIDs = map[string][]string{
	cryptoutilSharedMagic.SMProductName:       {cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.OTLPServiceSMIM},
	cryptoutilSharedMagic.JoseProductName:     {cryptoutilSharedMagic.OTLPServiceJoseJA},
	cryptoutilSharedMagic.PKIProductName:      {cryptoutilSharedMagic.OTLPServicePKICA},
	cryptoutilSharedMagic.IdentityProductName: {cryptoutilSharedMagic.OTLPServiceIdentityAuthz, cryptoutilSharedMagic.OTLPServiceIdentityIDP, cryptoutilSharedMagic.OTLPServiceIdentityRS, cryptoutilSharedMagic.OTLPServiceIdentityRP, cryptoutilSharedMagic.OTLPServiceIdentitySPA},
	cryptoutilSharedMagic.SkeletonProductName: {cryptoutilSharedMagic.OTLPServiceSkeletonTemplate},
}

// allPSIDs is the ordered list of all 10 PS-IDs in the suite.
var allPSIDs = []string{
	cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.OTLPServiceSMIM, cryptoutilSharedMagic.OTLPServiceJoseJA, cryptoutilSharedMagic.OTLPServicePKICA,
	cryptoutilSharedMagic.OTLPServiceIdentityAuthz, cryptoutilSharedMagic.OTLPServiceIdentityIDP, cryptoutilSharedMagic.OTLPServiceIdentityRS, cryptoutilSharedMagic.OTLPServiceIdentityRP, cryptoutilSharedMagic.OTLPServiceIdentitySPA,
	cryptoutilSharedMagic.OTLPServiceSkeletonTemplate,
}

// psIDSet is a lookup set of valid PS-IDs for quick validation.
var psIDSet = func() map[string]bool {
	m := make(map[string]bool, len(allPSIDs))
	for _, id := range allPSIDs {
		m[id] = true
	}

	return m
}()

// ResolveTier determines the tier type and PS-IDs for a canonical tier ID.
func ResolveTier(tierID string) (TierType, []string, error) {
	return resolveTierInternal(tierID)
}

func resolveTierInternal(tierID string) (TierType, []string, error) {
	if tierID == suiteID {
		return TierSuite, allPSIDs, nil
	}

	if psIDs, ok := productToPSIDs[tierID]; ok {
		return TierProduct, psIDs, nil
	}

	if psIDSet[tierID] {
		return TierService, []string{tierID}, nil
	}

	return 0, nil, fmt.Errorf("unknown tier ID %q: must be %q, a product (sm, jose, pki, identity, skeleton), or a PS-ID", tierID, suiteID)
}

// AppInstances returns the 4 canonical app instance suffixes for a PS-ID.
func AppInstances(psID string) []string {
	return []string{
		psID + "-app-sqlite-1",
		psID + "-app-sqlite-2",
		psID + "-app-postgresql-1",
		psID + "-app-postgresql-2",
	}
}

// PKIDomains returns the 3 PKI domain identifiers for a PS-ID.
// sqlite-1 and sqlite-2 each get their own domain; postgresql-1 and postgresql-2 share one.
func PKIDomains(psID string) []string {
	return []string{
		psID + "-app-sqlite-1",
		psID + "-app-sqlite-2",
		psID + "-app-postgresql-ALL",
	}
}

// ClientRealms returns the 4 client realm names per PKI domain.
func ClientRealms() []string {
	return []string{
		"browser-realm-file",
		"browser-realm-db",
		"service-realm-file",
		"service-realm-db",
	}
}
