// Copyright (c) 2025-2026 Justin Cranford.
//
//

package realm

import (
	"context"
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestFederationManager_MapTenantFromClaims(t *testing.T) {
	t.Parallel()

	manager := NewFederationManager(nil)

	err := manager.RegisterProvider(&FederatedProvider{
		ID:        testMappingProvider,
		IssuerURL: testHTTPSIssuerExampleCom,
		Type:      FederationTypeOIDC,
		TenantMappings: []TenantMapping{
			{ClaimName: testOrg, ClaimValue: testAcme, TenantID: testTenantAcme, Priority: cryptoutilSharedMagic.JoseJADefaultMaxMaterials},
			{ClaimName: testGroups, ClaimValue: testAdmin, TenantID: testTenantAdmin, Priority: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries},
			{ClaimName: cryptoutilSharedMagic.ClaimSub, ClaimValue: "user123", TenantID: "tenant-user", Priority: cryptoutilSharedMagic.MaxErrorDisplay},
		},
	})
	require.NoError(t, err)

	err = manager.RegisterProvider(&FederatedProvider{
		ID:        testNoMappings,
		IssuerURL: "https://no-mappings.example.com",
		Type:      FederationTypeOIDC,
	})
	require.NoError(t, err)

	tests := []struct {
		name       string
		providerID string
		claims     map[string]any
		wantTenant string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       testNonExistingProvider,
			providerID: testNonExisting,
			claims:     map[string]any{testOrg: testAcme},
			wantErr:    true,
			errMsg:     "not found",
		},
		{
			name:       "no tenant mappings",
			providerID: testNoMappings,
			claims:     map[string]any{testOrg: testAcme},
			wantErr:    true,
			errMsg:     "no tenant mappings configured",
		},
		{
			name:       "string claim match",
			providerID: testMappingProvider,
			claims:     map[string]any{testOrg: testAcme},
			wantTenant: testTenantAcme,
			wantErr:    false,
		},
		{
			name:       "array claim match",
			providerID: testMappingProvider,
			claims:     map[string]any{testGroups: []any{testUser, testAdmin, "developer"}},
			wantTenant: testTenantAdmin,
			wantErr:    false,
		},
		{
			name:       "priority ordering (groups has higher priority than org)",
			providerID: testMappingProvider,
			claims:     map[string]any{testOrg: testAcme, testGroups: []any{testAdmin}},
			wantTenant: testTenantAdmin,
			wantErr:    false,
		},
		{
			name:       "no matching claim",
			providerID: testMappingProvider,
			claims:     map[string]any{"other": "value"},
			wantErr:    true,
			errMsg:     "no matching tenant mapping found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tenantID, err := manager.MapTenantFromClaims(tc.providerID, tc.claims)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantTenant, tenantID)
			}
		})
	}
}

func TestFederationManager_ValidateAudience(t *testing.T) {
	t.Parallel()

	manager := NewFederationManager(nil)

	err := manager.RegisterProvider(&FederatedProvider{
		ID:               testAudProvider,
		IssuerURL:        testHTTPSIssuerExampleCom,
		Type:             FederationTypeOIDC,
		AllowedAudiences: []string{testClient1, testClient2},
	})
	require.NoError(t, err)

	err = manager.RegisterProvider(&FederatedProvider{
		ID:        testNoAudProvider,
		IssuerURL: "https://no-aud.example.com",
		Type:      FederationTypeOIDC,
	})
	require.NoError(t, err)

	tests := []struct {
		name       string
		providerID string
		audience   []string
		wantErr    bool
	}{
		{
			name:       testNonExistingProvider,
			providerID: testNonExisting,
			audience:   []string{testClient1},
			wantErr:    true,
		},
		{
			name:       "no audience restriction",
			providerID: testNoAudProvider,
			audience:   []string{"any-client"},
			wantErr:    false,
		},
		{
			name:       "valid audience",
			providerID: testAudProvider,
			audience:   []string{testClient1},
			wantErr:    false,
		},
		{
			name:       "multiple audiences with one valid",
			providerID: testAudProvider,
			audience:   []string{cryptoutilSharedMagic.UNKNOWN, testClient2},
			wantErr:    false,
		},
		{
			name:       "invalid audience",
			providerID: testAudProvider,
			audience:   []string{cryptoutilSharedMagic.UNKNOWN},
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := manager.ValidateAudience(tc.providerID, tc.audience)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFederationManager_GetJWKSURL(t *testing.T) {
	t.Parallel()

	t.Run(testNonExistingProvider, func(t *testing.T) {
		t.Parallel()

		manager := NewFederationManager(nil)

		_, err := manager.GetJWKSURL(context.Background(), testNonExisting)
		require.Error(t, err)
	})

	t.Run("configured JWKS URL", func(t *testing.T) {
		t.Parallel()

		manager := NewFederationManager(nil)

		err := manager.RegisterProvider(&FederatedProvider{
			ID:        "jwks-configured",
			IssuerURL: testHTTPSIssuerExampleCom,
			Type:      FederationTypeOIDC,
			JWKSURL:   "https://custom.example.com/jwks",
		})
		require.NoError(t, err)

		jwksURL, err := manager.GetJWKSURL(context.Background(), "jwks-configured")
		require.NoError(t, err)
		require.Equal(t, "https://custom.example.com/jwks", jwksURL)
	})

	t.Run("discover JWKS URL", func(t *testing.T) {
		t.Parallel()

		discoveryDoc := OIDCDiscoveryDocument{
			Issuer:  testHTTPSIssuerExampleCom,
			JWKSURI: "https://issuer.example.com/jwks",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(discoveryDoc); err != nil {
				http.Error(w, "encoding error", http.StatusInternalServerError)
			}
		}))
		defer server.Close()

		manager := NewFederationManager(nil)

		err := manager.RegisterProvider(&FederatedProvider{
			ID:        "jwks-discover",
			IssuerURL: server.URL,
			Type:      FederationTypeOIDC,
		})
		require.NoError(t, err)

		jwksURL, err := manager.GetJWKSURL(context.Background(), "jwks-discover")
		require.NoError(t, err)
		require.Equal(t, "https://issuer.example.com/jwks", jwksURL)
	})
}

func TestFederationProviderTypes(t *testing.T) {
	t.Parallel()

	require.Equal(t, FederationProviderType("oidc"), FederationTypeOIDC)
}

func TestTenantIsolationModes(t *testing.T) {
	t.Parallel()

	require.Equal(t, TenantIsolationMode("schema"), TenantIsolationSchema)
	require.Equal(t, TenantIsolationMode("row"), TenantIsolationRow)
	require.Equal(t, TenantIsolationMode(cryptoutilSharedMagic.RealmStorageTypeDatabase), TenantIsolationDatabase)
}
