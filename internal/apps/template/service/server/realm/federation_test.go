// Copyright (c) 2025 Justin Cranford
//
//

package realm

import (
	"context"
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFederationManager(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config *FederationManagerConfig
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
		},
		{
			name:   "zero timeout uses default",
			config: &FederationManagerConfig{HTTPTimeout: 0},
		},
		{
			name: "custom config",
			config: &FederationManagerConfig{
				HTTPTimeout:       defaultHTTPTimeout * 2,
				DiscoveryCacheTTL: defaultDiscoveryCacheTTL * 2,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			manager := NewFederationManager(tc.config)
			require.NotNil(t, manager)
			require.NotNil(t, manager.providers)
			require.NotNil(t, manager.discoveryCache)
			require.NotNil(t, manager.httpClient)
		})
	}
}

func TestFederationManager_RegisterProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider *FederatedProvider
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "nil provider",
			provider: nil,
			wantErr:  true,
			errMsg:   "provider cannot be nil",
		},
		{
			name: "empty ID",
			provider: &FederatedProvider{
				ID:        "",
				IssuerURL: "https://issuer.example.com",
				Type:      FederationTypeOIDC,
			},
			wantErr: true,
			errMsg:  "provider ID is required",
		},
		{
			name: "empty issuer URL",
			provider: &FederatedProvider{
				ID:        "provider1",
				IssuerURL: "",
				Type:      FederationTypeOIDC,
			},
			wantErr: true,
			errMsg:  "issuer URL is required",
		},
		{
			name: "unsupported provider type",
			provider: &FederatedProvider{
				ID:        "provider1",
				IssuerURL: "https://issuer.example.com",
				Type:      "invalid",
			},
			wantErr: true,
			errMsg:  "unsupported provider type",
		},
		{
			name: "valid OIDC provider",
			provider: &FederatedProvider{
				ID:        "oidc-provider",
				Name:      "OIDC Provider",
				IssuerURL: "https://issuer.example.com",
				Type:      FederationTypeOIDC,
				ClientID:  "client-id",
				Enabled:   true,
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			manager := NewFederationManager(nil)
			err := manager.RegisterProvider(tc.provider)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFederationManager_RegisterProvider_Duplicate(t *testing.T) {
	t.Parallel()

	manager := NewFederationManager(nil)

	provider := &FederatedProvider{
		ID:        "dup-provider",
		IssuerURL: "https://issuer.example.com",
		Type:      FederationTypeOIDC,
	}

	err := manager.RegisterProvider(provider)
	require.NoError(t, err)

	err = manager.RegisterProvider(provider)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")
}

func TestFederationManager_GetProvider(t *testing.T) {
	t.Parallel()

	manager := NewFederationManager(nil)

	provider := &FederatedProvider{
		ID:        "test-provider",
		Name:      "Test Provider",
		IssuerURL: "https://issuer.example.com",
		Type:      FederationTypeOIDC,
	}

	err := manager.RegisterProvider(provider)
	require.NoError(t, err)

	t.Run("existing provider", func(t *testing.T) {
		t.Parallel()

		found, ok := manager.GetProvider("test-provider")
		require.True(t, ok)
		require.NotNil(t, found)
		require.Equal(t, "test-provider", found.ID)
	})

	t.Run("non-existing provider", func(t *testing.T) {
		t.Parallel()

		found, ok := manager.GetProvider("non-existing")
		require.False(t, ok)
		require.Nil(t, found)
	})
}

func TestFederationManager_GetProviderByIssuer(t *testing.T) {
	t.Parallel()

	manager := NewFederationManager(nil)

	provider := &FederatedProvider{
		ID:        "issuer-provider",
		IssuerURL: "https://issuer.example.com/",
		Type:      FederationTypeOIDC,
	}

	err := manager.RegisterProvider(provider)
	require.NoError(t, err)

	tests := []struct {
		name      string
		issuerURL string
		wantFound bool
	}{
		{
			name:      "exact match",
			issuerURL: "https://issuer.example.com/",
			wantFound: true,
		},
		{
			name:      "without trailing slash",
			issuerURL: "https://issuer.example.com",
			wantFound: true,
		},
		{
			name:      "non-matching issuer",
			issuerURL: "https://other.example.com",
			wantFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			found, ok := manager.GetProviderByIssuer(tc.issuerURL)
			require.Equal(t, tc.wantFound, ok)

			if tc.wantFound {
				require.NotNil(t, found)
			} else {
				require.Nil(t, found)
			}
		})
	}
}

func TestFederationManager_ListProviders(t *testing.T) {
	t.Parallel()

	manager := NewFederationManager(nil)

	t.Run("empty list", func(t *testing.T) {
		providers := manager.ListProviders()
		require.Empty(t, providers)
	})

	err := manager.RegisterProvider(&FederatedProvider{
		ID:        "provider1",
		IssuerURL: "https://issuer1.example.com",
		Type:      FederationTypeOIDC,
	})
	require.NoError(t, err)

	err = manager.RegisterProvider(&FederatedProvider{
		ID:        "provider2",
		IssuerURL: "https://issuer2.example.com",
		Type:      FederationTypeOIDC,
	})
	require.NoError(t, err)

	t.Run("two providers", func(t *testing.T) {
		providers := manager.ListProviders()
		require.Len(t, providers, 2)
	})
}

func TestFederationManager_UnregisterProvider(t *testing.T) {
	t.Parallel()

	manager := NewFederationManager(nil)

	provider := &FederatedProvider{
		ID:        "unregister-provider",
		IssuerURL: "https://issuer.example.com",
		Type:      FederationTypeOIDC,
	}

	err := manager.RegisterProvider(provider)
	require.NoError(t, err)

	t.Run("unregister existing", func(t *testing.T) {
		err := manager.UnregisterProvider("unregister-provider")
		require.NoError(t, err)

		_, ok := manager.GetProvider("unregister-provider")
		require.False(t, ok)
	})

	t.Run("unregister non-existing", func(t *testing.T) {
		err := manager.UnregisterProvider("non-existing")
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
	})
}

func TestFederationManager_DiscoverOIDC(t *testing.T) {
	t.Parallel()

	t.Run("successful discovery", func(t *testing.T) {
		t.Parallel()

		// Create mock OIDC discovery server for this test only.
		discoveryDoc := OIDCDiscoveryDocument{
			Issuer:                "https://issuer.example.com",
			AuthorizationEndpoint: "https://issuer.example.com/authorize",
			TokenEndpoint:         "https://issuer.example.com/token",
			JWKSURI:               "https://issuer.example.com/.well-known/jwks.json",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/.well-known/openid-configuration" {
				w.Header().Set("Content-Type", "application/json")

				if err := json.NewEncoder(w).Encode(discoveryDoc); err != nil {
					http.Error(w, "encoding error", http.StatusInternalServerError)
				}

				return
			}

			http.NotFound(w, r)
		}))
		defer server.Close()

		manager := NewFederationManager(nil)

		err := manager.RegisterProvider(&FederatedProvider{
			ID:        "oidc-discovery",
			IssuerURL: server.URL,
			Type:      FederationTypeOIDC,
		})
		require.NoError(t, err)

		doc, err := manager.DiscoverOIDC(context.Background(), "oidc-discovery")
		require.NoError(t, err)
		require.NotNil(t, doc)
		require.Equal(t, "https://issuer.example.com", doc.Issuer)
	})

	t.Run("non-existing provider", func(t *testing.T) {
		t.Parallel()

		manager := NewFederationManager(nil)

		doc, err := manager.DiscoverOIDC(context.Background(), "non-existing")
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
		require.Nil(t, doc)
	})
}

func TestFederationManager_DiscoverOIDC_Caching(t *testing.T) {
	t.Parallel()

	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++

		w.Header().Set("Content-Type", "application/json")

		doc := OIDCDiscoveryDocument{Issuer: "https://issuer.example.com"}
		if err := json.NewEncoder(w).Encode(doc); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	manager := NewFederationManager(nil)

	err := manager.RegisterProvider(&FederatedProvider{
		ID:        "cached-provider",
		IssuerURL: server.URL,
		Type:      FederationTypeOIDC,
	})
	require.NoError(t, err)

	// First call should fetch.
	_, err = manager.DiscoverOIDC(context.Background(), "cached-provider")
	require.NoError(t, err)
	require.Equal(t, 1, callCount)

	// Second call should use cache.
	_, err = manager.DiscoverOIDC(context.Background(), "cached-provider")
	require.NoError(t, err)
	require.Equal(t, 1, callCount)
}

func TestFederationManager_MapTenantFromClaims(t *testing.T) {
	t.Parallel()

	manager := NewFederationManager(nil)

	err := manager.RegisterProvider(&FederatedProvider{
		ID:        "mapping-provider",
		IssuerURL: "https://issuer.example.com",
		Type:      FederationTypeOIDC,
		TenantMappings: []TenantMapping{
			{ClaimName: "org", ClaimValue: "acme", TenantID: "tenant-acme", Priority: 10},
			{ClaimName: "groups", ClaimValue: "admin", TenantID: "tenant-admin", Priority: 5},
			{ClaimName: "sub", ClaimValue: "user123", TenantID: "tenant-user", Priority: 20},
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
	require.Equal(t, TenantIsolationMode("database"), TenantIsolationDatabase)
}
