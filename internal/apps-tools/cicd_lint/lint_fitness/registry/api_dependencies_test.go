// Copyright (c) 2025-2026 Justin Cranford.

package registry

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestValidateAPIDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dependencies []string
		wantError    string
	}{
		{name: "known registered target", dependencies: []string{cryptoutilSharedMagic.OTLPServiceIdentityAuthz}},
		{name: "empty target", dependencies: []string{""}, wantError: "target PS-ID must not be empty"},
		{
			name:         "duplicate target",
			dependencies: []string{cryptoutilSharedMagic.OTLPServiceIdentityAuthz, cryptoutilSharedMagic.OTLPServiceIdentityAuthz},
			wantError:    "duplicate target PS-ID " + cryptoutilSharedMagic.OTLPServiceIdentityAuthz,
		},
		{
			name: "self target", dependencies: []string{cryptoutilSharedMagic.OTLPServiceSMKMS},
			wantError: "service must not depend on its own API",
		},
		{name: "unknown target", dependencies: []string{"not-registered"}, wantError: "unknown target PS-ID not-registered"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			registry := *allRegistryFile
			registry.ProductServices = append([]RegistryProductService(nil), allRegistryFile.ProductServices...)
			registry.ProductServices[0].APIDependencies = append([]string(nil), test.dependencies...)

			err := validateRegistry(&registry)
			if test.wantError == "" {
				require.NoError(t, err)

				return
			}

			require.ErrorContains(t, err, test.wantError)
		})
	}
}

func TestAllAPIServiceDependenciesReturnsCanonicalIndependentEdges(t *testing.T) {
	t.Parallel()

	want := []APIServiceDependencyInfo{
		{
			SourcePSID: cryptoutilSharedMagic.OTLPServiceIdentityRP,
			TargetPSID: cryptoutilSharedMagic.OTLPServiceIdentityAuthz,
		},
	}

	first := AllAPIServiceDependencies()
	require.Equal(t, want, first)
	first[0].SourcePSID = "mutated"

	require.Equal(t, want, AllAPIServiceDependencies())
}

func TestAPIServiceDependenciesSortBySourceAndTarget(t *testing.T) {
	t.Parallel()

	services := []RegistryProductService{
		{
			PSID: cryptoutilSharedMagic.OTLPServiceSMKMS,
			APIDependencies: []string{
				cryptoutilSharedMagic.OTLPServiceIdentityRP,
			},
		},
		{
			PSID: cryptoutilSharedMagic.OTLPServiceIdentityRP,
			APIDependencies: []string{
				cryptoutilSharedMagic.OTLPServiceIdentityIDP,
				cryptoutilSharedMagic.OTLPServiceIdentityAuthz,
			},
		},
	}

	got := apiServiceDependencies(services)
	want := []APIServiceDependencyInfo{
		{
			SourcePSID: cryptoutilSharedMagic.OTLPServiceIdentityRP,
			TargetPSID: cryptoutilSharedMagic.OTLPServiceIdentityAuthz,
		},
		{
			SourcePSID: cryptoutilSharedMagic.OTLPServiceIdentityRP,
			TargetPSID: cryptoutilSharedMagic.OTLPServiceIdentityIDP,
		},
		{
			SourcePSID: cryptoutilSharedMagic.OTLPServiceSMKMS,
			TargetPSID: cryptoutilSharedMagic.OTLPServiceIdentityRP,
		},
	}

	require.Equal(t, want, got)
}
