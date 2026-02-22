// Copyright (c) 2025 Justin Cranford
//
//

package hash

import (
	"strings"
	"testing"

	cryptoutilSharedCryptoDigests "cryptoutil/internal/shared/crypto/digests"

	"github.com/stretchr/testify/require"
)

// TestParameterSetRegistry_GetDefaultParameterSet tests retrieving the default parameter set.
func TestParameterSetRegistry_GetDefaultParameterSet(t *testing.T) {
	t.Parallel()

	registry := NewParameterSetRegistry()
	params := registry.GetDefaultParameterSet()

	require.NotNil(t, params, "default parameter set should not be nil")
	require.Equal(t, "1", params.Version, "default version should be '1'")
}

// TestParameterSetRegistry_GetParameterSet tests retrieving specific parameter sets.
func TestParameterSetRegistry_GetParameterSet(t *testing.T) {
	t.Parallel()

	registry := NewParameterSetRegistry()

	tests := []struct {
		name            string
		version         string
		expectError     bool
		expectedVersion string
	}{
		{name: "Version1", version: "1", expectError: false, expectedVersion: "1"},
		{name: "Version2", version: "2", expectError: false, expectedVersion: "2"},
		{name: "Version3", version: "3", expectError: false, expectedVersion: "3"},
		{name: "InvalidVersion", version: "99", expectError: true, expectedVersion: ""},
		{name: "EmptyVersion", version: "", expectError: true, expectedVersion: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params, err := registry.GetParameterSet(tt.version)

			if tt.expectError {
				require.Error(t, err, "expected error for invalid version")
				require.Nil(t, params, "params should be nil for invalid version")
				require.Contains(t, err.Error(), "not found", "error should mention version not found")
			} else {
				require.NoError(t, err, "should retrieve valid parameter set")
				require.NotNil(t, params, "params should not be nil")
				require.Equal(t, tt.expectedVersion, params.Version, "version should match")
			}
		})
	}
}

// TestParameterSetRegistry_ListVersions tests listing all registered versions.
func TestParameterSetRegistry_ListVersions(t *testing.T) {
	t.Parallel()

	registry := NewParameterSetRegistry()
	versions := registry.ListVersions()

	require.Len(t, versions, 3, "should have exactly 3 registered versions")
	require.Contains(t, versions, "1", "should contain version 1")
	require.Contains(t, versions, "2", "should contain version 2")
	require.Contains(t, versions, "3", "should contain version 3")
}

// TestParameterSetRegistry_GetDefaultVersion tests retrieving the default version string.
func TestParameterSetRegistry_GetDefaultVersion(t *testing.T) {
	t.Parallel()

	registry := NewParameterSetRegistry()
	defaultVersion := registry.GetDefaultVersion()

	require.Equal(t, "1", defaultVersion, "default version should be '1'")
}

// TestParameterSetRegistry_HashWithAllVersions tests hashing with all parameter sets.
func TestParameterSetRegistry_HashWithAllVersions(t *testing.T) {
	t.Parallel()

	registry := NewParameterSetRegistry()
	secret := "test-password-12345"

	tests := []struct {
		name    string
		version string
	}{
		{name: "V1_600k_iterations", version: "1"},
		{name: "V2_1M_iterations", version: "2"},
		{name: "V3_2M_iterations", version: "3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params, err := registry.GetParameterSet(tt.version)
			require.NoError(t, err, "should retrieve parameter set")

			hash, err := cryptoutilSharedCryptoDigests.PBKDF2WithParams(secret, params)
			require.NoError(t, err, "should hash secret")
			require.NotEmpty(t, hash, "hash should not be empty")
			require.True(t, strings.HasPrefix(hash, "{"+tt.version+"}$"), "hash should have correct version prefix")

			// Verify hash can be validated.
			valid, err := cryptoutilSharedCryptoDigests.VerifySecret(hash, secret)
			require.NoError(t, err, "should verify hash")
			require.True(t, valid, "hash should verify correctly")
		})
	}
}

// TestParameterSetRegistry_CrossVersionVerification tests that hashes from different versions verify correctly.
func TestParameterSetRegistry_CrossVersionVerification(t *testing.T) {
	t.Parallel()

	registry := NewParameterSetRegistry()
	secret := "cross-version-test"

	// Generate hashes with V1, V2, V3.
	hashV1, err := cryptoutilSharedCryptoDigests.PBKDF2WithParams(secret, registry.GetDefaultParameterSet())
	require.NoError(t, err, "should hash with V1")

	paramsV2, err := registry.GetParameterSet("2")
	require.NoError(t, err, "should get V2 params")
	hashV2, err := cryptoutilSharedCryptoDigests.PBKDF2WithParams(secret, paramsV2)
	require.NoError(t, err, "should hash with V2")

	paramsV3, err := registry.GetParameterSet("3")
	require.NoError(t, err, "should get V3 params")
	hashV3, err := cryptoutilSharedCryptoDigests.PBKDF2WithParams(secret, paramsV3)
	require.NoError(t, err, "should hash with V3")

	// All hashes should verify correctly.
	tests := []struct {
		name string
		hash string
	}{
		{name: "V1_hash", hash: hashV1},
		{name: "V2_hash", hash: hashV2},
		{name: "V3_hash", hash: hashV3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			valid, err := cryptoutilSharedCryptoDigests.VerifySecret(tt.hash, secret)
			require.NoError(t, err, "should verify %s", tt.name)
			require.True(t, valid, "%s should verify correctly", tt.name)

			// Wrong password should not verify.
			invalid, err := cryptoutilSharedCryptoDigests.VerifySecret(tt.hash, "wrong-password")
			require.NoError(t, err, "should not error on wrong password")
			require.False(t, invalid, "%s should not verify with wrong password", tt.name)
		})
	}
}

// TestParameterSetRegistry_ConcurrentAccess tests thread-safe registry access.
func TestParameterSetRegistry_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	registry := NewParameterSetRegistry()
	iterations := 100

	// Spawn goroutines to access registry concurrently.
	done := make(chan bool, iterations)

	for i := range iterations {
		go func(idx int) {
			defer func() { done <- true }()

			// Alternate between different operations.
			switch idx % 4 {
			case 0:
				params := registry.GetDefaultParameterSet()
				require.NotNil(t, params)
			case 1:
				params, err := registry.GetParameterSet("1")
				require.NoError(t, err)
				require.NotNil(t, params)
			case 2:
				versions := registry.ListVersions()
				require.Len(t, versions, 3)
			case 3:
				version := registry.GetDefaultVersion()
				require.Equal(t, "1", version)
			}
		}(i)
	}

	// Wait for all goroutines to complete.
	for range iterations {
		<-done
	}
}

// TestGlobalRegistry tests the global registry instance.
func TestGlobalRegistry(t *testing.T) {
	t.Parallel()

	registry := GetGlobalRegistry()
	require.NotNil(t, registry, "global registry should not be nil")

	// Verify it's properly initialized.
	defaultParams := registry.GetDefaultParameterSet()
	require.NotNil(t, defaultParams, "default parameter set should not be nil")
	require.Equal(t, "1", defaultParams.Version, "default version should be '1'")

	versions := registry.ListVersions()
	require.Len(t, versions, 3, "should have exactly 3 registered versions")
}

// TestGetDefaultParameterSet_MissingDefaultPanics tests that GetDefaultParameterSet panics
// when the default version parameter set is missing (programming error condition).
func TestGetDefaultParameterSet_MissingDefaultPanics(t *testing.T) {
	t.Parallel()

	registry := NewParameterSetRegistry()

	// Corrupt the registry by removing the default version.
	registry.mu.Lock()
	delete(registry.parameterSets, registry.defaultVersion)
	registry.mu.Unlock()

	require.Panics(t, func() {
		registry.GetDefaultParameterSet()
	})
}
