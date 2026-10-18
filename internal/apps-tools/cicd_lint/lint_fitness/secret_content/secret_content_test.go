// Copyright (c) 2025-2026 Justin Cranford.
package secret_content

import (
	"os"
	"path/filepath"
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

func TestFindViolationsInDir_ValidDeployments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFiles map[string]string
	}{
		{
			name: "valid service deployment",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsHashPepperV3Secret:     "sm-kms-hash-pepper-v3-" + base64url43A(),
				testDeploymentsSmKMSSecretsBrowserUsernameSecret:  "sm-kms-browser-user",
				testDeploymentsSmKMSSecretsBrowserPasswordSecret:  "sm-kms-browser-pass-" + base64url43B(),
				testDeploymentsSmKMSSecretsServiceUsernameSecret:  "sm-kms-service-user",
				testDeploymentsSmKMSSecretsServicePasswordSecret:  "sm-kms-service-pass-" + base64url43C(),
				testDeploymentsSmKMSSecretsPostgresDatabaseSecret: "sm_kms_database",
				testDeploymentsSmKMSSecretsPostgresUsernameSecret: "sm_kms_database_user",
				testDeploymentsSmKMSSecretsPostgresPasswordSecret: "sm_kms_database_pass-" + base64url43D(),
				testDeploymentsSmKMSSecretsPostgresURLSecret:      "postgres://sm_kms_database_user:sm_kms_database_pass-" + base64url43D() + "@shared-postgres-leader:5432/sm_kms_database",
			},
		},
		{
			name: "valid product deployment",
			setupFiles: map[string]string{
				"deployments/sm/secrets/hash-pepper-v3.secret":         "sm-hash-pepper-v3-" + base64url43A(),
				testDeploymentsSmSecretsBrowserUsernameSecretNever:     NeverMarkerProduct,
				testDeploymentsSmSecretsBrowserPasswordSecretNever:     NeverMarkerProduct,
				"deployments/sm/secrets/service-username.secret.never": NeverMarkerProduct,
				"deployments/sm/secrets/service-password.secret.never": NeverMarkerProduct,
				"deployments/sm/secrets/postgres-database.secret":      "sm_database",
				"deployments/sm/secrets/postgres-username.secret":      "sm_database_user",
				"deployments/sm/secrets/postgres-password.secret":      "sm_database_pass-" + base64url43B(),
				"deployments/sm/secrets/postgres-url.secret":           "postgres://sm_database_user:sm_database_pass-" + base64url43B() + "@shared-postgres-leader:5432/sm_database",
			},
		},
		{
			name: "valid suite deployment",
			setupFiles: map[string]string{
				"deployments/cryptoutil/secrets/hash-pepper-v3.secret":         "cryptoutil-hash-pepper-v3-" + base64url43A(),
				"deployments/cryptoutil/secrets/browser-username.secret.never": NeverMarkerSuite,
				testDeploymentsCryptoutilSecretsBrowserPasswordSecretNever:     NeverMarkerSuite,
				"deployments/cryptoutil/secrets/service-username.secret.never": NeverMarkerSuite,
				testDeploymentsCryptoutilSecretsServicePasswordSecretNever:     NeverMarkerSuite,
				"deployments/cryptoutil/secrets/postgres-database.secret":      "cryptoutil_database",
				"deployments/cryptoutil/secrets/postgres-username.secret":      "cryptoutil_database_user",
				"deployments/cryptoutil/secrets/postgres-password.secret":      "cryptoutil_database_pass-" + base64url43C(),
				"deployments/cryptoutil/secrets/postgres-url.secret":           "postgres://cryptoutil_database_user:cryptoutil_database_pass-" + base64url43C() + "@shared-postgres-leader:5432/cryptoutil_database",
			},
		},
		{
			name: "infrastructure deployment is skipped",
			setupFiles: map[string]string{
				"deployments/shared-postgres/secrets/postgres-database.secret": "cryptoutil_db",
				"deployments/shared-postgres/secrets/postgres-username.secret": "cryptoutil_admin",
				"deployments/shared-postgres/secrets/postgres-password.secret": "CHANGE_ME_IN_PRODUCTION",
			},
		},
		{
			name: "deployment without secrets dir is skipped",
			setupFiles: map[string]string{
				"deployments/sm-kms/.gitkeep": "",
			},
		},
		{
			name: "content with trailing newline is valid",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsHashPepperV3Secret: "sm-kms-hash-pepper-v3-" + base64url43A() + "\n",
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

func TestFindViolationsInDir_InvalidContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFiles     map[string]string
		wantViolations int
		wantSubstring  string
	}{
		{
			name: "wrong prefix in hash pepper",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsHashPepperV3Secret: "wrong-hash-pepper-v3-" + base64url43A(),
			},
			wantViolations: 1,
			wantSubstring:  testDoesNotMatchExpectedPattern,
		},
		{
			name: "empty hash pepper file",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsHashPepperV3Secret: "",
			},
			wantViolations: 1,
			wantSubstring:  cryptoutilSharedMagic.EMPTY,
		},
		{
			name: "wrong browser username",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsBrowserUsernameSecret: testWrongBrowserUser,
			},
			wantViolations: 1,
			wantSubstring:  testDoesNotMatchExpectedPattern,
		},
		{
			name: "wrong browser password prefix",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsBrowserPasswordSecret: "wrong-browser-pass-" + base64url43A(),
			},
			wantViolations: 1,
			wantSubstring:  testDoesNotMatchExpectedPattern,
		},
		{
			name: "browser password too short base64url",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsBrowserPasswordSecret: "sm-kms-browser-pass-abc123",
			},
			wantViolations: 1,
			wantSubstring:  testDoesNotMatchExpectedPattern,
		},
		{
			name: "wrong service username",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsServiceUsernameSecret: "wrong-service-user",
			},
			wantViolations: 1,
			wantSubstring:  testDoesNotMatchExpectedPattern,
		},
		{
			name: "wrong service password prefix",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsServicePasswordSecret: "wrong-service-pass-" + base64url43A(),
			},
			wantViolations: 1,
			wantSubstring:  testDoesNotMatchExpectedPattern,
		},
		{
			name: "wrong postgres database prefix",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsPostgresDatabaseSecret: testWrongDatabase,
			},
			wantViolations: 1,
			wantSubstring:  testDoesNotMatchExpectedPattern,
		},
		{
			name: "wrong postgres username prefix",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsPostgresUsernameSecret: "wrong_database_user",
			},
			wantViolations: 1,
			wantSubstring:  testDoesNotMatchExpectedPattern,
		},
		{
			name: "wrong postgres password prefix",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsPostgresPasswordSecret: "wrong_database_pass-" + base64url43A(),
			},
			wantViolations: 1,
			wantSubstring:  testDoesNotMatchExpectedPattern,
		},
		{
			name: "wrong postgres url host",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsPostgresURLSecret: "postgres://sm_kms_database_user:sm_kms_database_pass-" + base64url43A() + "@wrong-postgres:5432/sm_kms_database",
			},
			wantViolations: 1,
			wantSubstring:  testDoesNotMatchExpectedPattern,
		},
		{
			name: "empty postgres url",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsPostgresURLSecret: "",
			},
			wantViolations: 1,
			wantSubstring:  cryptoutilSharedMagic.EMPTY,
		},
		{
			name: "wrong never marker at product tier",
			setupFiles: map[string]string{
				testDeploymentsSmSecretsBrowserUsernameSecretNever: testWrongContent,
			},
			wantViolations: 1,
			wantSubstring:  testExpected,
		},
		{
			name: "wrong never marker at suite tier",
			setupFiles: map[string]string{
				testDeploymentsCryptoutilSecretsServicePasswordSecretNever: testWrongContent,
			},
			wantViolations: 1,
			wantSubstring:  testExpected,
		},
		{
			name: "product tier marker has suite text",
			setupFiles: map[string]string{
				testDeploymentsSmSecretsBrowserPasswordSecretNever: NeverMarkerSuite,
			},
			wantViolations: 1,
			wantSubstring:  testExpected,
		},
		{
			name: "suite tier marker has product text",
			setupFiles: map[string]string{
				testDeploymentsCryptoutilSecretsBrowserPasswordSecretNever: NeverMarkerProduct,
			},
			wantViolations: 1,
			wantSubstring:  testExpected,
		},
		{
			name: "multiple violations in one deployment",
			setupFiles: map[string]string{
				testDeploymentsSmKMSSecretsHashPepperV3Secret:     "",
				testDeploymentsSmKMSSecretsBrowserUsernameSecret:  testWrongBrowserUser,
				testDeploymentsSmKMSSecretsPostgresDatabaseSecret: testWrongDatabase,
			},
			wantViolations: 3,
			wantSubstring:  "",
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
					if containsSubstring(v, tc.wantSubstring) {
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

func TestFindViolationsInDir_NonDirectoryEntryAndUnknownDeploymentSkipped(t *testing.T) {
	t.Parallel()

	rootDir := setupTempDir(t, map[string]string{
		"deployments/README.md":                           "deployment notes",
		"deployments/unregistered/secrets/example.secret": "ignored",
	})

	violations, err := FindViolationsInDir(rootDir)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckInDir_MissingDeploymentsDirectory(t *testing.T) {
	t.Parallel()

	err := CheckInDir(cryptoutilCmdCicdCommon.NewLogger("test"), t.TempDir())
	require.ErrorContains(t, err, "failed to check secret content")
}

func TestCheckInDir_WithViolation(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		testDeploymentsSmKMSSecretsHashPepperV3Secret: "wrong-hash-pepper-v3-" + base64url43A(),
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
		testDeploymentsSmKMSSecretsHashPepperV3Secret: "sm-kms-hash-pepper-v3-" + base64url43A(),
	}
	rootDir := setupTempDir(t, files)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)
	if err != nil {
		require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "unexpected error: %v", err)
	}
}

// Sequential: modifies the process working directory, which is global state.
func TestCheck_DelegatesToWorkspaceRoot(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "cannot get cwd: %v", err)
	}

	projectRoot := filepath.Join(orig, "..", "..", "..", "..", "..")

	if err := os.Chdir(projectRoot); err != nil {
		require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "cannot chdir to project root: %v", err)
	}

	defer func() { _ = os.Chdir(orig) }()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	if err := Check(logger); err != nil {
		require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "unexpected violation in real workspace: %v", err)
	}
}

// Sequential: modifies package-level secretSchemasYAML.
func TestFindViolationsInDir_InvalidYAML(t *testing.T) {
	orig := secretSchemasYAML
	secretSchemasYAML = []byte("not: valid: yaml: [unclosed")

	defer func() { secretSchemasYAML = orig }()

	tmpDir := t.TempDir()

	_, err := FindViolationsInDir(tmpDir)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestFindViolationsInDir_ReadError(t *testing.T) {
	t.Parallel()

	// Create a directory where the secret file is expected, so ReadFile returns
	// a non-NotExist error (is-a-directory error).
	rootDir := t.TempDir()
	secretsDir := filepath.Join(rootDir, "deployments", cryptoutilSharedMagic.KMSServiceID, "secrets")

	if err := os.MkdirAll(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute); err != nil {
		require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "failed to create secrets dir: %v", err)
	}

	// Create a directory with the secret file name (forces read error).
	dirAsFile := filepath.Join(secretsDir, "hash-pepper-v3.secret")
	if err := os.Mkdir(dirAsFile, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute); err != nil {
		require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "failed to create dir-as-file: %v", err)
	}

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "unexpected error: %v", err)
	}

	if len(violations) != 1 {
		require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "expected 1 violation for read error, got %d: %v", len(violations), violations)
	}

	if !containsSubstring(violations[0], "failed to read") {
		assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected violation about read failure, got: %q", violations[0])
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

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}

// Unique 43-char base64url values for testing (32 random bytes base64url-encoded without padding).
func base64url43A() string {
	return "V5Oa5USQnAu2UpPS0keFoQuLyEJ3nR2Xptwq2fODkQ4"
}

func base64url43B() string {
	return "7xIwPIo-c7W6wnWtbSB97VllYD8SyS8Zg-3rCcF0ba4"
}

func base64url43C() string {
	return "sBfVvCzrL62XOZXqzXMs1xO2rAaic-w-OnEkj7Bfa7w"
}

func base64url43D() string {
	return "KcgUfvNiZf2NdKPiXMw_nYFd1u-xPfZbiLS5ZlgaEIg"
}
