// Copyright (c) 2025-2026 Justin Cranford.
//

package tls_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestGenerate_AllThirteenTiers verifies that Generate succeeds for all 13 valid tier IDs:
// 1 suite, 4 products, and 8 PS-IDs. Uses stub crypto (no real key generation) so the
// test runs fast enough for t.Parallel() on all subtests.
func TestGenerate_AllThirteenTiers(t *testing.T) {
	t.Parallel()

	// All 13 valid tier IDs: 1 suite + 4 products + 8 PS-IDs.
	tiers := []string{
		// Suite (1)
		cryptoutilSharedMagic.DefaultOTLPServiceDefault,

		// Products (4)
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SkeletonProductName,

		// PS-IDs (8)
		cryptoutilSharedMagic.OTLPServiceSMKMS,
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
