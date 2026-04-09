// Copyright (c) 2025 Justin Cranford

package circular_deps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCheck_NoGoMod(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Check() reads go.mod from ".". Package test directory has no go.mod,
	// so Check() returns an error "failed to read go.mod".
	err := Check(logger)
	require.ErrorContains(t, err, "failed to read go.mod")
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_FreshCheckFindsCircularDeps(t *testing.T) {
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), cryptoutilSharedMagic.CacheFilePermissions))

	_ = os.Remove(cryptoutilSharedMagic.CircularDepCacheFileName)

	stubGoListFn := func() ([]byte, error) {
		return []byte("{\"ImportPath\": \"testmod/a\", \"Imports\": [\"testmod/b\"]}\n{\"ImportPath\": \"testmod/b\", \"Imports\": [\"testmod/a\"]}"), nil
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test-circular-fresh")

	err = check(logger, stubGoListFn)
	require.ErrorContains(t, err, "circular dependency")

	_ = os.Remove(cryptoutilSharedMagic.CircularDepCacheFileName)
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_Integration(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(projectRoot))

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	cacheFile := cryptoutilSharedMagic.CircularDepCacheFileName
	_ = os.Remove(cacheFile)

	logger := cryptoutilCmdCicdCommon.NewLogger("test-circulardeps")

	// First call: cache miss - performs actual check.
	err = Check(logger)
	require.NoError(t, err, "Project should have no circular dependencies")

	// Second call: cache hit - uses cached result.
	err = Check(logger)
	require.NoError(t, err, "Cached result should indicate no circular dependencies")

	_ = os.Remove(cacheFile)
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_CacheStates(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping - cannot find project root (no go.mod)")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(projectRoot))

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	goModStat, err := os.Stat("go.mod")
	require.NoError(t, err)

	tests := []struct {
		name    string
		cache   cryptoutilSharedMagic.CircularDepCache
		wantErr string
	}{
		{
			name: "cached with circular deps",
			cache: cryptoutilSharedMagic.CircularDepCache{
				LastCheck:       time.Now().UTC(),
				GoModModTime:    goModStat.ModTime(),
				HasCircularDeps: true,
				CircularDeps:    []string{"pkg/a -> pkg/b -> pkg/a"},
			},
			wantErr: "circular dependencies detected (cached)",
		},
		{
			name: "expired cache triggers fresh check",
			cache: cryptoutilSharedMagic.CircularDepCache{
				LastCheck:       time.Now().UTC().Add(-cryptoutilSharedMagic.CircularDepCacheValidDuration - time.Hour),
				GoModModTime:    goModStat.ModTime(),
				HasCircularDeps: false,
				CircularDeps:    []string{},
			},
		},
		{
			name: "go.mod changed triggers fresh check",
			cache: cryptoutilSharedMagic.CircularDepCache{
				LastCheck:       time.Now().UTC(),
				GoModModTime:    time.Now().UTC().Add(-time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
				HasCircularDeps: false,
				CircularDeps:    []string{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cacheFile := cryptoutilSharedMagic.CircularDepCacheFileName

			require.NoError(t, SaveCircularDepCache(cacheFile, tc.cache))

			defer func() { _ = os.Remove(cacheFile) }()

			logger := cryptoutilCmdCicdCommon.NewLogger("test-circulardeps")

			err := Check(logger)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_GoListError(t *testing.T) {
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	goModContent := "module testmodule\n\ngo 1.21\n\nrequire nonexistent.example.com/fake/module v999.999.999\n"
	require.NoError(t, os.WriteFile("go.mod", []byte(goModContent), cryptoutilSharedMagic.CacheFilePermissions))

	goFileContent := "package main\n\nimport \"nonexistent.example.com/fake/module\"\n\nfunc main() { module.Do() }\n"
	require.NoError(t, os.WriteFile("main.go", []byte(goFileContent), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test-circulardeps")

	err = Check(logger)
	require.ErrorContains(t, err, "failed to run go list")
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_FreshCheckWithActualCircularDeps(t *testing.T) {
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module testcircular\n\ngo 1.21\n"), cryptoutilSharedMagic.CacheFilePermissions))

	require.NoError(t, os.MkdirAll("internal/a", cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile("internal/a/a.go", []byte("package a\n\nimport \"testcircular/internal/b\"\n\nfunc A() { b.B() }\n"), cryptoutilSharedMagic.CacheFilePermissions))

	require.NoError(t, os.MkdirAll("internal/b", cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile("internal/b/b.go", []byte("package b\n\nimport \"testcircular/internal/a\"\n\nfunc B() { a.A() }\n"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test-circulardeps")

	// go list may fail with import cycle error before CheckDependencies runs.
	err = Check(logger)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to run go list") ||
		strings.Contains(err.Error(), "circular"),
		"Expected error about go list failure or circular deps, got: %v", err)
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_SaveCacheError(t *testing.T) {
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), cryptoutilSharedMagic.CacheFilePermissions))

	require.NoError(t, os.MkdirAll("internal/pkg", cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile("internal/pkg/hello.go", []byte("package pkg\n\nfunc Hello() string { return \"hello\" }\n"), cryptoutilSharedMagic.CacheFilePermissions))

	// Create .cicd as a regular FILE (not dir) to make SaveCircularDepCache fail.
	require.NoError(t, os.WriteFile(cryptoutilSharedMagic.CICDOutputDir, []byte("blocker"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Check should succeed (no circular deps) but log a warning about failed cache save.
	err = Check(logger)
	require.NoError(t, err)
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}

		dir = parent
	}
}
