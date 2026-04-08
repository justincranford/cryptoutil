// Copyright (c) 2025 Justin Cranford
//
//

package magic

// ProductToPSIDs maps product IDs to their constituent PS-IDs.
var ProductToPSIDs = map[string][]string{
	SMProductName:       {OTLPServiceSMKMS, OTLPServiceSMIM},
	JoseProductName:     {OTLPServiceJoseJA},
	PKIProductName:      {OTLPServicePKICA},
	IdentityProductName: {OTLPServiceIdentityAuthz, OTLPServiceIdentityIDP, OTLPServiceIdentityRS, OTLPServiceIdentityRP, OTLPServiceIdentitySPA},
	SkeletonProductName: {OTLPServiceSkeletonTemplate},
}

// AllPSIDs is the ordered list of all 10 PS-IDs in the suite.
var AllPSIDs = []string{
	OTLPServiceSMKMS, OTLPServiceSMIM, OTLPServiceJoseJA, OTLPServicePKICA,
	OTLPServiceIdentityAuthz, OTLPServiceIdentityIDP, OTLPServiceIdentityRS, OTLPServiceIdentityRP, OTLPServiceIdentitySPA,
	OTLPServiceSkeletonTemplate,
}

// PSIDSet is a lookup set of valid PS-IDs for quick validation.
var PSIDSet = func() map[string]bool {
	m := make(map[string]bool, len(AllPSIDs))
	for _, id := range AllPSIDs {
		m[id] = true
	}

	return m
}()
