// Copyright (c) 2025 Justin Cranford

package lint_golangci

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

func TestLint_NoYAMLFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{}

	err := Lint(logger, filesByExtension)
	require.NoError(t, err, "Lint should succeed with no YAML files")
}

func TestLint_ValidConfig(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create a valid golangci-lint v2 config file.
	configContent := "version: \"2\"\nlinters:\n  enable:\n    - errcheck\n"
	configPath := filepath.Join(tmpDir, ".golangci.yml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), cryptoutilSharedMagic.CacheFilePermissions))

	filesByExtension := map[string][]string{
		"yml": {configPath},
	}

	// Lint should succeed for a valid config.
	err := Lint(logger, filesByExtension)
	require.NoError(t, err, "Lint should succeed with valid golangci config")
}

func TestLint_DeprecatedV1Options(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create a config file with deprecated v1 option under linters-settings.
	configContent := "linters-settings:\n  wsl:\n    force-err-cuddling: true\n"
	configPath := filepath.Join(tmpDir, ".golangci.yml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), cryptoutilSharedMagic.CacheFilePermissions))

	filesByExtension := map[string][]string{
		"yml": {configPath},
	}

	err := Lint(logger, filesByExtension)
	require.Error(t, err, "Lint should fail with deprecated v1 options")
}
