// Copyright (c) 2025-2026 Justin Cranford.
package secret_naming

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	// Run from project root (6 levels up from this test file).
	err := CheckInDir(logger, filepath.Join("..", "..", "..", "..", ".."))
	if err != nil {
		require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "unexpected violation in real workspace: %v", err)
	}
}

// Sequential: changes the process working directory to exercise Check.
func TestCheck_WorkspaceRoot(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", "..", "..", ".."))
	require.NoError(t, err)
	original, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))
	defer func() {
		require.NoError(t, os.Chdir(original))
	}()

	require.NoError(t, Check(cryptoutilCmdCicdCommon.NewLogger("test-check-workspace-root")))
}

func TestFindViolationsInDir_ValidSecrets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFiles map[string]string
	}{
		{
			name: "hyphenated secret files",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsUnseal1of5SecretValue4011E2E5: testContent,
				"deployments/sm-kms/secrets/hash-pepper-v3.secret":       testContent,
				"deployments/sm-kms/secrets/postgres-password.secret":    testContent,
			},
		},
		{
			name: "secret.never files are valid",
			setupFiles: map[string]string{
				"deployments/sm/secrets/browser-password.secret.never": testContent,
				"deployments/sm/secrets/service-password.secret.never": testContent,
			},
		},
		{
			name: "deployment without secrets dir is skipped",
			setupFiles: map[string]string{
				"deployments/shared-postgres/.gitkeep": "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := setupTempDir(t, tc.setupFiles)

			violations, err := FindViolationsInDir(rootDir)
			if err != nil {
				require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "unexpected error: %v", err)
			}

			if len(violations) > 0 {
				assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected no violations, got %d: %v", len(violations), violations)
			}
		})
	}
}

func TestFindViolationsInDir_InvalidSecrets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFiles     map[string]string
		wantViolations int
		wantSubstring  string
	}{
		{
			name: "underscore in filename",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsUnseal1of5Secret: testContent,
			},
			wantViolations: 1,
			wantSubstring:  "underscore",
		},
		{
			name: "missing .secret extension",
			setupFiles: map[string]string{
				"deployments/sm-kms/secrets/password.txt": testContent,
			},
			wantViolations: 1,
			wantSubstring:  "extension",
		},
		{
			name: "uppercase in filename",
			setupFiles: map[string]string{
				"deployments/sm-kms/secrets/Unseal-1of5.secret": testContent,
			},
			wantViolations: 1,
			wantSubstring:  "lowercase",
		},
		{
			name: "multiple violations in one file",
			setupFiles: map[string]string{
				"deployments/sm-kms/secrets/Unseal_1of5.txt": testContent,
			},
			wantViolations: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := setupTempDir(t, tc.setupFiles)

			violations, err := FindViolationsInDir(rootDir)
			if err != nil {
				require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "unexpected error: %v", err)
			}

			if len(violations) != tc.wantViolations {
				assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected %d violations, got %d: %v", tc.wantViolations, len(violations), violations)
			}

			if tc.wantSubstring != "" && len(violations) > 0 {
				found := false

				for _, v := range violations {
					if strings.Contains(v, tc.wantSubstring) {
						found = true

						break
					}
				}

				if !found {
					assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected violation containing %q, got: %v", tc.wantSubstring, violations)
				}
			}
		})
	}
}

func TestFindViolationsInDir_MissingDeploymentsDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	_, err := FindViolationsInDir(tmpDir)
	if err == nil {
		t.Fatal("expected error for missing deployments/ directory")
	}
}

func TestFindViolationsInDir_NonDirectoryAndUnreadableSecretsPath(t *testing.T) {
	t.Parallel()

	rootDir := setupTempDir(t, map[string]string{
		"deployments/standalone-file": testContent,
		"deployments/sm-kms/secrets":  testContent,
	})

	violations, err := FindViolationsInDir(rootDir)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0], "failed to read secrets/ directory")
}

func TestCheckInDir_MissingDeploymentsDirectory(t *testing.T) {
	t.Parallel()

	err := CheckInDir(cryptoutilCmdCicdCommon.NewLogger("test-check-missing-deployments"), t.TempDir())
	require.ErrorContains(t, err, "failed to check secret naming")
	require.ErrorContains(t, err, "failed to read deployments/")
}

func TestCheckInDir_WithViolation(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		testDeploymentsSmKMSSecretsUnseal1of5Secret: testContent,
	}
	rootDir := setupTempDir(t, files)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)
	if err == nil {
		t.Fatal("expected error from CheckInDir with violations")
	}
}

func TestCheckInDir_NoViolation(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		testDeploymentsSmKMSSecretsUnseal1of5SecretValue4011E2E5: testContent,
	}
	rootDir := setupTempDir(t, files)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)
	if err != nil {
		require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "unexpected error: %v", err)
	}
}

// setupTempDir creates a temporary directory with the given file structure.
func setupTempDir(t *testing.T, files map[string]string) string {
	t.Helper()

	rootDir := t.TempDir()

	for relPath, content := range files {
		absPath := filepath.Join(rootDir, filepath.FromSlash(relPath))

		if err := os.MkdirAll(filepath.Dir(absPath), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute); err != nil {
			require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "failed to create directory: %v", err)
		}

		if err := os.WriteFile(absPath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions); err != nil {
			require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "failed to write file: %v", err)
		}
	}

	return rootDir
}
