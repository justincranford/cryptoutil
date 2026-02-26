// Copyright (c) 2025 Justin Cranford

package admin_port_exposure

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

func TestCheckComposeFile_InvalidFile(t *testing.T) {
	t.Parallel()

	violations, err := CheckComposeFile("/nonexistent/compose.yml")
	require.Error(t, err, "should error on invalid file")
	require.Nil(t, violations)
}

// TestCheckComposeFile_FileOpenError tests the error path when compose file cannot be opened.
func TestCheckComposeFile_FileOpenError(t *testing.T) {
	t.Parallel()

	violations, err := CheckComposeFile("/nonexistent/path/to/compose.yml")
	require.Error(t, err, "should return error for non-existent file")
	require.Nil(t, violations, "should return nil violations on error")
	require.Contains(t, err.Error(), "failed to open file", "error should indicate file open failure")
}

func TestCheck_NoComposeFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {"/some/file.go"},
	}

	err := Check(logger, filesByExtension)
	require.NoError(t, err, "No compose files should return nil")
}

func TestCheck_WithCleanComposeFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	composeFile := filepath.Join(tmpDir, "compose.yml")
	content := "services:\n  app:\n    ports:\n      - 8080:8080\n"
	require.NoError(t, os.WriteFile(composeFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {composeFile},
	}

	err := Check(logger, filesByExtension)
	require.NoError(t, err, "Non-admin port should not be a violation")
}

func TestCheck_WithAdminPortExposure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	composeFile := filepath.Join(tmpDir, "compose.yml")
	content := "services:\n  admin:\n    ports:\n      - 9090:9090\n"
	require.NoError(t, os.WriteFile(composeFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {composeFile},
	}

	err := Check(logger, filesByExtension)
	require.Error(t, err, "Admin port 9090 exposure should be a violation")
	require.Contains(t, err.Error(), "admin port exposure violations")
}

func TestCheck_FailedToOpenComposeFile(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {"/nonexistent/path/compose.yml"},
	}

	// FindComposeFiles returns files named compose.yml/compose.yaml/etc.
	// This file doesn't exist, so CheckComposeFile returns an error.
	// Check() logs a warning and continues, so the result is nil.
	err := Check(logger, filesByExtension)
	require.NoError(t, err, "Warning on non-existent file should not fail Check()")
}

func TestCheckComposeFile_NoPortsSection(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	composeFile := filepath.Join(tmpDir, "compose.yml")
	content := "services:\n  app:\n    image: alpine\n"
	require.NoError(t, os.WriteFile(composeFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	violations, err := CheckComposeFile(composeFile)
	require.NoError(t, err)
	require.Empty(t, violations, "No ports section means no violations")
}

func TestCheckComposeFile_PortRangeToAdmin(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	composeFile := filepath.Join(tmpDir, "compose.yml")
	content := "services:\n  admin:\n    ports:\n      - 9080-9089:9090\n"
	require.NoError(t, os.WriteFile(composeFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	violations, err := CheckComposeFile(composeFile)
	require.NoError(t, err)
	require.Len(t, violations, 1, "Port range to :9090 is a violation")
}

func TestCheckComposeFile_CommentedPortLine(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	composeFile := filepath.Join(tmpDir, "compose.yml")
	content := "services:\n  app:\n    ports:\n      # - 9090:9090\n"
	require.NoError(t, os.WriteFile(composeFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	violations, err := CheckComposeFile(composeFile)
	require.NoError(t, err)
	require.Empty(t, violations, "Commented-out port should not be a violation")
}

func TestCheckComposeFile_ExitsPortsSection(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	composeFile := filepath.Join(tmpDir, "compose.yml")
	// After "volumes:", we are no longer in a ports section.
	content := "services:\n  app:\n    ports:\n      - 8080:8080\n    volumes:\n      - ./data:/data\n"
	require.NoError(t, os.WriteFile(composeFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	violations, err := CheckComposeFile(composeFile)
	require.NoError(t, err)
	require.Empty(t, violations, "No admin port exposure after exiting ports section")
}

func TestCheckComposeFile_DifferentHostPort(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	composeFile := filepath.Join(tmpDir, "compose.yml")
	content := "services:\n  admin:\n    ports:\n      - 19090:9090\n"
	require.NoError(t, os.WriteFile(composeFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	violations, err := CheckComposeFile(composeFile)
	require.NoError(t, err)
	require.Len(t, violations, 1, "Mapping 19090->9090 exposes admin port")
	require.Contains(t, violations[0].Content, "9090")
}

