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

func TestAllProductServices_InternalAppsDirMatchesProductAndService(t *testing.T) {
	t.Parallel()

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		t.Run(ps.PSID, func(t *testing.T) {
			t.Parallel()

			expected := ps.Product + "/" + ps.Service + "/"
			assert.Equal(t, expected, ps.InternalAppsDir, "InternalAppsDir must match product/service/")
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

	// Mutate first copy and verify second is unaffected.
	ps1[0].PSID = "mutated"
	require.NotEqual(t, ps1[0].PSID, ps2[0].PSID, "AllProductServices must return independent copies")
}
