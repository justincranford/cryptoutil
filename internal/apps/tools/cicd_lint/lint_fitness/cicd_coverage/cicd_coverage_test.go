// Copyright (c) 2025 Justin Cranford

package cicd_coverage_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	. "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/cicd_coverage"
)

// buildValidContent generates file content containing all lint-* and format-*
// commands from ValidCommands so the coverage check always passes.
func buildValidContent() string {
	var sb strings.Builder

	for cmd := range cryptoutilSharedMagic.ValidCommands {
		if strings.HasPrefix(cmd, "lint-") || strings.HasPrefix(cmd, "format-") {
			sb.WriteString(fmt.Sprintf("./cicd %s\n", cmd))
		}
	}

	return sb.String()
}

// createValidDir creates a temp directory with valid coverage files for all two required paths.
func createValidDir(t *testing.T) string {
	t.Helper()

	rootDir := t.TempDir()
	validContent := buildValidContent()

	for _, relPath := range []string{
		".pre-commit-config.yaml",
		filepath.Join(cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "workflows", "ci-quality.yml"),
	} {
		fullPath := filepath.Join(rootDir, relPath)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), cryptoutilSharedMagic.CICDTempDirPermissions))
		require.NoError(t, os.WriteFile(fullPath, []byte(validContent), cryptoutilSharedMagic.CacheFilePermissions))
	}

	return rootDir
}

func TestCheckInDir_AllValid(t *testing.T) {
	t.Parallel()

	rootDir := createValidDir(t)
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}

func TestCheckInDir_MissingPreCommitFile(t *testing.T) {
	t.Parallel()

	rootDir := createValidDir(t)
	require.NoError(t, os.Remove(filepath.Join(rootDir, ".pre-commit-config.yaml")))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), ".pre-commit-config.yaml")
	require.Contains(t, err.Error(), "cannot read file")
}

func TestCheckInDir_MissingWorkflowFile(t *testing.T) {
	t.Parallel()

	rootDir := createValidDir(t)
	require.NoError(t, os.Remove(filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "workflows", "ci-quality.yml")))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "ci-quality.yml")
	require.Contains(t, err.Error(), "cannot read file")
}

func TestCheckInDir_MissingLinterInWorkflow(t *testing.T) {
	t.Parallel()

	// Find a lint command to omit.
	var omitCmd string

	for cmd := range cryptoutilSharedMagic.ValidCommands {
		if strings.HasPrefix(cmd, "lint-") {
			omitCmd = cmd

			break
		}
	}

	require.NotEmpty(t, omitCmd, "expected at least one lint-* command in ValidCommands")

	rootDir := t.TempDir()
	validContent := buildValidContent()

	// Build workflow content that is missing the omit command.
	var lines []string

	for _, line := range strings.Split(validContent, "\n") {
		if !strings.Contains(line, omitCmd) {
			lines = append(lines, line)
		}
	}

	workflowContent := strings.Join(lines, "\n")

	for _, entry := range []struct {
		relPath string
		content string
	}{
		{".pre-commit-config.yaml", validContent},
		{filepath.Join(cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "workflows", "ci-quality.yml"), workflowContent},
	} {
		fullPath := filepath.Join(rootDir, entry.relPath)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), cryptoutilSharedMagic.CICDTempDirPermissions))
		require.NoError(t, os.WriteFile(fullPath, []byte(entry.content), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "ci-quality.yml")
	require.Contains(t, err.Error(), omitCmd)
}

func TestCheckInDir_MissingFormatterInPreCommit(t *testing.T) {
	t.Parallel()

	// Find a format command to omit.
	var omitCmd string

	for cmd := range cryptoutilSharedMagic.ValidCommands {
		if strings.HasPrefix(cmd, "format-") {
			omitCmd = cmd

			break
		}
	}

	require.NotEmpty(t, omitCmd, "expected at least one format-* command in ValidCommands")

	rootDir := t.TempDir()
	validContent := buildValidContent()

	// Build pre-commit content missing the formatter.
	var lines []string

	for _, line := range strings.Split(validContent, "\n") {
		if !strings.Contains(line, omitCmd) {
			lines = append(lines, line)
		}
	}

	preCommitContent := strings.Join(lines, "\n")

	for _, entry := range []struct {
		relPath string
		content string
	}{
		{".pre-commit-config.yaml", preCommitContent},
		{filepath.Join(cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "workflows", "ci-quality.yml"), validContent},
	} {
		fullPath := filepath.Join(rootDir, entry.relPath)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), cryptoutilSharedMagic.CICDTempDirPermissions))
		require.NoError(t, os.WriteFile(fullPath, []byte(entry.content), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), ".pre-commit-config.yaml")
	require.Contains(t, err.Error(), omitCmd)
}
