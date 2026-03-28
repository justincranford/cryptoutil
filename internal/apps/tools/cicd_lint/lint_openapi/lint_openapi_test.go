// Copyright (c) 2025 Justin Cranford

package lint_openapi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestLint_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Lint(logger, map[string][]string{})

	require.NoError(t, err)
}

func TestLint_CleanSpecs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	apiDir := filepath.Join(tmpDir, "api", "test-service")
	err := os.MkdirAll(apiDir, cryptoutilSharedMagic.CICDTempDirPermissions)
	require.NoError(t, err)

	// Create a valid OpenAPI spec file.
	specFile := filepath.Join(apiDir, "openapi_spec.yaml")
	specContent := `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths: {}
`

	err = os.WriteFile(specFile, []byte(specContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Lint(logger, map[string][]string{"yaml": {specFile}})

	require.NoError(t, err)
}

func TestLint_VersionViolation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	apiDir := filepath.Join(tmpDir, "api", "test-service")
	err := os.MkdirAll(apiDir, cryptoutilSharedMagic.CICDTempDirPermissions)
	require.NoError(t, err)

	specFile := filepath.Join(apiDir, "openapi_spec.yaml")
	specContent := `openapi: '3.1.0'
info:
  title: Test API
  version: 1.0.0
`

	err = os.WriteFile(specFile, []byte(specContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Lint(logger, map[string][]string{"yaml": {specFile}})

	require.Error(t, err)
	require.Contains(t, err.Error(), "lint-openapi failed")
}

func TestLint_CodegenConfigViolation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	apiDir := filepath.Join(tmpDir, "api", "test-service")
	err := os.MkdirAll(apiDir, cryptoutilSharedMagic.CICDTempDirPermissions)
	require.NoError(t, err)

	configFile := filepath.Join(apiDir, "openapi-gen_config_server.yaml")
	configContent := `package: server
output-options:
  additional-initialisms:
    - JWT
`

	err = os.WriteFile(configFile, []byte(configContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Lint(logger, map[string][]string{"yaml": {configFile}})

	require.Error(t, err)
	require.Contains(t, err.Error(), "lint-openapi failed")
}
