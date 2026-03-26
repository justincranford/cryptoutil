// Copyright (c) 2025 Justin Cranford

package template_consistency_test

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessTemplateConsistency "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/template_consistency"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func makeSecretsDir(t *testing.T, rootDir string) string {
	t.Helper()

	secretsDir := filepath.Join(rootDir, "deployments", cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, "secrets")
	require.NoError(t, os.MkdirAll(secretsDir, cryptoutilSharedMagic.DirPermissions))

	return secretsDir
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

func TestFindViolationsInDir_EmptySecretsDir_NoViolations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	makeSecretsDir(t, tmpDir)

	violations, err := lintFitnessTemplateConsistency.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_HyphenatedNames_NoViolations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := makeSecretsDir(t, tmpDir)

	for _, name := range []string{
		"browser-password.secret",
		"hash-pepper-v3.secret",
		"postgres-database.secret",
		"unseal-1of5.secret",
		"browser-password.secret.never",
	} {
		require.NoError(t, os.WriteFile(
			filepath.Join(secretsDir, name),
			[]byte("value"),
			cryptoutilSharedMagic.FilePermissions,
		))
	}

	violations, err := lintFitnessTemplateConsistency.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_UnderscoreInSecretName_Violation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := makeSecretsDir(t, tmpDir)

	require.NoError(t, os.WriteFile(
		filepath.Join(secretsDir, "browser_password.secret"),
		[]byte("value"),
		cryptoutilSharedMagic.FilePermissions,
	))

	violations, err := lintFitnessTemplateConsistency.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	require.Len(t, violations, 1)
	assert.Equal(t, "browser_password.secret", violations[0])
}

func TestFindViolationsInDir_UnderscoreInNeverName_Violation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := makeSecretsDir(t, tmpDir)

	require.NoError(t, os.WriteFile(
		filepath.Join(secretsDir, "browser_password.secret.never"),
		[]byte("value"),
		cryptoutilSharedMagic.FilePermissions,
	))

	violations, err := lintFitnessTemplateConsistency.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	require.Len(t, violations, 1)
	assert.Equal(t, "browser_password.secret.never", violations[0])
}

func TestFindViolationsInDir_MultipleViolations_AllReported(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := makeSecretsDir(t, tmpDir)

	violating := []string{
		"browser_password.secret",
		"postgres_database.secret",
	}
	valid := []string{
		"hash-pepper-v3.secret",
	}

	for _, name := range append(violating, valid...) {
		require.NoError(t, os.WriteFile(
			filepath.Join(secretsDir, name),
			[]byte("value"),
			cryptoutilSharedMagic.FilePermissions,
		))
	}

	violations, err := lintFitnessTemplateConsistency.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Len(t, violations, 2)
}

func TestFindViolationsInDir_NonSecretFilesIgnored(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := makeSecretsDir(t, tmpDir)

	require.NoError(t, os.WriteFile(
		filepath.Join(secretsDir, "has_underscore.txt"),
		[]byte("not a secret"),
		cryptoutilSharedMagic.FilePermissions,
	))

	violations, err := lintFitnessTemplateConsistency.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_SubdirectoriesIgnored(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := makeSecretsDir(t, tmpDir)

	require.NoError(t, os.MkdirAll(
		filepath.Join(secretsDir, "under_score_dir"),
		cryptoutilSharedMagic.DirPermissions,
	))

	violations, err := lintFitnessTemplateConsistency.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_MissingSecretsDir_ReturnsError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	violations, err := lintFitnessTemplateConsistency.FindViolationsInDir(tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "skeleton-template secrets directory not found")
	assert.Nil(t, violations)
}

func TestCheckInDir_ValidNames(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := makeSecretsDir(t, tmpDir)

	require.NoError(t, os.WriteFile(
		filepath.Join(secretsDir, "hash-pepper-v3.secret"),
		[]byte("value"),
		cryptoutilSharedMagic.FilePermissions,
	))

	err := lintFitnessTemplateConsistency.CheckInDir(newTestLogger(), tmpDir)

	require.NoError(t, err)
}

func TestCheckInDir_InvalidName_ReturnsError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := makeSecretsDir(t, tmpDir)

	require.NoError(t, os.WriteFile(
		filepath.Join(secretsDir, "postgres_password.secret"),
		[]byte("value"),
		cryptoutilSharedMagic.FilePermissions,
	))

	err := lintFitnessTemplateConsistency.CheckInDir(newTestLogger(), tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "template-consistency: found 1 secret file")
}

func TestCheckInDir_MissingSecretsDir_ReturnsError(t *testing.T) {
	t.Parallel()

	err := lintFitnessTemplateConsistency.CheckInDir(newTestLogger(), "/nonexistent/root")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check template consistency")
}

// TestCheck_Integration runs the linter against the real workspace.
func TestCheck_Integration(t *testing.T) {
	root := findProjectRoot(t)

	err := lintFitnessTemplateConsistency.CheckInDir(newTestLogger(), root)

	require.NoError(t, err)
}

// TestCheck_FromWorkspaceRoot verifies Check() (no rootDir) works from project root.
// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_FromWorkspaceRoot(t *testing.T) {
	root := findProjectRoot(t)

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	err = lintFitnessTemplateConsistency.Check(newTestLogger())
	require.NoError(t, err)
}
