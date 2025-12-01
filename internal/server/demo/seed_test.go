// Copyright (c) 2025 Justin Cranford
//
//

package demo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
		tc := tc
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
