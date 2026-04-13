// Copyright (c) 2025 Justin Cranford

package secrets_compliance

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// projectRoot returns the path to the project root for integration tests.
func projectRoot() string {
	return filepath.Join("..", "..", "..", "..", "..", "..")
}

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	// Integration test: run against the real project root (6 levels up from this file).
	logger := cryptoutilCmdCicdCommon.NewLogger("test-secrets-compliance")
	err := checkInDir(logger, projectRoot(), defaultComplianceFn)
	require.NoError(t, err)
}

func TestCheckInDir_Success(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-secrets-compliance")
	called := false

	err := checkInDir(logger, ".", func(_ string) ([]string, error) {
		called = true

		return nil, nil
	})

	require.NoError(t, err)
	require.True(t, called)
}

func TestCheckInDir_Violations(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-secrets-compliance")

	err := checkInDir(logger, ".", func(_ string) ([]string, error) {
		return []string{"sm-kms/secrets/unseal-1of5.secret: missing expected secret file"}, nil
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "secrets-compliance violations")
	require.Contains(t, err.Error(), "unseal-1of5.secret")
}

func TestCheckInDir_LoadError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-secrets-compliance")

	err := checkInDir(logger, ".", func(_ string) ([]string, error) {
		return nil, os.ErrPermission
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to check secrets compliance")
}

func TestCheckSecretsDir_AllPresent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := filepath.Join(tmpDir, "secrets")
	require.NoError(t, os.MkdirAll(secretsDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	// Create all 14 expected PS-ID secrets files.
	for _, f := range expectedPSIDSecrets {
		require.NoError(t, os.WriteFile(filepath.Join(secretsDir, f), []byte("test-content"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	violations := checkSecretsDir(secretsDir, cryptoutilSharedMagic.OTLPServiceSMKMS, expectedPSIDSecrets)
	require.Empty(t, violations)
}

func TestCheckSecretsDir_MissingFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := filepath.Join(tmpDir, "secrets")
	require.NoError(t, os.MkdirAll(secretsDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	// Create all files except unseal-1of5.secret.
	for _, f := range expectedPSIDSecrets[1:] {
		require.NoError(t, os.WriteFile(filepath.Join(secretsDir, f), []byte("test-content"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	violations := checkSecretsDir(secretsDir, cryptoutilSharedMagic.OTLPServiceSMKMS, expectedPSIDSecrets)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0], "unseal-1of5.secret")
	require.Contains(t, violations[0], "missing expected secret file")
}

func TestCheckSecretsDir_UnexpectedFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := filepath.Join(tmpDir, "secrets")
	require.NoError(t, os.MkdirAll(secretsDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	// Create all expected files.
	for _, f := range expectedPSIDSecrets {
		require.NoError(t, os.WriteFile(filepath.Join(secretsDir, f), []byte("test-content"), cryptoutilSharedMagic.CacheFilePermissions))
	}
	// Add an unexpected .secret file.
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "unexpected.secret"), []byte("oops"), cryptoutilSharedMagic.CacheFilePermissions))

	violations := checkSecretsDir(secretsDir, cryptoutilSharedMagic.OTLPServiceSMKMS, expectedPSIDSecrets)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0], "unexpected.secret")
	require.Contains(t, violations[0], "unexpected secret file")
}

func TestCheckSecretsDir_NonSecretFileAllowed(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := filepath.Join(tmpDir, "secrets")
	require.NoError(t, os.MkdirAll(secretsDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	// Create all expected files.
	for _, f := range expectedPSIDSecrets {
		require.NoError(t, os.WriteFile(filepath.Join(secretsDir, f), []byte("test-content"), cryptoutilSharedMagic.CacheFilePermissions))
	}
	// Add a non-secret file (should be allowed).
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "README.md"), []byte("# docs"), cryptoutilSharedMagic.CacheFilePermissions))

	violations := checkSecretsDir(secretsDir, cryptoutilSharedMagic.OTLPServiceSMKMS, expectedPSIDSecrets)
	require.Empty(t, violations)
}

func TestCheckSecretsDir_MissingDirectory(t *testing.T) {
	t.Parallel()

	secretsDir := filepath.Join(t.TempDir(), "nonexistent", "secrets")

	violations := checkSecretsDir(secretsDir, cryptoutilSharedMagic.OTLPServiceSMKMS, expectedPSIDSecrets)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0], "directory does not exist")
}

func TestCheckSecretsDir_ProductSecrets(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := filepath.Join(tmpDir, "secrets")
	require.NoError(t, os.MkdirAll(secretsDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	// Create all expected product secrets files.
	for _, f := range expectedProductSecrets {
		require.NoError(t, os.WriteFile(filepath.Join(secretsDir, f), []byte("test-content"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	violations := checkSecretsDir(secretsDir, cryptoutilSharedMagic.SMProductName, expectedProductSecrets)
	require.Empty(t, violations)
}

func TestCheckSecretsDir_SuiteSecrets(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := filepath.Join(tmpDir, "secrets")
	require.NoError(t, os.MkdirAll(secretsDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	// Create all expected suite secrets files.
	for _, f := range expectedSuiteSecrets {
		require.NoError(t, os.WriteFile(filepath.Join(secretsDir, f), []byte("test-content"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	violations := checkSecretsDir(secretsDir, cryptoutilSharedMagic.DefaultOTLPServiceDefault, expectedSuiteSecrets)
	require.Empty(t, violations)
}

func TestExpectedPSIDSecrets_Count(t *testing.T) {
	t.Parallel()

	require.Len(t, expectedPSIDSecrets, 15, "PS-ID secrets must have exactly 15 files (Decision 14: 14 + issuing-ca-key)")
}

func TestExpectedProductSecrets_Count(t *testing.T) {
	t.Parallel()

	require.Len(t, expectedProductSecrets, 15, "product secrets must have exactly 15 files")
}

func TestExpectedSuiteSecrets_Count(t *testing.T) {
	t.Parallel()

	require.Len(t, expectedSuiteSecrets, 15, "suite secrets must have exactly 15 files")
}
