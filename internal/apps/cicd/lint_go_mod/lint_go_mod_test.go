// Copyright (c) 2025 Justin Cranford

package lint_go_mod

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
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
