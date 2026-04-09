// Copyright (c) 2025 Justin Cranford

package circular_deps

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCheckDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		goListOutput string
		wantErr      string
	}{
		{name: "empty output", goListOutput: ""},
		{name: "single package no imports", goListOutput: `{"ImportPath": "example.com/pkg/a", "Imports": []}`},
		{
			name: "no cycles a-b-c",
			goListOutput: `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": ["example.com/pkg/c"]}
{"ImportPath": "example.com/pkg/c", "Imports": []}`,
		},
		{
			name: "external deps ignored",
			goListOutput: `{"ImportPath": "example.com/pkg/a", "Imports": ["github.com/external/pkg", "example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": ["fmt", "github.com/another/pkg"]}`,
		},
		{
			name: "multiple disconnected graphs",
			goListOutput: `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": []}
{"ImportPath": "example.com/pkg/c", "Imports": ["example.com/pkg/d"]}
{"ImportPath": "example.com/pkg/d", "Imports": []}`,
		},
		{
			name: "mixed module prefixes",
			goListOutput: `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": []}
{"ImportPath": "other.org/different/pkg", "Imports": ["other.org/different/another"]}`,
		},
		{
			name: "cycle a-b-c-a",
			goListOutput: `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": ["example.com/pkg/c"]}
{"ImportPath": "example.com/pkg/c", "Imports": ["example.com/pkg/a"]}`,
			wantErr: "circular dependency",
		},
		{
			name: "complex cycle via d",
			goListOutput: `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b", "example.com/pkg/c"]}
{"ImportPath": "example.com/pkg/b", "Imports": ["example.com/pkg/d"]}
{"ImportPath": "example.com/pkg/c", "Imports": ["example.com/pkg/d"]}
{"ImportPath": "example.com/pkg/d", "Imports": ["example.com/pkg/a"]}`,
			wantErr: "circular dependency",
		},
		{
			name:         "self reference",
			goListOutput: `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/a"]}`,
			wantErr:      "circular dependency",
		},
		{
			name: "malformed JSON",
			goListOutput: `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
invalid json line
{"ImportPath": "example.com/pkg/b", "Imports": []}`,
			wantErr: "failed to decode package info",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := CheckDependencies(tc.goListOutput)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGetModulePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		packages map[string][]string
		want     string
	}{
		{name: "empty packages", packages: map[string][]string{}, want: ""},
		{name: "single package", packages: map[string][]string{"example.com/pkg/a": {}}, want: "example.com"},
		{name: "multiple packages", packages: map[string][]string{"example.com/pkg/a": {}, "example.com/pkg/b": {}, "example.com/pkg/c": {}}, want: "example.com"},
		{name: "different prefixes", packages: map[string][]string{"github.com/user/repo/pkg/a": {}, "github.com/user/repo/pkg/b": {}}, want: "github.com"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := GetModulePath(tc.packages)
			require.Equal(t, tc.want, result)
		})
	}
}

func TestLoadCircularDepCache_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupPath func(t *testing.T) string
		wantErr   string
	}{
		{
			name:      "file not found",
			setupPath: func(_ *testing.T) string { return "nonexistent-file-12345.json" },
			wantErr:   "failed to read cache file",
		},
		{
			name: "invalid JSON",
			setupPath: func(t *testing.T) string {
				t.Helper()

				tmpFile := filepath.Join(t.TempDir(), "invalid-cache.json")
				require.NoError(t, os.WriteFile(tmpFile, []byte("{invalid json}"), cryptoutilSharedMagic.CacheFilePermissions))

				return tmpFile
			},
			wantErr: "failed to unmarshal cache JSON",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cache, err := LoadCircularDepCache(tc.setupPath(t))
			require.Error(t, err)
			require.Nil(t, cache)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestSaveLoadCircularDepCache_RoundTrip(t *testing.T) {
	t.Parallel()

	tmpFile := filepath.Join(t.TempDir(), "cache.json")

	original := cryptoutilSharedMagic.CircularDepCache{
		LastCheck:       time.Now().UTC().Truncate(time.Second),
		GoModModTime:    time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Second),
		HasCircularDeps: true,
		CircularDeps:    []string{"pkg/a -> pkg/b -> pkg/a"},
	}

	err := SaveCircularDepCache(tmpFile, original)
	require.NoError(t, err)

	_, err = os.Stat(tmpFile)
	require.NoError(t, err)

	loaded, err := LoadCircularDepCache(tmpFile)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	require.Equal(t, original.HasCircularDeps, loaded.HasCircularDeps)
	require.Equal(t, original.CircularDeps, loaded.CircularDeps)
	require.True(t, original.LastCheck.Equal(loaded.LastCheck))
	require.True(t, original.GoModModTime.Equal(loaded.GoModModTime))
}

func TestSaveCircularDepCache_CreateDirectory(t *testing.T) {
	t.Parallel()

	cacheFile := filepath.Join(t.TempDir(), "subdir", "cache.json")

	cache := cryptoutilSharedMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    time.Now().UTC(),
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	err := SaveCircularDepCache(cacheFile, cache)
	require.NoError(t, err)

	_, err = os.Stat(cacheFile)
	require.NoError(t, err)
}

func TestSaveCircularDepCache_DirectoryCreationError(t *testing.T) {
	t.Parallel()

	cache := cryptoutilSharedMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    time.Now().UTC(),
		HasCircularDeps: false,
		CircularDeps:    nil,
	}

	tempFile, err := os.CreateTemp("", "notadir")
	require.NoError(t, err)

	defer func() { _ = os.Remove(tempFile.Name()) }()

	_ = tempFile.Close()

	invalidPath := tempFile.Name() + "/cache.json"
	err = SaveCircularDepCache(invalidPath, cache)
	require.ErrorContains(t, err, "failed to create output directory")
}

// Sequential: test uses filesystem permissions (os.Chmod on temp file).
func TestSaveCircularDepCache_WriteFileError(t *testing.T) {
	cache := cryptoutilSharedMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    time.Now().UTC(),
		HasCircularDeps: false,
		CircularDeps:    nil,
	}

	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "subdir")
	require.NoError(t, os.MkdirAll(cacheDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	cacheFile := filepath.Join(cacheDir, "cache.json")
	require.NoError(t, os.WriteFile(cacheFile, []byte("existing"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.Chmod(cacheFile, 0o000))

	defer func() { _ = os.Chmod(cacheFile, cryptoutilSharedMagic.CacheFilePermissions) }()

	err := SaveCircularDepCache(cacheFile, cache)
	require.ErrorContains(t, err, "failed to write cache file")
}
