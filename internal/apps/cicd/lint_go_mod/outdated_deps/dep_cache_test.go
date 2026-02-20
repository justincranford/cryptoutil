// Copyright (c) 2025 Justin Cranford

package outdated_deps

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)
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
