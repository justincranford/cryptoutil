// Copyright (c) 2025 Justin Cranford

package gen_config_initialisms

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), cryptoutilSharedMagic.FilePermOwnerReadWriteOnly)
	require.NoError(t, err)

	return path
}

func makeTestAPIDir(t *testing.T) (rootDir, apiDir string) {
	t.Helper()
	rootDir = t.TempDir()
	apiDir = filepath.Join(rootDir, "api")
	require.NoError(t, os.MkdirAll(apiDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	return rootDir, apiDir
}

var fullBaseList = "output-options:\n  name-normalizer: ToCamelCaseWithInitialisms\n  additional-initialisms:\n    - IDS\n    - JWT\n    - JWK\n    - JWE\n    - JWS\n    - OIDC\n    - SAML\n    - AES\n    - GCM\n    - CBC\n    - RSA\n    - EC\n    - HMAC\n    - SHA\n    - TLS\n    - IP\n    - AI\n    - ML\n    - KEM\n    - PEM\n    - DER\n    - DSA\n    - IKM\n"

func TestCheckMissingInitialisms_AllPresent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTestFile(t, dir, genConfigServerFileName, fullBaseList)
	missing, err := checkMissingInitialisms(path)
	require.NoError(t, err)
	require.Empty(t, missing)
}

func TestCheckMissingInitialisms_MissingOne(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// All base initialisms except DSA.
	noDSA := "output-options:\n  additional-initialisms:\n    - IDS\n    - JWT\n    - JWK\n    - JWE\n    - JWS\n    - OIDC\n    - SAML\n    - AES\n    - GCM\n    - CBC\n    - RSA\n    - EC\n    - HMAC\n    - SHA\n    - TLS\n    - IP\n    - AI\n    - ML\n    - KEM\n    - PEM\n    - DER\n    - IKM\n"
	path := writeTestFile(t, dir, genConfigServerFileName, noDSA)
	missing, err := checkMissingInitialisms(path)
	require.NoError(t, err)
	require.Equal(t, []string{"DSA"}, missing)
}

func TestCheckMissingInitialisms_EmptyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTestFile(t, dir, genConfigServerFileName, "")
	missing, err := checkMissingInitialisms(path)
	require.NoError(t, err)
	require.Len(t, missing, len(baseInitialisms))
}

func TestCheckMissingInitialisms_WithDomainExtras(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	content := fullBaseList + "    - CSR\n    - CA\n"
	path := writeTestFile(t, dir, genConfigServerFileName, content)
	missing, err := checkMissingInitialisms(path)
	require.NoError(t, err)
	require.Empty(t, missing)
}

func TestCheckMissingInitialisms_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := checkMissingInitialisms("/nonexistent/path/openapi-gen_config_server.yaml")
	require.Error(t, err)
}

// TestCheckMissingInitialisms_ScannerError verifies that a line exceeding bufio.MaxScanTokenSize triggers scanner.Err().
func TestCheckMissingInitialisms_ScannerError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, genConfigServerFileName)
	// A single line exceeding 64KB triggers bufio.ErrTooLong from scanner.Err().
	hugeLine := "- IDS " + strings.Repeat("x", 70000) + "\n"
	require.NoError(t, os.WriteFile(path, []byte(hugeLine), cryptoutilSharedMagic.FilePermOwnerReadWriteOnly))

	_, err := checkMissingInitialisms(path)
	require.Error(t, err)
}

func TestCheckInDir_Clean(t *testing.T) {
	t.Parallel()
	rootDir, apiDir := makeTestAPIDir(t)
	writeTestFile(t, apiDir, genConfigServerFileName, fullBaseList)

	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.NoError(t, err)
}

func TestCheckInDir_Violation(t *testing.T) {
	t.Parallel()
	rootDir, apiDir := makeTestAPIDir(t)
	writeTestFile(t, apiDir, genConfigServerFileName, "output-options:\n  additional-initialisms:\n    - JWT\n")

	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "gen config initialisms violation")
}

func TestCheckInDir_NonServerConfigs(t *testing.T) {
	t.Parallel()
	rootDir, apiDir := makeTestAPIDir(t)
	writeTestFile(t, apiDir, "openapi-gen_config_client.yaml", "output-options:\n  name-normalizer: ToCamelCaseWithInitialisms\n")
	writeTestFile(t, apiDir, "openapi-gen_config_models.yaml", "output-options:\n  skip-prune: true\n")

	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.NoError(t, err)
}

func TestCheckInDir_SkipsGitAndVendor(t *testing.T) {
	t.Parallel()
	rootDir, apiDir := makeTestAPIDir(t)
	gitDir := filepath.Join(apiDir, cryptoutilSharedMagic.CICDExcludeDirGit)
	require.NoError(t, os.MkdirAll(gitDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	writeTestFile(t, gitDir, genConfigServerFileName, "")

	vendorDir := filepath.Join(apiDir, cryptoutilSharedMagic.CICDExcludeDirVendor)
	require.NoError(t, os.MkdirAll(vendorDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	writeTestFile(t, vendorDir, genConfigServerFileName, "")

	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.NoError(t, err)
}

func TestCheckInDir_MultipleServices(t *testing.T) {
	t.Parallel()
	rootDir, apiDir := makeTestAPIDir(t)
	svc1Dir := filepath.Join(apiDir, cryptoutilSharedMagic.OTLPServiceJoseJA)
	require.NoError(t, os.MkdirAll(svc1Dir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	writeTestFile(t, svc1Dir, genConfigServerFileName, fullBaseList+"    - JWKS\n")

	svc2Dir := filepath.Join(apiDir, cryptoutilSharedMagic.OTLPServiceSMKMS)
	require.NoError(t, os.MkdirAll(svc2Dir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	writeTestFile(t, svc2Dir, genConfigServerFileName, "output-options:\n  additional-initialisms:\n    - JWT\n")

	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.Error(t, err)
}

func TestCheckInDir_NoAPIDir(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "api directory not found")
}

func TestCheckInDir_NoServerConfigs(t *testing.T) {
	t.Parallel()
	rootDir, apiDir := makeTestAPIDir(t)
	writeTestFile(t, apiDir, "openapi-gen_config_client.yaml", "# client config\n")

	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.NoError(t, err)
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_Integration(t *testing.T) {
	projectRoot, findErr := findProjectRoot()
	if findErr != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(projectRoot))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms-integration")
	err = Check(logger)
	require.NoError(t, err, "All server gen configs should contain the base initialisms list")
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}

		dir = parent
	}
}

// TestCheckInDir_WalkError verifies error propagation when WalkDir itself fails.
// Sequential: mutates package-level filepathWalkDir seam.
func TestCheckInDir_WalkError(t *testing.T) {
	origWalkDir := filepathWalkDir

	defer func() { filepathWalkDir = origWalkDir }()

	filepathWalkDir = func(_ string, _ fs.WalkDirFunc) error {
		return fmt.Errorf("mock walk error")
	}

	rootDir, _ := makeTestAPIDir(t)
	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to walk api directory")
}

// TestCheckInDir_FileCheckError verifies error propagation when checkMissingInitialisms returns an error.
// Sequential: mutates package-level checkMissingInitialismsFunc seam.
func TestCheckInDir_FileCheckError(t *testing.T) {
	orig := checkMissingInitialismsFunc

	defer func() { checkMissingInitialismsFunc = orig }()

	checkMissingInitialismsFunc = func(_ string) ([]string, error) {
		return nil, fmt.Errorf("mock file read error")
	}

	rootDir, apiDir := makeTestAPIDir(t)
	writeTestFile(t, apiDir, genConfigServerFileName, fullBaseList)

	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to walk api directory")
}

// TestCheckInDir_WalkCallbackEntryError verifies entry-level error in WalkDir callback.
// Sequential: mutates package-level filepathWalkDir seam.
func TestCheckInDir_WalkCallbackEntryError(t *testing.T) {
	origWalkDir := filepathWalkDir

	defer func() { filepathWalkDir = origWalkDir }()

	filepathWalkDir = func(root string, fn fs.WalkDirFunc) error {
		// Simulate a WalkDir callback error for a directory entry.
		return fn(root, nil, fmt.Errorf("mock entry error"))
	}

	rootDir, _ := makeTestAPIDir(t)
	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to walk api directory")
}
