// Copyright (c) 2025 Justin Cranford
//
//

package realm

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"database/sql"
	"encoding/base64"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	gormsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // Use modernc CGO-free SQLite.
)

// errReader is a reader that always returns an error.
type errReader struct{}

func (e *errReader) Read(_ []byte) (int, error) {
	return 0, fmt.Errorf("simulated read error")
}

func (e *errReader) Close() error { return nil }

// errBodyTransport is an http.RoundTripper that returns a response with a failing body.
type errBodyTransport struct{}

func (t *errBodyTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(&errReader{}),
	}, nil
}

// TestAuthenticate_UnsupportedRealmType covers the default branch in the Authenticate switch.
func TestAuthenticate_UnsupportedRealmType(t *testing.T) {
	t.Parallel()

	realmID := googleUuid.Must(googleUuid.NewV7()).String()

	auth, err := NewAuthenticator(&Config{
		Realms: []RealmConfig{
			{
				ID:      realmID,
				Name:    "unsupported-realm",
				Type:    RealmTypeFile, // valid type for config validation
				Enabled: true,
			},
		},
		Defaults: RealmDefaults{PasswordPolicy: DefaultPasswordPolicy()},
	})
	require.NoError(t, err)

	// Directly inject unsupported realm type to bypass config validation.
	auth.mu.Lock()
	auth.realmMap[realmID].Type = "unsupported"
	auth.mu.Unlock()

	result := auth.Authenticate(context.Background(), realmID, "user", "pass")
	require.NotNil(t, result)
	require.Equal(t, "unsupported realm type", result.Error)
}

// TestVerifyPassword_InvalidHashEncoding covers the hash base64 decode error path.
func TestVerifyPassword_InvalidHashEncoding(t *testing.T) {
	t.Parallel()

	auth, err := NewAuthenticator(&Config{
		Realms: []RealmConfig{
			{
				ID:      googleUuid.Must(googleUuid.NewV7()).String(),
				Name:    "test-realm",
				Type:    RealmTypeFile,
				Enabled: true,
			},
		},
		Defaults: RealmDefaults{PasswordPolicy: DefaultPasswordPolicy()},
	})
	require.NoError(t, err)

	policy := &PasswordPolicyConfig{
		Algorithm:  cryptoutilSharedMagic.PBKDF2DefaultAlgorithm,
		Iterations: cryptoutilSharedMagic.IMPBKDF2Iterations,
		SaltBytes:  cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes,
		HashBytes:  cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes,
	}

	// Valid salt (base64) but invalid hash (non-base64).
	validSalt := base64.StdEncoding.EncodeToString([]byte("valid-salt-32-bytes-padding-here"))
	invalidHash := "$pbkdf2-sha256$600000$" + validSalt + "$!!!invalid-base64!!!"

	err = auth.verifyPassword("password", invalidHash, policy)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid hash encoding")
}

// TestDiscoverOIDC_NonOIDCProvider covers the non-OIDC provider type error.
func TestDiscoverOIDC_NonOIDCProvider(t *testing.T) {
	t.Parallel()

	manager := NewFederationManager(nil)

	err := manager.RegisterProvider(&FederatedProvider{
		ID:        "saml-provider",
		IssuerURL: "https://issuer.example.com",
		Type:      FederationTypeOIDC,
	})
	require.NoError(t, err)

	// Directly set type to non-OIDC after registration.
	manager.mu.Lock()
	manager.providers["saml-provider"].Type = "saml"
	manager.mu.Unlock()

	doc, err := manager.DiscoverOIDC(context.Background(), "saml-provider")
	require.Error(t, err)
	require.Nil(t, doc)
	require.Contains(t, err.Error(), "not OIDC type")
}

// TestDiscoverOIDC_ErrorPaths covers HTTP error, bad status, and bad JSON responses.
func TestDiscoverOIDC_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		errMsg  string
	}{
		{
			name: "non-200 status",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			errMsg: "returned status 500",
		},
		{
			name: "invalid JSON",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte("{invalid json"))
			},
			errMsg: "failed to parse discovery document",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(tc.handler)
			defer server.Close()

			manager := NewFederationManager(nil)

			providerID := "discover-" + tc.name

			err := manager.RegisterProvider(&FederatedProvider{
				ID:        providerID,
				IssuerURL: server.URL,
				Type:      FederationTypeOIDC,
			})
			require.NoError(t, err)

			doc, err := manager.DiscoverOIDC(context.Background(), providerID)
			require.Error(t, err)
			require.Nil(t, doc)
			require.Contains(t, err.Error(), tc.errMsg)
		})
	}
}

// TestDiscoverOIDC_HTTPError covers the HTTP Do error path (unreachable server).
func TestDiscoverOIDC_HTTPError(t *testing.T) {
	t.Parallel()

	manager := NewFederationManager(nil)

	err := manager.RegisterProvider(&FederatedProvider{
		ID:        "unreachable-provider",
		IssuerURL: "http://127.0.0.1:1", // port 1 — connection refused
		Type:      FederationTypeOIDC,
	})
	require.NoError(t, err)

	doc, err := manager.DiscoverOIDC(context.Background(), "unreachable-provider")
	require.Error(t, err)
	require.Nil(t, doc)
	require.Contains(t, err.Error(), "failed to fetch discovery document")
}

// TestDiscoverOIDC_ReadAllError covers the io.ReadAll error path.
func TestDiscoverOIDC_ReadAllError(t *testing.T) {
	t.Parallel()

	manager := NewFederationManager(nil)
	manager.httpClient = &http.Client{Transport: &errBodyTransport{}}

	err := manager.RegisterProvider(&FederatedProvider{
		ID:        "readall-error-provider",
		IssuerURL: "https://issuer.example.com",
		Type:      FederationTypeOIDC,
	})
	require.NoError(t, err)

	doc, err := manager.DiscoverOIDC(context.Background(), "readall-error-provider")
	require.Error(t, err)
	require.Nil(t, doc)
	require.Contains(t, err.Error(), "failed to read discovery response")
}

// TestUnregisterProvider_WithCacheEntries covers the cache cleanup loop in UnregisterProvider.
func TestUnregisterProvider_WithCacheEntries(t *testing.T) {
	t.Parallel()

	// Create mock OIDC discovery server.
	discoveryDoc := OIDCDiscoveryDocument{
		Issuer:  "https://issuer.example.com",
		JWKSURI: "https://issuer.example.com/.well-known/jwks.json",
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
		ID:        "cached-unregister",
		IssuerURL: server.URL,
		Type:      FederationTypeOIDC,
	})
	require.NoError(t, err)

	// Populate discovery cache by calling DiscoverOIDC.
	doc, err := manager.DiscoverOIDC(context.Background(), "cached-unregister")
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Verify cache entry exists.
	manager.mu.RLock()
	_, cached := manager.discoveryCache["cached-unregister:discovery"]
	manager.mu.RUnlock()
	require.True(t, cached)

	// Unregister — should clear cache entries.
	err = manager.UnregisterProvider("cached-unregister")
	require.NoError(t, err)

	// Verify cache entry was cleared.
	manager.mu.RLock()
	_, cached = manager.discoveryCache["cached-unregister:discovery"]
	manager.mu.RUnlock()
	require.False(t, cached)
}

// TestMapTenantFromClaims_NonStringArrayElement covers the non-string element in []any.
func TestMapTenantFromClaims_NonStringArrayElement(t *testing.T) {
	t.Parallel()

	manager := NewFederationManager(nil)

	err := manager.RegisterProvider(&FederatedProvider{
		ID:        "array-provider",
		IssuerURL: "https://issuer.example.com",
		Type:      FederationTypeOIDC,
		TenantMappings: []TenantMapping{
			{
				ClaimName:  "groups",
				ClaimValue: "admin",
				TenantID:   "tenant-1",
			},
		},
	})
	require.NoError(t, err)

	// Array with non-string elements — should skip without matching.
	claims := map[string]any{
		"groups": []any{123, true, nil},
	}

	tenantID, err := manager.MapTenantFromClaims("array-provider", claims)
	require.Error(t, err)
	require.Empty(t, tenantID)
	require.Contains(t, err.Error(), "no matching tenant mapping found")
}

// TestMapTenantFromClaims_DefaultTypeConversion covers the default type case (non-string, non-array).
func TestMapTenantFromClaims_DefaultTypeConversion(t *testing.T) {
	t.Parallel()

	manager := NewFederationManager(nil)

	err := manager.RegisterProvider(&FederatedProvider{
		ID:        "default-type-provider",
		IssuerURL: "https://issuer.example.com",
		Type:      FederationTypeOIDC,
		TenantMappings: []TenantMapping{
			{
				ClaimName:  "org_id",
				ClaimValue: "42",
				TenantID:   "tenant-42",
			},
		},
	})
	require.NoError(t, err)

	// Integer claim value — hits default case with fmt.Sprintf conversion.
	claims := map[string]any{
		"org_id": cryptoutilSharedMagic.AnswerToLifeUniverseEverything,
	}

	tenantID, err := manager.MapTenantFromClaims("default-type-provider", claims)
	require.NoError(t, err)
	require.Equal(t, "tenant-42", tenantID)
}

// TestGetJWKSURL_DiscoverError covers the DiscoverOIDC failure path in GetJWKSURL.
func TestGetJWKSURL_DiscoverError(t *testing.T) {
	t.Parallel()

	manager := NewFederationManager(nil)

	err := manager.RegisterProvider(&FederatedProvider{
		ID:        "no-jwks-provider",
		IssuerURL: "http://127.0.0.1:1", // unreachable
		Type:      FederationTypeOIDC,
		JWKSURL:   "", // no configured JWKS URL — forces discovery
	})
	require.NoError(t, err)

	url, err := manager.GetJWKSURL(context.Background(), "no-jwks-provider")
	require.Error(t, err)
	require.Empty(t, url)
	require.Contains(t, err.Error(), "failed to discover JWKS URL")
}

// TestLoadConfig_PermissionError covers the non-ENOENT file read error path.
func TestLoadConfig_PermissionError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "realms.yml")

	// Create the file then make it unreadable.
	err := os.WriteFile(configPath, []byte("test"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	err = os.Chmod(configPath, 0o000)
	require.NoError(t, err)

	config, err := LoadConfig(tmpDir)
	require.Error(t, err)
	require.Nil(t, config)
	require.Contains(t, err.Error(), "failed to read realms.yml")
}

// TestCreateTenantSchema_DBError covers the Exec error path in createTenantSchema.
func TestCreateTenantSchema_DBError(t *testing.T) {
	t.Parallel()

	// Open a real SQLite DB then close the underlying sql.DB to force errors.
	dbName := fmt.Sprintf("file:%s?mode=memory&cache=shared", googleUuid.Must(googleUuid.NewV7()).String())

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dbName)
	require.NoError(t, err)

	db, err := gorm.Open(gormsqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	require.NoError(t, err)

	err = sqlDB.Close()
	require.NoError(t, err)

	manager, err := NewTenantManager(db, &TenantManagerConfig{
		IsolationMode: TenantIsolationSchema,
	})
	require.NoError(t, err)

	tenant := &TenantConfig{
		ID:         googleUuid.Must(googleUuid.NewV7()).String(),
		SchemaName: fmt.Sprintf("test_schema_%s", googleUuid.Must(googleUuid.NewV7()).String()),
	}

	err = manager.createTenantSchema(context.Background(), tenant)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create schema")
}

// TestDropTenantSchema_DBError covers the Exec error path in dropTenantSchema.
func TestDropTenantSchema_DBError(t *testing.T) {
	t.Parallel()

	dbName := fmt.Sprintf("file:%s?mode=memory&cache=shared", googleUuid.Must(googleUuid.NewV7()).String())

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dbName)
	require.NoError(t, err)

	db, err := gorm.Open(gormsqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	require.NoError(t, err)

	err = sqlDB.Close()
	require.NoError(t, err)

	manager, err := NewTenantManager(db, &TenantManagerConfig{
		IsolationMode: TenantIsolationSchema,
	})
	require.NoError(t, err)

	tenant := &TenantConfig{
		ID:         googleUuid.Must(googleUuid.NewV7()).String(),
		SchemaName: fmt.Sprintf("test_schema_%s", googleUuid.Must(googleUuid.NewV7()).String()),
	}

	err = manager.dropTenantSchema(context.Background(), tenant)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to drop schema")
}
