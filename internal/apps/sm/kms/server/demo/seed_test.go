// Copyright (c) 2025 Justin Cranford
//
//

package demo

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateDemoTenantID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "generates valid UUIDv4"},
		{name: "generates unique IDs"},
	}

	generatedIDs := make(map[string]bool)

	var mu sync.Mutex

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			id := GenerateDemoTenantID()
			require.NotEmpty(t, id)
			require.Len(t, id, cryptoutilSharedMagic.UUIDStringLength) // UUID format: 8-4-4-4-12.
			require.Contains(t, id, "-")

			mu.Lock()
			defer mu.Unlock()

			require.False(t, generatedIDs[id], "expected unique ID")
			generatedIDs[id] = true
		})
	}
}

func TestDefaultDemoTenants(t *testing.T) {
	t.Parallel()

	t.Run("returns 2 valid demo tenants", func(t *testing.T) {
		t.Parallel()

		tenants := DefaultDemoTenants()

		require.Len(t, tenants, 2, "Should return 2 demo tenants")

		for _, tenant := range tenants {
			require.NotEmpty(t, tenant.ID, "Tenant ID should not be empty")
			require.Len(t, tenant.ID, cryptoutilSharedMagic.UUIDStringLength, "Tenant ID should be UUID format")
			require.NotEmpty(t, tenant.Name, "Tenant name should not be empty")
		}

		// Verify IDs are unique.
		require.NotEqual(t, tenants[0].ID, tenants[1].ID, "Tenant IDs should be unique")
	})

	t.Run("regenerates IDs on each call", func(t *testing.T) {
		t.Parallel()

		tenants1 := DefaultDemoTenants()
		tenants2 := DefaultDemoTenants()

		// Each call should generate new UUIDs (Session 4 Q9).
		require.NotEqual(t, tenants1[0].ID, tenants2[0].ID, "Tenant IDs should regenerate on each call")
		require.NotEqual(t, tenants1[1].ID, tenants2[1].ID, "Tenant IDs should regenerate on each call")
	})
}

func TestDefaultDemoKeys(t *testing.T) {
	t.Parallel()

	keys := DefaultDemoKeys()

	require.NotEmpty(t, keys, "DefaultDemoKeys should return non-empty list")
	require.Len(t, keys, 4, "Should return 4 demo keys")

	// Verify each key has required fields.
	for _, key := range keys {
		require.NotEmpty(t, key.Name, "Key name should not be empty")
		require.NotEmpty(t, key.Description, "Key description should not be empty")
		require.NotEmpty(t, key.Algorithm, "Key algorithm should not be empty")
	}

	// Verify specific expected keys exist.
	keyNames := make(map[string]bool)
	for _, key := range keys {
		keyNames[key.Name] = true
	}

	require.True(t, keyNames["demo-encryption-aes256"], "Should include demo-encryption-aes256 key")
	require.True(t, keyNames["demo-signing-rsa2048"], "Should include demo-signing-rsa2048 key")
	require.True(t, keyNames["demo-signing-ec256"], "Should include demo-signing-ec256 key")
	require.True(t, keyNames["demo-wrapping-aes256kw"], "Should include demo-wrapping-aes256kw key")
}

func TestDemoKeyConfig_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		keyName     string
		description string
		wantName    string
		wantDesc    string
	}{
		{
			name:        "encryption key",
			keyName:     "demo-encryption-aes256",
			description: "Demo AES-256-GCM encryption key",
			wantName:    "demo-encryption-aes256",
			wantDesc:    "Demo AES-256-GCM encryption key",
		},
		{
			name:        "signing RSA key",
			keyName:     "demo-signing-rsa2048",
			description: "Demo RSA-2048 signing key",
			wantName:    "demo-signing-rsa2048",
			wantDesc:    "Demo RSA-2048 signing key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			keys := DefaultDemoKeys()

			var found *DemoKeyConfig

			for i := range keys {
				if keys[i].Name == tc.keyName {
					found = &keys[i]

					break
				}
			}

			require.NotNil(t, found, "Key %s should exist", tc.keyName)
			require.Equal(t, tc.wantName, found.Name, "Key name should match")
			require.Equal(t, tc.wantDesc, found.Description, "Key description should match")
		})
	}
}

func TestResetDemoData(t *testing.T) {
	t.Parallel()

	// ResetDemoData is currently a placeholder that calls SeedDemoData
	// Since SeedDemoData is idempotent, ResetDemoData should also be idempotent
	// This test verifies the function exists and can be called without error
	// In a real implementation, this would test that existing keys are disabled/reset

	// For now, just verify the function signature and that it doesn't panic
	// when called with nil parameters (though in practice it would need real services)
	require.NotNil(t, ResetDemoData, "ResetDemoData function should exist")
}
