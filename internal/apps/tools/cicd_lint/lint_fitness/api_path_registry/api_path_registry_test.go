// Copyright (c) 2025 Justin Cranford

package api_path_registry

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err, "failed to get working directory")

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("skipping integration test: cannot find project root (no go.mod)")
		}

		dir = parent
	}
}

// setupSpecFile creates api/{psID}/openapi_spec.yaml with the given paths.
func setupSpecFile(t *testing.T, tmpDir, psID string, paths []string) {
	t.Helper()

	apiDir := filepath.Join(tmpDir, "api", psID)
	require.NoError(t, os.MkdirAll(apiDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	content := "paths:\n"
	for _, p := range paths {
		content += fmt.Sprintf("  %s:\n    get:\n      summary: test\n", p)
	}

	require.NoError(t, os.WriteFile(
		filepath.Join(apiDir, "openapi_spec.yaml"),
		[]byte(content),
		cryptoutilSharedMagic.FilePermissions,
	))
}

// setupAllValidAPIDirs creates api/{ps-id}/openapi_spec.yaml files for all services
// that have api_resources in the registry, with matching paths.
func setupAllValidAPIDirs(t *testing.T, tmpDir string) {
	t.Helper()

	for _, info := range lintFitnessRegistry.AllAPIResources() {
		if len(info.Resources) == 0 {
			continue
		}

		setupSpecFile(t, tmpDir, info.PSID, info.Resources)
	}
}

func TestCheck_DelegatesToCheckInDir(t *testing.T) {
	// Not parallel: changes process working directory.
	root := findProjectRoot(t)

	orig, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	defer func() { _ = os.Chdir(orig) }()

	err = Check(newTestLogger())
	require.NoError(t, err, "Check() should pass on real workspace")
}

func TestCheckInDir_RealWorkspace(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	err := CheckInDir(newTestLogger(), root, os.ReadDir, os.ReadFile)
	require.NoError(t, err, "all registry services with api_resources should have matching OpenAPI spec paths")
}

func TestCheckInDir_AllPathsMatch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllValidAPIDirs(t, tmpDir)

	err := CheckInDir(newTestLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.NoError(t, err)
}

func TestCheckInDir_MissingFromSpec(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Set up all services with valid matching paths.
	setupAllValidAPIDirs(t, tmpDir)

	// Overwrite sm-kms spec with a subset of registry paths (missing one).
	specPaths := []string{
		"/elastic-keys",
		// Deliberately omit other paths to trigger a "missing from spec" violation.
	}

	setupSpecFile(t, tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS, specPaths)

	err := CheckInDir(newTestLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
	assert.Contains(t, err.Error(), "declared in registry but missing from OpenAPI spec")
}

func TestCheckInDir_UndeclaredInSpec(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Set up all services with valid matching paths.
	setupAllValidAPIDirs(t, tmpDir)

	// Look up the registry resources for sm-kms and add an extra path.
	var smKMSPaths []string

	for _, info := range lintFitnessRegistry.AllAPIResources() {
		if info.PSID == cryptoutilSharedMagic.OTLPServiceSMKMS {
			smKMSPaths = append(smKMSPaths, info.Resources...)

			break
		}
	}

	// Add an extra path that is NOT in the registry.
	smKMSPaths = append(smKMSPaths, "/undeclared-extra-path")
	setupSpecFile(t, tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS, smKMSPaths)

	err := CheckInDir(newTestLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
	assert.Contains(t, err.Error(), "in OpenAPI spec but not declared in registry api_resources")
	assert.Contains(t, err.Error(), "/undeclared-extra-path")
}

func TestCheckInDir_SkipsEmptyResources(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Only create directories for services WITH api_resources.
	// Services with empty api_resources (identity-rp, identity-spa) have NO api dirs.
	setupAllValidAPIDirs(t, tmpDir)

	// Verify that services with empty api_resources don't cause errors
	// (no spec file exists for them, and that's fine).
	err := CheckInDir(newTestLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.NoError(t, err)
}

func TestCheckInDir_MissingAPIDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Set up all services EXCEPT sm-kms — leave its api dir missing.
	for _, info := range lintFitnessRegistry.AllAPIResources() {
		if len(info.Resources) == 0 || info.PSID == cryptoutilSharedMagic.OTLPServiceSMKMS {
			continue
		}

		setupSpecFile(t, tmpDir, info.PSID, info.Resources)
	}

	err := CheckInDir(newTestLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
	assert.Contains(t, err.Error(), "cannot read api directory")
}

func TestCheckInDir_NoSpecFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Set up all services with valid paths.
	setupAllValidAPIDirs(t, tmpDir)

	// Remove all spec files from sm-kms (leave only an empty directory).
	smKMSDir := filepath.Join(tmpDir, "api", cryptoutilSharedMagic.OTLPServiceSMKMS)
	entries, err := os.ReadDir(smKMSDir)
	require.NoError(t, err)

	for _, e := range entries {
		require.NoError(t, os.Remove(filepath.Join(smKMSDir, e.Name())))
	}

	err = CheckInDir(newTestLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
	assert.Contains(t, err.Error(), "no openapi_spec*.yaml files found")
}

func TestCheckInDir_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Set up all services with valid paths.
	setupAllValidAPIDirs(t, tmpDir)

	// Replace sm-kms spec with malformed YAML.
	smKMSDir := filepath.Join(tmpDir, "api", cryptoutilSharedMagic.OTLPServiceSMKMS)
	require.NoError(t, os.WriteFile(
		filepath.Join(smKMSDir, "openapi_spec.yaml"),
		[]byte("paths: {invalid: [yaml: content"),
		cryptoutilSharedMagic.FilePermissions,
	))

	err := CheckInDir(newTestLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
}

func TestCheckInDir_SpecFileNoPathsBlock(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Set up all services with valid paths.
	setupAllValidAPIDirs(t, tmpDir)

	// Replace sm-kms spec with valid YAML but no "paths:" top-level key.
	smKMSDir := filepath.Join(tmpDir, "api", cryptoutilSharedMagic.OTLPServiceSMKMS)
	require.NoError(t, os.WriteFile(
		filepath.Join(smKMSDir, "openapi_spec.yaml"),
		[]byte("info:\n  title: test\n  version: 1.0.0\n"),
		cryptoutilSharedMagic.FilePermissions,
	))

	// A spec file with no paths block means 0 spec paths but registry has paths —
	// so every registry path is "missing from spec".
	err := CheckInDir(newTestLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
	assert.Contains(t, err.Error(), "declared in registry but missing from OpenAPI spec")
}

func TestCheckInDir_PathsBlockNotMapping(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Set up all services with valid paths.
	setupAllValidAPIDirs(t, tmpDir)

	// Replace sm-kms spec with valid YAML but "paths:" is a scalar (not a mapping).
	smKMSDir := filepath.Join(tmpDir, "api", cryptoutilSharedMagic.OTLPServiceSMKMS)
	require.NoError(t, os.WriteFile(
		filepath.Join(smKMSDir, "openapi_spec.yaml"),
		[]byte("paths: some_scalar_value\n"),
		cryptoutilSharedMagic.FilePermissions,
	))

	err := CheckInDir(newTestLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
	assert.Contains(t, err.Error(), "paths block is not a YAML mapping")
}

func TestCheckInDir_ReadDirError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllValidAPIDirs(t, tmpDir)

	err := CheckInDir(newTestLogger(), tmpDir, func(path string) ([]os.DirEntry, error) {
		if filepath.Base(path) == cryptoutilSharedMagic.OTLPServiceSMKMS {
			return nil, fmt.Errorf("injected OS error: ReadDir failed")
		}

		return os.ReadDir(path)
	}, os.ReadFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
	assert.Contains(t, err.Error(), "cannot read api directory")
}

func TestCheckInDir_ReadFileError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllValidAPIDirs(t, tmpDir)

	err := CheckInDir(newTestLogger(), tmpDir, os.ReadDir, func(path string) ([]byte, error) {
		if filepath.Base(path) == "openapi_spec.yaml" {
			return nil, fmt.Errorf("injected OS error: ReadFile failed")
		}

		return os.ReadFile(path)
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot parse")
}

func TestIsSpecFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		file string
		want bool
	}{
		{
			name: "standard spec file",
			file: "openapi_spec.yaml",
			want: true,
		},
		{
			name: "enrollment spec file",
			file: "openapi_spec_enrollment.yaml",
			want: true,
		},
		{
			name: "paths-only spec file",
			file: "openapi_spec_paths.yaml",
			want: true,
		},
		{
			name: "components excluded",
			file: "openapi_spec_components.yaml",
			want: false,
		},
		{
			name: "security components excluded",
			file: "openapi_spec_security_components.yaml",
			want: false,
		},
		{
			name: "gen_config excluded",
			file: "openapi_spec_gen_config.yaml",
			want: false,
		},
		{
			name: "gen_config variant excluded",
			file: "openapi_spec_gen_config_server.yaml",
			want: false,
		},
		{
			name: "not yaml extension excluded",
			file: "openapi_spec.json",
			want: false,
		},
		{
			name: "non-spec file excluded",
			file: "generate.go",
			want: false,
		},
		{
			name: "wrong prefix excluded",
			file: "spec.yaml",
			want: false,
		},
		{
			name: "README excluded",
			file: "README.md",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isSpecFile(tt.file)
			assert.Equal(t, tt.want, got, "isSpecFile(%q)", tt.file)
		})
	}
}
