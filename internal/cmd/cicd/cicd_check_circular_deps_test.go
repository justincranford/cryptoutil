// Package cicd provides tests for circular dependency checking functionality.
package cicd

import (
	"os"
	"testing"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"

	"github.com/stretchr/testify/require"
)

func TestCheckCircularDependencies_NoPackages(t *testing.T) {
	// Test with empty JSON output
	err := checkCircularDependencies("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no packages found")
}

func TestCheckCircularDependencies_InvalidJSON(t *testing.T) {
	// Test with invalid JSON
	err := checkCircularDependencies(`{"invalid": json}`)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse package info")
}

func TestCheckCircularDependencies_NoCircularDeps(t *testing.T) {
	// Test with valid packages but no circular dependencies
	jsonOutput := `{
		"ImportPath": "cryptoutil/internal/common/util",
		"Imports": ["fmt", "strings"]
	}{
		"ImportPath": "cryptoutil/internal/common/config",
		"Imports": ["cryptoutil/internal/common/util", "os"]
	}`

	err := checkCircularDependencies(jsonOutput)
	require.NoError(t, err)
}

func TestCheckCircularDependencies_WithCircularDeps(t *testing.T) {
	// Test with circular dependencies
	jsonOutput := `{
		"ImportPath": "cryptoutil/internal/common/util",
		"Imports": ["cryptoutil/internal/common/config"]
	}{
		"ImportPath": "cryptoutil/internal/common/config",
		"Imports": ["cryptoutil/internal/common/util"]
	}`

	err := checkCircularDependencies(jsonOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular dependencies detected")
	require.Contains(t, err.Error(), "Chain 1")
	require.Contains(t, err.Error(), "cryptoutil/internal/common/util")
	require.Contains(t, err.Error(), "cryptoutil/internal/common/config")
}

func TestCheckCircularDependencies_ComplexCircularDeps(t *testing.T) {
	// Test with more complex circular dependencies (A -> B -> C -> A)
	jsonOutput := `{
		"ImportPath": "cryptoutil/internal/a",
		"Imports": ["cryptoutil/internal/b"]
	}{
		"ImportPath": "cryptoutil/internal/b",
		"Imports": ["cryptoutil/internal/c"]
	}{
		"ImportPath": "cryptoutil/internal/c",
		"Imports": ["cryptoutil/internal/a"]
	}`

	err := checkCircularDependencies(jsonOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular dependencies detected")
	require.Contains(t, err.Error(), "Chain 1")
	require.Contains(t, err.Error(), "cryptoutil/internal/a")
	require.Contains(t, err.Error(), "cryptoutil/internal/b")
	require.Contains(t, err.Error(), "cryptoutil/internal/c")
}

func TestCheckCircularDependencies_IgnoresExternalDeps(t *testing.T) {
	// Test that external dependencies (not starting with cryptoutil/) are ignored
	jsonOutput := `{
		"ImportPath": "cryptoutil/internal/common/util",
		"Imports": ["fmt", "strings", "github.com/stretchr/testify/require"]
	}{
		"ImportPath": "github.com/stretchr/testify/require",
		"Imports": ["cryptoutil/internal/common/util"]
	}`

	// Should not detect circular dependency because external package importing internal is ignored
	err := checkCircularDependencies(jsonOutput)
	require.NoError(t, err)
}

func TestCheckCircularDependencies_MultipleChains(t *testing.T) {
	// Test with multiple separate circular dependency chains
	jsonOutput := `{
		"ImportPath": "cryptoutil/internal/a",
		"Imports": ["cryptoutil/internal/b"]
	}{
		"ImportPath": "cryptoutil/internal/b",
		"Imports": ["cryptoutil/internal/a"]
	}{
		"ImportPath": "cryptoutil/internal/x",
		"Imports": ["cryptoutil/internal/y"]
	}{
		"ImportPath": "cryptoutil/internal/y",
		"Imports": ["cryptoutil/internal/x"]
	}`

	err := checkCircularDependencies(jsonOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular dependencies detected")
	require.Contains(t, err.Error(), "2 circular dependency chain(s)")
}

func TestLoadSaveCircularDepCache(t *testing.T) {
	tests := []struct {
		name  string
		cache cryptoutilMagic.CircularDepCache
	}{
		{
			name: "cache with no circular deps",
			cache: cryptoutilMagic.CircularDepCache{
				LastCheck:       time.Now().UTC(),
				GoModModTime:    time.Now().Add(-1 * time.Hour).UTC(),
				HasCircularDeps: false,
				CircularDeps:    []string{},
			},
		},
		{
			name: "cache with circular deps",
			cache: cryptoutilMagic.CircularDepCache{
				LastCheck:       time.Now().UTC(),
				GoModModTime:    time.Now().Add(-1 * time.Hour).UTC(),
				HasCircularDeps: true,
				CircularDeps:    []string{"circular dependency: pkg1 -> pkg2 -> pkg1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cacheFile := tmpDir + "/test-cache.json"

			// Test save
			err := saveCircularDepCache(cacheFile, tt.cache)
			require.NoError(t, err, "Failed to save cache")

			// Test load
			loadedCache, err := loadCircularDepCache(cacheFile)
			require.NoError(t, err, "Failed to load cache")
			require.NotNil(t, loadedCache, "Loaded cache should not be nil")
			require.Equal(t, tt.cache.HasCircularDeps, loadedCache.HasCircularDeps)
			require.Equal(t, tt.cache.CircularDeps, loadedCache.CircularDeps)
		})
	}
}

func TestLoadCircularDepCache_Errors(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T) string
		expectedError string
	}{
		{
			name: "cache file does not exist",
			setupFunc: func(t *testing.T) string {
				t.Helper()

				return "/nonexistent/cache.json"
			},
			expectedError: "failed to read cache file",
		},
		{
			name: "invalid JSON in cache file",
			setupFunc: func(t *testing.T) string {
				t.Helper()

				tmpDir := t.TempDir()
				cacheFile := tmpDir + "/invalid.json"
				err := os.WriteFile(cacheFile, []byte("{invalid json}"), 0o600)
				require.NoError(t, err)

				return cacheFile
			},
			expectedError: "failed to unmarshal cache JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cacheFile := tt.setupFunc(t)
			_, err := loadCircularDepCache(cacheFile)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestCircularDepCache_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	cacheFile := tmpDir + "/circ-dep-cache.json"

	// Create a cache entry
	originalCache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    time.Now().Add(-2 * time.Hour).UTC(),
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	// Save it
	err := saveCircularDepCache(cacheFile, originalCache)
	require.NoError(t, err)

	// Load it back
	loadedCache, err := loadCircularDepCache(cacheFile)
	require.NoError(t, err)
	require.Equal(t, originalCache.HasCircularDeps, loadedCache.HasCircularDeps)

	// Verify file was created
	_, err = os.Stat(cacheFile)
	require.NoError(t, err)
}
