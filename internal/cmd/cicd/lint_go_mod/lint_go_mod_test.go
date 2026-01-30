// Copyright (c) 2025 Justin Cranford

package lint_go_mod

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestLint_NoGoMod(t *testing.T) {
	t.Parallel()

	// This test would fail if run in a directory without go.mod.
	// Since we're in a Go project, it should work.
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Note: This will actually check dependencies, which may pass or fail.
	// We're testing that the function doesn't panic.
	_ = Lint(logger)
}

func TestCheckDependencyUpdates_Empty(t *testing.T) {
	t.Parallel()

	outdated, err := checkDependencyUpdates("", map[string]bool{})

	require.NoError(t, err)
	require.Empty(t, outdated)
}

func TestCheckDependencyUpdates_NoUpdates(t *testing.T) {
	t.Parallel()

	goListOutput := `example.com/mymodule
github.com/pkg/errors v0.9.1
github.com/stretchr/testify v1.8.0`

	directDeps := map[string]bool{
		"github.com/pkg/errors":       true,
		"github.com/stretchr/testify": true,
	}

	outdated, err := checkDependencyUpdates(goListOutput, directDeps)

	require.NoError(t, err)
	require.Empty(t, outdated)
}

func TestCheckDependencyUpdates_WithUpdates(t *testing.T) {
	t.Parallel()

	goListOutput := `example.com/mymodule
github.com/pkg/errors v0.9.1 [v0.9.2]
github.com/stretchr/testify v1.8.0`

	directDeps := map[string]bool{
		"github.com/pkg/errors":       true,
		"github.com/stretchr/testify": true,
	}

	outdated, err := checkDependencyUpdates(goListOutput, directDeps)

	require.NoError(t, err)
	require.Len(t, outdated, 1)
	require.Contains(t, outdated[0], "github.com/pkg/errors")
}

func TestCheckDependencyUpdates_IndirectNotIncluded(t *testing.T) {
	t.Parallel()

	goListOutput := `example.com/mymodule
github.com/pkg/errors v0.9.1 [v0.9.2]
github.com/indirect/dep v1.0.0 [v1.1.0]`

	// Only direct deps are included.
	directDeps := map[string]bool{
		"github.com/pkg/errors": true,
	}

	outdated, err := checkDependencyUpdates(goListOutput, directDeps)

	require.NoError(t, err)
	require.Len(t, outdated, 1)
	require.Contains(t, outdated[0], "github.com/pkg/errors")
}

func TestGetDirectDependencies(t *testing.T) {
	t.Parallel()

	goModContent := []byte(`module example.com/mymodule

go 1.25.5

require (
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.0
	github.com/indirect/dep v1.0.0 // indirect
)
`)

	directDeps, err := getDirectDependencies(goModContent)

	require.NoError(t, err)
	require.True(t, directDeps["github.com/pkg/errors"])
	require.True(t, directDeps["github.com/stretchr/testify"])
	require.False(t, directDeps["github.com/indirect/dep"])
}

func TestGetDirectDependencies_SingleLineRequire(t *testing.T) {
	t.Parallel()

	goModContent := []byte(`module example.com/mymodule

go 1.25.5

require github.com/pkg/errors v0.9.1
`)

	directDeps, err := getDirectDependencies(goModContent)

	require.NoError(t, err)
	require.True(t, directDeps["github.com/pkg/errors"])
}

func TestGetDirectDependencies_Empty(t *testing.T) {
	t.Parallel()

	goModContent := []byte(`module example.com/mymodule

go 1.25.5
`)

	directDeps, err := getDirectDependencies(goModContent)

	require.NoError(t, err)
	require.Empty(t, directDeps)
}

func TestCheckAndUseDepCache_CacheNotFound(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	goModFile := filepath.Join(tmpDir, "go.mod")
	err := os.WriteFile(goModFile, []byte("module test\ngo 1.25.5\n"), 0o600)
	require.NoError(t, err)

	goSumFile := filepath.Join(tmpDir, "go.sum")
	err = os.WriteFile(goSumFile, []byte(""), 0o600)
	require.NoError(t, err)

	goModStat, err := os.Stat(goModFile)
	require.NoError(t, err)

	goSumStat, err := os.Stat(goSumFile)
	require.NoError(t, err)

	used, state, err := checkAndUseDepCache("nonexistent_cache.json", "direct", goModStat, goSumStat, logger)
	require.NoError(t, err)
	require.False(t, used)
	require.Equal(t, "cache not found", state)
}

func TestLoadDepCache_FileNotFound(t *testing.T) {
	t.Parallel()

	cache, err := loadDepCache("nonexistent_cache.json")
	require.Error(t, err)
	require.Nil(t, cache)
	require.Contains(t, err.Error(), "failed to read cache file")
}

func TestLoadDepCache_InvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "invalid_cache.json")
	err := os.WriteFile(cacheFile, []byte("invalid json"), 0o600)
	require.NoError(t, err)

	cache, err := loadDepCache(cacheFile)
	require.Error(t, err)
	require.Nil(t, cache)
	require.Contains(t, err.Error(), "failed to unmarshal cache JSON")
}

func TestSaveDepCache_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "test_cache.json")

	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC(),
		GoModModTime: time.Now().UTC(),
		GoSumModTime: time.Now().UTC(),
		OutdatedDeps: []string{"example.com/dep v1.0.0 [v1.1.0]"},
		Mode:         "direct",
	}

	err := saveDepCache(cacheFile, cache)
	require.NoError(t, err)

	content, err := os.ReadFile(cacheFile)
	require.NoError(t, err)
	require.Contains(t, string(content), "example.com/dep")
	require.Contains(t, string(content), "direct")
}

func TestCheckDependencyUpdates_MalformedLine(t *testing.T) {
	t.Parallel()

	goListOutput := `example.com/mymodule
malformed_line_without_space [v0.9.2]
github.com/pkg/errors v0.9.1`

	directDeps := map[string]bool{
		"github.com/pkg/errors": true,
	}

	outdated, err := checkDependencyUpdates(goListOutput, directDeps)
	require.NoError(t, err)
	require.Empty(t, outdated, "Should skip lines with updates but no matching direct deps")
}

func TestCheckDependencyUpdates_IncompleteUpdate(t *testing.T) {
	t.Parallel()

	goListOutput := `example.com/mymodule
github.com/pkg/errors v0.9.1 [v0.9.2
github.com/stretchr/testify v1.8.0 [incomplete`

	directDeps := map[string]bool{
		"github.com/pkg/errors":       true,
		"github.com/stretchr/testify": true,
	}

	outdated, err := checkDependencyUpdates(goListOutput, directDeps)
	require.NoError(t, err)
	require.Empty(t, outdated, "Should skip lines without closing bracket")
}

func TestGetDirectDependencies_WithComments(t *testing.T) {
	t.Parallel()

	goModContent := []byte(`module example.com/mymodule

go 1.25.5

require (
	// This is a comment
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.0 // needed for tests
)
`)

	directDeps, err := getDirectDependencies(goModContent)
	require.NoError(t, err)
	require.Len(t, directDeps, 2)
	require.True(t, directDeps["github.com/pkg/errors"])
	require.True(t, directDeps["github.com/stretchr/testify"])
}

func TestGetDirectDependencies_MixedRequireFormats(t *testing.T) {
	t.Parallel()

	goModContent := []byte(`module example.com/mymodule

go 1.25.5

require github.com/pkg/errors v0.9.1

require (
	github.com/stretchr/testify v1.8.0
	github.com/indirect/dep v1.0.0 // indirect
)
`)

	directDeps, err := getDirectDependencies(goModContent)
	require.NoError(t, err)
	require.Len(t, directDeps, 2)
	require.True(t, directDeps["github.com/pkg/errors"])
	require.True(t, directDeps["github.com/stretchr/testify"])
	require.False(t, directDeps["github.com/indirect/dep"])
}

func TestCheckDependencyUpdates_MultipleUpdates(t *testing.T) {
	t.Parallel()

	goListOutput := `example.com/mymodule
github.com/pkg/errors v0.9.1 [v0.9.2]
github.com/stretchr/testify v1.8.0 [v1.9.0]
github.com/indirect/dep v1.0.0 [v1.1.0]`

	directDeps := map[string]bool{
		"github.com/pkg/errors":       true,
		"github.com/stretchr/testify": true,
	}

	outdated, err := checkDependencyUpdates(goListOutput, directDeps)
	require.NoError(t, err)
	require.Len(t, outdated, 2)
	require.Contains(t, outdated[0], "github.com/pkg/errors")
	require.Contains(t, outdated[1], "github.com/stretchr/testify")
}

func TestCheckDependencyUpdates_NoDirectDeps(t *testing.T) {
	t.Parallel()

	goListOutput := `example.com/mymodule
github.com/pkg/errors v0.9.1 [v0.9.2]`

	directDeps := map[string]bool{}

	outdated, err := checkDependencyUpdates(goListOutput, directDeps)
	require.NoError(t, err)
	require.Empty(t, outdated, "Should return empty when no direct dependencies specified")
}

func TestCheckAndUseDepCache_ModeMismatch(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	goModFile := filepath.Join(tmpDir, "go.mod")
	err := os.WriteFile(goModFile, []byte("module test\ngo 1.25.5\n"), 0o600)
	require.NoError(t, err)

	goSumFile := filepath.Join(tmpDir, "go.sum")
	err = os.WriteFile(goSumFile, []byte(""), 0o600)
	require.NoError(t, err)

	goModStat, err := os.Stat(goModFile)
	require.NoError(t, err)

	goSumStat, err := os.Stat(goSumFile)
	require.NoError(t, err)

	cacheFile := filepath.Join(tmpDir, "cache.json")
	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC(),
		GoModModTime: goModStat.ModTime(),
		GoSumModTime: goSumStat.ModTime(),
		OutdatedDeps: []string{},
		Mode:         "indirect",
	}
	err = saveDepCache(cacheFile, cache)
	require.NoError(t, err)

	used, state, err := checkAndUseDepCache(cacheFile, "direct", goModStat, goSumStat, logger)
	require.NoError(t, err)
	require.False(t, used)
	require.Equal(t, "cache mode mismatch", state)
}

func TestCheckAndUseDepCache_Expired(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	goModFile := filepath.Join(tmpDir, "go.mod")
	err := os.WriteFile(goModFile, []byte("module test\ngo 1.25.5\n"), 0o600)
	require.NoError(t, err)

	goSumFile := filepath.Join(tmpDir, "go.sum")
	err = os.WriteFile(goSumFile, []byte(""), 0o600)
	require.NoError(t, err)

	goModStat, err := os.Stat(goModFile)
	require.NoError(t, err)

	goSumStat, err := os.Stat(goSumFile)
	require.NoError(t, err)

	cacheFile := filepath.Join(tmpDir, "cache.json")
	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC().Add(-2 * time.Hour),
		GoModModTime: goModStat.ModTime(),
		GoSumModTime: goSumStat.ModTime(),
		OutdatedDeps: []string{},
		Mode:         "direct",
	}
	err = saveDepCache(cacheFile, cache)
	require.NoError(t, err)

	used, state, err := checkAndUseDepCache(cacheFile, "direct", goModStat, goSumStat, logger)
	require.NoError(t, err)
	require.False(t, used)
	require.Equal(t, "cache expired", state)
}

func TestCheckAndUseDepCache_GoModChanged(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	goModFile := filepath.Join(tmpDir, "go.mod")
	err := os.WriteFile(goModFile, []byte("module test\ngo 1.25.5\n"), 0o600)
	require.NoError(t, err)

	goSumFile := filepath.Join(tmpDir, "go.sum")
	err = os.WriteFile(goSumFile, []byte(""), 0o600)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	goModStat, err := os.Stat(goModFile)
	require.NoError(t, err)

	goSumStat, err := os.Stat(goSumFile)
	require.NoError(t, err)

	cacheFile := filepath.Join(tmpDir, "cache.json")
	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC(),
		GoModModTime: goModStat.ModTime().Add(-1 * time.Hour),
		GoSumModTime: goSumStat.ModTime(),
		OutdatedDeps: []string{},
		Mode:         "direct",
	}
	err = saveDepCache(cacheFile, cache)
	require.NoError(t, err)

	used, state, err := checkAndUseDepCache(cacheFile, "direct", goModStat, goSumStat, logger)
	require.NoError(t, err)
	require.False(t, used)
	require.Equal(t, "go.mod or go.sum changed", state)
}

func TestCheckAndUseDepCache_ValidCacheWithOutdated(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	goModFile := filepath.Join(tmpDir, "go.mod")
	err := os.WriteFile(goModFile, []byte("module test\ngo 1.25.5\n"), 0o600)
	require.NoError(t, err)

	goSumFile := filepath.Join(tmpDir, "go.sum")
	err = os.WriteFile(goSumFile, []byte(""), 0o600)
	require.NoError(t, err)

	goModStat, err := os.Stat(goModFile)
	require.NoError(t, err)

	goSumStat, err := os.Stat(goSumFile)
	require.NoError(t, err)

	cacheFile := filepath.Join(tmpDir, "cache.json")
	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC(),
		GoModModTime: goModStat.ModTime(),
		GoSumModTime: goSumStat.ModTime(),
		OutdatedDeps: []string{"example.com/dep v1.0.0 [v1.1.0]"},
		Mode:         "direct",
	}
	err = saveDepCache(cacheFile, cache)
	require.NoError(t, err)

	used, state, err := checkAndUseDepCache(cacheFile, "direct", goModStat, goSumStat, logger)
	require.True(t, used)
	require.Empty(t, state)
	require.Error(t, err)
	require.Contains(t, err.Error(), "outdated dependencies found")
}

func TestCheckAndUseDepCache_ValidCacheClean(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	goModFile := filepath.Join(tmpDir, "go.mod")
	err := os.WriteFile(goModFile, []byte("module test\ngo 1.25.5\n"), 0o600)
	require.NoError(t, err)

	goSumFile := filepath.Join(tmpDir, "go.sum")
	err = os.WriteFile(goSumFile, []byte(""), 0o600)
	require.NoError(t, err)

	goModStat, err := os.Stat(goModFile)
	require.NoError(t, err)

	goSumStat, err := os.Stat(goSumFile)
	require.NoError(t, err)

	cacheFile := filepath.Join(tmpDir, "cache.json")
	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC(),
		GoModModTime: goModStat.ModTime(),
		GoSumModTime: goSumStat.ModTime(),
		OutdatedDeps: []string{},
		Mode:         "direct",
	}
	err = saveDepCache(cacheFile, cache)
	require.NoError(t, err)

	used, state, err := checkAndUseDepCache(cacheFile, "direct", goModStat, goSumStat, logger)
	require.True(t, used)
	require.Empty(t, state)
	require.NoError(t, err)
}

func TestLoadDepCache_ValidCache(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "cache.json")

	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC(),
		GoModModTime: time.Now().UTC(),
		GoSumModTime: time.Now().UTC(),
		OutdatedDeps: []string{"example.com/dep v1.0.0 [v1.1.0]"},
		Mode:         "direct",
	}
	err := saveDepCache(cacheFile, cache)
	require.NoError(t, err)

	loadedCache, err := loadDepCache(cacheFile)
	require.NoError(t, err)
	require.NotNil(t, loadedCache)
	require.Equal(t, "direct", loadedCache.Mode)
	require.Len(t, loadedCache.OutdatedDeps, 1)
}

func TestSaveDepCache_CreateDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "subdir", "cache.json")

	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC(),
		GoModModTime: time.Now().UTC(),
		GoSumModTime: time.Now().UTC(),
		OutdatedDeps: []string{},
		Mode:         "direct",
	}

	err := saveDepCache(cacheFile, cache)
	require.NoError(t, err)

	_, err = os.Stat(cacheFile)
	require.NoError(t, err, "Cache file should be created with parent directory")
}

func TestGetDirectDependencies_EmptyLines(t *testing.T) {
	t.Parallel()

	goModContent := []byte(`module example.com/mymodule

go 1.25.5

require (

	github.com/pkg/errors v0.9.1

	github.com/stretchr/testify v1.8.0

)
`)

	directDeps, err := getDirectDependencies(goModContent)
	require.NoError(t, err)
	require.Len(t, directDeps, 2)
	require.True(t, directDeps["github.com/pkg/errors"])
	require.True(t, directDeps["github.com/stretchr/testify"])
}

func TestCheckDependencyUpdates_OnlyEmptyLines(t *testing.T) {
	t.Parallel()

	goListOutput := "\n\n\n"

	directDeps := map[string]bool{
		"github.com/pkg/errors": true,
	}

	outdated, err := checkDependencyUpdates(goListOutput, directDeps)
	require.NoError(t, err)
	require.Empty(t, outdated)
}

func TestGetDirectDependencies_MultipleSingleLineRequires(t *testing.T) {
	t.Parallel()

	goModContent := []byte(`module example.com/mymodule

go 1.25.5

require github.com/pkg/errors v0.9.1
require github.com/stretchr/testify v1.8.0
require github.com/indirect/dep v1.0.0 // indirect
`)

	directDeps, err := getDirectDependencies(goModContent)
	require.NoError(t, err)
	require.Len(t, directDeps, 2)
	require.True(t, directDeps["github.com/pkg/errors"])
	require.True(t, directDeps["github.com/stretchr/testify"])
	require.False(t, directDeps["github.com/indirect/dep"])
}

func TestCheckDependencyUpdates_LinesWithoutSpace(t *testing.T) {
	t.Parallel()

	goListOutput := `example.com/mymodule
no-space-line[v0.9.2]
github.com/pkg/errors v0.9.1 [v0.9.2]`

	directDeps := map[string]bool{
		"github.com/pkg/errors": true,
	}

	outdated, err := checkDependencyUpdates(goListOutput, directDeps)
	require.NoError(t, err)
	require.Len(t, outdated, 1)
	require.Contains(t, outdated[0], "github.com/pkg/errors")
}

func TestCheckDependencyUpdates_UpdateMarkersWithoutClosingBracket(t *testing.T) {
	t.Parallel()

	goListOutput := `example.com/mymodule
github.com/pkg/errors v0.9.1 [v0.9.2
github.com/stretchr/testify v1.8.0 [v1.9.0]`

	directDeps := map[string]bool{
		"github.com/pkg/errors":       true,
		"github.com/stretchr/testify": true,
	}

	outdated, err := checkDependencyUpdates(goListOutput, directDeps)
	require.NoError(t, err)
	require.Len(t, outdated, 1, "Should only include lines with complete update markers")
	require.Contains(t, outdated[0], "github.com/stretchr/testify")
}

func TestGetDirectDependencies_NoRequireBlock(t *testing.T) {
	t.Parallel()

	goModContent := []byte(`module example.com/mymodule

go 1.25.5

replace example.com/old => example.com/new v1.0.0
`)

	directDeps, err := getDirectDependencies(goModContent)
	require.NoError(t, err)
	require.Empty(t, directDeps)
}

func TestCheckDependencyUpdates_LargeNumberOfDeps(t *testing.T) {
	t.Parallel()

	goListOutput := `example.com/mymodule
github.com/dep1/pkg v1.0.0 [v1.1.0]
github.com/dep2/pkg v1.0.0 [v1.1.0]
github.com/dep3/pkg v1.0.0 [v1.1.0]
github.com/dep4/pkg v1.0.0 [v1.1.0]
github.com/dep5/pkg v1.0.0 [v1.1.0]
github.com/indirect/dep v1.0.0 [v1.1.0]`

	directDeps := map[string]bool{
		"github.com/dep1/pkg": true,
		"github.com/dep2/pkg": true,
		"github.com/dep3/pkg": true,
		"github.com/dep4/pkg": true,
		"github.com/dep5/pkg": true,
	}

	outdated, err := checkDependencyUpdates(goListOutput, directDeps)
	require.NoError(t, err)
	require.Len(t, outdated, 5, "Should find all direct outdated dependencies")
}

// TestCheckOutdatedDeps_NoGoMod tests checkOutdatedDeps when go.mod is missing.
func TestCheckOutdatedDeps_NoGoMod(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = checkOutdatedDeps(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read go.mod")
}

// TestCheckOutdatedDeps_NoGoSum tests checkOutdatedDeps when go.sum is missing.
func TestCheckOutdatedDeps_NoGoSum(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod only (no go.sum).
	err = os.WriteFile("go.mod", []byte("module test\ngo 1.25.5\n"), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = checkOutdatedDeps(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read go.sum")
}

// TestCheckOutdatedDeps_CacheUsed tests checkOutdatedDeps when valid cache exists.
func TestCheckOutdatedDeps_CacheUsed(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod and go.sum.
	err = os.WriteFile("go.mod", []byte("module test\ngo 1.25.5\n"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile("go.sum", []byte(""), 0o600)
	require.NoError(t, err)

	// Get file stats for cache.
	goModStat, err := os.Stat("go.mod")
	require.NoError(t, err)
	goSumStat, err := os.Stat("go.sum")
	require.NoError(t, err)

	// Create a valid cache file with no outdated deps.
	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC(),
		GoModModTime: goModStat.ModTime(),
		GoSumModTime: goSumStat.ModTime(),
		OutdatedDeps: []string{},
		Mode:         cryptoutilSharedMagic.ModeNameDirect,
	}
	err = saveDepCache(cryptoutilSharedMagic.DepCacheFileName, cache)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = checkOutdatedDeps(logger)
	require.NoError(t, err)
}

// TestCheckOutdatedDeps_CacheWithError tests checkOutdatedDeps when cache has outdated deps.
func TestCheckOutdatedDeps_CacheWithError(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod and go.sum.
	err = os.WriteFile("go.mod", []byte("module test\ngo 1.25.5\n"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile("go.sum", []byte(""), 0o600)
	require.NoError(t, err)

	// Get file stats for cache.
	goModStat, err := os.Stat("go.mod")
	require.NoError(t, err)
	goSumStat, err := os.Stat("go.sum")
	require.NoError(t, err)

	// Create a valid cache file WITH outdated deps.
	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC(),
		GoModModTime: goModStat.ModTime(),
		GoSumModTime: goSumStat.ModTime(),
		OutdatedDeps: []string{"example.com/dep v1.0.0 [v1.1.0]"},
		Mode:         cryptoutilSharedMagic.ModeNameDirect,
	}
	err = saveDepCache(cryptoutilSharedMagic.DepCacheFileName, cache)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = checkOutdatedDeps(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cached dependency check failed")
}

// TestCheckOutdatedDeps_GoListError tests checkOutdatedDeps when go list command fails.
func TestCheckOutdatedDeps_GoListError(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a malformed go.mod file that will make go list fail.
	// Using invalid syntax to force a parsing error.
	err = os.WriteFile("go.mod", []byte("invalid go.mod content without module directive\n"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile("go.sum", []byte(""), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = checkOutdatedDeps(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to check dependencies")
}

// TestCheckOutdatedDeps_NoOutdatedDeps tests checkOutdatedDeps with up-to-date deps (fresh check).
func TestCheckOutdatedDeps_NoOutdatedDeps(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a valid go module that go list can process.
	// Using "example.com/test" which won't have any real dependencies.
	err = os.WriteFile("go.mod", []byte("module example.com/test\ngo 1.25.5\n"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile("go.sum", []byte(""), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = checkOutdatedDeps(logger)
	// This should succeed because there are no dependencies.
	require.NoError(t, err)
}

// TestSaveDepCache_WriteError tests saveDepCache when directory creation fails.
func TestSaveDepCache_WriteError(t *testing.T) {
	t.Parallel()

	// Create a read-only directory structure.
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	err := os.Mkdir(readOnlyDir, 0o500)
	require.NoError(t, err)

	// Try to write to a nested path inside the read-only directory.
	cacheFile := filepath.Join(readOnlyDir, "nested", "cache.json")

	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC(),
		GoModModTime: time.Now().UTC(),
		GoSumModTime: time.Now().UTC(),
		OutdatedDeps: []string{},
		Mode:         "direct",
	}

	err = saveDepCache(cacheFile, cache)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create output directory")
}

// TestLint_WithLinterError tests the Lint function when a linter returns an error.
func TestLint_WithLinterError(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// No go.mod file, so the linter will fail.
	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Lint(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "lint-go-mod failed with 1 errors")
}

// TestLint_Success tests the Lint function when all linters pass.
func TestLint_Success(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a valid go module setup.
	err = os.WriteFile("go.mod", []byte("module example.com/test\ngo 1.25.5\n"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile("go.sum", []byte(""), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Lint(logger)
	require.NoError(t, err)
}

// TestCheckOutdatedDeps_WithOutdatedDeps tests checkOutdatedDeps finding outdated deps (fresh check).
func TestCheckOutdatedDeps_WithOutdatedDeps(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a go module that has an outdated dependency.
	// We need a real dependency that might have updates.
	goModContent := `module example.com/test

go 1.25.5

require github.com/pkg/errors v0.8.0
`
	err = os.WriteFile("go.mod", []byte(goModContent), 0o600)
	require.NoError(t, err)
	err = os.WriteFile("go.sum", []byte("github.com/pkg/errors v0.8.0 h1:WdK/asTD0HN+q6hsWO3/vpuAkAr+tw6aNJNDFFf0+qw=\ngithub.com/pkg/errors v0.8.0/go.mod h1:bwawxfHBFNV+L2hUp1rHADufV3IMtnDRdf1r5NINEl0=\n"), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = checkOutdatedDeps(logger)
	// This should fail because pkg/errors v0.8.0 has updates available.
	require.Error(t, err)
	require.Contains(t, err.Error(), "outdated dependencies found")
}

// TestCheckOutdatedDeps_SaveCacheError tests warning when cache save fails.
func TestCheckOutdatedDeps_SaveCacheError(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a valid go module.
	err = os.WriteFile("go.mod", []byte("module example.com/test\ngo 1.25.5\n"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile("go.sum", []byte(""), 0o600)
	require.NoError(t, err)

	// Create a file at the cache directory location to prevent directory creation.
	cacheDir := filepath.Dir(cryptoutilSharedMagic.DepCacheFileName)
	err = os.WriteFile(cacheDir, []byte("blocking file"), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	// This should succeed but with a warning about cache save failure.
	err = checkOutdatedDeps(logger)
	// The check itself should still pass (no outdated deps).
	require.NoError(t, err)
}

// TestSaveDepCache_WriteFileError tests saveDepCache when file write fails.
func TestSaveDepCache_WriteFileError(t *testing.T) {
	t.Parallel()

	// Create a directory at the cache file location to make write fail.
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "cache.json")

	// Create a directory with the same name as the intended file.
	err := os.Mkdir(cacheFile, 0o755)
	require.NoError(t, err)

	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC(),
		GoModModTime: time.Now().UTC(),
		GoSumModTime: time.Now().UTC(),
		OutdatedDeps: []string{},
		Mode:         "direct",
	}

	err = saveDepCache(cacheFile, cache)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write cache file")
}
