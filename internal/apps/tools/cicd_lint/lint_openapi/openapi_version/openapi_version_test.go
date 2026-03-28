// Copyright (c) 2025 Justin Cranford

package openapi_version

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestCheck_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger, map[string][]string{})

	require.NoError(t, err)
}

func TestCheck_ValidSpec(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a path that includes "api/" so the filter matches.
	apiDir := filepath.Join(tmpDir, "api", "test-service")
	err := os.MkdirAll(apiDir, cryptoutilSharedMagic.CICDTempDirPermissions)
	require.NoError(t, err)

	specFile := filepath.Join(apiDir, "openapi_spec.yaml")
	content := `# OpenAPI spec
openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths: {}
`

	err = os.WriteFile(specFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"yaml": {specFile}})

	require.NoError(t, err)
}

func TestCheck_WrongVersion(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	apiDir := filepath.Join(tmpDir, "api", "test-service")
	err := os.MkdirAll(apiDir, cryptoutilSharedMagic.CICDTempDirPermissions)
	require.NoError(t, err)

	specFile := filepath.Join(apiDir, "openapi_spec.yaml")
	content := `openapi: "3.1.0"
info:
  title: Test API
  version: 1.0.0
`

	err = os.WriteFile(specFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"yaml": {specFile}})

	require.Error(t, err)
	require.Contains(t, err.Error(), "openapi-version")
}

func TestCheck_SwaggerVersion(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	apiDir := filepath.Join(tmpDir, "api", "test-service")
	err := os.MkdirAll(apiDir, cryptoutilSharedMagic.CICDTempDirPermissions)
	require.NoError(t, err)

	specFile := filepath.Join(apiDir, "openapi_spec.yaml")
	content := `openapi: '2.0'
info:
  title: Legacy API
  version: 1.0.0
`

	err = os.WriteFile(specFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"yaml": {specFile}})

	require.Error(t, err)
	require.Contains(t, err.Error(), "openapi-version")
}

func TestCheck_ComponentFileNoVersion(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	apiDir := filepath.Join(tmpDir, "api", "test-service")
	err := os.MkdirAll(apiDir, cryptoutilSharedMagic.CICDTempDirPermissions)
	require.NoError(t, err)

	specFile := filepath.Join(apiDir, "openapi_spec_components.yaml")
	content := `# Components file — no openapi version field
components:
  schemas:
    Error:
      type: object
`

	err = os.WriteFile(specFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"yaml": {specFile}})

	require.NoError(t, err, "component files without openapi field should pass")
}

func TestCheck_NonAPIFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	yamlFile := filepath.Join(tmpDir, "config.yaml")
	content := `key: value
`

	err := os.WriteFile(yamlFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"yaml": {yamlFile}})

	require.NoError(t, err, "non-api YAML files should be skipped")
}

func TestExtractOpenAPIVersion_NonexistentFile(t *testing.T) {
	t.Parallel()

	_, err := extractOpenAPIVersion("/nonexistent/file.yaml")

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to open file")
}

func TestCheck_QuotedVersion(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	apiDir := filepath.Join(tmpDir, "api", "test-service")
	err := os.MkdirAll(apiDir, cryptoutilSharedMagic.CICDTempDirPermissions)
	require.NoError(t, err)

	specFile := filepath.Join(apiDir, "openapi_spec.yaml")
	content := `openapi: '3.0.3'
info:
  title: Test API
  version: 1.0.0
`

	err = os.WriteFile(specFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"yaml": {specFile}})

	require.NoError(t, err, "single-quoted version should be accepted")
}
