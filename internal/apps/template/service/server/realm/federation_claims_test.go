// Copyright (c) 2025 Justin Cranford
//
//

package realm

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFederationManager_MapTenantFromClaims(t *testing.T) {
	t.Parallel()

	manager := NewFederationManager(nil)

	err := manager.RegisterProvider(&FederatedProvider{
		ID:        "mapping-provider",
		IssuerURL: "https://issuer.example.com",
		Type:      FederationTypeOIDC,
		TenantMappings: []TenantMapping{
			{ClaimName: "org", ClaimValue: "acme", TenantID: "tenant-acme", Priority: cryptoutilSharedMagic.JoseJADefaultMaxMaterials},
			{ClaimName: "groups", ClaimValue: "admin", TenantID: "tenant-admin", Priority: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries},
			{ClaimName: cryptoutilSharedMagic.ClaimSub, ClaimValue: "user123", TenantID: "tenant-user", Priority: cryptoutilSharedMagic.MaxErrorDisplay},
		},
	})
	require.NoError(t, err)

	err = manager.RegisterProvider(&FederatedProvider{
		ID:        "no-mappings",
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
			name:       "non-existing provider",
			providerID: "non-existing",
			claims:     map[string]any{"org": "acme"},
			wantErr:    true,
			errMsg:     "not found",
		},
		{
			name:       "no tenant mappings",
			providerID: "no-mappings",
			claims:     map[string]any{"org": "acme"},
			wantErr:    true,
			errMsg:     "no tenant mappings configured",
		},
		{
			name:       "string claim match",
			providerID: "mapping-provider",
			claims:     map[string]any{"org": "acme"},
			wantTenant: "tenant-acme",
			wantErr:    false,
		},
		{
			name:       "array claim match",
			providerID: "mapping-provider",
			claims:     map[string]any{"groups": []any{"user", "admin", "developer"}},
			wantTenant: "tenant-admin",
			wantErr:    false,
		},
		{
			name:       "priority ordering (groups has higher priority than org)",
			providerID: "mapping-provider",
			claims:     map[string]any{"org": "acme", "groups": []any{"admin"}},
			wantTenant: "tenant-admin",
			wantErr:    false,
		},
		{
			name:       "no matching claim",
			providerID: "mapping-provider",
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
		ID:               "aud-provider",
		IssuerURL:        "https://issuer.example.com",
		Type:             FederationTypeOIDC,
		AllowedAudiences: []string{"client1", "client2"},
	})
	require.NoError(t, err)

	err = manager.RegisterProvider(&FederatedProvider{
		ID:        "no-aud-provider",
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
			name:       "non-existing provider",
			providerID: "non-existing",
			audience:   []string{"client1"},
			wantErr:    true,
		},
		{
			name:       "no audience restriction",
			providerID: "no-aud-provider",
			audience:   []string{"any-client"},
			wantErr:    false,
		},
		{
			name:       "valid audience",
			providerID: "aud-provider",
			audience:   []string{"client1"},
			wantErr:    false,
		},
		{
			name:       "multiple audiences with one valid",
			providerID: "aud-provider",
			audience:   []string{"unknown", "client2"},
			wantErr:    false,
		},
		{
			name:       "invalid audience",
			providerID: "aud-provider",
			audience:   []string{"unknown"},
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

	t.Run("non-existing provider", func(t *testing.T) {
		t.Parallel()

		manager := NewFederationManager(nil)

		_, err := manager.GetJWKSURL(context.Background(), "non-existing")
		require.Error(t, err)
	})

	t.Run("configured JWKS URL", func(t *testing.T) {
		t.Parallel()

		manager := NewFederationManager(nil)

		err := manager.RegisterProvider(&FederatedProvider{
			ID:        "jwks-configured",
			IssuerURL: "https://issuer.example.com",
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
			Issuer:  "https://issuer.example.com",
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
