// Copyright (c) 2025 Justin Cranford

package config_rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("config-rules-test")
}

// findProjectRoot walks up from cwd to find go.mod.
func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		require.NotEqual(t, dir, parent, "go.mod not found")

		dir = parent
	}
}

// --- Integration tests against real workspace ---

func TestCheckKeyNaming_RealWorkspace(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)
	logger := newTestLogger()

	err := checkKeyNamingInDir(logger, root)
	assert.NoError(t, err, "config-key-naming should pass on real workspace")
}

func TestCheckHeaderIdentity_RealWorkspace(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)
	logger := newTestLogger()

	err := checkHeaderIdentityInDir(logger, root)
	assert.NoError(t, err, "config-header-identity should pass on real workspace")
}

func TestCheckInstanceMinimal_RealWorkspace(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)
	logger := newTestLogger()

	err := checkInstanceMinimalInDir(logger, root)
	assert.NoError(t, err, "config-instance-minimal should pass on real workspace")
}

func TestCheckCommonComplete_RealWorkspace(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)
	logger := newTestLogger()

	err := checkCommonCompleteInDir(logger, root)
	assert.NoError(t, err, "config-common-complete should pass on real workspace")
}

// --- Public wrapper tests ---

// TestCheckKeyNaming_PublicWrapper verifies the public wrapper delegates correctly.
// From the test package cwd, no deployment config dirs exist, so the glob returns
// no files and the function succeeds (no violations = no error).
func TestCheckKeyNaming_PublicWrapper(t *testing.T) {
	t.Parallel()

	logger := newTestLogger()

	// From non-root cwd, glob finds no deployment configs — no violations.
	err := CheckKeyNaming(logger)
	assert.NoError(t, err)
}

func TestCheckHeaderIdentity_PublicWrapper(t *testing.T) {
	t.Parallel()

	logger := newTestLogger()

	err := CheckHeaderIdentity(logger)
	assert.Error(t, err)
}

// TestCheckInstanceMinimal_PublicWrapper verifies the public wrapper delegates correctly.
// From the test package cwd, no instance config files exist, so the function succeeds.
func TestCheckInstanceMinimal_PublicWrapper(t *testing.T) {
	t.Parallel()

	logger := newTestLogger()

	// From non-root cwd, glob finds no instance configs — no violations.
	err := CheckInstanceMinimal(logger)
	assert.NoError(t, err)
}

func TestCheckCommonComplete_PublicWrapper(t *testing.T) {
	t.Parallel()

	logger := newTestLogger()

	err := CheckCommonComplete(logger)
	assert.Error(t, err)
}

// --- collectNonKebabKeys edge cases ---

func TestCollectNonKebabKeys_NilNode(t *testing.T) {
	t.Parallel()

	result := collectNonKebabKeys(nil, "")
	assert.Empty(t, result)
}

func TestCollectNonKebabKeys_SequenceWithMappings(t *testing.T) {
	t.Parallel()

	input := "- bad_key: value\n- good-key: value\n"

	var node yaml.Node

	require.NoError(t, yaml.Unmarshal([]byte(input), &node))

	violations := collectNonKebabKeys(&node, "")
	assert.Len(t, violations, 1)
	assert.Contains(t, violations[0], "bad_key")
}

// --- Helpers ---

// setupMinimalStructure creates deployment and config dirs for sm-kms only.
func setupMinimalStructure(t *testing.T, root string) {
	t.Helper()

	for _, ps := range allTestPSIDs() {
		// Create deployment config dir.
		configDir := filepath.Join(root, "deployments", ps, "config")
		require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.CICDTempDirPermissions))

		// Create standalone config dir and framework + domain files.
		standaloneDir := filepath.Join(root, cryptoutilSharedMagic.CICDConfigsDir, ps)
		require.NoError(t, os.MkdirAll(standaloneDir, cryptoutilSharedMagic.CICDTempDirPermissions))

		writeFile(t, filepath.Join(standaloneDir, ps+"-framework.yml"),
			"# "+ps+" Framework Configuration\nbind-public-address: 127.0.0.1\n")
		writeFile(t, filepath.Join(standaloneDir, ps+"-domain.yml"),
			"# "+ps+" Domain Configuration\n")
	}
}

// allTestPSIDs returns all PS-IDs from the registry for minimal test scaffolding.
func allTestPSIDs() []string {
	psIDs := make([]string, 0)

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		psIDs = append(psIDs, ps.PSID)
	}

	return psIDs
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	dir := filepath.Dir(path)

	require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.CICDTempDirPermissions))

	content = strings.ReplaceAll(content, "\r\n", "\n")

	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))
}
