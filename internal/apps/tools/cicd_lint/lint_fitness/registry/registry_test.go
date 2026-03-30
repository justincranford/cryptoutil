// Copyright (c) 2025 Justin Cranford

package registry_test

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

func TestAllProducts_Count(t *testing.T) {
	t.Parallel()

	products := lintFitnessRegistry.AllProducts()
	require.Len(t, products, cryptoutilSharedMagic.SuiteProductCount, "expected exactly 5 products in registry")
}

func TestAllProductServices_Count(t *testing.T) {
	t.Parallel()

	services := lintFitnessRegistry.AllProductServices()
	require.Len(t, services, cryptoutilSharedMagic.SuiteServiceCount, "expected exactly 10 product-services in registry")
}

func TestAllSuites_Count(t *testing.T) {
	t.Parallel()

	suites := lintFitnessRegistry.AllSuites()
	require.Len(t, suites, 1, "expected exactly 1 suite in registry")
}

func TestAllProducts_FieldsNonEmpty(t *testing.T) {
	t.Parallel()

	for _, p := range lintFitnessRegistry.AllProducts() {
		t.Run(p.ID, func(t *testing.T) {
			t.Parallel()

			assert.NotEmpty(t, p.ID, "product ID must not be empty")
			assert.NotEmpty(t, p.DisplayName, "product DisplayName must not be empty")
			assert.NotEmpty(t, p.InternalAppsDir, "product InternalAppsDir must not be empty")
			assert.NotEmpty(t, p.CmdDir, "product CmdDir must not be empty")
		})
	}
}

func TestAllProductServices_FieldsNonEmpty(t *testing.T) {
	t.Parallel()

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		t.Run(ps.PSID, func(t *testing.T) {
			t.Parallel()

			assert.NotEmpty(t, ps.PSID, "PSID must not be empty")
			assert.NotEmpty(t, ps.Product, "Product must not be empty")
			assert.NotEmpty(t, ps.Service, "Service must not be empty")
			assert.NotEmpty(t, ps.DisplayName, "DisplayName must not be empty")
			assert.NotEmpty(t, ps.InternalAppsDir, "InternalAppsDir must not be empty")
			assert.NotEmpty(t, ps.MagicFile, "MagicFile must not be empty")
		})
	}
}

func TestAllSuites_FieldsNonEmpty(t *testing.T) {
	t.Parallel()

	for _, s := range lintFitnessRegistry.AllSuites() {
		t.Run(s.ID, func(t *testing.T) {
			t.Parallel()

			assert.NotEmpty(t, s.ID, "suite ID must not be empty")
			assert.NotEmpty(t, s.DisplayName, "suite DisplayName must not be empty")
			assert.NotEmpty(t, s.CmdDir, "suite CmdDir must not be empty")
		})
	}
}

func TestAllProductServices_PSIDEqualsProductDashService(t *testing.T) {
	t.Parallel()

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		t.Run(ps.PSID, func(t *testing.T) {
			t.Parallel()

			expected := ps.Product + "-" + ps.Service
			assert.Equal(t, expected, ps.PSID, "PSID must equal product-service")
		})
	}
}

func TestAllProductServices_InternalAppsDirMatchesPSID(t *testing.T) {
	t.Parallel()

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		t.Run(ps.PSID, func(t *testing.T) {
			t.Parallel()

			expected := ps.PSID + "/"
			assert.Equal(t, expected, ps.InternalAppsDir, "InternalAppsDir must match PSID/")
		})
	}
}

func TestAllProductServices_ContainsExpectedPSIDs(t *testing.T) {
	t.Parallel()

	services := lintFitnessRegistry.AllProductServices()
	psIDs := make(map[string]bool, len(services))

	for _, ps := range services {
		psIDs[ps.PSID] = true
	}

	expectedPSIDs := []string{
		cryptoutilSharedMagic.OTLPServiceIdentityAuthz,
		cryptoutilSharedMagic.OTLPServiceIdentityIDP,
		cryptoutilSharedMagic.OTLPServiceIdentityRP,
		cryptoutilSharedMagic.OTLPServiceIdentityRS,
		cryptoutilSharedMagic.OTLPServiceIdentitySPA,
		cryptoutilSharedMagic.OTLPServiceJoseJA,
		cryptoutilSharedMagic.OTLPServicePKICA,
		cryptoutilSharedMagic.OTLPServiceSkeletonTemplate,
		cryptoutilSharedMagic.OTLPServiceSMIM,
		cryptoutilSharedMagic.OTLPServiceSMKMS,
	}

	for _, id := range expectedPSIDs {
		assert.True(t, psIDs[id], "registry must contain PS-ID %q", id)
	}
}

func TestAllProducts_ContainsExpectedProductIDs(t *testing.T) {
	t.Parallel()

	products := lintFitnessRegistry.AllProducts()
	productIDs := make(map[string]bool, len(products))

	for _, p := range products {
		productIDs[p.ID] = true
	}

	expectedProductIDs := []string{
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SMProductName,
	}

	for _, id := range expectedProductIDs {
		assert.True(t, productIDs[id], "registry must contain product ID %q", id)
	}
}

func TestAllReturnsIndependentCopies(t *testing.T) {
	t.Parallel()

	ps1 := lintFitnessRegistry.AllProductServices()
	ps2 := lintFitnessRegistry.AllProductServices()

	require.Equal(t, ps1, ps2, "two calls must return equal slices")

	const mutatedPSID = "mutated"

	// Mutate first copy and verify second is unaffected.
	ps1[0].PSID = mutatedPSID
	require.NotEqual(t, ps1[0].PSID, ps2[0].PSID, "AllProductServices must return independent copies")
}

func TestAllPorts_Count(t *testing.T) {
	t.Parallel()

	ports := lintFitnessRegistry.AllPorts()
	require.Len(t, ports, cryptoutilSharedMagic.SuiteServiceCount, "expected exactly 10 port entries")
}

func TestAllPorts_FieldsNonZero(t *testing.T) {
	t.Parallel()

	for _, p := range lintFitnessRegistry.AllPorts() {
		t.Run(p.PSID, func(t *testing.T) {
			t.Parallel()

			assert.NotEmpty(t, p.PSID, "port PSID must not be empty")
			assert.Greater(t, p.BasePort, 0, "base_port must be positive")
			assert.Greater(t, p.PGHostPort, 0, "pg_host_port must be positive")
		})
	}
}

func TestAllPorts_ContainsExpectedPSIDs(t *testing.T) {
	t.Parallel()

	ports := lintFitnessRegistry.AllPorts()
	psIDs := make(map[string]bool, len(ports))

	for _, p := range ports {
		psIDs[p.PSID] = true
	}

	assert.True(t, psIDs[cryptoutilSharedMagic.OTLPServiceSMKMS], "AllPorts must contain sm-kms")
	assert.True(t, psIDs[cryptoutilSharedMagic.OTLPServiceSMIM], "AllPorts must contain sm-im")
}

func TestAllPorts_NoDuplicatePSIDsOrPorts(t *testing.T) {
	t.Parallel()

	ports := lintFitnessRegistry.AllPorts()
	seenPSID := make(map[string]bool, len(ports))
	seenBase := make(map[int]bool, len(ports))
	seenPG := make(map[int]bool, len(ports))

	for _, p := range ports {
		assert.False(t, seenPSID[p.PSID], "duplicate PSID in AllPorts: %s", p.PSID)
		assert.False(t, seenBase[p.BasePort], "duplicate base_port in AllPorts: %d", p.BasePort)
		assert.False(t, seenPG[p.PGHostPort], "duplicate pg_host_port in AllPorts: %d", p.PGHostPort)
		seenPSID[p.PSID] = true
		seenBase[p.BasePort] = true
		seenPG[p.PGHostPort] = true
	}
}

func TestAllMigrationRanges_Count(t *testing.T) {
	t.Parallel()

	ranges := lintFitnessRegistry.AllMigrationRanges()
	require.Len(t, ranges, cryptoutilSharedMagic.SuiteServiceCount, "expected exactly 10 migration range entries")
}

func TestAllMigrationRanges_FieldsValid(t *testing.T) {
	t.Parallel()

	for _, mr := range lintFitnessRegistry.AllMigrationRanges() {
		t.Run(mr.PSID, func(t *testing.T) {
			t.Parallel()

			assert.NotEmpty(t, mr.PSID, "migration range PSID must not be empty")
			assert.GreaterOrEqual(t, mr.Start, 2001, "migration_range_start must be >= 2001")
			assert.Greater(t, mr.End, mr.Start, "migration_range_end must be > start")
		})
	}
}

func TestAllMigrationRanges_NoOverlap(t *testing.T) {
	t.Parallel()

	ranges := lintFitnessRegistry.AllMigrationRanges()

	for i := range ranges {
		for j := i + 1; j < len(ranges); j++ {
			a := ranges[i]
			b := ranges[j]
			overlaps := a.Start <= b.End && b.Start <= a.End
			assert.False(t, overlaps,
				"migration ranges for %s [%d,%d] and %s [%d,%d] must not overlap",
				a.PSID, a.Start, a.End, b.PSID, b.Start, b.End)
		}
	}
}

func TestAllAPIResources_Count(t *testing.T) {
	t.Parallel()

	resources := lintFitnessRegistry.AllAPIResources()
	require.Len(t, resources, cryptoutilSharedMagic.SuiteServiceCount, "expected exactly 10 API resource entries")
}

func TestAllAPIResources_FieldsNonEmpty(t *testing.T) {
	t.Parallel()

	for _, ar := range lintFitnessRegistry.AllAPIResources() {
		t.Run(ar.PSID, func(t *testing.T) {
			t.Parallel()

			assert.NotEmpty(t, ar.PSID, "API resource PSID must not be empty")
			// Resources may be empty for services with no OpenAPI spec (e.g. identity-spa).
			// We only assert the PSID is set.
		})
	}
}

func TestAllAPIResources_ResourcesAreIndependentCopies(t *testing.T) {
	t.Parallel()

	r1 := lintFitnessRegistry.AllAPIResources()
	r2 := lintFitnessRegistry.AllAPIResources()

	require.Equal(t, r1, r2, "two AllAPIResources calls must return equal results")

	// Mutate first slice and verify second is unaffected.
	if len(r1) > 0 && len(r1[0].Resources) > 0 {
		original := r1[0].Resources[0]
		r1[0].Resources[0] = "mutated"
		require.Equal(t, original, r2[0].Resources[0], "AllAPIResources must return independent copies")
	}
}
