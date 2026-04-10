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
  - name: file-size-limits
    directory: file_size_limits
    description: Enforces file size limits.
    category: architecture
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

func TestLoadFitnessRegistry_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		yaml    string
		useRaw  bool
		wantErr string
		wantNil bool
	}{
		{name: "file not found", useRaw: true, wantErr: "failed to read", wantNil: true},
		{name: "invalid YAML", yaml: "!!! not valid: yaml: [", wantErr: "failed to parse", wantNil: true},
		{name: "empty registry", yaml: "sub_linters: []\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var rootDir string
			if tc.useRaw {
				rootDir = t.TempDir()
			} else {
				rootDir = buildRegistryRoot(t, tc.yaml, nil)
			}

			reg, err := LoadFitnessRegistry(rootDir, os.ReadFile)

			switch {
			case tc.wantErr != "":
				require.Error(t, err)
				require.Nil(t, reg)
				require.Contains(t, err.Error(), tc.wantErr)
			case tc.wantNil:
				require.NoError(t, err)
				require.Nil(t, reg)
			default:
				require.NoError(t, err)
				require.Empty(t, reg.SubLinters)
			}
		})
	}
}

// -----------------------------------------------------------------------
// CheckRegistryCompleteness
// -----------------------------------------------------------------------

func TestCheckRegistryCompleteness_SubdirVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		subdirs      []string
		wantOrphaned []string
		wantMissing  []string
	}{
		{
			name:    "all match",
			subdirs: []string{"cgo_free_sqlite", "file_size_limits"},
		},
		{
			name:         "orphaned dir",
			subdirs:      []string{"cgo_free_sqlite", "file_size_limits", "extra_linter"},
			wantOrphaned: []string{"extra_linter"},
		},
		{
			name:        "missing dir",
			subdirs:     []string{"cgo_free_sqlite"},
			wantMissing: []string{"file_size_limits"},
		},
		{
			name:    "skips registry dir",
			subdirs: []string{"cgo_free_sqlite", "file_size_limits", "registry"},
		},
		{
			name:         "both orphaned and missing",
			subdirs:      []string{"cgo_free_sqlite", "extra_unknown"},
			wantOrphaned: []string{"extra_unknown"},
			wantMissing:  []string{"file_size_limits"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := buildRegistryRoot(t, minimalRegistryYAML(), tc.subdirs)

			orphaned, missing, err := CheckRegistryCompleteness(rootDir, os.ReadFile, os.ReadDir)

			require.NoError(t, err)
			require.Equal(t, tc.wantOrphaned, orphaned)
			require.Equal(t, tc.wantMissing, missing)
		})
	}
}

func TestCheckRegistryCompleteness_SkipsNonDirectories(t *testing.T) {
	t.Parallel()

	rootDir := buildRegistryRoot(t, minimalRegistryYAML(), []string{"cgo_free_sqlite", "file_size_limits"})

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

	rootDir := t.TempDir() // no YAML file.

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
	rootDir := buildRegistryRoot(t, yaml, []string{"zzz_extra", "aaa_extra"})

	orphaned, missing, err := CheckRegistryCompleteness(rootDir, os.ReadFile, os.ReadDir)

	require.NoError(t, err)
	require.Equal(t, []string{"aaa_extra", "zzz_extra"}, orphaned, "orphaned should be sorted")
	require.Equal(t, []string{"aaa_first", "zzz_last"}, missing, "missing should be sorted")
}

// -----------------------------------------------------------------------
// CheckInDir
// -----------------------------------------------------------------------

func TestCheckInDir_SubdirVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		subdirs    []string
		wantErr    bool
		wantSubstr string
	}{
		{name: "all match", subdirs: []string{"cgo_free_sqlite", "file_size_limits"}},
		{name: "orphaned", subdirs: []string{"cgo_free_sqlite", "file_size_limits", "orphan_dir"}, wantErr: true, wantSubstr: "ORPHANED"},
		{name: "missing", subdirs: []string{"cgo_free_sqlite"}, wantErr: true, wantSubstr: "MISSING"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := buildRegistryRoot(t, minimalRegistryYAML(), tc.subdirs)
			logger := cryptoutilCmdCicdCommon.NewLogger("test")

			err := CheckInDir(logger, rootDir)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantSubstr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckInDir_ManifestError(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir() // no manifest.
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

func TestFindFitnessProjectRoot_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		getwdFn func() (string, error)
		wantErr string
	}{
		{
			name:    "getwd error",
			getwdFn: func() (string, error) { return "", fmt.Errorf("simulated getwd error") },
			wantErr: "simulated getwd error",
		},
		{
			name: "go.mod not found",
			getwdFn: func() (string, error) {
				root := filepath.VolumeName(os.TempDir())
				if root == "" {
					root = "/"
				} else {
					root += "\\"
				}

				return root, nil
			},
			wantErr: "go.mod not found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := findFitnessProjectRoot(tc.getwdFn)

			require.Error(t, err)
			require.Empty(t, result)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
