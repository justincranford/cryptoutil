// Copyright (c) 2025 Justin Cranford

package database_key_uniformity

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func newLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

// findProjectRoot walks up directories until it finds go.mod.
func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (go.mod not found)")
		}

		dir = parent
	}
}

// writeYML creates a YAML file in a temporary directory.
func writeYML(t *testing.T, dir, name, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(dir, 0o700))

	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.KeyFilePermissions))
}

// setupConfigDir creates a deployments/{psID}/config/ directory with the given files.
func setupConfigDir(t *testing.T, rootDir, psID string, files map[string]string) string {
	t.Helper()

	configDir := filepath.Join(rootDir, "deployments", psID, "config")

	for name, content := range files {
		writeYML(t, configDir, name, content)
	}

	return configDir
}

// TestCheck_RealWorkspace verifies that the real project workspace passes the linter.
func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRoot(t)

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(projectRoot))

	t.Cleanup(func() {
		require.NoError(t, os.Chdir(origDir))
	})

	err = Check(newLogger())
	assert.NoError(t, err)
}

// TestCheckInDir_AllCorrect verifies a setup where all files use database-url:.
func TestCheckInDir_AllCorrect(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	psID := cryptoutilSharedMagic.KMSServiceID

	setupConfigDir(t, tmpDir, psID, map[string]string{
		psID + "-app-common.yml":       "bind-public-address: \"0.0.0.0\"\n",
		psID + "-app-sqlite-1.yml":     "database-url: \"sqlite://file::memory:?cache=shared\"\n",
		psID + "-app-sqlite-2.yml":     "database-url: \"sqlite://file::memory:?cache=shared\"\n",
		psID + "-app-postgresql-1.yml": fmt.Sprintf("otlp-service: %s-postgres-1\n", psID),
		psID + "-app-postgresql-2.yml": fmt.Sprintf("otlp-service: %s-postgres-2\n", psID),
	})

	err := CheckInDir(newLogger(), tmpDir)
	assert.NoError(t, err)
}

// TestCheckInDir_NoConfigDir verifies that missing config directories are skipped.
func TestCheckInDir_NoConfigDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Only create the deployments root; no config subdirectories for any PS-ID.
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "deployments"), 0o700))

	err := CheckInDir(newLogger(), tmpDir)
	assert.NoError(t, err)
}

// TestCheckInDir_NestedDatabaseMapping verifies that nested database: mapping is detected.
func TestCheckInDir_NestedDatabaseMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{
			name: "type and dsn sub-keys",
			content: `database:
  type: postgres
  dsn: "${IDENTITY_DB_DSN}"
`,
		},
		{
			name: "single nested key",
			content: `database:
  url: "postgres://localhost/mydb"
`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			psID := cryptoutilSharedMagic.OTLPServiceSMIM
			filename := psID + "-domain.yml"

			setupConfigDir(t, tmpDir, psID, map[string]string{filename: tc.content})

			err := CheckInDir(newLogger(), tmpDir)
			require.Error(t, err)
			assert.Contains(t, err.Error(), psID)
			assert.Contains(t, err.Error(), filename)
			assert.Contains(t, err.Error(), "deprecated nested 'database:' mapping")
			assert.Contains(t, err.Error(), "database-url:")
		})
	}
}

// TestCheckInDir_ScalarDatabaseKeyAllowed verifies that a scalar database: value is allowed.
func TestCheckInDir_ScalarDatabaseKeyAllowed(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	psID := cryptoutilSharedMagic.OTLPServiceJoseJA

	// A scalar database: value (not a mapping) should not trigger a violation.
	setupConfigDir(t, tmpDir, psID, map[string]string{
		psID + "-app-common.yml": "database: \"sqlite://file::memory:?cache=shared\"\n",
	})

	err := CheckInDir(newLogger(), tmpDir)
	assert.NoError(t, err)
}

// TestCheckInDir_EmptyConfigDir verifies an empty config dir produces no violations.
func TestCheckInDir_EmptyConfigDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	psID := cryptoutilSharedMagic.PKICAServiceID

	// Create an empty config dir (no yml files).
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "deployments", psID, "config"), 0o700))

	err := CheckInDir(newLogger(), tmpDir)
	assert.NoError(t, err)
}

// TestCheckInDir_NonYMLFilesIgnored verifies that non-yml files in config dir are skipped.
func TestCheckInDir_NonYMLFilesIgnored(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	psID := cryptoutilSharedMagic.OTLPServiceSkeletonTemplate
	configDir := filepath.Join(tmpDir, "deployments", psID, "config")

	require.NoError(t, os.MkdirAll(configDir, 0o700))

	// Write a non-yml file that contains forbidden content.
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "readme.txt"), []byte("database:\n  type: postgres\n"), cryptoutilSharedMagic.KeyFilePermissions))

	err := CheckInDir(newLogger(), tmpDir)
	assert.NoError(t, err)
}

// TestCheckInDir_SubdirIgnored verifies that subdirectories in config dir are not recursed into.
func TestCheckInDir_SubdirIgnored(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	psID := cryptoutilSharedMagic.KMSServiceID
	configDir := filepath.Join(tmpDir, "deployments", psID, "config")
	subDir := filepath.Join(configDir, "nested")

	require.NoError(t, os.MkdirAll(subDir, 0o700))

	// Nested yml with forbidden content should not be detected.
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "nested.yml"), []byte("database:\n  type: postgres\n"), cryptoutilSharedMagic.KeyFilePermissions))

	err := CheckInDir(newLogger(), tmpDir)
	assert.NoError(t, err)
}

// TestCheckInDir_InvalidYAML verifies that unparseable YAML files produce an error.
func TestCheckInDir_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	psID := cryptoutilSharedMagic.KMSServiceID
	filename := psID + "-app-bad.yml"

	setupConfigDir(t, tmpDir, psID, map[string]string{filename: ":\n  invalid: [yaml"})

	err := CheckInDir(newLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), psID)
	assert.Contains(t, err.Error(), filename)
	assert.Contains(t, err.Error(), "YAML parse error")
}

// TestCheckInDir_ReadDirFnError verifies that a ReadDir error is reported.
//
// Sequential: mutates readDirFn package-level state.
func TestCheckInDir_ReadDirFnError(t *testing.T) {
	origReadDir := readDirFn

	defer func() { readDirFn = origReadDir }()

	sentinelErr := errors.New("simulated readdir failure")
	readDirFn = func(_ string) ([]os.DirEntry, error) { return nil, sentinelErr }

	tmpDir := t.TempDir()
	psID := cryptoutilSharedMagic.KMSServiceID
	// Create the config dir so os.Stat passes, then readDirFn will fail.
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "deployments", psID, "config"), 0o700))

	err := CheckInDir(newLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot read config dir")
}

// TestCheckInDir_ReadFileFnError verifies that a read error is reported.
//
// Sequential: mutates readFileFn package-level state.
func TestCheckInDir_ReadFileFnError(t *testing.T) {
	orig := readFileFn

	defer func() { readFileFn = orig }()

	sentinelErr := errors.New("simulated read failure")
	readFileFn = func(_ string) ([]byte, error) { return nil, sentinelErr }

	tmpDir := t.TempDir()
	psID := cryptoutilSharedMagic.KMSServiceID
	filename := psID + "-app-common.yml"

	setupConfigDir(t, tmpDir, psID, map[string]string{filename: "bind-public-address: \"0.0.0.0\"\n"})

	err := CheckInDir(newLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot read file")
}

// TestCheckInDir_MultipleViolations verifies that all violations across PS-IDs are reported.
func TestCheckInDir_MultipleViolations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	badContent := "database:\n  type: postgres\n  dsn: \"${DSN}\"\n"

	for _, psID := range []string{cryptoutilSharedMagic.KMSServiceID, cryptoutilSharedMagic.IMServiceID} {
		filename := fmt.Sprintf("%s-domain.yml", psID)
		setupConfigDir(t, tmpDir, psID, map[string]string{filename: badContent})
	}

	err := CheckInDir(newLogger(), tmpDir)
	require.Error(t, err)

	// Both PS-IDs should be reported.
	assert.True(t, strings.Contains(err.Error(), cryptoutilSharedMagic.KMSServiceID) && strings.Contains(err.Error(), cryptoutilSharedMagic.IMServiceID),
		"expected both %s and %s in error: %s", cryptoutilSharedMagic.KMSServiceID, cryptoutilSharedMagic.IMServiceID, err)
}

// TestCheck_DelegatesToCheckInDir verifies Check() delegates to CheckInDir(".").
//
// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_DelegatesToCheckInDir(t *testing.T) {
	tmpDir := t.TempDir()
	psID := cryptoutilSharedMagic.KMSServiceID
	filename := psID + "-app-common.yml"

	setupConfigDir(t, tmpDir, psID, map[string]string{filename: "bind-public-address: \"0.0.0.0\"\n"})

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(tmpDir))

	t.Cleanup(func() {
		require.NoError(t, os.Chdir(origDir))
	})

	// Should work because Check() calls CheckInDir(".") and tmpDir has a valid sm-kms config.
	err = Check(newLogger())
	assert.NoError(t, err)
}
