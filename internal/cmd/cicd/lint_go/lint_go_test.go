// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCheckDependencies_NoCycles(t *testing.T) {
	t.Parallel()

	// Simulate go list -json output with no cycles.
	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": ["example.com/pkg/c"]}
{"ImportPath": "example.com/pkg/c", "Imports": []}`

	err := CheckDependencies(goListOutput)
	require.NoError(t, err, "Should not detect cycles in acyclic graph")
}

func TestCheckDependencies_WithCycle(t *testing.T) {
	t.Parallel()

	// Simulate go list -json output with a cycle: a -> b -> c -> a.
	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": ["example.com/pkg/c"]}
{"ImportPath": "example.com/pkg/c", "Imports": ["example.com/pkg/a"]}`

	err := CheckDependencies(goListOutput)
	require.Error(t, err, "Should detect cycle")
	require.Contains(t, err.Error(), "circular dependency", "Error should mention circular dependency")
}

func TestCheckDependencies_EmptyOutput(t *testing.T) {
	t.Parallel()

	err := CheckDependencies("")
	require.NoError(t, err, "Empty output should not cause error")
}

func TestCheckDependencies_SinglePackage(t *testing.T) {
	t.Parallel()

	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": []}`

	err := CheckDependencies(goListOutput)
	require.NoError(t, err, "Single package with no imports should not cause error")
}

func TestCheckDependencies_ExternalDepsIgnored(t *testing.T) {
	t.Parallel()

	// External dependencies should be ignored.
	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["github.com/external/pkg", "example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": ["fmt", "github.com/another/pkg"]}`

	err := CheckDependencies(goListOutput)
	require.NoError(t, err, "External dependencies should be ignored")
}

func TestGetModulePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		packages map[string][]string
		expected string
	}{
		{
			name:     "empty packages",
			packages: map[string][]string{},
			expected: "",
		},
		{
			name: "single package",
			packages: map[string][]string{
				"example.com/pkg/a": {},
			},
			expected: "example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := getModulePath(tc.packages)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestLoadCircularDepCache_FileNotFound tests loading cache when file doesn't exist.
func TestLoadCircularDepCache_FileNotFound(t *testing.T) {
	t.Parallel()

	cache, err := loadCircularDepCache("nonexistent-file-12345.json")
	require.Error(t, err)
	require.Nil(t, cache)
	require.Contains(t, err.Error(), "failed to read cache file")
}

// TestLoadCircularDepCache_InvalidJSON tests loading cache with malformed JSON.
func TestLoadCircularDepCache_InvalidJSON(t *testing.T) {
	t.Parallel()

	// Create temp file with invalid JSON.
	tmpFile := filepath.Join(t.TempDir(), "invalid-cache.json")
	err := os.WriteFile(tmpFile, []byte("{invalid json}"), 0o600)
	require.NoError(t, err)

	cache, err := loadCircularDepCache(tmpFile)
	require.Error(t, err)
	require.Nil(t, cache)
	require.Contains(t, err.Error(), "failed to unmarshal cache JSON")
}

// TestSaveLoadCircularDepCache_RoundTrip tests save/load cycle.
func TestSaveLoadCircularDepCache_RoundTrip(t *testing.T) {
	t.Parallel()

	tmpFile := filepath.Join(t.TempDir(), "cache.json")

	// Create cache data.
	original := cryptoutilSharedMagic.CircularDepCache{
		LastCheck:       time.Now().UTC().Truncate(time.Second), // Truncate to avoid precision loss.
		GoModModTime:    time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Second),
		HasCircularDeps: true,
		CircularDeps:    []string{"pkg/a -> pkg/b -> pkg/a"},
	}

	// Save cache.
	err := saveCircularDepCache(tmpFile, original)
	require.NoError(t, err)

	// Verify file exists.
	_, err = os.Stat(tmpFile)
	require.NoError(t, err)

	// Load cache.
	loaded, err := loadCircularDepCache(tmpFile)
	require.NoError(t, err)
	require.NotNil(t, loaded)

	// Verify data matches.
	require.Equal(t, original.HasCircularDeps, loaded.HasCircularDeps)
	require.Equal(t, original.CircularDeps, loaded.CircularDeps)
	require.True(t, original.LastCheck.Equal(loaded.LastCheck))
	require.True(t, original.GoModModTime.Equal(loaded.GoModModTime))
}

// TestSaveCircularDepCache_CreateDirectory tests directory creation.
func TestSaveCircularDepCache_CreateDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "subdir", "cache.json")

	cache := cryptoutilSharedMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    time.Now().UTC(),
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	err := saveCircularDepCache(cacheFile, cache)
	require.NoError(t, err)

	// Verify file and directory exist.
	_, err = os.Stat(cacheFile)
	require.NoError(t, err)
}

func TestLint(t *testing.T) {
	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Lint(logger)

	require.Error(t, err, "Lint fails when go.mod not in current directory")
	require.Contains(t, err.Error(), "lint-go failed")
}

func TestCheckDependencies_MalformedJSON(t *testing.T) {
	t.Parallel()

	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
invalid json line
{"ImportPath": "example.com/pkg/b", "Imports": []}`

	err := CheckDependencies(goListOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode package info")
}

func TestCheckDependencies_ComplexCycle(t *testing.T) {
	t.Parallel()

	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b", "example.com/pkg/c"]}
{"ImportPath": "example.com/pkg/b", "Imports": ["example.com/pkg/d"]}
{"ImportPath": "example.com/pkg/c", "Imports": ["example.com/pkg/d"]}
{"ImportPath": "example.com/pkg/d", "Imports": ["example.com/pkg/a"]}`

	err := CheckDependencies(goListOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular dependency")
}

func TestCheckDependencies_SelfReference(t *testing.T) {
	t.Parallel()

	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/a"]}`

	err := CheckDependencies(goListOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular dependency")
}

func TestCheckDependencies_MultipleDisconnectedGraphs(t *testing.T) {
	t.Parallel()

	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": []}
{"ImportPath": "example.com/pkg/c", "Imports": ["example.com/pkg/d"]}
{"ImportPath": "example.com/pkg/d", "Imports": []}`

	err := CheckDependencies(goListOutput)
	require.NoError(t, err)
}

func TestGetModulePath_MultiplePackages(t *testing.T) {
	t.Parallel()

	packages := map[string][]string{
		"example.com/pkg/a": {},
		"example.com/pkg/b": {},
		"example.com/pkg/c": {},
	}

	result := getModulePath(packages)
	require.Equal(t, "example.com", result)
}

func TestGetModulePath_DifferentPrefixes(t *testing.T) {
	t.Parallel()

	packages := map[string][]string{
		"github.com/user/repo/pkg/a": {},
		"github.com/user/repo/pkg/b": {},
	}

	result := getModulePath(packages)
	require.Equal(t, "github.com", result)
}
