// Copyright (c) 2025 Justin Cranford

package fitness_registry_completeness

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------
// Test helpers
// -----------------------------------------------------------------------

// buildRegistryRoot creates a temporary root directory with the fitness sub-linter
// YAML manifest and optional subdirectories under lint_fitness/.
func buildRegistryRoot(t *testing.T, yamlContent string, subdirs []string) string {
	t.Helper()

	rootDir := t.TempDir()

	// Create directory for the YAML file.
	yamlDir := filepath.Dir(filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDFitnessRegistryFile)))
	require.NoError(t, os.MkdirAll(yamlDir, 0o700))
	require.NoError(t, os.WriteFile(
		filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDFitnessRegistryFile)),
		[]byte(yamlContent),
		cryptoutilSharedMagic.FilePermissionsDefault,
	))

	// Create fitness dir and subdirs.
	fitnessDir := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDLintFitnessDir))
	require.NoError(t, os.MkdirAll(fitnessDir, 0o700))

	for _, sd := range subdirs {
		require.NoError(t, os.MkdirAll(filepath.Join(fitnessDir, sd), 0o700))
	}

	return rootDir
}

// minimalRegistryYAML returns a valid registry YAML with two sub-linters.
func minimalRegistryYAML() string {
	return `sub_linters:
  - name: cgo-free-sqlite
    directory: cgo_free_sqlite
    description: Ensures CGO-free SQLite.
    category: architecture
  - name: crypto-rand
    directory: crypto_rand
    description: Detects math/rand vs crypto/rand.
    category: security
`
}

// -----------------------------------------------------------------------
// LoadFitnessRegistry
// -----------------------------------------------------------------------

func TestLoadFitnessRegistry_HappyPath(t *testing.T) {
	t.Parallel()

	rootDir := buildRegistryRoot(t, minimalRegistryYAML(), nil)

	reg, err := LoadFitnessRegistry(rootDir, os.ReadFile)

	require.NoError(t, err)
	require.Len(t, reg.SubLinters, 2)
	require.Equal(t, "cgo-free-sqlite", reg.SubLinters[0].Name)
	require.Equal(t, "cgo_free_sqlite", reg.SubLinters[0].Directory)
	require.Equal(t, "architecture", reg.SubLinters[0].Category)
}

func TestLoadFitnessRegistry_FileNotFound(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	reg, err := LoadFitnessRegistry(rootDir, os.ReadFile)

	require.Error(t, err)
	require.Nil(t, reg)
	require.Contains(t, err.Error(), "failed to read")
}

func TestLoadFitnessRegistry_InvalidYAML(t *testing.T) {
	t.Parallel()

	rootDir := buildRegistryRoot(t, "!!! not valid: yaml: [", nil)

	reg, err := LoadFitnessRegistry(rootDir, os.ReadFile)

	require.Error(t, err)
	require.Nil(t, reg)
	require.Contains(t, err.Error(), "failed to parse")
}

func TestLoadFitnessRegistry_EmptyRegistry(t *testing.T) {
	t.Parallel()

	rootDir := buildRegistryRoot(t, "sub_linters: []\n", nil)

	reg, err := LoadFitnessRegistry(rootDir, os.ReadFile)

	require.NoError(t, err)
	require.Empty(t, reg.SubLinters)
}

// -----------------------------------------------------------------------
// CheckRegistryCompleteness
// -----------------------------------------------------------------------

func TestCheckRegistryCompleteness_AllMatch(t *testing.T) {
	t.Parallel()

	rootDir := buildRegistryRoot(t, minimalRegistryYAML(), []string{"cgo_free_sqlite", "crypto_rand"})

	orphaned, missing, err := CheckRegistryCompleteness(rootDir, os.ReadFile, os.ReadDir)

	require.NoError(t, err)
	require.Empty(t, orphaned)
	require.Empty(t, missing)
}

func TestCheckRegistryCompleteness_OrphanedDir(t *testing.T) {
	t.Parallel()

	// Filesystem has extra_linter but it's not in the YAML.
	rootDir := buildRegistryRoot(t, minimalRegistryYAML(), []string{"cgo_free_sqlite", "crypto_rand", "extra_linter"})

	orphaned, missing, err := CheckRegistryCompleteness(rootDir, os.ReadFile, os.ReadDir)

	require.NoError(t, err)
	require.Equal(t, []string{"extra_linter"}, orphaned)
	require.Empty(t, missing)
}

func TestCheckRegistryCompleteness_MissingDir(t *testing.T) {
	t.Parallel()

	// YAML has crypto_rand but filesystem only has cgo_free_sqlite.
	rootDir := buildRegistryRoot(t, minimalRegistryYAML(), []string{"cgo_free_sqlite"})

	orphaned, missing, err := CheckRegistryCompleteness(rootDir, os.ReadFile, os.ReadDir)

	require.NoError(t, err)
	require.Empty(t, orphaned)
	require.Equal(t, []string{"crypto_rand"}, missing)
}

func TestCheckRegistryCompleteness_SkipsRegistryDir(t *testing.T) {
	t.Parallel()

	// Filesystem has a "registry" directory which should be excluded.
	rootDir := buildRegistryRoot(t, minimalRegistryYAML(), []string{"cgo_free_sqlite", "crypto_rand", "registry"})

	orphaned, missing, err := CheckRegistryCompleteness(rootDir, os.ReadFile, os.ReadDir)

	require.NoError(t, err)
	require.Empty(t, orphaned, "registry directory should be excluded from filesystem scan")
	require.Empty(t, missing)
}

func TestCheckRegistryCompleteness_SkipsNonDirectories(t *testing.T) {
	t.Parallel()

	// A regular file in the fitness dir should not be treated as a sub-linter.
	rootDir := buildRegistryRoot(t, minimalRegistryYAML(), []string{"cgo_free_sqlite", "crypto_rand"})

	// Create a file (not a directory) in the fitness dir to verify non-dirs are ignored.
	fitnessDir := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDLintFitnessDir))
	require.NoError(t, os.WriteFile(filepath.Join(fitnessDir, "some-extra-file.yaml"), []byte(""), cryptoutilSharedMagic.FilePermissionsDefault))

	orphaned, missing, err := CheckRegistryCompleteness(rootDir, os.ReadFile, os.ReadDir)

	require.NoError(t, err)
	require.Empty(t, orphaned)
	require.Empty(t, missing)
}

func TestCheckRegistryCompleteness_ManifestLoadError(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir() // no YAML file

	orphaned, missing, err := CheckRegistryCompleteness(rootDir, os.ReadFile, os.ReadDir)

	require.Error(t, err)
	require.Nil(t, orphaned)
	require.Nil(t, missing)
}

func TestCheckRegistryCompleteness_ReadDirError(t *testing.T) {
	t.Parallel()

	stubReadDirFn := func(_ string) ([]os.DirEntry, error) {
		return nil, fmt.Errorf("simulated ReadDir error")
	}

	rootDir := buildRegistryRoot(t, minimalRegistryYAML(), nil)

	orphaned, missing, err := CheckRegistryCompleteness(rootDir, os.ReadFile, stubReadDirFn)

	require.Error(t, err)
	require.Nil(t, orphaned)
	require.Nil(t, missing)
	require.Contains(t, err.Error(), "simulated ReadDir error")
}

func TestCheckRegistryCompleteness_BothOrphanedAndMissing(t *testing.T) {
	t.Parallel()

	// YAML has crypto_rand (missing from FS) + FS has extra_unknown (not in YAML).
	rootDir := buildRegistryRoot(t, minimalRegistryYAML(), []string{"cgo_free_sqlite", "extra_unknown"})

	orphaned, missing, err := CheckRegistryCompleteness(rootDir, os.ReadFile, os.ReadDir)

	require.NoError(t, err)
	require.Equal(t, []string{"extra_unknown"}, orphaned)
	require.Equal(t, []string{"crypto_rand"}, missing)
}

func TestCheckRegistryCompleteness_SortedResults(t *testing.T) {
	t.Parallel()

	yaml := `sub_linters:
  - name: zzz-last
    directory: zzz_last
    description: Last linter.
    category: architecture
  - name: aaa-first
    directory: aaa_first
    description: First linter.
    category: architecture
`
	// Filesystem has two extra dirs and none from YAML.
	rootDir := buildRegistryRoot(t, yaml, []string{"zzz_extra", "aaa_extra"})

	orphaned, missing, err := CheckRegistryCompleteness(rootDir, os.ReadFile, os.ReadDir)

	require.NoError(t, err)
	require.Equal(t, []string{"aaa_extra", "zzz_extra"}, orphaned, "orphaned should be sorted")
	require.Equal(t, []string{"aaa_first", "zzz_last"}, missing, "missing should be sorted")
}

// -----------------------------------------------------------------------
// CheckInDir
// -----------------------------------------------------------------------

func TestCheckInDir_AllMatch(t *testing.T) {
	t.Parallel()

	rootDir := buildRegistryRoot(t, minimalRegistryYAML(), []string{"cgo_free_sqlite", "crypto_rand"})
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}

func TestCheckInDir_Orphaned(t *testing.T) {
	t.Parallel()

	rootDir := buildRegistryRoot(t, minimalRegistryYAML(), []string{"cgo_free_sqlite", "crypto_rand", "orphan_dir"})
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "ORPHANED")
	require.Contains(t, err.Error(), "orphan_dir")
}

func TestCheckInDir_Missing(t *testing.T) {
	t.Parallel()

	rootDir := buildRegistryRoot(t, minimalRegistryYAML(), []string{"cgo_free_sqlite"})
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "MISSING")
	require.Contains(t, err.Error(), "crypto_rand")
}

func TestCheckInDir_ManifestError(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir() // no manifest
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "fitness-registry-completeness")
}

func TestCheckInDir_ReadDirError(t *testing.T) {
	t.Parallel()

	stubReadDirFn := func(_ string) ([]os.DirEntry, error) {
		return nil, fmt.Errorf("simulated CheckInDir ReadDir error")
	}

	rootDir := buildRegistryRoot(t, minimalRegistryYAML(), nil)
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := checkInDir(logger, rootDir, os.ReadFile, stubReadDirFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "simulated CheckInDir ReadDir error")
}

// -----------------------------------------------------------------------
// Check (integration test using real project)
// -----------------------------------------------------------------------

func TestCheck_Integration(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-fitness-registry-completeness")
	err := Check(logger)

	require.NoError(t, err, "fitness-registry-completeness should pass on real project files")
}

func TestCheck_ProjectRootError(t *testing.T) {
	t.Parallel()

	stubGetwdFn := func() (string, error) {
		return "", fmt.Errorf("simulated root error")
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := check(logger, stubGetwdFn, os.ReadFile, os.ReadDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "simulated root error")
}

func TestFindFitnessProjectRoot_GetwdError(t *testing.T) {
	t.Parallel()

	stubGetwdFn := func() (string, error) {
		return "", fmt.Errorf("simulated getwd error")
	}

	result, err := findFitnessProjectRoot(stubGetwdFn)

	require.Error(t, err)
	require.Empty(t, result)
	require.Contains(t, err.Error(), "simulated getwd error")
}

func TestFindFitnessProjectRoot_GoModNotFound(t *testing.T) {
	t.Parallel()

	// Use the filesystem root; it will never have go.mod.
	stubGetwdFn := func() (string, error) {
		// Return a path that has no go.mod in itself or any parent.
		// On Windows this is e.g. C:\ on Unix it is /
		root := filepath.VolumeName(t.TempDir())
		if root == "" {
			root = "/"
		} else {
			root += "\\"
		}

		return root, nil
	}

	result, err := findFitnessProjectRoot(stubGetwdFn)

	require.Error(t, err)
	require.Empty(t, result)
	require.Contains(t, err.Error(), "go.mod not found")
}
