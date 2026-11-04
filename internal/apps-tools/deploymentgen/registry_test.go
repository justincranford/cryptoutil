// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"slices"
	"testing"

	entityregistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"

	"github.com/stretchr/testify/require"
)

func TestRegisteredScopesDeriveCanonicalRegistry(t *testing.T) {
	t.Parallel()

	scopes := RegisteredScopes()
	services := entityregistry.AllProductServices()
	products := entityregistry.AllProducts()
	suites := entityregistry.AllSuites()

	require.Len(t, services, 7)
	require.Len(t, products, 3)
	require.Len(t, suites, 1)
	require.Len(t, scopes, 11)

	tests := []struct {
		name        string
		tier        Tier
		expectedIDs []string
	}{
		{name: "product services", tier: TierService, expectedIDs: serviceIDs(services)},
		{name: "products", tier: TierProduct, expectedIDs: productIDs(products)},
		{name: "suites", tier: TierSuite, expectedIDs: suiteIDs(suites)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var gotIDs []string

			for _, scope := range scopes {
				if scope.Tier == test.tier {
					gotIDs = append(gotIDs, scope.ID)
				}
			}

			require.Equal(t, test.expectedIDs, gotIDs)
		})
	}

	for _, product := range products {
		t.Run("members/"+product.ID, func(t *testing.T) {
			t.Parallel()

			scope, found := findScope(scopes, TierProduct, product.ID)
			require.True(t, found)

			var expected []string

			for _, service := range services {
				if service.Product == product.ID {
					expected = append(expected, service.PSID)
				}
			}

			slices.Sort(expected)

			require.NotEmpty(t, expected)
			require.Equal(t, expected, scope.Members)
		})
	}

	for _, service := range services {
		t.Run("service/"+service.PSID, func(t *testing.T) {
			t.Parallel()

			scope, found := findScope(scopes, TierService, service.PSID)
			require.True(t, found)
			require.Equal(t, []string{service.PSID}, scope.Members)
		})
	}

	for _, suite := range suites {
		t.Run("suite/"+suite.ID, func(t *testing.T) {
			t.Parallel()

			scope, found := findScope(scopes, TierSuite, suite.ID)
			require.True(t, found)

			expected := serviceIDs(services)
			require.Equal(t, expected, scope.Members)
		})
	}
}

func findScope(scopes []Scope, tier Tier, id string) (Scope, bool) {
	for _, scope := range scopes {
		if scope.Tier == tier && scope.ID == id {
			return scope, true
		}
	}

	return Scope{}, false
}

func serviceIDs(services []entityregistry.ProductService) []string {
	ids := make([]string, 0, len(services))
	for _, service := range services {
		ids = append(ids, service.PSID)
	}

	slices.Sort(ids)

	return ids
}

func productIDs(products []entityregistry.Product) []string {
	ids := make([]string, 0, len(products))
	for _, product := range products {
		ids = append(ids, product.ID)
	}

	slices.Sort(ids)

	return ids
}

func suiteIDs(suites []entityregistry.Suite) []string {
	ids := make([]string, 0, len(suites))
	for _, suite := range suites {
		ids = append(ids, suite.ID)
	}

	slices.Sort(ids)

	return ids
}
