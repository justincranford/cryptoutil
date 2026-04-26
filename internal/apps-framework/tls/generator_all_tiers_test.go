// Copyright (c) 2025 Justin Cranford
//

package tls_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestGenerate_AllSixteenTiers verifies that Generate succeeds for all 16 valid tier IDs:
// 1 suite, 5 products, and 10 PS-IDs. Uses stub crypto (no real key generation) so the
// test runs fast enough for t.Parallel() on all subtests.
func TestGenerate_AllSixteenTiers(t *testing.T) {
	t.Parallel()

	// All 16 valid tier IDs: 1 suite + 5 products + 10 PS-IDs.
	tiers := []string{
		// Suite (1)
		cryptoutilSharedMagic.DefaultOTLPServiceDefault,

		// Products (5)
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SkeletonProductName,

		// PS-IDs (10)
		cryptoutilSharedMagic.OTLPServiceSMKMS,
		cryptoutilSharedMagic.OTLPServiceSMIM,
		cryptoutilSharedMagic.OTLPServiceJoseJA,
		cryptoutilSharedMagic.OTLPServicePKICA,
		cryptoutilSharedMagic.OTLPServiceIdentityAuthz,
		cryptoutilSharedMagic.OTLPServiceIdentityIDP,
		cryptoutilSharedMagic.OTLPServiceIdentityRS,
		cryptoutilSharedMagic.OTLPServiceIdentityRP,
		cryptoutilSharedMagic.OTLPServiceIdentitySPA,
		cryptoutilSharedMagic.OTLPServiceSkeletonTemplate,
	}

	for _, tierID := range tiers {
		t.Run(tierID, func(t *testing.T) {
			t.Parallel()

			gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
				os.MkdirAll,
				stubWriteFile,
				stubCreateCA,
				stubCreateLeaf,
				stubGetKeyPair,
				stubEncodePKCS12,
				stubEncodeTrustPKCS12,
				stubGetRealmsForPSID,
			)

			err := gen.Generate(tierID, t.TempDir())
			require.NoError(t, err, "Generate(%q) should succeed", tierID)
		})
	}
}

// TestGenerate_InvalidTierID verifies that Generate rejects unknown tier IDs with a descriptive error.
func TestGenerate_InvalidTierID(t *testing.T) {
	t.Parallel()

	gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
		stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair,
		stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
	)

	err := gen.Generate("not-a-valid-tier", t.TempDir())
	require.ErrorContains(t, err, "unknown tier ID")
}
