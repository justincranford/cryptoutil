// Copyright (c) 2025 Justin Cranford

package go_check_circular_package_dependencies

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	testify "github.com/stretchr/testify/require"

	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilFiles "cryptoutil/internal/common/util/files"
)

func TestCheckDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		jsonOutput   string
		wantError    bool
		wantContains []string
	}{
		{
			name: "no_cycle",
			jsonOutput: "\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/a\",\"Imports\":[\"cryptoutil/pkg/b\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/b\",\"Imports\":[\"cryptoutil/pkg/c\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/c\",\"Imports\":[]}\n",
			wantError: false,
		},
		{
			name: "with_cycle",
			jsonOutput: "\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/a\",\"Imports\":[\"cryptoutil/pkg/b\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/b\",\"Imports\":[\"cryptoutil/pkg/c\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/c\",\"Imports\":[\"cryptoutil/pkg/a\"]}\n",
			wantError: true,
			wantContains: []string{
				"circular dependencies detected",
				"cryptoutil/pkg/a",
				"cryptoutil/pkg/b",
				"cryptoutil/pkg/c",
			},
		},
		{
			name: "multiple_cycles",
			jsonOutput: "\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/a\",\"Imports\":[\"cryptoutil/pkg/b\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/b\",\"Imports\":[\"cryptoutil/pkg/a\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/x\",\"Imports\":[\"cryptoutil/pkg/y\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/y\",\"Imports\":[\"cryptoutil/pkg/x\"]}\n",
			wantError: true,
			wantContains: []string{
				"circular dependencies detected",
			},
		},
		{
			name: "external_package_ignored",
			jsonOutput: "\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/a\",\"Imports\":[\"github.com/external/pkg\",\"cryptoutil/pkg/b\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/b\",\"Imports\":[\"fmt\"]}\n",
			wantError: false,
		},
		{
			name:       "empty_output",
			jsonOutput: "",
			wantError:  true,
			wantContains: []string{
				"no packages found",
			},
		},
		{
			name:       "invalid_json",
			jsonOutput: "{\"ImportPath\":\"cryptoutil/pkg/a\",\"Imports\":[\"invalid json",
			wantError:  true,
			wantContains: []string{
				"failed to parse package info",
			},
		},
		{
			name: "self_cycle",
			jsonOutput: "\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/a\",\"Imports\":[\"cryptoutil/pkg/a\"]}\n",
			wantError: true,
			wantContains: []string{
				"circular dependencies detected",
			},
		},
		{
			name: "complex_cycle",
			jsonOutput: "\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/a\",\"Imports\":[\"cryptoutil/pkg/b\",\"cryptoutil/pkg/c\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/b\",\"Imports\":[\"cryptoutil/pkg/d\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/c\",\"Imports\":[\"cryptoutil/pkg/d\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/d\",\"Imports\":[\"cryptoutil/pkg/a\"]}\n",
			wantError: true,
			wantContains: []string{
				"circular dependencies detected",
				"Chain",
			},
		},
		{
			name: "long_chain",
			jsonOutput: "\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/a\",\"Imports\":[\"cryptoutil/pkg/b\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/b\",\"Imports\":[\"cryptoutil/pkg/c\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/c\",\"Imports\":[\"cryptoutil/pkg/d\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/d\",\"Imports\":[\"cryptoutil/pkg/e\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/e\",\"Imports\":[\"cryptoutil/pkg/f\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/f\",\"Imports\":[]}\n",
			wantError: false,
		},
		{
			name: "mixed_internal_external",
			jsonOutput: "\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/a\",\"Imports\":[\"github.com/external/x\",\"cryptoutil/pkg/b\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/b\",\"Imports\":[\"golang.org/x/tools\",\"cryptoutil/pkg/c\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/c\",\"Imports\":[\"fmt\",\"encoding/json\"]}\n",
			wantError: false,
		},
		{
			name: "error_message_format",
			jsonOutput: "\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/a\",\"Imports\":[\"cryptoutil/pkg/b\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/b\",\"Imports\":[\"cryptoutil/pkg/a\"]}\n",
			wantError: true,
			wantContains: []string{
				"circular dependencies detected:",
				"Chain 1",
				"packages)",
				"→",
				"Consider refactoring to break these cycles",
			},
		},
		{
			name:       "no_packages_in_graph",
			jsonOutput: "",
			wantError:  true,
			wantContains: []string{
				"no packages found",
			},
		},
		{
			name: "package_with_no_imports",
			jsonOutput: "\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/standalone\",\"Imports\":[]}\n",
			wantError: false,
		},
		{
			name: "only_external_imports",
			jsonOutput: "\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/a\",\"Imports\":[\"github.com/foo/bar\",\"golang.org/x/tools\"]}\n" +
				"{\"ImportPath\":\"cryptoutil/pkg/b\",\"Imports\":[\"fmt\",\"encoding/json\"]}\n",
			wantError: false,
		},
		{
			name: "stress_test",
			jsonOutput: func() string {
				var builder strings.Builder

				numPackages := 100
				for i := 0; i < numPackages; i++ {
					imports := "[]"
					if i > 0 {
						imports = fmt.Sprintf("[\"cryptoutil/pkg/p%d\"]", i-1)
					}

					fmt.Fprintf(&builder, "{\"ImportPath\":\"cryptoutil/pkg/p%d\",\"Imports\":%s}\n", i, imports)
				}

				return builder.String()
			}(),
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := CheckDependencies(tc.jsonOutput)

			if tc.wantError {
				testify.Error(t, err, "Expected error for test case: %s", tc.name)

				for _, contains := range tc.wantContains {
					testify.Contains(t, err.Error(), contains, "Error should contain: %s", contains)
				}
			} else {
				testify.NoError(t, err, "Expected no error for test case: %s", tc.name)
			}
		})
	}
}

func TestCacheOperations(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, ".cicd", "circular-deps-cache.json")

	// Test save cache
	cache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    time.Now().UTC(),
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	err := saveCircularDepCache(cacheFile, cache)
	testify.NoError(t, err, "Save cache should succeed")

	// Verify file exists
	testify.FileExists(t, cacheFile, "Cache file should exist")

	// Test load cache
	loadedCache, err := loadCircularDepCache(cacheFile)
	testify.NoError(t, err, "Load cache should succeed")
	testify.NotNil(t, loadedCache, "Loaded cache should not be nil")
	testify.Equal(t, cache.HasCircularDeps, loadedCache.HasCircularDeps, "HasCircularDeps should match")
	testify.Equal(t, len(cache.CircularDeps), len(loadedCache.CircularDeps), "CircularDeps length should match")
}

func TestLoadCircularDepCache_NonExistentFile(t *testing.T) {
	t.Parallel()

	_, err := loadCircularDepCache("/nonexistent/path/cache.json")
	testify.Error(t, err, "Load should fail for non-existent file")
	testify.Contains(t, err.Error(), "failed to read cache file", "Error should mention read failure")
}

func TestLoadCircularDepCache_InvalidJSON(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "invalid-cache.json")

	// Write invalid JSON
	err := cryptoutilFiles.WriteFile(cacheFile, []byte("invalid json content"), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Write should succeed")

	_, err = loadCircularDepCache(cacheFile)
	testify.Error(t, err, "Load should fail for invalid JSON")
	testify.Contains(t, err.Error(), "failed to unmarshal cache JSON", "Error should mention unmarshal failure")
}

func TestSaveCircularDepCache_WithCircularDeps(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, ".cicd", "cache.json")

	cache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    time.Now().UTC(),
		HasCircularDeps: true,
		CircularDeps:    []string{"cryptoutil/pkg/a -> cryptoutil/pkg/b -> cryptoutil/pkg/a"},
	}

	err := saveCircularDepCache(cacheFile, cache)
	testify.NoError(t, err, "Save should succeed")

	// Verify cache can be loaded and contains correct data
	loadedCache, err := loadCircularDepCache(cacheFile)
	testify.NoError(t, err, "Load should succeed")
	testify.True(t, loadedCache.HasCircularDeps, "Should have circular deps flag set")
	testify.Len(t, loadedCache.CircularDeps, 1, "Should have one circular dep entry")
	testify.Contains(t, loadedCache.CircularDeps[0], "cryptoutil/pkg/a", "Should contain cycle info")
}

func TestSaveCircularDepCache_DirectoryCreation(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "deep", "nested", "path", "cache.json")

	cache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    time.Now().UTC(),
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	err := saveCircularDepCache(cacheFile, cache)
	testify.NoError(t, err, "Save should succeed and create directories")

	// Verify file exists
	testify.FileExists(t, cacheFile, "Cache file should exist")

	// Verify directory structure was created
	testify.DirExists(t, filepath.Dir(cacheFile), "Cache directory should exist")
}

func TestCacheJSONFormat(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "cache.json")

	cache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
		GoModModTime:    time.Date(2025, 1, 14, 10, 0, 0, 0, time.UTC),
		HasCircularDeps: true,
		CircularDeps:    []string{"cycle1", "cycle2"},
	}

	err := saveCircularDepCache(cacheFile, cache)
	testify.NoError(t, err, "Save should succeed")

	// Read raw JSON and verify format
	content, err := os.ReadFile(cacheFile)
	testify.NoError(t, err, "Read should succeed")

	// Verify it's valid JSON
	var decoded cryptoutilMagic.CircularDepCache

	err = json.Unmarshal(content, &decoded)
	testify.NoError(t, err, "JSON should be valid")

	// Verify formatting (indented)
	testify.Contains(t, string(content), "  ", "JSON should be indented")
	testify.Contains(t, string(content), "last_check", "JSON should contain last_check field")
	testify.Contains(t, string(content), "has_circular_deps", "JSON should contain has_circular_deps field")
}

func TestCachePermissions(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, ".cicd", "cache.json")

	cache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    time.Now().UTC(),
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	err := saveCircularDepCache(cacheFile, cache)
	testify.NoError(t, err, "Save should succeed")

	// Check file exists (permission check is platform-specific, so we skip it)
	testify.FileExists(t, cacheFile, "Cache file should exist")
}
